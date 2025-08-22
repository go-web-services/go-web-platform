package platform

import (
	"github.com/Lomank123/go-web-platform/constants"

	"github.com/Lomank123/go-web-platform/logger"
	"github.com/Lomank123/go-web-platform/middleware"
	"github.com/Lomank123/go-web-platform/transport/http"
	"github.com/Lomank123/go-web-platform/types"
	"github.com/gin-gonic/gin"
)

// SetupPlatform injects platform specific entities into the project. Entrypoint for platform specific code.
func SetupPlatform(
	router *gin.Engine,
	log logger.Logger,
	readinessFunc func() error,
	loggingConfig types.LoggingConfig,
	errorHandler types.ErrorHandler,
	env types.Environment,
) {
	if env == constants.Production {
		log.Info("Setting production mode")
		gin.SetMode(gin.ReleaseMode)
	}

	// Add recovery middleware (throw 500 instead of crashes)
	router.Use(gin.Recovery())

	// Apply middlewares to the router before adding routes
	router.Use(middleware.LoggingMiddleware(log, loggingConfig))
	router.Use(middleware.ErrorHandlerMiddleware(log, errorHandler))

	// Add platform routes
	http.AddPlatformRoutes(router, log, readinessFunc, env)
}
