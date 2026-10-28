// Package v3 exposes the hand-maintained Conversational AI v3 preview resources.
// Its wire contract is checked against the repository-root openapi.json snapshot;
// it is not produced by the Fern generator. The module's /v2 suffix is the Go
// module major version, not the REST API version.
package v3

import (
	"encoding/json"
	"net/http"
)

// Response retains headers such as ETag and Agora-Request-Id alongside a body.
type Response[T any] struct {
	Body T
	// Header is a detached copy of all response headers.
	Header     http.Header
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
	Text              string `json:"text"`
	URL               string `json:"url,omitempty"`
	DownloadTimeoutMs *int   `json:"download_timeout_ms,omitempty"`
	PCMSampleRate     *int   `json:"pcm_sample_rate,omitempty"`
}

// ProviderModule keeps provider-specific params open while typing the stable
// credential envelope. Values returned by the API omit write-only credential fields.
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
	Mode   string            `json:"mode"`
	Fields map[string]string `json:"fields,omitempty"`
}

// Pipeline holds the cascaded preview pipeline and its provider modules. Extra
// provider-specific fields can be supplied in the maps without losing typing
// of the stable top-level shape.
type Pipeline struct {
	Mode          string         `json:"mode"`
	ASR           map[string]any `json:"asr,omitempty"`
	LLM           map[string]any `json:"llm,omitempty"`
	TTS           map[string]any `json:"tts,omitempty"`
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
	Name       *string   `json:"name,omitempty"`
	Hostname   string    `json:"hostname"`
	Port       *int      `json:"port,omitempty"`
	Transport  string    `json:"transport,omitempty"`
	Auth       *SipAuth  `json:"auth,omitempty"`
	AllowedIPs *[]string `json:"allowed_ips,omitempty"`
}
type Inbound struct {
	Agent      *string        `json:"agent"`
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

type Session struct {
	AgentSessionID string `json:"agent_session_id"`
	CreateTS       int64  `json:"create_ts,omitempty"`
	// CreatedAt accepts the rollout-era response field. New responses use CreateTS.
	CreatedAt json.RawMessage `json:"created_at,omitempty"`
	StartTS   int64           `json:"start_ts"`
	StopTS    *int64          `json:"stop_ts"`
	Status    string          `json:"status"`
	AgentName string          `json:"agent_name"`
	Message   *string         `json:"message"`
}

// SessionPatch updates only a running session's RTC token and supported agent
// overrides. Pipeline params are complete replacements when supplied.
type SessionPatch struct {
	Transport *SessionTransportPatch `json:"transport,omitempty"`
	Agent     *SessionAgentPatch     `json:"agent,omitempty"`
}
type SessionTransportPatch struct {
	Token string `json:"token"`
}
type SessionAgentPatch struct {
	Overrides *SessionAgentOverridesPatch `json:"overrides,omitempty"`
}
type SessionAgentOverridesPatch struct {
	Messages *[]Message            `json:"messages,omitempty"`
	Pipeline *SessionPipelinePatch `json:"pipeline,omitempty"`
}
type SessionPipelinePatch struct {
	LLM *SessionModulePatch `json:"llm,omitempty"`
}
type SessionModulePatch struct {
	Params map[string]any `json:"params"`
}

func (p SessionModulePatch) MarshalJSON() ([]byte, error) {
	body := map[string]any{}
	if p.Params != nil {
		body["params"] = p.Params
	}
	return json.Marshal(body)
}

// SessionTransport describes the RTC connection associated with a session.
type SessionTransport struct {
	Type          string              `json:"type"`
	Channel       string              `json:"channel,omitempty"`
	Token         string              `json:"token,omitempty"`
	UID           string              `json:"uid,omitempty"`
	SubscribeUIDs []string            `json:"subscribe_uids,omitempty"`
	Encryption    *RTCEncryption      `json:"encryption,omitempty"`
	AudioScenario string              `json:"audio_scenario,omitempty"`
	Geofence      *Geofence           `json:"geofence,omitempty"`
	From          string              `json:"from,omitempty"`
	To            string              `json:"to,omitempty"`
	CallPolicy    *CampaignCallPolicy `json:"call_policy,omitempty"`
}

type RTCEncryption struct {
	Mode string `json:"mode"`
	Key  string `json:"key"`
	Salt string `json:"salt,omitempty"`
}

type Geofence struct {
	Area        string `json:"area"`
	ExcludeArea string `json:"exclude_area,omitempty"`
}

// SessionLifecycle contains optional session timeouts in milliseconds.
type SessionLifecycle struct {
	IdleTimeoutMS *int          `json:"idle_timeout_ms,omitempty"`
	MaxDurationMS *int          `json:"max_duration_ms,omitempty"`
	GracefulStop  *GracefulStop `json:"graceful_stop,omitempty"`
}

type GracefulStop struct {
	Enabled   bool `json:"enabled"`
	TimeoutMS *int `json:"timeout_ms,omitempty"`
}

type DataPolicy struct {
	Retention string `json:"retention,omitempty"`
}

// SessionStartRequest starts a session from an existing saved agent.
type SessionStartRequest struct {
	AgentID    string            `json:"agent_id"`
	Values     map[string]any    `json:"values,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Transport  SessionTransport  `json:"transport"`
	Lifecycle  *SessionLifecycle `json:"lifecycle,omitempty"`
	DataPolicy *DataPolicy       `json:"data_policy,omitempty"`
}

// EphemeralAgent is an inline agent configuration that is not saved.
type EphemeralAgent struct {
	Instructions *string                       `json:"instructions,omitempty"`
	Messages     []Message                     `json:"messages,omitempty"`
	Greeting     any                           `json:"greeting,omitempty"`
	Variables    map[string]VariableDefinition `json:"variables,omitempty"`
	Pipeline     *Pipeline                     `json:"pipeline,omitempty"`
	FillerWords  any                           `json:"filler_words,omitempty"`
}

// SessionStartEphemeralRequest starts a session from an unsaved agent config.
type SessionStartEphemeralRequest struct {
	Labels     map[string]string `json:"labels,omitempty"`
	Agent      EphemeralAgent    `json:"agent"`
	Values     map[string]any    `json:"values,omitempty"`
	Transport  SessionTransport  `json:"transport"`
	Lifecycle  *SessionLifecycle `json:"lifecycle,omitempty"`
	DataPolicy *DataPolicy       `json:"data_policy,omitempty"`
}

type SessionWriteOptions struct {
	IdempotencyKey string
}
type SessionListOptions struct {
	ListOptions
	Channel       string
	Status        string
	StartedAfter  *int64
	StartedBefore *int64
}

// UnmarshalJSON rejects removed preview fields instead of silently discarding
// them when a caller loads an agent configuration from JSON.
func (p *Pipeline) UnmarshalJSON(data []byte) error {
	if err := validatePreviewPipeline(json.RawMessage(data)); err != nil {
		return err
	}
	type wire Pipeline
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*p = Pipeline(decoded)
	return nil
}

func (p *SessionPipelinePatch) UnmarshalJSON(data []byte) error {
	if err := validatePreviewPipeline(json.RawMessage(data)); err != nil {
		return err
	}
	type wire SessionPipelinePatch
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*p = SessionPipelinePatch(decoded)
	return nil
}
