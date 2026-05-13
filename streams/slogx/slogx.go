package slogx

import (
	"bytes"
	"context"
	"log/slog"
	"os"

	logslogx "github.com/pumpingbytes/go-kit/log/slogx"
	streamspkg "github.com/pumpingbytes/go-kit/streams"
)

// writer adapts slog.Logger to io.Writer and trims trailing line endings so each
// Write becomes one log record.
type writer struct {
	logger *slog.Logger
	level  slog.Level
}

func (w writer) Write(p []byte) (int, error) {
	n := len(p)
	p = bytes.TrimRight(p, "\r\n")
	if len(p) == 0 {
		return n, nil
	}

	logger := w.logger
	if logger == nil {
		logger = slog.Default()
	}
	logger.Log(nil, w.level, string(p))
	return n, nil
}

// New returns slog-backed streams using the provided logger.
func New(logger *slog.Logger, outLevel, errLevel slog.Level) streamspkg.Basic {
	return streamspkg.NewBasic(
		os.Stdin,
		writer{logger: logger, level: outLevel},
		writer{logger: logger, level: errLevel},
	)
}

// NewCtx returns slog-backed streams using the logger from context.
func NewCtx(ctx context.Context, outLevel, errLevel slog.Level) streamspkg.Basic {
	return New(logslogx.GetLogger(ctx), outLevel, errLevel)
}
