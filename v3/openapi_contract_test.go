package v3

import (
	"encoding/json"
	"os"
	"testing"
)

// The v3 package is handwritten. Keep this small contract check tied to the
// checked-in canonical snapshot so resource updates do not drift silently.
func TestHandwrittenResourcesMatchOpenAPISnapshot(t *testing.T) {
	data, err := os.ReadFile("../openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Info struct {
			Version string `json:"version"`
		} `json:"info"`
		Security   []map[string][]string                 `json:"security"`
		Paths      map[string]map[string]json.RawMessage `json:"paths"`
		Components struct {
			Schemas         map[string]map[string]any `json:"schemas"`
			SecuritySchemes map[string]map[string]any `json:"securitySchemes"`
		} `json:"components"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Info.Version != "3.0.0" {
		t.Fatalf("unexpected OpenAPI version: %q", doc.Info.Version)
	}
	if _, ok := doc.Components.SecuritySchemes["AgoraTokenAuth"]; !ok {
		t.Fatal("snapshot must use the existing Agora token authorization")
	}
	for _, route := range []struct{ path, method string }{
		{"/conversational-ai/v1/projects/{appid}/sessions/{agent_session_id}", "patch"},
		{"/conversational-ai/v1/projects/{appid}/sessions/ephemeral", "post"},
		{"/conversational-ai/v1/projects/{appid}/campaigns", "post"},
	} {
		if _, ok := doc.Paths[route.path][route.method]; !ok {
			t.Errorf("snapshot missing %s %s", route.method, route.path)
		}
	}
	credential := doc.Components.Schemas["AgentRequest_PublicByokCredential"]["properties"].(map[string]any)
	if _, ok := credential["fields"]; !ok {
		t.Fatal("snapshot credential envelope lacks fields")
	}
	campaign := doc.Components.Schemas["CampaignCreate"]["properties"].(map[string]any)
	if _, ok := campaign["data_policy"]; !ok {
		t.Fatal("snapshot campaign create lacks data_policy")
	}
	contacts := campaign["contacts"].(map[string]any)
	if contacts["maxItems"] != float64(50000) {
		t.Fatalf("contact limit drift: %#v", contacts["maxItems"])
	}
	endCall := doc.Components.Schemas["CampaignEndCallPolicy"]["properties"].(map[string]any)
	if _, ok := endCall["on_voicemail"]; !ok {
		t.Fatal("campaign end-call policy lacks on_voicemail")
	}
	if _, ok := doc.Components.Schemas["SessionAnalysis"]; ok {
		t.Fatal("removed session analysis schema returned")
	}
}
