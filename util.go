package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
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

// Encode Marshals data into JSON and Base64-encodes the result.
func cookiesEncode(data interface{}) (string, error) {
	if data == nil {
		return "", nil
	}

	bytes, err := json.Marshal(data)
	return base64.StdEncoding.EncodeToString(bytes), err
}

// Decodes Base64-decodes the cookie and unmarshals the data from JSON.
func cookiesDecode(cookie string, data interface{}) error {
	if cookie == "" {
		return nil
	}

	if bytes, err := base64.StdEncoding.DecodeString(cookie); err != nil {
		return err
	} else {
		return json.Unmarshal(bytes, data)
	}
}
