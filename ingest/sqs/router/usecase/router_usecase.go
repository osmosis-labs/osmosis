package usecase

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mvc"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/log"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/route"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/routertesting/parsing"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
)

var _ mvc.RouterUsecase = &routerUseCaseImpl{}

type routerUseCaseImpl struct {
	contextTimeout   time.Duration
	routerRepository mvc.RouterRepository
	poolsUsecase     mvc.PoolsUsecase
	config           domain.RouterConfig
	logger           log.Logger
}

// NewRouterUsecase will create a new pools use case object
func NewRouterUsecase(timeout time.Duration, routerRepository mvc.RouterRepository, poolsUsecase mvc.PoolsUsecase, config domain.RouterConfig, logger log.Logger) mvc.RouterUsecase {
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
	router := r.initializeRouter()

	candidateRoutes, err := r.handleRoutes(ctx, router, tokenIn.Denom, tokenOutDenom)
	if err != nil {
		r.logger.Error("error handling routes", zap.Error(err))
		return nil, err
	}

	for _, route := range candidateRoutes.Routes {
		r.logger.Info("filtered_candidate_route", zap.Any("route", route))
	}

	// Note that retrieving pools and taker fees is done in separate transactions.
	// This is fine because taker fees don't change often.
	takerFees, err := r.routerRepository.GetAllTakerFees(ctx)
	if err != nil {
		return nil, err
	}

	routes, err := r.poolsUsecase.GetRoutesFromCandidates(ctx, candidateRoutes, takerFees, tokenIn.Denom, tokenOutDenom)
	if err != nil {
		return nil, err
	}

	return router.getOptimalQuote(tokenIn, routes)
}

// GetBestSingleRouteQuote returns the best single route quote to be done directly without a split.
func (r *routerUseCaseImpl) GetBestSingleRouteQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string) (domain.Quote, error) {
	router := r.initializeRouter()

	candidateRoutes, err := r.handleRoutes(ctx, router, tokenIn.Denom, tokenOutDenom)
	if err != nil {
		return nil, err
	}
	// TODO: abstract this

	takerFees, err := r.routerRepository.GetAllTakerFees(ctx)
	if err != nil {
		return nil, err
	}

	routes, err := r.poolsUsecase.GetRoutesFromCandidates(ctx, candidateRoutes, takerFees, tokenIn.Denom, tokenOutDenom)
	if err != nil {
		return nil, err
	}

	return router.getBestSingleRouteQuote(tokenIn, routes)
}

// GetCandidateRoutes implements domain.RouterUsecase.
func (r *routerUseCaseImpl) GetCandidateRoutes(ctx context.Context, tokenInDenom string, tokenOutDenom string) (route.CandidateRoutes, error) {
	router := r.initializeRouter()

	routes, err := r.handleRoutes(ctx, router, tokenInDenom, tokenOutDenom)
	if err != nil {
		return route.CandidateRoutes{}, err
	}

	return routes, nil
}

// initializeRouter initializes the router per configuration defined on the use case
// Returns error if:
// - there is an error retrieving pools from the store
// - there is an error retrieving taker fees from the store
// TODO: test
func (r *routerUseCaseImpl) initializeRouter() *Router {
	router := NewRouter([]uint64{}, r.config.MaxPoolsPerRoute, r.config.MaxRoutes, r.config.MaxSplitRoutes, r.config.MaxSplitIterations, r.config.MinOSMOLiquidity, r.logger)
	router = WithRouterRepository(router, r.routerRepository)
	router = WithPoolsUsecase(router, r.poolsUsecase)

	r.logger.Info("sorted pools", zap.Int("num_pools", len(router.sortedPools)))
	for _, pool := range router.sortedPools {
		r.logger.Debug("sorted pool", zap.Uint64("pool_id", pool.GetId()), zap.Stringer("tvl", pool.GetTotalValueLockedUOSMO()))
	}

	return router
}

// handleRoutes attempts to retrieve routes from the cache. If no routes are cached, it will
// compute, persist in cache and return them.
// Returns routes on success
// Errors if:
// - there is an error retrieving routes from cache
// - there are no routes cached and there is an error computing them
// - fails to persist the computed routes in cache
func (r *routerUseCaseImpl) handleRoutes(ctx context.Context, router *Router, tokenInDenom, tokenOutDenom string) (candidateRoutes route.CandidateRoutes, err error) {
	r.logger.Info("getting routes")

	// Check cache for routes if enabled
	if r.config.RouteCacheEnabled {
		candidateRoutes, err = r.routerRepository.GetRoutes(ctx, tokenInDenom, tokenOutDenom)
		if err != nil {
			return route.CandidateRoutes{}, err
		}
	}

	// TODO: swithch to debug
	r.logger.Info("cached routes", zap.Int("num_routes", len(candidateRoutes.Routes)))

	// If no routes are cached, find them
	if len(candidateRoutes.Routes) == 0 {
		r.logger.Info("calculating routes")

		r.logger.Info("retrieving pools")
		allPools, err := r.poolsUsecase.GetAllPools(ctx)
		if err != nil {
			return route.CandidateRoutes{}, err
		}
		r.logger.Info("retrieved pools", zap.Int("num_pools", len(allPools)))
		router = WithSortedPools(router, allPools)

		candidateRoutes, err = router.GetCandidateRoutes(tokenInDenom, tokenOutDenom)
		if err != nil {
			return route.CandidateRoutes{}, err
		}

		r.logger.Info("calculated routes", zap.Int("num_routes", len(candidateRoutes.Routes)))

		// Persist routes
		if len(candidateRoutes.Routes) > 0 && r.config.RouteCacheEnabled {
			r.logger.Info("persisting routes", zap.Int("num_routes", len(candidateRoutes.Routes)))

			if err := r.routerRepository.SetRoutes(ctx, tokenInDenom, tokenOutDenom, candidateRoutes); err != nil {
				return route.CandidateRoutes{}, err
			}
		}
	}

	return candidateRoutes, nil
}

// StoreRouterStateFiles implements domain.RouterUsecase.
// TODO: clean up
func (r *routerUseCaseImpl) StoreRouterStateFiles(ctx context.Context) error {
	// These pools do not contain tick model
	pools, err := r.poolsUsecase.GetAllPools(ctx)

	if err != nil {
		return err
	}

	concentratedpoolIDs := make([]uint64, 0, len(pools))
	for _, pool := range pools {
		if pool.GetType() == poolmanagertypes.Concentrated {
			concentratedpoolIDs = append(concentratedpoolIDs, pool.GetId())
		}
	}

	tickModelMap, err := r.poolsUsecase.GetTickModelMap(ctx, concentratedpoolIDs)
	if err != nil {
		return err
	}

	if err := parsing.StorePools(pools, tickModelMap, "pools.json"); err != nil {
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
