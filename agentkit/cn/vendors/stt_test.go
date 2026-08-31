package vendors

import (
	"encoding/json"
	"reflect"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
)

var _ func(...FengmingSTTOptions) *FengmingSTT = NewFengmingSTT

func TestFengmingSTTKeywordsMatchGeneratedASR(t *testing.T) {
	wantKeywords := []string{"Shengwang", "Conversational AI"}
	config := NewFengmingSTT(FengmingSTTOptions{
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
		t.Fatalf("marshal Fengming config: %v", err)
	}
	var generated Agora.Asr
	if err := json.Unmarshal(payload, &generated); err != nil {
		t.Fatalf("unmarshal Fengming config: %v", err)
	}
	if generated.Fengming == nil || generated.Fengming.Params == nil {
		t.Fatalf("generated Fengming params are nil: %#v", generated)
	}
	if !reflect.DeepEqual(generated.Fengming.Keywords, wantKeywords) {
		t.Fatalf("generated keywords = %#v, want %#v", generated.Fengming.Keywords, wantKeywords)
	}
	if (*generated.Fengming.Params)["custom_param"] != true {
		t.Fatalf("generated params lost custom_param: %#v", generated.Fengming.Params)
	}
}

func TestFengmingSTTKeepsNoArgumentCompatibility(t *testing.T) {
	first := *NewFengmingSTT()
	second := *NewFengmingSTT()
	if first != second {
		t.Fatal("default Fengming values should retain empty-struct equality semantics")
	}

	config := NewFengmingSTT().ToConfig()
	if _, exists := config["params"]; exists {
		t.Fatalf("empty Fengming config should omit params: %#v", config)
	}

	var nilFengming *FengmingSTT
	if got := nilFengming.ToConfig()["vendor"]; got != "fengming" {
		t.Fatalf("nil receiver vendor = %v, want fengming", got)
	}
}

func TestFengmingSTTAllowsAdditionalParams(t *testing.T) {
	config := NewFengmingSTT(FengmingSTTOptions{
		AdditionalParams: map[string]interface{}{
			"custom_param": true,
			"sample_rate":  16000,
		},
	}).ToConfig()

	params, ok := config["params"].(map[string]interface{})
	if !ok {
		t.Fatalf("params = %#v, want map[string]interface{}", config["params"])
	}
	if params["custom_param"] != true || params["sample_rate"] != 16000 {
		t.Fatalf("additional params were not preserved: %#v", params)
	}
}

func TestFengmingSTTRejectsMultipleOptions(t *testing.T) {
	defer func() {
		if got := recover(); got != "NewFengmingSTT accepts at most one options value" {
			t.Fatalf("panic = %v, want multiple-options error", got)
		}
	}()
	NewFengmingSTT(FengmingSTTOptions{}, FengmingSTTOptions{})
}
