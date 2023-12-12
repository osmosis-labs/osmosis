package usecase

import (
	"context"
	"errors"
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
// Uses caching strategies for optimal performance.
// Currently, supports candidate route caching. If candidate routes for the given token in and token out denoms
// are present in cache, they are used without re-computing them. Otherwise, they are computed and cached.
// In the future, we will support caching of ranked routes that are constructed from candidate and sorted
// by the decreasing amount out within an order of magnitude of token in. Similarly, We will also support optimal split caching
// Returns error if:
// - fails to estimate direct quotes for ranked routes
// - fails to retrieve candidate routes
// -
func (r *routerUseCaseImpl) GetOptimalQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string) (domain.Quote, error) {
	// TODO: implement and check ranked route cache
	hasRankedRoutesInCache := false

	var (
		rankedRoutes        []route.RouteImpl
		topSingleRouteQuote domain.Quote
		err                 error
	)

	router := r.initializeRouter()

	if hasRankedRoutesInCache {
		// TODO: if top routes are present in cache, estimate the quotes and return the best.
		topSingleRouteQuote, rankedRoutes, err = estimateDirectQuote(router, rankedRoutes, tokenIn)
		if err != nil {
			return nil, err
		}
	} else {
		// If top routes are not present in cache, retrieve unranked candidate routes
		candidateRoutes, err := r.handleCandidateRoutes(ctx, router, tokenIn.Denom, tokenOutDenom)
		if err != nil {
			r.logger.Error("error handling routes", zap.Error(err))
			return nil, err
		}

		for _, route := range candidateRoutes.Routes {
			r.logger.Debug("filtered_candidate_route", zap.Any("route", route))
		}

		// Rank candidate routes by estimating direct quotes
		topSingleRouteQuote, rankedRoutes, err = r.rankRoutesByDirectQuote(ctx, router, candidateRoutes, tokenIn, tokenOutDenom)
		if err != nil {
			r.logger.Error("error getting top routes", zap.Error(err))
			return nil, err
		}

		if len(rankedRoutes) == 0 {
			return nil, fmt.Errorf("no ranked routes found")
		}

		// TODO: Cache ranked routes
	}

	if len(rankedRoutes) == 1 {
		return topSingleRouteQuote, nil
	}

	// Compute split route quote
	topSplitQuote, err := router.GetSplitQuote(rankedRoutes, tokenIn)
	if err != nil {
		return nil, err
	}

	// TODO: Cache split route proportions

	finalQuote := topSingleRouteQuote

	// If the split route quote is better than the single route quote, return the split route quote
	if topSplitQuote.GetAmountOut().GT(topSingleRouteQuote.GetAmountOut()) {
		routes := topSplitQuote.GetRoute()

		r.logger.Debug("split route selected", zap.Int("route_count", len(routes)))
		for _, route := range routes {
			r.logger.Debug("route", zap.Stringer("route", route))
		}

		finalQuote = topSplitQuote
	}

	r.logger.Debug("single route selected", zap.Stringer("route", finalQuote.GetRoute()[0]))

	if finalQuote.GetAmountOut().IsZero() {
		return nil, errors.New("best we can do is no tokens out")
	}

	return finalQuote, nil
}

// rankRoutesByDirectQuote ranks the given candidate routes by estimating direct quotes over each route.
// Returns the top quote as well as the ranked routes in decrease order of amount out.
// Returns error if:
// - fails to read taker fees
// - fails to convert candidate routes to routes
// - fails to estimate direct quotes
func (r *routerUseCaseImpl) rankRoutesByDirectQuote(ctx context.Context, router *Router, candidateRoutes route.CandidateRoutes, tokenIn sdk.Coin, tokenOutDenom string) (domain.Quote, []route.RouteImpl, error) {
	// Note that retrieving pools and taker fees is done in separate transactions.
	// This is fine because taker fees don't change often.
	// TODO: this can be refactored to only retrieve the relevant taker fees.
	takerFees, err := r.routerRepository.GetAllTakerFees(ctx)
	if err != nil {
		return nil, nil, err
	}

	routes, err := r.poolsUsecase.GetRoutesFromCandidates(ctx, candidateRoutes, takerFees, tokenIn.Denom, tokenOutDenom)
	if err != nil {
		return nil, nil, err
	}

	topQuote, routes, err := estimateDirectQuote(router, routes, tokenIn)
	if err != nil {
		return nil, nil, err
	}

	return topQuote, routes, nil
}

// estimateDirectQuote estimates and returns the direct quote for the given routes, token in and token out denom.
// Also, returns the routes ranked by amount out in decreasing order.
// Returns error if:
// - fails to estimate direct quotes
func estimateDirectQuote(router *Router, routes []route.RouteImpl, tokenIn sdk.Coin) (domain.Quote, []route.RouteImpl, error) {
	topQuote, routesSortedByAmtOut, err := router.estimateBestSingleRouteQuote(routes, tokenIn)
	if err != nil {
		return nil, nil, err
	}

	numRoutes := len(routesSortedByAmtOut)

	// If split routes are disabled, return a single the top route
	if router.maxSplitRoutes == 0 && numRoutes > 0 {
		numRoutes = 1
		// If there are more routes than the max split routes, keep only the top routes
	} else if len(routesSortedByAmtOut) > router.maxSplitRoutes {
		// Keep only top routes for splits
		routes = routes[:router.maxSplitRoutes]
		numRoutes = router.maxSplitRoutes
	}

	// Convert routes sorted by amount out to routes
	for i := 0; i < numRoutes; i++ {
		// Update routes with the top routes
		routes[i] = routesSortedByAmtOut[i].RouteImpl
	}

	return topQuote, routes, nil
}

// GetBestSingleRouteQuote returns the best single route quote to be done directly without a split.
func (r *routerUseCaseImpl) GetBestSingleRouteQuote(ctx context.Context, tokenIn sdk.Coin, tokenOutDenom string) (domain.Quote, error) {
	router := r.initializeRouter()

	candidateRoutes, err := r.handleCandidateRoutes(ctx, router, tokenIn.Denom, tokenOutDenom)
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

	candidateRoutes, err := r.handleCandidateRoutes(ctx, router, tokenIn.Denom, tokenOutDenom)
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

	routes, err := r.handleCandidateRoutes(ctx, router, tokenInDenom, tokenOutDenom)
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

// GetCachedCandidateRoutes implements mvc.RouterUsecase.
func (r *routerUseCaseImpl) GetCachedCandidateRoutes(ctx context.Context, tokenInDenom string, tokenOutDenom string) (route.CandidateRoutes, error) {
	if !r.config.RouteCacheEnabled {
		return route.CandidateRoutes{}, fmt.Errorf("route cache is disabled")
	}

	cachedCandidateRoutes, err := r.routerRepository.GetRoutes(ctx, tokenInDenom, tokenOutDenom)
	if err != nil {
		return route.CandidateRoutes{}, err
	}

	return cachedCandidateRoutes, nil
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

// handleCandidateRoutes attempts to retrieve candidate routes from the cache. If no routes are cached, it will
// compute, persist in cache and return them.
// Returns routes on success
// Errors if:
// - there is an error retrieving routes from cache
// - there are no routes cached and there is an error computing them
// - fails to persist the computed routes in cache
func (r *routerUseCaseImpl) handleCandidateRoutes(ctx context.Context, router *Router, tokenInDenom, tokenOutDenom string) (candidateRoutes route.CandidateRoutes, err error) {
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
