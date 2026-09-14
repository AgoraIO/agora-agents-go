package agentkit

import (
	"encoding/json"
	"strings"
	"testing"
)

const debugTestAPIKey = "test-google-api-key"

func TestRedactSecretsWalksNestedStructures(t *testing.T) {
	redacted := RedactSecrets(map[string]interface{}{
		"appid": "81190c52971d4004b7244bdcd93e2f34",
		"properties": map[string]interface{}{
			"token": "007eJxTYKhsrH10",
			"asr": map[string]interface{}{
				"vendor": "gemini",
				"params": map[string]interface{}{"api_key": debugTestAPIKey, "model": "m"},
			},
			"mcp_servers": []interface{}{
				map[string]interface{}{"name": "a", "headers": map[string]interface{}{"authorization": "Bearer x"}},
			},
		},
	})

	assertJSONEqual(t, redacted, `{
		"appid": "[REDACTED]",
		"properties": {
			"token": "[REDACTED]",
			"asr": {"vendor": "gemini", "params": {"api_key": "[REDACTED]", "model": "m"}},
			"mcp_servers": [{"name": "a", "headers": {"authorization": "[REDACTED]"}}]
		}
	}`)
}

func TestRedactSecretsCoversGeminiURLAndGoogleTTSCredentials(t *testing.T) {
	redacted := RedactSecrets(map[string]interface{}{
		"llm": map[string]interface{}{
			"url": "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:streamGenerateContent?alt=sse&key=" + debugTestAPIKey,
		},
		"tts": map[string]interface{}{
			"params": map[string]interface{}{"credentials": debugTestAPIKey},
		},
	})

	got, err := json.Marshal(redacted)
	if err != nil {
		t.Fatalf("marshal redacted config: %v", err)
	}
	if strings.Contains(string(got), debugTestAPIKey) {
		t.Errorf("live Google key leaked into debug dump: %s", got)
	}
}

func TestRedactSecretsLeavesEmptyValuesVisible(t *testing.T) {
	redacted := RedactSecrets(map[string]interface{}{"params": map[string]interface{}{"api_key": ""}})

	assertJSONEqual(t, redacted, `{"params": {"api_key": ""}}`)
}

func TestRedactSecretsDoesNotMutateInput(t *testing.T) {
	params := map[string]interface{}{"api_key": debugTestAPIKey}
	original := map[string]interface{}{"asr": map[string]interface{}{"params": params}}

	RedactSecrets(original)

	if params["api_key"] != debugTestAPIKey {
		t.Errorf("input mutated: api_key = %v", params["api_key"])
	}
}

func assertJSONEqual(t *testing.T, got interface{}, wantJSON string) {
	t.Helper()

	gotBytes, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal got: %v", err)
	}
	var gotValue interface{}
	if err := json.Unmarshal(gotBytes, &gotValue); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	var wantValue interface{}
	if err := json.Unmarshal([]byte(wantJSON), &wantValue); err != nil {
		t.Fatalf("unmarshal want: %v", err)
	}

	gotNormalized, err := json.Marshal(gotValue)
	if err != nil {
		t.Fatalf("marshal normalized got: %v", err)
	}
	wantNormalized, err := json.Marshal(wantValue)
	if err != nil {
		t.Fatalf("marshal normalized want: %v", err)
	}
	if string(gotNormalized) != string(wantNormalized) {
		t.Errorf("wire shape mismatch\n got: %s\nwant: %s", gotNormalized, wantNormalized)
	}
}
