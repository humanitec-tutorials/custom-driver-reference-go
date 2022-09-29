package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"

	herrors "humanitec.io/custom-reference-driver/internal/errors"
)

var validID = regexp.MustCompile(`^[a-z0-9][a-z0-9-]+[a-z0-9]$`)

func isValidAsID(str string) bool {
	return validID.MatchString(str)
}

func getCredentials(driverSecrets map[string]interface{}) (string, string, error) {
	account, ok := driverSecrets["account"].(map[string]interface{})
	if !ok {
		return "", "", errors.New("driver secrets should contain 'account' map")
	}
	accessKeyId, ok := account["aws_access_key_id"].(string)
	if !ok {
		return "", "", errors.New("'account' details should include 'aws_access_key_id'")
	}
	secretAccessKey, ok := account["aws_secret_access_key"].(string)
	if !ok {
		return "", "", errors.New("'account' details should include 'aws_secret_access_key'")
	}
	return accessKeyId, secretAccessKey, nil
}

// writeAsText Writes the status code and the object (marshled into a string) into the response.
func writeAsText(w http.ResponseWriter, statusCode int, data interface{}) {
	dataText := fmt.Sprintf("%+v", data)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	_, err := w.Write([]byte(dataText))
	if err != nil {
		log.Println(err)
	}
}

// writeAsJSON writes the supplied object to a response along with the status code.
func writeAsJSON(w http.ResponseWriter, statusCode int, obj interface{}) {
	jsonObj, err := json.Marshal(obj)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(jsonObj)
}

// readAsJSON parses request content from JSON
func readAsJSON(r io.Reader, obj interface{}) error {
	if r == nil {
		return errors.New("content stream is not available")
	}
	return json.NewDecoder(r).Decode(obj)
}

// writeError sends the error message as an HTTP Response
func writeError(w http.ResponseWriter, httpStatusCode int, message string, err error) {
	if httpStatusCode <= 0 {
		httpStatusCode = http.StatusInternalServerError
	}

	var humanitecErr *herrors.HumanitecError
	if err == nil || !errors.As(err, &humanitecErr) {
		humanitecErr = herrors.New("RES-001", message, map[string]interface{}{"error": fmt.Sprintf("%v", err)}, err)
	} else {
		if humanitecErr.Details == nil {
			humanitecErr.Details = map[string]interface{}{}
		}
		humanitecErr.Details["error"] = fmt.Sprintf("%v", err)
	}

	log.Printf("(HTTP %d) %v\n", httpStatusCode, humanitecErr)
	writeAsJSON(w, httpStatusCode, humanitecErr)
}
