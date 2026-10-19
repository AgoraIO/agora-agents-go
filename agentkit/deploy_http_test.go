package agentkit

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/option"
	"github.com/AgoraIO/agora-agents-go/v2/v3"
)

type v3HTTPFunc func(*http.Request) (*http.Response, error)

func (f v3HTTPFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }
func v3Reply(code int, body, etag string) *http.Response {
	h := http.Header{}
	if etag != "" {
		h.Set("ETag", etag)
	}
	return &http.Response{StatusCode: code, Header: h, Body: io.NopCloser(strings.NewReader(body))}
}

func TestDeployUpdateUsesFreshETagAndClearsStale(t *testing.T) {
	calls := 0
	httpClient := v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Path != "/conversational-ai/v3/projects/app/agents/id" {
			t.Fatalf("path: %s", r.URL.Path)
		}
		switch calls {
		case 1:
			if r.Method != "GET" {
				t.Fatalf("first request: %s", r.Method)
			}
			return v3Reply(200, `{"agent_id":"id","agent_name":"Old","create_ts":1,"update_ts":2,"pipeline":{"mode":"cascaded","llm":{"vendor":"openai"}}}`, "e1"), nil
		case 2:
			if r.Method != "PATCH" || r.Header.Get("If-Match") != "e1" {
				t.Fatalf("patch headers: %s %v", r.Method, r.Header)
			}
			b, _ := io.ReadAll(r.Body)
			var patch map[string]any
			if err := json.Unmarshal(b, &patch); err != nil {
				t.Fatal(err)
			}
			pipe := patch["pipeline"].(map[string]any)
			if value, ok := pipe["llm"]; !ok || value != nil {
				t.Fatalf("stale LLM not cleared: %s", b)
			}
			return v3Reply(200, `{"agent_id":"id","agent_name":"New","create_ts":1,"update_ts":3}`, "e2"), nil
		default:
			t.Fatalf("unexpected request %d", calls)
			return nil, nil
		}
	})
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: httpClient, V3: &V3Options{BaseURL: "https://stage.example", Headers: http.Header{"X-Studio-V2-Gateway-Secret": []string{"gateway"}}}})
	a := NewAgent(c)
	d, err := a.Deploy(context.Background(), DeployOptions{Name: "New", AgentID: "id"})
	if err != nil {
		t.Fatal(err)
	}
	if d.ID != "id" || d.ETag != "e2" || calls != 2 {
		t.Fatalf("deployed: %#v calls=%d", d, calls)
	}
}

func TestAssignUnassignNumberRetriesOneConflict(t *testing.T) {
	calls := 0
	httpClient := v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch calls {
		case 1:
			if r.Method != "GET" || !strings.HasSuffix(r.URL.Path, "/numbers") {
				t.Fatalf("list: %s %s", r.Method, r.URL.Path)
			}
			return v3Reply(200, `{"data":{"count":1,"list":[{"number_id":"n1","number":"+12025550123","create_ts":1}]},"has_more":false,"next_cursor":null}`, ""), nil
		case 2, 4, 6:
			if r.Method != "GET" || !strings.HasSuffix(r.URL.Path, "/numbers/n1") {
				t.Fatalf("get: %s %s", r.Method, r.URL.Path)
			}
			etag := "e1"
			if calls == 4 {
				etag = "e2"
			}
			return v3Reply(200, `{"number_id":"n1","number":"+12025550123","create_ts":1}`, etag), nil
		case 3:
			if r.Method != "PATCH" {
				t.Fatalf("patch: %s", r.Method)
			}
			return v3Reply(412, `{"detail":"stale"}`, ""), nil
		case 5, 7:
			if calls == 5 && r.Header.Get("If-Match") != "e2" {
				t.Fatalf("retry used stale ETag: %q", r.Header.Get("If-Match"))
			}
			b, _ := io.ReadAll(r.Body)
			var body map[string]any
			json.Unmarshal(b, &body)
			in := body["inbound"].(map[string]any)
			if calls == 5 && in["agent"] != "a1" {
				t.Fatalf("assign: %s", b)
			}
			if calls == 7 {
				if v, ok := in["agent"]; !ok || v != nil {
					t.Fatalf("unassign: %s", b)
				}
			}
			if _, ok := in["sip_trunk"]; ok {
				t.Fatalf("echoed number details: %s", b)
			}
			return v3Reply(200, `{"number_id":"n1","number":"+12025550123","create_ts":1}`, "e2"), nil
		default:
			t.Fatalf("unexpected request %d", calls)
			return nil, nil
		}
	})
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: httpClient, V3: &V3Options{BaseURL: "https://stage.example"}})
	d := &DeployedAgent{ID: "a1", Raw: c.V3.Agents, client: c}
	if _, err := d.AssignNumber(context.Background(), "+1 (202) 555-0123", AssignNumberOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := d.UnassignNumber(context.Background(), v3.Number{NumberID: "n1"}); err != nil {
		t.Fatal(err)
	}
	if calls != 7 {
		t.Fatalf("requests: %d", calls)
	}
}

func TestListNumbersScansPagesWithoutDetailGets(t *testing.T) {
	calls := 0
	httpClient := v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != "GET" || !strings.HasSuffix(r.URL.Path, "/numbers") {
			t.Fatalf("unexpected detail request: %s %s", r.Method, r.URL.Path)
		}
		if calls == 1 {
			if r.URL.Query().Get("cursor") != "" {
				t.Fatal("first cursor")
			}
			return v3Reply(200, `{"data":{"count":3,"list":[{"number_id":"n1","number":"+111","create_ts":1,"inbound":{"agent":"a1"}},{"number_id":"n2","number":"+222","create_ts":1,"inbound":{"agent":"other"}},{"number_id":"n3","number":"+333","create_ts":1,"inbound":null}]},"has_more":true,"next_cursor":"page2"}`, ""), nil
		}
		if calls == 2 {
			if r.URL.Query().Get("cursor") != "page2" {
				t.Fatal("second cursor")
			}
			return v3Reply(200, `{"data":{"count":1,"list":[{"number_id":"n4","number":"+444","create_ts":1,"inbound":{"agent":"a1"}}]},"has_more":false,"next_cursor":null}`, ""), nil
		}
		t.Fatal("unexpected third page")
		return nil, nil
	})
	c := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: "app", Token: "token", HTTPClient: httpClient, V3: &V3Options{BaseURL: "https://stage.example"}})
	d := &DeployedAgent{ID: "a1", Raw: c.V3.Agents, client: c}
	numbers, err := d.ListNumbers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || len(numbers) != 2 || numbers[0].NumberID != "n1" || numbers[1].NumberID != "n4" {
		t.Fatalf("scan: calls=%d numbers=%#v", calls, numbers)
	}
}
