package agentkit

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/AgoraIO/agora-agents-go/v2/option"
	"github.com/AgoraIO/agora-agents-go/v2/v3"
)

func TestPlatformProviderMappingsFromDocs(t *testing.T) {
	ref := SecretRef("models", "key")
	for _, tc := range []struct {
		name   string
		config map[string]interface{}
		field  string
	}{
		{"deepgram", vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{APIKey: ref, Model: "nova-3", Language: "fr"}).ToConfig(), "api_key"},
		{"speechmatics", map[string]interface{}{"vendor": "speechmatics", "params": map[string]interface{}{"key": ref, "uri": "wss://speech.example", "language": "de"}}, "key"},
		{"assemblyai", map[string]interface{}{"vendor": "assemblyai", "params": map[string]interface{}{"api_key": ref, "ws_url": "wss://assembly.example"}}, "api_key"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, bad := convertModule("asr", tc.config)
			if len(bad) > 0 {
				t.Fatal(bad)
			}
			fields := got["credential"].(map[string]any)["fields"].(map[string]string)
			if fields[tc.field] != ref {
				t.Fatal(fields)
			}
			params, _ := got["params"].(map[string]any)
			if _, ok := params["language"]; ok {
				t.Fatal("duplicate language")
			}
			if tc.name == "deepgram" && got["language"] != "fr" {
				t.Fatal(got)
			}
			if tc.name != "deepgram" && got["url"] == nil {
				t.Fatal("endpoint missing", got)
			}
		})
	}
	got, bad := convertModule("asr", vendors.NewAresSTT(vendors.AresSTTOptions{Keywords: []string{"Agora"}}).ToConfig())
	if len(bad) > 0 || got["language"] != "en-US" {
		t.Fatal(got, bad)
	}
	if got["params"].(map[string]any)["keywords"].([]string)[0] != "Agora" {
		t.Fatal(got)
	}
	for _, endpoint := range []string{"", "https://gemini.example/generate?other=value&key=old"} {
		got, bad := convertModule("llm", vendors.NewGemini(vendors.GeminiOptions{APIKey: ref, Model: "gemini-model", URL: endpoint}).ToConfig())
		if len(bad) > 0 {
			t.Fatal(bad)
		}
		if got["credential"].(map[string]any)["fields"].(map[string]string)["api_key"] != ref {
			t.Fatal(got)
		}
		if strings.Contains(got["url"].(string), "key=") || strings.Contains(got["url"].(string), "secrets") {
			t.Fatal("key leaked into URL", got)
		}
	}
	got, bad = convertModule("avatar", map[string]interface{}{"vendor": "heygen", "params": map[string]interface{}{"api_key": ref}})
	if len(bad) > 0 || got["vendor"] != "liveavatar" {
		t.Fatal(got, bad)
	}
	managed, bad := convertModule("llm", vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-5.6-luna"}).ToConfig())
	if len(bad) > 0 || managed["vendor"] != "openai" || managed["credential"].(map[string]any)["mode"] != "managed" {
		t.Fatal(managed, bad)
	}
	bedrock, bad := convertModule("llm", vendors.NewAmazonBedrock(vendors.AmazonBedrockOptions{APIKey: ref, Region: "us-east-1", Model: "model"}).ToConfig())
	if len(bad) > 0 || bedrock["vendor"] != "amazon" || bedrock["credential"].(map[string]any)["fields"].(map[string]string)["api_key"] != ref {
		t.Fatal(bedrock, bad)
	}
}

func TestMiniMaxOptionsMergeWithoutMutation(t *testing.T) {
	extra := map[string]interface{}{"voice_setting": map[string]interface{}{"voice_id": "old", "speed": 0.8, "custom": true}, "audio_setting": map[string]interface{}{"format": "pcm"}}
	config := vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{Model: "speech-2.8-turbo", VoiceID: "new", Speed: Float64(1.2), SampleRate: Int(24000), AdditionalParams: extra}).ToConfig()
	got, bad := convertModule("tts", config)
	if len(bad) > 0 {
		t.Fatal(bad)
	}
	params := got["params"].(map[string]any)
	voice := params["voice_setting"].(map[string]any)
	if voice["voice_id"] != "new" || voice["speed"] != 1.2 || voice["custom"] != true {
		t.Fatal(voice)
	}
	if extra["voice_setting"].(map[string]any)["voice_id"] != "old" {
		t.Fatal("mutated input")
	}
	audio := params["audio_setting"].(map[string]any)
	if audio["format"] != "pcm" || audio["sample_rate"] != 24000 {
		t.Fatal(audio)
	}
	minimal, _ := convertModule("tts", vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{Model: "speech-2.8-turbo"}).ToConfig())
	if minimal["params"].(map[string]any)["voice_setting"].(map[string]any)["speed"] != 1.0 {
		t.Fatal(minimal)
	}
}

func TestExplicitEphemeralAndTelephoneSessions(t *testing.T) {
	var paths []string
	var bodies []map[string]any
	httpClient := v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		paths = append(paths, r.URL.Path)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		bodies = append(bodies, body)
		return v3Reply(200, `{"agent_session_id":"session","status":"RUNNING"}`, ""), nil
	})
	client := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: httpClient})
	config := NewAgent(client).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-5.6-luna"}))
	if _, err := config.CreateEphemeralSession(CreateSessionOptions{Channel: "room", RemoteUIDs: []string{"2"}}).Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	policy := &CallPolicy{EndCall: &EndCallPolicy{OnUserRequest: true}}
	call := config.Call(CallOptions{From: "number", To: "+12025550123", CallPolicy: policy})
	policy.EndCall.OnUserRequest = false
	if _, err := call.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	saved := &SavedAgent{ID: "saved", client: client}
	if _, err := saved.Call(CallOptions{From: "number", To: "+12025550123"}).Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(paths[0], "/sessions/ephemeral") || !strings.HasSuffix(paths[1], "/sessions/ephemeral") || !strings.HasSuffix(paths[2], "/sessions") {
		t.Fatal(paths)
	}
	transport := bodies[1]["transport"].(map[string]any)
	if transport["type"] != "telephony" || transport["from"] != "number" {
		t.Fatal(transport)
	}
	if _, ok := transport["channel"]; ok {
		t.Fatal(transport)
	}
	if transport["call_policy"].(map[string]any)["end_call"].(map[string]any)["on_user_request"] != true {
		t.Fatal("call policy not snapshotted")
	}
	if _, err := config.Call(CallOptions{From: "number", To: "+12025550123", CreateSessionOptions: CreateSessionOptions{Channel: "bad"}}).Start(context.Background()); err == nil {
		t.Fatal("mixed transport accepted")
	}
}

func TestPlatformTokenGenerationOnlyWithExplicitUID(t *testing.T) {
	client := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: strings.Repeat("a", 32), AppCertificate: strings.Repeat("b", 32), Token: "rest"})
	saved := &SavedAgent{ID: "saved", client: client}
	for _, uid := range []string{"", "0", "1", "agent-name"} {
		session := saved.CreateSession(CreateSessionOptions{Channel: "room", AgentUID: uid, RemoteUIDs: []string{"2"}})
		transport, _, _, err := session.v3StartRequestParts()
		if err != nil {
			t.Fatal(err)
		}
		if (uid == "" || uid == "0") != (transport.Token == "") {
			t.Fatal(uid, transport.Token)
		}
		again, _, _, err := session.v3StartRequestParts()
		if err != nil || again.Token != transport.Token {
			t.Fatal("retry token changed")
		}
	}
}

func TestInboundEndCallPolicy(t *testing.T) {
	calls := 0
	client := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method == "GET" {
			return v3Reply(200, `{"number_id":"number","number":"+12025550123"}`, "etag"), nil
		}
		var b map[string]any
		json.NewDecoder(r.Body).Decode(&b)
		p := b["inbound"].(map[string]any)["call_policy"].(map[string]any)["end_call"].(map[string]any)
		if p["on_user_request"] != true || p["on_fax"] != false {
			t.Fatal(p)
		}
		return v3Reply(200, `{"number_id":"number"}`, "next"), nil
	})})
	_, err := (&SavedAgent{ID: "saved", client: client}).AssignNumber(context.Background(), "number", AssignNumberOptions{CallPolicy: &InboundCallPolicy{EndCall: &InboundEndCallPolicy{OnUserRequest: true}}})
	if err != nil || calls != 2 {
		t.Fatal(err, calls)
	}
}

func TestSessionAnalysisAndSecretKeyMetadata(t *testing.T) {
	var response v3.Session
	if err := json.Unmarshal([]byte(`{"agent_session_id":"s","analysis":{"status":"completed","result":{"resolved":true},"complete_ts":123}}`), &response); err != nil {
		t.Fatal(err)
	}
	if response.Analysis == nil || response.Analysis.Result["resolved"] != true {
		t.Fatal(response)
	}
	var secret v3.SecretSet
	if err := json.Unmarshal([]byte(`{"name":"prod","keys":["old",{"key":"new","create_ts":1,"update_ts":2}]}`), &secret); err != nil {
		t.Fatal(err)
	}
	if secret.Keys[0].Key != "old" || secret.Keys[0].CreateTS != nil || *secret.Keys[1].UpdateTS != 2 {
		t.Fatal(secret)
	}
}
