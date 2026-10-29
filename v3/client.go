package v3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/AgoraIO/agora-agents-go/v2/core"
	"github.com/AgoraIO/agora-agents-go/v2/internal"
	"github.com/AgoraIO/agora-agents-go/v2/option"
)

const defaultHost = "https://preview.ai.agora.io"
const apiVersionPath = "/conversational-ai/v1/projects/"

type Client struct {
	Agents    *AgentsClient
	Secrets   *SecretsClient
	Numbers   *NumbersClient
	Sessions  *SessionsClient
	Campaigns *CampaignsClient
	shared    *transport
}
type transport struct {
	appID     string
	baseURL   string
	originErr error
	options   *core.RequestOptions
	caller    *internal.Caller
}

func (t *transport) forLatestRead() *transport {
	clone := *t
	options := *t.options
	options.HTTPHeader = t.options.HTTPHeader.Clone()
	for name := range options.HTTPHeader {
		if strings.EqualFold(name, "If-Match") || strings.EqualFold(name, "If-None-Match") || strings.EqualFold(name, "Idempotency-Key") {
			delete(options.HTTPHeader, name)
		}
	}
	clone.options = &options
	return &clone
}

// PreconditionFailedError retains a 412 and can fetch the latest resource.
type PreconditionFailedError struct {
	Cause error
	fetch func(context.Context) (any, string, error)
}

func (e *PreconditionFailedError) Error() string {
	return "v3: precondition failed: " + e.Cause.Error()
}
func (e *PreconditionFailedError) Unwrap() error { return e.Cause }
func (e *PreconditionFailedError) Latest(ctx context.Context) (any, string, error) {
	if e.fetch == nil {
		return nil, "", errors.New("v3: latest resource unavailable")
	}
	return e.fetch(ctx)
}
func precondition(ctx context.Context, err error, fetch func(context.Context) (any, string, error)) error {
	_ = ctx
	var api *core.APIError
	if !errors.As(err, &api) || api.StatusCode != http.StatusPreconditionFailed {
		return err
	}
	return &PreconditionFailedError{Cause: err, fetch: fetch}
}

// NewClient takes the project App ID and standard SDK request options.
// Base URL overrides are origins; the v3 project path is appended by this client.
func NewClient(appID string, opts ...option.RequestOption) *Client {
	o := core.NewRequestOptions(opts...)
	base := o.BaseURL
	if env := os.Getenv("AGORA_AGENTS_V3_BASE_URL"); env != "" && base == "" {
		base = env
	}
	if base == "" {
		base = defaultHost
	}
	base = strings.TrimRight(base, "/")
	parsedBase, parseErr := url.Parse(base)
	var originErr error
	if parseErr != nil || parsedBase == nil || (parsedBase.Scheme != "http" && parsedBase.Scheme != "https") || parsedBase.Host == "" || (parsedBase.Path != "" && parsedBase.Path != "/") || parsedBase.RawQuery != "" || parsedBase.ForceQuery || parsedBase.Fragment != "" || strings.Contains(base, "#") || parsedBase.User != nil {
		originErr = fmt.Errorf("v3: base URL must be an http(s) origin without credentials, path, query, or fragment")
	}
	t := &transport{appID: appID, baseURL: base, originErr: originErr, options: o, caller: internal.NewCaller(&internal.CallerParams{Client: o.HTTPClient, MaxAttempts: o.MaxAttempts})}
	return &Client{Agents: &AgentsClient{t}, Secrets: &SecretsClient{t}, Numbers: &NumbersClient{t}, Sessions: &SessionsClient{t}, Campaigns: &CampaignsClient{t}, shared: t}
}

// ForAppID binds a project to a low-level client created without one.
func (c *Client) ForAppID(appID string) *Client {
	t := *c.shared
	t.appID = appID
	return &Client{Agents: &AgentsClient{&t}, Secrets: &SecretsClient{&t}, Numbers: &NumbersClient{&t}, Sessions: &SessionsClient{&t}, Campaigns: &CampaignsClient{&t}, shared: &t}
}

func segment(s, label string) (string, error) {
	if strings.TrimSpace(s) == "" {
		return "", fmt.Errorf("v3: %s is required", label)
	}
	if s == "." || s == ".." {
		return "", fmt.Errorf("v3: %s must not be a dot path segment", label)
	}
	return url.PathEscape(s), nil
}
func (t *transport) path(resource string, ids ...string) (string, error) {
	a, err := segment(t.appID, "appID")
	if err != nil {
		return "", err
	}
	u := t.baseURL + apiVersionPath + a + "/" + resource
	for _, id := range ids {
		p, e := segment(id, "resource ID")
		if e != nil {
			return "", e
		}
		u += "/" + p
	}
	return u, nil
}
func (t *transport) call(ctx context.Context, method, path string, query url.Values, body, output any, headers http.Header, maxAttempts uint) (Response[any], error) {
	if t.originErr != nil {
		return Response[any]{}, t.originErr
	}
	if path == "" {
		return Response[any]{}, errors.New("v3: empty endpoint")
	}
	base := t.options.ToHeader()
	base = internal.MergeHeaders(base, headers)
	if err := core.ValidatePublicAuthorization(base); err != nil {
		return Response[any]{}, err
	}
	if hasHeader(base, "Idempotency-Key") {
		for name, values := range base {
			if !strings.EqualFold(name, "Idempotency-Key") {
				continue
			}
			for _, value := range values {
				if err := validateIdempotencyKey(value); err != nil {
					return Response[any]{}, err
				}
			}
		}
	}
	if method == http.MethodPost && (strings.HasSuffix(path, "/agents") || strings.HasSuffix(path, "/secrets") || strings.HasSuffix(path, "/numbers")) && hasHeader(base, "Idempotency-Key") {
		return Response[any]{}, errors.New("v3: Idempotency-Key is not supported for agent, secret, or number creation")
	}
	for name := range base {
		lower := strings.ToLower(name)
		if strings.HasPrefix(lower, "x-studio-") || lower == "x-consumer-username" {
			return Response[any]{}, fmt.Errorf("v3: internal header %q is not supported", name)
		}
	}
	if body != nil {
		base.Set("Content-Type", "application/json")
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	if method != http.MethodGet {
		maxAttempts = 1
	}
	request := body
	if body != nil {
		request = &body
	} // the shared caller expects a nil-able request value
	raw, err := t.caller.Call(ctx, &internal.CallParams{URL: path, Method: method, Headers: base, Request: request, Response: output, ResponseIsOptional: output == nil, MaxAttempts: maxAttempts, Client: t.options.HTTPClient})
	if err != nil {
		return Response[any]{}, err
	}
	return Response[any]{Body: output, Header: raw.Header.Clone(), ETag: raw.Header.Get("ETag"), RequestID: raw.Header.Get("Agora-Request-Id"), StatusCode: raw.StatusCode}, nil
}

func validateIdempotencyKey(value string) error {
	if len(value) < 1 || len(value) > 256 {
		return errors.New("v3: Idempotency-Key must contain 1-256 printable ASCII characters")
	}
	for _, char := range value {
		if char < 0x21 || char > 0x7e {
			return errors.New("v3: Idempotency-Key must contain 1-256 printable ASCII characters")
		}
	}
	return nil
}

func hasHeader(headers http.Header, target string) bool {
	for name := range headers {
		if strings.EqualFold(name, target) {
			return true
		}
	}
	return false
}

func listQuery(o ListOptions) url.Values {
	q := url.Values{}
	if o.Limit > 0 {
		q.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.Cursor != "" {
		q.Set("cursor", o.Cursor)
	}
	if o.Order != "" {
		q.Set("order", o.Order)
	}
	return q
}
func etagHeader(v string) http.Header {
	h := http.Header{}
	if v != "" {
		h.Set("If-Match", v)
	}
	return h
}
func requiredETag(v string) error {
	if strings.TrimSpace(v) == "" {
		return errors.New("v3: If-Match ETag is required")
	}
	return nil
}
func response[T any](r Response[any], body T) *Response[T] {
	return &Response[T]{Body: body, Header: r.Header.Clone(), ETag: r.ETag, RequestID: r.RequestID, StatusCode: r.StatusCode}
}

type AgentsClient struct{ t *transport }

func (c *AgentsClient) Create(ctx context.Context, body AgentCreate) (*Response[Agent], error) {
	if err := validatePreviewPipeline(body.Pipeline); err != nil {
		return nil, err
	}
	p, e := c.t.path("agents")
	if e != nil {
		return nil, e
	}
	var out Agent
	// Agent create does not declare Idempotency-Key in the public contract.
	r, e := c.t.call(ctx, http.MethodPost, p, nil, body, &out, nil, 1)
	if e != nil {
		return nil, e
	}
	return response(r, out), nil
}
func (c *AgentsClient) List(ctx context.Context, opts ListOptions) (*Response[Page[AgentListItem]], error) {
	p, e := c.t.path("agents")
	if e != nil {
		return nil, e
	}
	var out Page[AgentListItem]
	r, e := c.t.call(ctx, http.MethodGet, p, listQuery(opts), nil, &out, nil, 0)
	if e != nil {
		return nil, e
	}
	return response(r, out), nil
}
func (c *AgentsClient) Get(ctx context.Context, id string) (*Response[Agent], error) {
	p, e := c.t.path("agents", id)
	if e != nil {
		return nil, e
	}
	var out Agent
	r, e := c.t.call(ctx, http.MethodGet, p, nil, nil, &out, nil, 0)
	if e != nil {
		return nil, e
	}
	return response(r, out), nil
}
func (c *AgentsClient) Update(ctx context.Context, id, ifMatch string, patch AgentPatch) (*Response[Agent], error) {
	if err := validatePreviewPipeline(patch["pipeline"]); err != nil {
		return nil, err
	}
	if e := requiredETag(ifMatch); e != nil {
		return nil, e
	}
	p, e := c.t.path("agents", id)
	if e != nil {
		return nil, e
	}
	var out Agent
	r, e := c.t.call(ctx, http.MethodPatch, p, nil, patch, &out, etagHeader(ifMatch), 0)
	if e != nil {
		return nil, precondition(ctx, e, func(ctx context.Context) (any, string, error) {
			r, err := (&AgentsClient{t: c.t.forLatestRead()}).Get(ctx, id)
			if err != nil {
				return nil, "", err
			}
			return r.Body, r.ETag, nil
		})
	}
	return response(r, out), nil
}
func (c *AgentsClient) Delete(ctx context.Context, id string) (*Response[Deleted], error) {
	p, e := c.t.path("agents", id)
	if e != nil {
		return nil, e
	}
	var out Deleted
	r, e := c.t.call(ctx, http.MethodDelete, p, nil, nil, &out, nil, 0)
	if e != nil {
		return nil, e
	}
	return response(r, out), nil
}

type SecretsClient struct{ t *transport }

func validateSecretName(value string) error {
	if len(value) < 1 || len(value) > 64 {
		return errors.New("v3: secret set/key names must be 1-64 ASCII letters, digits, or underscores")
	}
	for _, char := range value {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return errors.New("v3: secret set/key names must be 1-64 ASCII letters, digits, or underscores")
		}
	}
	return nil
}

func validateSecretMap(values map[string]string) error {
	if len(values) == 0 {
		return errors.New("v3: secret map must contain at least one key")
	}
	return validateSecretMapKeys(values)
}

func validateSecretMapKeys[T any](values map[string]T) error {
	for key := range values {
		if err := validateSecretName(key); err != nil {
			return fmt.Errorf("v3: invalid secret key %q: %w", key, err)
		}
	}
	return nil
}

func validateSecretUpdateKeys(values map[string]*string) error {
	for key, value := range values {
		if err := validateSecretName(key); err == nil {
			continue
		}
		if value == nil && validLegacySecretKey(key) {
			continue
		}
		return fmt.Errorf("v3: secret update key %q must use ASCII letters, digits, or underscores; legacy hyphenated keys may only be deleted with null", key)
	}
	return nil
}

func validLegacySecretKey(value string) bool {
	if len(value) < 1 || len(value) > 64 || !strings.Contains(value, "-") {
		return false
	}
	for _, char := range value {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_' || char == '-') {
			return false
		}
	}
	return true
}

func (c *SecretsClient) Create(ctx context.Context, body SecretCreate, idempotencyKey string) (*Response[SecretSet], error) {
	if idempotencyKey != "" {
		return nil, errors.New("v3: idempotency keys are not supported for secret creation")
	}
	if err := validateSecretName(body.Name); err != nil {
		return nil, err
	}
	if err := validateSecretMap(body.Secrets); err != nil {
		return nil, err
	}
	p, e := c.t.path("secrets")
	if e != nil {
		return nil, e
	}
	var out SecretSet
	r, e := c.t.call(ctx, http.MethodPost, p, nil, body, &out, nil, 1)
	if e != nil {
		return nil, e
	}
	return response(r, out), nil
}
func (c *SecretsClient) List(ctx context.Context, opts ListOptions) (*Response[Page[SecretSet]], error) {
	p, e := c.t.path("secrets")
	if e != nil {
		return nil, e
	}
	var out Page[SecretSet]
	r, e := c.t.call(ctx, http.MethodGet, p, listQuery(opts), nil, &out, nil, 0)
	if e != nil {
		return nil, e
	}
	return response(r, out), nil
}
func (c *SecretsClient) Get(ctx context.Context, name string) (*Response[SecretSet], error) {
	if e := validateSecretName(name); e != nil {
		return nil, e
	}
	p, e := c.t.path("secrets", name)
	if e != nil {
		return nil, e
	}
	var out SecretSet
	r, e := c.t.call(ctx, http.MethodGet, p, nil, nil, &out, nil, 0)
	if e != nil {
		return nil, e
	}
	return response(r, out), nil
}
func (c *SecretsClient) Update(ctx context.Context, name, ifMatch string, body SecretUpdate) (*Response[SecretSet], error) {
	if err := validateSecretName(name); err != nil {
		return nil, err
	}
	if len(body.Secrets) == 0 {
		return nil, errors.New("v3: secret update requires at least one key")
	}
	if err := validateSecretUpdateKeys(body.Secrets); err != nil {
		return nil, err
	}
	p, e := c.t.path("secrets", name)
	if e != nil {
		return nil, e
	}
	var out SecretSet
	r, e := c.t.call(ctx, http.MethodPatch, p, nil, body, &out, etagHeader(ifMatch), 0)
	if e != nil {
		return nil, precondition(ctx, e, func(ctx context.Context) (any, string, error) {
			r, err := (&SecretsClient{t: c.t.forLatestRead()}).Get(ctx, name)
			if err != nil {
				return nil, "", err
			}
			return r.Body, r.ETag, nil
		})
	}
	return response(r, out), nil
}
func (c *SecretsClient) Replace(ctx context.Context, name, ifMatch string, body SecretReplace) (*Response[SecretSet], error) {
	if err := validateSecretName(name); err != nil {
		return nil, err
	}
	if err := validateSecretMap(body.Secrets); err != nil {
		return nil, err
	}
	if e := requiredETag(ifMatch); e != nil {
		return nil, e
	}
	p, e := c.t.path("secrets", name)
	if e != nil {
		return nil, e
	}
	var out SecretSet
	r, e := c.t.call(ctx, http.MethodPut, p, nil, body, &out, etagHeader(ifMatch), 0)
	if e != nil {
		return nil, precondition(ctx, e, func(ctx context.Context) (any, string, error) {
			r, err := (&SecretsClient{t: c.t.forLatestRead()}).Get(ctx, name)
			if err != nil {
				return nil, "", err
			}
			return r.Body, r.ETag, nil
		})
	}
	return response(r, out), nil
}
func (c *SecretsClient) Delete(ctx context.Context, name, ifMatch string) (*Response[Deleted], error) {
	if err := validateSecretName(name); err != nil {
		return nil, err
	}
	if e := requiredETag(ifMatch); e != nil {
		return nil, e
	}
	p, e := c.t.path("secrets", name)
	if e != nil {
		return nil, e
	}
	var out Deleted
	r, e := c.t.call(ctx, http.MethodDelete, p, nil, nil, &out, etagHeader(ifMatch), 0)
	if e != nil {
		return nil, precondition(ctx, e, func(ctx context.Context) (any, string, error) {
			r, err := (&SecretsClient{t: c.t.forLatestRead()}).Get(ctx, name)
			if err != nil {
				return nil, "", err
			}
			return r.Body, r.ETag, nil
		})
	}
	return response(r, out), nil
}
func (c *SecretsClient) SetKey(ctx context.Context, name, key string, body SecretKeyWrite, opts SecretKeyOptions) (*Response[SecretKey], error) {
	if err := validateSecretName(name); err != nil {
		return nil, err
	}
	if err := validateSecretName(key); err != nil {
		return nil, fmt.Errorf("v3: invalid secret key: %w", err)
	}
	if body.Value == "" {
		return nil, errors.New("v3: secret key value must not be empty")
	}
	p, e := c.t.path("secrets", name, "keys", key)
	if e != nil {
		return nil, e
	}
	var out SecretKey
	h := etagHeader(opts.IfMatch)
	if opts.IfNoneMatch != "" {
		h.Set("If-None-Match", opts.IfNoneMatch)
	}
	r, e := c.t.call(ctx, http.MethodPut, p, nil, body, &out, h, 0)
	if e != nil {
		return nil, precondition(ctx, e, func(ctx context.Context) (any, string, error) {
			r, err := (&SecretsClient{t: c.t.forLatestRead()}).Get(ctx, name)
			if err != nil {
				return nil, "", err
			}
			return r.Body, r.ETag, nil
		})
	}
	return response(r, out), nil
}
func (c *SecretsClient) DeleteKey(ctx context.Context, name, key, ifMatch string) (*Response[Deleted], error) {
	if err := validateSecretName(name); err != nil {
		return nil, err
	}
	if err := validateSecretName(key); err != nil {
		return nil, fmt.Errorf("v3: invalid secret key: %w", err)
	}
	p, e := c.t.path("secrets", name, "keys", key)
	if e != nil {
		return nil, e
	}
	var out Deleted
	r, e := c.t.call(ctx, http.MethodDelete, p, nil, nil, &out, etagHeader(ifMatch), 0)
	if e != nil {
		return nil, precondition(ctx, e, func(ctx context.Context) (any, string, error) {
			r, err := (&SecretsClient{t: c.t.forLatestRead()}).Get(ctx, name)
			if err != nil {
				return nil, "", err
			}
			return r.Body, r.ETag, nil
		})
	}
	return response(r, out), nil
}

type NumbersClient struct{ t *transport }

func (c *NumbersClient) Create(ctx context.Context, body NumberCreate, idempotencyKey string) (*Response[Number], error) {
	if idempotencyKey != "" {
		return nil, errors.New("v3: idempotency keys are not supported for number creation")
	}
	p, e := c.t.path("numbers")
	if e != nil {
		return nil, e
	}
	var out Number
	r, e := c.t.call(ctx, http.MethodPost, p, nil, body, &out, nil, 1)
	if e != nil {
		return nil, e
	}
	return response(r, out), nil
}
func (c *NumbersClient) List(ctx context.Context, opts ListOptions) (*Response[Page[NumberListItem]], error) {
	p, e := c.t.path("numbers")
	if e != nil {
		return nil, e
	}
	var out Page[NumberListItem]
	r, e := c.t.call(ctx, http.MethodGet, p, listQuery(opts), nil, &out, nil, 0)
	if e != nil {
		return nil, e
	}
	return response(r, out), nil
}
func (c *NumbersClient) Get(ctx context.Context, id string) (*Response[Number], error) {
	p, e := c.t.path("numbers", id)
	if e != nil {
		return nil, e
	}
	var out Number
	r, e := c.t.call(ctx, http.MethodGet, p, nil, nil, &out, nil, 0)
	if e != nil {
		return nil, e
	}
	return response(r, out), nil
}
func (c *NumbersClient) Update(ctx context.Context, id, ifMatch string, patch NumberPatch) (*Response[Number], error) {
	if e := requiredETag(ifMatch); e != nil {
		return nil, e
	}
	p, e := c.t.path("numbers", id)
	if e != nil {
		return nil, e
	}
	var out Number
	r, e := c.t.call(ctx, http.MethodPatch, p, nil, patch, &out, etagHeader(ifMatch), 0)
	if e != nil {
		return nil, precondition(ctx, e, func(ctx context.Context) (any, string, error) {
			r, err := (&NumbersClient{t: c.t.forLatestRead()}).Get(ctx, id)
			if err != nil {
				return nil, "", err
			}
			return r.Body, r.ETag, nil
		})
	}
	return response(r, out), nil
}
func (c *NumbersClient) Delete(ctx context.Context, id string) (*Response[Deleted], error) {
	p, e := c.t.path("numbers", id)
	if e != nil {
		return nil, e
	}
	var out Deleted
	r, e := c.t.call(ctx, http.MethodDelete, p, nil, nil, &out, nil, 0)
	if e != nil {
		return nil, e
	}
	return response(r, out), nil
}

type SessionsClient struct{ t *transport }

func (c *SessionsClient) List(ctx context.Context, opts SessionListOptions) (*Response[Page[Session]], error) {
	p, e := c.t.path("sessions")
	if e != nil {
		return nil, e
	}
	var out Page[Session]
	q := listQuery(opts.ListOptions)
	if opts.Channel != "" {
		q.Set("channel", opts.Channel)
	}
	if opts.Status != "" {
		q.Set("status", opts.Status)
	}
	if opts.StartedAfter != nil {
		q.Set("started_after", strconv.FormatInt(*opts.StartedAfter, 10))
	}
	if opts.StartedBefore != nil {
		q.Set("started_before", strconv.FormatInt(*opts.StartedBefore, 10))
	}
	r, e := c.t.call(ctx, http.MethodGet, p, q, nil, &out, nil, 0)
	if e != nil {
		return nil, e
	}
	for i := range out.Data.List {
		normalizeSessionResponse(&out.Data.List[i])
	}
	return response(r, out), nil
}
func (c *SessionsClient) Get(ctx context.Context, id string) (*Response[Session], error) {
	p, e := c.t.path("sessions", id)
	if e != nil {
		return nil, e
	}
	var out Session
	r, e := c.t.call(ctx, http.MethodGet, p, nil, nil, &out, nil, 0)
	if e != nil {
		return nil, e
	}
	normalizeSessionResponse(&out)
	return response(r, out), nil
}

// Update applies the session's limited runtime patch. Params are full
// replacements when included; credentials and provider selection cannot change.
func (c *SessionsClient) Update(ctx context.Context, id string, patch SessionPatch) (*Response[Session], error) {
	if patch.Transport != nil && strings.TrimSpace(patch.Transport.Token) == "" {
		return nil, errors.New("v3: session update token must be non-empty")
	}
	if patch.Agent != nil && patch.Agent.Overrides != nil && patch.Agent.Overrides.Pipeline != nil {
		if module := patch.Agent.Overrides.Pipeline.LLM; module != nil {
			for key := range module.Params {
				switch strings.ToLower(key) {
				case "url", "base_url", "api_base_url", "endpoint":
					return nil, fmt.Errorf("v3: session update cannot change provider endpoint through params.%s", key)
				}
			}
		}
	}
	p, err := c.t.path("sessions", id)
	if err != nil {
		return nil, err
	}
	var out Session
	raw, err := c.t.call(ctx, http.MethodPatch, p, nil, patch, &out, nil, 1)
	if err != nil {
		return nil, err
	}
	normalizeSessionResponse(&out)
	return response(raw, out), nil
}

// Start creates a running session from a saved agent configuration.
func (c *SessionsClient) Start(ctx context.Context, body SessionStartRequest, opts ...SessionWriteOptions) (*Response[Session], error) {
	if len(opts) > 1 {
		return nil, errors.New("v3: at most one session write options value is accepted")
	}
	headers, err := sessionWriteHeaders(opts)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(body.AgentID) == "" {
		return nil, errors.New("v3: agentID is required")
	}
	if err := validateSessionTransport(body.Transport); err != nil {
		return nil, err
	}
	if err := validateSessionLifecycle(body.Lifecycle); err != nil {
		return nil, err
	}
	if err := validateDataPolicy(body.DataPolicy); err != nil {
		return nil, err
	}
	if err := validateSessionValues(body.Values); err != nil {
		return nil, err
	}
	p, err := c.t.path("sessions")
	if err != nil {
		return nil, err
	}
	var out Session
	r, err := c.t.call(ctx, http.MethodPost, p, nil, body, &out, headers, 0)
	if err != nil {
		return nil, err
	}
	normalizeSessionResponse(&out)
	return response(r, out), nil
}

// StartEphemeral creates a running session from an inline, unsaved agent.
func (c *SessionsClient) StartEphemeral(ctx context.Context, body SessionStartEphemeralRequest, opts ...SessionWriteOptions) (*Response[Session], error) {
	if err := validatePreviewPipeline(body.Agent.Pipeline); err != nil {
		return nil, err
	}
	if len(opts) > 1 {
		return nil, errors.New("v3: at most one session write options value is accepted")
	}
	headers, err := sessionWriteHeaders(opts)
	if err != nil {
		return nil, err
	}
	if err := validateSessionTransport(body.Transport); err != nil {
		return nil, err
	}
	if err := validateSessionLifecycle(body.Lifecycle); err != nil {
		return nil, err
	}
	if err := validateDataPolicy(body.DataPolicy); err != nil {
		return nil, err
	}
	if err := validateSessionValues(body.Values); err != nil {
		return nil, err
	}
	p, err := c.t.path("sessions", "ephemeral")
	if err != nil {
		return nil, err
	}
	var out Session
	r, err := c.t.call(ctx, http.MethodPost, p, nil, body, &out, headers, 0)
	if err != nil {
		return nil, err
	}
	normalizeSessionResponse(&out)
	return response(r, out), nil
}

func sessionWriteHeaders(opts []SessionWriteOptions) (http.Header, error) {
	h := http.Header{}
	if len(opts) > 0 && opts[0].IdempotencyKey != "" {
		key := opts[0].IdempotencyKey
		if err := validateIdempotencyKey(key); err != nil {
			return nil, err
		}
		h.Set("Idempotency-Key", key)
	}
	return h, nil
}

func normalizeSessionResponse(session *Session) {
	if session == nil {
		return
	}
	session.Status = strings.ToUpper(session.Status)
	if session.CreateTS == 0 && len(session.CreatedAt) > 0 {
		_ = json.Unmarshal(session.CreatedAt, &session.CreateTS)
	}
}

// Stop stops a running session. The endpoint has no request body.
func (c *SessionsClient) Stop(ctx context.Context, id string) (*Response[map[string]any], error) {
	p, err := c.t.path("sessions", id, "stop")
	if err != nil {
		return nil, err
	}
	var out map[string]any
	r, err := c.t.call(ctx, http.MethodPost, p, nil, nil, &out, nil, 0)
	if err != nil {
		return nil, err
	}
	return response(r, out), nil
}

func validateSessionTransport(t SessionTransport) error {
	if t.Type == "telephony" {
		if strings.TrimSpace(t.From) == "" || len(t.From) > 64 {
			return errors.New("v3: telephony transport requires a from number ID of at most 64 characters")
		}
		if !validE164(t.To) {
			return errors.New("v3: telephony transport to must be an E.164 phone number")
		}
		if t.Channel != "" || t.UID != "" || len(t.SubscribeUIDs) > 0 || t.Token != "" || t.Encryption != nil || t.AudioScenario != "" || t.Geofence != nil {
			return errors.New("v3: telephony transport cannot include RTC fields")
		}
		return nil
	}
	if t.Type != "rtc" {
		return errors.New("v3: session transport type must be rtc or telephony")
	}
	if strings.TrimSpace(t.Channel) == "" {
		return errors.New("v3: session transport channel is required")
	}
	if t.UID != "" && !validRTCUID(t.UID, true) {
		return errors.New("v3: session transport uid must be a numeric UID string or a non-empty string UID")
	}
	if len(t.SubscribeUIDs) != 1 || !validRTCUID(t.SubscribeUIDs[0], false) {
		return errors.New("v3: session transport requires exactly one valid subscribe UID")
	}
	if t.UID != "" && t.UID != "0" && isNumericUID(t.UID) != isNumericUID(t.SubscribeUIDs[0]) {
		return errors.New("v3: agent and subscriber UIDs must use the same UID type")
	}
	if t.AudioScenario != "" && t.AudioScenario != "aiserver" && t.AudioScenario != "chorus" {
		return errors.New("v3: audio_scenario must be aiserver or chorus")
	}
	if t.Encryption != nil {
		modes := map[string]bool{"aes_128_xts": true, "aes_128_ecb": true, "aes_256_xts": true, "sm4_128_ecb": true, "aes_128_gcm": true, "aes_256_gcm": true, "aes_128_gcm2": true, "aes_256_gcm2": true}
		if !modes[t.Encryption.Mode] || strings.TrimSpace(t.Encryption.Key) == "" {
			return errors.New("v3: RTC encryption requires a supported mode and non-empty key")
		}
	}
	if t.Geofence != nil {
		areas := map[string]bool{"global": true, "north_america": true, "europe": true, "asia": true, "india": true, "japan": true}
		if !areas[t.Geofence.Area] || (t.Geofence.ExcludeArea != "" && (t.Geofence.Area != "global" || !areas[t.Geofence.ExcludeArea] || t.Geofence.ExcludeArea == "global")) {
			return errors.New("v3: invalid RTC geofence")
		}
	}
	return nil
}

func validRTCUID(value string, allowZero bool) bool {
	if strings.TrimSpace(value) == "" || value == "*" {
		return false
	}
	if isNumericUID(value) {
		if len(value) > 10 {
			return false
		}
		if !allowZero && value == "0" {
			return false
		}
		if len(value) > 1 && value[0] == '0' {
			return false
		}
		return true
	}
	return true
}

func validE164(value string) bool {
	if len(value) < 3 || len(value) > 16 || value[0] != '+' || value[1] < '1' || value[1] > '9' {
		return false
	}
	for _, r := range value[2:] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isNumericUID(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func validateSessionLifecycle(lifecycle *SessionLifecycle) error {
	if lifecycle == nil {
		return nil
	}
	if lifecycle.IdleTimeoutMS != nil && *lifecycle.IdleTimeoutMS < 0 {
		return errors.New("v3: idle_timeout_ms must be non-negative")
	}
	if lifecycle.MaxDurationMS != nil && (*lifecycle.MaxDurationMS < 1000 || *lifecycle.MaxDurationMS > 3600000) {
		return errors.New("v3: max_duration_ms must be between 1000 and 3600000")
	}
	if lifecycle.IdleTimeoutMS != nil && *lifecycle.IdleTimeoutMS > 3600000 {
		return errors.New("v3: idle_timeout_ms must not exceed 3600000")
	}
	if lifecycle.GracefulStop != nil && lifecycle.GracefulStop.TimeoutMS != nil && (*lifecycle.GracefulStop.TimeoutMS < 0 || *lifecycle.GracefulStop.TimeoutMS > 120000) {
		return errors.New("v3: graceful_stop.timeout_ms must be between 0 and 120000")
	}
	return nil
}

func validateDataPolicy(policy *DataPolicy) error {
	if policy != nil && policy.Retention != "" && policy.Retention != "30_days" && policy.Retention != "none" {
		return errors.New("v3: data_policy.retention must be 30_days or none")
	}
	return nil
}

var sessionValueNameRE = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

func validateSessionValues(values map[string]any) error {
	for name, value := range values {
		if !sessionValueNameRE.MatchString(name) {
			return fmt.Errorf("v3: invalid session value name %q", name)
		}
		switch value.(type) {
		case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, json.Number:
		default:
			return fmt.Errorf("v3: session value %q must be a string, number, or boolean", name)
		}
	}
	return nil
}

// validatePreviewPipeline also checks map-based merge patches before transport.
func validatePreviewPipeline(value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var pipeline map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &pipeline); err != nil {
		return err
	}
	if _, ok := pipeline["mllm"]; ok {
		return errors.New("v3: MLLM is not supported by the preview API")
	}
	if raw, ok := pipeline["mode"]; ok && string(raw) != "null" {
		var mode string
		if err := json.Unmarshal(raw, &mode); err != nil {
			return err
		}
		if mode != "cascaded" {
			return errors.New("v3: preview pipeline mode must be cascaded")
		}
	}
	return nil
}
