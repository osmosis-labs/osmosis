package sqs

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/middleware"
	poolsHttpDelivery "github.com/osmosis-labs/osmosis/v20/ingest/sqs/pools/delivery/http"
	poolsRedisRepository "github.com/osmosis-labs/osmosis/v20/ingest/sqs/pools/repository/redis"
	poolsUseCase "github.com/osmosis-labs/osmosis/v20/ingest/sqs/pools/usecase"
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
func NewSideCarQueryServer(dbHost, dbPort, sideCarQueryServerAddress string, useCaseTimeoutDuration int) (SideCarQueryServer, error) {
	// Handle SIGINT and SIGTERM signals to initiate shutdown
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGTERM)

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
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
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	// Create redis client and ensure that it is up.
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", dbHost, dbPort),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	redisStatus := redisClient.Ping(ctx)
	_, err := redisStatus.Result()
	if err != nil {
		return nil, err
	}

	// Initialize pools repository, usecase and HTTP handler
	poolsRepository := poolsRedisRepository.NewRedisPoolsRepo(redisClient)
	timeoutContext := time.Duration(useCaseTimeoutDuration) * time.Second
	poolsUseCase := poolsUseCase.NewPoolsUsecase(timeoutContext, poolsRepository)
	poolsHttpDelivery.NewPoolsHandler(e, poolsUseCase)

	// Start server in a separate goroutine
	go func() {
		err = e.Start(sideCarQueryServerAddress)
		if err != nil {
			panic(err)
		}
	}()

	return &sideCarQueryServer{
		poolsRepository: poolsRepository,
	}, nil
}
