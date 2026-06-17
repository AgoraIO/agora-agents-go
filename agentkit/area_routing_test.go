package agentkit

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/client"
	"github.com/AgoraIO/agora-agents-go/v2/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type captureStartHTTPClient struct {
	lastRequest *http.Request
	lastBody    []byte
}

func (c *captureStartHTTPClient) Do(req *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(req.Body)
	c.lastRequest = req.Clone(req.Context())
	c.lastBody = body

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{"agent_id":"agent_123","status":"RUNNING"}`)),
	}, nil
}

func TestSessionStartWithAreaUsesConvoAIPathAndAppCredentialsAuth(t *testing.T) {
	httpClient := &captureStartHTTPClient{}
	rawClient := client.NewClient(
		option.WithArea(option.AreaUS),
		option.WithHTTPClient(httpClient),
	)

	agoraClient := &AgoraClient{
		Agents:         rawClient.Agents,
		AppID:          "0123456789abcdef0123456789abcdef",
		AppCertificate: "fedcba9876543210fedcba9876543210",
		AuthMode:       AuthModeAppCredentials,
	}

	agent := NewAgent(WithName("support-agent"))
	session := agent.CreateSession(agoraClient, CreateSessionOptions{
		Channel:    "room-1",
		Token:      "rtc-token",
		AgentUID:   "1",
		RemoteUIDs: []string{"100", "101"},
		Preset: []string{
			AgentPresets.Asr.DeepgramNova3,
			AgentPresets.Llm.OpenAIGpt4oMini,
			AgentPresets.Tts.OpenAITts1,
		},
	})

	_, err := session.Start(context.Background())
	require.NoError(t, err)
	require.NotNil(t, httpClient.lastRequest)

	assert.Equal(t, "/api/conversational-ai-agent/v2/projects/0123456789abcdef0123456789abcdef/join", httpClient.lastRequest.URL.Path)
	assert.Contains(t, httpClient.lastRequest.Header.Get("Authorization"), "agora token=")
	assert.NotEqual(t, "agora token=rtc-token", httpClient.lastRequest.Header.Get("Authorization"))

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(httpClient.lastBody, &payload))
	properties, ok := payload["properties"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "rtc-token", properties["token"])
	remoteUIDs, ok := properties["remote_rtc_uids"].([]interface{})
	require.True(t, ok)
	require.Len(t, remoteUIDs, 2)
	assert.Equal(t, "100", remoteUIDs[0])
	assert.Equal(t, "101", remoteUIDs[1])
}

func TestSessionStartAutoGeneratesRESTAuthTokenWhenSessionTokenMissing(t *testing.T) {
	httpClient := &captureStartHTTPClient{}
	rawClient := client.NewClient(
		option.WithArea(option.AreaUS),
		option.WithHTTPClient(httpClient),
	)

	agoraClient := &AgoraClient{
		Agents:         rawClient.Agents,
		AppID:          "0123456789abcdef0123456789abcdef",
		AppCertificate: "fedcba9876543210fedcba9876543210",
		AuthMode:       AuthModeAppCredentials,
	}

	agent := NewAgent(WithName("support-agent"))
	session := agent.CreateSession(agoraClient, CreateSessionOptions{
		Channel:    "room-1",
		AgentUID:   "1",
		RemoteUIDs: []string{"100", "101"},
		Preset: []string{
			AgentPresets.Asr.DeepgramNova3,
			AgentPresets.Llm.OpenAIGpt4oMini,
			AgentPresets.Tts.OpenAITts1,
		},
	})

	_, err := session.Start(context.Background())
	require.NoError(t, err)
	require.NotNil(t, httpClient.lastRequest)

	assert.Equal(t, "/api/conversational-ai-agent/v2/projects/0123456789abcdef0123456789abcdef/join", httpClient.lastRequest.URL.Path)
	assert.Contains(t, httpClient.lastRequest.Header.Get("Authorization"), "agora token=")
	assert.NotEqual(t, "agora token=", httpClient.lastRequest.Header.Get("Authorization"))
}
