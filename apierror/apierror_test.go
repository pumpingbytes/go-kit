package apierror

import (
	"bytes"
	stdcontext "context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/pumpingbytes/go-kit/apperror"
	"github.com/pumpingbytes/go-kit/context"
	"github.com/pumpingbytes/go-kit/httpmw"
)

func TestAPIErrorError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  *APIError
		want string
	}{
		{
			name: "with_code",
			err:  &APIError{Code: "INVALID", Message: "bad input"},
			want: "INVALID: bad input",
		},
		{
			name: "without_code",
			err:  &APIError{Message: "bad input"},
			want: "bad input",
		},
		{
			name: "nil_receiver",
			err:  nil,
			want: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.err.Error(); got != tc.want {
				t.Fatalf("Error() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAPIErrorWithContext(t *testing.T) {
	t.Parallel()

	baseWithContext := &APIError{
		Code:    "INVALID",
		Message: "bad input",
		Context: json.RawMessage(`{"a":1}`),
	}

	cases := []struct {
		name          string
		base          *APIError
		ctx           context.Context
		want          *APIError
		baseUnchanged *APIError
	}{
		{
			name:          "nil_context",
			base:          baseWithContext,
			ctx:           nil,
			want:          baseWithContext,
			baseUnchanged: baseWithContext,
		},
		{
			name:          "empty_context",
			base:          baseWithContext,
			ctx:           context.Context{},
			want:          baseWithContext,
			baseUnchanged: baseWithContext,
		},
		{
			name: "valid_context",
			base: &APIError{Code: "INVALID", Message: "bad input"},
			ctx:  context.Context{"request_id": "r1"},
			want: &APIError{
				Code:    "INVALID",
				Message: "bad input",
				Context: mustJSON(t, context.Context{"request_id": "r1"}),
			},
			baseUnchanged: &APIError{Code: "INVALID", Message: "bad input"},
		},
		{
			name:          "marshal_error",
			base:          baseWithContext,
			ctx:           context.Context{"bad": make(chan int)},
			want:          baseWithContext,
			baseUnchanged: baseWithContext,
		},
		{
			name:          "nil_receiver",
			base:          nil,
			ctx:           context.Context{"request_id": "r1"},
			want:          nil,
			baseUnchanged: nil,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.base.WithContext(tc.ctx)
			assertAPIErrorPtr(t, got, tc.want)
			assertAPIErrorPtr(t, tc.base, tc.baseUnchanged)
		})
	}
}

func TestAPIErrorWithDebug(t *testing.T) {
	t.Parallel()

	baseWithDebug := &APIError{
		Code:    "INVALID",
		Message: "bad input",
		Debug:   json.RawMessage(`{"d":true}`),
	}

	cases := []struct {
		name          string
		base          *APIError
		debug         context.Context
		want          *APIError
		baseUnchanged *APIError
	}{
		{
			name:          "nil_debug",
			base:          baseWithDebug,
			debug:         nil,
			want:          baseWithDebug,
			baseUnchanged: baseWithDebug,
		},
		{
			name:          "empty_debug",
			base:          baseWithDebug,
			debug:         context.Context{},
			want:          baseWithDebug,
			baseUnchanged: baseWithDebug,
		},
		{
			name:  "valid_debug",
			base:  &APIError{Code: "INVALID", Message: "bad input"},
			debug: context.Context{"trace": "t1"},
			want: &APIError{
				Code:    "INVALID",
				Message: "bad input",
				Debug:   mustJSON(t, context.Context{"trace": "t1"}),
			},
			baseUnchanged: &APIError{Code: "INVALID", Message: "bad input"},
		},
		{
			name:          "marshal_error",
			base:          baseWithDebug,
			debug:         context.Context{"bad": make(chan int)},
			want:          baseWithDebug,
			baseUnchanged: baseWithDebug,
		},
		{
			name:          "nil_receiver",
			base:          nil,
			debug:         context.Context{"trace": "t1"},
			want:          nil,
			baseUnchanged: nil,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.base.WithDebug(tc.debug)
			assertAPIErrorPtr(t, got, tc.want)
			assertAPIErrorPtr(t, tc.base, tc.baseUnchanged)
		})
	}
}

func TestAPIErrorWithStatus(t *testing.T) {
	t.Parallel()

	base := &APIError{Status: 400, Code: "INVALID", Message: "bad input"}

	cases := []struct {
		name          string
		base          *APIError
		status        int
		want          *APIError
		baseUnchanged *APIError
	}{
		{
			name:          "nil_receiver",
			base:          nil,
			status:        500,
			want:          nil,
			baseUnchanged: nil,
		},
		{
			name:          "invalid_status",
			base:          base,
			status:        0,
			want:          base,
			baseUnchanged: base,
		},
		{
			name:          "same_status",
			base:          base,
			status:        400,
			want:          base,
			baseUnchanged: base,
		},
		{
			name:   "changes_status",
			base:   base,
			status: 422,
			want: &APIError{
				Status:  422,
				Code:    "INVALID",
				Message: "bad input",
			},
			baseUnchanged: base,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.base.WithStatus(tc.status)
			assertAPIErrorPtr(t, got, tc.want)
			assertAPIErrorPtr(t, tc.base, tc.baseUnchanged)
		})
	}
}

func TestAPIErrorCleanup(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		base          *APIError
		want          *APIError
		baseUnchanged *APIError
	}{
		{
			name:          "removes_debug",
			base:          &APIError{Code: "INVALID", Message: "bad input", Debug: json.RawMessage(`{"d":true}`)},
			want:          &APIError{Code: "INVALID", Message: "bad input"},
			baseUnchanged: &APIError{Code: "INVALID", Message: "bad input", Debug: json.RawMessage(`{"d":true}`)},
		},
		{
			name:          "no_debug",
			base:          &APIError{Code: "INVALID", Message: "bad input"},
			want:          &APIError{Code: "INVALID", Message: "bad input"},
			baseUnchanged: &APIError{Code: "INVALID", Message: "bad input"},
		},
		{
			name:          "nil_receiver",
			base:          nil,
			want:          nil,
			baseUnchanged: nil,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.base.Cleanup()
			assertAPIErrorPtr(t, got, tc.want)
			assertAPIErrorPtr(t, tc.base, tc.baseUnchanged)
		})
	}
}

func TestAPIErrorConstructors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		got  *APIError
		want *APIError
	}{
		{
			name: "new",
			got:  New("INVALID", "bad input"),
			want: &APIError{Code: "INVALID", Message: "bad input"},
		},
		{
			name: "new_invalid_request",
			got:  NewInvalidRequest("bad input"),
			want: &APIError{Code: CodeInvalidRequest, Message: "bad input"},
		},
		{
			name: "new_bad_gateway",
			got:  NewBadGateway("upstream down"),
			want: &APIError{Code: CodeBadGateway, Message: "upstream down"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assertAPIErrorPtr(t, tc.got, tc.want)
		})
	}
}

func TestAs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		err      error
		wantOK   bool
		wantCode string
	}{
		{
			name:     "direct_apierror",
			err:      &APIError{Code: "DIRECT", Message: "x"},
			wantOK:   true,
			wantCode: "DIRECT",
		},
		{
			name:     "wrapped_apierror",
			err:      fmt.Errorf("wrap: %w", &APIError{Code: "WRAPPED", Message: "x"}),
			wantOK:   true,
			wantCode: "WRAPPED",
		},
		{
			name:   "string_error",
			err:    fmt.Errorf("%v", APIError{Code: "VALUE", Message: "x"}),
			wantOK: false,
		},
		{
			name:   "other_error",
			err:    errors.New("nope"),
			wantOK: false,
		},
		{
			name:   "nil_error",
			err:    nil,
			wantOK: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, ok := As(tc.err)
			if ok != tc.wantOK {
				t.Fatalf("As() ok = %v, want %v", ok, tc.wantOK)
			}
			if !tc.wantOK {
				if got != nil {
					t.Fatalf("As() returned non-nil error when ok=false")
				}
				return
			}
			if got == nil {
				t.Fatalf("As() returned nil error when ok=true")
			}
			if got.Code != tc.wantCode {
				t.Fatalf("As() code = %q, want %q", got.Code, tc.wantCode)
			}
		})
	}
}

func TestFromAppError(t *testing.T) {
	t.Parallel()

	direct := &apperror.Error{Code: "INVALID", Message: "bad input", Status: 422}
	wrapped := fmt.Errorf("wrap: %w", &apperror.Error{Code: "FAILED", Message: "save failed", Status: 500})
	other := errors.New("nope")

	cases := []struct {
		name   string
		err    error
		wantOK bool
		want   *APIError
	}{
		{
			name:   "direct apperror",
			err:    direct,
			wantOK: true,
			want:   &APIError{Code: "INVALID", Message: "bad input", Status: 422},
		},
		{
			name:   "wrapped apperror",
			err:    wrapped,
			wantOK: true,
			want:   &APIError{Code: "FAILED", Message: "save failed", Status: 500},
		},
		{
			name:   "other error",
			err:    other,
			wantOK: false,
			want:   nil,
		},
		{
			name:   "nil error",
			err:    nil,
			wantOK: false,
			want:   nil,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, ok := FromAppError(tc.err)
			if ok != tc.wantOK {
				t.Fatalf("FromAppError() ok = %v, want %v", ok, tc.wantOK)
			}
			if !tc.wantOK {
				if got != nil {
					t.Fatalf("FromAppError() = %#v, want nil", got)
				}
				return
			}
			assertAPIErrorPtr(t, got, tc.want)
			if len(got.Context) != 0 {
				t.Fatalf("FromAppError() context = %s, want empty", string(got.Context))
			}
			if len(got.Debug) != 0 {
				t.Fatalf("FromAppError() debug = %s, want empty", string(got.Debug))
			}
		})
	}
}

func TestAPIErrorWithHTTPMWErrorContext(t *testing.T) {
	t.Parallel()

	ctx := stdcontext.Background()
	ctx = httpmw.PutRequestID(ctx, "req-1")
	ctx = httpmw.PutTraceIDs(ctx, "trace-1", "span-1")

	got := ErrInternalServerError.WithContext(httpmw.ErrorContext(ctx))
	want := &APIError{
		Status:  ErrInternalServerError.Status,
		Code:    CodeInternal,
		Message: "internal error",
		Context: mustJSON(t, context.Context{
			httpmw.RequestID: "req-1",
			httpmw.TraceID:   "trace-1",
			httpmw.SpanID:    "span-1",
		}),
	}

	assertAPIErrorPtr(t, got, want)
}

func mustJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	return b
}

func assertAPIError(t *testing.T, got, want APIError) {
	t.Helper()
	if got.Status != want.Status {
		t.Fatalf("Status = %d, want %d", got.Status, want.Status)
	}
	if got.Code != want.Code {
		t.Fatalf("Code = %q, want %q", got.Code, want.Code)
	}
	if got.Message != want.Message {
		t.Fatalf("Message = %q, want %q", got.Message, want.Message)
	}
	if !bytes.Equal(got.Context, want.Context) {
		t.Fatalf("Context = %s, want %s", string(got.Context), string(want.Context))
	}
	if !bytes.Equal(got.Debug, want.Debug) {
		t.Fatalf("Debug = %s, want %s", string(got.Debug), string(want.Debug))
	}
}

func assertAPIErrorPtr(t *testing.T, got, want *APIError) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Fatalf("APIError nil mismatch: got=%v want=%v", got, want)
		}
		return
	}
	if got.Status != want.Status {
		t.Fatalf("Status = %d, want %d", got.Status, want.Status)
	}
	if got.Code != want.Code {
		t.Fatalf("Code = %q, want %q", got.Code, want.Code)
	}
	if got.Message != want.Message {
		t.Fatalf("Message = %q, want %q", got.Message, want.Message)
	}
	if !bytes.Equal(got.Context, want.Context) {
		t.Fatalf("Context = %s, want %s", string(got.Context), string(want.Context))
	}
	if !bytes.Equal(got.Debug, want.Debug) {
		t.Fatalf("Debug = %s, want %s", string(got.Debug), string(want.Debug))
	}
}
