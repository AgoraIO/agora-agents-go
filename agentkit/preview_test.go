package agentkit

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
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

func newTestPreviewClient(rec *recordingClient) *AgoraClient {
	return NewAgoraClient(AgoraClientOptions{
		Area:           option.AreaUS,
		AppID:          "81190c52971d4004b7244bdcd93e2f34",
		AppCertificate: "0123456789abcdef0123456789abcdef",
		HTTPClient:     rec,
	})
}

// --- Vendor wire shapes -----------------------------------------------------

func TestGeminiSTTSerialisesDocumentedASRShape(t *testing.T) {
	config := vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
		APIKey:        previewAPIKey,
		LanguageCodes: []string{"en-US"},
	}).ToConfig()

	assertJSONEqual(t, config, `{
		"vendor": "gemini",
		"params": {
			"api_key": "test-google-api-key",
			"model": "gemini-3.5-transcribe-live",
			"sample_rate": 16000,
			"language_hints": ["en-US"]
		}
	}`)
}

func TestEmitsNoTopLevelLanguageOfItsOwn(t *testing.T) {
	// Every STT vendor leaves asr.language to the Agent, which derives it from
	// the turn detection language. A vendor-level copy would be a no-op the
	// builder overwrites, so this one does not offer the option at all.
	config := vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
		APIKey: previewAPIKey,
	}).ToConfig()

	if _, found := config["language"]; found {
		t.Error("asr.language is the Agent's to set, not the vendor's")
	}
	params := config["params"].(map[string]interface{})
	if _, found := params["language"]; found {
		t.Error("params.language should not be sent; language_codes replaced it")
	}
	// Nor does it invent language_codes — absent means auto-detect.
	if _, found := params["language_codes"]; found {
		t.Error("language_codes should be omitted unless the caller sets it")
	}
}

func TestLegacyLanguageCodesMapToGALanguageHints(t *testing.T) {
	single := vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
		APIKey:        previewAPIKey,
		LanguageCodes: []string{"es-ES"},
	}).ToConfig()
	assertJSONEqual(t, single["params"].(map[string]interface{})["language_hints"], `["es-ES"]`)

	multiple := vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
		APIKey:        previewAPIKey,
		LanguageCodes: []string{"en-US", "es-ES"},
	}).ToConfig()
	assertJSONEqual(t, multiple["params"].(map[string]interface{})["language_hints"], `["en-US","es-ES"]`)
}

func TestExplicitEmptyLanguageCodesStillReachesTheWire(t *testing.T) {
	// `[]` is the caller spelling auto-detect outright. Nil versus empty is the
	// distinction: nil omits the field, empty-non-nil sends [].
	config := vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
		APIKey:        previewAPIKey,
		LanguageCodes: []string{},
	}).ToConfig()

	assertJSONEqual(t, config["params"].(map[string]interface{})["language_hints"], `[]`)
}

func TestCustomVocabularyIsSentOnlyWhenSupplied(t *testing.T) {
	withVocab := vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
		APIKey:           previewAPIKey,
		CustomVocabulary: []string{"Agora", "Kubernetes"},
	}).ToConfig()
	assertJSONEqual(t, withVocab["params"].(map[string]interface{})["custom_vocabulary"], `["Agora","Kubernetes"]`)
	if _, found := withVocab["params"].(map[string]interface{})["word_timestamp"]; found {
		t.Error("word_timestamp should be omitted unless explicitly supplied")
	}

	withoutVocab := vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
		APIKey: previewAPIKey,
	}).ToConfig()
	if _, found := withoutVocab["params"].(map[string]interface{})["custom_vocabulary"]; found {
		t.Error("custom_vocabulary should be omitted when not supplied")
	}
}

func TestWordTimestampIsSentOnlyWhenExplicitlySupplied(t *testing.T) {
	withoutTimestamp := vendors.NewGeminiSTT(vendors.GeminiSTTOptions{APIKey: previewAPIKey}).ToConfig()
	if _, found := withoutTimestamp["params"].(map[string]interface{})["word_timestamp"]; found {
		t.Error("word_timestamp should be omitted unless explicitly supplied")
	}

	explicitTrue := true
	withTimestamp := vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
		APIKey:        previewAPIKey,
		WordTimestamp: &explicitTrue,
	}).ToConfig()
	if got := withTimestamp["params"].(map[string]interface{})["word_timestamp"]; got != true {
		t.Fatalf("word_timestamp = %v, want true", got)
	}
}

func TestCustomVocabularyRejectsEnabledWordTimestamps(t *testing.T) {
	explicitTrue := true
	tests := map[string]vendors.GeminiSTTOptions{
		"typed options": {
			APIKey:           previewAPIKey,
			CustomVocabulary: []string{"Agora"},
			WordTimestamp:    &explicitTrue,
		},
	}

	for name, options := range tests {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "CustomVocabulary cannot be used with WordTimestamp=true" {
					t.Fatalf("panic = %v, want incompatible-parameter error", got)
				}
			}()
			vendors.NewGeminiSTT(options).ToConfig()
		})
	}
}

func TestCustomVocabularyAllowsExplicitlyDisabledWordTimestamps(t *testing.T) {
	explicitFalse := false
	config := vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
		APIKey:           previewAPIKey,
		CustomVocabulary: []string{"Agora"},
		WordTimestamp:    &explicitFalse,
	}).ToConfig()
	params := config["params"].(map[string]interface{})

	assertJSONEqual(t, params["custom_vocabulary"], `["Agora"]`)
	if got := params["word_timestamp"]; got != false {
		t.Fatalf("word_timestamp = %v, want false", got)
	}
}

// --- Routing ----------------------------------------------------------------

func newGeminiASRSession(rec *recordingClient) *AgentSession {
	client := newTestPreviewClient(rec)
	agent := NewAgent(client).
		WithStt(vendors.NewGeminiSTT(vendors.GeminiSTTOptions{APIKey: previewAPIKey})).
		WithLlm(vendors.NewGemini(vendors.GeminiOptions{
			APIKey: previewAPIKey,
			Model:  "gemini-2.0-flash",
		})).
		WithTts(vendors.NewGoogleTTS(vendors.GoogleTTSOptions{
			Key:          previewAPIKey,
			VoiceName:    "en-US-Chirp3-HD-Charon",
			LanguageCode: "en-US",
		}))
	return agent.CreateSession(CreateSessionOptions{
		Channel: "preview", AgentUID: "1", RemoteUIDs: []string{"100"},
	})
}

func TestGeminiASRSessionLifecycleUsesProductionRouting(t *testing.T) {
	rec := &recordingClient{}
	session := newGeminiASRSession(rec)
	ctx := context.Background()
	if _, err := session.Start(ctx); err != nil {
		t.Fatal(err)
	}
	var startBody map[string]interface{}
	if err := json.Unmarshal(rec.bodies[0], &startBody); err != nil {
		t.Fatalf("decode start body: %v", err)
	}
	if got := startBody["appid"]; got != "81190c52971d4004b7244bdcd93e2f34" {
		t.Errorf("start body appid = %v, want SDK app ID", got)
	}
	_ = session.Say(ctx, "hello", nil, nil)
	_ = session.Interrupt(ctx)
	_, _ = session.Think(ctx, "think", nil, nil, nil, nil, nil)
	_ = session.Update(ctx, &Agora.UpdateAgentsRequestProperties{})
	_, _ = session.GetHistory(ctx)
	_, _ = session.GetInfo(ctx)
	_, _ = session.GetTurns(ctx)
	_ = session.Stop(ctx)

	if len(rec.requests) != 9 {
		t.Fatalf("captured %d requests, want 9 lifecycle requests", len(rec.requests))
	}
	for i, req := range rec.requests {
		if got := req.Header.Get(PreviewFeatureHeader); got != "" {
			t.Errorf("request %d: %s = %q, want no preview gate", i, PreviewFeatureHeader, got)
		}
		if strings.HasPrefix(req.URL.String(), PreviewAPIBaseURL) {
			t.Errorf("request %d unexpectedly used preview host: %s", i, req.URL)
		}
	}
}

func TestGeminiASRSessionPreservesPerCallHeaders(t *testing.T) {
	rec := &recordingClient{}
	session := newGeminiASRSession(rec)
	if _, err := session.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	custom := make(http.Header)
	custom.Set("x-custom", "kept")
	reqOpts := append([]option.RequestOption(nil), session.routingOpts...)
	reqOpts = append(reqOpts, option.WithHTTPHeader(custom))
	_ = session.client.Stop(context.Background(), &Agora.StopAgentsRequest{
		Appid:   "81190c52971d4004b7244bdcd93e2f34",
		AgentID: "agent-1",
	}, reqOpts...)

	req := rec.requests[len(rec.requests)-1]
	if got := req.Header.Get(PreviewFeatureHeader); got != "" {
		t.Errorf("%s = %q, want no preview gate", PreviewFeatureHeader, got)
	}
	if got := req.Header.Get("x-custom"); got != "kept" {
		t.Errorf("x-custom = %q, want %q", got, "kept")
	}
}

func TestStopAgentRemainsProductionOnly(t *testing.T) {
	rec := &recordingClient{}
	client := newTestPreviewClient(rec)

	_ = client.StopAgent(context.Background(), "agent-1")

	if got := rec.requests[0].Header.Get(PreviewFeatureHeader); got != "" {
		t.Errorf("%s = %q, want no preview gate", PreviewFeatureHeader, got)
	}
	if strings.HasPrefix(rec.requests[0].URL.String(), PreviewAPIBaseURL) {
		t.Errorf("StopAgent unexpectedly used preview host: %s", rec.requests[0].URL)
	}
}

func TestGASessionRemainsOnProductionRouting(t *testing.T) {
	rec := &recordingClient{}
	client := newTestPreviewClient(rec)
	agent := NewAgent(client).WithMllm(vendors.NewGeminiLive(vendors.GeminiLiveOptions{
		APIKey: "ga-key", Model: "gemini-live-2.5-flash-preview-native-audio-09-2025",
	}))
	session := agent.CreateSession(CreateSessionOptions{
		Channel: "ga", AgentUID: "1", RemoteUIDs: []string{"100"},
	})
	if _, err := session.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := rec.requests[0].Header.Get(PreviewFeatureHeader); got != "" {
		t.Errorf("%s = %q, want no preview gate", PreviewFeatureHeader, got)
	}
	if strings.HasPrefix(rec.requests[0].URL.String(), PreviewAPIBaseURL) {
		t.Errorf("GA session unexpectedly used preview host: %s", rec.requests[0].URL)
	}
}

// --- Preview feature detection ---------------------------------------------

func TestRequiredPreviewFeaturesLeavesGeminiASROnGA(t *testing.T) {
	asr := map[string]interface{}{"asr": map[string]interface{}{"vendor": "gemini"}}
	if got := RequiredPreviewFeatures(asr); len(got) != 0 {
		t.Errorf("RequiredPreviewFeatures(gemini asr) = %v, want none", got)
	}
}

func TestRequiredPreviewFeaturesLeavesAGAPipelineAlone(t *testing.T) {
	ga := map[string]interface{}{"asr": map[string]interface{}{"vendor": "microsoft"}}

	if got := RequiredPreviewFeatures(ga); len(got) != 0 {
		t.Errorf("RequiredPreviewFeatures(GA asr) = %v, want none", got)
	}
}

func TestDebugHTTPClientLogsFinalRedactedRequest(t *testing.T) {
	// The debug client must be below the preview gate: users need to see the
	// actual endpoint and feature header, not the partial request assembled by
	// AgentSession before routing options run.
	recorder := &recordingClient{}
	client := &previewGateClient{
		features: PreviewFeatureLiveModels,
		inner:    &debugHTTPClient{inner: recorder},
	}
	body := `{"appid":"secret-app-id","properties":{"mllm":{"api_key":"secret-api-key"}}}`
	req, err := http.NewRequest(http.MethodPost, PreviewAPIBaseURL+"/v1/projects/apps/app-1/agents?key=url-secret", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer secret-token")
	req.Header.Set("X-Goog-Api-Key", "google-secret")
	req.Header.Set("X-Request-ID", "trace-123")

	var output bytes.Buffer
	previousOutput := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previousOutput) })

	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()

	got := output.String()
	for _, secret := range []string{"secret-app-id", "secret-api-key", "secret-token", "google-secret", "url-secret", "app-1"} {
		if strings.Contains(got, secret) {
			t.Errorf("debug request log leaked %q: %s", secret, got)
		}
	}
	for _, want := range []string{
		`"method":"POST"`,
		PreviewAPIBaseURL + "/v1/projects/apps/%5BREDACTED%5D/agents",
		`"agora-feature":["live-models"]`,
		`"x-request-id":["trace-123"]`,
		`"api_key":"[REDACTED]"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("debug request log missing %q: %s", want, got)
		}
	}
	if len(recorder.bodies) != 1 || string(recorder.bodies[0]) != body {
		t.Errorf("debug logging changed body delivered to transport: %q", recorder.bodies)
	}
}

// --- Preview wire-shape translation -----------------------------------------

func TestGPTLiveJSONHeadersAreRedacted(t *testing.T) {
	config := map[string]interface{}{"params": map[string]interface{}{"headers": `{"Authorization":"private-value"}`}}
	redacted := RedactSecrets(config).(map[string]interface{})
	if redacted["params"].(map[string]interface{})["headers"] != Redacted {
		t.Fatal("JSON headers leaked")
	}
	if config["params"].(map[string]interface{})["headers"] != `{"Authorization":"private-value"}` {
		t.Fatal("mutated headers")
	}
}

func TestDebugLoggerLeavesNonReplayableBodyUntouched(t *testing.T) {
	body := io.NopCloser(strings.NewReader(`{"test":true}`))
	req, err := http.NewRequest(http.MethodPost, "https://example.test", body)
	if err != nil {
		t.Fatal(err)
	}
	if req.GetBody != nil {
		t.Fatal("test requires a non-replayable body")
	}
	_, err = readRequestBody(req)
	if err == nil {
		t.Fatal("expected unavailable-body error")
	}
	got, err := io.ReadAll(req.Body)
	if err != nil || string(got) != `{"test":true}` {
		t.Fatalf("debug read consumed body: %s, %v", got, err)
	}
}

func TestGPTLiveV3RoutesSessionLifecycle(t *testing.T) {
	rec := &recordingClient{}
	zero := 0
	session := NewAgent(newTestPreviewClient(rec)).WithMllm(vendors.NewOpenAIGPTLive(vendors.OpenAIGPTLiveOptions{
		APIKey: "test", Prompt: "Be brief", OutputIdleEndMs: &zero,
	})).CreateSession(CreateSessionOptions{Channel: "preview", AgentUID: "1", RemoteUIDs: []string{"100"}})
	ctx := context.Background()
	if _, err := session.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if err := session.Say(ctx, "hello", nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := session.Interrupt(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := session.GetHistory(ctx); err != nil {
		t.Fatal(err)
	}
	if err := session.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if len(rec.requests) != 5 {
		t.Fatalf("got %d requests", len(rec.requests))
	}
	for _, req := range rec.requests {
		if req.Header.Get(PreviewFeatureHeader) != PreviewFeatureLiveModels || !strings.HasPrefix(req.URL.String(), PreviewAPIBaseURL) {
			t.Fatal("lost preview route")
		}
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rec.bodies[0], &body); err != nil {
		t.Fatal(err)
	}
	mllm := body["properties"].(map[string]interface{})["mllm"].(map[string]interface{})
	if mllm["enable"] != true || mllm["url"] != "wss://api.openai.com/v1/live/sessions" {
		t.Fatalf("mllm = %#v", mllm)
	}
	assertJSONEqual(t, mllm["params"], `{"model":"gpt-live-1-diamond-alpha","alpha_selector":"quicksilver=v3","prompt":"Be brief","output_idle_end_ms":0}`)
}
