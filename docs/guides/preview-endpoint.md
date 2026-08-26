---
sidebar_position: 10
title: Preview Endpoint
description: How NewAgoraClient routes preview sessions and pins the gateway's agora-feature gate header.
---

# Preview Endpoint

Some providers ship on a preview gateway before they reach the production Conversational AI environment. `NewAgoraClient` detects those providers from each session's resolved start body and routes that session automatically.

Everything in this guide is temporary by design. When a preview provider goes GA it moves to the production gateway, and the corresponding entry disappears.

## Using a preview provider

```go
client := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
    Area:           option.AreaUS,
    AppID:          os.Getenv("AGORA_APP_ID"),
    AppCertificate: os.Getenv("AGORA_APP_CERTIFICATE"),
})

agent := agentkit.NewAgent(client).
    WithStt(vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
        APIKey:        os.Getenv("GOOGLE_API_KEY"),
        LanguageCodes: []string{"en-US"},
    })).
    WithLlm(vendors.NewGemini(vendors.GeminiOptions{
        APIKey: os.Getenv("GOOGLE_API_KEY"),
        Model:  "gemini-2.0-flash",
    })).
    WithTts(vendors.NewGoogleTTS(vendors.GoogleTTSOptions{
        Key:          os.Getenv("GOOGLE_API_KEY"),
        VoiceName:    "en-US-Chirp3-HD-Charon",
        LanguageCode: "en-US",
    }))

session := agent.CreateSession(agentkit.CreateSessionOptions{
    Channel:    "demo",
    AgentUID:   "1",
    RemoteUIDs: []string{"100"},
})
agentID, err := session.Start(ctx)
```

There is no separate preview client. On `Start`, the SDK calls `RequiredPreviewFeatures` on the resolved body. Preview sessions bind the preview base URL and gate transport for their full lifecycle; GA sessions keep production regional routing.

The sample above is compiled as `ExampleNewAgoraClient_previewRouting` in `agentkit/preview_example_test.go`, so it cannot drift from the API.

## The gate header

The gateway routes preview traffic on a single request header:

```
agora-feature: gemini-live
```

| Constant                   | Value           |
| -------------------------- | --------------- |
| `PreviewFeatureHeader`     | `agora-feature` |
| `PreviewFeatureGeminiLive` | `gemini-live`   |

The SDK derives the feature list from the resolved session body; callers do not select it manually.

### The header is not overridable

`option.WithHTTPHeader` **replaces the entire header map** rather than merging into it. That is normal Go SDK behavior, but it means a per-call header option would silently drop the gate — and a preview request that loses the header is not rejected, it routes to the production environment where the preview providers do not exist.

So the gate is applied below the option layer, in a wrapped `core.HTTPClient`:

```go
func (p *previewGateClient) Do(req *http.Request) (*http.Response, error) {
    req.Header.Set(PreviewFeatureHeader, p.features)
    return p.inner.Do(req)
}
```

A custom `HTTPClient` passed in `AgoraClientOptions` is **wrapped for preview sessions**, so supplying your own transport keeps the gate intact:

```go
client := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
    Area:       option.AreaUS,
    AppID:      appID,
    HTTPClient: myInstrumentedClient,
})
```

The header rides every session verb, not just `Start` — including `Say`, `Interrupt`, `Think`, `Update`, history/info/turn reads, and `Stop`. `StopAgent` has no session body from which to derive routing and remains production-only.

## Intake node behavior

The gateway decides where a request goes before it validates the body. That produces failure modes that look like outages but are routing problems.

| Symptom                                                    | What it means                                                                                      |
| ---------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| `503` `{"reason":"ServiceUnavailable"}` on `POST .../join` | The gate header was not recognised. The request died before validation, so the body is irrelevant. |
| `401` `Missing authorization header`                       | Routing worked; auth did not.                                                                      |
| `404` `no Route matched with those values`                 | The base URL path is wrong.                                                                        |
| `400` validation error                                     | You are past the gate. The header is fine and the body is the problem.                             |

The 503 is the one that misleads. It reads as a partner-side outage and invites waiting it out, when the fix is usually a one-line header change.

Observed on the `gemini-live` rollout in August 2026, when the gateway had not yet been configured to route on `agora-feature` and every request fell through to a 503. That was fixed server-side on 2026-08-09, so the 503 is not currently reproducible — the mapping is recorded here because it is the failure signature a newly provisioned preview family is most likely to hit first.

### Diagnosing without starting a billable agent

Two probes, neither of which allocates an agent:

1. **`GET .../v2/projects/{appId}/agents`** — a `200` proves host, auth, and routing are all healthy. If this succeeds while `join` fails, the problem is specific to the start path.
2. **A deliberately invalid start body** — send properties with no `llm` and no `mllm` at all. A `400` means you are past the gate; a `503` means you are not. This is what separates "my config is wrong" from "my header is wrong".

### Known gap (as of 2026-08-09)

A missing gate header is _intended_ to route to the production environment, where a preview config would fail. In practice an ungated start currently **succeeds** against the preview host, so the fallback is not observable from the client side.

The SDK cannot control the intake node, and this does not affect SDK users because the header is pinned. It matters only for callers hitting the REST API directly, who may get a request that appears to succeed while silently landing in the wrong environment. Flag it to the endpoint owner rather than working around it in SDK code.

## Session-scoped detection

`RequiredPreviewFeatures` reads the resolved request body rather than the vendor types, so hand-written configs and preset-enriched bodies are covered too. It keys on `asr.vendor`, using `previewASRVendors` in `preview_client.go`.

Routing state is stored on the `AgentSession`, not `AgoraClient`. One client can therefore start GA and preview sessions without leaking the preview host or gate header between them.

## Preview vendors

| Type           | Wire vendor             | Model                        |
| -------------- | ----------------------- | ---------------------------- |
| `NewGeminiSTT` | `asr.vendor = "gemini"` | `gemini-3.5-transcribe-live` |

`NewGeminiSTT` is an ASR stage, so it needs an LLM and a TTS vendor alongside it. The sample above uses Gemini LLM and Google TTS with the same Google API key. Mixing in other vendors is still valid; preview routing only triggers on `asr.vendor`.

### ASR language selection

Gemini Transcribe takes `params.language_codes`, an **array**, in place of the singular `params.language` other ASR vendors use.

```go
// Auto-detect (the default) — language_codes is not sent at all
vendors.NewGeminiSTT(vendors.GeminiSTTOptions{APIKey: apiKey})

// Commit to one language
vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
    APIKey:        apiKey,
    LanguageCodes: []string{"en-US"},
})

// Let the model choose between several
vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
    APIKey:        apiKey,
    LanguageCodes: []string{"en-US", "es-ES"},
})

// Auto-detect, stated outright
vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
    APIKey:        apiKey,
    LanguageCodes: []string{},
})
```

`LanguageCodes` is omitted from the request unless you supply it, which is how the provider spells auto-detect. Omitting the field and sending `[]` mean the same thing — nil versus empty is the distinction in Go: a `nil` slice omits `language_codes`, an empty-but-non-nil slice reaches the wire as `[]`.

`GeminiSTTOptions` has **no `Language` field**. The `Agent` always derives the top-level `asr.language` from the turn detection language — as it does for every STT vendor — so a vendor-level copy would be a no-op the builder overwrites. Set the interaction language on turn detection, and the transcription languages on `LanguageCodes`; they are separate settings and neither feeds the other.

| Setting                 | Where it belongs              | What it controls            |
| ----------------------- | ----------------------------- | --------------------------- |
| interaction language    | turn detection `Language`     | top-level `asr.language`    |
| transcription languages | `LanguageCodes` on the vendor | `asr.params.language_codes` |

`CustomVocabulary` biases recognition toward words the model would otherwise mis-hear — product names, jargon, proper nouns. It is omitted from the request entirely when nil.

```go
vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
    APIKey:           apiKey,
    CustomVocabulary: []string{"Agora", "Kubernetes"},
})
```

`WordTimestamp` is also omitted unless you set its pointer explicitly. Gemini does not support enabled word timestamps together with `CustomVocabulary`, so `ToConfig()` panics if both are requested. Explicit `WordTimestamp: Agora.Bool(false)` remains compatible with custom vocabulary.

```go
vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
    APIKey:        apiKey,
    WordTimestamp: Agora.Bool(true),
})
```

## The vendor type is not the whole wire shape

A vendor constructor emitting the right map is not proof of what ships, because the shared builder in `agentkit/core/properties.go` **also writes into the vendor config map** after the vendor is done with it — using the Agora schema spellings, which are correct for every GA provider but need not match what a preview route reads.

Every field the builder injects is listed below. Each one is a candidate for a silent mismatch on a preview route:

| Category | Key                                                                                         | Written from                                                                              | When                                     |
| -------- | ------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- | ---------------------------------------- |
| `mllm`   | `greeting_message`                                                                          | agent-level `Greeting`                                                                    | only when the key is absent              |
| `mllm`   | `failure_message`                                                                           | agent-level `FailureMessage`                                                              | only when the key is absent              |
| `mllm`   | `enable`                                                                                    | `WithMllm()`                                                                              | always                                   |
| `asr`    | `language`                                                                                  | turn detection `language`                                                                 | **always — overwrites any vendor value** |
| `llm`    | `system_messages`, `greeting_message`, `greeting_configs`, `failure_message`, `max_history` | agent-level `Instructions`, `Greeting`, `GreetingConfigs`, `FailureMessage`, `MaxHistory` | only when the key is absent              |

So when a preview provider documents a field that appears in that table, putting it on the vendor options struct is not the whole fix. One of three things applies:

- **The builder always overwrites it** (`asr.language`) — do not expose it on the options struct at all; it would be a field the builder silently discards. `GeminiSTTOptions` has no `Language` for exactly this reason.
- **It is an Agora engine field rather than the provider's** (`failure_message`) — leave it in the schema spelling.
- **The preview route spells it differently** — the translation belongs in `preview_client.go`, applied at session start, so it disappears with that file at GA rather than leaving a vestigial hook in the shared builder. Nothing in this release needs one, but a future preview family may.

### Verify against the request body, not the vendor output

`ToConfig()` returning the right map proves nothing about what ships, because the builder runs after it. Both checks are needed:

1. A unit test on the vendor constructor, for the keys the vendor owns.
2. An **end-to-end test that starts a session against a stub HTTP client and asserts on the captured request body** — the only check that sees the builder's injections. Every preview vendor has one in `agentkit/preview_test.go`.

The manual version is `Debug: true`, which logs the fully resolved body. Diff it against the payload the provider documented, key by key. A value sitting under a name the route ignores fails **silently** — no error, no validation complaint, the agent simply never greets. Go's `map[string]interface{}` bodies make this especially easy to miss: an unknown key is not a compile error, so nothing upstream of the gateway objects.

Wire parity across the three SDKs is a hard requirement, so a change here lands in Go, TypeScript, and Python together, verified by diffing the serialized bodies.

## Adding a future preview family

Everything preview-only lives in `agentkit/preview_client.go` and `agentkit/vendors/preview.go` so it can be deleted wholesale at GA. To add a family:

1. Add a `PreviewFeature*` constant in `preview_client.go` alongside `PreviewFeatureGeminiLive`. The value is what goes in the `agora-feature` header.
2. Add the vendor constructors to `vendors/preview.go`, returning the same config types as production vendors so the builder accepts them unchanged.
3. Register the detection keys — for an ASR family, the vendor name in `previewASRVendors` — so `RequiredPreviewFeatures()` recognises configs that need the new family.
4. **Diff the resolved request body against the payload the provider documented**, not the vendor constructor output — see [The vendor type is not the whole wire shape](#the-vendor-type-is-not-the-whole-wire-shape).
5. Add an end-to-end test that starts a session and asserts on the captured body, alongside the vendor unit test.

Generated files are overwritten on the next Fern run; `.fernignore` protects `agentkit/`. At GA, delete the preview files and move the vendor constructors into `vendors/stt.go`.

## Base URL

```
https://partner.ai.agora.io/preview/api/conversational-ai-agent
```

Request paths append to it exactly as they do in production — `POST {base}/v2/projects/{appId}/join`. The service path segment is part of the base: `https://partner.ai.agora.io/preview/api/` alone returns `404 no Route matched with those values`.

## Debug output

`Debug: true` marshals the resolved start request. The body is passed through `RedactSecrets` first, which replaces vendor API keys, the RTC token, and the App ID with `[REDACTED]` while leaving model names, voices, and instructions readable.

Empty strings are left visible on purpose: `""` is the signature of an unset environment variable, and hiding it would disguise the exact misconfiguration the debug output exists to surface.

## Related

- [Regional Routing](./regional-routing.md) — the production domain pool preview sessions bypass
- [Error Handling](./error-handling.md) — API error handling
