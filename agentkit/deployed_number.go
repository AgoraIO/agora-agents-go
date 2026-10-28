package agentkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/AgoraIO/agora-agents-go/v2/v3"
)

type AssignNumberOptions struct {
	CallPolicy             *InboundCallPolicy
	MaxCallDurationSeconds *int
	MaxSilenceDurationMs   *int
}

// ListNumbers scans this project's number pages for inbound assignments to d.ID.
// The v3 list endpoint has no agent filter, so the scan cost grows with the
// project's total number count. It does not make per-number GET requests.
func (d *SavedAgent) ListNumbers(ctx context.Context) ([]v3.Number, error) {
	if d == nil || d.client == nil || d.ID == "" {
		return nil, errors.New("agentkit: saved agent with ID and client is required")
	}
	out := []v3.Number{}
	cursor := ""
	seen := map[string]bool{}
	for {
		page, err := d.client.V3.Numbers.List(ctx, v3.ListOptions{Limit: 100, Cursor: cursor})
		if err != nil {
			return nil, err
		}
		for _, number := range page.Body.Data.List {
			if len(number.Inbound) == 0 || string(number.Inbound) == "null" {
				continue
			}
			var inbound struct {
				Agent *string `json:"agent"`
			}
			if err := json.Unmarshal(number.Inbound, &inbound); err != nil {
				return nil, fmt.Errorf("agentkit: invalid inbound value for number %s: %w", number.NumberID, err)
			}
			if inbound.Agent != nil && *inbound.Agent == d.ID {
				out = append(out, number)
			}
		}
		if !page.Body.HasMore {
			return out, nil
		}
		if page.Body.NextCursor == nil || *page.Body.NextCursor == "" || seen[*page.Body.NextCursor] {
			return nil, errors.New("agentkit: number pagination ended without a usable next cursor")
		}
		cursor = *page.Body.NextCursor
		seen[cursor] = true
	}
}

// AssignNumber resolves either a number ID or its E.164 value in this project,
// then binds its inbound agent with If-Match. A number must already exist.
func (d *SavedAgent) AssignNumber(ctx context.Context, number any, opts AssignNumberOptions) (*v3.Number, error) {
	policy := map[string]any{}
	if opts.CallPolicy != nil {
		encoded, err := json.Marshal(opts.CallPolicy)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(encoded, &policy); err != nil {
			return nil, err
		}
		if opts.CallPolicy.MaxCallDurationSeconds != nil && *opts.CallPolicy.MaxCallDurationSeconds < 1 {
			return nil, errors.New("agentkit: max call duration must be positive")
		}
		if opts.CallPolicy.MaxSilenceDurationMS != nil && *opts.CallPolicy.MaxSilenceDurationMS < 1 {
			return nil, errors.New("agentkit: max silence duration must be positive")
		}
	}
	if opts.MaxCallDurationSeconds != nil {
		if *opts.MaxCallDurationSeconds < 1 {
			return nil, errors.New("agentkit: max call duration must be positive")
		}
		policy["max_call_duration_seconds"] = *opts.MaxCallDurationSeconds
	}
	if opts.MaxSilenceDurationMs != nil {
		if *opts.MaxSilenceDurationMs < 1 {
			return nil, errors.New("agentkit: max silence duration must be positive")
		}
		policy["max_silence_duration_ms"] = *opts.MaxSilenceDurationMs
	}
	return d.setNumber(ctx, number, d.ID, policy)
}

func (d *SavedAgent) UnassignNumber(ctx context.Context, number any) (*v3.Number, error) {
	return d.setNumber(ctx, number, nil, nil)
}

func (d *SavedAgent) setNumber(ctx context.Context, number any, agent any, policy map[string]any) (*v3.Number, error) {
	current, err := d.resolveNumber(ctx, number)
	if err != nil {
		return nil, err
	}
	for attempt := 0; attempt < 2; attempt++ {
		inbound := map[string]any{"agent": agent}
		if len(policy) > 0 {
			inbound["call_policy"] = policy
		}
		r, err := d.client.V3.Numbers.Update(ctx, current.Body.NumberID, current.ETag, v3.NumberPatch{"inbound": inbound})
		if err == nil {
			return &r.Body, nil
		}
		var conflict *v3.PreconditionFailedError
		if !errors.As(err, &conflict) || attempt == 1 {
			return nil, translatePrecondition(err)
		}
		current, err = d.client.V3.Numbers.Get(ctx, current.Body.NumberID)
		if err != nil {
			return nil, err
		}
	}
	return nil, errors.New("agentkit: number update failed")
}

func (d *SavedAgent) resolveNumber(ctx context.Context, number any) (*v3.Response[v3.Number], error) {
	var input string
	switch n := number.(type) {
	case string:
		input = n
	case v3.Number:
		input = n.NumberID
	case *v3.Number:
		if n != nil {
			input = n.NumberID
		}
	default:
		return nil, fmt.Errorf("%w: invalid number type", ErrNumberNotFound)
	}
	if strings.TrimSpace(input) == "" {
		return nil, fmt.Errorf("%w: empty number", ErrNumberNotFound)
	}
	// The API's canonical identifier is number_id. E.164 values require a scan:
	// GET /numbers has no number filter in the staging schema.
	if !strings.HasPrefix(input, "+") {
		got, err := d.client.V3.Numbers.Get(ctx, input)
		if err == nil {
			return got, nil
		}
		return nil, err
	}
	cursor := ""
	seen := map[string]bool{}
	for {
		page, err := d.client.V3.Numbers.List(ctx, v3.ListOptions{Limit: 100, Cursor: cursor})
		if err != nil {
			return nil, err
		}
		for _, item := range page.Body.Data.List {
			if normalizeNumber(item.Number) == normalizeNumber(input) {
				return d.client.V3.Numbers.Get(ctx, item.NumberID)
			}
		}
		if !page.Body.HasMore || page.Body.NextCursor == nil || *page.Body.NextCursor == "" || seen[*page.Body.NextCursor] {
			break
		}
		cursor = *page.Body.NextCursor
		seen[cursor] = true
	}
	return nil, fmt.Errorf("%w: %s in project %s", ErrNumberNotFound, input, d.client.AppID())
}
func normalizeNumber(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || strings.ContainsRune("().-", r) {
			return -1
		}
		return r
	}, s)
}
