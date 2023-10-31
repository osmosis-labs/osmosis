package usecase

import (
	"sort"

	"go.uber.org/zap"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/log"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

type Router struct {
	sortedPools []domain.PoolI
	// The maximum number of hops to route through.
	maxHops int
	// The maximum number of routes to return.
	maxRoutes int
	// The maximum number of split iterations to perform
	maxSplitIterations int
	// The logger.
	logger log.Logger
}

type ratedPool struct {
	pool   domain.PoolI
	rating int64
}

const (
	// OSMO token precision
	osmoPrecisionMultiplier = 1000000

	// Pool ordering constants below:

	// 100 points for no error in TVL
	noTVLErrorPoints = 100

	// each pool gets a proportion of this value depending on its on-chain TVL.
	proRataTVLPointsTotal = 10000

	// points for being a preferred pool
	preferredPoints = 1000

	// points for being concentrated
	concentratedPoints = 20
)

// NewRouter returns a new Router.
// It initialized the routable pools where the given preferredPoolIDs take precedence.
// The rest of the pools are sorted by TVL.
// Each pool has a flag indicating whether there was an error in estimating its on-chain TVL.
// If that is the case, the pool is to be sorted towards the end. However, the preferredPoolIDs overwrites this rule
// and prioritizes the preferred pools.
func NewRouter(preferredPoolIDs []uint64, allPools []domain.PoolI, maxHops int, maxRoutes int, maxSplitIterations int, minOSMOTVL int, logger log.Logger) Router {
	if logger == nil {
		logger = &log.NoOpLogger{}
	}

	// TODO: consider mutating directly on allPools
	poolsCopy := make([]domain.PoolI, 0)
	totalTVL := osmomath.ZeroInt()

	minUOSMOTVL := osmomath.NewInt(int64(minOSMOTVL * osmoPrecisionMultiplier))

	// Make a copy and filter pools
	for _, pool := range allPools {
		if err := pool.Validate(minUOSMOTVL); err != nil {
			logger.Info("pool validation failed, skip silently", zap.Uint64("pool_id", pool.GetId()), zap.Error(err))
			continue
		}

		poolsCopy = append(poolsCopy, pool)

		totalTVL = totalTVL.Add(pool.GetTotalValueLockedUOSMO())
	}

	preferredPoolIDsMap := make(map[uint64]struct{})
	for _, poolID := range preferredPoolIDs {
		preferredPoolIDsMap[poolID] = struct{}{}
	}

	// TODO: move rating into a separate function
	// https://app.clickup.com/t/86a19n1ge
	ratedPools := make([]ratedPool, 0, len(poolsCopy))
	for _, pool := range poolsCopy {
		osmoTVL := pool.GetTotalValueLockedUOSMO()
		tvlProportion := osmoTVL.ToLegacyDec().Quo(totalTVL.ToLegacyDec())

		// Get points proportional to 100 depending on the total on-chain TVL.
		rating := tvlProportion.Mul(osmomath.NewDec(proRataTVLPointsTotal)).RoundInt64()

		// 100 points for no error in TVL
		if !pool.GetSQSPoolModel().IsErrorInTotalValueLocked {
			rating += noTVLErrorPoints
		}

		_, isPreferred := preferredPoolIDsMap[pool.GetId()]
		if isPreferred {
			rating += preferredPoints
		}

		isConcentrated := pool.GetType() == poolmanagertypes.Concentrated
		if isConcentrated {
			rating += concentratedPoints
		}

		ratedPools = append(ratedPools, ratedPool{
			pool:   pool,
			rating: rating,
		})
	}

	// Sort all pools by the rating score
	sort.Slice(ratedPools, func(i, j int) bool {
		return ratedPools[i].rating > ratedPools[j].rating
	})

	logger.Debug("pool count in router ", zap.Int("pool_count", len(ratedPools)))
	logger.Info("initial pool order")
	for i, pool := range ratedPools {
		logger.Info("pool", zap.Int("index", i), zap.Any("pool", pool.pool), zap.Int64("rate", pool.rating))
	}

	// Convert back to pools
	for i, ratedPool := range ratedPools {
		poolsCopy[i] = ratedPool.pool
	}

	return Router{
		sortedPools:        poolsCopy,
		maxHops:            maxHops,
		maxRoutes:          maxRoutes,
		logger:             logger,
		maxSplitIterations: maxSplitIterations,
	}
}
