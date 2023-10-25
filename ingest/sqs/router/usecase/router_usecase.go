package usecase

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/log"
)

var _ domain.RouterUsecase = &routerUseCase{}

type routerUseCase struct {
	contextTimeout time.Duration
	poolsUsecase   domain.PoolsUsecase
	config         domain.RouterConfig
	logger         log.Logger
}

// NewRouterUsecase will create a new pools use case object
func NewRouterUsecase(timeout time.Duration, poolsRepository domain.PoolsUsecase, config domain.RouterConfig, logger log.Logger) domain.RouterUsecase {
	return &routerUseCase{
		contextTimeout: timeout,
		poolsUsecase:   poolsRepository,
		config:         config,
		logger:         logger,
	}
}

// GetOptimalQuote returns the optimal quote by estimating the optimal route(s) through pools
// on the osmosis network.
func (a *routerUseCase) GetOptimalQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string) (domain.Quote, error) {
	allPools, err := a.poolsUsecase.GetAllPools(ctx)
	if err != nil {
		return nil, err
	}

	router := NewRouter([]uint64{}, allPools, a.config.MaxPoolsPerRoute, a.config.MaxRoutes, a.config.MaxSplitIterations, a.logger)

	return router.getQuote(tokenIn, tokenOutDenom)
}

// GetBestSingleRouteQuote returns the best single route quote to be done directly without a split.
func (a *routerUseCase) GetBestSingleRouteQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string) (domain.Quote, error) {
	allPools, err := a.poolsUsecase.GetAllPools(ctx)
	if err != nil {
		return nil, err
	}

	router := NewRouter([]uint64{}, allPools, a.config.MaxPoolsPerRoute, a.config.MaxRoutes, a.config.MaxSplitIterations, a.logger)

	routes, err := router.getCandidateRoutes(tokenIn.Denom, tokenOutDenom)
	if err != nil {
		return nil, err
	}

	router.logger.Debug("routes ", zap.Int("routes_count", len(routes)))

	// Validate the chosen routes.
	if routes, err = router.validateAndFilterRoutes(routes, tokenIn.Denom); err != nil {
		router.logger.Error("validateRoutes failed", zap.Error(err))
		// return nil, err
	}

	return router.getBestSingleRouteQuote(routes, tokenIn)
}
