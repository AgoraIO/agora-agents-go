package vendors

import "strings"

const (
	LiveAvatarRequiredSampleRate = SampleRate24kHz
	AkoolRequiredSampleRate      = SampleRate16kHz
)

type LiveAvatarAvatarOptions struct {
	APIKey              string
	Quality             string
	AgoraUID            string
	AgoraToken          string
	AvatarID            string
	Enable              *bool
	DisableIdleTimeout  *bool
	ActivityIdleTimeout *int
	AdditionalParams    map[string]interface{}
}

type LiveAvatarAvatar struct {
	options LiveAvatarAvatarOptions
}

func NewLiveAvatarAvatar(opts LiveAvatarAvatarOptions) *LiveAvatarAvatar {
	if opts.APIKey == "" {
		panic("LiveAvatarAvatar requires APIKey")
	}
	if opts.Quality == "" {
		panic("LiveAvatarAvatar requires Quality (low, medium, or high)")
	}
	if opts.Quality != "low" && opts.Quality != "medium" && opts.Quality != "high" {
		panic("LiveAvatarAvatar Quality must be one of: low, medium, high")
	}
	if opts.AgoraUID == "" {
		panic("LiveAvatarAvatar requires AgoraUID")
	}
	return &LiveAvatarAvatar{options: opts}
}

func (l *LiveAvatarAvatar) RequiredSampleRate() SampleRate {
	return LiveAvatarRequiredSampleRate
}

func (l *LiveAvatarAvatar) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range l.options.AdditionalParams {
		params[k] = v
	}
	params["api_key"] = l.options.APIKey
	params["quality"] = l.options.Quality
	params["agora_uid"] = l.options.AgoraUID
	if l.options.AgoraToken != "" {
		params["agora_token"] = l.options.AgoraToken
	}
	if l.options.AvatarID != "" {
		params["avatar_id"] = l.options.AvatarID
	}
	if l.options.DisableIdleTimeout != nil {
		params["disable_idle_timeout"] = *l.options.DisableIdleTimeout
	}
	if l.options.ActivityIdleTimeout != nil {
		params["activity_idle_timeout"] = *l.options.ActivityIdleTimeout
	}

	enable := true
	if l.options.Enable != nil {
		enable = *l.options.Enable
	}
	return map[string]interface{}{
		"enable": enable,
		"vendor": "liveavatar",
		"params": params,
	}
}

type AkoolAvatarOptions struct {
	APIKey           string
	AvatarID         string
	Enable           *bool
	AdditionalParams map[string]interface{}
}

type AkoolAvatar struct {
	options AkoolAvatarOptions
}

func NewAkoolAvatar(opts AkoolAvatarOptions) *AkoolAvatar {
	if opts.APIKey == "" {
		panic("AkoolAvatar requires APIKey")
	}
	return &AkoolAvatar{options: opts}
}

func (a *AkoolAvatar) RequiredSampleRate() SampleRate {
	return AkoolRequiredSampleRate
}

func (a *AkoolAvatar) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range a.options.AdditionalParams {
		params[k] = v
	}
	params["api_key"] = a.options.APIKey
	if a.options.AvatarID != "" {
		params["avatar_id"] = a.options.AvatarID
	}
	enable := true
	if a.options.Enable != nil {
		enable = *a.options.Enable
	}
	return map[string]interface{}{
		"enable": enable,
		"vendor": "akool",
		"params": params,
	}
}

type AnamAvatarOptions struct {
	APIKey           string
	AvatarID         string
	Enable           *bool
	AdditionalParams map[string]interface{}
}

type AnamAvatar struct {
	options AnamAvatarOptions
}

func NewAnamAvatar(opts AnamAvatarOptions) *AnamAvatar {
	if opts.APIKey == "" {
		panic("AnamAvatar requires APIKey")
	}
	return &AnamAvatar{options: opts}
}

func (a *AnamAvatar) RequiredSampleRate() SampleRate {
	return 0
}

func (a *AnamAvatar) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range a.options.AdditionalParams {
		params[k] = v
	}
	params["api_key"] = a.options.APIKey
	if a.options.AvatarID != "" {
		params["avatar_id"] = a.options.AvatarID
	}
	enable := true
	if a.options.Enable != nil {
		enable = *a.options.Enable
	}
	return map[string]interface{}{
		"enable": enable,
		"vendor": "anam",
		"params": params,
	}
}

type GenericAvatarOptions struct {
	APIKey           string
	APIBaseURL       string
	AvatarID         string
	AgoraUID         string
	AgoraToken       string
	AgoraAppID       string
	AgoraChannel     string
	Enable           *bool
	AdditionalParams map[string]interface{}
}

type GenericAvatar struct {
	options GenericAvatarOptions
}

func NewGenericAvatar(opts GenericAvatarOptions) *GenericAvatar {
	if opts.APIKey == "" {
		panic("GenericAvatar requires APIKey")
	}
	if opts.APIBaseURL == "" {
		panic("GenericAvatar requires APIBaseURL")
	}
	if opts.AvatarID == "" {
		panic("GenericAvatar requires AvatarID")
	}
	if opts.AgoraUID == "" {
		panic("GenericAvatar requires AgoraUID")
	}
	return &GenericAvatar{options: opts}
}

// TavusOptions retains the generic options with an optional provider URL.
type TavusOptions = GenericAvatarOptions

// Tavus reuses GenericAvatar configuration and session/token handling.
type Tavus struct{ *GenericAvatar }

func NewTavus(opts TavusOptions) *Tavus {
	generic := GenericAvatarOptions(opts)
	if generic.APIBaseURL == "" {
		generic.APIBaseURL = "https://tavusapi.com/v2/conversations/agora"
	}
	return &Tavus{NewGenericAvatar(generic)}
}

type ProtofaceOptions = GenericAvatarOptions

// Protoface reuses GenericAvatar configuration and session/token handling.
type Protoface struct{ *GenericAvatar }

func NewProtoface(opts ProtofaceOptions) *Protoface {
	generic := GenericAvatarOptions(opts)
	if generic.APIBaseURL == "" {
		generic.APIBaseURL = "https://api.protoface.com/v1/agora"
	}
	return &Protoface{NewGenericAvatar(generic)}
}

// LemonSliceOptions exposes generic fields directly for struct literals.
// Pointer fields distinguish omission from explicitly supplied empty values.
type LemonSliceOptions struct {
	APIKey           string
	APIBaseURL       string
	AvatarID         string
	AgoraUID         string
	AgoraToken       string
	AgoraAppID       string
	AgoraChannel     string
	Enable           *bool
	AdditionalParams map[string]interface{}
	AgentImageURL    *string
	AgentID          *string
	AgentImageBase64 *string
	AspectRatio      *string
}

// LemonSlice reuses GenericAvatar configuration and session/token handling.
type LemonSlice struct{ *GenericAvatar }

func NewLemonSlice(opts LemonSliceOptions) *LemonSlice {
	params := make(map[string]interface{}, len(opts.AdditionalParams)+4)
	for key, value := range opts.AdditionalParams {
		params[key] = value
	}
	for key, value := range map[string]*string{
		"agent_image_url":    opts.AgentImageURL,
		"agent_id":           opts.AgentID,
		"agent_image_base64": opts.AgentImageBase64,
		"aspect_ratio":       opts.AspectRatio,
	} {
		if value != nil {
			params[key] = *value
		}
	}
	selectors := 0
	for _, key := range []string{"agent_id", "agent_image_url", "agent_image_base64"} {
		if value, exists := params[key]; exists {
			text, ok := value.(string)
			if !ok || strings.TrimSpace(text) == "" {
				panic("LemonSlice requires nonempty string " + key)
			}
			selectors++
		}
	}
	if selectors != 1 {
		panic("LemonSlice requires exactly one of agent_id, agent_image_url, agent_image_base64")
	}
	if value, exists := params["aspect_ratio"]; exists {
		ratio, ok := value.(string)
		if !ok || (ratio != "2x3" && ratio != "9x16" && ratio != "1x1") {
			panic("LemonSlice aspect_ratio must be one of: 2x3, 9x16, 1x1")
		}
	}
	generic := GenericAvatarOptions{
		APIKey: opts.APIKey, APIBaseURL: opts.APIBaseURL, AvatarID: opts.AvatarID,
		AgoraUID: opts.AgoraUID, AgoraToken: opts.AgoraToken, AgoraAppID: opts.AgoraAppID,
		AgoraChannel: opts.AgoraChannel, Enable: opts.Enable, AdditionalParams: params,
	}
	if generic.APIBaseURL == "" {
		generic.APIBaseURL = "https://lemonslice.com/api/liveai/agora"
	}
	if generic.AvatarID == "" {
		generic.AvatarID = "lemonslice"
	}
	return &LemonSlice{NewGenericAvatar(generic)}
}

func (g *GenericAvatar) RequiredSampleRate() SampleRate {
	return 0
}

func (g *GenericAvatar) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range g.options.AdditionalParams {
		params[k] = v
	}
	params["api_key"] = g.options.APIKey
	params["api_base_url"] = g.options.APIBaseURL
	params["avatar_id"] = g.options.AvatarID
	params["agora_uid"] = g.options.AgoraUID
	if g.options.AgoraToken != "" {
		params["agora_token"] = g.options.AgoraToken
	}
	if g.options.AgoraAppID != "" {
		params["agora_appid"] = g.options.AgoraAppID
	}
	if g.options.AgoraChannel != "" {
		params["agora_channel"] = g.options.AgoraChannel
	}
	enable := true
	if g.options.Enable != nil {
		enable = *g.options.Enable
	}
	return map[string]interface{}{
		"enable": enable,
		"vendor": "generic",
		"params": params,
	}
}

// HeyGenAvatarOptions is deprecated.
//
// Deprecated: Use LiveAvatarAvatarOptions instead.
type HeyGenAvatarOptions = LiveAvatarAvatarOptions

// HeyGenAvatar is deprecated.
//
// Deprecated: Use LiveAvatarAvatar instead.
type HeyGenAvatar struct {
	options HeyGenAvatarOptions
}

// NewHeyGenAvatar is deprecated.
//
// Deprecated: Use NewLiveAvatarAvatar instead.
func NewHeyGenAvatar(opts HeyGenAvatarOptions) *HeyGenAvatar {
	if opts.APIKey == "" {
		panic("HeyGenAvatar requires APIKey")
	}
	if opts.Quality == "" {
		panic("HeyGenAvatar requires Quality (low, medium, or high)")
	}
	if opts.Quality != "low" && opts.Quality != "medium" && opts.Quality != "high" {
		panic("HeyGenAvatar Quality must be one of: low, medium, high")
	}
	if opts.AgoraUID == "" {
		panic("HeyGenAvatar requires AgoraUID")
	}
	return &HeyGenAvatar{options: opts}
}

func (h *HeyGenAvatar) RequiredSampleRate() SampleRate {
	return LiveAvatarRequiredSampleRate
}

func (h *HeyGenAvatar) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range h.options.AdditionalParams {
		params[k] = v
	}
	params["api_key"] = h.options.APIKey
	params["quality"] = h.options.Quality
	params["agora_uid"] = h.options.AgoraUID
	if h.options.AgoraToken != "" {
		params["agora_token"] = h.options.AgoraToken
	}
	if h.options.AvatarID != "" {
		params["avatar_id"] = h.options.AvatarID
	}
	if h.options.DisableIdleTimeout != nil {
		params["disable_idle_timeout"] = *h.options.DisableIdleTimeout
	}
	if h.options.ActivityIdleTimeout != nil {
		params["activity_idle_timeout"] = *h.options.ActivityIdleTimeout
	}

	enable := true
	if h.options.Enable != nil {
		enable = *h.options.Enable
	}
	return map[string]interface{}{
		"enable": enable,
		"vendor": "heygen",
		"params": params,
	}
}
