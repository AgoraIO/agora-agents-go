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

func TestMLLMVendorsSupportToolsAndMCPServers(t *testing.T) {
	tool := &Agora.LlmTool{
		Function: &Agora.LlmToolFunction{Name: "lookup"},
		Server: &Agora.LlmToolServer{
			Method: Agora.LlmToolServerMethodPost,
			URL:    "https://tools.example.com/lookup",
		},
	}
	server := &Agora.McpServer{
		Name:     "catalog",
		Endpoint: "https://mcp.example.com",
	}
	turnDetection := &Agora.MllmTurnDetection{
		Mode: Agora.MllmTurnDetectionModeServerVad.Ptr(),
	}
	tests := []struct {
		name   string
		config func() map[string]interface{}
	}{
		{
			name: "OpenAI Realtime",
			config: func() map[string]interface{} {
				return NewOpenAIRealtime(OpenAIRealtimeOptions{
					APIKey:     "key",
					Tools:      []*Agora.LlmTool{tool},
					McpServers: []*Agora.McpServer{server},
				}).ToConfig()
			},
		},
		{
			name: "OpenAI GPT Live",
			config: func() map[string]interface{} {
				return NewOpenAIGPTLive(OpenAIGPTLiveOptions{
					APIKey:           "key",
					Tools:            []*Agora.LlmTool{tool},
					McpServerConfigs: []*Agora.McpServer{server},
				}).ToConfig()
			},
		},
		{
			name: "Azure OpenAI Realtime",
			config: func() map[string]interface{} {
				return NewAzureOpenAIRealtime(AzureOpenAIRealtimeOptions{
					APIKey:        "key",
					URL:           "wss://azure.example.com/realtime",
					TurnDetection: turnDetection,
					Tools:         []*Agora.LlmTool{tool},
					McpServers:    []*Agora.McpServer{server},
				}).ToConfig()
			},
		},
		{
			name: "xAI Grok",
			config: func() map[string]interface{} {
				return NewXaiGrok(XaiGrokOptions{
					APIKey:     "key",
					Tools:      []*Agora.LlmTool{tool},
					McpServers: []*Agora.McpServer{server},
				}).ToConfig()
			},
		},
		{
			name: "Gemini Live",
			config: func() map[string]interface{} {
				return NewGeminiLive(GeminiLiveOptions{
					APIKey:     "key",
					Tools:      []*Agora.LlmTool{tool},
					McpServers: []*Agora.McpServer{server},
				}).ToConfig()
			},
		},
		{
			name: "Gemini Live legacy model",
			config: func() map[string]interface{} {
				return NewGeminiLive(GeminiLiveOptions{
					APIKey:     "key",
					Model:      "gemini-live-2.5-flash",
					Tools:      []*Agora.LlmTool{tool},
					McpServers: []*Agora.McpServer{server},
				}).ToConfig()
			},
		},
		{
			name: "Vertex AI",
			config: func() map[string]interface{} {
				return NewVertexAI(VertexAIOptions{
					ProjectID:           "project",
					ADCredentialsString: "credentials",
					Tools:               []*Agora.LlmTool{tool},
					McpServers:          []*Agora.McpServer{server},
				}).ToConfig()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config()
			if got := config["tools"]; !reflect.DeepEqual(got, []*Agora.LlmTool{tool}) {
				t.Fatalf("tools = %#v, want configured REST tools", got)
			}
			if got := config["mcp_servers"]; !reflect.DeepEqual(got, []*Agora.McpServer{server}) {
				t.Fatalf("mcp_servers = %#v, want configured MCP servers", got)
			}

			payload, err := json.Marshal(config)
			if err != nil {
				t.Fatalf("marshal MLLM config: %v", err)
			}
			var generated Agora.Mllm
			if err := json.Unmarshal(payload, &generated); err != nil {
				t.Fatalf("unmarshal generated MLLM: %v", err)
			}
			if len(generated.Tools) != 1 || len(generated.McpServers) != 1 {
				t.Fatalf("generated tools/mcp_servers = %#v/%#v, want one of each", generated.Tools, generated.McpServers)
			}
		})
	}
}

func TestOpenAIGPTLiveWireShape(t *testing.T) {
	servers := []map[string]interface{}{{"name": "lookup", "endpoint": "https://tools.example/mcp"}}
	config := NewOpenAIGPTLive(OpenAIGPTLiveOptions{
		APIKey:          "openai-key",
		GreetingMessage: "Hello from GPT Live",
		McpServers:      servers,
	}).ToConfig()

	want := map[string]interface{}{
		"vendor":           "openai_gpt_live",
		"api_key":          "openai-key",
		"url":              "wss://api.openai.com/v1/live/sessions",
		"greeting_message": "Hello from GPT Live",
		"params": map[string]interface{}{
			"model": "gpt-live-1",
		},
		"mcp_servers": []map[string]interface{}{{"name": "lookup", "endpoint": "https://tools.example/mcp", "transport": "streamable_http"}},
	}
	if !reflect.DeepEqual(config, want) {
		t.Fatalf("gpt live config = %#v, want %#v", config, want)
	}
	if _, ok := servers[0]["transport"]; ok {
		t.Fatal("mutated MCP input")
	}
}

func TestOpenAIGPTLiveSupportsToolsAndTypedMCP(t *testing.T) {
	tool := &Agora.LlmTool{
		Function: &Agora.LlmToolFunction{Name: "lookup"},
		Server: &Agora.LlmToolServer{
			Method: Agora.LlmToolServerMethodPost,
			URL:    "https://tools.example.com/lookup",
		},
	}
	server := &Agora.McpServer{
		Name:     "catalog",
		Endpoint: "https://mcp.example.com",
	}
	config := NewOpenAIGPTLive(OpenAIGPTLiveOptions{
		APIKey:           "openai-key",
		Tools:            []*Agora.LlmTool{tool},
		McpServers:       []map[string]interface{}{{"name": "legacy"}},
		McpServerConfigs: []*Agora.McpServer{server},
	}).ToConfig()

	tools, ok := config["tools"].([]*Agora.LlmTool)
	if !ok || len(tools) != 1 || tools[0] != tool {
		t.Fatalf("tools = %#v, want typed REST tool", config["tools"])
	}
	servers, ok := config["mcp_servers"].([]*Agora.McpServer)
	if !ok || len(servers) != 1 || servers[0] != server {
		t.Fatalf("mcp_servers = %#v, want typed MCP server", config["mcp_servers"])
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
		"Tools",
		"McpServers",
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
