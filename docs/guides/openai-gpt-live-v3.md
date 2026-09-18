# GPT Live v3

GPT Live targets `gpt-live-1` on `/v1/live/sessions` and runs through the production Conversational AI gateway.

<!-- snippet: fragment -->
```go
toolsEnabled := true
live := vendors.NewOpenAIGPTLive(vendors.OpenAIGPTLiveOptions{
    APIKey: openaiKey,
    Voice: "cedar",
    ToolEnabled: &toolsEnabled,
    Prompt: "You are a helpful assistant. Keep responses concise.",
    GreetingMessage: "Hello! I'm GPT Live. How can I help you today?",
    Tools: []*Agora.LlmTool{{
        Function: &Agora.LlmToolFunction{Name: "lookup"},
        Server: &Agora.LlmToolServer{
            Method: Agora.LlmToolServerMethodPost,
            URL: "https://tools.example/lookup",
        },
    }},
    McpServerConfigs: []*Agora.McpServer{{
        Name: "catalog",
        Endpoint: "https://tools.example/mcp",
    }},
})
// Pass live to your agent's WithMllm method.
// Enable advanced_features.enable_tools with WithTools(true).
```

Set REST tools and typed MCP servers on `OpenAIGPTLiveOptions`; they serialize as `mllm.tools` and `mllm.mcp_servers`. The deprecated map-based `McpServers` option remains supported for compatibility, while `McpServerConfigs` takes precedence when both are set. Enable `advanced_features.enable_tools` with the existing tools builder. To advertise graph tools to GPT Live, also set `ToolEnabled` to true.

## Request placement

GPT Live places MCP at `properties.mllm.mcp_servers`, the tool gate at `properties.advanced_features.enable_tools`, and silence settings at `properties.parameters.silence_config`. MCP and MAIN settings are outside `properties.mllm.params`.

```json
{
  "properties": {
    "mllm": {
      "enable": true,
      "vendor": "openai_gpt_live",
      "api_key": "<openai-api-key>",
      "url": "wss://api.openai.com/v1/live/sessions",
      "greeting_message": "Hello! I'm GPT Live. How can I help you today?",
      "params": {
        "model": "gpt-live-1",
        "voice": "cedar",
        "prompt": "You are a helpful assistant.",
        "tool_enabled": true
      },
      "tools": [{
        "function": {"name": "lookup"},
        "server": {
          "method": "POST",
          "url": "https://tools.example/lookup"
        }
      }],
      "mcp_servers": [{
        "name": "catalog",
        "endpoint": "https://tools.example/mcp",
        "transport": "streamable_http"
      }]
    },
    "advanced_features": {"enable_tools": true}
  }
}
```

This fragment omits the normal name, channel, token and UID fields populated by the SDK session. GPT Live sessions use the same production regional routing as other GA vendors and do not send an `agora-feature` gate.

The SDK omits `AlphaSelector` by default. Set it only when a future preview contract requires an `OpenAI-Alpha` selector.

## Silence and backend rollout

Keep silence settings in the existing agent parameters builder, never in vendor params. The public API spelling is `silence_config`, with `{timeout_ms, action, content}`. The supplied extension contract describes internal `parameters.main.silence` and supports `action: "think"`; the public documentation currently says `silence_config` does not apply to MLLM. Serialization is covered by tests, but the public documentation does not establish that the production backend maps it to GPT Live's internal MAIN setting. Confirm that backend mapping before relying on silence nudges. The SDK does not invent a new public `main` field.

## Advanced options

The supplied backend contract marks custom voice objects, `responses_params`, and first-class context management as pending PR #1522. The SDK does not expose typed options for those fields. Use raw params only once your target backend supports that PR. Until then, context management is reachable via `session_params.context_management`.

`session_params` cannot override `model`, `delegation`, `audio`, `instructions`, or `input`; put modelled values in their dedicated options instead. Unknown session fields may still be rejected by the provider API. Delegation and voice are fixed for a session: apply changes when creating a new session. See the [vendor reference](../reference/vendors.md) for every option and provider default.
