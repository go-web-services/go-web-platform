package dto

type ErrorDTO struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

type ValidationErrorDTO struct {
	ErrorDTO
	Errors []ValidationErrorItem `json:"errors"`
}

type ValidationErrorItem struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
