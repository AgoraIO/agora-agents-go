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
			name:    "legacy basic auth option",
			headers: core.NewRequestOptions(&core.BasicAuthOption{Username: "customer", Password: "secret"}).ToHeader(),
		},
		{name: "basic authorization", headers: http.Header{"Authorization": []string{"  bAsIc abc  "}}},
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
