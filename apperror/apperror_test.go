package apperror

import (
	stderrors "errors"
	"fmt"
	"testing"
)

func TestErrorString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "nil receiver",
			err:  nil,
			want: "",
		},
		{
			name: "code and message",
			err:  &Error{Code: "INVALID", Message: "bad input"},
			want: "INVALID: bad input",
		},
		{
			name: "code only",
			err:  &Error{Code: "INVALID"},
			want: "INVALID",
		},
		{
			name: "message only",
			err:  &Error{Message: "bad input"},
			want: "bad input",
		},
		{
			name: "empty error",
			err:  &Error{},
			want: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.err.Error(); got != tc.want {
				t.Fatalf("Error() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	t.Parallel()

	cause := stderrors.New("boom")
	err := &Error{Code: "FAILED", Cause: cause}
	if got := err.Unwrap(); got != cause {
		t.Fatalf("Unwrap() = %v, want %v", got, cause)
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	err := New("INVALID", "bad input", 400)
	if err == nil {
		t.Fatal("New() = nil, want non-nil")
	}
	if err.Code != "INVALID" || err.Message != "bad input" || err.Status != 400 || err.Cause != nil {
		t.Fatalf("New() = %#v, want code/message/status set and nil cause", err)
	}
}

func TestWrap(t *testing.T) {
	t.Parallel()

	cause := stderrors.New("boom")

	cases := []struct {
		name    string
		cause   error
		code    string
		message string
		status  int
		want    *Error
	}{
		{
			name:    "nil cause behaves like new",
			cause:   nil,
			code:    "INVALID",
			message: "bad input",
			status:  400,
			want:    &Error{Code: "INVALID", Message: "bad input", Status: 400},
		},
		{
			name:    "wraps cause",
			cause:   cause,
			code:    "FAILED",
			message: "save failed",
			status:  500,
			want:    &Error{Code: "FAILED", Message: "save failed", Status: 500, Cause: cause},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := Wrap(tc.cause, tc.code, tc.message, tc.status)
			if got == nil {
				t.Fatal("Wrap() = nil, want non-nil")
			}
			if got.Code != tc.want.Code || got.Message != tc.want.Message || got.Status != tc.want.Status || got.Cause != tc.want.Cause {
				t.Fatalf("Wrap() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestAs(t *testing.T) {
	t.Parallel()

	direct := &Error{Code: "DIRECT", Message: "x", Status: 400}
	wrapped := fmt.Errorf("wrap: %w", &Error{Code: "WRAPPED", Message: "y", Status: 500})
	other := stderrors.New("other")

	cases := []struct {
		name   string
		err    error
		wantOK bool
		want   *Error
	}{
		{name: "direct", err: direct, wantOK: true, want: direct},
		{name: "wrapped", err: wrapped, wantOK: true, want: &Error{Code: "WRAPPED", Message: "y", Status: 500}},
		{name: "other", err: other, wantOK: false, want: nil},
		{name: "nil", err: nil, wantOK: false, want: nil},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := As(tc.err)
			if ok != tc.wantOK {
				t.Fatalf("As() ok = %v, want %v", ok, tc.wantOK)
			}
			if !tc.wantOK {
				if got != nil {
					t.Fatalf("As() = %#v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("As() = nil, want non-nil")
			}
			if got.Code != tc.want.Code || got.Message != tc.want.Message || got.Status != tc.want.Status {
				t.Fatalf("As() = %#v, want %#v", got, tc.want)
			}
		})
	}
}
