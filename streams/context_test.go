package streams

import (
	"context"
	"io"
	"os"
	"testing"
)

func TestPutIOStreamsAndGetIOStreamsRoundTrip(t *testing.T) {
	t.Parallel()

	custom := NewBasic(os.Stdin, io.Discard, os.Stderr)
	ctx := PutIOStreams(context.Background(), custom)
	got := GetIOStreams(ctx)

	if got.In() != custom.In() {
		t.Fatalf("GetIOStreams().In() mismatch")
	}
	if got.Out() != custom.Out() {
		t.Fatalf("GetIOStreams().Out() mismatch")
	}
	if got.ErrOut() != custom.ErrOut() {
		t.Fatalf("GetIOStreams().ErrOut() mismatch")
	}
}

func TestWithIOStreamsCompatibilityWrapper(t *testing.T) {
	t.Parallel()

	custom := NewBasic(os.Stdin, os.Stdout, io.Discard)
	ctx := PutIOStreams(context.Background(), custom)
	got := GetIOStreams(ctx)

	if got.In() != custom.In() {
		t.Fatalf("GetIOStreams().In() mismatch")
	}
	if got.Out() != custom.Out() {
		t.Fatalf("GetIOStreams().Out() mismatch")
	}
	if got.ErrOut() != custom.ErrOut() {
		t.Fatalf("GetIOStreams().ErrOut() mismatch")
	}
}

func TestGetIOStreamsFallsBackToDefault(t *testing.T) {
	t.Parallel()

	got := GetIOStreams(context.Background())
	def := NewDefault()

	if got.In() != def.In() {
		t.Fatalf("GetIOStreams().In() = %v, want default stdin", got.In())
	}
	if got.Out() != def.Out() {
		t.Fatalf("GetIOStreams().Out() = %v, want default stdout", got.Out())
	}
	if got.ErrOut() != def.ErrOut() {
		t.Fatalf("GetIOStreams().ErrOut() = %v, want default stderr", got.ErrOut())
	}
}

func TestGetIOStreamsIgnoresWrongContextType(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), ioStreamsCtxKey{}, "not-streams")
	got := GetIOStreams(ctx)
	def := NewDefault()

	if got.In() != def.In() {
		t.Fatalf("GetIOStreams().In() = %v, want default stdin", got.In())
	}
	if got.Out() != def.Out() {
		t.Fatalf("GetIOStreams().Out() = %v, want default stdout", got.Out())
	}
	if got.ErrOut() != def.ErrOut() {
		t.Fatalf("GetIOStreams().ErrOut() = %v, want default stderr", got.ErrOut())
	}
}
