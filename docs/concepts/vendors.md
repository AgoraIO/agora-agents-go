---
sidebar_position: 4
title: Vendors
description: Vendor catalog — LLM, TTS, STT, MLLM, and Avatar constructors with configuration structs.
---

# Vendors

The `agentkit/vendors` package provides constructor functions for all supported third-party vendors. Each vendor implements one of five interfaces: `LLM`, `TTS`, `STT`, `MLLM`, or `Avatar`.

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
| `NewOpenAI` | `OpenAIOptions` | `Model` for Agora-managed models; `APIKey`, `BaseURL`, `Model` for BYOK | — |
| `NewAzureOpenAI` | `AzureOpenAIOptions` | `APIKey`, `Model`, `Endpoint`, `DeploymentName` | — |
| `NewAnthropic` | `AnthropicOptions` | `APIKey`, `Model`, `URL`, `Headers`, `MaxTokens` | — |
| `NewGemini` | `GeminiOptions` | `APIKey`, `Model` | — |
| `NewGroq` | `GroqOptions` | `APIKey`, `Model`, `BaseURL` | — |
| `NewVertexAILLM` | `VertexAILLMOptions` | `APIKey`, `Model`, `ProjectID`, `Location` | — |
| `NewAmazonBedrock` | `AmazonBedrockOptions` | `AccessKey`, `SecretKey`, `Region`, `Model` | — |
| `NewDify` | `DifyOptions` | `APIKey`, `URL`, `Model` | — |
| `NewCustomLLM` | `CustomLLMOptions` | `APIKey`, `BaseURL`, `Model` | — |

<!-- snippet: fragment -->
```go
llm := vendors.NewOpenAI(vendors.OpenAIOptions{
    Model: "gpt-5-mini",
})

agent := agentkit.NewAgent(...).WithLlm(llm)
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
| `NewRimeTTS` | `RimeTTSOptions` | `Key`, `Speaker`, `ModelID` |
| `NewFishAudioTTS` | `FishAudioTTSOptions` | `Key`, `ReferenceID`, `Backend` |
| `NewGroqTTS` | `GroqTTSOptions` | `Key` |
| `NewMiniMaxTTS` | `MiniMaxTTSOptions` | `Model` for supported Agora-managed MiniMax models; `Key`, `GroupID`, `Model`, `VoiceID`, `URL` for BYOK |
| `NewDeepgramTTS` | `DeepgramTTSOptions` | `APIKey`, `Model` |
| `NewSarvamTTS` | `SarvamTTSOptions` | `APIKey` |

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

Use `TurnDetectionConfig.Language` for Agora interaction language; it defaults to `en-US`. STT vendor `Language` fields are serialized under `asr.params` using each provider's own format. Ares does not take a provider language option; AgentKit uses `TurnDetectionConfig.Language` for REST `asr.language`.

| Constructor | Options Struct | Required Fields |
|---|---|---|
| `NewSpeechmaticsSTT` | `SpeechmaticsSTTOptions` | `APIKey`, `Language` |
| `NewDeepgramSTT` | `DeepgramSTTOptions` | `APIKey` for BYOK; none for supported Agora-managed Deepgram models; `Keyterm?` |
| `NewMicrosoftSTT` | `MicrosoftSTTOptions` | `Key`, `Region`, `Language` |
| `NewOpenAISTT` | `OpenAISTTOptions` | `APIKey` |
| `NewGoogleSTT` | `GoogleSTTOptions` | `ProjectID`, `Location`, `ADCCredentialsString`, `Language` |
| `NewAmazonSTT` | `AmazonSTTOptions` | `AccessKey`, `SecretKey`, `Region`, `Language` |
| `NewAssemblyAISTT` | `AssemblyAISTTOptions` | `APIKey`, `Language` |
| `NewAresSTT` | `AresSTTOptions` | None |
| `NewSarvamSTT` | `SarvamSTTOptions` | `APIKey`, `Language` |

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
