package core

import (
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
)

func TestLlmToolsSurviveAgentPropertiesBuildWithToolGate(t *testing.T) {
	tool := &Agora.LlmTool{
		Function: &Agora.LlmToolFunction{Name: "lookup"},
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
}
