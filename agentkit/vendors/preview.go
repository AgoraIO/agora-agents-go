package vendors

import "strings"

// Preview-only vendor types.
//
// Wire shapes here match the preview gateway contract exactly and are not
// served by the production environment. See agentkit/preview_client.go for the
// routing and the agora-feature gate.

// Preview MLLM model names.
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

// GeminiLivePreviewURL is the Gemini Developer API host for the 3.8 models.
const GeminiLivePreviewURL = "https://generativelanguage.googleapis.com"

// buildGeminiPreviewConfig assembles the 3.8 wire envelope from GeminiLive.
func buildGeminiPreviewConfig(o GeminiLiveOptions) map[string]interface{} {
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
		url = GeminiLivePreviewURL
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
	// to xAI Grok in the Agora schema, and the production Gemini Live provider
	// sends no language field at all.
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
	// "greeting", not "greeting_message": the preview Gemini models read this
	// spelling. See the preview-endpoint guide.
	if o.GreetingMessage != "" {
		config["greeting"] = o.GreetingMessage
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
	if o.TurnDetection != nil {
		config["turn_detection"] = o.TurnDetection
	}

	return config
}
