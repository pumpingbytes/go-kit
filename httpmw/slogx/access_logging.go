package slogx

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/pumpingbytes/go-kit/httpmw"
	"github.com/pumpingbytes/go-kit/log/slogx"
)

// AccessLoggerCtx returns an httpmw.AccessLogger backed by logger fetched from context.
func AccessLoggerCtx(ctx context.Context) httpmw.AccessLogger {
	return AccessLogger(slogx.GetLogger(ctx))
}

// AccessLogger returns an httpmw.AccessLogger backed by slog.
func AccessLogger(base *slog.Logger) httpmw.AccessLogger {
	return func(ctx context.Context, f httpmw.AccessLogFields) {
		if base == nil {
			return // it is the caller's responsibility to provide valid logger.
		}

		logger := EnrichLogger(ctx, base)

		attrs := []any{
			slog.String(string(httpmw.HTTPAccessMethod), f.Method),
			slog.String(string(httpmw.HTTPAccessPath), f.Path),
			slog.Int(string(httpmw.HTTPAccessStatus), f.Status),
			slog.Duration(string(httpmw.HTTPAccessLatency), f.Latency),
			slog.String(string(httpmw.HTTPAccessClientIP), f.ClientIP),
		}
		if f.UserAgent != "" {
			attrs = append(attrs, slog.String(string(httpmw.HTTPAccessUserAgent), f.UserAgent))
		}

		msg := "request"
		if f.Status >= http.StatusInternalServerError {
			logger.Error(msg, attrs...)
			return
		}
		logger.Info(msg, attrs...)
	}
}
