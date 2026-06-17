package vendors

import Agora "github.com/AgoraIO/agora-agents-go/v2"

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
	return &OpenAIRealtime{options: opts}
}

func (o *OpenAIRealtime) ToConfig() map[string]interface{} {
	// Match TS: `model` is the base; explicit Params entries override it.
	var params map[string]interface{}
	if o.options.Model != "" || o.options.Params != nil {
		params = map[string]interface{}{}
		params["model"] = o.options.Model
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
	}
	if o.options.URL != "" {
		config["url"] = o.options.URL
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
		"params":  params,
	}
	if g.options.URL != "" {
		config["url"] = g.options.URL
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

	config := map[string]interface{}{
		"vendor":                 "vertexai",
		"project_id":             v.options.ProjectID,
		"location":               v.options.Location,
		"adc_credentials_string": v.options.ADCredentialsString,
		"params":                 params,
	}

	if v.options.URL != "" {
		config["url"] = v.options.URL
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
