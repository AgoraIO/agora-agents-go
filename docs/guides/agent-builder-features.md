---
sidebar_position: 5
title: Agent Builder Features
description: Configure SAL, advanced features, parameters, geofence, labels, RTC, filler words, and more.
---

# Agent Builder Features

The Agent builder supports many configuration options beyond the core LLM, TTS, and STT vendors. This guide shows how to use each feature.

For string values with a finite set of options (e.g. `data_channel`, `sal_mode`, `area`), use the short constant aliases exported from the `agentkit` package (e.g. `agentkit.DataChannelRtm`, `agentkit.SalModeLocking`) instead of the verbose Fern-generated names or raw strings.

## Overview

| Feature | Option / Method | Description |
|---|---|---|
| `sal` | `WithSalConfig` / `WithSal` | Selective Attention Locking — speaker recognition and noise suppression |
| `advancedFeatures` | `WithAdvancedFeatures` | Enable MLLM, RTM, SAL, tools |
| `tools` | `WithTools` | Enable tool invocation for MCP servers and inline LLM tools |
| `parameters` | `WithParameters` | Silence config, farewell config, data channel |
| `failureMessage` | LLM/MLLM vendor option | Message spoken when LLM fails |
| `maxHistory` | LLM vendor option | Max conversation turns in LLM context |
| `geofence` | `WithGeofence` | Restrict backend server regions |
| `labels` | `WithLabels` | Custom key-value labels (returned in callbacks) |
| `rtc` | `WithRtc` | RTC media encryption |
| `fillerWords` | `WithFillerWords` | Filler words while waiting for LLM |

Use `NewAgent(client, opts...)` for constructor options, or chain methods on an existing agent. All methods return a **new** `*Agent` (immutable).

## SAL (Selective Attention Locking)

SAL helps the agent focus on the primary speaker and suppress background noise. Enable it via `WithAdvancedFeatures` and configure with `WithSal`:

```go
import (
    Agora "github.com/AgoraIO/agora-agents-go/v2"
    "github.com/AgoraIO/agora-agents-go/v2/agentkit"
    "github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
)

agent := agentkit.NewAgent(client,
    agentkit.WithAdvancedFeatures(&agentkit.AdvancedFeatures{
        EnableSal: Agora.Bool(true),
    }),
    agentkit.WithSalConfig(&agentkit.SalConfig{
        SalMode: agentkit.SalModeLocking.Ptr(),
        SampleUrls: map[string]string{
            "primary-speaker": "https://example.com/voiceprint.pcm",
        },
    }),
).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
    APIKey:  "<key>",
    BaseURL: "https://api.openai.com/v1/chat/completions",
    Model:   "gpt-4o-mini",
    SystemMessages: []map[string]interface{}{
        {"role": "system", "content": "You are a helpful assistant."},
    },
})).
  WithTts(vendors.NewElevenLabsTTS(/* ... */)).
  WithStt(vendors.NewDeepgramSTT(/* ... */))
```

`SalMode` can be `agentkit.SalModeLocking` or `agentkit.SalModeRecognition`.

## Advanced Features

Enable MLLM, RTM, SAL, or tools:

```go
// MLLM mode (see mllm-flow guide)
agent := agentkit.NewAgent(client).WithMllm(/* ... */)

// RTM signaling for custom data delivery
agent := agentkit.NewAgent(client,
    agentkit.WithAdvancedFeatures(&agentkit.AdvancedFeatures{
        EnableRtm: Agora.Bool(true),
    }),
)

// Enable tool invocation via MCP servers or inline LLM tools
agent := agentkit.NewAgent(client,
    agentkit.WithTools(true),
)
```

Note: Use `Agora.Bool(true)` for optional boolean fields — the API expects `*bool`.

## Session Parameters

Configure silence handling, farewell behavior, and data channel:

```go
agent := agentkit.NewAgent(client,
    agentkit.WithParameters(&agentkit.SessionParams{
        SilenceConfig: &agentkit.SilenceConfig{
            TimeoutMs: Agora.Int(10000),
            Action:    agentkit.SilenceActionSpeak.Ptr(),
            Content:   Agora.String("I'm still here. Take your time."),
        },
        FarewellConfig: &agentkit.FarewellConfig{
            GracefulEnabled:        Agora.Bool(true),
            GracefulTimeoutSeconds: Agora.Int(10),
        },
        DataChannel: agentkit.DataChannelRtm.Ptr(),
    }),
).WithLlm(/* ... */).WithTts(/* ... */).WithStt(/* ... */)
```

## Failure Message and Max History

```go
agent := agentkit.NewAgent(client).
    WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
        APIKey:         "<key>",
        BaseURL:        "https://api.openai.com/v1/chat/completions",
        Model:          "gpt-4o-mini",
        FailureMessage: "Something went wrong.",
        MaxHistory:     Agora.Int(15),
    })).WithTts(/* ... */).WithStt(/* ... */)
```

## Geofence

Restrict which geographic regions the backend can use:

```go
import Agora "github.com/AgoraIO/agora-agents-go/v2"

agent := agentkit.NewAgent(client,
    agentkit.WithGeofence(&agentkit.GeofenceConfig{
        Area: agentkit.GeofenceAreaNorthAmerica,
    }),
).WithLlm(/* ... */).WithTts(/* ... */).WithStt(/* ... */)

// Global with exclusion
agent := agentkit.NewAgent(client,
    agentkit.WithGeofence(&agentkit.GeofenceConfig{
        Area:        agentkit.GeofenceAreaGlobal,
        ExcludeArea: agentkit.GeofenceExcludeAreaEurope.Ptr(),
    }),
).WithLlm(/* ... */).WithTts(/* ... */).WithStt(/* ... */)
```

Available `GeofenceArea` constants: `agentkit.GeofenceAreaGlobal`, `GeofenceAreaNorthAmerica`, `GeofenceAreaEurope`, `GeofenceAreaAsia`, `GeofenceAreaIndia`, `GeofenceAreaJapan`.

## Labels

Attach custom labels returned in notification callbacks:

```go
agent := agentkit.NewAgent(client,
    agentkit.WithLabels(map[string]string{
        "environment": "production",
        "team":        "support",
        "version":     "1.2.0",
    }),
).WithLlm(/* ... */).WithTts(/* ... */).WithStt(/* ... */)
```

## RTC Encryption

Configure RTC media encryption:

```go
agent := agentkit.NewAgent(client,
    agentkit.WithRtc(&agentkit.RtcConfig{
        EncryptionKey:  Agora.String("your-32-byte-key"),
        EncryptionMode: Agora.Int(5), // AES_128_GCM
    }),
).WithLlm(/* ... */).WithTts(/* ... */).WithStt(/* ... */)
```

## Filler Words

Play filler words while waiting for the LLM response:

```go
agent := agentkit.NewAgent(client,
    agentkit.WithFillerWords(&agentkit.FillerWordsConfig{
        Enable: Agora.Bool(true),
        Trigger: &agentkit.FillerWordsTrigger{
            Mode: Agora.String(agentkit.FillerWordsTriggerModeFixedTime),
            FixedTimeConfig: &agentkit.FillerWordsTriggerFixedTimeConfig{
                ResponseWaitMs: Agora.Int(2000),
            },
        },
        Content: &agentkit.FillerWordsContent{
            Mode: agentkit.FillerWordsContentModeStatic.Ptr(),
            StaticConfig: &agentkit.FillerWordsContentStaticConfig{
                Phrases:      []string{"Let me think...", "One moment...", "Hmm..."},
                SelectionRule: agentkit.FillerWordsSelectionRuleShuffle.Ptr(),
            },
        },
    }),
).WithLlm(/* ... */).WithTts(/* ... */).WithStt(/* ... */)
```

Use generated mode to create a filler phrase from the last user message. When
`GeneratedConfig` is omitted, the service uses its default generator:

```go
agent := agentkit.NewAgent(client,
    agentkit.WithFillerWords(&agentkit.FillerWordsConfig{
        Enable: Agora.Bool(true),
        Trigger: &agentkit.FillerWordsTrigger{
            Mode: Agora.String(agentkit.FillerWordsTriggerModeFixedTime),
            FixedTimeConfig: &agentkit.FillerWordsTriggerFixedTimeConfig{
                ResponseWaitMs: Agora.Int(1200),
            },
        },
        Content: &agentkit.FillerWordsContent{
            Mode: agentkit.FillerWordsContentModeGenerated.Ptr(),
        },
    }),
).WithLlm(/* ... */).WithTts(/* ... */).WithStt(/* ... */)
```

You can also configure an OpenAI-compatible filler LLM independently from the
main business LLM:

```go
Content: &agentkit.FillerWordsContent{
    Mode: agentkit.FillerWordsContentModeGenerated.Ptr(),
    StaticConfig: &agentkit.FillerWordsContentStaticConfig{
        Phrases: []string{"One moment..."},
    },
    GeneratedConfig: &agentkit.FillerWordsContentGeneratedConfig{
        LlmProvider: &agentkit.FillerWordsContentGeneratedLlmProvider{
            URL:    "https://api.openai.com/v1/chat/completions",
            APIKey: os.Getenv("FILLER_LLM_API_KEY"),
            Params: map[string]interface{}{"model": "gpt-4o-mini"},
        },
        Prompt:           Agora.String("Reply with a short conversational filler phrase."),
        FallbackStrategy: Agora.String(agentkit.FillerWordsFallbackStrategyStatic),
    },
},
```

## Getters

Read back configuration via getter methods:

```go
agent := agentkit.NewAgent(client,
    agentkit.WithGeofence(&agentkit.GeofenceConfig{
        Area: agentkit.GeofenceAreaEurope,
    }),
    agentkit.WithLabels(map[string]string{"env": "staging"}),
)

agent.Geofence()       // *GeofenceConfig
agent.Labels()         // map[string]string
agent.Sal()            // *SalConfig
agent.AdvancedFeatures()
agent.Parameters()
agent.FailureMessage()
agent.Rtc()
agent.FillerWords()
```

## Chaining All Features

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

    agent := agentkit.NewAgent(client,
        agentkit.WithAdvancedFeatures(&agentkit.AdvancedFeatures{
            EnableRtm: Agora.Bool(true),
        }),
        agentkit.WithParameters(&agentkit.SessionParams{
            SilenceConfig: &agentkit.SilenceConfig{
                TimeoutMs: Agora.Int(8000),
                Action:    agentkit.SilenceActionSpeak.Ptr(),
                Content:   Agora.String("I'm listening."),
            },
            FarewellConfig: &agentkit.FarewellConfig{
                GracefulEnabled:        Agora.Bool(true),
                GracefulTimeoutSeconds: Agora.Int(5),
            },
        }),
        agentkit.WithGeofence(&agentkit.GeofenceConfig{
            Area: agentkit.GeofenceAreaNorthAmerica,
        }),
        agentkit.WithLabels(map[string]string{
            "app":     "voice-assistant",
            "version": "2.0",
        }),
        agentkit.WithFillerWords(&agentkit.FillerWordsConfig{
            Enable: Agora.Bool(true),
            Trigger: &agentkit.FillerWordsTrigger{
                Mode: Agora.String("fixed_time"),
                FixedTimeConfig: &agentkit.FillerWordsTriggerFixedTimeConfig{
                    ResponseWaitMs: Agora.Int(1500),
                },
            },
            Content: &agentkit.FillerWordsContent{
                Mode: agentkit.FillerWordsContentModeStatic.Ptr(),
                StaticConfig: &agentkit.FillerWordsContentStaticConfig{
                    Phrases:       []string{"Let me think...", "One moment please."},
                    SelectionRule: agentkit.FillerWordsSelectionRuleShuffle.Ptr(),
                },
            },
        }),
    ).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
        APIKey:  "<openai_key>",
        BaseURL: "https://api.openai.com/v1/chat/completions",
        Model:   "gpt-4o-mini",
        SystemMessages: []map[string]interface{}{
            {"role": "system", "content": "You are a helpful voice assistant."},
        },
        GreetingMessage: "Hello! How can I help?",
        FailureMessage:  "Sorry, I had trouble processing that.",
        MaxHistory:      Agora.Int(20),
    })).WithTts(vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
        Key:     "<elevenlabs_key>",
        ModelID:  "eleven_turbo_v2_5",
        VoiceID:    "<voice_id>",
        BaseURL:    "wss://api.elevenlabs.io/v1",
    })).WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
        APIKey:   "<deepgram_key>",
        Model:    "nova-2",
        Language: "en-US",
    }))

    session := agent.CreateSession(agentkit.CreateSessionOptions{
        Name:        fmt.Sprintf("conversation-%d", time.Now().UnixMilli()),
        Channel:     fmt.Sprintf("demo-channel-%d", time.Now().UnixMilli()),
        AgentUID:   "1001",
        RemoteUIDs: []string{"1002"},
    })

    ctx := context.Background()
    if _, err := session.Start(ctx); err != nil {
        log.Fatalf("Failed to start: %v", err)
    }
    defer session.Stop(ctx)
}
```

## Next steps

- [Agent Reference](../reference/agent.md) — full API signatures
- [Cascading Flow](./cascading-flow.md) — ASR → LLM → TTS setup
- [MLLM Flow](./mllm-flow.md) — multimodal flow with `mllm.enable`
- [Regional Routing](./regional-routing.md) — client area and geofence
