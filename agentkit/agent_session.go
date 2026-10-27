package agentkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	agentcore "github.com/AgoraIO/agora-agents-go/v2/agentkit/core"
	"github.com/AgoraIO/agora-agents-go/v2/agentmanagement"
	"github.com/AgoraIO/agora-agents-go/v2/agents"
	"github.com/AgoraIO/agora-agents-go/v2/core"
	"github.com/AgoraIO/agora-agents-go/v2/option"
	"github.com/AgoraIO/agora-agents-go/v2/v3"
)

// AgentSessionLifecycle is the local client lifecycle state for AgentSession.
type AgentSessionLifecycle string

const (
	AgentSessionLifecycleIdle     AgentSessionLifecycle = "idle"
	AgentSessionLifecycleStarting AgentSessionLifecycle = "starting"
	AgentSessionLifecycleRunning  AgentSessionLifecycle = "running"
	AgentSessionLifecycleStopping AgentSessionLifecycle = "stopping"
	AgentSessionLifecycleStopped  AgentSessionLifecycle = "stopped"
	AgentSessionLifecycleError    AgentSessionLifecycle = "error"
)

const (
	// Deprecated: Use AgentSessionLifecycleIdle.
	StatusIdle = AgentSessionLifecycleIdle
	// Deprecated: Use AgentSessionLifecycleStarting.
	StatusStarting = AgentSessionLifecycleStarting
	// Deprecated: Use AgentSessionLifecycleRunning.
	StatusRunning = AgentSessionLifecycleRunning
	// Deprecated: Use AgentSessionLifecycleStopping.
	StatusStopping = AgentSessionLifecycleStopping
	// Deprecated: Use AgentSessionLifecycleStopped.
	StatusStopped = AgentSessionLifecycleStopped
	// Deprecated: Use AgentSessionLifecycleError.
	StatusError = AgentSessionLifecycleError
)

type EventHandler func(data interface{})

type sessionBackend uint8

const (
	sessionBackendV2 sessionBackend = iota
	sessionBackendV3Saved
	sessionBackendV3Ephemeral
)

// ErrUnsupportedSessionOperation identifies lifecycle actions that the
// current v3 session contract does not expose.
var ErrUnsupportedSessionOperation = errors.New("agentkit: operation is not supported for v3 sessions")

type AgentSession struct {
	client          *agents.Client
	agentManagement *agentmanagement.Client
	httpClient      core.HTTPClient
	routingOpts     []option.RequestOption
	agent           agentcore.AgentRuntime
	appID           string
	appCertificate  string
	name            string
	channel         string
	token           string
	agentUID        string
	remoteUIDs      []string
	idleTimeout     *int
	maxDuration     *int
	labels          map[string]string
	values          map[string]any
	idempotencyKey  string
	encryption      *v3.RTCEncryption
	geofence        *v3.Geofence
	audioScenario   string
	gracefulStop    *GracefulStopOptions
	dataPolicy      *v3.DataPolicy
	enableStringUID *bool
	expiresIn       int  // Token lifetime in seconds (0 = use DefaultExpirySeconds)
	useAppCredsREST bool // When true, generate ConvoAI token per request for REST API auth
	preset          []string
	pipelineID      string
	debug           bool
	warn            func(string)
	backend         sessionBackend
	v3Sessions      *v3.SessionsClient
	savedAgentID    string
	initErr         error

	agentID        string
	status         AgentSessionLifecycle
	startSucceeded bool
	v3Revision     uint64
	v3ReadSequence uint64
	v3WritePending bool
	mu             sync.RWMutex
	handlers       map[string][]EventHandler
}

type AgentSessionOptions struct {
	Client                *agents.Client
	HTTPClient            core.HTTPClient
	AgentManagementClient *agentmanagement.Client
	Agent                 agentcore.AgentRuntime
	AppID                 string
	AppCertificate        string
	// Name is the agent instance identifier sent as the top-level /join "name" field.
	// When empty, NewAgentSession generates agent-<unix_timestamp>.
	Name            string
	ClientReference string
	Channel         string
	Token           string
	AgentUID        string
	RemoteUIDs      []string
	IdleTimeout     *int
	MaxDuration     *int
	Labels          map[string]string
	Values          map[string]any
	IdempotencyKey  string
	Encryption      *v3.RTCEncryption
	Geofence        *v3.Geofence
	AudioScenario   string
	GracefulStop    *GracefulStopOptions
	DataPolicy      *v3.DataPolicy
	EnableStringUID *bool
	// ExpiresIn is the token lifetime in seconds (default: 86400 = 24 hours, Agora maximum).
	// Only applies when the SDK auto-generates a token. Valid range: 1–86400.
	// Use ExpiresInHours() / ExpiresInMinutes() for clarity.
	ExpiresIn int
	// UseAppCredentialsForREST when true, generates a ConvoAI token per request for REST API
	// authentication. Use when the client was created without Basic Auth or token (app-credentials mode).
	UseAppCredentialsForREST bool
	Preset                   []string
	PipelineID               string
	Debug                    bool
	Warn                     func(string)
}

// NewAgentSession creates a session in StatusIdle.
// Prefer Agent.CreateSession for the public AgentKit flow; this constructor is for advanced wiring.
func NewAgentSession(opts AgentSessionOptions) *AgentSession {
	name := opts.Name
	if name == "" {
		name = fmt.Sprintf("agent-%d", time.Now().UnixMilli())
	}

	return &AgentSession{
		client:          opts.Client,
		agentManagement: opts.AgentManagementClient,
		httpClient:      opts.HTTPClient,
		agent:           opts.Agent,
		appID:           opts.AppID,
		appCertificate:  opts.AppCertificate,
		name:            name,
		channel:         opts.Channel,
		token:           opts.Token,
		agentUID:        opts.AgentUID,
		remoteUIDs:      append([]string(nil), opts.RemoteUIDs...),
		idleTimeout:     cloneIntPointer(opts.IdleTimeout),
		maxDuration:     cloneIntPointer(opts.MaxDuration),
		labels:          cloneStringMap(opts.Labels),
		values:          cloneAnyMap(opts.Values),
		idempotencyKey:  opts.IdempotencyKey,
		encryption:      cloneRTCEncryption(opts.Encryption),
		geofence:        cloneGeofence(opts.Geofence),
		audioScenario:   opts.AudioScenario,
		gracefulStop:    cloneGracefulStop(opts.GracefulStop),
		dataPolicy:      cloneDataPolicy(opts.DataPolicy),
		enableStringUID: cloneBoolPointer(opts.EnableStringUID),
		expiresIn:       opts.ExpiresIn,
		useAppCredsREST: opts.UseAppCredentialsForREST,
		preset:          append([]string(nil), opts.Preset...),
		pipelineID:      opts.PipelineID,
		debug:           opts.Debug,
		warn:            opts.Warn,
		status:          StatusIdle,
		handlers:        make(map[string][]EventHandler),
	}
}

func newV3AgentSession(agent agentcore.AgentRuntime, client *AgoraClient, opts CreateSessionOptions, savedAgentID string) *AgentSession {
	if agent == nil {
		agent = NewAgent(client)
	}
	session := newLegacyAgentSession(agent, opts)
	session.backend = sessionBackendV3Ephemeral
	if savedAgentID != "" {
		session.backend = sessionBackendV3Saved
	}
	session.savedAgentID = savedAgentID
	if config, ok := agent.(*Agent); ok && config != nil && config.base != nil {
		base := config.base
		if session.labels == nil {
			session.labels = cloneStringMap(base.Labels)
		}
		if session.audioScenario == "" && base.AudioScenario != nil {
			value := strings.ToLower(string(*base.AudioScenario))
			if value == "default" {
				value = "aiserver"
			}
			session.audioScenario = value
		}
		if session.geofence == nil && base.Geofence != nil {
			geofence := &v3.Geofence{Area: strings.ToLower(string(base.Geofence.Area))}
			if base.Geofence.ExcludeArea != nil {
				geofence.ExcludeArea = strings.ToLower(string(*base.Geofence.ExcludeArea))
			}
			session.geofence = geofence
		}
		if base.Parameters != nil {
			if session.gracefulStop == nil && base.Parameters.FarewellConfig != nil {
				legacy := base.Parameters.FarewellConfig
				graceful := &GracefulStopOptions{}
				if legacy.GracefulEnabled != nil {
					graceful.Enabled = *legacy.GracefulEnabled
				}
				if legacy.GracefulTimeoutSeconds != nil {
					milliseconds := *legacy.GracefulTimeoutSeconds * 1000
					graceful.TimeoutMS = &milliseconds
				}
				session.gracefulStop = graceful
			}
			if session.dataPolicy == nil && base.Parameters.OptOut != nil {
				retention := "30_days"
				if *base.Parameters.OptOut {
					retention = "none"
				}
				session.dataPolicy = &v3.DataPolicy{Retention: retention}
			}
		}
	}
	if client == nil || client.V3 == nil || client.V3.Sessions == nil {
		session.initErr = errors.New("agentkit: v3 sessions client is required")
		return session
	}
	session.v3Sessions = client.V3.Sessions
	if opts.ClientReference != "" {
		session.initErr = errors.New("agentkit: ClientReference was removed from the v3 session contract")
	}
	return session
}

func cloneStringMap(source map[string]string) map[string]string {
	if source == nil {
		return nil
	}
	cloned := make(map[string]string, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func cloneIntPointer(source *int) *int {
	if source == nil {
		return nil
	}
	cloned := *source
	return &cloned
}

func cloneBoolPointer(source *bool) *bool {
	if source == nil {
		return nil
	}
	cloned := *source
	return &cloned
}

func cloneAnyMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func cloneRTCEncryption(source *v3.RTCEncryption) *v3.RTCEncryption {
	if source == nil {
		return nil
	}
	cloned := *source
	return &cloned
}

func cloneGracefulStop(source *GracefulStopOptions) *GracefulStopOptions {
	if source == nil {
		return nil
	}
	cloned := *source
	cloned.TimeoutMS = cloneIntPointer(source.TimeoutMS)
	return &cloned
}

func cloneDataPolicy(source *v3.DataPolicy) *v3.DataPolicy {
	if source == nil {
		return nil
	}
	cloned := *source
	return &cloned
}

func cloneGeofence(source *v3.Geofence) *v3.Geofence {
	if source == nil {
		return nil
	}
	cloned := *source
	return &cloned
}

func (s *AgentSession) v3StartRequestParts() (v3.SessionTransport, *v3.SessionLifecycle, *v3.DataPolicy, error) {
	if s.initErr != nil {
		return v3.SessionTransport{}, nil, nil, s.initErr
	}
	if strings.TrimSpace(s.channel) == "" {
		return v3.SessionTransport{}, nil, nil, errors.New("agentkit: v3 session channel is required")
	}
	if s.expiresIn != 0 {
		return v3.SessionTransport{}, nil, nil, errors.New("agentkit: v3 sessions do not support ExpiresIn")
	}
	if len(s.preset) > 0 {
		return v3.SessionTransport{}, nil, nil, errors.New("agentkit: v3 sessions do not support presets")
	}
	if s.pipelineID != "" {
		return v3.SessionTransport{}, nil, nil, errors.New("agentkit: v3 sessions do not support PipelineID overrides")
	}
	transport := v3.SessionTransport{
		Type: "rtc", Channel: s.channel, Token: s.token, UID: s.agentUID,
		SubscribeUIDs: append([]string(nil), s.remoteUIDs...),
		Encryption:    cloneRTCEncryption(s.encryption), AudioScenario: s.audioScenario,
		Geofence: cloneGeofence(s.geofence),
	}
	var lifecycle *v3.SessionLifecycle
	if s.idleTimeout != nil || s.maxDuration != nil || s.gracefulStop != nil {
		lifecycle = &v3.SessionLifecycle{}
		if s.idleTimeout != nil {
			if *s.idleTimeout < 0 || *s.idleTimeout > 3600 {
				return v3.SessionTransport{}, nil, nil, errors.New("agentkit: IdleTimeout must be between 0 and 3600 seconds")
			}
			milliseconds := *s.idleTimeout * 1000
			lifecycle.IdleTimeoutMS = &milliseconds
		}
		if s.maxDuration != nil {
			if *s.maxDuration < 1 || *s.maxDuration > 3600 {
				return v3.SessionTransport{}, nil, nil, errors.New("agentkit: MaxDuration must be between 1 and 3600 seconds")
			}
			milliseconds := *s.maxDuration * 1000
			lifecycle.MaxDurationMS = &milliseconds
		}
		if s.gracefulStop != nil {
			lifecycle.GracefulStop = cloneGracefulStop(s.gracefulStop)
		}
	}
	if s.audioScenario != "" && s.audioScenario != "aiserver" && s.audioScenario != "chorus" {
		return v3.SessionTransport{}, nil, nil, errors.New("agentkit: AudioScenario must be aiserver or chorus")
	}
	return transport, lifecycle, cloneDataPolicy(s.dataPolicy), nil
}

func (s *AgentSession) startV3(ctx context.Context) (string, error) {
	s.mu.Lock()
	if s.idempotencyKey != "" && s.startSucceeded {
		s.mu.Unlock()
		return "", errors.New("agentkit: a session handle cannot reuse its IdempotencyKey after a successful start; create a new handle with a new key")
	}
	if s.status != StatusIdle && s.status != StatusStopped && s.status != StatusError {
		status := s.status
		s.mu.Unlock()
		return "", fmt.Errorf("cannot start session in %s state", status)
	}
	s.status = StatusStarting
	s.agentID = ""
	s.v3Revision++
	s.v3WritePending = true
	s.mu.Unlock()
	transport, lifecycle, dataPolicy, err := s.v3StartRequestParts()
	if err != nil {
		s.setV3StartFailure(err)
		return "", err
	}
	if s.v3Sessions == nil {
		err = errors.New("agentkit: v3 sessions client is required")
		s.setV3StartFailure(err)
		return "", err
	}
	var response *v3.Response[v3.Session]
	if s.backend == sessionBackendV3Saved {
		response, err = s.v3Sessions.Start(ctx, v3.SessionStartRequest{
			AgentID: s.savedAgentID, Values: cloneAnyMap(s.values), Labels: cloneStringMap(s.labels),
			Transport: transport, Lifecycle: lifecycle, DataPolicy: dataPolicy,
		}, v3.SessionWriteOptions{IdempotencyKey: s.idempotencyKey})
	} else {
		config, ok := s.agent.(*Agent)
		if !ok || config == nil {
			err = errors.New("agentkit: ephemeral v3 sessions require an Agent config")
		} else {
			storedShape, mapErr := config.toV3Session()
			if mapErr != nil {
				err = mapErr
			} else {
				response, err = s.v3Sessions.StartEphemeral(ctx, v3.SessionStartEphemeralRequest{
					Labels: cloneStringMap(s.labels), Values: cloneAnyMap(s.values),
					Agent: v3.EphemeralAgent{
						Instructions: storedShape.Instructions,
						Messages:     storedShape.Messages,
						Greeting:     storedShape.Greeting,
						Variables:    storedShape.Variables,
						Pipeline:     storedShape.Pipeline,
						FillerWords:  storedShape.FillerWords,
					},
					Transport: transport, Lifecycle: lifecycle, DataPolicy: dataPolicy,
				}, v3.SessionWriteOptions{IdempotencyKey: s.idempotencyKey})
			}
		}
	}
	if err != nil {
		s.setV3StartFailure(err)
		return "", err
	}
	if response == nil || response.Body.AgentSessionID == "" {
		err = errors.New("agentkit: v3 session response has no agent_session_id")
		s.setV3StartFailure(err)
		return "", err
	}
	s.mu.Lock()
	s.agentID = response.Body.AgentSessionID
	s.status = lifecycleStatus(response.Body.Status, StatusRunning)
	s.startSucceeded = true
	s.v3WritePending = false
	s.v3Revision++
	s.mu.Unlock()
	s.emit("started", map[string]string{"agent_session_id": response.Body.AgentSessionID})
	return response.Body.AgentSessionID, nil
}

func (s *AgentSession) setV3StartFailure(err error) {
	s.mu.Lock()
	s.status = StatusError
	s.v3WritePending = false
	s.v3Revision++
	s.mu.Unlock()
	s.emit("error", err)
}

func lifecycleStatus(status string, fallback AgentSessionLifecycle) AgentSessionLifecycle {
	switch strings.ToLower(status) {
	case "idle":
		return StatusIdle
	case "starting":
		return StatusStarting
	case "running":
		return StatusRunning
	case "stopping":
		return StatusStopping
	case "stopped":
		return StatusStopped
	case "failed", "error":
		return StatusError
	default:
		return fallback
	}
}

func (s *AgentSession) requireV2Operation(name string) error {
	if s.backend != sessionBackendV2 {
		return fmt.Errorf("%w: %s", ErrUnsupportedSessionOperation, name)
	}
	return nil
}

// convoAIRequestOpts returns per-request options with ConvoAI token when using app credentials.
func (s *AgentSession) convoAIRequestOpts(ctx context.Context) ([]option.RequestOption, error) {
	s.mu.RLock()
	routingOpts := append([]option.RequestOption(nil), s.routingOpts...)
	s.mu.RUnlock()
	if !s.useAppCredsREST {
		return routingOpts, nil
	}
	if s.appCertificate == "" {
		return nil, fmt.Errorf("appCertificate is required for app-credentials auth mode; pass AppCertificate when creating AgoraClient")
	}
	uid, err := agentcore.ParseNumericUID(s.agentUID, "agent UID")
	if err != nil {
		return nil, err
	}
	token, err := GenerateConvoAIToken(GenerateConvoAITokenOptions{
		AppID:          s.appID,
		AppCertificate: s.appCertificate,
		ChannelName:    s.channel,
		UID:            uid,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate ConvoAI token for app-credentials auth mode: %w", err)
	}
	return append([]option.RequestOption{option.WithToken(token)}, routingOpts...), nil
}

func (s *AgentSession) ID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.agentID
}

func (s *AgentSession) Status() AgentSessionLifecycle {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *AgentSession) Agent() agentcore.AgentRuntime {
	return s.agent
}

func (s *AgentSession) AppID() string {
	return s.appID
}

func (s *AgentSession) Raw() *agents.Client {
	if s.backend != sessionBackendV2 {
		return nil
	}
	return s.client
}

func (s *AgentSession) RawAgentManagement() *agentmanagement.Client {
	if s.backend != sessionBackendV2 {
		return nil
	}
	return s.agentManagement
}

func (s *AgentSession) warnf(message string) {
	if s.warn != nil {
		s.warn(message)
		return
	}
	log.Printf("agentkit: %s", message)
}

func (s *AgentSession) validateAvatarConfig() error {
	baseAgent := s.agent.BaseAgent()
	if s.agent == nil || baseAgent.Avatar == nil {
		return nil
	}
	if !agentcore.AvatarConfigEnabled(baseAgent.Avatar) {
		return nil
	}

	vendor, _ := baseAgent.Avatar["vendor"].(string)
	if vendor == "" {
		return nil
	}
	params, _ := baseAgent.Avatar["params"].(map[string]interface{})
	if err := agentcore.ValidateAvatarConfig(vendor, params); err != nil {
		return err
	}

	if baseAgent.TTSSampleRate != nil {
		if err := agentcore.ValidateTtsSampleRate(vendor, int(*baseAgent.TTSSampleRate)); err != nil {
			return err
		}
		return nil
	}

	switch vendor {
	case "heygen":
		s.warnf("Warning: HeyGen avatar detected but TTS sample_rate is not explicitly set. HeyGen requires 24,000 Hz. Please ensure your TTS provider is configured for 24kHz.")
	case "liveavatar":
		s.warnf("Warning: LiveAvatar avatar detected but TTS sample_rate is not explicitly set. LiveAvatar requires 24,000 Hz. Please ensure your TTS provider is configured for 24kHz.")
	case "akool":
		s.warnf("Warning: Akool avatar detected but TTS sample_rate is not explicitly set. Akool requires 16,000 Hz. Please ensure your TTS provider is configured for 16kHz.")
	}
	return nil
}

func (s *AgentSession) Start(ctx context.Context) (string, error) {
	if s.backend != sessionBackendV2 {
		return s.startV3(ctx)
	}
	s.mu.Lock()
	if s.status != StatusIdle && s.status != StatusStopped && s.status != StatusError {
		s.mu.Unlock()
		return "", fmt.Errorf("cannot start session in %s state", s.status)
	}
	baseAgent := s.agent.BaseAgent()
	if s.agent != nil && baseAgent.MLLM != nil && baseAgent.Avatar != nil && agentcore.AvatarConfigEnabled(baseAgent.Avatar) {
		s.mu.Unlock()
		return "", fmt.Errorf("avatar is only supported with cascading ASR/LLM/TTS sessions; remove the avatar configuration when using MLLM")
	}

	if baseAgent.Avatar != nil && agentcore.AvatarConfigEnabled(baseAgent.Avatar) && baseAgent.AvatarRequiredSampleRate != nil && *baseAgent.AvatarRequiredSampleRate != 0 && baseAgent.TTSSampleRate != nil {
		if *baseAgent.TTSSampleRate != *baseAgent.AvatarRequiredSampleRate {
			s.mu.Unlock()
			return "", fmt.Errorf(
				"avatar requires TTS sample rate of %d Hz, but TTS is configured with %d Hz",
				int(*baseAgent.AvatarRequiredSampleRate), int(*baseAgent.TTSSampleRate),
			)
		}
	}
	if err := s.validateAvatarConfig(); err != nil {
		s.mu.Unlock()
		return "", err
	}

	s.status = StatusStarting
	s.mu.Unlock()

	pipelineID := s.pipelineID
	if pipelineID == "" {
		pipelineID = baseAgent.PipelineID
	}

	skipCategories, allowMissingCategories := s.vendorValidationCategories(pipelineID)
	propOpts := ToPropertiesOptions{
		Channel:                        s.channel,
		AgentUID:                       s.agentUID,
		RemoteUIDs:                     s.remoteUIDs,
		Token:                          s.token,
		AppID:                          s.appID,
		AppCertificate:                 s.appCertificate,
		ExpiresIn:                      s.expiresIn,
		IdleTimeout:                    s.idleTimeout,
		EnableStringUID:                s.enableStringUID,
		SkipVendorValidationCategories: skipCategories,
		AllowMissingVendorCategories:   allowMissingCategories,
		Warn:                           s.warnf,
	}

	properties, err := agentcore.BuildPropertiesMap(s.agent.Profile(), baseAgent, propOpts, func(opts agentcore.GenerateConvoAITokenOptions) (string, error) {
		return GenerateConvoAIToken(GenerateConvoAITokenOptions(opts))
	})
	if err != nil {
		s.mu.Lock()
		s.status = StatusError
		s.mu.Unlock()
		s.emit("error", err)
		return "", err
	}

	resolvedPreset, resolvedProperties, err := ResolveSessionPresetsMap(s.preset, properties)
	if err != nil {
		s.mu.Lock()
		s.status = StatusError
		s.mu.Unlock()
		s.emit("error", err)
		return "", err
	}
	if err := validateEnrichedAvatarConfig(resolvedProperties); err != nil {
		s.mu.Lock()
		s.status = StatusError
		s.mu.Unlock()
		s.emit("error", err)
		return "", err
	}
	ApplyPreviewShape(resolvedProperties)

	previewFeatures := RequiredPreviewFeatures(resolvedProperties)
	s.mu.Lock()
	s.routingOpts = previewRequestOptions(previewFeatures, s.httpClient, s.debug)
	s.mu.Unlock()

	if s.debug {
		debugPayload := map[string]interface{}{
			"name":       s.name,
			"properties": RedactSecrets(resolvedProperties),
		}
		if resolvedPreset != "" {
			debugPayload["preset"] = resolvedPreset
		}
		if pipelineID != "" {
			debugPayload["pipeline_id"] = pipelineID
		}
		if payload, err := json.Marshal(debugPayload); err == nil {
			log.Printf("[Agora Debug] Starting agent session: %s", payload)
		} else {
			s.warnf(fmt.Sprintf("debug logging failed to marshal start request: %v", err))
		}
	}

	reqOpts, err := s.convoAIRequestOpts(ctx)
	if err != nil {
		s.mu.Lock()
		s.status = StatusError
		s.mu.Unlock()
		s.emit("error", err)
		return "", err
	}
	resp, err := startAgentsWithMapBody(ctx, s.client, s.appID, s.name, resolvedPreset, pipelineID, resolvedProperties, reqOpts...)
	if err != nil {
		s.mu.Lock()
		s.status = StatusError
		s.mu.Unlock()
		s.emit("error", err)
		return "", err
	}

	s.mu.Lock()
	if resp != nil && resp.AgentID != nil {
		s.agentID = *resp.AgentID
	}
	s.status = StatusRunning
	s.mu.Unlock()

	s.emit("started", map[string]string{"agent_id": s.agentID})
	return s.agentID, nil
}

func validateEnrichedAvatarConfig(properties map[string]interface{}) error {
	if properties == nil {
		return nil
	}
	avatar := asMap(properties["avatar"])
	if len(avatar) == 0 || !agentcore.AvatarConfigEnabled(avatar) {
		return nil
	}
	vendor, _ := avatar["vendor"].(string)
	params := agentcore.AsMap(avatar["params"])
	if err := agentcore.ValidateAvatarConfig(vendor, params); err != nil {
		return err
	}
	if agentcore.IsGenericAvatar(vendor) {
		if !agentcore.HasNonEmptyString(params, "agora_appid") {
			return fmt.Errorf("Generic avatar requires agora_appid")
		}
		if !agentcore.HasNonEmptyString(params, "agora_channel") {
			return fmt.Errorf("Generic avatar requires agora_channel")
		}
		if !agentcore.HasNonEmptyString(params, "agora_token") {
			return fmt.Errorf("Generic avatar requires agora_token")
		}
		return nil
	}
	if agentcore.IsAvatarTokenManaged(vendor) && agentcore.ParseUIDString(params["agora_uid"]) != "" && !agentcore.HasNonEmptyString(params, "agora_token") {
		return fmt.Errorf("%s avatar requires agora_token; pass AgoraToken on the avatar vendor or provide AppCertificate for automatic token generation", vendor)
	}
	return nil
}

func (s *AgentSession) vendorValidationCategories(pipelineID string) ([]string, []string) {
	skip := map[string]bool{}
	allowMissing := map[string]bool{}

	if pipelineID != "" {
		for _, category := range []string{"asr", "llm", "tts"} {
			allowMissing[category] = true
		}
	}

	for _, preset := range parsePresetInput(s.preset) {
		if category := getPresetCategory(preset); category != "" {
			skip[category] = true
			allowMissing[category] = true
		}
	}

	baseAgent := s.agent.BaseAgent()
	if preset, ok := inferASRPreset(baseAgent.STT); ok && preset != "" {
		skip["asr"] = true
	}
	if preset, ok := inferLLMPreset(baseAgent.LLM); ok && preset != "" {
		skip["llm"] = true
	}
	if preset, ok := inferTTSPreset(baseAgent.TTS); ok && preset != "" {
		skip["tts"] = true
	}

	return categorySetToSlice(skip), categorySetToSlice(allowMissing)
}

func categorySetToSlice(values map[string]bool) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, category := range []string{"asr", "llm", "tts"} {
		if values[category] {
			result = append(result, category)
		}
	}
	return result
}

func (s *AgentSession) Stop(ctx context.Context) error {
	if s.backend != sessionBackendV2 {
		return s.stopV3(ctx)
	}
	s.mu.Lock()
	if s.status != StatusRunning {
		s.mu.Unlock()
		return fmt.Errorf("cannot stop session in %s state", s.status)
	}
	if s.agentID == "" {
		s.mu.Unlock()
		return fmt.Errorf("no agent ID available")
	}
	s.status = StatusStopping
	s.mu.Unlock()

	reqOpts, err := s.convoAIRequestOpts(ctx)
	if err != nil {
		s.mu.Lock()
		s.status = StatusError
		s.mu.Unlock()
		s.emit("error", err)
		return err
	}
	err = s.client.Stop(ctx, &Agora.StopAgentsRequest{
		Appid:   s.appID,
		AgentID: s.agentID,
	}, reqOpts...)
	if err != nil {
		// Handle 404 "task not found" gracefully — agent is already stopped
		var apiErr *core.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			s.mu.Lock()
			s.status = StatusStopped
			s.mu.Unlock()
			s.emit("stopped", map[string]string{"agent_id": s.agentID})
			return nil
		}
		s.mu.Lock()
		s.status = StatusError
		s.mu.Unlock()
		s.emit("error", err)
		return err
	}

	s.mu.Lock()
	s.status = StatusStopped
	s.mu.Unlock()
	s.emit("stopped", map[string]string{"agent_id": s.agentID})
	return nil
}

func (s *AgentSession) stopV3(ctx context.Context) error {
	s.mu.Lock()
	if s.status != StatusRunning && s.status != StatusStarting {
		status := s.status
		s.mu.Unlock()
		return fmt.Errorf("cannot stop session in %s state", status)
	}

	if s.agentID == "" {
		s.mu.Unlock()
		return errors.New("no agent_session_id available")
	}
	s.status = StatusStopping
	s.v3Revision++
	s.v3WritePending = true
	revision := s.v3Revision
	id := s.agentID
	s.mu.Unlock()
	if s.v3Sessions == nil {
		err := errors.New("agentkit: v3 sessions client is required")
		s.mu.Lock()
		s.status = StatusError
		s.v3WritePending = false
		s.v3Revision++
		s.mu.Unlock()
		return err
	}
	_, err := s.v3Sessions.Stop(ctx, id)
	if err != nil {
		s.mu.Lock()
		if s.agentID == id && s.v3Revision == revision {
			s.status = sessionStatusAfterStopError(err)
			s.v3WritePending = false
			s.v3Revision++
		}
		s.mu.Unlock()
		s.emit("error", err)
		return err
	}
	s.mu.Lock()
	if s.agentID == id && s.v3Revision == revision {
		s.status = StatusStopped
		s.v3WritePending = false
		s.v3Revision++
	}
	s.mu.Unlock()
	s.emit("stopped", map[string]string{"agent_session_id": id})
	return nil
}

func sessionStatusAfterStopError(err error) AgentSessionLifecycle {
	var apiErr *core.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusConflict || apiErr.Unwrap() == nil {
		return StatusError
	}
	var payload struct {
		Reason string `json:"reason"`
	}
	if json.Unmarshal([]byte(apiErr.Unwrap().Error()), &payload) != nil {
		return StatusError
	}
	switch payload.Reason {
	case "TaskNotRunning":
		return StatusStopped
	case "TaskConflict":
		return StatusStopping
	default:
		return StatusError
	}
}

func (s *AgentSession) Say(ctx context.Context, text string, priority *Agora.SpeakAgentsRequestPriority, interruptable *bool) error {
	return s.SayWithOptions(ctx, text, &SayOptions{
		Priority:      priority,
		Interruptable: interruptable,
	})
}

type SayOptions struct {
	Priority      *Agora.SpeakAgentsRequestPriority
	Interruptable *bool
}

func (s *AgentSession) SayWithOptions(ctx context.Context, text string, opts *SayOptions) error {
	if err := s.requireV2Operation("Say"); err != nil {
		return err
	}
	s.mu.RLock()
	if s.status != StatusRunning {
		s.mu.RUnlock()
		return fmt.Errorf("cannot say in %s state", s.status)
	}
	if s.agentID == "" {
		s.mu.RUnlock()
		return fmt.Errorf("no agent ID available")
	}
	s.mu.RUnlock()

	req := &Agora.SpeakAgentsRequest{
		Appid:   s.appID,
		AgentID: s.agentID,
		Text:    text,
	}
	if opts != nil {
		req.Priority = opts.Priority
		req.Interruptable = opts.Interruptable
	}

	reqOpts, err := s.convoAIRequestOpts(ctx)
	if err != nil {
		return err
	}
	_, err = s.client.Speak(ctx, req, reqOpts...)
	return err
}

func (s *AgentSession) Interrupt(ctx context.Context) error {
	if err := s.requireV2Operation("Interrupt"); err != nil {
		return err
	}
	s.mu.RLock()
	if s.status != StatusRunning {
		s.mu.RUnlock()
		return fmt.Errorf("cannot interrupt in %s state", s.status)
	}
	if s.agentID == "" {
		s.mu.RUnlock()
		return fmt.Errorf("no agent ID available")
	}
	s.mu.RUnlock()

	reqOpts, err := s.convoAIRequestOpts(ctx)
	if err != nil {
		return err
	}
	_, err = s.client.Interrupt(ctx, &Agora.InterruptAgentsRequest{
		Appid:   s.appID,
		AgentID: s.agentID,
	}, reqOpts...)
	return err
}

func (s *AgentSession) Think(
	ctx context.Context,
	text string,
	onListeningAction *Agora.AgentThinkAgentManagementRequestOnListeningAction,
	onThinkingAction *Agora.AgentThinkAgentManagementRequestOnThinkingAction,
	onSpeakingAction *Agora.AgentThinkAgentManagementRequestOnSpeakingAction,
	interruptable *bool,
	metadata map[string]string,
) (*Agora.AgentThinkAgentManagementResponse, error) {
	return s.ThinkWithOptions(ctx, text, &ThinkOptions{
		OnListeningAction: onListeningAction,
		OnThinkingAction:  onThinkingAction,
		OnSpeakingAction:  onSpeakingAction,
		Interruptable:     interruptable,
		Metadata:          metadata,
	})
}

type ThinkOptions struct {
	OnListeningAction *Agora.AgentThinkAgentManagementRequestOnListeningAction
	OnThinkingAction  *Agora.AgentThinkAgentManagementRequestOnThinkingAction
	OnSpeakingAction  *Agora.AgentThinkAgentManagementRequestOnSpeakingAction
	Interruptable     *bool
	Metadata          map[string]string
}

func (s *AgentSession) ThinkWithOptions(
	ctx context.Context,
	text string,
	opts *ThinkOptions,
) (*Agora.AgentThinkAgentManagementResponse, error) {
	if err := s.requireV2Operation("Think"); err != nil {
		return nil, err
	}
	s.mu.RLock()
	if s.status != StatusRunning {
		s.mu.RUnlock()
		return nil, fmt.Errorf("cannot think in %s state", s.status)
	}
	if s.agentID == "" {
		s.mu.RUnlock()
		return nil, fmt.Errorf("no agent ID available")
	}
	s.mu.RUnlock()
	if s.agentManagement == nil {
		return nil, fmt.Errorf("agent management client is not configured")
	}

	req := &Agora.AgentThinkAgentManagementRequest{
		Appid:   s.appID,
		AgentID: s.agentID,
		Text:    text,
	}
	if opts != nil {
		req.OnListeningAction = opts.OnListeningAction
		req.OnThinkingAction = opts.OnThinkingAction
		req.OnSpeakingAction = opts.OnSpeakingAction
		req.Interruptable = opts.Interruptable
		req.Metadata = opts.Metadata
	}
	reqOpts, err := s.convoAIRequestOpts(ctx)
	if err != nil {
		return nil, err
	}
	return s.agentManagement.AgentThink(ctx, req, reqOpts...)
}

// Update applies a legacy properties update for v2 sessions or a v3.SessionPatch
// for v3 sessions. V3 patches only support token and message/LLM/MLLM overrides.
func (s *AgentSession) Update(ctx context.Context, properties *Agora.UpdateAgentsRequestProperties) error {
	if s.backend != sessionBackendV2 {
		if properties == nil {
			return errors.New("agentkit: session Update properties are required")
		}
		patch, err := sessionPatchFromLegacyProperties(properties)
		if err != nil {
			return err
		}
		s.mu.RLock()
		id, status := s.agentID, s.status
		s.mu.RUnlock()
		if status != StatusRunning || id == "" {
			return fmt.Errorf("cannot update v3 session in %s state", status)
		}
		if s.v3Sessions == nil {
			return errors.New("agentkit: v3 sessions client is required")
		}
		_, err = s.v3Sessions.Update(ctx, id, patch)
		return err
	}
	s.mu.RLock()
	if s.status != StatusRunning {
		s.mu.RUnlock()
		return fmt.Errorf("cannot update in %s state", s.status)
	}
	if s.agentID == "" {
		s.mu.RUnlock()
		return fmt.Errorf("no agent ID available")
	}
	s.mu.RUnlock()

	reqOpts, err := s.convoAIRequestOpts(ctx)
	if err != nil {
		return err
	}
	_, err = s.client.Update(ctx, &Agora.UpdateAgentsRequest{
		Appid:      s.appID,
		AgentID:    s.agentID,
		Properties: properties,
	}, reqOpts...)
	return err
}

func sessionPatchFromLegacyProperties(properties *Agora.UpdateAgentsRequestProperties) (v3.SessionPatch, error) {
	var patch v3.SessionPatch
	if len(properties.GetExtraProperties()) > 0 {
		return patch, fmt.Errorf("agentkit: v3 session Update does not support extra properties")
	}
	if token := properties.GetToken(); token != nil {
		patch.Transport = &v3.SessionTransportPatch{Token: *token}
	}
	if llm := properties.GetLlm(); llm != nil {
		if len(llm.GetExtraProperties()) > 0 {
			return patch, fmt.Errorf("agentkit: v3 session Update does not support extra LLM properties")
		}
		overrides := &v3.SessionAgentOverridesPatch{}
		if messages := llm.GetSystemMessages(); messages != nil {
			converted := make([]v3.Message, 0, len(messages))
			for i, message := range messages {
				for key := range message {
					if key != "role" && key != "content" {
						return patch, fmt.Errorf("agentkit: v3 session Update message %d has unsupported field %q", i, key)
					}
				}
				role, roleOK := message["role"].(string)
				content, contentOK := message["content"].(string)
				if !roleOK || !contentOK || (role != "system" && role != "user" && role != "assistant") {
					return patch, fmt.Errorf("agentkit: v3 session Update message %d requires a supported role and string content", i)
				}
				converted = append(converted, v3.Message{Role: role, Content: content})
			}
			overrides.Messages = &converted
		}
		if params := llm.GetParams(); params != nil {
			overrides.Pipeline = &v3.SessionPipelinePatch{LLM: &v3.SessionModulePatch{Params: params}}
		}
		if overrides.Messages != nil || overrides.Pipeline != nil {
			patch.Agent = &v3.SessionAgentPatch{Overrides: overrides}
		}
	}
	if mllm := properties.GetMllm(); mllm != nil {
		if len(mllm.GetExtraProperties()) > 0 {
			return patch, fmt.Errorf("agentkit: v3 session Update does not support extra MLLM properties")
		}
		if patch.Agent != nil && patch.Agent.Overrides != nil && patch.Agent.Overrides.Pipeline != nil {
			return patch, fmt.Errorf("agentkit: session Update cannot combine LLM and MLLM overrides")
		}
		params := mllm.GetParams()
		if params != nil {
			overrides := &v3.SessionAgentOverridesPatch{Pipeline: &v3.SessionPipelinePatch{MLLM: &v3.SessionModulePatch{Params: params}}}
			if patch.Agent != nil && patch.Agent.Overrides != nil {
				overrides.Messages = patch.Agent.Overrides.Messages
			}
			patch.Agent = &v3.SessionAgentPatch{Overrides: overrides}
		}
	}
	return patch, nil
}

func (s *AgentSession) GetHistory(ctx context.Context) (*Agora.GetHistoryAgentsResponse, error) {
	if err := s.requireV2Operation("GetHistory"); err != nil {
		return nil, err
	}
	s.mu.RLock()
	if s.agentID == "" {
		s.mu.RUnlock()
		return nil, fmt.Errorf("no agent ID available")
	}
	s.mu.RUnlock()

	reqOpts, err := s.convoAIRequestOpts(ctx)
	if err != nil {
		return nil, err
	}
	return s.client.GetHistory(ctx, &Agora.GetHistoryAgentsRequest{
		Appid:   s.appID,
		AgentID: s.agentID,
	}, reqOpts...)
}

func (s *AgentSession) GetInfo(ctx context.Context) (*Agora.GetAgentsResponse, error) {
	if s.backend != sessionBackendV2 {
		s.mu.Lock()
		id := s.agentID
		revision := s.v3Revision
		s.v3ReadSequence++
		sequence := s.v3ReadSequence
		s.mu.Unlock()
		if id == "" {
			return nil, errors.New("no agent_session_id available")
		}
		if s.v3Sessions == nil {
			return nil, errors.New("agentkit: v3 sessions client is required")
		}
		response, err := s.v3Sessions.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		encoded, err := json.Marshal(response.Body)
		if err != nil {
			return nil, err
		}
		body := &Agora.GetAgentsResponse{}
		if err := json.Unmarshal(encoded, body); err != nil {
			return nil, err
		}
		s.mu.Lock()
		if s.agentID == id && s.v3Revision == revision && s.v3ReadSequence == sequence && !s.v3WritePending {
			s.status = lifecycleStatus(response.Body.Status, s.status)
		}
		s.mu.Unlock()
		return body, nil
	}
	s.mu.RLock()
	if s.agentID == "" {
		s.mu.RUnlock()
		return nil, fmt.Errorf("no agent ID available")
	}
	s.mu.RUnlock()

	reqOpts, err := s.convoAIRequestOpts(ctx)
	if err != nil {
		return nil, err
	}
	return s.client.Get(ctx, &Agora.GetAgentsRequest{
		Appid:   s.appID,
		AgentID: s.agentID,
	}, reqOpts...)
}

type GetTurnsOptions struct {
	PageIndex *int
	PageSize  *int
}

type GetAllTurnsOptions struct {
	PageSize *int
}

func (s *AgentSession) GetTurns(ctx context.Context, opts ...GetTurnsOptions) (*Agora.GetTurnsAgentsResponse, error) {
	if err := s.requireV2Operation("GetTurns"); err != nil {
		return nil, err
	}
	s.mu.RLock()
	if s.agentID == "" {
		s.mu.RUnlock()
		return nil, fmt.Errorf("no agent ID available")
	}
	s.mu.RUnlock()

	req := &Agora.GetTurnsAgentsRequest{
		Appid:   s.appID,
		AgentID: s.agentID,
	}
	if len(opts) > 0 {
		req.PageIndex = opts[0].PageIndex
		req.PageSize = opts[0].PageSize
	}
	reqOpts, err := s.convoAIRequestOpts(ctx)
	if err != nil {
		return nil, err
	}
	return s.client.GetTurns(ctx, req, reqOpts...)
}

func (s *AgentSession) GetAllTurns(ctx context.Context, opts ...GetAllTurnsOptions) (*Agora.GetTurnsAgentsResponse, error) {
	pageIndex := 1
	pageSize := 50
	if len(opts) > 0 && opts[0].PageSize != nil {
		pageSize = *opts[0].PageSize
	}
	all := []*Agora.GetTurnsAgentsResponseTurnsItem{}
	var latest *Agora.GetTurnsAgentsResponse
	currentPage := 0
	for {
		resp, err := s.GetTurns(ctx, GetTurnsOptions{
			PageIndex: &pageIndex,
			PageSize:  &pageSize,
		})
		if err != nil {
			return nil, err
		}
		if resp == nil {
			return &Agora.GetTurnsAgentsResponse{Turns: all}, nil
		}
		latest = resp
		all = append(all, resp.Turns...)
		if resp.Pagination != nil {
			if resp.Pagination.IsLastPage != nil && *resp.Pagination.IsLastPage {
				latest.Turns = all
				return latest, nil
			}
			if resp.Pagination.TotalPages != nil && pageIndex >= *resp.Pagination.TotalPages {
				latest.Turns = all
				return latest, nil
			}
			if resp.Pagination.PageIndex != nil {
				if *resp.Pagination.PageIndex <= currentPage {
					return nil, fmt.Errorf("getAllTurns pagination did not advance: requested page %d, received page %d", pageIndex, *resp.Pagination.PageIndex)
				}
				currentPage = *resp.Pagination.PageIndex
			} else {
				currentPage = pageIndex
			}
		} else {
			latest.Turns = all
			return latest, nil
		}
		pageIndex++
	}
}

func (s *AgentSession) On(event string, handler EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[event] = append(s.handlers[event], handler)
}

func (s *AgentSession) Off(event string, handler EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	handlers := s.handlers[event]
	if len(handlers) == 0 {
		return
	}

	target := reflect.ValueOf(handler).Pointer()
	filtered := handlers[:0]
	for _, registered := range handlers {
		if reflect.ValueOf(registered).Pointer() == target {
			continue
		}
		filtered = append(filtered, registered)
	}
	if len(filtered) == 0 {
		delete(s.handlers, event)
		return
	}
	s.handlers[event] = filtered
}

func (s *AgentSession) emit(event string, data interface{}) {
	s.mu.RLock()
	handlers := s.handlers[event]
	s.mu.RUnlock()

	for _, h := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Log and continue so a panicking handler does not prevent
					// remaining handlers or session lifecycle from completing.
					s.warnf(fmt.Sprintf("recovered panic in %q event handler: %v", event, r))
				}
			}()
			h(data)
		}()
	}
}

func stringPtrOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
