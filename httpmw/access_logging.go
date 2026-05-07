package httpmw

import (
	"context"
	"time"
)

// AccessLogFields defines the standard access log schema.
type AccessLogFields struct {
	Method    string
	Path      string
	Status    int
	Latency   time.Duration
	ClientIP  string
	UserAgent string
}

// AccessLogger emits a single access-log record for the provided request context.
type AccessLogger func(ctx context.Context, f AccessLogFields)

