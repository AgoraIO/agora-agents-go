package vendors

import (
	"strings"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
)

// Preview-only vendor types.
//
// Wire shapes here match the preview gateway contract exactly and are not
// served by the production environment. See agentkit/preview_client.go for the
// routing and the agora-feature gate.

// Preview MLLM model names.
//
// Keep the working codenames as defaults until the new Google IDs launch.
// The models/ prefix on the 3.8 names is part of the requested model ID.
const (
	GeminiMLLMModelWalkieTalkie           = "walkie-talkie"
	GeminiMLLMModelCleverChatter          = "clever-chatter"
	GeminiMLLMModel38Live                 = "models/gemini-3.8-live"
	GeminiMLLMModel38LiveExtendedThinking = "models/gemini-3.8-live-extended-thinking"
)

// The working codename the MLLM sends by default.
//
// Keep the codenames while they work. Set Model to one of the 3.8 IDs when
// Google activates it; no SDK release or change to the preview gate is needed.
const GeminiMLLMDefaultModel = GeminiMLLMModelWalkieTalkie

// Gemini clever-chatter reasoning budgets.
const (
	GeminiThinkingLevelLow    = "low"
	GeminiThinkingLevelMedium = "medium"
	GeminiThinkingLevelHigh   = "high"
)

// GeminiMLLMURL is the Gemini Developer API host the MLLM uses.
const GeminiMLLMURL = "https://generativelanguage.googleapis.com"

// GeminiMLLMOptions configures either Gemini voice model.
//
// The API key is sent as mllm.api_key. The preview URL defaults to the
// Gemini Developer API host rather than a WebSocket endpoint.
type GeminiMLLMOptions struct {
	APIKey string
	// Model defaults to walkie-talkie until the 3.8 IDs launch.
	Model string
	// ThinkingLevel is sent only for clever-chatter and the extended-thinking model.
	ThinkingLevel string
	// Voice defaults to "Puck".
	Voice string
	// LanguageCodes are the languages the model should work in, sent as
	// params.language_codes. The preview MLLMs take the plural array, not the
	// singular "language" (which belongs to xAI Grok). Omitted when nil.
	LanguageCodes   []string
	URL             string
	Instructions    string
	TranscribeAgent *bool
	TranscribeUser  *bool
	AffectiveDialog *bool
	ProactiveAudio  *bool
	HttpOptions     map[string]interface{}
	// Greeting is serialized to "greeting". The preview Gemini models read this
	// spelling, not the "greeting_message" the production providers use.
	Greeting         string
	FailureMessage   string
	InputModalities  []string
	OutputModalities []string
	Messages         []map[string]interface{}
	AdditionalParams map[string]interface{}
	TurnDetection    *Agora.MllmTurnDetection
}

// buildGeminiMLLMConfig assembles the Gemini preview envelope.
func buildGeminiMLLMConfig(o GeminiMLLMOptions) map[string]interface{} {
	model := strings.TrimSpace(o.Model)
	if model == "" {
		model = GeminiMLLMDefaultModel
	}
	voice := o.Voice
	if voice == "" {
		voice = "Puck"
	}
	url := o.URL
	if url == "" {
		url = GeminiMLLMURL
	}

	params := map[string]interface{}{}
	for k, v := range o.AdditionalParams {
		params[k] = v
	}
	delete(params, "api_key")
	params["model"] = model
	params["voice"] = voice
	if model == GeminiMLLMModelCleverChatter || model == GeminiMLLMModel38LiveExtendedThinking {
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
	if o.Greeting != "" {
		config["greeting"] = o.Greeting
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

// GeminiMLLM configures both Gemini preview voice models in one type.
type GeminiMLLM struct {
	options GeminiMLLMOptions
}

// NewGeminiMLLM creates a Gemini preview MLLM with an optional model ID.
func NewGeminiMLLM(opts GeminiMLLMOptions) *GeminiMLLM {
	opts.APIKey = strings.TrimSpace(opts.APIKey)
	if opts.APIKey == "" {
		panic("GeminiMLLM requires APIKey")
	}
	switch opts.ThinkingLevel {
	case "", GeminiThinkingLevelLow, GeminiThinkingLevelMedium, GeminiThinkingLevelHigh:
	default:
		panic("GeminiMLLM ThinkingLevel must be low, medium, or high")
	}
	return &GeminiMLLM{options: opts}
}

// ToConfig serializes to the mllm field of the start request.
func (g *GeminiMLLM) ToConfig() map[string]interface{} {
	return buildGeminiMLLMConfig(g.options)
}
