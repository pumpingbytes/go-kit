package slogx

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pkglog "github.com/pumpingbytes/go-kit/log"
)

func TestLoadFormatSelection(t *testing.T) {
	tests := []struct {
		name       string
		format     pkglog.Format
		addSource  bool
		assertions func(t *testing.T, out string)
	}{
		{
			name:   "json format",
			format: pkglog.FormatJSON,
			assertions: func(t *testing.T, out string) {
				record := decodeSingleJSONRecord(t, out)
				if got := record["msg"]; got != "json message" {
					t.Fatalf("msg = %v, want json message", got)
				}
			},
		},
		{
			name:   "text format",
			format: pkglog.FormatText,
			assertions: func(t *testing.T, out string) {
				if strings.HasPrefix(strings.TrimSpace(out), "{") {
					t.Fatalf("text output unexpectedly looks like JSON: %q", out)
				}
				if !strings.Contains(out, "msg=\"text message\"") {
					t.Fatalf("text output = %q, want msg=\"text message\"", out)
				}
			},
		},
		{
			name:   "empty format defaults to json",
			format: "",
			assertions: func(t *testing.T, out string) {
				record := decodeSingleJSONRecord(t, out)
				if got := record["msg"]; got != "default json message" {
					t.Fatalf("msg = %v, want default json message", got)
				}
			},
		},
		{
			name:   "unknown format defaults to json",
			format: pkglog.Format("console"),
			assertions: func(t *testing.T, out string) {
				record := decodeSingleJSONRecord(t, out)
				if got := record["msg"]; got != "unknown format message" {
					t.Fatalf("msg = %v, want unknown format message", got)
				}
			},
		},
		{
			name:      "json add source",
			format:    pkglog.FormatJSON,
			addSource: true,
			assertions: func(t *testing.T, out string) {
				record := decodeSingleJSONRecord(t, out)
				if got := record["msg"]; got != "source message" {
					t.Fatalf("msg = %v, want source message", got)
				}
				if _, ok := record["source"]; !ok {
					t.Fatalf("expected source field in %v", record)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			appName := "formatapp"
			logger, closer, err := Load(pkglog.Options{Dir: dir, Format: tt.format, AddSource: tt.addSource}, appName)
			if err != nil {
				t.Fatalf("Load() error = %v, want nil", err)
			}

			msg := map[string]string{
				"json format":                     "json message",
				"text format":                     "text message",
				"empty format defaults to json":   "default json message",
				"unknown format defaults to json": "unknown format message",
				"json add source":                 "source message",
			}[tt.name]
			logger.Info(msg)
			if err := closer.Close(); err != nil {
				t.Fatalf("closer.Close() error = %v", err)
			}

			out := readLogFile(t, dir, appName)
			tt.assertions(t, out)
		})
	}
}

func TestLoadLevelSelection(t *testing.T) {
	tests := []struct {
		name             string
		level            string
		envOverride      string
		appName          string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:          "debug level logs debug and info",
			level:         "debug",
			appName:       "debugapp",
			shouldContain: []string{"debug message", "info message"},
		},
		{
			name:             "warn level filters info",
			level:            "warn",
			appName:          "warnapp",
			shouldContain:    []string{"warn message"},
			shouldNotContain: []string{"info message"},
		},
		{
			name:             "invalid level defaults to info",
			level:            "nope",
			appName:          "invalidapp",
			shouldContain:    []string{"info message"},
			shouldNotContain: []string{"debug message"},
		},
		{
			name:             "env override wins",
			level:            "error",
			envOverride:      "debug",
			appName:          "overrideapp",
			shouldContain:    []string{"debug message", "info message", "warn message"},
			shouldNotContain: []string{""},
		},
		{
			name:             "warning alias works",
			level:            "warning",
			appName:          "warningapp",
			shouldContain:    []string{"warn message"},
			shouldNotContain: []string{"info message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.envOverride != "" {
				t.Setenv(strings.ToUpper(tt.appName)+"_LOG_LEVEL", tt.envOverride)
			}

			logger, closer, err := Load(pkglog.Options{Dir: dir, Format: pkglog.FormatJSON, Level: tt.level}, tt.appName)
			if err != nil {
				t.Fatalf("Load() error = %v, want nil", err)
			}

			logger.Debug("debug message")
			logger.Info("info message")
			logger.Warn("warn message")
			if err := closer.Close(); err != nil {
				t.Fatalf("closer.Close() error = %v", err)
			}

			out := readLogFile(t, dir, tt.appName)
			for _, want := range tt.shouldContain {
				if want == "" {
					continue
				}
				if !strings.Contains(out, want) {
					t.Fatalf("output = %q, want to contain %q", out, want)
				}
			}
			for _, unwanted := range tt.shouldNotContain {
				if unwanted == "" {
					continue
				}
				if strings.Contains(out, unwanted) {
					t.Fatalf("output = %q, want not to contain %q", out, unwanted)
				}
			}
		})
	}
}

func TestLoadFallsBackToStderrOnOpenWriterError(t *testing.T) {
	tests := []struct {
		name string
		opts pkglog.Options
	}{
		{
			name: "mkdir failure falls back to stderr",
			opts: pkglog.Options{Dir: createFilePath(t, "not-a-dir"), Format: pkglog.FormatJSON},
		},
		{
			name: "open file failure falls back to stderr",
			opts: pkglog.Options{Dir: createDirWithDirectoryLogPath(t, "as-dir"), File: "as-dir", Format: pkglog.FormatJSON},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stderr := captureStderr(t)
			logger, closer, err := Load(tt.opts, "stderrapp")
			if err == nil {
				t.Fatal("Load() error = nil, want non-nil")
			}
			if closer == nil {
				t.Fatal("closer = nil, want fallback closer")
			}

			logger.Info("stderr fallback message", slog.String("kind", tt.name))
			if err := closer.Close(); err != nil {
				t.Fatalf("closer.Close() error = %v", err)
			}

			out := stderr.readAll(t)
			record := decodeSingleJSONRecord(t, out)
			if got := record["msg"]; got != "stderr fallback message" {
				t.Fatalf("msg = %v, want stderr fallback message", got)
			}
		})
	}
}

func readLogFile(t *testing.T, dir, appName string) string {
	t.Helper()
	path := filepath.Join(dir, "log", appName+".log")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return strings.TrimSpace(string(data))
}

func decodeSingleJSONRecord(t *testing.T, out string) map[string]any {
	t.Helper()
	line := strings.TrimSpace(out)
	if line == "" {
		t.Fatal("expected a log record, got empty output")
	}

	var record map[string]any
	if err := json.Unmarshal([]byte(line), &record); err != nil {
		t.Fatalf("unmarshal log record: %v\noutput: %s", err, line)
	}
	return record
}

type capturedStderr struct {
	old *os.File
	r   *os.File
	w   *os.File
}

func captureStderr(t *testing.T) *capturedStderr {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	stderrCapture := &capturedStderr{old: os.Stderr, r: r, w: w}
	os.Stderr = w
	t.Cleanup(func() {
		os.Stderr = stderrCapture.old
		_ = stderrCapture.w.Close()
		_ = stderrCapture.r.Close()
	})
	return stderrCapture
}

func (c *capturedStderr) readAll(t *testing.T) string {
	t.Helper()
	if err := c.w.Close(); err != nil {
		t.Fatalf("stderr writer Close() error = %v", err)
	}
	data, err := io.ReadAll(c.r)
	if err != nil {
		t.Fatalf("ReadAll(stderr) error = %v", err)
	}
	return strings.TrimSpace(string(data))
}

func createFilePath(t *testing.T, name string) string {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
	return path
}

func createDirWithDirectoryLogPath(t *testing.T, fileName string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "log", fileName), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	return dir
}
