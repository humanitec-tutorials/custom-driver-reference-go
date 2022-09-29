package cookies

import (
	"encoding/base64"
	"encoding/json"
)

// Encode Marshals data into JSON and Base64-encodes the result.
func Encode(data interface{}) (string, error) {
	if data == nil {
		return "", nil
	}

	bytes, err := json.Marshal(data)
	return base64.StdEncoding.EncodeToString(bytes), err
}

// Decodes Base64-decodes the cookie and unmarshals the data from JSON.
func Decode(cookie string, data interface{}) error {
	if cookie == "" {
		return nil
	}

	if bytes, err := base64.StdEncoding.DecodeString(cookie); err != nil {
		return err
	} else {
		return json.Unmarshal(bytes, data)
	}
}
