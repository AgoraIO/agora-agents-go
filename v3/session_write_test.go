package v3

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/option"
)

func TestSessionStartEphemeralAndStopContracts(t *testing.T) {
	ctx := context.Background()
	calls := 0
	c := NewClient("app", option.WithBaseURL("https://staging.example"), option.WithHTTPClient(httpFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch calls {
		case 1:
			if r.Method != http.MethodPost || r.URL.Path != "/conversational-ai/v3/projects/app/sessions" {
				t.Fatalf("start request: %s %s", r.Method, r.URL.Path)
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["agent_id"] != "agent1" || body["client_reference"] != "caller-1" {
				t.Fatalf("saved session body: %#v", body)
			}
			return reply(201, `{"agent_session_id":"session1","status":"running","created_at":"2026-10-19T00:00:00Z"}`, http.Header{"Location": []string{"/sessions/session1"}}), nil
		case 2:
			if r.Method != http.MethodPost || r.URL.Path != "/conversational-ai/v3/projects/app/sessions/ephemeral" {
				t.Fatalf("ephemeral request: %s %s", r.Method, r.URL.Path)
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if _, has := body["agent_id"]; has {
				t.Fatal("ephemeral request must not include agent_id")
			}
			agent := body["agent"].(map[string]any)
			if _, has := agent["agent_name"]; has {
				t.Fatal("ephemeral agent must not include agent_name")
			}
			return reply(201, `{"agent_session_id":"session2","status":"RUNNING","created_at":1781827200}`, nil), nil
		case 3:
			if r.Method != http.MethodPost || r.URL.Path != "/conversational-ai/v3/projects/app/sessions/session1/stop" {
				t.Fatalf("stop request: %s %s", r.Method, r.URL.Path)
			}
			if r.Body != nil {
				body, _ := io.ReadAll(r.Body)
				if len(body) != 0 {
					t.Fatalf("stop request should have no body: %s", body)
				}
			}
			return reply(200, `{}`, nil), nil
		default:
			t.Fatalf("unexpected request %d", calls)
			return nil, nil
		}
	})))
	transport := SessionTransport{Type: "rtc", Channel: "room", UID: "1001", SubscribeUIDs: []string{"1002"}}
	started, err := c.Sessions.Start(ctx, SessionStartRequest{AgentID: "agent1", ClientReference: "caller-1", Transport: transport})
	if err != nil || started.StatusCode != http.StatusCreated || started.Body.AgentSessionID != "session1" || started.Body.Status != "running" || len(started.Body.CreatedAt) == 0 || started.Header.Get("Location") != "/sessions/session1" {
		t.Fatalf("start response: %#v err=%v", started, err)
	}
	ephemeral, err := c.Sessions.StartEphemeral(ctx, SessionStartEphemeralRequest{
		Agent:     EphemeralAgent{Instructions: strPtr("say hello"), Pipeline: &Pipeline{Mode: "realtime"}},
		Transport: transport,
	})
	if err != nil || ephemeral.Body.AgentSessionID != "session2" || ephemeral.Body.Status != "RUNNING" {
		t.Fatalf("ephemeral response: %#v err=%v", ephemeral, err)
	}
	stopped, err := c.Sessions.Stop(ctx, "session1")
	if err != nil || stopped.Body == nil || len(stopped.Body) != 0 || calls != 3 {
		t.Fatalf("stop response: %#v err=%v calls=%d", stopped, err, calls)
	}
}

func TestSessionWritesRejectUnsupportedInputsBeforeHTTP(t *testing.T) {
	calls := 0
	c := NewClient("app", option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return reply(201, `{}`, nil), nil
	})))
	transport := SessionTransport{Type: "rtc", Channel: "room", UID: "1001"}
	if _, err := c.Sessions.Start(context.Background(), SessionStartRequest{AgentID: "agent1", Transport: transport, Lifecycle: &SessionLifecycle{IdleTimeoutMS: intPtrTest(-1)}}); err == nil {
		t.Fatal("expected invalid lifecycle error")
	}
	if _, err := c.Sessions.StartEphemeral(context.Background(), SessionStartEphemeralRequest{Agent: EphemeralAgent{Pipeline: &Pipeline{Mode: "realtime"}}, Transport: transport, Labels: map[string]string{"x": "y"}}); err == nil {
		t.Fatal("expected ephemeral-label error")
	}
	if calls != 0 {
		t.Fatalf("invalid requests reached transport: %d", calls)
	}
}

func TestSessionWritesAllowZeroLifecycleTimeout(t *testing.T) {
	calls := 0
	c := NewClient("app", option.WithHTTPClient(httpFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		lifecycle := body["lifecycle"].(map[string]any)
		if lifecycle["idle_timeout_ms"] != float64(0) {
			t.Fatalf("zero timeout lost: %#v", lifecycle)
		}
		return reply(201, `{"agent_session_id":"session1","status":"running"}`, nil), nil
	})))
	zero := 0
	_, err := c.Sessions.Start(context.Background(), SessionStartRequest{
		AgentID: "agent1", Transport: SessionTransport{Type: "rtc", Channel: "room", UID: "1"},
		Lifecycle: &SessionLifecycle{IdleTimeoutMS: &zero},
	})
	if err != nil || calls != 1 {
		t.Fatalf("zero timeout rejected: err=%v calls=%d", err, calls)
	}
}

func strPtr(s string) *string { return &s }
func intPtrTest(n int) *int   { return &n }
