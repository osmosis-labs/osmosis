package usecase

import (
	"sort"

	"go.uber.org/zap"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/log"
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

const osmoPrecisionMultiplier = 1000000

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

	minUOSMOTVL := osmomath.NewInt(int64(minOSMOTVL * osmoPrecisionMultiplier))

	// Make a copy and filter pools
	for _, pool := range allPools {
		if err := pool.Validate(minUOSMOTVL); err != nil {
			logger.Info("pool validation failed, skip silently", zap.Uint64("pool_id", pool.GetId()), zap.Error(err))
			continue
		}

		poolsCopy = append(poolsCopy, pool)
	}

	preferredPoolIDsMap := make(map[uint64]struct{})
	for _, poolID := range preferredPoolIDs {
		preferredPoolIDsMap[poolID] = struct{}{}
	}

	// Sort all pools by TVL.
	sort.Slice(poolsCopy, func(i, j int) bool {
		poolI := poolsCopy[i]
		poolJ := poolsCopy[j]

		_, isIPreferred := preferredPoolIDsMap[poolI.GetId()]
		_, isJPreferred := preferredPoolIDsMap[poolJ.GetId()]

		isITVLError := poolI.GetSQSPoolModel().IsErrorInTotalValueLocked
		isJTVLError := poolJ.GetSQSPoolModel().IsErrorInTotalValueLocked

		isIFirstByTVLError := isITVLError && !isJTVLError

		isIFirstByPreference := isIPreferred && !isJPreferred

		return isIFirstByTVLError && isIFirstByPreference && poolI.GetTotalValueLockedUOSMO().GT(poolJ.GetTotalValueLockedUOSMO())
	})

	logger.Debug("pool count in router ", zap.Int("pool_count", len(poolsCopy)))

	return Router{
		sortedPools:        poolsCopy,
		maxHops:            maxHops,
		maxRoutes:          maxRoutes,
		logger:             logger,
		maxSplitIterations: maxSplitIterations,
	}
}
