package v3

import (
	"context"
	"errors"
	"net/http"
)

// CampaignCreate contains the supported inline campaign creation contract.
// Contacts are immutable after creation. Source numbers must belong to the project.
type CampaignCreate struct {
	Name      string             `json:"name"`
	AgentID   string             `json:"agent_id"`
	Transport CampaignTransport  `json:"transport"`
	Contacts  []CampaignContact  `json:"contacts"`
	Schedule  CampaignSchedule   `json:"schedule"`
	Lifecycle *CampaignLifecycle `json:"lifecycle,omitempty"`
	Pacing    *CampaignPacing    `json:"pacing,omitempty"`
}
type CampaignContact struct {
	Phone  string         `json:"phone"`
	Values map[string]any `json:"values,omitempty"`
}

// From uses the canonical array form accepted by the API, including one number.
type CampaignTransport struct {
	Type       string              `json:"type"`
	From       []string            `json:"from"`
	CallPolicy *CampaignCallPolicy `json:"call_policy,omitempty"`
}
type CampaignCallPolicy struct {
	MaxRingDurationMS      *int                     `json:"max_ring_duration_ms,omitempty"`
	MaxCallDurationSeconds *int                     `json:"max_call_duration_seconds,omitempty"`
	MaxSilenceDurationMS   *int                     `json:"max_silence_duration_ms,omitempty"`
	Voicemail              *CampaignVoicemailPolicy `json:"voicemail,omitempty"`
	EndCall                *CampaignEndCallPolicy   `json:"end_call,omitempty"`
	Transfer               *CampaignTransferPolicy  `json:"transfer,omitempty"`
}
type CampaignVoicemailPolicy struct {
	Mode string `json:"mode"`
}
type CampaignEndCallPolicy struct {
	OnConversationComplete bool `json:"on_conversation_complete,omitempty"`
	OnUserRequest          bool `json:"on_user_request,omitempty"`
	OnFax                  bool `json:"on_fax,omitempty"`
	OnAIAssistant          bool `json:"on_ai_assistant,omitempty"`
}
type CampaignTransferPolicy struct {
	PhoneNumber string `json:"phone_number"`
	Description string `json:"description"`
}
type CampaignLifecycle struct {
	IdleTimeoutMS *int          `json:"idle_timeout_ms,omitempty"`
	MaxDurationMS *int          `json:"max_duration_ms,omitempty"`
	GracefulStop  *GracefulStop `json:"graceful_stop,omitempty"`
}

// Retry fields exist in the supplied schema but their runtime behavior is open.
// The SDK implements no campaign retry or dialing algorithm.
type CampaignPacing struct {
	MaxConcurrent     *int `json:"max_concurrent,omitempty"`
	MaxCallsPerSecond *int `json:"max_calls_per_second,omitempty"`
	MaxAttempts       *int `json:"max_attempts,omitempty"`
	RetryBackoffMS    *int `json:"retry_backoff_ms,omitempty"`
}

// Mode is immediate or scheduled. Scheduled requires timezone, start and end.
// Times are local to the IANA timezone; start/end use YYYY-MM-DDTHH:MM.
type CampaignSchedule struct {
	Mode     string               `json:"mode"`
	Timezone string               `json:"timezone,omitempty"`
	Start    string               `json:"start,omitempty"`
	End      string               `json:"end,omitempty"`
	Days     []CampaignCallingDay `json:"days,omitempty"`
}
type CampaignCallingDay struct {
	Weekday string              `json:"weekday"`
	Ranges  []CampaignTimeRange `json:"ranges"`
}
type CampaignTimeRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// CampaignPatch preserves omission, explicit null and empty objects.
// Allowed keys are lifecycle, pacing, schedule and transport (call_policy only).
// Pacing merges; schedule and call policy replace. After dialing starts only
// pacing is editable. Server validation errors are preserved unchanged.
type CampaignPatch map[string]any
type Campaign struct {
	CampaignID string            `json:"campaign_id"`
	Name       string            `json:"name"`
	Status     string            `json:"status"`
	AgentID    string            `json:"agent_id"`
	Transport  CampaignTransport `json:"transport"`
	Lifecycle  CampaignLifecycle `json:"lifecycle"`
	Schedule   CampaignSchedule  `json:"schedule"`
	Counts     CampaignCounts    `json:"counts"`
	CreateTS   int64             `json:"create_ts"`
	StartTS    *int64            `json:"start_ts"`
	CompleteTS *int64            `json:"complete_ts"`
	Failure    *CampaignFailure  `json:"failure"`
}
type CampaignCounts struct {
	Contacts  int `json:"contacts"`
	Pending   int `json:"pending"`
	Calling   int `json:"calling"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
	Canceled  int `json:"canceled"`
}
type CampaignFailure struct {
	Code     string  `json:"code"`
	Provider *string `json:"provider,omitempty"`
	Message  string  `json:"message"`
	TS       int64   `json:"ts"`
}
type CampaignDeleteResponse struct {
	Deleted bool `json:"deleted"`
}
type CampaignListOptions struct {
	ListOptions
	Status string
}
type CampaignWriteOptions struct{ IdempotencyKey string }
type CampaignsClient struct{ t *transport }

func campaignWriteHeaders(opts []CampaignWriteOptions) (http.Header, error) {
	if len(opts) > 1 {
		return nil, errors.New("v3: at most one campaign write options value is accepted")
	}
	headers := http.Header{}
	if len(opts) == 1 && opts[0].IdempotencyKey != "" {
		headers.Set("Idempotency-Key", opts[0].IdempotencyKey)
	}
	return headers, nil
}

func (c *CampaignsClient) Create(ctx context.Context, body CampaignCreate, opts ...CampaignWriteOptions) (*Response[Campaign], error) {
	headers, err := campaignWriteHeaders(opts)
	if err != nil {
		return nil, err
	}
	p, err := c.t.path("campaigns")
	if err != nil {
		return nil, err
	}
	var out Campaign
	raw, err := c.t.call(ctx, http.MethodPost, p, nil, body, &out, headers, 1)
	if err != nil {
		return nil, err
	}
	return response(raw, out), nil
}
func (c *CampaignsClient) List(ctx context.Context, opts CampaignListOptions) (*Response[Page[Campaign]], error) {
	p, err := c.t.path("campaigns")
	if err != nil {
		return nil, err
	}
	query := listQuery(opts.ListOptions)
	if opts.Status != "" {
		query.Set("status", opts.Status)
	}
	var out Page[Campaign]
	raw, err := c.t.call(ctx, http.MethodGet, p, query, nil, &out, nil, 0)
	if err != nil {
		return nil, err
	}
	return response(raw, out), nil
}
func (c *CampaignsClient) Get(ctx context.Context, id string) (*Response[Campaign], error) {
	p, err := c.t.path("campaigns", id)
	if err != nil {
		return nil, err
	}
	var out Campaign
	raw, err := c.t.call(ctx, http.MethodGet, p, nil, nil, &out, nil, 0)
	if err != nil {
		return nil, err
	}
	return response(raw, out), nil
}
func (c *CampaignsClient) Update(ctx context.Context, id, ifMatch string, body CampaignPatch) (*Response[Campaign], error) {
	if err := requiredETag(ifMatch); err != nil {
		return nil, err
	}
	p, err := c.t.path("campaigns", id)
	if err != nil {
		return nil, err
	}
	var out Campaign
	raw, err := c.t.call(ctx, http.MethodPatch, p, nil, body, &out, etagHeader(ifMatch), 1)
	if err != nil {
		return nil, precondition(ctx, err, func(ctx context.Context) (any, string, error) {
			r, e := (&CampaignsClient{t: c.t.forLatestRead()}).Get(ctx, id)
			if e != nil {
				return nil, "", e
			}
			return r.Body, r.ETag, nil
		})
	}
	return response(raw, out), nil
}
func (c *CampaignsClient) Delete(ctx context.Context, id string) (*Response[CampaignDeleteResponse], error) {
	p, err := c.t.path("campaigns", id)
	if err != nil {
		return nil, err
	}
	var out CampaignDeleteResponse
	raw, err := c.t.call(ctx, http.MethodDelete, p, nil, nil, &out, nil, 1)
	if err != nil {
		return nil, err
	}
	return response(raw, out), nil
}
func (c *CampaignsClient) action(ctx context.Context, id, action string, opts []CampaignWriteOptions) (*Response[Campaign], error) {
	headers, err := campaignWriteHeaders(opts)
	if err != nil {
		return nil, err
	}
	p, err := c.t.path("campaigns", id)
	if err != nil {
		return nil, err
	}
	var out Campaign
	raw, err := c.t.call(ctx, http.MethodPost, p+":"+action, nil, nil, &out, headers, 1)
	if err != nil {
		return nil, err
	}
	return response(raw, out), nil
}
func (c *CampaignsClient) Pause(ctx context.Context, id string, opts ...CampaignWriteOptions) (*Response[Campaign], error) {
	return c.action(ctx, id, "pause", opts)
}
func (c *CampaignsClient) Resume(ctx context.Context, id string, opts ...CampaignWriteOptions) (*Response[Campaign], error) {
	return c.action(ctx, id, "resume", opts)
}
func (c *CampaignsClient) Cancel(ctx context.Context, id string, opts ...CampaignWriteOptions) (*Response[Campaign], error) {
	return c.action(ctx, id, "cancel", opts)
}
