package agentkit

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	agentcore "github.com/AgoraIO/agora-agents-go/v2/agentkit/core"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/AgoraIO/agora-agents-go/v2/core"
	"github.com/AgoraIO/agora-agents-go/v2/option"
	"github.com/AgoraIO/agora-agents-go/v2/v3"
)

func TestAgentKitV3SentinelsFoldIntoValidation(t *testing.T) {
	for _, tc := range []struct {
		err     error
		message string
	}{
		{ErrUnsupportedField, "agentkit: fields cannot be mapped to v3"},
		{ErrNumberNotFound, "agentkit: number not found in this project"},
		{ErrUnsupportedSessionOperation, "agentkit: operation is not supported for v3 sessions"},
		{ErrSecretRefNotSupported, "secret references are only supported by v3 deploy"},
	} {
		if tc.err.Error() != tc.message {
			t.Errorf("message changed: %q", tc.err.Error())
		}
		var validation *v3.ValidationError
		var base v3.Error
		if !errors.Is(tc.err, v3.ErrValidation) || !errors.As(tc.err, &validation) || !errors.As(tc.err, &base) {
			t.Errorf("%q is not a v3 validation error", tc.message)
		}
		if errors.Is(tc.err, v3.ErrConfiguration) || errors.Is(tc.err, v3.ErrAPI) || errors.Is(tc.err, v3.ErrConnection) {
			t.Errorf("%q matched another category", tc.message)
		}
	}
	if ErrSecretRefNotSupported != agentcore.ErrSecretRefNotSupported {
		t.Fatal("ErrSecretRefNotSupported alias diverged")
	}
	// Wrapped sentinels still match both the AgentKit and the v3 category.
	wrapped := errors.Join(errors.New("context"), ErrUnsupportedSessionOperation)
	if !errors.Is(wrapped, ErrUnsupportedSessionOperation) || !errors.Is(wrapped, v3.ErrValidation) {
		t.Fatal("wrapped sentinel lost its identity")
	}
}

func TestAgentKitV3PathsReturnCategorizedErrors(t *testing.T) {
	ctx := context.Background()
	calls := 0
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", V3: &V3Options{BaseURL: "https://stage.example"}, HTTPClient: v3HTTPFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("unexpected request")
	})})

	_, err := NewAgent(c).WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{APIKey: "key"})).Save(ctx, SaveOptions{Name: "rt"})
	if !errors.Is(err, ErrUnsupportedField) || !errors.Is(err, v3.ErrValidation) || !strings.Contains(err.Error(), "MLLM") {
		t.Fatalf("unsupported field: %v", err)
	}

	saved := &SavedAgent{ID: "saved1", Name: "Support", client: c}
	_, err = saved.CreateSession(CreateSessionOptions{Channel: "c", AgentUID: "1", ExpiresIn: 60}).Start(ctx)
	if !errors.Is(err, v3.ErrValidation) || err.Error() != "agentkit: v3 sessions do not support ExpiresIn" {
		t.Fatalf("session option: %v", err)
	}

	if _, err = saved.AssignNumber(ctx, 42, AssignNumberOptions{}); !errors.Is(err, ErrNumberNotFound) || !errors.Is(err, v3.ErrValidation) {
		t.Fatalf("invalid number: %v", err)
	}

	if _, err = GetSavedAgent(ctx, nil, "id"); !errors.Is(err, v3.ErrConfiguration) {
		t.Fatalf("missing client: %v", err)
	}

	if calls != 0 {
		t.Fatalf("local errors reached HTTP %d times", calls)
	}

	// Transport failures from the low-level client surface unchanged.
	_, err = GetSavedAgent(ctx, c, "id")
	var conn *v3.ConnectionError
	if !errors.Is(err, v3.ErrConnection) || !errors.As(err, &conn) {
		t.Fatalf("connection: %v", err)
	}
}

func TestAgentKitV3SessionUpdateValidation(t *testing.T) {
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", V3: &V3Options{BaseURL: "https://stage.example"}})
	session := (&SavedAgent{ID: "saved1", client: c}).CreateSession(CreateSessionOptions{Channel: "c"})
	if err := session.Update(context.Background(), nil); !errors.Is(err, v3.ErrValidation) {
		t.Fatalf("nil properties: %v", err)
	}
	if err := session.Update(context.Background(), &Agora.UpdateAgentsRequestProperties{}); !errors.Is(err, v3.ErrValidation) {
		t.Fatalf("idle session: %v", err)
	}
}

func TestAgentKitPreconditionMatchesV3Taxonomy(t *testing.T) {
	calls := 0
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", V3: &V3Options{BaseURL: "https://stage.example"}, HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method == http.MethodGet {
			return v3Reply(200, `{"agent_id":"id","agent_name":"Old","create_ts":1,"update_ts":2}`, "e1"), nil
		}
		return v3Reply(412, `{"detail":"stale","reason":"PreconditionFailed"}`, ""), nil
	})})
	_, err := NewAgent(c).Save(context.Background(), SaveOptions{Name: "New", AgentID: "id"})
	var stale *v3.PreconditionFailedError
	var api *v3.APIError
	var legacy *core.APIError
	if !errors.Is(err, ErrPreconditionFailed) || !errors.Is(err, v3.ErrPreconditionFailed) || !errors.Is(err, v3.ErrAPI) ||
		!errors.As(err, &stale) || !errors.As(err, &api) || !errors.As(err, &legacy) {
		t.Fatalf("412 chain: %#v", err)
	}
	if api.StatusCode != 412 || api.Detail != "stale" || legacy.StatusCode != 412 || calls != 2 {
		t.Fatalf("412 fields: %#v calls=%d", api, calls)
	}
}

func assertContractViolation(t *testing.T, err error, status int, problem string) *v3.APIError {
	t.Helper()
	var api *v3.APIError
	if !errors.Is(err, v3.ErrAPI) || !errors.As(err, &api) || errors.Is(err, v3.ErrConnection) || errors.Is(err, v3.ErrValidation) {
		t.Fatalf("expected API contract violation, got %#v", err)
	}
	if api.StatusCode != status || !strings.HasPrefix(err.Error(), "v3: API error: ") || !strings.Contains(err.Error(), problem) {
		t.Fatalf("contract violation: status=%d message=%q", api.StatusCode, err.Error())
	}
	return api
}

func contractClient(handler func(*http.Request) *http.Response) *AgoraClient {
	return NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", V3: &V3Options{BaseURL: "https://stage.example"}, HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		return handler(r), nil
	})})
}

func TestAgentKitContractViolationsAreAPIErrors(t *testing.T) {
	ctx := context.Background()
	numberPage := func(body string) *http.Response {
		resp := v3Reply(200, body, "")
		resp.Header.Set("Agora-Request-Id", "req-page")
		return resp
	}

	t.Run("session without agent_session_id", func(t *testing.T) {
		c := contractClient(func(*http.Request) *http.Response {
			resp := v3Reply(200, `{"status":"running"}`, "")
			resp.Header.Set("Agora-Request-Id", "req-session")
			return resp
		})
		_, err := (&SavedAgent{ID: "saved1", client: c}).CreateSession(CreateSessionOptions{Channel: "c", AgentUID: "1", RemoteUIDs: []string{"2"}}).Start(ctx)
		api := assertContractViolation(t, err, 200, "session response has no agent_session_id")
		if api.RequestID != "req-session" || string(api.Body) != `{"status":"running"}` {
			t.Fatalf("fields: %#v", api)
		}
	})

	t.Run("saved agent without ETag", func(t *testing.T) {
		calls := 0
		c := contractClient(func(*http.Request) *http.Response {
			calls++
			return v3Reply(200, `{"agent_id":"id","agent_name":"Old","create_ts":1,"update_ts":2}`, "")
		})
		_, err := NewAgent(c).Save(ctx, SaveOptions{Name: "New", AgentID: "id"})
		assertContractViolation(t, err, 200, "agent response has no ETag")
		if calls != 1 {
			t.Fatalf("PATCH sent without an ETag: calls=%d", calls)
		}
	})

	t.Run("number without ETag", func(t *testing.T) {
		calls := 0
		c := contractClient(func(*http.Request) *http.Response {
			calls++
			return v3Reply(200, `{"number_id":"n1","number":"+12025550123","create_ts":1}`, "")
		})
		_, err := (&SavedAgent{ID: "a1", client: c}).AssignNumber(ctx, "n1", AssignNumberOptions{})
		assertContractViolation(t, err, 200, "number response has no ETag")
		if calls != 1 {
			t.Fatalf("PATCH sent without an ETag: calls=%d", calls)
		}
	})

	t.Run("list numbers repeated cursor", func(t *testing.T) {
		c := contractClient(func(*http.Request) *http.Response {
			return numberPage(`{"data":{"count":0,"list":[]},"has_more":true,"next_cursor":"same"}`)
		})
		_, err := (&SavedAgent{ID: "a1", client: c}).ListNumbers(ctx)
		api := assertContractViolation(t, err, 200, `repeated next cursor "same"`)
		if api.RequestID != "req-page" {
			t.Fatalf("request ID: %#v", api)
		}
	})

	t.Run("resolve number missing cursor", func(t *testing.T) {
		c := contractClient(func(*http.Request) *http.Response {
			return numberPage(`{"data":{"count":0,"list":[]},"has_more":true,"next_cursor":null}`)
		})
		_, err := (&SavedAgent{ID: "a1", client: c}).AssignNumber(ctx, "+12025550123", AssignNumberOptions{})
		assertContractViolation(t, err, 200, "has more results but no next cursor")
		if errors.Is(err, ErrNumberNotFound) {
			t.Fatal("a broken page must not be reported as number not found")
		}
	})

	t.Run("invalid inbound value", func(t *testing.T) {
		c := contractClient(func(*http.Request) *http.Response {
			return numberPage(`{"data":{"count":1,"list":[{"number_id":"n1","number":"+12025550123","create_ts":1,"inbound":"oops"}]},"has_more":false,"next_cursor":null}`)
		})
		_, err := (&SavedAgent{ID: "a1", client: c}).ListNumbers(ctx)
		assertContractViolation(t, err, 200, "invalid inbound value for number n1")
		var typeErr *json.UnmarshalTypeError
		if !errors.As(err, &typeErr) {
			t.Fatalf("decode cause not reachable: %#v", err)
		}
	})
}
