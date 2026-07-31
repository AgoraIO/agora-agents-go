package cn

import (
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testAgoraClient() *AgoraClient {
	return NewAgoraClient(ClientOptions{
		AppID:          "test-app-id",
		AppCertificate: "test-app-certificate",
	})
}

func basePropertiesOpts() ToPropertiesOptions {
	return ToPropertiesOptions{
		Channel:    "channel",
		Token:      "token",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
	}
}

func TestDefaultASRFallsBackToFengming(t *testing.T) {
	agent := NewAgent(testAgoraClient()).
		WithLlm(vendors.NewAliyun(vendors.AliyunOptions{
			APIKey:  "aliyun-key",
			Model:   "qwen-max",
			BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
		})).
		WithTts(vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
			Key:   "mm-key",
			Model: "speech-2.6-turbo",
			VoiceSetting: &vendors.MiniMaxVoiceSetting{
				VoiceID: "Chinese (Mandarin)_Cheerful_Female",
			},
		}))

	props, err := agent.ToPropertiesMap(basePropertiesOpts())
	require.NoError(t, err)

	asr := props["asr"].(map[string]interface{})
	assert.Equal(t, "fengming", asr["vendor"])
	assert.Equal(t, "en-US", asr["language"])
	assert.NotContains(t, asr, "params")
}

func TestQwenOmniMLLMProperties(t *testing.T) {
	turnDetection := &Agora.MllmTurnDetection{
		Mode: Agora.MllmTurnDetectionModeServerVad.Ptr(),
	}
	agent := NewAgent(testAgoraClient()).WithMllm(
		vendors.NewQwenOmni(vendors.QwenOmniOptions{
			APIKey:        "dashscope-key",
			Model:         "qwen3-omni-flash-realtime",
			Voice:         "Momo",
			TurnDetection: turnDetection,
		}),
	)

	props, err := agent.ToPropertiesMap(basePropertiesOpts())
	require.NoError(t, err)

	mllm := props["mllm"].(map[string]interface{})
	assert.Equal(t, "qwen_omni", mllm["vendor"])
	assert.Equal(t, true, mllm["enable"])
	assert.Equal(t, "dashscope-key", mllm["api_key"])
	assert.Equal(t, "wss://dashscope.aliyuncs.com/api-ws/v1/realtime", mllm["url"])
	assert.Equal(t, turnDetection, mllm["turn_detection"])
	params := mllm["params"].(map[string]interface{})
	assert.Equal(t, "qwen3-omni-flash-realtime", params["model"])
	assert.Equal(t, "Momo", params["voice"])
	assert.NotContains(t, props, "llm")
	assert.NotContains(t, props, "tts")
}
