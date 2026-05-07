package context

import (
	"encoding/json"

	"github.com/ygrebnov/keys"
)

// Context is a user-safe, serializable context payload for objects like apierror.APIError.
// It is not intended for internal use or control flow,
// so it is not a full replacement for context.Context.
// Prefer building it with Ctx for concise call sites.
//
// Keys should be safe for client exposure.
type Context map[keys.Key]any

// Ctx builds a Context from key/value pairs. Odd or non-keys.Key keys are ignored.
// This helper keeps call sites concise.
func Ctx(key keys.Key, value any, rest ...any) Context {
	ctx := Context{key: value}
	for i := 0; i+1 < len(rest); i += 2 {
		k, ok := rest[i].(keys.Key)
		if !ok {
			continue
		}
		ctx[k] = rest[i+1]
	}
	return ctx
}

// Marshal returns the JSON representation of the context payload.
func (c Context) Marshal() (json.RawMessage, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

