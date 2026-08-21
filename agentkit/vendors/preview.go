package vendors

import (
	Agora "github.com/AgoraIO/agora-agents-go/v2"
)

// Preview-only vendor types.
//
// Wire shapes here match the preview gateway contract exactly and are not
// served by the production environment. See agentkit/preview_client.go for the
// routing and the agora-feature gate.

// Preview MLLM model names.
//
// Google is renaming these from codenames to conventional model IDs
// (walkie-talkie -> gemini-3.5-live, clever-chatter -> gemini-3.5-live-thinking).
// Both spellings are listed because both are accepted during the rollout; see
// the GeminiPreviewDefaultModel* constants for which one the vendor types send.
const (
	GeminiPreviewModelWalkieTalkie   = "walkie-talkie"
	GeminiPreviewModelCleverChatter  = "clever-chatter"
	GeminiPreviewModel35Live         = "gemini-3.5-live"
	GeminiPreviewModel35LiveThinking = "gemini-3.5-live-thinking"
)

// The model name each preview MLLM type sends by default.
//
// Still the codenames: as of 2026-08-19 the renamed IDs are propagating through
// Google's fleet and are not live everywhere, while the codenames work. When
// Google confirms the rename has landed, point these two at
// GeminiPreviewModel35Live / GeminiPreviewModel35LiveThinking — that is the
// whole cutover. Callers who need the new names sooner can set Model per
// instance.
const (
	GeminiPreviewDefaultModelWalkieTalkie  = GeminiPreviewModelWalkieTalkie
	GeminiPreviewDefaultModelCleverChatter = GeminiPreviewModelCleverChatter
)

// Gemini clever-chatter reasoning budgets.
const (
	GeminiThinkingLevelLow    = "low"
	GeminiThinkingLevelMedium = "medium"
	GeminiThinkingLevelHigh   = "high"
)

// GeminiPreviewMLLMURL is the Gemini Developer API host the preview MLLMs use.
const GeminiPreviewMLLMURL = "https://generativelanguage.googleapis.com"

// GeminiPreviewMLLMOptions is the shared option set for the preview Gemini MLLMs.
//
// Differs from the production GeminiLive in two ways the preview gateway
// requires: params.api_key is repeated inside params, and URL defaults to the
// Gemini Developer API host rather than a WebSocket endpoint.
type GeminiPreviewMLLMOptions struct {
	APIKey string
	// Model overrides the per-type default, so a preview rename does not
	// require an SDK release.
	Model string
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

// buildGeminiPreviewMLLMConfig assembles the shared envelope. modelParams is
// merged after voice so key order matches the other SDKs.
func buildGeminiPreviewMLLMConfig(
	defaultModel string,
	o GeminiPreviewMLLMOptions,
	modelParams map[string]interface{},
) map[string]interface{} {
	model := o.Model
	if model == "" {
		model = defaultModel
	}
	voice := o.Voice
	if voice == "" {
		voice = "Puck"
	}
	url := o.URL
	if url == "" {
		url = GeminiPreviewMLLMURL
	}

	params := map[string]interface{}{}
	for k, v := range o.AdditionalParams {
		params[k] = v
	}
	params["api_key"] = o.APIKey
	params["model"] = model
	params["voice"] = voice
	for k, v := range modelParams {
		params[k] = v
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

// GeminiWalkieTalkie is the Gemini walkie-talkie MLLM (preview).
type GeminiWalkieTalkie struct {
	options GeminiPreviewMLLMOptions
}

// NewGeminiWalkieTalkie creates a walkie-talkie MLLM config.
func NewGeminiWalkieTalkie(opts GeminiPreviewMLLMOptions) *GeminiWalkieTalkie {
	if opts.APIKey == "" {
		panic("GeminiWalkieTalkie requires APIKey")
	}
	return &GeminiWalkieTalkie{options: opts}
}

// ToConfig serializes to the mllm field of the start request.
func (g *GeminiWalkieTalkie) ToConfig() map[string]interface{} {
	return buildGeminiPreviewMLLMConfig(GeminiPreviewDefaultModelWalkieTalkie, g.options, nil)
}

// GeminiCleverChatterOptions is GeminiPreviewMLLMOptions plus ThinkingLevel.
//
// ThinkingLevel is exclusive to clever-chatter — it is a separate type rather
// than a field on the shared options so the distinction is structural.
type GeminiCleverChatterOptions struct {
	GeminiPreviewMLLMOptions

	// ThinkingLevel is one of GeminiThinkingLevelLow, Medium, or High.
	ThinkingLevel string
}

// GeminiCleverChatter is the Gemini clever-chatter MLLM (preview).
type GeminiCleverChatter struct {
	options GeminiCleverChatterOptions
}

// NewGeminiCleverChatter creates a clever-chatter MLLM config.
func NewGeminiCleverChatter(opts GeminiCleverChatterOptions) *GeminiCleverChatter {
	if opts.APIKey == "" {
		panic("GeminiCleverChatter requires APIKey")
	}
	switch opts.ThinkingLevel {
	case "", GeminiThinkingLevelLow, GeminiThinkingLevelMedium, GeminiThinkingLevelHigh:
	default:
		panic("GeminiCleverChatter ThinkingLevel must be low, medium, or high")
	}
	return &GeminiCleverChatter{options: opts}
}

// ToConfig serializes to the mllm field of the start request.
func (g *GeminiCleverChatter) ToConfig() map[string]interface{} {
	var modelParams map[string]interface{}
	if g.options.ThinkingLevel != "" {
		modelParams = map[string]interface{}{"thinking_level": g.options.ThinkingLevel}
	}
	return buildGeminiPreviewMLLMConfig(
		GeminiPreviewDefaultModelCleverChatter,
		g.options.GeminiPreviewMLLMOptions,
		modelParams,
	)
}
