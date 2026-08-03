package vendors

import (
	"encoding/json"
	"reflect"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
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
			name: "generic",
			params: NewGenericTTS(GenericTTSOptions{
				URL:            "https://tts.example.com/v1/audio/speech",
				Headers:        map[string]string{"Authorization": "Bearer token"},
				APIKey:         "generic-key",
				Model:          "gpt-4o-mini-tts",
				Voice:          "alloy",
				ResponseFormat: "pcm",
				AdditionalParams: map[string]interface{}{
					"api_key":         "additional-key",
					"model":           "additional-model",
					"voice":           "additional-voice",
					"response_format": "mp3",
					"custom_param":    "custom-value",
				},
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"api_key":         "generic-key",
				"model":           "gpt-4o-mini-tts",
				"voice":           "alloy",
				"response_format": "pcm",
				"custom_param":    "custom-value",
			},
		},
		{
			name: "xai",
			params: NewXaiTTS(XaiTTSOptions{
				APIKey:   "xai-key",
				Language: "en-US",
				VoiceID:  "voice-1",
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"api_key":  "xai-key",
				"language": "en-US",
				"voice_id": "voice-1",
			},
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
			name: "minimax byok voice_id shortcut",
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
			name: "minimax byok additional params",
			params: NewMiniMaxTTS(MiniMaxTTSOptions{
				Key:     "minimax-key",
				GroupID: "group",
				Model:   "speech-02-turbo",
				URL:     "wss://api-uw.minimax.io/ws/v1/t2a_v2",
				AdditionalParams: map[string]interface{}{
					"voice_setting": map[string]interface{}{
						"voice_id":              "voice",
						"speed":                 1,
						"vol":                   1,
						"pitch":                 0,
						"emotion":               "happy",
						"latex_read":            true,
						"english_normalization": true,
					},
					"audio_setting": map[string]interface{}{
						"sample_rate": 16000,
					},
					"pronunciation_dict": map[string]interface{}{
						"tone": []string{"alpha/(ae1)(l)(f)(ah0)", "beta/(b)(ey1)(t)(ah0)"},
					},
					"language_boost": "auto",
				},
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"key":      "minimax-key",
				"group_id": "group",
				"model":    "speech-02-turbo",
				"voice_setting": map[string]interface{}{
					"voice_id":              "voice",
					"speed":                 1,
					"vol":                   1,
					"pitch":                 0,
					"emotion":               "happy",
					"latex_read":            true,
					"english_normalization": true,
				},
				"audio_setting": map[string]interface{}{"sample_rate": 16000},
				"pronunciation_dict": map[string]interface{}{
					"tone": []string{"alpha/(ae1)(l)(f)(ah0)", "beta/(b)(ey1)(t)(ah0)"},
				},
				"language_boost": "auto",
				"url":            "wss://api-uw.minimax.io/ws/v1/t2a_v2",
			},
		},
		{
			name: "minimax voice_id overrides additional voice_setting",
			params: NewMiniMaxTTS(MiniMaxTTSOptions{
				Key:     "minimax-key",
				GroupID: "group",
				Model:   "speech-02-turbo",
				VoiceID: "shortcut-voice",
				URL:     "wss://api-uw.minimax.io/ws/v1/t2a_v2",
				AdditionalParams: map[string]interface{}{
					"voice_setting": map[string]interface{}{
						"voice_id": "additional-voice",
						"speed":    1,
					},
				},
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"key":           "minimax-key",
				"group_id":      "group",
				"model":         "speech-02-turbo",
				"voice_setting": map[string]interface{}{"voice_id": "shortcut-voice"},
				"url":           "wss://api-uw.minimax.io/ws/v1/t2a_v2",
			},
		},
		{
			name: "minimax timber_weights",
			params: NewMiniMaxTTS(MiniMaxTTSOptions{
				Key:     "minimax-key",
				GroupID: "group",
				Model:   "speech-01-turbo",
				URL:     "wss://api-uw.minimax.io/ws/v1/t2a_v2",
				AdditionalParams: map[string]interface{}{
					"timber_weights": []map[string]interface{}{
						{"voice_id": "male-qn-qingse", "weight": 1},
						{"voice_id": "female-shaonv", "weight": 5},
					},
				},
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"key":      "minimax-key",
				"group_id": "group",
				"model":    "speech-01-turbo",
				"url":      "wss://api-uw.minimax.io/ws/v1/t2a_v2",
				"timber_weights": []map[string]interface{}{
					{"voice_id": "male-qn-qingse", "weight": 1},
					{"voice_id": "female-shaonv", "weight": 5},
				},
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
		{
			name: "gradium",
			params: NewGradiumTTS(GradiumTTSOptions{
				APIKey:     "gradium-key",
				URL:        "wss://api.gradium.ai/api/speech/tts",
				ModelName:  "default",
				VoiceID:    "voice-id",
				SampleRate: &sampleRate,
				AdditionalParams: map[string]interface{}{
					"api_key":      "override-key",
					"model_name":   "override-model",
					"sample_rate":  8000,
					"custom_param": "custom-value",
				},
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"api_key":      "gradium-key",
				"url":          "wss://api.gradium.ai/api/speech/tts",
				"model_name":   "default",
				"voice_id":     "voice-id",
				"sample_rate":  24000,
				"custom_param": "custom-value",
			},
		},
		{
			name: "mistral",
			params: NewMistralTTS(MistralTTSOptions{
				APIKey: "mistral-key",
				Model:  "voxtral-mini-tts-2603",
				Voice:  "voice-id",
				AdditionalParams: map[string]interface{}{
					"api_key":      "override-key",
					"model":        "override-model",
					"custom_param": "custom-value",
				},
			}).ToConfig()["params"].(map[string]interface{}),
			want: map[string]interface{}{
				"api_key":      "mistral-key",
				"model":        "voxtral-mini-tts-2603",
				"voice":        "voice-id",
				"custom_param": "custom-value",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !reflect.DeepEqual(tc.params, tc.want) {
				t.Fatalf("params mismatch\nwant: %#v\n got: %#v", tc.want, tc.params)
			}
		})
	}
}

func TestGradiumAndMistralTTSMatchGeneratedUnion(t *testing.T) {
	sampleRate := SampleRate16kHz
	tests := []struct {
		name   string
		config map[string]interface{}
		check  func(*testing.T, *Agora.Tts)
	}{
		{
			name: "gradium",
			config: NewGradiumTTS(GradiumTTSOptions{
				APIKey:     "gradium-key",
				URL:        "wss://api.gradium.ai/api/speech/tts",
				ModelName:  "default",
				VoiceID:    "voice-id",
				SampleRate: &sampleRate,
				AdditionalParams: map[string]interface{}{
					"custom_param": "custom-value",
				},
				SkipPatterns: []int{1, 2},
			}).ToConfig(),
			check: func(t *testing.T, generated *Agora.Tts) {
				t.Helper()
				if generated.Gradium == nil || generated.Gradium.Params == nil {
					t.Fatalf("generated Gradium vendor is nil: %#v", generated)
				}
				if got := generated.Gradium.Params.GetSampleRate(); got == nil || *got != 16000 {
					t.Fatalf("sample_rate = %v, want 16000", got)
				}
				if got := generated.Gradium.Params.GetExtraProperties()["custom_param"]; got != "custom-value" {
					t.Fatalf("custom_param = %v, want custom-value", got)
				}
			},
		},
		{
			name: "mistral",
			config: NewMistralTTS(MistralTTSOptions{
				APIKey: "mistral-key",
				Model:  "voxtral-mini-tts-2603",
				Voice:  "voice-id",
				AdditionalParams: map[string]interface{}{
					"custom_param": "custom-value",
				},
				SkipPatterns: []int{1, 2},
			}).ToConfig(),
			check: func(t *testing.T, generated *Agora.Tts) {
				t.Helper()
				if generated.Mistral == nil || generated.Mistral.Params == nil {
					t.Fatalf("generated Mistral vendor is nil: %#v", generated)
				}
				if got := generated.Mistral.Params.GetModel(); got == nil || *got != "voxtral-mini-tts-2603" {
					t.Fatalf("model = %v, want voxtral-mini-tts-2603", got)
				}
				if got := generated.Mistral.Params.GetExtraProperties()["custom_param"]; got != "custom-value" {
					t.Fatalf("custom_param = %v, want custom-value", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := json.Marshal(tt.config)
			if err != nil {
				t.Fatalf("marshal config: %v", err)
			}
			var generated Agora.Tts
			if err := json.Unmarshal(payload, &generated); err != nil {
				t.Fatalf("unmarshal into generated TTS union: %v", err)
			}
			tt.check(t, &generated)
			if _, err := json.Marshal(generated); err != nil {
				t.Fatalf("marshal generated TTS union: %v", err)
			}
		})
	}
}

func TestGradiumAndMistralTTSValidation(t *testing.T) {
	tests := []struct {
		name      string
		wantPanic string
		create    func()
	}{
		{
			name:      "gradium requires API key",
			wantPanic: "GradiumTTS requires APIKey",
			create: func() {
				NewGradiumTTS(GradiumTTSOptions{})
			},
		},
		{
			name:      "mistral requires API key",
			wantPanic: "MistralTTS requires APIKey",
			create: func() {
				NewMistralTTS(MistralTTSOptions{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertPanic(t, tt.wantPanic, tt.create)
		})
	}
}

func TestRimeTTSCredentialModes(t *testing.T) {
	tests := []struct {
		name string
		opts RimeTTSOptions
		want map[string]interface{}
	}{
		{
			name: "default BYOK mode",
			opts: RimeTTSOptions{
				Key:     "rime-key",
				Speaker: "speaker",
				ModelID: "mist",
			},
			want: map[string]interface{}{
				"vendor": "rime",
				"params": map[string]interface{}{
					"api_key": "rime-key",
					"speaker": "speaker",
					"modelId": "mist",
				},
			},
		},
		{
			name: "explicit BYOK mode",
			opts: RimeTTSOptions{
				CredentialMode: CredentialMode("byok"),
				Key:            "rime-key",
				Speaker:        "speaker",
				ModelID:        "mist",
			},
			want: map[string]interface{}{
				"vendor":          "rime",
				"credential_mode": "byok",
				"params": map[string]interface{}{
					"api_key": "rime-key",
					"speaker": "speaker",
					"modelId": "mist",
				},
			},
		},
		{
			name: "managed mode",
			opts: RimeTTSOptions{
				CredentialMode: CredentialMode("managed"),
				ModelID:        "mist",
				BaseURL:        "wss://managed.rime.example/ws",
			},
			want: map[string]interface{}{
				"vendor":          "rime",
				"credential_mode": "managed",
				"params": map[string]interface{}{
					"modelId":  "mist",
					"base_url": "wss://managed.rime.example/ws",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRimeTTS(tt.opts).ToConfig()
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("config mismatch\nwant: %#v\n got: %#v", tt.want, got)
			}
		})
	}
}

func TestRimeTTSValidation(t *testing.T) {
	tests := []struct {
		name      string
		opts      RimeTTSOptions
		wantPanic string
	}{
		{
			name:      "default mode requires key",
			opts:      RimeTTSOptions{Speaker: "speaker", ModelID: "mist"},
			wantPanic: "RimeTTS requires Key",
		},
		{
			name:      "default mode requires speaker",
			opts:      RimeTTSOptions{Key: "rime-key", ModelID: "mist"},
			wantPanic: "RimeTTS requires Speaker",
		},
		{
			name:      "default mode requires model ID",
			opts:      RimeTTSOptions{Key: "rime-key", Speaker: "speaker"},
			wantPanic: "RimeTTS requires ModelID",
		},
		{
			name: "BYOK mode requires key",
			opts: RimeTTSOptions{
				CredentialMode: CredentialMode("byok"),
				Speaker:        "speaker",
				ModelID:        "mist",
			},
			wantPanic: "RimeTTS requires Key",
		},
		{
			name: "BYOK mode requires speaker",
			opts: RimeTTSOptions{
				CredentialMode: CredentialMode("byok"),
				Key:            "rime-key",
				ModelID:        "mist",
			},
			wantPanic: "RimeTTS requires Speaker",
		},
		{
			name: "BYOK mode requires model ID",
			opts: RimeTTSOptions{
				CredentialMode: CredentialMode("byok"),
				Key:            "rime-key",
				Speaker:        "speaker",
			},
			wantPanic: "RimeTTS requires ModelID",
		},
		{
			name: "managed mode requires base URL",
			opts: RimeTTSOptions{
				CredentialMode: CredentialMode("managed"),
				ModelID:        "mist",
			},
			wantPanic: "RimeTTS requires BaseURL in managed credential mode",
		},
		{
			name: "managed mode requires model ID",
			opts: RimeTTSOptions{
				CredentialMode: CredentialMode("managed"),
				BaseURL:        "wss://managed.rime.example/ws",
			},
			wantPanic: "RimeTTS requires ModelID in managed credential mode",
		},
		{
			name:      "unsupported credential mode",
			opts:      RimeTTSOptions{CredentialMode: "invalid"},
			wantPanic: "RimeTTS CredentialMode must be one of: managed, byok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertPanic(t, tt.wantPanic, func() {
				NewRimeTTS(tt.opts)
			})
		})
	}
}

func TestGenericTTSRequiresOnlyURLAndMatchesGeneratedHTTPVendor(t *testing.T) {
	t.Run("URL only", func(t *testing.T) {
		config := NewGenericTTS(GenericTTSOptions{
			URL: "https://tts.example.com/v1/audio/speech",
		}).ToConfig()

		if got := config["vendor"]; got != "generic_http" {
			t.Fatalf("vendor = %v, want generic_http", got)
		}
		if _, ok := config["headers"]; ok {
			t.Fatalf("headers should be omitted when empty: %#v", config)
		}
		wantParams := map[string]interface{}{}
		if got := config["params"]; !reflect.DeepEqual(got, wantParams) {
			t.Fatalf("params mismatch\nwant: %#v\n got: %#v", wantParams, got)
		}

		payload, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("marshal config: %v", err)
		}
		var generated Agora.Tts
		if err := json.Unmarshal(payload, &generated); err != nil {
			t.Fatalf("unmarshal into generated TTS union: %v", err)
		}
		if generated.GenericHTTP == nil {
			t.Fatalf("generated GenericHTTP vendor is nil: %#v", generated)
		}
		if _, err := json.Marshal(generated); err != nil {
			t.Fatalf("marshal generated TTS union: %v", err)
		}
	})

	t.Run("missing URL", func(t *testing.T) {
		assertPanic(t, "GenericTTS requires URL", func() {
			NewGenericTTS(GenericTTSOptions{})
		})
	})
}

func TestGenericTTSURLValidation(t *testing.T) {
	validURLs := []struct {
		name string
		url  string
	}{
		{name: "HTTP", url: "http://tts.example.com/v1/audio/speech"},
		{name: "HTTPS", url: "https://tts.example.com/v1/audio/speech"},
	}
	for _, tc := range validURLs {
		t.Run(tc.name, func(t *testing.T) {
			config := NewGenericTTS(GenericTTSOptions{URL: tc.url}).ToConfig()
			if got := config["vendor"]; got != "generic_http" {
				t.Fatalf("vendor = %v, want generic_http", got)
			}
		})
	}

	invalidURLs := []struct {
		name      string
		url       string
		wantPanic string
	}{
		{name: "missing scheme", url: "tts.example.com/v1/audio/speech", wantPanic: "GenericTTS currently supports only HTTP and HTTPS URLs"},
		{name: "missing host", url: "https:///v1/audio/speech", wantPanic: "GenericTTS currently supports only HTTP and HTTPS URLs"},
		{name: "malformed", url: "https://tts.example.com/%zz", wantPanic: "GenericTTS currently supports only HTTP and HTTPS URLs"},
		{name: "WebSocket", url: "ws://tts.example.com/v1/audio/speech", wantPanic: "GenericTTS currently supports only HTTP and HTTPS URLs"},
		{name: "secure WebSocket", url: "wss://tts.example.com/v1/audio/speech", wantPanic: "GenericTTS currently supports only HTTP and HTTPS URLs"},
		{name: "FTP", url: "ftp://tts.example.com/v1/audio/speech", wantPanic: "GenericTTS currently supports only HTTP and HTTPS URLs"},
	}
	for _, tc := range invalidURLs {
		t.Run(tc.name, func(t *testing.T) {
			assertPanic(t, tc.wantPanic, func() {
				NewGenericTTS(GenericTTSOptions{URL: tc.url})
			})
		})
	}
}

func TestTypecastTTSMatchesGeneratedUnion(t *testing.T) {
	config := NewTypecastTTS(TypecastTTSOptions{
		APIKey:  "typecast-key",
		VoiceID: "voice-id",
		Model:   "ssfm-v30",
		AdditionalParams: map[string]interface{}{
			"api_key":      "override-key",
			"custom_param": true,
		},
		SkipPatterns: []int{1, 2},
	}).ToConfig()

	if got := config["vendor"]; got != "typecast" {
		t.Fatalf("vendor = %v, want typecast", got)
	}
	params := config["params"].(map[string]interface{})
	if got := params["api_key"]; got != "typecast-key" {
		t.Fatalf("api_key = %v, want typecast-key", got)
	}
	if got := params["voice_id"]; got != "voice-id" {
		t.Fatalf("voice_id = %v, want voice-id", got)
	}
	if got := params["model"]; got != "ssfm-v30" {
		t.Fatalf("model = %v, want ssfm-v30", got)
	}

	payload, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("marshal Typecast config: %v", err)
	}
	var generated Agora.Tts
	if err := json.Unmarshal(payload, &generated); err != nil {
		t.Fatalf("unmarshal Typecast config: %v", err)
	}
	if generated.Typecast == nil || generated.Typecast.Params == nil {
		t.Fatalf("generated Typecast vendor is nil: %#v", generated)
	}
	if generated.Typecast.Params.GetExtraProperties()["custom_param"] != true {
		t.Fatalf("generated params lost custom_param: %#v", generated.Typecast.Params)
	}
}

func TestTypecastTTSValidation(t *testing.T) {
	tests := []struct {
		name      string
		opts      TypecastTTSOptions
		wantPanic string
	}{
		{
			name:      "API key required",
			opts:      TypecastTTSOptions{VoiceID: "voice", Model: "model"},
			wantPanic: "TypecastTTS requires APIKey",
		},
		{
			name:      "voice ID required",
			opts:      TypecastTTSOptions{APIKey: "key", Model: "model"},
			wantPanic: "TypecastTTS requires VoiceID",
		},
		{
			name:      "model required",
			opts:      TypecastTTSOptions{APIKey: "key", VoiceID: "voice"},
			wantPanic: "TypecastTTS requires Model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertPanic(t, tt.wantPanic, func() {
				NewTypecastTTS(tt.opts)
			})
		})
	}
}

func ptrInt(v int) *int {
	return &v
}

func ptrFloat(v float64) *float64 {
	return &v
}
