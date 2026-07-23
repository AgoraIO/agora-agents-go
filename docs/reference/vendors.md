---
sidebar_position: 4
title: Vendors Reference
description: Complete API reference for all vendor constructors and configuration structs.
---

# Vendors Reference

Package: `github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors`

This reference describes the default/global vendor package. Mainland China vendor constructors live under `github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors`.

## Package layout

- Global/default package: `github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors`
- CN package: `github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors`

## SampleRate

<!-- snippet: fragment -->
```go
type SampleRate int

const (
    SampleRate8kHz  SampleRate = 8000
    SampleRate16kHz SampleRate = 16000
    SampleRate22kHz SampleRate = 22050
    SampleRate24kHz SampleRate = 24000
    SampleRate44kHz SampleRate = 44100
    SampleRate48kHz SampleRate = 48000
)
```

## Interfaces

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

---

## LLM Vendors

### NewOpenAI

<!-- snippet: fragment -->
```go
func NewOpenAI(opts OpenAIOptions) *OpenAI
```

Panics if `Model` is empty. Panics if `APIKey` is empty unless `Model` is one of the supported Agora-managed global/default OpenAI models (`gpt-4o-mini`, `gpt-4.1-mini`, `gpt-5-nano`, `gpt-5-mini`) and `BaseURL` / `Vendor` are not set.

#### OpenAIOptions

| Field             | Type                       | Required | Default                                        | Description             |
| ----------------- | -------------------------- | -------- | ---------------------------------------------- | ----------------------- |
| `APIKey`          | `string`                   | BYOK only | —                                             | OpenAI API key. Optional for supported Agora-managed global/default OpenAI models. |
| `Model`           | `string`                   | Yes      | —                                              | Model identifier        |
| `BaseURL`         | `string`                   | BYOK only | —                                             | API endpoint. Required when `APIKey` is set. |
| `Temperature`     | `*float64`                 | No       | —                                              | Sampling temperature    |
| `TopP`            | `*float64`                 | No       | —                                              | Nucleus sampling        |
| `MaxTokens`       | `*int`                     | No       | —                                              | Max tokens in response  |
| `SystemMessages`  | `[]map[string]interface{}` | No       | —                                              | System messages         |
| `GreetingMessage` | `string`                   | No       | —                                              | Initial greeting        |
| `FailureMessage`  | `string`                   | No       | —                                              | Fallback on error       |
| `InputModalities` | `[]string`                 | No       | `["text"]`                                     | Input modality types    |
| `OutputModalities` | `[]string`                | No       | —                                              | Output modality types   |
| `Params`          | `map[string]interface{}`   | No       | —                                              | Additional model params |
| `Headers`         | `map[string]string`        | No       | —                                              | Custom HTTP headers forwarded to the LLM provider |
| `GreetingConfigs` | `map[string]interface{}`   | No       | —                                              | Greeting playback configuration |
| `TemplateVariables` | `map[string]string`      | No       | —                                              | Template variables for messages |

### NewAzureOpenAI

<!-- snippet: fragment -->
```go
func NewAzureOpenAI(opts AzureOpenAIOptions) *AzureOpenAI
```

Panics if `APIKey`, `Model`, `Endpoint`, or `DeploymentName` is empty.

#### AzureOpenAIOptions

| Field             | Type                       | Required | Default                | Description          |
| ----------------- | -------------------------- | -------- | ---------------------- | -------------------- |
| `APIKey`          | `string`                   | Yes      | —                      | Azure OpenAI API key |
| `Endpoint`        | `string`                   | Yes      | —                      | Azure endpoint URL   |
| `DeploymentName`  | `string`                   | Yes      | —                      | Deployment name      |
| `APIVersion`      | `string`                   | No       | `"2024-08-01-preview"` | API version          |
| `Model`           | `string`                   | Yes      | —                      | Deployment's base model name (e.g., `"gpt-4o"`). Emitted as `params.model` for parity with the TypeScript SDK. |
| `Temperature`     | `*float64`                 | No       | —                      | Sampling temperature |
| `TopP`            | `*float64`                 | No       | —                      | Nucleus sampling     |
| `MaxTokens`       | `*int`                     | No       | —                      | Max tokens           |
| `SystemMessages`  | `[]map[string]interface{}` | No       | —                      | System messages      |
| `GreetingMessage` | `string`                   | No       | —                      | Initial greeting     |
| `FailureMessage`  | `string`                   | No       | —                      | Fallback on error    |
| `InputModalities` | `[]string`                 | No       | `["text"]`             | Input modality types |
| `OutputModalities` | `[]string`                | No       | —                      | Output modality types |
| `Params`          | `map[string]interface{}`   | No       | —                      | Additional model params |
| `Headers`         | `map[string]string`        | No       | —                      | Custom HTTP headers forwarded to the LLM provider |
| `GreetingConfigs` | `map[string]interface{}`   | No       | —                      | Greeting playback configuration |
| `TemplateVariables` | `map[string]string`      | No       | —                      | Template variables for messages |

### NewAnthropic

<!-- snippet: fragment -->
```go
func NewAnthropic(opts AnthropicOptions) *Anthropic
```

Panics if `APIKey`, `Model`, `URL`, `Headers`, or `MaxTokens` is empty.

#### AnthropicOptions

| Field             | Type                       | Required | Default                        | Description          |
| ----------------- | -------------------------- | -------- | ------------------------------ | -------------------- |
| `APIKey`          | `string`                   | Yes      | —                              | Anthropic API key    |
| `Model`           | `string`                   | Yes      | —                              | Model identifier     |
| `URL`             | `string`                   | Yes      | —                              | Anthropic messages endpoint URL |
| `Headers`         | `map[string]string`        | Yes      | —                              | Request headers, including Anthropic API version |
| `MaxTokens`       | `*int`                     | Yes      | —                              | Max tokens           |
| `Temperature`     | `*float64`                 | No       | —                              | Sampling temperature |
| `TopP`            | `*float64`                 | No       | —                              | Nucleus sampling     |
| `SystemMessages`  | `[]map[string]interface{}` | No       | —                              | System messages      |
| `GreetingMessage` | `string`                   | No       | —                              | Initial greeting     |
| `FailureMessage`  | `string`                   | No       | —                              | Fallback on error    |
| `InputModalities` | `[]string`                 | No       | `["text"]`                     | Input modality types |
| `OutputModalities` | `[]string`                | No       | —                              | Output modality types |
| `Params`          | `map[string]interface{}`   | No       | —                              | Additional model params |
| `GreetingConfigs` | `map[string]interface{}`   | No       | —                              | Greeting playback configuration |
| `TemplateVariables` | `map[string]string`      | No       | —                              | Template variables for messages |

### NewGemini

<!-- snippet: fragment -->
```go
func NewGemini(opts GeminiOptions) *Gemini
```

Panics if `APIKey` or `Model` is empty.

#### GeminiOptions

| Field             | Type                       | Required | Default                  | Description          |
| ----------------- | -------------------------- | -------- | ------------------------ | -------------------- |
| `APIKey`          | `string`                   | Yes      | —                        | Google AI API key |
| `Model`           | `string`                   | Yes      | —                        | Model identifier     |
| `URL`             | `string`                   | No       | auto-generated           | Optional full Gemini `streamGenerateContent` URL override |
| `Temperature`     | `*float64`                 | No       | —                        | Sampling temperature |
| `TopP`            | `*float64`                 | No       | —                        | Nucleus sampling     |
| `TopK`            | `*int`                     | No       | —                        | Top-K sampling       |
| `MaxOutputTokens` | `*int`                     | No       | —                        | Max output tokens    |
| `SystemMessages`  | `[]map[string]interface{}` | No       | —                        | System messages      |
| `GreetingMessage` | `string`                   | No       | —                        | Initial greeting     |
| `FailureMessage`  | `string`                   | No       | —                        | Fallback on error    |
| `InputModalities` | `[]string`                 | No       | `["text"]`               | Input modality types |
| `OutputModalities` | `[]string`                | No       | —                        | Output modality types |
| `Params`          | `map[string]interface{}`   | No       | —                        | Additional model params |
| `Headers`         | `map[string]string`        | No       | —                        | Custom HTTP headers forwarded to the LLM provider |
| `GreetingConfigs` | `map[string]interface{}`   | No       | —                        | Greeting playback configuration |
| `TemplateVariables` | `map[string]string`      | No       | —                        | Template variables for messages |

### Other LLM vendors

The SDK also includes named helpers for the remaining Agora-supported LLM providers. These helpers choose the correct request format internally.

| Constructor | Options Struct | Required Fields |
|---|---|---|
| `NewGroq` | `GroqOptions` | `APIKey`, `Model`, `BaseURL` |
| `NewVertexAILLM` | `VertexAILLMOptions` | `APIKey`, `Model`, `ProjectID`, `Location` |
| `NewXaiLLM` | `XaiLLMOptions` | `APIKey`, `Model`, `BaseURL` |
| `NewAmazonBedrock` | `AmazonBedrockOptions` | `AccessKey`, `SecretKey`, `Region`, `Model` |
| `NewDify` | `DifyOptions` | `APIKey`, `URL`, `Model` |
| `NewCustomLLM` | `CustomLLMOptions` | `APIKey`, `BaseURL`, `Model` |

`VertexAILLMOptions` accepts an optional `URL` override. When omitted, AgentKit builds the Vertex AI `streamGenerateContent` endpoint from `ProjectID`, `Location`, and `Model`.

### NewXaiLLM

<!-- snippet: fragment -->
```go
func NewXaiLLM(opts XaiLLMOptions) *XaiLLM
```

Panics if `APIKey`, `Model`, or `BaseURL` is empty.

#### XaiLLMOptions

`XaiLLMOptions` is an alias of `OpenAIOptions`.

Use `NewXaiLLM` for xAI Grok chat-completions style LLM integrations in the global/default vendor package. It emits `vendor: "xai"` with OpenAI-compatible request formatting.

---

## TTS Vendors

### NewElevenLabsTTS

<!-- snippet: fragment -->
```go
func NewElevenLabsTTS(opts ElevenLabsTTSOptions) *ElevenLabsTTS
```

Panics if `Key`, `ModelID`, `VoiceID`, or `BaseURL` is empty.

#### ElevenLabsTTSOptions

| Field          | Type          | Required | Description                                    |
| -------------- | ------------- | -------- | ---------------------------------------------- |
| `Key`          | `string`      | Yes      | ElevenLabs API key                             |
| `ModelID`      | `string`      | Yes      | Model identifier (e.g., `"eleven_turbo_v2_5"`) |
| `VoiceID`      | `string`      | Yes      | Voice identifier                               |
| `BaseURL`      | `string`      | Yes      | WebSocket base URL                             |
| `SampleRate`   | `*SampleRate` | No       | Output sample rate                             |
| `OptimizeStreamingLatency` | `*int` | No | Latency optimization level, 0-4 |
| `Stability` | `*float64` | No | Voice stability, 0.0-1.0 |
| `SimilarityBoost` | `*float64` | No | Voice similarity boost, 0.0-1.0 |
| `Style` | `*float64` | No | Voice style exaggeration, 0.0-1.0 |
| `UseSpeakerBoost` | `*bool` | No | Enable speaker boost |
| `SkipPatterns` | `[]int`       | No       | Patterns to skip in TTS output                 |

### NewMicrosoftTTS

<!-- snippet: fragment -->
```go
func NewMicrosoftTTS(opts MicrosoftTTSOptions) *MicrosoftTTS
```

Panics if `Key`, `Region`, or `VoiceName` is empty.

#### MicrosoftTTSOptions

| Field          | Type          | Required | Description                              |
| -------------- | ------------- | -------- | ---------------------------------------- |
| `Key`          | `string`      | Yes      | Azure Speech Services key                |
| `Region`       | `string`      | Yes      | Azure region (e.g., `"eastus"`)          |
| `VoiceName`    | `string`      | Yes      | Voice name (e.g., `"en-US-JennyNeural"`) |
| `SampleRate`   | `*SampleRate` | No       | Output sample rate                       |
| `Speed`        | `*float64`    | No       | Speaking rate multiplier                 |
| `Volume`       | `*float64`    | No       | Audio volume                             |
| `SkipPatterns` | `[]int`       | No       | Patterns to skip                         |

### NewOpenAITTS

<!-- snippet: fragment -->
```go
func NewOpenAITTS(opts OpenAITTSOptions) *OpenAITTS
```

Panics if `Voice` is empty. `APIKey`, `Model`, and `BaseURL` are required together for BYOK. `APIKey` is optional for the Agora-managed global/default `tts-1` path. Always returns `SampleRate24kHz` from `GetSampleRate()`.

#### OpenAITTSOptions

| Field            | Type       | Required | Description                        |
| ---------------- | ---------- | -------- | ---------------------------------- |
| `APIKey`         | `string`   | BYOK only | OpenAI API key. Optional for the Agora-managed global/default `tts-1` path. |
| `Voice`          | `string`   | Yes      | Voice name                         |
| `Model`          | `string`   | BYOK only | Model identifier                   |
| `BaseURL`        | `string`   | BYOK only | OpenAI TTS endpoint URL            |
| `Instructions`   | `string`   | No       | Custom instructions for voice style, accent, pace, and tone |
| `Speed`          | `*float64` | No       | Speech speed multiplier            |
| `SkipPatterns`   | `[]int`    | No       | Patterns to skip                   |

### NewGenericTTS

<!-- snippet: fragment -->
```go
func NewGenericTTS(opts GenericTTSOptions) *GenericTTS
```

Available in both `github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors` and `github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors`. Panics if `URL` is empty, malformed, lacks a host, or uses a scheme other than HTTP or HTTPS. Supported URLs serialize with `vendor: "generic_http"`.

`AdditionalParams` values are copied into `tts.params` first. Non-empty named fields take precedence over matching additional parameters.

#### GenericTTSOptions

| Field | Type | Required | Description |
|---|---|---|---|
| `URL` | `string` | Yes | Absolute Generic OpenAI-compatible TTS endpoint using HTTP or HTTPS |
| `Headers` | `map[string]string` | No | Custom HTTP headers forwarded to the TTS service |
| `APIKey` | `string` | No | API key sent as `params.api_key` |
| `Model` | `string` | No | TTS model name |
| `Voice` | `string` | No | Voice name |
| `Speed` | `*float64` | No | Speech rate |
| `SampleRate` | `*SampleRate` | No | Output sample rate |
| `ResponseFormat` | `string` | No | Output audio format sent as `params.response_format`; currently `pcm` is supported |
| `Instruction` | `string` | No | Additional style instruction |
| `AdditionalParams` | `map[string]interface{}` | No | Additional provider parameters forwarded under `tts.params` |
| `SkipPatterns` | `[]int` | No | Patterns to skip |

### NewCartesiaTTS

<!-- snippet: fragment -->
```go
func NewCartesiaTTS(opts CartesiaTTSOptions) *CartesiaTTS
```

Panics if `APIKey`, `VoiceID`, or `ModelID` is empty.

#### CartesiaTTSOptions

| Field          | Type          | Required | Description                                          |
| -------------- | ------------- | -------- | ---------------------------------------------------- |
| `APIKey`       | `string`      | Yes      | Cartesia API key                                     |
| `VoiceID`      | `string`      | Yes      | Voice identifier (serialized as `{"mode":"id","id":"..."}`) |
| `ModelID`      | `string`      | Yes      | Model identifier                                     |
| `BaseURL`      | `string`      | No       | WebSocket URL for the Cartesia streaming API         |
| `Language`     | `string`      | No       | Target language for speech synthesis                 |
| `SampleRate`   | `*SampleRate` | No       | Output sample rate                                   |
| `SkipPatterns` | `[]int`       | No       | Patterns to skip                                     |

### NewGoogleTTS

<!-- snippet: fragment -->
```go
func NewGoogleTTS(opts GoogleTTSOptions) *GoogleTTS
```

Panics if `Key` or `VoiceName` is empty.

#### GoogleTTSOptions

| Field          | Type     | Required | Description          |
| -------------- | -------- | -------- | -------------------- |
| `Key`          | `string` | Yes      | Google Cloud API key |
| `VoiceName`    | `string` | Yes      | Voice name           |
| `LanguageCode` | `string` | No       | Language code        |
| `SampleRate`   | `*SampleRate` | No | Output sample rate |
| `SkipPatterns` | `[]int`  | No       | Patterns to skip     |

### NewAmazonTTS

<!-- snippet: fragment -->
```go
func NewAmazonTTS(opts AmazonTTSOptions) *AmazonTTS
```

Panics if `AccessKey`, `SecretKey`, `Region`, `VoiceID`, or `Engine` is empty.

#### AmazonTTSOptions

| Field          | Type     | Required | Description      |
| -------------- | -------- | -------- | ---------------- |
| `AccessKey`    | `string` | Yes      | AWS access key   |
| `SecretKey`    | `string` | Yes      | AWS secret key   |
| `Region`       | `string` | Yes      | AWS region       |
| `VoiceID`      | `string` | Yes      | Polly voice ID   |
| `Engine`       | `string` | Yes      | Polly engine type |
| `SkipPatterns` | `[]int`  | No       | Patterns to skip |

### NewDeepgramTTS

<!-- snippet: fragment -->
```go
func NewDeepgramTTS(opts DeepgramTTSOptions) *DeepgramTTS
```

Panics if `APIKey` or `Model` is empty.

#### DeepgramTTSOptions

| Field          | Type                     | Required | Description |
| -------------- | ------------------------ | -------- | ----------- |
| `APIKey`       | `string`                 | Yes      | Deepgram API key |
| `Model`        | `string`                 | Yes      | Deepgram TTS model (e.g., `"aura-2-thalia-en"`) |
| `BaseURL`      | `string`                 | No       | WebSocket endpoint; defaults server-side to `wss://api.deepgram.com/v1/speak` |
| `SampleRate`   | `*SampleRate`            | No       | Output sample rate |
| `AdditionalParams` | `map[string]interface{}` | No       | Additional Deepgram TTS parameters, flattened into `params` |
| `SkipPatterns` | `[]int`                  | No       | Patterns to skip |

### NewHumeAITTS

<!-- snippet: fragment -->
```go
func NewHumeAITTS(opts HumeAITTSOptions) *HumeAITTS
```

Panics if `Key`, `VoiceID`, or `Provider` is empty.

#### HumeAITTSOptions

| Field          | Type     | Required | Description      |
| -------------- | -------- | -------- | ---------------- |
| `Key`          | `string` | Yes      | Hume AI API key  |
| `VoiceID`      | `string` | Yes      | Hume AI voice ID |
| `Provider`     | `string` | Yes      | Voice provider type, such as `CUSTOM_VOICE` or `HUME_AI` |
| `ConfigID`     | `string` | No       | Configuration ID |
| `BaseURL`      | `string` | No       | Base URL         |
| `Speed`        | `*float64` | No     | Playback speed |
| `TrailingSilence` | `*float64` | No | Trailing silence in seconds |
| `SkipPatterns` | `[]int`  | No       | Patterns to skip |

### NewRimeTTS

<!-- snippet: fragment -->
```go
func NewRimeTTS(opts RimeTTSOptions) *RimeTTS
```

When `CredentialMode` is `managed`, panics if `BaseURL` or `ModelID` is empty. When `CredentialMode` is omitted or set to `byok`, panics if `Key`, `Speaker`, or `ModelID` is empty. Panics if `CredentialMode` is any other value.

#### RimeTTSOptions

| Field            | Type             | Required     | Description                                   |
| ---------------- | ---------------- | ------------ | --------------------------------------------- |
| `CredentialMode` | `CredentialMode` | No           | Shared credential mode. Use `agentkit.CredentialModeManaged` or `agentkit.CredentialModeBYOK`. Defaults to BYOK behavior when omitted. |
| `Key`            | `string`         | BYOK only    | Rime API key                                  |
| `Speaker`        | `string`         | BYOK only    | Speaker identifier                            |
| `ModelID`        | `string`         | Yes          | Model identifier                              |
| `BaseURL`        | `string`         | Managed mode | WebSocket URL. Optional in BYOK mode.         |
| `SkipPatterns`   | `[]int`          | No           | Patterns to skip                              |

### NewFishAudioTTS

<!-- snippet: fragment -->
```go
func NewFishAudioTTS(opts FishAudioTTSOptions) *FishAudioTTS
```

Panics if `Key`, `ReferenceID`, or `Backend` is empty.

#### FishAudioTTSOptions

| Field          | Type     | Required | Description        |
| -------------- | -------- | -------- | ------------------ |
| `Key`          | `string` | Yes      | FishAudio API key  |
| `ReferenceID`  | `string` | Yes      | Reference audio ID |
| `Backend`      | `string` | Yes      | Backend model version |
| `SkipPatterns` | `[]int`  | No       | Patterns to skip   |

### NewMiniMaxTTS

<!-- snippet: fragment -->
```go
func NewMiniMaxTTS(opts MiniMaxTTSOptions) *MiniMaxTTS
```

Panics if `Model` is empty. `Key` is optional for supported Agora-managed global/default MiniMax models (`speech-2.6-turbo`, `speech_2_6_turbo`, `speech-2.8-turbo`, `speech_2_8_turbo`). BYOK requires `Key`, `GroupID`, `VoiceID`, and `URL`. In Agora-managed mode, `GroupID`, `VoiceID`, and `URL` are optional overrides.

#### MiniMaxTTSOptions

| Field          | Type     | Required | Description                               |
| -------------- | -------- | -------- | ----------------------------------------- |
| `Key`          | `string` | No       | MiniMax API key. Optional for supported Agora-managed global/default MiniMax models. |
| `GroupID`      | `string` | No       | MiniMax group ID. Required for BYOK.      |
| `Model`        | `string` | Yes      | Model name (e.g., `speech-02-turbo`)      |
| `VoiceID`      | `string` | No       | Voice style identifier. Required for BYOK; optional override for Agora-managed mode. |
| `URL`          | `string` | No       | WebSocket endpoint. Required for BYOK; optional override for Agora-managed mode. |
| `SkipPatterns` | `[]int`  | No       | Patterns to skip                          |

### NewMurfTTS

<!-- snippet: fragment -->
```go
func NewMurfTTS(opts MurfTTSOptions) *MurfTTS
```

Panics if `Key` is empty.

#### MurfTTSOptions

| Field          | Type     | Required | Description                              |
| -------------- | -------- | -------- | ---------------------------------------- |
| `Key`          | `string` | Yes      | Murf API key                             |
| `VoiceID`      | `string` | No       | Voice ID (e.g., `Ariana`, `Natalie`)     |
| `BaseURL`      | `string` | No       | WebSocket endpoint                       |
| `Locale`       | `string` | No       | Voice locale                             |
| `Rate`         | `*float64` | No      | Speech rate                              |
| `Pitch`        | `*float64` | No      | Pitch adjustment                         |
| `Model`        | `string` | No       | TTS model                                |
| `SampleRate`   | `*int`   | No       | Audio sample rate                        |
| `SkipPatterns` | `[]int`  | No       | Patterns to skip                         |

### NewSarvamTTS

<!-- snippet: fragment -->
```go
func NewSarvamTTS(opts SarvamTTSOptions) *SarvamTTS
```

Panics if `Key`, `Speaker`, or `TargetLanguageCode` is empty.

#### SarvamTTSOptions

| Field                | Type     | Required | Description          |
| -------------------- | -------- | -------- | -------------------- |
| `Key`                | `string` | Yes      | Sarvam API key       |
| `Speaker`            | `string` | Yes      | Speaker name         |
| `TargetLanguageCode` | `string` | Yes      | Target language code |
| `Pitch`              | `*float64` | No     | Pitch adjustment |
| `Pace`               | `*float64` | No     | Speed of speech |
| `Loudness`           | `*float64` | No     | Volume level |
| `SampleRate`         | `*int`   | No       | Audio sample rate |
| `SkipPatterns`       | `[]int`  | No       | Patterns to skip     |

### NewXaiTTS

<!-- snippet: fragment -->
```go
func NewXaiTTS(opts XaiTTSOptions) *XaiTTS
```

Panics if `APIKey` or `Language` is empty.

#### XaiTTSOptions

| Field | Type | Required | Description |
|---|---|---|---|
| `APIKey` | `string` | Yes | xAI API key |
| `VoiceID` | `string` | No | xAI voice identifier |
| `Language` | `string` | Yes | BCP-47 language code |
| `SampleRate` | `*SampleRate` | No | Output sample rate |
| `SkipPatterns` | `[]int` | No | Patterns to skip |

### NewGradiumTTS

<!-- snippet: fragment -->
```go
func NewGradiumTTS(opts GradiumTTSOptions) *GradiumTTS
```

Panics if `APIKey` is empty. Available only in `github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors`.

`AdditionalParams` values are copied into `tts.params` first. Non-empty named fields take precedence over matching additional parameters.

#### GradiumTTSOptions

| Field | Type | Required | Description |
|---|---|---|---|
| `APIKey` | `string` | Yes | Gradium API key |
| `URL` | `string` | No | WebSocket endpoint for streaming TTS output |
| `ModelName` | `string` | No | Gradium TTS model name |
| `VoiceID` | `string` | No | Gradium voice identifier |
| `SampleRate` | `*SampleRate` | No | Output sample rate |
| `AdditionalParams` | `map[string]interface{}` | No | Additional provider parameters forwarded under `tts.params` |
| `SkipPatterns` | `[]int` | No | Patterns to skip |

### NewMistralTTS

<!-- snippet: fragment -->
```go
func NewMistralTTS(opts MistralTTSOptions) *MistralTTS
```

Panics if `APIKey` is empty. Available only in `github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors`.

`AdditionalParams` values are copied into `tts.params` first. Non-empty named fields take precedence over matching additional parameters.

#### MistralTTSOptions

| Field | Type | Required | Description |
|---|---|---|---|
| `APIKey` | `string` | Yes | Mistral API key |
| `Model` | `string` | No | Mistral TTS model name |
| `Voice` | `string` | No | Mistral voice identifier |
| `AdditionalParams` | `map[string]interface{}` | No | Additional provider parameters forwarded under `tts.params` |
| `SkipPatterns` | `[]int` | No | Patterns to skip |

---

## STT Vendors

### NewSpeechmaticsSTT

<!-- snippet: fragment -->
```go
func NewSpeechmaticsSTT(opts SpeechmaticsSTTOptions) *SpeechmaticsSTT
```

Panics if `APIKey` or `Language` is empty.

#### SpeechmaticsSTTOptions

| Field | Type | Required | Description |
|---|---|---|---|
| `APIKey` | `string` | Yes | Speechmatics API key |
| `Language` | `string` | Yes | Speechmatics language code |
| `Model` | `string` | No | Model identifier |
| `URI` | `string` | No | Speechmatics streaming WebSocket URL |
| `AdditionalParams` | `map[string]interface{}` | No | Additional vendor params |

### NewDeepgramSTT

<!-- snippet: fragment -->
```go
func NewDeepgramSTT(opts DeepgramSTTOptions) *DeepgramSTT
```

Panics if `APIKey` is empty unless `Model` is one of the supported Agora-managed global/default Deepgram models (`nova-2`, `nova-3`).

#### DeepgramSTTOptions

| Field              | Type                     | Required | Description              |
| ------------------ | ------------------------ | -------- | ------------------------ |
| `APIKey`           | `string`                 | BYOK only | Deepgram API key. Optional only for Agora-managed global/default `nova-2` and `nova-3`. |
| `Model`            | `string`                 | No       | Model (e.g., `"nova-2"`) |
| `Language`         | `string`                 | No       | Language code            |
| `Keyterm`          | `string`                 | No       | Boost specialized terms and brands; serialized as `asr.params.keyterm` |
| `SmartFormat`      | `*bool`                  | No       | Enable smart formatting  |
| `Punctuation`      | `*bool`                  | No       | Enable punctuation       |
| `AdditionalParams` | `map[string]interface{}` | No       | Additional vendor params |

### NewMicrosoftSTT

<!-- snippet: fragment -->
```go
func NewMicrosoftSTT(opts MicrosoftSTTOptions) *MicrosoftSTT
```

Panics if `Key`, `Region`, or `Language` is empty.

#### MicrosoftSTTOptions

| Field      | Type     | Required | Description               |
| ---------- | -------- | -------- | ------------------------- |
| `Key`      | `string` | Yes      | Azure Speech Services key |
| `Region`   | `string` | Yes      | Azure region              |
| `Language` | `string` | Yes      | Language code             |

### NewOpenAISTT

<!-- snippet: fragment -->
```go
func NewOpenAISTT(opts OpenAISTTOptions) *OpenAISTT
```

Panics if `APIKey` is empty.

#### OpenAISTTOptions

| Field | Type | Required | Description |
|---|---|---|---|
| `APIKey` | `string` | Yes | OpenAI API key |
| `Model` | `string` | No | Transcription model |
| `Language` | `string` | No | OpenAI transcription language |
| `Prompt` | `string` | No | Prompt for OpenAI transcription |
| `InputAudioTranscription` | `map[string]interface{}` | No | OpenAI transcription settings |
| `AdditionalParams` | `map[string]interface{}` | No | Additional vendor params |

### NewGoogleSTT

<!-- snippet: fragment -->
```go
func NewGoogleSTT(opts GoogleSTTOptions) *GoogleSTT
```

Panics if `ProjectID`, `Location`, `ADCCredentialsString`, or `Language` is empty.

#### GoogleSTTOptions

| Field | Type | Required | Description |
|---|---|---|---|
| `ProjectID` | `string` | Yes | Google Cloud project ID |
| `Location` | `string` | Yes | Google Cloud region |
| `ADCCredentialsString` | `string` | Yes | Google service account credentials JSON string |
| `Language` | `string` | Yes | Google recognition language |
| `Model` | `string` | No | Model identifier |
| `AdditionalParams` | `map[string]interface{}` | No | Additional vendor params |

### NewAmazonSTT

<!-- snippet: fragment -->
```go
func NewAmazonSTT(opts AmazonSTTOptions) *AmazonSTT
```

Panics if `AccessKey`, `SecretKey`, `Region`, or `Language` is empty.

#### AmazonSTTOptions

| Field       | Type     | Required | Description    |
| ----------- | -------- | -------- | -------------- |
| `AccessKey` | `string` | Yes      | AWS access key |
| `SecretKey` | `string` | Yes      | AWS secret key |
| `Region`    | `string` | Yes      | AWS region     |
| `Language`  | `string` | Yes      | Amazon `language_code` |
| `AdditionalParams` | `map[string]interface{}` | No | Additional vendor params |

### NewAssemblyAISTT

<!-- snippet: fragment -->
```go
func NewAssemblyAISTT(opts AssemblyAISTTOptions) *AssemblyAISTT
```

Panics if `APIKey` or `Language` is empty.

#### AssemblyAISTTOptions

| Field | Type | Required | Description |
|---|---|---|---|
| `APIKey` | `string` | Yes | AssemblyAI API key |
| `Language` | `string` | Yes | AssemblyAI language code |
| `URI` | `string` | No | AssemblyAI streaming WebSocket URL |
| `AdditionalParams` | `map[string]interface{}` | No | Additional vendor params |

### NewSarvamSTT

<!-- snippet: fragment -->
```go
func NewSarvamSTT(opts SarvamSTTOptions) *SarvamSTT
```

Panics if `APIKey` is empty.

#### SarvamSTTOptions

| Field      | Type     | Required | Description      |
| ---------- | -------- | -------- | ---------------- |
| `APIKey`   | `string` | Yes      | Sarvam API key   |
| `Language` | `string` | No       | Language code    |
| `Model`    | `string` | No       | Model identifier |

### NewXaiSTT

<!-- snippet: fragment -->
```go
func NewXaiSTT(opts XaiSTTOptions) *XaiSTT
```

Panics if `APIKey` is empty.

#### XaiSTTOptions

| Field | Type | Required | Description |
|---|---|---|---|
| `APIKey` | `string` | Yes | xAI API key |
| `BaseURL` | `string` | No | xAI streaming STT WebSocket URL |
| `Language` | `string` | No | Recognition language |
| `SampleRate` | `*SampleRate` | No | Audio sample rate |
| `AdditionalParams` | `map[string]interface{}` | No | Additional vendor params |

---

## MLLM Vendors

### NewOpenAIRealtime

<!-- snippet: fragment -->
```go
func NewOpenAIRealtime(opts OpenAIRealtimeOptions) *OpenAIRealtime
```

Panics if `APIKey` is empty.

#### OpenAIRealtimeOptions

| Field             | Type       | Required | Default                     | Description                                        |
| ----------------- | ---------- | -------- | --------------------------- | -------------------------------------------------- |
| `APIKey`          | `string`   | Yes      | —                           | OpenAI API key                                     |
| `Model`           | `string`   | No       | `"gpt-4o-realtime-preview"` | Model identifier                                   |
| `URL`             | `string`   | No       | `"wss://api.openai.com/v1/realtime"` | Realtime WebSocket URL                  |
| `GreetingMessage` | `string`                   | No       | —                           | Initial greeting                                   |
| `FailureMessage`  | `string`                   | No       | —                           | Fallback message                                   |
| `InputModalities` | `[]string`                 | No       | —                           | Input modalities                                   |
| `OutputModalities` | `[]string`                | No       | —                           | Output modalities                                  |
| `Messages`        | `[]map[string]interface{}` | No       | —                           | Conversation messages for short-term memory        |
| `Params`          | `map[string]interface{}`   | No       | —                           | Additional realtime params such as `voice`         |
| `TurnDetection`   | `*Agora.MllmTurnDetection` | No | — | MLLM turn detection configuration; overrides top-level turn detection |

### NewGeminiLive

<!-- snippet: fragment -->
```go
func NewGeminiLive(opts GeminiLiveOptions) *GeminiLive
```

Panics if `APIKey` or `Model` is empty.

#### GeminiLiveOptions

| Field              | Type                       | Required | Default | Description |
| ------------------ | -------------------------- | -------- | ------- | ----------- |
| `APIKey`           | `string`                   | Yes      | —       | Google AI API key |
| `Model`            | `string`                   | Yes      | —       | Gemini Live model identifier |
| `URL`              | `string`                   | No       | —       | Custom realtime WebSocket URL |
| `Instructions`     | `string`                   | No       | —       | System instruction |
| `Voice`            | `string`                   | No       | —       | Voice name |
| `GreetingMessage`  | `string`                   | No       | —       | Initial greeting |
| `FailureMessage`   | `string`                   | No       | —       | Fallback message |
| `InputModalities`  | `[]string`                 | No       | —       | Input modalities |
| `OutputModalities` | `[]string`                 | No       | —       | Output modalities |
| `Messages`         | `[]map[string]interface{}` | No       | —       | Conversation messages |
| `AdditionalParams` | `map[string]interface{}`   | No       | —       | Additional Gemini params |
| `TurnDetection`    | `*Agora.MllmTurnDetection` | No | — | MLLM turn detection configuration; overrides top-level turn detection |

### NewXaiGrok

<!-- snippet: fragment -->
```go
func NewXaiGrok(opts XaiGrokOptions) *XaiGrok
```

xAI Grok MLLM vendor (`mllm.vendor`: `"xai"`). Panics if `APIKey` is empty. Defaults `URL` to `wss://api.x.ai/v1/realtime`.

> `NewXAIGrok` / `XAIGrokOptions` are deprecated aliases.

#### XaiGrokOptions

Same fields as `XAIGrokOptions` below.

### NewXAIGrok (deprecated)

<!-- snippet: fragment -->
```go
func NewXAIGrok(opts XAIGrokOptions) *XAIGrok
```

Deprecated. Use `NewXaiGrok` instead.

#### XAIGrokOptions

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `APIKey` | `string` | Yes | — | xAI API key |
| `URL` | `string` | No | `"wss://api.x.ai/v1/realtime"` | Realtime WebSocket URL |
| `Voice` | `string` | No | — | Voice identifier |
| `Language` | `string` | No | — | Language code |
| `SampleRate` | `*int` | No | — | Audio sample rate in Hz |
| `GreetingMessage` | `string` | No | — | Initial greeting |
| `FailureMessage` | `string` | No | — | Fallback message |
| `InputModalities` | `[]string` | No | — | Input modalities |
| `OutputModalities` | `[]string` | No | — | Output modalities |
| `Messages` | `[]map[string]interface{}` | No | — | Conversation messages |
| `Params` | `map[string]interface{}` | No | — | Additional xAI params |
| `TurnDetection` | `*Agora.MllmTurnDetection` | No | — | `agora_vad` / `server_vad` turn detection |

### NewVertexAI

<!-- snippet: fragment -->
```go
func NewVertexAI(opts VertexAIOptions) *VertexAI
```

 Panics if `ProjectID` or `ADCredentialsString` is empty.

#### VertexAIOptions

| Field             | Type       | Required | Default                  | Description                                     |
| ----------------- | ---------- | -------- | ------------------------ | ----------------------------------------------- |
| `ProjectID`       | `string`   | Yes      | —                        | GCP project ID                                  |
| `Location`        | `string`   | No       | `"us-central1"`          | GCP region                                      |
| `Model`           | `string`   | No       | `"gemini-2.0-flash-exp"` | Model identifier                                |
| `URL`             | `string`   | No       | —                        | Custom realtime WebSocket URL                   |
| `ADCredentialsString` | `string` | Yes     | —                        | ADC JSON credentials string                     |
| `Voice`           | `string`   | No       | —                        | Voice name                                      |
| `Instructions`    | `string`                   | No       | —                        | System instruction                              |
| `Messages`        | `[]map[string]interface{}` | No       | —                        | Conversation messages                           |
| `GreetingMessage` | `string`                   | No       | —                        | Initial greeting                                |
| `FailureMessage`  | `string`                   | No       | —                        | Fallback message                                |
| `InputModalities` | `[]string`                 | No       | —                        | Input modalities                                |
| `OutputModalities` | `[]string`                | No       | —                        | Output modalities                               |
| `AdditionalParams` | `map[string]interface{}`  | No       | —                        | Additional Vertex/Gemini params                 |
| `TurnDetection`    | `*Agora.MllmTurnDetection` | No | — | MLLM turn detection configuration; overrides top-level turn detection |

---

## Avatar Vendors

### NewLiveAvatarAvatar

<!-- snippet: fragment -->
```go
func NewLiveAvatarAvatar(opts LiveAvatarAvatarOptions) *LiveAvatarAvatar
```

Panics if `APIKey` or `AgoraUID` is empty, or if `Quality` is not one of `"low"`, `"medium"`, `"high"`.

Required TTS sample rate: **24kHz** (`SampleRate24kHz`)

#### LiveAvatarAvatarOptions

| Field                 | Type     | Required | Description                                      |
| --------------------- | -------- | -------- | ------------------------------------------------ |
| `APIKey`              | `string` | Yes      | LiveAvatar API key                               |
| `Quality`             | `string` | Yes      | `"low"`, `"medium"`, or `"high"`                 |
| `AgoraUID`            | `string` | Yes      | UID for avatar's video stream                    |
| `AgoraToken`          | `string` | No       | Avatar Agora token. Auto-generated when omitted. |
| `AvatarID`            | `string` | No       | LiveAvatar avatar ID                             |
| `Enable`              | `*bool`  | No       | Enable or disable the avatar (default: `true`)   |
| `DisableIdleTimeout`  | `*bool`  | No       | Disable the idle timeout                         |
| `ActivityIdleTimeout` | `*int`   | No       | Idle timeout in seconds (default: 120)           |

### NewAkoolAvatar

<!-- snippet: fragment -->
```go
func NewAkoolAvatar(opts AkoolAvatarOptions) *AkoolAvatar
```

Panics if `APIKey` is empty.

Required TTS sample rate: **16kHz** (`SampleRate16kHz`)

#### AkoolAvatarOptions

| Field              | Type                     | Required | Description |
| ------------------ | ------------------------ | -------- | ----------- |
| `APIKey`           | `string`                 | Yes      | Akool API key |
| `AvatarID`         | `string`                 | No       | Avatar ID |
| `Enable`           | `*bool`                  | No       | Enable or disable the avatar |
| `AdditionalParams` | `map[string]interface{}` | No       | Additional vendor params |

### NewHeyGenAvatar (deprecated)

<!-- snippet: fragment -->
```go
func NewHeyGenAvatar(opts HeyGenAvatarOptions) *HeyGenAvatar
```

Deprecated alias for the LiveAvatar-compatible avatar shape. Use `NewLiveAvatarAvatar` for new integrations.

### NewAnamAvatar

<!-- snippet: fragment -->
```go
func NewAnamAvatar(opts AnamAvatarOptions) *AnamAvatar
```

Panics if `APIKey` is empty.

#### AnamAvatarOptions

| Field | Type | Required | Description |
|---|---|---|---|
| `APIKey` | `string` | Yes | Anam API key |
| `AvatarID` | `string` | No | Anam avatar identifier (wire key: `params.avatar_id`) |
| `Enable` | `*bool` | No | Enable or disable the avatar |
| `AdditionalParams` | `map[string]interface{}` | No | Additional vendor params |

### NewGenericAvatar

<!-- snippet: fragment -->
```go
func NewGenericAvatar(opts GenericAvatarOptions) *GenericAvatar
```

Panics if `APIKey`, `APIBaseURL`, `AvatarID`, or `AgoraUID` is empty. `AgoraAppID`, `AgoraChannel`, and `AgoraToken` are optional; AgentKit fills them from the session on `Start()` when omitted.

Generic avatars do not enforce a fixed TTS sample rate. Use the sample rate required by your avatar provider.

#### GenericAvatarOptions

| Field | Type | Required | Description |
|---|---|---|---|
| `APIKey` | `string` | Yes | Generic avatar vendor API key |
| `APIBaseURL` | `string` | Yes | Generic avatar API endpoint |
| `AvatarID` | `string` | Yes | Avatar identifier |
| `AgoraUID` | `string` | Yes | UID for avatar video stream; use a different UID from `AgentUID` |
| `AgoraToken` | `string` | No | Avatar token; auto-generated with the same token format as agent tokens when omitted |
| `AgoraAppID` | `string` | No | Overrides session App ID |
| `AgoraChannel` | `string` | No | Overrides session channel |
| `Enable` | `*bool` | No | Enable or disable the avatar |
| `AdditionalParams` | `map[string]interface{}` | No | Additional vendor params |

## CN Avatar Vendors

CN avatar constructors live in `github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors` and are used with `agentkit/cn.Agent.WithAvatar`.

For CN STT, LLM, and TTS constructors, see [Vendors concept — CN Vendors](../concepts/vendors.md#cn-vendors-agentkitcnvendors).

### NewSensetimeAvatar

Package: `github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors`

Use with `agentkit/cn.Agent.WithAvatar`. See [Avatars Guide](../guides/avatars.md#sensetime-avatar-example-cn) for a full example.

<!-- snippet: fragment -->
```go
func NewSensetimeAvatar(opts SensetimeAvatarOptions) *SensetimeAvatar
```

Panics if `AgoraUID`, `AppID`, or `AppKey` is empty.

Required TTS sample rate: **not enforced** (`RequiredSampleRate()` returns `0`). Configure TTS according to your SenseTime deployment.

`AgoraToken` is optional; AgentKit auto-generates `agora_token` on `Start()` when it is omitted and `AppCertificate` is configured on the client.

#### SensetimeAvatarOptions

| Field | Type | Required | Description |
|---|---|---|---|
| `AgoraUID` | `string` | Yes | UID for avatar video stream; use a different UID from `AgentUID` |
| `AgoraToken` | `string` | No | Avatar Agora token; auto-generated when omitted |
| `AppID` | `string` | Yes | SenseTime app ID (wire key: `params.appId`) |
| `AppKey` | `string` | Yes | SenseTime app key |
| `SceneList` | `[]SensetimeScene` | No | Optional scene configuration with digital role settings |
| `Enable` | `*bool` | No | Enable or disable the avatar (default: `true`) |
| `AdditionalParams` | `map[string]interface{}` | No | Additional vendor params |

#### SensetimeScene

| Field | Type | Required | Description |
|---|---|---|---|
| `DigitalRole` | `SensetimeDigitalRole` | Yes | Digital role within the scene |

#### SensetimeDigitalRole

| Field | Type | Required | Description |
|---|---|---|---|
| `FaceFeatureID` | `string` | No | Face feature identifier (wire key: `params.sceneList[].digital_role.face_feature_id`) |
| `Position` | `SensetimePosition` | No | Avatar position (`x`, `y`) within the scene |
| `URL` | `string` | No | Avatar model package URL |

#### SensetimePosition

| Field | Type | Description |
|---|---|---|
| `X` | `int` | Horizontal position |
| `Y` | `int` | Vertical position |

### NewSpatiusAvatar

Package: `github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors`

Use with `agentkit/cn.Agent.WithAvatar`.

<!-- snippet: fragment -->
```go
func NewSpatiusAvatar(opts SpatiusAvatarOptions) *SpatiusAvatar
```

Panics if `SpatiusAPIKey`, `SpatiusAppID`, `SpatiusAvatarID`, or `AgoraUID` is empty.

Required TTS sample rate: **not enforced** (`RequiredSampleRate()` returns `0`). Configure TTS according to your Spatius deployment.

`AgoraToken` is optional; AgentKit auto-generates `agora_token` on `Start()` when it is omitted and `AppCertificate` is configured on the client.

#### SpatiusAvatarOptions

| Field | Type | Required | Description |
|---|---|---|---|
| `SpatiusAPIKey` | `string` | Yes | Spatius API key |
| `SpatiusAppID` | `string` | Yes | Spatius application ID |
| `SpatiusAvatarID` | `string` | Yes | Spatius avatar ID |
| `AgoraUID` | `string` | Yes | UID for avatar video stream; use a different UID from `AgentUID` |
| `AgoraToken` | `string` | No | Avatar Agora token; auto-generated when omitted |
| `Region` | `string` | No | Spatius service region, for example `cn-beijing` |
| `SampleRate` | `*SampleRate` | No | Avatar audio sample rate hint |
| `SessionExpireMinutes` | `*int` | No | Session validity duration in minutes |
| `Enable` | `*bool` | No | Enable or disable the avatar |
| `AdditionalParams` | `map[string]interface{}` | No | Additional vendor params |

---

## Sample Rate Constants

<!-- snippet: fragment -->
```go
const (
    LiveAvatarRequiredSampleRate = SampleRate24kHz
    AkoolRequiredSampleRate  = SampleRate16kHz  // 16000 Hz
)
```
