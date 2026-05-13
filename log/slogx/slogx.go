package slogx

import (
	"context"
	"log/slog"
	"os"
	"strings"

	kitcontext "github.com/pumpingbytes/go-kit/context"
	pkglog "github.com/pumpingbytes/go-kit/log"
)

// Load creates a slog.Logger using the shared log.Options/output machinery.
// Closer is returned for flushing any buffers.
//
// appName is used for:
// - default log filename when opts.Dir is set
// - env override: <APPNAME>_LOG_LEVEL (uppercased)
//
// Notes:
// - If Options.Path is set, logs go to that exact file path.
// - Else if Options.Dir is empty, logs go to stderr.
// - Else logs go to <Dir>/log/<File> (or <app>.log).
// - You can still override level with <APP>_LOG_LEVEL for quick local debugging.
func Load(opts pkglog.Options, appName string) (*slog.Logger, pkglog.Closer, error) {
	level := parseLevel(opts.Level)
	if v := os.Getenv(strings.ToUpper(appName) + "_LOG_LEVEL"); strings.TrimSpace(v) != "" {
		level = parseLevel(v)
	}

	hOpts := &slog.HandlerOptions{Level: level, AddSource: opts.AddSource}
	w, closer, err := pkglog.OpenWriter(opts, appName)

	switch opts.Format {
	case pkglog.FormatText:
		return slog.New(slog.NewTextHandler(w, hOpts)), closer, err
	case "", pkglog.FormatJSON:
		fallthrough
	default:
		return slog.New(slog.NewJSONHandler(w, hOpts)), closer, err
	}
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "info", "":
		fallthrough
	default:
		return slog.LevelInfo
	}
}

var defaultLogger = func() *slog.Logger {
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(h)
}()

func defaultStdErrLogger() *slog.Logger { return defaultLogger }

type loggerCtxKey struct{}

// PutLogger returns a new context with the provided logger.
func PutLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return kitcontext.PutValue(ctx, loggerCtxKey{}, logger)
}

// GetLogger returns the logger from the context, falling back to a stderr JSON logger.
func GetLogger(ctx context.Context) *slog.Logger {
	if logger := kitcontext.GetValue[*slog.Logger](ctx, loggerCtxKey{}, nil); logger != nil {
		return logger
	}
	return defaultStdErrLogger()
}
