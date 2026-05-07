# go-kit

Small, framework-agnostic Go helpers for service infrastructure and application plumbing.

This repository provides reusable building blocks for:

- API-friendly error payloads
- user-safe structured context
- request correlation helpers
- `net/http` middleware helpers, including CORS and access logging
- `log/slog` setup and request-scoped logger enrichment

The module is intentionally lightweight and stays close to the standard library.

## Requirements

- Go `1.25+`

## Installation

```bash
go get github.com/pumpingbytes/go-kit
```

## Packages

| Package | Purpose |
| --- | --- |
| `apierror` | Standardized JSON error envelope with stable codes, HTTP status, client-safe context, and internal debug payloads |
| `context` | Small serializable context map for user-safe metadata attached to errors or responses |
| `httpmw` | `net/http` helpers for request correlation, CORS support, and access log field definitions |
| `httpmw/slogx` | `slog` adapters for request context enrichment and HTTP access logging |
| `log` | Output configuration and writer setup for application logging |
| `log/slogx` | Shared `slog.Logger` loader backed by `log.Options` |

## Quick start

### API errors with request correlation

```go
package main

import (
    "context"
    "fmt"

    "github.com/pumpingbytes/go-kit/apierror"
    "github.com/pumpingbytes/go-kit/httpmw"
)

func main() {
    ctx := context.Background()
    ctx = httpmw.PutRequestID(ctx, "req-123")
    ctx = httpmw.PutTraceIDs(ctx, "trace-abc", "span-def")

    err := apierror.ErrInternalServerError.WithContext(httpmw.ErrorContext(ctx))
    fmt.Println(err.Error())
    // Output: INTERNAL: internal error
}
```

The serialized error shape is:

```json
{
  "code": "INTERNAL",
  "message": "internal error",
  "context": {
    "request.id": "req-123",
    "trace.id": "trace-abc",
    "span.id": "span-def"
  }
}
```

Use `Debug` only for internal diagnostics and call `Cleanup()` before returning an error to clients if needed.

### Request ID propagation

```go
package main

import (
    "net/http"

    "github.com/pumpingbytes/go-kit/httpmw"
)

func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx, requestID := httpmw.EnsureRequestID(r.Context(), r, httpmw.DefaultRequestIDHeader)
        w.Header().Set(httpmw.DefaultRequestIDHeader, requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

If the incoming request does not include `X-Request-Id`, a new random hex ID is generated.

### `slog` logger setup

```go
package main

import (
    "log/slog"

    kitlog "github.com/pumpingbytes/go-kit/log"
    logslogx "github.com/pumpingbytes/go-kit/log/slogx"
)

func main() {
    logger, closer, err := logslogx.Load(kitlog.Options{
        Dir:       "/var/my-service",
        Format:    kitlog.FormatJSON,
        Level:     "info",
        AddSource: false,
    }, "my-service")
    if err != nil {
        // Load falls back to stderr and still returns a usable logger.
        logger.Warn("log output fallback", slog.String("error", err.Error()))
    }
    defer closer.Close()

    logger.Info("service started")
}
```

Behavior summary:

- if `Options.Dir` is empty, logs go to `stderr`
- otherwise logs go to `<Dir>/log/<File>`
- if `File` is empty, the default file is `<appName>.log`
- log level can be overridden with `<APPNAME>_LOG_LEVEL`

### Request-scoped logging

```go
package main

import (
    "context"
    "log/slog"

    "github.com/pumpingbytes/go-kit/httpmw"
    httpslogx "github.com/pumpingbytes/go-kit/httpmw/slogx"
)

func main() {
    base := slog.Default()

    ctx := context.Background()
    ctx = httpmw.PutRequestID(ctx, "req-123")
    ctx = httpmw.PutTraceIDs(ctx, "trace-abc", "span-def")

    logger := httpslogx.EnrichLogger(ctx, base)
    logger.Info("handling request")
}
```

`httpmw/slogx.AccessLogger` emits a single access log entry using the standard access-log field names from `httpmw`.

### CORS helper

```go
package main

import (
    "net/http"

    "github.com/pumpingbytes/go-kit/httpmw"
)

func withCORS(next http.Handler) http.Handler {
    opts := httpmw.CORSOptions{
        AllowedOrigins: []string{"https://example.com"},
        AllowHeaders:   []string{"Authorization", "Content-Type"},
        AllowMethods:   []string{"GET", "POST", "OPTIONS"},
    }

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        preflight, status := httpmw.ApplyCORS(w, r, opts)
        if preflight {
            w.WriteHeader(status)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

`ApplyCORS` is framework-agnostic and works directly with `net/http`.

## API notes

### `apierror`

- `New(code, message)` creates a base `APIError`
- `WithStatus(status)` clones with HTTP status
- `WithContext(ctx)` attaches client-safe serialized context
- `WithDebug(ctx)` attaches internal-only serialized debug payload
- `Cleanup()` removes the debug payload
- `As(err)` unwraps an `error` into `*APIError`

Common predeclared errors include:

- `ErrInternalServerError`
- `ErrServiceUnavailable`
- `ErrMissingAuthorizationHeader`
- `ErrAuthorizationRequired`
- `ErrInvalidToken`
- `ErrMissingRequiredScope`
- `ErrTooManyRequests`
- `ErrEntitlementRequired`

### `context`

The `context` package is a small serializable map intended for safe metadata, not for request cancellation or deadlines.

```go
ctx := context.Ctx(httpmw.RequestID, "req-123", httpmw.TraceID, "trace-abc")
```

Keys use `github.com/ygrebnov/keys` for consistency.

## Development

Run the test suite:

```bash
go test ./...
```

## Status

This repository currently focuses on low-level helpers rather than a full application framework. The packages are small and composable so they can be embedded into HTTP services, background workers, and other Go applications without adopting a larger runtime stack.

## License

This project is licensed under the BSD-3-Clause License. See `LICENSE`.

