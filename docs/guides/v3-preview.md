# Conversational AI v3 preview (P1)

The Go module path stays `github.com/AgoraIO/agora-agents-go/v2`. Its `/v2` suffix is the Go module major version, not the Conversational AI REST API version. Existing `Agent.CreateSession` and the v2 resource clients are unchanged.

The preview adds `client.V3.Agents`, `.Secrets`, `.Numbers`, and read-only `.Sessions`. The resources cover all 20 operations in the staging OpenAPI snapshot. Each method returns a `v3.Response[T]` with `Body`, raw `ETag`, `RequestID`, and `StatusCode`. List bodies retain `Data.Count`, `Data.List`, `HasMore`, and `NextCursor`; pass `NextCursor` into the next `ListOptions.Cursor`. IDs are `agent_id`, `number_id`, and `agent_session_id`. Empty IDs are rejected before HTTP.

```go
client := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
    Area: option.AreaUS,
    AppID: appID,
    Token: token,
    V3: &agentkit.V3Options{
        BaseURL: "https://your-staging-gateway.example",
        Headers: http.Header{"X-Studio-V2-Gateway-Secret": {gatewaySecret}},
    },
})

agent := agentkit.NewAgent(client,
    agentkit.WithVariable("resolved", agentkit.VariableDefinition{
        Type: "bool", Extract: true, Description: "Was the issue resolved?",
    }),
)
deployed, err := agent.Deploy(ctx, agentkit.DeployOptions{Name: "Support", AgentID: existingID})
if err != nil { return err }
_, err = deployed.AssignNumber(ctx, "+1 (202) 555-0123", agentkit.AssignNumberOptions{})
```

`AgentID` empty creates a new stored agent; persist the returned `deployed.ID`. With an ID, `Deploy` GETs the latest ETag and PATCHes the stored config. It returns `ErrUnsupportedField` before HTTP if the builder uses a v2-only field or a vendor setting with no verified v3 mapping. `WithVariable` works both as an option and as an immutable method. `SecretRef("prod", "openai")` returns `$secrets.prod.openai` for a provider API key; the v2 session path rejects such references with `ErrSecretRefNotSupported`. `DeployedAgent.Raw` exposes the agents resource for direct calls. `GetDeployedAgent` loads an existing handle; `Delete` removes it.

Numbers must first be created in the same project through `client.V3.Numbers.Create`. `AssignNumber` and `UnassignNumber` accept a `number_id`, a number response, or an E.164 value. For an E.164 value, AgentKit scans the project's number pages, then GETs the matched number for its ETag. A 412 causes one fresh GET and retry. Assignment sends only the inbound agent and explicitly requested call policy fields. The low-level `v3.PreconditionFailedError.Latest(ctx)` fetches the current resource and ETag on demand.

The v3 base URL is an **origin**, not a full path. The SDK appends `/conversational-ai/v3/projects/{appid}`. `AGORA_AGENTS_V3_BASE_URL` is another staging override; `V3Options.BaseURL` takes precedence. The production path and regional/CN availability are awaiting backend confirmation. AgentKit rejects `Deploy` in CN until confirmed. Staging gateway headers belong in `V3Options.Headers` so they are not sent to v2. An app certificate client signs a fresh ConvoAI token for v3 requests; Basic and prebuilt token auth use the existing SDK options.

PATCH uses `application/json`, as specified by staging, with JSON merge-patch semantics. Redeploy recursively writes `null` for readable fields absent from the builder, including stale provider params. It does not echo server-owned IDs/timestamps or resolved credential metadata. A switch from BYOK to managed sends `api_key: null`. Provider API keys and secret values are write-only, so the SDK cannot reconstruct them from GET; supply credentials explicitly on redeploy. Managed credentials use only URLs already present in a vendor config. The staging requirement for a managed provider URL remains unresolved; the SDK does not invent defaults. POST agents and numbers use one HTTP attempt because idempotency is not reliable on staging; all v3 writes use one transport attempt.

Sessions only support list/get in P1. Starting or stopping sessions, outbound calls, session actions and campaigns await published schemas and have no AgentKit methods in this preview.
