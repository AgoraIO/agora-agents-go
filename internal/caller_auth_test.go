package internal

import (
	"context"
	"net/http"
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/core"
)

type authTestHTTPClient func(*http.Request) (*http.Response, error)

func (f authTestHTTPClient) Do(req *http.Request) (*http.Response, error) { return f(req) }

func TestCallerRejectsUnsupportedHeadersBeforeHTTP(t *testing.T) {
	tests := []struct {
		name    string
		headers http.Header
	}{
		{
			name:    "incomplete basic auth option",
			headers: core.NewRequestOptions(&core.BasicAuthOption{Username: "customer", Password: ""}).ToHeader(),
		},
		{name: "bearer authorization", headers: http.Header{"Authorization": []string{"Bearer abc"}}},
		{name: "other scheme after token", headers: http.Header{"Authorization": []string{"agora token=valid", "Digest abc"}}},
		{name: "case-variant other scheme", headers: http.Header{"authorization": []string{"Bearer abc"}}},
		{name: "basic authorization", headers: http.Header{"Authorization": []string{"  bAsIc abc  "}}},
		{name: "lowercase basic authorization", headers: http.Header{"authorization": []string{"Basic abc"}}},
		{name: "mixed-case second authorization value", headers: http.Header{"Authorization": []string{"Bearer token", "Basic abc"}}},
		{name: "studio routing header", headers: http.Header{"X-Studio-Region": []string{"internal"}}},
		{name: "consumer routing header", headers: http.Header{"X-Consumer-Username": []string{"internal"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			caller := NewCaller(&CallerParams{Client: authTestHTTPClient(func(*http.Request) (*http.Response, error) {
				calls++
				return nil, nil
			})})
			_, err := caller.Call(context.Background(), &CallParams{URL: "https://example.test", Method: http.MethodGet, Headers: tt.headers, ResponseIsOptional: true})
			if err == nil || calls != 0 {
				t.Fatalf("unsupported auth/routing header was not rejected locally: err=%v calls=%d", err, calls)
			}
		})
	}
}

func TestCallerSupportsBasicAuthorization(t *testing.T) {
	calls := 0
	caller := NewCaller(&CallerParams{Client: authTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		calls++
		user, password, ok := req.BasicAuth()
		if !ok || user != "customer" || password != "secret" {
			t.Fatal("incorrect Basic auth")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: http.NoBody}, nil
	})})
	_, err := caller.Call(context.Background(), &CallParams{URL: "https://example.test", Method: http.MethodGet, Headers: core.NewRequestOptions(&core.BasicAuthOption{Username: "customer", Password: "secret"}).ToHeader(), ResponseIsOptional: true})
	if err != nil || calls != 1 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
}
