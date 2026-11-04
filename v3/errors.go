package v3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/AgoraIO/agora-agents-go/v2/core"
)

// Category sentinels. Every error returned by a v3 operation matches exactly
// one of ErrValidation, ErrConfiguration, ErrConnection or ErrAPI with
// errors.Is. ErrTimeout additionally matches connection errors caused by a
// timeout, and ErrPreconditionFailed additionally matches 412 API errors.
//
//	switch {
//	case errors.Is(err, v3.ErrValidation):    // fix the input; nothing was sent
//	case errors.Is(err, v3.ErrConfiguration): // fix the client setup
//	case errors.Is(err, v3.ErrTimeout):       // retry with a longer deadline
//	case errors.Is(err, v3.ErrConnection):    // network failure; maybe retry
//	case errors.Is(err, v3.ErrAPI):           // inspect *v3.APIError
//	}
var (
	// ErrValidation matches input rejected locally before any HTTP request.
	ErrValidation = errors.New("v3: invalid request")
	// ErrConfiguration matches a misconfigured client, such as an invalid base
	// URL, a missing App ID, or missing or conflicting credentials.
	ErrConfiguration = errors.New("v3: invalid client configuration")
	// ErrConnection matches requests that could not complete: DNS, refused or
	// reset connections, TLS failures, cancellation, timeouts, and response
	// bodies that could not be read.
	ErrConnection = errors.New("v3: connection failed")
	// ErrTimeout matches the subset of connection errors caused by a timeout
	// or an expired context deadline.
	ErrTimeout = errors.New("v3: request timed out")
	// ErrAPI matches non-2xx responses and 2xx responses that violate the
	// API contract.
	ErrAPI = errors.New("v3: API error")
	// ErrPreconditionFailed matches 412 Precondition Failed API errors.
	ErrPreconditionFailed = errors.New("v3: precondition failed")
)

// Error is implemented by every categorized error returned by v3 operations
// and by AgentKit paths that call v3. Use errors.As with a v3.Error target to
// catch any of them.
type Error interface {
	error
	// Category returns the most specific category sentinel: ErrValidation,
	// ErrConfiguration, ErrConnection, ErrTimeout, ErrAPI or
	// ErrPreconditionFailed. Prefer errors.Is for checks, because a timeout is
	// also a connection error and a precondition failure is also an API error.
	Category() error
}

var (
	_ Error = (*ValidationError)(nil)
	_ Error = (*ConfigurationError)(nil)
	_ Error = (*ConnectionError)(nil)
	_ Error = (*APIError)(nil)
	_ Error = (*PreconditionFailedError)(nil)
)

// ValidationError reports input rejected locally before any HTTP request.
// It matches ErrValidation.
type ValidationError struct {
	// Message is the complete message, including its package prefix.
	Message string
	// Err is the underlying cause, if any.
	Err error
}

func (e *ValidationError) Error() string { return categorizedMessage(e.Message, ErrValidation, e.Err) }
func (e *ValidationError) Unwrap() error { return e.Err }

// Is reports whether target is ErrValidation.
func (e *ValidationError) Is(target error) bool { return target == ErrValidation }

// Category returns ErrValidation.
func (e *ValidationError) Category() error { return ErrValidation }

// ConfigurationError reports a misconfigured client, such as an invalid base
// URL, a missing App ID, or missing or conflicting credentials. It matches
// ErrConfiguration.
type ConfigurationError struct {
	// Message is the complete message, including its package prefix.
	Message string
	// Err is the underlying cause, if any.
	Err error
}

func (e *ConfigurationError) Error() string {
	return categorizedMessage(e.Message, ErrConfiguration, e.Err)
}
func (e *ConfigurationError) Unwrap() error { return e.Err }

// Is reports whether target is ErrConfiguration.
func (e *ConfigurationError) Is(target error) bool { return target == ErrConfiguration }

// Category returns ErrConfiguration.
func (e *ConfigurationError) Category() error { return ErrConfiguration }

func categorizedMessage(message string, sentinel, cause error) string {
	if message == "" {
		message = sentinel.Error()
	}
	if cause != nil {
		return message + ": " + cause.Error()
	}
	return message
}

// ConnectionError reports a request that could not complete. Err is the
// underlying transport error, so errors.As(err, &netErr) with a net.Error
// target and errors.Is(err, context.DeadlineExceeded) or
// errors.Is(err, context.Canceled) keep working. It matches ErrConnection,
// and also ErrTimeout when Timeout reports true.
type ConnectionError struct {
	Err error
}

func (e *ConnectionError) Error() string {
	prefix := "v3: request failed"
	if e.Timeout() {
		prefix = "v3: request timed out"
	}
	if e.Err == nil {
		return prefix
	}
	return prefix + ": " + e.Err.Error()
}
func (e *ConnectionError) Unwrap() error { return e.Err }

// Timeout reports whether the failure was caused by a timeout or an expired
// context deadline.
func (e *ConnectionError) Timeout() bool {
	if e == nil || e.Err == nil {
		return false
	}
	if errors.Is(e.Err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(e.Err, &netErr) && netErr.Timeout()
}

// Is reports whether target is ErrConnection, or ErrTimeout for timeouts.
func (e *ConnectionError) Is(target error) bool {
	return target == ErrConnection || (target == ErrTimeout && e.Timeout())
}

// Category returns ErrTimeout for timeouts and ErrConnection otherwise.
func (e *ConnectionError) Category() error {
	if e.Timeout() {
		return ErrTimeout
	}
	return ErrConnection
}

// APIError reports a non-2xx response, or a 2xx response that violates the API
// contract (an undecodable body, a missing ETag or ID, an unusable pagination
// cursor). For non-2xx responses it unwraps to the *core.APIError produced by
// the shared transport, so existing errors.As checks for *core.APIError keep
// matching; for contract violations Err describes the problem. It matches
// ErrAPI.
type APIError struct {
	// StatusCode is the HTTP status code.
	StatusCode int
	// Header holds the response headers.
	Header http.Header
	// Body is the raw response body.
	Body []byte
	// ParsedBody is the decoded JSON body, or nil when the body is not JSON.
	ParsedBody any
	// Detail is the body's "detail" string, when present.
	Detail string
	// Reason is the body's "reason" string, when present.
	Reason string
	// RequestID is the Agora-Request-Id response header, when present.
	RequestID string
	// Err is the underlying *core.APIError for non-2xx responses, or the
	// contract violation for 2xx responses. It may be nil.
	Err error
}

// ContractViolation returns an *APIError for a 2xx response that the SDK
// cannot use. It carries the response's status code, headers, body and
// request ID; cause describes the violation and stays reachable through
// Unwrap. AgentKit uses it for responses that lack required values.
func ContractViolation[T any](r *Response[T], cause error) *APIError {
	out := &APIError{Err: cause}
	if r == nil {
		return out
	}
	out.StatusCode = r.StatusCode
	out.Header = r.Header.Clone()
	out.RequestID = r.RequestID
	if out.RequestID == "" && out.Header != nil {
		out.RequestID = out.Header.Get("Agora-Request-Id")
	}
	out.Body = append([]byte(nil), r.raw...)
	out.parseBody()
	return out
}

func newAPIError(source *core.APIError) *APIError {
	out := &APIError{StatusCode: source.StatusCode, Header: source.Header.Clone(), Err: source}
	if out.Header != nil {
		out.RequestID = out.Header.Get("Agora-Request-Id")
	}
	if inner := source.Unwrap(); inner != nil {
		out.Body = []byte(inner.Error())
	}
	out.parseBody()
	return out
}

func (e *APIError) parseBody() {
	trimmed := bytes.TrimSpace(e.Body)
	if len(trimmed) == 0 || !json.Valid(trimmed) {
		return
	}
	var parsed any
	if json.Unmarshal(trimmed, &parsed) != nil {
		return
	}
	e.ParsedBody = parsed
	if fields, ok := parsed.(map[string]any); ok {
		e.Detail, _ = fields["detail"].(string)
		e.Reason, _ = fields["reason"].(string)
	}
}

func (e *APIError) Error() string { return "v3: API error: " + e.describe() }

// describe renders the status, reason, request ID and detail (or raw body).
func (e *APIError) describe() string {
	text := fmt.Sprintf("status %d", e.StatusCode)
	if e.Reason != "" {
		text += " " + e.Reason
	}
	if e.RequestID != "" {
		text += " (request ID " + e.RequestID + ")"
	}
	var source *core.APIError
	if e.Err != nil && !errors.As(e.Err, &source) {
		// A contract violation: the cause says what was wrong with the 2xx.
		return text + ": " + e.Err.Error()
	}
	if e.Detail != "" {
		return text + ": " + e.Detail
	}
	if body := bytes.TrimSpace(e.Body); len(body) > 0 {
		return text + ": " + string(body)
	}
	return text
}
func (e *APIError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Is reports whether target is ErrAPI.
func (e *APIError) Is(target error) bool { return target == ErrAPI }

// Category returns ErrAPI.
func (e *APIError) Category() error { return ErrAPI }

// PreconditionFailedError reports a 412 response. Cause is the *APIError, so
// errors.As matches *v3.APIError and *core.APIError through it. Conditional
// writes can reload the current resource with Latest. It matches
// ErrPreconditionFailed and ErrAPI.
type PreconditionFailedError struct {
	Cause error
	fetch func(context.Context) (any, string, error)
}

func (e *PreconditionFailedError) Error() string {
	var api *APIError
	if errors.As(e.Cause, &api) {
		return "v3: precondition failed: " + api.describe()
	}
	if e.Cause == nil {
		return ErrPreconditionFailed.Error()
	}
	return "v3: precondition failed: " + e.Cause.Error()
}
func (e *PreconditionFailedError) Unwrap() error { return e.Cause }

// Is reports whether target is ErrPreconditionFailed. ErrAPI matches through
// Cause.
func (e *PreconditionFailedError) Is(target error) bool { return target == ErrPreconditionFailed }

// Category returns ErrPreconditionFailed.
func (e *PreconditionFailedError) Category() error { return ErrPreconditionFailed }

// Latest performs a new GET for the resource the conditional write targeted
// and returns its body and ETag. It is unavailable for operations that are
// not ETag-guarded resource writes.
func (e *PreconditionFailedError) Latest(ctx context.Context) (any, string, error) {
	if e.fetch == nil {
		return nil, "", &ValidationError{Message: "v3: latest resource unavailable"}
	}
	return e.fetch(ctx)
}

// precondition attaches the latest-resource reload to a 412 from a
// conditional write.
func precondition(_ context.Context, err error, fetch func(context.Context) (any, string, error)) error {
	var stale *PreconditionFailedError
	if errors.As(err, &stale) && stale.fetch == nil {
		stale.fetch = fetch
	}
	return err
}

// classifyCallError maps an error from the shared caller to a v3 category.
// Local checks run before the caller, so anything else is a server response
// or a transport failure.
func classifyCallError(err error) error {
	var source *core.APIError
	if errors.As(err, &source) {
		api := newAPIError(source)
		if api.StatusCode == http.StatusPreconditionFailed {
			return &PreconditionFailedError{Cause: api}
		}
		return api
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr.Op == "parse" {
		return &ConfigurationError{Message: "v3: invalid request URL", Err: err}
	}
	var unsupportedType *json.UnsupportedTypeError
	var unsupportedValue *json.UnsupportedValueError
	var marshaler *json.MarshalerError
	if errors.As(err, &unsupportedType) || errors.As(err, &unsupportedValue) || errors.As(err, &marshaler) {
		return &ValidationError{Message: "v3: request body cannot be encoded as JSON", Err: err}
	}
	return &ConnectionError{Err: err}
}

func invalid(message string) error { return &ValidationError{Message: message} }

func invalidf(format string, args ...any) error {
	return &ValidationError{Message: fmt.Sprintf(format, args...)}
}
