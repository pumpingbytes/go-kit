package httpmw

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDPreservesIncomingHeaderAndContext(t *testing.T) {
	const requestID = "req-123"

	var gotRequestID string
	handler := WithRequestID(DefaultRequestIDHeader)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequestID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(DefaultRequestIDHeader, "  "+requestID+"  ")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if gotRequestID != requestID {
		t.Fatalf("GetRequestID(context) = %q, want %q", gotRequestID, requestID)
	}
	if got := rr.Header().Get(DefaultRequestIDHeader); got != requestID {
		t.Fatalf("response header %q = %q, want %q", DefaultRequestIDHeader, got, requestID)
	}
}

func TestRequestIDGeneratesMissingHeaderAndSetsContext(t *testing.T) {
	var gotRequestID string
	handler := WithRequestID(DefaultRequestIDHeader)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequestID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if gotRequestID == "" {
		t.Fatal("GetRequestID(context) = empty, want generated request ID")
	}
	if got := rr.Header().Get(DefaultRequestIDHeader); got == "" {
		t.Fatalf("response header %q = empty, want generated request ID", DefaultRequestIDHeader)
	} else if got != gotRequestID {
		t.Fatalf("response header %q = %q, want %q", DefaultRequestIDHeader, got, gotRequestID)
	}
}

func TestRequestIDUsesDefaultHeaderWhenEmptyHeaderNameProvided(t *testing.T) {
	const requestID = "req-default"

	var gotRequestID string
	handler := WithRequestID("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequestID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(DefaultRequestIDHeader, requestID)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if gotRequestID != requestID {
		t.Fatalf("GetRequestID(context) = %q, want %q", gotRequestID, requestID)
	}
	if got := rr.Header().Get(DefaultRequestIDHeader); got != requestID {
		t.Fatalf("response header %q = %q, want %q", DefaultRequestIDHeader, got, requestID)
	}
}

func TestRequestIDSupportsCustomHeader(t *testing.T) {
	const (
		headerName = "X-Correlation-Id"
		requestID  = "corr-123"
	)

	var gotRequestID string
	handler := WithRequestID(headerName)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequestID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(headerName, requestID)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if gotRequestID != requestID {
		t.Fatalf("GetRequestID(context) = %q, want %q", gotRequestID, requestID)
	}
	if got := rr.Header().Get(headerName); got != requestID {
		t.Fatalf("response header %q = %q, want %q", headerName, got, requestID)
	}
	if got := rr.Header().Get(DefaultRequestIDHeader); got != "" {
		t.Fatalf("unexpected default response header %q = %q", DefaultRequestIDHeader, got)
	}
}
