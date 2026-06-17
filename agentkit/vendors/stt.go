package vendors

import "strings"

type SpeechmaticsSTTOptions struct {
	APIKey           string
	Language         string
	Model            string
	URI              string
	AdditionalParams map[string]interface{}
}

type SpeechmaticsSTT struct {
	options SpeechmaticsSTTOptions
}

func NewSpeechmaticsSTT(opts SpeechmaticsSTTOptions) *SpeechmaticsSTT {
	if opts.APIKey == "" {
		panic("SpeechmaticsSTT requires APIKey")
	}
	if opts.Language == "" {
		panic("SpeechmaticsSTT requires Language")
	}
	return &SpeechmaticsSTT{options: opts}
}

func (s *SpeechmaticsSTT) ToConfig() map[string]interface{} {
	params := map[string]interface{}{
		"api_key":  s.options.APIKey,
		"language": s.options.Language,
	}
	if s.options.Model != "" {
		params["model"] = s.options.Model
	}
	if s.options.URI != "" {
		params["uri"] = s.options.URI
	}
	for k, v := range s.options.AdditionalParams {
		if _, exists := params[k]; !exists {
			params[k] = v
		}
	}

	config := map[string]interface{}{
		"vendor": "speechmatics",
		"params": params,
	}
	return config
}

type DeepgramSTTOptions struct {
	APIKey           string
	Model            string
	Language         string
	Keyterm          string
	SmartFormat      *bool
	Punctuation      *bool
	AdditionalParams map[string]interface{}
}

type DeepgramSTT struct {
	options DeepgramSTTOptions
}

func NewDeepgramSTT(opts DeepgramSTTOptions) *DeepgramSTT {
	if opts.APIKey == "" {
		switch strings.ToLower(strings.TrimSpace(opts.Model)) {
		case "nova-2", "nova-3":
		default:
			panic("DeepgramSTT requires APIKey unless using a supported Agora-managed model")
		}
	}
	return &DeepgramSTT{options: opts}
}

func (d *DeepgramSTT) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range d.options.AdditionalParams {
		params[k] = v
	}
	if d.options.APIKey != "" {
		params["key"] = d.options.APIKey
	}
	if d.options.Model != "" {
		params["model"] = d.options.Model
	}
	if d.options.Language != "" {
		params["language"] = d.options.Language
	}
	if d.options.SmartFormat != nil {
		params["smart_format"] = *d.options.SmartFormat
	}
	if d.options.Punctuation != nil {
		params["punctuation"] = *d.options.Punctuation
	}
	if d.options.Keyterm != "" {
		params["keyterm"] = d.options.Keyterm
	}

	config := map[string]interface{}{
		"vendor": "deepgram",
		"params": params,
	}
	return config
}

type MicrosoftSTTOptions struct {
	Key              string
	Region           string
	Language         string
	AdditionalParams map[string]interface{}
}

type MicrosoftSTT struct {
	options MicrosoftSTTOptions
}

func NewMicrosoftSTT(opts MicrosoftSTTOptions) *MicrosoftSTT {
	if opts.Key == "" {
		panic("MicrosoftSTT requires Key")
	}
	if opts.Region == "" {
		panic("MicrosoftSTT requires Region")
	}
	if opts.Language == "" {
		panic("MicrosoftSTT requires Language")
	}
	return &MicrosoftSTT{options: opts}
}

func (m *MicrosoftSTT) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range m.options.AdditionalParams {
		params[k] = v
	}
	params["key"] = m.options.Key
	params["region"] = m.options.Region
	if m.options.Language != "" {
		params["language"] = m.options.Language
	}

	config := map[string]interface{}{
		"vendor": "microsoft",
		"params": params,
	}
	return config
}

type OpenAISTTOptions struct {
	APIKey                  string
	Model                   string
	Language                string
	Prompt                  string
	InputAudioTranscription map[string]interface{}
	AdditionalParams        map[string]interface{}
}

type OpenAISTT struct {
	options OpenAISTTOptions
}

func NewOpenAISTT(opts OpenAISTTOptions) *OpenAISTT {
	if opts.APIKey == "" {
		panic("OpenAISTT requires APIKey")
	}
	return &OpenAISTT{options: opts}
}

func (o *OpenAISTT) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range o.options.AdditionalParams {
		params[k] = v
	}
	params["api_key"] = o.options.APIKey
	transcription := map[string]interface{}{"model": "gpt-4o-mini-transcribe"}
	for k, v := range o.options.InputAudioTranscription {
		transcription[k] = v
	}
	if o.options.Model != "" {
		transcription["model"] = o.options.Model
	}
	if o.options.Prompt != "" {
		transcription["prompt"] = o.options.Prompt
	}
	if o.options.Language != "" {
		transcription["language"] = o.options.Language
	}
	if v, _ := transcription["model"].(string); v == "" {
		panic("OpenAISTT: input_audio_transcription.model is required")
	}
	if v, _ := transcription["prompt"].(string); v == "" {
		panic("OpenAISTT: input_audio_transcription.prompt is required")
	}
	if v, _ := transcription["language"].(string); v == "" {
		panic("OpenAISTT: input_audio_transcription.language is required")
	}
	params["input_audio_transcription"] = transcription

	config := map[string]interface{}{
		"vendor": "openai",
		"params": params,
	}
	return config
}

type GoogleSTTOptions struct {
	ProjectID            string
	Location             string
	ADCCredentialsString string
	Language             string
	Model                string
	AdditionalParams     map[string]interface{}
}

type GoogleSTT struct {
	options GoogleSTTOptions
}

func NewGoogleSTT(opts GoogleSTTOptions) *GoogleSTT {
	if opts.ProjectID == "" {
		panic("GoogleSTT requires ProjectID")
	}
	if opts.Location == "" {
		panic("GoogleSTT requires Location")
	}
	if opts.ADCCredentialsString == "" {
		panic("GoogleSTT requires ADCCredentialsString")
	}
	if opts.Language == "" {
		panic("GoogleSTT requires Language")
	}
	return &GoogleSTT{options: opts}
}

func (g *GoogleSTT) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range g.options.AdditionalParams {
		params[k] = v
	}
	params["project_id"] = g.options.ProjectID
	params["location"] = g.options.Location
	params["adc_credentials_string"] = g.options.ADCCredentialsString
	if g.options.Language != "" {
		params["language"] = g.options.Language
	}
	if g.options.Model != "" {
		params["model"] = g.options.Model
	}

	config := map[string]interface{}{
		"vendor": "google",
		"params": params,
	}
	return config
}

type AmazonSTTOptions struct {
	AccessKey        string
	SecretKey        string
	Region           string
	Language         string
	AdditionalParams map[string]interface{}
}

type AmazonSTT struct {
	options AmazonSTTOptions
}

func NewAmazonSTT(opts AmazonSTTOptions) *AmazonSTT {
	if opts.AccessKey == "" {
		panic("AmazonSTT requires AccessKey")
	}
	if opts.SecretKey == "" {
		panic("AmazonSTT requires SecretKey")
	}
	if opts.Region == "" {
		panic("AmazonSTT requires Region")
	}
	if opts.Language == "" {
		panic("AmazonSTT requires Language")
	}
	return &AmazonSTT{options: opts}
}

func (a *AmazonSTT) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range a.options.AdditionalParams {
		params[k] = v
	}
	params["access_key_id"] = a.options.AccessKey
	params["secret_access_key"] = a.options.SecretKey
	params["region"] = a.options.Region
	if a.options.Language != "" {
		params["language_code"] = a.options.Language
	}

	config := map[string]interface{}{
		"vendor": "amazon",
		"params": params,
	}
	return config
}

type AssemblyAISTTOptions struct {
	APIKey           string
	Language         string
	URI              string
	AdditionalParams map[string]interface{}
}

type AssemblyAISTT struct {
	options AssemblyAISTTOptions
}

func NewAssemblyAISTT(opts AssemblyAISTTOptions) *AssemblyAISTT {
	if opts.APIKey == "" {
		panic("AssemblyAISTT requires APIKey")
	}
	if opts.Language == "" {
		panic("AssemblyAISTT requires Language")
	}
	return &AssemblyAISTT{options: opts}
}

func (a *AssemblyAISTT) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range a.options.AdditionalParams {
		params[k] = v
	}
	params["api_key"] = a.options.APIKey
	if a.options.Language != "" {
		params["language"] = a.options.Language
	}
	if a.options.URI != "" {
		params["uri"] = a.options.URI
	}

	config := map[string]interface{}{
		"vendor": "assemblyai",
		"params": params,
	}
	return config
}

type AresSTTOptions struct {
	AdditionalParams map[string]interface{}
}

type AresSTT struct {
	options AresSTTOptions
}

func NewAresSTT(opts AresSTTOptions) *AresSTT {
	return &AresSTT{options: opts}
}

func (a *AresSTT) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range a.options.AdditionalParams {
		params[k] = v
	}
	config := map[string]interface{}{
		"vendor": "ares",
	}
	if len(params) > 0 {
		config["params"] = params
	}
	return config
}

type SarvamSTTOptions struct {
	APIKey           string
	Language         string
	Model            string
	AdditionalParams map[string]interface{}
}

type SarvamSTT struct {
	options SarvamSTTOptions
}

func NewSarvamSTT(opts SarvamSTTOptions) *SarvamSTT {
	if opts.APIKey == "" {
		panic("SarvamSTT requires APIKey")
	}
	if opts.Language == "" {
		panic("SarvamSTT requires Language")
	}
	return &SarvamSTT{options: opts}
}

func (s *SarvamSTT) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range s.options.AdditionalParams {
		params[k] = v
	}
	params["api_key"] = s.options.APIKey
	params["language"] = s.options.Language
	if s.options.Model != "" {
		params["model"] = s.options.Model
	}

	config := map[string]interface{}{
		"vendor": "sarvam",
		"params": params,
	}
	return config
}
