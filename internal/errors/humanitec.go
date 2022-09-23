// TODO: extract this to a separate library
package errors

// HumanitecError represents a standard Humanitec Error
type HumanitecError struct {
	Err     error                  `json:"-"`
	Code    string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error allows the HumanitecError struct to be used as an error object.
func (e *HumanitecError) Error() string {
	errorText := e.Code + ": " + e.Message
	if nil != e.Err {
		errorText += ": " + e.Err.Error()
	}
	return errorText
}

// Unwrap allows the HumanitecError struct to be used with go built in errors.Is() and errors.As()
func (e *HumanitecError) Unwrap() error {
	return e.Err
}

// New wraps an error and creates a new Humanitec Error
func New(code, message string, details map[string]interface{}, err error) *HumanitecError {
	return &HumanitecError{
		Err:     err,
		Code:    code,
		Message: message,
		Details: details,
	}
}
