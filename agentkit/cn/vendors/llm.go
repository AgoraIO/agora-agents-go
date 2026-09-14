package vendors

import Agora "github.com/AgoraIO/agora-agents-go/v2"

type OpenAIOptions struct {
	APIKey            string
	Model             string
	BaseURL           string
	Temperature       *float64
	TopP              *float64
	MaxTokens         *int
	MaxHistory        *int
	SystemMessages    []map[string]interface{}
	GreetingMessage   string
	FailureMessage    string
	InputModalities   []string
	Params            map[string]interface{}
	Headers           map[string]string
	OutputModalities  []string
	GreetingConfigs   map[string]interface{}
	TemplateVariables map[string]string
	Vendor            string
	McpServers        []map[string]interface{}
	Tools             []*Agora.LlmTool
}

func ensureMcpTransport(servers []map[string]interface{}) []map[string]interface{} {
	if servers == nil {
		return nil
	}
	result := make([]map[string]interface{}, len(servers))
	for i, s := range servers {
		item := make(map[string]interface{})
		for k, v := range s {
			item[k] = v
		}
		if _, ok := item["transport"]; !ok {
			item["transport"] = "streamable_http"
		}
		result[i] = item
	}
	return result
}

type openAICompatibleLLM struct {
	options OpenAIOptions
	label   string
	vendor  string
}

func newOpenAICompatibleLLM(label string, vendor string, opts OpenAIOptions) *openAICompatibleLLM {
	if opts.APIKey == "" {
		panic(label + " requires APIKey")
	}
	if opts.Model == "" {
		panic(label + " requires Model")
	}
	if opts.BaseURL == "" {
		panic(label + " requires BaseURL")
	}
	return &openAICompatibleLLM{
		options: opts,
		label:   label,
		vendor:  vendor,
	}
}

func (o *openAICompatibleLLM) ToConfig() map[string]interface{} {
	params := map[string]interface{}{"model": o.options.Model}
	for k, v := range o.options.Params {
		params[k] = v
	}
	if o.options.Temperature != nil {
		params["temperature"] = *o.options.Temperature
	}
	if o.options.TopP != nil {
		params["top_p"] = *o.options.TopP
	}
	if o.options.MaxTokens != nil {
		params["max_tokens"] = *o.options.MaxTokens
	}

	inputMod := o.options.InputModalities
	if inputMod == nil {
		inputMod = []string{"text"}
	}

	config := map[string]interface{}{
		"url":              o.options.BaseURL,
		"api_key":          o.options.APIKey,
		"params":           params,
		"style":            "openai",
		"input_modalities": inputMod,
		"vendor":           o.vendor,
	}
	if o.options.Headers != nil {
		config["headers"] = o.options.Headers
	}
	if o.options.SystemMessages != nil {
		config["system_messages"] = o.options.SystemMessages
	}
	if o.options.GreetingMessage != "" {
		config["greeting_message"] = o.options.GreetingMessage
	}
	if o.options.FailureMessage != "" {
		config["failure_message"] = o.options.FailureMessage
	}
	if o.options.OutputModalities != nil {
		config["output_modalities"] = o.options.OutputModalities
	}
	if o.options.GreetingConfigs != nil {
		config["greeting_configs"] = o.options.GreetingConfigs
	}
	if o.options.TemplateVariables != nil {
		config["template_variables"] = o.options.TemplateVariables
	}
	if o.options.McpServers != nil {
		config["mcp_servers"] = ensureMcpTransport(o.options.McpServers)
	}
	if o.options.Tools != nil {
		config["tools"] = o.options.Tools
	}
	if o.options.MaxHistory != nil {
		config["max_history"] = *o.options.MaxHistory
	}
	return config
}

type AliyunOptions = OpenAIOptions
type BytedanceOptions = OpenAIOptions
type DeepSeekOptions = OpenAIOptions
type TencentLLMOptions = OpenAIOptions

type Aliyun struct{ *openAICompatibleLLM }
type Bytedance struct{ *openAICompatibleLLM }
type DeepSeek struct{ *openAICompatibleLLM }
type TencentLLM struct{ *openAICompatibleLLM }

func NewAliyun(opts AliyunOptions) *Aliyun {
	return &Aliyun{newOpenAICompatibleLLM("Aliyun", "aliyun", opts)}
}

func NewBytedance(opts BytedanceOptions) *Bytedance {
	return &Bytedance{newOpenAICompatibleLLM("Bytedance", "bytedance", opts)}
}

func NewDeepSeek(opts DeepSeekOptions) *DeepSeek {
	return &DeepSeek{newOpenAICompatibleLLM("DeepSeek", "deepseek", opts)}
}

func NewTencentLLM(opts TencentLLMOptions) *TencentLLM {
	return &TencentLLM{newOpenAICompatibleLLM("TencentLLM", "tencent", opts)}
}
