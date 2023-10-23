package usecase

import (
	"sort"

	"go.uber.org/zap"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/log"
)

type Router struct {
	sortedPools []domain.PoolI
	// The maximum number of hops to route through.
	maxHops int
	// The maximum number of routes to return.
	maxRoutes int
	// The logger.
	logger *zap.Logger
}

// NewRouter returns a new Router.
// It initialized the routable pools where the given preferredPoolIDs take precedence.
// The rest of the pools are sorted by TVL.
// Sets
func NewRouter(preferredPoolIDs []uint64, allPools []domain.PoolI, maxHops int, maxRoutes int, logger log.Logger) Router {
	// TODO: consider mutating directly on allPools
	poolsCopy := make([]domain.PoolI, len(allPools))
	copy(poolsCopy, allPools)

	preferredPoolIDsMap := make(map[uint64]struct{})
	for _, poolID := range preferredPoolIDs {
		preferredPoolIDsMap[poolID] = struct{}{}
	}

	// Sort all pools by TVL.
	sort.Slice(poolsCopy, func(i, j int) bool {

		_, isIPreferred := preferredPoolIDsMap[poolsCopy[i].GetId()]
		_, isJPreferred := preferredPoolIDsMap[poolsCopy[j].GetId()]

		isIFirstByPreference := isIPreferred && !isJPreferred

		return isIFirstByPreference && poolsCopy[i].GetTotalValueLockedUSDC().GT(poolsCopy[j].GetTotalValueLockedUSDC())
	})

	logger.Debug("pool count in router ", zap.Int("pool_count", len(poolsCopy)))

	return Router{
		sortedPools: poolsCopy,
		maxHops:     maxHops,
		maxRoutes:   maxRoutes,
	}
}

// getCandidateRoutes returns candidate routes from tokenInDenom to tokenOutDenom using DFS.
// Relies on the constructor to initialize the sorted pools with preferred pool IDs, max routes and max hops
func (r Router) getCandidateRoutes(tokenInDenom, tokenOutDenom string) ([]domain.Route, error) {
	r.logger.Debug("getting candidate routes", zap.String("token_in_denom", tokenInDenom), zap.String("token_out_denom", tokenOutDenom), zap.Int("sorted_pool_count", len(r.sortedPools)))
	return r.findRoutes(tokenInDenom, tokenOutDenom, &routeImpl{}, make([]bool, len(r.sortedPools)), nil)
}

// findRoutes returns routes from tokenInDenom to tokenOutDenom.
// This is a recursive algorithm that does depth-first search. As a result, it is not guaranteed to return the shortest path.
// The algorithm utilizes the pools defined on the router, max routes and max hops values.
// It does not do more than max hops recursive calls.
// It stops once max routes are found.
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

	// currentRoutePools := currentRoute.GetPools()
	// fmt.Println("currentRoute pools ", len(currentRoutePools))
	// for _, pool := range currentRoutePools {
	// 	fmt.Println("pool ", pool.GetId(), " ", pool.GetPoolDenoms(), " ", pool.GetTokenOutDenom())
	// }
	// fmt.Printf("\n")

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

		updatedPoolsUsed := make([]bool, len(poolsUsed))
		copy(updatedPoolsUsed, poolsUsed)
		updatedPoolsUsed[i] = true

		for _, poolDenom := range poolDenoms {
			// Skip if this is the previous token out denom
			if poolDenom == previousTokenOutDenom {
				continue
			}

			updatedCurrentRoute := currentRoute.DeepCopy()
			updatedCurrentRoute.AddPool(pool, poolDenom)

			updatedPreviousTokenOutDenoms := make([]string, len(previousTokenOutDenoms))
			copy(updatedPreviousTokenOutDenoms, previousTokenOutDenoms)
			updatedPreviousTokenOutDenoms = append(updatedPreviousTokenOutDenoms, poolDenom)

			newRoutes, err := r.findRoutes(tokenInDenom, tokenOutDenom, updatedCurrentRoute, updatedPoolsUsed, updatedPreviousTokenOutDenoms)
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
