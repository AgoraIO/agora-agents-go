---
sidebar_position: 2
title: MLLM Flow
description: Use multimodal models (OpenAI Realtime, Gemini Live) for real-time audio processing.
---

# MLLM Flow (Multimodal)

The MLLM flow uses a single multimodal model to process audio input and generate audio output directly, bypassing the ASR -> LLM -> TTS chain. This provides lower latency and more natural conversational behavior.

## Enabling MLLM Mode

Call `WithMllm(vendor)` to enable MLLM mode. The builder sets `mllm.enable = true` automatically. Use `agentkit` plus `agentkit/vendors` for global providers, and `agentkit/cn` plus `agentkit/cn/vendors` for mainland China Qwen Omni.

```go
agent := agentkit.NewAgent(client).WithMllm(vendor)
```

## OpenAI Realtime Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    Agora "github.com/AgoraIO/agora-agents-go/v2"
    "github.com/AgoraIO/agora-agents-go/v2/agentkit"
    "github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
    "github.com/AgoraIO/agora-agents-go/v2/option"
)

func main() {
    client := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
        Area:           option.AreaUS,
        AppID:          "<app_id>",
        AppCertificate: "<app_cert>",
    })

    agent := agentkit.NewAgent(client).WithMllm(
        vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
            APIKey: "<openai_key>",
            Model:  "gpt-4o-realtime-preview",
            Params: map[string]interface{}{
                "voice": "alloy",
            },
        }),
    )

    session := agent.CreateSession(agentkit.CreateSessionOptions{
        Name:        fmt.Sprintf("conversation-%d", time.Now().UnixMilli()),
        Channel:     fmt.Sprintf("demo-channel-%d", time.Now().UnixMilli()),
        AgentUID:   "1001",
        RemoteUIDs: []string{"1002"},
    })

    ctx := context.Background()

    agentID, err := session.Start(ctx)
    if err != nil {
        log.Fatalf("Failed to start: %v", err)
    }
    fmt.Println("Realtime agent running:", agentID)

    err = session.Stop(ctx)
    if err != nil {
        log.Fatalf("Failed to stop: %v", err)
    }
}
```

## Gemini Live Example

```go
agent := agentkit.NewAgent(client).WithMllm(
    vendors.NewGeminiLive(vendors.GeminiLiveOptions{
        APIKey:       "<google_ai_api_key>",
        Model:        "gemini-live-2.5-flash",
        Instructions: "You are a helpful assistant.",
        Voice:        "Puck",
    }),
)
```

## Azure OpenAI Realtime (Global)

Azure OpenAI Realtime is exposed only by the global vendor package. `URL` is the complete Azure Realtime WebSocket URL for your resource and deployment. `MaxHistory` maps to the Azure-only top-level `mllm.max_history` field.

```go
maxHistory := 20
messages := []map[string]interface{}{
    {
        "role":    "user",
        "content": "Can you please count to 3 after greeting?",
    },
}
agent := agentkit.NewAgent(client).WithMllm(
    vendors.NewAzureOpenAIRealtime(vendors.AzureOpenAIRealtimeOptions{
        APIKey:           "<azure_openai_key>",
        URL:              "wss://<resource>.openai.azure.com/openai/realtime?api-version=<version>&deployment=<deployment>",
        Model:            "gpt-realtime-2",
        Voice:            "alloy",
        Instructions:     "You are a Conversational AI Agent, developed by Agora.",
        Messages:         messages,
        OutputModalities: []string{"audio"},
        MaxHistory:       &maxHistory,
        GreetingMessage:  "Hey There Sir",
        TurnDetection: &Agora.MllmTurnDetection{
            Mode: Agora.MllmTurnDetectionModeServerVad.Ptr(),
        },
    }),
)
```

## Qwen Omni (Mainland China)

Qwen Omni is exposed only by `agentkit/cn/vendors`. Its default URL is the mainland China DashScope endpoint.

```go
import (
    Agora "github.com/AgoraIO/agora-agents-go/v2"
    agentkit "github.com/AgoraIO/agora-agents-go/v2/agentkit/cn"
    vendors "github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors"
)

agent := agentkit.NewAgent(client).WithMllm(
    vendors.NewQwenOmni(vendors.QwenOmniOptions{
        APIKey: "<dashscope_key>",
        Model:  "qwen3-omni-flash-realtime",
        Voice:  "Momo",
        TurnDetection: &Agora.MllmTurnDetection{
            Mode: Agora.MllmTurnDetectionModeServerVad.Ptr(),
        },
    }),
)
```

## MLLM with Turn Detection

Configure MLLM turn detection on the MLLM vendor with `TurnDetection`. It is required by Azure OpenAI Realtime and Qwen Omni. When set, `mllm.turn_detection` overrides the top-level `turn_detection` object.

Example:

```go
agent := agentkit.NewAgent(client).WithMllm(
    vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
        APIKey: "<openai_key>",
        TurnDetection: &Agora.MllmTurnDetection{
            Mode: Agora.MllmTurnDetectionModeServerVad.Ptr(),
            ServerVadConfig: &Agora.MllmTurnDetectionServerVadConfig{
                IdleTimeoutMs: Agora.Int(5000),
            },
        },
    }),
)
```

## Low-Level Usage

Use AgentKit for MLLM session starts. It keeps the app-credentials flow intact and generates the required REST and RTC tokens from `AppID` and `AppCertificate`. See [Low-Level API](./low-level-api.md) for generated REST operations outside the realtime session lifecycle.

## Pointer Helper Functions

MLLM configuration makes heavy use of pointer helpers for optional fields:

| Helper | Type | Example |
|---|---|---|
| `Agora.Bool(true)` | `*bool` | `Enable: Agora.Bool(true)` |
| `Agora.String("...")` | `*string` | `APIKey: Agora.String("<key>")` |
| `Agora.Int(120)` | `*int` | `IdleTimeout: Agora.Int(120)` |
| `Agora.Float64(0.5)` | `*float64` | `Threshold: Agora.Float64(0.5)` |

These exist because Go does not allow taking the address of a literal value (`&true` is invalid). The helpers return pointers to the given values.

## Key Differences from Cascading Flow

| Aspect | Cascading | MLLM |
|---|---|---|
| Vendors required | LLM + TTS (STT optional) | MLLM only |
| Audio processing | Three-step chain | Single model, end-to-end |
| Latency | Higher (3 network hops) | Lower (1 network hop) |
| `mllm.enable` | Not set or `false` | Must be `Agora.Bool(true)` |
