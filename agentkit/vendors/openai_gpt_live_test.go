package vendors

import (
	"reflect"
	"testing"
)

func TestGPTLiveV3Options(t *testing.T) {
	zero, inputIdle, rate, negative, disabled := 0, 1500, 24000, -1, false
	original := map[string]interface{}{"model": "other", "prompt": "other", "output_idle_end_ms": 900}
	config := NewOpenAIGPTLive(OpenAIGPTLiveOptions{
		APIKey: "test", Model: "caller-supplied-model", Voice: "cedar",
		Instructions: "alias", Prompt: "Be brief", OutputIdleEndMs: &zero,
		InputIdleEndMs: &inputIdle, OutputSilencePeak: &zero, OutputSampleRate: &rate,
		OutputBufferMs: &negative, InputBatchMs: &zero, ToolEnabled: &disabled,
		Delegation: "client", ResponsesModel: "delegate",
		AlphaSelector:       "custom=v4",
		InterruptOnUserTurn: &disabled, Headers: `{"X-Test":"yes"}`,
		SessionParams: map[string]interface{}{"context_management": map[string]interface{}{"type": "compaction"}}, Params: original,
	}).ToConfig()
	want := map[string]interface{}{
		"model": "caller-supplied-model", "voice": "cedar", "prompt": "Be brief",
		"alpha_selector": "custom=v4", "output_idle_end_ms": 0,
		"input_idle_end_ms": 1500, "output_silence_peak": 0, "output_sample_rate": 24000,
		"output_buffer_ms": -1, "input_batch_ms": 0, "tool_enabled": false,
		"delegation": "client", "responses_model": "delegate",
		"interrupt_on_user_turn": false, "headers": `{"X-Test":"yes"}`,
		"session_params": map[string]interface{}{"context_management": map[string]interface{}{"type": "compaction"}},
	}
	if !reflect.DeepEqual(config["params"], want) {
		t.Fatalf("params = %#v, want %#v", config["params"], want)
	}
	if original["output_idle_end_ms"] != 900 || original["prompt"] != "other" {
		t.Fatal("mutated input")
	}
}

func TestGPTLiveEndpoints(t *testing.T) {
	for _, tt := range []struct {
		opts OpenAIGPTLiveOptions
		want string
	}{
		{OpenAIGPTLiveOptions{}, "wss://api.openai.com/v1/live/sessions"},
		{OpenAIGPTLiveOptions{BaseURL: "wss://proxy.test/", Path: "custom"}, "wss://proxy.test/custom"},
		{OpenAIGPTLiveOptions{URL: "wss://proxy.test/v1/live?x=1", BaseURL: "wss://unused.test"}, "wss://proxy.test/v1/live?x=1"},
		{OpenAIGPTLiveOptions{URL: "wss://api.openai.com/v1/live?x=1"}, "wss://api.openai.com/v1/live/sessions?x=1"},
	} {
		tt.opts.APIKey = "test"
		if got := NewOpenAIGPTLive(tt.opts).ToConfig()["url"]; got != tt.want {
			t.Fatalf("url = %v, want %s", got, tt.want)
		}
	}
}

func TestGPTLiveProtectedSessionFields(t *testing.T) {
	for _, field := range []string{"model", "delegation", "audio", "instructions", "input"} {
		t.Run(field, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected rejection")
				}
			}()
			NewOpenAIGPTLive(OpenAIGPTLiveOptions{APIKey: "test", Params: map[string]interface{}{"session_params": map[string]interface{}{field: "override"}}}).ToConfig()
		})
	}
}

func TestGPTLiveLegacyAndPendingFields(t *testing.T) {
	config := NewOpenAIGPTLive(OpenAIGPTLiveOptions{APIKey: "test", Instructions: "Be brief", Params: map[string]interface{}{
		"turn_detection": map[string]interface{}{}, "voice": map[string]interface{}{"id": "voice_123"},
	}}).ToConfig()
	params := config["params"].(map[string]interface{})
	if params["prompt"] != "Be brief" {
		t.Fatal("lost instructions")
	}
	if _, ok := params["turn_detection"]; ok {
		t.Fatal("unsupported turn detection retained")
	}
	if _, ok := params["instructions"]; ok {
		t.Fatal("legacy field retained")
	}
	if !reflect.DeepEqual(params["voice"], map[string]interface{}{"id": "voice_123"}) {
		t.Fatal("lost custom voice")
	}
}

func TestGPTLiveRejectsInvalidHeadersAndEndpoints(t *testing.T) {
	for _, opts := range []OpenAIGPTLiveOptions{
		{APIKey: "test", URL: "https://api.openai.com/v1/live/sessions"},
		{APIKey: "test", URL: "/v1/live/sessions"},
		{APIKey: "test", Headers: "not-json"},
		{APIKey: "test", Headers: "[]"},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("expected rejection")
				}
			}()
			NewOpenAIGPTLive(opts).ToConfig()
		}()
	}
}
