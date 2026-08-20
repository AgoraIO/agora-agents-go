package core

import "testing"

func TestGPTLiveAgentGreetingUsesGreetingMessageField(t *testing.T) {
	base := NewBaseAgent(WithGreeting("hello from agent"))
	base.MLLM = map[string]interface{}{"vendor": "openai_gpt_live"}
	config := buildMllmConfigMap(base)
	if config["greeting_message"] != "hello from agent" {
		t.Fatalf("greeting_message = %#v, want agent greeting", config["greeting_message"])
	}
	if _, exists := config["greeting"]; exists {
		t.Fatalf("GPT Live must not receive greeting: %#v", config)
	}
}
