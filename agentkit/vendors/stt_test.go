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

func TestAresSTTKeywordsMatchGeneratedASR(t *testing.T) {
	wantKeywords := []string{"Agora", "ConvoAI"}
	config := NewAresSTT(AresSTTOptions{
		Keywords: wantKeywords,
		AdditionalParams: map[string]interface{}{
			"custom_param": true,
			"keywords":     []string{"overridden"},
		},
	}).ToConfig()

	params := config["params"].(map[string]interface{})
	if !reflect.DeepEqual(params["keywords"], wantKeywords) {
		t.Fatalf("keywords = %#v, want %#v", params["keywords"], wantKeywords)
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
	if !reflect.DeepEqual(generated.Ares.Params.Keywords, wantKeywords) {
		t.Fatalf("generated keywords = %#v, want %#v", generated.Ares.Params.Keywords, wantKeywords)
	}
	if generated.Ares.Params.GetExtraProperties()["custom_param"] != true {
		t.Fatalf("generated params lost custom_param: %#v", generated.Ares.Params)
	}
}

func TestAresSTTOmitsEmptyParams(t *testing.T) {
	config := NewAresSTT().ToConfig()
	if _, exists := config["params"]; exists {
		t.Fatalf("empty Ares config should omit params: %#v", config)
	}
}

func TestAresSTTRejectsMultipleOptions(t *testing.T) {
	defer func() {
		if got := recover(); got != "NewAresSTT accepts at most one options value" {
			t.Fatalf("panic = %v, want multiple-options error", got)
		}
	}()
	NewAresSTT(AresSTTOptions{}, AresSTTOptions{})
}
