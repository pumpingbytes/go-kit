// Package context provides two small, related building blocks:
//
//   - Context: a user-safe, serializable metadata payload for errors and responses
//   - PutValue/GetValue: tiny generic helpers for storing typed values in stdlib context.Context
//
// The serializable Context type is intended for safe metadata such as request and trace
// identifiers. It is not a replacement for stdlib context.Context cancellation, deadlines,
// or internal control-flow data.
//
// In packages that need request-scoped runtime values, prefer package-specific wrappers
// such as httpmw.PutRequestID, log/slogx.PutLogger, or streams.PutIOStreams when
// available. They keep call sites self-documenting while sharing the same typed helper
// behavior.
package context


