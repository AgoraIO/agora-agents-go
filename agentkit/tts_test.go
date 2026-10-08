package agentkit

import (
	"encoding/json"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSarvamTTSStartRequest(t *testing.T) {
	t.Parallel()
	agent := NewAgent(testAgoraClient()).WithStt(stubASR).WithLlm(stubLLM).
		WithTts(vendors.NewSarvamTTS(vendors.SarvamTTSOptions{
			Key: "sarvam-key", Speaker: "anushka", TargetLanguageCode: vendors.SarvamTTSLanguageEnIN,
			SpeechSampleRate: Agora.Int(24000), EnablePreprocessing: Agora.Bool(false), Model: "bulbul:v3",
			AdditionalParams: map[string]any{"custom_option": "value"},
		}))

	generated, err := agent.ToProperties(basePropertiesOpts())
	require.NoError(t, err)
	assert.Equal(t, Agora.Int(24000), generated.Tts.Sarvam.Params.SpeechSampleRate)

	payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
	require.NoError(t, err)
	properties, ok := payload["properties"].(map[string]any)
	require.True(t, ok)
	ttsJSON, err := json.Marshal(properties["tts"])
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"vendor":"sarvam","params":{
			"api_subscription_key":"sarvam-key","speaker":"anushka","target_language_code":"en-IN",
			"speech_sample_rate":24000,"enable_preprocessing":false,"model":"bulbul:v3","custom_option":"value"}
	}`, string(ttsJSON))
}

func TestSarvamTTSAvatarSampleRate(t *testing.T) {
	t.Parallel()
	avatar := vendors.NewLiveAvatarAvatar(vendors.LiveAvatarAvatarOptions{
		APIKey: "avatar-key", Quality: "low", AgoraUID: "2001", AgoraToken: "avatar-token",
	})
	tts16 := vendors.NewSarvamTTS(vendors.SarvamTTSOptions{
		Key: "key", Speaker: "anushka", TargetLanguageCode: "en-IN", SpeechSampleRate: Agora.Int(16000),
	})
	assert.Panics(t, func() { NewAgent(testAgoraClient()).WithTts(tts16).WithAvatar(avatar) })
	assert.Panics(t, func() { NewAgent(testAgoraClient()).WithAvatar(avatar).WithTts(tts16) })

	tts24 := vendors.NewSarvamTTS(vendors.SarvamTTSOptions{
		Key: "key", Speaker: "anushka", TargetLanguageCode: "en-IN", SpeechSampleRate: Agora.Int(24000),
	})
	agent := NewAgent(testAgoraClient()).WithStt(stubASR).WithLlm(stubLLM).WithTts(tts24).WithAvatar(avatar)
	_, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
	require.NoError(t, err)
}

type rawTTSConfig map[string]any

func (c rawTTSConfig) ToConfig() map[string]any           { return c }
func (c rawTTSConfig) GetSampleRate() *vendors.SampleRate { return nil }

func TestAvatarValidationReadsSpeechSampleRateFromWireParams(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		rate    any
		wantErr bool
	}{
		{name: "compatible integer", rate: 24000},
		{name: "compatible json number", rate: float64(24000)},
		{name: "incompatible integer", rate: 16000, wantErr: true},
		{name: "incompatible json number", rate: float64(16000), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewAgent(testAgoraClient()).WithStt(stubASR).WithLlm(stubLLM).
				WithTts(rawTTSConfig{"vendor": "sarvam", "params": map[string]any{
					"api_subscription_key": "key", "speaker": "anushka", "target_language_code": "en-IN",
					"speech_sample_rate": tt.rate,
				}}).WithAvatar(vendors.NewLiveAvatarAvatar(vendors.LiveAvatarAvatarOptions{
				APIKey: "avatar-key", Quality: "low", AgoraUID: "2001", AgoraToken: "avatar-token",
			}))
			_, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "24,000")
				assert.Contains(t, err.Error(), "16000")
				return
			}
			require.NoError(t, err)
		})
	}
}
