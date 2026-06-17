package vendors

import (
	"fmt"
	"strings"
)

// ensureMcpTransport sets transport to "streamable_http" on each MCP server
// if not already set (API requires it).
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
}

type OpenAI struct {
	options OpenAIOptions
}

func NewOpenAI(opts OpenAIOptions) *OpenAI {
	if opts.Model == "" {
		panic("OpenAI requires Model")
	}
	if opts.APIKey == "" {
		switch strings.ToLower(strings.TrimSpace(opts.Model)) {
		case "gpt-4o-mini", "gpt-4.1-mini", "gpt-5-nano", "gpt-5-mini":
			if opts.BaseURL != "" {
				panic("OpenAI Agora-managed mode does not allow BaseURL")
			}
			if opts.Vendor != "" {
				panic("OpenAI Agora-managed mode does not allow Vendor")
			}
		default:
			panic("OpenAI requires APIKey unless using a supported Agora-managed model")
		}
	}
	if opts.APIKey != "" && opts.BaseURL == "" {
		panic("OpenAI requires BaseURL when APIKey is set")
	}
	return &OpenAI{options: opts}
}

func (o *OpenAI) ToConfig() map[string]interface{} {
	url := o.options.BaseURL
	if url == "" {
		url = "https://api.openai.com/v1/chat/completions"
	}

	// model is the base; explicit Params entries override it.
	// Always build a fresh map so we never mutate the caller's Params.
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
		"url":              url,
		"params":           params,
		"style":            "openai",
		"input_modalities": inputMod,
	}
	if o.options.APIKey != "" {
		config["api_key"] = o.options.APIKey
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
	if o.options.Vendor != "" {
		config["vendor"] = o.options.Vendor
	}
	if o.options.McpServers != nil {
		config["mcp_servers"] = ensureMcpTransport(o.options.McpServers)
	}
	if o.options.MaxHistory != nil {
		config["max_history"] = *o.options.MaxHistory
	}

	return config
}

type AzureOpenAIOptions struct {
	APIKey string
	// Model is the deployment's underlying model name (e.g., "gpt-4o"). Sent in `params.model`
	// for parity with the TypeScript SDK; Azure ignores the field for chat completions because
	// the deployment determines the model, but downstream tooling and logs use it.
	Model             string
	Endpoint          string
	DeploymentName    string
	APIVersion        string
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
}

type AzureOpenAI struct {
	options AzureOpenAIOptions
}

func NewAzureOpenAI(opts AzureOpenAIOptions) *AzureOpenAI {
	if opts.APIKey == "" {
		panic("AzureOpenAI requires APIKey")
	}
	if opts.Endpoint == "" {
		panic("AzureOpenAI requires Endpoint")
	}
	if opts.DeploymentName == "" {
		panic("AzureOpenAI requires DeploymentName")
	}
	if opts.Model == "" {
		panic("AzureOpenAI requires Model")
	}
	if opts.APIVersion == "" {
		opts.APIVersion = "2024-08-01-preview"
	}
	return &AzureOpenAI{options: opts}
}

func (a *AzureOpenAI) ToConfig() map[string]interface{} {
	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s",
		a.options.Endpoint, a.options.DeploymentName, a.options.APIVersion)

	inputMod := a.options.InputModalities
	if inputMod == nil {
		inputMod = []string{"text"}
	}

	vendor := a.options.Vendor
	if vendor == "" {
		vendor = "azure"
	}
	config := map[string]interface{}{
		"url":              url,
		"api_key":          a.options.APIKey,
		"vendor":           vendor,
		"style":            "openai",
		"input_modalities": inputMod,
	}

	// model is the base; explicit Params entries override it; named fields (temperature/top_p/max_tokens) win.
	params := map[string]interface{}{}
	if a.options.Model != "" {
		params["model"] = a.options.Model
	}
	for k, v := range a.options.Params {
		params[k] = v
	}
	if a.options.Temperature != nil {
		params["temperature"] = *a.options.Temperature
	}
	if a.options.TopP != nil {
		params["top_p"] = *a.options.TopP
	}
	if a.options.MaxTokens != nil {
		params["max_tokens"] = *a.options.MaxTokens
	}
	if len(params) > 0 {
		config["params"] = params
	}
	if a.options.Headers != nil {
		config["headers"] = a.options.Headers
	}

	if a.options.SystemMessages != nil {
		config["system_messages"] = a.options.SystemMessages
	}
	if a.options.GreetingMessage != "" {
		config["greeting_message"] = a.options.GreetingMessage
	}
	if a.options.FailureMessage != "" {
		config["failure_message"] = a.options.FailureMessage
	}
	if a.options.OutputModalities != nil {
		config["output_modalities"] = a.options.OutputModalities
	}
	if a.options.GreetingConfigs != nil {
		config["greeting_configs"] = a.options.GreetingConfigs
	}
	if a.options.TemplateVariables != nil {
		config["template_variables"] = a.options.TemplateVariables
	}
	if a.options.McpServers != nil {
		config["mcp_servers"] = ensureMcpTransport(a.options.McpServers)
	}
	if a.options.MaxHistory != nil {
		config["max_history"] = *a.options.MaxHistory
	}

	return config
}

type AnthropicOptions struct {
	APIKey            string
	Model             string
	URL               string
	MaxTokens         *int
	Temperature       *float64
	TopP              *float64
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
}

type Anthropic struct {
	options AnthropicOptions
}

func NewAnthropic(opts AnthropicOptions) *Anthropic {
	if opts.APIKey == "" {
		panic("Anthropic requires APIKey")
	}
	if opts.MaxTokens == nil {
		panic("Anthropic requires MaxTokens")
	}
	if opts.URL == "" {
		panic("Anthropic requires URL")
	}
	if opts.Headers == nil {
		panic("Anthropic requires Headers")
	}
	if opts.Model == "" {
		panic("Anthropic requires Model")
	}
	return &Anthropic{options: opts}
}

func (a *Anthropic) ToConfig() map[string]interface{} {
	url := a.options.URL
	if url == "" {
		url = "https://api.anthropic.com/v1/messages"
	}

	inputMod := a.options.InputModalities
	if inputMod == nil {
		inputMod = []string{"text"}
	}

	// model is the base; explicit Params entries extend it; named fields win.
	params := map[string]interface{}{"model": a.options.Model}
	for k, v := range a.options.Params {
		params[k] = v
	}
	if a.options.MaxTokens != nil {
		params["max_tokens"] = *a.options.MaxTokens
	}
	if a.options.Temperature != nil {
		params["temperature"] = *a.options.Temperature
	}
	if a.options.TopP != nil {
		params["top_p"] = *a.options.TopP
	}

	config := map[string]interface{}{
		"url":              url,
		"api_key":          a.options.APIKey,
		"params":           params,
		"headers":          a.options.Headers,
		"style":            "anthropic",
		"input_modalities": inputMod,
	}

	if a.options.SystemMessages != nil {
		config["system_messages"] = a.options.SystemMessages
	}
	if a.options.GreetingMessage != "" {
		config["greeting_message"] = a.options.GreetingMessage
	}
	if a.options.FailureMessage != "" {
		config["failure_message"] = a.options.FailureMessage
	}
	if a.options.OutputModalities != nil {
		config["output_modalities"] = a.options.OutputModalities
	}
	if a.options.GreetingConfigs != nil {
		config["greeting_configs"] = a.options.GreetingConfigs
	}
	if a.options.TemplateVariables != nil {
		config["template_variables"] = a.options.TemplateVariables
	}
	if a.options.Vendor != "" {
		config["vendor"] = a.options.Vendor
	}
	if a.options.McpServers != nil {
		config["mcp_servers"] = ensureMcpTransport(a.options.McpServers)
	}
	if a.options.MaxHistory != nil {
		config["max_history"] = *a.options.MaxHistory
	}

	return config
}

type GeminiOptions struct {
	APIKey            string
	Model             string
	URL               string
	Temperature       *float64
	TopP              *float64
	TopK              *int
	MaxOutputTokens   *int
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
}

type Gemini struct {
	options GeminiOptions
}

func NewGemini(opts GeminiOptions) *Gemini {
	if opts.APIKey == "" {
		panic("Gemini requires APIKey")
	}
	if opts.Model == "" {
		panic("Gemini requires Model")
	}
	return &Gemini{options: opts}
}

func (g *Gemini) ToConfig() map[string]interface{} {
	url := g.options.URL
	if url == "" {
		url = "https://generativelanguage.googleapis.com/v1beta/models"
	}

	inputMod := g.options.InputModalities
	if inputMod == nil {
		inputMod = []string{"text"}
	}

	// model is the base; explicit Params entries extend it; named fields win.
	params := map[string]interface{}{"model": g.options.Model}
	for k, v := range g.options.Params {
		params[k] = v
	}
	if g.options.Temperature != nil {
		params["temperature"] = *g.options.Temperature
	}
	if g.options.TopP != nil {
		params["top_p"] = *g.options.TopP
	}
	if g.options.TopK != nil {
		params["top_k"] = *g.options.TopK
	}
	if g.options.MaxOutputTokens != nil {
		params["max_output_tokens"] = *g.options.MaxOutputTokens
	}

	config := map[string]interface{}{
		"url":              url,
		"api_key":          g.options.APIKey,
		"params":           params,
		"style":            "gemini",
		"input_modalities": inputMod,
	}

	if g.options.SystemMessages != nil {
		config["system_messages"] = g.options.SystemMessages
	}
	if g.options.Headers != nil {
		config["headers"] = g.options.Headers
	}
	if g.options.GreetingMessage != "" {
		config["greeting_message"] = g.options.GreetingMessage
	}
	if g.options.FailureMessage != "" {
		config["failure_message"] = g.options.FailureMessage
	}
	if g.options.OutputModalities != nil {
		config["output_modalities"] = g.options.OutputModalities
	}
	if g.options.GreetingConfigs != nil {
		config["greeting_configs"] = g.options.GreetingConfigs
	}
	if g.options.TemplateVariables != nil {
		config["template_variables"] = g.options.TemplateVariables
	}
	if g.options.Vendor != "" {
		config["vendor"] = g.options.Vendor
	}
	if g.options.McpServers != nil {
		config["mcp_servers"] = ensureMcpTransport(g.options.McpServers)
	}
	if g.options.MaxHistory != nil {
		config["max_history"] = *g.options.MaxHistory
	}

	return config
}

type GroqOptions = OpenAIOptions

type Groq struct {
	options GroqOptions
}

func NewGroq(opts GroqOptions) *Groq {
	if opts.APIKey == "" {
		panic("Groq requires APIKey")
	}
	if opts.Model == "" {
		panic("Groq requires Model")
	}
	if opts.BaseURL == "" {
		panic("Groq requires BaseURL")
	}
	return &Groq{options: opts}
}

func (g *Groq) ToConfig() map[string]interface{} {
	opts := g.options
	return (&OpenAI{options: opts}).ToConfig()
}

type CustomLLMOptions = OpenAIOptions

type CustomLLM struct {
	options CustomLLMOptions
}

func NewCustomLLM(opts CustomLLMOptions) *CustomLLM {
	if opts.APIKey == "" {
		panic("CustomLLM requires APIKey")
	}
	if opts.BaseURL == "" {
		panic("CustomLLM requires BaseURL")
	}
	if opts.Model == "" {
		panic("CustomLLM requires Model")
	}
	return &CustomLLM{options: opts}
}

func (c *CustomLLM) ToConfig() map[string]interface{} {
	opts := c.options
	if opts.Vendor == "" {
		opts.Vendor = "custom"
	}
	return (&OpenAI{options: opts}).ToConfig()
}

type VertexAILLMOptions struct {
	GeminiOptions
	ProjectID string
	Location  string
}

type VertexAILLM struct {
	options VertexAILLMOptions
}

func NewVertexAILLM(opts VertexAILLMOptions) *VertexAILLM {
	if opts.APIKey == "" {
		panic("VertexAILLM requires APIKey")
	}
	if opts.ProjectID == "" {
		panic("VertexAILLM requires ProjectID")
	}
	if opts.Location == "" {
		panic("VertexAILLM requires Location")
	}
	if opts.Model == "" {
		panic("VertexAILLM requires Model")
	}
	return &VertexAILLM{options: opts}
}

func (v *VertexAILLM) ToConfig() map[string]interface{} {
	opts := v.options.GeminiOptions
	opts.APIKey = v.options.APIKey
	opts.Model = v.options.Model
	if v.options.URL != "" {
		opts.URL = v.options.URL
	} else {
		opts.URL = fmt.Sprintf(
			"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:streamGenerateContent?alt=sse",
			v.options.Location, v.options.ProjectID, v.options.Location, v.options.Model,
		)
	}
	return (&Gemini{options: opts}).ToConfig()
}

type AmazonBedrockOptions struct {
	AccessKey         string
	SecretKey         string
	Region            string
	Model             string
	MaxHistory        *int
	MaxTokens         *int
	Temperature       *float64
	TopP              *float64
	TopK              *int
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
}

type AmazonBedrock struct {
	options AmazonBedrockOptions
}

func NewAmazonBedrock(opts AmazonBedrockOptions) *AmazonBedrock {
	if opts.AccessKey == "" {
		panic("AmazonBedrock requires AccessKey")
	}
	if opts.SecretKey == "" {
		panic("AmazonBedrock requires SecretKey")
	}
	if opts.Region == "" {
		panic("AmazonBedrock requires Region")
	}
	if opts.Model == "" {
		panic("AmazonBedrock requires Model")
	}
	return &AmazonBedrock{options: opts}
}

func (a *AmazonBedrock) ToConfig() map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range a.options.Params {
		params[k] = v
	}
	if a.options.MaxTokens != nil {
		params["max_tokens"] = *a.options.MaxTokens
	}
	if a.options.Temperature != nil {
		params["temperature"] = *a.options.Temperature
	}
	if a.options.TopP != nil {
		params["top_p"] = *a.options.TopP
	}
	if a.options.TopK != nil {
		params["top_k"] = *a.options.TopK
	}

	config := map[string]interface{}{
		"url":              fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/converse-stream", a.options.Region, a.options.Model),
		"access_key":       a.options.AccessKey,
		"secret_key":       a.options.SecretKey,
		"region":           a.options.Region,
		"model":            a.options.Model,
		"params":           params,
		"style":            "bedrock",
		"input_modalities": []string{"text"},
	}
	if a.options.InputModalities != nil {
		config["input_modalities"] = a.options.InputModalities
	}
	if a.options.OutputModalities != nil {
		config["output_modalities"] = a.options.OutputModalities
	}
	if a.options.Headers != nil {
		config["headers"] = a.options.Headers
	}
	if a.options.MaxHistory != nil {
		config["max_history"] = *a.options.MaxHistory
	}
	if a.options.SystemMessages != nil {
		config["system_messages"] = a.options.SystemMessages
	}
	if a.options.GreetingMessage != "" {
		config["greeting_message"] = a.options.GreetingMessage
	}
	if a.options.FailureMessage != "" {
		config["failure_message"] = a.options.FailureMessage
	}
	if a.options.GreetingConfigs != nil {
		config["greeting_configs"] = a.options.GreetingConfigs
	}
	if a.options.TemplateVariables != nil {
		config["template_variables"] = a.options.TemplateVariables
	}
	if a.options.Vendor != "" {
		config["vendor"] = a.options.Vendor
	}
	if a.options.McpServers != nil {
		config["mcp_servers"] = ensureMcpTransport(a.options.McpServers)
	}
	return config
}

type DifyOptions struct {
	APIKey            string
	URL               string
	Model             string
	User              string
	ConversationID    string
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
}

type Dify struct {
	options DifyOptions
}

func NewDify(opts DifyOptions) *Dify {
	if opts.APIKey == "" {
		panic("Dify requires APIKey")
	}
	if opts.URL == "" {
		panic("Dify requires URL")
	}
	if opts.Model == "" {
		panic("Dify requires Model")
	}
	return &Dify{options: opts}
}

func (d *Dify) ToConfig() map[string]interface{} {
	inputMod := d.options.InputModalities
	if inputMod == nil {
		inputMod = []string{"text"}
	}
	params := map[string]interface{}{"model": d.options.Model}
	for k, v := range d.options.Params {
		params[k] = v
	}
	if d.options.User != "" {
		params["user"] = d.options.User
	}
	if d.options.ConversationID != "" {
		params["conversation_id"] = d.options.ConversationID
	}
	config := map[string]interface{}{
		"url":              d.options.URL,
		"api_key":          d.options.APIKey,
		"params":           params,
		"style":            "dify",
		"input_modalities": inputMod,
	}
	if d.options.Headers != nil {
		config["headers"] = d.options.Headers
	}
	if d.options.SystemMessages != nil {
		config["system_messages"] = d.options.SystemMessages
	}
	if d.options.GreetingMessage != "" {
		config["greeting_message"] = d.options.GreetingMessage
	}
	if d.options.FailureMessage != "" {
		config["failure_message"] = d.options.FailureMessage
	}
	if d.options.OutputModalities != nil {
		config["output_modalities"] = d.options.OutputModalities
	}
	if d.options.GreetingConfigs != nil {
		config["greeting_configs"] = d.options.GreetingConfigs
	}
	if d.options.TemplateVariables != nil {
		config["template_variables"] = d.options.TemplateVariables
	}
	if d.options.Vendor != "" {
		config["vendor"] = d.options.Vendor
	}
	if d.options.McpServers != nil {
		config["mcp_servers"] = ensureMcpTransport(d.options.McpServers)
	}
	if d.options.MaxHistory != nil {
		config["max_history"] = *d.options.MaxHistory
	}
	return config
}
