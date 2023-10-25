package sqs

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/log"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/middleware"
	poolsHttpDelivery "github.com/osmosis-labs/osmosis/v20/ingest/sqs/pools/delivery/http"
	poolsRedisRepository "github.com/osmosis-labs/osmosis/v20/ingest/sqs/pools/repository/redis"
	poolsUseCase "github.com/osmosis-labs/osmosis/v20/ingest/sqs/pools/usecase"

	routerHttpDelivery "github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/delivery/http"
	routerUseCase "github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase"
)

// SideCarQueryServer defines an interface for sidecar query server (SQS).
// It encapsulates all logic for ingesting chain data into the server
// and exposes endpoints for querying formatter and processed data from frontend.
type SideCarQueryServer interface {
	GetPoolsRepository() domain.PoolsRepository
}

type sideCarQueryServer struct {
	poolsRepository domain.PoolsRepository
}

// GetPoolsRepository implements SideCarQueryServer.
func (sqs *sideCarQueryServer) GetPoolsRepository() domain.PoolsRepository {
	return sqs.poolsRepository
}

// NewSideCarQueryServer creates a new sidecar query server (SQS).
func NewSideCarQueryServer(appCodec codec.Codec, dbHost, dbPort, sideCarQueryServerAddress string, useCaseTimeoutDuration int) (SideCarQueryServer, error) {
	// Handle SIGINT and SIGTERM signals to initiate shutdown
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGTERM)

	// logger
	// TODO: figure out logging to file
	isProductionLogger := true
	logger, err := log.NewLogger(isProductionLogger)
	logger.Info("Starting sidecar query server")

	defer func() {
		if err := recover(); err != nil {
			logger.Error("panic occurred", zap.Any("error", err))
			exitChan <- syscall.SIGTERM
		}
	}()

	// Setup echo server
	e := echo.New()
	middleware := middleware.InitMiddleware()
	e.Use(middleware.CORS)

	// Use context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-exitChan
		cancel() // Trigger shutdown

		err := e.Shutdown(ctx)
		if err != nil {
			logger.Error("error shutting down server", zap.Error(err))
		}

		os.Exit(0)
	}()

	// Create redis client and ensure that it is up.
	redisAddress := fmt.Sprintf("%s:%s", dbHost, dbPort)
	logger.Info("Pinging redis", zap.String("redis_address", redisAddress))
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	redisStatus := redisClient.Ping(ctx)
	_, err = redisStatus.Result()
	if err != nil {
		return nil, err
	}

	// Initialize pools repository, usecase and HTTP handler
	poolsRepository := poolsRedisRepository.NewRedisPoolsRepo(appCodec, redisClient)
	timeoutContext := time.Duration(useCaseTimeoutDuration) * time.Second
	poolsUseCase := poolsUseCase.NewPoolsUsecase(timeoutContext, poolsRepository)
	poolsHttpDelivery.NewPoolsHandler(e, poolsUseCase)

	// TODO: move to config file
	routerConfig := domain.RouterConfig{
		PreferredPoolIDs:   []uint64{},
		MaxPoolsPerRoute:   4,
		MaxRoutes:          5,
		MaxSplitIterations: 10,
	}

	// Initialize router usecase and HTTP handler
	routerUsecase := routerUseCase.NewRouterUsecase(timeoutContext, poolsUseCase, routerConfig, logger)
	routerHttpDelivery.NewRouterHandler(e, routerUsecase)

	// Start server in a separate goroutine
	go func() {
		logger.Info("Starting sidecar query server", zap.String("address", sideCarQueryServerAddress))
		err = e.Start(sideCarQueryServerAddress)
		if err != nil {
			panic(err)
		}
	}()

	return &sideCarQueryServer{
		poolsRepository: poolsRepository,
	}, nil
}
