package main

import (
	_ "enclave-backend/docs"
	"enclave-backend/handlers"
	"enclave-backend/internal/logging"
	"enclave-backend/internal/metrics"
	"enclave-backend/internal/middleware"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	// APIVersion is the version of the API
	APIVersion = "api/v1"
	port       = ":8080"
)

// @title			Enclave API Server
// @version		0.0.1
// @description	API server for the Enclave project
// @termsOfService	http://swagger.io/terms/
// @contact.name	API Support
// @license.name	GNU General Public License v3.0
// @license.url	https://www.gnu.org/licenses/gpl-3.0.html
// @host			localhost:8080
// @BasePath		/api/v1
func main() {
	// Set up logger
	logger := logging.NewLogger()

	// Initialize metrics
	metrics.Init()

	// Start secret lifecycle management
	const secretRotationInterval = 5 * time.Second
	middleware.StartSecretLifecycle(logger, secretRotationInterval)

	router := gin.New()

	router.Use(gin.Recovery())

	api := router.Group("/" + APIVersion)

	router.Use(metrics.MetricsHandler())

	// Logging & metrics middleware
	router.Use(func(ctx *gin.Context) {
		// Log request received
		logger.WithFields(map[string]any{
			"method": ctx.Request.Method,
			"path":   ctx.Request.URL.Path,
		}).Info("request received")

		// Increment metrics
		metrics.RequestCount.WithLabelValues(ctx.Request.URL.Path, ctx.Request.Method).Inc()

		// Process request
		ctx.Next()

		// Log status if not successful
		status := ctx.Writer.Status()
		if status >= http.StatusBadRequest {
			logger.WithFields(map[string]any{
				"method": ctx.Request.Method,
				"path":   ctx.Request.URL.Path,
				"status": status,
			}).Warn("request failed")
		}
	})

	// Swagger UI-endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Prometheus metrics endpoint (no API group)
	router.GET("/metrics", func(ctx *gin.Context) {
		gin.WrapH(promhttp.Handler())(ctx)
	})

	// JWT issue endpoint
	api.GET("/issue-token", handlers.IssueToken)

	api.GET("/demo", handlers.Demo)

	logger.Info("Starting API-Service on " + port)
	if err := router.Run(port); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}
}
