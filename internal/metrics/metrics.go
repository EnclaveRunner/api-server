package metrics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

const badRequestThreshold = 400

var (
	RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "api_requests_total",
			Help:        "Total number of HTTP requests",
			Namespace:   "",
			Subsystem:   "",
			ConstLabels: nil,
		},
		[]string{"path", "status"},
	)
	ErrorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "api_requests_errors_total",
			Help:        "Total number of FAILED HTTP requests",
			Namespace:   "",
			Subsystem:   "",
			ConstLabels: nil,
		},
		[]string{"path", "status"},
	)
)

// Init registers all metrics with Prometheus
func Init() {
	prometheus.MustRegister(RequestCount)
	prometheus.MustRegister(ErrorCount)
}

// Handler returns the Prometheus HTTP handler
func MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		c.Next()
		status := c.Writer.Status()
		RequestCount.WithLabelValues(path, http.StatusText(status)).Inc()
		if status >= badRequestThreshold {
			ErrorCount.WithLabelValues(path, http.StatusText(status)).Inc()
		}
	}
}
