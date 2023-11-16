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
	redisrepo "github.com/osmosis-labs/osmosis/v20/ingest/sqs/repository/redis"
	routerRedisRepository "github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/repository/redis"
	tokensUseCase "github.com/osmosis-labs/osmosis/v20/ingest/sqs/tokens/usecase"

	routerHttpDelivery "github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/delivery/http"
	routerUseCase "github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase"
)

// SideCarQueryServer defines an interface for sidecar query server (SQS).
// It encapsulates all logic for ingesting chain data into the server
// and exposes endpoints for querying formatter and processed data from frontend.
type SideCarQueryServer interface {
	GetTxManager() domain.TxManager
	GetPoolsRepository() domain.PoolsRepository
	GetRouterRepository() domain.RouterRepository
	GetTokensUseCase() domain.TokensUsecase
	GetLogger() log.Logger
}

type sideCarQueryServer struct {
	txManager        domain.TxManager
	poolsRepository  domain.PoolsRepository
	routerRepository domain.RouterRepository
	tokensUseCase    domain.TokensUsecase
	logger           log.Logger
}

// GetTokensUseCase implements SideCarQueryServer.
func (sqs *sideCarQueryServer) GetTokensUseCase() domain.TokensUsecase {
	return sqs.tokensUseCase
}

// GetPoolsRepository implements SideCarQueryServer.
func (sqs *sideCarQueryServer) GetPoolsRepository() domain.PoolsRepository {
	return sqs.poolsRepository
}

// GetRouterRepository implements SideCarQueryServer.
func (sqs *sideCarQueryServer) GetRouterRepository() domain.RouterRepository {
	return sqs.routerRepository
}

// GetTxManager implements SideCarQueryServer.
func (sqs *sideCarQueryServer) GetTxManager() domain.TxManager {
	return sqs.txManager
}

// GetLogger implements SideCarQueryServer.
func (sqs *sideCarQueryServer) GetLogger() log.Logger {
	return sqs.logger
}

// NewSideCarQueryServer creates a new sidecar query server (SQS).
func NewSideCarQueryServer(appCodec codec.Codec, routerConfig domain.RouterConfig, dbHost, dbPort, sideCarQueryServerAddress string, useCaseTimeoutDuration int, logger log.Logger) (SideCarQueryServer, error) {
	// Handle SIGINT and SIGTERM signals to initiate shutdown
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGTERM)

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
	_, err := redisStatus.Result()
	if err != nil {
		return nil, err
	}

	// Creare repository manager
	redisTxManager := redisrepo.NewTxManager(redisClient)

	// Initialize pools repository, usecase and HTTP handler
	poolsRepository := poolsRedisRepository.NewRedisPoolsRepo(appCodec, redisTxManager)
	timeoutContext := time.Duration(useCaseTimeoutDuration) * time.Second
	poolsUseCase := poolsUseCase.NewPoolsUsecase(timeoutContext, poolsRepository, redisTxManager)
	poolsHttpDelivery.NewPoolsHandler(e, poolsUseCase)

	// Initialize router repository, usecase and HTTP handler
	routerRepository := routerRedisRepository.NewRedisRouterRepo(redisTxManager)
	routerUsecase := routerUseCase.NewRouterUsecase(timeoutContext, routerRepository, poolsUseCase, routerConfig, logger)
	routerHttpDelivery.NewRouterHandler(e, routerUsecase)

	// Initialized tokens usecase
	tokensUseCase := tokensUseCase.NewTokensUsecase(timeoutContext)

	// Start server in a separate goroutine
	go func() {
		logger.Info("Starting sidecar query server", zap.String("address", sideCarQueryServerAddress))
		err = e.Start(sideCarQueryServerAddress)
		if err != nil {
			panic(err)
		}
	}()

	return &sideCarQueryServer{
		txManager:        redisTxManager,
		poolsRepository:  poolsRepository,
		routerRepository: routerRepository,
		tokensUseCase:    tokensUseCase,
		logger:           logger,
	}, nil
}
