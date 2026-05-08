package httpmw

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAccessLogCapturesExplicitStatusAndBytes(t *testing.T) {
	var got AccessLogFields
	var calls int

	handler := AccessLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, "hello")
	}), func(ctx context.Context, f AccessLogFields) {
		calls++
		got = f
	})

	req := httptest.NewRequest(http.MethodPost, "http://example.com/api", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("User-Agent", "curl/8.0")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if calls != 1 {
		t.Fatalf("logger calls = %d, want 1", calls)
	}
	if got.Method != http.MethodPost {
		t.Fatalf("Method = %q, want %q", got.Method, http.MethodPost)
	}
	if got.Path != "/api" {
		t.Fatalf("Path = %q, want %q", got.Path, "/api")
	}
	if got.Status != http.StatusCreated {
		t.Fatalf("Status = %d, want %d", got.Status, http.StatusCreated)
	}
	if got.ResponseBytes != 5 {
		t.Fatalf("ResponseBytes = %d, want 5", got.ResponseBytes)
	}
	if got.ClientIP != "127.0.0.1" {
		t.Fatalf("ClientIP = %q, want %q", got.ClientIP, "127.0.0.1")
	}
	if got.UserAgent != "curl/8.0" {
		t.Fatalf("UserAgent = %q, want %q", got.UserAgent, "curl/8.0")
	}
	if got.Latency < 0 {
		t.Fatalf("Latency = %v, want non-negative", got.Latency)
	}
}

func TestAccessLogCapturesImplicitStatusAndZeroWrite(t *testing.T) {
	var got AccessLogFields

	handler := AccessLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}), func(ctx context.Context, f AccessLogFields) {
		got = f
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/healthz", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got.Status != http.StatusOK {
		t.Fatalf("Status = %d, want %d", got.Status, http.StatusOK)
	}
	if got.ResponseBytes != 2 {
		t.Fatalf("ResponseBytes = %d, want 2", got.ResponseBytes)
	}
}

func TestAccessLogToleratesUnixSocketRemoteAddr(t *testing.T) {
	var got AccessLogFields

	handler := AccessLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), func(ctx context.Context, f AccessLogFields) {
		got = f
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/local", nil)
	req.RemoteAddr = "/var/run/acre-agent.sock"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got.ClientIP != "/var/run/acre-agent.sock" {
		t.Fatalf("ClientIP = %q, want unix socket remote addr preserved", got.ClientIP)
	}
	if got.Status != http.StatusOK {
		t.Fatalf("Status = %d, want %d", got.Status, http.StatusOK)
	}
	if got.ResponseBytes != 0 {
		t.Fatalf("ResponseBytes = %d, want 0", got.ResponseBytes)
	}
}

func TestAccessLogWithOptionsSkipsLoggingButStillServesRequest(t *testing.T) {
	var calls int

	handler := AccessLogWithOptions(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = io.WriteString(w, "skipped")
	}), AccessLogOptions{
		Logger: func(context.Context, AccessLogFields) {
			calls++
		},
		Skip: SkipPaths("/healthz", " /ready "),
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/healthz", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if calls != 0 {
		t.Fatalf("logger calls = %d, want 0", calls)
	}
	if rr.Code != http.StatusAccepted {
		t.Fatalf("response status = %d, want %d", rr.Code, http.StatusAccepted)
	}
	if body := rr.Body.String(); body != "skipped" {
		t.Fatalf("response body = %q, want %q", body, "skipped")
	}
}

func TestSkipPathsMatchesExactPathsOnly(t *testing.T) {
	skip := SkipPaths("/healthz", " /ready ", "")

	if !skip(httptest.NewRequest(http.MethodGet, "http://example.com/healthz", nil)) {
		t.Fatal("SkipPaths did not match exact /healthz path")
	}
	if !skip(httptest.NewRequest(http.MethodGet, "http://example.com/ready", nil)) {
		t.Fatal("SkipPaths did not match trimmed /ready path")
	}
	if skip(httptest.NewRequest(http.MethodGet, "http://example.com/healthz/live", nil)) {
		t.Fatal("SkipPaths unexpectedly matched nested path")
	}
	if skip(nil) {
		t.Fatal("SkipPaths unexpectedly matched nil request")
	}
}

func TestAccessLogCountsBytesViaReadFrom(t *testing.T) {
	var got AccessLogFields
	payload := "streamed payload"

	handler := AccessLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(w, readerOnly{Reader: strings.NewReader(payload)})
	}), func(ctx context.Context, f AccessLogFields) {
		got = f
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/stream", nil)
	rw := &readerFromResponseWriter{header: make(http.Header)}

	handler.ServeHTTP(rw, req)

	if got.ResponseBytes != int64(len(payload)) {
		t.Fatalf("ResponseBytes = %d, want %d", got.ResponseBytes, len(payload))
	}
	if rw.body.String() != payload {
		t.Fatalf("body = %q, want %q", rw.body.String(), payload)
	}
	if got.Status != http.StatusOK {
		t.Fatalf("Status = %d, want %d", got.Status, http.StatusOK)
	}
}

func TestWrapAccessLogResponseWriterPreservesOptionalInterfaces(t *testing.T) {
	base := &flusherOnlyResponseWriter{header: make(http.Header)}
	_, wrapped := wrapAccessLogResponseWriter(base)

	if _, ok := wrapped.(http.Flusher); !ok {
		t.Fatal("wrapped writer does not implement http.Flusher")
	}
	if _, ok := wrapped.(http.Hijacker); ok {
		t.Fatal("wrapped writer unexpectedly implements http.Hijacker")
	}
	if _, ok := wrapped.(http.Pusher); ok {
		t.Fatal("wrapped writer unexpectedly implements http.Pusher")
	}
	if _, ok := wrapped.(io.ReaderFrom); ok {
		t.Fatal("wrapped writer unexpectedly implements io.ReaderFrom")
	}
}

type readerOnly struct {
	io.Reader
}

type flusherOnlyResponseWriter struct {
	header  http.Header
	status  int
	flushed bool
}

func (w *flusherOnlyResponseWriter) Header() http.Header {
	return w.header
}

func (w *flusherOnlyResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return len(p), nil
}

func (w *flusherOnlyResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *flusherOnlyResponseWriter) Flush() {
	w.flushed = true
}

type readerFromResponseWriter struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (w *readerFromResponseWriter) Header() http.Header {
	return w.header
}

func (w *readerFromResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.Write(p)
}

func (w *readerFromResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *readerFromResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return io.Copy(&w.body, r)
}

func (w *readerFromResponseWriter) Flush() {}

func (w *readerFromResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}
