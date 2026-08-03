package vendors

import (
	"encoding/json"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
)

func TestQwenOmniMatchesGeneratedMLLM(t *testing.T) {
	turnDetection := &Agora.MllmTurnDetection{
		Mode: Agora.MllmTurnDetectionModeServerVad.Ptr(),
	}
	config := NewQwenOmni(QwenOmniOptions{
		APIKey:       "dashscope-key",
		Model:        "qwen3-omni-flash-realtime",
		Voice:        "Momo",
		Instructions: "Be concise.",
		Params: map[string]interface{}{
			"model":        "qwen-omni-custom",
			"custom_param": true,
		},
		TurnDetection: turnDetection,
	}).ToConfig()

	if got := config["vendor"]; got != "qwen_omni" {
		t.Fatalf("vendor = %v, want qwen_omni", got)
	}
	if got := config["url"]; got != qwenOmniDefaultURL {
		t.Fatalf("url = %v, want %s", got, qwenOmniDefaultURL)
	}
	params := config["params"].(map[string]interface{})
	if got := params["model"]; got != "qwen-omni-custom" {
		t.Fatalf("model = %v, want qwen-omni-custom", got)
	}
	if got := params["voice"]; got != "Momo" {
		t.Fatalf("voice = %v, want Momo", got)
	}

	payload, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("marshal Qwen Omni config: %v", err)
	}
	var generated Agora.Mllm
	if err := json.Unmarshal(payload, &generated); err != nil {
		t.Fatalf("unmarshal Qwen Omni config: %v", err)
	}
	if generated.Vendor == nil || *generated.Vendor != Agora.MllmVendorQwenOmni {
		t.Fatalf("generated vendor = %v, want qwen_omni", generated.Vendor)
	}
	if generated.TurnDetection == nil {
		t.Fatal("generated turn_detection is nil")
	}
	if generated.Params == nil || generated.Params.GetExtraProperties()["custom_param"] != true {
		t.Fatalf("generated params lost custom_param: %#v", generated.Params)
	}
}

func TestQwenOmniValidation(t *testing.T) {
	turnDetection := &Agora.MllmTurnDetection{
		Mode: Agora.MllmTurnDetectionModeServerVad.Ptr(),
	}
	tests := []struct {
		name      string
		opts      QwenOmniOptions
		wantPanic string
	}{
		{
			name: "API key required",
			opts: QwenOmniOptions{
				Model:         "qwen3-omni-flash-realtime",
				TurnDetection: turnDetection,
			},
			wantPanic: "QwenOmni requires APIKey",
		},
		{
			name: "model required",
			opts: QwenOmniOptions{
				APIKey:        "dashscope-key",
				TurnDetection: turnDetection,
			},
			wantPanic: "QwenOmni requires Model",
		},
		{
			name: "turn detection required",
			opts: QwenOmniOptions{
				APIKey: "dashscope-key",
				Model:  "qwen3-omni-flash-realtime",
			},
			wantPanic: "QwenOmni requires TurnDetection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != tt.wantPanic {
					t.Fatalf("panic = %v, want %q", got, tt.wantPanic)
				}
			}()
			NewQwenOmni(tt.opts)
		})
	}
}
