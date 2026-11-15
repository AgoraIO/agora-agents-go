package agentkit

import (
	"encoding/json"
	"math"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	cnvendors "github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors"
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

func TestMiniMaxDuplexGeneratedConfig(t *testing.T) {
	t.Parallel()
	speed, zero, disabled := 1.2, 0.0, false
	extra := map[string]interface{}{"key": "overridden", "continuous_sound": true, "future_param": "kept"}
	tts := cnvendors.NewMiniMaxDuplexTTS(cnvendors.MiniMaxDuplexTTSOptions{
		Key: "mm-key", Model: "speech-2.8-turbo", URL: "wss://api.minimax.cn/ws/v1/t2a_v2_bidi",
		VoiceSetting: &cnvendors.MiniMaxDuplexVoiceSetting{
			VoiceID: "female-shaonv", Speed: &speed, Volume: &zero, Pitch: &zero,
			Emotion: "happy", LatexRead: &disabled,
		},
		AudioSetting:      &cnvendors.MiniMaxDuplexAudioSetting{SampleRate: 24000, Format: "pcm", Channels: 1},
		PronunciationDict: &cnvendors.MiniMaxPronunciationDict{Tone: []string{"测试/(ce4)(shi4)"}},
		VoiceModify: &cnvendors.MiniMaxDuplexVoiceModify{
			Pitch: &zero, Intensity: &zero, Timbre: &zero, SoundEffects: "spacious_echo",
		},
		LanguageBoost: "Chinese", TextNormalizationMode: "basic", SubtitleEnable: &disabled,
		SubtitleType: "word", ContinuousSound: &disabled, SessionID: "session-1",
		SkipPatterns: []int{1}, AdditionalParams: extra,
	})
	const expected = `{
		"vendor":"minimax_duplex", "params":{
			"key":"mm-key", "model":"speech-2.8-turbo", "url":"wss://api.minimax.cn/ws/v1/t2a_v2_bidi",
			"voice_setting":{"voice_id":"female-shaonv","speed":1.2,"vol":0,"pitch":0,"emotion":"happy","latex_read":false},
			"audio_setting":{"sample_rate":24000,"format":"pcm","channels":1},
			"pronunciation_dict":{"tone":["测试/(ce4)(shi4)"]},
			"voice_modify":{"pitch":0,"intensity":0,"timbre":0,"sound_effects":"spacious_echo"},
			"language_boost":"Chinese", "text_normalization_mode":"basic", "subtitle_enable":false,
			"subtitle_type":"word", "continuous_sound":false, "session_id":"session-1", "future_param":"kept"
		}, "skip_patterns":[1]
	}`
	assertJSONEqual(t, tts.ToConfig(), expected)
	if got := tts.GetSampleRate(); got == nil || *got != vendors.SampleRate24kHz {
		t.Fatalf("sample rate = %v", got)
	}
	payload, err := json.Marshal(tts.ToConfig())
	if err != nil {
		t.Fatal(err)
	}
	var generated Agora.Tts
	if err := json.Unmarshal(payload, &generated); err != nil {
		t.Fatal(err)
	}
	if generated.MinimaxDuplex == nil || generated.MinimaxDuplex.Params.VoiceSetting.VoiceID != "female-shaonv" {
		t.Fatalf("generated duplex variant missing: %#v", generated)
	}
	assertJSONEqual(t, &generated, expected)
	if extra["key"] != "overridden" || extra["continuous_sound"] != true {
		t.Fatalf("additional params mutated: %v", extra)
	}
}

func TestSonioxGeneratedConfig(t *testing.T) {
	t.Parallel()
	rate, speed, disabled := vendors.SampleRate24kHz, 0.7, false
	tts := vendors.NewSonioxTTS(vendors.SonioxTTSOptions{
		APIKey: "k", BaseURL: "wss://tts.soniox.com", Model: "model", Language: "en", Voice: "voice",
		SampleRate: &rate, Speed: &speed, ReduceSilence: &disabled, ReturnTimestamps: &disabled,
		ClientReferenceID: "ref", Dump: &disabled, SkipPatterns: []int{1},
		AdditionalParams: map[string]interface{}{"api_key": "overridden", "reduce_silence": true, "future_param": 42},
	})
	const expected = `{
		"vendor":"soniox", "params":{
			"api_key":"k", "base_url":"wss://tts.soniox.com", "model":"model", "language":"en", "voice":"voice",
			"sample_rate":24000, "speed":0.7, "reduce_silence":false, "return_timestamps":false,
			"client_reference_id":"ref", "dump":false, "future_param":42
		}, "skip_patterns":[1]
	}`
	assertJSONEqual(t, tts.ToConfig(), expected)
	if got := tts.GetSampleRate(); got == nil || *got != rate {
		t.Fatalf("sample rate = %v", got)
	}
	payload, err := json.Marshal(tts.ToConfig())
	if err != nil {
		t.Fatal(err)
	}
	var generated Agora.Tts
	if err := json.Unmarshal(payload, &generated); err != nil {
		t.Fatal(err)
	}
	if generated.Soniox == nil || generated.Soniox.Params.APIKey != "k" {
		t.Fatalf("generated Soniox variant missing: %#v", generated)
	}
	assertJSONEqual(t, &generated, expected)
}

func TestMiniMaxDuplexAndSonioxMinimalConfigs(t *testing.T) {
	t.Parallel()
	tts := cnvendors.NewMiniMaxDuplexTTS(cnvendors.MiniMaxDuplexTTSOptions{
		Key: "k", Model: "m", VoiceSetting: &cnvendors.MiniMaxDuplexVoiceSetting{VoiceID: "v"},
	})
	assertJSONEqual(t, tts.ToConfig(), `{
		"vendor":"minimax_duplex","params":{"key":"k","model":"m","voice_setting":{"voice_id":"v"}}
	}`)
	soniox := vendors.NewSonioxTTS(vendors.SonioxTTSOptions{APIKey: "k"})
	assertJSONEqual(t, soniox.ToConfig(), `{"vendor":"soniox","params":{"api_key":"k"}}`)
	if tts.GetSampleRate() != nil || soniox.GetSampleRate() != nil {
		t.Fatal("omitted sample rates must remain unspecified")
	}
}

func TestMiniMaxDuplexAndSonioxValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		call func()
	}{
		{name: "duplex key", call: func() {
			cnvendors.NewMiniMaxDuplexTTS(cnvendors.MiniMaxDuplexTTSOptions{})
		}},
		{name: "duplex model", call: func() {
			cnvendors.NewMiniMaxDuplexTTS(cnvendors.MiniMaxDuplexTTSOptions{Key: "k"})
		}},
		{name: "duplex voice", call: func() {
			cnvendors.NewMiniMaxDuplexTTS(cnvendors.MiniMaxDuplexTTSOptions{Key: "k", Model: "m"})
		}},
		{name: "duplex empty voice", call: func() {
			cnvendors.NewMiniMaxDuplexTTS(cnvendors.MiniMaxDuplexTTSOptions{
				Key: "k", Model: "m", VoiceSetting: &cnvendors.MiniMaxDuplexVoiceSetting{},
			})
		}},
		{name: "Soniox key", call: func() { vendors.NewSonioxTTS(vendors.SonioxTTSOptions{}) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected invalid options to panic")
				}
			}()
			tt.call()
		})
	}
	for _, speed := range []float64{0.6, 1.4, math.NaN()} {
		t.Run("Soniox speed", func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected invalid speed to panic")
				}
			}()
			vendors.NewSonioxTTS(vendors.SonioxTTSOptions{APIKey: "k", Speed: &speed})
		})
	}
}
