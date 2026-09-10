package core

import (
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
)

func TestLlmToolsSurviveAgentPropertiesBuildWithToolGate(t *testing.T) {
	tool := &Agora.LlmTool{
		Function:  &Agora.LlmToolFunction{Name: "lookup"},
		Execution: &Agora.LlmToolExecution{Mode: Agora.LlmToolExecutionModeSync.Ptr()},
		Server: &Agora.LlmToolServer{
			Method: Agora.LlmToolServerMethodGet,
			URL:    "https://example.com/items/{{args.id}}",
		},
	}
	base := NewBaseAgent(WithTools(true))
	base.LLM = map[string]interface{}{"tools": []*Agora.LlmTool{tool}}

	props, err := BuildPropertiesMap(ProfileGlobal, base, ToPropertiesOptions{
		Channel:              "tools-channel",
		AgentUID:             "1001",
		RemoteUIDs:           []string{"1002"},
		Token:                "token",
		SkipVendorValidation: true,
	}, nil)
	if err != nil {
		t.Fatalf("BuildPropertiesMap returned error: %v", err)
	}

	advanced := props["advanced_features"].(map[string]interface{})
	if advanced["enable_tools"] != true {
		t.Fatalf("enable_tools = %#v, want true", advanced["enable_tools"])
	}
	llm := props["llm"].(map[string]interface{})
	if _, ok := llm["tools"].([]*Agora.LlmTool); !ok {
		t.Fatalf("llm.tools = %#v, want generated tool definitions", llm["tools"])
	}
	if tool.Execution.Mode == nil || *tool.Execution.Mode != Agora.LlmToolExecutionModeSync {
		t.Fatalf("execution mode = %#v, want sync", tool.Execution.Mode)
	}
}

func TestMLLMKeepsMCPOutsideVendorParams(t *testing.T) {
	for _, vendor := range []string{"openai_gpt_live", "openai"} {
		t.Run(vendor, func(t *testing.T) {
			base := NewBaseAgent(WithTools(true))
			servers := []map[string]interface{}{{"name": "lookup", "endpoint": "https://tools.example/mcp", "transport": "streamable_http"}}
			base.MLLM = map[string]interface{}{"enable": true, "vendor": vendor, "mcp_servers": servers, "params": map[string]interface{}{"tool_enabled": true}}
			props, err := BuildPropertiesMap(ProfileGlobal, base, ToPropertiesOptions{Channel: "test", AgentUID: "1", RemoteUIDs: []string{"2"}, Token: "token"}, nil)
			if err != nil {
				t.Fatal(err)
			}
			_, exists := props["llm"]
			if vendor == "openai" {
				if exists {
					t.Fatal("changed public Realtime behavior")
				}
				return
			}
			if exists {
				t.Fatalf("unexpected llm = %#v", props["llm"])
			}
			if props["advanced_features"].(map[string]interface{})["enable_tools"] != true {
				t.Fatal("lost tool gate")
			}
			if props["mllm"].(map[string]interface{})["mcp_servers"] == nil {
				t.Fatal("MCP missing from mllm")
			}
		})
	}
}

func TestGPTLiveWarnsAndDropsAgentLevelTurnDetection(t *testing.T) {
	base := NewBaseAgent()
	base.MLLM = map[string]interface{}{"enable": true, "vendor": "openai_gpt_live"}
	base.TurnDetection = &TurnDetectionConfig{}
	var warning string
	props, err := BuildPropertiesMap(ProfileGlobal, base, ToPropertiesOptions{
		Channel: "test", AgentUID: "1", RemoteUIDs: []string{"2"}, Token: "token",
		Warn: func(message string) { warning = message },
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := props["turn_detection"]; exists {
		t.Fatal("retained unsupported agent-level turn_detection")
	}
	if warning == "" {
		t.Fatal("did not warn about unsupported agent-level turn_detection")
	}
}
