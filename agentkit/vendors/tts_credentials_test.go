package vendors

import "testing"

func TestCartesiaAndFishAudioTTSValidateCredentials(t *testing.T) {
	providers := []struct {
		name   string
		create func(CredentialMode, string)
		key    string
	}{
		{
			name: "CartesiaTTS",
			create: func(mode CredentialMode, key string) {
				NewCartesiaTTS(CartesiaTTSOptions{
					CredentialMode: mode, APIKey: key, VoiceID: "voice", ModelID: "sonic-2",
				})
			},
			key: "APIKey",
		},
		{
			name: "FishAudioTTS",
			create: func(mode CredentialMode, key string) {
				NewFishAudioTTS(FishAudioTTSOptions{
					CredentialMode: mode, Key: key, ReferenceID: "ref", Backend: "speech-1.5",
				})
			},
			key: "Key",
		},
	}
	for _, provider := range providers {
		for _, mode := range []CredentialMode{"", credentialModeBYOK} {
			t.Run(provider.name+"/missing key/"+string(mode), func(t *testing.T) {
				assertPanic(t, provider.name+" requires "+provider.key, func() { provider.create(mode, "") })
			})
		}
		t.Run(provider.name+"/invalid mode", func(t *testing.T) {
			assertPanic(t, provider.name+" CredentialMode must be one of: managed, byok", func() {
				provider.create("invalid", "tts-key")
			})
		})
	}

	tests := []struct {
		name   string
		create func()
		want   string
	}{
		{
			name: "managed Cartesia requires voice",
			create: func() {
				NewCartesiaTTS(CartesiaTTSOptions{CredentialMode: credentialModeManaged, ModelID: "sonic-2"})
			},
			want: "CartesiaTTS requires VoiceID",
		},
		{
			name: "managed Cartesia requires model",
			create: func() {
				NewCartesiaTTS(CartesiaTTSOptions{CredentialMode: credentialModeManaged, VoiceID: "voice"})
			},
			want: "CartesiaTTS requires ModelID",
		},
		{
			name: "managed Fish Audio requires reference",
			create: func() {
				NewFishAudioTTS(FishAudioTTSOptions{CredentialMode: credentialModeManaged, Backend: "speech-1.5"})
			},
			want: "FishAudioTTS requires ReferenceID",
		},
		{
			name: "managed Fish Audio requires backend",
			create: func() {
				NewFishAudioTTS(FishAudioTTSOptions{CredentialMode: credentialModeManaged, ReferenceID: "ref"})
			},
			want: "FishAudioTTS requires Backend",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertPanic(t, tt.want, tt.create) })
	}
}
