package agentkit

import (
	"net/url"
	"strings"
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
		if !strings.EqualFold(k, "key") {
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
