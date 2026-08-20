package vendors

import (
	"encoding/json"
	"reflect"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
)

var _ func(...AresSTTOptions) *AresSTT = NewAresSTT

func TestAresSTTKeywordsMatchGeneratedASR(t *testing.T) {
	wantKeywords := []string{"Agora", "ConvoAI"}
	config := NewAresSTT(AresSTTOptions{
		Keywords: wantKeywords,
		AdditionalParams: map[string]interface{}{
			"custom_param": true,
			"keywords":     []string{"overridden"},
		},
	}).ToConfig()

	if !reflect.DeepEqual(config["keywords"], wantKeywords) {
		t.Fatalf("keywords = %#v, want %#v", config["keywords"], wantKeywords)
	}
	params := config["params"].(map[string]interface{})
	if _, exists := params["keywords"]; exists {
		t.Fatalf("typed keywords must not also be sent in params: %#v", params)
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

func TestAresSTTRejectsMultipleOptions(t *testing.T) {
	defer func() {
		if got := recover(); got != "NewAresSTT accepts at most one options value" {
			t.Fatalf("panic = %v, want multiple-options error", got)
		}
	}()
	NewAresSTT(AresSTTOptions{}, AresSTTOptions{})
}
