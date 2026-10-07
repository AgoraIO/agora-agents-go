---
sidebar_position: 3
title: Avatars
description: Add visual avatars to your agent with token handling and sample rate requirements.
---

# Avatars

Avatars provide a visual representation of your AI agent. The SDK supports LiveAvatar, Generic Avatar, Anam, Akool, deprecated HeyGen integrations, and the CN-only SenseTime and Spatius avatar integrations. Avatar sessions currently require the cascading ASR/LLM/TTS pipeline; MLLM sessions do not support avatars. Vendors that publish an Agora video stream use a separate avatar UID and token from the voice agent.

For mainland China integrations, use `agentkit/cn` with `NewSensetimeAvatar` or `NewSpatiusAvatar` from `agentkit/cn/vendors`. See [CN AgentKit](./cn-agentkit.md) for routing details.

## Agent Token vs Avatar Token

Voice agents and video avatars both use ConvoAI-compatible Agora tokens. They must be scoped to different UIDs:

| Purpose | Field | UID | Default behavior |
|---|---|---|---|
| Voice agent | `properties.token` | `agent_rtc_uid` | Generated from session `AgentUID` when `Token` is omitted |
| Avatar video stream | `avatar.params.agora_token` | `avatar.params.agora_uid` | Auto-generated for LiveAvatar, HeyGen, Generic, SenseTime, and Spatius when `AgoraToken` is omitted |

Use a unique avatar `AgoraUID`; do not reuse the session `AgentUID`. If you provide `AgoraToken`, the SDK uses it as-is and does not overwrite it.

## Avatar Vendors

| Vendor | Constructor | Required Sample Rate | Required Fields |
|---|---|---|---|
| LiveAvatar | `vendors.NewLiveAvatarAvatar` | 24kHz (`SampleRate24kHz`) | `APIKey`, `Quality` (low/medium/high), `AgoraUID` |
| HeyGen (deprecated name) | `vendors.NewHeyGenAvatar` | 24kHz (`SampleRate24kHz`) | `APIKey`, `Quality` (low/medium/high), `AgoraUID` |
| Akool | `vendors.NewAkoolAvatar` | 16kHz (`SampleRate16kHz`) | `APIKey` |
| Anam | `vendors.NewAnamAvatar` | Provider-managed | `APIKey` |
| Generic | `vendors.NewGenericAvatar` | Vendor-dependent; not enforced by AgentKit | `APIKey`, `APIBaseURL`, `AvatarID`, `AgoraUID` |
| SenseTime (CN) | `cn/vendors.NewSensetimeAvatar` | Not enforced by AgentKit | `AgoraUID`, `AppID`, `AppKey` |
| Spatius (CN) | `cn/vendors.NewSpatiusAvatar` | Not enforced by AgentKit | `SpatiusAPIKey`, `SpatiusAppID`, `SpatiusAvatarID`, `AgoraUID` |

## Tavus and Protoface aliases

Import `github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors` to use
`Tavus` or `Protoface`. Both are Go type aliases of `GenericAvatar`;
`TavusOptions` and `ProtofaceOptions` alias `GenericAvatarOptions`.
Their constructors use the same validation and configuration, serialize with
`vendor: "generic"`, and inherit Generic Avatar token handling and sample rate behavior.

```go
tavus := vendors.NewTavus(vendors.TavusOptions{
    APIKey:     "<tavus_key>",
    APIBaseURL: "https://tavusapi.com/v2/conversations/agora",
    AvatarID:   "<tavus_avatar_id>",
    AgoraUID:   "2001",
})
protoface := vendors.NewProtoface(vendors.ProtofaceOptions{
    APIKey:     "<protoface_key>",
    APIBaseURL: "<protoface_api_base_url>",
    AvatarID:   "<protoface_avatar_id>",
    AgoraUID:   "2002",
})

agent := agentkit.NewAgent(client).WithAvatar(tavus)
// Alternatively: agentkit.NewAgent(client).WithAvatar(protoface)
```

Supply your provider's API base URL explicitly. Optional `AgoraToken`, `AgoraAppID`,
`AgoraChannel`, `Enable`, and `AdditionalParams` work exactly as for `GenericAvatar`.

## Generic Avatar Example

```go
sampleRate := vendors.SampleRate24kHz // or 16kHz, depending on your provider

agent := agentkit.NewAgent(client).WithLlm(
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

For Generic avatars, `agora_appid`, `agora_channel`, and `agora_token` are filled from the session when omitted. For LiveAvatar, HeyGen, Generic, SenseTime, and Spatius, AgentKit auto-generates `agora_token` when `agora_uid` is set and `agora_token` is omitted.

## LiveAvatar Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

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

    agent := agentkit.NewAgent(client).WithLlm(
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

agent := agentkit.NewAgent(client).WithLlm(
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

## SenseTime Avatar Example (CN)

SenseTime is available only from the CN facade (`agentkit/cn`). Use `agentkit/cn/vendors.NewSensetimeAvatar` with the CN agent builder. When `AgoraToken` is omitted, AgentKit auto-generates `agora_token` from the session channel and avatar `AgoraUID` (same behavior as LiveAvatar).

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
        AppID:          "<app_id>",
        AppCertificate: "<app_cert>",
    })

    agent := agentkit.NewAgent(client).WithStt(
        vendors.NewTencentSTT(vendors.TencentSTTOptions{
            Key:    "<key>",
            AppID:  "<app_id>",
            Secret: "<secret>",
        }),
    ).WithLlm(
        vendors.NewTencentLLM(vendors.TencentLLMOptions{
            APIKey:  "<api_key>",
            Model:   "hunyuan-turbos-latest",
            BaseURL: "https://api.hunyuan.cloud.tencent.com/v1/chat/completions",
        }),
    ).WithTts(
        vendors.NewTencentTTS(vendors.TencentTTSOptions{
            AppID:     "<app_id>",
            SecretID:  "<secret_id>",
            SecretKey: "<secret_key>",
        }),
    ).WithAvatar(
        vendors.NewSensetimeAvatar(vendors.SensetimeAvatarOptions{
            AgoraUID: "2001", // distinct from session AgentUID
            AppID:    "<sensetime_app_id>",
            AppKey:   "<sensetime_app_key>",
            SceneList: []vendors.SensetimeScene{
                {
                    DigitalRole: vendors.SensetimeDigitalRole{
                        FaceFeatureID: "<face_feature_id>",
                        Position: vendors.SensetimePosition{
                            X: 0,
                            Y: 0,
                        },
                        URL: "<avatar_model_package_url>",
                    },
                },
            },
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

See [Vendors Reference — CN Avatar Vendors](../reference/vendors.md#newsensetimeavatar) for field details.

## Spatius Avatar Example (CN)

Spatius is available only from the CN facade (`agentkit/cn`). Use `agentkit/cn/vendors.NewSpatiusAvatar` with the CN agent builder. When `AgoraToken` is omitted, AgentKit auto-generates `agora_token` from the session channel and avatar `AgoraUID`.

```go
agent := agentkit.NewAgent(client).WithAvatar(
    vendors.NewSpatiusAvatar(vendors.SpatiusAvatarOptions{
        SpatiusAPIKey:   "<spatius_api_key>",
        SpatiusAppID:    "<spatius_app_id>",
        SpatiusAvatarID: "<spatius_avatar_id>",
        AgoraUID:        "2001",
        Region:          "cn-beijing",
    }),
)
```

## Sample Rate Validation and Panic Behavior

The Go SDK enforces TTS/avatar sample rate constraints at runtime. This differs from TypeScript (compile-time phantom types) and Python (`ValueError`).

### When does the panic occur?

`WithAvatar()` panics immediately only when the avatar vendor is token-managed **and** publishes a non-zero `RequiredSampleRate()` that does not match an already-configured TTS sample rate. In practice this applies to **LiveAvatar and HeyGen (24kHz)**:

```
Avatar requires TTS sample rate of 24000 Hz, but TTS is configured with 16000 Hz. Please update your TTS sample_rate to 24000.
```

**Akool (16kHz)**, **Anam**, **Generic**, **SenseTime**, and **Spatius** do not panic at `WithAvatar()` time. Their constraints are checked in `AgentSession.Start()` via `ValidateTtsSampleRate` or provider-specific validation.

#### Anam options

| Field | Type | Required | Description |
|---|---|---|---|
| `APIKey` | `string` | Yes | Anam API key |
| `AvatarID` | `string` | No | Anam avatar identifier (wire key: `params.avatar_id`) |
| `Enable` | `*bool` | No | Enable or disable the avatar |
| `AdditionalParams` | `map[string]interface{}` | No | Additional vendor params |

### Why panic instead of returning an error?

In Go, `panic` is idiomatic for programmer errors — configuration mistakes that should be caught during development, not handled at runtime. A sample rate mismatch is a static configuration error, not a transient failure.

### How to prevent the panic

Always configure TTS with the correct sample rate **before** calling `WithAvatar()`:

```go
// Correct: TTS sample rate matches avatar requirement
sr := vendors.SampleRate24kHz
agent := agentkit.NewAgent(client).
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
agent := agentkit.NewAgent(client).
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

## SensetimeAvatarOptions Fields (CN)

| Field | Type | Required | Description |
|---|---|---|---|
| `AgoraUID` | `string` | Yes | UID for the avatar's video stream; must differ from session `AgentUID` |
| `AgoraToken` | `string` | No | Avatar Agora token. Auto-generated when omitted. |
| `AppID` | `string` | Yes | SenseTime application ID (serialized as `params.appId`) |
| `AppKey` | `string` | Yes | SenseTime application key |
| `SceneList` | `[]SensetimeScene` | No | Optional scene configuration with digital role, position, and model URL |
| `Enable` | `*bool` | No | Enable or disable the avatar (default: `true`) |
| `AdditionalParams` | `map[string]interface{}` | No | Additional vendor params merged into `avatar.params` |

Each `SensetimeScene` contains a `DigitalRole` (`FaceFeatureID`, `Position` with `X`/`Y`, and model package `URL`).

## SpatiusAvatarOptions Fields (CN)

| Field | Type | Required | Description |
|---|---|---|---|
| `SpatiusAPIKey` | `string` | Yes | Spatius API key |
| `SpatiusAppID` | `string` | Yes | Spatius application ID |
| `SpatiusAvatarID` | `string` | Yes | Spatius avatar ID |
| `AgoraUID` | `string` | Yes | UID for the avatar's video stream; must differ from session `AgentUID` |
| `AgoraToken` | `string` | No | Avatar Agora token. Auto-generated when omitted. |
| `Region` | `string` | No | Spatius service region, for example `cn-beijing` |
| `SampleRate` | `*vendors.SampleRate` | No | Avatar audio sample rate hint |
| `SessionExpireMinutes` | `*int` | No | Session validity duration in minutes |
| `Enable` | `*bool` | No | Enable or disable the avatar |
| `AdditionalParams` | `map[string]interface{}` | No | Additional vendor params merged into `avatar.params` |
