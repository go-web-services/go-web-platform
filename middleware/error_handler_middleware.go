package middleware

import (
	"errors"
	"net/http"

	platformConstants "github.com/Lomank123/go-web-platform/constants"

	"github.com/Lomank123/go-web-platform/dto"
	platformError "github.com/Lomank123/go-web-platform/error"
	"github.com/Lomank123/go-web-platform/logger"
	"github.com/Lomank123/go-web-platform/types"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ErrorHandlerMiddleware is a middleware that handles errors and returns appropriate responses
func ErrorHandlerMiddleware(log logger.Logger, errorHandler types.ErrorHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Only run if there are errors to handle
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			log.Error(err)

			var isHandled bool
			var payload types.ErrorHandlerPayload

			// Only try to handle if errorHandler is provided
			if errorHandler != nil {
				payload, isHandled = errorHandler.Handle(err)
			}

			// Used mostly in Gateways.
			// Here we handle all possible errors that
			// may be returned from shared modules as well
			var reqErr *platformError.RequestError
			var baseErr *platformError.BaseError

			switch {
			case isHandled:
				c.AbortWithStatusJSON(payload.StatusCode, payload.Payload)
			case errors.As(err, &reqErr):
				c.AbortWithStatusJSON(reqErr.StatusCode, reqErr.Payload)
			case errors.As(err, &baseErr):
				c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorDTO{
					Message:   baseErr.Message,
					ErrorCode: string(baseErr.Code),
				})
			case errors.As(err, &validator.ValidationErrors{}):
				errs := formatValidationErrors(err)
				c.AbortWithStatusJSON(http.StatusBadRequest, dto.ValidationErrorDTO{
					ErrorDTO: dto.ErrorDTO{
						Message:   "Validation error occurred",
						ErrorCode: string(platformConstants.ValidationError),
					},
					Errors: errs,
				})
			default:
				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					dto.ErrorDTO{
						Message:   "Internal server error",
						ErrorCode: string(platformConstants.InternalServerError),
					},
				)
			}
		}
	}
}

// formatValidationErrors formats validation errors
func formatValidationErrors(err error) []dto.ValidationErrorItem {
	var errItems []dto.ValidationErrorItem
	var validationErr validator.ValidationErrors
	errors.As(err, &validationErr)

	for _, e := range validationErr {
		errItems = append(errItems, dto.ValidationErrorItem{
			Field:   e.Field(),
			Message: mapMessageByTag(e),
		})
	}

	return errItems
}

// mapMessageByTag maps validation error tag to a user-friendly message
func mapMessageByTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email"
	case "min":
		return "Minimum length is " + fe.Param()
	case "max":
		return "Maximum length is " + fe.Param()
	}

	return "Invalid value"
}
