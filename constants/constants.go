package constants

import "github.com/Lomank123/go-web-platform/types"

const (
	EntityNotFound        types.ErrorCode = "ENTITY_NOT_FOUND"
	InvalidRequestPayload types.ErrorCode = "INVALID_REQUEST_PAYLOAD"
	ValidationError       types.ErrorCode = "VALIDATION_ERROR"
	RequestError          types.ErrorCode = "REQUEST_ERROR"
	UnauthorizedError     types.ErrorCode = "UNAUTHORIZED_ERROR"
	InternalServerError   types.ErrorCode = "INTERNAL_SERVER_ERROR"
	ForbiddenError        types.ErrorCode = "FORBIDDEN_ERROR"
)

const (
	Local       types.Environment = "local"
	Development types.Environment = "dev"
	Staging     types.Environment = "stage"
	Production  types.Environment = "prod"
)

// TraceIDHeader is the header key used for the trace ID
const TraceIDHeader = "X-Trace-ID"
