---
sidebar_position: 2
title: Agent
description: The Agent builder — functional options pattern, vendor chaining, and ToProperties conversion.
---

# Agent

The `agentkit.Agent` is the central configuration object. It defines what LLM, TTS, STT, MLLM, and avatar vendors your agent uses, along with session-level settings.

Every `Agent` is created with `agentkit.NewAgent(client, opts...)`. The `client` argument is required and must be a non-nil `*AgoraClient` from `NewAgoraClient`. Passing `nil` panics immediately with `NewAgent requires AgoraClient`.

Typical flow:

```go
client := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{ /* AppID, AppCertificate, Area */ })
agent := agentkit.NewAgent(client).WithLlm(/* ... */)
session := agent.CreateSession(agentkit.CreateSessionOptions{
    Name:        fmt.Sprintf("conversation-%d", time.Now().UnixMilli()),
    Channel:     fmt.Sprintf("demo-channel-%d", time.Now().UnixMilli()),
})
```

The same rule applies to `agentkit/cn`: create `cn.NewAgoraClient` first, then `cn.NewAgent(client, ...)`.

## Functional Options Pattern

`agentkit.NewAgent` uses Go's [functional options pattern](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis). Instead of a large config struct, you pass option functions that each configure one aspect of the agent:

<!-- snippet: fragment -->
```go
agent := agentkit.NewAgent(client).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
    APIKey:  "your-openai-key",
    BaseURL: "https://api.openai.com/v1/chat/completions",
    Model:   "gpt-4o-mini",
    SystemMessages: []map[string]interface{}{
        {"role": "system", "content": "You are a helpful voice assistant."},
    },
    GreetingMessage: "Hello! How can I help?",
    FailureMessage:  "Sorry, something went wrong.",
    MaxHistory:      intPtr(10),
}))
```

Each `With*` function has the signature `func(...) AgentOption`, where the option is applied while constructing a client-bound `Agent`. This pattern lets you:

- Omit any option you don't need (sensible defaults)
- Add new options without breaking existing code
- Compose options from helper functions

## AgentOption Functions

These are passed to `agentkit.NewAgent(client, opts ...AgentOption)`:

| Function | Parameter | Description |
|---|---|---|
| `WithPipelineID(pipelineID string)` | AI Studio pipeline ID | Default pipeline for sessions created from this agent |
| `WithInstructions(instructions string)` | System prompt | Deprecated. Use LLM vendor `SystemMessages` instead. |
| `WithGreeting(greeting string)` | Greeting text | Deprecated. Use LLM/MLLM vendor `GreetingMessage` instead. |
| `WithFailureMessage(msg string)` | Fallback message | Deprecated. Use LLM/MLLM vendor `FailureMessage` instead. |
| `WithMaxHistory(n int)` | History depth | Deprecated. Use LLM vendor `MaxHistory` instead. |
| `WithTurnDetectionConfig(td *TurnDetectionConfig)` | Turn detection config | Configure `turn_detection.language` and cascading-flow SOS/EOS detection |
| `WithSalConfig(sal *SalConfig)` | SAL config | Speech analytics configuration |
| `WithAdvancedFeatures(af *AdvancedFeatures)` | Feature flags | RTM, tools, and other advanced features |
| `WithParameters(params *SessionParams)` | Session params | Additional session parameters |
| `WithGeofence(gf *GeofenceConfig)` | Geofence config | Regional access restriction |
| `WithLabels(labels map[string]string)` | Labels map | Custom key-value labels (returned in callbacks) |
| `WithRtc(rtc *RtcConfig)` | RTC config | RTC media encryption |
| `WithFillerWords(fw *FillerWordsConfig)` | Filler words | Filler words while waiting for LLM |

Filler words support two content modes. Use `FillerWordsContentModeStatic` with
`FillerWordsContentStaticConfig` for a fixed phrase list, or
`FillerWordsContentModeGenerated` to generate a short phrase from the latest
user message. In generated mode, omit `GeneratedConfig` to use the service
default generator, or provide `FillerWordsContentGeneratedConfig` with an
OpenAI-compatible `FillerWordsContentGeneratedLlmProvider`. Set
`FillerWordsFallbackStrategyStatic` when a static fallback is configured.

## Vendor Chaining Methods

After creating an agent with `NewAgent`, attach vendors using method chaining. Each method returns a **new** `*Agent` (the original is not modified — immutable cloning):

<!-- snippet: fragment -->
```go
agent := agentkit.NewAgent(client).WithLlm(
    vendors.NewOpenAI(vendors.OpenAIOptions{
        APIKey:  "<key>",
        BaseURL: "https://api.openai.com/v1/chat/completions",
        Model:   "gpt-4o-mini",
        SystemMessages: []map[string]interface{}{
            {"role": "system", "content": "You are helpful."},
        },
    }),
).WithTts(
    vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
        Key:     "<key>",
        ModelID:    "eleven_turbo_v2_5",
        VoiceID:    "<voice_id>",
        BaseURL:    "wss://api.elevenlabs.io/v1",
    }),
).WithStt(
    vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
        APIKey: "<key>",
    }),
)
```

| Method | Parameter Type | Description |
|---|---|---|
| `WithLlm(vendor vendors.LLM)` | `vendors.LLM` interface | Set the LLM vendor |
| `WithTts(vendor vendors.TTS)` | `vendors.TTS` interface | Set the TTS vendor (also captures sample rate) |
| `WithStt(vendor vendors.STT)` | `vendors.STT` interface | Set the STT vendor |
| `WithMllm(vendor vendors.MLLM)` | `vendors.MLLM` interface | Set the MLLM vendor (for multimodal flow) |
| `WithAvatar(vendor vendors.Avatar)` | `vendors.Avatar` interface | Set the avatar vendor (validates sample rate) |
| `WithTurnDetection(td *TurnDetectionConfig)` | Pointer to config | Configure `turn_detection.language` and cascading-flow SOS/EOS detection; use interruption config for interruption behavior |
| `WithInstructions(instructions string)` | String | Deprecated. Use LLM vendor `SystemMessages` instead. |
| `WithGreeting(greeting string)` | String | Deprecated. Use LLM/MLLM vendor `GreetingMessage` instead. |
| `WithSal(sal *SalConfig)` | Pointer to config | Set SAL configuration |
| `WithAdvancedFeatures(af *AdvancedFeatures)` | Pointer to config | Set advanced features |
| `WithParameters(params *SessionParams)` | Pointer to config | Set session parameters |
| `WithFailureMessage(msg string)` | String | Deprecated. Use LLM/MLLM vendor `FailureMessage` instead. |
| `WithMaxHistory(n int)` | Int | Deprecated. Use LLM vendor `MaxHistory` instead. |
| `WithGeofence(gf *GeofenceConfig)` | Pointer to config | Set geofence configuration |
| `WithLabels(labels map[string]string)` | Map | Set custom labels |
| `WithRtc(rtc *RtcConfig)` | Pointer to config | Set RTC configuration |
| `WithFillerWords(fw *FillerWordsConfig)` | Pointer to config | Set filler words configuration |

Note: `WithInstructions`, `WithGreeting`, `WithFailureMessage`, and `WithMaxHistory` are compatibility shims. New code should configure instructions and messages on the LLM or MLLM vendor. `WithMaxHistory` applies only to the cascading LLM; Azure MLLM history uses `AzureOpenAIRealtimeOptions.MaxHistory`.

## Session name

The agent instance `name` sent on `/join` is set on `CreateSession`, not on the `Agent` builder. Pass `CreateSessionOptions.Name` when creating a session; if omitted, AgentKit generates `agent-<unix_timestamp>`. See [Session](./session.md).

## Agent Getters

<!-- snippet: fragment -->
```go
agent.Instructions() string
agent.Greeting() string
agent.FailureMessage() string
agent.MaxHistory() *int
agent.LlmConfig() map[string]interface{}
agent.TtsConfig() map[string]interface{}
agent.SttConfig() map[string]interface{}
agent.MllmConfig() map[string]interface{}
agent.TtsSampleRate() *vendors.SampleRate
agent.AvatarRequiredSampleRate() *vendors.SampleRate
agent.Avatar() map[string]interface{}
agent.TurnDetection() *TurnDetectionConfig
agent.Sal() *SalConfig
agent.AdvancedFeatures() *AdvancedFeatures
agent.Parameters() *SessionParams
agent.Geofence() *GeofenceConfig
agent.Labels() map[string]string
agent.Rtc() *RtcConfig
agent.FillerWords() *FillerWordsConfig
```

## ToProperties

`ToProperties` converts the agent configuration into a `*Agora.StartAgentsRequestProperties` suitable for the API:

<!-- snippet: fragment -->
```go
props, err := agent.ToProperties(agentkit.ToPropertiesOptions{
    Channel:        "my-channel",
    AgentUID:       "1001",
    RemoteUIDs:     []string{"1002"},
    AppID:          "<app_id>",
    AppCertificate: "<app_cert>",
})
if err != nil {
    log.Fatalf("Failed to build properties: %v", err)
}
```

### ToPropertiesOptions Fields

| Field | Type | Description |
|---|---|---|
| `Channel` | `string` | Agora channel name |
| `AgentUID` | `string` | UID for the agent in the channel |
| `RemoteUIDs` | `[]string` | UIDs of remote participants |
| `Token` | `string` | Pre-generated RTC token (if provided, skips generation) |
| `AppID` | `string` | Agora App ID (for token generation) |
| `AppCertificate` | `string` | Agora App Certificate (for token generation) |
| `UID` | `uint32` | Numeric UID for token generation |
| `TokenExpirySeconds` | `int` | Token expiry (default: 3600) |
| `IdleTimeout` | `*int` | Session idle timeout in seconds |
| `EnableStringUID` | `*bool` | Enable string UIDs |

If `Token` is empty, `ToProperties` generates one from `AppID` + `AppCertificate`. If both are empty, it returns an error.

In cascading mode, `ToProperties` requires both LLM and TTS to be configured — it returns an error if either is missing. In MLLM mode (when `mllm.enable` is `true`), LLM and TTS are not required.

## Type Aliases

The agentkit package defines convenient type aliases for common Agora types:

<!-- snippet: fragment -->
```go
type TurnDetectionConfig = Agora.StartAgentsRequestPropertiesTurnDetection
type SalConfig = Agora.StartAgentsRequestPropertiesSal
type AdvancedFeatures = Agora.StartAgentsRequestPropertiesAdvancedFeatures
type SessionParams = Agora.StartAgentsRequestPropertiesParameters
type GeofenceConfig = Agora.StartAgentsRequestPropertiesGeofence
type RtcConfig = Agora.StartAgentsRequestPropertiesRtc
type FillerWordsConfig = Agora.StartAgentsRequestPropertiesFillerWords
type LlmConfig = Agora.Llm
type MllmConfig = Agora.Mllm
type AsrConfig = Agora.Asr
type TtsConfig = Agora.Tts
type AvatarConfig = Agora.StartAgentsRequestPropertiesAvatar
```

Additional SOS/EOS turn detection aliases (`TurnDetectionNestedConfig`, `StartOfSpeechConfig`, `EndOfSpeechConfig`, etc.) are available — see the [Agent Reference](../reference/agent.md).
