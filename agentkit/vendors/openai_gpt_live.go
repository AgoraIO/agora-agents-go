package vendors

import (
	"encoding/json"
	"log"
	"net/url"
	"strings"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
)

// OpenAIGPTLiveOptions configures GPT Live v3 alpha. Not for production traffic.
// Explicit options override Params. Unset tuning options retain provider defaults.
type OpenAIGPTLiveOptions struct {
	APIKey string
	URL    string
	// Deprecated: Use Prompt. Serialized as prompt; an explicit Prompt wins.
	Instructions     string
	GreetingMessage  string
	FailureMessage   string
	InputModalities  []string
	OutputModalities []string
	Messages         []map[string]interface{}
	// MCP servers exposed to GPT Live. Requires agentkit.WithTools(true).
	McpServers []map[string]interface{}
	Params     map[string]interface{}
	// Deprecated: Unsupported in v3; setting this panics during ToConfig.
	InputAudioTranscription map[string]interface{}
	// Deprecated: Ignored with a warning; v3 performs endpointing internally.
	TurnDetection *Agora.MllmTurnDetection
	// Defaults to gpt-live-1-diamond-alpha.
	Model string
	// Output voice; provider default marin. Custom voice objects require PR #1522; use params after rollout.
	Voice string
	// Session instructions.
	Prompt string
	// Host when url is omitted; default wss://api.openai.com.
	BaseURL string
	// WebSocket path; default /v1/live/sessions.
	Path string
	// OpenAI-Alpha selector. Leave empty to use the required GPT Live v3 contract.
	AlphaSelector string
	// Extra provider request headers as a JSON string; protocol headers win.
	Headers string
	// Assistant silence boundary in ms; provider default 600. Zero disables inference.
	OutputIdleEndMs *int
	// Caller silence boundary in ms; provider default 1500.
	InputIdleEndMs *int
	// Speech amplitude threshold on the 16-bit scale; provider default 50.
	OutputSilencePeak *int
	// Graph PCM sample rate; provider default 24000.
	OutputSampleRate *int
	// Initial audio cushion; provider default 0. Negative disables pacing.
	OutputBufferMs *int
	// Mic append batching in ms. Join default 0; extension class default 100.
	InputBatchMs *int
	// Advertise graph tools; provider default false. Does not control delegate built-ins.
	ToolEnabled *bool
	// Tool delegation mode; provider default responses. Fixed for the session.
	Delegation string
	// Tool delegate model; provider default gpt-5.6-sol.
	ResponsesModel string
	// Interrupt playback on caller speech; provider default false.
	InterruptOnUserTurn *bool
	// Unmodelled v3 session fields. Cannot override model, delegation, audio, instructions or input.
	SessionParams map[string]interface{}
}

// OpenAIGPTLive is the preview GPT Live v3 MLLM vendor.
type OpenAIGPTLive struct{ options OpenAIGPTLiveOptions }

func NewOpenAIGPTLive(opts OpenAIGPTLiveOptions) *OpenAIGPTLive {
	if opts.APIKey == "" {
		panic("OpenAIGPTLive requires APIKey")
	}
	return &OpenAIGPTLive{options: opts}
}

func (o *OpenAIGPTLive) ToConfig() map[string]interface{} {
	opts := o.options
	params := map[string]interface{}{
		"model":          "gpt-live-1-diamond-alpha",
		"alpha_selector": "quicksilver=v3",
	}
	for k, v := range opts.Params {
		params[k] = v
	}
	if opts.Instructions != "" {
		params["prompt"] = opts.Instructions
	}
	if opts.Model != "" {
		params["model"] = opts.Model
	}
	if opts.Voice != "" {
		params["voice"] = opts.Voice
	}
	if opts.Prompt != "" {
		params["prompt"] = opts.Prompt
	}
	if opts.BaseURL != "" {
		params["base_url"] = opts.BaseURL
	}
	if opts.Path != "" {
		params["path"] = opts.Path
	}
	if opts.AlphaSelector != "" {
		params["alpha_selector"] = opts.AlphaSelector
	}
	if opts.Headers != "" {
		params["headers"] = opts.Headers
	}
	if opts.OutputIdleEndMs != nil {
		params["output_idle_end_ms"] = *opts.OutputIdleEndMs
	}
	if opts.InputIdleEndMs != nil {
		params["input_idle_end_ms"] = *opts.InputIdleEndMs
	}
	if opts.OutputSilencePeak != nil {
		params["output_silence_peak"] = *opts.OutputSilencePeak
	}
	if opts.OutputSampleRate != nil {
		params["output_sample_rate"] = *opts.OutputSampleRate
	}
	if opts.OutputBufferMs != nil {
		params["output_buffer_ms"] = *opts.OutputBufferMs
	}
	if opts.InputBatchMs != nil {
		params["input_batch_ms"] = *opts.InputBatchMs
	}
	if opts.ToolEnabled != nil {
		params["tool_enabled"] = *opts.ToolEnabled
	}
	if opts.Delegation != "" {
		params["delegation"] = opts.Delegation
	}
	if opts.ResponsesModel != "" {
		params["responses_model"] = opts.ResponsesModel
	}
	if opts.InterruptOnUserTurn != nil {
		params["interrupt_on_user_turn"] = *opts.InterruptOnUserTurn
	}
	if opts.SessionParams != nil {
		params["session_params"] = opts.SessionParams
	}
	if _, ok := params["input_audio_transcription"]; ok || opts.InputAudioTranscription != nil {
		panic("GPT Live v3 does not support input_audio_transcription")
	}
	if _, ok := params["turn_detection"]; ok || opts.TurnDetection != nil {
		log.Print("GPT Live v3 ignores turn_detection; endpointing is internal")
		delete(params, "turn_detection")
	}
	if mode, ok := params["delegation"]; ok && mode != "client" && mode != "responses" {
		panic("GPT Live delegation must be client or responses")
	}
	if raw, ok := params["headers"]; ok {
		value, ok := raw.(string)
		var headers map[string]interface{}
		if !ok || json.Unmarshal([]byte(value), &headers) != nil || headers == nil {
			panic("GPT Live headers must be a JSON object string")
		}
	}
	if raw, exists := params["session_params"]; exists {
		session, ok := raw.(map[string]interface{})
		if !ok {
			panic("GPT Live session_params must be an object")
		}
		for _, key := range []string{"model", "delegation", "audio", "instructions", "input"} {
			if _, ok := session[key]; ok {
				panic("GPT Live session_params cannot override " + key)
			}
		}
	}
	endpoint := opts.URL
	if endpoint == "" {
		base, path := "wss://api.openai.com", "/v1/live/sessions"
		if v, ok := params["base_url"].(string); ok {
			base = v
		}
		if v, ok := params["path"].(string); ok {
			path = v
		}
		endpoint = strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || (parsed.Scheme != "ws" && parsed.Scheme != "wss") || parsed.Hostname() == "" {
		panic("GPT Live url must be a full ws:// or wss:// endpoint")
	}
	if parsed.Hostname() == "api.openai.com" && parsed.Path == "/v1/live" {
		parsed.Path = "/v1/live/sessions"
		endpoint = parsed.String()
	}
	config := map[string]interface{}{"vendor": "openai_gpt_live", "api_key": opts.APIKey, "url": endpoint, "params": params}
	if opts.GreetingMessage != "" {
		config["greeting_message"] = opts.GreetingMessage
	}
	if opts.FailureMessage != "" {
		config["failure_message"] = opts.FailureMessage
	}
	if opts.InputModalities != nil {
		config["input_modalities"] = opts.InputModalities
	}
	if opts.OutputModalities != nil {
		config["output_modalities"] = opts.OutputModalities
	}
	if opts.Messages != nil {
		config["messages"] = opts.Messages
	}
	if opts.McpServers != nil {
		config["mcp_servers"] = ensureMcpTransport(opts.McpServers)
	}
	return config
}
