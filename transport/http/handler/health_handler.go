package http

import (
	"net/http"

	"github.com/Lomank123/go-web-platform/logger"
	"github.com/gin-gonic/gin"
)

type HealthHandler interface {
	HealthV1(c *gin.Context)
	ReadyV1(c *gin.Context)
}

type healthHandler struct {
	log           logger.Logger
	readinessFunc func() error
}

func NewHealthHandler(log logger.Logger, readinessFunc func() error) HealthHandler {
	if readinessFunc == nil {
		readinessFunc = func() error { return nil }
	}
	return &healthHandler{
		log:           log,
		readinessFunc: readinessFunc,
	}
}

// HealthV1
// @Summary Healthcheck endpoint.
// @Tags Health
// @Accept json
// @Produce json
// @Router /health [get]
func (h *healthHandler) HealthV1(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// ReadyV1
// @Summary Return 200 OK if the service is ready.
// @Tags Health
// @Accept json
// @Produce json
// @Router /ready [get]
func (h *healthHandler) ReadyV1(c *gin.Context) {
	if err := h.readinessFunc(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
