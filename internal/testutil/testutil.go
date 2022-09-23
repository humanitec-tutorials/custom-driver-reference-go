package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

// ExecuteTestRequest Makes a test request, returns response recorder.
func ExecuteTestRequest(t *testing.T, ctx context.Context, router *mux.Router, method, url, authHeader string, body interface{}, queryParams interface{}) *httptest.ResponseRecorder {
	var req *http.Request
	var err error
	if body == nil {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	} else {
		if b, ok := body.([]byte); ok {
			req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(b))
		} else if b, ok := body.(*bytes.Buffer); ok {
			req, err = http.NewRequestWithContext(ctx, method, url, b)
		} else {
			b, marshalErr := json.Marshal(body)
			if marshalErr != nil {
				t.Errorf("creating request: marshaling body to JSON: %v", marshalErr)
			}
			req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(b))
		}
	}
	if err != nil {
		t.Errorf("creating request: %v", err)
	}

	if authHeader != "" {
		req.Header.Add("From", authHeader)
	}

	if queryParams != nil {
		q := req.URL.Query()
		var encoder = schema.NewEncoder()
		encoder.Encode(queryParams, q)
		req.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}
