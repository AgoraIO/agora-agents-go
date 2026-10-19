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

func TestDeployUnicodeDescriptionsAndEmptyMessages(t *testing.T) {
	for _, length := range []int{1000, 2000} {
		t.Run(fmt.Sprintf("description_%d_characters", length), func(t *testing.T) {
			description := strings.Repeat("解", length)
			calls := 0
			c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", V3: &V3Options{BaseURL: "https://stage.example"}, HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				var body struct {
					Instructions *string                       `json:"instructions"`
					Messages     []v3.Message                  `json:"messages"`
					Variables    map[string]VariableDefinition `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body.Variables["outcome"].Description != description {
					t.Fatal("Unicode description changed")
				}
				if body.Instructions == nil || *body.Instructions != "" {
					t.Fatalf("empty system message lost: %#v", body.Instructions)
				}
				if len(body.Messages) != 1 || body.Messages[0].Role != "user" || body.Messages[0].Content != "" {
					t.Fatalf("empty user content lost: %#v", body.Messages)
				}
				return v3Reply(201, `{"agent_id":"id","agent_name":"Unicode","create_ts":1,"update_ts":1}`, "e1"), nil
			})})
			a := NewAgent(c, WithVariable("outcome", VariableDefinition{Type: "str", Extract: true, Description: description})).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
				APIKey: "key", BaseURL: "https://api.openai.com/v1/chat/completions", Model: "gpt-4o-mini",
				SystemMessages: []map[string]interface{}{{"role": "system", "content": ""}, {"role": "user", "content": ""}},
			}))
			if _, err := a.Deploy(context.Background(), DeployOptions{Name: "Unicode"}); err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Fatalf("calls=%d", calls)
			}
		})
	}
	t.Run("over limit", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("accepted description longer than 2000 Unicode characters")
			}
		}()
		WithVariable("outcome", VariableDefinition{Type: "str", Description: strings.Repeat("解", 2001)})
	})
}

func TestDeployRejectsMissingMessageContent(t *testing.T) {
	for _, message := range []map[string]interface{}{{"role": "user"}, {"role": "user", "content": nil}} {
		calls := 0
		c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) { calls++; return nil, errors.New("unexpected request") })})
		a := NewAgent(c).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{APIKey: "key", BaseURL: "https://api.openai.com/v1/chat/completions", Model: "gpt-4o-mini", SystemMessages: []map[string]interface{}{message}}))
		_, err := a.Deploy(context.Background(), DeployOptions{Name: "Missing content"})
		if !errors.Is(err, ErrUnsupportedField) || calls != 0 {
			t.Fatalf("message=%#v err=%v calls=%d", message, err, calls)
		}
	}
}

func TestDeployAudioGreetingSettingsStayAudioOnly(t *testing.T) {
	for _, audio := range []bool{false, true} {
		for _, field := range []string{"audio_download_timeout_ms", "audio_pcm_sample_rate"} {
			t.Run(fmt.Sprintf("audio_%t_%s", audio, field), func(t *testing.T) {
				calls := 0
				c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", V3: &V3Options{BaseURL: "https://stage.example"}, HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
					calls++
					var body struct {
						Greeting map[string]any `json:"greeting"`
					}
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatal(err)
					}
					wantField := "download_timeout_ms"
					if field == "audio_pcm_sample_rate" {
						wantField = "pcm_sample_rate"
					}
					if body.Greeting["mode"] != "audio" || body.Greeting[wantField] != float64(1000) {
						t.Fatalf("audio greeting=%#v", body.Greeting)
					}
					return v3Reply(201, `{"agent_id":"id","agent_name":"Audio","create_ts":1,"update_ts":1}`, "e1"), nil
				})})
				config := vendors.OpenAIOptions{APIKey: "key", BaseURL: "https://api.openai.com/v1/chat/completions", Model: "gpt-4o-mini", GreetingMessage: "Hello", GreetingConfigs: map[string]interface{}{field: 1000}}
				a := NewAgent(c).WithLlm(vendors.NewOpenAI(config))
				if audio {
					a = a.WithGreetingAudioURL("https://example.com/hello.wav")
				}
				_, err := a.Deploy(context.Background(), DeployOptions{Name: "Audio"})
				if audio {
					if err != nil || calls != 1 {
						t.Fatalf("audio deploy: err=%v calls=%d", err, calls)
					}
				} else if !errors.Is(err, ErrUnsupportedField) || calls != 0 {
					t.Fatalf("text greeting accepted audio settings: err=%v calls=%d", err, calls)
				}
			})
		}
	}
}

func TestSaveSavedAgentUpdateAndProjectResourceAliases(t *testing.T) {
	calls := 0
	httpClient := v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch calls {
		case 1:
			if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/agents") {
				t.Fatalf("create request: %s %s", r.Method, r.URL.Path)
			}
			return v3Reply(201, `{"agent_id":"a1","agent_name":"Support","create_ts":1,"update_ts":1}`, "e1"), nil
		case 2:
			if r.Method != "GET" || !strings.HasSuffix(r.URL.Path, "/agents/a1") {
				t.Fatalf("get request: %s %s", r.Method, r.URL.Path)
			}
			return v3Reply(200, `{"agent_id":"a1","agent_name":"Support","create_ts":1,"update_ts":1}`, "e1"), nil
		case 3:
			if r.Method != "PATCH" || r.Header.Get("If-Match") != "e1" {
				t.Fatalf("update request: %s %v", r.Method, r.Header)
			}
			var patch map[string]any
			if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
				t.Fatal(err)
			}
			if patch["agent_name"] != "Support" {
				t.Fatalf("update did not retain saved name: %#v", patch)
			}
			return v3Reply(200, `{"agent_id":"a1","agent_name":"Support","create_ts":1,"update_ts":2}`, "e2"), nil
		default:
			t.Fatalf("unexpected request %d", calls)
			return nil, nil
		}
	})
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: httpClient, V3: &V3Options{BaseURL: "https://stage.example"}})
	if c.Secrets != c.V3.Secrets || c.Numbers != c.V3.Numbers {
		t.Fatal("version-neutral resource accessors must share the V3 client resources")
	}
	a := NewAgent(c)
	saved, err := a.Save(context.Background(), SaveOptions{Name: "Support"})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := saved.Update(context.Background(), a.WithVariable("priority", VariableDefinition{Type: "str"}))
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID != "a1" || saved.ETag != "e1" || updated.ID != "a1" || updated.ETag != "e2" || updated.Name != "Support" || calls != 3 {
		t.Fatalf("save/update handles: saved=%#v updated=%#v calls=%d", saved, updated, calls)
	}
	if a.base.Variables != nil {
		t.Fatal("Update mutated the original Agent config")
	}
}

func TestSavedAgentUpdateRejectsDifferentClient(t *testing.T) {
	calls := 0
	transport := v3HTTPFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("unexpected request")
	})
	first := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: transport})
	second := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: transport})
	saved := &SavedAgent{ID: "a1", Name: "Support", client: first}
	_, err := saved.Update(context.Background(), NewAgent(second))
	if err == nil || !strings.Contains(err.Error(), "same AgoraClient instance") || calls != 0 {
		t.Fatalf("cross-client update: err=%v requests=%d", err, calls)
	}
}

func TestInlineConfigCanStartSessionAndSaveWithoutMutation(t *testing.T) {
	calls := 0
	inlineKey := "inline-provider-key"
	storedAgent := `{"agent_id":"saved1","agent_name":"Support","create_ts":1,"update_ts":1,"pipeline":{"mode":"realtime","mllm":{"vendor":"openai","credential":{"mode":"byok","api_key":"inline-provider-key"},"url":"wss://api.openai.com/v1/realtime","model":"gpt-4o-realtime-preview"}}}`
	httpClient := v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch calls {
		case 1:
			if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/join") {
				t.Fatalf("session request: %s %s", r.Method, r.URL.Path)
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			properties := body["properties"].(map[string]any)
			mllm := properties["mllm"].(map[string]any)
			if mllm["api_key"] != inlineKey {
				t.Fatalf("inline key missing from v2 session request: %#v", mllm)
			}
			return v3Reply(200, `{"agent_id":"session1","status":"RUNNING"}`, ""), nil
		case 2:
			if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/agents") {
				t.Fatalf("save request: %s %s", r.Method, r.URL.Path)
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			pipeline := body["pipeline"].(map[string]any)
			mllm := pipeline["mllm"].(map[string]any)
			credential := mllm["credential"].(map[string]any)
			if credential["api_key"] != inlineKey {
				t.Fatalf("inline key missing from saved config: %#v", mllm)
			}
			return v3Reply(201, storedAgent, "e1"), nil
		case 3:
			if r.Method != "GET" || !strings.HasSuffix(r.URL.Path, "/agents/saved1") {
				t.Fatalf("get request: %s %s", r.Method, r.URL.Path)
			}
			return v3Reply(200, storedAgent, "e1"), nil
		case 4:
			if r.Method != "PATCH" || r.Header.Get("If-Match") != "e1" {
				t.Fatalf("update request: %s %v", r.Method, r.Header)
			}
			var patch map[string]any
			if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
				t.Fatal(err)
			}
			pipeline := patch["pipeline"].(map[string]any)
			mllm := pipeline["mllm"].(map[string]any)
			credential := mllm["credential"].(map[string]any)
			if credential["api_key"] != inlineKey {
				t.Fatalf("inline key missing from updated config: %#v", mllm)
			}
			return v3Reply(200, storedAgent, "e2"), nil
		default:
			t.Fatalf("unexpected request %d", calls)
			return nil, nil
		}
	})
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: httpClient, V3: &V3Options{BaseURL: "https://stage.example"}})
	config := NewAgent(c).WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
		APIKey: inlineKey,
		Model:  "gpt-4o-realtime-preview",
	}))
	session := config.CreateSession(CreateSessionOptions{Channel: "support", Token: "rtc-token", AgentUID: "1", RemoteUIDs: []string{"100"}})
	if _, err := session.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	saved, err := config.Save(context.Background(), SaveOptions{Name: "Support"})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := saved.Update(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != "saved1" || updated.ETag != "e2" || calls != 4 {
		t.Fatalf("update result: %#v calls=%d", updated, calls)
	}
	properties, err := config.ToPropertiesMap(basePropertiesOpts())
	if err != nil {
		t.Fatal(err)
	}
	mllm := properties["mllm"].(map[string]any)
	if mllm["api_key"] != inlineKey {
		t.Fatalf("session/save/update mutated the source config: %#v", mllm)
	}
}
