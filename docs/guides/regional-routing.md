---
sidebar_position: 4
title: Regional Routing
description: Configure the Go SDK to route requests to the nearest Agora region.
---

# Regional Routing

Use `Area` on `agentkit.NewAgoraClient` to route session requests to the desired Agora region while keeping the app-credentials auth path.

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

## Area enum

| Constant | Region |
|---|---|
| `option.AreaUS` | United States (west + east) |
| `option.AreaEU` | Europe (west + central) |
| `option.AreaAP` | Asia-Pacific (southeast + northeast) |
| `option.AreaCN` | Chinese mainland (east + north) |

## AgentKit packages

`Area` on `agentkit.NewAgoraClient` controls regional API routing (domain pool and path). Vendor constructors are organized in separate packages:

| Package | Description |
|---|---|
| `agentkit` | Global/default agent builder; supports cascading and MLLM |
| `agentkit/vendors` | Global/default vendor constructors |
| `agentkit/cn` | Mainland China facade; sets `option.AreaCN` on the client |
| `agentkit/cn/vendors` | Mainland China vendor constructors |

MLLM constructors follow the same boundary:

| MLLM | Package | Agora routing |
|---|---|---|
| Azure OpenAI Realtime | `agentkit/vendors.NewAzureOpenAIRealtime` | Global `agentkit` client with `AreaUS`, `AreaEU`, or `AreaAP` as appropriate |
| Qwen Omni | `agentkit/cn/vendors.NewQwenOmni` | Mainland China `agentkit/cn` client, which uses `AreaCN` |

`Area` selects the Agora control-plane route. The vendor package selects the provider catalog for that route, so do not combine the global Azure constructor with the CN facade or the CN Qwen Omni constructor with the global facade.

Global/default example:

```go
package main

import (
    "github.com/AgoraIO/agora-agents-go/v2/agentkit"
    "github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
    "github.com/AgoraIO/agora-agents-go/v2/option"
)

func main() {
    client := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
        Area:           option.AreaUS,
        AppID:          "your-app-id",
        AppCertificate: "your-app-certificate",
    })

    _ = agentkit.NewAgent(client).WithStt(
        vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{Model: "nova-3", Language: "en-US"}),
    ).WithLlm(
        vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-4o-mini"}),
    ).WithTts(
        vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{Model: "speech_2_6_turbo", VoiceID: "English_captivating_female1"}),
    )
}
```

If you omit `WithStt()` in the global/default facade, AgentKit falls back to `asr.vendor = "ares"`.

CN example:

```go
package main

import (
    agentkit "github.com/AgoraIO/agora-agents-go/v2/agentkit/cn"
    vendors "github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors"
)

func main() {
    client := agentkit.NewAgoraClient(agentkit.ClientOptions{
        AppID:          "your-app-id",
        AppCertificate: "your-app-certificate",
    })

    _ = agentkit.NewAgent(client).WithStt(
        vendors.NewFengmingSTT(),
    ).WithLlm(
        vendors.NewAliyun(vendors.AliyunOptions{
            APIKey:  "api-key",
            BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
            Model:   "qwen-plus",
        }),
    ).WithTts(
        vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
            Key:   "minimax-key",
            Model: "speech-01-turbo",
            VoiceSetting: &vendors.MiniMaxVoiceSetting{
                VoiceID: "female-shaonv",
            },
        }),
    )
}
```

If you omit `WithStt()` in the CN facade, AgentKit falls back to `asr.vendor = "fengming"`.

## How the domain pool works

Each area has two regional domain prefixes and two domain suffixes. The pool:

1. Starts with the first regional prefix and the first domain suffix.
2. Resolves the best domain suffix via DNS every 30 seconds.
3. Constructs the full URL as `https://{prefix}.{suffix}{path}`.

## Region-to-domain mapping

| Area | Primary prefix | Fallback prefix | Primary suffix | Fallback suffix | Path |
|---|---|---|---|---|---|
| `option.AreaUS` | `api-us-west-1` | `api-us-east-1` | `agora.io` | `sd-rtn.com` | `/api/conversational-ai-agent` |
| `option.AreaEU` | `api-eu-west-1` | `api-eu-central-1` | `agora.io` | `sd-rtn.com` | `/api/conversational-ai-agent` |
| `option.AreaAP` | `api-ap-southeast-1` | `api-ap-northeast-1` | `agora.io` | `sd-rtn.com` | `/api/conversational-ai-agent` |
| `option.AreaCN` | `api-cn-east-1` | `api-cn-north-1` | `sd-rtn.com` | `agora.io` | `/cn/api/conversational-ai-agent` |

Note: `option.AreaCN` uses `sd-rtn.com` as the primary suffix and the `/cn/api/conversational-ai-agent` path by default.

## When to override lower-level routing

The generated `client` and `option` packages also expose lower-level routing primitives such as `option.WithBaseURL`, `option.WithPool`, and `core.NewPool`. Use those only when you are building direct generated-client integrations. For realtime agent sessions, prefer `AgoraClientOptions.Area` so auth, token generation, and regional routing stay together in AgentKit.
