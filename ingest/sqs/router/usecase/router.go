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
	// The maximum number of split iterations to perform
	maxSplitIterations int
	// The logger.
	logger log.Logger
}

// NewRouter returns a new Router.
// It initialized the routable pools where the given preferredPoolIDs take precedence.
// The rest of the pools are sorted by TVL.
// Sets
func NewRouter(preferredPoolIDs []uint64, allPools []domain.PoolI, maxHops int, maxRoutes int, maxSplitIterations int, logger log.Logger) Router {
	if logger == nil {
		logger = &log.NoOpLogger{}
	}

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

		return isIFirstByPreference && poolsCopy[i].GetTotalValueLockedUOSMO().GT(poolsCopy[j].GetTotalValueLockedUOSMO())
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
