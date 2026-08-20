package Agora

import (
	"encoding/json"
	"testing"
)

func TestAgentThinkAppendActionsSerialize(t *testing.T) {
	request := AgentThinkAgentManagementRequest{
		Appid:             "app",
		AgentID:           "agent",
		Text:              "queued instruction",
		OnListeningAction: AgentThinkAgentManagementRequestOnListeningActionAppend.Ptr(),
		OnThinkingAction:  AgentThinkAgentManagementRequestOnThinkingActionAppend.Ptr(),
		OnSpeakingAction:  AgentThinkAgentManagementRequestOnSpeakingActionAppend.Ptr(),
	}
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal think request: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode think request: %v", err)
	}
	for _, field := range []string{"on_listening_action", "on_thinking_action", "on_speaking_action"} {
		if decoded[field] != "append" {
			t.Fatalf("%s = %#v, want append", field, decoded[field])
		}
	}
}
