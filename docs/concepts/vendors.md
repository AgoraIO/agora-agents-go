---
sidebar_position: 4
title: Vendors
description: Vendor catalog — LLM, TTS, STT, MLLM, and Avatar constructors with configuration structs.
---

# Vendors

The default `agentkit/vendors` package provides constructor functions for the global/default vendor catalog. The `agentkit/cn/vendors` package provides constructors for mainland China providers.

## Package layout

- Global/default package: `github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors`
- CN package: `github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors`

## Vendor Interfaces

<!-- snippet: fragment -->
```go
type LLM interface {
    ToConfig() map[string]interface{}
}

type TTS interface {
    ToConfig() map[string]interface{}
    GetSampleRate() *SampleRate
}

type STT interface {
    ToConfig() map[string]interface{}
}

type MLLM interface {
    ToConfig() map[string]interface{}
}

type Avatar interface {
    ToConfig() map[string]interface{}
    RequiredSampleRate() SampleRate
}
```

## LLM Vendors

| Constructor | Options Struct | Required Fields | Default Model |
|---|---|---|---|
| `NewOpenAI` | `OpenAIOptions` | `Model` for Agora-managed global/default models; `APIKey`, `BaseURL`, `Model` for BYOK | — |
| `NewAzureOpenAI` | `AzureOpenAIOptions` | `APIKey`, `Model`, `Endpoint`, `DeploymentName` | — |
| `NewAnthropic` | `AnthropicOptions` | `APIKey`, `Model`, `URL`, `Headers`, `MaxTokens` | — |
| `NewGemini` | `GeminiOptions` | `APIKey`, `Model` | — |
| `NewGroq` | `GroqOptions` | `APIKey`, `Model`, `BaseURL` | — |
| `NewVertexAILLM` | `VertexAILLMOptions` | `APIKey`, `Model`, `ProjectID`, `Location` | — |
| `NewXaiLLM` | `XaiLLMOptions` | `APIKey`, `Model`, `BaseURL` | — |
| `NewAmazonBedrock` | `AmazonBedrockOptions` | `AccessKey`, `SecretKey`, `Region`, `Model` | — |
| `NewDify` | `DifyOptions` | `APIKey`, `URL`, `Model` | — |
| `NewCustomLLM` | `CustomLLMOptions` | `APIKey`, `BaseURL`, `Model` | — |

<!-- snippet: fragment -->
```go
llm := vendors.NewOpenAI(vendors.OpenAIOptions{
    Model: "gpt-5-mini",
})

agent := agentkit.NewAgent(client).WithLlm(llm)
```

## TTS Vendors

| Constructor | Options Struct | Required Fields |
|---|---|---|
| `NewElevenLabsTTS` | `ElevenLabsTTSOptions` | `Key`, `ModelID`, `VoiceID`, `BaseURL` |
| `NewMicrosoftTTS` | `MicrosoftTTSOptions` | `Key`, `Region`, `VoiceName` |
| `NewOpenAITTS` | `OpenAITTSOptions` | `Voice` for Agora-managed `tts-1`; `APIKey`, `Model`, `BaseURL`, `Voice` for BYOK |
| `NewCartesiaTTS` | `CartesiaTTSOptions` | `APIKey`, `VoiceID`, `ModelID` |
| `NewGoogleTTS` | `GoogleTTSOptions` | `Key`, `VoiceName` |
| `NewAmazonTTS` | `AmazonTTSOptions` | `AccessKey`, `SecretKey`, `Region`, `VoiceID`, `Engine` |
| `NewHumeAITTS` | `HumeAITTSOptions` | `Key`, `VoiceID`, `Provider` |
| `NewRimeTTS` | `RimeTTSOptions` | `BaseURL`, `ModelID` for managed credentials; `Key`, `Speaker`, `ModelID` for BYOK or omitted `CredentialMode` |
| `NewFishAudioTTS` | `FishAudioTTSOptions` | `Key`, `ReferenceID`, `Backend` |
| `NewMiniMaxTTS` | `MiniMaxTTSOptions` | `Model` for supported Agora-managed global/default MiniMax models; `Key`, `GroupID`, `Model`, `VoiceID`, `URL` for BYOK |
| `NewDeepgramTTS` | `DeepgramTTSOptions` | `APIKey`, `Model` |
| `NewGenericTTS` | `GenericTTSOptions` | `URL` |
| `NewSarvamTTS` | `SarvamTTSOptions` | `APIKey` |
| `NewMurfTTS` | `MurfTTSOptions` | `Key` |
| `NewXaiTTS` | `XaiTTSOptions` | `APIKey`, `Language` |
| `NewGradiumTTS` | `GradiumTTSOptions` | `APIKey` |
| `NewMistralTTS` | `MistralTTSOptions` | `APIKey` |
| `NewTypecastTTS` | `TypecastTTSOptions` | `APIKey`, `VoiceID`, `Model` |

`NewGenericTTS` is available in both `agentkit/vendors` and `agentkit/cn/vendors`. The required `URL` must be an absolute HTTP or HTTPS URL; other schemes are currently rejected. Supported URLs serialize with `vendor: "generic_http"`. All other fields are optional, and `AdditionalParams` forwards provider-specific values under `tts.params`; explicitly set named fields take precedence over matching additional parameters.

<!-- snippet: fragment -->
```go
tts := vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
    Key:        "<key>",
    ModelID:    "eleven_turbo_v2_5",
    VoiceID:    "<voice_id>",
    BaseURL:    "wss://api.elevenlabs.io/v1",
    SampleRate: &vendors.SampleRate24kHz,
})

agent = agent.WithTts(tts)
```

### SampleRate Constants

<!-- snippet: fragment -->
```go
vendors.SampleRate8kHz   // 8000
vendors.SampleRate16kHz  // 16000
vendors.SampleRate22kHz  // 22050
vendors.SampleRate24kHz  // 24000
vendors.SampleRate44kHz  // 44100
vendors.SampleRate48kHz  // 48000
```

Note: `OpenAITTS` always returns `SampleRate24kHz`. Other TTS vendors return their configured sample rate or `nil`.

## STT Vendors

Use `TurnDetectionConfig.Language` for Agora interaction language; it defaults to `en-US`. STT vendor `Language` fields are serialized under `asr.params` using each provider's own format. When global/default `agentkit` omits `WithStt()`, AgentKit falls back to `asr.vendor = "ares"`. Ares does not take a provider language option; AgentKit uses `TurnDetectionConfig.Language` for REST `asr.language`.

| Constructor | Options Struct | Required Fields |
|---|---|---|
| `NewSpeechmaticsSTT` | `SpeechmaticsSTTOptions` | `APIKey`, `Language` |
| `NewDeepgramSTT` | `DeepgramSTTOptions` | `APIKey` for BYOK; none for supported Agora-managed global/default Deepgram models; `Keyterm?` |
| `NewMicrosoftSTT` | `MicrosoftSTTOptions` | `Key`, `Region`, `Language` |
| `NewOpenAISTT` | `OpenAISTTOptions` | `APIKey` |
| `NewGoogleSTT` | `GoogleSTTOptions` | `ProjectID`, `Location`, `ADCCredentialsString`, `Language` |
| `NewAmazonSTT` | `AmazonSTTOptions` | `AccessKey`, `SecretKey`, `Region`, `Language` |
| `NewAssemblyAISTT` | `AssemblyAISTTOptions` | `APIKey`, `Language` |
| `NewAresSTT` | `AresSTTOptions` | None; optional `Keywords` |
| `NewSarvamSTT` | `SarvamSTTOptions` | `APIKey`, `Language` |
| `NewXaiSTT` | `XaiSTTOptions` | `APIKey` |

<!-- snippet: fragment -->
```go
stt := vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
    APIKey:   "<key>",
    Model:    "nova-2",
    Language: "en-US",
})

agent = agent.WithStt(stt)
```

## MLLM Vendors

| Constructor | Options Struct | Required Fields | Default Model |
|---|---|---|---|
| `NewOpenAIRealtime` | `OpenAIRealtimeOptions` | `APIKey` | `gpt-4o-realtime-preview` |
| `NewAzureOpenAIRealtime` (global) | `AzureOpenAIRealtimeOptions` | `APIKey`, `URL`, `TurnDetection` | — |
| `NewXaiGrok` | `XaiGrokOptions` | `APIKey`, `Model` | — |
| `NewGeminiLive` | `GeminiLiveOptions` | `APIKey`, `Model` | — |
| `NewVertexAI` | `VertexAIOptions` | `ProjectID` | `gemini-2.0-flash-exp` |

<!-- snippet: fragment -->
```go
mllm := vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
    APIKey: "<key>",
    Model:  "gpt-4o-realtime-preview",
    Params: map[string]interface{}{
        "voice": "alloy",
    },
})

agent = agent.WithMllm(mllm)
```

## Avatar Vendors

| Constructor | Options Struct | Required Fields | Required TTS Sample Rate |
|---|---|---|---|
| `NewLiveAvatarAvatar` | `LiveAvatarAvatarOptions` | `APIKey`, `Quality`, `AgoraUID` | 24kHz |
| `NewGenericAvatar` | `GenericAvatarOptions` | `APIKey`, `APIBaseURL`, `AvatarID`, `AgoraUID` | Provider-dependent |
| `NewAnamAvatar` | `AnamAvatarOptions` | `APIKey` | Provider-managed |
| `NewAkoolAvatar` | `AkoolAvatarOptions` | `APIKey` | 16kHz |
| `NewHeyGenAvatar` | `HeyGenAvatarOptions` | `APIKey`, `Quality`, `AgoraUID` | 24kHz; deprecated alias |
| `NewSensetimeAvatar` (CN) | `SensetimeAvatarOptions` | `AgoraUID`, `AppID`, `AppKey` | Not enforced; see [Avatars Guide](../guides/avatars.md) |
| `NewSpatiusAvatar` (CN) | `SpatiusAvatarOptions` | `SpatiusAPIKey`, `SpatiusAppID`, `SpatiusAvatarID`, `AgoraUID` | Not enforced; see [Avatars Guide](../guides/avatars.md) |

## CN Vendors (`agentkit/cn/vendors`)

Use with `agentkit/cn.Agent`. CN LLM constructors share `OpenAIOptions` shape and require `APIKey`, `Model`, and `BaseURL`.

### CN LLM

| Constructor | Wire `vendor` | Required fields |
|---|---|---|
| `NewAliyun` | `aliyun` | `APIKey`, `Model`, `BaseURL` |
| `NewBytedance` | `bytedance` | `APIKey`, `Model`, `BaseURL` |
| `NewDeepSeek` | `deepseek` | `APIKey`, `Model`, `BaseURL` |
| `NewTencentLLM` | `tencent` | `APIKey`, `Model`, `BaseURL` |

### CN MLLM

| Constructor | Wire `vendor` | Required fields | Default URL |
|---|---|---|---|
| `NewQwenOmni` | `qwen_omni` | `APIKey`, `Model`, `URL` | None |

### CN STT

REST `asr.language` comes from `TurnDetectionConfig.Language` (default `en-US`), not from CN STT constructors. When `agentkit/cn` omits `WithStt()`, AgentKit falls back to `asr.vendor = "fengming"`. Provider-specific language values go under `asr.params` when set.

| Constructor | Wire `vendor` | Required fields |
|---|---|---|
| `NewFengmingSTT()` / `NewFengmingSTT(FengmingSTTOptions)` | `fengming` | none; optional `Keywords` |
| `NewTencentSTT` | `tencent` | `Key`, `AppID`, `Secret` |
| `NewMicrosoftSTT` | `microsoft` | `Key`, `Region`, `Language` (in `params.language`) |
| `NewXfyunSTT` | `xfyun` | `APIKey`, `AppID`, `APISecret`; `Language` optional in `params` |
| `NewXfyunBigModelSTT` | `xfyun_bigmodel` | `APIKey`, `AppID`, `APISecret`; `Language` / `LanguageName` optional in `params` |
| `NewXfyunDialectSTT` | `xfyun_dialect` | `AppID`, `AccessKeyID`, `AccessKeySecret`; `Language` optional in `params` |

### CN TTS

All CN TTS option structs support `AdditionalParams` and `SkipPatterns`.

| Constructor | Wire `vendor` |
|---|---|
| `NewMiniMaxTTS` | `minimax` |
| `NewTencentTTS` | `tencent` |
| `NewMicrosoftTTS` | `microsoft` |
| `NewGenericTTS` | `generic_http` |
| `NewBytedanceTTS` | `bytedance` |
| `NewCosyVoiceTTS` | `cosyvoice` |
| `NewBytedanceDuplexTTS` | `bytedance_duplex` |
| `NewStepFunTTS` | `stepfun` |

### CN Avatar

| Constructor | Wire `vendor` | Required fields |
|---|---|---|
| `NewSensetimeAvatar` | `sensetime` | `AgoraUID`, `AppID`, `AppKey` |
| `NewSpatiusAvatar` | `spatius` | `SpatiusAPIKey`, `SpatiusAppID`, `SpatiusAvatarID`, `AgoraUID` |

See [CN AgentKit](../guides/cn-agentkit.md) and [Vendors Reference](../reference/vendors.md#cn-avatar-vendors) for examples.

<!-- snippet: fragment -->
```go
avatar := vendors.NewLiveAvatarAvatar(vendors.LiveAvatarAvatarOptions{
    APIKey:   "<key>",
    Quality:  "high",
    AgoraUID: "2001",
})

// TTS must be configured with matching sample rate BEFORE WithAvatar
agent = agent.WithTts(tts).WithAvatar(avatar)
```

See [Avatars Guide](../guides/avatars.md) for sample rate requirements and the panic behavior when rates mismatch.

## Validation

All vendor constructors validate required fields and `panic` if they are missing. For example:

<!-- snippet: fragment -->
```go
// This panics because gpt-4o requires an OpenAI API key.
vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-4o"})
```

This fails early so invalid vendor configuration does not reach session start. Provide BYOK fields when you use models outside the Agora-managed catalog.
