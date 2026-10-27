package v3

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
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
			if r.Method != http.MethodPost || r.URL.Path != "/conversational-ai/v1/projects/app/sessions" {
				t.Fatalf("start request: %s %s", r.Method, r.URL.Path)
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["agent_id"] != "agent1" || body["client_reference"] != nil || r.Header.Get("Idempotency-Key") != "session-key" {
				t.Fatalf("saved session body: %#v", body)
			}
			return reply(200, `{"agent_session_id":"session1","status":"starting","created_at":1781827200}`, http.Header{"Location": []string{"/sessions/session1"}}), nil
		case 2:
			if r.Method != http.MethodPost || r.URL.Path != "/conversational-ai/v1/projects/app/sessions/ephemeral" {
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
			if body["labels"].(map[string]any)["owner"] != "support" {
				t.Fatalf("ephemeral labels missing: %#v", body["labels"])
			}
			return reply(201, `{"agent_session_id":"session2","status":"RUNNING","created_at":1781827200}`, nil), nil
		case 3:
			if r.Method != http.MethodPost || r.URL.Path != "/conversational-ai/v1/projects/app/sessions/session1/stop" {
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
	started, err := c.Sessions.Start(ctx, SessionStartRequest{AgentID: "agent1", Labels: map[string]string{"owner": "x"}, Values: map[string]any{"customer": "Ada"}, Transport: transport}, SessionWriteOptions{IdempotencyKey: "session-key"})
	if err != nil || started.StatusCode != http.StatusOK || started.Body.AgentSessionID != "session1" || started.Body.Status != "STARTING" || started.Body.CreateTS != 1781827200 || len(started.Body.CreatedAt) == 0 || started.Header.Get("Location") != "/sessions/session1" {
		t.Fatalf("start response: %#v err=%v", started, err)
	}
	ephemeral, err := c.Sessions.StartEphemeral(ctx, SessionStartEphemeralRequest{
		Agent:     EphemeralAgent{Instructions: strPtr("say hello"), Pipeline: &Pipeline{Mode: "realtime"}},
		Labels:    map[string]string{"owner": "support"},
		Transport: transport,
	})
	if err != nil || ephemeral.Body.AgentSessionID != "session2" || ephemeral.Body.Status != "RUNNING" || ephemeral.Body.CreateTS != 1781827200 {
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
	if _, err := c.Sessions.Start(context.Background(), SessionStartRequest{AgentID: "agent1", Transport: transport, Lifecycle: &SessionLifecycle{MaxDurationMS: intPtrTest(0)}}); err == nil {
		t.Fatal("expected invalid lifecycle error")
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
		AgentID: "agent1", Transport: SessionTransport{Type: "rtc", Channel: "room", UID: "1", SubscribeUIDs: []string{"2"}},
		Lifecycle: &SessionLifecycle{IdleTimeoutMS: &zero},
	})
	if err != nil || calls != 1 {
		t.Fatalf("zero timeout rejected: err=%v calls=%d", err, calls)
	}
}

func TestSessionWriteRejectsWhitespaceOnlyIdempotencyKey(t *testing.T) {
	calls := 0
	c := NewClient("app", option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return reply(200, `{}`, nil), nil
	})))
	transport := SessionTransport{Type: "rtc", Channel: "room", UID: "1", SubscribeUIDs: []string{"2"}}
	if _, err := c.Sessions.Start(context.Background(), SessionStartRequest{AgentID: "agent", Transport: transport}, SessionWriteOptions{IdempotencyKey: " \t"}); err == nil {
		t.Fatal("whitespace-only session idempotency key should fail locally")
	}
	if _, err := c.Sessions.StartEphemeral(context.Background(), SessionStartEphemeralRequest{Agent: EphemeralAgent{}, Transport: transport}, SessionWriteOptions{IdempotencyKey: " \t"}); err == nil {
		t.Fatal("whitespace-only ephemeral idempotency key should fail locally")
	}
	if calls != 0 {
		t.Fatalf("invalid idempotency keys reached transport: calls=%d", calls)
	}
}

func TestSessionWriteIdempotencyKeyPreservesOpaqueNonblankValue(t *testing.T) {
	for _, key := range []string{"", "key:Opaque-!"} {
		headers, err := sessionWriteHeaders([]SessionWriteOptions{{IdempotencyKey: key}})
		if err != nil {
			t.Fatalf("sessionWriteHeaders(%q): %v", key, err)
		}
		if key == "" && len(headers) != 0 {
			t.Fatalf("empty default key should be omitted: %#v", headers)
		}
		if key != "" && headers.Get("Idempotency-Key") != key {
			t.Fatalf("opaque key was changed: got %q want %q", headers.Get("Idempotency-Key"), key)
		}
	}
}

func TestSessionUpdateUsesLimitedPatchWireShape(t *testing.T) {
	if encoded, err := json.Marshal(SessionModulePatch{}); err != nil || string(encoded) != `{}` {
		t.Fatalf("omitted params patch: %s %v", encoded, err)
	}
	if encoded, err := json.Marshal(SessionModulePatch{Params: map[string]any{}}); err != nil || string(encoded) != `{"params":{}}` {
		t.Fatalf("empty params replacement: %s %v", encoded, err)
	}
	calls := 0
	c := NewClient("app", option.WithBaseURL("https://stage.example"), option.WithHTTPClient(httpFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != http.MethodPatch || r.URL.Path != "/conversational-ai/v1/projects/app/sessions/0123456789abcdef0123456789abcdef" {
			t.Fatalf("request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if len(body) != 2 || body["transport"].(map[string]any)["token"] != "rtc-token" {
			t.Fatalf("patch body: %#v", body)
		}
		params := body["agent"].(map[string]any)["overrides"].(map[string]any)["pipeline"].(map[string]any)["mllm"].(map[string]any)["params"].(map[string]any)
		if len(params) != 1 || params["prompt"] != "updated" {
			t.Fatalf("params must replace: %#v", params)
		}
		return reply(200, `{"agent_session_id":"0123456789abcdef0123456789abcdef","create_ts":1,"status":"RUNNING"}`, http.Header{"ETag": []string{"s2"}}), nil
	})))
	_, err := c.Sessions.Update(context.Background(), "0123456789abcdef0123456789abcdef", SessionPatch{
		Transport: &SessionTransportPatch{Token: "rtc-token"},
		Agent:     &SessionAgentPatch{Overrides: &SessionAgentOverridesPatch{Pipeline: &SessionPipelinePatch{MLLM: &SessionModulePatch{Params: map[string]any{"prompt": "updated"}}}}},
	})
	if err != nil || calls != 1 {
		t.Fatalf("update: calls=%d err=%v", calls, err)
	}
}

func TestSessionUpdateMakesOneAttempt(t *testing.T) {
	calls := 0
	c := NewClient("app", option.WithBaseURL("https://stage.example"), option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return reply(503, `{"detail":"unavailable"}`, nil), nil
	})))
	_, err := c.Sessions.Update(context.Background(), "0123456789abcdef0123456789abcdef", SessionPatch{Transport: &SessionTransportPatch{Token: "token"}})
	if err == nil || calls != 1 {
		t.Fatalf("patch attempts=%d err=%v", calls, err)
	}
}

func TestSessionUpdateRejectsEndpointAliasParamsLocally(t *testing.T) {
	calls := 0
	c := NewClient("app", option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, io.ErrUnexpectedEOF
	})))
	_, err := c.Sessions.Update(context.Background(), "0123456789abcdef0123456789abcdef", SessionPatch{Agent: &SessionAgentPatch{Overrides: &SessionAgentOverridesPatch{Pipeline: &SessionPipelinePatch{LLM: &SessionModulePatch{Params: map[string]any{"api_base_url": "https://elsewhere"}}}}}})
	if err == nil || calls != 0 {
		t.Fatalf("endpoint alias was sent: calls=%d err=%v", calls, err)
	}
}

func TestRTCUIDMayBeOmittedForServerAssignment(t *testing.T) {
	transport := SessionTransport{Type: "rtc", Channel: "support", SubscribeUIDs: []string{"100"}}
	if err := validateSessionTransport(transport); err != nil {
		t.Fatalf("auto-assigned UID rejected: %v", err)
	}
	encoded, err := json.Marshal(transport)
	if err != nil || strings.Contains(string(encoded), `"uid"`) {
		t.Fatalf("UID should be omitted: %s %v", encoded, err)
	}
}

func TestSessionTransportRTCUIDAndTelephonyVariants(t *testing.T) {
	cases := []struct {
		name      string
		transport SessionTransport
		valid     bool
	}{
		{name: "zero selects string UID", transport: SessionTransport{Type: "rtc", Channel: "room", UID: "0", SubscribeUIDs: []string{"agent-user"}}, valid: true},
		{name: "wildcard subscriber", transport: SessionTransport{Type: "rtc", Channel: "room", UID: "agent-user", SubscribeUIDs: []string{"*"}}},
		{name: "blank string UID", transport: SessionTransport{Type: "rtc", Channel: "room", UID: " ", SubscribeUIDs: []string{"user"}}},
		{name: "telephony", transport: SessionTransport{Type: "telephony", From: "number-uuid", To: "+12025550123", CallPolicy: &CampaignCallPolicy{}}, valid: true},
		{name: "invalid phone", transport: SessionTransport{Type: "telephony", From: "number-uuid", To: "2025550123"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSessionTransport(tc.transport)
			if (err == nil) != tc.valid {
				t.Fatalf("validateSessionTransport(%#v) err=%v, valid=%t", tc.transport, err, tc.valid)
			}
		})
	}
}

func strPtr(s string) *string { return &s }
func intPtrTest(n int) *int   { return &n }
