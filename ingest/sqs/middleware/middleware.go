package middleware

import (
	"context"
	"time"

	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
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

		requestMethod := c.Request().Method
		requestPath, err := domain.ParseURLPath(c)
		if err != nil {
			return err
		}

		// Increment the request counter
		requestsTotal.WithLabelValues(requestMethod, requestPath).Inc()

		// Insert the request path into the context
		ctx := c.Request().Context()
		ctx = context.WithValue(ctx, domain.RequestPathCtxKey, requestPath)
		request := c.Request().WithContext(ctx)
		c.SetRequest(request)

		err = next(c)

		duration := time.Since(start).Seconds()

		// Observe the duration with the histogram
		requestLatency.WithLabelValues(requestMethod, requestPath).Observe(duration)

		return err
	}
}
