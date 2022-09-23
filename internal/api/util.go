package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/humanitec/golib/hlogger"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	humerr "humanitec.io/go-service-template/internal/errors"
)

// Unauthorized request
var ErrUnauthorized = errors.New("unauthorized")

var validID = regexp.MustCompile(`^[a-z0-9][a-z0-9-]+[a-z0-9]$`)

func isValidAsID(str string) bool {
	return validID.MatchString(str)
}

// getUser extracts user identifier from the HTTP "From" header.
// Returns ErrUnauthorized if the HTTP "From" header is not set or empty.
func getUser(r *http.Request) (string, error) {
	auth := r.Header.Get("from")
	if auth == "" {
		return "", ErrUnauthorized
	}

	return auth, nil
}

// writeAsText Writes the status code and the object (marshaled into a string) into the response.
func (s *apiServer) writeAsText(_ context.Context, w http.ResponseWriter, statusCode int, data interface{}) {
	dataText := fmt.Sprintf("%+v", data)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte(dataText))
}

// writeAsJSON Writes the status code and the object (marshaled into a JSON string) into the response.
func (s *apiServer) writeAsJSON(ctx context.Context, w http.ResponseWriter, statusCode int, obj interface{}) {
	dataJSON, err := json.Marshal(obj)
	if err != nil {
		details := hlogger.LogDetails(ctx, "err", err.Error())

		s.logger.Logger.Sugar().Errorw("failed marshalling json", details...)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(dataJSON)
}

// writeError sends the error message as an HTTP Response
func (s *apiServer) writeError(ctx context.Context, w http.ResponseWriter, httpStatusCode int, message string, err error) {
	if httpStatusCode <= 0 {
		httpStatusCode = http.StatusInternalServerError
	}

	var humanitecErr *humerr.HumanitecError
	if !errors.As(err, &humanitecErr) {
		humanitecErr = humerr.New("API-001", message, nil, err)
	}

	span, _ := tracer.StartSpanFromContext(ctx, "writeError")
	span.Finish(tracer.WithError(err))

	details := hlogger.LogDetails(ctx, "err", humanitecErr.Error(), "status", httpStatusCode)
	s.logger.Logger.Sugar().Errorw("http error", details...)
	s.writeAsJSON(ctx, w, httpStatusCode, humanitecErr)
}

// readAsJSON parses request content from JSON
func readAsJSON(r io.Reader, obj interface{}) error {
	if r == nil {
		return errors.New("content stream is not available")
	}
	return json.NewDecoder(r).Decode(obj)
}
