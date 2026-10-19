# Reference

The generated reference below covers the v2 REST clients. The handwritten v3 preview's 20 operations, response types, and conditional headers are documented in the [Conversational AI v3 preview guide](docs/guides/v3-preview.md#low-level-reference-all-20-staging-operations).
## Agents
<details><summary><code>client.Agents.Start(Appid, request) -> *Agora.StartAgentsResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create and start a Conversational AI agent instance.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.StartAgentsRequest{
        Appid: "appid",
        Name: "unique_name",
        Properties: &Agora.StartAgentsRequestProperties{
            Channel: "channel_name",
            Token: "token",
            AgentRtcUID: "1001",
            RemoteRtcUIDs: []string{
                "1002",
            },
            IdleTimeout: Agora.Int(
                120,
            ),
            Asr: &Agora.StartAgentsRequestPropertiesAsr{
                Language: Agora.String(
                    "en-US",
                ),
            },
            Tts: &Agora.Tts{
                Microsoft: &Agora.MicrosoftTts{
                    Params: &Agora.MicrosoftTtsParams{
                        Key: "key",
                        Region: "region",
                        VoiceName: "voice_name",
                    },
                },
            },
            Llm: &Agora.StartAgentsRequestPropertiesLlm{
                URL: "https://api.openai.com/v1/chat/completions",
                APIKey: Agora.String(
                    "<your_llm_key>",
                ),
                SystemMessages: []map[string]any{
                    map[string]any{
                        "role": "system",
                        "content": "You are a helpful chatbot.",
                    },
                },
                Params: map[string]any{
                    "model": "gpt-4o-mini",
                },
                MaxHistory: Agora.Int(
                    32,
                ),
                GreetingMessage: Agora.String(
                    "Hello, how can I assist you today?",
                ),
                FailureMessage: Agora.String(
                    "Please hold on a second.",
                ),
            },
            TurnDetection: &Agora.StartAgentsRequestPropertiesTurnDetection{
                Config: &Agora.StartAgentsRequestPropertiesTurnDetectionConfig{
                    EndOfSpeech: &Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeech{
                        Mode: Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeechModeSemantic.Ptr(),
                    },
                },
            },
        },
    }
client.Agents.Start(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**appid:** `string` — The App ID of the project.
    
</dd>
</dl>

<dl>
<dd>

**name:** `string` — The unique identifier of the agent. The same identifier cannot be used repeatedly.
    
</dd>
</dl>

<dl>
<dd>

**preset:** `*string` 

A comma-separated string of one or more presets. Each preset provides a predefined configuration for ASR, LLM, and TTS. You can specify a preset for any or all of ASR, LLM, and TTS. When a preset is specified, you do not need to provide the endpoint URL, API key, or model for the preset providers. Use the `asr`, `llm`, and `tts` fields to configure additional settings.

Available presets:
- ASR: `deepgram_nova_2`, `deepgram_nova_3`
- LLM: `openai_gpt_4o_mini`, `openai_gpt_4_1_mini`, `openai_gpt_5_nano`, `openai_gpt_5_mini`
- TTS: `minimax_speech_2_6_turbo`, `minimax_speech_2_8_turbo`, `openai_tts_1`
    
</dd>
</dl>

<dl>
<dd>

**pipelineID:** `*string` — The unique ID of a published agent in AI Studio. When provided, the saved agent configuration is used as the base configuration. Any fields specified in `properties` override the corresponding agent settings. When you specify a `pipeline_id`, the `asr`, `tts`, and `llm` fields in `properties` are optional.
    
</dd>
</dl>

<dl>
<dd>

**properties:** `*Agora.StartAgentsRequestProperties` — Configuration details of the agent.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Agents.List(Appid) -> *Agora.ListAgentsResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Retrieve a list of agents that meet the specified conditions.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.ListAgentsRequest{
        Appid: "appid",
    }
client.Agents.List(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**appid:** `string` — The App ID of the project.
    
</dd>
</dl>

<dl>
<dd>

**channel:** `*string` — The channel to query for a list of agents.
    
</dd>
</dl>

<dl>
<dd>

**fromTime:** `*float64` — The start timestamp (in seconds) for the query. Default is 2 hours ago.
    
</dd>
</dl>

<dl>
<dd>

**toTime:** `*float64` — The end timestamp (in seconds) for the query. Default is current time.
    
</dd>
</dl>

<dl>
<dd>

**state:** `*Agora.ListAgentsRequestState` 

The agent state to filter by. Only one state can be specified per query:
- `IDLE` (0): Agent is idle.
- `STARTING` (1): The agent is being started.
- `RUNNING` (2): The agent is running.
- `STOPPING` (3): The agent is stopping.
- `STOPPED` (4): The agent has exited.
- `FAILED` (6): The agent failed to execute.
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — The maximum number of entries returned per page.
    
</dd>
</dl>

<dl>
<dd>

**cursor:** `*string` — The paging cursor, indicating the starting position (`agent_id`) of the next page of results.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Agents.Get(Appid, AgentID) -> *Agora.GetAgentsResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get the current state information of the specified agent instance.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.GetAgentsRequest{
        Appid: "appid",
        AgentID: "agentId",
    }
client.Agents.Get(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**appid:** `string` — The App ID of the project.
    
</dd>
</dl>

<dl>
<dd>

**agentID:** `string` — The agent instance ID you obtained after successfully calling `join` to start a conversational AI agent.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Agents.GetHistory(Appid, AgentID) -> *Agora.GetHistoryAgentsResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get the history of the conversation between the user and the agent.

Call this endpoint while the agent is running to retrieve the conversation history. You can set the maximum number of cached entries using the `llm.max_history` parameter when calling the start agent endpoint. The default value is `32`.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.GetHistoryAgentsRequest{
        Appid: "appid",
        AgentID: "agentId",
    }
client.Agents.GetHistory(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**appid:** `string` — The App ID of the project.
    
</dd>
</dl>

<dl>
<dd>

**agentID:** `string` — The agent instance ID you obtained after successfully calling `join` to start a conversational AI agent.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Agents.GetTurns(Appid, AgentID) -> *Agora.GetTurnsAgentsResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Query conversation turn information for a conversational AI agent session.

After a conversation with the agent ends, use this endpoint to query the conversation turn information, including the start information, end information, and performance metrics of each conversation turn.

You can query sessions within the last 7 days.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.GetTurnsAgentsRequest{
        Appid: "appid",
        AgentID: "agentId",
    }
client.Agents.GetTurns(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**appid:** `string` — The App ID of the project.
    
</dd>
</dl>

<dl>
<dd>

**agentID:** `string` — The agent instance ID you obtained after successfully calling `join` to start a conversational AI agent.
    
</dd>
</dl>

<dl>
<dd>

**pageIndex:** `*int` — The page number. Starts from 1.
    
</dd>
</dl>

<dl>
<dd>

**pageSize:** `*int` — The number of dialogue turns returned per page.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Agents.Stop(Appid, AgentID) -> error</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Stop the specified conversational agent instance. The API responds after request parameters are validated, and the stop operation is processed asynchronously after the response is returned.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.StopAgentsRequest{
        Appid: "appid",
        AgentID: "agentId",
    }
client.Agents.Stop(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**appid:** `string` — The App ID of the project.
    
</dd>
</dl>

<dl>
<dd>

**agentID:** `string` — The agent instance ID you obtained after successfully calling `join` to start a conversational AI agent.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Agents.Update(Appid, AgentID, request) -> *Agora.UpdateAgentsResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Adjust Conversation AI Engine parameters at runtime.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.UpdateAgentsRequest{
        Appid: "appid",
        AgentID: "agentId",
        Properties: &Agora.UpdateAgentsRequestProperties{
            Token: Agora.String(
                "007eJxTYxxxxxxxxxxIaHMLAAAA0ex66",
            ),
            Llm: &Agora.UpdateAgentsRequestPropertiesLlm{
                SystemMessages: []map[string]any{
                    map[string]any{
                        "role": "system",
                        "content": "You are a helpful assistant. xxx",
                    },
                    map[string]any{
                        "role": "system",
                        "content": "Previously, user has talked about their favorite hobbies with some key topics: xxx",
                    },
                },
                Params: map[string]any{
                    "model": "abab6.5s-chat",
                    "max_token": 1024,
                },
            },
        },
    }
client.Agents.Update(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**appid:** `string` — The App ID of the project.
    
</dd>
</dl>

<dl>
<dd>

**agentID:** `string` — The agent instance ID you obtained after successfully calling `join` to start a conversational AI agent.
    
</dd>
</dl>

<dl>
<dd>

**properties:** `*Agora.UpdateAgentsRequestProperties` — Configuration properties to update.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Agents.Speak(Appid, AgentID, request) -> *Agora.SpeakAgentsResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Broadcast a custom message using the TTS module.

During a conversation with an agent, call this endpoint to immediately broadcast a custom message using the TTS module. Upon receiving the request, the system interrupts the agent's speech and thought process to deliver the message. This broadcast can be interrupted by human voice.

Note: The speak API is not supported when using `mllm` configuration.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.SpeakAgentsRequest{
        Appid: "appid",
        AgentID: "agentId",
        Text: "Sorry, the conversation content is not compliant.",
        Priority: Agora.SpeakAgentsRequestPriorityInterrupt.Ptr(),
        Interruptable: Agora.Bool(
            false,
        ),
    }
client.Agents.Speak(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**appid:** `string` — The App ID of the project.
    
</dd>
</dl>

<dl>
<dd>

**agentID:** `string` — The agent instance ID you obtained after successfully calling `join` to start a conversational AI agent.
    
</dd>
</dl>

<dl>
<dd>

**text:** `string` — The broadcast message text. The maximum length of the text content is 512 bytes.
    
</dd>
</dl>

<dl>
<dd>

**priority:** `*Agora.SpeakAgentsRequestPriority` 

Sets the priority of the message broadcast:
- `INTERRUPT`: High priority. The agent immediately interrupts the current interaction to announce the message.
- `APPEND`: Medium priority. The agent announces the message after the current interaction ends.
- `IGNORE`: Low priority. If the agent is busy interacting, it ignores and discards the broadcast; the message is only announced if the agent is not interacting.
    
</dd>
</dl>

<dl>
<dd>

**interruptable:** `*bool` 

Whether to allow users to interrupt the agent's broadcast by speaking:
- `true`: Allow
- `false`: Don't allow
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Agents.Interrupt(Appid, AgentID, request) -> *Agora.InterruptAgentsResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Interrupt the specified agent while speaking or thinking.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.InterruptAgentsRequest{
        Appid: "appid",
        AgentID: "agentId",
    }
client.Agents.Interrupt(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**appid:** `string` — The App ID of the project.
    
</dd>
</dl>

<dl>
<dd>

**agentID:** `string` — The agent instance ID you obtained after successfully calling `join` to start a conversational AI agent.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Agent Management
<details><summary><code>client.AgentManagement.AgentThink(Appid, AgentID, request) -> *Agora.AgentThinkAgentManagementResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Send a custom text instruction to the specified conversational AI agent instance.

The instruction is injected into the current conversation pipeline as user input, and the agent processes and responds to it following the standard user input logic.

Use this endpoint for the following scenarios:
- **Implicit instruction injection**: Inject hidden context or directives into the conversation.
- **Client-side event triggering**: Notify the agent of client-side events, such as a user clicking a button.
- **Voice and text collaboration**: Combine text instructions with voice input for richer interaction.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.AgentThinkAgentManagementRequest{
        Appid: "appid",
        AgentID: "agentId",
        Text: "The user just clicked the purchase button.",
        OnListeningAction: Agora.AgentThinkAgentManagementRequestOnListeningActionInject.Ptr(),
        OnThinkingAction: Agora.AgentThinkAgentManagementRequestOnThinkingActionInterrupt.Ptr(),
        OnSpeakingAction: Agora.AgentThinkAgentManagementRequestOnSpeakingActionIgnore.Ptr(),
        Interruptable: Agora.Bool(
            true,
        ),
        Metadata: map[string]string{
            "publisher": "user123",
            "model": "deepseek-r1",
        },
    }
client.AgentManagement.AgentThink(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**appid:** `string` — The App ID of the project.
    
</dd>
</dl>

<dl>
<dd>

**agentID:** `string` — The agent instance ID you obtained after successfully calling `join` to start a conversational AI agent.
    
</dd>
</dl>

<dl>
<dd>

**text:** `string` — The custom instruction text to inject into the current conversation pipeline. The system processes this as user input.
    
</dd>
</dl>

<dl>
<dd>

**onListeningAction:** `*Agora.AgentThinkAgentManagementRequestOnListeningAction` 

The action to take when the agent is in a listening state:
- `inject`: Inject the custom text instruction into the current turn without interrupting it.
- `interrupt`: Immediately interrupt the current flow and initiate a new round of dialogue.
- `ignore`: Ignore the request.
    
</dd>
</dl>

<dl>
<dd>

**onThinkingAction:** `*Agora.AgentThinkAgentManagementRequestOnThinkingAction` 

The action to take when the agent is in a thinking state:
- `interrupt`: Interrupt the current state and start a new conversation turn.
- `ignore`: Ignore the request.
    
</dd>
</dl>

<dl>
<dd>

**onSpeakingAction:** `*Agora.AgentThinkAgentManagementRequestOnSpeakingAction` 

The action to take when the agent is in a speaking state:
- `interrupt`: Interrupt the current state and start a new conversation turn.
- `ignore`: Ignore the request.
    
</dd>
</dl>

<dl>
<dd>

**interruptable:** `*bool` 

Whether user speech can interrupt the injected instruction:
- `true`: User speech can interrupt the instruction.
- `false`: User speech cannot interrupt the instruction.
    
</dd>
</dl>

<dl>
<dd>

**metadata:** `map[string]string` — Custom metadata in key-value pair format. Use this field to pass additional business information such as identifiers or model references.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Telephony
<details><summary><code>client.Telephony.List(Appid) -> *Agora.ListTelephonyResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Query historical call records for a specified appid based on the filter criteria.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.ListTelephonyRequest{
        Appid: "appid",
    }
client.Telephony.List(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**appid:** `string` — The App ID of the project.
    
</dd>
</dl>

<dl>
<dd>

**number:** `*string` — Filter by phone number. Can be either the calling number or the called number.
    
</dd>
</dl>

<dl>
<dd>

**fromTime:** `*int` — Query list start timestamp (in seconds). Default is 60 days ago.
    
</dd>
</dl>

<dl>
<dd>

**toTime:** `*int` — Query list end timestamp (in seconds). Default is current time.
    
</dd>
</dl>

<dl>
<dd>

**type_:** `*Agora.ListTelephonyRequestType` 

Call type filter:
- `inbound`: Inbound call.
- `outbound`: Outbound call.

If not specified, all call types are returned.
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — Maximum number of items returned in a single page.
    
</dd>
</dl>

<dl>
<dd>

**cursor:** `*string` — Pagination cursor. Use the `agent_id` from the previous page as the cursor for the next page.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Telephony.Call(Appid, request) -> *Agora.CallTelephonyResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Initiate an outbound call to a specified number and create an agent to join the specified RTC channel.

Use this endpoint to initiate an outbound call to the specified number and create an agent that joins the target RTC channel. The agent waits for the callee to answer.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.CallTelephonyRequest{
        Appid: "appid",
        Name: "customer_service",
        Sip: &Agora.CallTelephonyRequestSip{
            ToNumber: "+19876543210",
            FromNumber: "+11234567890",
            RtcUID: "100",
            RtcToken: "<agora_sip_rtc_token>",
        },
        PipelineID: Agora.String(
            "fzufjlweixxxxnlp",
        ),
        Properties: &Agora.CallTelephonyRequestProperties{
            Channel: "<agora_channel>",
            Token: "<agora_channel_token>",
            AgentRtcUID: "111",
            RemoteRtcUIDs: []string{
                "100",
            },
        },
    }
client.Telephony.Call(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**appid:** `string` — The App ID of the project.
    
</dd>
</dl>

<dl>
<dd>

**name:** `string` — The name identifier of the call session.
    
</dd>
</dl>

<dl>
<dd>

**sip:** `*Agora.CallTelephonyRequestSip` — SIP (Session Initiation Protocol) call configuration object.
    
</dd>
</dl>

<dl>
<dd>

**pipelineID:** `*string` — The unique ID of a published project in AI Studio.
    
</dd>
</dl>

<dl>
<dd>

**properties:** `*Agora.CallTelephonyRequestProperties` 

Call attribute configuration. The content of this field varies depending on the invocation method:
- **Using pipeline ID**: Simply pass in `channel`, `token`, `agent_rtc_uid`, and `remote_rtc_uids`.
- **Using complete configuration**: Pass in the complete parameters of the [Start a conversational AI agent](https://docs.agora.io/en/conversational-ai/rest-api/agent/join) `properties`, including all required fields such as `channel`, `token`, `agent_rtc_uid`, `remote_rtc_uids`, `tts`, and `llm`.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Telephony.Get(Appid, AgentID) -> *Agora.GetTelephonyResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Retrieve the call status and related information of a specified agent.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.GetTelephonyRequest{
        Appid: "appid",
        AgentID: "agent_id",
    }
client.Telephony.Get(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**appid:** `string` — The App ID of the project.
    
</dd>
</dl>

<dl>
<dd>

**agentID:** `string` — The agent ID you obtained after successfully calling the API to initiate an outbound call.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Telephony.Hangup(Appid, AgentID, request) -> *Agora.HangupTelephonyResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Instruct the agent to proactively hang up the ongoing call and leave the RTC channel.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.HangupTelephonyRequest{
        Appid: "appid",
        AgentID: "agent_id",
    }
client.Telephony.Hangup(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**appid:** `string` — The App ID of the project.
    
</dd>
</dl>

<dl>
<dd>

**agentID:** `string` — The agent ID you obtained after successfully calling the API to initiate an outbound call.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## PhoneNumbers
<details><summary><code>client.PhoneNumbers.List() -> []*Agora.ListPhoneNumbersResponseItem</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Retrieve a list of all imported phone numbers under the current account.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.PhoneNumbers.List(
        context.TODO(),
    )
}
```
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.PhoneNumbers.Add(request) -> *Agora.AddPhoneNumbersResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Import a pre-configured phone number that can be used for inbound or outbound calls.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.AddPhoneNumbersRequest{
        Provider: Agora.AddPhoneNumbersRequestProviderByo,
        PhoneNumber: "+19876543210",
        Label: "Sales Hotline",
        Inbound: Agora.Bool(
            true,
        ),
        Outbound: Agora.Bool(
            true,
        ),
        InboundConfig: &Agora.AddPhoneNumbersRequestInboundConfig{
            AllowedAddresses: []string{
                "112.126.15.64/27",
            },
        },
        OutboundConfig: &Agora.AddPhoneNumbersRequestOutboundConfig{
            Address: Agora.String(
                "xxx:xxx@sip.example.com",
            ),
            Transport: Agora.String(
                "tls",
            ),
        },
    }
client.PhoneNumbers.Add(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**provider:** `*Agora.AddPhoneNumbersRequestProvider` 

Number provider:
- `byo`: BYO (Bring Your Own)
- `twilio`: Twilio
    
</dd>
</dl>

<dl>
<dd>

**phoneNumber:** `string` — Telephone number in E.164 format.
    
</dd>
</dl>

<dl>
<dd>

**label:** `string` — A label used to identify the number.
    
</dd>
</dl>

<dl>
<dd>

**inbound:** `*bool` — Whether the number supports inbound calls.
    
</dd>
</dl>

<dl>
<dd>

**outbound:** `*bool` — Whether the number supports outbound calls.
    
</dd>
</dl>

<dl>
<dd>

**inboundConfig:** `*Agora.AddPhoneNumbersRequestInboundConfig` — SIP inbound call configuration.
    
</dd>
</dl>

<dl>
<dd>

**outboundConfig:** `*Agora.AddPhoneNumbersRequestOutboundConfig` — SIP outbound call configuration.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.PhoneNumbers.Get(PhoneNumber) -> *Agora.GetPhoneNumbersResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Retrieve detailed information for a specific phone number.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.GetPhoneNumbersRequest{
        PhoneNumber: "phone_number",
    }
client.PhoneNumbers.Get(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**phoneNumber:** `string` — Telephone number in E.164 format. For example, +11234567890.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.PhoneNumbers.Delete(PhoneNumber) -> error</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Remove an imported phone number from the system.

After calling this endpoint, the number stops receiving calls routed through this system. To delete the number from the service provider, remove it in the service provider's console.
> This operation only removes the number configuration from the Agora system; the number stored with the phone service provider is not deleted.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.DeletePhoneNumbersRequest{
        PhoneNumber: "phone_number",
    }
client.PhoneNumbers.Delete(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**phoneNumber:** `string` — Telephone number in E.164 format. For example, +11234567890.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.PhoneNumbers.Update(PhoneNumber, request) -> *Agora.UpdatePhoneNumbersResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Update the configuration for a phone number.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &Agora.UpdatePhoneNumbersRequest{
        PhoneNumber: "phone_number",
        InboundConfig: &Agora.UpdatePhoneNumbersRequestInboundConfig{
            PipelineID: Agora.String(
                "xxxxx",
            ),
        },
        OutboundConfig: &Agora.UpdatePhoneNumbersRequestOutboundConfig{
            PipelineID: Agora.String(
                "xxxxx",
            ),
        },
    }
client.PhoneNumbers.Update(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**phoneNumber:** `string` — Telephone number in E.164 format. For example, +11234567890.
    
</dd>
</dl>

<dl>
<dd>

**inboundConfig:** `*Agora.UpdatePhoneNumbersRequestInboundConfig` — Update inbound call configuration. Passing `null` will clear the configuration.
    
</dd>
</dl>

<dl>
<dd>

**outboundConfig:** `*Agora.UpdatePhoneNumbersRequestOutboundConfig` — Update outbound call configuration. Passing `null` will clear the configuration.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>
