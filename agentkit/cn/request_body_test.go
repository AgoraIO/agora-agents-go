package cn

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors"
	"github.com/AgoraIO/agora-agents-go/v2/client"
	"github.com/AgoraIO/agora-agents-go/v2/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type captureQwenStartHTTPClient struct {
	body []byte
}

func (c *captureQwenStartHTTPClient) Do(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	c.body = body

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{"agent_id":"agent_123","status":"RUNNING"}`)),
	}, nil
}

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
	agent := NewAgent(testAgoraClient()).WithMllm(
		vendors.NewQwenOmni(vendors.QwenOmniOptions{
			APIKey: "dashscope-key",
			Model:  "qwen3.5-omni-plus-realtime",
			URL:    "wss://dashscope.aliyuncs.com/api-ws/v1/realtime",
		}),
	)

	props, err := agent.ToPropertiesMap(basePropertiesOpts())
	require.NoError(t, err)

	assert.Equal(t, map[string]interface{}{
		"enable":  true,
		"vendor":  "qwen_omni",
		"url":     "wss://dashscope.aliyuncs.com/api-ws/v1/realtime",
		"api_key": "dashscope-key",
		"params": map[string]interface{}{
			"model": "qwen3.5-omni-plus-realtime",
		},
	}, props["mllm"])
	assert.NotContains(t, props, "llm")
	assert.NotContains(t, props, "tts")
}

func TestQwenOmniStartRequestContainsRequiredMLLMFields(t *testing.T) {
	httpClient := &captureQwenStartHTTPClient{}
	rawClient := client.NewClient(
		option.WithBaseURL("https://api.example.test"),
		option.WithHTTPClient(httpClient),
	)
	agoraClient := NewAgoraClient(ClientOptions{
		AppID:          "0123456789abcdef0123456789abcdef",
		AppCertificate: "fedcba9876543210fedcba9876543210",
	})
	agoraClient.Agents = rawClient.Agents

	agent := NewAgent(agoraClient).WithMllm(
		vendors.NewQwenOmni(vendors.QwenOmniOptions{
			APIKey: "dashscope-key",
			Model:  "qwen3.5-omni-plus-realtime",
			URL:    "wss://dashscope.aliyuncs.com/api-ws/v1/realtime",
		}),
	)
	session := agent.CreateSession(CreateSessionOptions{
		Name:       "qwen-test",
		Channel:    "channel",
		Token:      "rtc-token",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
	})

	_, err := session.Start(context.Background())
	require.NoError(t, err)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(httpClient.body, &payload))
	properties := payload["properties"].(map[string]interface{})
	assert.Equal(t, map[string]interface{}{
		"enable":  true,
		"vendor":  "qwen_omni",
		"url":     "wss://dashscope.aliyuncs.com/api-ws/v1/realtime",
		"api_key": "dashscope-key",
		"params": map[string]interface{}{
			"model": "qwen3.5-omni-plus-realtime",
		},
	}, properties["mllm"])
}
