package error

import (
	"net/http"

	"github.com/Lomank123/go-web-platform/constants"
)

var (
	ErrEntityNotFound        = NewErrorWithStatus(constants.EntityNotFound, "entity not found", http.StatusNotFound)
	ErrInvalidRequestPayload = NewError(constants.InvalidRequestPayload, "invalid request payload")
	ErrUnauthorized          = NewErrorWithStatus(constants.UnauthorizedError, "unauthorized", http.StatusUnauthorized)
	ErrInternalServerError   = NewErrorWithStatus(constants.InternalServerError, "internal server error", http.StatusInternalServerError)
	ErrForbidden             = NewErrorWithStatus(constants.ForbiddenError, "forbidden", http.StatusForbidden)
	ErrValidation            = NewError(constants.ValidationError, "validation error")
)
