package agentkit_test

// Compiles the code sample in docs/guides/preview-endpoint.md, so the doc
// cannot drift from the API.

import (
	"context"
	"os"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/AgoraIO/agora-agents-go/v2/option"
)

func ExampleGeminiLive() {
	client := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
		Area:           option.AreaUS,
		AppID:          os.Getenv("AGORA_APP_ID"),
		AppCertificate: os.Getenv("AGORA_APP_CERTIFICATE"),
	})

	agent := agentkit.NewAgent(client).
		WithMllm(vendors.NewGeminiLive(vendors.GeminiLiveOptions{
			APIKey:        os.Getenv("GOOGLE_API_KEY"),
			Model:         vendors.GeminiLiveModel38LiveExtendedThinking,
			ThinkingLevel: vendors.GeminiThinkingLevelMedium,
		}))

	session := agent.CreateSession(agentkit.CreateSessionOptions{
		Channel:    "demo",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
	})
	_, _ = session.Start(context.Background())
}
