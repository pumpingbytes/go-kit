package httpmw

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// AccessLogOptions configures net/http access logging middleware.
type AccessLogOptions struct {
	Logger AccessLogger
	Skip   func(*http.Request) bool
}

// AccessLog wraps next with access logging using the provided logger.
func AccessLog(next http.Handler, logger AccessLogger) http.Handler {
	return AccessLogWithOptions(next, AccessLogOptions{Logger: logger})
}

// AccessLogWithOptions wraps next with access logging using the provided options.
func AccessLogWithOptions(next http.Handler, opts AccessLogOptions) http.Handler {
	return WithAccessLogOptions(opts)(next)
}

// WithAccessLog returns a middleware that emits one access-log record per request.
func WithAccessLog(logger AccessLogger) func(http.Handler) http.Handler {
	return WithAccessLogOptions(AccessLogOptions{Logger: logger})
}

// WithAccessLogOptions returns a middleware that emits one access-log record per request.
func WithAccessLogOptions(opts AccessLogOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if next == nil {
			next = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
		}
		if opts.Logger == nil {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if opts.Skip != nil && opts.Skip(r) {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			recorder, wrapped := wrapAccessLogResponseWriter(w)

			next.ServeHTTP(wrapped, r)

			opts.Logger(r.Context(), AccessLogFields{
				Method:        r.Method,
				Path:          r.URL.Path,
				Status:        recorder.Status(),
				Latency:       time.Since(start),
				ClientIP:      clientIPFromRemoteAddr(r.RemoteAddr),
				ResponseBytes: recorder.BytesWritten(),
				UserAgent:     r.UserAgent(),
			})
		})
	}
}

// SkipPaths returns a skip predicate that suppresses access logs for exact request paths.
// Empty path entries are ignored.
func SkipPaths(paths ...string) func(*http.Request) bool {
	pathSet := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if path = strings.TrimSpace(path); path != "" {
			pathSet[path] = struct{}{}
		}
	}

	return func(r *http.Request) bool {
		if r == nil || r.URL == nil {
			return false
		}
		_, ok := pathSet[r.URL.Path]
		return ok
	}
}

type accessLogResponseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int64
}

func (w *accessLogResponseWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *accessLogResponseWriter) Write(p []byte) (int, error) {
	w.ensureStatusOK()
	n, err := w.ResponseWriter.Write(p)
	w.bytesWritten += int64(n)
	return n, err
}

func (w *accessLogResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	w.ensureStatusOK()
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		n, err := rf.ReadFrom(r)
		w.bytesWritten += n
		return n, err
	}
	return io.Copy(accessLogWriterOnly{Writer: w}, r)
}

func (w *accessLogResponseWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *accessLogResponseWriter) BytesWritten() int64 {
	return w.bytesWritten
}

func (w *accessLogResponseWriter) ensureStatusOK() {
	if w.status == 0 {
		w.status = http.StatusOK
	}
}

type accessLogWriterOnly struct {
	io.Writer
}

type accessLogFlusher struct{ *accessLogResponseWriter }

func (w accessLogFlusher) Flush() {
	w.ensureStatusOK()
	w.ResponseWriter.(http.Flusher).Flush()
}

type accessLogHijacker struct{ *accessLogResponseWriter }

func (w accessLogHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

type accessLogPusher struct{ *accessLogResponseWriter }

func (w accessLogPusher) Push(target string, opts *http.PushOptions) error {
	return w.ResponseWriter.(http.Pusher).Push(target, opts)
}

type accessLogReaderFrom struct{ *accessLogResponseWriter }

func (w accessLogReaderFrom) ReadFrom(r io.Reader) (int64, error) {
	return w.accessLogResponseWriter.ReadFrom(r)
}

func wrapAccessLogResponseWriter(w http.ResponseWriter) (*accessLogResponseWriter, http.ResponseWriter) {
	base := &accessLogResponseWriter{ResponseWriter: w}

	mask := 0
	if _, ok := w.(http.Flusher); ok {
		mask |= 1
	}
	if _, ok := w.(http.Hijacker); ok {
		mask |= 2
	}
	if _, ok := w.(http.Pusher); ok {
		mask |= 4
	}
	if _, ok := w.(io.ReaderFrom); ok {
		mask |= 8
	}

	switch mask {
	case 0:
		return base, base
	case 1:
		return base, struct {
			http.ResponseWriter
			http.Flusher
		}{ResponseWriter: base, Flusher: accessLogFlusher{base}}
	case 2:
		return base, struct {
			http.ResponseWriter
			http.Hijacker
		}{ResponseWriter: base, Hijacker: accessLogHijacker{base}}
	case 3:
		return base, struct {
			http.ResponseWriter
			http.Flusher
			http.Hijacker
		}{ResponseWriter: base, Flusher: accessLogFlusher{base}, Hijacker: accessLogHijacker{base}}
	case 4:
		return base, struct {
			http.ResponseWriter
			http.Pusher
		}{ResponseWriter: base, Pusher: accessLogPusher{base}}
	case 5:
		return base, struct {
			http.ResponseWriter
			http.Flusher
			http.Pusher
		}{ResponseWriter: base, Flusher: accessLogFlusher{base}, Pusher: accessLogPusher{base}}
	case 6:
		return base, struct {
			http.ResponseWriter
			http.Hijacker
			http.Pusher
		}{ResponseWriter: base, Hijacker: accessLogHijacker{base}, Pusher: accessLogPusher{base}}
	case 7:
		return base, struct {
			http.ResponseWriter
			http.Flusher
			http.Hijacker
			http.Pusher
		}{ResponseWriter: base, Flusher: accessLogFlusher{base}, Hijacker: accessLogHijacker{base}, Pusher: accessLogPusher{base}}
	case 8:
		return base, struct {
			http.ResponseWriter
			io.ReaderFrom
		}{ResponseWriter: base, ReaderFrom: accessLogReaderFrom{base}}
	case 9:
		return base, struct {
			http.ResponseWriter
			http.Flusher
			io.ReaderFrom
		}{ResponseWriter: base, Flusher: accessLogFlusher{base}, ReaderFrom: accessLogReaderFrom{base}}
	case 10:
		return base, struct {
			http.ResponseWriter
			http.Hijacker
			io.ReaderFrom
		}{ResponseWriter: base, Hijacker: accessLogHijacker{base}, ReaderFrom: accessLogReaderFrom{base}}
	case 11:
		return base, struct {
			http.ResponseWriter
			http.Flusher
			http.Hijacker
			io.ReaderFrom
		}{ResponseWriter: base, Flusher: accessLogFlusher{base}, Hijacker: accessLogHijacker{base}, ReaderFrom: accessLogReaderFrom{base}}
	case 12:
		return base, struct {
			http.ResponseWriter
			http.Pusher
			io.ReaderFrom
		}{ResponseWriter: base, Pusher: accessLogPusher{base}, ReaderFrom: accessLogReaderFrom{base}}
	case 13:
		return base, struct {
			http.ResponseWriter
			http.Flusher
			http.Pusher
			io.ReaderFrom
		}{ResponseWriter: base, Flusher: accessLogFlusher{base}, Pusher: accessLogPusher{base}, ReaderFrom: accessLogReaderFrom{base}}
	case 14:
		return base, struct {
			http.ResponseWriter
			http.Hijacker
			http.Pusher
			io.ReaderFrom
		}{ResponseWriter: base, Hijacker: accessLogHijacker{base}, Pusher: accessLogPusher{base}, ReaderFrom: accessLogReaderFrom{base}}
	default:
		return base, struct {
			http.ResponseWriter
			http.Flusher
			http.Hijacker
			http.Pusher
			io.ReaderFrom
		}{ResponseWriter: base, Flusher: accessLogFlusher{base}, Hijacker: accessLogHijacker{base}, Pusher: accessLogPusher{base}, ReaderFrom: accessLogReaderFrom{base}}
	}
}

func clientIPFromRemoteAddr(remoteAddr string) string {
	remoteAddr = strings.TrimSpace(remoteAddr)
	if remoteAddr == "" {
		return ""
	}

	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return remoteAddr
}
