package agentkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sync"
	"time"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/agentmanagement"
	"github.com/AgoraIO/agora-agents-go/v2/agents"
	"github.com/AgoraIO/agora-agents-go/v2/core"
	"github.com/AgoraIO/agora-agents-go/v2/option"
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

type AgentSession struct {
	client          *agents.Client
	agentManagement *agentmanagement.Client
	agent           *Agent
	appID           string
	appCertificate  string
	name            string
	channel         string
	token           string
	agentUID        string
	remoteUIDs      []string
	idleTimeout     *int
	enableStringUID *bool
	expiresIn       int  // Token lifetime in seconds (0 = use DefaultExpirySeconds)
	useAppCredsREST bool // When true, generate ConvoAI token per request for REST API auth
	preset          []string
	pipelineID      string
	debug           bool
	warn            func(string)

	agentID  string
	status   AgentSessionLifecycle
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

type AgentSessionOptions struct {
	Client                *agents.Client
	AgentManagementClient *agentmanagement.Client
	Agent                 *Agent
	AppID                 string
	AppCertificate        string
	Name                  string
	Channel               string
	Token                 string
	AgentUID              string
	RemoteUIDs            []string
	IdleTimeout           *int
	EnableStringUID       *bool
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

func NewAgentSession(opts AgentSessionOptions) *AgentSession {
	name := opts.Name
	if name == "" {
		name = fmt.Sprintf("agent-%d", time.Now().UnixMilli())
	}

	return &AgentSession{
		client:          opts.Client,
		agentManagement: opts.AgentManagementClient,
		agent:           opts.Agent,
		appID:           opts.AppID,
		appCertificate:  opts.AppCertificate,
		name:            name,
		channel:         opts.Channel,
		token:           opts.Token,
		agentUID:        opts.AgentUID,
		remoteUIDs:      opts.RemoteUIDs,
		idleTimeout:     opts.IdleTimeout,
		enableStringUID: opts.EnableStringUID,
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

// convoAIRequestOpts returns per-request options with ConvoAI token when using app credentials.
func (s *AgentSession) convoAIRequestOpts(ctx context.Context) ([]option.RequestOption, error) {
	if !s.useAppCredsREST {
		return nil, nil
	}
	if s.appCertificate == "" {
		return nil, fmt.Errorf("appCertificate is required for app-credentials auth mode; pass AppCertificate when creating AgoraClient")
	}
	uid, err := parseNumericUID(s.agentUID, "agent UID")
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
	return []option.RequestOption{option.WithToken(token)}, nil
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

func (s *AgentSession) Agent() *Agent {
	return s.agent
}

func (s *AgentSession) AppID() string {
	return s.appID
}

func (s *AgentSession) Raw() *agents.Client {
	return s.client
}

func (s *AgentSession) RawAgentManagement() *agentmanagement.Client {
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
	if s.agent == nil || s.agent.avatar == nil {
		return nil
	}
	if !avatarConfigEnabled(s.agent.avatar) {
		return nil
	}

	vendor, _ := s.agent.avatar["vendor"].(string)
	if vendor == "" {
		return nil
	}
	params, _ := s.agent.avatar["params"].(map[string]interface{})
	if err := ValidateAvatarConfig(vendor, params); err != nil {
		return err
	}

	if s.agent.ttsSampleRate != nil {
		if err := ValidateTtsSampleRate(vendor, int(*s.agent.ttsSampleRate)); err != nil {
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
	s.mu.Lock()
	if s.status != StatusIdle && s.status != StatusStopped && s.status != StatusError {
		s.mu.Unlock()
		return "", fmt.Errorf("cannot start session in %s state", s.status)
	}
	if s.agent != nil && s.agent.mllm != nil && s.agent.hasEnabledAvatar() {
		s.mu.Unlock()
		return "", fmt.Errorf("avatar is only supported with cascading ASR/LLM/TTS sessions; remove the avatar configuration when using MLLM")
	}

	if s.agent.avatar != nil && avatarConfigEnabled(s.agent.avatar) && s.agent.avatarRequiredSampleRate != nil && *s.agent.avatarRequiredSampleRate != 0 && s.agent.ttsSampleRate != nil {
		if *s.agent.ttsSampleRate != *s.agent.avatarRequiredSampleRate {
			s.mu.Unlock()
			return "", fmt.Errorf(
				"avatar requires TTS sample rate of %d Hz, but TTS is configured with %d Hz",
				int(*s.agent.avatarRequiredSampleRate), int(*s.agent.ttsSampleRate),
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
		pipelineID = s.agent.PipelineID()
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

	properties, err := s.agent.ToPropertiesMap(propOpts)
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

	if s.debug {
		debugPayload := map[string]interface{}{
			"name":       s.name,
			"properties": resolvedProperties,
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
	if len(avatar) == 0 || !avatarConfigEnabled(avatar) {
		return nil
	}
	vendor, _ := avatar["vendor"].(string)
	params := asMap(avatar["params"])
	if err := ValidateAvatarConfig(vendor, params); err != nil {
		return err
	}
	if IsGenericAvatar(vendor) {
		if !hasNonEmptyString(params, "agora_appid") {
			return fmt.Errorf("Generic avatar requires agora_appid")
		}
		if !hasNonEmptyString(params, "agora_channel") {
			return fmt.Errorf("Generic avatar requires agora_channel")
		}
		if !hasNonEmptyString(params, "agora_token") {
			return fmt.Errorf("Generic avatar requires agora_token")
		}
		return nil
	}
	if IsAvatarTokenManaged(vendor) && avatarUIDString(params["agora_uid"]) != "" && !hasNonEmptyString(params, "agora_token") {
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

	if preset, ok := inferASRPreset(s.agent.stt); ok && preset != "" {
		skip["asr"] = true
	}
	if preset, ok := inferLLMPreset(s.agent.llm); ok && preset != "" {
		skip["llm"] = true
	}
	if preset, ok := inferTTSPreset(s.agent.tts); ok && preset != "" {
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

func (s *AgentSession) Update(ctx context.Context, properties *Agora.UpdateAgentsRequestProperties) error {
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

func (s *AgentSession) GetHistory(ctx context.Context) (*Agora.GetHistoryAgentsResponse, error) {
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
