package agentkit

import (
	"strings"
	"testing"
	"time"

	"github.com/AgoraIO-Community/go-tokenbuilder/accesstoken2"
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
	if !ok || rtm.UserId != account || len(rtm.Privileges) != 1 || rtm.Privileges[accesstoken2.PrivilegeLogin] != tokenTTL {
		t.Fatal("RTM identity/privilege mismatch")
	}
}

func TestGenerateConvoAITokenSignsStringAccountsLikeNumericUIDs(t *testing.T) {
	appID, certificate := strings.Repeat("a", 32), strings.Repeat("b", 32)
	for _, account := range []string{"42", "agent-a"} {
		token, err := generateConvoAITokenForAccount(GenerateConvoAITokenOptions{AppID: appID, AppCertificate: certificate, ChannelName: "support", TokenExpire: 600}, account)
		if err != nil {
			t.Fatal(err)
		}
		assertConvoAIToken(t, token, appID, certificate, "support", account, 600, 600)
	}
}

func TestGenerateClientTokenGrantsRTCAndRTMWithoutConvoAI(t *testing.T) {
	appID, certificate := strings.Repeat("a", 32), strings.Repeat("b", 32)
	for _, tc := range []struct {
		uid                    string
		privilegeTTL, expected uint32
	}{{"42", 0, 600}, {"user-a", 120, 120}} {
		token, err := GenerateClientToken(GenerateClientTokenOptions{AppID: appID, AppCertificate: certificate, ChannelName: "room", UID: tc.uid, TokenExpire: 600, PrivilegeExpire: int(tc.privilegeTTL)})
		if err != nil {
			t.Fatal(err)
		}
		decoded := accesstoken2.CreateAccessToken()
		if ok, err := decoded.Parse(token); err != nil || !ok {
			t.Fatalf("parse token: ok=%v err=%v", ok, err)
		}
		if ok, err := decoded.VerifySignature(certificate); err != nil || !ok {
			t.Fatalf("verify token: ok=%v err=%v", ok, err)
		}
		if decoded.Expire != 600 || len(decoded.Services) != 2 {
			t.Fatalf("expected RTC and RTM only with 600s TTL, got %d services, TTL %d", len(decoded.Services), decoded.Expire)
		}
		if _, ok := decoded.Services[accesstoken2.ServiceTypeConvoAI]; ok {
			t.Fatal("client token must not carry the ConvoAI service")
		}
		rtc, ok := decoded.Services[accesstoken2.ServiceTypeRtc].(*accesstoken2.ServiceRtc)
		if !ok || rtc.ChannelName != "room" || rtc.Uid != tc.uid || len(rtc.Privileges) != 4 || rtc.Privileges[accesstoken2.PrivilegeJoinChannel] != tc.expected {
			t.Fatalf("RTC identity/privileges mismatch: %+v", rtc)
		}
		rtm, ok := decoded.Services[accesstoken2.ServiceTypeRtm].(*accesstoken2.ServiceRtm)
		if !ok || rtm.UserId != tc.uid || rtm.Privileges[accesstoken2.PrivilegeLogin] != 600 {
			t.Fatalf("RTM identity/privilege mismatch: %+v", rtm)
		}
	}
}

func TestGenerateClientTokenRejectsUIDThatCannotBeRTMUserID(t *testing.T) {
	for _, uid := range []string{"", "0", "000"} {
		_, err := GenerateClientToken(GenerateClientTokenOptions{AppID: strings.Repeat("a", 32), AppCertificate: strings.Repeat("b", 32), ChannelName: "room", UID: uid})
		if err == nil || !strings.Contains(err.Error(), "nonzero and nonempty") {
			t.Fatalf("uid %q: %v", uid, err)
		}
	}
}
