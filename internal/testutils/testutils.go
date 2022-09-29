package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func ExecuteTestRequest(ctx context.Context, t *testing.T, router *mux.Router, method, url string, headers map[string]string, body interface{}) *httptest.ResponseRecorder {
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

	for key, val := range headers {
		req.Header.Add(key, val)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}
