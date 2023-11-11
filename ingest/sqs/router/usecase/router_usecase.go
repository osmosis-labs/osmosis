package usecase

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/log"
)

var _ domain.RouterUsecase = &routerUseCase{}

type routerUseCase struct {
	contextTimeout   time.Duration
	routerRepository domain.RouterRepository
	poolsUsecase     domain.PoolsUsecase
	config           domain.RouterConfig
	logger           log.Logger
}

// NewRouterUsecase will create a new pools use case object
func NewRouterUsecase(timeout time.Duration, routerRepository domain.RouterRepository, poolsUsecase domain.PoolsUsecase, config domain.RouterConfig, logger log.Logger) domain.RouterUsecase {
	return &routerUseCase{
		contextTimeout:   timeout,
		routerRepository: routerRepository,
		poolsUsecase:     poolsUsecase,
		config:           config,
		logger:           logger,
	}
}

// GetOptimalQuote returns the optimal quote by estimating the optimal route(s) through pools
// on the osmosis network.
func (r *routerUseCase) GetOptimalQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string) (domain.Quote, error) {
	allPools, err := r.poolsUsecase.GetAllPools(ctx)
	if err != nil {
		return nil, err
	}

	// Note that retrieving pools and taker fees is done in separate transactions.
	// This is fine because taker fees don't change often.
	takerFees, err := r.routerRepository.GetAllTakerFees(ctx)
	if err != nil {
		return nil, err
	}

	router := NewRouter([]uint64{}, allPools, takerFees, r.config.MaxPoolsPerRoute, r.config.MaxRoutes, r.config.MaxSplitIterations, r.config.MinOSMOLiquidity, r.logger)

	return router.getOptimalQuote(tokenIn, tokenOutDenom)
}

// GetBestSingleRouteQuote returns the best single route quote to be done directly without a split.
func (r *routerUseCase) GetBestSingleRouteQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string) (domain.Quote, error) {
	allPools, err := r.poolsUsecase.GetAllPools(ctx)
	if err != nil {
		return nil, err
	}

	// Note that retrieving pools and taker fees is done in separate transactions.
	// This is fine because taker fees don't change often.
	takerFees, err := r.routerRepository.GetAllTakerFees(ctx)
	if err != nil {
		return nil, err
	}

	router := NewRouter([]uint64{}, allPools, takerFees, r.config.MaxPoolsPerRoute, r.config.MaxRoutes, r.config.MaxSplitIterations, r.config.MinOSMOLiquidity, r.logger)
	return router.getBestSingleRouteQuote(tokenIn, tokenOutDenom)
}
