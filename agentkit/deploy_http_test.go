package agentkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/AgoraIO/agora-agents-go/v2/option"
	"github.com/AgoraIO/agora-agents-go/v2/v3"
)

type v3HTTPFunc func(*http.Request) (*http.Response, error)

func (f v3HTTPFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }
func v3Reply(code int, body, etag string) *http.Response {
	h := http.Header{}
	if etag != "" {
		h.Set("ETag", etag)
	}
	return &http.Response{StatusCode: code, Header: h, Body: io.NopCloser(strings.NewReader(body))}
}

func TestDeployUpdateUsesFreshETagAndClearsStale(t *testing.T) {
	calls := 0
	httpClient := v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Path != "/conversational-ai/v3/projects/app/agents/id" {
			t.Fatalf("path: %s", r.URL.Path)
		}
		switch calls {
		case 1:
			if r.Method != "GET" {
				t.Fatalf("first request: %s", r.Method)
			}
			return v3Reply(200, `{"agent_id":"id","agent_name":"Old","create_ts":1,"update_ts":2,"pipeline":{"mode":"cascaded","llm":{"vendor":"openai"}}}`, "e1"), nil
		case 2:
			if r.Method != "PATCH" || r.Header.Get("If-Match") != "e1" {
				t.Fatalf("patch headers: %s %v", r.Method, r.Header)
			}
			b, _ := io.ReadAll(r.Body)
			var patch map[string]any
			if err := json.Unmarshal(b, &patch); err != nil {
				t.Fatal(err)
			}
			pipe := patch["pipeline"].(map[string]any)
			if value, ok := pipe["llm"]; !ok || value != nil {
				t.Fatalf("stale LLM not cleared: %s", b)
			}
			return v3Reply(200, `{"agent_id":"id","agent_name":"New","create_ts":1,"update_ts":3}`, "e2"), nil
		default:
			t.Fatalf("unexpected request %d", calls)
			return nil, nil
		}
	})
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: httpClient, V3: &V3Options{BaseURL: "https://stage.example", Headers: http.Header{"X-Studio-V2-Gateway-Secret": []string{"gateway"}}}})
	a := NewAgent(c)
	d, err := a.Deploy(context.Background(), DeployOptions{Name: "New", AgentID: "id"})
	if err != nil {
		t.Fatal(err)
	}
	if d.ID != "id" || d.ETag != "e2" || calls != 2 {
		t.Fatalf("deployed: %#v calls=%d", d, calls)
	}
}

func TestAssignUnassignNumberRetriesOneConflict(t *testing.T) {
	calls := 0
	httpClient := v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch calls {
		case 1:
			if r.Method != "GET" || !strings.HasSuffix(r.URL.Path, "/numbers") {
				t.Fatalf("list: %s %s", r.Method, r.URL.Path)
			}
			return v3Reply(200, `{"data":{"count":1,"list":[{"number_id":"n1","number":"+12025550123","create_ts":1}]},"has_more":false,"next_cursor":null}`, ""), nil
		case 2, 4, 6:
			if r.Method != "GET" || !strings.HasSuffix(r.URL.Path, "/numbers/n1") {
				t.Fatalf("get: %s %s", r.Method, r.URL.Path)
			}
			etag := "e1"
			if calls == 4 {
				etag = "e2"
			}
			return v3Reply(200, `{"number_id":"n1","number":"+12025550123","create_ts":1}`, etag), nil
		case 3:
			if r.Method != "PATCH" {
				t.Fatalf("patch: %s", r.Method)
			}
			return v3Reply(412, `{"detail":"stale"}`, ""), nil
		case 5, 7:
			if calls == 5 && r.Header.Get("If-Match") != "e2" {
				t.Fatalf("retry used stale ETag: %q", r.Header.Get("If-Match"))
			}
			b, _ := io.ReadAll(r.Body)
			var body map[string]any
			json.Unmarshal(b, &body)
			in := body["inbound"].(map[string]any)
			if calls == 5 && in["agent"] != "a1" {
				t.Fatalf("assign: %s", b)
			}
			if calls == 7 {
				if v, ok := in["agent"]; !ok || v != nil {
					t.Fatalf("unassign: %s", b)
				}
			}
			if _, ok := in["sip_trunk"]; ok {
				t.Fatalf("echoed number details: %s", b)
			}
			return v3Reply(200, `{"number_id":"n1","number":"+12025550123","create_ts":1}`, "e2"), nil
		default:
			t.Fatalf("unexpected request %d", calls)
			return nil, nil
		}
	})
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: httpClient, V3: &V3Options{BaseURL: "https://stage.example"}})
	d := &DeployedAgent{ID: "a1", Raw: c.V3.Agents, client: c}
	if _, err := d.AssignNumber(context.Background(), "+1 (202) 555-0123", AssignNumberOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := d.UnassignNumber(context.Background(), v3.Number{NumberID: "n1"}); err != nil {
		t.Fatal(err)
	}
	if calls != 7 {
		t.Fatalf("requests: %d", calls)
	}
}

func TestListNumbersScansPagesWithoutDetailGets(t *testing.T) {
	calls := 0
	httpClient := v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != "GET" || !strings.HasSuffix(r.URL.Path, "/numbers") {
			t.Fatalf("unexpected detail request: %s %s", r.Method, r.URL.Path)
		}
		if calls == 1 {
			if r.URL.Query().Get("cursor") != "" {
				t.Fatal("first cursor")
			}
			return v3Reply(200, `{"data":{"count":3,"list":[{"number_id":"n1","number":"+111","create_ts":1,"inbound":{"agent":"a1"}},{"number_id":"n2","number":"+222","create_ts":1,"inbound":{"agent":"other"}},{"number_id":"n3","number":"+333","create_ts":1,"inbound":null}]},"has_more":true,"next_cursor":"page2"}`, ""), nil
		}
		if calls == 2 {
			if r.URL.Query().Get("cursor") != "page2" {
				t.Fatal("second cursor")
			}
			return v3Reply(200, `{"data":{"count":1,"list":[{"number_id":"n4","number":"+444","create_ts":1,"inbound":{"agent":"a1"}}]},"has_more":false,"next_cursor":null}`, ""), nil
		}
		t.Fatal("unexpected third page")
		return nil, nil
	})
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: httpClient, V3: &V3Options{BaseURL: "https://stage.example"}})
	d := &DeployedAgent{ID: "a1", Raw: c.V3.Agents, client: c}
	numbers, err := d.ListNumbers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || len(numbers) != 2 || numbers[0].NumberID != "n1" || numbers[1].NumberID != "n4" {
		t.Fatalf("scan: calls=%d numbers=%#v", calls, numbers)
	}
}

func TestDeployPreservesCascadedInteractionLanguage(t *testing.T) {
	for _, withVendor := range []bool{false, true} {
		t.Run(fmt.Sprintf("explicit_vendor_%t", withVendor), func(t *testing.T) {
			calls := 0
			c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", V3: &V3Options{BaseURL: "https://stage.example"}, HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				var body struct {
					Pipeline struct {
						ASR  map[string]any `json:"asr"`
						Turn map[string]any `json:"turn_detection"`
					} `json:"pipeline"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				wantVendor := "ares"
				if withVendor {
					wantVendor = "deepgram"
				}
				if body.Pipeline.ASR["language"] != "fr-FR" || body.Pipeline.ASR["vendor"] != wantVendor {
					t.Fatalf("ASR language mapping: %#v", body.Pipeline)
				}
				if _, exists := body.Pipeline.Turn["language"]; exists {
					t.Fatal("language leaked into v3 turn detection")
				}
				return v3Reply(201, `{"agent_id":"id","agent_name":"French","create_ts":1,"update_ts":1}`, "e1"), nil
			})})
			language := Agora.AsrLanguage("fr-FR")
			a := NewAgent(c).WithTurnDetection(&TurnDetectionConfig{Language: &language})
			if withVendor {
				a = a.WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{APIKey: "asr-key", Model: "nova-3"}))
			}
			if _, err := a.Deploy(context.Background(), DeployOptions{Name: "French"}); err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Fatalf("calls = %d", calls)
			}
		})
	}
}

func TestDeployRejectsRealtimeInteractionLanguage(t *testing.T) {
	calls := 0
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) { calls++; return nil, errors.New("unexpected request") })})
	language := Agora.AsrLanguage("fr-FR")
	a := NewAgent(c).WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{APIKey: "key", TurnDetection: &Agora.MllmTurnDetection{Mode: Agora.MllmTurnDetectionModeAgoraVad.Ptr()}})).WithTurnDetection(&TurnDetectionConfig{Language: &language})
	_, err := a.Deploy(context.Background(), DeployOptions{Name: "French"})
	if !errors.Is(err, ErrUnsupportedField) || !strings.Contains(err.Error(), "turnDetection.language") || calls != 0 {
		t.Fatalf("realtime language: err=%v calls=%d", err, calls)
	}
}

func TestVariableOptionSnapshotsAtConstructionAndCanBeReused(t *testing.T) {
	calls := 0
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", V3: &V3Options{BaseURL: "https://stage.example"}, HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		var body struct {
			Variables map[string]VariableDefinition `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if got := body.Variables["choice"].Options; len(got) != 1 || got[0] != "original" {
			t.Fatalf("caller mutation leaked: %#v", got)
		}
		return v3Reply(201, `{"agent_id":"id","agent_name":"Options","create_ts":1,"update_ts":1}`, "e1"), nil
	})})
	definition := VariableDefinition{Type: "str", Options: []any{"original"}}
	variable := WithVariable("choice", definition)
	definition.Options[0] = "changed before application"
	first := NewAgent(c, variable)
	definition.Options[0] = "changed before reuse"
	second := NewAgent(c, variable)
	for _, agent := range []*Agent{first, second} {
		if _, err := agent.Deploy(context.Background(), DeployOptions{Name: "Options"}); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 2 {
		t.Fatalf("calls = %d", calls)
	}
}

func TestDeployNormalizesLegacyGreetingModes(t *testing.T) {
	cases := []struct {
		name     string
		config   map[string]interface{}
		greeting string
		want     string
		bad      bool
	}{
		{"each", map[string]interface{}{"mode": "single_every"}, "Hello", "each_join", false},
		{"first", map[string]interface{}{"mode": "single_first"}, "Hello", "first_join", false},
		{"explicit", map[string]interface{}{"mode": "single_every", "on": "first_join"}, "Hello", "first_join", false},
		{"unknown", map[string]interface{}{"mode": "unknown"}, "Hello", "", true},
		{"missing greeting", map[string]interface{}{"mode": "single_every"}, "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", V3: &V3Options{BaseURL: "https://stage.example"}, HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				var body struct {
					Greeting map[string]any `json:"greeting"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body.Greeting["mode"] != "text" || body.Greeting["text"] != "Hello" || body.Greeting["on"] != tc.want {
					t.Fatalf("greeting: %#v", body.Greeting)
				}
				return v3Reply(201, `{"agent_id":"id","agent_name":"Greeting","create_ts":1,"update_ts":1}`, "e1"), nil
			})})
			a := NewAgent(c).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{APIKey: "key", BaseURL: "https://api.openai.com/v1/chat/completions", Model: "gpt-4o-mini", GreetingMessage: tc.greeting, GreetingConfigs: tc.config}))
			_, err := a.Deploy(context.Background(), DeployOptions{Name: "Greeting"})
			if tc.bad {
				if !errors.Is(err, ErrUnsupportedField) || calls != 0 {
					t.Fatalf("expected local rejection: err=%v calls=%d", err, calls)
				}
			} else if err != nil || calls != 1 {
				t.Fatalf("deploy: err=%v calls=%d", err, calls)
			}
		})
	}
}
