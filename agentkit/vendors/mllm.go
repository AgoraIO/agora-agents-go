package vendors

import (
	"encoding/json"
	"log"
	"net/url"
	"strings"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
)

type OpenAIRealtimeOptions struct {
	APIKey                  string
	Model                   string
	Voice                   string
	Instructions            string
	InputAudioTranscription map[string]interface{}
	URL                     string
	GreetingMessage         string
	FailureMessage          string
	InputModalities         []string
	OutputModalities        []string
	Messages                []map[string]interface{}
	Params                  map[string]interface{}
	TurnDetection           *Agora.MllmTurnDetection
}

type OpenAIRealtime struct {
	options OpenAIRealtimeOptions
}

func NewOpenAIRealtime(opts OpenAIRealtimeOptions) *OpenAIRealtime {
	if opts.APIKey == "" {
		panic("OpenAIRealtime requires APIKey")
	}
	if opts.Model == "" {
		opts.Model = "gpt-4o-realtime-preview"
	}
	if opts.URL == "" {
		opts.URL = "wss://api.openai.com/v1/realtime"
	}
	return &OpenAIRealtime{options: opts}
}

func (o *OpenAIRealtime) ToConfig() map[string]interface{} {
	// Match TS: build params when any nested field is set; explicit Params entries override.
	var params map[string]interface{}
	if o.options.Model != "" ||
		o.options.Params != nil ||
		o.options.Voice != "" ||
		o.options.Instructions != "" ||
		o.options.InputAudioTranscription != nil {
		params = map[string]interface{}{}
		if o.options.Model != "" {
			params["model"] = o.options.Model
		}
		for k, v := range o.options.Params {
			params[k] = v
		}
		if o.options.Voice != "" {
			params["voice"] = o.options.Voice
		}
		if o.options.Instructions != "" {
			params["instructions"] = o.options.Instructions
		}
		if o.options.InputAudioTranscription != nil {
			params["input_audio_transcription"] = o.options.InputAudioTranscription
		}
	}

	config := map[string]interface{}{
		"vendor":  "openai",
		"api_key": o.options.APIKey,
		"url":     o.options.URL,
	}
	if params != nil {
		config["params"] = params
	}

	if o.options.GreetingMessage != "" {
		config["greeting_message"] = o.options.GreetingMessage
	}
	if o.options.FailureMessage != "" {
		config["failure_message"] = o.options.FailureMessage
	}
	if o.options.InputModalities != nil {
		config["input_modalities"] = o.options.InputModalities
	}
	if o.options.OutputModalities != nil {
		config["output_modalities"] = o.options.OutputModalities
	}
	if o.options.Messages != nil {
		config["messages"] = o.options.Messages
	}
	if o.options.TurnDetection != nil {
		config["turn_detection"] = o.options.TurnDetection
	}

	return config
}

// OpenAIGPTLiveOptions configures GPT Live v3.
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
	// Tools configures inline REST tools exposed to GPT Live.
	Tools []*Agora.LlmTool
	// Deprecated: Use McpServerConfigs.
	McpServers []map[string]interface{}
	// McpServerConfigs configures typed MCP servers and takes precedence over McpServers.
	McpServerConfigs []*Agora.McpServer
	Params           map[string]interface{}
	// Deprecated: Unsupported in v3; setting this panics during ToConfig.
	InputAudioTranscription map[string]interface{}
	// Deprecated: Ignored with a warning; v3 performs endpointing internally.
	TurnDetection *Agora.MllmTurnDetection
	// Defaults to gpt-live-1.
	Model string
	// Output voice; provider default marin. Custom voice objects require PR #1522; use params after rollout.
	Voice string
	// Session instructions.
	Prompt string
	// Host when url is omitted; default wss://api.openai.com.
	BaseURL string
	// WebSocket path; default /v1/live/sessions.
	Path string
	// Optional OpenAI-Alpha selector for preview contracts. Omitted when empty.
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

// OpenAIGPTLive is the GPT Live v3 MLLM vendor.
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
		"model": "gpt-live-1",
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
	config := map[string]interface{}{
		"vendor":  "openai_gpt_live",
		"api_key": opts.APIKey,
		"url":     endpoint,
		"params":  params,
	}
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
	if opts.Tools != nil {
		config["tools"] = opts.Tools
	}
	if opts.McpServerConfigs != nil {
		config["mcp_servers"] = opts.McpServerConfigs
	} else if opts.McpServers != nil {
		config["mcp_servers"] = ensureMcpTransport(opts.McpServers)
	}
	return config
}

// AzureOpenAIRealtimeOptions configures Azure OpenAI Realtime MLLM.
type AzureOpenAIRealtimeOptions struct {
	APIKey           string
	URL              string
	Messages         []map[string]interface{}
	Instructions     string
	Model            string
	Voice            string
	OutputModalities []string
	MaxHistory       *int
	GreetingMessage  string
	TurnDetection    *Agora.MllmTurnDetection
}

// AzureOpenAIRealtime is the global Azure OpenAI Realtime MLLM vendor.
type AzureOpenAIRealtime struct {
	options AzureOpenAIRealtimeOptions
}

// NewAzureOpenAIRealtime creates an Azure OpenAI Realtime MLLM configuration.
func NewAzureOpenAIRealtime(opts AzureOpenAIRealtimeOptions) *AzureOpenAIRealtime {
	if opts.APIKey == "" {
		panic("AzureOpenAIRealtime requires APIKey")
	}
	if opts.URL == "" {
		panic("AzureOpenAIRealtime requires URL")
	}
	if opts.TurnDetection == nil {
		panic("AzureOpenAIRealtime requires TurnDetection")
	}
	return &AzureOpenAIRealtime{options: opts}
}

// ToConfig returns the Azure OpenAI Realtime configuration expected by the API.
func (a *AzureOpenAIRealtime) ToConfig() map[string]interface{} {
	var params map[string]interface{}
	if a.options.Model != "" || a.options.Voice != "" || a.options.Instructions != "" {
		params = map[string]interface{}{}
		if a.options.Model != "" {
			params["model"] = a.options.Model
		}
		if a.options.Voice != "" {
			params["voice"] = a.options.Voice
		}
		if a.options.Instructions != "" {
			params["instructions"] = a.options.Instructions
		}
	}

	config := map[string]interface{}{
		"vendor":         "azure",
		"api_key":        a.options.APIKey,
		"url":            a.options.URL,
		"turn_detection": a.options.TurnDetection,
	}
	if params != nil {
		config["params"] = params
	}
	if a.options.MaxHistory != nil {
		config["max_history"] = *a.options.MaxHistory
	}
	if a.options.GreetingMessage != "" {
		config["greeting_message"] = a.options.GreetingMessage
	}
	if a.options.OutputModalities != nil {
		config["output_modalities"] = a.options.OutputModalities
	}
	if a.options.Messages != nil {
		config["messages"] = a.options.Messages
	}
	return config
}

// XaiGrokOptions configures the xAI Grok MLLM vendor (mllm.vendor "xai").
// Future xAI ASR/TTS wrappers should be named XaiSTT and XaiTTS, not XaiRealtime.
type XaiGrokOptions struct {
	APIKey           string
	URL              string
	Voice            string
	Language         string
	SampleRate       *int
	GreetingMessage  string
	FailureMessage   string
	InputModalities  []string
	OutputModalities []string
	Messages         []map[string]interface{}
	Params           map[string]interface{}
	TurnDetection    *Agora.MllmTurnDetection
}

// XaiGrok is the xAI Grok MLLM vendor (mllm.vendor "xai").
type XaiGrok struct {
	options XaiGrokOptions
}

// NewXaiGrok creates an xAI Grok MLLM vendor.
func NewXaiGrok(opts XaiGrokOptions) *XaiGrok {
	if opts.APIKey == "" {
		panic("XaiGrok requires APIKey")
	}
	if opts.URL == "" {
		opts.URL = "wss://api.x.ai/v1/realtime"
	}
	return &XaiGrok{options: opts}
}

// XAIGrokOptions is deprecated.
//
// Deprecated: Use XaiGrokOptions instead.
type XAIGrokOptions = XaiGrokOptions

// XAIGrok is deprecated.
//
// Deprecated: Use XaiGrok instead.
type XAIGrok = XaiGrok

// NewXAIGrok is deprecated.
//
// Deprecated: Use NewXaiGrok instead.
func NewXAIGrok(opts XAIGrokOptions) *XAIGrok {
	return NewXaiGrok(opts)
}

func (x *XaiGrok) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range x.options.Params {
		params[k] = v
	}
	if x.options.Voice != "" {
		params["voice"] = x.options.Voice
	}
	if x.options.Language != "" {
		params["language"] = x.options.Language
	}
	if x.options.SampleRate != nil {
		params["sample_rate"] = *x.options.SampleRate
	}

	config := map[string]interface{}{
		"vendor":  "xai",
		"api_key": x.options.APIKey,
		"url":     x.options.URL,
		"params":  params,
	}
	if x.options.GreetingMessage != "" {
		config["greeting_message"] = x.options.GreetingMessage
	}
	if x.options.FailureMessage != "" {
		config["failure_message"] = x.options.FailureMessage
	}
	if x.options.InputModalities != nil {
		config["input_modalities"] = x.options.InputModalities
	}
	if x.options.OutputModalities != nil {
		config["output_modalities"] = x.options.OutputModalities
	}
	if x.options.Messages != nil {
		config["messages"] = x.options.Messages
	}
	if x.options.TurnDetection != nil {
		config["turn_detection"] = x.options.TurnDetection
	}
	return config
}

type GeminiLiveOptions struct {
	APIKey           string
	Model            string
	URL              string
	Instructions     string
	Voice            string
	AffectiveDialog  *bool
	ProactiveAudio   *bool
	TranscribeAgent  *bool
	TranscribeUser   *bool
	HttpOptions      map[string]interface{}
	GreetingMessage  string
	FailureMessage   string
	InputModalities  []string
	OutputModalities []string
	Messages         []map[string]interface{}
	AdditionalParams map[string]interface{}
	TurnDetection    *Agora.MllmTurnDetection
}

type GeminiLive struct {
	options GeminiLiveOptions
}

func NewGeminiLive(opts GeminiLiveOptions) *GeminiLive {
	if opts.APIKey == "" {
		panic("GeminiLive requires APIKey")
	}
	if opts.Model == "" {
		panic("GeminiLive requires Model")
	}
	return &GeminiLive{options: opts}
}

func (g *GeminiLive) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range g.options.AdditionalParams {
		params[k] = v
	}
	params["model"] = g.options.Model
	if g.options.Instructions != "" {
		params["instructions"] = g.options.Instructions
	}
	if g.options.Voice != "" {
		params["voice"] = g.options.Voice
	}
	if g.options.AffectiveDialog != nil {
		params["affective_dialog"] = *g.options.AffectiveDialog
	}
	if g.options.ProactiveAudio != nil {
		params["proactive_audio"] = *g.options.ProactiveAudio
	}
	if g.options.TranscribeAgent != nil {
		params["transcribe_agent"] = *g.options.TranscribeAgent
	}
	if g.options.TranscribeUser != nil {
		params["transcribe_user"] = *g.options.TranscribeUser
	}
	if g.options.HttpOptions != nil {
		params["http_options"] = g.options.HttpOptions
	}

	config := map[string]interface{}{
		"vendor":  "gemini",
		"api_key": g.options.APIKey,
		"url":     g.options.URL,
		"params":  params,
	}
	if g.options.GreetingMessage != "" {
		config["greeting_message"] = g.options.GreetingMessage
	}
	if g.options.FailureMessage != "" {
		config["failure_message"] = g.options.FailureMessage
	}
	if g.options.InputModalities != nil {
		config["input_modalities"] = g.options.InputModalities
	}
	if g.options.OutputModalities != nil {
		config["output_modalities"] = g.options.OutputModalities
	}
	if g.options.Messages != nil {
		config["messages"] = g.options.Messages
	}
	if g.options.TurnDetection != nil {
		config["turn_detection"] = g.options.TurnDetection
	}
	return config
}

type VertexAIOptions struct {
	ProjectID           string
	Location            string
	Model               string
	URL                 string
	Voice               string
	AffectiveDialog     *bool
	ProactiveAudio      *bool
	TranscribeAgent     *bool
	TranscribeUser      *bool
	HttpOptions         map[string]interface{}
	Instructions        string
	Messages            []map[string]interface{}
	ADCredentialsString string
	AdditionalParams    map[string]interface{}
	GreetingMessage     string
	FailureMessage      string
	InputModalities     []string
	OutputModalities    []string
	TurnDetection       *Agora.MllmTurnDetection
}

type VertexAI struct {
	options VertexAIOptions
}

func NewVertexAI(opts VertexAIOptions) *VertexAI {
	if opts.ProjectID == "" {
		panic("VertexAI requires ProjectID")
	}
	if opts.ADCredentialsString == "" {
		panic("VertexAI requires ADCredentialsString")
	}
	if opts.Location == "" {
		opts.Location = "us-central1"
	}
	if opts.Model == "" {
		opts.Model = "gemini-2.0-flash-exp"
	}
	return &VertexAI{options: opts}
}

func (v *VertexAI) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, val := range v.options.AdditionalParams {
		params[k] = val
	}
	params["model"] = v.options.Model
	params["project_id"] = v.options.ProjectID
	params["location"] = v.options.Location
	params["adc_credentials_string"] = v.options.ADCredentialsString
	if v.options.Voice != "" {
		params["voice"] = v.options.Voice
	}
	if v.options.Instructions != "" {
		params["instructions"] = v.options.Instructions
	}
	if v.options.AffectiveDialog != nil {
		params["affective_dialog"] = *v.options.AffectiveDialog
	}
	if v.options.ProactiveAudio != nil {
		params["proactive_audio"] = *v.options.ProactiveAudio
	}
	if v.options.TranscribeAgent != nil {
		params["transcribe_agent"] = *v.options.TranscribeAgent
	}
	if v.options.TranscribeUser != nil {
		params["transcribe_user"] = *v.options.TranscribeUser
	}
	if v.options.HttpOptions != nil {
		params["http_options"] = v.options.HttpOptions
	}
	params["project_id"] = v.options.ProjectID
	params["location"] = v.options.Location
	params["adc_credentials_string"] = v.options.ADCredentialsString

	config := map[string]interface{}{
		"vendor": "vertexai",
		"url":    v.options.URL,
		"params": params,
	}
	if v.options.GreetingMessage != "" {
		config["greeting_message"] = v.options.GreetingMessage
	}
	if v.options.FailureMessage != "" {
		config["failure_message"] = v.options.FailureMessage
	}
	if v.options.InputModalities != nil {
		config["input_modalities"] = v.options.InputModalities
	}
	if v.options.OutputModalities != nil {
		config["output_modalities"] = v.options.OutputModalities
	}
	if v.options.Messages != nil {
		config["messages"] = v.options.Messages
	}
	if v.options.TurnDetection != nil {
		config["turn_detection"] = v.options.TurnDetection
	}

	return config
}
