package httpjson

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type handlerContextKey struct{}

// SPEC section 7 defines exact success codes. Handler must also forward the
// incoming context (including cancellation/authentication middleware values).
func TestHandlerSuccessContract(t *testing.T) {
	for _, tc := range []struct {
		method, path, contentType string
		status                    int
	}{
		{http.MethodPost, "/v0/listing", "", http.StatusNoContent},
		{http.MethodPost, "/v0/heartbeat", "", http.StatusNoContent},
		{http.MethodPost, "/v0/decline", "", http.StatusNoContent},
		{http.MethodPost, "/v0/report", "", http.StatusNoContent},
		{http.MethodGet, "/v0/jobs?node_id=contract-node", "application/json", http.StatusOK},
		{http.MethodGet, "/v0/fees", "application/json", http.StatusOK},
	} {
		t.Run(tc.path, func(t *testing.T) {
			c := &testCoordinator{}
			ctx := context.WithValue(context.Background(), handlerContextKey{}, tc.path)
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader("{}")).WithContext(ctx)
			w := httptest.NewRecorder()
			Handler(c).ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d", w.Code, tc.status)
			}
			if tc.contentType != "" && w.Header().Get("Content-Type") != tc.contentType {
				t.Errorf("Content-Type = %q, want %q", w.Header().Get("Content-Type"), tc.contentType)
			}
			if len(c.contexts) != 1 || c.contexts[0].Value(handlerContextKey{}) != tc.path {
				t.Error("incoming context marker was not forwarded to coordinator")
			}
		})
	}
}
