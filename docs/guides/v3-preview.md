---
sidebar_position: 10
title: Conversational AI v3 preview
description: Deploy stored agents and manage v3 secrets, numbers, and read-only sessions with Go.
---

# Conversational AI v3 preview (P1)

In v3, an **agent** is a saved configuration identified by `agent_id`. A running instance is a **session**, identified by `agent_session_id`. Existing `Agent.CreateSession(...).Start(ctx)` still starts a v2 `/join` session; `Deploy` saves a v3 configuration and does not start a session. The Go import path remains `github.com/AgoraIO/agora-agents-go/v2`: its `/v2` suffix is the module's semantic major version, independent of the REST API version.

## Quickstart on staging

Set the project App ID, a REST token, the staging gateway origin, and the two gateway headers. The gateway secret is sent separately by the service operator. Pass an **origin**, not a full v3 path: the SDK appends `/conversational-ai/v3/projects/{appid}`.

```sh
export AGORA_APP_ID='your-project-app-id'
export AGORA_REST_TOKEN='your-rest-token'
export AGORA_V3_STAGING_ORIGIN='https://your-staging-gateway.example'
export AGORA_V3_GATEWAY_SECRET='your-gateway-secret'
export AGORA_V3_COMPANY_ID='your-numeric-company-id'
```

Save as `main.go` and run `go run main.go`. This makes a read-only list request:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
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
        V3: &agentkit.V3Options{
            BaseURL: os.Getenv("AGORA_V3_STAGING_ORIGIN"),
            Headers: http.Header{
                "X-Studio-V2-Gateway-Secret": {os.Getenv("AGORA_V3_GATEWAY_SECRET")},
                "X-Consumer-Username": {os.Getenv("AGORA_V3_COMPANY_ID")},
            },
        },
    })
    page, err := c.V3.Agents.List(context.Background(), v3.ListOptions{Limit: 10})
    if err != nil { log.Fatal(err) }
    fmt.Printf("count=%d next=%v request=%s\n", page.Body.Data.Count,
        page.Body.NextCursor, page.RequestID)
}
```

An app certificate client can set `AppCertificate` in place of `Token`; AgentKit signs a fresh ConvoAI token for each v3 request. Basic auth uses `CustomerID` and `CustomerSecret`. `AGORA_AGENTS_V3_BASE_URL` is a second origin override; explicit `V3Options.BaseURL` takes precedence. Put staging-only headers in `V3Options.Headers` so v2 requests never receive them. Production path and regional availability are unconfirmed; AgentKit rejects CN deployment pending backend confirmation.

## AgentKit workflows

### Deploy and redeploy

The following body assumes `ctx context.Context`, `c *agentkit.AgoraClient`, and imports for `agentkit` and `agentkit/vendors`:

```go
agent := agentkit.NewAgent(c,
    agentkit.WithVariable("resolved", agentkit.VariableDefinition{
        Type: "bool", Extract: true, Description: "Was the issue resolved?",
    }),
).WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
    APIKey: agentkit.SecretRef("prod", "openai"),
    Instructions: "Help the caller briefly and politely.",
}))

// Create once; persist the returned ID.
deployed, err := agent.Deploy(ctx, agentkit.DeployOptions{Name: "Support"})
if err != nil { return err }

// Later, update the same stored configuration.
deployed, err = agent.Deploy(ctx, agentkit.DeployOptions{
    Name: "Support", AgentID: deployed.ID,
})
if err != nil { return err }
```

`WithVariable` also exists as `agent.WithVariable(name, def)`, returning a new immutable builder. Names match `^[a-z][a-z0-9_]{0,63}$`; types are `str`, `int`, `float`, `bool`, `date`, and `datetime`. `Options`, when provided, must contain 1–100 unique JSON scalar values. Descriptions are limited to 2000 Unicode characters; `Extract: true` requires a nonempty description. The no-error builder methods panic for invalid definitions. `agentkit.GetDeployedAgent(ctx, c, id)` loads a handle; `deployed.Delete(ctx)` deletes it. Handles expose `ID`, `Name`, `ETag`, and `Raw` (`*v3.AgentsClient`).

With no `AgentID`, `Deploy` POSTs a new agent. Save its ID: names are not unique upsert keys and agent POST has no supported idempotency key. With an ID, `Deploy` GETs the current ETag, then PATCHes. It sends recursive `null`s to clear readable stale fields, omitting server-owned IDs, timestamps, and resolved credential metadata. A 412 becomes `ErrPreconditionFailed` and retains the underlying v3 error; deploy does not retry the conflict automatically.

For cascaded deployments, an explicit `TurnDetection.Language` is preserved as `pipeline.asr.language`, taking precedence over the ASR vendor language as on v2. Without an explicit STT vendor, it uses the existing global `ares` default with managed credentials; endpoint availability remains a backend concern. The same interaction-language setting is rejected on realtime deployment because the published realtime schema has no equivalent.

Legacy greeting settings `mode: "single_every"` and `mode: "single_first"` map to `on: "each_join"` and `on: "first_join"`. An explicit `on` takes precedence. Unknown legacy modes and playback settings without an actual greeting are rejected before a request. Audio download timeout and PCM sample-rate settings require an audio greeting; supplying them for a text greeting is rejected. Empty message content remains valid. Variable descriptions allow up to 2000 Unicode characters, rather than 2000 UTF-8 bytes.

### Secrets and numbers

Create the secret set before deploying a provider that uses `SecretRef`. Values are write-only. The optional secret `Idempotency-Key` is a `string` argument; pass `""` to omit it. This code also assumes `os`, `v3`, and `agentkit` imports:

```go
secret, err := c.V3.Secrets.Create(ctx, v3.SecretCreate{
    Name: "prod",
    Secrets: map[string]string{"openai": os.Getenv("OPENAI_API_KEY")},
}, "secret-create-2026-10")
if err != nil { return err }
_ = secret.Body.Keys // Values are never returned.

number, err := c.V3.Numbers.Create(ctx, v3.NumberCreate{
    Number: "+12025550123",
    SipTrunk: v3.SipTrunk{Hostname: "sip.example.com", Transport: "tls"},
}, "")
if err != nil { return err }

maxCall := 900
_, err = deployed.AssignNumber(ctx, number.Body, agentkit.AssignNumberOptions{
    MaxCallDurationSeconds: &maxCall,
})
if err != nil { return err }
_, err = deployed.UnassignNumber(ctx, number.Body.NumberID)
if err != nil { return err }
```

Assignment accepts `number_id`, `v3.Number`, `*v3.Number`, or a formatted E.164 value. E.164 lookup scans this project's pages because the API has no number filter, then GETs the matched number for an ETag. Assignment PATCHes only the inbound agent and explicitly supplied call policy fields; one 412 causes one fresh GET and retry. A missing project number returns `ErrNumberNotFound`. Number POST accepts an idempotency argument but staging does not reliably honor it, so it is not automatically retried. `SecretRef("prod", "openai")` yields `$secrets.prod.openai`; passing it to the v2 session path returns `ErrSecretRefNotSupported`.

`deployed.ListNumbers(ctx)` returns the v3 numbers whose `inbound.agent` equals the deployed agent ID. It scans **every number page in the project**, with no per-number GETs, because staging has no agent filter. Its request cost grows with the project's total number count; use it deliberately for large projects. The v3 session list is also project-wide, but the published list and get responses provide no `agent_id` association. There is no `DeployedAgent.ListSessions` helper because matching the nonunique `agent_name` would be unreliable.

## Low-level reference: all 20 staging operations

Import `github.com/AgoraIO/agora-agents-go/v2/v3`. `agentkit.AgoraClient.V3` already has the App ID. The root `client.NewClient(...)` has no App ID option: use `api.V3.ForAppID(appID)` before calling its resources, or construct `v3.NewClient(appID, opts...)` directly. An empty App ID or resource ID fails locally before HTTP.

Every method takes `ctx context.Context` first and returns `(*v3.Response[T], error)`. Below are its **exact remaining arguments**, response body `T`, and relevant headers. A `Response[T]` contains `Body`, a detached `Header http.Header` with all response headers, raw `ETag`, `RequestID` (from `Agora-Request-Id`), and `StatusCode`. Use `response.Header.Get("Location")` or `response.Header.Get("Idempotent-Replayed")` to inspect those headers. `ifMatch` must be the literal ETag, including quotes if the server sent them. `""` omits optional header arguments.

| Method after `c.V3` | Arguments after `ctx` | Body `T` | Header/filter note |
|---|---|---|---|
| `Agents.Create` | `body v3.AgentCreate` | `v3.Agent` | No idempotency key |
| `Agents.List` | `opts v3.ListOptions` | `v3.Page[v3.AgentListItem]` | limit, cursor, order |
| `Agents.Get` | `id string` | `v3.Agent` | `agent_id` |
| `Agents.Update` | `id, ifMatch string, patch v3.AgentPatch` | `v3.Agent` | Required `If-Match` |
| `Agents.Delete` | `id string` | `v3.Deleted` | `agent_id` |
| `Secrets.Create` | `body v3.SecretCreate, idempotencyKey string` | `v3.SecretSet` | Optional `Idempotency-Key` |
| `Secrets.List` | `opts v3.ListOptions` | `v3.Page[v3.SecretSet]` | limit, cursor, order |
| `Secrets.Get` | `name string` | `v3.SecretSet` | Secret set name |
| `Secrets.Update` | `name, ifMatch string, body v3.SecretUpdate` | `v3.SecretSet` | Optional `If-Match`; nil values delete keys |
| `Secrets.Replace` | `name, ifMatch string, body v3.SecretReplace` | `v3.SecretSet` | Required `If-Match` |
| `Secrets.Delete` | `name, ifMatch string` | `v3.Deleted` | Required `If-Match` |
| `Secrets.SetKey` | `name, key string, body v3.SecretKeyWrite, opts v3.SecretKeyOptions` | `v3.SecretKey` | Optional `If-Match` and `If-None-Match` |
| `Secrets.DeleteKey` | `name, key, ifMatch string` | `v3.Deleted` | Optional `If-Match` |
| `Numbers.Create` | `body v3.NumberCreate, idempotencyKey string` | `v3.Number` | Key optional; unreliable on staging |
| `Numbers.List` | `opts v3.ListOptions` | `v3.Page[v3.NumberListItem]` | limit, cursor, order |
| `Numbers.Get` | `id string` | `v3.Number` | `number_id` |
| `Numbers.Update` | `id, ifMatch string, patch v3.NumberPatch` | `v3.Number` | Required `If-Match` |
| `Numbers.Delete` | `id string` | `v3.Deleted` | `number_id` |
| `Sessions.List` | `opts v3.SessionListOptions` | `v3.Page[v3.Session]` | channel, status, started_after/before, limit, cursor |
| `Sessions.Get` | `id string` | `v3.Session` | `agent_session_id` |

`v3.ListOptions` has `Limit int`, `Cursor string`, and `Order string`. `v3.SessionListOptions` embeds it and adds `Channel`, `Status`, `StartedAfter *int64`, and `StartedBefore *int64`; session list does not send `Order`. Each page has `Data.Count`, `Data.List`, `HasMore`, and `NextCursor`. `Count` is the page count, not a promised project total. A cursor loop:

The staging defaults and limits differ by resource: agents and secrets default to 50 items and allow 1–200; numbers default to 50 and allow 1–100; sessions default to 20 and require at least 1. Agents, secrets, and numbers default to `create_ts.desc` order. Leaving `SessionListOptions.Status` empty sends no filter, so the server defaults to **`RUNNING`**, not all sessions. To cover all published statuses, list separately with `IDLE`, `STARTING`, `RUNNING`, `STOPPING`, `STOPPED`, and `FAILED`, following each status's cursor independently.

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

## Compatibility and limits

`Deploy` collects unsupported v2 fields into one `ErrUnsupportedField` before HTTP. Current examples are `sal`, `advancedFeatures`, `geofence`, `labels`, `rtc`, `audioScenario`, `pipelineId`, non-silence `parameters`, nonempty `template_variables` (v3 variables have no defaults), unverified provider fields, and multipart credentials that cannot map to one API key. Remove them from the stored-agent builder or keep that use case on v2. Managed credentials use only a URL already present in a vendor config; the SDK invents no default URL.

Provider API keys and secret values are write-only. Supply credentials explicitly when redeploying; GET cannot reconstruct them. Redeploy writes `null` for readable stale nested fields, but actual backend handling of whole-pipeline merge patches still needs live verification. Session operations are **read-only and project-wide** in P1: no v3 start/stop/actions, outbound call methods, analysis shortcut on the v2 `AgentSession`, or campaigns exist yet. v2 telephony and phone-number clients keep their original semantics; they do not manage v3 project-scoped numbers.

## Maintenance and verification boundary

The v3 package is **handwritten against the staging OpenAPI snapshot** and protected by `.fernignore`; Fern has not generated it. Keep this guide and the types in sync as the schema evolves, and revisit generation for a production schema. Local tests cover paths, headers, pagination, conditional errors, AgentKit mapping, and number assignment. This preview has **not** been validated against a live staging or production gateway in this workspace. The production path, region/CN availability, managed URL requirement, and nested merge-patch behavior remain open backend questions.
