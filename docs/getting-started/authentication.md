---
sidebar_position: 2
title: Authentication
description: Configure the Go SDK with the app-credentials flow and understand the supported auth modes.
---

# Authentication

Use app credentials mode for production session integrations.

Generated REST tokens include the ConvoAI product service plus RTC and RTM services. Platform session routes require the ConvoAI service; an RTC/RTM-only token is insufficient. `GenerateConvoAIToken` uses `TokenExpire` for the whole token and RTM login; `PrivilegeExpire` controls only the RTC privileges and defaults to `TokenExpire`. The whole-token expiration still limits any longer privilege lifetime.

Create `AgoraClient` with `AppID` and `AppCertificate`, then let `AgentSession` generate the ConvoAI REST auth token and the RTC join token automatically.

## Recommended: app credentials

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

    _ = client
}
```

## Why this is the default

- The SDK handles ConvoAI REST auth and RTC join token generation for you.
- Your session code stays focused on agent behavior while the SDK handles auth details.
- Your quick start code stays vendor-key free when you use supported Agora-managed global models.

## Legacy auth modes

The generated REST client still contains legacy auth hooks for prebuilt REST tokens and HTTP Basic Auth. Do not use those for new session integrations. Use app credentials so AgentKit can mint short-lived ConvoAI REST auth and RTC join tokens for each session.

## Inspecting the resolved auth mode

```go
import "fmt"

fmt.Println(client.AuthMode) // "app-credentials"
```
