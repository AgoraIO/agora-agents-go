package agentkit

import (
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
)

func TestPlatformASRPreservesRequiredProviderLanguage(t *testing.T) {
	ref := SecretRef("models", "asr")
	configs := map[string]map[string]any{
		"assemblyai":   vendors.NewAssemblyAISTT(vendors.AssemblyAISTTOptions{APIKey: ref, Language: "hi-IN"}).ToConfig(),
		"google":       vendors.NewGoogleSTT(vendors.GoogleSTTOptions{ADCCredentialsString: ref, ProjectID: "project", Location: "global", Language: "hi-IN"}).ToConfig(),
		"microsoft":    vendors.NewMicrosoftSTT(vendors.MicrosoftSTTOptions{Key: ref, Region: "us", Language: "hi-IN"}).ToConfig(),
		"sarvam":       vendors.NewSarvamSTT(vendors.SarvamSTTOptions{APIKey: ref, Language: "hi-IN"}).ToConfig(),
		"speechmatics": vendors.NewSpeechmaticsSTT(vendors.SpeechmaticsSTTOptions{Key: ref, Language: "hi-IN"}).ToConfig(),
	}
	for vendor, input := range configs {
		t.Run(vendor, func(t *testing.T) {
			got, bad := convertModule("asr", input)
			if len(bad) != 0 {
				t.Fatal(bad)
			}
			if got["language"] != "hi-IN" || got["params"].(map[string]any)["language"] != "hi-IN" {
				t.Fatal(got)
			}
			if _, exists := input["language"]; exists {
				t.Fatal("mutated caller configuration")
			}

			language := Agora.AsrLanguage("en-US")
			agent := NewAgent(testAgoraClient()).WithTurnDetection(&TurnDetectionConfig{Language: &language})
			agent.base.STT = input
			body, err := agent.toV3("language-test")
			if err != nil {
				t.Fatal(err)
			}
			if body.Pipeline.ASR["language"] != "en-US" || body.Pipeline.ASR["params"].(map[string]any)["language"] != "hi-IN" {
				t.Fatal(body.Pipeline.ASR)
			}

			agent.base.STT = map[string]any{"vendor": vendor}
			body, err = agent.toV3("language-test")
			if err != nil {
				t.Fatal(err)
			}
			if body.Pipeline.ASR["params"].(map[string]any)["language"] != "en-US" {
				t.Fatal(body.Pipeline.ASR)
			}

			got, bad = convertModule("asr", map[string]any{"vendor": vendor, "language": "de-DE"})
			if len(bad) != 0 || got["params"].(map[string]any)["language"] != "de-DE" {
				t.Fatal(got, bad)
			}
		})
	}
}

func TestPlatformSarvamTTSModel(t *testing.T) {
	for _, tc := range []struct{ name, model, extra, want string }{
		{"default", "", "", "bulbul:v3"},
		{"typed override", "future-model", "", "future-model"},
		{"additional override", "", "additional-model", "additional-model"},
		{"typed wins", "typed-model", "additional-model", "typed-model"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			extra := map[string]any{}
			if tc.extra != "" {
				extra["model"] = tc.extra
			}
			config := vendors.NewSarvamTTS(vendors.SarvamTTSOptions{Key: SecretRef("models", "sarvam"), Speaker: "anushka", TargetLanguageCode: "hi-IN", Model: tc.model, AdditionalParams: extra}).ToConfig()
			got, bad := convertModule("tts", config)
			if len(bad) != 0 || got["params"].(map[string]any)["model"] != tc.want {
				t.Fatal(got, bad)
			}
			if tc.extra == "" && len(extra) != 0 {
				t.Fatal("mutated additional parameters")
			}
		})
	}
}
