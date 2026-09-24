package core

import (
	"encoding/json"
	"fmt"
	"strconv"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/agentmanagement"
	"github.com/AgoraIO/agora-agents-go/v2/agents"
	basecore "github.com/AgoraIO/agora-agents-go/v2/core"
)

type SampleRate int

const (
	SampleRate8kHz  SampleRate = 8000
	SampleRate16kHz SampleRate = 16000
	SampleRate22kHz SampleRate = 22050
	SampleRate24kHz SampleRate = 24000
	SampleRate44kHz SampleRate = 44100
	SampleRate48kHz SampleRate = 48000
)

type LLMVendor interface {
	ToConfig() map[string]interface{}
}

type TTSVendor interface {
	ToConfig() map[string]interface{}
	GetSampleRate() *SampleRate
}

type STTVendor interface {
	ToConfig() map[string]interface{}
}

type MLLMVendor interface {
	ToConfig() map[string]interface{}
}

type AvatarVendorConfig interface {
	ToConfig() map[string]interface{}
	RequiredSampleRate() SampleRate
}

type AgentOption func(*BaseAgent)

type ClientRuntime interface {
	AgentsClient() *agents.Client
	AgentManagementClient() *agentmanagement.Client
	AppID() string
	AppCertificate() string
	IsAppCredentialsMode() bool
}

type BaseAgent struct {
	Client                   ClientRuntime
	PipelineID               string
	Instructions             string
	Greeting                 string
	FailureMessage           string
	MaxHistory               *int
	LLM                      map[string]interface{}
	TTS                      map[string]interface{}
	STT                      map[string]interface{}
	MLLM                     map[string]interface{}
	TTSSampleRate            *SampleRate
	Avatar                   map[string]interface{}
	AvatarRequiredSampleRate *SampleRate
	TurnDetection            *TurnDetectionConfig
	Interruption             *InterruptionConfig
	GreetingConfigs          *LlmGreetingConfigs
	Sal                      *SalConfig
	AdvancedFeatures         *AdvancedFeatures
	Parameters               *SessionParams
	AudioScenario            *ParametersAudioScenario
	Geofence                 *GeofenceConfig
	Labels                   map[string]string
	RTC                      *RtcConfig
	FillerWords              *FillerWordsConfig
}

type AgentRuntime interface {
	BaseAgent() *BaseAgent
	Profile() Profile
}

type ToPropertiesOptions struct {
	Channel                        string
	AgentUID                       string
	RemoteUIDs                     []string
	Token                          string
	AppID                          string
	AppCertificate                 string
	ExpiresIn                      int
	IdleTimeout                    *int
	EnableStringUID                *bool
	SkipVendorValidation           bool
	SkipVendorValidationCategories []string
	AllowMissingVendorCategories   []string
	Warn                           func(string)
}

type TokenFactory func(GenerateConvoAITokenOptions) (string, error)

type GenerateConvoAITokenOptions struct {
	AppID           string
	AppCertificate  string
	ChannelName     string
	UID             int
	TokenExpire     int
	PrivilegeExpire int
}

func NewBaseAgent(opts ...AgentOption) *BaseAgent {
	a := &BaseAgent{}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func WithPipelineID(pipelineID string) AgentOption {
	return func(a *BaseAgent) {
		a.PipelineID = pipelineID
	}
}

func WithInstructions(instructions string) AgentOption {
	return func(a *BaseAgent) {
		a.Instructions = instructions
	}
}

func WithGreeting(greeting string) AgentOption {
	return func(a *BaseAgent) {
		a.Greeting = greeting
	}
}

func WithFailureMessage(msg string) AgentOption {
	return func(a *BaseAgent) {
		a.FailureMessage = msg
	}
}

func WithMaxHistory(n int) AgentOption {
	return func(a *BaseAgent) {
		a.MaxHistory = &n
	}
}

func WithTurnDetectionConfig(td *TurnDetectionConfig) AgentOption {
	return func(a *BaseAgent) {
		a.TurnDetection = td
	}
}

func WithInterruptionConfig(interruption *InterruptionConfig) AgentOption {
	return func(a *BaseAgent) {
		a.Interruption = interruption
	}
}

func WithGreetingConfigs(configs *LlmGreetingConfigs) AgentOption {
	return func(a *BaseAgent) {
		a.GreetingConfigs = configs
	}
}

func WithSalConfig(sal *SalConfig) AgentOption {
	return func(a *BaseAgent) {
		a.Sal = sal
	}
}

func WithAdvancedFeatures(af *AdvancedFeatures) AgentOption {
	return func(a *BaseAgent) {
		a.AdvancedFeatures = af
	}
}

func WithTools(enabled bool) AgentOption {
	return func(a *BaseAgent) {
		if a.AdvancedFeatures == nil {
			a.AdvancedFeatures = &AdvancedFeatures{}
		}
		a.AdvancedFeatures.EnableTools = &enabled
	}
}

func WithParameters(params *SessionParams) AgentOption {
	return func(a *BaseAgent) {
		a.Parameters = params
	}
}

func WithAudioScenario(audioScenario ParametersAudioScenario) AgentOption {
	return func(a *BaseAgent) {
		a.AudioScenario = &audioScenario
	}
}

func WithGeofence(gf *GeofenceConfig) AgentOption {
	return func(a *BaseAgent) {
		a.Geofence = gf
	}
}

func WithLabels(labels map[string]string) AgentOption {
	return func(a *BaseAgent) {
		a.Labels = labels
	}
}

func WithRtc(rtc *RtcConfig) AgentOption {
	return func(a *BaseAgent) {
		a.RTC = rtc
	}
}

func WithFillerWords(fw *FillerWordsConfig) AgentOption {
	return func(a *BaseAgent) {
		a.FillerWords = fw
	}
}

func WithGreetingAudioURL(url string) AgentOption {
	return func(a *BaseAgent) {
		if a.LLM == nil {
			a.LLM = map[string]interface{}{}
		}
		a.LLM["greeting_audio_url"] = url
	}
}

func WithSessionOptOut(optOut bool) AgentOption {
	return func(a *BaseAgent) {
		if a.Parameters == nil {
			a.Parameters = &SessionParams{}
		}
		a.Parameters.OptOut = &optOut
	}
}

func (a *BaseAgent) Clone() *BaseAgent {
	clone := *a
	if a.Labels != nil {
		clone.Labels = make(map[string]string, len(a.Labels))
		for k, v := range a.Labels {
			clone.Labels[k] = v
		}
	}
	return &clone
}

func (a *BaseAgent) ApplyLLMConfig(cfg map[string]interface{}) *BaseAgent {
	clone := a.Clone()
	clone.LLM = CloneConfig(cfg)
	return clone
}

func (a *BaseAgent) ApplySTTConfig(cfg map[string]interface{}) *BaseAgent {
	clone := a.Clone()
	clone.STT = CloneConfig(cfg)
	return clone
}

func (a *BaseAgent) ApplyTTSConfig(cfg map[string]interface{}, sr *SampleRate) *BaseAgent {
	clone := a.Clone()
	clone.TTS = CloneConfig(cfg)
	clone.TTSSampleRate = sr
	return clone
}

func (a *BaseAgent) ApplyMLLMConfig(cfg map[string]interface{}) *BaseAgent {
	clone := a.Clone()
	clone.MLLM = CloneConfig(cfg)
	if clone.MLLM != nil {
		clone.MLLM["enable"] = true
	}
	if clone.AdvancedFeatures != nil {
		clone.AdvancedFeatures.EnableMllm = nil
		if clone.AdvancedFeatures.EnableRtm == nil && clone.AdvancedFeatures.EnableSal == nil && clone.AdvancedFeatures.EnableTools == nil {
			clone.AdvancedFeatures = nil
		}
	}
	return clone
}

func (a *BaseAgent) ApplyAvatarConfig(cfg map[string]interface{}, requiredSR *SampleRate) *BaseAgent {
	clone := a.Clone()
	clone.Avatar = CloneConfig(cfg)
	if AvatarConfigEnabled(clone.Avatar) {
		clone.AvatarRequiredSampleRate = requiredSR
	} else {
		clone.AvatarRequiredSampleRate = nil
	}
	return clone
}

type turnDetectionLanguage string

const defaultTurnDetectionLanguage turnDetectionLanguage = "en-US"

func isTurnDetectionLanguage(language string) bool {
	_, err := Agora.NewAsrLanguageFromString(language)
	return err == nil
}

func ValidateTurnDetectionLanguage(language string) {
	if !isTurnDetectionLanguage(string(language)) {
		panic(fmt.Sprintf("invalid turn_detection.language: %s", language))
	}
}

func MapToStruct(m map[string]interface{}, target interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal config map: %w", err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal config into struct: %w", err)
	}
	return nil
}

func StructToMap(value interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func CloneConfig(config map[string]interface{}) map[string]interface{} {
	if config == nil {
		return nil
	}
	clone := make(map[string]interface{}, len(config))
	for k, v := range config {
		clone[k] = CloneValue(v)
	}
	return clone
}

func CloneValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		return CloneConfig(v)
	case []interface{}:
		clone := make([]interface{}, len(v))
		for i, item := range v {
			clone[i] = CloneValue(item)
		}
		return clone
	case []map[string]interface{}:
		clone := make([]map[string]interface{}, len(v))
		for i, item := range v {
			clone[i] = CloneConfig(item)
		}
		return clone
	case []string:
		if v == nil {
			return nil
		}
		// Use make so an empty non-nil slice stays non-nil (JSON []).
		clone := make([]string, len(v))
		copy(clone, v)
		return clone
	case []int:
		if v == nil {
			return nil
		}
		clone := make([]int, len(v))
		copy(clone, v)
		return clone
	case map[string]string:
		clone := make(map[string]string, len(v))
		for key, item := range v {
			clone[key] = item
		}
		return clone
	default:
		return value
	}
}

func BoolFromMap(m map[string]interface{}, key string) bool {
	if m == nil {
		return false
	}
	value, ok := m[key]
	if !ok {
		return false
	}
	b, ok := value.(bool)
	return ok && b
}

func HasNonEmptyString(m map[string]interface{}, key string) bool {
	if m == nil {
		return false
	}
	v, ok := m[key]
	if !ok {
		return false
	}
	s, ok := v.(string)
	return ok && s != ""
}

func AsMap(value interface{}) map[string]interface{} {
	m, _ := value.(map[string]interface{})
	return m
}

func SetStructMap(target map[string]interface{}, key string, value interface{}) error {
	valueMap, err := StructToMap(value)
	if err != nil {
		return fmt.Errorf("failed to convert %s config to map: %w", key, err)
	}
	target[key] = valueMap
	return nil
}

func AvatarConfigEnabled(avatar map[string]interface{}) bool {
	if avatar == nil {
		return false
	}
	enabled, ok := avatar["enable"].(bool)
	return !ok || enabled
}

func avatarUIDString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return fmt.Sprint(v)
	case int8:
		return fmt.Sprint(v)
	case int16:
		return fmt.Sprint(v)
	case int32:
		return fmt.Sprint(v)
	case int64:
		return fmt.Sprint(v)
	case uint:
		return fmt.Sprint(v)
	case uint8:
		return fmt.Sprint(v)
	case uint16:
		return fmt.Sprint(v)
	case uint32:
		return fmt.Sprint(v)
	case uint64:
		return fmt.Sprint(v)
	case float32:
		return fmt.Sprint(v)
	case float64:
		return fmt.Sprint(v)
	default:
		return ""
	}
}

func ParseUIDString(value interface{}) string {
	return avatarUIDString(value)
}

func ParseNumericUID(uid string, label string) (int, error) {
	value, err := strconv.Atoi(uid)
	if err != nil {
		return 0, fmt.Errorf("%s must be a numeric RTC UID when auto-generating a ConvoAI token", label)
	}
	return value, nil
}

func IsHeyGenAvatar(vendor string) bool {
	return vendor == "heygen"
}

func IsAkoolAvatar(vendor string) bool {
	return vendor == "akool"
}

func IsLiveAvatarAvatar(vendor string) bool {
	return vendor == "liveavatar"
}

func IsAnamAvatar(vendor string) bool {
	return vendor == "anam"
}

func IsSensetimeAvatar(vendor string) bool {
	return vendor == "sensetime"
}

func IsSpatiusAvatar(vendor string) bool {
	return vendor == "spatius"
}

func IsGenericAvatar(vendor string) bool {
	return vendor == "generic"
}

func IsAvatarTokenManaged(vendor string) bool {
	return IsHeyGenAvatar(vendor) || IsLiveAvatarAvatar(vendor) || IsGenericAvatar(vendor) || IsSensetimeAvatar(vendor) || IsSpatiusAvatar(vendor)
}

func ValidateAvatarConfig(vendor string, params map[string]interface{}) error {
	if IsHeyGenAvatar(vendor) || IsLiveAvatarAvatar(vendor) {
		label := "HeyGen"
		if IsLiveAvatarAvatar(vendor) {
			label = "LiveAvatar"
		}
		if params == nil {
			return fmt.Errorf("%s avatar requires params", label)
		}
		if !HasNonEmptyString(params, "api_key") {
			return fmt.Errorf("%s avatar requires api_key", label)
		}
		if q, ok := params["quality"]; !ok || !HasNonEmptyString(params, "quality") {
			return fmt.Errorf("%s avatar requires quality (low, medium, or high)", label)
		} else {
			qs, _ := q.(string)
			if qs != "low" && qs != "medium" && qs != "high" {
				return fmt.Errorf("invalid quality for %s: %v. Must be one of: low, medium, high", label, q)
			}
		}
		if avatarUIDString(params["agora_uid"]) == "" {
			return fmt.Errorf("%s avatar requires agora_uid", label)
		}
	} else if IsAkoolAvatar(vendor) {
		if params == nil {
			return fmt.Errorf("Akool avatar requires params")
		}
		if !HasNonEmptyString(params, "api_key") {
			return fmt.Errorf("Akool avatar requires api_key")
		}
	} else if IsAnamAvatar(vendor) {
		if params == nil {
			return fmt.Errorf("Anam avatar requires params")
		}
		if !HasNonEmptyString(params, "api_key") {
			return fmt.Errorf("Anam avatar requires api_key")
		}
	} else if IsSensetimeAvatar(vendor) {
		if params == nil {
			return fmt.Errorf("Sensetime avatar requires params")
		}
		if avatarUIDString(params["agora_uid"]) == "" {
			return fmt.Errorf("Sensetime avatar requires agora_uid")
		}
		if !HasNonEmptyString(params, "appId") {
			return fmt.Errorf("Sensetime avatar requires appId")
		}
		if !HasNonEmptyString(params, "app_key") {
			return fmt.Errorf("Sensetime avatar requires app_key")
		}
	} else if IsSpatiusAvatar(vendor) {
		if params == nil {
			return fmt.Errorf("Spatius avatar requires params")
		}
		if !HasNonEmptyString(params, "spatius_api_key") {
			return fmt.Errorf("Spatius avatar requires spatius_api_key")
		}
		if !HasNonEmptyString(params, "spatius_app_id") {
			return fmt.Errorf("Spatius avatar requires spatius_app_id")
		}
		if !HasNonEmptyString(params, "spatius_avatar_id") {
			return fmt.Errorf("Spatius avatar requires spatius_avatar_id")
		}
		if avatarUIDString(params["agora_uid"]) == "" {
			return fmt.Errorf("Spatius avatar requires agora_uid")
		}
	} else if IsGenericAvatar(vendor) {
		if params == nil {
			return fmt.Errorf("Generic avatar requires params")
		}
		if !HasNonEmptyString(params, "api_key") {
			return fmt.Errorf("Generic avatar requires api_key")
		}
		if !HasNonEmptyString(params, "api_base_url") {
			return fmt.Errorf("Generic avatar requires api_base_url")
		}
		if !HasNonEmptyString(params, "avatar_id") {
			return fmt.Errorf("Generic avatar requires avatar_id")
		}
		if avatarUIDString(params["agora_uid"]) == "" {
			return fmt.Errorf("Generic avatar requires agora_uid")
		}
	}
	return nil
}

func ValidateTtsSampleRate(avatarVendor string, sampleRate int) error {
	if IsHeyGenAvatar(avatarVendor) || IsLiveAvatarAvatar(avatarVendor) {
		label := "HeyGen"
		docURL := "https://docs.agora.io/en/conversational-ai/models/avatar/heygen"
		if IsLiveAvatarAvatar(avatarVendor) {
			label = "LiveAvatar"
			docURL = "https://docs.agora.io/en/conversational-ai/models/avatar/overview"
		}
		if sampleRate != 24000 {
			return fmt.Errorf(
				"%s avatars ONLY support 24,000 Hz sample rate. "+
					"Your TTS is configured with %d Hz. "+
					"Please update your TTS configuration to use 24kHz sample rate. "+
					"See: %s",
				label, sampleRate, docURL,
			)
		}
	} else if IsAkoolAvatar(avatarVendor) {
		if sampleRate != 16000 {
			return fmt.Errorf(
				"Akool avatars ONLY support 16,000 Hz sample rate. "+
					"Your TTS is configured with %d Hz. "+
					"Please update your TTS configuration to use 16kHz sample rate. "+
					"See: https://docs.agora.io/en/conversational-ai/models/avatar/akool",
				sampleRate,
			)
		}
	}
	return nil
}

func NewAreaRequestOption(area basecore.Area) *basecore.AreaRequestOption {
	return basecore.NewAreaRequestOption(area)
}
