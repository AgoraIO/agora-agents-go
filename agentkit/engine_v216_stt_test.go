package agentkit

import (
	"encoding/json"
	"reflect"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	cnvendors "github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/core"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
)

func TestEngineV216STTStartRequests(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		newVendor func(string, map[string]any) core.STTVendor
	}{
		{
			name: "elevenlabs",
			newVendor: func(key string, params map[string]any) core.STTVendor {
				return vendors.NewElevenLabsSTT(vendors.ElevenLabsSTTOptions{APIKey: key, AdditionalParams: params})
			},
		},
		{
			name: "fishaudio",
			newVendor: func(key string, params map[string]any) core.STTVendor {
				return vendors.NewFishAudioSTT(vendors.FishAudioSTTOptions{APIKey: key, AdditionalParams: params})
			},
		},
		{
			name: "qwen",
			newVendor: func(key string, params map[string]any) core.STTVendor {
				return cnvendors.NewQwenSTT(cnvendors.QwenSTTOptions{APIKey: key, AdditionalParams: params})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			additional := map[string]any{"api_key": "ignored", "enabled": false, "keywords": []string{}}
			stt := tt.newVendor("asr-key", additional)
			agent := NewAgent(testAgoraClient()).WithStt(stt).WithLlm(stubLLM).WithTts(stubTTS).
				WithTurnDetection(&TurnDetectionConfig{Language: Agora.AsrLanguageZhCn.Ptr()})

			// Verify the generated ASR union and the actual /join body both retain the new variant.
			props, err := agent.ToProperties(basePropertiesOpts())
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(props.Asr)
			if err != nil {
				t.Fatal(err)
			}
			generated := map[string]any{}
			if err := json.Unmarshal(data, &generated); err != nil {
				t.Fatal(err)
			}
			expected := map[string]any{
				"vendor": tt.name, "language": "zh-CN",
				"params": map[string]any{"api_key": "asr-key", "enabled": false, "keywords": []any{}},
			}
			if !reflect.DeepEqual(generated, expected) {
				t.Fatalf("generated asr = %#v, want %#v", generated, expected)
			}
			payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
			if err != nil {
				t.Fatal(err)
			}
			properties, ok := payload["properties"].(map[string]any)
			if !ok || !reflect.DeepEqual(properties["asr"], expected) {
				t.Fatalf("join properties = %#v, want asr %#v", payload["properties"], expected)
			}
			if additional["api_key"] != "ignored" {
				t.Fatal("serialization mutated additional params")
			}

			t.Run("missing api key", func(t *testing.T) {
				defer func() {
					if recover() == nil {
						t.Fatal("expected missing api key to panic")
					}
				}()
				tt.newVendor("", map[string]any{"api_key": "cannot-bypass-validation"})
			})
		})
	}
}
