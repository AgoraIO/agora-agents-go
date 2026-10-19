package v3

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

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
		if r.Method != http.MethodPut || r.URL.EscapedPath() != "/conversational-ai/v3/projects/a%2Fb/secrets/prod/keys/openai" {
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

func TestAgentPreconditionIncludesLatest(t *testing.T) {
	ctx := context.Background()
	calls := 0
	c := NewClient("app", option.WithBaseURL("https://staging.example"), option.WithHTTPClient(httpFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method == http.MethodPatch {
			if r.Header.Get("If-Match") != "old" {
				t.Fatal("If-Match missing")
			}
			return reply(412, `{"detail":"stale"}`, nil), nil
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
