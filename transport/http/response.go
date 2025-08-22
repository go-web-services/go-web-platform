package http

import (
	"errors"
	"net/http"

	"github.com/Lomank123/go-web-platform/constants"
	"github.com/Lomank123/go-web-platform/dto"
	"github.com/Lomank123/go-web-platform/types"

	platformError "github.com/Lomank123/go-web-platform/error"
	"github.com/gin-gonic/gin"
)

// Error sends an error response with the given error and status code
func Error(c *gin.Context, err error, statusCode int, errorCode types.ErrorCode) {
	c.AbortWithStatusJSON(statusCode, dto.ErrorDTO{
		ErrorCode: string(errorCode),
		Message:   err.Error(),
	})
}

func Ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

func BadRequest(c *gin.Context, err error, errorCode types.ErrorCode) {
	var reqErr *platformError.RequestError

	if errors.As(err, &reqErr) {
		c.AbortWithStatusJSON(http.StatusBadRequest, reqErr.Payload)
		return
	}

	Error(c, err, http.StatusBadRequest, errorCode)
}

func Unauthorized(c *gin.Context, err error) {
	Error(c, err, http.StatusUnauthorized, constants.UnauthorizedError)
}

func Forbidden(c *gin.Context, err error) {
	Error(c, err, http.StatusForbidden, constants.ForbiddenError)
}

func NotFound(c *gin.Context, err error) {
	Error(c, err, http.StatusNotFound, constants.EntityNotFound)
}

func InternalServerError(c *gin.Context, err error) {
	Error(c, err, http.StatusInternalServerError, constants.InternalServerError)
}
