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
	"github.com/AgoraIO/agora-agents-go/v2/core"
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
		if r.URL.Path != "/conversational-ai/v1/projects/app/agents/id" {
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
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: httpClient, V3: &V3Options{BaseURL: "https://stage.example"}})
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

func TestManagedProviderOmitsURLsButBYOKRetainsThem(t *testing.T) {
	managed, bad := convertModule("llm", map[string]interface{}{"vendor": "openai", "url": "https://managed.invalid", "params": map[string]interface{}{"base_url": "https://also.invalid", "temperature": 0.3}})
	if len(bad) > 0 {
		t.Fatalf("managed module conversion: %v", bad)
	}
	if _, ok := managed["url"]; ok {
		t.Fatalf("managed top-level URL should be omitted: %#v", managed)
	}
	if _, ok := managed["params"].(map[string]interface{})["url"]; ok {
		t.Fatalf("managed params URL should be omitted: %#v", managed)
	}
	byok, bad := convertModule("llm", map[string]interface{}{"vendor": "openai", "api_key": "secret", "url": "https://byok.example/v1", "params": map[string]interface{}{"temperature": 0.3}})
	if len(bad) > 0 {
		t.Fatalf("BYOK module conversion: %v", bad)
	}
	if byok["url"] != "https://byok.example/v1" || byok["credential"].(map[string]any)["mode"] != "byok" {
		t.Fatalf("BYOK URL/credential lost: %#v", byok)
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

func TestSavedAgentCreateSessionUsesV3SessionLifecycle(t *testing.T) {
	calls := 0
	httpClient := v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch calls {
		case 1:
			if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/sessions") {
				t.Fatalf("start request: %s %s", r.Method, r.URL.Path)
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if r.Header.Get("Idempotency-Key") != "session-key" {
				t.Fatalf("idempotency header: %v", r.Header)
			}
			if body["agent_id"] != "saved1" || body["client_reference"] != nil || body["labels"].(map[string]any)["team"] != "support" || body["values"].(map[string]any)["customer"] != "Ada" {
				t.Fatalf("start body: %#v", body)
			}
			transport := body["transport"].(map[string]any)
			if transport["type"] != "rtc" || transport["uid"] != "1001" || transport["channel"] != "support" || transport["token"] != "rtc-token" || transport["audio_scenario"] != "chorus" || transport["encryption"].(map[string]any)["mode"] != "aes_128_gcm" || transport["geofence"].(map[string]any)["exclude_area"] != "india" {
				t.Fatalf("transport: %#v", transport)
			}
			lifecycle := body["lifecycle"].(map[string]any)
			if lifecycle["idle_timeout_ms"] != float64(60000) || lifecycle["max_duration_ms"] != float64(120000) || lifecycle["graceful_stop"].(map[string]any)["timeout_ms"] != float64(5000) {
				t.Fatalf("lifecycle unit conversion: %#v", lifecycle)
			}
			if body["data_policy"].(map[string]any)["retention"] != "none" {
				t.Fatalf("data policy: %#v", body["data_policy"])
			}
			return v3Reply(200, `{"agent_session_id":"sess1","status":"starting","create_ts":1781827200}`, ""), nil
		case 2:
			if r.Method != http.MethodGet || !strings.HasSuffix(r.URL.Path, "/sessions/sess1") {
				t.Fatalf("get session request: %s %s", r.Method, r.URL.Path)
			}
			return v3Reply(200, `{"agent_session_id":"sess1","status":"RUNNING","start_ts":1781827200}`, ""), nil
		case 3:
			if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/sessions/sess1/stop") {
				t.Fatalf("stop request: %s %s", r.Method, r.URL.Path)
			}
			if r.Body != nil {
				body, _ := io.ReadAll(r.Body)
				if len(body) != 0 {
					t.Fatalf("stop body must be empty: %s", body)
				}
			}
			return v3Reply(200, `{}`, ""), nil
		default:
			t.Fatalf("unexpected request %d", calls)
			return nil, nil
		}
	})
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: httpClient, V3: &V3Options{BaseURL: "https://stage.example"}})
	idle, max := 60, 120
	session := (&SavedAgent{ID: "saved1", Name: "Support", client: c}).CreateSession(CreateSessionOptions{
		Channel: "support", Token: "rtc-token", AgentUID: "1001", RemoteUIDs: []string{"1002"},
		IdleTimeout: &idle, MaxDuration: &max, Labels: map[string]string{"team": "support"},
		IdempotencyKey: "session-key", Values: map[string]any{"customer": "Ada"},
		Encryption: &RTCEncryption{Mode: "aes_128_gcm", Key: "encryption-key"}, AudioScenario: "chorus",
		Geofence: &SessionGeofence{Area: "global", ExcludeArea: "india"}, GracefulStop: &GracefulStopOptions{Enabled: true, TimeoutMS: ptrInt(5000)},
		DataPolicy: &SessionDataPolicy{Retention: "none"},
	})
	id, err := session.Start(context.Background())
	if err != nil || id != "sess1" || session.ID() != "sess1" || session.Status() != StatusStarting {
		t.Fatalf("session start: id=%q status=%q err=%v", id, session.Status(), err)
	}
	info, err := session.GetInfo(context.Background())
	if err != nil || info.GetAgentSessionID() == nil || *info.GetAgentSessionID() != "sess1" || info.GetAgentID() != nil || info.GetStatus() == nil || string(*info.GetStatus()) != "RUNNING" || session.Status() != StatusRunning {
		t.Fatalf("session info: %#v err=%v", info, err)
	}
	if err := session.Stop(context.Background()); err != nil || session.Status() != StatusStopped || calls != 3 {
		t.Fatalf("session stop: status=%q err=%v calls=%d", session.Status(), err, calls)
	}
	if _, err := session.Start(context.Background()); err == nil || calls != 3 {
		t.Fatalf("idempotent session restart should fail locally: err=%v calls=%d", err, calls)
	}
}

func TestAgentConfigWithSecretRefStartsEphemeralV3Session(t *testing.T) {
	calls := 0
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", V3: &V3Options{BaseURL: "https://stage.example"}, HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/sessions/ephemeral1") {
			return v3Reply(200, `{"agent_session_id":"ephemeral1","status":"RUNNING"}`, ""), nil
		}
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/sessions/ephemeral") {
			t.Fatalf("ephemeral request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		agent := body["agent"].(map[string]any)
		pipeline := agent["pipeline"].(map[string]any)
		mllm := pipeline["mllm"].(map[string]any)
		credential := mllm["credential"].(map[string]any)
		if credential["api_key"] != "$secrets.prod.openai" {
			t.Fatalf("secret reference lost: %#v", mllm)
		}
		if len(agent["variables"].(map[string]any)) != 1 {
			t.Fatalf("variable missing from ephemeral agent: %#v", agent)
		}
		return v3Reply(201, `{"agent_session_id":"ephemeral1","status":"running","created_at":1781827200}`, ""), nil
	})})
	config := NewAgent(c).
		WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{APIKey: SecretRef("prod", "openai")})).
		WithVariable("priority", VariableDefinition{Type: "str", Description: "Main priority"})
	session := config.CreateSession(CreateSessionOptions{
		Channel: "support", AgentUID: "1001", RemoteUIDs: []string{"1002"},
	})
	id, err := session.Start(context.Background())
	if err != nil || id != "ephemeral1" || calls != 1 {
		t.Fatalf("ephemeral start: id=%q err=%v calls=%d", id, err, calls)
	}
	if err := session.Say(context.Background(), "hello", nil, nil); !errors.Is(err, ErrUnsupportedSessionOperation) {
		t.Fatalf("Say should be rejected locally for v3 session: %v", err)
	}
	info, err := session.GetInfo(context.Background())
	if err != nil || info.GetAgentSessionID() == nil || *info.GetAgentSessionID() != "ephemeral1" {
		t.Fatalf("GetInfo should retain the unified session interface: %#v %v", info, err)
	}
	if calls != 2 {
		t.Fatalf("expected only start and info requests: %d", calls)
	}
}

func TestV3SessionCanStopStartingAndPropagatesStopErrors(t *testing.T) {
	for _, tc := range []struct {
		code   int
		reason string
		status AgentSessionLifecycle
	}{{404, "TaskNotFound", StatusError}, {409, "TaskNotRunning", StatusStopped}, {409, "TaskConflict", StatusStopping}} {
		t.Run(tc.reason, func(t *testing.T) {
			calls := 0
			c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", V3: &V3Options{BaseURL: "https://stage.example"}, HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					return v3Reply(200, `{"agent_session_id":"session1","status":"STARTING","create_ts":1}`, ""), nil
				}
				return v3Reply(tc.code, `{"detail":"stop rejected","reason":"`+tc.reason+`"}`, ""), nil
			})})
			session := (&SavedAgent{ID: "agent1", client: c}).CreateSession(CreateSessionOptions{Channel: "room", AgentUID: "1001", RemoteUIDs: []string{"1002"}})
			if _, err := session.Start(context.Background()); err != nil || session.Status() != StatusStarting {
				t.Fatalf("start state: %q %v", session.Status(), err)
			}
			err := session.Stop(context.Background())
			var apiErr *core.APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != tc.code || calls != 2 || session.Status() != tc.status {
				t.Fatalf("stop error/state: err=%v status=%q calls=%d", err, session.Status(), calls)
			}
		})
	}
}

func boolPtrTest(v bool) *bool { return &v }

func TestV3SessionSnapshotsOptionsAndRejectsCN(t *testing.T) {
	calls := 0
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", V3: &V3Options{BaseURL: "https://stage.example"}, HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		transport := body["transport"].(map[string]any)
		if transport["subscribe_uids"].([]any)[0] != "original" {
			t.Fatalf("remote UIDs were not snapshotted: %#v", transport)
		}
		lifecycle := body["lifecycle"].(map[string]any)
		if lifecycle["idle_timeout_ms"] != float64(5000) {
			t.Fatalf("timeout was not snapshotted: %#v", lifecycle)
		}
		return v3Reply(201, `{"agent_session_id":"snapshot","status":"RUNNING"}`, ""), nil
	})})
	idle := 5
	remoteUIDs := []string{"original"}
	session := (&SavedAgent{ID: "saved", client: c}).CreateSession(CreateSessionOptions{Channel: "room", AgentUID: "agent-original", RemoteUIDs: remoteUIDs, IdleTimeout: &idle})
	remoteUIDs[0] = "mutated"
	idle = 100
	if _, err := session.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if session.Raw() != nil || session.RawAgentManagement() != nil {
		t.Fatal("v3 session exposed legacy raw clients")
	}

	cnCalls := 0
	cn := NewAgoraClient(AgoraClientOptions{Area: option.AreaCN, AppID: "app", Token: "token", V3: &V3Options{BaseURL: "https://stage.example"}, HTTPClient: v3HTTPFunc(func(*http.Request) (*http.Response, error) { cnCalls++; return nil, errors.New("unexpected request") })})
	blocked := (&SavedAgent{ID: "saved", client: cn}).CreateSession(CreateSessionOptions{Channel: "room", AgentUID: "10"})
	if _, err := blocked.Start(context.Background()); err == nil || cnCalls != 0 {
		t.Fatalf("CN v3 session was not blocked locally: err=%v calls=%d", err, cnCalls)
	}
}

func TestV3SessionRejectsLegacyAndUndocumentedOptionsLocally(t *testing.T) {
	calls := 0
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", V3: &V3Options{BaseURL: "https://stage.example"}, HTTPClient: v3HTTPFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("unexpected request")
	})})
	saved := &SavedAgent{ID: "saved1", Name: "Support", client: c}
	cases := []struct {
		name string
		opts CreateSessionOptions
	}{
		{"client reference", CreateSessionOptions{Channel: "c", AgentUID: "1", ClientReference: "stale"}},
		{"expires in", CreateSessionOptions{Channel: "c", AgentUID: "1", ExpiresIn: 60}},
		{"preset", CreateSessionOptions{Channel: "c", AgentUID: "1", Preset: []string{"llm.openai.gpt-4o-mini"}}},
		{"pipeline override", CreateSessionOptions{Channel: "c", AgentUID: "1", PipelineID: "pipeline"}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.opts.AgentUID = "1"
			session := saved.CreateSession(tt.opts)
			if _, err := session.Start(context.Background()); err == nil {
				t.Fatal("expected local unsupported-option error")
			}
		})
	}
	if calls != 0 {
		t.Fatalf("unsupported options reached transport: %d", calls)
	}
}

func TestAvatarOnlySecretReferenceSelectsEphemeralSession(t *testing.T) {
	calls := 0
	client := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Path != "/conversational-ai/v1/projects/app/sessions/ephemeral" {
			t.Fatalf("avatar reference routed to %s", r.URL.Path)
		}
		var body struct {
			Agent struct {
				Pipeline struct {
					Avatar struct {
						Credential struct {
							APIKey string `json:"api_key"`
						} `json:"credential"`
					} `json:"avatar"`
				} `json:"pipeline"`
			} `json:"agent"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Agent.Pipeline.Avatar.Credential.APIKey != "$secrets.avatars.anam" {
			t.Fatalf("avatar reference lost: %#v", body)
		}
		return v3Reply(200, `{"agent_session_id":"avatar-session","status":"RUNNING","create_ts":1}`, ""), nil
	})})
	config := NewAgent(client).WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{APIKey: "inline-provider-key"})).WithAvatar(vendors.NewAnamAvatar(vendors.AnamAvatarOptions{APIKey: SecretRef("avatars", "anam"), AvatarID: "avatar"}))
	session := config.CreateSession(CreateSessionOptions{Channel: "room", AgentUID: "1", RemoteUIDs: []string{"2"}})
	id, err := session.Start(context.Background())
	if err != nil || id != "avatar-session" || calls != 1 {
		t.Fatalf("avatar reference start failed: id=%q error=%v calls=%d", id, err, calls)
	}
}

func TestEphemeralSessionInheritsBuilderLabelsWithoutSavingThem(t *testing.T) {
	for _, override := range []bool{false, true} {
		t.Run(fmt.Sprint(override), func(t *testing.T) {
			expected := "builder"
			if override {
				expected = "session"
			}
			client := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body["labels"].(map[string]any)["source"] != expected {
					t.Fatalf("labels lost: %#v", body)
				}
				if _, ok := body["agent"].(map[string]any)["labels"]; ok {
					t.Fatal("session labels leaked into stored shape")
				}
				return v3Reply(200, `{"agent_session_id":"s","status":"RUNNING","create_ts":1}`, ""), nil
			})})
			config := NewAgent(client, WithLabels(map[string]string{"source": "builder"})).WithVariable("value", VariableDefinition{Type: "str"})
			opts := CreateSessionOptions{Channel: "room", AgentUID: "1", RemoteUIDs: []string{"2"}}
			if override {
				opts.Labels = map[string]string{"source": "session"}
			}
			session := config.CreateSession(opts)
			if _, err := session.Start(context.Background()); err != nil {
				t.Fatal(err)
			}
			if config.base.Labels["source"] != "builder" {
				t.Fatal("builder labels mutated")
			}
		})
	}
}

func TestSessionInfoCannotOverwriteCompletedStopOrRestart(t *testing.T) {
	for _, restart := range []bool{false, true} {
		t.Run(fmt.Sprint(restart), func(t *testing.T) {
			readEntered := make(chan struct{})
			releaseRead := make(chan struct{})
			readDone := make(chan error, 1)
			starts := 0
			client := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method == http.MethodGet {
					close(readEntered)
					<-releaseRead
					status := "RUNNING"
					if restart {
						status = "STOPPED"
					}
					return v3Reply(200, fmt.Sprintf(`{"agent_session_id":"first","status":%q}`, status), ""), nil
				}
				if strings.HasSuffix(r.URL.Path, "/stop") {
					return v3Reply(200, `{}`, ""), nil
				}
				starts++
				id := "first"
				if starts > 1 {
					id = "second"
				}
				return v3Reply(200, fmt.Sprintf(`{"agent_session_id":%q,"status":"RUNNING","create_ts":1}`, id), ""), nil
			})})
			saved := &SavedAgent{client: client, ID: "saved"}
			session := saved.CreateSession(CreateSessionOptions{Channel: "room", AgentUID: "1", RemoteUIDs: []string{"2"}})
			if _, err := session.Start(context.Background()); err != nil {
				t.Fatal(err)
			}
			go func() { _, err := session.GetInfo(context.Background()); readDone <- err }()
			<-readEntered
			if err := session.Stop(context.Background()); err != nil {
				t.Fatal(err)
			}
			expected := StatusStopped
			if restart {
				if _, err := session.Start(context.Background()); err != nil {
					t.Fatal(err)
				}
				expected = StatusRunning
			}
			close(releaseRead)
			if err := <-readDone; err != nil {
				t.Fatal(err)
			}
			if session.Status() != expected {
				t.Fatalf("stale read replaced lifecycle: got %s want %s", session.Status(), expected)
			}
		})
	}
}

func TestRestartDoesNotControlPreviousSessionWhileAwaitingIdentity(t *testing.T) {
	restartEntered := make(chan struct{})
	releaseRestart := make(chan struct{})
	restartDone := make(chan error, 1)
	starts, stops, reads := 0, 0, 0
	client := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method == http.MethodGet {
			reads++
			return v3Reply(200, `{}`, ""), nil
		}
		if strings.HasSuffix(r.URL.Path, "/stop") {
			stops++
			return v3Reply(200, `{}`, ""), nil
		}
		starts++
		id := "first"
		if starts == 2 {
			close(restartEntered)
			<-releaseRestart
			id = "second"
		}
		return v3Reply(200, fmt.Sprintf(`{"agent_session_id":%q,"status":"RUNNING","create_ts":1}`, id), ""), nil
	})})
	session := (&SavedAgent{client: client, ID: "saved"}).CreateSession(CreateSessionOptions{Channel: "room", AgentUID: "1", RemoteUIDs: []string{"2"}})
	if _, err := session.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := session.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	go func() { _, err := session.Start(context.Background()); restartDone <- err }()
	<-restartEntered
	if err := session.Stop(context.Background()); err == nil {
		t.Error("stop during pending restart should have no identity")
	}
	if _, err := session.GetInfo(context.Background()); err == nil {
		t.Error("get during pending restart should have no identity")
	}
	close(releaseRestart)
	if err := <-restartDone; err != nil {
		t.Fatal(err)
	}
	if starts != 2 || stops != 1 || reads != 0 {
		t.Fatalf("previous identity controlled during restart: starts=%d stops=%d reads=%d", starts, stops, reads)
	}
	if session.ID() != "second" || session.Status() != StatusRunning {
		t.Fatalf("restart did not establish new identity: %q %s", session.ID(), session.Status())
	}
}

func TestInfoDuringStopCannotSupersedeWriteAndConflictCanBePolled(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		t.Run(fmt.Sprint(conflict), func(t *testing.T) {
			stopEntered := make(chan struct{})
			releaseStop := make(chan struct{})
			stopDone := make(chan error, 1)
			readStatus := "RUNNING"
			client := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method == http.MethodGet {
					return v3Reply(200, fmt.Sprintf(`{"agent_session_id":"s","status":%q}`, readStatus), ""), nil
				}
				if strings.HasSuffix(r.URL.Path, "/stop") {
					close(stopEntered)
					<-releaseStop
					if conflict {
						return v3Reply(409, `{"reason":"TaskConflict"}`, ""), nil
					}
					return v3Reply(200, `{}`, ""), nil
				}
				return v3Reply(200, `{"agent_session_id":"s","status":"RUNNING","create_ts":1}`, ""), nil
			})})
			session := (&SavedAgent{client: client, ID: "saved"}).CreateSession(CreateSessionOptions{Channel: "room", AgentUID: "1", RemoteUIDs: []string{"2"}})
			if _, err := session.Start(context.Background()); err != nil {
				t.Fatal(err)
			}
			go func() { stopDone <- session.Stop(context.Background()) }()
			<-stopEntered
			if err := session.Stop(context.Background()); err == nil {
				t.Error("concurrent stop should fail locally")
			}
			if _, err := session.Start(context.Background()); err == nil {
				t.Error("start during pending stop should fail locally")
			}

			if _, err := session.GetInfo(context.Background()); err != nil {
				t.Fatal(err)
			}
			if session.Status() != StatusStopping {
				t.Errorf("read overrode pending stop: %s", session.Status())
			}
			close(releaseStop)
			err := <-stopDone
			if conflict {
				if err == nil || session.Status() != StatusStopping {
					t.Fatalf("conflict state: %v %s", err, session.Status())
				}
			} else if err != nil || session.Status() != StatusStopped {
				t.Fatalf("stop result: %v %s", err, session.Status())
			}
			if !conflict {
				if err := session.Stop(context.Background()); err == nil {
					t.Error("already stopped session issued another stop")
				}
			}
			readStatus = "STOPPED"
			if _, err := session.GetInfo(context.Background()); err != nil {
				t.Fatal(err)
			}
			if session.Status() != StatusStopped {
				t.Fatalf("completed conflict prevented polling: %s", session.Status())
			}
		})
	}
}

func TestLatestIssuedSessionReadWins(t *testing.T) {
	firstReadEntered := make(chan struct{})
	releaseFirstRead := make(chan struct{})
	firstDone := make(chan error, 1)
	reads := 0
	client := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method == http.MethodGet {
			reads++
			if reads == 1 {
				close(firstReadEntered)
				<-releaseFirstRead
				return v3Reply(200, `{"agent_session_id":"s","status":"RUNNING"}`, ""), nil
			}
			return v3Reply(200, `{"agent_session_id":"s","status":"STOPPED"}`, ""), nil
		}
		return v3Reply(200, `{"agent_session_id":"s","status":"RUNNING","create_ts":1}`, ""), nil
	})})
	session := (&SavedAgent{client: client, ID: "saved"}).CreateSession(CreateSessionOptions{Channel: "room", AgentUID: "1", RemoteUIDs: []string{"2"}})
	if _, err := session.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	go func() { _, err := session.GetInfo(context.Background()); firstDone <- err }()
	<-firstReadEntered
	if _, err := session.GetInfo(context.Background()); err != nil {
		t.Fatal(err)
	}
	close(releaseFirstRead)
	if err := <-firstDone; err != nil {
		t.Fatal(err)
	}
	if session.Status() != StatusStopped {
		t.Fatalf("older read overwrote newer read: %s", session.Status())
	}
}

func TestAppCredentialsSignLegacyAndPreviewResourcesAndPreserveExplicitTokens(t *testing.T) {
	calls := 0
	client := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: strings.Repeat("a", 32), AppCertificate: strings.Repeat("b", 32), HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		auth := r.Header.Get("Authorization")
		if calls == 2 {
			if auth != "agora token=per-call-token" {
				t.Fatalf("per-call token overwritten: %q", auth)
			}
		} else if !strings.HasPrefix(auth, "agora token=") || len(auth) < 30 {
			t.Fatalf("app credentials did not sign call %d: %q", calls, auth)
		}
		return v3Reply(200, `{}`, ""), nil
	})})
	request := &Agora.GetAgentsRequest{Appid: strings.Repeat("a", 32), AgentID: "running"}
	if _, err := client.Agents.Get(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Agents.Get(context.Background(), request, option.WithToken("per-call-token")); err != nil {
		t.Fatal(err)
	}
	if _, err := client.V3.Agents.Get(context.Background(), "saved"); err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Fatalf("expected legacy+override+preview calls, got %d", calls)
	}
}

func TestTokenOnlyPublicClientRejectsOtherSchemesAcrossResources(t *testing.T) {
	for _, opts := range [][]option.RequestOption{
		{option.WithAuthorization("")},
		{option.WithHTTPHeader(http.Header{"authorization": []string{""}})},
		{option.WithAuthorization("Bearer obsolete")},
		{option.WithHTTPHeader(http.Header{"authorization": []string{"Digest obsolete"}})},
		{option.WithHTTPHeader(http.Header{"Authorization": []string{"agora token=valid", "Bearer obsolete"}})},
	} {
		calls := 0
		transport := v3HTTPFunc(func(*http.Request) (*http.Response, error) { calls++; return v3Reply(200, `{}`, ""), nil })
		client := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: strings.Repeat("a", 32), AppCertificate: strings.Repeat("b", 32), HTTPClient: transport})
		if _, err := client.Agents.Get(context.Background(), &Agora.GetAgentsRequest{Appid: "app", AgentID: "id"}, opts...); err == nil {
			t.Error("legacy resource accepted non-Agora auth")
		}
		preview := v3.NewClient("app", append(opts, option.WithHTTPClient(transport))...)
		if _, err := preview.Agents.Get(context.Background(), "id"); err == nil {
			t.Error("preview resource accepted non-Agora auth")
		}
		if calls != 0 {
			t.Fatalf("invalid authorization sent %d requests", calls)
		}
	}
}

func TestAppCredentialSignerPreservesCaseVariantExplicitToken(t *testing.T) {
	calls := 0
	signer := &appCredentialsHTTPClient{transport: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if values := r.Header["authorization"]; len(values) != 1 || values[0] != "agora token=explicit" {
			t.Fatalf("explicit token lost: %v", r.Header)
		}
		if _, exists := r.Header["Authorization"]; exists {
			t.Fatalf("signer injected a second token: %v", r.Header)
		}
		return v3Reply(200, `{}`, ""), nil
	})}
	req, _ := http.NewRequest(http.MethodGet, "https://example.test", nil)
	req.Header["authorization"] = []string{"agora token=explicit"}
	if _, err := signer.Do(req); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("explicit token request was not sent: %d", calls)
	}
}
