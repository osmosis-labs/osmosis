package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"strconv"

	"go.uber.org/zap"

	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mvc"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/log"

	"github.com/labstack/echo"
)

type SystemHandler struct {
	logger       log.Logger
	redisAddress string
	grpcAddress  string
	CIUsecase    mvc.ChainInfoUsecase
}

const heightTolerance = 10

// NewSystemHandler will initialize the /debug/ppof resources endpoint
func NewSystemHandler(e *echo.Echo, redisAddress, grpcAddress string, logger log.Logger, us mvc.ChainInfoUsecase) {
	handler := &SystemHandler{
		logger:       logger,
		redisAddress: redisAddress,
		grpcAddress:  grpcAddress,
		CIUsecase:    us,
	}

	e.GET("/debug/pprof/*", echo.WrapHandler(http.DefaultServeMux))
	e.GET("/healthcheck", handler.GetHealthStatus)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
}

// GetHealthStatus handles health check requests for both GRPC gateway and Redis
func (h *SystemHandler) GetHealthStatus(c echo.Context) error {
	ctx := c.Request().Context()

	// Check GRPC Gateway status
	url := h.grpcAddress + "/status"
	resp, err := http.Get(url)
	if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
		h.logger.Error("Error checking GRPC gateway status", zap.Error(err))
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Error connecting to the Osmosis chain via GRPC gateway")
	} else {
		defer resp.Body.Close()
	}

	// Check the latest height from chain info use case
	latestStoreHeight, err := h.CIUsecase.GetLatestHeight(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get latest height from Redis")
	}

	// Parse the response from the GRPC Gateway status endpoint
	type JsonResponse struct {
		Result struct {
			SyncInfo struct {
				LatestBlockHeight string `json:"latest_block_height"`
			} `json:"sync_info"`
		} `json:"result"`
	}

	var statusResponse JsonResponse

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

	// allow 10 blocks of difference before claiming node is not synced

	latestChainHeight, err := strconv.ParseUint(statusResponse.Result.SyncInfo.LatestBlockHeight, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to parse latest block height from GRPC gateway")
	}

	// If the node is not synced, return HTTP 503
	if latestChainHeight-latestStoreHeight > heightTolerance {
		return echo.NewHTTPError(http.StatusServiceUnavailable, fmt.Sprintf("Node is not synced, chain height (%d), store height (%d), tolerance (%d)", latestChainHeight, latestStoreHeight, heightTolerance))
	}

	// Check Redis status
	rdb := redis.NewClient(&redis.Options{
		Addr: h.redisAddress,
	})

	if _, err := rdb.Ping().Result(); err != nil {
		h.logger.Error("Error connecting to Redis", zap.Error(err))
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Error connecting to Redis", err)
	}

	// Return combined status
	return c.JSON(http.StatusOK, map[string]string{
		"grpc_gateway_status": "running",
		"redis_status":        "running",
		"chain_latest_height": fmt.Sprint(latestChainHeight),
		"store_latest_height": fmt.Sprint(latestStoreHeight),
	})
}
