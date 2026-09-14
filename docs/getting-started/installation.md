---
sidebar_position: 1
title: Installation
description: Install the Agora Conversational AI Go SDK and configure your project.
---

# Installation

## Prerequisites

- **Go 1.21** or later

## Install

```sh
go get github.com/AgoraIO/agora-agents-go/v2@latest
```

## Import Paths

The SDK is organized into several packages. Import the ones you need:

```go
// Root package — type definitions, pointer helpers (Agora.String(), Agora.Bool(), etc.), environments
Agora "github.com/AgoraIO/agora-agents-go/v2"

// Fern-generated client — low-level API access
"github.com/AgoraIO/agora-agents-go/v2/client"

// Request options — authentication, base URL, retries, HTTP client
"github.com/AgoraIO/agora-agents-go/v2/option"

// Wrapper layer — Agent builder with functional options, session lifecycle
"github.com/AgoraIO/agora-agents-go/v2/agentkit"

// Vendor constructors — LLM, TTS, STT, MLLM, and Avatar vendors
"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
```

## Verify Installation

```go
package main

import (
    "fmt"

    Agora "github.com/AgoraIO/agora-agents-go/v2"
)

func main() {
    fmt.Println("Default API endpoint:", Agora.Environments.Default)
}
```

```sh
go run main.go
# Output: Default API endpoint: https://api.agora.io/api/conversational-ai-agent
```

## Next Steps

- [Authentication](./authentication.md) — configure your credentials
- [Quick Start](./quick-start.md) — build your first conversational agent

## Migrating from a previous module path

In v2.0.0, the repository remains [AgoraIO/agora-agents-go](https://github.com/AgoraIO/agora-agents-go), and the Go module path is `github.com/AgoraIO/agora-agents-go/v2`. Replace `github.com/AgoraIO-Conversational-AI/agent-server-sdk-go` in `go.mod` and in all import paths, then run `go mod tidy`.

For the full change list, see [changelog — v2.0.0](../../changelog.md#v200--2026-05-21).
