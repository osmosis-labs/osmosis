package usecase

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/log"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase/routertesting/parsing"
)

var _ domain.RouterUsecase = &routerUseCaseImpl{}

type routerUseCaseImpl struct {
	contextTimeout   time.Duration
	routerRepository domain.RouterRepository
	poolsUsecase     domain.PoolsUsecase
	config           domain.RouterConfig
	logger           log.Logger
}

// NewRouterUsecase will create a new pools use case object
func NewRouterUsecase(timeout time.Duration, routerRepository domain.RouterRepository, poolsUsecase domain.PoolsUsecase, config domain.RouterConfig, logger log.Logger) domain.RouterUsecase {
	return &routerUseCaseImpl{
		contextTimeout:   timeout,
		routerRepository: routerRepository,
		poolsUsecase:     poolsUsecase,
		config:           config,
		logger:           logger,
	}
}

// GetOptimalQuote returns the optimal quote by estimating the optimal route(s) through pools
// on the osmosis network.
func (r *routerUseCaseImpl) GetOptimalQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string) (domain.Quote, error) {
	router, err := r.initializeRouter(ctx)
	if err != nil {
		return nil, err
	}

	routes, err := r.handleRoutes(ctx, router, tokenIn.Denom, tokenOutDenom)
	if err != nil {
		r.logger.Error("error handling routes", zap.Error(err))
		return nil, err
	}

	return router.getOptimalQuote(tokenIn, tokenOutDenom, routes)
}

// GetBestSingleRouteQuote returns the best single route quote to be done directly without a split.
func (r *routerUseCaseImpl) GetBestSingleRouteQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string) (domain.Quote, error) {
	router, err := r.initializeRouter(ctx)
	if err != nil {
		return nil, err
	}

	routes, err := r.handleRoutes(ctx, router, tokenIn.Denom, tokenOutDenom)
	if err != nil {
		return nil, err
	}

	return router.getBestSingleRouteQuote(tokenIn, tokenOutDenom, routes)
}

// GetCandidateRoutes implements domain.RouterUsecase.
func (r *routerUseCaseImpl) GetCandidateRoutes(ctx context.Context, tokenInDenom string, tokenOutDenom string) ([]domain.Route, error) {
	router, err := r.initializeRouter(ctx)
	if err != nil {
		return nil, err
	}

	routes, err := r.handleRoutes(ctx, router, tokenInDenom, tokenOutDenom)
	if err != nil {
		return nil, err
	}

	return routes, nil
}

// initializeRouter initializes the router per configuration defined on the use case
// Retrieves pools and taker fees from the store. Sorts pools and returns the final initialized router.
// Returns error if:
// - there is an error retrieving pools from the store
// - there is an error retrieving taker fees from the store
// TODO: test
func (r *routerUseCaseImpl) initializeRouter(ctx context.Context) (*Router, error) {
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

	router := NewRouter([]uint64{}, takerFees, r.config.MaxPoolsPerRoute, r.config.MaxRoutes, r.config.MaxSplitIterations, r.config.MinOSMOLiquidity, r.logger)
	router = WithSortedPools(router, allPools)

	r.logger.Info("sorted pools")
	for _, pool := range router.sortedPools {
		r.logger.Info("sorted pool", zap.Uint64("pool_id", pool.GetId()), zap.Stringer("tvl", pool.GetTotalValueLockedUOSMO()))
	}

	return router, nil
}

// handleRoutes attempts to retrieve routes from the cache. If no routes are cached, it will
// compute, persist in cache and return them.
// Returns routes on success
// Errors if:
// - there is an error retrieving routes from cache
// - there are no routes cached and there is an error computing them
// - fails to persist the computed routes in cache
func (r *routerUseCaseImpl) handleRoutes(ctx context.Context, router *Router, tokenInDenom, tokenOutDenom string) (routes []domain.Route, err error) {
	r.logger.Info("getting routes")

	// Check cache for routes if enabled
	if r.config.RouteCacheEnabled {
		routes, err = r.routerRepository.GetRoutes(ctx, tokenInDenom, tokenOutDenom)
		if err != nil {
			return nil, err
		}
	}

	// TODO: swithch to debug
	r.logger.Info("cached routes", zap.Int("num_routes", len(routes)))

	// If no routes are cached, find them
	if len(routes) == 0 {
		// TODO: swithch to debug
		r.logger.Info("calculating routes")

		routes, err = router.GetCandidateRoutes(tokenInDenom, tokenOutDenom)
		if err != nil {
			return nil, err
		}

		// Persist routes
		if len(routes) > 0 && r.config.RouteCacheEnabled {

			r.logger.Info("persisting routes", zap.Int("num_routes", len(routes)))

			if err := r.routerRepository.SetRoutes(ctx, tokenInDenom, tokenOutDenom, routes); err != nil {
				return nil, err
			}
		}
	}

	return routes, nil
}

// StoreRouterStateFiles implements domain.RouterUsecase.
// TODO: clean up
func (r *routerUseCaseImpl) StoreRouterStateFiles(ctx context.Context) error {
	pools, err := r.poolsUsecase.GetAllPools(ctx)

	if err != nil {
		return err
	}

	if err := parsing.StorePools(pools, "pools.json"); err != nil {
		return err
	}

	takerFeesMap, err := r.routerRepository.GetAllTakerFees(ctx)
	if err != nil {
		return err
	}

	if err := parsing.StoreTakerFees("taker_fees.json", takerFeesMap); err != nil {
		return err
	}

	return nil
}
