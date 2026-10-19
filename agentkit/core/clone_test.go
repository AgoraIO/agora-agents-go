package core

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestCloneValueKeepsEmptyStringSliceNonNil(t *testing.T) {
	src := []string{}
	got, ok := CloneValue(src).([]string)
	if !ok {
		t.Fatalf("CloneValue type = %T, want []string", CloneValue(src))
	}
	if got == nil {
		t.Fatal("CloneValue turned empty []string into nil")
	}
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestCloneValueKeepsEmptyIntSliceNonNil(t *testing.T) {
	src := []int{}
	got, ok := CloneValue(src).([]int)
	if !ok {
		t.Fatalf("CloneValue type = %T, want []int", CloneValue(src))
	}
	if got == nil {
		t.Fatal("CloneValue turned empty []int into nil")
	}
}

func TestCloneValueNilSlicesStayNil(t *testing.T) {
	if CloneValue(([]string)(nil)) != nil {
		t.Fatalf("nil []string clone = %#v, want nil", CloneValue(([]string)(nil)))
	}
	if CloneValue(([]int)(nil)) != nil {
		t.Fatalf("nil []int clone = %#v, want nil", CloneValue(([]int)(nil)))
	}
}

func TestCloneConfigCopiesSlicesIndependently(t *testing.T) {
	src := map[string]interface{}{
		"dicts":    []string{"a", "b"},
		"patterns": []int{1, 2},
	}
	clone := CloneConfig(src)

	src["dicts"].([]string)[0] = "mutated"
	src["patterns"].([]int)[0] = 99

	if got := clone["dicts"].([]string); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("dicts = %#v, want independent copy", got)
	}
	if got := clone["patterns"].([]int); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("patterns = %#v, want independent copy", got)
	}
}

func TestCloneConfigEmptyPronunciationDictsJSON(t *testing.T) {
	cfg := map[string]interface{}{
		"params": map[string]interface{}{
			"pronunciation_dicts": []string{},
		},
	}
	clone := CloneConfig(cfg)
	payload, err := json.Marshal(clone)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(payload), `"pronunciation_dicts":[]`) {
		t.Fatalf("payload = %s, want pronunciation_dicts as []", payload)
	}
}
