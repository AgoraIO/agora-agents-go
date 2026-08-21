---
sidebar_position: 10
title: Preview Endpoint
description: How NewAgoraPreviewClient routes to preview providers, and how the gateway's agora-feature gate header behaves.
---

# Preview Endpoint

Some providers ship on a preview gateway before they reach the production Conversational AI environment. `NewAgoraPreviewClient` returns a client pointed at that gateway with the header the gateway routes on pinned to every request.

Everything in this guide is temporary by design. When a preview provider goes GA it moves to the production gateway, and the corresponding entry disappears.

## Using the preview client

```go
client := agentkit.NewAgoraPreviewClient(agentkit.AgoraPreviewClientOptions{
    AgoraClientOptions: agentkit.AgoraClientOptions{
        Area:           option.AreaUS,
        AppID:          os.Getenv("AGORA_APP_ID"),
        AppCertificate: os.Getenv("AGORA_APP_CERTIFICATE"),
    },
})

agent := agentkit.NewAgent(client).
    WithMllm(vendors.NewGeminiCleverChatter(vendors.GeminiCleverChatterOptions{
        GeminiPreviewMLLMOptions: vendors.GeminiPreviewMLLMOptions{
            APIKey: os.Getenv("GOOGLE_API_KEY"),
        },
        ThinkingLevel: vendors.GeminiThinkingLevelMedium,
    }))

session := agent.CreateSession(agentkit.CreateSessionOptions{
    Channel:    "demo",
    AgentUID:   "1",
    RemoteUIDs: []string{"100"},
})
agentID, err := session.Start(ctx)
```

`NewAgoraPreviewClient` returns an `*AgoraClient`, not a distinct type, so it is accepted everywhere a production client is. Auth, the `Agent` builder, `AgentSession`, and every vendor type work exactly as they do with `NewAgoraClient`. Only two things differ: where requests go, and the gate header.

The sample above is compiled as `ExampleNewAgoraPreviewClient` in `agentkit/preview_example_test.go`, so it cannot drift from the API.

## The gate header

The gateway routes preview traffic on a single request header:

```
agora-feature: gemini-live
```

| Constant | Value |
|---|---|
| `PreviewFeatureHeader` | `agora-feature` |
| `PreviewFeatureGeminiLive` | `gemini-live` |

Select the families a client requests with `Features`:

```go
client := agentkit.NewAgoraPreviewClient(agentkit.AgoraPreviewClientOptions{
    AgoraClientOptions: base,
    Features:           []string{agentkit.PreviewFeatureGeminiLive}, // default
})
```

### The header is not overridable

`option.WithHTTPHeader` **replaces the entire header map** rather than merging into it. That is normal Go SDK behavior, but it means a per-call header option would silently drop the gate — and a preview request that loses the header is not rejected, it routes to the production environment where the preview providers do not exist.

So the gate is applied below the option layer, in a wrapped `core.HTTPClient`:

```go
func (p *previewGateClient) Do(req *http.Request) (*http.Response, error) {
    req.Header.Set(PreviewFeatureHeader, p.features)
    return p.inner.Do(req)
}
```

A custom `HTTPClient` passed in `AgoraPreviewClientOptions` is **wrapped, not replaced**, so supplying your own transport keeps the gate intact:

```go
client := agentkit.NewAgoraPreviewClient(agentkit.AgoraPreviewClientOptions{
    AgoraClientOptions: base,
    HTTPClient:         myInstrumentedClient, // still gated
})
```

The header rides every verb, not just `Start` — `Say`, `Interrupt`, `Stop`, the `Agents` sub-client, and `StopAgent`. `StopAgent` matters in particular, because services typically build a fresh client in an end-call handler rather than holding the session that started the agent.

## Intake node behavior

The gateway decides where a request goes before it validates the body. That produces failure modes that look like outages but are routing problems.

| Symptom | What it means |
|---|---|
| `503` `{"reason":"ServiceUnavailable"}` on `POST .../join` | The gate header was not recognised. The request died before validation, so the body is irrelevant. |
| `401` `Missing authorization header` | Routing worked; auth did not. |
| `404` `no Route matched with those values` | The base URL path is wrong. |
| `400` validation error | You are past the gate. The header is fine and the body is the problem. |

The 503 is the one that misleads. It reads as a partner-side outage and invites waiting it out, when the fix is usually a one-line header change.

Observed on the `gemini-live` rollout in August 2026, when the gateway had not yet been configured to route on `agora-feature` and every request fell through to a 503. That was fixed server-side on 2026-08-09, so the 503 is not currently reproducible — the mapping is recorded here because it is the failure signature a newly provisioned preview family is most likely to hit first.

### Diagnosing without starting a billable agent

Two probes, neither of which allocates an agent:

1. **`GET .../v2/projects/{appId}/agents`** — a `200` proves host, auth, and routing are all healthy. If this succeeds while `join` fails, the problem is specific to the start path.
2. **A deliberately invalid start body** — send properties with no `llm` and no `mllm` at all. A `400` means you are past the gate; a `503` means you are not. This is what separates "my config is wrong" from "my header is wrong".

### Known gap (as of 2026-08-09)

A missing gate header is *intended* to route to the production environment, where a preview config would fail. In practice an ungated start currently **succeeds** against the preview host, so the fallback is not observable from the client side.

The SDK cannot control the intake node, and this does not affect SDK users because the header is pinned. It matters only for callers hitting the REST API directly, who may get a request that appears to succeed while silently landing in the wrong environment. Flag it to the endpoint owner rather than working around it in SDK code.

## Guard against the wrong client

Preview providers used with a production client fail before any request is sent:

```
this agent uses preview providers that require the preview endpoint (gemini-live):
create the session with NewAgoraPreviewClient instead of NewAgoraClient, and make
sure its Features include: gemini-live
```

`AssertPreviewSupport` reads the resolved request body rather than the vendor types, so hand-written configs are covered too. `mllm.vendor` is `"gemini"` for both the production Gemini Live provider and the preview ones, so the vendor name alone settles nothing — see [How a config is recognised as preview](#how-a-config-is-recognised-as-preview) for the two signals it uses instead.

`NewSession` fills the session's feature list from the bound client via an optional interface, so a production client simply reports none and the guard rejects preview configs.

## Preview vendors

| Type | Wire vendor | Model |
|---|---|---|
| `NewGeminiWalkieTalkie` | `mllm.vendor = "gemini"` | `walkie-talkie` |
| `NewGeminiCleverChatter` | `mllm.vendor = "gemini"` | `clever-chatter` |

MLLM sessions have no ASR, LLM, or TTS stage — the model handles the whole exchange end to end, so `TranscribeAgent` and `TranscribeUser` are the only source of transcripts.

`ThinkingLevel` is exclusive to clever-chatter, which is why it is a separate type rather than a field on the shared options — the distinction is structural rather than a runtime check.

The preview MLLMs differ from the production `GeminiLive` in two ways the gateway requires: `params.api_key` is repeated inside `params`, and `URL` defaults to the Gemini Developer API host rather than a WebSocket endpoint.

Language is set with `LanguageCodes`, which serializes to `params.language_codes` and is omitted when nil. Note the plural: the singular `mllm.params.language` belongs to xAI Grok in the Agora schema, and the [production Gemini Live provider](https://docs.agora.io/en/ai/models/mllm/gemini) sends no language field at all — so the preview MLLMs never send it.

The greeting field is `Greeting`, not `GreetingMessage`, and it serializes to `greeting`. The new Gemini preview models read that spelling; `greeting_message` is what the production providers use.

Both placements work — `agentkit.WithGreeting(...)` and the vendor's `Greeting` field both reach the model. That is not automatic: the `Agent` writes an agent-level greeting to `mllm.greeting_message`, the production spelling, which these models ignore. `ApplyPreviewShape` retargets it at session start — see [Preview wire-shape translation](#preview-wire-shape-translation).

### The MLLM rename in flight

Google is renaming the MLLM codenames to conventional model IDs:

| Type | Model sent today | Post-rename model |
|---|---|---|
| `NewGeminiWalkieTalkie` | `walkie-talkie` | `gemini-3.5-live` |
| `NewGeminiCleverChatter` | `clever-chatter` | `gemini-3.5-live-thinking` |

The types still send the codenames. As of 2026-08-19 the new IDs are propagating through Google's fleet and are not live everywhere, while the codenames work — Google will confirm when the rename lands. Both spellings are in the known-model set, and a name that is in neither is still recognised by its envelope (see [How a config is recognised as preview](#how-a-config-is-recognised-as-preview)), so `Model` opts a single instance into a new name safely:

```go
vendors.NewGeminiWalkieTalkie(vendors.GeminiPreviewMLLMOptions{
    APIKey: apiKey,
    Model:  vendors.GeminiPreviewModel35Live,
})
```

When the rename lands, repoint `GeminiPreviewDefaultModelWalkieTalkie` / `GeminiPreviewDefaultModelCleverChatter` in `vendors/preview.go`. That is the entire cutover.

## Preview wire-shape translation

The preview route spells two MLLM fields differently from the Agora schema. Where the vendor types own the field they simply emit the preview spelling, but `greeting` is different: the shared `Agent` builder writes it, using the production spelling that is correct for every GA vendor.

So `ApplyPreviewShape` runs at session start, immediately after `AssertPreviewSupport`, and retargets those fields on a preview config:

| Written by the builder | Sent to the preview route |
|---|---|
| `mllm.greeting_message` | `mllm.greeting` |

The vendor's own value wins; the production-spelled one is the fallback. That also migrates a hand-written `greeting_message` onto the preview key, so an existing MLLM config keeps working after only swapping the model.

`mllm.failure_message` is deliberately **not** translated — it is an Agora engine feature rather than a Gemini one, so it keeps its schema spelling.

A production MLLM is left untouched, so the translation is safe to run unconditionally. Teaching the shared builder about preview spellings was the alternative; keeping it here means the whole thing disappears with `preview_client.go` at GA instead of leaving a vestigial hook behind.

### How a config is recognised as preview

Two independent signals, either one sufficient:

1. **`mllm.params.model` is in the known-model set** — the four names in `preview_client.go` (both codenames and both post-rename IDs).
2. **The config carries the preview envelope** — `params.api_key` set *and* `url` on the Gemini Developer API host. This is exactly what the preview vendor types emit and what the production `GeminiLive` vendor does not: it leaves `params.api_key` unset and defaults `url` to `""` or a WebSocket endpoint.

The second signal exists because the first one fails silently on its own. Overriding `Model` is the documented escape hatch for a rename that lands ahead of an SDK release, so a name this SDK has never seen is reachable by following the advice above — and with model-name matching alone, that config would quietly stop getting `ApplyPreviewShape`. No error, no warning: the greeting would land in `greeting_message`, which these models ignore, and the agent would simply never greet.

Adding a newly announced model name to the set is still worth doing — it keeps recognition working for hand-written configs, which carry no envelope. But forgetting to is no longer a silent failure.

### The vendor type is not the whole wire shape

This translation exists because of a bug worth not repeating. `NewGeminiWalkieTalkie` emitted `greeting` correctly and nothing in `vendors/preview.go` was wrong — but an agent-level greeting still never reached the model, because the shared builder in `agentkit/core/properties.go` **also writes into the vendor config map**, using production spellings, after the vendor is done with it.

Every field the builder injects is listed below. Each one is a candidate for the same bug on a preview route:

| Category | Key | Written from | When |
|---|---|---|---|
| `mllm` | `greeting_message` | agent-level `Greeting` | only when the key is absent |
| `mllm` | `failure_message` | agent-level `FailureMessage` | only when the key is absent |
| `mllm` | `enable` | `WithMllm()` | always |
| `asr` | `language` | turn detection `language` | **always — overwrites any vendor value** |
| `llm` | `system_messages`, `greeting_message`, `greeting_configs`, `failure_message`, `max_history` | agent-level `Instructions`, `Greeting`, `GreetingConfigs`, `FailureMessage`, `MaxHistory` | only when the key is absent |

So when a preview provider documents a field that appears in that table, putting it on the vendor options struct is not the whole fix. One of three things applies:

- **The preview route spells it differently** — add an entry to `previewMLLMFieldRenames` in `preview_client.go`. One entry, and both the vendor-level and agent-level placements start working.
- **The builder always overwrites it** (`asr.language`) — do not expose it on the options struct at all; it would be a field the builder silently discards.
- **It is an Agora engine field rather than the provider's** (`failure_message`) — leave it in the schema spelling.

### Verify against the request body, not the vendor output

`ToConfig()` returning the right map proves nothing about what ships, because the builder runs after it. Both checks are needed:

1. A unit test on the vendor constructor, for the keys the vendor owns.
2. An **end-to-end test that starts a session against a stub HTTP client and asserts on the captured request body** — the only check that sees the builder's injections and the preview translation. Every preview vendor has one in `agentkit/preview_test.go`.

The manual version is `Debug: true`, which logs the fully resolved body. Diff it against the payload the provider documented, key by key. A value sitting under a name the route ignores fails **silently** — no error, no validation complaint, the agent simply never greets. Go's `map[string]interface{}` bodies make this especially easy to miss: an unknown key is not a compile error, so nothing upstream of the gateway objects.

Wire parity across the three SDKs is a hard requirement, so a change here lands in Go, TypeScript, and Python together, verified by diffing the serialized bodies.

## Adding a future preview family

Everything preview-only lives in `agentkit/preview_client.go` and `agentkit/vendors/preview.go` so it can be deleted wholesale at GA. To add a family:

1. Add a `PreviewFeature*` constant in `preview_client.go` alongside `PreviewFeatureGeminiLive`. The value is what goes in the `agora-feature` header.
2. Add the vendor constructors to `vendors/preview.go`, returning the same config types as production vendors so the builder accepts them unchanged.
3. Register the detection keys — for an MLLM family, the model names in `previewMLLMModels` plus an envelope check — so `RequiredPreviewFeatures()` recognises configs that need the new family.
4. **Diff the resolved request body against the payload the provider documented**, not the vendor constructor output. Anything the shared builder writes under a different name belongs in `previewMLLMFieldRenames` — see [The vendor type is not the whole wire shape](#the-vendor-type-is-not-the-whole-wire-shape).
5. Add an end-to-end test that starts a session and asserts on the captured body, alongside the vendor unit test.

Generated files are overwritten on the next Fern run; `.fernignore` protects `agentkit/`. At GA, delete the preview files and move the vendor constructors into `vendors/mllm.go`.

## Base URL

```
https://partner.ai.agora.io/preview/api/conversational-ai-agent
```

Request paths append to it exactly as they do in production — `POST {base}/v2/projects/{appId}/join`. The service path segment is part of the base: `https://partner.ai.agora.io/preview/api/` alone returns `404 no Route matched with those values`.

Override the host for a staging gateway with `BaseURL`.

## Debug output

`Debug: true` marshals the resolved start request. The body is passed through `RedactSecrets` first, which replaces vendor API keys, the RTC token, and the App ID with `[REDACTED]` while leaving model names, voices, and instructions readable.

Empty strings are left visible on purpose: `""` is the signature of an unset environment variable, and hiding it would disguise the exact misconfiguration the debug output exists to surface.

## Related

- [Regional Routing](./regional-routing.md) — the production domain pool the preview client bypasses
- [Error Handling](./error-handling.md) — API error handling
