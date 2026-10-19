---
sidebar_position: 9
title: Low-Level API
description: Use generated clients for advanced operations while keeping app-credentials auth as the default.
---

# Low-Level API

For stored Conversational AI v3 configurations, secrets, project-scoped numbers, and read-only sessions, see the [v3 preview guide](./v3-preview.md). That client is handwritten against the staging schema; the clients below are generated for v2.

For starting and managing realtime agent sessions, prefer the AgentKit builder:

- create `agentkit.NewAgoraClient` with `AppID` and `AppCertificate`
- pass that client to `agentkit.NewAgent(client, ...)`
- configure vendors with `WithStt`, `WithLlm`, `WithTts`, or `WithMllm`
- call `session.Start(ctx)`, `session.Say(ctx, ...)`, `session.Update(ctx, ...)`, and `session.Stop(ctx)`

That path generates ConvoAI REST auth and RTC join tokens from app credentials, so application code does not need prebuilt REST tokens, RTC tokens, Customer ID, or Customer Secret.

## Generated Clients

The SDK also exposes generated clients for API surfaces that are not part of the AgentKit session lifecycle:

- `client.Telephony` for call status and hangup operations
- `client.PhoneNumbers` for phone-number list, create, retrieve, update, and delete operations
- `client.AgentManagement` for management actions such as `agent-think`

Use these when you need REST API operations outside the AgentKit session lifecycle. For new session starts, use AgentKit instead of manually constructing `StartAgentsRequest`; AgentKit handles token generation and supported Agora-managed global models.

## App-Credentials Client

```go
package main

import (
    "github.com/AgoraIO/agora-agents-go/v2/agentkit"
    "github.com/AgoraIO/agora-agents-go/v2/option"
)

func main() {
    client := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
        Area:           option.AreaUS,
        AppID:          "your-app-id",
        AppCertificate: "your-app-certificate",
    })

    _ = client.Telephony
    _ = client.PhoneNumbers
    _ = client.AgentManagement
}
```

## Stopping By Agent ID

If you need to stop an agent from a separate request handler and do not have the original `AgentSession`, use `StopAgent`. It uses the same app credentials to mint the required ConvoAI REST auth header.

```go
if err := client.StopAgent(ctx, agentID); err != nil {
    return err
}
```

## Raw Types

Generated request and response types are still useful for strongly typed values, pointer helpers, and advanced payload construction. The root package exposes helpers such as `Agora.String`, `Agora.Bool`, `Agora.Int`, and `Agora.Float64` for optional fields.

```go
import Agora "github.com/AgoraIO/agora-agents-go/v2"

idleTimeout := Agora.Int(120)
greeting := Agora.String("Hello! How can I help?")
```

For full generated method signatures, see [reference.md](../../reference.md).
