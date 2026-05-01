package http

import (
	"github.com/go-web-services/go-web-platform/constants"
	"github.com/go-web-services/go-web-platform/logger"
	http "github.com/go-web-services/go-web-platform/transport/http/handler"
	"github.com/go-web-services/go-web-platform/types"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

// AddPlatformRoutes adds all platform routes to provided router
func AddPlatformRoutes(
	router *gin.Engine,
	log logger.Logger,
	readinessFunc func() error,
	env types.Environment,
) *gin.Engine {
	// Add Prometheus metrics middleware and endpoint

	healthHandler := http.NewHealthHandler(log, readinessFunc)
	v1 := router.Group("/")
	{
		v1.GET("/health", healthHandler.HealthV1)
		v1.GET("/ready", healthHandler.ReadyV1)
	}

	// Swagger
	if env == constants.Development || env == constants.Staging || env == constants.Local {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	return router
}
