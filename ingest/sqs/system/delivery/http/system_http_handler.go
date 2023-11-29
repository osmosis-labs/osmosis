package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"

	"go.uber.org/zap"

	"github.com/go-redis/redis"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain/mvc"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/log"

	"github.com/labstack/echo"
)

type SystemHandler struct {
	logger    log.Logger
	host      string // Host for the GRPC gateway
	grpcPort  string // GRPC gateway port
	redisHost string // Redis host
	redisPort string // Redis port
	CIUsecase mvc.ChainInfoUsecase
}

// NewSystemHandler will initialize the /debug/ppof resources endpoint
func NewSystemHandler(e *echo.Echo, redisAddress, grpcAddress string, logger log.Logger, us mvc.ChainInfoUsecase) {
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
		CIUsecase: us,
	}

	e.GET("/debug/pprof/*", echo.WrapHandler(http.DefaultServeMux))
	e.GET("/healthcheck", handler.GetHealthStatus)

	// // Register pprof handlers on "/debug/pprof"
	// e.GET("/debug/pprof/*", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
}

// GetHealthStatus handles health check requests for both GRPC gateway and Redis
func (h *SystemHandler) GetHealthStatus(c echo.Context) error {
	ctx := c.Request().Context()

	// Check GRPC Gateway status
	grpcStatus := "running"
	url := fmt.Sprintf("http://%s:%s/status", h.host, h.grpcPort)
	resp, err := http.Get(url)
	if err != nil {
		grpcStatus = "down"
		h.logger.Error("Error checking GRPC gateway status", zap.Error(err))
	} else {
		defer resp.Body.Close()
	}

	// Check the latest height from chain info use case
	latestHeight, err := h.CIUsecase.GetLatestHeight(ctx)
	if err != nil {
		return err
	}

	// Parse the response from the GRPC Gateway status endpoint
	var statusResponse struct {
		Result struct {
			SyncInfo struct {
				LatestBlockHeight string `json:"latest_block_height"`
			} `json:"sync_info"`
		} `json:"result"`
	}

	if resp != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read response body")
		}

		err = json.Unmarshal(bodyBytes, &statusResponse)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to parse JSON response")
		}
	}

	// Compare latestHeight with latest_block_height from the status endpoint
	nodeStatus := "synced"
	if statusResponse.Result.SyncInfo.LatestBlockHeight != fmt.Sprint(latestHeight) {
		nodeStatus = "not_synced"
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
		"node_status":         nodeStatus,
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
