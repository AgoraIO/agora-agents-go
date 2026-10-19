package agentkit_test

import (
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
)

func TestGeminiAndGPTLiveStayOnProduction(t *testing.T) {
	gemini := vendors.NewGeminiLive(vendors.GeminiLiveOptions{APIKey: "test-key"})
	config := gemini.ToConfig()
	geminiProperties := map[string]interface{}{"mllm": config}
	if got := agentkit.RequiredPreviewFeatures(geminiProperties); len(got) != 0 {
		t.Fatalf("Gemini 3.8 production gate = %v", got)
	}
	config["greeting_message"] = "Hello"
	agentkit.ApplyPreviewShape(geminiProperties)
	if config["greeting_message"] != "Hello" || config["greeting"] != nil {
		t.Fatalf("Gemini production greeting shape = %v", config)
	}
	legacy := vendors.NewGeminiLive(vendors.GeminiLiveOptions{
		APIKey: "test-key", Model: "future-live-model", URL: vendors.GeminiLivePreviewURL,
	}).ToConfig()
	legacyProperties := map[string]interface{}{"mllm": legacy}
	if got := agentkit.RequiredPreviewFeatures(legacyProperties); len(got) != 0 {
		t.Fatalf("legacy Gemini production gate = %v", got)
	}
	legacy["greeting_message"] = "Hello"
	agentkit.ApplyPreviewShape(legacyProperties)
	if legacy["greeting_message"] != "Hello" || legacy["greeting"] != nil {
		t.Fatalf("legacy Gemini production greeting = %v", legacy)
	}

	gpt := vendors.NewOpenAIGPTLive(vendors.OpenAIGPTLiveOptions{APIKey: "test-key"})
	if got := agentkit.RequiredPreviewFeatures(map[string]interface{}{"mllm": gpt.ToConfig()}); len(got) != 0 {
		t.Fatalf("GPT Live should remain on production routing, got gate = %v", got)
	}
}
