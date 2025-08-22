package error

import (
	"github.com/Lomank123/go-web-platform/constants"
)

var (
	ErrEntityNotFound        = NewError(constants.EntityNotFound, "entity not found")
	ErrInvalidRequestPayload = NewError(constants.InvalidRequestPayload, "invalid request payload")
	ErrUnauthorized          = NewError(constants.UnauthorizedError, "unauthorized")
	ErrInternalServerError   = NewError(constants.InternalServerError, "internal server error")
	ErrForbidden             = NewError(constants.ForbiddenError, "forbidden")
	ErrValidation            = NewError(constants.ValidationError, "validation error")
)
