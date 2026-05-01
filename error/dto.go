package error

// ErrorDTO is the standard JSON error response shape.
type ErrorDTO struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

// ValidationErrorDTO is the error response shape for validation failures.
type ValidationErrorDTO struct {
	ErrorDTO
	Errors []ValidationErrorItem `json:"errors"`
}

// ValidationErrorItem describes a single field-level validation failure.
type ValidationErrorItem struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
