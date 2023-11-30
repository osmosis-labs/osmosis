package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"

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

	// // Register pprof handlers on "/debug/pprof"
	// e.GET("/debug/pprof/*", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
}

// GetHealthStatus handles health check requests for both GRPC gateway and Redis
func (h *SystemHandler) GetHealthStatus(c echo.Context) error {
	ctx := c.Request().Context()

	// Check GRPC Gateway status
	grpcStatus := "running"
	url := h.grpcAddress + "/status"
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

	// Compare latestHeight with latest_block_height from the status endpoint
	nodeStatus := "synced"

	// allow 10 blocks of difference before claiming node is not synced
	if fmt.Sprint(int64(latestHeight)+10) < statusResponse.Result.SyncInfo.LatestBlockHeight {
		nodeStatus = "not_synced"
	}

	// Check Redis status
	redisStatus := "running"
	rdb := redis.NewClient(&redis.Options{
		Addr: h.redisAddress,
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
