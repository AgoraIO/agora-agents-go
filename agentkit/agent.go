package agentkit

import (
	"fmt"
	"time"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	agentcore "github.com/AgoraIO/agora-agents-go/v2/agentkit/core"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
)

type TurnDetectionConfig = agentcore.TurnDetectionConfig
type SalConfig = agentcore.SalConfig
type SalMode = agentcore.SalMode
type AdvancedFeatures = agentcore.AdvancedFeatures
type SessionParams = agentcore.SessionParams
type SessionParamsInput = agentcore.SessionParamsInput
type GeofenceConfig = agentcore.GeofenceConfig
type RtcConfig = agentcore.RtcConfig
type FillerWordsConfig = agentcore.FillerWordsConfig
type TurnDetectionNestedConfig = agentcore.TurnDetectionNestedConfig
type StartOfSpeechConfig = agentcore.StartOfSpeechConfig
type StartOfSpeechMode = agentcore.StartOfSpeechMode
type StartOfSpeechVadConfig = agentcore.StartOfSpeechVadConfig
type StartOfSpeechKeywordsConfig = agentcore.StartOfSpeechKeywordsConfig
type StartOfSpeechDisabledConfig = agentcore.StartOfSpeechDisabledConfig
type StartOfSpeechDisabledConfigStrategy = agentcore.StartOfSpeechDisabledConfigStrategy
type EndOfSpeechConfig = agentcore.EndOfSpeechConfig
type EndOfSpeechMode = agentcore.EndOfSpeechMode
type EndOfSpeechVadConfig = agentcore.EndOfSpeechVadConfig
type EndOfSpeechSemanticConfig = agentcore.EndOfSpeechSemanticConfig
type InterruptionConfig = agentcore.InterruptionConfig
type InterruptionMode = agentcore.InterruptionMode
type TurnDetectionType = agentcore.TurnDetectionType
type InterruptMode = agentcore.InterruptMode
type Eagerness = agentcore.Eagerness
type LlmGreetingConfigs = agentcore.LlmGreetingConfigs
type LlmGreetingConfigsMode = agentcore.LlmGreetingConfigsMode
type McpServersItem = agentcore.McpServersItem
type SilenceConfig = agentcore.SilenceConfig
type SilenceAction = agentcore.SilenceAction
type FarewellConfig = agentcore.FarewellConfig
type ParametersDataChannel = agentcore.ParametersDataChannel
type ParametersAudioScenario = agentcore.ParametersAudioScenario
type GeofenceArea = agentcore.GeofenceArea
type GeofenceExcludeArea = agentcore.GeofenceExcludeArea
type LlmConfig = agentcore.LlmConfig
type MllmConfig = agentcore.MllmConfig
type MllmTurnDetectionConfig = agentcore.MllmTurnDetectionConfig
type MllmTurnDetectionMode = agentcore.MllmTurnDetectionMode
type AsrConfig = agentcore.AsrConfig
type SttConfig = agentcore.SttConfig
type LlmStyle = agentcore.LlmStyle
type SttVendor = agentcore.SttVendor
type MllmVendor = agentcore.MllmVendor
type AvatarVendor = agentcore.AvatarVendor
type AgentConfig = agentcore.AgentConfig
type AgentConfigUpdate = agentcore.AgentConfigUpdate
type SessionInfo = agentcore.SessionInfo
type SessionListResponse = agentcore.SessionListResponse
type SessionSummary = agentcore.SessionSummary
type SessionStatus = agentcore.SessionStatus
type ConversationHistory = agentcore.ConversationHistory
type ConversationTurn = agentcore.ConversationTurn
type ConversationRole = agentcore.ConversationRole
type ConversationTurns = agentcore.ConversationTurns
type ConversationSessionTurn = agentcore.ConversationSessionTurn
type ThinkResponse = agentcore.ThinkResponse
type ThinkOnListeningAction = agentcore.ThinkOnListeningAction
type ThinkOnThinkingAction = agentcore.ThinkOnThinkingAction
type ThinkOnSpeakingAction = agentcore.ThinkOnSpeakingAction
type SpeakPriority = agentcore.SpeakPriority
type Labels = agentcore.Labels
type TtsConfig = agentcore.TtsConfig
type AvatarConfig = agentcore.AvatarConfig
type FillerWordsTrigger = agentcore.FillerWordsTrigger
type FillerWordsTriggerFixedTimeConfig = agentcore.FillerWordsTriggerFixedTimeConfig
type FillerWordsContent = agentcore.FillerWordsContent
type FillerWordsContentMode = agentcore.FillerWordsContentMode
type FillerWordsContentStaticConfig = agentcore.FillerWordsContentStaticConfig
type FillerWordsContentSelectionRule = agentcore.FillerWordsContentSelectionRule
type FillerWordsContentGeneratedConfig = agentcore.FillerWordsContentGeneratedConfig
type FillerWordsGeneratedLlmProvider = agentcore.FillerWordsGeneratedLlmProvider
type GeminiAsrConfig = agentcore.GeminiAsrConfig
type GeminiAsrParams = agentcore.GeminiAsrParams
type LlmTool = agentcore.LlmTool
type LlmToolFunction = agentcore.LlmToolFunction
type LlmToolFunctionParameters = agentcore.LlmToolFunctionParameters
type LlmToolExecution = agentcore.LlmToolExecution
type LlmToolServer = agentcore.LlmToolServer
type LlmToolServerMethod = agentcore.LlmToolServerMethod

const (
	ParametersAudioScenarioDefault  ParametersAudioScenario = agentcore.ParametersAudioScenarioDefault
	ParametersAudioScenarioChorus   ParametersAudioScenario = agentcore.ParametersAudioScenarioChorus
	ParametersAudioScenarioAIServer ParametersAudioScenario = agentcore.ParametersAudioScenarioAIServer
)

type AgentOption = agentcore.AgentOption
type ToPropertiesOptions = agentcore.ToPropertiesOptions
type AgentRuntime = agentcore.AgentRuntime

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

func WithPipelineID(pipelineID string) AgentOption { return agentcore.WithPipelineID(pipelineID) }
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
func WithTurnDetectionConfig(td *TurnDetectionConfig) AgentOption {
	return agentcore.WithTurnDetectionConfig(td)
}
func WithInterruptionConfig(interruption *InterruptionConfig) AgentOption {
	return agentcore.WithInterruptionConfig(interruption)
}
func WithGreetingConfigs(configs *LlmGreetingConfigs) AgentOption {
	return agentcore.WithGreetingConfigs(configs)
}
func WithSalConfig(sal *SalConfig) AgentOption { return agentcore.WithSalConfig(sal) }
func WithAdvancedFeatures(af *AdvancedFeatures) AgentOption {
	return agentcore.WithAdvancedFeatures(af)
}
func WithTools(enabled bool) AgentOption               { return agentcore.WithTools(enabled) }
func WithParameters(params *SessionParams) AgentOption { return agentcore.WithParameters(params) }
func WithAudioScenario(audioScenario ParametersAudioScenario) AgentOption {
	return agentcore.WithAudioScenario(audioScenario)
}
func WithGeofence(gf *GeofenceConfig) AgentOption       { return agentcore.WithGeofence(gf) }
func WithLabels(labels map[string]string) AgentOption   { return agentcore.WithLabels(labels) }
func WithRtc(rtc *RtcConfig) AgentOption                { return agentcore.WithRtc(rtc) }
func WithFillerWords(fw *FillerWordsConfig) AgentOption { return agentcore.WithFillerWords(fw) }

func (a *Agent) BaseAgent() *agentcore.BaseAgent { return a.base }
func (a *Agent) Profile() agentcore.Profile      { return agentcore.ProfileGlobal }
func (a *Agent) Client() agentcore.ClientRuntime {
	return a.base.Client
}

func (a *Agent) WithLlm(vendor agentcore.LLMVendor) *Agent {
	return &Agent{base: a.base.ApplyLLMConfig(vendor.ToConfig())}
}

func (a *Agent) WithTts(vendor agentcore.TTSVendor) *Agent {
	var sr *agentcore.SampleRate
	if current := vendor.GetSampleRate(); current != nil {
		converted := agentcore.SampleRate(*current)
		sr = &converted
	}
	clone := a.base.ApplyTTSConfig(vendor.ToConfig(), sr)
	if clone.AvatarRequiredSampleRate != nil && *clone.AvatarRequiredSampleRate != 0 && clone.TTSSampleRate != nil {
		if *clone.TTSSampleRate != *clone.AvatarRequiredSampleRate {
			panic(fmt.Sprintf(
				"TTS sample rate %d Hz is incompatible with the configured avatar, which requires %d Hz. "+
					"Please update your TTS sample_rate to %d.",
				int(*clone.TTSSampleRate), int(*clone.AvatarRequiredSampleRate), int(*clone.AvatarRequiredSampleRate),
			))
		}
	}
	return &Agent{base: clone}
}

func (a *Agent) WithStt(vendor agentcore.STTVendor) *Agent {
	return &Agent{base: a.base.ApplySTTConfig(vendor.ToConfig())}
}

func (a *Agent) WithMllm(vendor agentcore.MLLMVendor) *Agent {
	return &Agent{base: a.base.ApplyMLLMConfig(vendor.ToConfig())}
}

func (a *Agent) WithAvatar(vendor agentcore.AvatarVendorConfig) *Agent {
	requiredSR := agentcore.SampleRate(vendor.RequiredSampleRate())
	avatarConfig := vendor.ToConfig()
	if agentcore.IsAvatarTokenManaged(vendorName(avatarConfig)) && requiredSR != 0 && a.base.TTSSampleRate != nil && *a.base.TTSSampleRate != requiredSR {
		panic(fmt.Sprintf(
			"Avatar requires TTS sample rate of %d Hz, but TTS is configured with %d Hz. "+
				"Please update your TTS sample_rate to %d.",
			int(requiredSR), int(*a.base.TTSSampleRate), int(requiredSR),
		))
	}
	return &Agent{base: a.base.ApplyAvatarConfig(avatarConfig, &requiredSR)}
}

func (a *Agent) WithTurnDetection(td *TurnDetectionConfig) *Agent {
	clone := a.base.Clone()
	clone.TurnDetection = td
	return &Agent{base: clone}
}

func (a *Agent) WithInterruption(interruption *InterruptionConfig) *Agent {
	clone := a.base.Clone()
	clone.Interruption = interruption
	return &Agent{base: clone}
}

func (a *Agent) WithGreetingConfigs(configs *LlmGreetingConfigs) *Agent {
	clone := a.base.Clone()
	clone.GreetingConfigs = configs
	return &Agent{base: clone}
}

func (a *Agent) WithInstructions(instructions string) *Agent {
	clone := a.base.Clone()
	clone.Instructions = instructions
	return &Agent{base: clone}
}

func (a *Agent) WithGreeting(greeting string) *Agent {
	clone := a.base.Clone()
	clone.Greeting = greeting
	return &Agent{base: clone}
}

func (a *Agent) WithGreetingAudioURL(url string) *Agent {
	clone := a.base.Clone()
	if clone.LLM == nil {
		clone.LLM = map[string]interface{}{}
	}
	clone.LLM["greeting_audio_url"] = url
	return &Agent{base: clone}
}

func (a *Agent) WithSal(sal *SalConfig) *Agent {
	clone := a.base.Clone()
	clone.Sal = sal
	return &Agent{base: clone}
}

func (a *Agent) WithAdvancedFeatures(af *AdvancedFeatures) *Agent {
	clone := a.base.Clone()
	clone.AdvancedFeatures = af
	return &Agent{base: clone}
}

func (a *Agent) WithTools(enabled bool) *Agent {
	clone := a.base.Clone()
	if clone.AdvancedFeatures == nil {
		clone.AdvancedFeatures = &AdvancedFeatures{}
	} else {
		advancedFeatures := *clone.AdvancedFeatures
		clone.AdvancedFeatures = &advancedFeatures
	}
	clone.AdvancedFeatures.EnableTools = &enabled
	return &Agent{base: clone}
}

func (a *Agent) WithParameters(params *SessionParams) *Agent {
	clone := a.base.Clone()
	clone.Parameters = params
	return &Agent{base: clone}
}

func (a *Agent) WithSessionOptOut(optOut bool) *Agent {
	clone := a.base.Clone()
	if clone.Parameters == nil {
		clone.Parameters = &SessionParams{}
	}
	clone.Parameters.OptOut = &optOut
	return &Agent{base: clone}
}

func (a *Agent) WithAudioScenario(audioScenario ParametersAudioScenario) *Agent {
	clone := a.base.Clone()
	clone.AudioScenario = &audioScenario
	return &Agent{base: clone}
}

func (a *Agent) WithFailureMessage(msg string) *Agent {
	clone := a.base.Clone()
	clone.FailureMessage = msg
	return &Agent{base: clone}
}

func (a *Agent) WithMaxHistory(n int) *Agent {
	clone := a.base.Clone()
	clone.MaxHistory = &n
	return &Agent{base: clone}
}

func (a *Agent) WithGeofence(gf *GeofenceConfig) *Agent {
	clone := a.base.Clone()
	clone.Geofence = gf
	return &Agent{base: clone}
}

func (a *Agent) WithLabels(labels map[string]string) *Agent {
	clone := a.base.Clone()
	clone.Labels = labels
	return &Agent{base: clone}
}

func (a *Agent) WithRtc(rtc *RtcConfig) *Agent {
	clone := a.base.Clone()
	clone.RTC = rtc
	return &Agent{base: clone}
}

func (a *Agent) WithFillerWords(fw *FillerWordsConfig) *Agent {
	clone := a.base.Clone()
	clone.FillerWords = fw
	return &Agent{base: clone}
}

func (a *Agent) PipelineID() string                { return a.base.PipelineID }
func (a *Agent) Instructions() string              { return a.base.Instructions }
func (a *Agent) Greeting() string                  { return a.base.Greeting }
func (a *Agent) LlmConfig() map[string]interface{} { return a.base.LLM }
func (a *Agent) TtsConfig() map[string]interface{} { return a.base.TTS }
func (a *Agent) Stt() map[string]interface{}       { return a.base.STT }

func (a *Agent) SttConfig() map[string]interface{}  { return a.Stt() }
func (a *Agent) MllmConfig() map[string]interface{} { return a.base.MLLM }
func (a *Agent) TtsSampleRate() *vendors.SampleRate {
	if a.base.TTSSampleRate == nil {
		return nil
	}
	sr := vendors.SampleRate(*a.base.TTSSampleRate)
	return &sr
}
func (a *Agent) AvatarRequiredSampleRate() *vendors.SampleRate {
	if a.base.AvatarRequiredSampleRate == nil {
		return nil
	}
	sr := vendors.SampleRate(*a.base.AvatarRequiredSampleRate)
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

// CreateSessionOptions configures a new AgentSession before Start.
type CreateSessionOptions struct {
	// Name is the agent instance identifier sent as the top-level /join "name" field.
	// When empty, CreateSession generates agent-<unix_timestamp>.
	Name            string
	Channel         string
	Token           string
	AgentUID        string
	RemoteUIDs      []string
	IdleTimeout     *int
	EnableStringUID *bool
	ExpiresIn       int
	Preset          []string
	PipelineID      string
	Debug           bool
	Warn            func(string)
}

// CreateSession builds an AgentSession from this agent configuration.
// Set Name on opts to identify the agent instance; omit Name to auto-generate one.
func (a *Agent) CreateSession(opts CreateSessionOptions) *AgentSession {
	return NewSession(a, opts)
}

// NewSession creates an AgentSession from any AgentRuntime implementation.
func NewSession(agent agentcore.AgentRuntime, opts CreateSessionOptions) *AgentSession {
	clientProvider, ok := agent.(interface {
		Client() agentcore.ClientRuntime
	})
	if !ok || clientProvider.Client() == nil {
		panic("agent must be bound to an AgoraClient before creating a session")
	}
	client := clientProvider.Client()
	name := opts.Name
	if name == "" {
		name = fmt.Sprintf("agent-%d", time.Now().UnixMilli())
	}

	return NewAgentSession(AgentSessionOptions{
		Client:                   client.AgentsClient(),
		AgentManagementClient:    client.AgentManagementClient(),
		Agent:                    agent,
		AppID:                    client.AppID(),
		AppCertificate:           client.AppCertificate(),
		Name:                     name,
		Channel:                  opts.Channel,
		Token:                    opts.Token,
		AgentUID:                 opts.AgentUID,
		RemoteUIDs:               opts.RemoteUIDs,
		IdleTimeout:              opts.IdleTimeout,
		EnableStringUID:          opts.EnableStringUID,
		ExpiresIn:                opts.ExpiresIn,
		UseAppCredentialsForREST: client.IsAppCredentialsMode(),
		Preset:                   opts.Preset,
		PipelineID:               opts.PipelineID,
		Debug:                    opts.Debug,
		Warn:                     opts.Warn,
	})
}

func (a *Agent) ToProperties(opts ToPropertiesOptions) (*Agora.StartAgentsRequestProperties, error) {
	propsMap, err := a.ToPropertiesMap(opts)
	if err != nil {
		return nil, err
	}
	var props Agora.StartAgentsRequestProperties
	if err := agentcore.MapToStruct(propsMap, &props); err != nil {
		return nil, fmt.Errorf("failed to convert properties map: %w", err)
	}
	return &props, nil
}

func (a *Agent) ToPropertiesMap(opts ToPropertiesOptions) (map[string]interface{}, error) {
	return BuildPropertiesMap(a.base, opts, GenerateConvoAIToken)
}

func BuildPropertiesMap(base *agentcore.BaseAgent, opts ToPropertiesOptions, tokenFactory func(GenerateConvoAITokenOptions) (string, error)) (map[string]interface{}, error) {
	return agentcore.BuildPropertiesMap(agentcore.ProfileGlobal, base, agentcore.ToPropertiesOptions(opts), func(coreOpts agentcore.GenerateConvoAITokenOptions) (string, error) {
		return tokenFactory(GenerateConvoAITokenOptions(coreOpts))
	})
}

func (a *Agent) hasEnabledAvatar() bool {
	return a != nil && a.base.Avatar != nil && agentcore.AvatarConfigEnabled(a.base.Avatar)
}

func vendorName(config map[string]interface{}) string {
	if config == nil {
		return ""
	}
	vendor, _ := config["vendor"].(string)
	return vendor
}
