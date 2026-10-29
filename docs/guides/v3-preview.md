---
sidebar_position: 10
title: Conversational AI v3 preview
description: Manage agents, sessions, secrets, numbers, and campaigns with the v1 REST path from the v3 preview SDK.
---

# Conversational AI v3 preview (P1)

In v3, an **agent** is a saved configuration identified by `agent_id`. A running instance is a **session**, identified by `agent_session_id`. Use `Agent.CreateEphemeralSession(...).Start(ctx)` to explicitly start an unsaved Platform session. `Agent.CreateSession` retains automatic routing for compatibility. `Save` creates or updates a saved configuration; it does not start a session. The Go import path remains `github.com/AgoraIO/agora-agents-go/v2`: its `/v2` suffix is the module's semantic major version, independent of the REST API version.

## Quickstart

Set the project App ID and a REST token. The default API origin is `https://preview.ai.agora.io`; the SDK appends `/conversational-ai/v1/projects/{appid}`. `V3Options.BaseURL` can override the origin for local gateways.

```sh
export AGORA_APP_ID='your-project-app-id'
export AGORA_REST_TOKEN='your-rest-token'
# Optional origin override; omit to use https://preview.ai.agora.io
export AGORA_V3_BASE_URL='https://preview.ai.agora.io'
```

Save as `main.go` and run `go run main.go`. This makes a read-only list request:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/AgoraIO/agora-agents-go/v2/agentkit"
    "github.com/AgoraIO/agora-agents-go/v2/option"
    "github.com/AgoraIO/agora-agents-go/v2/v3"
)

func main() {
    c := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
        Area: option.AreaUS,
        AppID: os.Getenv("AGORA_APP_ID"),
        Token: os.Getenv("AGORA_REST_TOKEN"),
        V3: &agentkit.V3Options{BaseURL: os.Getenv("AGORA_V3_BASE_URL")},
    })
    page, err := c.V3.Agents.List(context.Background(), v3.ListOptions{Limit: 10})
    if err != nil { log.Fatal(err) }
    fmt.Printf("count=%d next=%v request=%s\n", page.Body.Data.Count,
        page.Body.NextCursor, page.RequestID)
}
```

A client can use an Agora REST `Token` or set `AppCertificate`; AgentKit signs a fresh ConvoAI token for each request in app-credentials mode. Basic authentication and internal `X-Studio-*` / `X-Consumer-Username` headers are rejected. The `AGORA_AGENTS_V3_BASE_URL` environment variable and `V3Options.BaseURL` override the default origin; the explicit option takes precedence. The SDK does not hard-code regional availability; follow the preview rollout guidance for the target project.

## AgentKit workflows

### Save and update a stored agent

The following body assumes `ctx context.Context`, `c *agentkit.AgoraClient`, and imports for `agentkit` and `agentkit/vendors`:

```go
agent := agentkit.NewAgent(c,
    agentkit.WithVariable("resolved", agentkit.VariableDefinition{
        Type: "bool", Extract: true, Description: "Was the issue resolved?",
    }),
).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
    Model: "gpt-5.6-luna",
})).WithInstructions("Help the caller briefly and politely.")

// Create once; persist the returned ID.
saved, err := agent.Save(ctx, agentkit.SaveOptions{Name: "Support"})
if err != nil { return err }

// Later, update the same stored configuration with a new immutable config.
updatedAgentConfig := agent.WithVariable("priority", agentkit.VariableDefinition{
    Type: "str", Description: "What should the agent prioritize?",
})
saved, err = saved.Update(ctx, updatedAgentConfig)
if err != nil { return err }
```

`WithVariable` also exists as `agent.WithVariable(name, def)`, returning a new immutable builder. Names match `^[a-z][a-z0-9_]{0,63}$`; types are `str`, `int`, `float`, `bool`, `date`, and `datetime`. `Options`, when provided, must contain 1–100 unique JSON scalar values. Descriptions are limited to 2000 Unicode characters; `Extract: true` requires a nonempty description. The no-error builder methods panic for invalid definitions. `agentkit.GetSavedAgent(ctx, c, id)` loads a handle; `saved.Delete(ctx)` deletes it. `saved.Update(ctx, config)` reuses its ID and name and returns a handle with the fresh ETag. Handles expose `ID`, `Name`, `ETag`, and `Raw` (`*v3.AgentsClient`). The old `Deploy`, `DeployOptions`, `DeployedAgent`, and `GetDeployedAgent` names remain available as compatibility aliases. Agent creation is not an upsert yet: repeated POSTs can create distinct agents. Agent PATCH uses JSON merge patch with conditional ETags.

`Save` without an ID POSTs a new agent. Save its returned ID: names are not unique upsert keys and agent POST has no supported idempotency key. `SavedAgent.Update` GETs the current ETag, then PATCHes. It sends recursive `null`s to clear readable stale fields, omitting server-owned IDs, timestamps, and resolved credential metadata. A 412 becomes `ErrPreconditionFailed` and retains the underlying v3 error; save does not retry the conflict automatically.

For cascaded saves, an explicit `TurnDetection.Language` is preserved as `pipeline.asr.language`, taking precedence over the ASR vendor language as on v2. Otherwise the ASR vendor language is promoted to the block-level field, and Ares defaults to `en-US`. Without an explicit STT vendor, it uses the existing global `ares` default with managed credentials; endpoint availability remains a backend concern. The preview API supports cascading pipelines only; MLLM configurations cannot be saved or used with preview session features. Existing MLLM sessions remain supported through the original session API.

Legacy greeting settings `mode: "single_every"` and `mode: "single_first"` map to `on: "each_join"` and `on: "first_join"`. An explicit `on` takes precedence. Unknown legacy modes and playback settings without an actual greeting are rejected before a request. Audio download timeout and PCM sample-rate settings require an audio greeting; supplying them for a text greeting is rejected. Empty message content remains valid. Variable descriptions allow up to 2000 Unicode characters, rather than 2000 UTF-8 bytes.

### Secrets and numbers

Create the secret set before saving a provider that uses `SecretRef`. Values are write-only. Secret and number POSTs do not support `Idempotency-Key`; pass `""` to their compatibility string arguments. A nonempty value is rejected locally. New secret set/key names use ASCII letters, digits, and underscores; legacy hyphenated keys can only be deleted with a null-valued `Secrets.Update`. Session and campaign create/action calls do accept optional idempotency keys, retained for 24 hours. This code also assumes `os` and `agentkit` imports:

```go
secret, err := c.Secrets.Create(ctx, agentkit.SecretCreate{
    Name: "prod",
    Secrets: map[string]string{"openai": os.Getenv("OPENAI_API_KEY")},
}, "")
if err != nil { return err }
_ = secret.Body.Keys // Values are never returned.

number, err := c.Numbers.Create(ctx, agentkit.NumberCreate{
    Number: "+12025550123",
    SipTrunk: agentkit.SipTrunk{Hostname: "sip.example.com", Transport: "tls"},
}, "")
if err != nil { return err }

maxCall := 900
_, err = saved.AssignNumber(ctx, number.Body, agentkit.AssignNumberOptions{
    MaxCallDurationSeconds: &maxCall,
})
if err != nil { return err }
_, err = saved.UnassignNumber(ctx, number.Body.NumberID)
if err != nil { return err }
```

Assignment accepts `number_id`, `agentkit.Number`, `*agentkit.Number`, or a formatted E.164 value. E.164 lookup scans this project's pages because the API has no number filter, then GETs the matched number for an ETag. Assignment PATCHes only the inbound agent and explicitly supplied call policy fields; one 412 causes one fresh GET and retry. A missing project number returns `ErrNumberNotFound`. Number POST rejects a nonempty idempotency argument because the API does not accept that header. `SecretRef("prod", "openai")` yields `$secrets.prod.openai`; passing it to the v2 session path returns `ErrSecretRefNotSupported`.

`saved.ListNumbers(ctx)` returns the numbers whose `inbound.agent` equals the saved agent ID. It scans **every number page in the project**, with no per-number GETs, because staging has no agent filter. Its request cost grows with the project's total number count; use it deliberately for large projects. The v3 session list is also project-wide, but the published list and get responses provide no `agent_id` association. There is no `SavedAgent.ListSessions` helper because matching the nonunique `agent_name` would be unreliable.

## Low-level reference: all 32 operations

Import `github.com/AgoraIO/agora-agents-go/v2/v3`. `agentkit.AgoraClient.V3` already has the App ID. The root `client.NewClient(...)` has no App ID option: use `api.V3.ForAppID(appID)` before calling its resources, or construct `v3.NewClient(appID, opts...)` directly. An empty App ID or resource ID fails locally before HTTP.

Every method takes `ctx context.Context` first and returns `(*v3.Response[T], error)`. Below are its **exact remaining arguments**, response body `T`, and relevant headers. A `Response[T]` contains `Body`, a detached `Header http.Header` with all response headers, raw `ETag`, `RequestID` (from `Agora-Request-Id`), and `StatusCode`. Use `response.Header.Get("Location")` or `response.Header.Get("Idempotent-Replayed")` to inspect those headers. `ifMatch` must be the literal ETag, including quotes if the server sent them. `""` omits optional header arguments.

| Method after `c.V3` | Arguments after `ctx` | Body `T` | Header/filter note |
|---|---|---|---|
| `Agents.Create` | `body v3.AgentCreate` | `v3.Agent` | No idempotency key |
| `Agents.List` | `opts v3.ListOptions` | `v3.Page[v3.AgentListItem]` | limit, cursor, order |
| `Agents.Get` | `id string` | `v3.Agent` | `agent_id` |
| `Agents.Update` | `id, ifMatch string, patch v3.AgentPatch` | `v3.Agent` | Required `If-Match` |
| `Agents.Delete` | `id string` | `v3.Deleted` | `agent_id` |
| `Secrets.Create` | `body v3.SecretCreate, idempotencyKey string` | `v3.SecretSet` | Pass `""`; a nonempty key is rejected |
| `Secrets.List` | `opts v3.ListOptions` | `v3.Page[v3.SecretSet]` | limit, cursor, order |
| `Secrets.Get` | `name string` | `v3.SecretSet` | Secret set name |
| `Secrets.Update` | `name, ifMatch string, body v3.SecretUpdate` | `v3.SecretSet` | Optional `If-Match`; nil values delete keys |
| `Secrets.Replace` | `name, ifMatch string, body v3.SecretReplace` | `v3.SecretSet` | Required `If-Match` |
| `Secrets.Delete` | `name, ifMatch string` | `v3.Deleted` | Required `If-Match` |
| `Secrets.SetKey` | `name, key string, body v3.SecretKeyWrite, opts v3.SecretKeyOptions` | `v3.SecretKey` | Optional `If-Match` and `If-None-Match` |
| `Secrets.DeleteKey` | `name, key, ifMatch string` | `v3.Deleted` | Optional `If-Match` |
| `Numbers.Create` | `body v3.NumberCreate, idempotencyKey string` | `v3.Number` | Pass `""`; a nonempty key is rejected |
| `Numbers.List` | `opts v3.ListOptions` | `v3.Page[v3.NumberListItem]` | limit, cursor, order |
| `Numbers.Get` | `id string` | `v3.Number` | `number_id` |
| `Numbers.Update` | `id, ifMatch string, patch v3.NumberPatch` | `v3.Number` | Required `If-Match` |
| `Numbers.Delete` | `id string` | `v3.Deleted` | `number_id` |
| `Sessions.Start` | `body v3.SessionStartRequest, opts ...v3.SessionWriteOptions` | `v3.Session` | Saved `agent_id`, values, labels, RTC transport, lifecycle, data policy, optional idempotency key |
| `Sessions.StartEphemeral` | `body v3.SessionStartEphemeralRequest, opts ...v3.SessionWriteOptions` | `v3.Session` | Inline agent config, values, labels, transport, lifecycle, data policy, optional idempotency key |
| `Sessions.Stop` | `id string` | `map[string]any` | `agent_session_id`; no request body; 404 and 409 errors are preserved |
| `Sessions.List` | `opts v3.SessionListOptions` | `v3.Page[v3.Session]` | channel, comma-separated status, started_after/before, limit, cursor |
| `Sessions.Get` | `id string` | `v3.Session` | `agent_session_id` |
| `Sessions.Update` | `id string, patch v3.SessionPatch` | `v3.Session` | Optional RTC token and message/LLM params overrides only |
| `Campaigns.Create` | `body v3.CampaignCreate, opts ...v3.CampaignWriteOptions` | `v3.Campaign` | Optional `Idempotency-Key` |
| `Campaigns.List` | `opts v3.CampaignListOptions` | `v3.Page[v3.Campaign]` | status, limit, cursor, order |
| `Campaigns.Get` | `id string` | `v3.Campaign` | `campaign_id` |
| `Campaigns.Update` | `id, ifMatch string, body v3.CampaignPatch` | `v3.Campaign` | Required `If-Match`; 412 exposes `Latest(ctx)` |
| `Campaigns.Delete` | `id string` | `v3.CampaignDeleteResponse` | Only final campaigns can be deleted |
| `Campaigns.Pause` | `id string, opts ...v3.CampaignWriteOptions` | `v3.Campaign` | Optional `Idempotency-Key` |
| `Campaigns.Resume` | `id string, opts ...v3.CampaignWriteOptions` | `v3.Campaign` | Optional `Idempotency-Key` |
| `Campaigns.Cancel` | `id string, opts ...v3.CampaignWriteOptions` | `v3.Campaign` | Optional `Idempotency-Key` |

`v3.ListOptions` has `Limit int`, `Cursor string`, and `Order string`. `v3.SessionListOptions` embeds it and adds `Channel`, `Status`, `StartedAfter *int64`, and `StartedBefore *int64`; session list does not send `Order`. Each page has `Data.Count`, `Data.List`, `HasMore`, and `NextCursor`. `Count` is the page count, not a promised project total. A cursor loop:

The staging defaults and limits differ by resource: agents and secrets default to 50 items and allow 1–200; numbers default to 50 and allow 1–100; sessions default to 20 and require at least 1. Agents, secrets, and numbers default to `create_ts.desc` order. Leaving `SessionListOptions.Status` empty sends no filter, so the server defaults to **`RUNNING`**, not all sessions. Use comma-separated status values such as `RUNNING,STOPPED,FAILED` to filter by multiple states; pagination cursors apply to the combined result.

```go
cursor := ""
for {
    page, err := c.V3.Agents.List(ctx, v3.ListOptions{Limit: 100, Cursor: cursor})
    if err != nil { return err }
    for _, item := range page.Body.Data.List { fmt.Println(item.AgentID, item.AgentName) }
    if !page.Body.HasMore || page.Body.NextCursor == nil || *page.Body.NextCursor == "" { break }
    cursor = *page.Body.NextCursor
}
```

Conditional writes use ETags from GET responses. A stale tag returns `*v3.PreconditionFailedError`. `Latest(ctx)` performs a **new GET only when called** and returns `(any, string, error)` for latest body and ETag; the body is a `v3.Agent`, `v3.SecretSet`, or `v3.Number` depending on the resource. For example:

```go
current, err := c.V3.Agents.Get(ctx, agentID)
if err != nil { return err }
_, err = c.V3.Agents.Update(ctx, agentID, current.ETag,
    v3.AgentPatch{"instructions": "Updated instructions"})
var stale *v3.PreconditionFailedError
if errors.As(err, &stale) {
    latest, tag, readErr := stale.Latest(ctx)
    if readErr != nil { return readErr }
    fmt.Printf("latest ETag: %s; resource: %#v\n", tag, latest)
}
return err
```

`AgentPatch` and `NumberPatch` are map types; explicit `nil` sends a merge-patch `null`. PATCH uses `application/json`, matching staging. All v3 writes make one transport attempt; the caller decides whether retrying a failed write is safe.

## Session options and campaign limits

For deterministic Platform routing, use `Agent.CreateEphemeralSession` even when the configuration has no variables or secrets. AgentKit routes a saved handle through `SavedAgent.CreateSession`. Inline configs with secret references or variable definitions use the ephemeral endpoint automatically; plain inline configs, including token-backed `/join` sessions, retain v2 behavior. New v3 options such as `Values`, `IdempotencyKey`, `DataPolicy`, RTC encryption, audio scenario, geofence, and maximum duration select ephemeral v3 sessions. `ClientReference` remains only as a deprecated compatibility field and is rejected on v3 starts.

`AgentSession.Update(ctx, *Agora.UpdateAgentsRequestProperties)` keeps the same call shape on v2 and v3. For v3 sessions, only token, LLM system messages, and complete LLM params replacements are sent; other legacy update fields are rejected. The versioned low-level client accepts `v3.SessionPatch` directly.

`IdleTimeout` and `MaxDuration` are public seconds and serialize to milliseconds. Idle timeout accepts 0–3600 seconds; max duration accepts 1–3600 seconds. `GracefulStop.TimeoutMS` already uses wire milliseconds (0–120000). RTC subscriber UID is required; agent UID may be omitted for server assignment. When both are explicit, their UID types must match. Numeric IDs may be digits; string UIDs are also accepted. Response statuses normalize to uppercase; start accepts both current `200/create_ts` and rollout-era `201/created_at` responses.

Campaign contacts are inline at create time (1–50,000 contacts; the create payload limit is 10 MiB); there is no contacts API. Campaign pacing supports `MaxConcurrent` and `MaxCallsPerSecond`, with no retry fields or SDK dialing/retry algorithm. Campaign create and pause/resume/cancel accept optional idempotency keys, which the service retains for 24 hours. `CampaignCreate.DataPolicy` controls retention; campaign end-call policy includes voicemail.

## Compatibility and limits

`Save` collects unsupported v2 fields into one `ErrUnsupportedField` before HTTP. Current examples are `sal`, `advancedFeatures`, agent-level `geofence`, `labels`, `rtc`, `audioScenario`, `pipelineId`, unsupported `parameters`, nonempty `template_variables` (v3 variables have no defaults), unverified provider fields, and multipart credentials that cannot map to one API key. Geofence, audio scenario, farewell, and opt-out map to session transport, lifecycle, and data policy when starting an eligible inline v3 session; they are not stored agent properties. Managed credentials omit endpoint URLs from the v3 payload; BYOK credentials retain explicitly configured URLs.

Provider credentials use `credential.fields` with provider-specific keys. Saved-agent writes accept only `$secrets.<set>.<key>` references; ephemeral inline session configs may use direct values. Secret values are write-only, so GET does not reconstruct credentials. Updates clear stale credential fields with merge-patch nulls alongside readable stale values. Sessions can be started from saved agent IDs or ephemeral inline configs, stopped, retrieved, and updated with the limited token/messages/LLM params patch. Optional typed `Session.Analysis` exposes extraction status, result, error, and completion metadata when returned by the server. Rollout responses may omit analysis. AgentKit exposes these through the same `AgentSession` handle; v3 session IDs remain distinct from v2 agent IDs, so legacy actions, history, turns, and `Raw` clients are unavailable for v3 sessions. Labels are supported on both saved-agent and ephemeral starts. The low-level `Campaigns` resource covers create, list, retrieve, conditional update, delete, pause, resume, and cancel; it does not implement a dialing or retry state machine. Contact ingestion APIs, per-contact results, events, and webhooks are not included. v2 telephony and phone-number clients keep their original semantics; they do not manage v3 project-scoped numbers.

## Maintenance and verification boundary

The v3 package is **handwritten against the repository-root `openapi.json` / `openapi.yaml` contract snapshot** and protected by `.fernignore`; Fern has not generated it. A contract test checks key routes and schema fields. Keep this guide and the types in sync as the schema evolves. Local tests cover paths, headers, pagination, conditional errors, AgentKit mapping, and number assignment. This preview has **not** been validated against a live staging or production gateway in this workspace. Regional availability follows the preview rollout; the SDK adds no region guard. Nested patch replacement semantics follow the checked-in schema. Managed credentials omit provider URLs; BYOK requests retain the configured URL. SIP trunk `allowed_ips` accepts the schema-defined IPv4 address/CIDR entries, `*`, and explicit empty-list clearing.

## Install the preview

Use the slashless preview branch (Go rejects a slash in a branch version query):

```sh
go get github.com/AgoraIO/agora-agents-go/v2@v2.14.0-alpha.3
```

The alpha.3 preview uses tag `v2.14.0-alpha.3`, compatible with the existing `/v2` module path. This checkout prepares that release; the command is available after the tag is published. The slashless `v3-api-preview` branch remains available for testing before publication.

## Explicit sessions and outbound calls

```go
session := agent.CreateEphemeralSession(agentkit.CreateSessionOptions{
    Channel: "support", AgentUID: "1002", RemoteUIDs: []string{"1001"},
})
_, err := session.Start(ctx)
if err != nil { return err }

call := saved.Call(agentkit.CallOptions{
    From: number.Body.NumberID, To: "+12025550123",
    CallPolicy: &agentkit.CallPolicy{
        EndCall: &agentkit.EndCallPolicy{OnUserRequest: true},
    },
})
_, err = call.Start(ctx)
if err != nil { return err }
```

`Agent.Call` also supports an unsaved configuration. Call returns a configured session; `Start`, `GetInfo`, and `Stop` share the existing lifecycle. `From` is a registered number ID, and `To` is an E.164 number. RTC options cannot be combined with telephone calls. `AssignNumberOptions.CallPolicy` accepts `InboundCallPolicy`, including the four supported `InboundEndCallPolicy` conditions. Existing duration shortcuts override the corresponding nested policy values when supplied.

For Platform RTC sessions, an explicit token is preserved. When a nonzero agent UID and App Certificate are available, AgentKit generates the RTC token using its existing token helpers. With an omitted or auto-assigned UID, it leaves token generation to the service. Generated tokens are retained for keyed retries so the request body stays stable; a new unkeyed session after a successful session refreshes the generated token, while failed start retries keep it.

`SecretSet.Keys` contains `SecretKey` values with optional `CreateTS`/`UpdateTS`; the client also accepts the rollout-era array of plain names and normalizes each name to `SecretKey{Key: name}`. `SecretKeyWrite`, `SecretKeyOptions`, interruption sub-configurations, VAD/semantic constants, and pointer helpers are available from `agentkit`.

Provider mapping uses explicit vendor defaults. OpenAI managed-model availability is validated by the service rather than a hardcoded SDK list. MiniMax exposes `Speed`, `Volume`, `Pitch`, `Emotion`, `SampleRate`, `Bitrate`, `Format`, and `Channel`, merging typed fields with `AdditionalParams`; the Platform adapter supplies a default speed of 1. Gemini credentials stay in `credential.fields.api_key` on Platform requests, and HeyGen maps to `liveavatar`. Bedrock accepts either `APIKey` or the existing access-key pair using vendor `amazon`; the two credential forms cannot be combined.
