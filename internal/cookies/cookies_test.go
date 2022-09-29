package cookies

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestEncode(t *testing.T) {
	type testStructType struct {
		Prop1 string      `json:"p1"`
		Prop2 interface{} `json:"p2,omitempty"`
	}
	cyclicStruct := testStructType{
		Prop1: "Cyclic struct",
	}
	cyclicStruct.Prop2 = &cyclicStruct

	var tests = []struct {
		Name     string
		Data     interface{}
		Expected string
		Error    error
	}{
		// Success Path
		//
		{
			Name:     "Should handle nil",
			Data:     nil,
			Expected: "",
		},
		{
			Name:     "Should encode simple string",
			Data:     "Test",
			Expected: "IlRlc3Qi",
		},
		{
			Name: "Should encode map",
			Data: map[string]interface{}{
				"key": "value",
			},
			Expected: "eyJrZXkiOiJ2YWx1ZSJ9",
		},
		{
			Name: "Should encode struct",
			Data: testStructType{
				Prop1: "Test struct",
			},
			Expected: "eyJwMSI6IlRlc3Qgc3RydWN0In0=",
		},

		// Errors Handling
		//
		{
			Name:  "Should report JSON marshaling errors",
			Data:  cyclicStruct,
			Error: errors.New("encountered a cycle"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			cookie, err := Encode(tt.Data)

			if tt.Error == nil {
				// On Success
				assert.NilError(t, err, "Unexpected error: %v", err)
				assert.DeepEqual(t, cookie, tt.Expected)
			} else {
				// On Error
				assert.Assert(t, err != nil, "Should return error.")
				assert.Assert(t, strings.Contains(err.Error(), tt.Error.Error()),
					"Error message should contain: '%v'. Actual: '%v'.", tt.Error, err)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	type testStructType struct {
		Prop1 string      `json:"p1"`
		Prop2 interface{} `json:"p2,omitempty"`
	}

	var tests = []struct {
		Name     string
		Cookie   string
		Expected interface{}
		Error    error
	}{
		// Success Path
		//
		{
			Name:     "Should accept empty cookie",
			Cookie:   "",
			Expected: nil,
		},
		{
			Name:     "Should decode simple string",
			Cookie:   "IlRlc3Qi",
			Expected: "Test",
		},
		{
			Name:   "Should decode map",
			Cookie: "eyJrZXkiOiJ2YWx1ZSJ9",
			Expected: map[string]interface{}{
				"key": "value",
			},
		},
		{
			Name:   "Should decode struct",
			Cookie: "eyJwMSI6IlRlc3Qgc3RydWN0In0=",
			Expected: testStructType{
				Prop1: "Test struct",
			},
		},

		// Errors Handling
		//
		{
			Name:   "Should report Base64-decoding errors",
			Cookie: "{ not a base 64 string }",
			Error:  errors.New("illegal base64 data"),
		},
		{
			Name:   "Should report JSON unmarshaling errors",
			Cookie: "YWJjMTIzIT8kKiYoKSctPUB+",
			Error:  errors.New("invalid character"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			var data = interface{}(nil)

			// NOTE: For proper JSON Unmarshalling, ensure data is of the same type as tt.Expected and not just an interface{}
			if tt.Expected != nil {
				data = reflect.New(reflect.TypeOf(tt.Expected)).Interface()
			}

			err := Decode(tt.Cookie, data)

			if tt.Error == nil {
				// On Success
				assert.NilError(t, err, "Unexpected error: %v", err)

				// NOTE: Unwrap data back to the same type as tt.Expected
				if reflect.TypeOf(data) != nil {
					data = reflect.ValueOf(data).Elem().Interface()
				}
				assert.DeepEqual(t, data, tt.Expected)
			} else {
				// On Error
				assert.Assert(t, err != nil, "Should return error.")
				assert.Assert(t, strings.Contains(err.Error(), tt.Error.Error()),
					"Error message should contain: '%v'. Actual: '%v'.", tt.Error, err)
			}
		})
	}
}
