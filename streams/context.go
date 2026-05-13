package streams

import (
	"context"

	kitcontext "github.com/pumpingbytes/go-kit/context"
)

type ioStreamsCtxKey struct{}

// PutIOStreams returns a new context with the provided IO streams.
func PutIOStreams(ctx context.Context, s IOStreams) context.Context {
	return kitcontext.PutValue(ctx, ioStreamsCtxKey{}, s)
}

// GetIOStreams returns the IO streams from the context, falling back to stdio.
func GetIOStreams(ctx context.Context) IOStreams {
	return kitcontext.GetValue[IOStreams](ctx, ioStreamsCtxKey{}, NewDefault())
}
