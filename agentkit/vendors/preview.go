package vendors

// Preview provider vendor types.
//
// These follow the same shape as the GA vendor types in this package — options
// struct in, snake_case wire config out — so they drop into agent.WithStt()
// unchanged. AgentSession detects them and routes to the preview endpoint.

// GeminiSTTModel35Live is the default Gemini preview transcription model.
const GeminiSTTModel35Live = "gemini-3.5-transcribe-live"

// GeminiSTTOptions configures [GeminiSTT].
type GeminiSTTOptions struct {
	// APIKey is the Google API key.
	APIKey string
	// Model name. Defaults to gemini-3.5-transcribe-live.
	Model string
	// LanguageCodes are the languages the model should transcribe, sent as
	// params.language_codes.
	//
	// Omitted from the request when nil, which is how the provider spells
	// auto-detect — the SDK does not pin a language the caller never asked for.
	// Pass one code to commit to a language, several to let the model choose
	// between them, or an empty non-nil slice to request auto-detect outright.
	//
	// This is the only language setting on this vendor. The top-level
	// asr.language is supplied by Agent from turn detection, as it is for every
	// STT vendor.
	LanguageCodes []string
	// CustomVocabulary biases recognition toward these words and phrases —
	// product names, jargon, proper nouns the model would otherwise mis-hear.
	CustomVocabulary []string
	// SampleRate is the audio sample rate in Hz. Defaults to 16000.
	SampleRate int
	// WordTimestamp emits per-word timestamps in transcription results.
	// It is omitted unless explicitly set and cannot be true when
	// CustomVocabulary is set.
	WordTimestamp *bool
	// AdditionalParams are additional vendor-specific parameters.
	AdditionalParams map[string]interface{}
}

// GeminiSTT is the Gemini 3.5 Transcribe ASR vendor (preview).
//
// Example:
//
//	agent.WithStt(vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
//		APIKey:        apiKey,
//		LanguageCodes: []string{"en-US"},
//	}))
type GeminiSTT struct {
	options GeminiSTTOptions
}

// NewGeminiSTT creates a Gemini 3.5 Transcribe ASR vendor (preview).
func NewGeminiSTT(opts GeminiSTTOptions) *GeminiSTT {
	if opts.APIKey == "" {
		panic("GeminiSTT requires APIKey")
	}
	return &GeminiSTT{options: opts}
}

// ToConfig serializes to the asr field of the start request.
func (g *GeminiSTT) ToConfig() map[string]interface{} {
	model := g.options.Model
	if model == "" {
		model = GeminiSTTModel35Live
	}
	sampleRate := g.options.SampleRate
	if sampleRate == 0 {
		sampleRate = 16000
	}
	// AdditionalParams first so that explicit fields always win.
	params := map[string]interface{}{}
	for k, v := range g.options.AdditionalParams {
		params[k] = v
	}
	params["api_key"] = g.options.APIKey
	params["model"] = model
	params["sample_rate"] = sampleRate
	// Omitted unless the caller asked for it: no language_codes is how the
	// provider spells auto-detect. Nil versus empty matters — an empty non-nil
	// slice is an explicit auto-detect and still goes on the wire as [].
	if g.options.LanguageCodes != nil {
		params["language_codes"] = g.options.LanguageCodes
	}
	if g.options.CustomVocabulary != nil {
		params["custom_vocabulary"] = g.options.CustomVocabulary
	}
	if g.options.WordTimestamp != nil {
		params["word_timestamp"] = *g.options.WordTimestamp
	}
	if _, hasCustomVocabulary := params["custom_vocabulary"]; hasCustomVocabulary && params["word_timestamp"] == true {
		panic("CustomVocabulary cannot be used with WordTimestamp=true")
	}

	// No top-level "language": Agent sets it from turn detection,
	// the same as every other STT vendor.
	return map[string]interface{}{
		"vendor": "gemini",
		"params": params,
	}
}
