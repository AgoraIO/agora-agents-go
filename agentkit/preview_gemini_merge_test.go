package agentkit_test

import (
	"reflect"
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
)

func TestPreviewGeminiAndGPTLiveGates(t *testing.T) {
	gemini := vendors.NewGeminiLive(vendors.GeminiLiveOptions{APIKey: "test-key"})
	config := gemini.ToConfig()
	geminiProperties := map[string]interface{}{"mllm": config}
	if got := agentkit.RequiredPreviewFeatures(geminiProperties); !reflect.DeepEqual(got, []string{agentkit.PreviewFeatureGeminiLive}) {
		t.Fatalf("Gemini gate = %v", got)
	}
	config["greeting_message"] = "Hello"
	agentkit.ApplyPreviewShape(geminiProperties)
	if config["greeting"] != "Hello" || config["greeting_message"] != nil {
		t.Fatalf("Gemini greeting shape = %v", config)
	}

	gpt := vendors.NewOpenAIGPTLive(vendors.OpenAIGPTLiveOptions{APIKey: "test-key"})
	if got := agentkit.RequiredPreviewFeatures(map[string]interface{}{"mllm": gpt.ToConfig()}); !reflect.DeepEqual(got, []string{agentkit.PreviewFeatureLiveModels}) {
		t.Fatalf("GPT Live gate = %v", got)
	}
}
