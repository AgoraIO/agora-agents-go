package vendors_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
)

// Function signatures verify public option and return type identity, not just
// interface compatibility, from an external consumer package.
var avatarAliases = map[string]func(vendors.GenericAvatarOptions) *vendors.GenericAvatar{
	"Tavus":      vendors.NewTavus,
	"Protoface":  vendors.NewProtoface,
	"LemonSlice": vendors.NewLemonSlice,
}

var (
	_ *vendors.Tavus               = (*vendors.GenericAvatar)(nil)
	_ *vendors.Protoface           = (*vendors.GenericAvatar)(nil)
	_ *vendors.LemonSlice          = (*vendors.GenericAvatar)(nil)
	_ vendors.GenericAvatarOptions = vendors.TavusOptions{}
	_ vendors.GenericAvatarOptions = vendors.ProtofaceOptions{}
	_ vendors.GenericAvatarOptions = vendors.LemonSliceOptions{}
	_ vendors.Avatar               = (*vendors.Tavus)(nil)
	_ vendors.Avatar               = (*vendors.Protoface)(nil)
	_ vendors.Avatar               = (*vendors.LemonSlice)(nil)
)

func TestAvatarAliasesSerialization(t *testing.T) {
	disabled := false
	for name, constructor := range avatarAliases {
		t.Run(name, func(t *testing.T) {
			opts := vendors.GenericAvatarOptions{
				APIKey: "key", APIBaseURL: "https://avatar.example", AvatarID: "avatar", AgoraUID: "2001",
				AgoraToken: "token", AgoraAppID: "app", AgoraChannel: "channel", Enable: &disabled,
				AdditionalParams: map[string]interface{}{"custom": "value", "api_key": "overridden"},
			}
			avatar := constructor(opts)
			if avatar.RequiredSampleRate() != 0 {
				t.Fatal("alias must retain generic sample rate behavior")
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

func TestAvatarAliasesValidation(t *testing.T) {
	for name, constructor := range avatarAliases {
		for _, field := range []string{"APIKey", "APIBaseURL", "AvatarID", "AgoraUID"} {
			t.Run(name+"/"+field, func(t *testing.T) {
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
