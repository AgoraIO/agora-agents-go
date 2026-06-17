# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/).

## [v2.2.0] — 2026-06-05

### Added

- **Expanded provider surface** — Added generated API support for the latest Conversational AI vendors and configuration types, including Dify LLM and Generic Avatar.
- **Interaction language handling** — AgentKit now consistently derives REST `asr.language` from `TurnDetectionConfig.Language` while keeping provider-specific STT language values under `asr.params`.
- **Deepgram keyterm** — Added `Keyterm` support on `DeepgramSTT`, serialized as `asr.params.keyterm`.

### Changed

- **MiniMax managed presets** — MiniMax preset-backed TTS now keeps the preset model as an internal hint while sending only supported partial TTS settings such as `voice_setting.voice_id`.
- **Vertex AI LLM routing** — `NewVertexAILLM` now keeps project and location in the generated endpoint URL instead of duplicating them in `llm.params`.

### Fixed

- **Provider wire keys** — Corrected alias-sensitive TTS payloads so Google TTS emits `VoiceSelectionParams` and `AudioConfig`, Rime TTS emits `modelId`, and Murf TTS preserves `voiceId`.
- **AgentKit request validation** — Start request validation preserves generated wire keys while still allowing preset and pipeline-backed partial configs.
- **Request body coverage** — Added regression tests for BYOK, preset-backed, mixed preset/BYOK, and pipeline override request shapes across provider configurations.
- **TTS vendor tests** — Added Microsoft TTS coverage to the vendor helper test matrix.

## [v2.1.0] — 2026-06-02

### Added

- **Turn detection language** — AgentKit now manages Agora interaction language through `TurnDetectionConfig.Language`, validates it against the supported BCP-47 language list, and sends the default `en-US` when no language is provided.
- **Provider parameter parity** — ASR, LLM, MLLM, TTS, and avatar wrappers expose typed provider parameters plus passthrough fields where the generated core supports additional properties.

### Changed

- **Generated core refresh** — Regenerated core types from the v2.1 API schema.
- **Deepgram TTS passthrough** — `NewDeepgramTTS` now uses `AdditionalParams` for passthrough fields and flattens them into `tts.params`; the removed nested `params.params` shape is no longer documented or emitted.
- **OpenAI TTS** — Docs and tests now reflect the generated core shape, including `Instructions` and `Speed` under `tts.params`.
- **TTS provider docs** — Updated TTS provider reference tables to match implemented wrapper fields and generated core params.

### Fixed

- **Managed-provider validation** — AgentKit validation now distinguishes preset-backed providers from BYOK providers so required provider fields are only required when credentials are caller-supplied.
- **Language placement** — Provider-specific STT language values remain under `asr.params`; the REST `asr.language` field is populated from `turn_detection.language`.
- **Deepgram TTS explicit fields** — `AdditionalParams` can no longer override explicit `APIKey`, `Model`, `BaseURL`, or `SampleRate` values.

## [v2.0.0] — 2026-05-21

AgentKit alignment for Conversational AI v2.7.

### Added

- **Alias parity** — Exported `SttConfig`, session/conversation type aliases, `IsAvatarTokenManaged`, think type aliases, and cross-SDK discovery table in `docs/reference/agent.md`.
- **`AgoraClient.Telephony` and `AgoraClient.PhoneNumbers`** — AgentKit callers can reach the v2.7 telephony and phone-number REST endpoints without rebuilding the generated client.
- **`vendors.NewXaiGrok`** — xAI Grok MLLM sessions (`mllm.vendor`: `"xai"`), matching the TypeScript `XaiGrok` shape. `NewXAIGrok` remains as a deprecated alias.
- **`vendors.NewGenericAvatar` and `IsGenericAvatar`** — Generic avatar wrapper for custom avatar providers.
- **Avatar parameter enrichment** — Generic avatars get `agora_appid`, `agora_channel`, and `agora_token` from the session when omitted; LiveAvatar and HeyGen get `agora_token` auto-generated when omitted.
- **`WithGreetingConfigs`** — `llm.greeting_configs`, including v2.7 `interruptable`.
- **`GetTurnsOptions` and `GetAllTurns`** — Turn pagination helpers. `GetAllTurns` returns the full response with aggregated `Turns`.
- **Think action constants** — `ThinkOnListeningAction*`, `ThinkOnThinkingAction*`, and `ThinkOnSpeakingAction*` for v2.7 Think actions.
- **Interruption constants** — `InterruptionModeStartOfSpeech`, `InterruptionModeKeywords`, `InterruptionDisabledStrategyAppend`, and `InterruptionDisabledStrategyIgnore` for the v2.7 `interruption` object.
- **Speak priority constants** — `SpeakPriorityInterrupt`, `SpeakPriorityAppend`, and `SpeakPriorityIgnore` for `AgentSession.Say`.
- **MLLM turn detection constants** — `MllmTurnDetectionModeAgoraVad`, `MllmTurnDetectionModeServerVad`, and `MllmTurnDetectionModeSemanticVad` for the MLLM `turn_detection.mode` field.
- **`AzureOpenAIOptions.Model`** — Emits `params.model` for parity with the TypeScript SDK; Azure ignores the value for chat completions, but downstream tooling and logs surface it.

### Changed

- **Repository and module path** — The repository has been updated to [`AgoraIO/agora-agents-go`](https://github.com/AgoraIO/agora-agents-go) (formerly `AgoraIO-Conversational-AI/agent-server-sdk-go`). Update imports to `github.com/AgoraIO/agora-agents-go/v2`.
- **Go module v2 path alignment** — `go.mod` now declares `module github.com/AgoraIO/agora-agents-go/v2`, and all SDK imports/examples use `/v2` paths so standard Go module resolution works for `v2.0.0`.
- **ConvoAI token options** — `GenerateConvoAIToken()` now accepts an integer `UID` and handles the internal token string conversion for users, agents, and avatars.
- **Avatar token generation** — Removed the dedicated `GenerateAvatarRtcToken()` wrapper; avatar RTC tokens use the existing ConvoAI token helper.
- **Session lifecycle naming** — Renamed the AgentKit lifecycle type to `AgentSessionLifecycle`; `SessionStatus` is now the generated API status alias.
- **`AgentSession.Start`** — Sends a map-based join payload after preset resolution, preventing generated structs from reintroducing empty provider-owned fields such as `llm.url`, `llm.api_key`, or `tts.params.key`.
- **`ToPropertiesMap`** — Builds vendor configs from maps directly for closer parity with Python and TypeScript AgentKit.
- **`GetTurns`** — Supports `page_index` and `page_size`; callers with more than one page should paginate or call `GetAllTurns`.
- **`Agent.ToPropertiesMap`** — Rejects MLLM + enabled avatar combinations before generating tokens or building properties; avatars currently require the cascading ASR/LLM/TTS pipeline.
- **Avatar vendor `ToConfig()`** — HeyGen, LiveAvatar, Akool, Anam, and Generic now spread `AdditionalParams` first so required fields like `api_key`, `quality`, and `agora_uid` always take precedence over caller overrides.
- **`OpenAIRealtime.ToConfig`** — Explicit `Params["model"]` overrides the named `Model`, matching the TypeScript SDK and the existing Gemini/Vertex AI/xAI Grok behavior.

### Documentation

- Documented v2.7 error `reason` codes, Think default behavior, event `112`, avatar token handling, Generic avatars, xAI Grok, Deepgram TTS, Murf, Anam, LiveAvatar, `pause_state_enabled`, `audio_scenario`, and `pipeline_id`.

### Migration Notes

- Deprecated aliases remain for compatibility. Use `NewXaiGrok` / `XaiGrok` / `XaiGrokOptions` instead of `NewXAIGrok` / `XAIGrok` / `XAIGrokOptions`, and `NewLiveAvatarAvatar` / `LiveAvatarAvatar` / `LiveAvatarAvatarOptions` instead of `NewHeyGenAvatar` / `HeyGenAvatar` / `HeyGenAvatarOptions`.
- In v2.7, omitting `ThinkOptions.OnListeningAction` uses the server default `interrupt`. Pass `agentkit.ThinkOnListeningActionInject.Ptr()` to preserve inject-style behavior.
- Avatar `AgoraUID` should be distinct from the session `AgentUID`. The SDK warns on collisions and preserves explicitly provided avatar tokens.
- Go consumers should install and import via the `/v2` module path, for example: `go get github.com/AgoraIO/agora-agents-go/v2@v2.0.0`.

## [v1.4.0] — 2026-05-13

### Added

- **`NewDeepgramTTS`** — New TTS vendor wrapper for Deepgram (Beta). Accepts `APIKey`, `Model`, `BaseURL`, `SampleRate`, `AdditionalParams`, and `SkipPatterns`.
- **`Agent.WithTools(enabled bool)` / `WithTools` option** — Dedicated builder method and constructor option to enable MCP tool invocation (`advanced_features.enable_tools`). Replaces the raw `WithAdvancedFeatures(&AdvancedFeatures{EnableTools: Agora.Bool(true)})` call.
- **LLM vendors: `Headers` field** — All four LLM vendors (`OpenAI`, `AzureOpenAI`, `Anthropic`, `Gemini`) now accept an optional `Headers map[string]string` field. Use this to pass custom HTTP headers to the LLM provider (e.g., tenant identifiers, routing headers).
- **`AgentSession.Think()` / `ThinkWithOptions()`** — Send a custom instruction to a running agent through the agent management API.
- **`Agent.WithInterruption()` / `WithInterruptionConfig()`** — Configure the new top-level `interruption` object for unified interruption control.
- **MLLM turn detection** — `NewOpenAIRealtime`, `NewGeminiLive`, and `NewVertexAI` now accept `TurnDetection`, which maps to `mllm.turn_detection` and overrides top-level turn detection for MLLM sessions.
- **`AudioScenario` AgentKit support** — Session params and AgentKit request construction now expose the top-level `parameters.audio_scenario` field.

### Fixed

- **MiniMax TTS preset stripping** — When a MiniMax reseller preset is inferred (`minimax_speech_2_6_turbo` or `minimax_speech_2_8_turbo`), the `group_id` and `url` fields are now correctly stripped from `tts.params` alongside `key` and `model`. Previously they were forwarded to the API, causing request failures.
- **MLLM enable flag** — `Agent.WithMllm()` now sets `mllm.enable = true` and removes the deprecated `advanced_features.enable_mllm` flag from generated requests.
- **MLLM wrapper shape** — MLLM vendors no longer emit removed fields such as `style`; docs and tests now reflect the v2.6 MLLM contract.
- **Preset-backed OpenAI TTS** — `NewOpenAITTS` no longer requires `APIKey` when a reseller preset supplies credentials server-side.
- **AgentKit parity coverage** — Added regression coverage for interruption, MLLM turn detection, Deepgram TTS, LLM headers, and deprecated MLLM flag cleanup.

## [v1.3.4] — 2026-04-28

### Fixed

- **Managed preset payload fields** — AgentKit start payloads now stay in a dynamic map shape after preset resolution so provider-owned fields removed for managed presets do not reappear as generated Go zero values.
- **Preset inference and stripping** — MiniMax and OpenAI TTS preset fields are inferred and stripped correctly during start request construction.
- **Keyless ARES ASR payloads** — Preset-backed ARES ASR configurations without explicit credentials are preserved instead of being overwritten by empty struct fields.

## [v1.3.3] — 2026-04-27

### Fixed

- **Region base URL construction** — Corrected geofence/region routing when building the Conversational AI API base URL. Added unit tests for area routing behavior.

## [v1.3.2] — 2026-04-23

### Fixed

- **`GeminiLive` and `VertexAI` MLLM `url` parameter** — The optional WebSocket URL override supported by `OpenAIRealtime` was not exposed on the other two MLLM vendors, making it impossible to point them at a custom endpoint. Both constructors now accept an optional `URL` field that is passed through to the API request.
- **Reseller API key support** — Preset-backed vendor configurations now correctly handle reseller-managed API keys without requiring caller-supplied credentials.

## [v1.3.0] — 2026-04-02

### Added

- **`AgentSession.GetTurns()`** — Retrieve conversation turn analytics for a running session.
- **Session-level `Preset` and `PipelineID`** — `CreateSessionOptions` now accepts `Preset` and `PipelineID`. Automatic reseller preset inference for supported Deepgram (nova-2, nova-3), OpenAI (gpt-4o-mini, gpt-4.1-mini, gpt-5-nano, gpt-5-mini), and MiniMax (speech-2.6-turbo, speech-2.8-turbo) models.
- **`AgentPresets` constants** — Type-safe preset constant map for discoverable session preset composition.
- **`LiveAvatarAvatar` and `AnamAvatar`** — New avatar vendor wrappers. `LiveAvatarAvatar` is the successor to `HeyGenAvatar` and emits `vendor: "liveavatar"`.

### Changed

- **`GeminiLive` MLLM wrapper** — Output now matches the Agora low-level MLLM contract; `messages` stays at top level.
- **Avatar sample-rate validation** — `CreateSession` validates TTS sample rate against avatar vendor requirements and emits a warning (via `WarnHook`) when using `HeyGenAvatar` or `LiveAvatarAvatar` without an explicit 24 kHz TTS sample rate.

### Fixed

- **MLLM wrapper fields** — Removed unsupported wrapper-only fields that caused the Go surface to diverge from the generated Agora API contract.

## [v1.2.0] — 2026-03-27

### Fixed

- **`AresSTT`** — Removed redundant `language` key from the `Params` map. Ares only selects the provider; AgentKit populates REST `asr.language` from `TurnDetectionConfig.Language`.
- **`OpenAIRealtime` / `GeminiLive` (MLLM)** — Agent-level `greeting`, `FailureMessage`, and `MaxHistory` overrides are now correctly applied in MLLM mode.
- **`GeminiLive` (MLLM)** — `messages` is now correctly placed inside `params` as required by the Gemini Live API.

### Changed

- **`OpenAITTS`** — Constructor field renamed `Key` → `APIKey`. ⚠️ **Breaking change.**
- **`CartesiaTTS`** — Constructor field renamed `Key` → `APIKey`. Voice is now serialized as `{"mode": "id", "id": "<voice_id>"}`. ⚠️ **Breaking change.**
- **`HeyGenAvatar`** — Removed legacy fields `AvatarName`, `VoiceID`, `Language`, `Version`. Added `AgoraToken`, `AvatarID`, `Enable`, `DisableIdleTimeout`, `ActivityIdleTimeout`. Config now includes a top-level `enable` field (defaults `true`). ⚠️ **Breaking change.**

### Added

- **`OpenAITTS`** — New optional fields: `Instructions` (string) and `Speed` (float64).
- **`CartesiaTTS`** — `VoiceID` user-facing field is preserved; serialized to the required nested object format automatically.
- **`RimeTTS`** — New optional fields: `Lang` (string), `SamplingRate` (int), `SpeedAlpha` (float64).
- **`OpenAIRealtime`** — New optional fields: `PredefinedTools` ([]string), `FailureMessage` (string), `MaxHistory` (int).
- **`GeminiLive` (MLLM)** — New optional fields: `PredefinedTools` ([]string), `FailureMessage` (string), `MaxHistory` (int).

## [v1.1.0] — 2026-03-17

### Added

- **`NewMurfTTS`** — New TTS vendor wrapper for Murf.
- **`NewSarvamTTS`** — New TTS vendor wrapper for Sarvam.
- **`NewSarvamSTT`** — New STT vendor wrapper for Sarvam.
- All LLM vendors: added `MaxHistory` field for conversation history caching.
- **`AzureOpenAI`** — Added `Params` escape hatch for arbitrary API parameters.
- **`Anthropic`** — Added `URL` for custom endpoints and `Params` escape hatch.
- **`Gemini`** — Added `URL` for custom endpoints and `Params` escape hatch.
- **`SpeechmaticsSTT`**, **`SarvamSTT`** — Added optional `Model` field.

### Fixed

- **`MiniMaxTTS`** — Added required `GroupID`, `URL`, and correctly nested `VoiceSetting.VoiceID` fields.
- **`SarvamTTS`** — Corrected schema to `Key` + `Speaker` + `TargetLanguageCode`.

## [v1.0.0] — 2026-03-11

Initial stable release of Agora Agents Go.

### Added

- `Agent` builder with immutable fluent API (`WithLlm()`, `WithTts()`, `WithStt()`, `WithMllm()`, `WithAvatar()`) and `NewAgent(opts...)` constructor.
- `AgentSession` for session lifecycle management (`Start()`, `Stop()`, `Say()`, `Interrupt()`, `Update()`, `GetHistory()`, `GetInfo()`).
- App-credentials auth — pass `AppID` + `AppCertificate` to `AgoraClientOptions` and ConvoAI tokens are generated per request.
- Token utilities: `GenerateRtcToken`, `GenerateConvoAIToken`.
- Turn detection configuration via `TurnDetectionConfig` with full `StartOfSpeechConfig` / `EndOfSpeechConfig` nested types and constants.
- SAL (Selective Attention Locking) via `SalConfig` with `SalModeLocking` / `SalModeRecognition` constants.
- Filler words support via `FillerWordsConfig`.
- Session parameters: `SessionParams`, `SilenceConfig`, `FarewellConfig`.
- Geofencing via `GeofenceConfig` with `GeofenceArea` constants.
- Advanced features (MLLM mode) via `AdvancedFeatures`.
- `WarnHook` callback for runtime warnings (avatar sample rate mismatches, handler panics).
- Vendor integrations:
  - **LLM**: `NewOpenAI`, `NewAzureOpenAI`, `NewAnthropic`, `NewGemini`
  - **MLLM**: `NewOpenAIRealtime`, `NewGeminiLive`
  - **TTS**: `NewElevenLabsTTS`, `NewMicrosoftTTS`, `NewOpenAITTS`, `NewCartesiaTTS`, `NewGoogleTTS`, `NewAmazonTTS`, `NewHumeAITTS`, `NewRimeTTS`, `NewFishAudioTTS`, `NewMiniMaxTTS`
  - **STT**: `NewDeepgramSTT`, `NewMicrosoftSTT`, `NewOpenAISTT`, `NewGoogleSTT`, `NewAmazonSTT`, `NewAssemblyAISTT`, `NewAresSTT`, `NewSpeechmaticsSTT`
  - **Avatar**: `NewHeyGenAvatar`, `NewAkoolAvatar`
