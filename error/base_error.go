package error

import (
	"github.com/Lomank123/go-web-platform/types"
)

// BaseError represents an error with a message and a code
type BaseError struct {
	Code    types.ErrorCode
	Message string
}

// Error implements the error interface for BaseError
func (e *BaseError) Error() string {
	return e.Message
}

// NewError creates a new BaseError with the given code and message
func NewError(code types.ErrorCode, message string) error {
	return &BaseError{
		Code:    code,
		Message: message,
	}
}
