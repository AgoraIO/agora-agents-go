---
sidebar_position: 2
title: Authentication
description: Configure the Go SDK with the app-credentials flow and understand the supported auth modes.
---

# Authentication

Use app credentials mode for production session integrations.

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

## Client tokens

Browsers and apps need their own token to join the channel and log in to RTM, which carries agent transcripts and state. Mint it on your server with `agentkit.GenerateClientToken`:

```go
token, err := agentkit.GenerateClientToken(agentkit.GenerateClientTokenOptions{
    AppID:          os.Getenv("AGORA_APP_ID"),
    AppCertificate: os.Getenv("AGORA_APP_CERTIFICATE"),
    ChannelName:    "my-room",
    UID:            "100", // or a string user account; also the RTM user ID
    TokenExpire:    3600,
})
```

The token grants RTC (publisher) and RTM login only. Never hand a client a token from `GenerateConvoAIToken`: its ConvoAI service authenticates REST calls, so a client holding it could control agents in your project. Keep ConvoAI tokens on the server.
