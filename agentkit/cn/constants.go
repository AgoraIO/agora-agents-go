package cn

import base "github.com/AgoraIO/agora-agents-go/v2/agentkit"

var (
	SalModeLocking     = base.SalModeLocking
	SalModeRecognition = base.SalModeRecognition
)

var (
	DataChannelRtm        = base.DataChannelRtm
	DataChannelDatastream = base.DataChannelDatastream
)

var (
	AudioScenarioDefault  = base.AudioScenarioDefault
	AudioScenarioChorus   = base.AudioScenarioChorus
	AudioScenarioAIServer = base.AudioScenarioAIServer
)

var (
	SilenceActionSpeak = base.SilenceActionSpeak
	SilenceActionThink = base.SilenceActionThink
)

var (
	GeofenceAreaGlobal       = base.GeofenceAreaGlobal
	GeofenceAreaNorthAmerica = base.GeofenceAreaNorthAmerica
	GeofenceAreaEurope       = base.GeofenceAreaEurope
	GeofenceAreaAsia         = base.GeofenceAreaAsia
	GeofenceAreaIndia        = base.GeofenceAreaIndia
	GeofenceAreaJapan        = base.GeofenceAreaJapan

	GeofenceExcludeAreaNorthAmerica = base.GeofenceExcludeAreaNorthAmerica
	GeofenceExcludeAreaEurope       = base.GeofenceExcludeAreaEurope
	GeofenceExcludeAreaAsia         = base.GeofenceExcludeAreaAsia
	GeofenceExcludeAreaIndia        = base.GeofenceExcludeAreaIndia
	GeofenceExcludeAreaJapan        = base.GeofenceExcludeAreaJapan
)

var (
	TurnDetectionTypeAgoraVad    = base.TurnDetectionTypeAgoraVad
	TurnDetectionTypeServerVad   = base.TurnDetectionTypeServerVad
	TurnDetectionTypeSemanticVad = base.TurnDetectionTypeSemanticVad
)

var (
	FillerWordsSelectionRuleShuffle    = base.FillerWordsSelectionRuleShuffle
	FillerWordsSelectionRuleRoundRobin = base.FillerWordsSelectionRuleRoundRobin
	FillerWordsContentModeStatic       = base.FillerWordsContentModeStatic
	FillerWordsContentModeGenerated    = base.FillerWordsContentModeGenerated
)

var (
	ThinkOnListeningActionInject    = base.ThinkOnListeningActionInject
	ThinkOnListeningActionInterrupt = base.ThinkOnListeningActionInterrupt
	ThinkOnListeningActionIgnore    = base.ThinkOnListeningActionIgnore
	ThinkOnListeningActionAppend    = base.ThinkOnListeningActionAppend
	ThinkOnThinkingActionInterrupt  = base.ThinkOnThinkingActionInterrupt
	ThinkOnThinkingActionIgnore     = base.ThinkOnThinkingActionIgnore
	ThinkOnThinkingActionAppend     = base.ThinkOnThinkingActionAppend
	ThinkOnSpeakingActionInterrupt  = base.ThinkOnSpeakingActionInterrupt
	ThinkOnSpeakingActionIgnore     = base.ThinkOnSpeakingActionIgnore
	ThinkOnSpeakingActionAppend     = base.ThinkOnSpeakingActionAppend
)

var (
	InterruptionModeStartOfSpeech = base.InterruptionModeStartOfSpeech
	InterruptionModeKeywords      = base.InterruptionModeKeywords

	InterruptionDisabledStrategyAppend = base.InterruptionDisabledStrategyAppend
	InterruptionDisabledStrategyIgnore = base.InterruptionDisabledStrategyIgnore
)

var (
	SpeakPriorityInterrupt = base.SpeakPriorityInterrupt
	SpeakPriorityAppend    = base.SpeakPriorityAppend
	SpeakPriorityIgnore    = base.SpeakPriorityIgnore
)

var (
	StartOfSpeechModeManual = base.StartOfSpeechModeManual
	EndOfSpeechModeManual   = base.EndOfSpeechModeManual
)
