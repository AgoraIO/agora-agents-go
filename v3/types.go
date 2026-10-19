// Package v3 exposes the staged Conversational AI v3 resource API. The Go
// module's /v2 suffix is its semantic major version, not the REST API version.
package v3

import "encoding/json"

// Response retains headers such as ETag and Agora-Request-Id alongside a body.
type Response[T any] struct {
	Body       T
	ETag       string
	RequestID  string
	StatusCode int
}

type Page[T any] struct {
	Data struct {
		Count int `json:"count"`
		List  []T `json:"list"`
	} `json:"data"`
	HasMore    bool    `json:"has_more"`
	NextCursor *string `json:"next_cursor"`
}

type ListOptions struct {
	Limit  int
	Cursor string
	Order  string
}

type VariableDefinition struct {
	Type        string `json:"type,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Options     []any  `json:"options,omitempty"`
	Extract     bool   `json:"extract,omitempty"`
	Description string `json:"description,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type Greeting struct {
	Mode              string `json:"mode"`
	On                string `json:"on,omitempty"`
	DelayMs           *int   `json:"delay_ms,omitempty"`
	Text              string `json:"text,omitempty"`
	URL               string `json:"url,omitempty"`
	DownloadTimeoutMs *int   `json:"download_timeout_ms,omitempty"`
	PCMSampleRate     *int   `json:"pcm_sample_rate,omitempty"`
}

// ProviderModule keeps provider-specific params open while typing the stable
// credential envelope. Values returned by the API omit write-only api_key.
type ProviderModule struct {
	Vendor           string            `json:"vendor"`
	Credential       *Credential       `json:"credential,omitempty"`
	URL              string            `json:"url,omitempty"`
	Model            string            `json:"model,omitempty"`
	Style            string            `json:"style,omitempty"`
	Language         string            `json:"language,omitempty"`
	Enabled          *bool             `json:"enabled,omitempty"`
	Params           map[string]any    `json:"params,omitempty"`
	Headers          map[string]string `json:"headers,omitempty"`
	InputModalities  []string          `json:"input_modalities,omitempty"`
	OutputModalities []string          `json:"output_modalities,omitempty"`
	MaxHistory       *int              `json:"max_history,omitempty"`
	FailureMessage   string            `json:"failure_message,omitempty"`
	MCPServers       []map[string]any  `json:"mcp_servers,omitempty"`
	Tools            []map[string]any  `json:"tools,omitempty"`
	SkipPatterns     []string          `json:"skip_patterns,omitempty"`
}
type Credential struct {
	Mode   string `json:"mode"`
	APIKey string `json:"api_key,omitempty"`
}

// Pipeline holds the common union discriminator and provider modules. Extra
// provider-specific fields can be supplied in the maps without losing typing
// of the stable top-level shape.
type Pipeline struct {
	Mode          string         `json:"mode"`
	ASR           map[string]any `json:"asr,omitempty"`
	LLM           map[string]any `json:"llm,omitempty"`
	TTS           map[string]any `json:"tts,omitempty"`
	MLLM          map[string]any `json:"mllm,omitempty"`
	Avatar        map[string]any `json:"avatar,omitempty"`
	TurnDetection map[string]any `json:"turn_detection,omitempty"`
	SilenceConfig map[string]any `json:"silence_config,omitempty"`
}

type AgentCreate struct {
	AgentName    string                        `json:"agent_name"`
	Instructions *string                       `json:"instructions,omitempty"`
	Messages     []Message                     `json:"messages,omitempty"`
	Greeting     any                           `json:"greeting,omitempty"`
	Variables    map[string]VariableDefinition `json:"variables,omitempty"`
	Pipeline     *Pipeline                     `json:"pipeline,omitempty"`
	FillerWords  any                           `json:"filler_words,omitempty"`
}

// AgentPatch is an RFC 7396 merge-patch object. Nil values clear fields.
type AgentPatch map[string]any
type Agent struct {
	AgentID      string                        `json:"agent_id"`
	CreateTS     int64                         `json:"create_ts"`
	UpdateTS     int64                         `json:"update_ts"`
	AgentName    string                        `json:"agent_name"`
	Instructions *string                       `json:"instructions"`
	Messages     []Message                     `json:"messages"`
	Greeting     json.RawMessage               `json:"greeting"`
	Variables    map[string]VariableDefinition `json:"variables"`
	Pipeline     json.RawMessage               `json:"pipeline"`
	FillerWords  json.RawMessage               `json:"filler_words"`
}
type AgentListItem struct {
	AgentID   string `json:"agent_id"`
	AgentName string `json:"agent_name"`
	CreateTS  int64  `json:"create_ts"`
	UpdateTS  int64  `json:"update_ts"`
}
type Deleted struct {
	Deleted bool `json:"deleted"`
}

type SecretCreate struct {
	Name    string            `json:"name"`
	Secrets map[string]string `json:"secrets"`
}
type SecretReplace struct {
	Secrets map[string]string `json:"secrets"`
}
type SecretUpdate struct {
	Secrets map[string]*string `json:"secrets"`
}
type SecretSet struct {
	Name     string   `json:"name"`
	Keys     []string `json:"keys"`
	CreateTS int64    `json:"create_ts"`
	UpdateTS int64    `json:"update_ts"`
}
type SecretKey struct {
	Key string `json:"key"`
}
type SecretKeyWrite struct {
	Value string `json:"value"`
}
type SecretKeyOptions struct {
	IfMatch     string
	IfNoneMatch string
}

type SipAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type SipTrunk struct {
	Name      *string  `json:"name,omitempty"`
	Hostname  string   `json:"hostname"`
	Port      *int     `json:"port,omitempty"`
	Transport string   `json:"transport,omitempty"`
	Auth      *SipAuth `json:"auth,omitempty"`
}
type Inbound struct {
	Agent      *string        `json:"agent,omitempty"`
	CallPolicy map[string]any `json:"call_policy,omitempty"`
}
type NumberCreate struct {
	Number      string   `json:"number"`
	Description *string  `json:"description,omitempty"`
	SipTrunk    SipTrunk `json:"sip_trunk"`
	Inbound     *Inbound `json:"inbound,omitempty"`
}
type NumberPatch map[string]any
type Number struct {
	NumberID    string          `json:"number_id"`
	Number      string          `json:"number"`
	Description *string         `json:"description"`
	SipTrunk    json.RawMessage `json:"sip_trunk"`
	Inbound     json.RawMessage `json:"inbound"`
	CreateTS    int64           `json:"create_ts"`
}
type NumberListItem = Number

type SessionAnalysis struct {
	Status     string         `json:"status"`
	Result     map[string]any `json:"result"`
	SkipReason *string        `json:"skip_reason"`
	CompleteTS *int64         `json:"complete_ts"`
	Error      *string        `json:"error"`
}
type Session struct {
	AgentSessionID string           `json:"agent_session_id"`
	StartTS        int64            `json:"start_ts"`
	StopTS         *int64           `json:"stop_ts"`
	Status         string           `json:"status"`
	AgentName      string           `json:"agent_name"`
	Message        *string          `json:"message"`
	Analysis       *SessionAnalysis `json:"analysis"`
}
type SessionListOptions struct {
	ListOptions
	Channel       string
	Status        string
	StartedAfter  *int64
	StartedBefore *int64
}
