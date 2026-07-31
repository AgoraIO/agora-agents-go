package vendors

import "github.com/AgoraIO/agora-agents-go/v2/agentkit/core"

type SampleRate = core.SampleRate

const (
	SampleRate8kHz  = core.SampleRate8kHz
	SampleRate16kHz = core.SampleRate16kHz
	SampleRate22kHz = core.SampleRate22kHz
	SampleRate24kHz = core.SampleRate24kHz
	SampleRate44kHz = core.SampleRate44kHz
	SampleRate48kHz = core.SampleRate48kHz
)

type (
	LLM    = core.LLMVendor
	TTS    = core.TTSVendor
	STT    = core.STTVendor
	MLLM   = core.MLLMVendor
	Avatar = core.AvatarVendorConfig
)
