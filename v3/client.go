package v3

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/AgoraIO/agora-agents-go/v2/core"
	"github.com/AgoraIO/agora-agents-go/v2/internal"
	"github.com/AgoraIO/agora-agents-go/v2/option"
)

const defaultHost = "https://api.agora.io"

type Client struct {
	Agents   *AgentsClient
	Secrets  *SecretsClient
	Numbers  *NumbersClient
	Sessions *SessionsClient
	shared   *transport
}
type transport struct {
	appID   string
	baseURL string
	options *core.RequestOptions
	caller  *internal.Caller
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
	if env := os.Getenv("AGORA_AGENTS_V3_BASE_URL"); env != "" && (base == "" || strings.HasSuffix(strings.TrimRight(base, "/"), "/api/conversational-ai-agent")) {
		base = env
	}
	if base == "" {
		base = defaultHost
	}
	base = strings.TrimSuffix(strings.TrimRight(base, "/"), "/api/conversational-ai-agent")
	t := &transport{appID: appID, baseURL: base, options: o, caller: internal.NewCaller(&internal.CallerParams{Client: o.HTTPClient, MaxAttempts: o.MaxAttempts})}
	return &Client{Agents: &AgentsClient{t}, Secrets: &SecretsClient{t}, Numbers: &NumbersClient{t}, Sessions: &SessionsClient{t}, shared: t}
}

// ForAppID binds a project to a low-level client created without one.
func (c *Client) ForAppID(appID string) *Client {
	t := *c.shared
	t.appID = appID
	return &Client{Agents: &AgentsClient{&t}, Secrets: &SecretsClient{&t}, Numbers: &NumbersClient{&t}, Sessions: &SessionsClient{&t}, shared: &t}
}

func segment(s, label string) (string, error) {
	if strings.TrimSpace(s) == "" {
		return "", fmt.Errorf("v3: %s is required", label)
	}
	return url.PathEscape(s), nil
}
func (t *transport) path(resource string, ids ...string) (string, error) {
	a, err := segment(t.appID, "appID")
	if err != nil {
		return "", err
	}
	u := t.baseURL + "/conversational-ai/v3/projects/" + a + "/" + resource
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
	if path == "" {
		return Response[any]{}, errors.New("v3: empty endpoint")
	}
	base := t.options.ToHeader()
	if t.options.Username != "" && t.options.Password != "" {
		base.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(t.options.Username+":"+t.options.Password)))
	}
	base = internal.MergeHeaders(base, headers)
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
	return Response[any]{Body: output, ETag: raw.Header.Get("ETag"), RequestID: raw.Header.Get("Agora-Request-Id"), StatusCode: raw.StatusCode}, nil
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
	return &Response[T]{Body: body, ETag: r.ETag, RequestID: r.RequestID, StatusCode: r.StatusCode}
}

type AgentsClient struct{ t *transport }

func (c *AgentsClient) Create(ctx context.Context, body AgentCreate) (*Response[Agent], error) {
	p, e := c.t.path("agents")
	if e != nil {
		return nil, e
	}
	var out Agent
	// The staging API does not honor Idempotency-Key for agents.
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
			r, err := c.Get(ctx, id)
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

func (c *SecretsClient) Create(ctx context.Context, body SecretCreate, idempotencyKey string) (*Response[SecretSet], error) {
	p, e := c.t.path("secrets")
	if e != nil {
		return nil, e
	}
	var out SecretSet
	h := http.Header{}
	if idempotencyKey != "" {
		h.Set("Idempotency-Key", idempotencyKey)
	}
	r, e := c.t.call(ctx, http.MethodPost, p, nil, body, &out, h, 1)
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
	p, e := c.t.path("secrets", name)
	if e != nil {
		return nil, e
	}
	var out SecretSet
	r, e := c.t.call(ctx, http.MethodPatch, p, nil, body, &out, etagHeader(ifMatch), 0)
	if e != nil {
		return nil, precondition(ctx, e, func(ctx context.Context) (any, string, error) {
			r, err := c.Get(ctx, name)
			if err != nil {
				return nil, "", err
			}
			return r.Body, r.ETag, nil
		})
	}
	return response(r, out), nil
}
func (c *SecretsClient) Replace(ctx context.Context, name, ifMatch string, body SecretReplace) (*Response[SecretSet], error) {
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
			r, err := c.Get(ctx, name)
			if err != nil {
				return nil, "", err
			}
			return r.Body, r.ETag, nil
		})
	}
	return response(r, out), nil
}
func (c *SecretsClient) Delete(ctx context.Context, name, ifMatch string) (*Response[Deleted], error) {
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
			r, err := c.Get(ctx, name)
			if err != nil {
				return nil, "", err
			}
			return r.Body, r.ETag, nil
		})
	}
	return response(r, out), nil
}
func (c *SecretsClient) SetKey(ctx context.Context, name, key string, body SecretKeyWrite, opts SecretKeyOptions) (*Response[SecretKey], error) {
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
			r, err := c.Get(ctx, name)
			if err != nil {
				return nil, "", err
			}
			return r.Body, r.ETag, nil
		})
	}
	return response(r, out), nil
}
func (c *SecretsClient) DeleteKey(ctx context.Context, name, key, ifMatch string) (*Response[Deleted], error) {
	p, e := c.t.path("secrets", name, "keys", key)
	if e != nil {
		return nil, e
	}
	var out Deleted
	r, e := c.t.call(ctx, http.MethodDelete, p, nil, nil, &out, etagHeader(ifMatch), 0)
	if e != nil {
		return nil, precondition(ctx, e, func(ctx context.Context) (any, string, error) {
			r, err := c.Get(ctx, name)
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
	p, e := c.t.path("numbers")
	if e != nil {
		return nil, e
	}
	var out Number
	h := http.Header{}
	if idempotencyKey != "" {
		h.Set("Idempotency-Key", idempotencyKey)
	}
	// Number create is not reliably idempotent on staging even with this header.
	r, e := c.t.call(ctx, http.MethodPost, p, nil, body, &out, h, 1)
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
			r, err := c.Get(ctx, id)
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
	return response(r, out), nil
}
