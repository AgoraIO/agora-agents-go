package agentkit

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/AgoraIO/agora-agents-go/v2/option"
)

// Wire-shape expectations are copied from the TypeScript suite so the three
// SDKs stay byte-identical on the wire.

const previewAPIKey = "test-google-api-key"

// recordingClient captures every outgoing request and answers with a generic
// success body.
type recordingClient struct {
	mu       sync.Mutex
	requests []*http.Request
	bodies   [][]byte
}

func (r *recordingClient) Do(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
	}
	r.requests = append(r.requests, req.Clone(req.Context()))
	r.bodies = append(r.bodies, body)
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"agent_id":"agent-1","data":{"list":[]}}`)),
		Header:     make(http.Header),
	}, nil
}

func newTestPreviewClient(rec *recordingClient, features ...string) *AgoraClient {
	return NewAgoraPreviewClient(AgoraPreviewClientOptions{
		AgoraClientOptions: AgoraClientOptions{
			Area:           option.AreaUS,
			AppID:          "81190c52971d4004b7244bdcd93e2f34",
			AppCertificate: "0123456789abcdef0123456789abcdef",
		},
		Features:   features,
		HTTPClient: rec,
	})
}

// --- Vendor wire shapes -----------------------------------------------------

func TestWalkieTalkieSerialisesDocumentedMLLMShape(t *testing.T) {
	config := vendors.NewGeminiWalkieTalkie(vendors.GeminiPreviewMLLMOptions{
		APIKey: previewAPIKey,
	}).ToConfig()

	assertJSONEqual(t, config, `{
		"vendor": "gemini",
		"api_key": "test-google-api-key",
		"url": "https://generativelanguage.googleapis.com",
		"params": {
			"api_key": "test-google-api-key",
			"model": "walkie-talkie",
			"voice": "Puck"
		}
	}`)
}

func TestCleverChatterIncludesThinkingLevel(t *testing.T) {
	config := vendors.NewGeminiCleverChatter(vendors.GeminiCleverChatterOptions{
		GeminiPreviewMLLMOptions: vendors.GeminiPreviewMLLMOptions{
			APIKey: previewAPIKey,
			Voice:  "Puck",
		},
		ThinkingLevel: vendors.GeminiThinkingLevelMedium,
	}).ToConfig()

	assertJSONEqual(t, config, `{
		"vendor": "gemini",
		"api_key": "test-google-api-key",
		"url": "https://generativelanguage.googleapis.com",
		"params": {
			"api_key": "test-google-api-key",
			"model": "clever-chatter",
			"voice": "Puck",
			"thinking_level": "medium"
		}
	}`)
}

func TestPreviewMLLMCarriesInstructionsAndTranscriptionFlags(t *testing.T) {
	enabled := true
	config := vendors.NewGeminiWalkieTalkie(vendors.GeminiPreviewMLLMOptions{
		APIKey:          previewAPIKey,
		Instructions:    "You are Ada.",
		TranscribeAgent: &enabled,
		TranscribeUser:  &enabled,
	}).ToConfig()

	params := config["params"].(map[string]interface{})
	if params["instructions"] != "You are Ada." {
		t.Errorf("instructions = %v, want %q", params["instructions"], "You are Ada.")
	}
	if params["transcribe_agent"] != true || params["transcribe_user"] != true {
		t.Errorf("transcription flags = %v / %v, want true / true",
			params["transcribe_agent"], params["transcribe_user"])
	}
}

func TestSendsGreetingNeverGreetingMessage(t *testing.T) {
	// The preview Gemini models read "greeting". "greeting_message" is what the
	// production providers use.
	withGreeting := vendors.NewGeminiWalkieTalkie(vendors.GeminiPreviewMLLMOptions{
		APIKey:   previewAPIKey,
		Greeting: "Hi there.",
	}).ToConfig()
	if withGreeting["greeting"] != "Hi there." {
		t.Errorf("greeting = %v, want %q", withGreeting["greeting"], "Hi there.")
	}
	if _, found := withGreeting["greeting_message"]; found {
		t.Error("greeting_message is the production spelling; preview must not send it")
	}

	withoutGreeting := vendors.NewGeminiWalkieTalkie(vendors.GeminiPreviewMLLMOptions{
		APIKey: previewAPIKey,
	}).ToConfig()
	for _, key := range []string{"greeting", "greeting_message"} {
		if _, found := withoutGreeting[key]; found {
			t.Errorf("%s should be omitted when no greeting is set", key)
		}
	}
}

func TestCleverChatterRejectsInvalidThinkingLevel(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected a panic for an invalid thinking level")
		}
	}()
	vendors.NewGeminiCleverChatter(vendors.GeminiCleverChatterOptions{
		GeminiPreviewMLLMOptions: vendors.GeminiPreviewMLLMOptions{APIKey: previewAPIKey},
		ThinkingLevel:            "not-a-level",
	})
}

// --- Routing ----------------------------------------------------------------

func TestPreviewClientReportsPreviewBaseURL(t *testing.T) {
	client := newTestPreviewClient(&recordingClient{})

	if got := client.CurrentURL(); got != PreviewAPIBaseURL {
		t.Errorf("CurrentURL() = %q, want %q", got, PreviewAPIBaseURL)
	}
}

func TestPreviewClientAcceptsBaseURLOverride(t *testing.T) {
	client := NewAgoraPreviewClient(AgoraPreviewClientOptions{
		AgoraClientOptions: AgoraClientOptions{
			Area:           option.AreaUS,
			AppID:          "81190c52971d4004b7244bdcd93e2f34",
			AppCertificate: "0123456789abcdef0123456789abcdef",
		},
		BaseURL: "https://staging.example.com/preview/api/conversational-ai-agent",
	})

	if got := client.CurrentURL(); got != "https://staging.example.com/preview/api/conversational-ai-agent" {
		t.Errorf("CurrentURL() = %q, want the staging override", got)
	}
}

func TestPreviewClientSendsGateHeader(t *testing.T) {
	rec := &recordingClient{}
	client := newTestPreviewClient(rec)

	_ = client.StopAgent(context.Background(), "agent-1")

	if len(rec.requests) == 0 {
		t.Fatal("no request captured")
	}
	req := rec.requests[0]
	if got := req.Header.Get(PreviewFeatureHeader); got != "gemini-live" {
		t.Errorf("%s = %q, want %q", PreviewFeatureHeader, got, "gemini-live")
	}
	if !strings.HasPrefix(req.URL.String(), PreviewAPIBaseURL) {
		t.Errorf("request went to %q, want the preview host", req.URL)
	}
	// Exactly one header may carry the preview value — the gateway accepts other
	// spellings that are not part of the public contract, and none may ship here.
	var carriers []string
	for name, values := range req.Header {
		for _, value := range values {
			if strings.Contains(value, "gemini-live") {
				carriers = append(carriers, name)
			}
		}
	}
	if len(carriers) != 1 || !strings.EqualFold(carriers[0], PreviewFeatureHeader) {
		t.Errorf("preview value carried by %v, want only %s", carriers, PreviewFeatureHeader)
	}
}

func TestPerCallHeaderOptionCannotDropTheGate(t *testing.T) {
	// option.WithHTTPHeader replaces the whole header map, so the gate is pinned
	// below the option layer. A request that loses it is routed to production,
	// where the preview providers do not exist.
	rec := &recordingClient{}
	client := newTestPreviewClient(rec)

	custom := make(http.Header)
	custom.Set("x-custom", "kept")
	_ = client.Agents.Stop(context.Background(), &Agora.StopAgentsRequest{
		Appid:   "81190c52971d4004b7244bdcd93e2f34",
		AgentID: "agent-1",
	}, option.WithHTTPHeader(custom))

	if len(rec.requests) == 0 {
		t.Fatal("no request captured")
	}
	req := rec.requests[0]
	if got := req.Header.Get(PreviewFeatureHeader); got != "gemini-live" {
		t.Errorf("%s = %q, want it to survive a per-call header option", PreviewFeatureHeader, got)
	}
	if got := req.Header.Get("x-custom"); got != "kept" {
		t.Errorf("x-custom = %q, want %q", got, "kept")
	}
}

func TestPreviewClientHonoursCustomFeatureList(t *testing.T) {
	rec := &recordingClient{}
	client := newTestPreviewClient(rec, "gemini-live", "another-preview")

	_ = client.StopAgent(context.Background(), "agent-1")

	if got := rec.requests[0].Header.Get(PreviewFeatureHeader); got != "gemini-live,another-preview" {
		t.Errorf("%s = %q, want the joined feature list", PreviewFeatureHeader, got)
	}
}

// --- Preview support guard --------------------------------------------------

func TestRequiredPreviewFeaturesFlagsPreviewMLLMModels(t *testing.T) {
	mllm := map[string]interface{}{"mllm": map[string]interface{}{
		"vendor": "gemini",
		"params": map[string]interface{}{"model": "walkie-talkie"},
	}}
	if got := RequiredPreviewFeatures(mllm); len(got) != 1 || got[0] != PreviewFeatureGeminiLive {
		t.Errorf("RequiredPreviewFeatures(walkie-talkie) = %v, want [gemini-live]", got)
	}
}

func TestRequiredPreviewFeaturesRecognisesBothSidesOfTheRename(t *testing.T) {
	// Google is renaming the codenames; the guard has to keep working through
	// the rollout whichever spelling a caller sends.
	models := []string{
		vendors.GeminiPreviewModelWalkieTalkie,
		vendors.GeminiPreviewModelCleverChatter,
		vendors.GeminiPreviewModel35Live,
		vendors.GeminiPreviewModel35LiveThinking,
	}
	for _, model := range models {
		properties := map[string]interface{}{"mllm": map[string]interface{}{
			"vendor": "gemini",
			"params": map[string]interface{}{"model": model},
		}}
		if got := RequiredPreviewFeatures(properties); len(got) != 1 || got[0] != PreviewFeatureGeminiLive {
			t.Errorf("RequiredPreviewFeatures(%s) = %v, want [gemini-live]", model, got)
		}
	}
}

func TestRequiredPreviewFeaturesLeavesGAGeminiLiveAlone(t *testing.T) {
	ga := map[string]interface{}{"mllm": map[string]interface{}{
		"vendor": "gemini",
		"params": map[string]interface{}{"model": "gemini-live-2.5-flash-preview-native-audio-09-2025"},
	}}

	if got := RequiredPreviewFeatures(ga); len(got) != 0 {
		t.Errorf("RequiredPreviewFeatures(GA gemini live) = %v, want none", got)
	}
}

func TestAssertPreviewSupportRejectsProductionClient(t *testing.T) {
	properties := map[string]interface{}{"mllm": map[string]interface{}{
		"vendor": "gemini",
		"params": map[string]interface{}{"model": "walkie-talkie"},
	}}

	err := AssertPreviewSupport(nil, properties)
	if err == nil {
		t.Fatal("expected an error for a production client")
	}
	if !strings.Contains(err.Error(), "require the preview endpoint") {
		t.Errorf("error = %q, want it to name the preview endpoint", err)
	}

	if err := AssertPreviewSupport([]string{PreviewFeatureGeminiLive}, properties); err != nil {
		t.Errorf("preview client rejected: %v", err)
	}
}

// --- Debug redaction --------------------------------------------------------

func TestRedactSecretsWalksNestedStructures(t *testing.T) {
	redacted := RedactSecrets(map[string]interface{}{
		"appid": "81190c52971d4004b7244bdcd93e2f34",
		"properties": map[string]interface{}{
			"token": "007eJxTYKhsrH10",
			"asr": map[string]interface{}{
				"vendor": "gemini",
				"params": map[string]interface{}{"api_key": previewAPIKey, "model": "m"},
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

func TestRedactSecretsLeavesEmptyValuesVisible(t *testing.T) {
	// "" is the signature of an unset env var and must stay diagnosable.
	redacted := RedactSecrets(map[string]interface{}{"params": map[string]interface{}{"api_key": ""}})

	assertJSONEqual(t, redacted, `{"params": {"api_key": ""}}`)
}

func TestRedactSecretsDoesNotMutateInput(t *testing.T) {
	params := map[string]interface{}{"api_key": previewAPIKey}
	original := map[string]interface{}{"asr": map[string]interface{}{"params": params}}

	RedactSecrets(original)

	if params["api_key"] != previewAPIKey {
		t.Errorf("input mutated: api_key = %v", params["api_key"])
	}
}

// --- helpers ----------------------------------------------------------------

func assertJSONEqual(t *testing.T, got interface{}, wantJSON string) {
	t.Helper()

	gotBytes, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal got: %v", err)
	}
	var gotAny, wantAny interface{}
	if err := json.Unmarshal(gotBytes, &gotAny); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	if err := json.Unmarshal([]byte(wantJSON), &wantAny); err != nil {
		t.Fatalf("unmarshal want: %v", err)
	}
	gotNorm, _ := json.Marshal(gotAny)
	wantNorm, _ := json.Marshal(wantAny)
	if string(gotNorm) != string(wantNorm) {
		t.Errorf("wire shape mismatch\n got: %s\nwant: %s", gotNorm, wantNorm)
	}
}

func TestPreviewMLLMSendsLanguageCodesNeverSingularLanguage(t *testing.T) {
	// params.language belongs to xAI Grok in the Agora schema, and the production
	// Gemini Live provider sends no language field at all. The preview MLLMs take
	// the plural array.
	config := vendors.NewGeminiWalkieTalkie(vendors.GeminiPreviewMLLMOptions{
		APIKey:        previewAPIKey,
		LanguageCodes: []string{"en-US"},
	}).ToConfig()

	params := config["params"].(map[string]interface{})
	assertJSONEqual(t, params["language_codes"], `["en-US"]`)
	if _, found := params["language"]; found {
		t.Error("params.language is xAI's field; the preview MLLM must not send it")
	}
}

func TestPreviewMLLMOmitsLanguageCodesWhenUnset(t *testing.T) {
	config := vendors.NewGeminiWalkieTalkie(vendors.GeminiPreviewMLLMOptions{
		APIKey: previewAPIKey,
	}).ToConfig()

	params := config["params"].(map[string]interface{})
	for _, key := range []string{"language", "language_codes"} {
		if _, found := params[key]; found {
			t.Errorf("%s should be omitted when LanguageCodes is nil", key)
		}
	}
}

// --- Preview wire-shape translation -----------------------------------------

func previewMLLMProperties(extra map[string]interface{}) map[string]interface{} {
	mllm := map[string]interface{}{
		"vendor": "gemini",
		"params": map[string]interface{}{"model": "walkie-talkie"},
	}
	for k, v := range extra {
		mllm[k] = v
	}
	return map[string]interface{}{"mllm": mllm}
}

func TestAgentLevelGreetingIsRetargetedToGreeting(t *testing.T) {
	// The Agent writes agent-level greetings to greeting_message, the production
	// spelling. Without this step the preview model never greets.
	properties := previewMLLMProperties(map[string]interface{}{"greeting_message": "AGENT"})

	ApplyPreviewShape(properties)

	mllm := properties["mllm"].(map[string]interface{})
	if mllm["greeting"] != "AGENT" {
		t.Errorf("greeting = %v, want AGENT", mllm["greeting"])
	}
	if _, found := mllm["greeting_message"]; found {
		t.Error("greeting_message should have been retargeted, not kept")
	}
}

func TestHandWrittenGreetingMessageMigratesOntoGreeting(t *testing.T) {
	// Lets an existing MLLM config keep working after only swapping the model.
	properties := previewMLLMProperties(map[string]interface{}{"greeting_message": "EXISTING"})

	ApplyPreviewShape(properties)

	mllm := properties["mllm"].(map[string]interface{})
	if mllm["greeting"] != "EXISTING" {
		t.Errorf("greeting = %v, want EXISTING", mllm["greeting"])
	}
}

func TestVendorGreetingWinsOverAgentLevel(t *testing.T) {
	properties := previewMLLMProperties(map[string]interface{}{
		"greeting":         "VENDOR",
		"greeting_message": "AGENT",
	})

	ApplyPreviewShape(properties)

	mllm := properties["mllm"].(map[string]interface{})
	if mllm["greeting"] != "VENDOR" {
		t.Errorf("greeting = %v, want VENDOR", mllm["greeting"])
	}
	if _, found := mllm["greeting_message"]; found {
		t.Error("greeting_message should be dropped once retargeted")
	}
}

func TestFailureMessageIsLeftAlone(t *testing.T) {
	// failure_message is an Agora engine feature, not a Gemini one.
	properties := previewMLLMProperties(map[string]interface{}{"failure_message": "One moment."})

	ApplyPreviewShape(properties)

	mllm := properties["mllm"].(map[string]interface{})
	if mllm["failure_message"] != "One moment." {
		t.Errorf("failure_message = %v, want it untouched", mllm["failure_message"])
	}
}

func TestProductionMLLMIsLeftUntouched(t *testing.T) {
	properties := map[string]interface{}{"mllm": map[string]interface{}{
		"vendor":           "gemini",
		"greeting_message": "PROD",
		"params":           map[string]interface{}{"model": "gemini-2.0-flash-live-001"},
	}}

	ApplyPreviewShape(properties)

	mllm := properties["mllm"].(map[string]interface{})
	if mllm["greeting_message"] != "PROD" {
		t.Errorf("greeting_message = %v, want PROD", mllm["greeting_message"])
	}
	if _, found := mllm["greeting"]; found {
		t.Error("a production MLLM must not gain the preview spelling")
	}
}

func TestApplyPreviewShapeIsANoOpWithoutMLLM(t *testing.T) {
	properties := map[string]interface{}{"asr": map[string]interface{}{"vendor": "gemini"}}

	ApplyPreviewShape(properties)

	if len(properties) != 1 {
		t.Errorf("properties = %v, want unchanged", properties)
	}
}

func TestStartedRequestCarriesGreetingNotGreetingMessage(t *testing.T) {
	// End-to-end: proves ApplyPreviewShape is wired into Start(), not merely
	// exported. An agent-level greeting is the natural usage and must reach the
	// preview model.
	rec := &recordingClient{}
	client := newTestPreviewClient(rec)

	agent := NewAgent(client, WithGreeting("AGENT")).
		WithMllm(vendors.NewGeminiWalkieTalkie(vendors.GeminiPreviewMLLMOptions{APIKey: previewAPIKey}))
	session := agent.CreateSession(CreateSessionOptions{
		Channel: "c", AgentUID: "1", RemoteUIDs: []string{"100"},
	})
	if _, err := session.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	if len(rec.bodies) == 0 {
		t.Fatal("no request captured")
	}

	var body struct {
		Properties struct {
			Mllm map[string]interface{} `json:"mllm"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(rec.bodies[0], &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Properties.Mllm["greeting"] != "AGENT" {
		t.Errorf("greeting = %v, want AGENT", body.Properties.Mllm["greeting"])
	}
	if _, found := body.Properties.Mllm["greeting_message"]; found {
		t.Error("greeting_message must not reach the preview route")
	}
}

func TestUnrecognisedModelOnAPreviewVendorStillGetsTheGreetingRename(t *testing.T) {
	// Overriding Model is the documented escape hatch for a Google rename that
	// lands ahead of an SDK release, so a name this SDK has never seen is
	// reachable by following its own advice. Matching on the model name alone
	// made that case fail silently: no rename, and an agent-level greeting would
	// land in a field these models ignore.
	rec := &recordingClient{}
	client := newTestPreviewClient(rec)

	agent := NewAgent(client, WithGreeting("AGENT")).
		WithMllm(vendors.NewGeminiWalkieTalkie(vendors.GeminiPreviewMLLMOptions{
			APIKey: previewAPIKey,
			Model:  "gemini-4.0-live-unreleased",
		}))
	session := agent.CreateSession(CreateSessionOptions{
		Channel: "c", AgentUID: "1", RemoteUIDs: []string{"100"},
	})
	if _, err := session.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	if len(rec.bodies) == 0 {
		t.Fatal("no request captured")
	}

	var body struct {
		Properties struct {
			Mllm map[string]interface{} `json:"mllm"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(rec.bodies[0], &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Properties.Mllm["greeting"] != "AGENT" {
		t.Errorf("greeting = %v, want AGENT", body.Properties.Mllm["greeting"])
	}
	if _, found := body.Properties.Mllm["greeting_message"]; found {
		t.Error("greeting_message must not reach the preview route")
	}
}

func TestUnrecognisedModelOnAPreviewVendorStillRequiresThePreviewEndpoint(t *testing.T) {
	properties := map[string]interface{}{
		"mllm": vendors.NewGeminiWalkieTalkie(vendors.GeminiPreviewMLLMOptions{
			APIKey: previewAPIKey,
			Model:  "gemini-4.0-live-unreleased",
		}).ToConfig(),
	}

	features := RequiredPreviewFeatures(properties)
	if len(features) != 1 || features[0] != PreviewFeatureGeminiLive {
		t.Errorf("features = %v, want [%s]", features, PreviewFeatureGeminiLive)
	}
}

func TestEnvelopeCheckDoesNotCatchAProductionGeminiLiveConfig(t *testing.T) {
	// GeminiLive leaves params.api_key unset and does not point url at the
	// Gemini Developer API host, so widening detection must not retarget
	// greeting_message for a GA caller.
	properties := map[string]interface{}{"mllm": map[string]interface{}{
		"vendor":           "gemini",
		"api_key":          "prod-key",
		"url":              "wss://generativelanguage.googleapis.com/ws",
		"greeting_message": "PROD",
		"params":           map[string]interface{}{"model": "gemini-2.0-flash-live-001"},
	}}

	ApplyPreviewShape(properties)

	mllm := properties["mllm"].(map[string]interface{})
	if mllm["greeting_message"] != "PROD" {
		t.Errorf("greeting_message = %v, want PROD", mllm["greeting_message"])
	}
	if _, found := mllm["greeting"]; found {
		t.Error("a production MLLM must not gain the preview spelling")
	}
	if features := RequiredPreviewFeatures(properties); len(features) != 0 {
		t.Errorf("features = %v, want none", features)
	}
}
