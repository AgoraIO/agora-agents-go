---
sidebar_position: 5
title: CN AgentKit
description: Use the CN AgentKit facade for Chinese mainland API routing.
---

# CN AgentKit

Use the CN AgentKit facade when your integration should route to Chinese mainland API endpoints (`option.AreaCN`).

`NewAgent(client, opts...)` requires a non-nil `*AgoraClient` from `NewAgoraClient`. Passing `nil` panics with `NewAgent requires AgoraClient`.

```go
package main

import (
    "context"
    "fmt"
    "time"

    agentkit "github.com/AgoraIO/agora-agents-go/v2/agentkit/cn"
    vendors "github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors"
)

func main() {
    client := agentkit.NewAgoraClient(agentkit.ClientOptions{
        AppID:          "your-app-id",
        AppCertificate: "your-app-certificate",
    })

    agent := agentkit.NewAgent(client).WithStt(
        vendors.NewTencentSTT(vendors.TencentSTTOptions{
            Key:    "secret-key",
            AppID:  "app-id",
            Secret: "secret",
        }),
    ).WithLlm(
        vendors.NewTencentLLM(vendors.TencentLLMOptions{
            APIKey:  "api-key",
            Model:   "hunyuan-turbos-latest",
            BaseURL: "https://api.hunyuan.cloud.tencent.com/v1/chat/completions",
        }),
    ).WithTts(
        vendors.NewTencentTTS(vendors.TencentTTSOptions{
            AppID:     "app-id",
            SecretID:  "secret-id",
            SecretKey: "secret-key",
        }),
    )

    session := agent.CreateSession(agentkit.CreateSessionOptions{
        Name:        fmt.Sprintf("conversation-%d", time.Now().UnixMilli()),
        Channel:     fmt.Sprintf("demo-channel-%d", time.Now().UnixMilli()),
        AgentUID:   "1001",
        RemoteUIDs: []string{"1002"},
    })

    _, _ = session.Start(context.Background())
}
```

## Routing Behavior

`agentkit/cn.NewAgoraClient(...)` always uses `option.AreaCN` internally. You do not need to pass an `Area` value.

When `WithStt()` is omitted on `agentkit/cn.Agent`, the request falls back to `asr.vendor = "fengming"` automatically.

Qwen Omni MLLM is available only from the CN vendor package and uses the mainland DashScope WebSocket endpoint by default:

```go
import Agora "github.com/AgoraIO/agora-agents-go/v2"

agent := agentkit.NewAgent(client).WithMllm(
    vendors.NewQwenOmni(vendors.QwenOmniOptions{
        APIKey: "your-dashscope-key",
        Model:  "qwen3-omni-flash-realtime",
        TurnDetection: &Agora.MllmTurnDetection{
            Mode: Agora.MllmTurnDetectionModeServerVad.Ptr(),
        },
    }),
)
```

## Package Layout

- Global/default package: `github.com/AgoraIO/agora-agents-go/v2/agentkit`
- Global/default vendors: `github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors`
- CN facade: `github.com/AgoraIO/agora-agents-go/v2/agentkit/cn`
- CN vendors: `github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors`

## Supported Surfaces

The CN facade currently supports:

- Cascading ASR / LLM / TTS flows and Qwen Omni MLLM flows
- `WithStt`, `WithLlm`, `WithTts`, `WithMllm`, and `WithAvatar` on `cn.Agent`
- CN-specific STT, LLM, TTS, MLLM, and avatar vendors from `agentkit/cn/vendors`
- Session-level options via `NewAgent(client, opts...)` such as `WithTurnDetectionConfig`, `WithInterruptionConfig`, `WithAdvancedFeatures`, `WithParameters`, `WithGeofence`, `WithRtc`, and `WithFillerWords`

The CN facade does **not** support:

- Chain methods such as `WithTurnDetection`, `WithSal`, or `WithLabels` on `cn.Agent` (use the matching `AgentOption` on `NewAgent` instead where available)
- `WithSalConfig`, `WithGreetingConfigs`, or `WithLabels` as `AgentOption` helpers
- `AgentPresets` or `ResolveSessionPresets*` helpers

For CN TTS vendors, `AdditionalParams map[string]interface{}` and `SkipPatterns []int` are available consistently across all TTS option structs.

### Global vs CN Agent API

| Capability | Global `agentkit` | `agentkit/cn` |
|---|---|---|
| Client options | `AgoraClientOptions` with `Area` | `ClientOptions` (fixed `AreaCN`) |
| Vendor chaining | `WithStt`, `WithLlm`, `WithTts`, `WithMllm`, `WithAvatar` | `WithStt`, `WithLlm`, `WithTts`, `WithMllm`, `WithAvatar` |
| MLLM vendors | OpenAI Realtime, Azure OpenAI Realtime, Gemini Live, Vertex AI, xAI Grok | Qwen Omni |
| Avatar vendors | LiveAvatar, Generic, Anam, Akool, HeyGen | SenseTime only |
| Turn detection | `WithTurnDetectionConfig` or `WithTurnDetection` | `WithTurnDetectionConfig` on `NewAgent` only |

See [Vendors](./concepts/vendors.md#cn-vendors-agentkitcnvendors) for the CN constructor catalog.
