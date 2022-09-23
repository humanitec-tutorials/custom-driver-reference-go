package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func executeTestGet(ctx context.Context, router *mux.Router, url string, headers map[string]string) (*httptest.ResponseRecorder, error) {
	return executeTestRequest(ctx, router, http.MethodGet, url, nil, headers)
}

func executeTestRequest(ctx context.Context, router *mux.Router, method, url string, body interface{}, headers map[string]string) (*httptest.ResponseRecorder, error) {
	var req *http.Request
	var err error
	if body == nil {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	} else {
		if b, ok := body.(*bytes.Buffer); ok {
			req, err = http.NewRequestWithContext(ctx, method, url, b)
		} else if b, ok := body.([]byte); ok {
			req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(b))
		} else {
			var b []byte
			if b, err = json.Marshal(body); err == nil {
				req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(b))
				req.Header.Add("Content-Type", "application/json")
			}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	for hdr, val := range headers {
		req.Header.Add(hdr, val)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w, nil
}

func TestValidIDs(t *testing.T) {
	assert.True(t, isValidAsID("valid-id"))                   // "valid-id" is a valid id
	assert.True(t, isValidAsID("01-valid-id-2"))              // "01-valid-id-2" is a valid id
	assert.True(t, isValidAsID("jahgsdo87iq28ui3hdgkuyqxl3")) // "jahgsdo87iq28ui3hdgkuyqxl3" is a valid id

	assert.False(t, isValidAsID("-invalid-id")) // "-invalid-id" is not a valid id
	assert.False(t, isValidAsID("Invalid ID"))  // "Invalid ID" is not a not avalid id
	assert.False(t, isValidAsID(""))            // "" is a not a valid id
	assert.False(t, isValidAsID("a"))           // "a" is a not a valid id
}

func TestGetUser(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://some.domain.test", nil)
	assert.NoError(t, err)

	_, err = getUser(r)
	assert.ErrorIs(t, err, ErrUnauthorized)

	r.Header.Set("authorization", "")
	_, err = getUser(r)
	assert.ErrorIs(t, err, ErrUnauthorized)

	r.Header.Set("authorization", "Bearer 1234")
	_, err = getUser(r)
	assert.ErrorIs(t, err, ErrUnauthorized)

	r.Header.Set("authorization", "JWT 1234.1234.1234")
	_, err = getUser(r)
	assert.ErrorIs(t, err, ErrUnauthorized)

	r.Header.Set("from", "")
	_, err = getUser(r)
	assert.ErrorIs(t, err, ErrUnauthorized)

	r.Header.Set("from", "user@domain.com")
	userId, err := getUser(r)
	assert.NoError(t, err)
	assert.Equal(t, "user@domain.com", userId)
}
