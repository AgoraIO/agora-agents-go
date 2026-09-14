package vendors

import (
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
)

func TestDomesticOpenAICompatibleLLMVendorsSetVendorHint(t *testing.T) {
	cases := []struct {
		name   string
		config map[string]interface{}
		vendor string
	}{
		{
			name: "aliyun",
			config: NewAliyun(AliyunOptions{
				APIKey:  "key",
				Model:   "qwen-max",
				BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
			}).ToConfig(),
			vendor: "aliyun",
		},
		{
			name: "bytedance",
			config: NewBytedance(BytedanceOptions{
				APIKey:  "key",
				Model:   "doubao-seed-1-6",
				BaseURL: "https://ark.cn-beijing.volces.com/api/v3/chat/completions",
			}).ToConfig(),
			vendor: "bytedance",
		},
		{
			name: "deepseek",
			config: NewDeepSeek(DeepSeekOptions{
				APIKey:  "key",
				Model:   "deepseek-chat",
				BaseURL: "https://api.deepseek.com/chat/completions",
			}).ToConfig(),
			vendor: "deepseek",
		},
		{
			name: "tencent",
			config: NewTencentLLM(TencentLLMOptions{
				APIKey:  "key",
				Model:   "hunyuan-turbos-latest",
				BaseURL: "https://api.hunyuan.cloud.tencent.com/v1/chat/completions",
			}).ToConfig(),
			vendor: "tencent",
		},
	}

	for _, tc := range cases {
		if tc.config["vendor"] != tc.vendor {
			t.Fatalf("%s unexpected vendor: %#v", tc.name, tc.config)
		}
		if tc.config["style"] != "openai" {
			t.Fatalf("%s unexpected style: %#v", tc.name, tc.config)
		}
		params := tc.config["params"].(map[string]interface{})
		if params["model"] == "" {
			t.Fatalf("%s missing model in params: %#v", tc.name, tc.config)
		}
	}
}

func TestDomesticOpenAICompatibleLLMEmitsInlineTools(t *testing.T) {
	tool := &Agora.LlmTool{
		Function: &Agora.LlmToolFunction{Name: "lookup"},
		Server: &Agora.LlmToolServer{
			Method: Agora.LlmToolServerMethodPost,
			URL:    "https://example.com/items",
		},
	}
	config := NewDeepSeek(DeepSeekOptions{
		APIKey:  "key",
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com/chat/completions",
		Tools:   []*Agora.LlmTool{tool},
	}).ToConfig()

	if _, ok := config["tools"].([]*Agora.LlmTool); !ok {
		t.Fatalf("tools = %#v, want generated tool definitions", config["tools"])
	}
}

func TestCNMicrosoftSTTParams(t *testing.T) {
	config := NewMicrosoftSTT(MicrosoftSTTOptions{
		Key:        "ms-key",
		Region:     "eastus",
		Language:   "zh-CN",
		PhraseList: []string{"shengwang", "convoai"},
	}).ToConfig()

	if config["vendor"] != "microsoft" {
		t.Fatalf("unexpected vendor: %#v", config)
	}
	params := config["params"].(map[string]interface{})
	if params["key"] != "ms-key" || params["region"] != "eastus" || params["language"] != "zh-CN" {
		t.Fatalf("unexpected params: %#v", params)
	}
}
