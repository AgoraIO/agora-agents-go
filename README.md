# Agora Agents Go

[![fern shield](https://img.shields.io/badge/%F0%9F%8C%BF-Built%20with%20Fern-brightgreen)](https://buildwithfern.com?utm_source=github&utm_medium=github&utm_campaign=readme&utm_source=https%3A%2F%2Fgithub.com%2FAgoraIO%2Fagora-agents-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/AgoraIO/agora-agents-go/v2.svg)](https://pkg.go.dev/github.com/AgoraIO/agora-agents-go/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/AgoraIO/agora-agents-go)](https://goreportcard.com/report/github.com/AgoraIO/agora-agents-go)
[![Release](https://img.shields.io/github/v/release/AgoraIO/agora-agents-go?sort=semver)](https://github.com/AgoraIO/agora-agents-go/releases)

The Agora Conversational AI SDK provides convenient access to the Agora Conversational AI APIs, 
enabling you to build voice-powered AI agents with support for both cascading flows (ASR -> LLM -> TTS) 
and multimodal flows (MLLM) for real-time audio processing.

## Install

```sh
go get github.com/AgoraIO/agora-agents-go/v2@v2.0.0
```

## Requirements

- Go 1.21+

## Quick Start

Start with the `Agent` builder: create a client with app credentials, choose your ASR, LLM, and TTS providers, then start a session. Omit vendor API keys for supported Agora-managed models, or provide keys when you want BYOK.
Set Agora interaction language with `TurnDetectionConfig.Language`; provider-specific STT language values remain under `asr.params`. Ares uses only the REST `asr.language` value sourced from `TurnDetectionConfig.Language`.

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"

    Agora "github.com/AgoraIO/agora-agents-go/v2"
    "github.com/AgoraIO/agora-agents-go/v2/agentkit"
    "github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
    "github.com/AgoraIO/agora-agents-go/v2/option"
)

const (
    agentPrompt = "You are a concise, technically credible voice assistant. Keep replies short unless the user asks for detail."
    greeting    = "Hi there! I am your Agora voice assistant. How can I help?"
)

func stringPtr(v string) *string { return &v }
func intPtr(v int) *int { return &v }
func float64Ptr(v float64) *float64 { return &v }
func boolPtr(v bool) *bool { return &v }

func requireEnv(name string) (string, error) {
    value := os.Getenv(name)
    if value == "" {
        return "", fmt.Errorf("missing required environment variable: %s", name)
    }
    return value, nil
}

func startConversation(ctx context.Context) (string, error) {
    appID, err := requireEnv("AGORA_APP_ID")
    if err != nil {
        return "", err
    }
    appCertificate, err := requireEnv("AGORA_APP_CERTIFICATE")
    if err != nil {
        return "", err
    }
    expiresIn, err := agentkit.ExpiresInHours(1)
    if err != nil {
        return "", err
    }

    client := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
        Area:           option.AreaUS,
        AppID:          appID,
        AppCertificate: appCertificate,
    })

    agent := agentkit.NewAgent(
        agentkit.WithName(fmt.Sprintf("conversation-%d", time.Now().UnixMilli())),
        agentkit.WithTurnDetectionConfig(&agentkit.TurnDetectionConfig{
            Language: Agora.AsrLanguageEnUs.Ptr(),
            Config: &agentkit.TurnDetectionNestedConfig{
                SpeechThreshold: float64Ptr(0.5),
                StartOfSpeech: &agentkit.StartOfSpeechConfig{
                    Mode: agentkit.StartOfSpeechMode("vad"),
                    VadConfig: &agentkit.StartOfSpeechVadConfig{
                        InterruptDurationMs: intPtr(160),
                        PrefixPaddingMs:     intPtr(300),
                    },
                },
                EndOfSpeech: &agentkit.EndOfSpeechConfig{
                    Mode: agentkit.EndOfSpeechMode("vad"),
                    VadConfig: &agentkit.EndOfSpeechVadConfig{
                        SilenceDurationMs: intPtr(480),
                    },
                },
            },
        }),
        agentkit.WithAdvancedFeatures(&agentkit.AdvancedFeatures{
            EnableRtm:   boolPtr(true),
            EnableTools: boolPtr(true),
        }),
        agentkit.WithParameters(&agentkit.SessionParams{
            DataChannel:        &agentkit.DataChannelRtm,
            EnableErrorMessage: boolPtr(true),
        }),
    ).WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
        Model: "nova-3",
        Language: "en",
    })).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
        Model:           "gpt-4o-mini",
        SystemMessages:  []map[string]interface{}{{"role": "system", "content": agentPrompt}},
        GreetingMessage: greeting,
        FailureMessage:  "Please wait a moment.",
        MaxHistory:      intPtr(50),
        Params: map[string]interface{}{
            "max_tokens": 1024,
            "temperature": 0.7,
            "top_p": 0.95,
        },
    })).WithTts(vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
        Model:   "speech_2_6_turbo",
        VoiceID: "English_captivating_female1",
    }))

    session := agent.CreateSession(client, agentkit.CreateSessionOptions{
        Channel:     fmt.Sprintf("demo-channel-%d", time.Now().UnixMilli()),
        AgentUID:    "123456",
        RemoteUIDs:  []string{"*"},
        IdleTimeout: intPtr(30),
        ExpiresIn:   expiresIn,
        Debug:       false,
    })

    return session.Start(ctx)
}

func main() {
    agentID, err := startConversation(context.Background())
    if err != nil {
        panic(err)
    }
    fmt.Println(agentID)
}
```

### Why no token or vendor key in the example?

`AgoraClient` generates the required ConvoAI REST auth and RTC join tokens automatically when you provide `AppID` and `AppCertificate`. For supported Agora-managed models, leave vendor API keys unset; provide keys when you want BYOK.

## AI Studio pipeline IDs

Use `WithPipelineID` when you want a published AI Studio pipeline to provide the base agent configuration:

```go
agent := agentkit.NewAgent(
    agentkit.WithName("support"),
    agentkit.WithPipelineID("studio-pipeline-id"),
)

session := agent.CreateSession(client, agentkit.CreateSessionOptions{
    Channel:    "support-room",
    AgentUID:   "1",
    RemoteUIDs: []string{"100"},
})
```

You can override it per session:

```go
session := agent.CreateSession(client, agentkit.CreateSessionOptions{
    Channel:    "support-room",
    AgentUID:   "1",
    RemoteUIDs: []string{"100"},
    PipelineID: "session-pipeline-id",
})
```

AgentKit sends the resolved value as the top-level `/join` field `pipeline_id`, not inside `properties`. Explicit Agent config such as `WithLlm`, `WithTts`, `WithStt`, `WithMllm`, and advanced features may send `properties` fields that override the saved pipeline settings.

### BYOK version

Use the same `Agent` builder shape, but provide credentials explicitly when you want vendor-managed billing and routing instead of Agora-managed models.

```go
agent := agentkit.NewAgent(agentkit.WithTurnDetectionConfig(&agentkit.TurnDetectionConfig{
    Language: Agora.AsrLanguageEnUs.Ptr(),
})).WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
    APIKey:   os.Getenv("DEEPGRAM_API_KEY"),
    Model:    "nova-3",
    Language: "en",
})).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
    APIKey:          os.Getenv("OPENAI_API_KEY"),
    BaseURL:         "https://api.openai.com/v1/chat/completions",
    Model:           "gpt-4o-mini",
    SystemMessages:  []map[string]interface{}{{"role": "system", "content": agentPrompt}},
    GreetingMessage: greeting,
    MaxTokens:       intPtr(1024),
    Temperature:     float64Ptr(0.7),
    TopP:            float64Ptr(0.95),
})).WithTts(vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
    Key:     os.Getenv("MINIMAX_API_KEY"),
    GroupID: os.Getenv("MINIMAX_GROUP_ID"),
    Model:   "speech_2_6_turbo",
    VoiceID: "English_captivating_female1",
    URL:     "wss://api-uw.minimax.io/ws/v1/t2a_v2",
}))
```

Migrating from `github.com/AgoraIO-Conversational-AI/agent-server-sdk-go`? Update your module path and imports to `github.com/AgoraIO/agora-agents-go/v2` — see the [v2.0.0 changelog](./changelog.md#v200--2026-05-21) or [installation guide](./docs/getting-started/installation.md#migrating-from-a-previous-module-path).

## BYOK

If you want to bring your own vendor credentials instead of using Agora-managed models, use the BYOK guide:

- [BYOK Guide](./docs/guides/byok.md)

## MLLM (Realtime / Multimodal)

Use `WithMllm()` for OpenAI Realtime, Gemini Live, Vertex AI, or xAI Grok. No STT, LLM, or TTS vendor is needed when MLLM mode is enabled.

```go
agent := agentkit.NewAgent(
    agentkit.WithName("realtime-assistant"),
).WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
    APIKey:          os.Getenv("OPENAI_API_KEY"),
    Model:           "gpt-4o-realtime-preview",
    GreetingMessage: "Hello! Ready to chat.",
}))
```

See the [MLLM Flow guide](./docs/guides/mllm-flow.md) for full examples with Gemini Live, Vertex AI, and xAI Grok.

> Avatars are not supported with MLLM. The avatar publisher requires the cascading ASR + LLM + TTS pipeline; combining `WithMllm()` with `WithAvatar()` returns an error from `Agent.ToProperties()` and `AgentSession.Start()`.

## Avatars

AgentKit supports LiveAvatar, Generic Avatar, Anam, Akool, and deprecated HeyGen. Avatar `AgoraToken` is optional: when omitted, `session.Start()` generates a token using the same ConvoAI token format as the agent token, scoped to the avatar `AgoraUID`. Avatars require the cascading ASR + LLM + TTS pipeline (not MLLM).

See the [Avatar Integration guide](./docs/guides/avatars.md) for sample-rate requirements and Generic Avatar setup.

## Documentation

API reference documentation is available [here](https://docs.agora.io/en/conversational-ai/overview).

## Reference

A full reference for this library is available [here](https://github.com/AgoraIO/agora-agents-go/blob/HEAD/./reference.md).

## Usage

Instantiate the high-level client with app credentials:

```go
package example

import (
    "github.com/AgoraIO/agora-agents-go/v2/agentkit"
    option "github.com/AgoraIO/agora-agents-go/v2/option"
)

func do() {
    client := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
        Area:           option.AreaUS,
        AppID:          "your-app-id",
        AppCertificate: "your-app-certificate",
    })
    _ = client
}
```

## Contributing

While we value open-source contributions to this SDK, this library is generated programmatically.
Additions made directly to this library would have to be moved over to our generation code,
otherwise they would be overwritten upon the next generated release. Feel free to open a PR as
a proof of concept, but know that we will not be able to merge it as-is. We suggest opening
an issue first to discuss with us!

On the other hand, contributions to the README are always very welcome!
