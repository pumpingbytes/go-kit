// Package gokit provides small, framework-agnostic helpers for service
// infrastructure and application plumbing.
//
// The module is organized into focused subpackages for application errors,
// API error envelopes, user-safe context payloads, typed context value helpers,
// HTTP middleware, slog setup, database error classification interfaces, and CLI
// streams. Postgres/pgx-specific database classification lives in the separate
// nested module github.com/pumpingbytes/go-kit/dberror/postgres.
package gokit
