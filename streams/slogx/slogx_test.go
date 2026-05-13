package slogx

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	logslogx "github.com/pumpingbytes/go-kit/log/slogx"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}})
	logger := slog.New(h)

	s := New(logger, slog.LevelInfo, slog.LevelError)

	if _, err := s.Out().Write([]byte("hello info\n")); err != nil {
		t.Fatalf("Out().Write() error = %v", err)
	}
	if _, err := s.ErrOut().Write([]byte("boom err\n")); err != nil {
		t.Fatalf("ErrOut().Write() error = %v", err)
	}

	time.Sleep(10 * time.Millisecond)
	got := buf.String()
	if !strings.Contains(got, "level=INFO") || !strings.Contains(got, `msg="hello info"`) {
		t.Fatalf("missing info log in output: %q", got)
	}
	if !strings.Contains(got, "level=ERROR") || !strings.Contains(got, `msg="boom err"`) {
		t.Fatalf("missing error log in output: %q", got)
	}
}

func TestNewCtx(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}})
	logger := slog.New(h)
	ctx := logslogx.PutLogger(context.Background(), logger)

	s := NewCtx(ctx, slog.LevelWarn, slog.LevelError)

	if _, err := s.Out().Write([]byte("warn msg\n")); err != nil {
		t.Fatalf("Out().Write() error = %v", err)
	}
	if _, err := s.ErrOut().Write([]byte("error msg\n")); err != nil {
		t.Fatalf("ErrOut().Write() error = %v", err)
	}

	time.Sleep(10 * time.Millisecond)
	got := buf.String()
	if !strings.Contains(got, "level=WARN") || !strings.Contains(got, `msg="warn msg"`) {
		t.Fatalf("missing warn log in output: %q", got)
	}
	if !strings.Contains(got, "level=ERROR") || !strings.Contains(got, `msg="error msg"`) {
		t.Fatalf("missing error log in output: %q", got)
	}
}
