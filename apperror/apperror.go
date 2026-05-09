package apperror

import "errors"

// Error is a structured, transport-agnostic error type intended to be
// returned by service-layer code.
//
// It carries:
//   - Code: stable machine-readable code
//   - Message: safe end-user message (especially for non-5xx)
//   - Status: suggested HTTP status for adapters (0 means "unspecified")
//   - Cause: optional wrapped error
//
// This package intentionally has no dependencies on net/http or any HTTP framework.
// It can be reused across services/CLIs.
//
// NOTE: we keep field names short and consistent across repos.
// We may add Meta map[string]any later, in order to provide richer context.
type Error struct {
	Code    string
	Message string
	Status  int
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Code != "" {
		if e.Message != "" {
			return e.Code + ": " + e.Message
		}
		return e.Code
	}
	if e.Message == "" && e.Cause != nil {
		return e.Cause.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Cause }

func New(code, message string, status int) *Error {
	return &Error{Code: code, Message: message, Status: status}
}

func Wrap(err error, code, message string, status int) *Error {
	if err == nil {
		return New(code, message, status)
	}
	return &Error{Code: code, Message: message, Status: status, Cause: err}
}

func As(err error) (*Error, bool) {
	var ae *Error
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
