package vendors

import (
	"strings"
	"testing"
)

func TestGroqSerializesAsOpenAICompatible(t *testing.T) {
	config := NewGroq(GroqOptions{
		APIKey:  "groq-key",
		Model:   "llama-3.3-70b-versatile",
		BaseURL: "https://api.groq.com/openai/v1/chat/completions",
	}).ToConfig()

	if config["url"] != "https://api.groq.com/openai/v1/chat/completions" {
		t.Fatalf("unexpected url: %v", config["url"])
	}
	if config["style"] != "openai" {
		t.Fatalf("unexpected style: %v", config["style"])
	}
}

func TestCustomLLMMarksRequestAsCustom(t *testing.T) {
	config := NewCustomLLM(CustomLLMOptions{
		APIKey:  "key",
		Model:   "model",
		BaseURL: "https://llm.example.com/chat",
	}).ToConfig()

	if config["vendor"] != "custom" {
		t.Fatalf("unexpected vendor: %v", config["vendor"])
	}
	if config["style"] != "openai" {
		t.Fatalf("unexpected style: %v", config["style"])
	}
}

func TestAnthropicSerializesRequiredClaudeFields(t *testing.T) {
	maxTokens := 1024
	config := NewAnthropic(AnthropicOptions{
		APIKey:    "anthropic-key",
		Model:     "claude-3-5-sonnet-20241022",
		URL:       "https://api.anthropic.com/v1/messages",
		Headers:   map[string]string{"anthropic-version": "2023-06-01"},
		MaxTokens: &maxTokens,
	}).ToConfig()

	params := config["params"].(map[string]interface{})
	headers := config["headers"].(map[string]string)
	if config["url"] != "https://api.anthropic.com/v1/messages" || config["style"] != "anthropic" {
		t.Fatalf("unexpected config: %#v", config)
	}
	if headers["anthropic-version"] != "2023-06-01" {
		t.Fatalf("unexpected headers: %#v", headers)
	}
	if params["model"] != "claude-3-5-sonnet-20241022" || params["max_tokens"] != 1024 {
		t.Fatalf("unexpected params: %#v", params)
	}
}

func TestAzureOpenAIIncludesRequiredModelParam(t *testing.T) {
	config := NewAzureOpenAI(AzureOpenAIOptions{
		APIKey:         "azure-key",
		Endpoint:       "https://example.openai.azure.com",
		DeploymentName: "deployment",
		Model:          "gpt-4o",
	}).ToConfig()

	params := config["params"].(map[string]interface{})
	if config["vendor"] != "azure" || config["style"] != "openai" {
		t.Fatalf("unexpected config: %#v", config)
	}
	if params["model"] != "gpt-4o" {
		t.Fatalf("unexpected params: %#v", params)
	}
}

func TestVertexAILLMBuildsCorrectURLAndExcludesProjectFromParams(t *testing.T) {
	config := NewVertexAILLM(VertexAILLMOptions{
		GeminiOptions: GeminiOptions{
			APIKey: "vertex-token",
			Model:  "gemini-2.0-flash",
		},
		ProjectID: "project",
		Location:  "us-central1",
	}).ToConfig()

	params := config["params"].(map[string]interface{})
	if config["style"] != "gemini" {
		t.Fatalf("unexpected style: %v", config["style"])
	}
	url, _ := config["url"].(string)
	if !strings.Contains(url, "project") || !strings.Contains(url, "us-central1") {
		t.Fatalf("URL does not contain project routing: %v", url)
	}
	if _, ok := params["project_id"]; ok {
		t.Fatalf("project_id should not be in params, got: %#v", params)
	}
	if _, ok := params["location"]; ok {
		t.Fatalf("location should not be in params, got: %#v", params)
	}
}

func TestAmazonBedrockSerializesAsBedrockStyle(t *testing.T) {
	config := NewAmazonBedrock(AmazonBedrockOptions{
		AccessKey: "aws-access",
		SecretKey: "aws-secret",
		Region:    "us-east-1",
		Model:     "anthropic.claude-3-5-sonnet-20241022-v2:0",
	}).ToConfig()

	if config["style"] != "bedrock" {
		t.Fatalf("unexpected style: %v", config["style"])
	}
	if config["url"] != "https://bedrock-runtime.us-east-1.amazonaws.com/model/anthropic.claude-3-5-sonnet-20241022-v2:0/converse-stream" {
		t.Fatalf("unexpected url: %v", config["url"])
	}
	if config["access_key"] != "aws-access" || config["secret_key"] != "aws-secret" || config["region"] != "us-east-1" {
		t.Fatalf("unexpected config: %#v", config)
	}
}

func TestDifySerializesConversationFields(t *testing.T) {
	config := NewDify(DifyOptions{
		APIKey:         "dify-key",
		URL:            "https://api.dify.ai/v1/chat-messages",
		Model:          "default",
		User:           "user-1",
		ConversationID: "conversation-1",
	}).ToConfig()

	params := config["params"].(map[string]interface{})
	if config["style"] != "dify" {
		t.Fatalf("unexpected style: %v", config["style"])
	}
	if params["model"] != "default" {
		t.Fatalf("unexpected params: %#v", params)
	}
	if params["user"] != "user-1" || params["conversation_id"] != "conversation-1" {
		t.Fatalf("unexpected params: %#v", params)
	}
}

func TestLLMVendorsRequireModels(t *testing.T) {
	maxTokens := 1024
	assertPanic(t, "OpenAI requires Model", func() {
		NewOpenAI(OpenAIOptions{
			APIKey:  "openai-key",
			BaseURL: "https://api.openai.com/v1/chat/completions",
		})
	})
	assertPanic(t, "Anthropic requires Model", func() {
		NewAnthropic(AnthropicOptions{
			APIKey:    "anthropic-key",
			URL:       "https://api.anthropic.com/v1/messages",
			Headers:   map[string]string{"anthropic-version": "2023-06-01"},
			MaxTokens: &maxTokens,
		})
	})
	assertPanic(t, "Gemini requires Model", func() {
		NewGemini(GeminiOptions{APIKey: "google-key"})
	})
	assertPanic(t, "Groq requires Model", func() {
		NewGroq(GroqOptions{
			APIKey:  "groq-key",
			BaseURL: "https://api.groq.com/openai/v1/chat/completions",
		})
	})
	assertPanic(t, "VertexAILLM requires Model", func() {
		NewVertexAILLM(VertexAILLMOptions{
			GeminiOptions: GeminiOptions{APIKey: "vertex-token"},
			ProjectID:     "project",
			Location:      "us-central1",
		})
	})
	assertPanic(t, "AmazonBedrock requires Model", func() {
		NewAmazonBedrock(AmazonBedrockOptions{
			AccessKey: "aws-access",
			SecretKey: "aws-secret",
			Region:    "us-east-1",
		})
	})
}

func TestOpenAIManagedModeIsRestrictedToSupportedModels(t *testing.T) {
	config := NewOpenAI(OpenAIOptions{Model: "gpt-5-mini"}).ToConfig()
	params := config["params"].(map[string]interface{})
	if params["model"] != "gpt-5-mini" {
		t.Fatalf("unexpected params: %#v", params)
	}

	assertPanic(t, "OpenAI requires APIKey unless using a supported Agora-managed model", func() {
		NewOpenAI(OpenAIOptions{Model: "gpt-4o"})
	})
	assertPanic(t, "OpenAI Agora-managed mode does not allow Vendor", func() {
		NewOpenAI(OpenAIOptions{Model: "gpt-5-mini", Vendor: "custom"})
	})
}

func assertPanic(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		got := recover()
		if got == nil {
			t.Fatalf("expected panic %q", want)
		}
		if got != want {
			t.Fatalf("expected panic %q, got %q", want, got)
		}
	}()
	fn()
}
