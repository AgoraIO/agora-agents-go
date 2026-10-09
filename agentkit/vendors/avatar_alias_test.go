package vendors_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit"
	agentcore "github.com/AgoraIO/agora-agents-go/v2/agentkit/core"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
)

// External consumer checks preserve constructor names and Avatar compatibility.
var avatarProviders = map[string]func(vendors.GenericAvatarOptions) vendors.Avatar{
	"Tavus": func(o vendors.GenericAvatarOptions) vendors.Avatar { return vendors.NewTavus(vendors.TavusOptions(o)) },
	"Protoface": func(o vendors.GenericAvatarOptions) vendors.Avatar {
		return vendors.NewProtoface(vendors.ProtofaceOptions(o))
	},
	"LemonSlice": func(o vendors.GenericAvatarOptions) vendors.Avatar {
		return vendors.NewLemonSlice(vendors.LemonSliceOptions{
			APIKey: o.APIKey, APIBaseURL: o.APIBaseURL, AvatarID: o.AvatarID,
			AgoraUID: o.AgoraUID, AgoraToken: o.AgoraToken, AgoraAppID: o.AgoraAppID,
			AgoraChannel: o.AgoraChannel, Enable: o.Enable, AdditionalParams: o.AdditionalParams,
			AgentID: str("agent"),
		})
	},
}

var (
	_ vendors.Avatar                                      = (*vendors.Tavus)(nil)
	_ vendors.Avatar                                      = (*vendors.Protoface)(nil)
	_ vendors.Avatar                                      = (*vendors.LemonSlice)(nil)
	_ func(vendors.TavusOptions) *vendors.Tavus           = vendors.NewTavus
	_ func(vendors.ProtofaceOptions) *vendors.Protoface   = vendors.NewProtoface
	_ func(vendors.LemonSliceOptions) *vendors.LemonSlice = vendors.NewLemonSlice
)

func str(s string) *string { return &s }

func TestAvatarProvidersSerialization(t *testing.T) {
	disabled := false
	for name, constructor := range avatarProviders {
		t.Run(name, func(t *testing.T) {
			opts := vendors.GenericAvatarOptions{
				APIKey: "key", APIBaseURL: "https://avatar.example", AvatarID: "avatar", AgoraUID: "2001",
				AgoraToken: "token", AgoraAppID: "app", AgoraChannel: "channel", Enable: &disabled,
				AdditionalParams: map[string]interface{}{"custom": "value", "api_key": "overridden"},
			}
			avatar := constructor(opts)
			if avatar.RequiredSampleRate() != 0 {
				t.Fatal("provider must retain generic sample rate behavior")
			}
			// Verify the public builder accepts the alias.
			agentkit.NewAgent(&agentkit.AgoraClient{}).WithAvatar(avatar)
			data, err := json.Marshal(avatar.ToConfig())
			if err != nil {
				t.Fatal(err)
			}
			var got map[string]interface{}
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			want := map[string]interface{}{
				"enable": false, "vendor": "generic",
				"params": map[string]interface{}{
					"api_key": "key", "api_base_url": "https://avatar.example", "avatar_id": "avatar", "agora_uid": "2001",
					"agora_token": "token", "agora_appid": "app", "agora_channel": "channel", "custom": "value",
				},
			}
			if name == "LemonSlice" {
				want["params"].(map[string]interface{})["agent_id"] = "agent"
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("serialized config = %#v, want %#v", got, want)
			}
			opts.Enable = nil
			opts.AgoraToken, opts.AgoraAppID, opts.AgoraChannel = "", "", ""
			config := constructor(opts).ToConfig()
			if config["enable"] != true {
				t.Fatal("avatar must be enabled by default")
			}
			params := config["params"].(map[string]interface{})
			for _, key := range []string{"agora_token", "agora_appid", "agora_channel"} {
				if _, exists := params[key]; exists {
					t.Fatalf("empty optional field %s serialized", key)
				}
			}
		})
	}
}

func TestAvatarProvidersValidation(t *testing.T) {
	for name, constructor := range avatarProviders {
		for _, field := range []string{"APIKey", "AvatarID", "AgoraUID"} {
			t.Run(name+"/"+field, func(t *testing.T) {
				if name == "LemonSlice" && field == "AvatarID" {
					return
				}
				opts := vendors.GenericAvatarOptions{APIKey: "key", APIBaseURL: "https://avatar.example", AvatarID: "avatar", AgoraUID: "2001"}
				reflect.ValueOf(&opts).Elem().FieldByName(field).SetString("")
				defer func() {
					if got := recover(); got != "GenericAvatar requires "+field {
						t.Fatalf("panic = %v, want GenericAvatar requires %s", got, field)
					}
				}()
				constructor(opts)
			})
		}
	}
}

func TestProviderDefaults(t *testing.T) {
	for name, constructor := range avatarProviders {
		opts := vendors.GenericAvatarOptions{APIKey: "key", AvatarID: "avatar", AgoraUID: "2001"}
		urls := map[string]string{"Tavus": "https://tavusapi.com/v2/conversations/agora", "Protoface": "https://api.protoface.com/v1/agora", "LemonSlice": "https://lemonslice.com/api/liveai/agora"}
		if got := constructor(opts).ToConfig()["params"].(map[string]interface{})["api_base_url"]; got != urls[name] {
			t.Fatalf("%s URL = %v", name, got)
		}
	}
	params := vendors.NewLemonSlice(vendors.LemonSliceOptions{APIKey: "key", AgoraUID: "2001", AgentID: str("id")}).ToConfig()["params"].(map[string]interface{})
	if params["avatar_id"] != "lemonslice" {
		t.Fatal(params)
	}
	if _, ok := params["aspect_ratio"]; ok {
		t.Fatal("unset aspect_ratio serialized")
	}
	mustPanic(t, func() {
		vendors.NewGenericAvatar(vendors.GenericAvatarOptions{APIKey: "key", AvatarID: "id", AgoraUID: "2001"})
	})
}

func mustPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("expected panic")
		}
	}()
	f()
}

func TestLemonSliceEffectiveParams(t *testing.T) {
	for _, key := range []string{"agent_id", "agent_image_url", "agent_image_base64"} {
		for _, typed := range []bool{false, true} {
			t.Run(key+fmt.Sprint(typed), func(t *testing.T) {
				input := map[string]interface{}{key: "additional", "custom": "value"}
				opts := vendors.LemonSliceOptions{APIKey: "key", AgoraUID: "2001", AdditionalParams: input}
				want := "additional"
				if typed {
					want = "typed"
					switch key {
					case "agent_id":
						opts.AgentID = str(want)
					case "agent_image_url":
						opts.AgentImageURL = str(want)
					case "agent_image_base64":
						opts.AgentImageBase64 = str(want)
					}
				}
				avatar := vendors.NewLemonSlice(opts)
				params := avatar.ToConfig()["params"].(map[string]interface{})
				if params[key] != want || params["custom"] != "value" {
					t.Fatal(params)
				}
				if input[key] != "additional" || len(input) != 2 || opts.APIBaseURL != "" || opts.AvatarID != "" {
					t.Fatal("input mutated")
				}
				input[key] = "changed"
				params[key] = "changed config"
				if avatar.ToConfig()["params"].(map[string]interface{})[key] != want {
					t.Fatal("provider retained caller map")
				}
			})
		}
	}
	for _, ratio := range []string{"2x3", "9x16", "1x1"} {
		opts := vendors.LemonSliceOptions{APIKey: "key", AgoraUID: "2001", AgentID: str("id"), AspectRatio: str(ratio), AdditionalParams: map[string]interface{}{"aspect_ratio": "bad", "agent_id": 123}}
		if vendors.NewLemonSlice(opts).ToConfig()["params"].(map[string]interface{})["aspect_ratio"] != ratio {
			t.Fatal(ratio)
		}
		opts.AspectRatio = nil
		opts.AdditionalParams = map[string]interface{}{"aspect_ratio": ratio}
		if vendors.NewLemonSlice(opts).ToConfig()["params"].(map[string]interface{})["aspect_ratio"] != ratio {
			t.Fatal(ratio)
		}
	}
	for _, params := range []map[string]interface{}{
		{}, {"agent_id": "id", "agent_image_url": "url"},
		{"agent_id": "id", "agent_image_url": "url", "agent_image_base64": "base64"},
		{"agent_id": ""}, {"agent_id": "  "}, {"agent_id": 123},
		{"agent_id": "id", "agent_image_url": ""}, {"agent_id": "id", "agent_image_base64": nil},
		{"agent_id": "id", "aspect_ratio": ""}, {"agent_id": "id", "aspect_ratio": "16x9"},
		{"agent_id": "id", "aspect_ratio": 123},
	} {
		mustPanic(t, func() {
			vendors.NewLemonSlice(vendors.LemonSliceOptions{APIKey: "key", AgoraUID: "2001", AdditionalParams: params})
		})
	}
	mustPanic(t, func() {
		vendors.NewLemonSlice(vendors.LemonSliceOptions{APIKey: "key", AgoraUID: "2001", AgentID: str(""), AdditionalParams: map[string]interface{}{"agent_id": "valid"}})
	})
	mustPanic(t, func() {
		vendors.NewLemonSlice(vendors.LemonSliceOptions{APIKey: "key", AgoraUID: "2001", AgentID: str("id"), AgentImageURL: str("url")})
	})
}

func TestProviderSessionAndTokens(t *testing.T) {
	for name, constructor := range avatarProviders {
		for _, explicit := range []bool{false, true} {
			t.Run(name+fmt.Sprint(explicit), func(t *testing.T) {
				opts := vendors.GenericAvatarOptions{APIKey: "key", AvatarID: "avatar", AgoraUID: "2001"}
				if explicit {
					opts.AgoraToken = "explicit-token"
					opts.AgoraAppID = "explicit-app"
					opts.AgoraChannel = "explicit-channel"
				}
				avatar := constructor(opts)
				base := &agentcore.BaseAgent{Avatar: avatar.ToConfig()}
				calls := 0
				props, err := agentcore.BuildPropertiesMap(agentcore.ProfileGlobal, base, agentcore.ToPropertiesOptions{
					Channel: "session-channel", AgentUID: "1001", RemoteUIDs: []string{"1002"},
					AppID: "session-app", AppCertificate: "cert", ExpiresIn: 3600, SkipVendorValidation: true,
				}, func(o agentcore.GenerateConvoAITokenOptions) (string, error) { calls++; return "generated-token", nil })
				if err != nil {
					t.Fatal(err)
				}
				got := props["avatar"].(map[string]interface{})
				params := got["params"].(map[string]interface{})
				token, app, channel := "generated-token", "session-app", "session-channel"
				if explicit {
					token, app, channel = "explicit-token", "explicit-app", "explicit-channel"
				}
				if got["vendor"] != "generic" || params["agora_token"] != token || params["agora_appid"] != app || params["agora_channel"] != channel {
					t.Fatal(got)
				}
				if !agentcore.IsAvatarTokenManaged("generic") {
					t.Fatal("generic tokens unmanaged")
				}
				if !explicit && calls == 0 {
					t.Fatal("no token generated")
				}
				if !explicit {
					original := avatar.ToConfig()["params"].(map[string]interface{})
					if _, exists := original["agora_token"]; exists {
						t.Fatal("session mutated provider")
					}
				}
			})
		}
	}
}

func TestLemonSliceMalformedAlternativeSelectors(t *testing.T) {
	for _, key := range []string{"agent_id", "agent_image_url", "agent_image_base64"} {
		for _, value := range []interface{}{"", " 	 ", nil, 42, false} {
			params := map[string]interface{}{key: value}
			other := "agent_id"
			if key == other {
				other = "agent_image_url"
			}
			params[other] = "valid"
			mustPanic(t, func() {
				vendors.NewLemonSlice(vendors.LemonSliceOptions{APIKey: "key", AgoraUID: "2001", AdditionalParams: params})
			})
		}
	}
}
