# Conversational AI v3 API preview

The preview API supports cascaded ASR/LLM/TTS pipelines only. MLLM saves, ephemeral preview sessions, and preview session updates are rejected locally. Existing MLLM vendor builders and original session operations remain available.

This checkout adds handwritten Go resources for saved agents, secrets, project numbers, sessions, and campaigns, with AgentKit save/session workflows. Session updates support the documented token, message, and LLM params patch fields. Saved-agent BYOK credentials use secret references; ephemeral sessions also accept direct credential values.

Campaign creation accepts up to 50,000 contacts, data-retention policy, and the current pacing and call-policy fields. Session and campaign create/lifecycle operations support the service's 24-hour idempotency keys.

The preview rollout is currently US-only. The SDK does not enforce a region restriction; confirm that the target project is enabled before using the preview. These APIs are available from a build of the `feat/v3-api-preview` branch; the published `@latest` Go module does not select this branch.
