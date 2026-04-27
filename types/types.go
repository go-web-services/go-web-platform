package types

import "github.com/Lomank123/go-web-platform/dto"

type ErrorCode string
type Environment string

type ErrorHandlerPayload struct {
	StatusCode int          `json:"status_code"`
	Payload    dto.ErrorDTO `json:"payload,omitempty"`
}

// Pagination

type PaginationInputParams struct {
	Page  int64 `json:"page,omitempty"`
	Limit int64 `json:"limit,omitempty"`
}

// PaginationOutputParams represents the output (response) for pagination
type PaginationOutputParams struct {
	Page       int64 `json:"page"`
	TotalPages int64 `json:"total_pages"`
	PerPage    int64 `json:"per_page"`
	Total      int64 `json:"total"`
}

// LoggingConfig holds configuration for the logging middleware
type LoggingConfig struct {
	// SensitiveFields contains field names that should be masked in logs
	SensitiveFields []string
	// LogRequestBody determines if request body should be logged
	LogRequestBody bool
	// LogResponseBody determines if response body should be logged
	LogResponseBody bool
	// MaxFieldLength is the maximum length for field values in logs
	MaxFieldLength int
	// PrettyLog when true uses multi-line messages and indented JSON; when false, logs are one line with compact JSON.
	// DefaultLoggingConfig sets this to true; the zero value is false (compact).
	PrettyLog bool
}

// ErrorHandler is an interface for handling known errors
// If you have a set of domain-specific errors,
// implement this interface and use in ErrorHandler middleware
type ErrorHandler interface {
	// Handle returns a payload with relative data if error is handled.
	// "ok" indicates whether the error was handled or not.
	Handle(err error) (payload ErrorHandlerPayload, ok bool)
}
