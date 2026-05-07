package slogx

import (
	"context"
	"log/slog"

	"github.com/pumpingbytes/go-kit/httpmw"
)

// EnrichLogger returns a logger pre-populated with request/trace fields.
func EnrichLogger(ctx context.Context, logger *slog.Logger) *slog.Logger {
	if logger == nil {
		return nil
	}

	attrs := make([]any, 0, 6)
	if rid := httpmw.GetRequestID(ctx); rid != "" {
		attrs = append(attrs, slog.String(string(httpmw.RequestID), rid))
	}
	if tid := httpmw.GetTraceID(ctx); tid != "" {
		attrs = append(attrs, slog.String(string(httpmw.TraceID), tid))
	}
	if sid := httpmw.GetSpanID(ctx); sid != "" {
		attrs = append(attrs, slog.String(string(httpmw.SpanID), sid))
	}
	if len(attrs) == 0 {
		return logger
	}
	return logger.With(attrs...)
}
