package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	platformConstants "github.com/go-web-services/go-web-platform/constants"
	platformError "github.com/go-web-services/go-web-platform/error"
	"github.com/go-web-services/go-web-platform/logger"
)

// ErrorHandlerMiddleware converts errors pushed via c.Error() into JSON responses.
// Priority: *BaseError (self-describing status + code) → ValidationErrors → 500 fallback.
func ErrorHandlerMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		log.Error(err)

		var baseErr *platformError.BaseError

		switch {
		case errors.As(err, &baseErr):
			c.AbortWithStatusJSON(baseErr.HTTPStatus(), platformError.ErrorDTO{
				ErrorCode: string(baseErr.HTTPErrorCode()),
				Message:   err.Error(),
			})
		case errors.As(err, &validator.ValidationErrors{}):
			c.AbortWithStatusJSON(http.StatusBadRequest, platformError.ValidationErrorDTO{
				ErrorDTO: platformError.ErrorDTO{
					Message:   "validation error occurred",
					ErrorCode: string(platformConstants.ValidationError),
				},
				Errors: formatValidationErrors(err),
			})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, platformError.ErrorDTO{
				Message:   "internal server error",
				ErrorCode: string(platformConstants.InternalServerError),
			})
		}
	}
}

func formatValidationErrors(err error) []platformError.ValidationErrorItem {
	var validationErr validator.ValidationErrors
	errors.As(err, &validationErr)

	items := make([]platformError.ValidationErrorItem, 0, len(validationErr))
	for _, e := range validationErr {
		items = append(items, platformError.ValidationErrorItem{
			Field:   e.Field(),
			Message: mapMessageByTag(e),
		})
	}
	return items
}

func mapMessageByTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "invalid email"
	case "min":
		return "minimum length is " + fe.Param()
	case "max":
		return "maximum length is " + fe.Param()
	}
	return "invalid value"
}
