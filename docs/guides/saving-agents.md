---
sidebar_position: 9
title: Configure, run, and save agents
description: Use one immutable Agent config with existing sessions or save it for later updates and number assignment.
---

# Configure, run, and save agents

Build an `Agent` configuration once, then choose whether to start an existing session or save the configuration for reuse. The Go module path remains `/v2` for its semantic major version. Existing `CreateSession(...).Start(ctx)` keeps its v2 `/join` behavior. On the `feat/v3-api-preview` branch, `Agent.Save` stores a v3 configuration, and `SavedAgent.CreateSession` starts a session from its saved ID. Configurations that use secret references or variables start through the v3 ephemeral endpoint automatically; plain inline configurations keep the existing v2 `/join` behavior.

## Run a session with inline credentials

For an existing v2 session, configure the provider with its credential and start a session as usual. The snippets assume `ctx context.Context` and imports for `agentkit`, `agentkit/vendors`, and `option`.

```go
client := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
    Area: option.AreaUS,
    AppID: "your-app-id",
    AppCertificate: "your-app-certificate",
})
config := agentkit.NewAgent(client).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
    APIKey: "your-openai-key",
    BaseURL: "https://api.openai.com/v1/chat/completions",
    Model: "gpt-4o-mini",
    SystemMessages: []map[string]interface{}{{"role": "system", "content": "You are a helpful voice assistant."}},
}))
session := config.CreateSession(agentkit.CreateSessionOptions{
    Name: "support-call",
    Channel: "support-call-room",
})
agentSessionID, err := session.Start(ctx)
if err != nil { return err }
```

The agent remains an immutable builder: chaining `With*` methods returns a new configuration. The original `config` is unchanged if you derive a variant for another session or for saving.

## Save and update a configuration

Use `Save` when you need a stored configuration and a stable `agent_id`. Keep the returned `SavedAgent` handle: its `Update` method reuses the saved ID and name and returns a handle with the latest ETag.

```go
saved, err := config.Save(ctx, agentkit.SaveOptions{Name: "Support"})
if err != nil { return err }

updatedConfig := config.WithVariable("priority", agentkit.VariableDefinition{
    Type: "str", Description: "What should the agent prioritize?",
})
saved, err = saved.Update(ctx, updatedConfig)
if err != nil { return err }
```

`Save` without `AgentID` creates a stored agent. `saved.Update(ctx, config)` updates it conditionally using the current ETag; a stale ETag returns `agentkit.ErrPreconditionFailed`. `agentkit.GetSavedAgent(ctx, client, agentID)` reloads a handle. Start a session from the saved handle with `saved.CreateSession(opts)`, then use the same `Start`, `GetInfo`, and `Stop` methods as an inline session. `GetInfo` reports the v3 `agent_session_id` separately from the v2 `agent_id`.

For this stored-agent preview, `AgoraClientOptions.V3.BaseURL` optionally overrides the default `https://preview.ai.agora.io` origin. Token007 and Basic authentication are supported; internal Studio headers are rejected. Use the `feat/v3-api-preview` branch build; the normal `@latest` installation command does not select it. See the [advanced v3 guide](./v3-preview.md) for transport, pagination, response headers, and API limits.

## Store credentials as secret references

These examples also assume `client *agentkit.AgoraClient`, `openAIKey string`, and the `agentkit/vendors` import. Create a project secret set with the version-neutral `client.Secrets` accessor, then use `SecretRef` in the provider config. Secret values are write-only: keep them in your secret manager and supply them on create or update.

```go
_, err := client.Secrets.Create(ctx, agentkit.SecretCreate{
    Name: "production",
    Secrets: map[string]string{"openai": openAIKey},
}, "")
if err != nil { return err }

storedConfig := agentkit.NewAgent(client).WithMllm(vendors.NewOpenAIRealtime(
    vendors.OpenAIRealtimeOptions{
        APIKey: agentkit.SecretRef("production", "openai"),
        Instructions: "Help the caller politely.",
    },
))
saved, err := storedConfig.Save(ctx, agentkit.SaveOptions{Name: "Realtime support"})
if err != nil { return err }
```

Create the secret set before saving the agent that references it. `SecretRef("production", "openai")` produces `$secrets.production.openai`. When an inline config contains a secret reference or variable, `CreateSession` selects the v3 ephemeral session endpoint automatically. The v2 session path still requires inline credentials and returns `ErrSecretRefNotSupported` for a secret reference.

## Create and assign a number

Project numbers use the version-neutral `client.Numbers` accessor. After creating a number, assign it through the saved-agent handle:

```go
number, err := client.Numbers.Create(ctx, agentkit.NumberCreate{
    Number: "+12025550123",
    SipTrunk: agentkit.SipTrunk{Hostname: "sip.example.com", Transport: "tls"},
}, "")
if err != nil { return err }

maxCall := 900
_, err = saved.AssignNumber(ctx, number.Body, agentkit.AssignNumberOptions{
    MaxCallDurationSeconds: &maxCall,
})
if err != nil { return err }
```

`AssignNumber` also accepts a number ID or E.164 value. E.164 lookup scans the project's number pages because the API has no number filter, so it becomes more expensive as the project grows. `saved.ListNumbers(ctx)` has the same project-wide scan cost. See the [v3 guide](./v3-preview.md#secrets-and-numbers) for conditional updates, retries, and pagination limits.
