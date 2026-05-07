package apierror

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pumpingbytes/go-kit/context"
)

// APIError is the standardized JSON error envelope.
//
// It is intentionally framework-agnostic.
// Services/handlers can serialize it as-is.
//
// Contract: {"code": "...", "message": "...", "context": {...}}.
//
//   - code: stable machine-readable error code (optional but preferred)
//   - message: safe end-user message
//   - context: optional user-safe debug/support details (e.g. request ID)
//
// NOTE: Code may be omitted for backwards-compatibility, but we prefer always setting it.
type APIError struct {
	Status  int             `json:"-"`
	Code    string          `json:"code,omitempty"`
	Message string          `json:"message"`
	Context json.RawMessage `json:"context,omitempty"`
	Debug   json.RawMessage `json:"debug,omitempty"` // Debug is for internal use and should not be sent to clients.
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	if e.Code != "" {
		return e.Code + ": " + e.Message
	}
	return e.Message
}

// WithContext returns a copy of the error with context set.
// If the ctx is empty or invalid, it returns the original error.
func (e *APIError) WithContext(ctx context.Context) *APIError {
	if e == nil {
		return nil
	}
	if ctx == nil || len(ctx) == 0 {
		return e
	}
	b, err := ctx.Marshal()
	if err != nil {
		// If context can't be serialized, return the original error without context.
		// TODO: consider a more robust logging strategy here.
		return e
	}
	clone := *e
	clone.Context = b
	return &clone
}

// WithDebug returns a copy of the error with debug information set.
// If ctx is empty or invalid, it returns the original error.
func (e *APIError) WithDebug(ctx context.Context) *APIError {
	if e == nil {
		return nil
	}
	if ctx == nil || len(ctx) == 0 {
		return e
	}
	b, err := ctx.Marshal()
	if err != nil {
		// If debug can't be serialized, return the original error without debug.
		return e
	}
	clone := *e
	clone.Debug = b
	return &clone
}

// WithStatus returns a copy of the error with the provided HTTP status.
// If the status is unchanged or invalid, it returns the original error.
func (e *APIError) WithStatus(status int) *APIError {
	if e == nil {
		return nil
	}
	if status <= 0 || status == e.Status {
		return e
	}
	clone := *e
	clone.Status = status
	return &clone
}

// Cleanup removes any debug information from the error, for safe exposure to clients.
// If there is no debug information, it returns the original error.
func (e *APIError) Cleanup() *APIError {
	if e == nil {
		return nil
	}
	if len(e.Debug) == 0 {
		return e
	}
	clone := *e
	clone.Debug = nil
	return &clone
}

// New creates a base APIError with a code and message.
func New(code, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// NewInvalidRequest creates a standard invalid request error.
func NewInvalidRequest(message string) *APIError {
	return New(CodeInvalidRequest, message)
}

// NewBadGateway creates a standard bad gateway error.
func NewBadGateway(message string) *APIError {
	return New(CodeBadGateway, message)
}

// As unwraps an error into an APIError pointer if possible.
func As(err error) (*APIError, bool) {
	var ae *APIError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

var ErrInternalServerError = New(CodeInternal, "internal error").
	WithStatus(http.StatusInternalServerError)

var ErrServiceUnavailable = New(CodeServiceUnavailable, "service temporarily unavailable").
	WithStatus(http.StatusServiceUnavailable)

var ErrMissingAuthorizationHeader = New(CodeAuthRequired, "missing authorization header").
	WithStatus(http.StatusUnauthorized)

var ErrAuthorizationRequired = New(CodeAuthRequired, "authorization required").
	WithStatus(http.StatusUnauthorized)

var ErrInvalidToken = New(CodeInvalidToken, "invalid token").
	WithStatus(http.StatusUnauthorized)

var ErrMissingRequiredScope = New(CodeForbidden, "missing required scope").
	WithStatus(http.StatusForbidden)

var ErrTooManyRequests = New(CodeTooManyRequests, "too many requests").
	WithStatus(http.StatusTooManyRequests)

var ErrEntitlementRequired = New(CodeEntitlementRequired, "missing required entitlement").
	WithStatus(http.StatusForbidden)

// Common stable error codes.
const (
	CodeInternal            = "INTERNAL"
	CodeUnknown             = "UNKNOWN"
	CodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	CodeAuthRequired        = "AUTH_REQUIRED"
	CodeForbidden           = "FORBIDDEN"
	CodeInvalidToken        = "INVALID_TOKEN"
	CodeTokenExpired        = "TOKEN_EXPIRED"
	CodeInvalidArgument     = "INVALID_ARGUMENT"
	CodeNotFound            = "NOT_FOUND"
	CodeConflict            = "RESOURCE_CONFLICT"
	CodeInvalidRequest      = "INVALID_REQUEST"
	CodeBadGateway          = "BAD_GATEWAY"
	CodeTooManyRequests     = "TOO_MANY_REQUESTS"
	CodeEntitlementRequired = "ENTITLEMENT_REQUIRED"
)
