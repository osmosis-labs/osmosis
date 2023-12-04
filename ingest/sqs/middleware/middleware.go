package middleware

import (
	"net/url"
	"time"

	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus"
)

// GoMiddleware represent the data-struct for middleware
type GoMiddleware struct {
	// another stuff , may be needed by middleware
}

var (
	// total number of requests counter
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sqs_requests_total",
			Help: "Total number of requests.",
		},
		[]string{"method", "endpoint"},
	)

	// request latency histogram
	requestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sqs_request_duration_seconds",
			Help:    "Histogram of request latencies.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestLatency)
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

// InstrumentMiddleware will handle the instrumentation middleware
func (m *GoMiddleware) InstrumentMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		parsedURL, err := url.Parse(c.Request().RequestURI)
		if err != nil {
			return err
		}

		requestMethod := c.Request().Method
		requestPath := parsedURL.Path

		// Increment the request counter
		requestsTotal.WithLabelValues(requestMethod, requestPath).Inc()

		err = next(c)

		duration := time.Since(start).Seconds()

		// Observe the duration with the histogram
		requestLatency.WithLabelValues(requestMethod, requestPath).Observe(duration)

		return err
	}
}
