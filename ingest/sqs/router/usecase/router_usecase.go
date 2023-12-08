package usecase

import (
	"context"
	"fmt"
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
		r.logger.Debug("filtered_candidate_route", zap.Any("route", route))
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

// GetCustomQuote implements mvc.RouterUsecase.
func (r *routerUseCaseImpl) GetCustomQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string, poolIDs []uint64) (domain.Quote, error) {
	// TODO: abstract this
	router := r.initializeRouter()

	candidateRoutes, err := r.handleRoutes(ctx, router, tokenIn.Denom, tokenOutDenom)
	if err != nil {
		return nil, err
	}

	takerFees, err := r.routerRepository.GetAllTakerFees(ctx)
	if err != nil {
		return nil, err
	}

	routes, err := r.poolsUsecase.GetRoutesFromCandidates(ctx, candidateRoutes, takerFees, tokenIn.Denom, tokenOutDenom)
	if err != nil {
		return nil, err
	}

	routeIndex := -1

	for curRouteIndex, route := range routes {
		routePools := route.GetPools()

		// Skip routes that do not match the pool length.
		if len(routePools) != len(poolIDs) {
			continue
		}

		for i, pool := range routePools {
			poolID := pool.GetId()

			desiredPoolID := poolIDs[i]

			// Break out of the loop if the poolID does not match the desired poolID
			if poolID != desiredPoolID {
				break
			}

			// Found a route that matches the poolIDs
			if i == len(routePools)-1 {
				routeIndex = curRouteIndex
			}
		}

		// If the routeIndex is not -1, then we found a route that matches the poolIDs
		// Break out of the loop
		if routeIndex != -1 {
			break
		}
	}

	// Validate routeIndex
	if routeIndex == -1 {
		return nil, fmt.Errorf("no route found for poolIDs: %v", poolIDs)
	}
	if routeIndex >= len(routes) {
		return nil, fmt.Errorf("routeIndex %d is out of bounds", routeIndex)
	}

	// Compute direct quote
	foundRoute := routes[routeIndex]
	quote, _, err := router.estimateBestSingleRouteQuote([]route.RouteImpl{foundRoute}, tokenIn)
	if err != nil {
		return nil, err
	}

	return quote, nil
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

// GetTakerFee implements mvc.RouterUsecase.
func (r *routerUseCaseImpl) GetTakerFee(ctx context.Context, poolID uint64) ([]domain.TakerFeeForPair, error) {
	takerFees, err := r.routerRepository.GetAllTakerFees(ctx)
	if err != nil {
		return []domain.TakerFeeForPair{}, err
	}

	pool, err := r.poolsUsecase.GetPool(ctx, poolID)
	if err != nil {
		return []domain.TakerFeeForPair{}, err
	}

	poolDenoms := pool.GetPoolDenoms()

	result := make([]domain.TakerFeeForPair, 0)

	for i := range poolDenoms {
		for j := i + 1; j < len(poolDenoms); j++ {
			denom0 := poolDenoms[i]
			denom1 := poolDenoms[j]

			takerFee := takerFees.GetTakerFee(denom0, denom1)

			result = append(result, domain.TakerFeeForPair{
				Denom0:   denom0,
				Denom1:   denom1,
				TakerFee: takerFee,
			})
		}
	}

	return result, nil
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
	r.logger.Debug("getting routes")

	// Check cache for routes if enabled
	if r.config.RouteCacheEnabled {
		candidateRoutes, err = r.routerRepository.GetRoutes(ctx, tokenInDenom, tokenOutDenom)
		if err != nil {
			return route.CandidateRoutes{}, err
		}
	}

	r.logger.Info("cached routes", zap.Int("num_routes", len(candidateRoutes.Routes)))

	// If no routes are cached, find them
	if len(candidateRoutes.Routes) == 0 {
		r.logger.Debug("calculating routes")
		allPools, err := r.poolsUsecase.GetAllPools(ctx)
		if err != nil {
			return route.CandidateRoutes{}, err
		}
		r.logger.Debug("retrieved pools", zap.Int("num_pools", len(allPools)))
		router = WithSortedPools(router, allPools)

		candidateRoutes, err = router.GetCandidateRoutes(tokenInDenom, tokenOutDenom)
		if err != nil {
			return route.CandidateRoutes{}, err
		}

		r.logger.Info("calculated routes", zap.Int("num_routes", len(candidateRoutes.Routes)))

		// Persist routes
		if len(candidateRoutes.Routes) > 0 && r.config.RouteCacheEnabled {
			r.logger.Debug("persisting routes", zap.Int("num_routes", len(candidateRoutes.Routes)))
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
