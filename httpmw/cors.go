package httpmw

import (
	"net/http"
	"strings"
)

// CORSOptions defines a simple CORS policy.
//
// AllowedOrigins:
//   - empty => allow all origins (echo back request Origin)
//   - otherwise => only echo origins present in the list
//
// AllowHeaders/AllowMethods are emitted as comma-separated lists.
//
// Note: This is intentionally framework-agnostic (net/http only).
// A framework adapter (gin/echo/chi) can call ApplyCORS and then decide
// how to abort/short-circuit the pipeline on preflight.

type CORSOptions struct {
	AllowedOrigins     []string
	AllowHeaders       []string
	AllowMethods       []string
	AllowCredentials   *bool
	VaryOriginHeader   *bool
	PreflightNoContent *bool
}

func (o CORSOptions) normalized() CORSOptions {
	out := o
	t := true
	if len(out.AllowHeaders) == 0 {
		out.AllowHeaders = []string{"Authorization", "Content-Type"}
	}
	if len(out.AllowMethods) == 0 {
		out.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if out.VaryOriginHeader == nil {
		// default true
		out.VaryOriginHeader = &t
	}
	if out.PreflightNoContent == nil {
		// default true
		out.PreflightNoContent = &t
	}
	if out.AllowCredentials == nil {
		// default true
		out.AllowCredentials = &t
	}
	return out
}

type allowedOriginsSet struct {
	allowAll bool
	allowed  map[string]struct{}
}

func newAllowedOriginsSet(allowedOrigins []string) allowedOriginsSet {
	allowAll := len(allowedOrigins) == 0
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		if s := strings.TrimSpace(o); s != "" {
			allowed[s] = struct{}{}
		}
	}
	return allowedOriginsSet{allowAll: allowAll, allowed: allowed}
}

func (s allowedOriginsSet) isAllowed(origin string) bool {
	if origin == "" {
		return false
	}
	if s.allowAll {
		return true
	}
	_, ok := s.allowed[origin]
	return ok
}

// ApplyCORS applies CORS headers to the response. If the request is a preflight
// (OPTIONS), ApplyCORS writes the appropriate headers and returns (true, statusCode).
//
// The caller is responsible for terminating the request flow when preflight=true.
func ApplyCORS(w http.ResponseWriter, r *http.Request, opts CORSOptions) (preflight bool, statusCode int) {
	opts = opts.normalized()

	origin := r.Header.Get("Origin")
	if origin == "" {
		// Non-CORS request.
		if r.Method == http.MethodOptions {
			return true, http.StatusNoContent
		}
		return false, 0
	}

	allowed := newAllowedOriginsSet(opts.AllowedOrigins)
	if allowed.isAllowed(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		if *opts.VaryOriginHeader {
			w.Header().Set("Vary", "Origin")
		}
		if *opts.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(opts.AllowHeaders, ", "))
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(opts.AllowMethods, ", "))
	}

	if r.Method == http.MethodOptions {
		if *opts.PreflightNoContent {
			return true, http.StatusNoContent
		}
		return true, http.StatusOK
	}
	return false, 0
}
