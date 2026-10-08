package agentkit

import (
	"encoding/json"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpeakBatchStartRequest(t *testing.T) {
	t.Parallel()
	agent := NewAgent(testAgoraClient(), WithParameters(&SessionParams{
		Speak:         &SpeakConfig{Batch: Agora.Bool(false)},
		SilenceConfig: &SilenceConfig{TimeoutMs: Agora.Int(15000), Action: SilenceActionThink.Ptr()},
	})).WithStt(stubASR).WithLlm(stubLLM).WithTts(stubTTS)

	generated, err := agent.ToProperties(basePropertiesOpts())
	require.NoError(t, err)
	require.NotNil(t, generated.Parameters.Speak)
	assert.Equal(t, Agora.Bool(false), generated.Parameters.Speak.Batch)

	payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
	require.NoError(t, err)
	properties, ok := payload["properties"].(map[string]any)
	require.True(t, ok)
	parametersJSON, err := json.Marshal(properties["parameters"])
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"speak":{"batch":false},"silence_config":{"timeout_ms":15000,"action":"think"},"audio_scenario":"default"
	}`, string(parametersJSON))
}

func TestSpeakBatchOptionalValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		speak    *ParametersSpeak
		expected string
	}{
		{name: "omitted speak", expected: `{"audio_scenario":"chorus"}`},
		{name: "omitted batch", speak: &SpeakConfig{}, expected: `{"audio_scenario":"chorus","speak":{}}`},
		{
			name: "false", speak: &SpeakConfig{Batch: Agora.Bool(false)},
			expected: `{"audio_scenario":"chorus","speak":{"batch":false}}`,
		},
		{
			name: "true", speak: &SpeakConfig{Batch: Agora.Bool(true)},
			expected: `{"audio_scenario":"chorus","speak":{"batch":true}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewAgent(testAgoraClient()).WithStt(stubASR).WithLlm(stubLLM).WithTts(stubTTS).
				WithParameters(&SessionParamsInput{Speak: tt.speak}).WithAudioScenario(ParametersAudioScenarioChorus)
			props, err := agent.ToProperties(basePropertiesOpts())
			require.NoError(t, err)
			payload, err := json.Marshal(props.Parameters)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(payload))
		})
	}
}

func TestNewAgentRequiresClient(t *testing.T) {
	require.PanicsWithValue(t, "NewAgent requires AgoraClient", func() {
		NewAgent(nil)
	})
}

func TestValidateAvatarConfigAllowsSensetimeWithoutSceneList(t *testing.T) {
	err := ValidateAvatarConfig("sensetime", map[string]interface{}{
		"agora_uid": "2001",
		"appId":     "sensetime-app",
		"app_key":   "sensetime-key",
	})
	require.NoError(t, err)
}
