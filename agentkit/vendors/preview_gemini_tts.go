package vendors

import "strings"

// Gemini 3.8 Flash TTS preview model.
const (
	GeminiTTSModelFlash38 = "gemini-3.8-flash-tts"
)

// GeminiTTSOptions configures preview TTS. Model names are sent verbatim.
type GeminiTTSOptions struct {
	APIKey string
	// Model defaults to GeminiTTSModelFlash38.
	Model string
	// Voice defaults to Puck.
	Voice string
	// Style is an optional natural-language speaking instruction.
	Style string
}

// GeminiTTS uses the gemini-live preview gate through AgentSession.
type GeminiTTS struct{ options GeminiTTSOptions }

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

// GetSampleRate returns nil: the preview contract does not expose a sample rate.
func (g *GeminiTTS) GetSampleRate() *SampleRate { return nil }

func (g *GeminiTTS) ToConfig() map[string]interface{} {
	o := g.options
	params := map[string]interface{}{"api_key": o.APIKey, "model": o.Model, "voice": o.Voice}
	if o.Style != "" {
		params["style"] = o.Style
	}
	return map[string]interface{}{"vendor": "gemini", "params": params}
}
