package platform

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/go-web-services/go-web-platform/constants"
	"github.com/go-web-services/go-web-platform/logger"
	"github.com/go-web-services/go-web-platform/middleware"
	"github.com/go-web-services/go-web-platform/transport/http"
	"github.com/go-web-services/go-web-platform/types"
)

// SetupPlatform wires platform middleware and routes into the provided router.
// Call this once in main.go before registering application routes.
func SetupPlatform(
	router *gin.Engine,
	log logger.Logger,
	readinessFunc func() error,
	loggingConfig types.LoggingConfig,
	env types.Environment,
) {
	if env == constants.Production {
		log.Info("Setting production mode")
		gin.SetMode(gin.ReleaseMode)
	}

	// Make validator use json tag names so validation errors report "email" not "Email".
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	}

	router.Use(gin.Recovery())
	router.Use(middleware.LoggingMiddleware(log, loggingConfig))
	router.Use(middleware.ErrorHandlerMiddleware(log))

	http.AddPlatformRoutes(router, log, readinessFunc, env)
}
