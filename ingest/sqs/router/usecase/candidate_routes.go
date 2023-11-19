package usecase

import (
	"sort"

	"go.uber.org/zap"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/log"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase/route"
)

// GetCandidateRoutes returns candidate routes from tokenInDenom to tokenOutDenom using DFS.
// Relies on the constructor to initialize the sorted pools with preferred pool IDs, max routes and max hops.
// It.
// Once the initial routes are found using DFS, those are sorted by number of hops.
// If the are overlaps in pool IDs between routes, the routes with more hops are filtered out.
// TODO: clean up and reduce code duplication.
func (r Router) GetCandidateRoutes(tokenInDenom, tokenOutDenom string) ([]domain.Route, error) {
	r.logger.Info("getting candidate routes", zap.String("token_in_denom", tokenInDenom), zap.String("token_out_denom", tokenOutDenom), zap.Int("sorted_pool_count", len(r.sortedPools)))
	routes, err := r.findRoutes(tokenInDenom, tokenOutDenom, &route.RouteImpl{}, make([]bool, len(r.sortedPools)), nil)
	if err != nil {
		return nil, err
	}

	r.logger.Info("found routes ", zap.Int("routes_count", len(routes)))

	for k, route := range routes {
		r.logger.Info("route", zap.Int("num", k), zap.Stringer("route", route))
	}

	// Validate the chosen routes.
	routes, err = r.validateAndFilterRoutes(routes, tokenInDenom)
	if err != nil {
		r.logger.Error("validateRoutes failed", zap.Error(err))
		return nil, err
	}

	r.logger.Info("filtered routes ", zap.Int("routes_count", len(routes)))

	inverseRoutes, err := r.findRoutes(tokenOutDenom, tokenInDenom, &route.RouteImpl{}, make([]bool, len(r.sortedPools)), nil)
	if err != nil {
		return nil, err
	}

	r.logger.Info("found inverse routes ", zap.Int("routes_count", len(inverseRoutes)))

	for k, route := range inverseRoutes {
		r.logger.Info("route", zap.Int("num", k), zap.Stringer("route", route))
	}

	for _, route := range inverseRoutes {
		if err := route.Reverse(tokenOutDenom); err != nil {
			return nil, err
		}
	}

	// Validate the chosen routes.
	inverseRoutes, err = r.validateAndFilterRoutes(inverseRoutes, tokenInDenom)
	if err != nil {
		r.logger.Error("validateRoutes failed", zap.Error(err))
		return nil, err
	}

	r.logger.Info("filtered inverse routes ", zap.Int("routes_count", len(inverseRoutes)))

	routes = append(routes, inverseRoutes...)

	// filter the final routes again to remove duplicates
	routes, err = r.validateAndFilterRoutes(routes, tokenInDenom)
	if err != nil {
		return nil, err
	}

	// Sort routes by number of hops.
	sort.Slice(routes, func(i, j int) bool {
		return len(routes[i].GetPools()) < len(routes[j].GetPools())
	})

	return routes, nil
}

// findRoutes returns routes from tokenInDenom to tokenOutDenom.
// This is a recursive algorithm that does depth-first search. As a result, it is not guaranteed to return the shortest path.
// The algorithm utilizes the pools defined on the router, max routes and max hops values.
// It does not do more than max hops recursive calls.
// It stops once max routes are found.
// It aims to avoid considering the same pool twice. If such a case occurs, it will skip the pool and the route.
//
// CONTRACT: The routable pool IDs are already sorted by preference (e.g. TVL, preferred)
//
// Errors if:
// - currentRoute is nil
// - sortedPools and poolsUsed have different lengths
// - sortedPools and pools in the route have different lengths
func (r Router) findRoutes(tokenInDenom, tokenOutDenom string, currentRoute domain.Route, poolsUsed []bool, previousTokenOutDenoms []string) ([]domain.Route, error) {
	if currentRoute == nil {
		return nil, ErrNilCurrentRoute
	}

	// Sorted pools and pools used should have the same length.
	if len(r.sortedPools) != len(poolsUsed) {
		return nil, SortedPoolsAndPoolsUsedLengthMismatchError{
			SortedPoolsLen: len(r.sortedPools),
			PoolsUsedLen:   len(poolsUsed),
		}
	}

	poolsInCurrentRoute := currentRoute.GetPools()
	numPoolInCurrentRoute := len(poolsInCurrentRoute)

	// Pools in the route should not be longer than the sorted pools.
	if numPoolInCurrentRoute > len(r.sortedPools) {
		return nil, SortedPoolsAndPoolsInRouteLengthMismatchError{
			SortedPoolsLen: len(r.sortedPools),
			PoolsInRoute:   numPoolInCurrentRoute,
		}
	}

	// Base case - route found
	if numPoolInCurrentRoute > 0 && poolsInCurrentRoute[numPoolInCurrentRoute-1].GetTokenOutDenom() == tokenOutDenom {
		return []domain.Route{currentRoute}, nil
	}

	// Unable to find - max hops reached
	if numPoolInCurrentRoute == r.maxHops {
		return []domain.Route{}, nil
	}

	result := []domain.Route{}

	if len(previousTokenOutDenoms) == 0 {
		previousTokenOutDenoms = make([]string, 0, r.maxHops)
		previousTokenOutDenoms = append(previousTokenOutDenoms, tokenInDenom)
	}

	previousTokenOutDenom := previousTokenOutDenoms[len(previousTokenOutDenoms)-1]

	for i, pool := range r.sortedPools {
		// Max number of routes reached - end early.
		if len(result) >= r.maxRoutes {
			break
		}

		if poolsUsed[i] {
			continue
		}

		// Check if previous token out denom is in the current pool
		poolDenoms := pool.GetPoolDenoms()
		isPreviousTokenOutDenomInPool := false
		for _, poolDenom := range poolDenoms {
			if poolDenom == previousTokenOutDenom {
				isPreviousTokenOutDenomInPool = true
				break
			}
		}

		// Skip if not
		if !isPreviousTokenOutDenomInPool {
			continue
		}

		// Note that pools used is reused across the recursive calls
		// This is to prevent having multiple routes that overlap on the same pool
		// We currently filter out routes with duplicate pools.
		// In the future, we may switch to using a prefix tree where this could be improved
		// and optimized: https://app.clickup.com/t/86a17rrmx
		copyPoolsUsed := make([]bool, len(poolsUsed))
		copy(copyPoolsUsed, poolsUsed)
		copyPoolsUsed[i] = true

		for _, poolDenom := range poolDenoms {
			// Skip if this is the previous token out denom
			if poolDenom == previousTokenOutDenom {
				continue
			}

			if poolDenom == tokenInDenom {
				continue
			}

			var updatedPreviousTokenOutDenoms []string

			updatedCurrentRoute := currentRoute.DeepCopy()

			updatedPreviousTokenOutDenoms = make([]string, len(previousTokenOutDenoms))
			copy(updatedPreviousTokenOutDenoms, previousTokenOutDenoms)
			// nolint: makezero
			updatedPreviousTokenOutDenoms = append(updatedPreviousTokenOutDenoms, poolDenom)

			takerFee, err := r.takerFeeMap.GetTakerFee(previousTokenOutDenom, poolDenom)
			if err != nil {
				return nil, err
			}

			updatedCurrentRoute.AddPool(pool, poolDenom, takerFee)

			newRoutes, err := r.findRoutes(tokenInDenom, tokenOutDenom, updatedCurrentRoute, copyPoolsUsed, updatedPreviousTokenOutDenoms)
			if err != nil {
				return nil, err
			}

			// Append new routes to result up until the max number of routes is reached.
			for i := 0; i < len(newRoutes) && len(result) < r.maxRoutes; i++ {
				result = append(result, newRoutes[i])
			}
		}
	}

	return result, nil
}

// GetSortedPoolIDs returns the sorted pool IDs.
// The sorting is initialized in NewRouter() by preferredPoolIDs and TVL.
func (r Router) GetSortedPoolIDs() []uint64 {
	sortedPoolIDs := make([]uint64, len(r.sortedPools))
	for i, pool := range r.sortedPools {
		sortedPoolIDs[i] = pool.GetId()
	}
	return sortedPoolIDs
}

// GetMaxHops returns the maximum number of hops configured.
func (r Router) GetMaxHops() int {
	return r.maxHops
}

// GetMaxRoutes returns the maximum number of routes configured.
func (r Router) GetMaxRoutes() int {
	return r.maxRoutes
}

// GetMaxSplitIterations returns the maximum number of iterations when searching for split routes.
func (r Router) GetMaxSplitIterations() int {
	return r.maxSplitIterations
}

// GetLogger returns the logger.
func (r Router) GetLogger() log.Logger {
	return r.logger
}

func (r Router) GetTakerFeeMap() domain.TakerFeeMap {
	return r.takerFeeMap
}
