package agentkit

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/AgoraIO/agora-agents-go/v2/core"
)

// Redaction for debug session logging.
//
// The debug dump is routinely pasted into issues and chat threads, so it must
// never carry a usable credential. Nothing redacted the request body, which is
// where vendor API keys and RTC tokens live.

// Redacted is the marker substituted for a redacted value.
const Redacted = "[REDACTED]"

// sensitiveBodyKeys are body fields whose values are credentials or account
// identifiers. Compared case-insensitively, and matched on both snake_case and
// camelCase spellings so this keeps working if a caller hand-builds a config.
var sensitiveBodyKeys = map[string]struct{}{
	// Vendor credentials
	"api_key":                {},
	"apikey":                 {},
	"key":                    {},
	"secret":                 {},
	"api_secret":             {},
	"apisecret":              {},
	"password":               {},
	"access_key_id":          {},
	"accesskeyid":            {},
	"secret_access_key":      {},
	"secretaccesskey":        {},
	"adc_credentials_string": {},
	"adccredentialsstring":   {},
	"credentials":            {},
	"subscription_key":       {},
	"subscriptionkey":        {},
	// Agora credentials and account identifiers
	"token":           {},
	"agora_token":     {},
	"agoratoken":      {},
	"authorization":   {},
	"headers":         {},
	"x-api-key":       {},
	"x_api_key":       {},
	"cookie":          {},
	"set-cookie":      {},
	"appid":           {},
	"app_id":          {},
	"agora_appid":     {},
	"agoraappid":      {},
	"app_certificate": {},
	"appcertificate":  {},
	"customer_secret": {},
	"customersecret":  {},
}

func isSensitiveKey(key string) bool {
	_, found := sensitiveBodyKeys[strings.ToLower(key)]
	return found
}

// redactQueryKeys strips Gemini-style `key=` query values from URL strings.
func redactQueryKeys(value string) string {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.RawQuery == "" {
		return value
	}
	query := parsed.Query()
	changed := false
	for k, vals := range query {
		if !isSensitiveKey(k) {
			continue
		}
		for i, v := range vals {
			if v != "" {
				vals[i] = Redacted
				changed = true
			}
		}
	}
	if !changed {
		return value
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

// redactURL removes credentials in query parameters and project IDs in API paths.
func redactURL(value string) string {
	value = redactQueryKeys(value)
	parsed, err := url.Parse(value)
	if err != nil {
		return value
	}
	segments := strings.Split(parsed.Path, "/")
	for i, segment := range segments {
		if segment != "projects" {
			continue
		}
		projectIndex := i + 1
		if projectIndex < len(segments) && segments[projectIndex] == "apps" {
			projectIndex++
		}
		if projectIndex < len(segments) && segments[projectIndex] != "" {
			segments[projectIndex] = Redacted
		}
		break
	}
	parsed.Path = strings.Join(segments, "/")
	return parsed.String()
}

// debugHTTPClient logs the final HTTP request just before it leaves the SDK.
// It sits below request options and preview routing so the trace includes the
// resolved URL and all SDK-added headers.
type debugHTTPClient struct {
	inner core.HTTPClient
}

func (c *debugHTTPClient) Do(req *http.Request) (*http.Response, error) {
	logDebugHTTPRequest(req)
	return c.inner.Do(req)
}

func logDebugHTTPRequest(req *http.Request) {
	body, readErr := readRequestBody(req)
	payload := map[string]interface{}{
		"method":         req.Method,
		"url":            redactURL(req.URL.Redacted()),
		"headers":        redactHTTPHeaders(req.Header),
		"content_length": req.ContentLength,
	}
	if readErr != nil {
		payload["body"] = "[unavailable: failed to read request body]"
	} else {
		payload["body"] = redactHTTPBody(body)
	}

	if encoded, err := json.Marshal(payload); err == nil {
		log.Printf("[Agora Debug] REST request: %s", encoded)
	} else {
		log.Printf("[Agora Debug] REST request logging failed: %v", err)
	}
}

func readRequestBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}
	// Reading the live body can consume streaming requests or swallow read errors.
	// Use the replay copy when available; otherwise leave the stream untouched.
	if req.GetBody == nil {
		return nil, errors.New("request body is not replayable")
	}
	copy, err := req.GetBody()
	if err != nil {
		return nil, err
	}
	defer copy.Close()
	return io.ReadAll(copy)
}

func redactHTTPHeaders(headers http.Header) map[string][]string {
	redacted := make(map[string][]string, len(headers))
	for key, values := range headers {
		copied := append([]string(nil), values...)
		if isSensitiveHeaderKey(key) {
			for i, value := range copied {
				if value != "" {
					copied[i] = Redacted
				}
			}
		}
		// HTTP field names are case-insensitive (and HTTP/2 serializes them in
		// lowercase). Normalize the display too, so a trace reflects the wire
		// convention instead of Go's canonical map-key spelling.
		key = strings.ToLower(key)
		redacted[key] = append(redacted[key], copied...)
	}
	return redacted
}

func isSensitiveHeaderKey(key string) bool {
	key = strings.ToLower(key)
	if isSensitiveKey(key) {
		return true
	}
	return strings.Contains(key, "authorization") ||
		strings.Contains(key, "api-key") ||
		strings.Contains(key, "api_key") ||
		strings.Contains(key, "token") ||
		strings.Contains(key, "secret") ||
		strings.Contains(key, "credential") ||
		strings.Contains(key, "cookie")
}

func redactHTTPBody(body []byte) interface{} {
	if len(body) == 0 {
		return nil
	}
	var decoded interface{}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return "[unavailable: non-JSON request body]"
	}
	return RedactSecrets(decoded)
}

// RedactSecrets deep-copies value, replacing credential fields with Redacted.
//
// Empty strings are left visible on purpose: "" is the signature of an unset
// environment variable, and hiding it behind [REDACTED] would disguise the exact
// misconfiguration the debug output exists to surface.
//
// Never mutates the input — the request that goes on the wire is untouched.
func RedactSecrets(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			if str, ok := item.(string); ok && isSensitiveKey(key) && str != "" {
				result[key] = Redacted
				continue
			}
			if str, ok := item.(string); ok {
				result[key] = redactQueryKeys(str)
				continue
			}
			result[key] = RedactSecrets(item)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(typed))
		for i, item := range typed {
			result[i] = RedactSecrets(item)
		}
		return result
	case []map[string]interface{}:
		result := make([]interface{}, len(typed))
		for i, item := range typed {
			result[i] = RedactSecrets(item)
		}
		return result
	default:
		return value
	}
}
