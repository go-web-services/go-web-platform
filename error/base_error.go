package error

import (
	"net/http"

	"github.com/Lomank123/go-web-platform/types"
)

// BaseError is the general-purpose domain error. It implements HTTPError.
// Status defaults to 400 when zero.
type BaseError struct {
	Code    types.ErrorCode
	Message string
	Status  int
}

func (e *BaseError) Error() string {
	return e.Message
}

func (e *BaseError) HTTPStatus() int {
	if e.Status != 0 {
		return e.Status
	}
	return http.StatusBadRequest
}

func (e *BaseError) HTTPErrorCode() types.ErrorCode {
	return e.Code
}

// NewError creates a BaseError that maps to HTTP 400.
func NewError(code types.ErrorCode, message string) error {
	return &BaseError{Code: code, Message: message}
}

// NewErrorWithStatus creates a BaseError with an explicit HTTP status code.
func NewErrorWithStatus(code types.ErrorCode, message string, status int) error {
	return &BaseError{Code: code, Message: message, Status: status}
}
