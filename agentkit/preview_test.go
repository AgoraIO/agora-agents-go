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
			"language_codes": ["en-US"]
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

func TestLanguageCodesIsSentVerbatimWhenSupplied(t *testing.T) {
	single := vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
		APIKey:        previewAPIKey,
		LanguageCodes: []string{"es-ES"},
	}).ToConfig()
	assertJSONEqual(t, single["params"].(map[string]interface{})["language_codes"], `["es-ES"]`)

	multiple := vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
		APIKey:        previewAPIKey,
		LanguageCodes: []string{"en-US", "es-ES"},
	}).ToConfig()
	assertJSONEqual(t, multiple["params"].(map[string]interface{})["language_codes"], `["en-US","es-ES"]`)
}

func TestExplicitEmptyLanguageCodesStillReachesTheWire(t *testing.T) {
	// `[]` is the caller spelling auto-detect outright. Nil versus empty is the
	// distinction: nil omits the field, empty-non-nil sends [].
	config := vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
		APIKey:        previewAPIKey,
		LanguageCodes: []string{},
	}).ToConfig()

	assertJSONEqual(t, config["params"].(map[string]interface{})["language_codes"], `[]`)
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
		"explicit empty vocabulary": {
			APIKey:           previewAPIKey,
			CustomVocabulary: []string{},
			WordTimestamp:    &explicitTrue,
		},
		"additional params": {
			APIKey: previewAPIKey,
			AdditionalParams: map[string]interface{}{
				"custom_vocabulary": []string{"Agora"},
				"word_timestamp":    true,
			},
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

func newPreviewSession(rec *recordingClient) *AgentSession {
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

func TestNewAgoraClientRoutesPreviewSessionLifecycle(t *testing.T) {
	rec := &recordingClient{}
	session := newPreviewSession(rec)
	ctx := context.Background()
	if _, err := session.Start(ctx); err != nil {
		t.Fatal(err)
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
		if got := req.Header.Get(PreviewFeatureHeader); got != PreviewFeatureGeminiLive {
			t.Errorf("request %d: %s = %q", i, PreviewFeatureHeader, got)
		}
		if !strings.HasPrefix(req.URL.String(), PreviewAPIBaseURL) {
			t.Errorf("request %d went to %q, want preview host", i, req.URL)
		}
	}
}

func TestPerCallHeaderOptionCannotDropTheGate(t *testing.T) {
	// option.WithHTTPHeader replaces the whole header map, so the gate is pinned
	// below the option layer. A request that loses it is routed to production,
	// where the preview providers do not exist.
	rec := &recordingClient{}
	session := newPreviewSession(rec)
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
	if got := req.Header.Get(PreviewFeatureHeader); got != "gemini-live" {
		t.Errorf("%s = %q, want it to survive a per-call header option", PreviewFeatureHeader, got)
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

func TestRequiredPreviewFeaturesFlagsTheGeminiASRVendor(t *testing.T) {
	asr := map[string]interface{}{"asr": map[string]interface{}{"vendor": "gemini"}}
	if got := RequiredPreviewFeatures(asr); len(got) != 1 || got[0] != PreviewFeatureGeminiLive {
		t.Errorf("RequiredPreviewFeatures(gemini asr) = %v, want [gemini-live]", got)
	}
}

func TestRequiredPreviewFeaturesLeavesAGAPipelineAlone(t *testing.T) {
	ga := map[string]interface{}{"asr": map[string]interface{}{"vendor": "microsoft"}}

	if got := RequiredPreviewFeatures(ga); len(got) != 0 {
		t.Errorf("RequiredPreviewFeatures(GA asr) = %v, want none", got)
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

func TestRedactSecretsCoversGeminiURLAndGoogleTTSCredentials(t *testing.T) {
	redacted := RedactSecrets(map[string]interface{}{
		"llm": map[string]interface{}{
			"url": "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:streamGenerateContent?alt=sse&key=" + previewAPIKey,
		},
		"tts": map[string]interface{}{
			"params": map[string]interface{}{"credentials": previewAPIKey},
		},
	})

	got, _ := json.Marshal(redacted)
	if strings.Contains(string(got), previewAPIKey) {
		t.Errorf("live Google key leaked into debug dump: %s", got)
	}
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

// --- Preview wire-shape translation -----------------------------------------
