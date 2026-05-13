package streams

import (
	"bytes"
	"io"
	"os"
	"sync"
)

// Buffers captures output into bytes.Buffers. Use this when you want to
// accumulate messages and flush or inspect them after a command or operation
// completes. It is not safe for concurrent writers; see
// ThreadSafeBuffers for a synchronized variant.
type Buffers struct {
	InR    io.Reader
	OutBuf *bytes.Buffer
	ErrBuf *bytes.Buffer
}

// NewBuffers creates a new Buffers with fresh buffers for Out and ErrOut.
func NewBuffers() *Buffers {
	return &Buffers{
		InR:    os.Stdin,
		OutBuf: &bytes.Buffer{},
		ErrBuf: &bytes.Buffer{},
	}
}

func (b *Buffers) In() io.Reader     { return b.InR }
func (b *Buffers) Out() io.Writer    { return b.OutBuf }
func (b *Buffers) ErrOut() io.Writer { return b.ErrBuf }

// Strings returns the current contents of the Out and ErrOut buffers as strings.
func (b *Buffers) Strings() (out, err string) {
	return b.OutBuf.String(), b.ErrBuf.String()
}

// Reset clears both Out and ErrOut buffers.
func (b *Buffers) Reset() {
	b.OutBuf.Reset()
	b.ErrBuf.Reset()
}

// tsBuf is a minimal mutex-protected buffer.
type tsBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (t *tsBuf) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.b.Write(p)
}

func (t *tsBuf) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.b.String()
}

func (t *tsBuf) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.b.Reset()
}

// ThreadSafeBuffers captures output into mutex-protected buffers and is
// safe for concurrent writers.
type ThreadSafeBuffers struct {
	InR    io.Reader
	OutBuf *tsBuf
	ErrBuf *tsBuf
}

// NewThreadSafeBuffers creates a new thread-safe buffers stream set.
func NewThreadSafeBuffers() *ThreadSafeBuffers {
	return &ThreadSafeBuffers{
		InR:    os.Stdin,
		OutBuf: &tsBuf{},
		ErrBuf: &tsBuf{},
	}
}

func (b *ThreadSafeBuffers) In() io.Reader     { return b.InR }
func (b *ThreadSafeBuffers) Out() io.Writer    { return b.OutBuf }
func (b *ThreadSafeBuffers) ErrOut() io.Writer { return b.ErrBuf }

// Strings returns the current contents of the Out and ErrOut buffers as strings.
func (b *ThreadSafeBuffers) Strings() (string, string) {
	return b.OutBuf.String(), b.ErrBuf.String()
}

// Reset clears both Out and ErrOut buffers.
func (b *ThreadSafeBuffers) Reset() {
	b.OutBuf.Reset()
	b.ErrBuf.Reset()
}
