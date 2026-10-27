package agentkit

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strings"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/AgoraIO/agora-agents-go/v2/option"
	"github.com/AgoraIO/agora-agents-go/v2/v3"
)

func TestDeployMappingAndStrictFields(t *testing.T) {
	a := NewAgent(NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token"}), WithVariable("result", VariableDefinition{Type: "bool", Extract: true, Description: "Resolved?"}))
	a.base.LLM = map[string]interface{}{"vendor": "openai", "api_key": SecretRef("prod", "openai"), "params": map[string]interface{}{"model": "gpt-4o"}, "system_messages": []map[string]string{{"role": "system", "content": "Help"}, {"role": "user", "content": "Hi"}}}
	a.base.TTS = map[string]interface{}{"vendor": "elevenlabs", "params": map[string]interface{}{"key": "tts-key", "voice_id": "v1"}}
	body, err := a.toV3("Support")
	if err != nil {
		t.Fatal(err)
	}
	if body.Instructions == nil || *body.Instructions != "Help" || len(body.Messages) != 1 {
		t.Fatalf("messages not split: %#v", body)
	}
	if body.Pipeline.LLM["model"] != "gpt-4o" {
		t.Fatalf("model missing: %#v", body.Pipeline.LLM)
	}
	credential := body.Pipeline.LLM["credential"].(map[string]any)
	if credential["fields"].(map[string]string)["api_key"] != "$secrets.prod.openai" {
		t.Fatalf("credential missing: %#v", credential)
	}
	if _, ok := body.Pipeline.LLM["api_key"]; ok {
		t.Fatal("v2 api_key leaked")
	}
	if body.Variables["result"].Description != "Resolved?" {
		t.Fatal("variable missing")
	}
	a = a.WithGeofence(&GeofenceConfig{})
	_, err = a.toV3("Support")
	if !errors.Is(err, ErrUnsupportedField) || !strings.Contains(err.Error(), "geofence") {
		t.Fatalf("strict mapping: %v", err)
	}
}

func TestVariableImmutableAndV2SecretGuard(t *testing.T) {
	def := VariableDefinition{Type: "str", Options: []any{"one"}}
	a := NewAgent(NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token"}))
	b := a.WithVariable("choice", def)
	def.Options[0] = "other"
	if len(a.base.Variables) != 0 || b.base.Variables["choice"].Options[0] != "one" {
		t.Fatal("variable option mutated original")
	}
	b.base.LLM = map[string]interface{}{"api_key": SecretRef("prod", "openai")}
	_, err := b.ToPropertiesMap(ToPropertiesOptions{})
	if !errors.Is(err, ErrSecretRefNotSupported) {
		t.Fatalf("v2 guard: %v", err)
	}
}

func TestV3AvatarCredentialsLeaveParams(t *testing.T) {
	for _, tc := range []struct {
		vendor string
		key    string
	}{
		{vendor: "sensetime", key: "app_key"},
		{vendor: "spatius", key: "spatius_api_key"},
	} {
		t.Run(tc.vendor, func(t *testing.T) {
			secret := "provider-secret"
			converted, _ := convertModule("avatar", map[string]interface{}{
				"vendor": tc.vendor,
				"params": map[string]interface{}{tc.key: secret},
			})
			params, _ := converted["params"].(map[string]interface{})
			if _, leaked := params[tc.key]; leaked {
				t.Fatalf("provider credential remained in params: %#v", params)
			}
			credential, _ := converted["credential"].(map[string]any)
			if credential["fields"].(map[string]string)[tc.key] != secret {
				t.Fatalf("provider credential was not extracted: %#v", credential)
			}

			a := NewAgent(NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token"}))
			a.base.Avatar = map[string]interface{}{"vendor": tc.vendor, "params": map[string]interface{}{tc.key: secret}}
			body, err := a.toV3("supported avatar")
			if err != nil {
				t.Fatalf("supported v3 avatar vendor should map: %v", err)
			}
			avatar := body.Pipeline.Avatar
			if avatar["vendor"] != tc.vendor || avatar["credential"].(map[string]any)["fields"].(map[string]string)[tc.key] != secret {
				t.Fatalf("avatar credentials were not mapped: %#v", avatar)
			}
			if params, _ := avatar["params"].(map[string]any); params != nil {
				if _, leaked := params[tc.key]; leaked {
					t.Fatalf("provider credential leaked into v3 params: %#v", params)
				}
			}
		})
	}
}

func TestV3CredentialFieldsUseProviderNames(t *testing.T) {
	cases := []struct {
		module, vendor, source, want string
	}{
		{"asr", "microsoft", "api_key", "key"},
		{"asr", "google", "api_key", "adc_credentials_string"},
		{"tts", "elevenlabs", "api_key", "key"},
		{"tts", "google", "api_key", "credentials"},
		{"tts", "sarvam", "api_key", "api_subscription_key"},
	}
	for _, tc := range cases {
		t.Run(tc.module+"/"+tc.vendor, func(t *testing.T) {
			converted, bad := convertModule(tc.module, map[string]interface{}{"vendor": tc.vendor, "api_key": "provider-secret"})
			if len(bad) > 0 {
				t.Fatalf("conversion rejected: %v", bad)
			}
			credential := converted["credential"].(map[string]any)
			fields := credential["fields"].(map[string]string)
			if fields[tc.want] != "provider-secret" {
				t.Fatalf("wrong credential field: %#v", fields)
			}
			if _, exists := fields[tc.source]; tc.source != tc.want && exists {
				t.Fatalf("generic key leaked: %#v", fields)
			}
		})
	}
}

func TestAmazonBedrockV3MappingUsesAmazonCredentialsAndRegionParam(t *testing.T) {
	config := vendors.NewAmazonBedrock(vendors.AmazonBedrockOptions{AccessKey: "access", SecretKey: "secret", Region: "us-east-1", Model: "foundation-model"})
	a := NewAgent(NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token"})).WithLlm(config)
	body, err := a.toV3("Bedrock")
	if err != nil {
		t.Fatal(err)
	}
	llm := body.Pipeline.LLM
	if llm["vendor"] != "amazon" || llm["style"] != "bedrock" {
		t.Fatalf("Bedrock provider mapping: %#v", llm)
	}
	fields := llm["credential"].(map[string]any)["fields"].(map[string]string)
	if fields["access_key"] != "access" || fields["secret_key"] != "secret" {
		t.Fatalf("Bedrock credentials: %#v", fields)
	}
	if llm["params"].(map[string]any)["region"] != "us-east-1" {
		t.Fatalf("Bedrock region missing: %#v", llm["params"])
	}
}

func TestRedeployClearsOnlyOldVendorCredentialFields(t *testing.T) {
	old := v3.Agent{AgentID: "id", AgentName: "Old", Pipeline: []byte(`{"mode":"cascaded","llm":{"vendor":"openai","credential":{"mode":"byok"}}}`)}
	next := v3.AgentCreate{AgentName: "New", Pipeline: &v3.Pipeline{Mode: "cascaded", LLM: map[string]any{"vendor": "amazon", "credential": map[string]any{"mode": "byok", "fields": map[string]string{"access_key": "$secrets.prod.access", "secret_key": "$secrets.prod.secret"}}}}}
	patch, err := deployPatch(old, next)
	if err != nil {
		t.Fatal(err)
	}
	fields := patch["pipeline"].(map[string]any)["llm"].(map[string]any)["credential"].(map[string]any)["fields"].(map[string]any)
	if fields["api_key"] != nil || fields["access_key"] != "$secrets.prod.access" || fields["secret_key"] != "$secrets.prod.secret" || len(fields) != 3 {
		t.Fatalf("credential patch has stale or unnecessary fields: %#v", fields)
	}
}

func TestRedeployClearsNestedStaleFieldsAndKeepsMetadataOut(t *testing.T) {
	old := v3.Agent{AgentID: "id", AgentName: "Old", Pipeline: []byte(`{"mode":"cascaded","llm":{"vendor":"openai","params":{"old":1,"keep":2},"credential":{"mode":"byok","resolved_key_id":"server"}}}`)}
	next := v3.AgentCreate{AgentName: "New", Pipeline: &v3.Pipeline{Mode: "cascaded", LLM: map[string]any{"vendor": "openai", "params": map[string]any{"keep": 2}, "credential": map[string]any{"mode": "managed"}}}}
	patch, err := deployPatch(old, next)
	if err != nil {
		t.Fatal(err)
	}
	pipeline := patch["pipeline"].(map[string]any)
	llm := pipeline["llm"].(map[string]any)
	params := llm["params"].(map[string]any)
	if value, ok := params["old"]; !ok || value != nil {
		t.Fatalf("stale param not cleared: %#v", params)
	}
	cred := llm["credential"].(map[string]any)
	if value, ok := cred["fields"]; !ok || value != nil {
		t.Fatalf("managed reset absent: %#v", cred)
	}
	if _, ok := cred["resolved_key_id"]; ok {
		t.Fatal("server metadata leaked")
	}
	if _, ok := patch["agent_id"]; ok {
		t.Fatal("server ID leaked")
	}
}

func TestActualRealtimeVendorTurnDetection(t *testing.T) {
	mode := Agora.MllmTurnDetectionMode("server_vad")
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token"})
	a := NewAgent(c).WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{APIKey: "key", Instructions: "vendor instruction", TurnDetection: &Agora.MllmTurnDetection{Mode: &mode}}))
	body, err := a.toV3("Realtime")
	if err != nil {
		t.Fatal(err)
	}
	if body.Pipeline.Mode != "realtime" || body.Pipeline.TurnDetection["type"] != "server_vad" {
		t.Fatalf("turn detection: %#v", body.Pipeline.TurnDetection)
	}
	if body.Instructions == nil || *body.Instructions != "vendor instruction" {
		t.Fatalf("instructions: %#v", body.Instructions)
	}
	if params, ok := body.Pipeline.MLLM["params"].(map[string]interface{}); ok {
		if _, exists := params["instructions"]; exists {
			t.Fatal("instructions duplicated in params")
		}
	}
	def := "default"
	b := NewAgent(c).WithTurnDetection(&TurnDetectionConfig{Mode: &def})
	if _, err = b.toV3("Default turn"); err != nil {
		t.Fatalf("modern default turn mode: %v", err)
	}
}

func TestVariableOptionsRequireJSONSemanticUniqueness(t *testing.T) {
	cases := []struct {
		name    string
		options []any
		invalid bool
	}{
		{"strings", []any{"same", "same"}, true},
		{"booleans", []any{true, true}, true},
		{"integer and float", []any{1, float64(1)}, true},
		{"integer widths", []any{int8(1), uint64(1)}, true},
		{"signed zero", []any{0, math.Copysign(0, -1)}, true},
		{"serialized floats", []any{float32(0.1), float64(0.1)}, true},
		{"different JSON types", []any{"1", 1, true}, false},
		{"distinct large integers", []any{uint64(9007199254740993), float64(9007199254740992)}, false},
		{"scientific notation", []any{1e-7, 2e-7}, false},
		{"nonfinite", []any{math.NaN()}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			def := VariableDefinition{Type: "str", Options: tc.options}
			for _, build := range []func(){
				func() { WithVariable("choice", def) },
				func() {
					NewAgent(NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token"})).WithVariable("choice", def)
				},
			} {
				panicked := false
				func() {
					defer func() { panicked = recover() != nil }()
					build()
				}()
				if panicked != tc.invalid {
					t.Fatalf("constructor panic = %t, want %t", panicked, tc.invalid)
				}
			}
			// A custom option or externally constructed configuration can bypass
			// the fluent validator. Deploy must apply the same validation.
			calls := 0
			client := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				return v3Reply(201, `{"agent_id":"id","agent_name":"Choice","create_ts":1,"update_ts":1}`, "e1"), nil
			})})
			agent := NewAgent(client)
			agent.base.Variables = map[string]VariableDefinition{"choice": def}
			_, err := agent.Deploy(context.Background(), DeployOptions{Name: "Choice"})
			if tc.invalid {
				if !errors.Is(err, ErrUnsupportedField) || !strings.Contains(err.Error(), "variables.choice") || calls != 0 {
					t.Fatalf("deploy err=%v calls=%d", err, calls)
				}
			} else if err != nil || calls != 1 {
				t.Fatalf("valid deploy err=%v calls=%d", err, calls)
			}
		})
	}
}
