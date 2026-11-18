package agentkit

import (
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func baseAgentForSTTLanguage() *Agent {
	return NewAgent(testAgoraClient()).
		WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{APIKey: "llm-key", Model: "gpt-4o-mini", BaseURL: "https://api.openai.com/v1/chat/completions"})).
		WithTts(vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
			Key:     "tts-key",
			VoiceID: "voice",
			ModelID: "eleven_flash_v2_5",
			BaseURL: "wss://api.elevenlabs.io/v1",
		}))
}

func propertiesForSTTLanguage(t *testing.T, agent *Agent) map[string]interface{} {
	t.Helper()

	props, err := agent.ToPropertiesMap(ToPropertiesOptions{
		Channel:    "channel",
		Token:      "token",
		AgentUID:   "1001",
		RemoteUIDs: []string{"1002"},
	})
	require.NoError(t, err)
	return props
}

func asrFromProperties(t *testing.T, props map[string]interface{}) map[string]interface{} {
	t.Helper()

	asr, ok := props["asr"].(map[string]interface{})
	require.True(t, ok)
	return asr
}

func TestSTTLanguageSerializesBCP47ToProviderParams(t *testing.T) {
	props := propertiesForSTTLanguage(t, baseAgentForSTTLanguage().
		WithStt(vendors.NewSpeechmaticsSTT(vendors.SpeechmaticsSTTOptions{
			APIKey:   "stt-key",
			Language: "en",
		})))

	asr := asrFromProperties(t, props)
	assert.Equal(t, "speechmatics", asr["vendor"])
	assert.Equal(t, "en-US", asr["language"])
	assert.Equal(t, "en-US", props["turn_detection"].(map[string]interface{})["language"])
	assert.Equal(t, "en", asr["params"].(map[string]interface{})["language"])
}

func TestSTTProviderLanguageDoesNotSetTurnDetectionLanguage(t *testing.T) {
	props := propertiesForSTTLanguage(t, baseAgentForSTTLanguage().
		WithStt(vendors.NewSpeechmaticsSTT(vendors.SpeechmaticsSTTOptions{
			APIKey:   "stt-key",
			Language: "en",
		})))

	asr := asrFromProperties(t, props)
	assert.Equal(t, "en-US", asr["language"])
	assert.Equal(t, "en-US", props["turn_detection"].(map[string]interface{})["language"])
	assert.Equal(t, "en", asr["params"].(map[string]interface{})["language"])
}

func TestTurnDetectionLanguageCanDifferFromProviderLanguage(t *testing.T) {
	props := propertiesForSTTLanguage(t, baseAgentForSTTLanguage().
		WithTurnDetection(&TurnDetectionConfig{
			Language: Agora.AsrLanguageFrFr.Ptr(),
		}).
		WithStt(vendors.NewSpeechmaticsSTT(vendors.SpeechmaticsSTTOptions{
			APIKey:   "stt-key",
			Language: "en",
		})))

	asr := asrFromProperties(t, props)
	assert.Equal(t, "fr-FR", asr["language"])
	assert.Equal(t, "fr-FR", props["turn_detection"].(map[string]interface{})["language"])
	assert.Equal(t, "en", asr["params"].(map[string]interface{})["language"])
}

func TestInvalidTurnDetectionLanguagePanics(t *testing.T) {
	assert.PanicsWithValue(t, "invalid turn_detection.language: xx", func() {
		baseAgentForSTTLanguage().WithTurnDetection(&TurnDetectionConfig{
			Language: Agora.AsrLanguage("xx").Ptr(),
		}).ToPropertiesMap(ToPropertiesOptions{
			Channel:    "channel",
			Token:      "token",
			AgentUID:   "1001",
			RemoteUIDs: []string{"1002"},
		})
	})
}

func TestGlobalDefaultASRIsAresWhenSTTIsOmitted(t *testing.T) {
	props := propertiesForSTTLanguage(t, baseAgentForSTTLanguage())

	assert.Equal(t, map[string]interface{}{"vendor": "ares", "language": "en-US"}, props["asr"])
	assert.Equal(t, map[string]interface{}{"language": "en-US"}, props["turn_detection"])
}

func TestSTTVendorParamsMatchDocumentedShapes(t *testing.T) {
	deepgramManaged := vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{Model: "nova-3", Language: "en-US"}).ToConfig()
	assert.NotContains(t, deepgramManaged, "language")
	assert.Equal(t, map[string]interface{}{
		"model":    "nova-3",
		"language": "en-US",
	}, deepgramManaged["params"])

	assert.PanicsWithValue(t, "DeepgramSTT requires APIKey unless using a supported Agora-managed model", func() {
		vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{Model: "enhanced"})
	})

	assert.Equal(t, map[string]interface{}{
		"api_key":  "dg-key",
		"language": "en",
	}, vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{APIKey: "dg-key", Language: "en"}).ToConfig()["params"])

	assert.Equal(t, map[string]interface{}{
		"api_key": "openai-key",
		"input_audio_transcription": map[string]interface{}{
			"model":    "gpt-4o-mini-transcribe",
			"language": "en",
			"prompt":   "Transcribe English speech",
		},
	}, vendors.NewOpenAISTT(vendors.OpenAISTTOptions{
		APIKey:   "openai-key",
		Model:    "gpt-4o-mini-transcribe",
		Language: "en",
		Prompt:   "Transcribe English speech",
	}).ToConfig()["params"])

	assert.PanicsWithValue(t, "OpenAISTT: input_audio_transcription.prompt is required", func() {
		vendors.NewOpenAISTT(vendors.OpenAISTTOptions{
			APIKey:   "openai-key",
			Language: "en",
		}).ToConfig()
	})

	assert.PanicsWithValue(t, "OpenAISTT: input_audio_transcription.language is required", func() {
		vendors.NewOpenAISTT(vendors.OpenAISTTOptions{
			APIKey: "openai-key",
			Prompt: "Transcribe speech",
		}).ToConfig()
	})

	assert.Equal(t, map[string]interface{}{
		"project_id":             "project",
		"location":               "global",
		"adc_credentials_string": "{}",
		"language":               "en-US",
		"model":                  "long",
	}, vendors.NewGoogleSTT(vendors.GoogleSTTOptions{
		ProjectID:            "project",
		Location:             "global",
		ADCCredentialsString: "{}",
		Language:             "en-US",
		Model:                "long",
	}).ToConfig()["params"])

	assert.Equal(t, map[string]interface{}{
		"access_key_id":     "access",
		"secret_access_key": "secret",
		"region":            "us-east-1",
		"language_code":     "en-US",
	}, vendors.NewAmazonSTT(vendors.AmazonSTTOptions{
		AccessKey: "access",
		SecretKey: "secret",
		Region:    "us-east-1",
		Language:  "en-US",
	}).ToConfig()["params"])

	assemblyAIConfig := vendors.NewAssemblyAISTT(vendors.AssemblyAISTTOptions{
		APIKey:   "assembly-key",
		Language: "en-US",
		WsURL:    "wss://example.test/ws",
	}).ToConfig()
	assert.NotContains(t, assemblyAIConfig, "language")
	assert.Equal(t, map[string]interface{}{
		"api_key":  "assembly-key",
		"language": "en-US",
		"ws_url":   "wss://example.test/ws",
	}, assemblyAIConfig["params"])

	xaiSampleRate := vendors.SampleRate24kHz
	assert.Equal(t, map[string]interface{}{
		"api_key":     "xai-key",
		"base_url":    "wss://api.x.ai/v1/stt",
		"language":    "en-US",
		"sample_rate": 24000,
	}, vendors.NewXaiSTT(vendors.XaiSTTOptions{
		APIKey:     "xai-key",
		BaseURL:    "wss://api.x.ai/v1/stt",
		Language:   "en-US",
		SampleRate: &xaiSampleRate,
	}).ToConfig()["params"])
}

func TestAssemblyAIParamsStayNestedAndASRLanguageComesFromTurnDetection(t *testing.T) {
	props := propertiesForSTTLanguage(t, baseAgentForSTTLanguage().
		WithTurnDetection(&TurnDetectionConfig{
			Language: Agora.AsrLanguageFrFr.Ptr(),
		}).
		WithStt(vendors.NewAssemblyAISTT(vendors.AssemblyAISTTOptions{
			APIKey:   "assembly-key",
			Language: "en-US",
			WsURL:    "wss://example.test/ws",
		})))

	assert.Equal(t, map[string]interface{}{
		"vendor":   "assemblyai",
		"language": "fr-FR",
		"params": map[string]interface{}{
			"api_key":  "assembly-key",
			"language": "en-US",
			"ws_url":   "wss://example.test/ws",
		},
	}, props["asr"])
	assert.Equal(t, map[string]interface{}{"language": "fr-FR"}, props["turn_detection"])
}

func TestTurnDetectionLanguagesJoinIntoASRLanguageAndFallBackToVAD(t *testing.T) {
	props := propertiesForSTTLanguage(t, baseAgentForSTTLanguage().
		WithTurnDetection(&TurnDetectionConfig{
			Languages: []Agora.AsrLanguage{Agora.AsrLanguageEnUs, Agora.AsrLanguageEsEs, Agora.AsrLanguageEnUs},
		}))

	asr := asrFromProperties(t, props)
	assert.Equal(t, "ares", asr["vendor"])
	assert.Equal(t, "en-US,es-ES", asr["language"])
	assert.Equal(t, map[string]interface{}{
		"language": "en-US,es-ES",
		"config": map[string]interface{}{
			"end_of_speech": map[string]interface{}{"mode": "vad"},
		},
	}, props["turn_detection"])
}

func TestTurnDetectionLanguagesKeepCallerVADConfig(t *testing.T) {
	vad := Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeechModeVad
	silence := 640
	props := propertiesForSTTLanguage(t, baseAgentForSTTLanguage().
		WithTurnDetection(&TurnDetectionConfig{
			Languages: []Agora.AsrLanguage{Agora.AsrLanguageEnUs, Agora.AsrLanguageHiIn},
			Config: &TurnDetectionNestedConfig{
				EndOfSpeech: &EndOfSpeechConfig{
					Mode:      &vad,
					VadConfig: &EndOfSpeechVadConfig{SilenceDurationMs: &silence},
				},
			},
		}))

	assert.Equal(t, "en-US,hi-IN", asrFromProperties(t, props)["language"])
	assert.Equal(t, map[string]interface{}{
		"mode":       "vad",
		"vad_config": map[string]interface{}{"silence_duration_ms": float64(640)},
	}, props["turn_detection"].(map[string]interface{})["config"].(map[string]interface{})["end_of_speech"])
}

func TestTurnDetectionSingleLanguageListHasNoVADFallback(t *testing.T) {
	props := propertiesForSTTLanguage(t, baseAgentForSTTLanguage().
		WithTurnDetection(&TurnDetectionConfig{Languages: []Agora.AsrLanguage{Agora.AsrLanguageJaJp}}))

	assert.Equal(t, map[string]interface{}{"language": "ja-JP"}, props["turn_detection"])
}

func TestTurnDetectionSemanticEndOfSpeechAllowsOneEnglishOrChineseLanguage(t *testing.T) {
	semantic := Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeechModeSemantic
	props := propertiesForSTTLanguage(t, baseAgentForSTTLanguage().
		WithTurnDetection(&TurnDetectionConfig{
			Language: Agora.AsrLanguageZhCn.Ptr(),
			Config:   &TurnDetectionNestedConfig{EndOfSpeech: &EndOfSpeechConfig{Mode: &semantic}},
		}))

	assert.Equal(t, "zh-CN", asrFromProperties(t, props)["language"])
}

func TestTurnDetectionLanguageErrors(t *testing.T) {
	semantic := Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeechModeSemantic
	semanticConfig := &TurnDetectionNestedConfig{EndOfSpeech: &EndOfSpeechConfig{Mode: &semantic}}
	cases := map[string]struct {
		config  *TurnDetectionConfig
		message string
	}{
		"languages with semantic": {
			config: &TurnDetectionConfig{
				Languages: []Agora.AsrLanguage{Agora.AsrLanguageEnUs, Agora.AsrLanguageZhCn},
				Config:    semanticConfig,
			},
			message: "cannot be used with semantic end_of_speech",
		},
		"semantic with unsupported language": {
			config:  &TurnDetectionConfig{Language: Agora.AsrLanguageFrFr.Ptr(), Config: semanticConfig},
			message: "supports only English and Chinese",
		},
		"language and languages": {
			config: &TurnDetectionConfig{
				Language:  Agora.AsrLanguageEnUs.Ptr(),
				Languages: []Agora.AsrLanguage{Agora.AsrLanguageEsEs},
			},
			message: "not both",
		},
		"empty languages": {
			config:  &TurnDetectionConfig{Languages: []Agora.AsrLanguage{}},
			message: "must not be empty",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := baseAgentForSTTLanguage().WithTurnDetection(tc.config).ToPropertiesMap(ToPropertiesOptions{
				Channel:    "channel",
				Token:      "token",
				AgentUID:   "1001",
				RemoteUIDs: []string{"1002"},
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.message)
		})
	}
}
