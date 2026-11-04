package v3

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AgoraIO/agora-agents-go/v2/core"
	"github.com/AgoraIO/agora-agents-go/v2/option"
)

var allCategories = []error{ErrValidation, ErrConfiguration, ErrConnection, ErrAPI}

// assertCategory checks that err belongs to exactly one top-level category,
// is a v3.Error, and keeps the "v3: " / "agentkit: " message style.
func assertCategory(t *testing.T, err, want error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %v, got nil", want)
	}
	for _, category := range allCategories {
		if got := errors.Is(err, category); got != (category == want) {
			t.Fatalf("errors.Is(%v, %v) = %v; want category %v", err, category, got, want)
		}
	}
	var base Error
	if !errors.As(err, &base) {
		t.Fatalf("%T does not implement v3.Error", err)
	}
	if !strings.HasPrefix(err.Error(), "v3: ") {
		t.Fatalf("message lost v3 prefix: %q", err.Error())
	}
}

func countingClient(calls *int, opts ...option.RequestOption) *Client {
	transport := httpFunc(func(*http.Request) (*http.Response, error) {
		*calls++
		return reply(200, `{}`, nil), nil
	})
	return NewClient("app", append([]option.RequestOption{option.WithBaseURL("https://staging.example"), option.WithHTTPClient(transport)}, opts...)...)
}

func TestValidationErrorsAreReturnedBeforeHTTP(t *testing.T) {
	ctx := context.Background()
	calls := 0
	c := countingClient(&calls)
	cases := map[string]func() error{
		"empty agent ID": func() error { _, err := c.Agents.Get(ctx, ""); return err },
		"dot segment":    func() error { _, err := c.Numbers.Get(ctx, ".."); return err },
		"missing ETag":   func() error { _, err := c.Agents.Update(ctx, "a1", "", AgentPatch{}); return err },
		"mllm pipeline": func() error {
			_, err := c.Agents.Update(ctx, "a1", "e1", AgentPatch{"pipeline": map[string]any{"mllm": map[string]any{}}})
			return err
		},
		"bad pipeline type": func() error { _, err := c.Agents.Update(ctx, "a1", "e1", AgentPatch{"pipeline": "x"}); return err },
		"secret name":       func() error { _, err := c.Secrets.Get(ctx, "bad-name"); return err },
		"secret create key": func() error { _, err := c.Secrets.Create(ctx, SecretCreate{Name: "s"}, "k"); return err },
		"idempotency key": func() error {
			_, err := c.Sessions.Start(ctx, SessionStartRequest{}, SessionWriteOptions{IdempotencyKey: "has space"})
			return err
		},
		"session transport": func() error { _, err := c.Sessions.Start(ctx, SessionStartRequest{AgentID: "a1"}); return err },
		"campaign options": func() error {
			_, err := c.Campaigns.Create(ctx, CampaignCreate{}, CampaignWriteOptions{}, CampaignWriteOptions{})
			return err
		},
		"unencodable body": func() error {
			_, err := c.Agents.Update(ctx, "a1", "e1", AgentPatch{"agent_name": make(chan int)})
			return err
		},
		"session update ptrs": func() error {
			_, err := c.Sessions.Update(ctx, "s1", SessionPatch{Transport: &SessionTransportPatch{}})
			return err
		},
	}
	internal := NewClient("app", option.WithBaseURL("https://staging.example"), option.WithHTTPHeader(http.Header{"X-Studio-Route": []string{"x"}}), option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return reply(200, `{}`, nil), nil
	})))
	cases["internal header"] = func() error { _, err := internal.Agents.List(ctx, ListOptions{}); return err }
	for name, run := range cases {
		t.Run(name, func(t *testing.T) {
			err := run()
			assertCategory(t, err, ErrValidation)
			var validation *ValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("errors.As *ValidationError failed for %T", err)
			}
			if errors.Is(err, ErrTimeout) || errors.Is(err, ErrPreconditionFailed) {
				t.Fatalf("validation matched a subtype: %v", err)
			}
			if validation.Category() != ErrValidation {
				t.Fatalf("category = %v", validation.Category())
			}
		})
	}
	if calls != 0 {
		t.Fatalf("validation errors reached HTTP %d times", calls)
	}
}

func TestValidationErrorPreservesCause(t *testing.T) {
	calls := 0
	c := countingClient(&calls)
	_, err := c.Secrets.Create(context.Background(), SecretCreate{Name: "prod", Secrets: map[string]string{"bad-key": "v"}}, "")
	assertCategory(t, err, ErrValidation)
	want := `v3: invalid secret key "bad-key": v3: secret set/key names must be 1-64 ASCII letters, digits, or underscores`
	if err.Error() != want {
		t.Fatalf("message = %q", err.Error())
	}
	var outer *ValidationError
	if !errors.As(err, &outer) || outer.Err == nil || !errors.Is(outer.Err, ErrValidation) {
		t.Fatalf("cause chain lost: %#v", err)
	}
	_, err = c.Agents.Update(context.Background(), "a1", "e1", AgentPatch{"agent_name": make(chan int)})
	if errors.Unwrap(err) == nil {
		t.Fatalf("encoding cause lost: %#v", err)
	}
}

func TestConfigurationErrors(t *testing.T) {
	ctx := context.Background()
	calls := 0
	transport := option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return reply(200, `{}`, nil), nil
	}))
	cases := map[string]*Client{
		"base URL with path":   NewClient("app", option.WithBaseURL("https://staging.example/v1"), transport),
		"base URL credentials": NewClient("app", option.WithBaseURL("https://user:pass@staging.example"), transport),
		"missing app ID":       NewClient("", option.WithBaseURL("https://staging.example"), transport),
		"conflicting auth":     NewClient("app", option.WithBaseURL("https://staging.example"), option.WithHTTPHeader(http.Header{"Authorization": []string{"agora token=secret-one", "agora token=secret-two"}}), transport),
		"unsupported auth":     NewClient("app", option.WithBaseURL("https://staging.example"), option.WithAuthorization("Bearer secret-one"), transport),
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := c.Agents.List(ctx, ListOptions{})
			assertCategory(t, err, ErrConfiguration)
			var config *ConfigurationError
			if !errors.As(err, &config) || config.Category() != ErrConfiguration {
				t.Fatalf("errors.As *ConfigurationError failed: %#v", err)
			}
			for _, secret := range []string{"secret-one", "secret-two", "pass"} {
				if strings.Contains(err.Error(), secret) {
					t.Fatalf("message leaked credentials: %q", err.Error())
				}
			}
		})
	}
	if calls != 0 {
		t.Fatalf("configuration errors reached HTTP %d times", calls)
	}
}

func TestConnectionErrorKeepsTransportCause(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	origin := server.URL
	server.Close() // connection refused from now on
	c := NewClient("app", option.WithBaseURL(origin), option.WithToken("secret-token"), option.WithHTTPClient(&http.Client{}))
	_, err := c.Agents.Get(context.Background(), "a1")
	assertCategory(t, err, ErrConnection)
	var conn *ConnectionError
	if !errors.As(err, &conn) || conn.Timeout() || errors.Is(err, ErrTimeout) || conn.Category() != ErrConnection {
		t.Fatalf("connection error: %#v", err)
	}
	var opErr *net.OpError
	if !errors.As(err, &opErr) {
		t.Fatalf("net.OpError not reachable: %#v", err)
	}
	var netErr net.Error
	if !errors.As(err, &netErr) {
		t.Fatalf("net.Error not reachable: %#v", err)
	}
	if strings.Contains(err.Error(), "secret-token") {
		t.Fatalf("message leaked token: %q", err.Error())
	}
}

type failingBody struct{ err error }

func (b failingBody) Read([]byte) (int, error) { return 0, b.err }
func (failingBody) Close() error               { return nil }

func TestConnectionErrorFromCustomTransportAndBodyRead(t *testing.T) {
	cause := errors.New("connection reset by peer")
	c := NewClient("app", option.WithBaseURL("https://staging.example"), option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
		return nil, cause
	})))
	_, err := c.Sessions.Stop(context.Background(), "s1")
	assertCategory(t, err, ErrConnection)
	if !errors.Is(err, cause) {
		t.Fatalf("cause not reachable: %#v", err)
	}
	// Failing to READ a 2xx body is a transport failure, not an API error.
	readErr := errors.New("unexpected EOF while reading body")
	aborted := NewClient("app", option.WithBaseURL("https://staging.example"), option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: failingBody{readErr}}, nil
	})))
	_, err = aborted.Agents.Get(context.Background(), "a1")
	assertCategory(t, err, ErrConnection)
	if !errors.Is(err, readErr) {
		t.Fatalf("read cause not reachable: %#v", err)
	}
}

func TestContractViolationsAreAPIErrors(t *testing.T) {
	ctx := context.Background()
	requestID := http.Header{"Agora-Request-Id": []string{"req-200"}}
	t.Run("undecodable body", func(t *testing.T) {
		c := NewClient("app", option.WithBaseURL("https://staging.example"), option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
			return reply(200, `{"agent_id":`, requestID.Clone()), nil
		})))
		_, err := c.Agents.Get(ctx, "a1")
		assertCategory(t, err, ErrAPI)
		var api *APIError
		if !errors.As(err, &api) || api.StatusCode != 200 || api.RequestID != "req-200" || string(api.Body) != `{"agent_id":` || api.ParsedBody != nil {
			t.Fatalf("fields: %#v", api)
		}
		var legacy *core.APIError
		if errors.As(err, &legacy) {
			t.Fatal("a 2xx contract violation must not pretend to be a core.APIError")
		}
		var syntax *json.SyntaxError
		if !errors.As(err, &syntax) && !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Fatalf("decode cause not reachable: %#v", err)
		}
		if got := err.Error(); !strings.HasPrefix(got, "v3: API error: status 200 (request ID req-200): response body cannot be decoded") {
			t.Fatalf("message = %q", got)
		}
	})
	t.Run("empty body", func(t *testing.T) {
		c := NewClient("app", option.WithBaseURL("https://staging.example"), option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
			return reply(201, ``, nil), nil
		})))
		_, err := c.Sessions.Stop(ctx, "s1")
		assertCategory(t, err, ErrAPI)
		var api *APIError
		if !errors.As(err, &api) || api.StatusCode != 201 || len(api.Body) != 0 {
			t.Fatalf("fields: %#v", api)
		}
	})
	t.Run("latest without ETag", func(t *testing.T) {
		c := NewClient("app", option.WithBaseURL("https://staging.example"), option.WithHTTPClient(httpFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method == http.MethodGet {
				return reply(200, `{"agent_id":"a1","agent_name":"x","create_ts":1,"update_ts":2}`, requestID.Clone()), nil
			}
			return reply(412, `{"detail":"stale"}`, nil), nil
		})))
		_, err := c.Agents.Update(ctx, "a1", "old", AgentPatch{})
		var stale *PreconditionFailedError
		if !errors.As(err, &stale) {
			t.Fatalf("412: %#v", err)
		}
		_, _, err = stale.Latest(ctx)
		assertCategory(t, err, ErrAPI)
		var api *APIError
		if !errors.As(err, &api) || api.StatusCode != 200 || api.RequestID != "req-200" || !strings.Contains(string(api.Body), `"agent_id":"a1"`) || errors.Is(err, ErrPreconditionFailed) {
			t.Fatalf("fields: %#v", api)
		}
		if !strings.HasSuffix(err.Error(), ": response has no ETag") {
			t.Fatalf("message = %q", err.Error())
		}
	})
	t.Run("constructed without core error", func(t *testing.T) {
		var nilErr *APIError
		if nilErr.Unwrap() != nil {
			t.Fatal("nil Unwrap")
		}
		err := ContractViolation[Agent](nil, nil)
		if err.Unwrap() != nil || err.Error() != "v3: API error: status 0" || !errors.Is(err, ErrAPI) {
			t.Fatalf("empty contract violation: %q", err.Error())
		}
		resp := &Response[Agent]{StatusCode: 200, Header: http.Header{"Agora-Request-Id": []string{"r1"}}, raw: []byte(`{"detail":"ignored"}`)}
		err = ContractViolation(resp, errors.New("missing field"))
		if err.RequestID != "r1" || err.Detail != "ignored" || err.Error() != "v3: API error: status 200 (request ID r1): missing field" {
			t.Fatalf("contract violation: %#v %q", err, err.Error())
		}
	})
}

func TestTimeoutIsAConnectionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	}))
	defer server.Close()

	t.Run("context deadline", func(t *testing.T) {
		c := NewClient("app", option.WithBaseURL(server.URL), option.WithHTTPClient(&http.Client{}))
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		_, err := c.Agents.Get(ctx, "a1")
		assertCategory(t, err, ErrConnection)
		var conn *ConnectionError
		if !errors.Is(err, ErrTimeout) || !errors.As(err, &conn) || !conn.Timeout() || conn.Category() != ErrTimeout {
			t.Fatalf("timeout not classified: %#v", err)
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("context.DeadlineExceeded not reachable: %#v", err)
		}
		if !strings.HasPrefix(err.Error(), "v3: request timed out") {
			t.Fatalf("message = %q", err.Error())
		}
	})
	t.Run("client timeout", func(t *testing.T) {
		c := NewClient("app", option.WithBaseURL(server.URL), option.WithHTTPClient(&http.Client{Timeout: 20 * time.Millisecond}))
		_, err := c.Agents.Get(context.Background(), "a1")
		assertCategory(t, err, ErrConnection)
		var netErr net.Error
		if !errors.Is(err, ErrTimeout) || !errors.As(err, &netErr) || !netErr.Timeout() {
			t.Fatalf("client timeout not classified: %#v", err)
		}
	})
	t.Run("canceled is not a timeout", func(t *testing.T) {
		c := NewClient("app", option.WithBaseURL(server.URL), option.WithHTTPClient(&http.Client{}))
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := c.Agents.Get(ctx, "a1")
		assertCategory(t, err, ErrConnection)
		if errors.Is(err, ErrTimeout) || !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation: %#v", err)
		}
	})
}

func TestAPIErrorFields(t *testing.T) {
	c := NewClient("app", option.WithBaseURL("https://staging.example"), option.WithToken("secret-token"), option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
		h := http.Header{}
		h.Set("Agora-Request-Id", "req-123")
		h.Set("Content-Type", "application/json")
		return reply(404, `{"detail":"agent a1 not found","reason":"AgentNotFound"}`, h), nil
	})))
	_, err := c.Agents.Get(context.Background(), "a1")
	assertCategory(t, err, ErrAPI)
	var api *APIError
	if !errors.As(err, &api) {
		t.Fatalf("errors.As *APIError failed: %#v", err)
	}
	if api.StatusCode != 404 || api.Detail != "agent a1 not found" || api.Reason != "AgentNotFound" || api.RequestID != "req-123" || api.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("fields: %#v", api)
	}
	if string(api.Body) != `{"detail":"agent a1 not found","reason":"AgentNotFound"}` {
		t.Fatalf("body = %s", api.Body)
	}
	if parsed, ok := api.ParsedBody.(map[string]any); !ok || parsed["reason"] != "AgentNotFound" {
		t.Fatalf("parsed body = %#v", api.ParsedBody)
	}
	if got, want := err.Error(), "v3: API error: status 404 AgentNotFound (request ID req-123): agent a1 not found"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
	if errors.Is(err, ErrPreconditionFailed) || api.Category() != ErrAPI {
		t.Fatalf("404 matched precondition: %#v", err)
	}
	// Backward compatibility: the generated core error is still reachable.
	var legacy *core.APIError
	if !errors.As(err, &legacy) || legacy.StatusCode != 404 || !strings.Contains(legacy.Error(), "AgentNotFound") {
		t.Fatalf("core.APIError not reachable: %#v", err)
	}
	if strings.Contains(err.Error(), "secret-token") {
		t.Fatalf("message leaked token: %q", err.Error())
	}
}

func TestAPIErrorWithNonJSONBody(t *testing.T) {
	c := NewClient("app", option.WithBaseURL("https://staging.example"), option.WithMaxAttempts(1), option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
		return reply(502, "<html>bad gateway</html>", nil), nil
	})))
	_, err := c.Agents.List(context.Background(), ListOptions{})
	assertCategory(t, err, ErrAPI)
	var api *APIError
	if !errors.As(err, &api) || api.ParsedBody != nil || api.Detail != "" || api.Reason != "" || api.RequestID != "" || string(api.Body) != "<html>bad gateway</html>" {
		t.Fatalf("fields: %#v", api)
	}
	if err.Error() != "v3: API error: status 502: <html>bad gateway</html>" {
		t.Fatalf("message = %q", err.Error())
	}
}

func TestPreconditionFailedIsAnAPIError(t *testing.T) {
	ctx := context.Background()
	calls := 0
	c := NewClient("app", option.WithBaseURL("https://staging.example"), option.WithHTTPClient(httpFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		h := http.Header{}
		if r.Method == http.MethodGet {
			h.Set("ETag", "fresh")
			return reply(200, `{"number_id":"n1","number":"+12025550123","create_ts":1}`, h), nil
		}
		h.Set("Agora-Request-Id", "req-412")
		return reply(412, `{"detail":"stale etag","reason":"PreconditionFailed"}`, h), nil
	})))
	_, err := c.Numbers.Update(ctx, "n1", "old", NumberPatch{})
	assertCategory(t, err, ErrAPI)
	var stale *PreconditionFailedError
	var api *APIError
	var legacy *core.APIError
	if !errors.Is(err, ErrPreconditionFailed) || !errors.As(err, &stale) || !errors.As(err, &api) || !errors.As(err, &legacy) {
		t.Fatalf("412 chain: %#v", err)
	}
	if api.StatusCode != 412 || api.RequestID != "req-412" || api.Reason != "PreconditionFailed" || legacy.StatusCode != 412 || stale.Category() != ErrPreconditionFailed {
		t.Fatalf("412 fields: %#v %#v", api, legacy)
	}
	if got, want := err.Error(), "v3: precondition failed: status 412 PreconditionFailed (request ID req-412): stale etag"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
	latest, etag, err := stale.Latest(ctx)
	if err != nil || etag != "fresh" || calls != 2 {
		t.Fatalf("latest: %v %q calls=%d", err, etag, calls)
	}
	if _, ok := latest.(Number); !ok {
		t.Fatalf("latest body: %#v", latest)
	}

	// A 412 from a write without a reload still classifies as a precondition
	// failure; Latest reports that no reload is available.
	_, err = c.Sessions.Start(ctx, SessionStartRequest{AgentID: "a1", Transport: SessionTransport{Type: "rtc", Channel: "c", SubscribeUIDs: []string{"2"}}}, SessionWriteOptions{IdempotencyKey: "k1"})
	if !errors.As(err, &stale) || !errors.Is(err, ErrAPI) {
		t.Fatalf("session 412: %#v", err)
	}
	if _, _, err := stale.Latest(ctx); !errors.Is(err, ErrValidation) {
		t.Fatalf("latest without reload: %v", err)
	}
}

func TestCategoryErrorZeroValues(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want string
	}{
		{&ValidationError{}, "v3: invalid request"},
		{&ConfigurationError{}, "v3: invalid client configuration"},
		{&ConnectionError{}, "v3: request failed"},
		{&PreconditionFailedError{}, "v3: precondition failed"},
		{&APIError{StatusCode: 500}, "v3: API error: status 500"},
	} {
		if tc.err.Error() != tc.want {
			t.Errorf("%T message = %q, want %q", tc.err, tc.err.Error(), tc.want)
		}
	}
}
