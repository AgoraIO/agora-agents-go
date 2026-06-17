package agentkit

import (
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func basePropertiesOpts() ToPropertiesOptions {
	return ToPropertiesOptions{
		Channel:    "channel",
		Token:      "token",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
	}
}

var stubLLM = vendors.NewOpenAI(vendors.OpenAIOptions{
	APIKey:  "stub-llm-key",
	BaseURL: "https://api.openai.com/v1/chat/completions",
	Model:   "gpt-4o",
})

var stubTTS = vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
	Key:     "stub-tts-key",
	VoiceID: "stub-voice",
	ModelID: "eleven_flash_v2_5",
	BaseURL: "wss://api.elevenlabs.io/v1",
})

var stubASR = vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
	APIKey:   "stub-asr-key",
	Language: "en",
})

// ─────────────────────────────────────────────────────────────────────────────
// Scenario 1 — BYOK pipeline (properties shape)
// ─────────────────────────────────────────────────────────────────────────────

func TestRequestBodyScenario1BYOKPropertiesShape(t *testing.T) {
	agent := NewAgent(WithName("byok-agent")).
		WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
			APIKey:   "dg-key",
			Language: "en",
		})).
		WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
			APIKey:  "openai-key",
			BaseURL: "https://api.openai.com/v1/chat/completions",
			Model:   "gpt-4o",
		})).
		WithTts(vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
			Key:     "el-key",
			ModelID: "eleven_flash_v2_5",
			VoiceID: "some-voice",
			BaseURL: "wss://api.elevenlabs.io/v1",
		}))

	props, err := agent.ToPropertiesMap(basePropertiesOpts())
	require.NoError(t, err)

	// ASR
	asr := props["asr"].(map[string]interface{})
	assert.Equal(t, "deepgram", asr["vendor"])
	asrParams := asr["params"].(map[string]interface{})
	assert.Equal(t, "dg-key", asrParams["key"])
	assert.Equal(t, "en", asrParams["language"])

	// LLM
	llm := props["llm"].(map[string]interface{})
	assert.Equal(t, "openai-key", llm["api_key"])
	assert.Equal(t, "https://api.openai.com/v1/chat/completions", llm["url"])
	assert.Equal(t, "openai", llm["style"])
	llmParams := llm["params"].(map[string]interface{})
	assert.Equal(t, "gpt-4o", llmParams["model"])

	// TTS
	tts := props["tts"].(map[string]interface{})
	assert.Equal(t, "elevenlabs", tts["vendor"])
	ttsParams := tts["params"].(map[string]interface{})
	assert.Equal(t, "el-key", ttsParams["key"])
	assert.Equal(t, "eleven_flash_v2_5", ttsParams["model_id"])
	assert.Equal(t, "some-voice", ttsParams["voice_id"])
	assert.Equal(t, "wss://api.elevenlabs.io/v1", ttsParams["base_url"])
}

// ─────────────────────────────────────────────────────────────────────────────
// Scenario 2a — Preset-backed pipeline (full start request)
// ─────────────────────────────────────────────────────────────────────────────

func TestRequestBodyScenario2aPresetBackedPipeline(t *testing.T) {
	agent := NewAgent(WithName("preset-agent")).
		WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
			Model:    "nova-2",
			Language: "en-US",
		})).
		WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
			Model: "gpt-4o-mini",
		})).
		WithTts(vendors.NewOpenAITTS(vendors.OpenAITTSOptions{
			Voice: "alloy",
		}))

	payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
	require.NoError(t, err)

	assert.Equal(t, "deepgram_nova_2,openai_gpt_4o_mini,openai_tts_1", payload["preset"])

	properties := payload["properties"].(map[string]interface{})

	// LLM api_key absent (managed)
	llm := properties["llm"].(map[string]interface{})
	assert.NotContains(t, llm, "api_key")

	// ASR model absent from params (managed)
	asr := properties["asr"].(map[string]interface{})
	asrParams, hasParams := asr["params"].(map[string]interface{})
	if hasParams {
		assert.NotContains(t, asrParams, "model")
	}

	// TTS params: voice present, api_key/model absent
	tts := properties["tts"].(map[string]interface{})
	ttsParams := tts["params"].(map[string]interface{})
	assert.Equal(t, "alloy", ttsParams["voice"])
	assert.NotContains(t, ttsParams, "api_key")
	assert.NotContains(t, ttsParams, "model")
}

// ─────────────────────────────────────────────────────────────────────────────
// Scenario 3 — LLM vendor config wins over agent-level convenience fields
// ─────────────────────────────────────────────────────────────────────────────

func TestRequestBodyScenario3VendorGreetingWins(t *testing.T) {
	opts := basePropertiesOpts()
	opts.AllowMissingVendorCategories = []string{"asr", "tts"}

	agent := NewAgent(WithName("greeting-agent")).
		WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
			APIKey:          "openai-key",
			BaseURL:         "https://api.openai.com/v1/chat/completions",
			Model:           "gpt-4o",
			GreetingMessage: "vendor greeting",
			SystemMessages: []map[string]interface{}{
				{"role": "system", "content": "You are helpful."},
			},
		})).
		WithGreeting("agent greeting")

	props, err := agent.ToPropertiesMap(opts)
	require.NoError(t, err)

	llm := props["llm"].(map[string]interface{})
	// Vendor greeting wins
	assert.Equal(t, "vendor greeting", llm["greeting_message"])
	// System messages from vendor
	assert.NotNil(t, llm["system_messages"])
}

func TestRequestBodyScenario3AgentGreetingFillsIn(t *testing.T) {
	opts := basePropertiesOpts()
	opts.AllowMissingVendorCategories = []string{"asr", "tts"}

	// LLM without greeting; agent has greeting → agent value fills in
	agent := NewAgent(WithName("fill-agent")).
		WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
			APIKey:  "openai-key",
			BaseURL: "https://api.openai.com/v1/chat/completions",
			Model:   "gpt-4o",
		})).
		WithGreeting("agent greeting")

	props, err := agent.ToPropertiesMap(opts)
	require.NoError(t, err)

	llm := props["llm"].(map[string]interface{})
	assert.Equal(t, "agent greeting", llm["greeting_message"])
}

// ─────────────────────────────────────────────────────────────────────────────
// Scenario 4 — VertexAI URL construction
// ─────────────────────────────────────────────────────────────────────────────

func TestRequestBodyScenario4VertexAIURLConstruction(t *testing.T) {
	opts := basePropertiesOpts()
	opts.AllowMissingVendorCategories = []string{"asr", "tts"}

	agent := NewAgent(WithName("vertex-agent")).
		WithLlm(vendors.NewVertexAILLM(vendors.VertexAILLMOptions{
			GeminiOptions: vendors.GeminiOptions{
				APIKey: "vertex-key",
				Model:  "gemini-2.0-flash",
			},
			ProjectID: "my-project",
			Location:  "us-central1",
		}))

	props, err := agent.ToPropertiesMap(opts)
	require.NoError(t, err)

	llm := props["llm"].(map[string]interface{})
	expectedURL := "https://us-central1-aiplatform.googleapis.com/v1/projects/my-project/locations/us-central1/publishers/google/models/gemini-2.0-flash:streamGenerateContent?alt=sse"
	assert.Equal(t, expectedURL, llm["url"])
	assert.Equal(t, "gemini", llm["style"])

	params := llm["params"].(map[string]interface{})
	assert.Equal(t, "gemini-2.0-flash", params["model"])
	assert.NotContains(t, params, "project_id")
	assert.NotContains(t, params, "location")
}

func TestRequestBodyScenario4VertexAIExplicitURLOverride(t *testing.T) {
	opts := basePropertiesOpts()
	opts.AllowMissingVendorCategories = []string{"asr", "tts"}

	agent := NewAgent(WithName("vertex-agent")).
		WithLlm(vendors.NewVertexAILLM(vendors.VertexAILLMOptions{
			GeminiOptions: vendors.GeminiOptions{
				APIKey: "vertex-key",
				Model:  "gemini-2.0-flash",
				URL:    "https://custom.vertex.example.com",
			},
			ProjectID: "my-project",
			Location:  "us-central1",
		}))

	props, err := agent.ToPropertiesMap(opts)
	require.NoError(t, err)

	llm := props["llm"].(map[string]interface{})
	assert.Equal(t, "https://custom.vertex.example.com", llm["url"])
}

// ─────────────────────────────────────────────────────────────────────────────
// Scenario 5 — OpenAISTT sub-cases
// ─────────────────────────────────────────────────────────────────────────────

func baseOpenAISTTOpts() ToPropertiesOptions {
	opts := basePropertiesOpts()
	opts.AllowMissingVendorCategories = []string{"llm", "tts"}
	return opts
}

func TestRequestBodyScenario5aOpenAISTTBasicConfig(t *testing.T) {
	agent := NewAgent(WithName("oai-stt")).
		WithStt(vendors.NewOpenAISTT(vendors.OpenAISTTOptions{
			APIKey:   "stt-key",
			Model:    "gpt-4o-mini-transcribe",
			Language: "en",
			Prompt:   "transcribe accurately",
		}))

	props, err := agent.ToPropertiesMap(baseOpenAISTTOpts())
	require.NoError(t, err)

	asr := props["asr"].(map[string]interface{})
	assert.Equal(t, "openai", asr["vendor"])
	asrParams := asr["params"].(map[string]interface{})
	assert.Equal(t, "stt-key", asrParams["api_key"])

	transcription := asrParams["input_audio_transcription"].(map[string]interface{})
	assert.Equal(t, "gpt-4o-mini-transcribe", transcription["model"])
	assert.Equal(t, "en", transcription["language"])
	assert.Equal(t, "transcribe accurately", transcription["prompt"])
}

func TestRequestBodyScenario5bOpenAISTTModelInTranscription(t *testing.T) {
	agent := NewAgent(WithName("oai-stt")).
		WithStt(vendors.NewOpenAISTT(vendors.OpenAISTTOptions{
			APIKey:   "stt-key",
			Model:    "gpt-4o-transcribe",
			Language: "fr",
			Prompt:   "parlez français",
		}))

	props, err := agent.ToPropertiesMap(baseOpenAISTTOpts())
	require.NoError(t, err)

	asr := props["asr"].(map[string]interface{})
	asrParams := asr["params"].(map[string]interface{})
	transcription := asrParams["input_audio_transcription"].(map[string]interface{})
	assert.Equal(t, "gpt-4o-transcribe", transcription["model"])
	assert.Equal(t, "fr", transcription["language"])
}

func TestRequestBodyScenario5cOpenAISTTLanguageInTranscription(t *testing.T) {
	agent := NewAgent(WithName("oai-stt")).
		WithStt(vendors.NewOpenAISTT(vendors.OpenAISTTOptions{
			APIKey:   "stt-key",
			Model:    "gpt-4o-mini-transcribe",
			Language: "de",
			Prompt:   "transcribe",
		}))

	props, err := agent.ToPropertiesMap(baseOpenAISTTOpts())
	require.NoError(t, err)

	asr := props["asr"].(map[string]interface{})
	asrParams := asr["params"].(map[string]interface{})
	transcription := asrParams["input_audio_transcription"].(map[string]interface{})
	assert.Equal(t, "de", transcription["language"])
}

func TestRequestBodyScenario5dOpenAISTTVendorIsOpenAI(t *testing.T) {
	agent := NewAgent(WithName("oai-stt")).
		WithStt(vendors.NewOpenAISTT(vendors.OpenAISTTOptions{
			APIKey:   "stt-key",
			Model:    "gpt-4o-mini-transcribe",
			Language: "en",
			Prompt:   "transcribe",
		}))

	props, err := agent.ToPropertiesMap(baseOpenAISTTOpts())
	require.NoError(t, err)

	asr := props["asr"].(map[string]interface{})
	assert.Equal(t, "openai", asr["vendor"])
}

// ─────────────────────────────────────────────────────────────────────────────
// Scenario 6 — Mixed preset + BYOK
// ─────────────────────────────────────────────────────────────────────────────

func TestRequestBodyScenario6aManagedASRBYOKLLMManagedTTS(t *testing.T) {
	agent := NewAgent(WithName("mixed")).
		WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
			Model:    "nova-2",
			Language: "en-US",
		})).
		WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
			APIKey:  "byok-llm-key",
			BaseURL: "https://api.openai.com/v1/chat/completions",
			Model:   "gpt-4o",
		})).
		WithTts(vendors.NewOpenAITTS(vendors.OpenAITTSOptions{
			Voice: "alloy",
		}))

	payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
	require.NoError(t, err)

	// ASR and TTS managed presets, LLM is BYOK
	preset := payload["preset"].(string)
	assert.Contains(t, preset, "deepgram_nova_2")
	assert.Contains(t, preset, "openai_tts_1")

	properties := payload["properties"].(map[string]interface{})
	llm := properties["llm"].(map[string]interface{})
	assert.Equal(t, "byok-llm-key", llm["api_key"])
}

func TestRequestBodyScenario6bBYOKASRManagedLLMBYOKTTS(t *testing.T) {
	// Use nova-2 model with BYOK key to exercise the key-detection path in inferASRPreset.
	// Without a key, even nova-2 would be managed; this test confirms the key gates BYOK correctly.
	agent := NewAgent(WithName("mixed")).
		WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
			APIKey:   "byok-asr-key",
			Model:    "nova-2",
			Language: "en",
		})).
		WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
			Model: "gpt-4o-mini",
		})).
		WithTts(vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
			Key:     "byok-tts-key",
			ModelID: "eleven_flash_v2_5",
			VoiceID: "some-voice",
			BaseURL: "wss://api.elevenlabs.io/v1",
		}))

	payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
	require.NoError(t, err)

	// LLM is managed preset; ASR is BYOK (key present) — no Deepgram preset inferred
	preset := payload["preset"].(string)
	assert.Contains(t, preset, "openai_gpt_4o_mini")
	assert.NotContains(t, preset, "deepgram_nova_2")

	properties := payload["properties"].(map[string]interface{})

	// ASR is BYOK — key and model both retained (nothing stripped)
	asr := properties["asr"].(map[string]interface{})
	asrParams := asr["params"].(map[string]interface{})
	assert.Equal(t, "byok-asr-key", asrParams["key"])
	assert.Equal(t, "nova-2", asrParams["model"])

	// LLM is managed — no api_key
	llm := properties["llm"].(map[string]interface{})
	assert.NotContains(t, llm, "api_key")

	// TTS is BYOK
	tts := properties["tts"].(map[string]interface{})
	assert.Equal(t, "elevenlabs", tts["vendor"])
	ttsParams := tts["params"].(map[string]interface{})
	assert.Equal(t, "byok-tts-key", ttsParams["key"])
}

// ─────────────────────────────────────────────────────────────────────────────
// Scenario 7 — Pipeline ID (7b and 7c)
// ─────────────────────────────────────────────────────────────────────────────

func TestRequestBodyScenario7bPipelineIDWithBYOKTTSOnly(t *testing.T) {
	agent := NewAgent(WithName("support"), WithPipelineID("studio-pipeline")).
		WithTts(vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
			Key:     "el-key",
			ModelID: "eleven_flash_v2_5",
			VoiceID: "some-voice",
			BaseURL: "wss://api.elevenlabs.io/v1",
		}))

	payload := startPipelineIDSession(t, agent, basePipelineSessionOptions())

	assert.Equal(t, "studio-pipeline", payload["pipeline_id"])
	properties := payload["properties"].(map[string]interface{})
	assert.NotContains(t, properties, "asr")
	assert.NotContains(t, properties, "llm")

	tts := properties["tts"].(map[string]interface{})
	assert.Equal(t, "elevenlabs", tts["vendor"])
	ttsParams := tts["params"].(map[string]interface{})
	assert.Equal(t, "el-key", ttsParams["key"])
}

func TestRequestBodyScenario7cPipelineIDWithBYOKASRAndTTS(t *testing.T) {
	agent := NewAgent(WithName("support"), WithPipelineID("studio-pipeline")).
		WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
			APIKey:   "dg-key",
			Language: "en",
		})).
		WithTts(vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
			Key:     "el-key",
			ModelID: "eleven_flash_v2_5",
			VoiceID: "some-voice",
			BaseURL: "wss://api.elevenlabs.io/v1",
		}))

	payload := startPipelineIDSession(t, agent, basePipelineSessionOptions())

	assert.Equal(t, "studio-pipeline", payload["pipeline_id"])
	properties := payload["properties"].(map[string]interface{})
	assert.NotContains(t, properties, "llm")

	asr := properties["asr"].(map[string]interface{})
	assert.Equal(t, "deepgram", asr["vendor"])

	tts := properties["tts"].(map[string]interface{})
	assert.Equal(t, "elevenlabs", tts["vendor"])
}

// ─────────────────────────────────────────────────────────────────────────────
// Scenario 8 — MLLM mode
// ─────────────────────────────────────────────────────────────────────────────

func TestRequestBodyScenario8aOpenAIRealtimeMLLM(t *testing.T) {
	agent := NewAgent(WithName("mllm-agent")).
		WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
			APIKey: "realtime-key",
			Model:  "gpt-4o-realtime-preview",
			Voice:  "coral",
		}))

	payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
	require.NoError(t, err)

	properties := payload["properties"].(map[string]interface{})
	assert.Contains(t, properties, "mllm")
	assert.NotContains(t, properties, "asr")
	assert.NotContains(t, properties, "llm")
	assert.NotContains(t, properties, "tts")

	mllm := properties["mllm"].(map[string]interface{})
	assert.Equal(t, "openai", mllm["vendor"])
	assert.Equal(t, true, mllm["enable"])
	assert.Equal(t, "realtime-key", mllm["api_key"])
	params := mllm["params"].(map[string]interface{})
	assert.Equal(t, "gpt-4o-realtime-preview", params["model"])
	assert.Equal(t, "coral", params["voice"])
}

func TestRequestBodyScenario8bMLLMAgentGreetingFillIn(t *testing.T) {
	opts := basePropertiesOpts()

	agent := NewAgent(WithName("mllm-agent")).
		WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
			APIKey: "realtime-key",
			Model:  "gpt-4o-realtime-preview",
			Voice:  "coral",
		})).
		WithGreeting("hello from agent")

	props, err := agent.ToPropertiesMap(opts)
	require.NoError(t, err)

	mllm := props["mllm"].(map[string]interface{})
	assert.Equal(t, "hello from agent", mllm["greeting_message"])
}

func TestRequestBodyScenario8cMLLMVendorGreetingWins(t *testing.T) {
	opts := basePropertiesOpts()

	agent := NewAgent(WithName("mllm-agent")).
		WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
			APIKey:          "realtime-key",
			Model:           "gpt-4o-realtime-preview",
			Voice:           "coral",
			GreetingMessage: "vendor greeting",
		})).
		WithGreeting("agent greeting")

	props, err := agent.ToPropertiesMap(opts)
	require.NoError(t, err)

	mllm := props["mllm"].(map[string]interface{})
	assert.Equal(t, "vendor greeting", mllm["greeting_message"])
}

// ─────────────────────────────────────────────────────────────────────────────
// BYOK ASR Vendor Shapes
// ─────────────────────────────────────────────────────────────────────────────

func TestBYOKASRVendorShapes(t *testing.T) {
	asrOpts := func() ToPropertiesOptions {
		opts := basePropertiesOpts()
		opts.AllowMissingVendorCategories = []string{"llm", "tts"}
		return opts
	}

	t.Run("Deepgram", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
				APIKey:   "dg-key",
				Model:    "nova-2",
				Language: "en",
			}))
		props, err := agent.ToPropertiesMap(asrOpts())
		require.NoError(t, err)
		asr := props["asr"].(map[string]interface{})
		assert.Equal(t, "deepgram", asr["vendor"])
		p := asr["params"].(map[string]interface{})
		assert.Equal(t, "dg-key", p["key"])
		assert.Equal(t, "nova-2", p["model"])
		assert.Equal(t, "en", p["language"])
	})

	t.Run("Deepgram/keyterm", func(t *testing.T) {
		// APIKey → wire key "key"; keyterm passes through unchanged
		config := vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
			APIKey:   "dg-key",
			Model:    "nova-3",
			Language: "en",
			Keyterm:  "term",
		}).ToConfig()
		p := config["params"].(map[string]interface{})
		assert.Equal(t, "dg-key", p["key"])
		assert.Equal(t, "nova-3", p["model"])
		assert.Equal(t, "en", p["language"])
		assert.Equal(t, "term", p["keyterm"])
	})

	t.Run("Microsoft", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithStt(vendors.NewMicrosoftSTT(vendors.MicrosoftSTTOptions{
				Key:      "ms-key",
				Region:   "eastus",
				Language: "en-US",
			}))
		props, err := agent.ToPropertiesMap(asrOpts())
		require.NoError(t, err)
		asr := props["asr"].(map[string]interface{})
		assert.Equal(t, "microsoft", asr["vendor"])
		p := asr["params"].(map[string]interface{})
		assert.Equal(t, "ms-key", p["key"])
		assert.Equal(t, "eastus", p["region"])
		assert.Equal(t, "en-US", p["language"])
	})

	t.Run("OpenAISTT", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithStt(vendors.NewOpenAISTT(vendors.OpenAISTTOptions{
				APIKey:   "stt-key",
				Model:    "gpt-4o-mini-transcribe",
				Language: "en",
				Prompt:   "transcribe",
			}))
		props, err := agent.ToPropertiesMap(asrOpts())
		require.NoError(t, err)
		asr := props["asr"].(map[string]interface{})
		assert.Equal(t, "openai", asr["vendor"])
		p := asr["params"].(map[string]interface{})
		assert.Equal(t, "stt-key", p["api_key"])
		transcription := p["input_audio_transcription"].(map[string]interface{})
		assert.Equal(t, "gpt-4o-mini-transcribe", transcription["model"])
		assert.Equal(t, "transcribe", transcription["prompt"])
		assert.Equal(t, "en", transcription["language"])
	})

	t.Run("Google", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithStt(vendors.NewGoogleSTT(vendors.GoogleSTTOptions{
				ProjectID:            "proj",
				Location:             "global",
				ADCCredentialsString: `{"type":"service_account"}`,
				Language:             "en-US",
				Model:                "long",
			}))
		props, err := agent.ToPropertiesMap(asrOpts())
		require.NoError(t, err)
		asr := props["asr"].(map[string]interface{})
		assert.Equal(t, "google", asr["vendor"])
		p := asr["params"].(map[string]interface{})
		assert.Equal(t, "proj", p["project_id"])
		assert.Equal(t, "global", p["location"])
		assert.Equal(t, `{"type":"service_account"}`, p["adc_credentials_string"])
		assert.Equal(t, "en-US", p["language"])
		assert.Equal(t, "long", p["model"])
	})

	t.Run("Amazon", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithStt(vendors.NewAmazonSTT(vendors.AmazonSTTOptions{
				AccessKey: "ak",
				SecretKey: "sk",
				Region:    "us-east-1",
				Language:  "en-US",
			}))
		props, err := agent.ToPropertiesMap(asrOpts())
		require.NoError(t, err)
		asr := props["asr"].(map[string]interface{})
		assert.Equal(t, "amazon", asr["vendor"])
		p := asr["params"].(map[string]interface{})
		assert.Equal(t, "ak", p["access_key_id"])
		assert.Equal(t, "sk", p["secret_access_key"])
		assert.Equal(t, "us-east-1", p["region"])
		assert.Equal(t, "en-US", p["language_code"])
	})

	t.Run("AssemblyAI", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithStt(vendors.NewAssemblyAISTT(vendors.AssemblyAISTTOptions{
				APIKey:   "aai-key",
				Language: "en-US",
				URI:      "wss://example.com/ws",
			}))
		props, err := agent.ToPropertiesMap(asrOpts())
		require.NoError(t, err)
		asr := props["asr"].(map[string]interface{})
		assert.Equal(t, "assemblyai", asr["vendor"])
		p := asr["params"].(map[string]interface{})
		assert.Equal(t, "aai-key", p["api_key"])
		assert.Equal(t, "en-US", p["language"])
		assert.Equal(t, "wss://example.com/ws", p["uri"])
	})

	t.Run("Ares", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithStt(vendors.NewAresSTT(vendors.AresSTTOptions{}))
		props, err := agent.ToPropertiesMap(asrOpts())
		require.NoError(t, err)
		asr := props["asr"].(map[string]interface{})
		assert.Equal(t, "ares", asr["vendor"])
		// Ares has no params key when empty
		assert.NotContains(t, asr, "params")
	})

	t.Run("Speechmatics", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithStt(vendors.NewSpeechmaticsSTT(vendors.SpeechmaticsSTTOptions{
				APIKey:   "sm-key",
				Language: "en",
			}))
		props, err := agent.ToPropertiesMap(asrOpts())
		require.NoError(t, err)
		asr := props["asr"].(map[string]interface{})
		assert.Equal(t, "speechmatics", asr["vendor"])
		p := asr["params"].(map[string]interface{})
		assert.Equal(t, "sm-key", p["api_key"])
		assert.Equal(t, "en", p["language"])
	})

	t.Run("Sarvam", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithStt(vendors.NewSarvamSTT(vendors.SarvamSTTOptions{
				APIKey:   "sv-key",
				Language: "en-IN",
			}))
		props, err := agent.ToPropertiesMap(asrOpts())
		require.NoError(t, err)
		asr := props["asr"].(map[string]interface{})
		assert.Equal(t, "sarvam", asr["vendor"])
		p := asr["params"].(map[string]interface{})
		assert.Equal(t, "sv-key", p["api_key"])
		assert.Equal(t, "en-IN", p["language"])
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// BYOK LLM Vendor Shapes
// ─────────────────────────────────────────────────────────────────────────────

func TestBYOKLLMVendorShapes(t *testing.T) {
	llmOpts := func() ToPropertiesOptions {
		opts := basePropertiesOpts()
		opts.AllowMissingVendorCategories = []string{"asr", "tts"}
		return opts
	}

	t.Run("OpenAI", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
				APIKey:  "openai-key",
				BaseURL: "https://api.openai.com/v1/chat/completions",
				Model:   "gpt-4o",
			}))
		props, err := agent.ToPropertiesMap(llmOpts())
		require.NoError(t, err)
		llm := props["llm"].(map[string]interface{})
		assert.Equal(t, "openai-key", llm["api_key"])
		assert.Equal(t, "https://api.openai.com/v1/chat/completions", llm["url"])
		assert.Equal(t, "openai", llm["style"])
		assert.Equal(t, "gpt-4o", llm["params"].(map[string]interface{})["model"])
	})

	t.Run("AzureOpenAI", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithLlm(vendors.NewAzureOpenAI(vendors.AzureOpenAIOptions{
				APIKey:         "az-key",
				Endpoint:       "https://myres.openai.azure.com",
				DeploymentName: "gpt-4o",
				Model:          "gpt-4o",
			}))
		props, err := agent.ToPropertiesMap(llmOpts())
		require.NoError(t, err)
		llm := props["llm"].(map[string]interface{})
		assert.Equal(t, "az-key", llm["api_key"])
		assert.Equal(t, "openai", llm["style"])
		assert.Equal(t, "azure", llm["vendor"])
		url := llm["url"].(string)
		assert.Contains(t, url, "myres.openai.azure.com")
		assert.Contains(t, url, "gpt-4o")
	})

	t.Run("Anthropic", func(t *testing.T) {
		maxTok := 1024
		agent := NewAgent(WithName("a")).
			WithLlm(vendors.NewAnthropic(vendors.AnthropicOptions{
				APIKey:    "ant-key",
				Model:     "claude-3-opus-20240229",
				URL:       "https://api.anthropic.com/v1/messages",
				MaxTokens: &maxTok,
				Headers:   map[string]string{"anthropic-version": "2023-06-01"},
			}))
		props, err := agent.ToPropertiesMap(llmOpts())
		require.NoError(t, err)
		llm := props["llm"].(map[string]interface{})
		assert.Equal(t, "ant-key", llm["api_key"])
		assert.Equal(t, "anthropic", llm["style"])
		assert.Equal(t, "https://api.anthropic.com/v1/messages", llm["url"])
		headers := llm["headers"].(map[string]string)
		assert.Equal(t, "2023-06-01", headers["anthropic-version"])
		params := llm["params"].(map[string]interface{})
		assert.Equal(t, "claude-3-opus-20240229", params["model"])
		assert.Equal(t, 1024, params["max_tokens"])
	})

	t.Run("Gemini", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithLlm(vendors.NewGemini(vendors.GeminiOptions{
				APIKey: "gem-key",
				Model:  "gemini-1.5-flash",
			}))
		props, err := agent.ToPropertiesMap(llmOpts())
		require.NoError(t, err)
		llm := props["llm"].(map[string]interface{})
		assert.Equal(t, "gem-key", llm["api_key"])
		assert.Equal(t, "gemini", llm["style"])
		assert.Equal(t, "gemini-1.5-flash", llm["params"].(map[string]interface{})["model"])
	})

	t.Run("Groq", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithLlm(vendors.NewGroq(vendors.GroqOptions{
				APIKey:  "groq-key",
				Model:   "llama-3.3-70b-versatile",
				BaseURL: "https://api.groq.com/openai/v1/chat/completions",
			}))
		props, err := agent.ToPropertiesMap(llmOpts())
		require.NoError(t, err)
		llm := props["llm"].(map[string]interface{})
		assert.Equal(t, "groq-key", llm["api_key"])
		assert.Equal(t, "openai", llm["style"])
		assert.Equal(t, "llama-3.3-70b-versatile", llm["params"].(map[string]interface{})["model"])
	})

	t.Run("CustomLLM", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithLlm(vendors.NewCustomLLM(vendors.CustomLLMOptions{
				APIKey:  "custom-key",
				BaseURL: "https://llm.example.com/chat",
				Model:   "my-model",
			}))
		props, err := agent.ToPropertiesMap(llmOpts())
		require.NoError(t, err)
		llm := props["llm"].(map[string]interface{})
		assert.Equal(t, "custom-key", llm["api_key"])
		assert.Equal(t, "custom", llm["vendor"])
		assert.Equal(t, "openai", llm["style"])
	})

	t.Run("VertexAILLM", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithLlm(vendors.NewVertexAILLM(vendors.VertexAILLMOptions{
				GeminiOptions: vendors.GeminiOptions{
					APIKey: "vtx-key",
					Model:  "gemini-2.0-flash",
				},
				ProjectID: "my-project",
				Location:  "us-central1",
			}))
		props, err := agent.ToPropertiesMap(llmOpts())
		require.NoError(t, err)
		llm := props["llm"].(map[string]interface{})
		assert.Equal(t, "gemini", llm["style"])
		params := llm["params"].(map[string]interface{})
		assert.NotContains(t, params, "project_id")
		assert.NotContains(t, params, "location")
		assert.Contains(t, llm["url"].(string), "my-project")
		assert.Contains(t, llm["url"].(string), "us-central1")
	})

	t.Run("AmazonBedrock", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithLlm(vendors.NewAmazonBedrock(vendors.AmazonBedrockOptions{
				AccessKey: "ak",
				SecretKey: "sk",
				Region:    "us-east-1",
				Model:     "anthropic.claude-3",
			}))
		props, err := agent.ToPropertiesMap(llmOpts())
		require.NoError(t, err)
		llm := props["llm"].(map[string]interface{})
		assert.Equal(t, "ak", llm["access_key"])
		assert.Equal(t, "sk", llm["secret_key"])
		assert.Equal(t, "us-east-1", llm["region"])
		assert.Equal(t, "bedrock", llm["style"])
	})

	t.Run("Dify", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithLlm(vendors.NewDify(vendors.DifyOptions{
				APIKey: "dify-key",
				URL:    "https://dify.example.com",
				Model:  "gpt-4o",
			}))
		props, err := agent.ToPropertiesMap(llmOpts())
		require.NoError(t, err)
		llm := props["llm"].(map[string]interface{})
		assert.Equal(t, "dify-key", llm["api_key"])
		assert.Equal(t, "dify", llm["style"])
		assert.Equal(t, "gpt-4o", llm["params"].(map[string]interface{})["model"])
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// BYOK TTS Vendor Shapes
// ─────────────────────────────────────────────────────────────────────────────

func TestBYOKTTSVendorShapes(t *testing.T) {
	ttsOpts := func() ToPropertiesOptions {
		opts := basePropertiesOpts()
		opts.AllowMissingVendorCategories = []string{"asr", "tts"}
		return opts
	}
	agentWithTTS := func(tts vendors.TTS) *Agent {
		return NewAgent(WithName("a")).
			WithLlm(stubLLM).
			WithTts(tts)
	}

	t.Run("ElevenLabs", func(t *testing.T) {
		agent := agentWithTTS(vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
			Key:     "el-key",
			ModelID: "eleven_flash_v2_5",
			VoiceID: "some-voice",
			BaseURL: "wss://api.elevenlabs.io/v1",
		}))
		props, err := agent.ToPropertiesMap(ttsOpts())
		require.NoError(t, err)
		tts := props["tts"].(map[string]interface{})
		assert.Equal(t, "elevenlabs", tts["vendor"])
		p := tts["params"].(map[string]interface{})
		assert.Equal(t, "el-key", p["key"])
		assert.Equal(t, "eleven_flash_v2_5", p["model_id"])
		assert.Equal(t, "some-voice", p["voice_id"])
		assert.Equal(t, "wss://api.elevenlabs.io/v1", p["base_url"])
	})

	t.Run("MicrosoftTTS", func(t *testing.T) {
		agent := agentWithTTS(vendors.NewMicrosoftTTS(vendors.MicrosoftTTSOptions{
			Key:       "ms-key",
			Region:    "eastus",
			VoiceName: "en-US-JennyNeural",
		}))
		props, err := agent.ToPropertiesMap(ttsOpts())
		require.NoError(t, err)
		tts := props["tts"].(map[string]interface{})
		assert.Equal(t, "microsoft", tts["vendor"])
		p := tts["params"].(map[string]interface{})
		assert.Equal(t, "ms-key", p["key"])
		assert.Equal(t, "eastus", p["region"])
		assert.Equal(t, "en-US-JennyNeural", p["voice_name"])
	})

	t.Run("OpenAITTS_BYOK", func(t *testing.T) {
		agent := agentWithTTS(vendors.NewOpenAITTS(vendors.OpenAITTSOptions{
			APIKey:  "tts-key",
			BaseURL: "https://api.openai.com/v1/audio/speech",
			Model:   "tts-1-hd",
			Voice:   "alloy",
		}))
		props, err := agent.ToPropertiesMap(ttsOpts())
		require.NoError(t, err)
		tts := props["tts"].(map[string]interface{})
		assert.Equal(t, "openai", tts["vendor"])
		p := tts["params"].(map[string]interface{})
		assert.Equal(t, "tts-key", p["api_key"])
		assert.Equal(t, "tts-1-hd", p["model"])
		assert.Equal(t, "alloy", p["voice"])
		assert.Equal(t, "https://api.openai.com/v1/audio/speech", p["base_url"])
	})

	t.Run("Cartesia", func(t *testing.T) {
		agent := agentWithTTS(vendors.NewCartesiaTTS(vendors.CartesiaTTSOptions{
			APIKey:  "cart-key",
			VoiceID: "some-voice-id",
			ModelID: "sonic-english",
		}))
		props, err := agent.ToPropertiesMap(ttsOpts())
		require.NoError(t, err)
		tts := props["tts"].(map[string]interface{})
		assert.Equal(t, "cartesia", tts["vendor"])
		p := tts["params"].(map[string]interface{})
		assert.Equal(t, "cart-key", p["api_key"])
		voice := p["voice"].(map[string]interface{})
		assert.Equal(t, "id", voice["mode"])
		assert.Equal(t, "some-voice-id", voice["id"])
	})

	t.Run("GoogleTTS", func(t *testing.T) {
		agent := agentWithTTS(vendors.NewGoogleTTS(vendors.GoogleTTSOptions{
			Key:       "google-creds",
			VoiceName: "en-US-Wavenet-A",
		}))
		props, err := agent.ToPropertiesMap(ttsOpts())
		require.NoError(t, err)
		tts := props["tts"].(map[string]interface{})
		assert.Equal(t, "google", tts["vendor"])
		p := tts["params"].(map[string]interface{})
		assert.Equal(t, "google-creds", p["credentials"])
		voiceSelection := p["VoiceSelectionParams"].(map[string]interface{})
		assert.Equal(t, "en-US-Wavenet-A", voiceSelection["name"])
	})

	t.Run("AmazonTTS", func(t *testing.T) {
		agent := agentWithTTS(vendors.NewAmazonTTS(vendors.AmazonTTSOptions{
			AccessKey: "ak",
			SecretKey: "sk",
			Region:    "us-east-1",
			VoiceID:   "Joanna",
			Engine:    "neural",
		}))
		props, err := agent.ToPropertiesMap(ttsOpts())
		require.NoError(t, err)
		tts := props["tts"].(map[string]interface{})
		assert.Equal(t, "amazon", tts["vendor"])
		p := tts["params"].(map[string]interface{})
		assert.Equal(t, "ak", p["aws_access_key_id"])
		assert.Equal(t, "sk", p["aws_secret_access_key"])
		assert.Equal(t, "us-east-1", p["region_name"])
		assert.Equal(t, "Joanna", p["voice"])
	})

	t.Run("DeepgramTTS", func(t *testing.T) {
		agent := agentWithTTS(vendors.NewDeepgramTTS(vendors.DeepgramTTSOptions{
			APIKey: "dg-key",
			Model:  "aura-2-en-us",
		}))
		props, err := agent.ToPropertiesMap(ttsOpts())
		require.NoError(t, err)
		tts := props["tts"].(map[string]interface{})
		assert.Equal(t, "deepgram", tts["vendor"])
		p := tts["params"].(map[string]interface{})
		assert.Equal(t, "dg-key", p["api_key"])
		assert.Equal(t, "aura-2-en-us", p["model"])
	})

	t.Run("HumeAI", func(t *testing.T) {
		agent := agentWithTTS(vendors.NewHumeAITTS(vendors.HumeAITTSOptions{
			Key:      "hume-key",
			VoiceID:  "v1",
			Provider: "hume_ai",
		}))
		props, err := agent.ToPropertiesMap(ttsOpts())
		require.NoError(t, err)
		tts := props["tts"].(map[string]interface{})
		assert.Equal(t, "humeai", tts["vendor"])
		p := tts["params"].(map[string]interface{})
		assert.Equal(t, "hume-key", p["key"])
		assert.Equal(t, "v1", p["voice_id"])
		assert.Equal(t, "hume_ai", p["provider"])
	})

	t.Run("Rime", func(t *testing.T) {
		agent := agentWithTTS(vendors.NewRimeTTS(vendors.RimeTTSOptions{
			Key:     "rime-key",
			Speaker: "eva",
			ModelID: "mist",
		}))
		props, err := agent.ToPropertiesMap(ttsOpts())
		require.NoError(t, err)
		tts := props["tts"].(map[string]interface{})
		assert.Equal(t, "rime", tts["vendor"])
		p := tts["params"].(map[string]interface{})
		assert.Equal(t, "rime-key", p["api_key"])
		assert.Equal(t, "eva", p["speaker"])
		assert.Equal(t, "mist", p["modelId"])
	})

	t.Run("FishAudio", func(t *testing.T) {
		agent := agentWithTTS(vendors.NewFishAudioTTS(vendors.FishAudioTTSOptions{
			Key:         "fish-key",
			ReferenceID: "ref1",
			Backend:     "speech-1.6",
		}))
		props, err := agent.ToPropertiesMap(ttsOpts())
		require.NoError(t, err)
		tts := props["tts"].(map[string]interface{})
		assert.Equal(t, "fishaudio", tts["vendor"])
		p := tts["params"].(map[string]interface{})
		assert.Equal(t, "fish-key", p["api_key"])
		assert.Equal(t, "ref1", p["reference_id"])
		assert.Equal(t, "speech-1.6", p["backend"])
	})

	t.Run("MiniMax_BYOK", func(t *testing.T) {
		agent := agentWithTTS(vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
			Key:     "mm-key",
			GroupID: "g1",
			Model:   "speech-02-turbo",
			VoiceID: "English_captivating_female1",
			URL:     "https://api.minimax.io/v1/tts",
		}))
		props, err := agent.ToPropertiesMap(ttsOpts())
		require.NoError(t, err)
		tts := props["tts"].(map[string]interface{})
		assert.Equal(t, "minimax", tts["vendor"])
		p := tts["params"].(map[string]interface{})
		assert.Equal(t, "mm-key", p["key"])
		assert.Equal(t, "g1", p["group_id"])
	})

	t.Run("SarvamTTS", func(t *testing.T) {
		agent := agentWithTTS(vendors.NewSarvamTTS(vendors.SarvamTTSOptions{
			Key:                "sv-key",
			Speaker:            "meera",
			TargetLanguageCode: "hi-IN",
		}))
		props, err := agent.ToPropertiesMap(ttsOpts())
		require.NoError(t, err)
		tts := props["tts"].(map[string]interface{})
		assert.Equal(t, "sarvam", tts["vendor"])
		p := tts["params"].(map[string]interface{})
		assert.Equal(t, "sv-key", p["api_subscription_key"])
		assert.Equal(t, "meera", p["speaker"])
		assert.Equal(t, "hi-IN", p["target_language_code"])
	})

	t.Run("Murf", func(t *testing.T) {
		agent := agentWithTTS(vendors.NewMurfTTS(vendors.MurfTTSOptions{
			Key: "murf-key",
		}))
		props, err := agent.ToPropertiesMap(ttsOpts())
		require.NoError(t, err)
		tts := props["tts"].(map[string]interface{})
		assert.Equal(t, "murf", tts["vendor"])
		p := tts["params"].(map[string]interface{})
		assert.Equal(t, "murf-key", p["api_key"])
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// MLLM Vendor Shapes
// ─────────────────────────────────────────────────────────────────────────────

func TestMLLMVendorShapes(t *testing.T) {
	mllmOpts := basePropertiesOpts()

	t.Run("OpenAIRealtime", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
				APIKey: "rt-key",
				Model:  "gpt-4o-realtime-preview",
				Voice:  "coral",
			}))
		props, err := agent.ToPropertiesMap(mllmOpts)
		require.NoError(t, err)
		mllm := props["mllm"].(map[string]interface{})
		assert.Equal(t, "openai", mllm["vendor"])
		assert.Equal(t, true, mllm["enable"])
		assert.Equal(t, "rt-key", mllm["api_key"])
		params := mllm["params"].(map[string]interface{})
		assert.Equal(t, "gpt-4o-realtime-preview", params["model"])
		assert.Equal(t, "coral", params["voice"])
	})

	t.Run("GeminiLive", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithMllm(vendors.NewGeminiLive(vendors.GeminiLiveOptions{
				APIKey: "gl-key",
				Model:  "gemini-live-2.5-flash",
			}))
		props, err := agent.ToPropertiesMap(mllmOpts)
		require.NoError(t, err)
		mllm := props["mllm"].(map[string]interface{})
		assert.Equal(t, "gemini", mllm["vendor"])
		assert.Equal(t, true, mllm["enable"])
		assert.Equal(t, "gl-key", mllm["api_key"])
		params := mllm["params"].(map[string]interface{})
		assert.Equal(t, "gemini-live-2.5-flash", params["model"])
	})

	t.Run("VertexAI", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithMllm(vendors.NewVertexAI(vendors.VertexAIOptions{
				ProjectID:           "my-project",
				Location:            "us-central1",
				Model:               "gemini-live-2.5-flash",
				ADCredentialsString: `{"type":"service_account"}`,
			}))
		props, err := agent.ToPropertiesMap(mllmOpts)
		require.NoError(t, err)
		mllm := props["mllm"].(map[string]interface{})
		assert.Equal(t, "vertexai", mllm["vendor"])
		assert.Equal(t, true, mllm["enable"])
		assert.Equal(t, "my-project", mllm["project_id"])
		assert.Equal(t, "us-central1", mllm["location"])
		assert.Equal(t, `{"type":"service_account"}`, mllm["adc_credentials_string"])
		params := mllm["params"].(map[string]interface{})
		assert.Equal(t, "gemini-live-2.5-flash", params["model"])
	})

	t.Run("XaiGrok", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithMllm(vendors.NewXaiGrok(vendors.XaiGrokOptions{
				APIKey: "xai-key",
			}))
		props, err := agent.ToPropertiesMap(mllmOpts)
		require.NoError(t, err)
		mllm := props["mllm"].(map[string]interface{})
		assert.Equal(t, "xai", mllm["vendor"])
		assert.Equal(t, true, mllm["enable"])
		assert.Equal(t, "xai-key", mllm["api_key"])
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Preset Coverage Matrix
// ─────────────────────────────────────────────────────────────────────────────

func TestPresetCoverageMatrix(t *testing.T) {
	// For each preset, verify it's inferred and correct fields are stripped.

	t.Run("ASR_deepgram_nova_2", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
				Model:    "nova-2",
				Language: "en-US",
			})).
			WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-4o-mini"})).
			WithTts(vendors.NewOpenAITTS(vendors.OpenAITTSOptions{Voice: "alloy"}))

		payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
		require.NoError(t, err)

		preset := payload["preset"].(string)
		assert.Contains(t, preset, AgentPresets.Asr.DeepgramNova2)

		props := payload["properties"].(map[string]interface{})
		asr := props["asr"].(map[string]interface{})
		asrParams, hasParams := asr["params"].(map[string]interface{})
		if hasParams {
			assert.NotContains(t, asrParams, "key")
			assert.NotContains(t, asrParams, "model")
		}
	})

	t.Run("ASR_deepgram_nova_3", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
				Model:    "nova-3",
				Language: "en-US",
			})).
			WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-4o-mini"})).
			WithTts(vendors.NewOpenAITTS(vendors.OpenAITTSOptions{Voice: "alloy"}))

		payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
		require.NoError(t, err)

		preset := payload["preset"].(string)
		assert.Contains(t, preset, AgentPresets.Asr.DeepgramNova3)
	})

	t.Run("LLM_openai_gpt_4o_mini", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-4o-mini"})).
			WithTts(vendors.NewOpenAITTS(vendors.OpenAITTSOptions{Voice: "alloy"}))

		payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
		require.NoError(t, err)

		preset := payload["preset"].(string)
		assert.Contains(t, preset, AgentPresets.Llm.OpenAIGpt4oMini)

		props := payload["properties"].(map[string]interface{})
		llm := props["llm"].(map[string]interface{})
		assert.NotContains(t, llm, "api_key")
	})

	t.Run("LLM_openai_gpt_4_1_mini", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-4.1-mini"})).
			WithTts(vendors.NewOpenAITTS(vendors.OpenAITTSOptions{Voice: "alloy"}))

		payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
		require.NoError(t, err)

		preset := payload["preset"].(string)
		assert.Contains(t, preset, AgentPresets.Llm.OpenAIGpt41Mini)
	})

	t.Run("LLM_openai_gpt_5_nano", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-5-nano"})).
			WithTts(vendors.NewOpenAITTS(vendors.OpenAITTSOptions{Voice: "alloy"}))

		payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
		require.NoError(t, err)

		preset := payload["preset"].(string)
		assert.Contains(t, preset, AgentPresets.Llm.OpenAIGpt5Nano)
	})

	t.Run("LLM_openai_gpt_5_mini", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-5-mini"})).
			WithTts(vendors.NewOpenAITTS(vendors.OpenAITTSOptions{Voice: "alloy"}))

		payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
		require.NoError(t, err)

		preset := payload["preset"].(string)
		assert.Contains(t, preset, AgentPresets.Llm.OpenAIGpt5Mini)
	})

	t.Run("TTS_minimax_speech_2_6_turbo", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-4o-mini"})).
			WithTts(vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
				Model:   "speech-2.6-turbo",
				VoiceID: "English_captivating_female1",
			}))

		payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
		require.NoError(t, err)

		preset := payload["preset"].(string)
		assert.Contains(t, preset, AgentPresets.Tts.MiniMaxSpeech26Turbo)

		props := payload["properties"].(map[string]interface{})
		tts := props["tts"].(map[string]interface{})
		ttsParams := tts["params"].(map[string]interface{})
		assert.NotContains(t, ttsParams, "key")
		assert.NotContains(t, ttsParams, "model")
	})

	t.Run("TTS_minimax_speech_2_8_turbo", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-4o-mini"})).
			WithTts(vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
				Model:   "speech-2.8-turbo",
				VoiceID: "English_captivating_female1",
			}))

		payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
		require.NoError(t, err)

		preset := payload["preset"].(string)
		assert.Contains(t, preset, AgentPresets.Tts.MiniMaxSpeech28Turbo)
	})

	t.Run("TTS_openai_tts_1", func(t *testing.T) {
		agent := NewAgent(WithName("a")).
			WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-4o-mini"})).
			WithTts(vendors.NewOpenAITTS(vendors.OpenAITTSOptions{Voice: "alloy"}))

		payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
		require.NoError(t, err)

		preset := payload["preset"].(string)
		assert.Contains(t, preset, AgentPresets.Tts.OpenAITts1)

		props := payload["properties"].(map[string]interface{})
		tts := props["tts"].(map[string]interface{})
		ttsParams := tts["params"].(map[string]interface{})
		assert.NotContains(t, ttsParams, "api_key")
		assert.NotContains(t, ttsParams, "model")
		assert.Equal(t, "alloy", ttsParams["voice"])
	})
}

func TestExplicitMiniMaxPresetStripsInternalHint(t *testing.T) {
	// When the caller supplies the MiniMax TTS preset explicitly (not inferred),
	// the internal _minimax_preset_model hint set by MiniMaxTTS must still be removed.
	opts := basePipelineSessionOptions()
	opts.Preset = []string{AgentPresets.Tts.MiniMaxSpeech28Turbo}

	agent := NewAgent(WithName("a")).
		WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-4o-mini"})).
		WithTts(vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
			Model:   "speech-2.8-turbo",
			VoiceID: "English_captivating_female1",
		}))

	payload, err := startPresetValidationSession(t, agent, opts)
	require.NoError(t, err)

	preset := payload["preset"].(string)
	assert.Contains(t, preset, AgentPresets.Tts.MiniMaxSpeech28Turbo)

	props := payload["properties"].(map[string]interface{})
	tts := props["tts"].(map[string]interface{})
	assert.NotContains(t, tts, "_minimax_preset_model")
}
