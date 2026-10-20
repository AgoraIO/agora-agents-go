package vendors

import (
	"slices"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit/core"
)

// RTZRSTTOptions configures the global RTZR streaming speech-to-text provider.
type RTZRSTTOptions struct {
	ClientID     string
	ClientSecret string
	APIBase      string
	ModelName    string
	// Language is the recognition language; the service defaults to Korean (ko).
	// The top-level ASR interaction language comes from turn detection.
	Language            string
	SampleRate          *int
	Encoding            string
	UseITN              *bool
	UseDisfluencyFilter *bool
	UseProfanityFilter  *bool
	UsePunctuation      *bool
	// Keywords is omitted when nil; an empty slice is sent explicitly.
	Keywords []string
	// AdditionalParams are flattened into params. Explicit options take precedence.
	AdditionalParams map[string]any
}

// RTZRSTT configures RTZR streaming speech recognition.
type RTZRSTT struct {
	options RTZRSTTOptions
}

var _ core.STTVendor = (*RTZRSTT)(nil)

// NewRTZRSTT creates an RTZR configuration and requires both client credentials.
func NewRTZRSTT(opts RTZRSTTOptions) *RTZRSTT {
	if opts.ClientID == "" {
		panic("RTZRSTT requires ClientID")
	}
	if opts.ClientSecret == "" {
		panic("RTZRSTT requires ClientSecret")
	}
	return &RTZRSTT{options: opts}
}

// ToConfig returns the RTZR ASR configuration expected by the API.
func (r *RTZRSTT) ToConfig() map[string]any {
	params := core.CloneConfig(r.options.AdditionalParams)
	if params == nil {
		params = map[string]any{}
	}
	params["client_id"] = r.options.ClientID
	params["client_secret"] = r.options.ClientSecret
	if r.options.APIBase != "" {
		params["api_base"] = r.options.APIBase
	}
	if r.options.ModelName != "" {
		params["model_name"] = r.options.ModelName
	}
	if r.options.Language != "" {
		params["language"] = r.options.Language
	}
	if r.options.SampleRate != nil {
		params["sample_rate"] = *r.options.SampleRate
	}
	if r.options.Encoding != "" {
		params["encoding"] = r.options.Encoding
	}
	if r.options.UseITN != nil {
		params["use_itn"] = *r.options.UseITN
	}
	if r.options.UseDisfluencyFilter != nil {
		params["use_disfluency_filter"] = *r.options.UseDisfluencyFilter
	}
	if r.options.UseProfanityFilter != nil {
		params["use_profanity_filter"] = *r.options.UseProfanityFilter
	}
	if r.options.UsePunctuation != nil {
		params["use_punctuation"] = *r.options.UsePunctuation
	}
	if r.options.Keywords != nil {
		params["keywords"] = slices.Clone(r.options.Keywords)
	}
	return map[string]any{"vendor": "rtzr", "params": params}
}
