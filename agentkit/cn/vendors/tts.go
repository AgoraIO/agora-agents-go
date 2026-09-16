package vendors

import (
	"net/url"
	"strings"
)

type MiniMaxVoiceSetting struct {
	VoiceID              string
	Speed                *int
	Volume               *int
	Pitch                *int
	Emotion              string
	LatexRead            *bool
	EnglishNormalization *bool
}

type MiniMaxAudioSetting struct {
	SampleRate int
}

type MiniMaxPronunciationDict struct {
	Tone []string
}

type MiniMaxTimberWeight struct {
	VoiceID string
	Weight  int
}

type MiniMaxTTSOptions struct {
	Key               string
	Model             string
	VoiceSetting      *MiniMaxVoiceSetting
	AudioSetting      *MiniMaxAudioSetting
	PronunciationDict *MiniMaxPronunciationDict
	TimberWeights     []MiniMaxTimberWeight
	LanguageBoost     string
	AdditionalParams  map[string]interface{}
	SkipPatterns      []int
}

type MiniMaxTTS struct {
	options MiniMaxTTSOptions
}

func NewMiniMaxTTS(opts MiniMaxTTSOptions) *MiniMaxTTS {
	if opts.Key == "" {
		panic("MiniMaxTTS requires Key")
	}
	if opts.Model == "" {
		panic("MiniMaxTTS requires Model")
	}
	if len(opts.TimberWeights) == 0 {
		if opts.VoiceSetting == nil || opts.VoiceSetting.VoiceID == "" {
			panic("MiniMaxTTS requires voice_setting.voice_id or TimberWeights")
		}
	}
	return &MiniMaxTTS{options: opts}
}

func (m *MiniMaxTTS) GetSampleRate() *SampleRate {
	return nil
}

func (m *MiniMaxTTS) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range m.options.AdditionalParams {
		params[k] = v
	}
	params["key"] = m.options.Key
	params["model"] = m.options.Model
	if m.options.VoiceSetting != nil {
		voiceSetting := map[string]interface{}{
			"voice_id": m.options.VoiceSetting.VoiceID,
		}
		if m.options.VoiceSetting.Speed != nil {
			voiceSetting["speed"] = *m.options.VoiceSetting.Speed
		}
		if m.options.VoiceSetting.Volume != nil {
			voiceSetting["vol"] = *m.options.VoiceSetting.Volume
		}
		if m.options.VoiceSetting.Pitch != nil {
			voiceSetting["pitch"] = *m.options.VoiceSetting.Pitch
		}
		if m.options.VoiceSetting.Emotion != "" {
			voiceSetting["emotion"] = m.options.VoiceSetting.Emotion
		}
		if m.options.VoiceSetting.LatexRead != nil {
			voiceSetting["latex_read"] = *m.options.VoiceSetting.LatexRead
		}
		if m.options.VoiceSetting.EnglishNormalization != nil {
			voiceSetting["english_normalization"] = *m.options.VoiceSetting.EnglishNormalization
		}
		params["voice_setting"] = voiceSetting
	}
	if len(m.options.TimberWeights) > 0 {
		timberWeights := make([]map[string]interface{}, 0, len(m.options.TimberWeights))
		for _, item := range m.options.TimberWeights {
			timberWeights = append(timberWeights, map[string]interface{}{
				"voice_id": item.VoiceID,
				"weight":   item.Weight,
			})
		}
		params["timber_weights"] = timberWeights
	}
	if m.options.AudioSetting != nil {
		params["audio_setting"] = map[string]interface{}{
			"sample_rate": m.options.AudioSetting.SampleRate,
		}
	}
	if m.options.PronunciationDict != nil {
		params["pronunciation_dict"] = map[string]interface{}{
			"tone": m.options.PronunciationDict.Tone,
		}
	}
	if m.options.LanguageBoost != "" {
		params["language_boost"] = m.options.LanguageBoost
	}

	config := map[string]interface{}{
		"vendor": "minimax",
		"params": params,
	}
	if m.options.SkipPatterns != nil {
		config["skip_patterns"] = m.options.SkipPatterns
	}
	return config
}

type TencentTTSOptions struct {
	AppID            string
	SecretID         string
	SecretKey        string
	VoiceType        *int
	Volume           *float64
	Speed            *float64
	EmotionCategory  string
	EmotionIntensity *int
	AdditionalParams map[string]interface{}
	SkipPatterns     []int
}

type TencentTTS struct {
	options TencentTTSOptions
}

func NewTencentTTS(opts TencentTTSOptions) *TencentTTS {
	if opts.AppID == "" {
		panic("TencentTTS requires AppID")
	}
	if opts.SecretID == "" {
		panic("TencentTTS requires SecretID")
	}
	if opts.SecretKey == "" {
		panic("TencentTTS requires SecretKey")
	}
	return &TencentTTS{options: opts}
}

func (t *TencentTTS) GetSampleRate() *SampleRate {
	return nil
}

func (t *TencentTTS) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range t.options.AdditionalParams {
		params[k] = v
	}
	params["app_id"] = t.options.AppID
	params["secret_id"] = t.options.SecretID
	params["secret_key"] = t.options.SecretKey
	if t.options.VoiceType != nil {
		params["voice_type"] = *t.options.VoiceType
	}
	if t.options.Volume != nil {
		params["volume"] = *t.options.Volume
	}
	if t.options.Speed != nil {
		params["speed"] = *t.options.Speed
	}
	if t.options.EmotionCategory != "" {
		params["emotion_category"] = t.options.EmotionCategory
	}
	if t.options.EmotionIntensity != nil {
		params["emotion_intensity"] = *t.options.EmotionIntensity
	}
	config := map[string]interface{}{
		"vendor": "tencent",
		"params": params,
	}
	if t.options.SkipPatterns != nil {
		config["skip_patterns"] = t.options.SkipPatterns
	}
	return config
}

type MicrosoftTTSOptions struct {
	Key              string
	Region           string
	VoiceName        string
	SampleRate       *SampleRate
	Speed            *float64
	Volume           *float64
	AdditionalParams map[string]interface{}
	SkipPatterns     []int
}

type MicrosoftTTS struct {
	options MicrosoftTTSOptions
}

func NewMicrosoftTTS(opts MicrosoftTTSOptions) *MicrosoftTTS {
	if opts.Key == "" {
		panic("MicrosoftTTS requires Key")
	}
	if opts.Region == "" {
		panic("MicrosoftTTS requires Region")
	}
	if opts.VoiceName == "" {
		panic("MicrosoftTTS requires VoiceName")
	}
	return &MicrosoftTTS{options: opts}
}

func (m *MicrosoftTTS) GetSampleRate() *SampleRate {
	return m.options.SampleRate
}

func (m *MicrosoftTTS) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range m.options.AdditionalParams {
		params[k] = v
	}
	params["key"] = m.options.Key
	params["region"] = m.options.Region
	params["voice_name"] = m.options.VoiceName
	if m.options.SampleRate != nil {
		params["sample_rate"] = int(*m.options.SampleRate)
	}
	if m.options.Speed != nil {
		params["speed"] = *m.options.Speed
	}
	if m.options.Volume != nil {
		params["volume"] = *m.options.Volume
	}

	config := map[string]interface{}{
		"vendor": "microsoft",
		"params": params,
	}
	if m.options.SkipPatterns != nil {
		config["skip_patterns"] = m.options.SkipPatterns
	}
	return config
}

const genericTTSURLValidationMessage = "GenericTTS currently supports only HTTP and HTTPS URLs"

type GenericTTSOptions struct {
	URL              string
	Headers          map[string]string
	APIKey           string
	Model            string
	Voice            string
	Speed            *float64
	SampleRate       *SampleRate
	ResponseFormat   string
	Instruction      string
	AdditionalParams map[string]interface{}
	SkipPatterns     []int
}

type GenericTTS struct {
	options GenericTTSOptions
	vendor  string
}

func NewGenericTTS(opts GenericTTSOptions) *GenericTTS {
	if opts.URL == "" {
		panic("GenericTTS requires URL")
	}
	return &GenericTTS{
		options: opts,
		vendor:  genericTTSVendorForURL(opts.URL),
	}
}

func (g *GenericTTS) GetSampleRate() *SampleRate {
	return g.options.SampleRate
}

func (g *GenericTTS) ToConfig() map[string]interface{} {
	params := make(map[string]interface{}, len(g.options.AdditionalParams)+1)
	for key, value := range g.options.AdditionalParams {
		params[key] = value
	}
	if g.options.Model != "" {
		params["model"] = g.options.Model
	}
	if g.options.Voice != "" {
		params["voice"] = g.options.Voice
	}
	if g.options.APIKey != "" {
		params["api_key"] = g.options.APIKey
	}
	if g.options.Speed != nil {
		params["speed"] = *g.options.Speed
	}
	if g.options.SampleRate != nil {
		params["sample_rate"] = int(*g.options.SampleRate)
	}
	if g.options.ResponseFormat != "" {
		params["response_format"] = g.options.ResponseFormat
	}
	if g.options.Instruction != "" {
		params["instruction"] = g.options.Instruction
	}

	config := map[string]interface{}{
		"vendor": g.vendor,
		"url":    g.options.URL,
		"params": params,
	}
	if len(g.options.Headers) > 0 {
		config["headers"] = g.options.Headers
	}
	if g.options.SkipPatterns != nil {
		config["skip_patterns"] = g.options.SkipPatterns
	}
	return config
}

func genericTTSVendorForURL(rawURL string) string {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil || parsedURL.Host == "" {
		panic(genericTTSURLValidationMessage)
	}

	switch strings.ToLower(parsedURL.Scheme) {
	case "http", "https":
		return "generic_http"
	default:
		panic(genericTTSURLValidationMessage)
	}
}

type BytedanceTTSOptions struct {
	Token            string
	AppID            string
	Cluster          string
	VoiceType        string
	SpeedRatio       *float64
	VolumeRatio      *float64
	PitchRatio       *float64
	Emotion          string
	AdditionalParams map[string]interface{}
	SkipPatterns     []int
}

type BytedanceTTS struct {
	options BytedanceTTSOptions
}

func NewBytedanceTTS(opts BytedanceTTSOptions) *BytedanceTTS {
	if opts.Token == "" {
		panic("BytedanceTTS requires Token")
	}
	if opts.AppID == "" {
		panic("BytedanceTTS requires AppID")
	}
	if opts.Cluster == "" {
		panic("BytedanceTTS requires Cluster")
	}
	if opts.VoiceType == "" {
		panic("BytedanceTTS requires VoiceType")
	}
	return &BytedanceTTS{options: opts}
}

func (b *BytedanceTTS) GetSampleRate() *SampleRate {
	return nil
}

func (b *BytedanceTTS) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range b.options.AdditionalParams {
		params[k] = v
	}
	params["token"] = b.options.Token
	params["app_id"] = b.options.AppID
	params["cluster"] = b.options.Cluster
	params["voice_type"] = b.options.VoiceType
	if b.options.SpeedRatio != nil {
		params["speed_ratio"] = *b.options.SpeedRatio
	}
	if b.options.VolumeRatio != nil {
		params["volume_ratio"] = *b.options.VolumeRatio
	}
	if b.options.PitchRatio != nil {
		params["pitch_ratio"] = *b.options.PitchRatio
	}
	if b.options.Emotion != "" {
		params["emotion"] = b.options.Emotion
	}
	config := map[string]interface{}{
		"vendor": "bytedance",
		"params": params,
	}
	if b.options.SkipPatterns != nil {
		config["skip_patterns"] = b.options.SkipPatterns
	}
	return config
}

type CosyVoiceTTSOptions struct {
	APIKey           string
	Model            string
	Voice            string
	SampleRate       *int
	AdditionalParams map[string]interface{}
	SkipPatterns     []int
}

type CosyVoiceTTS struct {
	options CosyVoiceTTSOptions
}

func NewCosyVoiceTTS(opts CosyVoiceTTSOptions) *CosyVoiceTTS {
	if opts.APIKey == "" {
		panic("CosyVoiceTTS requires APIKey")
	}
	if opts.Model == "" {
		panic("CosyVoiceTTS requires Model")
	}
	if opts.Voice == "" {
		panic("CosyVoiceTTS requires Voice")
	}
	return &CosyVoiceTTS{options: opts}
}

func (c *CosyVoiceTTS) GetSampleRate() *SampleRate {
	if c.options.SampleRate == nil {
		return nil
	}
	sr := SampleRate(*c.options.SampleRate)
	return &sr
}

func (c *CosyVoiceTTS) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range c.options.AdditionalParams {
		params[k] = v
	}
	params["api_key"] = c.options.APIKey
	params["model"] = c.options.Model
	params["voice"] = c.options.Voice
	if c.options.SampleRate != nil {
		params["sample_rate"] = *c.options.SampleRate
	}
	config := map[string]interface{}{
		"vendor": "cosyvoice",
		"params": params,
	}
	if c.options.SkipPatterns != nil {
		config["skip_patterns"] = c.options.SkipPatterns
	}
	return config
}

type BytedanceDuplexTTSOptions struct {
	AppID            string
	Token            string
	Speaker          string
	AdditionalParams map[string]interface{}
	SkipPatterns     []int
}

type BytedanceDuplexTTS struct {
	options BytedanceDuplexTTSOptions
}

func NewBytedanceDuplexTTS(opts BytedanceDuplexTTSOptions) *BytedanceDuplexTTS {
	if opts.AppID == "" {
		panic("BytedanceDuplexTTS requires AppID")
	}
	if opts.Token == "" {
		panic("BytedanceDuplexTTS requires Token")
	}
	if opts.Speaker == "" {
		panic("BytedanceDuplexTTS requires Speaker")
	}
	return &BytedanceDuplexTTS{options: opts}
}

func (b *BytedanceDuplexTTS) GetSampleRate() *SampleRate {
	return nil
}

func (b *BytedanceDuplexTTS) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range b.options.AdditionalParams {
		params[k] = v
	}
	params["app_id"] = b.options.AppID
	params["token"] = b.options.Token
	params["speaker"] = b.options.Speaker
	config := map[string]interface{}{
		"vendor": "bytedance_duplex",
		"params": params,
	}
	if b.options.SkipPatterns != nil {
		config["skip_patterns"] = b.options.SkipPatterns
	}
	return config
}

type StepFunTTSOptions struct {
	APIKey           string
	Model            string
	VoiceID          string
	AdditionalParams map[string]interface{}
	SkipPatterns     []int
}

type StepFunTTS struct {
	options StepFunTTSOptions
}

func NewStepFunTTS(opts StepFunTTSOptions) *StepFunTTS {
	if opts.APIKey == "" {
		panic("StepFunTTS requires APIKey")
	}
	if opts.Model == "" {
		panic("StepFunTTS requires Model")
	}
	if opts.VoiceID == "" {
		panic("StepFunTTS requires VoiceID")
	}
	return &StepFunTTS{options: opts}
}

func (s *StepFunTTS) GetSampleRate() *SampleRate {
	return nil
}

func (s *StepFunTTS) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range s.options.AdditionalParams {
		params[k] = v
	}
	params["api_key"] = s.options.APIKey
	params["model"] = s.options.Model
	params["voice_id"] = s.options.VoiceID
	config := map[string]interface{}{
		"vendor": "stepfun",
		"params": params,
	}
	if s.options.SkipPatterns != nil {
		config["skip_patterns"] = s.options.SkipPatterns
	}
	return config
}
