package log

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Format controls handler output.
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// Options is meant to be embedded into the app config (yaml/env).
//
// Suggested config shape:
//
//	log:
//	  dir: "/var/log/component"   # optional; default: stderr
//	  file: "component.log"      # optional; default: <app>.log when dir is set
//	  format: "json"                            # json|text
//	  level: "info"                             # debug|info|warn|error
//	  add_source: false
type Options struct {
	Dir       string `yaml:"dir" env:"LOG_DIR"`
	File      string `yaml:"file" env:"LOG_FILE"`
	Format    Format `yaml:"format" env:"LOG_FORMAT" default:"json"`
	Level     string `yaml:"level" env:"LOG_LEVEL" default:"info"`
	AddSource bool   `yaml:"add_source" env:"LOG_ADD_SOURCE" default:"false"`
}

type Closer interface{ Close() error }

type nopCloser struct{}

func (nopCloser) Close() error { return nil }

type flushOnClose struct {
	*bufio.Writer
	c io.Closer
}

func (f *flushOnClose) Close() error {
	_ = f.Writer.Flush()
	return f.c.Close()
}

// OpenWriter resolves the configured output destination.
//
// If file-backed logging cannot be opened, stderr is returned together with the
// encountered error so callers can choose whether to fail hard or fall back.
func OpenWriter(opts Options, appName string) (io.Writer, Closer, error) {
	if strings.TrimSpace(opts.Dir) == "" {
		return os.Stderr, nopCloser{}, nil
	}

	logDir := filepath.Join(opts.Dir, "log")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return os.Stderr, nopCloser{}, err
	}

	file := strings.TrimSpace(opts.File)
	if file == "" {
		file = appName + ".log"
	}

	logPath := filepath.Join(logDir, file)
	// O_APPEND to avoid truncation, 0600 as conservative default
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return os.Stderr, nopCloser{}, err
	}

	// Buffer writes to reduce syscalls; flush on Close.
	bufw := bufio.NewWriterSize(f, 64<<10) // 64KB
	sink := &flushOnClose{Writer: bufw, c: f}
	return sink, sink, nil
}

// Deprecated CLI helper: kept for compatibility with earlier code.
// Prefer wiring log.Options via app config.
func GetDir(appName string) (string, error) {
	configPathEnvVarName := strings.ToUpper(appName) + "_CONFIG_PATH"

	configPath := os.Getenv(configPathEnvVarName)
	if configPath != "" {
		return filepath.Dir(configPath), nil
	}

	userConfigDir := os.Getenv("XDG_CONFIG_HOME")
	if userConfigDir != "" {
		return filepath.Join(userConfigDir, appName), nil
	}

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine user config dir: %w", err)
	}
	return filepath.Join(userConfigDir, appName), nil
}
