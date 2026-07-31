package vendors

import Agora "github.com/AgoraIO/agora-agents-go/v2"

const qwenOmniDefaultURL = "wss://dashscope.aliyuncs.com/api-ws/v1/realtime"

// QwenOmniOptions configures the mainland China Qwen Omni Realtime MLLM provider.
type QwenOmniOptions struct {
	APIKey           string
	Model            string
	URL              string
	Voice            string
	Instructions     string
	GreetingMessage  string
	FailureMessage   string
	InputModalities  []string
	OutputModalities []string
	Messages         []map[string]interface{}
	Params           map[string]interface{}
	TurnDetection    *Agora.MllmTurnDetection
}

// QwenOmni is the mainland China Qwen Omni Realtime MLLM provider.
type QwenOmni struct {
	options QwenOmniOptions
}

// NewQwenOmni creates a Qwen Omni MLLM configuration.
func NewQwenOmni(opts QwenOmniOptions) *QwenOmni {
	if opts.APIKey == "" {
		panic("QwenOmni requires APIKey")
	}
	if opts.Model == "" {
		panic("QwenOmni requires Model")
	}
	if opts.TurnDetection == nil {
		panic("QwenOmni requires TurnDetection")
	}
	if opts.URL == "" {
		opts.URL = qwenOmniDefaultURL
	}
	return &QwenOmni{options: opts}
}

// ToConfig returns the Qwen Omni configuration expected by the API.
func (q *QwenOmni) ToConfig() map[string]interface{} {
	params := map[string]interface{}{"model": q.options.Model}
	for key, value := range q.options.Params {
		params[key] = value
	}
	if q.options.Voice != "" {
		params["voice"] = q.options.Voice
	}
	if q.options.Instructions != "" {
		params["instructions"] = q.options.Instructions
	}

	config := map[string]interface{}{
		"vendor":         "qwen_omni",
		"api_key":        q.options.APIKey,
		"url":            q.options.URL,
		"params":         params,
		"turn_detection": q.options.TurnDetection,
	}
	if q.options.GreetingMessage != "" {
		config["greeting_message"] = q.options.GreetingMessage
	}
	if q.options.FailureMessage != "" {
		config["failure_message"] = q.options.FailureMessage
	}
	if q.options.InputModalities != nil {
		config["input_modalities"] = q.options.InputModalities
	}
	if q.options.OutputModalities != nil {
		config["output_modalities"] = q.options.OutputModalities
	}
	if q.options.Messages != nil {
		config["messages"] = q.options.Messages
	}
	return config
}
