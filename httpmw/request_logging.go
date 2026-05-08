package httpmw

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"
)

const DefaultRequestIDHeader = "X-Request-Id"

// NewRequestID returns a new random request id (hex string).
func NewRequestID() string {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return hex.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(b[:])
}

// EnsureRequestID extracts (or generates) request id and returns an updated context.
// It also returns the request id value to set in response headers.
func EnsureRequestID(ctx context.Context, req *http.Request, headerName string) (context.Context, string) {
	if headerName == "" {
		headerName = DefaultRequestIDHeader
	}

	rid := strings.TrimSpace(req.Header.Get(headerName))
	if rid == "" {
		rid = NewRequestID()
	}

	ctx = PutRequestID(ctx, rid)
	return ctx, rid
}

// WithRequestID ensures each request has a request ID, stores it in the request
// context, and writes it back to the response header.
//
// If headerName is empty, DefaultRequestIDHeader is used.
func WithRequestID(headerName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			effectiveHeaderName := headerName
			if effectiveHeaderName == "" {
				effectiveHeaderName = DefaultRequestIDHeader
			}

			ctx, rid := EnsureRequestID(r.Context(), r, effectiveHeaderName)
			w.Header().Set(effectiveHeaderName, rid)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
