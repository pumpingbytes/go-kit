package httpmw

import (
	"github.com/ygrebnov/keys"
)

const (
	segmentHTTP    = "http"
	segmentRequest = "request"
	segmentAccess  = "access"
)

var (
	newRequestKey     = keys.Factory(keys.WithSegments(segmentRequest))
	newHTTPRequestKey = keys.Factory(keys.WithSegments(segmentHTTP, segmentRequest))
	newHTTPAccessKey  = keys.Factory(keys.WithSegments(segmentHTTP, segmentAccess))
)

var (
	RequestID = newRequestKey("id")
	SpanID    = keys.New("id", keys.WithSegments("span"))
	TraceID   = keys.New("id", keys.WithSegments("trace"))

	HTTPRequestMethod     = newHTTPRequestKey("method")
	HTTPRequestURL        = newHTTPRequestKey("url")
	HTTPRequestDurationMS = newHTTPRequestKey("duration_ms")
	HTTPRequestError      = newHTTPRequestKey("error")
	HTTPRequestStatus     = newHTTPRequestKey("status")

	HTTPAccessMethod        = newHTTPAccessKey("method")
	HTTPAccessPath          = newHTTPAccessKey("path")
	HTTPAccessStatus        = newHTTPAccessKey("status")
	HTTPAccessLatency       = newHTTPAccessKey("latency")
	HTTPAccessClientIP      = newHTTPAccessKey("client_ip")
	HTTPAccessResponseBytes = newHTTPAccessKey("response_bytes")
	HTTPAccessUserAgent     = newHTTPAccessKey("user_agent")
)
