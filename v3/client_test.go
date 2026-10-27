package v3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/AgoraIO/agora-agents-go/v2/core"
	"github.com/AgoraIO/agora-agents-go/v2/option"
)

type httpFunc func(*http.Request) (*http.Response, error)

func (f httpFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }
func reply(code int, body string, h http.Header) *http.Response {
	return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader(body)), Header: h}
}

func TestCRUDHeadersPathsAndNoWriteRetry(t *testing.T) {
	ctx := context.Background()
	calls := 0
	c := NewClient("a/b", option.WithBaseURL("https://staging.example"), option.WithHTTPClient(httpFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != http.MethodPut || r.URL.EscapedPath() != "/conversational-ai/v1/projects/a%2Fb/secrets/prod/keys/openai" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.EscapedPath())
		}
		if r.Header.Get("If-Match") != "e1" || r.Header.Get("If-None-Match") != "*" {
			t.Fatalf("missing condition headers: %v", r.Header)
		}
		return reply(503, `{"detail":"unavailable"}`, nil), nil
	})))
	_, err := c.Secrets.SetKey(ctx, "prod", "openai", SecretKeyWrite{Value: "x"}, SecretKeyOptions{IfMatch: "e1", IfNoneMatch: "*"})
	if err == nil || calls != 1 {
		t.Fatalf("write retries: calls=%d err=%v", calls, err)
	}
	if _, err := c.Agents.Get(ctx, ""); err == nil {
		t.Fatal("empty ID should fail before HTTP")
	}
}

func TestDefaultOriginAuthAndInternalHeaderFiltering(t *testing.T) {
	if got := NewClient("app").shared.baseURL; got != "https://preview.ai.agora.io" {
		t.Fatalf("default base URL = %q", got)
	}
	calls := 0
	c := NewClient("app", option.WithToken("token"), option.WithHTTPClient(httpFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Path != "/conversational-ai/v1/projects/app/agents" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "agora token=token" {
			t.Fatalf("token auth missing: %v", r.Header)
		}
		return reply(200, `{"data":{"count":0,"list":[]},"has_more":false}`, nil), nil
	})))
	if _, err := c.Agents.List(context.Background(), ListOptions{}); err != nil || calls != 1 {
		t.Fatalf("token request failed: %v calls=%d", err, calls)
	}
	internalHeaders := NewClient("app", option.WithHTTPHeader(http.Header{"X-Studio-Region": []string{"internal"}, "X-Consumer-Username": []string{"internal"}}), option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) { calls++; return nil, nil })))
	if _, err := internalHeaders.Agents.List(context.Background(), ListOptions{}); err == nil || calls != 1 {
		t.Fatalf("internal headers should be rejected before HTTP: %v calls=%d", err, calls)
	}
	basic := NewClient("app", &core.BasicAuthOption{Username: "customer", Password: "secret"}, option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("basic request reached transport")
		return nil, nil
	})))
	if _, err := basic.Agents.List(context.Background(), ListOptions{}); err == nil {
		t.Fatal("basic auth should be rejected")
	}
	for _, headers := range []http.Header{
		{"Authorization": []string{"Basic Zm9vOmJhcg=="}},
		{"authorization": []string{"Basic Zm9vOmJhcg=="}},
		{"Authorization": []string{"agora token=t", "Basic Zm9vOmJhcg=="}},
	} {
		headerBasic := NewClient("app", option.WithHTTPHeader(headers), option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
			t.Fatal("basic header reached transport")
			return nil, nil
		})))
		if _, err := headerBasic.Agents.List(context.Background(), ListOptions{}); err == nil {
			t.Fatalf("basic authorization header should be rejected: %v", headers)
		}
	}
}

func TestBaseURLOverrideMustBeOrigin(t *testing.T) {
	for _, baseURL := range []string{
		"ftp://api.example.test",
		"https://user:pass@api.example.test",
		"https://api.example.test/prefix",
		"https://api.example.test?route=other",
		"https://api.example.test?",
		"https://api.example.test#fragment",
		"https://api.example.test#",
	} {
		t.Run(baseURL, func(t *testing.T) {
			calls := 0
			client := NewClient("app", option.WithBaseURL(baseURL), option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return reply(200, `{}`, nil), nil
			})))
			if _, err := client.Agents.List(context.Background(), ListOptions{}); err == nil || calls != 0 {
				t.Fatalf("invalid origin was not rejected before HTTP: base=%q err=%v calls=%d", baseURL, err, calls)
			}
		})
	}
}

func TestUnsupportedCreationIdempotencyHeaderRejectedBeforeHTTP(t *testing.T) {
	for _, resource := range []string{"agents", "secrets", "numbers"} {
		t.Run(resource, func(t *testing.T) {
			calls := 0
			client := NewClient("app", option.WithHTTPHeader(http.Header{"Idempotency-Key": []string{"unsupported"}}), option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return reply(200, `{}`, nil), nil
			})))
			var err error
			switch resource {
			case "agents":
				_, err = client.Agents.Create(context.Background(), AgentCreate{AgentName: "Agent"})
			case "secrets":
				_, err = client.Secrets.Create(context.Background(), SecretCreate{Name: "set"}, "")
			case "numbers":
				_, err = client.Numbers.Create(context.Background(), NumberCreate{}, "")
			}
			if err == nil || calls != 0 {
				t.Fatalf("unsupported idempotency header was not rejected locally: err=%v calls=%d", err, calls)
			}
		})
	}
}

func TestIdempotencyHeaderMustBeNonemptyOnSupportedOperations(t *testing.T) {
	for _, method := range []string{"session", "campaign"} {
		t.Run(method, func(t *testing.T) {
			calls := 0
			client := NewClient("app", option.WithHTTPHeader(http.Header{"Idempotency-Key": []string{" \t"}}), option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return reply(200, `{}`, nil), nil
			})))
			var err error
			switch method {
			case "session":
				_, err = client.Sessions.Start(context.Background(), SessionStartRequest{AgentID: "a", Transport: SessionTransport{Type: "rtc", Channel: "room", UID: "1", SubscribeUIDs: []string{"2"}}})
			case "campaign":
				_, err = client.Campaigns.Pause(context.Background(), "c")
			}
			if err == nil || calls != 0 {
				t.Fatalf("blank global idempotency header was not rejected locally: err=%v calls=%d", err, calls)
			}
		})
	}
}

func TestAgentPreconditionIncludesLatest(t *testing.T) {
	ctx := context.Background()
	calls := 0
	c := NewClient("app", option.WithBaseURL("https://staging.example"), option.WithHTTPHeader(http.Header{"If-Match": []string{"global"}, "If-None-Match": []string{"*"}, "Idempotency-Key": []string{"write-key"}}), option.WithHTTPClient(httpFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method == http.MethodPatch {
			if r.Header.Get("If-Match") != "old" {
				t.Fatal("If-Match missing")
			}
			return reply(412, `{"detail":"stale"}`, nil), nil
		}
		for _, name := range []string{"If-Match", "If-None-Match", "Idempotency-Key"} {
			if r.Header.Get(name) != "" {
				t.Fatalf("latest GET retained conditional header %s: %v", name, r.Header)
			}
		}
		h := http.Header{}
		h.Set("ETag", "new")
		return reply(200, `{"agent_id":"a1","agent_name":"Latest","create_ts":1,"update_ts":2}`, h), nil
	})))
	_, err := c.Agents.Update(ctx, "a1", "old", AgentPatch{"agent_name": "next"})
	var stale *PreconditionFailedError
	if !errors.As(err, &stale) || calls != 1 {
		t.Fatalf("412 latest: %#v calls=%d", err, calls)
	}
	latest, etag, err := stale.Latest(ctx)
	if err != nil || etag != "new" || calls != 2 {
		t.Fatalf("lazy latest: %v %q calls=%d", err, etag, calls)
	}
	if got, ok := latest.(Agent); !ok || got.AgentName != "Latest" {
		t.Fatalf("latest body: %#v", latest)
	}
}

func TestCursorAndRawETag(t *testing.T) {
	ctx := context.Background()
	c := NewClient("app", option.WithBaseURL("https://staging.example"), option.WithHTTPClient(httpFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("cursor") != "next" || r.URL.Query().Get("limit") != "10" {
			t.Fatalf("query: %s", r.URL.RawQuery)
		}
		h := http.Header{}
		h.Set("ETag", "e2")
		h.Set("Agora-Request-Id", "req1")
		return reply(200, `{"data":{"count":1,"list":[{"agent_id":"a1","agent_name":"A"}]},"has_more":false,"next_cursor":null}`, h), nil
	})))
	r, err := c.Agents.List(ctx, ListOptions{Cursor: "next", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if r.ETag != "e2" || r.RequestID != "req1" || len(r.Body.Data.List) != 1 {
		t.Fatalf("bad page: %#v", r)
	}
}

func TestTypedCreatePreservesRequiredNullAndEmptyFields(t *testing.T) {
	ctx := context.Background()
	calls := 0
	c := NewClient("app", option.WithBaseURL("https://staging.example"), option.WithHTTPClient(httpFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		switch calls {
		case 1:
			var inbound map[string]json.RawMessage
			if err := json.Unmarshal(body["inbound"], &inbound); err != nil {
				t.Fatal(err)
			}
			if got, ok := inbound["agent"]; !ok || string(got) != "null" {
				t.Fatalf("required nullable agent was omitted: %s", body["inbound"])
			}
			return reply(201, `{"number_id":"n1","number":"+12025550123","create_ts":1}`, nil), nil
		case 2:
			var greeting map[string]json.RawMessage
			if err := json.Unmarshal(body["greeting"], &greeting); err != nil {
				t.Fatal(err)
			}
			if got, ok := greeting["text"]; !ok || string(got) != `""` {
				t.Fatalf("required empty text was omitted: %s", body["greeting"])
			}
			return reply(201, `{"agent_id":"a1","agent_name":"Greeting","create_ts":1,"update_ts":1}`, nil), nil
		default:
			t.Fatal("unexpected request")
			return nil, nil
		}
	})))
	if _, err := c.Numbers.Create(ctx, NumberCreate{Number: "+12025550123", SipTrunk: SipTrunk{Hostname: "sip.example"}, Inbound: &Inbound{CallPolicy: map[string]any{"max_call_duration_seconds": 900}}}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Agents.Create(ctx, AgentCreate{AgentName: "Greeting", Greeting: Greeting{Mode: "text", Text: ""}}); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d", calls)
	}
}

func TestResponsePreservesDetachedRawHeaders(t *testing.T) {
	header := http.Header{}
	header.Set("ETag", "e1")
	header.Set("Agora-Request-Id", "request-1")
	header.Set("Location", "/secrets/test")
	header.Set("Idempotent-Replayed", "true")
	header["X-Multi"] = []string{"first", "second"}
	c := NewClient("app", option.WithBaseURL("https://staging.example"), option.WithHTTPClient(httpFunc(func(r *http.Request) (*http.Response, error) {
		return reply(201, `{"name":"test","keys":["key"],"create_ts":1,"update_ts":1}`, header), nil
	})))
	if _, err := c.Secrets.Create(context.Background(), SecretCreate{Name: "test"}, "once"); err == nil {
		t.Fatal("secret idempotency key should be rejected")
	}
	got, err := c.Secrets.Create(context.Background(), SecretCreate{Name: "test", Secrets: map[string]string{"key": "value"}}, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Header.Get("Location") != "/secrets/test" || got.Header.Get("Idempotent-Replayed") != "true" || got.ETag != "e1" || got.RequestID != "request-1" || got.StatusCode != 201 {
		t.Fatalf("response headers lost: %#v", got)
	}
	header["X-Multi"][0] = "changed by transport"
	header.Set("Location", "changed")
	if got.Header["X-Multi"][0] != "first" || got.Header.Get("Location") != "/secrets/test" {
		t.Fatal("response headers alias transport headers")
	}
	got.Header["X-Multi"][1] = "changed by caller"
	if header["X-Multi"][1] != "second" {
		t.Fatal("transport headers alias caller response")
	}
}

func TestNumberCreateRejectsIdempotencyKey(t *testing.T) {
	calls := 0
	c := NewClient("app", option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) { calls++; return reply(201, `{}`, nil), nil })))
	if _, err := c.Numbers.Create(context.Background(), NumberCreate{Number: "+12025550123"}, "retry"); err == nil || calls != 0 {
		t.Fatalf("unsupported idempotency key was not rejected locally: err=%v calls=%d", err, calls)
	}
}

func TestDotSegmentsNeverReachHTTP(t *testing.T) {
	for _, id := range []string{".", ".."} {
		t.Run(id, func(t *testing.T) {
			calls := 0
			transport := option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) { calls++; return reply(200, `{}`, nil), nil }))
			c := NewClient("app", transport)
			if _, err := c.Agents.Get(context.Background(), id); err == nil {
				t.Error("dot agent ID accepted")
			}
			if _, err := c.Campaigns.Cancel(context.Background(), id); err == nil {
				t.Error("dot campaign ID accepted")
			}
			if _, err := NewClient(id, transport).Agents.List(context.Background(), ListOptions{}); err == nil {
				t.Error("dot project ID accepted")
			}
			if calls != 0 {
				t.Fatalf("invalid path reached HTTP %d times", calls)
			}
		})
	}
}
func TestWriteDoesNotWaitRetryAfterWhenRetriesDisabled(t *testing.T) {
	for _, status := range []int{429, 503} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			calls := 0
			c := NewClient("app", option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return reply(status, `{"reason":"retry-later"}`, http.Header{"Retry-After": []string{"1"}, "Agora-Request-Id": []string{"no-retry-request"}}), nil
			})))
			began := time.Now()
			_, err := c.Campaigns.Create(context.Background(), CampaignCreate{})
			elapsed := time.Since(began)
			var apiError *core.APIError
			if !errors.As(err, &apiError) || apiError.StatusCode != status || !strings.Contains(apiError.Error(), "retry-later") {
				t.Fatalf("backend error lost: %v", err)
			}
			if apiError.Header.Get("Agora-Request-Id") != "no-retry-request" {
				t.Fatalf("request ID lost: %#v", apiError.Header)
			}
			if calls != 1 {
				t.Fatalf("write was repeated %d times", calls)
			}
			if elapsed >= 500*time.Millisecond {
				t.Fatalf("no-retry write unnecessarily waited %v", elapsed)
			}
		})
	}
}
