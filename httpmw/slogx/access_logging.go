package slogx

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/pumpingbytes/go-kit/httpmw"
	"github.com/pumpingbytes/go-kit/log/slogx"
)

// AccessLoggerOptions configures slog-backed access log emission.
type AccessLoggerOptions struct {
	LevelPolicy func(httpmw.AccessLogFields) slog.Level
}

// AccessLoggerCtx returns an httpmw.AccessLogger backed by logger fetched from context.
func AccessLoggerCtx(ctx context.Context) httpmw.AccessLogger {
	return AccessLoggerWithOptions(slogx.GetLogger(ctx), AccessLoggerOptions{})
}

// AccessLoggerCtxWithOptions returns an httpmw.AccessLogger backed by logger fetched from context.
func AccessLoggerCtxWithOptions(ctx context.Context, opts AccessLoggerOptions) httpmw.AccessLogger {
	return AccessLoggerWithOptions(slogx.GetLogger(ctx), opts)
}

// AccessLogger returns an httpmw.AccessLogger backed by slog.
func AccessLogger(base *slog.Logger) httpmw.AccessLogger {
	return AccessLoggerWithOptions(base, AccessLoggerOptions{})
}

// AccessLoggerWithOptions returns an httpmw.AccessLogger backed by slog.
func AccessLoggerWithOptions(base *slog.Logger, opts AccessLoggerOptions) httpmw.AccessLogger {
	levelPolicy := opts.LevelPolicy
	if levelPolicy == nil {
		levelPolicy = defaultAccessLogLevelPolicy
	}

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
			slog.Int64(string(httpmw.HTTPAccessResponseBytes), f.ResponseBytes),
		}
		if f.UserAgent != "" {
			attrs = append(attrs, slog.String(string(httpmw.HTTPAccessUserAgent), f.UserAgent))
		}

		logger.Log(ctx, levelPolicy(f), "request", attrs...)
	}
}

func defaultAccessLogLevelPolicy(f httpmw.AccessLogFields) slog.Level {
	if f.Status >= http.StatusInternalServerError {
		return slog.LevelError
	}
	return slog.LevelInfo
}
