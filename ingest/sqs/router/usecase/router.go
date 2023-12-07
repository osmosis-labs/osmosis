package usecase

import (
	"sort"

	"go.uber.org/zap"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mvc"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/log"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
)

type Router struct {
	preferredPoolIDS []uint64
	sortedPools      []domain.PoolI
	// The maximum number of hops to route through.
	maxHops int
	// The maximum number of routes to return.
	maxRoutes int

	// The maximum number of routes to split across
	// Must be smaller than or equal to maxRoutes.
	maxSplitRoutes int

	// The maximum number of split iterations to perform
	maxSplitIterations int

	minOSMOTVL int

	routerRepository mvc.RouterRepository

	poolsUsecase mvc.PoolsUsecase

	// The logger.
	logger log.Logger
}

type ratedPool struct {
	pool   domain.PoolI
	rating osmomath.Int
}

const (
	// OSMO token precision
	osmoPrecisionMultiplier = 1000000

	// Pool ordering constants below:

	noTotalValueLockedError = ""
)

// NewRouter returns a new Router.
// It initialized the routable pools where the given preferredPoolIDs take precedence.
// The rest of the pools are sorted by TVL.
// Each pool has a flag indicating whether there was an error in estimating its on-chain TVL.
// If that is the case, the pool is to be sorted towards the end. However, the preferredPoolIDs overwrites this rule
// and prioritizes the preferred pools.
func NewRouter(preferredPoolIDs []uint64, maxHops, maxRoutes, maxSplitRoutes, maxSplitIterations int, minOSMOTVL int, logger log.Logger) *Router {
	if logger == nil {
		logger = &log.NoOpLogger{}
	}

	return &Router{
		maxHops:            maxHops,
		maxRoutes:          maxRoutes,
		logger:             logger,
		maxSplitIterations: maxSplitIterations,
		maxSplitRoutes:     maxSplitRoutes,
		minOSMOTVL:         minOSMOTVL,
	}
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

func (r Router) GetSortedPools() []domain.PoolI {
	return r.sortedPools
}

func WithSortedPools(router *Router, allPools []domain.PoolI) *Router {
	// TODO: consider mutating directly on allPools
	router.sortedPools = make([]domain.PoolI, 0)
	totalTVL := osmomath.ZeroInt()

	minUOSMOTVL := osmomath.NewInt(int64(router.minOSMOTVL * osmoPrecisionMultiplier))

	// Make a copy and filter pools
	for _, pool := range allPools {
		if err := pool.Validate(minUOSMOTVL); err != nil {
			router.logger.Debug("pool validation failed, skip silently", zap.Uint64("pool_id", pool.GetId()), zap.Error(err))
			continue
		}

		router.sortedPools = append(router.sortedPools, pool)

		totalTVL = totalTVL.Add(pool.GetTotalValueLockedUOSMO())
	}

	preferredPoolIDsMap := make(map[uint64]struct{})
	for _, poolID := range router.preferredPoolIDS {
		preferredPoolIDsMap[poolID] = struct{}{}
	}

	// sort pools so that the appropriate pools are at the top
	router.sortedPools = sortPools(router.sortedPools, totalTVL, preferredPoolIDsMap, router.logger)

	return router
}

// WithRouterRepository instruments router by setting a router repository on it and returns the router.
func WithRouterRepository(router *Router, routerRepository mvc.RouterRepository) *Router {
	router.routerRepository = routerRepository
	return router
}

// WithPoolsUsecase instruments router by setting a pools usecase on it and returns the router.
func WithPoolsUsecase(router *Router, poolsUsecase mvc.PoolsUsecase) *Router {
	router.poolsUsecase = poolsUsecase
	return router
}

// sortPools sorts the given pools so that the most appropriate pools are at the top.
// The details of the sorting follow. Assign a rating to each pool based on the following criteria:
// - Initial rating equals to the pool's total value locked denominated in OSMO.
// - If the pool has no error in TVL, add 1/100 of total value locked across all pools to the rating.
// - If the pool is a preferred pool, add the total value locked across all pools to the rating.
// - If the pool is a concentrated pool, add 1/2 of total value locked across all pools to the rating.
// - If the pool is a transmuter pool, add 3/2 of total value locked across all pools to the rating.
// - Sort all pools by the rating score.
//
// This sorting exists to pursue the following heuristics:
// - The TVL is the main metric to sort pools by.
// - Preferred pools are prioritized by getting a boost.
// - Transmuter pools are the most efficient due to no slippage swaps so they get a boost.
// - Concentrated pools follow so they get a smaller boost.
// - Pools with no error in TVL are prioritized by getting an even smaller boost.
//
// These heuristics are imperfect and subject to change.
func sortPools(pools []domain.PoolI, totalTVL osmomath.Int, preferredPoolIDsMap map[uint64]struct{}, logger log.Logger) []domain.PoolI {
	logger.Debug("total tvl", zap.Stringer("total_tvl", totalTVL))

	ratedPools := make([]ratedPool, 0, len(pools))
	for _, pool := range pools {
		// Initialize rating to TVL.
		rating := pool.GetTotalValueLockedUOSMO()

		// 1/ 100 of toal value locked across all pools for no error in TVL
		if pool.GetSQSPoolModel().TotalValueLockedError == noTotalValueLockedError {
			rating = rating.Add(totalTVL.QuoRaw(100))
		}

		// Preferred pools get a boost equal to the total value locked across all pools
		_, isPreferred := preferredPoolIDsMap[pool.GetId()]
		if isPreferred {
			rating = rating.Add(totalTVL)
		}

		// Concentrated pools get a boost equal to 1/2 of total value locked across all pools
		isConcentrated := pool.GetType() == poolmanagertypes.Concentrated
		if isConcentrated {
			rating = rating.Add(totalTVL.QuoRaw(2))
		}

		// Transmuter pools get a boost equal to 3/2 of total value locked across all pools
		if isTransmuter := pool.GetType() == poolmanagertypes.CosmWasm; isTransmuter {
			rating = rating.Add(totalTVL.MulRaw(3).QuoRaw(2))
		}

		ratedPools = append(ratedPools, ratedPool{
			pool:   pool,
			rating: rating,
		})
	}

	// Sort all pools by the rating score
	sort.Slice(ratedPools, func(i, j int) bool {
		return ratedPools[i].rating.GT(ratedPools[j].rating)
	})

	logger.Info("pool count in router ", zap.Int("pool_count", len(ratedPools)))
	// Convert back to pools
	for i, ratedPool := range ratedPools {
		pool := ratedPool.pool

		sqsModel := pool.GetSQSPoolModel()
		logger.Info("pool", zap.Int("index", i), zap.Any("pool", pool.GetId()), zap.Stringer("rate", ratedPool.rating), zap.Stringer("tvl", sqsModel.TotalValueLockedUSDC), zap.String("tvl_error", sqsModel.TotalValueLockedError))
		pools[i] = ratedPool.pool
	}
	return pools
}
