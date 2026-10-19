package agentkit

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
)

func TestGeminiTTSPreviewLifecycle(t *testing.T) {
	for _, model := range []string{vendors.GeminiTTSModelFlash38} {
		t.Run(model, func(t *testing.T) {
			rec := &recordingClient{}
			client := newTestPreviewClient(rec)
			tts := vendors.NewGeminiTTS(vendors.GeminiTTSOptions{APIKey: "test-key", Model: model, Voice: "Puck", Style: "warm and reassuring"})
			expected := map[string]interface{}{"vendor": "gemini", "params": map[string]interface{}{
				"api_key": "test-key", "model": model, "voice": "Puck", "style": "warm and reassuring"}}
			if !reflect.DeepEqual(tts.ToConfig(), expected) {
				t.Fatalf("config = %v", tts.ToConfig())
			}
			session := NewAgent(client).
				WithStt(vendors.NewGeminiSTT(vendors.GeminiSTTOptions{APIKey: "test-key"})).
				WithLlm(vendors.NewGemini(vendors.GeminiOptions{APIKey: "test-key", Model: "gemini-3.6-flash"})).
				WithTts(tts).CreateSession(CreateSessionOptions{Channel: "test", AgentUID: "1", RemoteUIDs: []string{"100"}})
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
			if _, err := session.GetHistory(ctx); err != nil {
				t.Fatal(err)
			}
			if err := session.Stop(ctx); err != nil {
				t.Fatal(err)
			}
			if len(rec.requests) != 5 {
				t.Fatalf("requests = %d", len(rec.requests))
			}
			for _, req := range rec.requests {
				if !strings.HasPrefix(req.URL.String(), PreviewAPIBaseURL) || req.Header.Get(PreviewFeatureHeader) != "gemini-live" {
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
		})
	}
}

func TestGeminiTTSDefaultsAndDetection(t *testing.T) {
	config := vendors.NewGeminiTTS(vendors.GeminiTTSOptions{APIKey: "test-key"}).ToConfig()
	assertJSONEqual(t, config, `{"vendor":"gemini","params":{"api_key":"test-key","model":"gemini-3.8-flash-tts","voice":"Puck"}}`)
	raw := map[string]interface{}{"tts": map[string]interface{}{"vendor": "gemini"}}
	if got := RequiredPreviewFeatures(raw); !reflect.DeepEqual(got, []string{"gemini-live"}) {
		t.Fatalf("features = %v", got)
	}
	raw["mllm"] = map[string]interface{}{"vendor": "openai_gpt_live"}
	if got := RequiredPreviewFeatures(raw); !reflect.DeepEqual(got, []string{"gemini-live"}) {
		t.Fatalf("features = %v", got)
	}
	for _, options := range []vendors.GeminiTTSOptions{{}, {APIKey: " "}, {APIKey: "test", Model: " "}, {APIKey: "test", Voice: " "}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("expected invalid options to panic")
				}
			}()
			vendors.NewGeminiTTS(options)
		}()
	}
}
