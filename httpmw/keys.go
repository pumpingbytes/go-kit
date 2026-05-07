package httpmw

import (
	"github.com/ygrebnov/keys"
)

const separator = '.'

const (
	segmentHTTP    = "http"
	segmentRequest = "request"
	segmentAccess  = "access"
)

var (
	newRequestKey     = keys.Factory(separator, keys.WithSegments(segmentRequest))
	newHTTPRequestKey = keys.Factory(separator, keys.WithSegments(segmentHTTP, segmentRequest))
	newHTTPAccessKey  = keys.Factory(separator, keys.WithSegments(segmentHTTP, segmentAccess))
)

var (
	RequestID = newRequestKey("id")
	SpanID    = keys.New("id", separator, keys.WithSegments("span"))
	TraceID   = keys.New("id", separator, keys.WithSegments("trace"))

	HTTPRequestMethod     = newHTTPRequestKey("method")
	HTTPRequestURL        = newHTTPRequestKey("url")
	HTTPRequestDurationMS = newHTTPRequestKey("duration_ms")
	HTTPRequestError      = newHTTPRequestKey("error")
	HTTPRequestStatus     = newHTTPRequestKey("status")

	HTTPAccessMethod    = newHTTPAccessKey("method")
	HTTPAccessPath      = newHTTPAccessKey("path")
	HTTPAccessStatus    = newHTTPAccessKey("status")
	HTTPAccessLatency   = newHTTPAccessKey("latency")
	HTTPAccessClientIP  = newHTTPAccessKey("client_ip")
	HTTPAccessUserAgent = newHTTPAccessKey("user_agent")
)
