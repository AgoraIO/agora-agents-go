package agentkit

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/AgoraIO/agora-agents-go/v2/core"
	"github.com/AgoraIO/agora-agents-go/v2/option"
)

func TestGeminiTTSProductionLifecycle(t *testing.T) {
	t.Setenv("AGORA_AGENTS_API_BASE_URL", "")
	tests := []struct {
		name string
		area option.Area
		raw  bool
	}{
		{name: "US vendor", area: option.AreaUS},
		{name: "EU vendor", area: option.AreaEU},
		{name: "AP vendor", area: option.AreaAP},
		{name: "US raw config", area: option.AreaUS, raw: true},
		{name: "EU raw config", area: option.AreaEU, raw: true},
		{name: "AP raw config", area: option.AreaAP, raw: true},
	}
	for _, tt := range tests {
		for _, model := range []string{vendors.GeminiTTSModelFlash38, "future-tts-model"} {
			t.Run(tt.name+"/"+model, func(t *testing.T) {
				rec := &recordingClient{}
				client := NewAgoraClient(AgoraClientOptions{
					Area: tt.area, AppID: "81190c52971d4004b7244bdcd93e2f34",
					AppCertificate: "0123456789abcdef0123456789abcdef", HTTPClient: rec,
				})
				pool, err := core.NewPool(tt.area)
				if err != nil {
					t.Fatal(err)
				}
				tts := vendors.NewGeminiTTS(vendors.GeminiTTSOptions{
					APIKey: "test-key", Model: model, Voice: "Puck", Style: "warm and reassuring",
				})
				expected := map[string]interface{}{"vendor": "gemini", "params": map[string]interface{}{
					"api_key": "test-key", "model": model, "voice": "Puck", "style": "warm and reassuring"}}
				if !reflect.DeepEqual(tts.ToConfig(), expected) {
					t.Fatalf("config = %v", tts.ToConfig())
				}
				agent := NewAgent(client).
					WithStt(vendors.NewGeminiSTT(vendors.GeminiSTTOptions{APIKey: "test-key"})).
					WithLlm(vendors.NewGemini(vendors.GeminiOptions{APIKey: "test-key", Model: "gemini-3.6-flash"})).
					WithTts(tts)
				if tt.raw {
					agent = agent.WithTts(rawTTSConfig(expected))
				}
				session := agent.CreateSession(CreateSessionOptions{
					Channel: "test", AgentUID: "1", RemoteUIDs: []string{"100"},
				})
				ctx := context.Background()
				if _, err := session.Start(ctx); err != nil {
					t.Fatal(err)
				}
				if err := session.Say(ctx, "hello", nil, nil); err != nil {
					t.Fatal(err)
				}
				if err := session.Interrupt(ctx); err != nil {
					t.Fatal(err)
				}
				if _, err := session.ThinkWithOptions(ctx, "think", nil); err != nil {
					t.Fatal(err)
				}
				if err := session.Update(ctx, &Agora.UpdateAgentsRequestProperties{}); err != nil {
					t.Fatal(err)
				}
				if _, err := session.GetHistory(ctx); err != nil {
					t.Fatal(err)
				}
				if _, err := session.GetInfo(ctx); err != nil {
					t.Fatal(err)
				}
				if _, err := session.GetTurns(ctx); err != nil {
					t.Fatal(err)
				}
				if err := session.Stop(ctx); err != nil {
					t.Fatal(err)
				}
				if len(rec.requests) != 9 {
					t.Fatalf("requests = %d", len(rec.requests))
				}
				for _, req := range rec.requests {
					if !strings.HasPrefix(req.URL.String(), pool.GetCurrentURL()+"/") || req.Header.Get(PreviewFeatureHeader) != "" {
						t.Fatalf("wrong route: %s %v", req.URL, req.Header)
					}
				}
				var body map[string]interface{}
				if err := json.Unmarshal(rec.bodies[0], &body); err != nil {
					t.Fatal(err)
				}
				if got := body["properties"].(map[string]interface{})["tts"]; !reflect.DeepEqual(got, expected) {
					t.Fatalf("wire TTS = %v", got)
				}
				if err := client.StopAgent(ctx, "agent-1"); err != nil {
					t.Fatal(err)
				}
				if req := rec.requests[9]; !strings.HasPrefix(req.URL.String(), pool.GetCurrentURL()+"/") ||
					req.Header.Get(PreviewFeatureHeader) != "" {
					t.Fatalf("shared client route = %s %v", req.URL, req.Header)
				}
			})
		}
	}
}

func TestGeminiTTSDefaultsAndDetection(t *testing.T) {
	tts := vendors.NewGeminiTTS(vendors.GeminiTTSOptions{APIKey: "test-key"})
	config := tts.ToConfig()
	assertJSONEqual(t, config, `{
		"vendor":"gemini","params":{"api_key":"test-key","model":"gemini-3.8-flash-tts","voice":"Puck"}
	}`)
	if tts.GetSampleRate() != nil {
		t.Fatal("Gemini TTS must preserve the preview-era unknown sample rate")
	}
	raw := map[string]interface{}{"tts": map[string]interface{}{"vendor": "gemini"}}
	if got := RequiredPreviewFeatures(raw); len(got) != 0 {
		t.Fatalf("features = %v", got)
	}
	raw["mllm"] = map[string]interface{}{"vendor": "openai_gpt_live"}
	if got := RequiredPreviewFeatures(raw); len(got) != 0 {
		t.Fatalf("features = %v", got)
	}
	tests := []struct {
		name    string
		options vendors.GeminiTTSOptions
	}{
		{name: "missing API key"},
		{name: "blank API key", options: vendors.GeminiTTSOptions{APIKey: " "}},
		{name: "blank model", options: vendors.GeminiTTSOptions{APIKey: "test", Model: " "}},
		{name: "blank voice", options: vendors.GeminiTTSOptions{APIKey: "test", Voice: " "}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("expected invalid options to panic")
				}
			}()
			vendors.NewGeminiTTS(tt.options)
		})
	}
}

func TestGeminiTTSAdditionalParams(t *testing.T) {
	additional := map[string]interface{}{
		"temperature": 0.7,
		"style":       "overridden",
		"voice":       "overridden",
	}
	config := vendors.NewGeminiTTS(vendors.GeminiTTSOptions{
		APIKey:           "test-key",
		Style:            "warm",
		AdditionalParams: additional,
		SkipPatterns:     []int{1},
	}).ToConfig()
	params := config["params"].(map[string]interface{})
	if params["temperature"] != 0.7 || params["style"] != "warm" || params["voice"] != "Puck" {
		t.Fatalf("params = %#v", params)
	}
	if !reflect.DeepEqual(config["skip_patterns"], []int{1}) {
		t.Fatalf("skip_patterns = %#v", config["skip_patterns"])
	}
	if additional["style"] != "overridden" || additional["voice"] != "overridden" {
		t.Fatalf("mutated additional params = %#v", additional)
	}
	payload, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	var generated Agora.Tts
	if err := json.Unmarshal(payload, &generated); err != nil {
		t.Fatal(err)
	}
	if generated.Gemini == nil || generated.Gemini.Params == nil {
		t.Fatalf("generated Gemini TTS = %#v", generated)
	}
	style := generated.Gemini.Params.GetStyle()
	if style == nil || *style != "warm" || generated.Gemini.Params.GetExtraProperties()["temperature"] != 0.7 {
		t.Fatalf("generated params = %#v", generated.Gemini.Params)
	}
	if !reflect.DeepEqual(generated.Gemini.GetSkipPatterns(), []int{1}) {
		t.Fatalf("generated skip_patterns = %#v", generated.Gemini.GetSkipPatterns())
	}
}
