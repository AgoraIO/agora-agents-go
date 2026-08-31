package core

import Agora "github.com/AgoraIO/agora-agents-go/v2"

type Profile string

const (
	ProfileGlobal Profile = "global"
	ProfileCN     Profile = "cn"
)

// CredentialMode identifies how credentials are supplied to a provider.
type CredentialMode string

const (
	// CredentialModeManaged uses credentials managed by Agora.
	CredentialModeManaged CredentialMode = "managed"
	// CredentialModeBYOK uses credentials supplied by the caller.
	CredentialModeBYOK CredentialMode = "byok"
)

// TurnDetectionConfig configures conversation turn detection.
//
// AgentKit owns this type rather than aliasing the generated SDK struct so the
// agentkit-specific `language` field (which drives ASR language; see
// resolveTurnDetectionConfig) survives SDK regeneration. The remaining fields
// mirror Agora.StartAgentsRequestPropertiesTurnDetection. AgentKit serializes
// this config to a JSON map (see StructToMap) when building the join payload, so
// only JSON tags are required.
type TurnDetectionConfig struct {
	// Language is the BCP-47 language tag identifying the primary language used
	// for agent interaction. AgentKit-only field; not present in the generated SDK.
	Language            *Agora.AsrLanguage                                            `json:"language,omitempty"`
	Mode                *string                                                       `json:"mode,omitempty"`
	Config              *Agora.StartAgentsRequestPropertiesTurnDetectionConfig        `json:"config,omitempty"`
	Type                *Agora.StartAgentsRequestPropertiesTurnDetectionType          `json:"type,omitempty"`
	InterruptMode       *Agora.StartAgentsRequestPropertiesTurnDetectionInterruptMode `json:"interrupt_mode,omitempty"`
	InterruptDurationMs *float64                                                      `json:"interrupt_duration_ms,omitempty"`
	InterruptKeywords   []string                                                      `json:"interrupt_keywords,omitempty"`
	PrefixPaddingMs     *int                                                          `json:"prefix_padding_ms,omitempty"`
	SilenceDurationMs   *int                                                          `json:"silence_duration_ms,omitempty"`
	Threshold           *float64                                                      `json:"threshold,omitempty"`
	CreateResponse      *bool                                                         `json:"create_response,omitempty"`
	InterruptResponse   *bool                                                         `json:"interrupt_response,omitempty"`
	Eagerness           *Agora.StartAgentsRequestPropertiesTurnDetectionEagerness     `json:"eagerness,omitempty"`
}
type (
	SalConfig          = Agora.StartAgentsRequestPropertiesSal
	SalMode            = Agora.StartAgentsRequestPropertiesSalSalMode
	AdvancedFeatures   = Agora.StartAgentsRequestPropertiesAdvancedFeatures
	SessionParams      = Agora.StartAgentsRequestPropertiesParameters
	SessionParamsInput = SessionParams
	GeofenceConfig     = Agora.StartAgentsRequestPropertiesGeofence
	RtcConfig          = Agora.StartAgentsRequestPropertiesRtc
	FillerWordsConfig  = Agora.StartAgentsRequestPropertiesFillerWords
)

type (
	TurnDetectionNestedConfig           = Agora.StartAgentsRequestPropertiesTurnDetectionConfig
	StartOfSpeechConfig                 = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeech
	StartOfSpeechMode                   = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeechMode
	StartOfSpeechVadConfig              = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeechVadConfig
	StartOfSpeechKeywordsConfig         = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeechKeywordsConfig
	StartOfSpeechDisabledConfig         = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeechDisabledConfig
	StartOfSpeechDisabledConfigStrategy = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeechDisabledConfigStrategy
	EndOfSpeechConfig                   = Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeech
	EndOfSpeechMode                     = Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeechMode
	EndOfSpeechVadConfig                = Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeechVadConfig
	EndOfSpeechSemanticConfig           = Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeechSemanticConfig
	InterruptionConfig                  = Agora.StartAgentsRequestPropertiesInterruption
	InterruptionMode                    = Agora.StartAgentsRequestPropertiesInterruptionMode
)

type (
	TurnDetectionType = Agora.StartAgentsRequestPropertiesTurnDetectionType
	InterruptMode     = Agora.StartAgentsRequestPropertiesTurnDetectionInterruptMode
	Eagerness         = Agora.StartAgentsRequestPropertiesTurnDetectionEagerness
)

type (
	LlmGreetingConfigs     = map[string]interface{}
	LlmGreetingConfigsMode = string
	McpServersItem         = map[string]interface{}
)

type (
	SilenceConfig         = Agora.StartAgentsRequestPropertiesParametersSilenceConfig
	SilenceAction         = Agora.StartAgentsRequestPropertiesParametersSilenceConfigAction
	FarewellConfig        = Agora.StartAgentsRequestPropertiesParametersFarewellConfig
	ParametersDataChannel = Agora.StartAgentsRequestPropertiesParametersDataChannel
)

type ParametersAudioScenario string

const (
	ParametersAudioScenarioDefault  ParametersAudioScenario = "default"
	ParametersAudioScenarioChorus   ParametersAudioScenario = "chorus"
	ParametersAudioScenarioAIServer ParametersAudioScenario = "aiserver"
)

type (
	GeofenceArea        = Agora.StartAgentsRequestPropertiesGeofenceArea
	GeofenceExcludeArea = Agora.StartAgentsRequestPropertiesGeofenceExcludeArea
)

type (
	LlmConfig               = Agora.Llm
	MllmConfig              = Agora.Mllm
	MllmTurnDetectionConfig = Agora.MllmTurnDetection
	MllmTurnDetectionMode   = Agora.MllmTurnDetectionMode
	AsrConfig               = Agora.Asr
	SttConfig               = AsrConfig
	LlmStyle                = Agora.LlmStyle
	SttVendor               = string
	MllmVendor              = Agora.MllmVendor
	AvatarVendor            = Agora.StartAgentsRequestPropertiesAvatarVendor
	AgentConfig             = Agora.StartAgentsRequestProperties
	AgentConfigUpdate       = Agora.UpdateAgentsRequestProperties
	SessionInfo             = Agora.GetAgentsResponse
	SessionListResponse     = Agora.ListAgentsResponse
	SessionSummary          = Agora.ListAgentsResponseDataListItem
	SessionStatus           = Agora.ListAgentsResponseDataListItemStatus
	ConversationHistory     = Agora.GetHistoryAgentsResponse
	ConversationTurn        = Agora.GetHistoryAgentsResponseContentsItem
	ConversationRole        = Agora.GetHistoryAgentsResponseContentsItemRole
	ConversationTurns       = Agora.GetTurnsAgentsResponse
	ConversationSessionTurn = Agora.GetTurnsAgentsResponseTurnsItem
	ThinkResponse           = Agora.AgentThinkAgentManagementResponse
	ThinkOnListeningAction  = Agora.AgentThinkAgentManagementRequestOnListeningAction
	ThinkOnThinkingAction   = Agora.AgentThinkAgentManagementRequestOnThinkingAction
	ThinkOnSpeakingAction   = Agora.AgentThinkAgentManagementRequestOnSpeakingAction
	SpeakPriority           = Agora.SpeakAgentsRequestPriority
	Labels                  = map[string]string
	TtsConfig               = Agora.Tts
	AvatarConfig            = Agora.StartAgentsRequestPropertiesAvatar
)

type (
	FillerWordsTrigger                     = Agora.StartAgentsRequestPropertiesFillerWordsTrigger
	FillerWordsTriggerFixedTimeConfig      = Agora.StartAgentsRequestPropertiesFillerWordsTriggerFixedTimeConfig
	FillerWordsContent                     = Agora.StartAgentsRequestPropertiesFillerWordsContent
	FillerWordsContentMode                 = Agora.StartAgentsRequestPropertiesFillerWordsContentMode
	FillerWordsContentStaticConfig         = Agora.StartAgentsRequestPropertiesFillerWordsContentStaticConfig
	FillerWordsContentSelectionRule        = Agora.StartAgentsRequestPropertiesFillerWordsContentStaticConfigSelectionRule
	FillerWordsContentGeneratedConfig      = Agora.StartAgentsRequestPropertiesFillerWordsContentGeneratedConfig
	FillerWordsContentGeneratedLlmProvider = Agora.StartAgentsRequestPropertiesFillerWordsContentGeneratedConfigLlmProvider
)

type (
	LlmTool                   = Agora.LlmTool
	LlmToolFunction           = Agora.LlmToolFunction
	LlmToolFunctionParameters = Agora.LlmToolFunctionParameters
	LlmToolExecution          = Agora.LlmToolExecution
	LlmToolServer             = Agora.LlmToolServer
	LlmToolServerMethod       = Agora.LlmToolServerMethod
)
