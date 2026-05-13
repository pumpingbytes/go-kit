package context_test

import (
	stdcontext "context"
	"fmt"

	kitcontext "github.com/pumpingbytes/go-kit/context"
	"github.com/ygrebnov/keys"
)

type tenantKey struct{}

func ExampleCtx() {
	payload := kitcontext.Ctx(
		keys.Key("request.id"), "req-123",
		keys.Key("trace.id"), "trace-abc",
	)
	b, _ := payload.Marshal()
	fmt.Println(string(b))
	// Output: {"request.id":"req-123","trace.id":"trace-abc"}
}

func ExamplePutValue() {
	ctx := kitcontext.PutValue(stdcontext.Background(), tenantKey{}, "tenant-42")
	fmt.Println(kitcontext.GetValue(ctx, tenantKey{}, ""))
	fmt.Println(kitcontext.GetValue(stdcontext.Background(), tenantKey{}, "fallback"))
	// Output:
	// tenant-42
	// fallback
}

