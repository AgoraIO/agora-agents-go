---
sidebar_position: 2
title: Authentication
description: Configure the Go SDK with the app-credentials flow and understand the supported auth modes.
---

# Authentication

Use app credentials mode for production session integrations.

Generated REST tokens include the ConvoAI product service plus RTC and RTM services. When using token authentication, Platform session routes require the ConvoAI service; an RTC/RTM-only token is insufficient. `GenerateConvoAIToken` uses `TokenExpire` for the whole token and RTM login; `PrivilegeExpire` controls only the RTC privileges and defaults to `TokenExpire`. The whole-token expiration still limits any longer privilege lifetime.

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

## Prebuilt token or Basic authentication

Both Token007 and HTTP Basic authentication are supported for REST requests. Set `Token` to a raw token containing the ConvoAI service, or supply the complete `CustomerID` / `CustomerSecret` pair from Agora Console:

```go
client := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
    Area: option.AreaUS,
    AppID: "your-app-id",
    CustomerID: "your-customer-id",
    CustomerSecret: "your-customer-secret",
    // Optional: AppCertificate generates RTC transport tokens independently.
})
```

Basic credentials and `Token` are mutually exclusive. A customer ID cannot contain a colon. `AppCertificate` can accompany either mode to generate RTC tokens; it does not replace REST Basic authentication. Without a certificate, provide an RTC token when required by the session transport.

Low-level clients support `option.WithBasicAuth(customerID, customerSecret)`, `option.WithToken(token)`, and a complete `option.WithAuthorization("Basic ...")` header. Internal Studio routing headers remain unsupported.

## Inspecting the resolved auth mode

```go
import "fmt"

fmt.Println(client.AuthMode) // "app-credentials"
```
