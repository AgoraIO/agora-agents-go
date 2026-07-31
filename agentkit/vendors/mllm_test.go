package vendors

import (
	"encoding/json"
	"reflect"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
)

func TestOpenAIRealtimeURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "default URL",
			want: "wss://api.openai.com/v1/realtime",
		},
		{
			name: "custom URL",
			url:  "wss://realtime.example.com/v1",
			want: "wss://realtime.example.com/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewOpenAIRealtime(OpenAIRealtimeOptions{
				APIKey: "openai-key",
				URL:    tt.url,
			}).ToConfig()

			if got := config["url"]; got != tt.want {
				t.Errorf("url = %v, want %q", got, tt.want)
			}
		})
	}
}

func TestAzureOpenAIRealtimeMatchesGeneratedMLLM(t *testing.T) {
	maxHistory := 20
	turnDetection := &Agora.MllmTurnDetection{
		Mode: Agora.MllmTurnDetectionModeServerVad.Ptr(),
	}
	const azureRealtimeURL = "wss://example.openai.azure.com/openai/realtime?" +
		"api-version=2025-04-01-preview&deployment=gpt-realtime"
	config := NewAzureOpenAIRealtime(AzureOpenAIRealtimeOptions{
		APIKey:        "azure-key",
		URL:           azureRealtimeURL,
		Model:         "gpt-realtime",
		Voice:         "alloy",
		Instructions:  "Be concise.",
		MaxHistory:    &maxHistory,
		TurnDetection: turnDetection,
	}).ToConfig()

	if got := config["vendor"]; got != "azure" {
		t.Fatalf("vendor = %v, want azure", got)
	}
	if got := config["max_history"]; got != 20 {
		t.Fatalf("max_history = %v, want 20", got)
	}
	params := config["params"].(map[string]interface{})
	if got := params["model"]; got != "gpt-realtime" {
		t.Fatalf("model = %v, want gpt-realtime", got)
	}
	if got := params["voice"]; got != "alloy" {
		t.Fatalf("voice = %v, want alloy", got)
	}

	payload, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("marshal Azure MLLM config: %v", err)
	}
	var generated Agora.Mllm
	if err := json.Unmarshal(payload, &generated); err != nil {
		t.Fatalf("unmarshal Azure MLLM config: %v", err)
	}
	if generated.Vendor == nil || *generated.Vendor != Agora.MllmVendorAzure {
		t.Fatalf("generated vendor = %v, want azure", generated.Vendor)
	}
	if generated.MaxHistory == nil || *generated.MaxHistory != 20 {
		t.Fatalf("generated max_history = %v, want 20", generated.MaxHistory)
	}
	if generated.TurnDetection == nil {
		t.Fatal("generated turn_detection is nil")
	}
	if generated.Params == nil || generated.Params.Model == nil || *generated.Params.Model != "gpt-realtime" {
		t.Fatalf("generated params model = %#v, want gpt-realtime", generated.Params)
	}
	if len(generated.Params.GetExtraProperties()) != 0 {
		t.Fatalf("generated params contain unexpected properties: %#v", generated.Params.GetExtraProperties())
	}
}

func TestAzureOpenAIRealtimeOptionsSurface(t *testing.T) {
	// Azure is intentionally a strict allowlist rather than a generic MLLM passthrough.
	optionType := reflect.TypeOf(AzureOpenAIRealtimeOptions{})
	got := make([]string, optionType.NumField())
	for i := 0; i < optionType.NumField(); i++ {
		got[i] = optionType.Field(i).Name
	}
	want := []string{
		"APIKey",
		"URL",
		"Messages",
		"Instructions",
		"Model",
		"Voice",
		"OutputModalities",
		"MaxHistory",
		"GreetingMessage",
		"TurnDetection",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Azure options fields = %v, want %v", got, want)
	}
}

func TestAzureOpenAIRealtimeValidation(t *testing.T) {
	turnDetection := &Agora.MllmTurnDetection{
		Mode: Agora.MllmTurnDetectionModeServerVad.Ptr(),
	}
	tests := []struct {
		name      string
		opts      AzureOpenAIRealtimeOptions
		wantPanic string
	}{
		{
			name: "API key required",
			opts: AzureOpenAIRealtimeOptions{
				URL:           "wss://azure.example/realtime",
				TurnDetection: turnDetection,
			},
			wantPanic: "AzureOpenAIRealtime requires APIKey",
		},
		{
			name: "URL required",
			opts: AzureOpenAIRealtimeOptions{
				APIKey:        "azure-key",
				TurnDetection: turnDetection,
			},
			wantPanic: "AzureOpenAIRealtime requires URL",
		},
		{
			name: "turn detection required",
			opts: AzureOpenAIRealtimeOptions{
				APIKey: "azure-key",
				URL:    "wss://azure.example/realtime",
			},
			wantPanic: "AzureOpenAIRealtime requires TurnDetection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != tt.wantPanic {
					t.Fatalf("panic = %v, want %q", got, tt.wantPanic)
				}
			}()
			NewAzureOpenAIRealtime(tt.opts)
		})
	}
}
