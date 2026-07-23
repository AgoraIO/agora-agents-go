package vendors

import "testing"

func TestOpenAIRealtimeURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "default URL",
			want: "wss://api.openai.com/v1/realtime",
		},
		{
			name: "custom URL",
			url:  "wss://realtime.example.com/v1",
			want: "wss://realtime.example.com/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewOpenAIRealtime(OpenAIRealtimeOptions{
				APIKey: "openai-key",
				URL:    tt.url,
			}).ToConfig()

			if got := config["url"]; got != tt.want {
				t.Errorf("url = %v, want %q", got, tt.want)
			}
		})
	}
}
