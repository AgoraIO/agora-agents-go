package agentkit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/AgoraIO/agora-agents-go/v2/option"
	"github.com/AgoraIO/agora-agents-go/v2/v3"
)

func TestBasicAuthSessionPathsAndIndependentRTCToken(t *testing.T) {
	for _, certificate := range []string{"", strings.Repeat("b", 32)} {
		t.Run(fmt.Sprintf("certificate=%t", certificate != ""), func(t *testing.T) {
			calls := 0
			client := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: strings.Repeat("a", 32), AppCertificate: certificate, CustomerID: "customer", CustomerSecret: "secret", HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				user, password, ok := r.BasicAuth()
				if !ok || user != "customer" || password != "secret" {
					t.Fatal("REST Basic auth changed")
				}
				if r.Method == http.MethodGet {
					return v3Reply(200, `{"data":{"list":[],"count":0},"has_more":false}`, ""), nil
				}
				if strings.Contains(r.URL.Path, "/sessions") && !strings.HasSuffix(r.URL.Path, "/stop") {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatal(err)
					}
					transport := body["transport"].(map[string]any)
					if certificate != "" && (transport["token"] == nil || transport["token"] == "") {
						t.Fatal("RTC token was not generated in Basic mode")
					}
					return v3Reply(200, `{"agent_session_id":"session","status":"RUNNING"}`, ""), nil
				}
				return v3Reply(200, `{"agent_id":"legacy"}`, ""), nil
			})})
			if client.AuthMode != AuthModeBasic || client.IsAppCredentialsMode() {
				t.Fatal("wrong auth mode")
			}
			ctx := context.Background()
			if _, err := client.V3.Sessions.List(ctx, v3.SessionListOptions{}); err != nil {
				t.Fatal(err)
			}
			saved := (&SavedAgent{ID: "saved", client: client}).CreateSession(CreateSessionOptions{Channel: "room", AgentUID: "1", RemoteUIDs: []string{"2"}})
			if _, err := saved.Start(ctx); err != nil {
				t.Fatal(err)
			}
			if err := saved.Stop(ctx); err != nil {
				t.Fatal(err)
			}
			config := NewAgent(client).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-4o-mini"})).WithTts(vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{Model: "speech-2.8-turbo"}))
			if _, err := config.CreateEphemeralSession(CreateSessionOptions{Channel: "room", AgentUID: "1", RemoteUIDs: []string{"2"}}).Start(ctx); err != nil {
				t.Fatal(err)
			}
			if _, err := config.CreateSession(CreateSessionOptions{Channel: "room", Token: "caller-rtc", AgentUID: "1", RemoteUIDs: []string{"2"}}).Start(ctx); err != nil {
				t.Fatal(err)
			}
			if calls != 5 {
				t.Fatalf("calls=%d", calls)
			}
		})
	}
}

func TestBasicAuthRejectsInvalidOptions(t *testing.T) {
	for _, opts := range []AgoraClientOptions{
		{CustomerID: "id"}, {CustomerSecret: "secret"}, {CustomerID: " ", CustomerSecret: "secret"}, {CustomerID: "id", CustomerSecret: " "}, {CustomerID: "bad:id", CustomerSecret: "secret"}, {CustomerID: "id", CustomerSecret: "secret", Token: "token"},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("expected invalid authentication options to panic")
				}
			}()
			NewAgoraClient(opts)
		}()
	}
}
