package log

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenWriterUsesStderrWhenDirIsEmpty(t *testing.T) {
	t.Run("empty dir", func(t *testing.T) {
		writer, closer, err := OpenWriter(Options{}, "app")
		if err != nil {
			t.Fatalf("OpenWriter() error = %v, want nil", err)
		}
		if writer != os.Stderr {
			t.Fatalf("writer = %T, want os.Stderr", writer)
		}
		if err := closer.Close(); err != nil {
			t.Fatalf("closer.Close() error = %v, want nil", err)
		}
	})

	t.Run("whitespace dir", func(t *testing.T) {
		writer, closer, err := OpenWriter(Options{Dir: "   \t  "}, "app")
		if err != nil {
			t.Fatalf("OpenWriter() error = %v, want nil", err)
		}
		if writer != os.Stderr {
			t.Fatalf("writer = %T, want os.Stderr", writer)
		}
		if err := closer.Close(); err != nil {
			t.Fatalf("closer.Close() error = %v, want nil", err)
		}
	})
}

func TestOpenWriterCreatesDefaultLogFileAndFlushesOnClose(t *testing.T) {
	dir := t.TempDir()

	writer, closer, err := OpenWriter(Options{Dir: dir}, "myapp")
	if err != nil {
		t.Fatalf("OpenWriter() error = %v, want nil", err)
	}

	if _, err := io.WriteString(writer, "hello world\n"); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}

	logPath := filepath.Join(dir, "log", "myapp.log")
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("stat log file before close: %v", err)
	}

	beforeClose, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile(before close) error = %v", err)
	}
	if string(beforeClose) != "" {
		t.Fatalf("file contents before close = %q, want empty because writes are buffered", string(beforeClose))
	}

	if err := closer.Close(); err != nil {
		t.Fatalf("closer.Close() error = %v", err)
	}

	afterClose, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile(after close) error = %v", err)
	}
	if string(afterClose) != "hello world\n" {
		t.Fatalf("file contents after close = %q, want %q", string(afterClose), "hello world\n")
	}
}

func TestOpenWriterUsesTrimmedCustomFileAndAppends(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "log", "custom.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(logPath, []byte("existing\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	writer, closer, err := OpenWriter(Options{Dir: dir, File: "  custom.log  "}, "ignored")
	if err != nil {
		t.Fatalf("OpenWriter() error = %v, want nil", err)
	}

	if _, err := io.WriteString(writer, "next\n"); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	if err := closer.Close(); err != nil {
		t.Fatalf("closer.Close() error = %v", err)
	}

	got, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "existing\nnext\n" {
		t.Fatalf("file contents = %q, want %q", string(got), "existing\nnext\n")
	}
}

func TestOpenWriterReturnsStderrAndErrorWhenLogDirCannotBeCreated(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "not-a-dir")
	if err := os.WriteFile(filePath, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	writer, closer, err := OpenWriter(Options{Dir: filePath}, "app")
	if err == nil {
		t.Fatal("OpenWriter() error = nil, want non-nil")
	}
	if writer != os.Stderr {
		t.Fatalf("writer = %T, want os.Stderr", writer)
	}
	if closer == nil {
		t.Fatal("closer = nil, want fallback nop closer")
	}
	if err := closer.Close(); err != nil {
		t.Fatalf("closer.Close() error = %v, want nil", err)
	}
}

func TestOpenWriterReturnsStderrAndErrorWhenLogPathIsDirectory(t *testing.T) {
	dir := t.TempDir()
	logDir := filepath.Join(dir, "log")
	if err := os.MkdirAll(filepath.Join(logDir, "as-dir"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	writer, closer, err := OpenWriter(Options{Dir: dir, File: "as-dir"}, "app")
	if err == nil {
		t.Fatal("OpenWriter() error = nil, want non-nil")
	}
	if writer != os.Stderr {
		t.Fatalf("writer = %T, want os.Stderr", writer)
	}
	if closer == nil {
		t.Fatal("closer = nil, want fallback nop closer")
	}
	if err := closer.Close(); err != nil {
		t.Fatalf("closer.Close() error = %v, want nil", err)
	}
}


