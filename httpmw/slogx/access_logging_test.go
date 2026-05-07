package slogx

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/pumpingbytes/go-kit/httpmw"
)

func TestAccessLoggerIncludesContextAndAccessFields(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logAccess := AccessLogger(logger)

	ctx := context.Background()
	ctx = httpmw.PutRequestID(ctx, "req-1")
	ctx = httpmw.PutTraceIDs(ctx, "trace-1", "span-1")

	logAccess(ctx, httpmw.AccessLogFields{
		Method:    "GET",
		Path:      "/healthz",
		Status:    200,
		Latency:   125 * time.Millisecond,
		ClientIP:  "127.0.0.1",
		UserAgent: "curl/8.0",
	})

	record := decodeSingleRecord(t, buf.String())

	if got := record["level"]; got != "INFO" {
		t.Fatalf("level = %v, want INFO", got)
	}
	if got := record["msg"]; got != "request" {
		t.Fatalf("msg = %v, want request", got)
	}
	if got := record[string(httpmw.RequestID)]; got != "req-1" {
		t.Fatalf("request_id = %v, want req-1", got)
	}
	if got := record[string(httpmw.TraceID)]; got != "trace-1" {
		t.Fatalf("trace_id = %v, want trace-1", got)
	}
	if got := record[string(httpmw.SpanID)]; got != "span-1" {
		t.Fatalf("span_id = %v, want span-1", got)
	}
	if got := record[string(httpmw.HTTPAccessMethod)]; got != "GET" {
		t.Fatalf("method = %v, want GET", got)
	}
	if got := record[string(httpmw.HTTPAccessPath)]; got != "/healthz" {
		t.Fatalf("path = %v, want /healthz", got)
	}
	if got := record[string(httpmw.HTTPAccessStatus)]; got != float64(200) {
		t.Fatalf("status = %v, want 200", got)
	}
	if got := record[string(httpmw.HTTPAccessClientIP)]; got != "127.0.0.1" {
		t.Fatalf("client_ip = %v, want 127.0.0.1", got)
	}
	if got := record[string(httpmw.HTTPAccessUserAgent)]; got != "curl/8.0" {
		t.Fatalf("user_agent = %v, want curl/8.0", got)
	}
	if _, ok := record[string(httpmw.HTTPAccessLatency)]; !ok {
		t.Fatalf("missing latency field")
	}
}

func TestAccessLoggerUsesErrorLevelForServerErrors(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logAccess := AccessLogger(logger)

	logAccess(context.Background(), httpmw.AccessLogFields{
		Method:   "POST",
		Path:     "/fail",
		Status:   500,
		Latency:  time.Second,
		ClientIP: "10.0.0.1",
	})

	record := decodeSingleRecord(t, buf.String())
	if got := record["level"]; got != "ERROR" {
		t.Fatalf("level = %v, want ERROR", got)
	}
}

func decodeSingleRecord(t *testing.T, out string) map[string]any {
	t.Helper()

	line := strings.TrimSpace(out)
	if line == "" {
		t.Fatal("expected a log record, got empty output")
	}

	var record map[string]any
	if err := json.Unmarshal([]byte(line), &record); err != nil {
		t.Fatalf("unmarshal log record: %v\noutput: %s", err, line)
	}
	return record
}
