package vendors

import (
	"reflect"
	"testing"
)

func TestTTSVendorParamsMatchGeneratedCoreShapes(t *testing.T) {
	sampleRate := SampleRate24kHz

	cases := []struct {
		name   string
		params map[string]interface{}
		want   map[string]interface{}
	}{
		{
			name: "microsoft",
			params: NewMicrosoftTTS(MicrosoftTTSOptions{
				Key:       "ms-key",
				Region:    "eastus",
				VoiceName: "en-US-JennyNeural",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"key":        "ms-key",
				"region":     "eastus",
				"voice_name": "en-US-JennyNeural",
			},
		},
		{
			name: "amazon",
			params: NewAmazonTTS(AmazonTTSOptions{
				AccessKey: "access",
				SecretKey: "secret",
				Region:    "us-east-1",
				VoiceID:   "Joanna",
				Engine:    "neural",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"aws_access_key_id":     "access",
				"aws_secret_access_key": "secret",
				"region_name":           "us-east-1",
				"voice":                 "Joanna",
				"engine":                "neural",
			},
		},
		{
			name: "google",
			params: NewGoogleTTS(GoogleTTSOptions{
				Key:          "{}",
				VoiceName:    "en-US-JennyNeural",
				LanguageCode: "en-US",
				SampleRate:   &sampleRate,
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"credentials":          "{}",
				"VoiceSelectionParams": map[string]interface{}{"name": "en-US-JennyNeural", "language_code": "en-US"},
				"AudioConfig":          map[string]interface{}{"sample_rate_hertz": 24000},
			},
		},
		{
			name: "cartesia",
			params: NewCartesiaTTS(CartesiaTTSOptions{
				APIKey:     "cartesia-key",
				VoiceID:    "voice",
				ModelID:    "sonic-2",
				SampleRate: &sampleRate,
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"api_key":       "cartesia-key",
				"model_id":      "sonic-2",
				"voice":         map[string]interface{}{"mode": "id", "id": "voice"},
				"output_format": map[string]interface{}{"container": "raw", "sample_rate": 24000},
			},
		},
		{
			name: "rime",
			params: NewRimeTTS(RimeTTSOptions{
				Key:     "rime-key",
				Speaker: "speaker",
				ModelID: "mist",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{"api_key": "rime-key", "speaker": "speaker", "modelId": "mist"},
		},
		{
			name: "fish",
			params: NewFishAudioTTS(FishAudioTTSOptions{
				Key:         "fish-key",
				ReferenceID: "ref",
				Backend:     "speech-1.5",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{"api_key": "fish-key", "reference_id": "ref", "backend": "speech-1.5"},
		},
		{
			name: "elevenlabs",
			params: NewElevenLabsTTS(ElevenLabsTTSOptions{
				Key:     "eleven-key",
				ModelID: "eleven_flash_v2_5",
				VoiceID: "voice",
				BaseURL: "wss://api.elevenlabs.io/v1",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"key":      "eleven-key",
				"base_url": "wss://api.elevenlabs.io/v1",
				"model_id": "eleven_flash_v2_5",
				"voice_id": "voice",
			},
		},
		{
			name: "deepgram",
			params: NewDeepgramTTS(DeepgramTTSOptions{
				APIKey:     "deepgram-key",
				Model:      "aura-2-thalia-en",
				BaseURL:    "wss://api.deepgram.com/v1/speak",
				SampleRate: &sampleRate,
				AdditionalParams: map[string]interface{}{
					"api_key":     "override-key",
					"model":       "override-model",
					"base_url":    "wss://override.example.com",
					"sample_rate": 16000,
					"encoding":    "linear16",
				},
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"api_key":     "deepgram-key",
				"model":       "aura-2-thalia-en",
				"base_url":    "wss://api.deepgram.com/v1/speak",
				"sample_rate": 24000,
				"encoding":    "linear16",
			},
		},
		{
			name: "openai byok",
			params: NewOpenAITTS(OpenAITTSOptions{
				APIKey:       "openai-key",
				Voice:        "coral",
				Model:        "gpt-4o-mini-tts",
				BaseURL:      "https://api.openai.com/v1",
				Instructions: "speak clearly",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"voice":        "coral",
				"api_key":      "openai-key",
				"base_url":     "https://api.openai.com/v1",
				"model":        "gpt-4o-mini-tts",
				"instructions": "speak clearly",
			},
		},
		{
			name:   "openai preset",
			params: NewOpenAITTS(OpenAITTSOptions{Voice: "coral"}).ToConfig()["params"].(map[string]interface{}),
			want:   map[string]interface{}{"voice": "coral"},
		},
		{
			name: "humeai",
			params: NewHumeAITTS(HumeAITTSOptions{
				Key:      "hume-key",
				VoiceID:  "voice",
				Provider: "CUSTOM_VOICE",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"key":      "hume-key",
				"voice_id": "voice",
				"provider": "CUSTOM_VOICE",
			},
		},
		{
			name: "minimax byok",
			params: NewMiniMaxTTS(MiniMaxTTSOptions{
				Key:     "minimax-key",
				GroupID: "group",
				Model:   "speech-02-turbo",
				VoiceID: "voice",
				URL:     "wss://api-uw.minimax.io/ws/v1/t2a_v2",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"key":           "minimax-key",
				"group_id":      "group",
				"model":         "speech-02-turbo",
				"voice_setting": map[string]interface{}{"voice_id": "voice"},
				"url":           "wss://api-uw.minimax.io/ws/v1/t2a_v2",
			},
		},
		{
			name: "sarvam",
			params: NewSarvamTTS(SarvamTTSOptions{
				Key:                "sarvam-key",
				Speaker:            "anushka",
				TargetLanguageCode: "en-IN",
				SampleRate:         ptrInt(24000),
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"api_subscription_key": "sarvam-key",
				"speaker":              "anushka",
				"target_language_code": "en-IN",
				"sample_rate":          24000,
			},
		},
		{
			name: "murf",
			params: NewMurfTTS(MurfTTSOptions{
				Key:        "murf-key",
				VoiceID:    "Ariana",
				BaseURL:    "wss://murf.example/ws",
				Locale:     "en-US",
				Rate:       ptrFloat(0),
				Pitch:      ptrFloat(0),
				Model:      "FALCON",
				SampleRate: ptrInt(24000),
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"api_key":     "murf-key",
				"base_url":    "wss://murf.example/ws",
				"voiceId":     "Ariana",
				"locale":      "en-US",
				"rate":        float64(0),
				"pitch":       float64(0),
				"model":       "FALCON",
				"sample_rate": 24000,
			},
		},
		{
			name: "murf minimal",
			params: NewMurfTTS(MurfTTSOptions{
				Key: "murf-key",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{"api_key": "murf-key"},
		},
	}

	for _, tc := range cases {
		if !reflect.DeepEqual(tc.params, tc.want) {
			t.Fatalf("%s params mismatch\nwant: %#v\n got: %#v", tc.name, tc.want, tc.params)
		}
	}
}

func ptrInt(v int) *int {
	return &v
}

func ptrFloat(v float64) *float64 {
	return &v
}
