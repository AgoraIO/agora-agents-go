package cn

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/AgoraIO-Community/go-tokenbuilder/accesstoken2"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit"
	"github.com/AgoraIO/agora-agents-go/v2/option"
	"github.com/AgoraIO/agora-agents-go/v2/v3"
)

type clientAuthTransport func(*http.Request) (*http.Response, error)

func (f clientAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCNClientAuthenticationParity(t *testing.T) {
	appID, certificate := strings.Repeat("a", 32), strings.Repeat("b", 32)
	for _, tc := range []struct {
		name string
		opts ClientOptions
		mode agentkit.AuthMode
	}{
		{"basic", ClientOptions{AppID: appID, CustomerID: "customer", CustomerSecret: "secret", AppCertificate: certificate}, agentkit.AuthModeBasic},
		{"token", ClientOptions{AppID: appID, Token: "provided-token", AppCertificate: certificate}, agentkit.AuthModeToken},
		{"app credentials", ClientOptions{AppID: appID, AppCertificate: certificate}, agentkit.AuthModeAppCredentials},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// The CN facade uses the standard HTTP client; replace its transport without a network listener.
			original := http.DefaultClient
			t.Cleanup(func() { http.DefaultClient = original })
			calls := 0
			http.DefaultClient = &http.Client{Transport: clientAuthTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				switch tc.mode {
				case agentkit.AuthModeBasic:
					user, password, ok := r.BasicAuth()
					if !ok || user != "customer" || password != "secret" {
						t.Fatal("Basic credentials not forwarded")
					}
				case agentkit.AuthModeToken:
					if r.Header.Get("Authorization") != "agora token=provided-token" {
						t.Fatal("supplied token not forwarded")
					}
				default:
					token := accesstoken2.CreateAccessToken()
					ok, err := token.Parse(strings.TrimPrefix(r.Header.Get("Authorization"), "agora token="))
					if err != nil || !ok {
						t.Fatal("invalid generated token", err)
					}
					if valid, err := token.VerifySignature(certificate); err != nil || !valid {
						t.Fatal("invalid signature", err)
					}
					if token.Services[accesstoken2.ServiceTypeConvoAI] == nil {
						t.Fatal("missing ConvoAI service")
					}
				}
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"data":{"list":[],"count":0},"has_more":false}`))}, nil
			})}
			client := NewAgoraClient(tc.opts)
			if client.AuthMode != tc.mode || client.AppCertificate() != certificate || client.Area != option.AreaCN {
				t.Fatal("client options not forwarded")
			}
			if _, err := client.V3.Sessions.List(context.Background(), v3.SessionListOptions{}); err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Fatalf("requests=%d", calls)
			}
		})
	}
}
