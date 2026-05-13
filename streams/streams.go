// Package streams provides small user-facing input/output stream adapters.
//
// It offers ready-to-use implementations backed by stdio, arbitrary writers,
// discard sinks, and in-memory buffers, plus helpers for propagating stream
// sets through context.Context. Slog-backed adapters live in streams/slogx.
package streams

import (
	"io"
	"os"
)

// IOStreams defines the minimal contract for user‑facing streams.
type IOStreams interface {
	In() io.Reader
	Out() io.Writer
	ErrOut() io.Writer
}

// Basic is a simple, zero‑dependency implementation of IOStreams.
// It forwards writes to the supplied io.Writer targets. Use the helpers
// NewBasic, NewDefault, and NewSilent to construct values quickly.
type Basic struct {
	in     io.Reader
	out    io.Writer
	errOut io.Writer
}

func (s Basic) In() io.Reader     { return s.in }
func (s Basic) Out() io.Writer    { return s.out }
func (s Basic) ErrOut() io.Writer { return s.errOut }

// NewBasic returns a Basic backed by the provided input, output, and error writers.
func NewBasic(in io.Reader, out, errOut io.Writer) Basic {
	return Basic{
		in:     in,
		out:    out,
		errOut: errOut,
	}
}

// NewDefault returns a Basic backed by os.Stdin, os.Stdout and os.Stderr.
func NewDefault() Basic {
	return NewBasic(os.Stdin, os.Stdout, os.Stderr)
}

// NewSilent returns a Basic that drops all output (useful for "--silent").
func NewSilent() Basic {
	return NewBasic(os.Stdin, io.Discard, io.Discard)
}
