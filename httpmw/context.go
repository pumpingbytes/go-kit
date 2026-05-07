package httpmw

import (
	"context"

	kitcontext "github.com/pumpingbytes/go-kit/context"
)

type requestIDCtxKey struct{}

// PutRequestID returns a new context containing the request ID.
func PutRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDCtxKey{}, requestID)
}

// GetRequestID returns the request ID from context, if present.
func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(requestIDCtxKey{}); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

type traceIDs struct {
	TraceID string
	SpanID  string
}

type traceIDsCtxKey struct{}

// PutTraceIDs returns a new context containing trace and span IDs.
func PutTraceIDs(ctx context.Context, traceID, spanID string) context.Context {
	return context.WithValue(ctx, traceIDsCtxKey{}, traceIDs{TraceID: traceID, SpanID: spanID})
}

// GetTraceID returns the trace ID from context, if present.
func GetTraceID(ctx context.Context) string {
	if v := ctx.Value(traceIDsCtxKey{}); v != nil {
		if t, ok := v.(traceIDs); ok {
			return t.TraceID
		}
	}
	return ""
}

// GetSpanID returns the span ID from context, if present.
func GetSpanID(ctx context.Context) string {
	if v := ctx.Value(traceIDsCtxKey{}); v != nil {
		if t, ok := v.(traceIDs); ok {
			return t.SpanID
		}
	}
	return ""
}

// ErrorContext extracts user-safe request correlation fields from the runtime context
// and returns them as a serializable payload suitable for apierror.APIError.WithContext.
//
// Only non-empty identifiers are included. The returned payload is intended for client-
// visible/support-safe context, so it contains correlation identifiers only and omits
// request details such as headers, paths, or internal errors.
//
// If no request-scoped identifiers are present, ErrorContext returns nil.
//
// Example:
//
//	err := apierror.ErrInternalServerError.WithContext(httpmw.ErrorContext(ctx))
func ErrorContext(ctx context.Context) kitcontext.Context {
	rid := GetRequestID(ctx)
	tid := GetTraceID(ctx)
	sid := GetSpanID(ctx)

	switch {
	case rid != "":
		rest := make([]any, 0, 4)
		if tid != "" {
			rest = append(rest, TraceID, tid)
		}
		if sid != "" {
			rest = append(rest, SpanID, sid)
		}
		return kitcontext.Ctx(RequestID, rid, rest...)
	case tid != "":
		if sid != "" {
			return kitcontext.Ctx(TraceID, tid, SpanID, sid)
		}
		return kitcontext.Ctx(TraceID, tid)
	case sid != "":
		return kitcontext.Ctx(SpanID, sid)
	default:
		return nil
	}
}
