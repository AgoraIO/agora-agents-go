package cn

import (
	Agora "github.com/AgoraIO/agora-agents-go/v2"
	base "github.com/AgoraIO/agora-agents-go/v2/agentkit"
	cnvendors "github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors"
	agentcore "github.com/AgoraIO/agora-agents-go/v2/agentkit/core"
)

type (
	AgentOption          = agentcore.AgentOption
	CreateSessionOptions = base.CreateSessionOptions
	AgentSession         = base.AgentSession
	AgentRuntime         = base.AgentRuntime
	ToPropertiesOptions  = base.ToPropertiesOptions
)

type (
	SalConfig                           = agentcore.SalConfig
	SalMode                             = agentcore.SalMode
	TurnDetectionConfig                 = agentcore.TurnDetectionConfig
	TurnDetectionNestedConfig           = agentcore.TurnDetectionNestedConfig
	StartOfSpeechConfig                 = agentcore.StartOfSpeechConfig
	StartOfSpeechMode                   = agentcore.StartOfSpeechMode
	StartOfSpeechVadConfig              = agentcore.StartOfSpeechVadConfig
	StartOfSpeechKeywordsConfig         = agentcore.StartOfSpeechKeywordsConfig
	StartOfSpeechDisabledConfig         = agentcore.StartOfSpeechDisabledConfig
	StartOfSpeechDisabledConfigStrategy = agentcore.StartOfSpeechDisabledConfigStrategy
	EndOfSpeechConfig                   = agentcore.EndOfSpeechConfig
	EndOfSpeechMode                     = agentcore.EndOfSpeechMode
	EndOfSpeechVadConfig                = agentcore.EndOfSpeechVadConfig
	EndOfSpeechSemanticConfig           = agentcore.EndOfSpeechSemanticConfig
	InterruptionConfig                  = agentcore.InterruptionConfig
	InterruptionMode                    = agentcore.InterruptionMode
	TurnDetectionType                   = agentcore.TurnDetectionType
	InterruptMode                       = agentcore.InterruptMode
	Eagerness                           = agentcore.Eagerness
	AdvancedFeatures                    = agentcore.AdvancedFeatures
	SessionParams                       = agentcore.SessionParams
	SessionParamsInput                  = agentcore.SessionParamsInput
	GeofenceConfig                      = agentcore.GeofenceConfig
	GeofenceArea                        = agentcore.GeofenceArea
	GeofenceExcludeArea                 = agentcore.GeofenceExcludeArea
	RtcConfig                           = agentcore.RtcConfig
	FillerWordsConfig                   = agentcore.FillerWordsConfig
	FillerWordsTrigger                  = agentcore.FillerWordsTrigger
	FillerWordsTriggerFixedTimeConfig   = agentcore.FillerWordsTriggerFixedTimeConfig
	FillerWordsContent                  = agentcore.FillerWordsContent
	FillerWordsContentMode              = agentcore.FillerWordsContentMode
	FillerWordsContentStaticConfig      = agentcore.FillerWordsContentStaticConfig
	FillerWordsContentSelectionRule     = agentcore.FillerWordsContentSelectionRule
	FillerWordsContentGeneratedConfig   = agentcore.FillerWordsContentGeneratedConfig
	FillerWordsGeneratedLlmProvider     = agentcore.FillerWordsGeneratedLlmProvider
	LlmTool                             = agentcore.LlmTool
	LlmToolFunction                     = agentcore.LlmToolFunction
	LlmToolFunctionParameters           = agentcore.LlmToolFunctionParameters
	LlmToolExecution                    = agentcore.LlmToolExecution
	LlmToolServer                       = agentcore.LlmToolServer
	LlmToolServerMethod                 = agentcore.LlmToolServerMethod
	LlmGreetingConfigs                  = agentcore.LlmGreetingConfigs
	LlmGreetingConfigsMode              = agentcore.LlmGreetingConfigsMode
	McpServersItem                      = agentcore.McpServersItem
	SilenceConfig                       = agentcore.SilenceConfig
	SilenceAction                       = agentcore.SilenceAction
	FarewellConfig                      = agentcore.FarewellConfig
	ParametersDataChannel               = agentcore.ParametersDataChannel
	ParametersAudioScenario             = agentcore.ParametersAudioScenario
	LlmConfig                           = agentcore.LlmConfig
	MllmConfig                          = agentcore.MllmConfig
	MllmTurnDetectionConfig             = agentcore.MllmTurnDetectionConfig
	MllmTurnDetectionMode               = agentcore.MllmTurnDetectionMode
	AsrConfig                           = agentcore.AsrConfig
	SttConfig                           = agentcore.SttConfig
	LlmStyle                            = agentcore.LlmStyle
	SttVendor                           = agentcore.SttVendor
	MllmVendor                          = agentcore.MllmVendor
	AvatarVendor                        = agentcore.AvatarVendor
	AgentConfig                         = agentcore.AgentConfig
	AgentConfigUpdate                   = agentcore.AgentConfigUpdate
	SessionInfo                         = agentcore.SessionInfo
	SessionListResponse                 = agentcore.SessionListResponse
	SessionSummary                      = agentcore.SessionSummary
	SessionStatus                       = agentcore.SessionStatus
	ConversationHistory                 = agentcore.ConversationHistory
	ConversationTurn                    = agentcore.ConversationTurn
	ConversationRole                    = agentcore.ConversationRole
	ConversationTurns                   = agentcore.ConversationTurns
	ConversationSessionTurn             = agentcore.ConversationSessionTurn
	ThinkResponse                       = agentcore.ThinkResponse
	ThinkOnListeningAction              = agentcore.ThinkOnListeningAction
	ThinkOnThinkingAction               = agentcore.ThinkOnThinkingAction
	ThinkOnSpeakingAction               = agentcore.ThinkOnSpeakingAction
	SpeakPriority                       = agentcore.SpeakPriority
	Labels                              = agentcore.Labels
	TtsConfig                           = agentcore.TtsConfig
	AvatarConfig                        = agentcore.AvatarConfig
)

type Agent struct {
	base *agentcore.BaseAgent
}

func NewAgent(client *AgoraClient, opts ...AgentOption) *Agent {
	if client == nil {
		panic("NewAgent requires AgoraClient")
	}
	baseAgent := agentcore.NewBaseAgent(opts...)
	baseAgent.Client = client
	return &Agent{base: baseAgent}
}

func WithPipelineID(id string) AgentOption { return agentcore.WithPipelineID(id) }
func WithInstructions(instructions string) AgentOption {
	return agentcore.WithInstructions(instructions)
}
func WithGreeting(greeting string) AgentOption  { return agentcore.WithGreeting(greeting) }
func WithFailureMessage(msg string) AgentOption { return agentcore.WithFailureMessage(msg) }
func WithMaxHistory(n int) AgentOption          { return agentcore.WithMaxHistory(n) }
func WithGreetingAudioURL(url string) AgentOption {
	return agentcore.WithGreetingAudioURL(url)
}
func WithSessionOptOut(optOut bool) AgentOption { return agentcore.WithSessionOptOut(optOut) }
func WithTurnDetectionConfig(cfg *TurnDetectionConfig) AgentOption {
	return agentcore.WithTurnDetectionConfig(cfg)
}

func WithInterruptionConfig(cfg *InterruptionConfig) AgentOption {
	return agentcore.WithInterruptionConfig(cfg)
}

func WithAdvancedFeatures(cfg *AdvancedFeatures) AgentOption {
	return agentcore.WithAdvancedFeatures(cfg)
}
func WithTools(enabled bool) AgentOption            { return agentcore.WithTools(enabled) }
func WithParameters(cfg *SessionParams) AgentOption { return agentcore.WithParameters(cfg) }
func WithAudioScenario(audioScenario ParametersAudioScenario) AgentOption {
	return agentcore.WithAudioScenario(audioScenario)
}
func WithGeofence(cfg *GeofenceConfig) AgentOption       { return agentcore.WithGeofence(cfg) }
func WithRtc(cfg *RtcConfig) AgentOption                 { return agentcore.WithRtc(cfg) }
func WithFillerWords(cfg *FillerWordsConfig) AgentOption { return agentcore.WithFillerWords(cfg) }

func (a *Agent) BaseAgent() *agentcore.BaseAgent { return a.base }
func (a *Agent) Profile() agentcore.Profile      { return agentcore.ProfileCN }
func (a *Agent) Client() agentcore.ClientRuntime {
	return a.base.Client
}

func (a *Agent) WithStt(vendor agentcore.STTVendor) *Agent {
	return &Agent{base: a.base.ApplySTTConfig(vendor.ToConfig())}
}

func (a *Agent) WithLlm(vendor agentcore.LLMVendor) *Agent {
	return &Agent{base: a.base.ApplyLLMConfig(vendor.ToConfig())}
}

func (a *Agent) WithTts(vendor agentcore.TTSVendor) *Agent {
	return &Agent{base: a.base.ApplyTTSConfig(vendor.ToConfig(), vendor.GetSampleRate())}
}

// WithMllm configures a mainland China MLLM vendor such as Qwen Omni.
func (a *Agent) WithMllm(vendor agentcore.MLLMVendor) *Agent {
	return &Agent{base: a.base.ApplyMLLMConfig(vendor.ToConfig())}
}

func (a *Agent) WithAvatar(vendor agentcore.AvatarVendorConfig) *Agent {
	requiredSR := agentcore.SampleRate(vendor.RequiredSampleRate())
	avatarConfig := vendor.ToConfig()
	if agentcore.IsAvatarTokenManaged(vendorName(avatarConfig)) && requiredSR != 0 && a.base.TTSSampleRate != nil && *a.base.TTSSampleRate != requiredSR {
		panic("Avatar requires a TTS sample rate compatible with the configured CN avatar vendor")
	}
	return &Agent{base: a.base.ApplyAvatarConfig(avatarConfig, &requiredSR)}
}

// CreateSession builds a session from the CN agent builder.
// Pass CreateSessionOptions.Name to set the /join agent instance identifier.
func (a *Agent) CreateSession(opts CreateSessionOptions) *AgentSession {
	return base.NewSession(a, opts)
}

func NewSession(agent *Agent, opts CreateSessionOptions) *AgentSession {
	return base.NewSession(agent, opts)
}

func (a *Agent) PipelineID() string                { return a.base.PipelineID }
func (a *Agent) Instructions() string              { return a.base.Instructions }
func (a *Agent) Greeting() string                  { return a.base.Greeting }
func (a *Agent) LlmConfig() map[string]interface{} { return a.base.LLM }
func (a *Agent) TtsConfig() map[string]interface{} { return a.base.TTS }
func (a *Agent) Stt() map[string]interface{}       { return a.base.STT }
func (a *Agent) SttConfig() map[string]interface{} { return a.Stt() }
func (a *Agent) MllmConfig() map[string]interface{} {
	return a.base.MLLM
}

func (a *Agent) TtsSampleRate() *cnvendors.SampleRate {
	if a.base.TTSSampleRate == nil {
		return nil
	}
	sr := cnvendors.SampleRate(*a.base.TTSSampleRate)
	return &sr
}

func (a *Agent) AvatarRequiredSampleRate() *cnvendors.SampleRate {
	if a.base.AvatarRequiredSampleRate == nil {
		return nil
	}
	sr := cnvendors.SampleRate(*a.base.AvatarRequiredSampleRate)
	return &sr
}
func (a *Agent) FailureMessage() string               { return a.base.FailureMessage }
func (a *Agent) MaxHistory() *int                     { return a.base.MaxHistory }
func (a *Agent) Avatar() map[string]interface{}       { return a.base.Avatar }
func (a *Agent) TurnDetection() *TurnDetectionConfig  { return a.base.TurnDetection }
func (a *Agent) Interruption() *InterruptionConfig    { return a.base.Interruption }
func (a *Agent) GreetingConfigs() *LlmGreetingConfigs { return a.base.GreetingConfigs }
func (a *Agent) Sal() *SalConfig                      { return a.base.Sal }
func (a *Agent) AdvancedFeatures() *AdvancedFeatures  { return a.base.AdvancedFeatures }
func (a *Agent) Parameters() *SessionParams           { return a.base.Parameters }
func (a *Agent) Geofence() *GeofenceConfig            { return a.base.Geofence }
func (a *Agent) Labels() map[string]string            { return a.base.Labels }
func (a *Agent) Rtc() *RtcConfig                      { return a.base.RTC }
func (a *Agent) FillerWords() *FillerWordsConfig      { return a.base.FillerWords }

func (a *Agent) ToProperties(opts ToPropertiesOptions) (*Agora.StartAgentsRequestProperties, error) {
	propsMap, err := a.ToPropertiesMap(opts)
	if err != nil {
		return nil, err
	}
	var props Agora.StartAgentsRequestProperties
	if err := agentcore.MapToStruct(propsMap, &props); err != nil {
		return nil, err
	}
	return &props, nil
}

func (a *Agent) ToPropertiesMap(opts ToPropertiesOptions) (map[string]interface{}, error) {
	return agentcore.BuildPropertiesMap(agentcore.ProfileCN, a.base, agentcore.ToPropertiesOptions(opts), func(coreOpts agentcore.GenerateConvoAITokenOptions) (string, error) {
		return base.GenerateConvoAIToken(base.GenerateConvoAITokenOptions(coreOpts))
	})
}

func ExpiresInHours(hours float64) (int, error) {
	return base.ExpiresInHours(hours)
}

func ExpiresInMinutes(minutes float64) (int, error) {
	return base.ExpiresInMinutes(minutes)
}

func vendorName(config map[string]interface{}) string {
	if config == nil {
		return ""
	}
	vendor, _ := config["vendor"].(string)
	return vendor
}
