package middleware

import (
	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus"
)

// GoMiddleware represent the data-struct for middleware
type GoMiddleware struct {
	// another stuff , may be needed by middleware
}

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sqs_requests_total",
			Help: "Total number of requests.",
		},
		[]string{"method"},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal)
}

// CORS will handle the CORS middleware
func (m *GoMiddleware) CORS(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		return next(c)
	}
}

// InitMiddleware initialize the middleware
func InitMiddleware() *GoMiddleware {
	return &GoMiddleware{}
}

func (m *GoMiddleware) InstrumentMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Increment the request counter
		requestsTotal.WithLabelValues(c.Request().Method).Inc()
		return next(c)
	}
}
