package router

import (
	"sort"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
)

type Router struct {
	sortedPools []domain.PoolI
	// The maximum number of hops to route through.
	maxHops int
	// The maximum number of routes to return.
	maxRoutes int
}

type Route interface {
	GetPools() []RoutablePool
	DeepCopy() Route
	AddPool(pool domain.PoolI, tokenOut string)
}

var _ RoutablePool = &routablePoolImpl{}

type routablePoolImpl struct {
	domain.PoolI
	tokenOutDenom string
}

// GetTokenOutDenom implements RoutablePool.
func (rp *routablePoolImpl) GetTokenOutDenom() string {
	return rp.tokenOutDenom
}

// // SetTokenOutDenom implements RoutablePool.
// func (rp *routablePoolImpl) SetTokenOutDenom(tokenOutDenom string) {
// 	rp.tokenOutDenom = tokenOutDenom
// }

type RoutablePool interface {
	domain.PoolI
	GetTokenOutDenom() string
}

var _ Route = &routeImpl{}

type routeImpl struct {
	pools []RoutablePool
}

// GetPools implements Route.
func (r *routeImpl) GetPools() []RoutablePool {
	return r.pools
}

func (r routeImpl) DeepCopy() Route {
	poolsCopy := make([]RoutablePool, len(r.pools))
	copy(poolsCopy, r.pools)
	return &routeImpl{
		pools: poolsCopy,
	}
}

func (r *routeImpl) AddPool(pool domain.PoolI, tokenOutDenom string) {
	routablePool := &routablePoolImpl{
		PoolI:         pool,
		tokenOutDenom: tokenOutDenom,
	}
	r.pools = append(r.pools, routablePool)
}

// NewRouter returns a new Router.
// It initialized the routable pools where the given preferredPoolIDs take precedence.
// The rest of the pools are sorted by TVL.
// Sets
func NewRouter(preferredPoolIDs []uint64, allPools []domain.PoolI, maxHops int, maxRoutes int) Router {
	// TODO: consider mutating directly on allPools
	poolsCopy := make([]domain.PoolI, len(allPools))
	copy(poolsCopy, allPools)

	preferredPoolIDsMap := make(map[uint64]struct{})
	for _, poolID := range preferredPoolIDs {
		preferredPoolIDsMap[poolID] = struct{}{}
	}

	// Sort all pools by TVL.
	sort.Slice(allPools, func(i, j int) bool {

		_, isIPreferred := preferredPoolIDsMap[allPools[i].GetId()]
		_, isJPreferred := preferredPoolIDsMap[allPools[j].GetId()]

		isIFirstByPreference := isIPreferred && !isJPreferred

		return isIFirstByPreference && allPools[i].GetTotalValueLockedUSDC().GT(allPools[j].GetTotalValueLockedUSDC())
	})

	return Router{
		sortedPools: allPools,
		maxHops:     maxHops,
		maxRoutes:   maxRoutes,
	}
}

// getCandidateRoutes returns candidate routes from tokenInDenom to tokenOutDenom using DFS.
// Relies on the constructor to initialize the sorted pools with preferred pool IDs, max routes and max hops
func (r Router) getCandidateRoutes(tokenInDenom, tokenOutDenom string) ([]Route, error) {
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
func (r Router) findRoutes(tokenInDenom, tokenOutDenom string, currentRoute Route, poolsUsed []bool, previousTokenOutDenoms []string) ([]Route, error) {
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
		return []Route{currentRoute}, nil
	}

	// Unable to find - max hops reached
	if numPoolInCurrentRoute == r.maxHops {
		return []Route{}, nil
	}

	result := []Route{}

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
