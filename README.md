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
| `apperror` | Transport-agnostic application/service error type with stable code, safe message, suggested status, and wrapped cause |
| `apierror` | Standardized JSON error envelope with stable codes, HTTP status, client-safe context, and internal debug payloads |
| `context` | User-safe serializable metadata payloads plus small generic helpers for typed `context.Context` values |
| `dberror` | Small interfaces for database error classification |
| `dberror/postgres` | Postgres/pgx classifier implementation for `dberror.Classifier` |
| `httpmw` | `net/http` helpers for request correlation, CORS support, and access log field definitions |
| `httpmw/slogx` | `slog` adapters for request context enrichment and HTTP access logging |
| `log` | Output configuration and writer setup for application logging |
| `log/slogx` | Shared `slog.Logger` loader plus request-scoped logger context helpers |
| `streams` | User-facing input/output stream abstractions backed by stdio, writers, discard sinks, and buffers |
| `streams/slogx` | `slog` adapter for `streams.IOStreams` |

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

### Error sink examples

For transport layers, a good pattern is to normalize everything into `apierror.APIError` before writing the response.

#### Plain `net/http`

```go
package main

import (
    "encoding/json"
    "log/slog"
    "net/http"

    "github.com/pumpingbytes/go-kit/apierror"
    "github.com/pumpingbytes/go-kit/httpmw"
)

func writeAPIError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
    if err == nil {
        return
    }

    var out *apierror.APIError

    if ae, ok := apierror.As(err); ok {
        out = ae
        if out.Status == 0 {
            out = out.WithStatus(http.StatusBadRequest)
        }
    } else if ae, ok := apierror.FromAppError(err); ok {
        out = ae
        if out.Status == 0 {
            out = out.WithStatus(http.StatusInternalServerError)
        }
    } else {
        out = apierror.ErrInternalServerError
    }

    if ctx := httpmw.ErrorContext(r.Context()); ctx != nil {
        out = out.WithContext(ctx)
    }

    logger.Error("request failed",
        slog.String("path", r.URL.Path),
        slog.String("error", err.Error()),
    )

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(out.Status)
    _ = json.NewEncoder(w).Encode(out.Cleanup())
}
```

#### Gin

```go
package api

import (
    "log/slog"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/pumpingbytes/go-kit/apierror"
    "github.com/pumpingbytes/go-kit/httpmw"
)

func handleError(ctx *gin.Context, logger *slog.Logger, err error) {
    if err == nil {
        return
    }

    var out *apierror.APIError

    if ae, ok := apierror.As(err); ok {
        out = ae
        if out.Status == 0 {
            out = out.WithStatus(http.StatusBadRequest)
        }
    } else if ae, ok := apierror.FromAppError(err); ok {
        out = ae
        if out.Status == 0 {
            out = out.WithStatus(http.StatusInternalServerError)
        }
    } else {
        out = apierror.ErrInternalServerError
    }

    if reqCtx := httpmw.ErrorContext(ctx.Request.Context()); reqCtx != nil {
        out = out.WithContext(reqCtx)
    }

    logger.Error("request failed",
        slog.String("path", ctx.FullPath()),
        slog.String("error", err.Error()),
    )

    ctx.AbortWithStatusJSON(out.Status, out.Cleanup())
}
```

These examples intentionally keep policy small and explicit:

- `apierror.As(err)` wins when the error is already transport-ready
- `apierror.FromAppError(err)` adapts `apperror.Error`
- unknown errors fall back to `apierror.ErrInternalServerError`

### Request ID propagation

```go
package main

import (
    "net/http"

    "github.com/pumpingbytes/go-kit/httpmw"
)

func middleware(next http.Handler) http.Handler {
    return httpmw.WithRequestID(httpmw.DefaultRequestIDHeader)(next)
}
```

If the incoming request does not include `X-Request-Id`, a new random hex ID is generated.

### Typed `context.Context` values

The `context` package also provides small generic helpers for storing typed values in
stdlib `context.Context`. Package-level wrappers like `httpmw.PutRequestID(...)`,
`logslogx.PutLogger(...)`, and `streams.PutIOStreams(...)` build on these helpers.

```go
package main

import (
    "context"
    "fmt"

    kitcontext "github.com/pumpingbytes/go-kit/context"
)

type tenantKey struct{}

func main() {
    ctx := kitcontext.PutValue(context.Background(), tenantKey{}, "tenant-42")
    tenantID := kitcontext.GetValue(ctx, tenantKey{}, "")
    fmt.Println(tenantID)
    // Output: tenant-42
}
```

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
        Path:      "/var/log/my-service/service.log",
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

- if `Options.Path` is set, logs go to that exact file path
- else if `Options.Dir` is empty, logs go to `stderr`
- otherwise logs go to `<Dir>/log/<File>`
- if `File` is empty, the default file is `<appName>.log`
- log level can be overridden with `<APPNAME>_LOG_LEVEL`

### CLI streams with `slog`

```go
package main

import (
    "fmt"
    "log/slog"

    streamsslogx "github.com/pumpingbytes/go-kit/streams/slogx"
)

func main() {
    logger := slog.Default()
    streams := streamsslogx.New(logger, slog.LevelInfo, slog.LevelError)

    _, _ = fmt.Fprintln(streams.Out(), "operation started")
    _, _ = fmt.Fprintln(streams.ErrOut(), "validation failed")
}
```

### Request-scoped logging

```go
package main

import (
    "context"
    "log/slog"
    "net/http"

    "github.com/pumpingbytes/go-kit/httpmw"
    httpslogx "github.com/pumpingbytes/go-kit/httpmw/slogx"
    logslogx "github.com/pumpingbytes/go-kit/log/slogx"
)

func main() {
    base := slog.Default()
    appCtx := logslogx.PutLogger(context.Background(), base)

    accessLogger := httpslogx.AccessLoggerCtxWithOptions(appCtx, httpslogx.AccessLoggerOptions{
        LevelPolicy: func(f httpmw.AccessLogFields) slog.Level {
            switch {
            case f.Status >= http.StatusInternalServerError:
                return slog.LevelError
            case f.Status >= http.StatusBadRequest:
                return slog.LevelWarn
            default:
                return slog.LevelInfo
            }
        },
    })

    handler := httpmw.WithRequestID(httpmw.DefaultRequestIDHeader)(
        httpmw.WithAccessLogOptions(httpmw.AccessLogOptions{
            Logger: accessLogger,
            Skip:   httpmw.SkipPaths("/healthz", "/readyz"),
        })(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := logslogx.PutLogger(r.Context(), httpslogx.EnrichLogger(r.Context(), base))
            r = r.WithContext(ctx)

            logslogx.GetLogger(r.Context()).Info("handling request")
            _, _ = w.Write([]byte("ok\n"))
        })),
    )

    _ = handler
}
```

`httpmw/slogx.AccessLogger` and `AccessLoggerWithOptions` emit a single access log entry using the standard access-log field names from `httpmw`.

The primary filter API is `Skip func(*http.Request) bool`, which keeps the middleware flexible across transports and routing setups. `httpmw.SkipPaths(...)` is a convenience helper for exact path matches.

If you want custom slog levels for access logs, use `httpslogx.AccessLoggerWithOptions(...)` with a `LevelPolicy` function as shown above.

Rule of thumb:

- for plain `net/http`, use `WithRequestID(...)` and `WithAccessLogOptions(...)` directly
- for Gin and similar frameworks, use `EnsureRequestID(...)`, then populate `httpmw.AccessLogFields` manually from framework-specific request/response data

### Gin integration

```go
package middleware

import (
    "log/slog"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/pumpingbytes/go-kit/httpmw"
    httpslogx "github.com/pumpingbytes/go-kit/httpmw/slogx"
    logslogx "github.com/pumpingbytes/go-kit/log/slogx"
)

const RequestIDHeader = httpmw.DefaultRequestIDHeader

func RequestLogging(base *slog.Logger) gin.HandlerFunc {
	logAccess := httpslogx.AccessLoggerWithOptions(base, httpslogx.AccessLoggerOptions{
		LevelPolicy: func(f httpmw.AccessLogFields) slog.Level {
			switch {
			case f.Status >= 500:
				return slog.LevelError
			case f.Status >= 400:
				return slog.LevelWarn
			default:
				return slog.LevelInfo
			}
		},
	})

    return func(c *gin.Context) {
        start := time.Now()

        ctx, rid := httpmw.EnsureRequestID(c.Request.Context(), c.Request, RequestIDHeader)
        c.Header(RequestIDHeader, rid)

        ctx = logslogx.PutLogger(ctx, httpslogx.EnrichLogger(ctx, base))
        c.Request = c.Request.WithContext(ctx)

        c.Next()

        logAccess(c.Request.Context(), httpmw.AccessLogFields{
            Method:        c.Request.Method,
            Path:          c.FullPath(),
            Status:        c.Writer.Status(),
            Latency:       time.Since(start),
            ClientIP:      c.ClientIP(),
            ResponseBytes: int64(c.Writer.Size()),
            UserAgent:     c.Request.UserAgent(),
        })
    }
}
```

For Gin and other frameworks, the current `httpmw` building blocks are usually the best fit: use `EnsureRequestID(...)`, populate `httpmw.AccessLogFields`, and emit through `httpmw/slogx`.

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
- `FromAppError(err)` conservatively adapts `apperror.Error` without exposing wrapped causes or internal debug data

### `apperror`

- `New(code, message, status)` creates a transport-agnostic application error
- `Wrap(err, code, message, status)` preserves an upstream cause while attaching application-level metadata
- `As(err)` unwraps an `error` into `*apperror.Error`

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

The `context` package serves two related purposes:

- `Context` is a small serializable map intended for client-safe metadata
- `PutValue(...)` / `GetValue(...)` are tiny generic helpers for typed values in stdlib `context.Context`

For examples that import both packages, alias this package to avoid confusion with the
stdlib package name.

```go
import (
    stdcontext "context"

    kitcontext "github.com/pumpingbytes/go-kit/context"
)

payload := kitcontext.Ctx(httpmw.RequestID, "req-123", httpmw.TraceID, "trace-abc")
ctx := kitcontext.PutValue(stdcontext.Background(), requestKey{}, "req-123")
```

Keys use `github.com/ygrebnov/keys` for consistency.

### `dberror`

- `Classifier` defines a tiny interface for database-specific error classification
- `dberror/postgres.Classifier` provides a pgx/Postgres implementation that can be reused from repositories without coupling service code to transport-layer errors

### `httpmw`

- `WithRequestID(headerName)` is the plain `net/http` request ID middleware; an empty header name uses `DefaultRequestIDHeader`
- `PutRequestID(...)`, `GetRequestID(...)`, `PutTraceIDs(...)`, `GetTraceID(...)`, and `GetSpanID(...)` expose request-correlation values directly through `context.Context`
- `WithAccessLogOptions(...)` wraps `net/http` handlers and emits `AccessLogFields` after the response is written
- `SkipPaths(...)` is an exact-path helper; for anything more advanced, prefer `Skip func(*http.Request) bool`

### `httpmw/slogx`

- `AccessLogger(...)` keeps the default access-log level policy: `ERROR` for `>= 500`, `INFO` otherwise
- `AccessLoggerWithOptions(...)` lets you customize slog level selection with `LevelPolicy`
- `EnrichLogger(...)` adds request correlation fields from context to a base logger

### `log`

- `Options.Path` writes to an exact file path and takes precedence over `Dir` and `File`
- when `Path` is empty and `Dir` is set, logs go to `<Dir>/log/<File-or-app.log>`
- when both `Path` and `Dir` are empty, logging falls back to `stderr`

### `log/slogx`

- `Load(...)` builds a `*slog.Logger` from `log.Options`
- `PutLogger(...)` and `GetLogger(...)` propagate request-scoped loggers via `context.Context`
- `GetLogger(...)` falls back to a default stderr JSON logger when no logger is stored in context

### `streams`

- `NewBasic(in, out, errOut)` builds a basic stream set explicitly
- `NewDefault()` uses `os.Stdin`, `os.Stdout`, and `os.Stderr`
- `NewSilent()` disables both output streams while keeping stdin
- `NewBuffers()` and `NewThreadSafeBuffers()` cover common CLI/test scenarios
- `PutIOStreams(...)` and `GetIOStreams(...)` propagate stream sets via `context.Context`

### `streams/slogx`

- `New(logger, outLevel, errLevel)` creates slog-backed streams
- `NewCtx(ctx, outLevel, errLevel)` uses the logger from `log/slogx.GetLogger(ctx)`

## Development

Run the test suite:

```bash
go test ./...
```

## Status

This repository currently focuses on low-level helpers rather than a full application framework. The packages are small and composable so they can be embedded into HTTP services, background workers, and other Go applications without adopting a larger runtime stack.

## License

This project is licensed under the BSD-3-Clause License. See `LICENSE`.

