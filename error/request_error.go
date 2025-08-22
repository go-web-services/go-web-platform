package error

import (
	"fmt"

	"github.com/Lomank123/go-web-platform/constants"
)

// RequestError represents an error with a message and a code
// Used when we need to return an error from service to gateway
type RequestError struct {
	BaseError
	StatusCode int            `json:"status_code"`
	ErrorCode  string         `json:"error_code"`
	Payload    map[string]any `json:"errors,omitempty"`
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("Code %s: %s. Status code: %d", e.Code, e.Message, e.StatusCode)
}

// NewRequestError creates a new RequestError with
// the given status code, message and payload
func NewRequestError(statusCode int, payload map[string]any) error {
	var errorCode string
	if code, ok := payload["error_code"]; ok {
		if strCode, ok := code.(string); ok {
			errorCode = strCode
		}
	}

	return &RequestError{
		BaseError: BaseError{
			Code:    constants.RequestError,
			Message: "Request failed",
		},
		StatusCode: statusCode,
		ErrorCode:  errorCode,
		Payload:    payload,
	}
}
