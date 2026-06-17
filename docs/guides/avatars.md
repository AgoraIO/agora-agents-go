---
sidebar_position: 3
title: Avatars
description: Add visual avatars to your agent with token handling and sample rate requirements.
---

# Avatars

Avatars provide a visual representation of your AI agent. The SDK supports LiveAvatar, Generic Avatar, Anam, Akool, and deprecated HeyGen integrations. Avatar sessions currently require the cascading ASR/LLM/TTS pipeline; MLLM sessions do not support avatars. Vendors that publish an Agora video stream use a separate avatar UID and token from the voice agent.

## Agent Token vs Avatar Token

Voice agents and video avatars both use ConvoAI-compatible Agora tokens. They must be scoped to different UIDs:

| Purpose | Field | UID | Default behavior |
|---|---|---|---|
| Voice agent | `properties.token` | `agent_rtc_uid` | Generated from session `AgentUID` when `Token` is omitted |
| Avatar video stream | `avatar.params.agora_token` | `avatar.params.agora_uid` | Generated from avatar `AgoraUID` when `AgoraToken` is omitted |

Use a unique avatar `AgoraUID`; do not reuse the session `AgentUID`. If you provide `AgoraToken`, the SDK uses it as-is and does not overwrite it.

## Avatar Vendors

| Vendor | Constructor | Required Sample Rate | Required Fields |
|---|---|---|---|
| LiveAvatar | `vendors.NewLiveAvatarAvatar` | 24kHz (`SampleRate24kHz`) | `APIKey`, `Quality` (low/medium/high), `AgoraUID` |
| HeyGen (deprecated name) | `vendors.NewHeyGenAvatar` | 24kHz (`SampleRate24kHz`) | `APIKey`, `Quality` (low/medium/high), `AgoraUID` |
| Akool | `vendors.NewAkoolAvatar` | 16kHz (`SampleRate16kHz`) | `APIKey` |
| Anam | `vendors.NewAnamAvatar` | Provider-managed | `APIKey` |
| Generic | `vendors.NewGenericAvatar` | Vendor-dependent; not enforced by AgentKit | `APIKey`, `APIBaseURL`, `AvatarID`, `AgoraUID` |

## Generic Avatar Example

```go
sampleRate := vendors.SampleRate24kHz // or 16kHz, depending on your provider

agent := agentkit.NewAgent(
    agentkit.WithName("generic-avatar"),
).WithLlm(
    vendors.NewOpenAI(vendors.OpenAIOptions{APIKey: "<openai_key>"}),
).WithTts(
    vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
        Key:        "<elevenlabs_key>",
        ModelID:    "eleven_turbo_v2_5",
        VoiceID:    "<voice_id>",
        // Choose the sample rate required by your generic avatar provider.
        SampleRate: &sampleRate,
    }),
).WithAvatar(
    vendors.NewGenericAvatar(vendors.GenericAvatarOptions{
        APIKey:     "<avatar_vendor_key>",
        APIBaseURL: "https://avatar.example.com",
        AvatarID:   "<avatar_id>",
        AgoraUID:   "2001", // distinct from session AgentUID
    }),
)
```

For Generic avatars, `agora_appid`, `agora_channel`, and `agora_token` are filled from the session when omitted. For LiveAvatar and HeyGen, AgentKit auto-generates only `agora_token` when `agora_uid` is set and `agora_token` is omitted.

## Generic Avatar Example

```go
sampleRate := vendors.SampleRate24kHz // or 16kHz, depending on your provider

agent := agentkit.NewAgent(
    agentkit.WithName("generic-avatar"),
).WithLlm(
    vendors.NewOpenAI(vendors.OpenAIOptions{
        APIKey:  "<openai_key>",
        BaseURL: "https://api.openai.com/v1/chat/completions",
    }),
).WithTts(
    vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
        Key:        "<elevenlabs_key>",
        ModelID:    "eleven_turbo_v2_5",
        VoiceID:    "<voice_id>",
        BaseURL:    "wss://api.elevenlabs.io/v1",
        // Choose the sample rate required by your generic avatar provider.
        SampleRate: &sampleRate,
    }),
).WithAvatar(
    vendors.NewGenericAvatar(vendors.GenericAvatarOptions{
        APIKey:     "<avatar_vendor_key>",
        APIBaseURL: "https://avatar.example.com",
        AvatarID:   "<avatar_id>",
        AgoraUID:   "2001", // distinct from session AgentUID
    }),
)
```

For Generic avatars, `agora_appid`, `agora_channel`, and `agora_token` are filled from the session when omitted. For LiveAvatar and HeyGen, AgentKit auto-generates only `agora_token` when `agora_uid` is set and `agora_token` is omitted.

## LiveAvatar Example

```go
package main

import (
    "context"
    "fmt"
    "log"

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

    sr := vendors.SampleRate24kHz

    agent := agentkit.NewAgent(
        agentkit.WithName("avatar-agent"),
    ).WithLlm(
        vendors.NewOpenAI(vendors.OpenAIOptions{
            APIKey:  "<openai_key>",
            BaseURL: "https://api.openai.com/v1/chat/completions",
            Model:   "gpt-4o-mini",
            SystemMessages: []map[string]interface{}{
                {"role": "system", "content": "You are a friendly virtual assistant with a visual avatar."},
            },
            GreetingMessage: "Hello! I can see you and you can see me!",
        }),
    ).WithTts(
        // TTS sample rate MUST match the avatar's required rate (24kHz for LiveAvatar).
        vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
            Key:        "<elevenlabs_key>",
            ModelID:    "eleven_turbo_v2_5",
            VoiceID:    "<voice_id>",
            BaseURL:    "wss://api.elevenlabs.io/v1",
            SampleRate: &sr,
        }),
    ).WithStt(
        vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
            APIKey: "<deepgram_key>",
        }),
    ).WithAvatar(
        vendors.NewLiveAvatarAvatar(vendors.LiveAvatarAvatarOptions{
            APIKey:   "<liveavatar_key>",
            Quality:  "high",
            AgoraUID: "2001",
        }),
    )

    session := agent.CreateSession(client, agentkit.CreateSessionOptions{
        Channel:    "avatar-channel",
        AgentUID:   "1001",
        RemoteUIDs: []string{"1002"},
    })

    ctx := context.Background()

    agentID, err := session.Start(ctx)
    if err != nil {
        log.Fatalf("Failed to start: %v", err)
    }
    fmt.Println("Avatar agent running:", agentID)

    err = session.Stop(ctx)
    if err != nil {
        log.Fatalf("Failed to stop: %v", err)
    }
}
```

## Akool Avatar Example

```go
sr := vendors.SampleRate16kHz

agent := agentkit.NewAgent(
    agentkit.WithName("akool-avatar"),
).WithLlm(
    vendors.NewOpenAI(vendors.OpenAIOptions{
        APIKey:  "<openai_key>",
        BaseURL: "https://api.openai.com/v1/chat/completions",
        Model:   "gpt-4o-mini",
        SystemMessages: []map[string]interface{}{
            {"role": "system", "content": "You are a virtual presenter."},
        },
    }),
).WithTts(
    // Akool requires 16kHz
    vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
        Key:        "<elevenlabs_key>",
        ModelID:    "eleven_turbo_v2_5",
        VoiceID:    "<voice_id>",
        BaseURL:    "wss://api.elevenlabs.io/v1",
        SampleRate: &sr,
    }),
).WithStt(
    vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
        APIKey: "<deepgram_key>",
    }),
).WithAvatar(
    vendors.NewAkoolAvatar(vendors.AkoolAvatarOptions{
        APIKey: "<akool_key>",
    }),
)
```

## Sample Rate Validation and Panic Behavior

The Go SDK enforces the TTS/avatar sample rate constraint at runtime using `panic()`. This differs from TypeScript (compile-time phantom types) and Python (`ValueError`).

### When does the panic occur?

`WithAvatar()` checks if the agent already has a TTS configured with a mismatched sample rate. If so, it panics immediately:

```
Avatar requires TTS sample rate of 24000 Hz, but TTS is configured with 16000 Hz. Please update your TTS sample_rate to 24000.
```

### Why panic instead of returning an error?

In Go, `panic` is idiomatic for programmer errors — configuration mistakes that should be caught during development, not handled at runtime. A sample rate mismatch is a static configuration error, not a transient failure.

### How to prevent the panic

Always configure TTS with the correct sample rate **before** calling `WithAvatar()`:

```go
// Correct: TTS sample rate matches avatar requirement
sr := vendors.SampleRate24kHz
agent := agentkit.NewAgent(...).
    WithTts(vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
        Key:        "<key>",
        ModelID:    "<model>",
        VoiceID:    "<voice>",
        BaseURL:    "wss://api.elevenlabs.io/v1",
        SampleRate: &sr,           // 24kHz for LiveAvatar
    })).
    WithAvatar(vendors.NewLiveAvatarAvatar(vendors.LiveAvatarAvatarOptions{
        APIKey:   "<key>",
        Quality:  "high",
        AgoraUID: "2001",
    }))
```

```go
// Wrong: This panics because TTS is 16kHz but LiveAvatar requires 24kHz.
sr := vendors.SampleRate16kHz
agent := agentkit.NewAgent(...).
    WithTts(vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
        Key:        "<key>",
        ModelID:    "<model>",
        VoiceID:    "<voice>",
        BaseURL:    "wss://api.elevenlabs.io/v1",
        SampleRate: &sr,           // 16kHz — mismatch!
    })).
    WithAvatar(vendors.NewLiveAvatarAvatar(vendors.LiveAvatarAvatarOptions{
        APIKey:   "<key>",
        Quality:  "high",
        AgoraUID: "2001",
    }))
// panic: Avatar requires TTS sample rate of 24000 Hz, but TTS is configured with 16000 Hz. Please update your TTS sample_rate to 24000.
```

### Additional validation at Start

`AgentSession.Start()` also validates the sample rate match before making the API call. If the mismatch was introduced after `WithAvatar()` (e.g., by cloning the agent with a different TTS), `Start()` returns an error instead of panicking.

## LiveAvatarAvatarOptions Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `APIKey` | `string` | Yes | LiveAvatar API key |
| `Quality` | `string` | Yes | `"low"`, `"medium"`, or `"high"` |
| `AgoraUID` | `string` | Yes | UID for the avatar's video stream |
| `AgoraToken` | `string` | No | Avatar token. Auto-generated when omitted. |
| `AvatarID` | `string` | No | LiveAvatar avatar ID |
| `Enable` | `*bool` | No | Enable or disable the avatar |
| `DisableIdleTimeout` | `*bool` | No | Disable the idle timeout |
| `ActivityIdleTimeout` | `*int` | No | Idle timeout in seconds |
| `AdditionalParams` | `map[string]interface{}` | No | Additional vendor params |

## AkoolAvatarOptions Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `APIKey` | `string` | Yes | Akool API key |
