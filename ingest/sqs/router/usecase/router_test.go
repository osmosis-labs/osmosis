package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/app/apptesting"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/log"
	routerusecase "github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase"
)

type RouterTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}

// This test validates a happy path expected behavior that
// when router is created, it first takes the preferred pool IDs,
// then sorts by TVL.
// Other configurations parameters are also validated.
func (s *RouterTestSuite) TestNewRouter() {
	s.Setup()

	// Prepare all supported pools.
	allPool := s.PrepareAllSupportedPools()

	// Get balancer pool
	balancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, allPool.BalancerPoolID)
	s.Require().NoError(err)

	// Get stableswap pool
	stableswapPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, allPool.StableSwapPoolID)
	s.Require().NoError(err)

	// Get CL pool
	concentratedPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, allPool.ConcentratedPoolID)
	s.Require().NoError(err)

	// Get CosmWasm pool
	cosmWasmPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, allPool.CosmWasmPoolID)
	s.Require().NoError(err)

	var (
		// Inputs
		preferredPoolIDs   = []uint64{allPool.BalancerPoolID, allPool.StableSwapPoolID}
		maxHops            = 3
		maxRoutes          = 5
		maxSplitIterations = 10
		logger, _          = log.NewLogger(false)
		defaultAllPools    = []domain.PoolI{
			&domain.PoolWrapper{
				ChainModel: balancerPool,
				SQSModel: domain.SQSPool{
					TotalValueLockedUSDC: osmomath.NewInt(5), // 5
				},
			},
			&domain.PoolWrapper{
				ChainModel: stableswapPool,
				SQSModel: domain.SQSPool{
					TotalValueLockedUSDC: osmomath.OneInt(), // 1
				},
			},
			&domain.PoolWrapper{
				ChainModel: concentratedPool,
				SQSModel: domain.SQSPool{
					TotalValueLockedUSDC: osmomath.NewInt(4), // 4
				},
			},
			&domain.PoolWrapper{
				ChainModel: cosmWasmPool,
				SQSModel: domain.SQSPool{
					TotalValueLockedUSDC: osmomath.NewInt(3), // 3
				},
			},
		}

		// Expected
		// First, preferred pool IDs, then sorted by TVL.
		expectedSortedPoolIDs = []uint64{allPool.BalancerPoolID, allPool.StableSwapPoolID, allPool.ConcentratedPoolID, allPool.CosmWasmPoolID}
	)

	// System under test
	router := routerusecase.NewRouter(preferredPoolIDs, defaultAllPools, maxHops, maxRoutes, maxSplitIterations, logger)

	// Assert
	s.Require().Equal(maxHops, router.GetMaxHops())
	s.Require().Equal(maxRoutes, router.GetMaxRoutes())
	s.Require().Equal(maxSplitIterations, router.GetMaxSplitIterations())
	s.Require().Equal(logger, router.GetLogger())
	s.Require().Equal(expectedSortedPoolIDs, router.GetSortedPoolIDs())
}
