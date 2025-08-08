package main

import (
	_ "enclave-backend/docu"
	"enclave-backend/handlers"
	"enclave-backend/internal/logging"
	"enclave-backend/internal/metrics"
	"enclave-backend/internal/middleware"
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

// @title           Enclave API Server
// @version         0.0.1
// @termsOfService  http://swagger.io/terms/

// @license.name  GNU General Public License v3.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath/api/v1

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	// Set up logger
	logger := logging.NewLogger()

	// Initialize metrics
	metrics.Init()

	// Start secret lifecycle management
	middleware.StartSecretLifecycle(logger, 5*time.Second)

	r := gin.New()
	r.Use(gin.Recovery())

	r.Use(metrics.MetricsHandler())

	// Logging & metrics middleware
	r.Use(func(c *gin.Context) {
		// Log request received
		logger.WithFields(map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
		}).Info("request received")

		// Increment metrics
		metrics.RequestCount.WithLabelValues(c.Request.URL.Path, c.Request.Method).Inc()

		// Process request
		c.Next()

		// Log status if not successful
		status := c.Writer.Status()
		if status >= 400 {
			logger.WithFields(map[string]interface{}{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
				"status": status,
			}).Warn("request failed")
		}
	})

	// Swagger UI-endpoint
	r.GET(apiPath("/swagger/*any"), ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Prometheus metrics endpoint
	r.GET("/metrics", func(c *gin.Context) {
		gin.WrapH(promhttp.Handler())(c)
	})

	// JWT issue endpoint
	r.GET(apiPath("/issue-token"), handlers.IssueToken)

	r.GET(apiPath("/demo"), handlers.Demo)

	logger.Info("Starting API-Service on " + apiPath("") + port)
	r.Run(port)
}

func apiPath(path string) string {
	return "/" + APIVersion + path
}
