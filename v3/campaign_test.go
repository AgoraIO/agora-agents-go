package v3

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/option"
)

func TestCampaignEightOperations(t *testing.T) {
	ctx := context.Background()
	count := 0
	c := NewClient("project", option.WithToken("token"), option.WithHTTPClient(httpFunc(func(r *http.Request) (*http.Response, error) {
		count++
		if r.Header.Get("Authorization") != "agora token=token" {
			t.Fatal("missing shared auth")
		}
		var data []byte
		if r.Body != nil {
			var err error
			data, err = io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
		}
		prefix := "/conversational-ai/v1/projects/project/campaigns"
		switch count {
		case 1:
			if r.Method != "POST" || r.URL.Path != prefix || r.Header.Get("Idempotency-Key") != "create-key" {
				t.Fatalf("create %s %s %v", r.Method, r.URL, r.Header)
			}
			var body map[string]any
			if err := json.Unmarshal(data, &body); err != nil {
				t.Fatal(err)
			}
			if body["name"] != "Outreach" || body["agent_id"] != "agent" || body["idempotency_key"] != nil {
				t.Fatalf("body %s", data)
			}
			if len(body["contacts"].([]any)) != 1 || body["transport"].(map[string]any)["from"].([]any)[0] != "number" {
				t.Fatalf("body %s", data)
			}
		case 2:
			if r.Method != "GET" || r.URL.Query().Get("status") != "running" || r.URL.Query().Get("cursor") != "next" || r.URL.Query().Get("limit") != "5" || r.URL.Query().Get("order") != "create_ts.asc" {
				t.Fatal(r.URL)
			}
			return reply(200, `{"data":{"count":1,"list":[{"campaign_id":"c"}]},"has_more":true,"next_cursor":"after"}`, nil), nil
		case 3:
			if r.Method != "GET" || r.URL.EscapedPath() != prefix+"/campaign%2Fa" {
				t.Fatal(r.URL)
			}
		case 4:
			if r.Method != "PATCH" || r.Header.Get("If-Match") != "etag" {
				t.Fatal(r.Header)
			}
			if string(data) != `{"lifecycle":null,"transport":{"call_policy":{}}}` {
				t.Fatalf("null/empty semantics %s", data)
			}
		case 5:
			if r.Method != "DELETE" || len(data) != 0 {
				t.Fatalf("delete %s", data)
			}
			return reply(200, `{"deleted":true}`, nil), nil
		default:
			action := []string{"pause", "resume", "cancel"}[count-6]
			if r.Method != "POST" || r.URL.Path != prefix+"/c:"+action || len(data) != 0 || r.Header.Get("Idempotency-Key") != action+"-key" {
				t.Fatalf("action %s %s body=%s headers=%v", r.Method, r.URL, data, r.Header)
			}
		}
		return reply(200, `{"campaign_id":"c","status":"running"}`, http.Header{"Etag": []string{"etag"}, "Agora-Request-Id": []string{"trace"}, "Idempotent-Replayed": []string{"true"}}), nil
	})))
	created, err := c.Campaigns.Create(ctx, CampaignCreate{Name: "Outreach", AgentID: "agent", Transport: CampaignTransport{Type: "telephony", From: []string{"number"}}, Contacts: []CampaignContact{{Phone: "+12025550123", Values: map[string]any{"name": "Ada"}}}, Schedule: CampaignSchedule{Mode: "immediate"}}, CampaignWriteOptions{IdempotencyKey: "create-key"})
	if err != nil || created.Body.CampaignID != "c" || created.ETag != "etag" || created.RequestID != "trace" || created.Header.Get("Idempotent-Replayed") != "true" {
		t.Fatalf("create response %#v %v", created, err)
	}
	page, err := c.Campaigns.List(ctx, CampaignListOptions{ListOptions: ListOptions{Limit: 5, Cursor: "next", Order: "create_ts.asc"}, Status: "running"})
	if err != nil || len(page.Body.Data.List) != 1 || !page.Body.HasMore || *page.Body.NextCursor != "after" {
		t.Fatalf("page %#v %v", page, err)
	}
	if _, err = c.Campaigns.Get(ctx, "campaign/a"); err != nil {
		t.Fatal(err)
	}
	if _, err = c.Campaigns.Update(ctx, "c", "etag", CampaignPatch{"lifecycle": nil, "transport": map[string]any{"call_policy": map[string]any{}}}); err != nil {
		t.Fatal(err)
	}
	if _, err = c.Campaigns.Delete(ctx, "c"); err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"pause", "resume", "cancel"} {
		var err error
		opts := CampaignWriteOptions{IdempotencyKey: action + "-key"}
		switch action {
		case "pause":
			_, err = c.Campaigns.Pause(ctx, "c", opts)
		case "resume":
			_, err = c.Campaigns.Resume(ctx, "c", opts)
		case "cancel":
			_, err = c.Campaigns.Cancel(ctx, "c", opts)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if count != 8 {
		t.Fatalf("operations %d", count)
	}
}

func TestCampaignConflictLatestAndWriteRetrySafety(t *testing.T) {
	calls := 0
	ctx := context.Background()
	c := NewClient("p", option.WithHTTPClient(httpFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method == "PATCH" {
			return reply(412, `{"reason":"PreconditionFailed"}`, nil), nil
		}
		if r.Method == "GET" {
			if r.Header.Get("If-Match") != "" {
				t.Fatal("write condition leaked into reload")
			}
			return reply(200, `{"campaign_id":"c","name":"Latest"}`, http.Header{"Etag": []string{"new"}}), nil
		}
		return reply(503, `{"reason":"Busy"}`, nil), nil
	})))
	_, err := c.Campaigns.Update(ctx, "c", "old", CampaignPatch{"pacing": map[string]any{"max_concurrent": 2}})
	var conflict *PreconditionFailedError
	if !errors.As(err, &conflict) || calls != 1 {
		t.Fatalf("error %v calls %d", err, calls)
	}
	latest, etag, err := conflict.Latest(ctx)
	if err != nil || etag != "new" || latest.(Campaign).Name != "Latest" {
		t.Fatalf("reload %v %s %v", latest, etag, err)
	}
	for _, action := range []func() error{func() error { _, e := c.Campaigns.Create(ctx, CampaignCreate{}); return e }, func() error { _, e := c.Campaigns.Pause(ctx, "c"); return e }, func() error { _, e := c.Campaigns.Resume(ctx, "c"); return e }, func() error { _, e := c.Campaigns.Cancel(ctx, "c"); return e }, func() error { _, e := c.Campaigns.Delete(ctx, "c"); return e }} {
		before := calls
		if err := action(); err == nil || calls != before+1 {
			t.Fatalf("unexpected retry %v %d", err, calls-before)
		}
	}
}
func TestCampaignLocalPathAndOptionValidation(t *testing.T) {
	c := NewClient("p", option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) { t.Fatal("must not send"); return nil, nil })))
	ctx := context.Background()
	for _, call := range []func() error{func() error { _, e := c.Campaigns.Get(ctx, ""); return e }, func() error { _, e := c.Campaigns.Delete(ctx, " "); return e }, func() error { _, e := c.Campaigns.Update(ctx, "c", "", CampaignPatch{}); return e }, func() error {
		_, e := c.Campaigns.Create(ctx, CampaignCreate{}, CampaignWriteOptions{}, CampaignWriteOptions{})
		return e
	}, func() error {
		_, e := c.Campaigns.Pause(ctx, "c", CampaignWriteOptions{}, CampaignWriteOptions{})
		return e
	}} {
		if err := call(); err == nil {
			t.Fatal("invalid call accepted")
		}
	}
}

func TestCampaignWriteRejectsWhitespaceOnlyIdempotencyKey(t *testing.T) {
	calls := 0
	c := NewClient("p", option.WithHTTPClient(httpFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return reply(200, `{}`, nil), nil
	})))
	ctx := context.Background()
	for _, call := range []func() error{
		func() error {
			_, err := c.Campaigns.Create(ctx, CampaignCreate{}, CampaignWriteOptions{IdempotencyKey: " \t"})
			return err
		},
		func() error {
			_, err := c.Campaigns.Pause(ctx, "c", CampaignWriteOptions{IdempotencyKey: " \t"})
			return err
		},
		func() error {
			_, err := c.Campaigns.Resume(ctx, "c", CampaignWriteOptions{IdempotencyKey: " \t"})
			return err
		},
		func() error {
			_, err := c.Campaigns.Cancel(ctx, "c", CampaignWriteOptions{IdempotencyKey: " \t"})
			return err
		},
	} {
		if err := call(); err == nil {
			t.Fatal("whitespace-only campaign idempotency key should fail locally")
		}
	}
	if calls != 0 {
		t.Fatalf("invalid campaign idempotency keys reached transport: calls=%d", calls)
	}
}

func TestCampaignRequiredEmptyAndZeroFieldsMarshal(t *testing.T) {
	zero := 0
	off := false
	body := CampaignCreate{Name: "name", AgentID: "id", Transport: CampaignTransport{Type: "telephony", From: []string{"number"}, CallPolicy: &CampaignCallPolicy{EndCall: &CampaignEndCallPolicy{}, Voicemail: &CampaignVoicemailPolicy{Mode: "continue"}}}, Contacts: []CampaignContact{}, Schedule: CampaignSchedule{Mode: "scheduled", Timezone: "America/New_York", Start: "2027-01-01T09:00", End: "2027-01-02T17:00"}, Pacing: &CampaignPacing{RetryBackoffMS: &zero}, Lifecycle: &CampaignLifecycle{IdleTimeoutMS: &zero, GracefulStop: &GracefulStop{Enabled: off}}}
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{`"contacts":[]`, `"retry_backoff_ms":0`, `"idle_timeout_ms":0`, `"end_call":{}`} {
		if !strings.Contains(string(encoded), fragment) {
			t.Fatalf("lost field %s in %s", fragment, encoded)
		}
	}
}
