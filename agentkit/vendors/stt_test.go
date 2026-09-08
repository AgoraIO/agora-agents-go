package vendors

import (
	"encoding/json"
	"reflect"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
)

var _ func(...AresSTTOptions) *AresSTT = NewAresSTT

func TestSpeechmaticsSTTNormalizesCredentialToKey(t *testing.T) {
	tests := []struct {
		name string
		opts SpeechmaticsSTTOptions
		want string
	}{
		{
			name: "key",
			opts: SpeechmaticsSTTOptions{Key: "new-key", Language: "en"},
			want: "new-key",
		},
		{
			name: "deprecated APIKey",
			opts: SpeechmaticsSTTOptions{APIKey: "legacy-key", Language: "en"},
			want: "legacy-key",
		},
		{
			name: "key takes precedence",
			opts: SpeechmaticsSTTOptions{Key: "new-key", APIKey: "legacy-key", Language: "en"},
			want: "new-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := NewSpeechmaticsSTT(tt.opts).ToConfig()["params"].(map[string]interface{})
			if params["key"] != tt.want {
				t.Fatalf("key = %#v, want %#v", params["key"], tt.want)
			}
			if _, exists := params["api_key"]; exists {
				t.Fatalf("deprecated api_key leaked onto the wire: %#v", params)
			}
		})
	}
}

func TestGeneratedSpeechmaticsParamsNormalizesDeprecatedAPIKey(t *testing.T) {
	payload, err := json.Marshal(&Agora.SpeechmaticsAsrParams{
		APIKey:   "legacy-key",
		Language: "en",
	})
	if err != nil {
		t.Fatalf("marshal Speechmatics params: %v", err)
	}

	var params map[string]interface{}
	if err := json.Unmarshal(payload, &params); err != nil {
		t.Fatalf("unmarshal Speechmatics params: %v", err)
	}
	if params["key"] != "legacy-key" {
		t.Fatalf("key = %#v, want legacy-key", params["key"])
	}
	if _, exists := params["api_key"]; exists {
		t.Fatalf("deprecated api_key leaked onto the wire: %#v", params)
	}
}

func TestGeminiSTTMatchesGeneratedASRSchema(t *testing.T) {
	wordTimestamp := true
	config := NewGeminiSTT(GeminiSTTOptions{
		APIKey:        "gemini-key",
		Model:         "gemini-3.7-transcribe-live",
		Language:      "en-US",
		SampleRate:    16000,
		WordTimestamp: &wordTimestamp,
	}).ToConfig()

	payload, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("marshal Gemini config: %v", err)
	}
	var generated Agora.Asr
	if err := json.Unmarshal(payload, &generated); err != nil {
		t.Fatalf("unmarshal Gemini config into generated ASR: %v", err)
	}
	if generated.Gemini == nil || generated.Gemini.Params == nil {
		t.Fatalf("generated Gemini params are nil: %#v", generated)
	}

	params := generated.Gemini.Params
	if params.APIKey != "gemini-key" {
		t.Fatalf("api_key = %q, want gemini-key", params.APIKey)
	}
	if params.Model != "gemini-3.7-transcribe-live" {
		t.Fatalf("model = %q, want gemini-3.7-transcribe-live", params.Model)
	}
	if params.Language == nil || *params.Language != "en-US" {
		t.Fatalf("language = %#v, want en-US", params.Language)
	}
	if params.SampleRate == nil || *params.SampleRate != 16000 {
		t.Fatalf("sample_rate = %#v, want 16000", params.SampleRate)
	}
	if params.WordTimestamp == nil || !*params.WordTimestamp {
		t.Fatalf("word_timestamp = %#v, want true", params.WordTimestamp)
	}
}

func TestGeminiSTTLanguageHints(t *testing.T) {
	config := NewGeminiSTT(GeminiSTTOptions{
		APIKey:           "gemini-key",
		LanguageHints:    []string{"en-US", "es-ES"},
		CustomVocabulary: []string{"Agora", "ConvoAI"},
	}).ToConfig()
	params := config["params"].(map[string]interface{})

	if params["model"] != GeminiSTTModel35Live {
		t.Fatalf("model = %#v, want %q", params["model"], GeminiSTTModel35Live)
	}
	if params["sample_rate"] != 16000 {
		t.Fatalf("sample_rate = %#v, want 16000", params["sample_rate"])
	}
	if !reflect.DeepEqual(params["language_hints"], []string{"en-US", "es-ES"}) {
		t.Fatalf("language_hints = %#v", params["language_hints"])
	}
	if !reflect.DeepEqual(params["custom_vocabulary"], []string{"Agora", "ConvoAI"}) {
		t.Fatalf("custom_vocabulary = %#v", params["custom_vocabulary"])
	}

	payload, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("marshal Gemini config: %v", err)
	}
	var generated Agora.Asr
	if err := json.Unmarshal(payload, &generated); err != nil {
		t.Fatalf("unmarshal Gemini config into generated ASR: %v", err)
	}
	if !reflect.DeepEqual(generated.Gemini.Params.LanguageHints, []string{"en-US", "es-ES"}) {
		t.Fatalf("generated language_hints = %#v", generated.Gemini.Params.LanguageHints)
	}
	if !reflect.DeepEqual(generated.Gemini.Params.CustomVocabulary, []string{"Agora", "ConvoAI"}) {
		t.Fatalf("generated custom_vocabulary = %#v", generated.Gemini.Params.CustomVocabulary)
	}
}

func TestGeminiSTTPreservesSliceSemantics(t *testing.T) {
	tests := []struct {
		name       string
		opts       GeminiSTTOptions
		key        string
		want       []string
		wantExists bool
	}{
		{
			name: "nil language hints are omitted",
			opts: GeminiSTTOptions{APIKey: "gemini-key"},
			key:  "language_hints",
		},
		{
			name:       "empty language hints are sent",
			opts:       GeminiSTTOptions{APIKey: "gemini-key", LanguageHints: []string{}},
			key:        "language_hints",
			want:       []string{},
			wantExists: true,
		},
		{
			name: "nil custom vocabulary is omitted",
			opts: GeminiSTTOptions{APIKey: "gemini-key"},
			key:  "custom_vocabulary",
		},
		{
			name:       "empty custom vocabulary is sent",
			opts:       GeminiSTTOptions{APIKey: "gemini-key", CustomVocabulary: []string{}},
			key:        "custom_vocabulary",
			want:       []string{},
			wantExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := NewGeminiSTT(tt.opts).ToConfig()["params"].(map[string]interface{})
			got, exists := params[tt.key]
			if exists != tt.wantExists {
				t.Fatalf("%s existence = %t, want %t", tt.key, exists, tt.wantExists)
			}
			if tt.wantExists && !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("%s = %#v, want %#v", tt.key, got, tt.want)
			}
		})
	}
}

func TestGeminiSTTLanguageHintsTakePriorityOverLanguageCodes(t *testing.T) {
	tests := []struct {
		name          string
		languageHints []string
		languageCodes []string
		expected      []string
	}{
		{
			name:          "language hints override language codes",
			languageHints: []string{"en-US"},
			languageCodes: []string{"fr-FR"},
			expected:      []string{"en-US"},
		},
		{
			name:          "empty language hints override language codes",
			languageHints: []string{},
			languageCodes: []string{"fr-FR"},
			expected:      []string{},
		},
		{
			name:          "language codes remain a fallback",
			languageCodes: []string{"fr-FR"},
			expected:      []string{"fr-FR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewGeminiSTT(GeminiSTTOptions{
				APIKey:        "gemini-key",
				LanguageHints: tt.languageHints,
				LanguageCodes: tt.languageCodes,
			}).ToConfig()
			params := config["params"].(map[string]interface{})

			if !reflect.DeepEqual(params["language_hints"], tt.expected) {
				t.Fatalf("language_hints = %#v, want %#v", params["language_hints"], tt.expected)
			}
		})
	}
}

func TestGeminiSTTRejectsParameterConflict(t *testing.T) {
	wordTimestamp := true
	defer func() {
		if got := recover(); got != "CustomVocabulary cannot be used with WordTimestamp=true" {
			t.Fatalf("panic = %v", got)
		}
	}()

	NewGeminiSTT(GeminiSTTOptions{
		APIKey:           "gemini-key",
		CustomVocabulary: []string{"Agora"},
		WordTimestamp:    &wordTimestamp,
	}).ToConfig()
}

func TestGeminiSTTExplicitFieldsOverrideAdditionalParams(t *testing.T) {
	wordTimestamp := false
	config := NewGeminiSTT(GeminiSTTOptions{
		APIKey:           "gemini-key",
		Model:            "gemini-3.7-transcribe-live",
		Language:         "en-US",
		LanguageHints:    []string{"en-US"},
		LanguageCodes:    []string{"es-ES"},
		CustomVocabulary: []string{"Agora"},
		WordTimestamp:    &wordTimestamp,
		AdditionalParams: map[string]interface{}{
			"api_key":           "wrong-key",
			"model":             "wrong-model",
			"language":          "fr-FR",
			"language_hints":    []string{"fr-FR"},
			"custom_vocabulary": []string{"wrong"},
			"word_timestamp":    true,
			"custom_parameter":  "kept",
		},
	}).ToConfig()

	params := config["params"].(map[string]interface{})
	if params["api_key"] != "gemini-key" || params["model"] != "gemini-3.7-transcribe-live" {
		t.Fatalf("explicit credentials/model did not win: %#v", params)
	}
	if params["language"] != "en-US" || params["word_timestamp"] != false {
		t.Fatalf("explicit optional fields did not win: %#v", params)
	}
	if !reflect.DeepEqual(params["language_hints"], []string{"en-US"}) ||
		!reflect.DeepEqual(params["custom_vocabulary"], []string{"Agora"}) {
		t.Fatalf("explicit compatibility fields did not win: %#v", params)
	}
	if params["custom_parameter"] != "kept" {
		t.Fatalf("additional parameter was lost: %#v", params)
	}
}

func TestGeminiSTTRequiresAPIKey(t *testing.T) {
	defer func() {
		if got := recover(); got != "GeminiSTT requires APIKey" {
			t.Fatalf("panic = %v, want %q", got, "GeminiSTT requires APIKey")
		}
	}()
	NewGeminiSTT(GeminiSTTOptions{Model: "gemini-3.7-transcribe-live"})
}

func TestAresSTTKeywordsMatchGeneratedASR(t *testing.T) {
	wantKeywords := []string{"Agora", "ConvoAI"}
	config := NewAresSTT(AresSTTOptions{
		Keywords: wantKeywords,
		AdditionalParams: map[string]interface{}{
			"custom_param": true,
		},
	}).ToConfig()

	if !reflect.DeepEqual(config["keywords"], wantKeywords) {
		t.Fatalf("keywords = %#v, want %#v", config["keywords"], wantKeywords)
	}

	payload, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("marshal Ares config: %v", err)
	}
	var generated Agora.Asr
	if err := json.Unmarshal(payload, &generated); err != nil {
		t.Fatalf("unmarshal Ares config: %v", err)
	}
	if generated.Ares == nil || generated.Ares.Params == nil {
		t.Fatalf("generated Ares params are nil: %#v", generated)
	}
	if !reflect.DeepEqual(generated.Ares.Keywords, wantKeywords) {
		t.Fatalf("generated keywords = %#v, want %#v", generated.Ares.Keywords, wantKeywords)
	}
	if (*generated.Ares.Params)["custom_param"] != true {
		t.Fatalf("generated params lost custom_param: %#v", generated.Ares.Params)
	}
}

func TestAresSTTOmitsEmptyParams(t *testing.T) {
	config := NewAresSTT().ToConfig()
	if _, exists := config["params"]; exists {
		t.Fatalf("empty Ares config should omit params: %#v", config)
	}
}

func TestAresSTTRejectsKeywordsInAdditionalParams(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("expected NewAresSTT to reject keywords in AdditionalParams")
		}
	}()
	NewAresSTT(AresSTTOptions{AdditionalParams: map[string]interface{}{"keywords": []string{"Agora"}}})
}

func TestAresSTTRejectsMultipleOptions(t *testing.T) {
	defer func() {
		if got := recover(); got != "NewAresSTT accepts at most one options value" {
			t.Fatalf("panic = %v, want multiple-options error", got)
		}
	}()
	NewAresSTT(AresSTTOptions{}, AresSTTOptions{})
}
