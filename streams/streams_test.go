package streams

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestDefaultIOStreams(t *testing.T) {
	s := NewDefault()

	// Check identity for In(). We won't write to Out/ErrOut to avoid polluting test output.
	if s.In() != os.Stdin {
		t.Fatalf("DefaultIOStreams.In() should be os.Stdin")
	}
	if s.Out() == nil || s.ErrOut() == nil {
		t.Fatalf("DefaultIOStreams Out/ErrOut must be non-nil")
	}
	// Spot-check that type is our concrete BasicIOStreams
	if _, ok := any(s).(Basic); !ok {
		t.Fatalf("DefaultIOStreams() must return BasicIOStreams")
	}
}

func TestWriters(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	s := NewBasic(os.Stdin, &outBuf, &errBuf)

	n, err := s.Out().Write([]byte("hello out\n"))
	if err != nil || n != len("hello out\n") {
		t.Fatalf("Out() write failed: n=%d err=%v", n, err)
	}
	n, err = s.ErrOut().Write([]byte("hello err\n"))
	if err != nil || n != len("hello err\n") {
		t.Fatalf("ErrOut() write failed: n=%d err=%v", n, err)
	}

	if got := outBuf.String(); got != "hello out\n" {
		t.Fatalf("Out buffer = %q, want %q", got, "hello out\n")
	}
	if got := errBuf.String(); got != "hello err\n" {
		t.Fatalf("Err buffer = %q, want %q", got, "hello err\n")
	}

	// Type check
	if _, ok := any(s).(Basic); !ok {
		t.Fatalf("Writers() must return BasicIOStreams")
	}
}

func TestNew(t *testing.T) {
	in := strings.NewReader("input")
	var outBuf, errBuf bytes.Buffer
	s := NewBasic(in, &outBuf, &errBuf)

	if s.In() != in {
		t.Fatalf("New().In() mismatch")
	}
	if s.Out() != &outBuf {
		t.Fatalf("New().Out() mismatch")
	}
	if s.ErrOut() != &errBuf {
		t.Fatalf("New().ErrOut() mismatch")
	}
}

func TestDiscard(t *testing.T) {
	s := NewSilent()

	// Writes should be accepted with full length, but nothing is captured.
	for _, w := range []io.Writer{s.Out(), s.ErrOut()} {
		n, err := w.Write([]byte("dropped\n"))
		if err != nil || n != len("dropped\n") {
			t.Fatalf("discard write failed: n=%d err=%v", n, err)
		}
	}

	// Type check
	if _, ok := any(s).(Basic); !ok {
		t.Fatalf("Discard() must return BasicIOStreams")
	}
}

func TestBuffersStreams(t *testing.T) {
	bs := NewBuffers()

	// Writes accumulate in buffers.
	if _, err := bs.Out().Write([]byte("info 1\n")); err != nil {
		t.Fatalf("write to Out: %v", err)
	}
	if _, err := bs.ErrOut().Write([]byte("err 1\n")); err != nil {
		t.Fatalf("write to ErrOut: %v", err)
	}

	out, errS := bs.Strings()
	if out != "info 1\n" || errS != "err 1\n" {
		t.Fatalf("Strings() = %q / %q, want %q / %q", out, errS, "info 1\n", "err 1\n")
	}

	// Reset clears both.
	bs.Reset()
	out, errS = bs.Strings()
	if out != "" || errS != "" {
		t.Fatalf("after Reset, got %q / %q, want empty / empty", out, errS)
	}

	// In() should be os.Stdin by default.
	if bs.In() != os.Stdin {
		t.Fatalf("BuffersStreams.In() should be os.Stdin")
	}
}

func TestThreadSafeBuffersStreams(t *testing.T) {
	ts := NewThreadSafeBuffers()

	var wg sync.WaitGroup
	wg.Add(200)

	// Concurrent writers on Out and ErrOut.
	for i := 0; i < 100; i++ {
		go func(i int) {
			defer wg.Done()
			_, _ = ts.Out().Write([]byte("O"))
		}(i)
		go func(i int) {
			defer wg.Done()
			_, _ = ts.ErrOut().Write([]byte("E"))
		}(i)
	}
	wg.Wait()

	out, errS := ts.Strings()
	if len(out) != 100 || strings.Count(out, "O") != 100 {
		t.Fatalf("Out length/count mismatch, got len=%d, content=%q", len(out), out)
	}
	if len(errS) != 100 || strings.Count(errS, "E") != 100 {
		t.Fatalf("ErrOut length/count mismatch, got len=%d, content=%q", len(errS), errS)
	}

	// Reset clears both.
	ts.Reset()
	out, errS = ts.Strings()
	if out != "" || errS != "" {
		t.Fatalf("after Reset, got %q / %q, want empty / empty", out, errS)
	}

	// In() should be os.Stdin by default.
	if ts.In() != os.Stdin {
		t.Fatalf("ThreadSafeBuffersStreams.In() should be os.Stdin")
	}
}
