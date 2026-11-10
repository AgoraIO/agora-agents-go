package agentkit

import (
	"reflect"
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
)

// MLLM mode is enabled with advanced_features.enable_mllm only. mllm.enable in
// the /join body, alone or next to enable_mllm, stops the agent from sending
// RTM message.metrics (TTFB) events, so the SDK must never send it.

func assertMLLMEnabledViaAdvancedFeatures(t *testing.T, properties map[string]interface{}, wantAdvanced map[string]interface{}) {
	t.Helper()
	mllm, ok := properties["mllm"].(map[string]interface{})
	if !ok {
		t.Fatalf("mllm missing: %#v", properties)
	}
	if _, exists := mllm["enable"]; exists {
		t.Fatalf("mllm.enable = %#v, want absent", mllm["enable"])
	}
	got, _ := properties["advanced_features"].(map[string]interface{})
	if !reflect.DeepEqual(got, wantAdvanced) {
		t.Fatalf("advanced_features = %#v, want %#v", got, wantAdvanced)
	}
}

func TestMLLMEnabledViaAdvancedFeaturesOnWire(t *testing.T) {
	for _, model := range []string{vendors.GeminiLiveModel38Live, vendors.GeminiLiveModel38LiveExtendedThinking} {
		t.Run("gemini/"+model, func(t *testing.T) {
			agent := NewAgent(testAgoraClient()).WithMllm(vendors.NewGeminiLive(vendors.GeminiLiveOptions{APIKey: "gl-key", Model: model}))
			payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
			if err != nil {
				t.Fatal(err)
			}
			assertMLLMEnabledViaAdvancedFeatures(t, payload["properties"].(map[string]interface{}), map[string]interface{}{"enable_mllm": true})
		})
	}

	t.Run("openai-realtime", func(t *testing.T) {
		agent := NewAgent(testAgoraClient()).WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
			APIKey: "rt-key", Model: "gpt-4o-realtime-preview", Voice: "coral",
		}))
		payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
		if err != nil {
			t.Fatal(err)
		}
		assertMLLMEnabledViaAdvancedFeatures(t, payload["properties"].(map[string]interface{}), map[string]interface{}{"enable_mllm": true})
	})

	t.Run("openai-gpt-live", func(t *testing.T) {
		agent := NewAgent(testAgoraClient()).WithMllm(vendors.NewOpenAIGPTLive(vendors.OpenAIGPTLiveOptions{APIKey: "live-key"}))
		props, err := agent.ToPropertiesMap(basePropertiesOpts())
		if err != nil {
			t.Fatal(err)
		}
		assertMLLMEnabledViaAdvancedFeatures(t, props, map[string]interface{}{"enable_mllm": true})
	})
}

func TestMLLMEnableFlagMergesWithCallerAdvancedFeatures(t *testing.T) {
	enableRTM := true
	enableSAL := false
	want := map[string]interface{}{"enable_mllm": true, "enable_rtm": true, "enable_sal": false}
	realtime := vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{APIKey: "rt-key", Model: "gpt-4o-realtime-preview"})

	orders := map[string]*Agent{
		"advanced features before mllm": NewAgent(testAgoraClient()).
			WithAdvancedFeatures(&AdvancedFeatures{EnableRtm: &enableRTM, EnableSal: &enableSAL}).
			WithMllm(realtime),
		"advanced features after mllm": NewAgent(testAgoraClient()).
			WithMllm(realtime).
			WithAdvancedFeatures(&AdvancedFeatures{EnableRtm: &enableRTM, EnableSal: &enableSAL}),
		"agent option": NewAgent(testAgoraClient(),
			WithAdvancedFeatures(&AdvancedFeatures{EnableRtm: &enableRTM, EnableSal: &enableSAL})).
			WithMllm(realtime),
	}
	for name, agent := range orders {
		t.Run(name, func(t *testing.T) {
			props, err := agent.ToPropertiesMap(basePropertiesOpts())
			if err != nil {
				t.Fatal(err)
			}
			assertMLLMEnabledViaAdvancedFeatures(t, props, want)
			// enable_rtm still selects the RTM data channel.
			if got := props["parameters"].(map[string]interface{})["data_channel"]; got != "rtm" {
				t.Fatalf("parameters.data_channel = %#v, want rtm", got)
			}
		})
	}

	t.Run("caller struct is not mutated", func(t *testing.T) {
		af := &AdvancedFeatures{EnableRtm: &enableRTM}
		agent := NewAgent(testAgoraClient()).WithAdvancedFeatures(af).WithMllm(realtime)
		if _, err := agent.ToPropertiesMap(basePropertiesOpts()); err != nil {
			t.Fatal(err)
		}
		if af.EnableMllm != nil {
			t.Fatalf("caller AdvancedFeatures.EnableMllm = %v, want nil", *af.EnableMllm)
		}
	})

	t.Run("caller enable_mllm false is overridden for an MLLM session", func(t *testing.T) {
		disabled := false
		agent := NewAgent(testAgoraClient()).
			WithAdvancedFeatures(&AdvancedFeatures{EnableMllm: &disabled, EnableRtm: &enableRTM}).
			WithMllm(realtime)
		props, err := agent.ToPropertiesMap(basePropertiesOpts())
		if err != nil {
			t.Fatal(err)
		}
		assertMLLMEnabledViaAdvancedFeatures(t, props, map[string]interface{}{"enable_mllm": true, "enable_rtm": true})
	})

	t.Run("session keeps tools and rtm", func(t *testing.T) {
		agent := NewAgent(testAgoraClient()).
			WithAdvancedFeatures(&AdvancedFeatures{EnableRtm: &enableRTM}).
			WithTools(true).
			WithMllm(vendors.NewGeminiLive(vendors.GeminiLiveOptions{APIKey: "gl-key", Model: vendors.GeminiLiveModel38Live}))
		payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
		if err != nil {
			t.Fatal(err)
		}
		assertMLLMEnabledViaAdvancedFeatures(t, payload["properties"].(map[string]interface{}),
			map[string]interface{}{"enable_mllm": true, "enable_rtm": true, "enable_tools": true})
	})
}

// rawMLLM is a hand-written MLLM vendor config that still carries the legacy
// mllm.enable key.
type rawMLLM map[string]interface{}

func (r rawMLLM) ToConfig() map[string]interface{} { return r }

func TestRawMLLMEnableIsDropped(t *testing.T) {
	agent := NewAgent(testAgoraClient()).WithMllm(rawMLLM{
		"enable":  true,
		"vendor":  "openai",
		"api_key": "rt-key",
		"params":  map[string]interface{}{"model": "gpt-4o-realtime-preview"},
	})
	if _, exists := agent.MllmConfig()["enable"]; exists {
		t.Fatal("stored MLLM config kept mllm.enable")
	}
	payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
	if err != nil {
		t.Fatal(err)
	}
	assertMLLMEnabledViaAdvancedFeatures(t, payload["properties"].(map[string]interface{}), map[string]interface{}{"enable_mllm": true})
}

func TestCascadingSessionHasNoEnableMLLM(t *testing.T) {
	enableRTM := true
	enabled := true
	cascading := func() *Agent {
		return NewAgent(testAgoraClient()).WithStt(stubASR).WithLlm(stubLLM).WithTts(stubTTS)
	}

	t.Run("no advanced features", func(t *testing.T) {
		payload, err := startPresetValidationSession(t, cascading(), basePipelineSessionOptions())
		if err != nil {
			t.Fatal(err)
		}
		properties := payload["properties"].(map[string]interface{})
		if _, exists := properties["advanced_features"]; exists {
			t.Fatalf("advanced_features = %#v, want absent", properties["advanced_features"])
		}
		if _, exists := properties["mllm"]; exists {
			t.Fatal("cascading body has mllm")
		}
	})

	t.Run("caller advanced features are kept without enable_mllm", func(t *testing.T) {
		agent := cascading().WithAdvancedFeatures(&AdvancedFeatures{EnableRtm: &enableRTM, EnableMllm: &enabled})
		props, err := agent.ToPropertiesMap(basePropertiesOpts())
		if err != nil {
			t.Fatal(err)
		}
		if got := props["advanced_features"]; !reflect.DeepEqual(got, map[string]interface{}{"enable_rtm": true}) {
			t.Fatalf("advanced_features = %#v, want only enable_rtm", got)
		}
	})

	t.Run("only enable_mllm set removes advanced features", func(t *testing.T) {
		agent := cascading().WithAdvancedFeatures(&AdvancedFeatures{EnableMllm: &enabled})
		props, err := agent.ToPropertiesMap(basePropertiesOpts())
		if err != nil {
			t.Fatal(err)
		}
		if _, exists := props["advanced_features"]; exists {
			t.Fatalf("advanced_features = %#v, want absent", props["advanced_features"])
		}
	})
}
