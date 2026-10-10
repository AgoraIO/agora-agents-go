package vendors

import "strings"

// Gemini 3.8 Flash TTS model.
const (
	GeminiTTSModelFlash38 = "gemini-3.8-flash-tts"
)

// GeminiTTSOptions configures Gemini TTS. Model names are sent verbatim.
type GeminiTTSOptions struct {
	APIKey string
	// Model defaults to GeminiTTSModelFlash38.
	Model string
	// Voice defaults to Puck.
	Voice string
	// Style is an optional natural-language speaking instruction.
	Style string
	// AdditionalParams forwards provider-specific parameters under tts.params.
	AdditionalParams map[string]interface{}
	// SkipPatterns controls whether bracketed content is omitted from speech.
	SkipPatterns []int
}

// GeminiTTS uses the client's configured regional production endpoint.
type GeminiTTS struct{ options GeminiTTSOptions }

// NewGeminiTTS creates a Gemini TTS vendor using the preview-era options and defaults.
func NewGeminiTTS(opts GeminiTTSOptions) *GeminiTTS {
	if strings.TrimSpace(opts.APIKey) == "" {
		panic("GeminiTTS requires APIKey")
	}
	if opts.Model == "" {
		opts.Model = GeminiTTSModelFlash38
	}
	if opts.Voice == "" {
		opts.Voice = "Puck"
	}
	if strings.TrimSpace(opts.Model) == "" {
		panic("GeminiTTS requires Model")
	}
	if strings.TrimSpace(opts.Voice) == "" {
		panic("GeminiTTS requires Voice")
	}
	return &GeminiTTS{options: opts}
}

// GetSampleRate returns nil: the Gemini TTS contract does not expose a sample rate.
func (g *GeminiTTS) GetSampleRate() *SampleRate { return nil }

// ToConfig preserves the Gemini TTS wire shape, including credentials inside params.
func (g *GeminiTTS) ToConfig() map[string]interface{} {
	o := g.options
	params := make(map[string]interface{}, len(o.AdditionalParams)+4)
	for key, value := range o.AdditionalParams {
		params[key] = value
	}
	params["api_key"] = o.APIKey
	params["model"] = o.Model
	params["voice"] = o.Voice
	if o.Style != "" {
		params["style"] = o.Style
	}
	config := map[string]interface{}{"vendor": "gemini", "params": params}
	if o.SkipPatterns != nil {
		config["skip_patterns"] = o.SkipPatterns
	}
	return config
}
