package httpmw

import (
	"context"
	"reflect"
	"testing"

	kitcontext "github.com/pumpingbytes/go-kit/context"
)

func TestRequestIDContextHelpers(t *testing.T) {
	t.Run("missing returns empty", func(t *testing.T) {
		if got := GetRequestID(context.Background()); got != "" {
			t.Fatalf("GetRequestID() = %q, want empty", got)
		}
	})

	t.Run("round trip", func(t *testing.T) {
		ctx := PutRequestID(context.Background(), "req-123")
		if got := GetRequestID(ctx); got != "req-123" {
			t.Fatalf("GetRequestID() = %q, want req-123", got)
		}
	})

	t.Run("overwrite keeps latest", func(t *testing.T) {
		ctx := PutRequestID(context.Background(), "req-1")
		ctx = PutRequestID(ctx, "req-2")
		if got := GetRequestID(ctx); got != "req-2" {
			t.Fatalf("GetRequestID() = %q, want req-2", got)
		}
	})

	t.Run("wrong type returns empty", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), requestIDCtxKey{}, 42)
		if got := GetRequestID(ctx); got != "" {
			t.Fatalf("GetRequestID() = %q, want empty", got)
		}
	})
}

func TestTraceContextHelpers(t *testing.T) {
	t.Run("missing returns empty", func(t *testing.T) {
		ctx := context.Background()
		if got := GetTraceID(ctx); got != "" {
			t.Fatalf("GetTraceID() = %q, want empty", got)
		}
		if got := GetSpanID(ctx); got != "" {
			t.Fatalf("GetSpanID() = %q, want empty", got)
		}
	})

	t.Run("round trip", func(t *testing.T) {
		ctx := PutTraceIDs(context.Background(), "trace-123", "span-456")
		if got := GetTraceID(ctx); got != "trace-123" {
			t.Fatalf("GetTraceID() = %q, want trace-123", got)
		}
		if got := GetSpanID(ctx); got != "span-456" {
			t.Fatalf("GetSpanID() = %q, want span-456", got)
		}
	})

	t.Run("overwrite keeps latest pair", func(t *testing.T) {
		ctx := PutTraceIDs(context.Background(), "trace-1", "span-1")
		ctx = PutTraceIDs(ctx, "trace-2", "span-2")
		if got := GetTraceID(ctx); got != "trace-2" {
			t.Fatalf("GetTraceID() = %q, want trace-2", got)
		}
		if got := GetSpanID(ctx); got != "span-2" {
			t.Fatalf("GetSpanID() = %q, want span-2", got)
		}
	})

	t.Run("empty span is preserved", func(t *testing.T) {
		ctx := PutTraceIDs(context.Background(), "trace-123", "")
		if got := GetTraceID(ctx); got != "trace-123" {
			t.Fatalf("GetTraceID() = %q, want trace-123", got)
		}
		if got := GetSpanID(ctx); got != "" {
			t.Fatalf("GetSpanID() = %q, want empty", got)
		}
	})

	t.Run("wrong type returns empty", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), traceIDsCtxKey{}, "bad")
		if got := GetTraceID(ctx); got != "" {
			t.Fatalf("GetTraceID() = %q, want empty", got)
		}
		if got := GetSpanID(ctx); got != "" {
			t.Fatalf("GetSpanID() = %q, want empty", got)
		}
	})
}

func TestErrorContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want kitcontext.Context
	}{
		{
			name: "empty context returns nil",
			ctx:  context.Background(),
			want: nil,
		},
		{
			name: "request id only",
			ctx:  PutRequestID(context.Background(), "req-1"),
			want: kitcontext.Context{RequestID: "req-1"},
		},
		{
			name: "trace and span only",
			ctx:  PutTraceIDs(context.Background(), "trace-1", "span-1"),
			want: kitcontext.Context{TraceID: "trace-1", SpanID: "span-1"},
		},
		{
			name: "span only is preserved",
			ctx:  context.WithValue(context.Background(), traceIDsCtxKey{}, traceIDs{SpanID: "span-1"}),
			want: kitcontext.Context{SpanID: "span-1"},
		},
		{
			name: "all identifiers",
			ctx:  PutTraceIDs(PutRequestID(context.Background(), "req-1"), "trace-1", "span-1"),
			want: kitcontext.Context{RequestID: "req-1", TraceID: "trace-1", SpanID: "span-1"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := ErrorContext(tt.ctx)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ErrorContext() mismatch\n got: %#v\nwant: %#v", got, tt.want)
			}
		})
	}
}
