package apierror_test

import (
	stdcontext "context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pumpingbytes/go-kit/apierror"
	"github.com/pumpingbytes/go-kit/apperror"
	"github.com/pumpingbytes/go-kit/httpmw"
)

func ExampleAPIError_WithContext() {
	ctx := stdcontext.Background()
	ctx = httpmw.PutRequestID(ctx, "req-123")

	err := apierror.ErrInternalServerError.WithContext(httpmw.ErrorContext(ctx))
	b, _ := json.Marshal(err.Cleanup())
	fmt.Println(string(b))
	// Output: {"code":"INTERNAL","message":"internal error","context":{"request.id":"req-123"}}
}

func ExampleFromAppError() {
	err := apperror.Wrap(io.EOF, "USER_NOT_FOUND", "user not found", http.StatusNotFound)
	out, _ := apierror.FromAppError(err)
	fmt.Printf("%d %s %s\n", out.Status, out.Code, out.Message)
	// Output: 404 USER_NOT_FOUND user not found
}

