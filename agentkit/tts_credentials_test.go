package agentkit

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
)

func TestCartesiaAndFishAudioTTSCredentialModes(t *testing.T) {
	sampleRate := vendors.SampleRate24kHz
	providers := []struct {
		name   string
		create func(CredentialMode, string) vendors.TTS
		params map[string]interface{}
	}{
		{
			name: "cartesia",
			create: func(mode CredentialMode, key string) vendors.TTS {
				return vendors.NewCartesiaTTS(vendors.CartesiaTTSOptions{
					CredentialMode: mode,
					APIKey:         key,
					VoiceID:        "voice",
					ModelID:        "sonic-2",
					BaseURL:        "wss://api.cartesia.ai/tts/websocket",
					Language:       "en",
					SampleRate:     &sampleRate,
					SkipPatterns:   []int{1, 2},
				})
			},
			params: map[string]interface{}{
				"model_id": "sonic-2",
				"voice":    map[string]interface{}{"mode": "id", "id": "voice"},
				"base_url": "wss://api.cartesia.ai/tts/websocket",
				"language": "en",
				"output_format": map[string]interface{}{
					"container": "raw", "sample_rate": 24000,
				},
			},
		},
		{
			name: "fishaudio",
			create: func(mode CredentialMode, key string) vendors.TTS {
				return vendors.NewFishAudioTTS(vendors.FishAudioTTSOptions{
					CredentialMode: mode,
					Key:            key,
					ReferenceID:    "ref",
					Backend:        "speech-1.5",
					SkipPatterns:   []int{1, 2},
				})
			},
			params: map[string]interface{}{"reference_id": "ref", "backend": "speech-1.5"},
		},
	}
	modes := []struct {
		name string
		mode CredentialMode
		key  string
	}{
		{name: "implicit BYOK", key: "tts-key"},
		{name: "explicit BYOK", mode: CredentialModeBYOK, key: "tts-key"},
		{name: "managed without key", mode: CredentialModeManaged},
		{name: "managed with key", mode: CredentialModeManaged, key: "tts-key"},
	}
	for _, provider := range providers {
		for _, mode := range modes {
			t.Run(provider.name+"/"+mode.name, func(t *testing.T) {
				params := make(map[string]interface{}, len(provider.params)+1)
				for key, value := range provider.params {
					params[key] = value
				}
				if mode.key != "" {
					params["api_key"] = mode.key
				}
				want := map[string]interface{}{
					"vendor": provider.name, "params": params, "skip_patterns": []int{1, 2},
				}
				if mode.mode != "" {
					want["credential_mode"] = string(mode.mode)
				}
				vendor := provider.create(mode.mode, mode.key)
				if got := vendor.ToConfig(); !reflect.DeepEqual(got, want) {
					t.Fatalf("ToConfig() = %#v, want %#v", got, want)
				}

				// Exercise AgentKit's conversion to generated request types and the final JSON payload.
				opts := basePropertiesOpts()
				opts.AllowMissingVendorCategories = []string{"asr"}
				props, err := NewAgent(testAgoraClient()).WithLlm(stubLLM).WithTts(vendor).ToProperties(opts)
				if err != nil {
					t.Fatal(err)
				}
				payload, err := json.Marshal(props.Tts)
				if err != nil {
					t.Fatal(err)
				}
				wantJSON, err := json.Marshal(want)
				if err != nil {
					t.Fatal(err)
				}
				var gotWire, wantWire map[string]interface{}
				if err := json.Unmarshal(payload, &gotWire); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(wantJSON, &wantWire); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(gotWire, wantWire) {
					t.Fatalf("wire payload = %s, want %s", payload, wantJSON)
				}
			})
		}
	}
}
