package vendors

import "testing"

func TestAnamAvatarSerializesAvatarID(t *testing.T) {
	config := NewAnamAvatar(AnamAvatarOptions{
		APIKey:   "anam-key",
		AvatarID: "avatar-1",
	}).ToConfig()

	if config["vendor"] != "anam" {
		t.Fatalf("unexpected vendor: %v", config["vendor"])
	}

	params := config["params"].(map[string]interface{})
	if params["api_key"] != "anam-key" {
		t.Fatalf("unexpected params: %#v", params)
	}
	if params["avatar_id"] != "avatar-1" {
		t.Fatalf("unexpected params: %#v", params)
	}
	if _, ok := params["persona_id"]; ok {
		t.Fatalf("unexpected legacy param present: %#v", params)
	}
}

func TestAnamAvatarSerializesPortraitOptions(t *testing.T) {
	videoWidth := 720
	videoHeight := 1280
	config := NewAnamAvatar(AnamAvatarOptions{
		APIKey:      "anam-key",
		AvatarModel: "cara_mk4",
		VideoWidth:  &videoWidth,
		VideoHeight: &videoHeight,
		AdditionalParams: map[string]interface{}{
			"avatar_model": "overridden-model",
			"video_width":  1,
			"video_height": 2,
		},
	}).ToConfig()

	params := config["params"].(map[string]interface{})
	if params["avatar_model"] != "cara_mk4" {
		t.Fatalf("unexpected avatar_model: %#v", params)
	}
	if params["video_width"] != videoWidth || params["video_height"] != videoHeight {
		t.Fatalf("unexpected video dimensions: %#v", params)
	}
}

func TestAnamAvatarRejectsIncompleteVideoDimensions(t *testing.T) {
	videoWidth := 720
	defer func() {
		if recover() != "AnamAvatar requires VideoWidth and VideoHeight together" {
			t.Fatal("expected AnamAvatar to reject incomplete video dimensions")
		}
	}()

	NewAnamAvatar(AnamAvatarOptions{APIKey: "anam-key", VideoWidth: &videoWidth})
}
