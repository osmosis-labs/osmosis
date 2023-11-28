package http

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	"go.uber.org/zap"

	"github.com/go-redis/redis"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/log"

	"github.com/labstack/echo"
)

type SystemHandler struct {
	logger    log.Logger
	host      string // Host for the GRPC gateway
	grpcPort  string // GRPC gateway port
	redisHost string // Redis host
	redisPort string // Redis port
}

// NewSystemHandler will initialize the /debug/ppof resources endpoint
func NewSystemHandler(e *echo.Echo, logger log.Logger) {
	// GRPC Gateway configuration
	host := getEnvOrDefault("GRPC_GATEWAY_HOST", "localhost")
	grpcPort := getEnvOrDefault("GRPC_GATEWAY_PORT", "1317")

	// Redis configuration
	redisHost := getEnvOrDefault("REDIS_HOST", "localhost")
	redisPort := getEnvOrDefault("REDIS_PORT", "6379")

	handler := &SystemHandler{
		logger:    logger,
		host:      host,
		grpcPort:  grpcPort,
		redisHost: redisHost,
		redisPort: redisPort,
	}

	e.GET("/debug/pprof/*", echo.WrapHandler(http.DefaultServeMux))
	e.GET("/health", handler.GetHealthStatus)

	// // Register pprof handlers on "/debug/pprof"
	// e.GET("/debug/pprof/*", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
}

// GetHealthStatus handles health check requests for both GRPC gateway and Redis
func (h *SystemHandler) GetHealthStatus(c echo.Context) error {
	// Check GRPC Gateway status
	grpcStatus := "running"
	url := fmt.Sprintf("http://%s:%s/status", h.host, h.grpcPort)
	if _, err := http.Get(url); err != nil {
		grpcStatus = "down"
		h.logger.Error("Error checking GRPC gateway status", zap.Error(err))
	}

	// Check Redis status
	redisStatus := "running"
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", h.redisHost, h.redisPort),
	})

	if _, err := rdb.Ping().Result(); err != nil {
		redisStatus = "down"
		h.logger.Error("Error connecting to Redis", zap.Error(err))
	}

	// Return combined status
	return c.JSON(http.StatusOK, map[string]string{
		"grpc_gateway_status": grpcStatus,
		"redis_status":        redisStatus,
	})
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
