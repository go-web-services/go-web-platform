package types

type ErrorCode string
type Environment string

type PaginationInputParams struct {
	Page  int64 `json:"page,omitempty"`
	Limit int64 `json:"limit,omitempty"`
}

type PaginationOutputParams struct {
	Page       int64 `json:"page"`
	TotalPages int64 `json:"total_pages"`
	PerPage    int64 `json:"per_page"`
	Total      int64 `json:"total"`
}

type LoggingConfig struct {
	SensitiveFields []string
	LogRequestBody  bool
	LogResponseBody bool
	MaxFieldLength  int
	PrettyLog       bool
}
