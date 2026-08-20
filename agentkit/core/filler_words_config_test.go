package core

import (
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
)

func TestGeneratedFillerWordsConfigSupportsGeneratedMode(t *testing.T) {
	mode := Agora.StartAgentsRequestPropertiesFillerWordsContentModeGenerated
	config := &Agora.StartAgentsRequestPropertiesFillerWords{
		Enable: Agora.Bool(true),
		Content: &Agora.StartAgentsRequestPropertiesFillerWordsContent{
			Mode: &mode,
			StaticConfig: &Agora.StartAgentsRequestPropertiesFillerWordsContentStaticConfig{
				Phrases: []string{"one moment"},
			},
			GeneratedConfig: &Agora.StartAgentsRequestPropertiesFillerWordsContentGeneratedConfig{
				LlmProvider: &Agora.StartAgentsRequestPropertiesFillerWordsContentGeneratedConfigLlmProvider{
					BaseURL: "https://llm.example.com",
					APIKey:  "key",
				},
				Prompt: "Generate a short filler phrase",
			},
		},
	}
	serialized, err := StructToMap(config)
	if err != nil {
		t.Fatalf("serialize generated filler words: %v", err)
	}
	content := serialized["content"].(map[string]interface{})
	if content["mode"] != "generated" || content["generated_config"] == nil {
		t.Fatalf("serialized filler words = %#v", serialized)
	}
}
