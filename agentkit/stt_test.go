package agentkit

import (
	"encoding/json"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rawSTTConfig map[string]any

func (c rawSTTConfig) ToConfig() map[string]any { return c }

func TestRTZRSTTStartRequest(t *testing.T) {
	t.Parallel()
	agent := NewAgent(testAgoraClient()).
		WithStt(vendors.NewRTZRSTT(vendors.RTZRSTTOptions{
			ClientID: "client-id", ClientSecret: "client-secret", Language: "ko",
			UseDisfluencyFilter: Agora.Bool(false), Keywords: []string{},
		})).
		WithLlm(stubLLM).
		WithTts(stubTTS)

	generated, err := agent.ToProperties(basePropertiesOpts())
	require.NoError(t, err)
	require.NotNil(t, generated.Asr.Rtzr)
	assert.Equal(t, "client-id", generated.Asr.Rtzr.Params.ClientID)
	assert.Equal(t, Agora.String("ko"), generated.Asr.Rtzr.Params.Language)

	payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
	require.NoError(t, err)
	properties, ok := payload["properties"].(map[string]any)
	require.True(t, ok)
	asrJSON, err := json.Marshal(properties["asr"])
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"vendor":"rtzr","language":"en-US",
		"params":{"client_id":"client-id","client_secret":"client-secret","language":"ko",
			"use_disfluency_filter":false,"keywords":[]}
	}`, string(asrJSON))
}

func TestDeepgramCredentialNormalizationPreservesBYOK(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		params   map[string]any
		expected string
	}{
		{name: "new field", params: map[string]any{"api_key": "new-key", "model": "nova-2"}, expected: "new-key"},
		{name: "legacy alias", params: map[string]any{"key": "legacy-key", "model": "nova-2"}, expected: "legacy-key"},
		{
			name: "new field wins", params: map[string]any{"api_key": "new-key", "key": "legacy-key", "model": "nova-2"},
			expected: "new-key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := rawSTTConfig{"vendor": "deepgram", "params": tt.params}
			agent := NewAgent(testAgoraClient()).WithStt(config).
				WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-4o-mini"})).WithTts(stubTTS)
			generated, err := agent.ToProperties(basePropertiesOpts())
			require.NoError(t, err)
			assert.Equal(t, tt.expected, generated.Asr.Deepgram.Params.APIKey)
			assert.NotContains(t, generated.Asr.Deepgram.Params.ExtraProperties, "key")

			payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
			require.NoError(t, err)
			assert.Equal(t, AgentPresets.Llm.OpenAIGpt4oMini, payload["preset"])
			properties, ok := payload["properties"].(map[string]any)
			require.True(t, ok)
			asr := asrFromProperties(t, properties)
			params, ok := asr["params"].(map[string]any)
			require.True(t, ok)
			assert.Equal(t, tt.expected, params["api_key"])
			assert.Equal(t, "nova-2", params["model"])
			assert.NotContains(t, params, "key")
			assert.Equal(t, tt.params, agent.BaseAgent().STT["params"])
		})
	}
}
