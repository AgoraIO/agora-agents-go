package vendors

import "strings"

// Gemini 3.8 MLLM model names.
//
// The models/ prefix is part of the requested model ID.
const (
	GeminiLiveModel38Live                 = "models/gemini-3.8-live"
	GeminiLiveModel38LiveExtendedThinking = "models/gemini-3.8-live-extended-thinking"
)

// The low-latency Gemini voice model is the default.
const GeminiLiveDefaultModel = GeminiLiveModel38Live

// Gemini extended-thinking reasoning budgets.
const (
	GeminiThinkingLevelLow    = "low"
	GeminiThinkingLevelMedium = "medium"
	GeminiThinkingLevelHigh   = "high"
)

// GeminiLiveURL is the Gemini Developer API host for Gemini Live models.
const GeminiLiveURL = "https://generativelanguage.googleapis.com"

// GeminiLivePreviewURL is retained for source compatibility with preview callers.
const GeminiLivePreviewURL = GeminiLiveURL

// buildGemini38Config assembles the 3.8 production wire envelope.
func buildGemini38Config(o GeminiLiveOptions) map[string]interface{} {
	model := strings.TrimSpace(o.Model)
	if model == "" {
		model = GeminiLiveDefaultModel
	}
	voice := o.Voice
	if voice == "" {
		voice = "Puck"
	}
	url := o.URL
	if url == "" {
		url = GeminiLiveURL
	}

	params := map[string]interface{}{}
	for k, v := range o.AdditionalParams {
		params[k] = v
	}
	delete(params, "api_key")
	params["model"] = model
	params["voice"] = voice
	if model == GeminiLiveModel38LiveExtendedThinking {
		if o.ThinkingLevel != "" {
			params["thinking_level"] = o.ThinkingLevel
		}
	} else {
		delete(params, "thinking_level")
	}
	// Plural array, and omitted when nil. The singular params.language belongs
	// to xAI Grok.
	if o.LanguageCodes != nil {
		params["language_codes"] = o.LanguageCodes
	}

	if o.Instructions != "" {
		params["instructions"] = o.Instructions
	}
	if o.TranscribeAgent != nil {
		params["transcribe_agent"] = *o.TranscribeAgent
	}
	if o.TranscribeUser != nil {
		params["transcribe_user"] = *o.TranscribeUser
	}
	if o.AffectiveDialog != nil {
		params["affective_dialog"] = *o.AffectiveDialog
	}
	if o.ProactiveAudio != nil {
		params["proactive_audio"] = *o.ProactiveAudio
	}
	if o.HttpOptions != nil {
		params["http_options"] = o.HttpOptions
	}

	config := map[string]interface{}{
		"vendor":  "gemini",
		"api_key": o.APIKey,
		"url":     url,
		"params":  params,
	}
	if o.Messages != nil {
		config["messages"] = o.Messages
	}
	if o.GreetingMessage != "" {
		config["greeting_message"] = o.GreetingMessage
	}
	if o.FailureMessage != "" {
		config["failure_message"] = o.FailureMessage
	}
	if o.InputModalities != nil {
		config["input_modalities"] = o.InputModalities
	}
	if o.OutputModalities != nil {
		config["output_modalities"] = o.OutputModalities
	}
	addMllmTools(config, o.Tools, o.McpServers)
	if o.TurnDetection != nil {
		config["turn_detection"] = o.TurnDetection
	}

	return config
}
