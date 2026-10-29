package agentkit

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AgoraIO-Community/go-tokenbuilder/accesstoken2"
	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/AgoraIO/agora-agents-go/v2/option"
	"github.com/AgoraIO/agora-agents-go/v2/v3"
)

func assertConvoAIToken(t *testing.T, value, appID, certificate, channel, account string, tokenTTL, privilegeTTL uint32) {
	t.Helper()
	decoded := accesstoken2.CreateAccessToken()
	if ok, err := decoded.Parse(value); err != nil || !ok {
		t.Fatalf("parse token: ok=%v err=%v", ok, err)
	}
	if ok, err := decoded.VerifySignature(certificate); err != nil || !ok {
		t.Fatalf("verify token: ok=%v err=%v", ok, err)
	}
	if decoded.AppId != appID || decoded.Expire != tokenTTL {
		t.Fatalf("token app/TTL: %s/%d", decoded.AppId, decoded.Expire)
	}
	now := time.Now().Unix()
	if int64(decoded.IssueTs) < now-10 || int64(decoded.IssueTs) > now+1 {
		t.Fatalf("issue time: %d", decoded.IssueTs)
	}
	if len(decoded.Services) != 3 {
		t.Fatalf("expected RTC, RTM and ConvoAI services, got %d", len(decoded.Services))
	}
	product, ok := decoded.Services[accesstoken2.ServiceTypeConvoAI].(*accesstoken2.ServiceConvoAI)
	if !ok || len(product.Privileges) != 0 {
		t.Fatal("missing ConvoAI service or unexpected product privileges")
	}
	rtc, ok := decoded.Services[accesstoken2.ServiceTypeRtc].(*accesstoken2.ServiceRtc)
	if !ok || rtc.ChannelName != channel || rtc.Uid != account {
		t.Fatal("RTC identity mismatch")
	}
	if len(rtc.Privileges) != 4 {
		t.Fatalf("RTC privileges: %v", rtc.Privileges)
	}
	for _, privilege := range []uint16{accesstoken2.PrivilegeJoinChannel, accesstoken2.PrivilegePublishAudioStream, accesstoken2.PrivilegePublishVideoStream, accesstoken2.PrivilegePublishDataStream} {
		if rtc.Privileges[privilege] != privilegeTTL {
			t.Fatalf("RTC privilege %d TTL: %d", privilege, rtc.Privileges[privilege])
		}
	}
	rtm, ok := decoded.Services[accesstoken2.ServiceTypeRtm].(*accesstoken2.ServiceRtm)
	if !ok || rtm.UserId != account || len(rtm.Privileges) != 1 || rtm.Privileges[accesstoken2.PrivilegeLogin] != privilegeTTL {
		t.Fatal("RTM identity/privilege mismatch")
	}
}

func TestGenerateConvoAITokenServicesAndExpiration(t *testing.T) {
	for _, tc := range []struct {
		name                                                          string
		uid, tokenTTL, privilegeTTL, expectedToken, expectedPrivilege int
	}{
		{"defaults and UID zero", 0, 0, 0, 86400, 86400},
		{"custom lifetime", 42, 3600, 0, 3600, 3600},
		{"independent privilege lifetime", 17, 7200, 900, 7200, 900},
	} {
		t.Run(tc.name, func(t *testing.T) {
			appID, certificate := strings.Repeat("a", 32), strings.Repeat("b", 32)
			token, err := GenerateConvoAIToken(GenerateConvoAITokenOptions{AppID: appID, AppCertificate: certificate, ChannelName: "support", UID: tc.uid, TokenExpire: tc.tokenTTL, PrivilegeExpire: tc.privilegeTTL})
			if err != nil {
				t.Fatal(err)
			}
			assertConvoAIToken(t, token, appID, certificate, "support", strconv.Itoa(tc.uid), uint32(tc.expectedToken), uint32(tc.expectedPrivilege))
		})
	}
}

func TestPlatformSessionAuthorizationContainsConvoAIService(t *testing.T) {
	appID, certificate := strings.Repeat("a", 32), strings.Repeat("b", 32)
	requests := 0
	client := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: appID, AppCertificate: certificate, HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		requests++
		if !strings.Contains(r.URL.Path, "/sessions") {
			t.Fatalf("unexpected route: %s", r.URL.Path)
		}
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "agora token=") {
			t.Fatal("missing token authorization")
		}
		assertConvoAIToken(t, strings.TrimPrefix(auth, "agora token="), appID, certificate, "v3", "0", 86400, 86400)
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/sessions") {
			return v3Reply(200, `{"data":{"list":[],"count":0},"has_more":false}`, ""), nil
		}
		return v3Reply(200, `{"agent_session_id":"session","status":"RUNNING"}`, ""), nil
	})})
	ctx := context.Background()
	if _, err := client.V3.Sessions.List(ctx, v3.SessionListOptions{}); err != nil {
		t.Fatal(err)
	}
	saved := (&SavedAgent{ID: "saved", client: client}).CreateSession(CreateSessionOptions{Channel: "room", AgentUID: "1", RemoteUIDs: []string{"2"}})
	if _, err := saved.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := saved.GetInfo(ctx); err != nil {
		t.Fatal(err)
	}
	if err := saved.Update(ctx, &Agora.UpdateAgentsRequestProperties{Token: String("replacement-rtc-token")}); err != nil {
		t.Fatal(err)
	}
	if err := saved.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	config := NewAgent(client).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-5.6-luna"}))
	if _, err := config.CreateEphemeralSession(CreateSessionOptions{Channel: "room", AgentUID: "1", RemoteUIDs: []string{"2"}}).Start(ctx); err != nil {
		t.Fatal(err)
	}
	if requests != 6 {
		t.Fatalf("requests: %d", requests)
	}
}

func TestPlatformPrebuiltAuthorizationIsPreserved(t *testing.T) {
	requests := 0
	client := NewAgoraClient(AgoraClientOptions{Area: option.AreaUS, AppID: strings.Repeat("a", 32), Token: "caller-token", HTTPClient: v3HTTPFunc(func(r *http.Request) (*http.Response, error) {
		requests++
		if r.Header.Get("Authorization") != "agora token=caller-token" {
			t.Fatal("prebuilt authorization was replaced")
		}
		return v3Reply(200, `{"data":{"list":[],"count":0},"has_more":false}`, ""), nil
	})})
	if _, err := client.V3.Sessions.List(context.Background(), v3.SessionListOptions{}); err != nil {
		t.Fatal(err)
	}
	if requests != 1 {
		t.Fatalf("requests: %d", requests)
	}
}
