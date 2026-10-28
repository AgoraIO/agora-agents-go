# Conversational AI v3 API preview

The preview API supports cascaded ASR/LLM/TTS pipelines only. MLLM saves, ephemeral preview sessions, and preview session updates are rejected locally. Existing MLLM vendor builders and original session operations remain available.

This checkout adds handwritten Go resources for saved agents, secrets, project numbers, sessions, and campaigns, with AgentKit save/session workflows. Session updates support the documented token, message, and LLM params patch fields. Saved-agent BYOK credentials use secret references; ephemeral sessions also accept direct credential values.

Campaign creation accepts up to 50,000 contacts, data-retention policy, and the current pacing and call-policy fields. Session and campaign create/lifecycle operations support the service's 24-hour idempotency keys.

The preview rollout is currently US-only. The SDK does not enforce a region restriction; confirm that the target project is enabled before using the preview. These APIs are available from a build of the `feat/v3-api-preview` branch; the published `@latest` Go module does not select this branch.

## Documentation feedback fixes (2026-10-28)

Added explicit ephemeral sessions and outbound `Call` helpers; fixed LLM vendor defaults, managed-model gating, Gemini credentials, ARES keywords/default language, Deepgram credentials and ASR language/endpoint mapping. MiniMax has typed voice/audio options and preserves additional nested settings. Saved-number assignment supports inbound end-call policy. Optional session analysis and both secret-key metadata response shapes are decoded. Platform RTC tokens are generated when an explicit UID and certificate permit it; otherwise server token defaults apply. AgentKit now exposes secret-key types, interruption sub-configs, VAD/semantic constants, and pointer helpers.

Install the slashless `v3-api-preview` branch or pin its resolved pseudo-version. The Go module remains `/v2`, so a v3 tag is not a compatible module release.
