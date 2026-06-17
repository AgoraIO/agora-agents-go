package agentkit

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
)

type turnDetectionLanguage string

const defaultTurnDetectionLanguage turnDetectionLanguage = "en-US"

func isTurnDetectionLanguage(language string) bool {
	_, err := Agora.NewAsrLanguageFromString(language)
	return err == nil
}

func validateTurnDetectionLanguage(language string) {
	if !isTurnDetectionLanguage(string(language)) {
		panic(fmt.Sprintf("invalid turn_detection.language: %s", language))
	}
}

func mapToStruct(m map[string]interface{}, target interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal config map: %w", err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal config into struct: %w", err)
	}
	return nil
}

func structToMap(value interface{}) (map[string]interface{}, error) {
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

func cloneConfig(config map[string]interface{}) map[string]interface{} {
	if config == nil {
		return nil
	}
	clone := make(map[string]interface{}, len(config))
	for k, v := range config {
		clone[k] = cloneValue(v)
	}
	return clone
}

func cloneValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		return cloneConfig(v)
	case []interface{}:
		clone := make([]interface{}, len(v))
		for i, item := range v {
			clone[i] = cloneValue(item)
		}
		return clone
	case []map[string]interface{}:
		clone := make([]map[string]interface{}, len(v))
		for i, item := range v {
			clone[i] = cloneConfig(item)
		}
		return clone
	case []string:
		return append([]string(nil), v...)
	case []int:
		return append([]int(nil), v...)
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

func boolFromMap(m map[string]interface{}, key string) bool {
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

// =============================================================================
// Top-level configuration aliases
// =============================================================================

type TurnDetectionConfig = Agora.StartAgentsRequestPropertiesTurnDetection
type SalConfig = Agora.StartAgentsRequestPropertiesSal
type SalMode = Agora.StartAgentsRequestPropertiesSalSalMode
type AdvancedFeatures = Agora.StartAgentsRequestPropertiesAdvancedFeatures
type SessionParams = Agora.StartAgentsRequestPropertiesParameters

// SessionParamsInput is an alias for SessionParams. Use WithAudioScenario for
// ergonomic audio scenario configuration.
type SessionParamsInput = SessionParams
type GeofenceConfig = Agora.StartAgentsRequestPropertiesGeofence
type RtcConfig = Agora.StartAgentsRequestPropertiesRtc
type FillerWordsConfig = Agora.StartAgentsRequestPropertiesFillerWords

// =============================================================================
// SOS/EOS turn detection aliases (preferred — replaces deprecated types below)
// =============================================================================

// TurnDetectionNestedConfig is the detailed nested config within TurnDetectionConfig.Config.
type TurnDetectionNestedConfig = Agora.StartAgentsRequestPropertiesTurnDetectionConfig

// StartOfSpeechConfig configures when the agent detects the start of a user's speech.
type StartOfSpeechConfig = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeech

// StartOfSpeechMode is the detection mode: "vad" | "keywords" | "disabled".
type StartOfSpeechMode = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeechMode

// StartOfSpeechVadConfig holds VAD settings for SoS detection.
type StartOfSpeechVadConfig = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeechVadConfig

// StartOfSpeechKeywordsConfig holds keyword-trigger settings for SoS detection.
type StartOfSpeechKeywordsConfig = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeechKeywordsConfig

// StartOfSpeechDisabledConfig holds settings when SoS detection is disabled.
type StartOfSpeechDisabledConfig = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeechDisabledConfig

// StartOfSpeechDisabledConfigStrategy is the voice processing strategy when SoS is disabled: "append" | "ignored".
type StartOfSpeechDisabledConfigStrategy = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeechDisabledConfigStrategy

// EndOfSpeechConfig configures when the agent detects the end of a user's speech.
type EndOfSpeechConfig = Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeech

// EndOfSpeechMode is the detection mode: "vad" | "semantic".
type EndOfSpeechMode = Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeechMode

// EndOfSpeechVadConfig holds VAD settings for EoS detection.
type EndOfSpeechVadConfig = Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeechVadConfig

// EndOfSpeechSemanticConfig holds semantic model settings for EoS detection.
type EndOfSpeechSemanticConfig = Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeechSemanticConfig

// InterruptionConfig configures unified interruption handling (top-level `interruption`).
type InterruptionConfig = Agora.StartAgentsRequestPropertiesInterruption

// InterruptionMode controls interruption trigger mode: "start_of_speech" | "keywords".
type InterruptionMode = Agora.StartAgentsRequestPropertiesInterruptionMode

// =============================================================================
// Deprecated turn detection aliases
// =============================================================================

// Deprecated: Use TurnDetectionConfig with TurnDetectionNestedConfig.StartOfSpeech and
// TurnDetectionNestedConfig.EndOfSpeech instead. The Type field and agora_vad / server_vad /
// semantic_vad values are being removed in a future release.
type TurnDetectionType = Agora.StartAgentsRequestPropertiesTurnDetectionType

// Deprecated: Use StartOfSpeechConfig with Mode "vad" | "keywords" | "disabled" and the
// corresponding VadConfig, KeywordsConfig, or DisabledConfig instead.
type InterruptMode = Agora.StartAgentsRequestPropertiesTurnDetectionInterruptMode

// Deprecated: Only applies to server_vad / semantic_vad modes with OpenAI Realtime API (MLLM).
// Has no equivalent in the standard ASR + LLM + TTS pipeline.
type Eagerness = Agora.StartAgentsRequestPropertiesTurnDetectionEagerness

// =============================================================================
// LLM sub-type aliases
// =============================================================================

// LlmGreetingConfigs configures how greeting messages are broadcast when multiple
// remote users are in the channel (llm.greeting_configs).
type LlmGreetingConfigs = map[string]interface{}

// LlmGreetingConfigsMode is the greeting broadcast mode: "single_every" | "single_first".
type LlmGreetingConfigsMode = string

// McpServersItem is a single MCP server config entry (llm.mcp_servers[]).
type McpServersItem = map[string]interface{}

// =============================================================================
// Parameters (SessionParams) sub-type aliases
// =============================================================================

// SilenceConfig configures agent behaviour during user silence (parameters.silence_config).
type SilenceConfig = Agora.StartAgentsRequestPropertiesParametersSilenceConfig

// SilenceAction is the action taken after the silence timeout elapses.
type SilenceAction = Agora.StartAgentsRequestPropertiesParametersSilenceConfigAction

// FarewellConfig configures graceful hang-up behaviour (parameters.farewell_config).
type FarewellConfig = Agora.StartAgentsRequestPropertiesParametersFarewellConfig

// ParametersDataChannel is the agent data-transmission channel: "rtm" | "datastream".
type ParametersDataChannel = Agora.StartAgentsRequestPropertiesParametersDataChannel

// ParametersAudioScenario is the RTC audio scenario used by the agent session.
type ParametersAudioScenario string

const (
	ParametersAudioScenarioDefault  ParametersAudioScenario = "default"
	ParametersAudioScenarioChorus   ParametersAudioScenario = "chorus"
	ParametersAudioScenarioAIServer ParametersAudioScenario = "aiserver"
)

// =============================================================================
// Geofence sub-type aliases
// =============================================================================

// GeofenceArea is an allowed geographic region for server access.
type GeofenceArea = Agora.StartAgentsRequestPropertiesGeofenceArea

// GeofenceExcludeArea is a geographic region to exclude when Area is "GLOBAL".
type GeofenceExcludeArea = Agora.StartAgentsRequestPropertiesGeofenceExcludeArea

// =============================================================================
// Concrete API payload config aliases (for constructing or inspecting ToProperties output)
// =============================================================================

// LlmConfig is the concrete LLM configuration payload (start_agents_request_properties.llm).
type LlmConfig = Agora.Llm

// MllmConfig is the concrete MLLM configuration payload (start_agents_request_properties.mllm).
type MllmConfig = Agora.Mllm

// MllmTurnDetectionConfig configures MLLM turn detection (`mllm.turn_detection`).
type MllmTurnDetectionConfig = Agora.MllmTurnDetection

// MllmTurnDetectionMode controls MLLM turn detection mode.
type MllmTurnDetectionMode = Agora.MllmTurnDetectionMode

// AsrConfig is the concrete ASR/STT configuration payload (start_agents_request_properties.asr).
type AsrConfig = Agora.Asr

// SttConfig is an alias for AsrConfig (wire field remains `asr`).
type SttConfig = AsrConfig

// LlmStyle is the LLM request style (openai, gemini, anthropic, dify).
type LlmStyle = Agora.LlmStyle

// SttVendor is the ASR/STT vendor identifier.
type SttVendor = string

// MllmVendor is the MLLM vendor identifier.
type MllmVendor = Agora.MllmVendor

// AvatarVendor is the avatar vendor identifier.
type AvatarVendor = Agora.StartAgentsRequestPropertiesAvatarVendor

// AgentConfig is the full agent start payload (start_agents_request_properties).
type AgentConfig = Agora.StartAgentsRequestProperties

// AgentConfigUpdate is the agent update payload (update_agents_request_properties).
type AgentConfigUpdate = Agora.UpdateAgentsRequestProperties

// SessionInfo is the response from GetAgents.
type SessionInfo = Agora.GetAgentsResponse

// SessionListResponse is the response from ListAgents.
type SessionListResponse = Agora.ListAgentsResponse

// SessionSummary is a single entry in a session list response.
type SessionSummary = Agora.ListAgentsResponseDataListItem

// SessionStatus is the API list-item status returned by session list responses.
type SessionStatus = Agora.ListAgentsResponseDataListItemStatus

// ConversationHistory is the response from GetHistoryAgents.
type ConversationHistory = Agora.GetHistoryAgentsResponse

// ConversationTurn is a single turn in conversation history.
type ConversationTurn = Agora.GetHistoryAgentsResponseContentsItem

// ConversationRole is the role of a participant in a conversation turn.
type ConversationRole = Agora.GetHistoryAgentsResponseContentsItemRole

// ConversationTurns is the response from GetTurnsAgents.
type ConversationTurns = Agora.GetTurnsAgentsResponse

// ConversationSessionTurn is a single turn in a paginated turns response.
type ConversationSessionTurn = Agora.GetTurnsAgentsResponseTurnsItem

// ThinkResponse is the response from AgentThink.
type ThinkResponse = Agora.AgentThinkAgentManagementResponse

// ThinkOnListeningAction is the action when the agent is listening.
type ThinkOnListeningAction = Agora.AgentThinkAgentManagementRequestOnListeningAction

// ThinkOnThinkingAction is the action when the agent is thinking.
type ThinkOnThinkingAction = Agora.AgentThinkAgentManagementRequestOnThinkingAction

// ThinkOnSpeakingAction is the action when the agent is speaking.
type ThinkOnSpeakingAction = Agora.AgentThinkAgentManagementRequestOnSpeakingAction

// SpeakPriority is the priority for speak requests.
type SpeakPriority = Agora.SpeakAgentsRequestPriority

// Labels is a string map of session metadata labels.
type Labels = map[string]string

// TtsConfig is the concrete TTS configuration payload (start_agents_request_properties.tts).
type TtsConfig = Agora.Tts

// AvatarConfig is the concrete Avatar configuration payload (start_agents_request_properties.avatar).
type AvatarConfig = Agora.StartAgentsRequestPropertiesAvatar

// =============================================================================
// FillerWords sub-type aliases
// =============================================================================

// FillerWordsTrigger configures when filler words are played (filler_words.trigger).
type FillerWordsTrigger = Agora.StartAgentsRequestPropertiesFillerWordsTrigger

// FillerWordsTriggerFixedTimeConfig holds the fixed-time trigger threshold (trigger.fixed_time_config).
type FillerWordsTriggerFixedTimeConfig = Agora.StartAgentsRequestPropertiesFillerWordsTriggerFixedTimeConfig

// FillerWordsContent configures the source and selection of filler words (filler_words.content).
type FillerWordsContent = Agora.StartAgentsRequestPropertiesFillerWordsContent

// FillerWordsContentStaticConfig configures a static list of filler words (content.static_config).
type FillerWordsContentStaticConfig = Agora.StartAgentsRequestPropertiesFillerWordsContentStaticConfig

// FillerWordsContentSelectionRule is the filler word selection rule: "shuffle" | "round_robin".
type FillerWordsContentSelectionRule = Agora.StartAgentsRequestPropertiesFillerWordsContentStaticConfigSelectionRule

type Agent struct {
	name                     string
	pipelineID               string
	instructions             string
	greeting                 string
	failureMessage           string
	maxHistory               *int
	llm                      map[string]interface{}
	tts                      map[string]interface{}
	stt                      map[string]interface{}
	mllm                     map[string]interface{}
	ttsSampleRate            *vendors.SampleRate
	avatar                   map[string]interface{}
	avatarRequiredSampleRate *vendors.SampleRate
	turnDetection            *TurnDetectionConfig
	interruption             *InterruptionConfig
	greetingConfigs          *LlmGreetingConfigs
	sal                      *SalConfig
	advancedFeatures         *AdvancedFeatures
	parameters               *SessionParams
	audioScenario            *ParametersAudioScenario
	geofence                 *GeofenceConfig
	labels                   map[string]string
	rtc                      *RtcConfig
	fillerWords              *FillerWordsConfig
}

type AgentOption func(*Agent)

func NewAgent(opts ...AgentOption) *Agent {
	a := &Agent{}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func WithName(name string) AgentOption {
	return func(a *Agent) {
		a.name = name
	}
}

// WithPipelineID sets the published AI Studio pipeline ID to use as this agent's
// base configuration. Explicit Agent config such as WithLlm, WithTts, WithStt,
// advanced features, and other builder options may send fields in properties
// that override the saved pipeline settings.
func WithPipelineID(pipelineID string) AgentOption {
	return func(a *Agent) {
		a.pipelineID = pipelineID
	}
}

// Deprecated: Configure system messages on the LLM vendor instead.
func WithInstructions(instructions string) AgentOption {
	return func(a *Agent) {
		a.instructions = instructions
	}
}

// Deprecated: Configure the greeting on the LLM or MLLM vendor instead.
func WithGreeting(greeting string) AgentOption {
	return func(a *Agent) {
		a.greeting = greeting
	}
}

// Deprecated: Configure the failure message on the LLM or MLLM vendor instead.
func WithFailureMessage(msg string) AgentOption {
	return func(a *Agent) {
		a.failureMessage = msg
	}
}

// Deprecated: Configure max history on the LLM vendor instead.
func WithMaxHistory(n int) AgentOption {
	return func(a *Agent) {
		a.maxHistory = &n
	}
}

func WithTurnDetectionConfig(td *TurnDetectionConfig) AgentOption {
	return func(a *Agent) {
		a.turnDetection = td
	}
}

func WithInterruptionConfig(interruption *InterruptionConfig) AgentOption {
	return func(a *Agent) {
		a.interruption = interruption
	}
}

// Deprecated: Configure greeting playback on the LLM vendor instead.
func WithGreetingConfigs(configs *LlmGreetingConfigs) AgentOption {
	return func(a *Agent) {
		a.greetingConfigs = configs
	}
}

func WithSalConfig(sal *SalConfig) AgentOption {
	return func(a *Agent) {
		a.sal = sal
	}
}

func WithAdvancedFeatures(af *AdvancedFeatures) AgentOption {
	return func(a *Agent) {
		a.advancedFeatures = af
	}
}

func WithTools(enabled bool) AgentOption {
	return func(a *Agent) {
		if a.advancedFeatures == nil {
			a.advancedFeatures = &AdvancedFeatures{}
		}
		a.advancedFeatures.EnableTools = &enabled
	}
}

func WithParameters(params *SessionParams) AgentOption {
	return func(a *Agent) {
		a.parameters = params
	}
}

func WithAudioScenario(audioScenario ParametersAudioScenario) AgentOption {
	return func(a *Agent) {
		a.audioScenario = &audioScenario
	}
}

func WithGeofence(gf *GeofenceConfig) AgentOption {
	return func(a *Agent) {
		a.geofence = gf
	}
}

func WithLabels(labels map[string]string) AgentOption {
	return func(a *Agent) {
		a.labels = labels
	}
}

func WithRtc(rtc *RtcConfig) AgentOption {
	return func(a *Agent) {
		a.rtc = rtc
	}
}

func WithFillerWords(fw *FillerWordsConfig) AgentOption {
	return func(a *Agent) {
		a.fillerWords = fw
	}
}

func (a *Agent) WithLlm(vendor vendors.LLM) *Agent {
	clone := a.clone()
	clone.llm = vendor.ToConfig()
	return clone
}

func (a *Agent) WithTts(vendor vendors.TTS) *Agent {
	clone := a.clone()
	clone.tts = vendor.ToConfig()
	clone.ttsSampleRate = vendor.GetSampleRate()
	// If an avatar is already set, verify the new TTS sample rate matches.
	// Mirrors the check in WithAvatar so both call orderings fail fast.
	if clone.avatarRequiredSampleRate != nil && *clone.avatarRequiredSampleRate != 0 && clone.ttsSampleRate != nil {
		if *clone.ttsSampleRate != *clone.avatarRequiredSampleRate {
			panic(fmt.Sprintf(
				"TTS sample rate %d Hz is incompatible with the configured avatar, which requires %d Hz. "+
					"Please update your TTS sample_rate to %d.",
				int(*clone.ttsSampleRate), int(*clone.avatarRequiredSampleRate), int(*clone.avatarRequiredSampleRate),
			))
		}
	}
	return clone
}

func (a *Agent) WithStt(vendor vendors.STT) *Agent {
	clone := a.clone()
	clone.stt = vendor.ToConfig()
	return clone
}

func (a *Agent) WithMllm(vendor vendors.MLLM) *Agent {
	clone := a.clone()
	clone.mllm = vendor.ToConfig()
	if clone.mllm != nil {
		clone.mllm["enable"] = true
	}
	if clone.advancedFeatures != nil {
		clone.advancedFeatures.EnableMllm = nil
		if clone.advancedFeatures.EnableRtm == nil && clone.advancedFeatures.EnableSal == nil && clone.advancedFeatures.EnableTools == nil {
			clone.advancedFeatures = nil
		}
	}
	return clone
}

func (a *Agent) WithAvatar(vendor vendors.Avatar) *Agent {
	requiredSR := vendor.RequiredSampleRate()
	avatarConfig := vendor.ToConfig()
	// If a TTS is already set, verify sample rate compatibility now.
	// Mirrors the check in WithTts so both call orderings fail fast.
	// AgentSession.Start also validates as a final safety net.
	if avatarConfigEnabled(avatarConfig) && requiredSR != 0 && a.ttsSampleRate != nil && *a.ttsSampleRate != requiredSR {
		panic(fmt.Sprintf(
			"Avatar requires TTS sample rate of %d Hz, but TTS is configured with %d Hz. "+
				"Please update your TTS sample_rate to %d.",
			int(requiredSR), int(*a.ttsSampleRate), int(requiredSR),
		))
	}
	clone := a.clone()
	clone.avatar = avatarConfig
	if avatarConfigEnabled(avatarConfig) {
		clone.avatarRequiredSampleRate = &requiredSR
	} else {
		clone.avatarRequiredSampleRate = nil
	}
	return clone
}

func (a *Agent) WithTurnDetection(td *TurnDetectionConfig) *Agent {
	clone := a.clone()
	clone.turnDetection = td
	return clone
}

func (a *Agent) WithInterruption(interruption *InterruptionConfig) *Agent {
	clone := a.clone()
	clone.interruption = interruption
	return clone
}

// Deprecated: Configure greeting playback on the LLM vendor instead.
func (a *Agent) WithGreetingConfigs(configs *LlmGreetingConfigs) *Agent {
	clone := a.clone()
	clone.greetingConfigs = configs
	return clone
}

// Deprecated: Configure system messages on the LLM vendor instead.
func (a *Agent) WithInstructions(instructions string) *Agent {
	clone := a.clone()
	clone.instructions = instructions
	return clone
}

// Deprecated: Configure the greeting on the LLM or MLLM vendor instead.
func (a *Agent) WithGreeting(greeting string) *Agent {
	clone := a.clone()
	clone.greeting = greeting
	return clone
}

func (a *Agent) WithName(name string) *Agent {
	clone := a.clone()
	clone.name = name
	return clone
}

func (a *Agent) WithSal(sal *SalConfig) *Agent {
	clone := a.clone()
	clone.sal = sal
	return clone
}

func (a *Agent) WithAdvancedFeatures(af *AdvancedFeatures) *Agent {
	clone := a.clone()
	clone.advancedFeatures = af
	return clone
}

func (a *Agent) WithTools(enabled bool) *Agent {
	clone := a.clone()
	if clone.advancedFeatures == nil {
		clone.advancedFeatures = &AdvancedFeatures{}
	} else {
		advancedFeatures := *clone.advancedFeatures
		clone.advancedFeatures = &advancedFeatures
	}
	clone.advancedFeatures.EnableTools = &enabled
	return clone
}

func (a *Agent) WithParameters(params *SessionParams) *Agent {
	clone := a.clone()
	clone.parameters = params
	return clone
}

func (a *Agent) WithAudioScenario(audioScenario ParametersAudioScenario) *Agent {
	clone := a.clone()
	clone.audioScenario = &audioScenario
	return clone
}

// Deprecated: Configure the failure message on the LLM or MLLM vendor instead.
func (a *Agent) WithFailureMessage(msg string) *Agent {
	clone := a.clone()
	clone.failureMessage = msg
	return clone
}

// Deprecated: Configure max history on the LLM vendor instead.
func (a *Agent) WithMaxHistory(n int) *Agent {
	clone := a.clone()
	clone.maxHistory = &n
	return clone
}

func (a *Agent) WithGeofence(gf *GeofenceConfig) *Agent {
	clone := a.clone()
	clone.geofence = gf
	return clone
}

func (a *Agent) WithLabels(labels map[string]string) *Agent {
	clone := a.clone()
	clone.labels = labels
	return clone
}

func (a *Agent) WithRtc(rtc *RtcConfig) *Agent {
	clone := a.clone()
	clone.rtc = rtc
	return clone
}

func (a *Agent) WithFillerWords(fw *FillerWordsConfig) *Agent {
	clone := a.clone()
	clone.fillerWords = fw
	return clone
}

func (a *Agent) Name() string                      { return a.name }
func (a *Agent) PipelineID() string                { return a.pipelineID }
func (a *Agent) Instructions() string              { return a.instructions }
func (a *Agent) Greeting() string                  { return a.greeting }
func (a *Agent) LlmConfig() map[string]interface{} { return a.llm }
func (a *Agent) TtsConfig() map[string]interface{} { return a.tts }
func (a *Agent) Stt() map[string]interface{}       { return a.stt }

// Deprecated: Use Stt.
func (a *Agent) SttConfig() map[string]interface{}             { return a.Stt() }
func (a *Agent) MllmConfig() map[string]interface{}            { return a.mllm }
func (a *Agent) TtsSampleRate() *vendors.SampleRate            { return a.ttsSampleRate }
func (a *Agent) AvatarRequiredSampleRate() *vendors.SampleRate { return a.avatarRequiredSampleRate }
func (a *Agent) FailureMessage() string                        { return a.failureMessage }
func (a *Agent) MaxHistory() *int                              { return a.maxHistory }
func (a *Agent) Avatar() map[string]interface{}                { return a.avatar }
func (a *Agent) TurnDetection() *TurnDetectionConfig           { return a.turnDetection }
func (a *Agent) Interruption() *InterruptionConfig             { return a.interruption }
func (a *Agent) GreetingConfigs() *LlmGreetingConfigs          { return a.greetingConfigs }
func (a *Agent) Sal() *SalConfig                               { return a.sal }
func (a *Agent) AdvancedFeatures() *AdvancedFeatures           { return a.advancedFeatures }
func (a *Agent) Parameters() *SessionParams                    { return a.parameters }
func (a *Agent) Geofence() *GeofenceConfig                     { return a.geofence }
func (a *Agent) Labels() map[string]string                     { return a.labels }
func (a *Agent) Rtc() *RtcConfig                               { return a.rtc }
func (a *Agent) FillerWords() *FillerWordsConfig               { return a.fillerWords }

type CreateSessionOptions struct {
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

func (a *Agent) CreateSession(client *AgoraClient, opts CreateSessionOptions) *AgentSession {
	name := opts.Name
	if name == "" {
		if a.name != "" {
			name = a.name
		} else {
			name = fmt.Sprintf("agent-%d", time.Now().UnixMilli())
		}
	}

	return NewAgentSession(AgentSessionOptions{
		Client:                   client.Agents,
		AgentManagementClient:    client.AgentManagement,
		Agent:                    a,
		AppID:                    client.AppID,
		AppCertificate:           client.AppCertificate,
		Name:                     name,
		Channel:                  opts.Channel,
		Token:                    opts.Token,
		AgentUID:                 opts.AgentUID,
		RemoteUIDs:               opts.RemoteUIDs,
		IdleTimeout:              opts.IdleTimeout,
		EnableStringUID:          opts.EnableStringUID,
		ExpiresIn:                opts.ExpiresIn,
		UseAppCredentialsForREST: client.AuthMode == AuthModeAppCredentials,
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
	if err := mapToStruct(propsMap, &props); err != nil {
		return nil, fmt.Errorf("failed to convert properties map: %w", err)
	}
	return &props, nil
}

func (a *Agent) ToPropertiesMap(opts ToPropertiesOptions) (map[string]interface{}, error) {
	// Reject incompatible combinations before any work (token generation, etc.).
	// Avatars are currently supported only with the cascading ASR/LLM/TTS pipeline.
	if a.mllm != nil && a.hasEnabledAvatar() {
		return nil, fmt.Errorf("avatar is only supported with cascading ASR/LLM/TTS sessions; remove the avatar configuration when using MLLM")
	}

	expiry := opts.ExpiresIn
	if expiry != 0 {
		var err error
		expiry, err = ValidateExpiresIn(expiry)
		if err != nil {
			return nil, fmt.Errorf("invalid expiresIn: %w", err)
		}
	}
	opts.ExpiresIn = expiry

	token := opts.Token
	if token == "" {
		if opts.AppID == "" || opts.AppCertificate == "" {
			return nil, fmt.Errorf("either token or app_id+app_certificate must be provided")
		}
		uid, err := parseNumericUID(opts.AgentUID, "agent UID")
		if err != nil {
			return nil, err
		}
		token, err = GenerateConvoAIToken(GenerateConvoAITokenOptions{
			AppID:          opts.AppID,
			AppCertificate: opts.AppCertificate,
			ChannelName:    opts.Channel,
			UID:            uid,
			TokenExpire:    expiry,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to generate token: %w", err)
		}
	}

	if len(opts.RemoteUIDs) == 0 {
		return nil, fmt.Errorf("AgentSessionOptions.RemoteUIDs is required and must contain at least one UID")
	}

	propsMap := map[string]interface{}{
		"channel":         opts.Channel,
		"token":           token,
		"agent_rtc_uid":   opts.AgentUID,
		"remote_rtc_uids": opts.RemoteUIDs,
	}
	if opts.IdleTimeout != nil {
		propsMap["idle_timeout"] = *opts.IdleTimeout
	}
	if opts.EnableStringUID != nil {
		propsMap["enable_string_uid"] = *opts.EnableStringUID
	}
	if a.mllm != nil {
		propsMap["mllm"] = a.buildMllmConfigMap()
	}
	if a.interruption != nil {
		if err := setStructMap(propsMap, "interruption", a.interruption); err != nil {
			return nil, err
		}
	}
	if a.sal != nil {
		if err := setStructMap(propsMap, "sal", a.sal); err != nil {
			return nil, err
		}
	}
	if a.avatar != nil {
		avatar, err := a.enrichAvatarParams(opts)
		if err != nil {
			return nil, err
		}
		propsMap["avatar"] = avatar
	}
	if a.advancedFeatures != nil {
		if err := setStructMap(propsMap, "advanced_features", a.advancedFeatures); err != nil {
			return nil, err
		}
	}
	if a.parameters != nil {
		if err := setStructMap(propsMap, "parameters", a.parameters); err != nil {
			return nil, err
		}
	}
	if a.audioScenario != nil {
		parameters, ok := propsMap["parameters"].(map[string]interface{})
		if !ok || parameters == nil {
			parameters = map[string]interface{}{}
			propsMap["parameters"] = parameters
		}
		parameters["audio_scenario"] = string(*a.audioScenario)
	}
	if a.geofence != nil {
		if err := setStructMap(propsMap, "geofence", a.geofence); err != nil {
			return nil, err
		}
	}
	if len(a.labels) > 0 {
		propsMap["labels"] = cloneValue(a.labels)
	}
	if a.rtc != nil {
		if err := setStructMap(propsMap, "rtc", a.rtc); err != nil {
			return nil, err
		}
	}
	if a.fillerWords != nil {
		if err := setStructMap(propsMap, "filler_words", a.fillerWords); err != nil {
			return nil, err
		}
	}

	if a.advancedFeatures != nil && a.advancedFeatures.EnableRtm != nil && *a.advancedFeatures.EnableRtm {
		parameters, ok := propsMap["parameters"].(map[string]interface{})
		if !ok || parameters == nil {
			parameters = map[string]interface{}{}
			propsMap["parameters"] = parameters
		}
		if _, exists := parameters["data_channel"]; !exists {
			parameters["data_channel"] = "rtm"
		}
	}

	if a.mllm != nil {
		if a.turnDetection != nil {
			if err := setStructMap(propsMap, "turn_detection", a.turnDetection); err != nil {
				return nil, err
			}
		}
		return propsMap, nil
	}

	skipCategories := map[string]bool{}
	for _, category := range opts.SkipVendorValidationCategories {
		skipCategories[category] = true
	}
	allowMissingCategories := map[string]bool{}
	for _, category := range opts.AllowMissingVendorCategories {
		allowMissingCategories[category] = true
	}
	if opts.SkipVendorValidation {
		for _, category := range []string{"asr", "llm", "tts"} {
			skipCategories[category] = true
			allowMissingCategories[category] = true
		}
	}

	turnDetection, err := a.resolveTurnDetectionConfig()
	if err != nil {
		return nil, err
	}
	if a.stt != nil || !allowMissingCategories["asr"] {
		propsMap["asr"] = a.resolveAsrConfig(turnDetection)
	}
	propsMap["turn_detection"] = turnDetection

	if a.tts == nil && !skipCategories["tts"] && !allowMissingCategories["tts"] {
		return nil, fmt.Errorf("TTS configuration is required; use WithTts() to set it")
	}
	if a.llm == nil && !skipCategories["llm"] && !allowMissingCategories["llm"] {
		return nil, fmt.Errorf("LLM configuration is required; use WithLlm() to set it")
	}

	if a.llm != nil {
		propsMap["llm"] = a.buildLlmConfigMap()
	}
	if a.tts != nil {
		propsMap["tts"] = cloneConfig(a.tts)
	}

	return propsMap, nil
}

func (a *Agent) resolveAsrConfig(turnDetection map[string]interface{}) map[string]interface{} {
	asrConfig := cloneConfig(a.stt)
	if asrConfig == nil {
		asrConfig = map[string]interface{}{"vendor": "ares"}
	}
	if len(asrConfig) == 0 {
		asrConfig["vendor"] = "ares"
	}
	asrConfig["language"] = turnDetection["language"]
	return asrConfig
}

func (a *Agent) resolveTurnDetectionConfig() (map[string]interface{}, error) {
	turnDetection := map[string]interface{}{}
	if a.turnDetection != nil {
		var err error
		turnDetection, err = structToMap(a.turnDetection)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize turn detection config: %w", err)
		}
	}

	language := ""
	if existing, ok := turnDetection["language"].(string); ok && existing != "" {
		language = existing
	}
	if language == "" {
		language = string(defaultTurnDetectionLanguage)
	}
	validateTurnDetectionLanguage(language)
	turnDetection["language"] = language
	return turnDetection, nil
}

func (a *Agent) buildMllmConfigMap() map[string]interface{} {
	mllmConfig := cloneConfig(a.mllm)
	if a.greeting != "" {
		if _, exists := mllmConfig["greeting_message"]; !exists {
			mllmConfig["greeting_message"] = a.greeting
		}
	}
	if a.failureMessage != "" {
		if _, exists := mllmConfig["failure_message"]; !exists {
			mllmConfig["failure_message"] = a.failureMessage
		}
	}
	return mllmConfig
}

func (a *Agent) buildLlmConfigMap() map[string]interface{} {
	llmConfig := cloneConfig(a.llm)
	if a.instructions != "" {
		if _, exists := llmConfig["system_messages"]; !exists {
			llmConfig["system_messages"] = []map[string]interface{}{
				{"role": "system", "content": a.instructions},
			}
		}
	}
	if a.greeting != "" {
		if _, exists := llmConfig["greeting_message"]; !exists {
			llmConfig["greeting_message"] = a.greeting
		}
	}
	if a.failureMessage != "" {
		if _, exists := llmConfig["failure_message"]; !exists {
			llmConfig["failure_message"] = a.failureMessage
		}
	}
	if a.maxHistory != nil {
		if _, exists := llmConfig["max_history"]; !exists {
			llmConfig["max_history"] = *a.maxHistory
		}
	}
	if a.greetingConfigs != nil {
		if _, exists := llmConfig["greeting_configs"]; !exists {
			if value, err := structToMap(a.greetingConfigs); err == nil {
				llmConfig["greeting_configs"] = value
			} else {
				llmConfig["greeting_configs"] = a.greetingConfigs
			}
		}
	}
	return llmConfig
}

func setStructMap(target map[string]interface{}, key string, value interface{}) error {
	valueMap, err := structToMap(value)
	if err != nil {
		return fmt.Errorf("failed to convert %s config to map: %w", key, err)
	}
	target[key] = valueMap
	return nil
}

func (a *Agent) enrichAvatarParams(opts ToPropertiesOptions) (map[string]interface{}, error) {
	avatar := cloneConfig(a.avatar)
	if !avatarConfigEnabled(avatar) {
		return avatar, nil
	}
	vendor, _ := avatar["vendor"].(string)
	params, _ := avatar["params"].(map[string]interface{})
	if params == nil {
		params = map[string]interface{}{}
		avatar["params"] = params
	}

	if IsGenericAvatar(vendor) {
		if _, exists := params["agora_appid"]; !exists && opts.AppID != "" {
			params["agora_appid"] = opts.AppID
		}
		if _, exists := params["agora_channel"]; !exists && opts.Channel != "" {
			params["agora_channel"] = opts.Channel
		}
	}

	avatarUID := avatarUIDString(params["agora_uid"])
	if IsAvatarTokenManaged(vendor) && avatarUID != "" {
		if avatarUID == opts.AgentUID && opts.Warn != nil {
			opts.Warn("avatar agora_uid matches agent_rtc_uid; use a distinct UID so the avatar video stream does not collide with the voice agent")
		}
		if token, _ := params["agora_token"].(string); token == "" {
			if opts.AppCertificate == "" {
				return nil, fmt.Errorf("cannot auto-generate avatar agora_token: appCertificate is required; pass AppCertificate when creating AgoraClient, or set AgoraToken on the avatar vendor")
			}
			uid, err := parseNumericUID(avatarUID, "avatar agora_uid")
			if err != nil {
				return nil, err
			}
			generated, err := GenerateConvoAIToken(GenerateConvoAITokenOptions{
				AppID:          opts.AppID,
				AppCertificate: opts.AppCertificate,
				ChannelName:    opts.Channel,
				UID:            uid,
				TokenExpire:    opts.ExpiresIn,
			})
			if err != nil {
				return nil, err
			}
			params["agora_token"] = generated
		}
	}

	return avatar, nil
}

func IsAvatarTokenManaged(vendor string) bool {
	return IsHeyGenAvatar(vendor) || IsLiveAvatarAvatar(vendor) || IsGenericAvatar(vendor)
}

func IsAvatarTokenManagedFromConfig(avatar map[string]interface{}) bool {
	vendor, _ := avatar["vendor"].(string)
	return IsAvatarTokenManaged(vendor)
}

func avatarConfigEnabled(avatar map[string]interface{}) bool {
	if avatar == nil {
		return false
	}
	enabled, ok := avatar["enable"].(bool)
	return !ok || enabled
}

func (a *Agent) hasEnabledAvatar() bool {
	return a != nil && a.avatar != nil && avatarConfigEnabled(a.avatar)
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

func parseNumericUID(uid string, label string) (int, error) {
	value, err := strconv.Atoi(uid)
	if err != nil {
		return 0, fmt.Errorf("%s must be a numeric RTC UID when auto-generating a ConvoAI token", label)
	}
	return value, nil
}

type ToPropertiesOptions struct {
	Channel        string
	AgentUID       string
	RemoteUIDs     []string
	Token          string
	AppID          string
	AppCertificate string
	// ExpiresIn is the token lifetime in seconds (default: 86400 = 24 hours, Agora maximum).
	// Valid range: 1–86400. Use ExpiresInHours() / ExpiresInMinutes() for clarity.
	ExpiresIn       int
	IdleTimeout     *int
	EnableStringUID *bool
	// Deprecated: use SkipVendorValidationCategories and AllowMissingVendorCategories instead.
	SkipVendorValidation           bool
	SkipVendorValidationCategories []string
	AllowMissingVendorCategories   []string
	Warn                           func(string)
}

func (a *Agent) clone() *Agent {
	clone := *a
	if a.labels != nil {
		clone.labels = make(map[string]string, len(a.labels))
		for k, v := range a.labels {
			clone.labels[k] = v
		}
	}
	return &clone
}
