package agentkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	agentcore "github.com/AgoraIO/agora-agents-go/v2/agentkit/core"
	"github.com/AgoraIO/agora-agents-go/v2/core"
	"github.com/AgoraIO/agora-agents-go/v2/option"
	"github.com/AgoraIO/agora-agents-go/v2/v3"
)

var (
	ErrUnsupportedField      = errors.New("agentkit: fields cannot be mapped to v3")
	ErrPreconditionFailed    = errors.New("agentkit: precondition failed; fetch a fresh ETag")
	ErrNumberNotFound        = errors.New("agentkit: number not found in this project")
	ErrSecretRefNotSupported = agentcore.ErrSecretRefNotSupported
)

var variableNameRE = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
var secretPartRE = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// SecretRef returns a reference accepted by v3 provider credentials.
// Invalid components panic because the returned string has no error channel.
func SecretRef(set, key string) string {
	if !secretPartRE.MatchString(set) || !secretPartRE.MatchString(key) {
		panic("agentkit: invalid secret reference component")
	}
	return "$secrets." + set + "." + key
}

type VariableDefinition = v3.VariableDefinition

// Project resource aliases let AgentKit users provision secrets and numbers
// without importing the versioned low-level package.
type SecretCreate = v3.SecretCreate
type SecretSet = v3.SecretSet
type SecretUpdate = v3.SecretUpdate
type SecretReplace = v3.SecretReplace
type NumberCreate = v3.NumberCreate
type Number = v3.Number
type SipTrunk = v3.SipTrunk
type SipAuth = v3.SipAuth
type Inbound = v3.Inbound

func shouldStartEphemeralV3(base *agentcore.BaseAgent, opts CreateSessionOptions) bool {
	if base == nil {
		return false
	}
	if len(base.Variables) > 0 || opts.ClientReference != "" || opts.MaxDuration != nil || opts.Labels != nil || opts.Values != nil || opts.Encryption != nil || opts.Geofence != nil || opts.AudioScenario != "" || opts.GracefulStop != nil || opts.DataPolicy != nil || opts.IdempotencyKey != "" || base.Geofence != nil || base.AudioScenario != nil || (base.Parameters != nil && (base.Parameters.FarewellConfig != nil || base.Parameters.OptOut != nil)) {
		return true
	}
	for _, module := range []map[string]interface{}{base.STT, base.LLM, base.TTS, base.MLLM} {
		if containsSecretReference(module) {
			return true
		}
	}
	return false
}

// toV3Session removes legacy agent-level fields that the session API carries
// in transport, lifecycle, or data_policy instead.
func (a *Agent) toV3Session() (v3.AgentCreate, error) {
	if a == nil || a.base == nil {
		return v3.AgentCreate{}, errors.New("agentkit: agent config is required")
	}
	copyAgent := *a
	copyBase := *a.base
	copyBase.Geofence = nil
	copyBase.AudioScenario = nil
	if copyBase.Parameters != nil {
		parameters := *copyBase.Parameters
		parameters.FarewellConfig = nil
		parameters.OptOut = nil
		copyBase.Parameters = &parameters
	}
	copyAgent.base = &copyBase
	return copyAgent.toV3("")
}

func containsSecretReference(value any) bool {
	switch item := value.(type) {
	case string:
		return strings.HasPrefix(strings.TrimSpace(item), "$secrets.")
	case map[string]interface{}:
		for _, child := range item {
			if containsSecretReference(child) {
				return true
			}
		}
	case []interface{}:
		for _, child := range item {
			if containsSecretReference(child) {
				return true
			}
		}
	}
	return false
}

func WithVariable(name string, def VariableDefinition) AgentOption {
	if err := validateVariable(name, def); err != nil {
		panic(err)
	}
	// Capture caller-owned slices now, then copy again for each application.
	// Reusing an option must not alias builders or mutate this closure.
	def.Options = append([]any(nil), def.Options...)
	return func(a *agentcore.BaseAgent) {
		if a.Variables == nil {
			a.Variables = map[string]VariableDefinition{}
		}
		value := def
		value.Options = append([]any(nil), def.Options...)
		a.Variables[name] = value
	}
}

func (a *Agent) WithVariable(name string, def VariableDefinition) *Agent {
	if err := validateVariable(name, def); err != nil {
		panic(err)
	}
	clone := a.base.Clone()
	if clone.Variables == nil {
		clone.Variables = map[string]VariableDefinition{}
	}
	def.Options = append([]any(nil), def.Options...)
	clone.Variables[name] = def
	return &Agent{base: clone}
}
func validateVariable(name string, def VariableDefinition) error {
	if !variableNameRE.MatchString(name) {
		return fmt.Errorf("agentkit: invalid variable name %q", name)
	}
	switch def.Type {
	case "str", "int", "float", "bool", "date", "datetime":
	default:
		return fmt.Errorf("agentkit: invalid variable type for %s", name)
	}
	if def.Extract && strings.TrimSpace(def.Description) == "" {
		return fmt.Errorf("agentkit: extracted variable %s requires description", name)
	}
	if utf8.RuneCountInString(def.Description) > 2000 {
		return fmt.Errorf("agentkit: variable %s description exceeds 2000 characters", name)
	}
	if def.Options != nil {
		if len(def.Options) < 1 || len(def.Options) > 100 {
			return fmt.Errorf("agentkit: variable %s options must contain 1-100 scalars", name)
		}
		seen := make(map[string]bool, len(def.Options))
		for _, v := range def.Options {
			var kind string
			switch v.(type) {
			case string:
				kind = "string:"
			case bool:
				kind = "bool:"
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
				kind = "number:"
			default:
				return fmt.Errorf("agentkit: variable %s option is not a scalar", name)
			}
			encoded, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("agentkit: variable %s option is not a JSON scalar: %w", name, err)
			}
			value := string(encoded)
			if kind == "number:" {
				// Compare the numbers actually sent as JSON, including integral
				// floats and signed zero, without rounding large integer values.
				number, ok := new(big.Rat).SetString(value)
				if !ok {
					return fmt.Errorf("agentkit: variable %s option is not a JSON number", name)
				}
				value = number.RatString()
			}
			key := kind + value
			if seen[key] {
				return fmt.Errorf("agentkit: variable %s options must be unique", name)
			}
			seen[key] = true
		}
	}
	return nil
}

// SaveOptions identifies the stored configuration to create or update.
// Leave AgentID empty to create a new saved agent.
type SaveOptions struct {
	Name    string
	AgentID string
}

// DeployOptions is retained for source compatibility. New code should use SaveOptions.
type DeployOptions = SaveOptions

// SavedAgent represents a stored agent. Raw exposes the low-level client.
type SavedAgent struct {
	ID     string
	ETag   string
	Name   string
	Raw    *v3.AgentsClient
	client *AgoraClient
}

// DeployedAgent is retained as a source-compatible name for SavedAgent.
type DeployedAgent = SavedAgent

// Save creates a stored configuration, or updates the saved agent identified
// by opts.AgentID. It does not start a running session.
func (a *Agent) Save(ctx context.Context, opts SaveOptions) (*SavedAgent, error) {
	return a.save(ctx, opts.Name, opts.AgentID)
}

// Deploy is retained for source compatibility. New code should use Save.
func (a *Agent) Deploy(ctx context.Context, opts DeployOptions) (*DeployedAgent, error) {
	return a.save(ctx, opts.Name, opts.AgentID)
}

func (a *Agent) save(ctx context.Context, name, id string) (*SavedAgent, error) {
	if a == nil || a.base == nil {
		return nil, errors.New("agentkit: agent is required")
	}
	c, ok := a.base.Client.(*AgoraClient)
	if !ok || c == nil {
		return nil, errors.New("agentkit: save requires AgoraClient")
	}
	if c.Area == option.AreaCN {
		return nil, errors.New("agentkit: saved agents in CN are pending backend confirmation")
	}
	if name == "" {
		name = fmt.Sprintf("agent-%d", time.Now().Unix())
	}
	body, err := a.toV3(name)
	if err != nil {
		return nil, err
	}
	if id == "" {
		r, err := c.V3.Agents.Create(ctx, body)
		if err != nil {
			return nil, err
		}
		return &DeployedAgent{ID: r.Body.AgentID, ETag: r.ETag, Name: r.Body.AgentName, Raw: c.V3.Agents, client: c}, nil
	}
	current, err := c.V3.Agents.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	patch, err := deployPatch(current.Body, body)
	if err != nil {
		return nil, err
	}
	r, err := c.V3.Agents.Update(ctx, id, current.ETag, patch)
	if err != nil {
		return nil, translatePrecondition(err)
	}
	return &SavedAgent{ID: r.Body.AgentID, ETag: r.ETag, Name: r.Body.AgentName, Raw: c.V3.Agents, client: c}, nil
}

// GetSavedAgent loads a stored configuration by ID.
func GetSavedAgent(ctx context.Context, c *AgoraClient, id string) (*SavedAgent, error) {
	if c == nil {
		return nil, errors.New("agentkit: client is required")
	}
	r, err := c.V3.Agents.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &SavedAgent{ID: r.Body.AgentID, ETag: r.ETag, Name: r.Body.AgentName, Raw: c.V3.Agents, client: c}, nil
}

// GetDeployedAgent is retained for source compatibility. New code should use GetSavedAgent.
func GetDeployedAgent(ctx context.Context, c *AgoraClient, id string) (*DeployedAgent, error) {
	return GetSavedAgent(ctx, c, id)
}

// Update saves config over this stored agent and returns a handle with the new
// ETag. Both handles must belong to the same AgoraClient instance.
func (d *SavedAgent) Update(ctx context.Context, config *Agent) (*SavedAgent, error) {
	if d == nil || d.client == nil || d.ID == "" {
		return nil, errors.New("agentkit: saved agent with ID and client is required")
	}
	if config == nil || config.base == nil {
		return nil, errors.New("agentkit: agent config is required")
	}
	c, ok := config.base.Client.(*AgoraClient)
	if !ok || c != d.client {
		return nil, errors.New("agentkit: saved agent update requires the same AgoraClient instance")
	}
	return config.Save(ctx, SaveOptions{Name: d.Name, AgentID: d.ID})
}

// CreateSession builds a v3 session that starts this saved agent. V3 session
// objects have their own lifecycle; legacy actions such as Say are unsupported.
func (d *SavedAgent) CreateSession(opts CreateSessionOptions) *AgentSession {
	if d == nil || d.client == nil || d.ID == "" {
		return &AgentSession{
			backend:  sessionBackendV3Saved,
			status:   StatusIdle,
			handlers: make(map[string][]EventHandler),
			initErr:  errors.New("agentkit: saved agent with ID and client is required"),
		}
	}
	session := newV3AgentSession(NewAgent(d.client), d.client, opts, d.ID)
	return session
}

func (d *SavedAgent) Delete(ctx context.Context) error {
	_, err := d.Raw.Delete(ctx, d.ID)
	return err
}

func translatePrecondition(err error) error {
	var apiErr *core.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusPreconditionFailed {
		return errors.Join(ErrPreconditionFailed, err)
	}
	return err
}

func (a *Agent) toV3(name string) (v3.AgentCreate, error) {
	b := a.base
	unsupported := []string{}
	if b.PipelineID != "" {
		unsupported = append(unsupported, "pipelineId")
	}
	if b.Sal != nil {
		unsupported = append(unsupported, "sal")
	}
	if b.AdvancedFeatures != nil {
		unsupported = append(unsupported, "advancedFeatures")
	}
	if b.Geofence != nil {
		unsupported = append(unsupported, "geofence")
	}
	if len(b.Labels) > 0 {
		unsupported = append(unsupported, "labels")
	}
	if b.RTC != nil {
		unsupported = append(unsupported, "rtc")
	}
	if b.AudioScenario != nil {
		unsupported = append(unsupported, "audioScenario")
	}
	if b.Parameters != nil {
		m, _ := agentcore.StructToMap(b.Parameters)
		for k, v := range m {
			if k != "silence_config" && v != nil {
				unsupported = append(unsupported, "parameters."+k)
			}
		}
	}
	if b.MLLM != nil && (b.LLM != nil || b.TTS != nil || b.STT != nil) {
		unsupported = append(unsupported, "cascaded modules with mllm")
	}
	for name, def := range b.Variables {
		if err := validateVariable(name, def); err != nil {
			unsupported = append(unsupported, fmt.Sprintf("variables.%s (%v)", name, err))
		}
	}
	pipe := &v3.Pipeline{Mode: "cascaded"}
	if b.MLLM != nil {
		pipe.Mode = "realtime"
	}
	var modules []struct {
		name   string
		config map[string]interface{}
		target *map[string]any
	}
	if b.MLLM != nil {
		modules = append(modules, struct {
			name   string
			config map[string]interface{}
			target *map[string]any
		}{"mllm", b.MLLM, &pipe.MLLM})
	} else {
		if b.STT != nil {
			modules = append(modules, struct {
				name   string
				config map[string]interface{}
				target *map[string]any
			}{"asr", b.STT, &pipe.ASR})
		}
		if b.LLM != nil {
			modules = append(modules, struct {
				name   string
				config map[string]interface{}
				target *map[string]any
			}{"llm", b.LLM, &pipe.LLM})
		}
		if b.TTS != nil {
			modules = append(modules, struct {
				name   string
				config map[string]interface{}
				target *map[string]any
			}{"tts", b.TTS, &pipe.TTS})
		}
	}
	if b.Avatar != nil {
		modules = append(modules, struct {
			name   string
			config map[string]interface{}
			target *map[string]any
		}{"avatar", b.Avatar, &pipe.Avatar})
	}
	for _, m := range modules {
		converted, bad := convertModule(m.name, m.config)
		*m.target = converted
		unsupported = append(unsupported, bad...)
	}
	turn := b.TurnDetection
	if turn != nil && turn.Language != nil {
		if pipe.Mode == "realtime" {
			unsupported = append(unsupported, "turnDetection.language in realtime pipeline")
		} else {
			// Match the global v2 ASR default and language precedence. Do not
			// invent a provider endpoint; managed endpoint validation is server-side.
			if pipe.ASR == nil {
				pipe.ASR = map[string]any{"vendor": "ares", "credential": map[string]any{"mode": "managed"}}
			}
			language := string(*turn.Language)
			if language == "" {
				language = "en-US"
			}
			pipe.ASR["language"] = language
		}
	}
	var vendorTurn map[string]interface{}
	if b.MLLM != nil {
		if x, ok := b.MLLM["turn_detection"]; ok && x != nil {
			raw, marshalErr := json.Marshal(x)
			if marshalErr == nil {
				marshalErr = json.Unmarshal(raw, &vendorTurn)
			}
			if marshalErr != nil || vendorTurn == nil {
				unsupported = append(unsupported, "mllm.turn_detection")
			}
		}
	}
	if vendorTurn != nil || turn != nil {
		var td map[string]any
		var bad []string
		if vendorTurn != nil {
			td, bad = convertTurnMap(vendorTurn, b.Interruption, pipe.Mode)
		} else {
			td, bad = convertTurnDetection(turn, b.Interruption, pipe.Mode)
		}
		pipe.TurnDetection = td
		unsupported = append(unsupported, bad...)
	} else if b.Interruption != nil {
		if pipe.Mode == "realtime" {
			unsupported = append(unsupported, "interruption in realtime pipeline")
		} else {
			m, _ := agentcore.StructToMap(b.Interruption)
			pipe.TurnDetection = map[string]any{"interruption": convertInterruption(m)}
		}
	}
	if b.Parameters != nil && b.Parameters.SilenceConfig != nil {
		if pipe.Mode == "realtime" {
			unsupported = append(unsupported, "parameters.silence_config in realtime pipeline")
		} else {
			m, _ := agentcore.StructToMap(b.Parameters.SilenceConfig)
			m["enabled"] = true
			pipe.SilenceConfig = m
		}
	}
	var out v3.AgentCreate
	out.AgentName = name
	out.Pipeline = pipe
	out.Variables = b.Variables
	llm := b.LLM
	if b.MLLM != nil {
		llm = b.MLLM
	}
	if llm != nil {
		if x, ok := llm["template_variables"]; ok && x != nil && !emptyValue(x) {
			unsupported = append(unsupported, "llm.template_variables")
		}
		if x, ok := llm["system_messages"]; ok && x != nil {
			msgs, bad := convertMessages(x)
			out.Instructions, out.Messages = splitMessages(msgs)
			unsupported = append(unsupported, bad...)
		}
		if x, ok := llm["messages"]; ok && x != nil {
			msgs, bad := convertMessages(x)
			out.Instructions, out.Messages = splitMessages(msgs)
			unsupported = append(unsupported, bad...)
		}
		if params, ok := llm["params"].(map[string]interface{}); ok && b.MLLM != nil {
			if s, ok := params["instructions"].(string); ok && s != "" {
				if out.Instructions == nil {
					out.Instructions = &s
				}
				if pipe.MLLM != nil {
					if pm, ok := pipe.MLLM["params"].(map[string]interface{}); ok {
						delete(pm, "instructions")
					}
				}
			}
		}
	}
	if out.Instructions == nil && b.Instructions != "" {
		out.Instructions = &b.Instructions
	}
	if pipe.LLM != nil {
		if _, ok := pipe.LLM["max_history"]; !ok && b.MaxHistory != nil {
			pipe.LLM["max_history"] = *b.MaxHistory
		}
		if _, ok := pipe.LLM["failure_message"]; !ok && b.FailureMessage != "" {
			pipe.LLM["failure_message"] = b.FailureMessage
		}
	} else {
		if b.MaxHistory != nil {
			unsupported = append(unsupported, "maxHistory")
		}
		if b.FailureMessage != "" {
			unsupported = append(unsupported, "failureMessage")
		}
	}
	greetingText := b.Greeting
	if llm != nil {
		if s, ok := llm["greeting_message"].(string); ok && s != "" {
			greetingText = s
		}
	}
	var gc map[string]interface{}
	if b.GreetingConfigs != nil {
		gc = *b.GreetingConfigs
	}
	if llm != nil {
		if x, ok := llm["greeting_configs"]; ok && x != nil {
			if m, ok := x.(map[string]interface{}); ok {
				gc = m
			}
		}
	}
	greet := map[string]any{}
	if greetingText != "" {
		greet["mode"] = "text"
		greet["text"] = greetingText
	}
	if llm != nil {
		if audio, ok := llm["greeting_audio_url"].(string); ok && audio != "" {
			greet["mode"] = "audio"
			greet["url"] = audio
			greet["text"] = greetingText
		}
	}
	if len(gc) > 0 && len(greet) == 0 {
		unsupported = append(unsupported, "greetingConfigs without greeting")
	}
	if gc != nil {
		if mode, exists := gc["mode"]; exists {
			var on string
			switch mode {
			case "single_every":
				on = "each_join"
			case "single_first":
				on = "first_join"
			default:
				unsupported = append(unsupported, "greetingConfigs.mode")
			}
			if _, explicit := gc["on"]; !explicit && on != "" {
				greet["on"] = on
			}
		}
		for k, v := range gc {
			switch k {
			case "mode": // Legacy aliases are normalized above; explicit on wins.
			case "on", "delay_ms":
				greet[k] = v
			case "audio_download_timeout_ms", "audio_pcm_sample_rate":
				if v == nil {
					continue
				}
				if greet["mode"] != "audio" {
					unsupported = append(unsupported, "greetingConfigs."+k+" requires audio greeting")
					continue
				}
				field := "download_timeout_ms"
				if k == "audio_pcm_sample_rate" {
					field = "pcm_sample_rate"
				}
				greet[field] = v
			default:
				unsupported = append(unsupported, "greetingConfigs."+k)
			}
		}
	}
	if len(greet) > 0 {
		out.Greeting = greet
	}
	if b.FillerWords != nil {
		m, _ := agentcore.StructToMap(b.FillerWords)
		fw, bad := convertFillerWords(m)
		out.FillerWords = fw
		unsupported = append(unsupported, bad...)
	}
	if len(unsupported) > 0 {
		sort.Strings(unsupported)
		return v3.AgentCreate{}, fmt.Errorf("%w: %s", ErrUnsupportedField, strings.Join(unsupported, ", "))
	}
	return out, nil
}

func convertMessages(x any) ([]v3.Message, []string) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, []string{"messages"}
	}
	// Pointer content distinguishes a valid empty string from an absent or
	// null required field, which decoding directly into v3.Message cannot do.
	var decoded []struct {
		Role    string  `json:"role"`
		Content *string `json:"content"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(b)))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&decoded); err != nil {
		return nil, []string{"messages"}
	}
	out := make([]v3.Message, 0, len(decoded))
	for _, m := range decoded {
		if (m.Role != "system" && m.Role != "user" && m.Role != "assistant") || m.Content == nil {
			return nil, []string{"messages"}
		}
		out = append(out, v3.Message{Role: m.Role, Content: *m.Content})
	}
	return out, nil
}

func splitMessages(messages []v3.Message) (*string, []v3.Message) {
	system := []string{}
	other := []v3.Message{}
	for _, m := range messages {
		if m.Role == "system" {
			system = append(system, m.Content)
		} else {
			other = append(other, m)
		}
	}
	if len(system) == 0 {
		return nil, other
	}
	s := strings.Join(system, "\n")
	return &s, other
}
func emptyValue(x any) bool {
	switch v := x.(type) {
	case map[string]any:
		return len(v) == 0
	case []any:
		return len(v) == 0
	}
	return false
}

func convertFillerWords(f map[string]interface{}) (map[string]any, []string) {
	out := map[string]any{}
	bad := []string{}
	for k, v := range f {
		switch k {
		case "enable":
			out["enable"] = v
		case "trigger":
			m, ok := v.(map[string]interface{})
			if !ok {
				bad = append(bad, "fillerWords.trigger")
				continue
			}
			t := map[string]any{}
			if x, ok := m["mode"]; ok {
				t["mode"] = x
			}
			for key, val := range m {
				switch key {
				case "mode":
				case "fixed_time_config":
					if sub, ok := val.(map[string]interface{}); ok {
						for sk, sv := range sub {
							t[sk] = sv
						}
					} else {
						bad = append(bad, "fillerWords.trigger.fixed_time_config")
					}
				default:
					bad = append(bad, "fillerWords.trigger."+key)
				}
			}
			out["trigger"] = t
		case "content":
			m, ok := v.(map[string]interface{})
			if !ok {
				bad = append(bad, "fillerWords.content")
				continue
			}
			c := map[string]any{}
			mode, _ := m["mode"].(string)
			c["mode"] = mode
			if mode == "static" {
				if sub, ok := m["static_config"].(map[string]interface{}); ok {
					for sk, sv := range sub {
						c[sk] = sv
					}
				}
			} else if mode == "generated" {
				if sub, ok := m["generated_config"].(map[string]interface{}); ok {
					for sk, sv := range sub {
						switch sk {
						case "prompt":
							c["instructions"] = sv
						case "fallback_strategy":
							c[sk] = sv
						default:
							bad = append(bad, "fillerWords.content.generated_config."+sk)
						}
					}
				}
			} else {
				bad = append(bad, "fillerWords.content.mode")
			}
			for key := range m {
				if key != "mode" && key != "static_config" && key != "generated_config" {
					bad = append(bad, "fillerWords.content."+key)
				}
			}
			out["content"] = c
		default:
			bad = append(bad, "fillerWords."+k)
		}
	}
	return out, bad
}

func convertModule(name string, input map[string]interface{}) (map[string]any, []string) {
	allowed := map[string]bool{"vendor": true, "params": true, "url": true}
	for _, key := range map[string][]string{"asr": {"language", "model"}, "llm": {"style", "model", "input_modalities", "output_modalities", "max_history", "failure_message", "mcp_servers", "tools", "headers"}, "tts": {"headers", "skip_patterns"}, "mllm": {"model", "input_modalities", "output_modalities", "mcp_servers", "headers"}}[name] {
		allowed[key] = true
	}
	consumed := map[string]bool{"api_key": true, "system_messages": true, "messages": true, "greeting_message": true, "greeting_audio_url": true, "template_variables": true, "greeting_configs": true, "turn_detection": true, "enable": true, "_minimax_preset_model": true}
	out := map[string]any{}
	bad := []string{}
	for k, v := range input {
		if allowed[k] {
			out[k] = agentcore.CloneValue(v)
		} else if !consumed[k] && v != nil {
			bad = append(bad, name+"."+k)
		}
	}
	params, _ := out["params"].(map[string]interface{})
	if params == nil {
		params = map[string]interface{}{}
	}
	key, _ := input["api_key"].(string)
	for _, k := range []string{"api_key", "key", "api_subscription_key", "app_key", "spatius_api_key"} {
		if s, ok := params[k].(string); ok && s != "" {
			if key != "" && key != s {
				bad = append(bad, name+".params."+k+" (ambiguous credential)")
			} else {
				key = s
			}
			delete(params, k)
		}
	}
	for _, k := range []string{"credentials", "access_key_id", "secret_access_key", "aws_access_key_id", "aws_secret_access_key", "access_key", "secret_key", "adc_credentials_string"} {
		if v, ok := params[k]; ok && v != nil && v != "" {
			bad = append(bad, name+".params."+k)
		}
	}
	for _, k := range []string{"url", "base_url"} {
		if _, ok := out["url"]; !ok {
			if v, ok := params[k]; ok {
				out["url"] = v
			}
		}
		delete(params, k)
	}
	if name == "llm" || name == "mllm" || name == "asr" {
		if _, ok := out["model"]; !ok {
			if v, ok := params["model"]; ok {
				out["model"] = v
			}
		}
		delete(params, "model")
	}
	if name == "tts" {
		if v, ok := input["_minimax_preset_model"]; ok {
			params["model"] = v
		}
	}
	if name == "avatar" {
		vendor, _ := input["vendor"].(string)
		switch vendor {
		case "akool", "liveavatar", "anam", "sensetime", "spatius", "generic":
		default:
			bad = append(bad, name+".vendor")
		}
		if v, ok := input["enable"]; ok {
			out["enabled"] = v
		} else {
			out["enabled"] = true
		}
	} else if v, ok := input["enable"].(bool); ok && !v {
		bad = append(bad, name+".enable")
	}
	if name == "tts" {
		if v, ok := out["skip_patterns"]; ok && !emptyValue(v) {
			bad = append(bad, name+".skip_patterns")
		}
	}
	if len(params) > 0 {
		out["params"] = params
	} else {
		delete(out, "params")
	}
	if key != "" {
		out["credential"] = map[string]any{"mode": "byok", "api_key": key}
	} else {
		out["credential"] = map[string]any{"mode": "managed"}
		delete(out, "url")
	}
	return out, bad
}

func convertTurnDetection(td *TurnDetectionConfig, interruption *InterruptionConfig, mode string) (map[string]any, []string) {
	m, err := agentcore.StructToMap(td)
	if err != nil {
		return nil, []string{"turnDetection"}
	}
	return convertTurnMap(m, interruption, mode)
}
func convertTurnMap(m map[string]interface{}, interruption *InterruptionConfig, mode string) (map[string]any, []string) {
	bad := []string{}
	out := map[string]any{}
	if mode == "realtime" {
		for k, v := range m {
			switch k {
			case "type":
				out[k] = v
			case "mode":
				if _, ok := out["type"]; !ok {
					out["type"] = v
				}
			case "threshold", "prefix_padding_ms", "silence_duration_ms", "eagerness", "create_response", "interrupt_response":
				out[k] = v
			case "agora_vad_config", "server_vad_config", "semantic_vad_config":
				if nested, ok := v.(map[string]interface{}); ok {
					for nk, nv := range nested {
						out[nk] = nv
					}
				} else {
					bad = append(bad, "turnDetection."+k)
				}
			case "language":
				if v != nil {
					bad = append(bad, "turnDetection.language in realtime pipeline")
				}
			default:
				if v != nil {
					bad = append(bad, "turnDetection."+k)
				}
			}
		}
		if interruption != nil {
			bad = append(bad, "interruption")
		}
		return out, bad
	}
	if threshold, ok := m["threshold"]; ok {
		bad = append(bad, "turnDetection.threshold")
		_ = threshold
	}
	if config, ok := m["config"].(map[string]interface{}); ok {
		if threshold, ok := config["speech_threshold"]; ok {
			out["speech_threshold"] = threshold
		}
		for _, key := range []string{"start_of_speech", "end_of_speech"} {
			if value, ok := config[key].(map[string]interface{}); ok {
				mode, _ := value["mode"].(string)
				if mode == "keywords" || mode == "disabled" {
					out["interruption"] = convertInterruption(value)
					continue
				}
				part := map[string]any{}
				part["type"] = mode
				if nested, ok := value[mode+"_config"].(map[string]interface{}); ok {
					for nk, nv := range nested {
						part[nk] = nv
					}
				}
				for k, v := range value {
					if k != "mode" && k != mode+"_config" && v != nil {
						bad = append(bad, "turnDetection.config."+key+"."+k)
					}
				}
				if wait, ok := part["max_wait_ms"].(float64); ok && wait < 500 {
					bad = append(bad, "turnDetection.config."+key+".max_wait_ms")
				}
				out[key] = part
			}
		}
		for k, v := range config {
			if k != "speech_threshold" && k != "start_of_speech" && k != "end_of_speech" && v != nil {
				bad = append(bad, "turnDetection.config."+k)
			}
		}
	}
	for k, v := range m {
		if k != "config" && k != "language" && !(k == "mode" && v == "default") && v != nil {
			bad = append(bad, "turnDetection."+k)
		}
	}
	if interruption != nil {
		im, _ := agentcore.StructToMap(interruption)
		out["interruption"] = convertInterruption(im)
	}
	return out, bad
}
func convertInterruption(m map[string]interface{}) map[string]any {
	mode, _ := m["mode"].(string)
	if enable, ok := m["enable"].(bool); ok && !enable || mode == "disabled" {
		out := map[string]any{"mode": "off"}
		if c, ok := m["disabled_config"].(map[string]interface{}); ok {
			out["when_off"] = c["strategy"]
		}
		return out
	}
	if mode == "keywords" {
		out := map[string]any{"mode": "keywords"}
		if c, ok := m["keywords_config"].(map[string]interface{}); ok {
			if v, ok := c["trigger_keywords"]; ok {
				out["keywords"] = v
			} else {
				out["keywords"] = c["triggered_keywords"]
			}
		}
		return out
	}
	return map[string]any{"mode": "start_of_speech"}
}

func deployPatch(old v3.Agent, new v3.AgentCreate) (v3.AgentPatch, error) {
	b, err := json.Marshal(new)
	if err != nil {
		return nil, err
	}
	var fresh map[string]any
	if err = json.Unmarshal(b, &fresh); err != nil {
		return nil, err
	}
	oldMap := map[string]any{}
	for k, v := range map[string]any{"agent_name": old.AgentName, "instructions": old.Instructions, "messages": old.Messages, "greeting": old.Greeting, "variables": old.Variables, "pipeline": old.Pipeline, "filler_words": old.FillerWords} {
		if raw, ok := v.(json.RawMessage); ok {
			if len(raw) > 0 {
				var decoded any
				if json.Unmarshal(raw, &decoded) == nil {
					oldMap[k] = decoded
				}
			}
		} else if v != nil {
			encoded, _ := json.Marshal(v)
			var decoded any
			if json.Unmarshal(encoded, &decoded) == nil {
				oldMap[k] = decoded
			}
		}
	}
	return clearMissing(oldMap, fresh, ""), nil
}

func clearMissing(old, new map[string]any, path string) map[string]any {
	allowed := map[string]map[string]bool{
		"":         {"agent_name": true, "instructions": true, "messages": true, "greeting": true, "variables": true, "pipeline": true, "filler_words": true},
		"pipeline": {"mode": true, "asr": true, "llm": true, "tts": true, "mllm": true, "avatar": true, "turn_detection": true, "silence_config": true},
	}
	moduleAllowed := map[string]bool{"credential": true, "vendor": true, "model": true, "style": true, "language": true, "enabled": true, "params": true, "url": true, "headers": true, "input_modalities": true, "output_modalities": true, "max_history": true, "failure_message": true, "mcp_servers": true, "tools": true, "skip_patterns": true}
	for k, ov := range old {
		if path == "" && !allowed[""][k] {
			continue
		}
		if path == "pipeline" && !allowed["pipeline"][k] {
			continue
		}
		isModule := path == "pipeline.asr" || path == "pipeline.llm" || path == "pipeline.tts" || path == "pipeline.mllm" || path == "pipeline.avatar"
		if isModule && !moduleAllowed[k] {
			continue
		}
		if isModule && k == "credential" {
			continue
		} // resolved credentials can contain server-owned metadata
		if nv, ok := new[k]; !ok {
			new[k] = nil
		} else if nm, ok := nv.(map[string]any); ok {
			if om, ok := ov.(map[string]any); ok {
				new[k] = clearMissing(om, nm, joinPath(path, k))
			}
		}
	}
	if path == "pipeline.asr" || path == "pipeline.llm" || path == "pipeline.tts" || path == "pipeline.mllm" || path == "pipeline.avatar" {
		if credential, ok := new["credential"].(map[string]any); ok && credential["mode"] == "managed" {
			credential["api_key"] = nil
		}
	}
	return new
}
func joinPath(a, b string) string {
	if a == "" {
		return b
	}
	return a + "." + b
}
