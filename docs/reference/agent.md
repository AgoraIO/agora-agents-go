---
sidebar_position: 2
title: Agent Reference
description: Complete API reference for agentkit.Agent — functional options, methods, and ToProperties.
---

# Agent Reference

Package: `github.com/AgoraIO/agora-agents-go/v2/agentkit`

## NewAgent

<!-- snippet: fragment -->
```go
func NewAgent(client *AgoraClient, opts ...AgentOption) *Agent
```

Creates a new `Agent` bound to the provided `AgoraClient` and configured with the given functional options.

`client` is required. Pass a non-nil `*AgoraClient` from `NewAgoraClient`. If `client` is `nil`, `NewAgent` panics with `NewAgent requires AgoraClient`.

## AgentOption Type

<!-- snippet: fragment -->
```go
type AgentOption func(*BaseAgent)
```

## AgentOption Functions

### WithPipelineID

<!-- snippet: fragment -->
```go
func WithPipelineID(pipelineID string) AgentOption
```

Sets the published AI Studio pipeline ID to use as this agent's base configuration. Explicit Agent config such as `WithLlm`, `WithTts`, `WithStt`, advanced features, and other builder options may send fields in `properties` that override the saved pipeline settings. Session-level `CreateSessionOptions.PipelineID` overrides this value.

### WithInstructions

<!-- snippet: fragment -->
```go
func WithInstructions(instructions string) AgentOption
```

Deprecated. Configure `SystemMessages` on the LLM vendor instead.

### WithGreeting

<!-- snippet: fragment -->
```go
func WithGreeting(greeting string) AgentOption
```

Deprecated. Configure `GreetingMessage` on the LLM or MLLM vendor instead.

### WithFailureMessage

<!-- snippet: fragment -->
```go
func WithFailureMessage(msg string) AgentOption
```

Deprecated. Configure `FailureMessage` on the LLM or MLLM vendor instead.

### WithMaxHistory

<!-- snippet: fragment -->
```go
func WithMaxHistory(n int) AgentOption
```

Deprecated. Configure `MaxHistory` on the LLM vendor instead.

### WithTurnDetectionConfig

<!-- snippet: fragment -->
```go
func WithTurnDetectionConfig(td *TurnDetectionConfig) AgentOption
```

Sets cascading-flow turn detection configuration. Use `Language` for the Agora interaction language, `Config.StartOfSpeech` and `Config.EndOfSpeech` for SOS/EOS detection, interruption config for interruption behavior, and MLLM vendor `TurnDetection` for MLLM turn detection.

### WithInterruptionConfig

<!-- snippet: fragment -->
```go
func WithInterruptionConfig(interruption *InterruptionConfig) AgentOption
```

Sets unified interruption control using the top-level `interruption` object. Use the `agentkit.InterruptionModeStartOfSpeech` / `InterruptionModeKeywords` and `InterruptionDisabledStrategyAppend` / `InterruptionDisabledStrategyIgnore` convenience constants when populating `InterruptionConfig.Mode` and `InterruptionConfig.DisabledConfig.Strategy`.

### WithGreetingConfigs

<!-- snippet: fragment -->
```go
func WithGreetingConfigs(configs *LlmGreetingConfigs) AgentOption
```

Deprecated. Configure `GreetingConfigs` on the LLM vendor instead.

### WithGreetingAudioURL

<!-- snippet: fragment -->
```go
func WithGreetingAudioURL(url string) AgentOption
```

Sets `llm.greeting_audio_url` for the cascading-flow LLM configuration.

### WithSalConfig

<!-- snippet: fragment -->
```go
func WithSalConfig(sal *SalConfig) AgentOption
```

Sets the speech analytics configuration.

### WithAdvancedFeatures

<!-- snippet: fragment -->
```go
func WithAdvancedFeatures(af *AdvancedFeatures) AgentOption
```

Sets advanced feature flags such as RTM or tool invocation.

### WithTools

<!-- snippet: fragment -->
```go
func WithTools(enabled bool) AgentOption
```

Enables or disables tool invocation by setting `AdvancedFeatures.EnableTools`.
Use this together with `Tools` or `McpServers` on the selected LLM vendor options.

### WithParameters

<!-- snippet: fragment -->
```go
func WithParameters(params *SessionParams) AgentOption
```

Sets additional session parameters.

### WithSessionOptOut

<!-- snippet: fragment -->
```go
func WithSessionOptOut(optOut bool) AgentOption
```

Sets `parameters.opt_out` for the session start request.

### WithAudioScenario

<!-- snippet: fragment -->
```go
func WithAudioScenario(audioScenario ParametersAudioScenario) AgentOption
```

Sets `parameters.audio_scenario` (`default`, `chorus`, or `aiserver`).

### WithGeofence

<!-- snippet: fragment -->
```go
func WithGeofence(gf *GeofenceConfig) AgentOption
```

Sets geofence configuration (restricts backend server regions).

### WithLabels

<!-- snippet: fragment -->
```go
func WithLabels(labels map[string]string) AgentOption
```

Sets custom labels (key-value pairs returned in notification callbacks).

### WithRtc

<!-- snippet: fragment -->
```go
func WithRtc(rtc *RtcConfig) AgentOption
```

Sets RTC configuration.

### WithFillerWords

<!-- snippet: fragment -->
```go
func WithFillerWords(fw *FillerWordsConfig) AgentOption
```

Sets filler words configuration (played while waiting for LLM response).

## Agent Methods

All vendor-chaining methods return a **new** `*Agent` (immutable clone). The original agent is not modified.

Accessor methods include `PipelineID()`, vendor config getters, and getters for optional advanced configuration.

### WithLlm

<!-- snippet: fragment -->
```go
func (a *Agent) WithLlm(vendor vendors.LLM) *Agent
```

### WithTts

<!-- snippet: fragment -->
```go
func (a *Agent) WithTts(vendor vendors.TTS) *Agent
```

Also captures the vendor's sample rate for avatar validation.

### WithStt

<!-- snippet: fragment -->
```go
func (a *Agent) WithStt(vendor vendors.STT) *Agent
```

### WithMllm

<!-- snippet: fragment -->
```go
func (a *Agent) WithMllm(vendor vendors.MLLM) *Agent
```

### WithAvatar

<!-- snippet: fragment -->
```go
func (a *Agent) WithAvatar(vendor vendors.Avatar) *Agent
```

**Panics** if TTS is already configured with a sample rate that doesn't match the avatar's requirement. See [Avatars Guide](../guides/avatars.md).

### WithTurnDetection (method)

<!-- snippet: fragment -->
```go
func (a *Agent) WithTurnDetection(td *TurnDetectionConfig) *Agent
```

Sets cascading-flow turn detection configuration. Use `Language` for the Agora interaction language, `Config.StartOfSpeech` and `Config.EndOfSpeech` for SOS/EOS detection, interruption config for interruption behavior, and MLLM vendor `TurnDetection` for MLLM turn detection.

Example with `pause_state_enabled`:

```go
enabled := true
mode := "default"
eosMode := Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeechModeSemantic

agent := agentkit.NewAgent(client,
    agentkit.WithTurnDetectionConfig(&agentkit.TurnDetectionConfig{
        Mode: &mode,
        Config: &agentkit.TurnDetectionNestedConfig{
            EndOfSpeech: &agentkit.EndOfSpeechConfig{
                Mode: &eosMode,
                SemanticConfig: &agentkit.EndOfSpeechSemanticConfig{
                    PauseStateEnabled: &enabled,
                },
            },
        },
    }),
)
```

### WithInstructions (method)

<!-- snippet: fragment -->
```go
func (a *Agent) WithInstructions(instructions string) *Agent
```

Deprecated. Configure `SystemMessages` on the LLM vendor instead.

### WithGreeting (method)

<!-- snippet: fragment -->
```go
func (a *Agent) WithGreeting(greeting string) *Agent
```

Deprecated. Configure `GreetingMessage` on the LLM or MLLM vendor instead.

### WithSal (method)

<!-- snippet: fragment -->
```go
func (a *Agent) WithSal(sal *SalConfig) *Agent
```

### WithAdvancedFeatures (method)

<!-- snippet: fragment -->
```go
func (a *Agent) WithAdvancedFeatures(af *AdvancedFeatures) *Agent
```

### WithTools (method)

<!-- snippet: fragment -->
```go
func (a *Agent) WithTools(enabled bool) *Agent
```

Enables or disables tool invocation by setting `AdvancedFeatures.EnableTools`.

### WithAudioScenario (method)

<!-- snippet: fragment -->
```go
func (a *Agent) WithAudioScenario(audioScenario ParametersAudioScenario) *Agent
```

Sets `parameters.audio_scenario` on immutable agent clones.

### WithParameters (method)

<!-- snippet: fragment -->
```go
func (a *Agent) WithParameters(params *SessionParams) *Agent
```

### WithFailureMessage (method)

<!-- snippet: fragment -->
```go
func (a *Agent) WithFailureMessage(msg string) *Agent
```

Deprecated. Configure `FailureMessage` on the LLM or MLLM vendor instead.

### WithMaxHistory (method)

<!-- snippet: fragment -->
```go
func (a *Agent) WithMaxHistory(n int) *Agent
```

Deprecated. Configure `MaxHistory` on the LLM vendor instead.

### WithGeofence (method)

<!-- snippet: fragment -->
```go
func (a *Agent) WithGeofence(gf *GeofenceConfig) *Agent
```

### WithLabels (method)

<!-- snippet: fragment -->
```go
func (a *Agent) WithLabels(labels map[string]string) *Agent
```

### WithRtc (method)

<!-- snippet: fragment -->
```go
func (a *Agent) WithRtc(rtc *RtcConfig) *Agent
```

### WithFillerWords (method)

<!-- snippet: fragment -->
```go
func (a *Agent) WithFillerWords(fw *FillerWordsConfig) *Agent
```

## Getters

<!-- snippet: fragment -->
```go
func (a *Agent) Instructions() string
func (a *Agent) Greeting() string
func (a *Agent) FailureMessage() string
func (a *Agent) MaxHistory() *int
func (a *Agent) LlmConfig() map[string]interface{}
func (a *Agent) TtsConfig() map[string]interface{}
func (a *Agent) SttConfig() map[string]interface{}
func (a *Agent) MllmConfig() map[string]interface{}
func (a *Agent) TtsSampleRate() *vendors.SampleRate
func (a *Agent) AvatarRequiredSampleRate() *vendors.SampleRate
func (a *Agent) Avatar() map[string]interface{}
func (a *Agent) TurnDetection() *TurnDetectionConfig
func (a *Agent) Interruption() *InterruptionConfig
func (a *Agent) GreetingConfigs() *LlmGreetingConfigs
func (a *Agent) Sal() *SalConfig
func (a *Agent) AdvancedFeatures() *AdvancedFeatures
func (a *Agent) Parameters() *SessionParams
func (a *Agent) Geofence() *GeofenceConfig
func (a *Agent) Labels() map[string]string
func (a *Agent) Rtc() *RtcConfig
func (a *Agent) FillerWords() *FillerWordsConfig
```

## CreateSession

<!-- snippet: fragment -->
```go
func (a *Agent) CreateSession(opts CreateSessionOptions) *AgentSession
```

Creates an `AgentSession` from the agent builder. Set `CreateSessionOptions.Name` to identify the agent instance on `/join`. If `Name` is empty, AgentKit generates `agent-<unix_timestamp>`.

### CreateSessionOptions

<!-- snippet: fragment -->
```go
type CreateSessionOptions struct {
    Name            string
    Channel         string
    Token           string
    AgentUID        string
    RemoteUIDs      []string
    IdleTimeout     *int
    EnableStringUID *bool
    ExpiresIn       int
    Preset          []string
    PipelineID      string
    Debug           bool
    Warn            func(string)
}
```

| Field | Type | Description |
|---|---|---|
| `Name` | `string` | Agent instance name for `/join`. Defaults to `agent-<unix_timestamp>` when empty. |
| `Channel` | `string` | Agora channel name |
| `Token` | `string` | Pre-generated RTC token |
| `AgentUID` | `string` | Agent UID in the channel |
| `RemoteUIDs` | `[]string` | Remote participant UIDs |
| `IdleTimeout` | `*int` | Idle timeout in seconds |
| `EnableStringUID` | `*bool` | Enable string UID mode |
| `ExpiresIn` | `int` | Auto-generated token lifetime in seconds |
| `Preset` | `[]string` | Advanced preset value for project-specific routing |
| `PipelineID` | `string` | Overrides `agent.PipelineID()` for this session |
| `Debug` | `bool` | Log the start request payload |
| `Warn` | `func(string)` | Custom warning sink |

## ToProperties

<!-- snippet: fragment -->
```go
func (a *Agent) ToProperties(opts ToPropertiesOptions) (*Agora.StartAgentsRequestProperties, error)
```

Converts the agent configuration into API request properties. Handles token generation, LLM/TTS config merging, and validation.

Returns an error if:
- Neither `Token` nor `AppID`+`AppCertificate` is provided
- In cascading mode: LLM or TTS is not configured
- Config marshaling fails

### ToPropertiesOptions

<!-- snippet: fragment -->
```go
type ToPropertiesOptions struct {
    Channel                      string
    AgentUID                     string
    RemoteUIDs                   []string
    Token                        string
    AppID                        string
    AppCertificate               string
    ExpiresIn                    int
    IdleTimeout                  *int
    EnableStringUID               *bool
    SkipVendorValidation         bool
    SkipVendorValidationCategories []string
    AllowMissingVendorCategories []string
    Warn                         func(string)
}
```

| Field | Type | Description |
|---|---|---|
| `Channel` | `string` | Agora channel name |
| `AgentUID` | `string` | Agent's UID in the channel |
| `RemoteUIDs` | `[]string` | Remote participant UIDs |
| `Token` | `string` | Pre-generated RTC+RTM token (skips generation if set) |
| `AppID` | `string` | Agora App ID (for token generation) |
| `AppCertificate` | `string` | Agora App Certificate (for token generation) |
| `ExpiresIn` | `int` | Token lifetime in seconds (default: `86400` = 24 h, Agora max). Use `ExpiresInHours()` / `ExpiresInMinutes()` for clarity. Valid range: 1–86400. |
| `IdleTimeout` | `*int` | Session idle timeout |
| `EnableStringUID` | `*bool` | Enable string UID mode |
| `SkipVendorValidation` | `bool` | **Deprecated.** Advanced escape hatch that disables all category validation. Use `SkipVendorValidationCategories` instead. |
| `SkipVendorValidationCategories` | `[]string` | Skip validation only for listed categories (`"asr"`, `"llm"`, `"tts"`). |
| `AllowMissingVendorCategories` | `[]string` | Allow missing config in listed categories (`"asr"`, `"llm"`, `"tts"`). |
| `Warn` | `func(string)` | Warning sink for recoverable config issues |

## Type Aliases

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
type SttConfig = AsrConfig
type LlmStyle = Agora.LlmStyle
type SessionInfo = Agora.GetAgentsResponse
type ThinkResponse = Agora.AgentThinkAgentManagementResponse
```

Additional SOS/EOS turn detection aliases: `TurnDetectionNestedConfig`, `StartOfSpeechConfig`, `EndOfSpeechConfig`, and related sub-types. Session/conversation aliases: `SessionListResponse`, `ConversationHistory`, `ConversationTurns`, etc. Think type aliases: `ThinkOnListeningAction`, `ThinkOnThinkingAction`, `ThinkOnSpeakingAction`.

## Cross-SDK discovery map

| Concept | Go | TypeScript | Python |
|---|---|---|---|
| STT payload alias (wire: `asr`) | `AsrConfig` / `SttConfig` | `SttConfig` / `AsrConfig` | `SttConfig` / `AsrConfig` |
| xAI MLLM (primary) | `XaiGrok` / `NewXaiGrok` | `XaiGrok` | `XaiGrok` |
| Avatar token auto-generation | Managed vendors: `liveavatar`, `heygen`, `generic`, `sensetime` (see [Avatars Guide](../guides/avatars.md)) | `isAvatarTokenManaged` | `is_avatar_token_managed` |
| Think inject constant | `ThinkOnListeningActionInject` | `ThinkOnListeningActionInject` | `ThinkOnListeningActionInject` |

## Token Generation

<!-- snippet: fragment -->
```go
func GenerateRtcToken(opts GenerateTokenOptions) (string, error)
func GenerateRtcTokenWithAccount(opts GenerateRtcTokenWithAccountOptions) (string, error)
func GenerateConvoAIToken(opts GenerateConvoAITokenOptions) (string, error)
```

### GenerateTokenOptions

<!-- snippet: fragment -->
```go
type GenerateTokenOptions struct {
    AppID          string
    AppCertificate string
    Channel        string
    UID            uint32
    Role           int
    ExpirySeconds  int
}
```

### Constants

<!-- snippet: fragment -->
```go
const (
    RolePublisher      = 1
    RoleSubscriber     = 2
    DefaultExpirySeconds = 86400
)
```

Avatar token automation uses `GenerateConvoAIToken` with the avatar `AgoraUID` as `UID`.
