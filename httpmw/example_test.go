package httpmw_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/pumpingbytes/go-kit/httpmw"
)

func ExampleWithRequestID() {
	handler := httpmw.WithRequestID(httpmw.DefaultRequestIDHeader)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, httpmw.GetRequestID(r.Context()))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(httpmw.DefaultRequestIDHeader, "req-123")
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	fmt.Println(resp.Header().Get(httpmw.DefaultRequestIDHeader))
	fmt.Println(resp.Body.String())
	// Output:
	// req-123
	// req-123
}

