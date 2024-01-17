package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/log"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase"
	routerusecase "github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/route"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/routertesting"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/routertesting/parsing"
)

type RouterTestSuite struct {
	routertesting.RouterTestHelper
}

const (
	relativePathMainnetFiles = "./routertesting/parsing/"
	poolsFileName            = "pools.json"
	takerFeesFileName        = "taker_fees.json"
)

var defaultRouterConfig = domain.RouterConfig{
	PreferredPoolIDs:          []uint64{},
	MaxRoutes:                 4,
	MaxPoolsPerRoute:          4,
	MaxSplitRoutes:            4,
	MaxSplitIterations:        10,
	MinOSMOLiquidity:          20000,
	RouteUpdateHeightInterval: 0,
	RouteCacheEnabled:         false,
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}

const dummyTotalValueLockedErrorStr = "total value locked error string"

var (
	// Concentrated liquidity constants
	Denom0 = ETH
	Denom1 = USDC

	DefaultCurrentTick = routertesting.DefaultCurrentTick

	DefaultAmt0 = routertesting.DefaultAmt0
	DefaultAmt1 = routertesting.DefaultAmt1

	DefaultCoin0 = routertesting.DefaultCoin0
	DefaultCoin1 = routertesting.DefaultCoin1

	DefaultLiquidityAmt = routertesting.DefaultLiquidityAmt

	// router specific variables
	defaultTickModel = &domain.TickModel{
		Ticks:            []domain.LiquidityDepthsWithRange{},
		CurrentTickIndex: 0,
		HasNoLiquidity:   false,
	}

	noTakerFee = osmomath.ZeroDec()
)

// This test validates a happy path expected behavior that
// when router is created, it first takes the preferred pool IDs,
// then sorts by TVL.
// Other configurations parameters are also validated.
func (s *RouterTestSuite) TestNewRouter() {
	s.Setup()

	// Prepare all supported pools.
	allPool := s.PrepareAllSupportedPools()

	// Create additional pools for edge cases
	var (
		secondBalancerPoolPoolID = s.PrepareBalancerPool()
		thirdBalancerPoolID      = s.PrepareBalancerPool()

		// Note that these default denoms might not actually match the pool denoms for simplicity.
		defaultDenoms = []string{"foo", "bar"}
	)

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

	// Get second & third balancer pools
	// Note that his pool is preferred but has TVL error flag set.
	secondBalancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, secondBalancerPoolPoolID)
	s.Require().NoError(err)

	// Note that his pool is not preferred and has TVL error flag set.
	thirdBalancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, thirdBalancerPoolID)
	s.Require().NoError(err)

	var (
		// Inputs
		preferredPoolIDs   = []uint64{allPool.BalancerPoolID, allPool.StableSwapPoolID, secondBalancerPoolPoolID}
		maxHops            = 3
		maxRoutes          = 5
		maxSplitRoutes     = 5
		maxSplitIterations = 10
		minOsmoLiquidity   = 2
		logger, _          = log.NewLogger(false, "", "")
		defaultAllPools    = []domain.PoolI{
			&domain.PoolWrapper{
				ChainModel: balancerPool,
				SQSModel: domain.SQSPool{
					TotalValueLockedUSDC: osmomath.NewInt(5 * usecase.OsmoPrecisionMultiplier), // 5
					PoolDenoms:           defaultDenoms,
				},
			},
			&domain.PoolWrapper{
				ChainModel: stableswapPool,
				SQSModel: domain.SQSPool{
					TotalValueLockedUSDC: osmomath.NewInt(int64(minOsmoLiquidity) - 1), // 1
					PoolDenoms:           defaultDenoms,
				},
			},
			&domain.PoolWrapper{
				ChainModel: concentratedPool,
				SQSModel: domain.SQSPool{
					TotalValueLockedUSDC: osmomath.NewInt(4 * usecase.OsmoPrecisionMultiplier), // 4
					PoolDenoms:           defaultDenoms,
				},
				TickModel: &domain.TickModel{
					Ticks: []domain.LiquidityDepthsWithRange{
						{
							LowerTick:       0,
							UpperTick:       100,
							LiquidityAmount: osmomath.NewDec(100),
						},
					},
					CurrentTickIndex: 0,
					HasNoLiquidity:   false,
				},
			},
			&domain.PoolWrapper{
				ChainModel: cosmWasmPool,
				SQSModel: domain.SQSPool{
					TotalValueLockedUSDC: osmomath.NewInt(3 * usecase.OsmoPrecisionMultiplier), // 3
					PoolDenoms:           defaultDenoms,
				},
			},

			// Note that the pools below have higher TVL.
			// However, since they have TVL error flag set, they
			// should be sorted after other pools, unless overriden by preferredPoolIDs.
			&domain.PoolWrapper{
				ChainModel: secondBalancerPool,
				SQSModel: domain.SQSPool{
					TotalValueLockedUSDC:  osmomath.NewInt(10 * usecase.OsmoPrecisionMultiplier), // 10
					PoolDenoms:            defaultDenoms,
					TotalValueLockedError: dummyTotalValueLockedErrorStr,
				},
			},
			&domain.PoolWrapper{
				ChainModel: thirdBalancerPool,
				SQSModel: domain.SQSPool{
					TotalValueLockedUSDC:  osmomath.NewInt(11 * usecase.OsmoPrecisionMultiplier), // 11
					PoolDenoms:            defaultDenoms,
					TotalValueLockedError: dummyTotalValueLockedErrorStr,
				},
			},
		}

		// Expected
		// First, preferred pool IDs, then sorted by TVL.
		expectedSortedPoolIDs = []uint64{
			// Transmuter pool is first due to no slippage swaps
			allPool.CosmWasmPoolID,

			allPool.ConcentratedPoolID,

			thirdBalancerPoolID, // non-preferred pool ID with TVL error flag set

			secondBalancerPoolPoolID, // preferred pool ID with TVL error flag set

			// Balancer is above concentrated pool due to higher TVL
			allPool.BalancerPoolID,
		}
	)

	// System under test
	router := routerusecase.NewRouter(preferredPoolIDs, maxHops, maxRoutes, maxSplitRoutes, maxSplitIterations, minOsmoLiquidity, logger)
	router = routerusecase.WithSortedPools(router, defaultAllPools)

	// Assert
	s.Require().Equal(maxHops, router.GetMaxHops())
	s.Require().Equal(maxRoutes, router.GetMaxRoutes())
	s.Require().Equal(maxSplitIterations, router.GetMaxSplitIterations())
	s.Require().Equal(logger, router.GetLogger())
	s.Require().Equal(expectedSortedPoolIDs, router.GetSortedPoolIDs())
}

// getTakerFeeMapForAllPoolTokenPairs returns a map of all pool token pairs to their taker fees.
func (s *RouterTestSuite) getTakerFeeMapForAllPoolTokenPairs(pools []domain.PoolI) domain.TakerFeeMap {
	pairs := make(domain.TakerFeeMap, 0)

	for _, pool := range pools {
		poolDenoms := pool.GetPoolDenoms()

		for i := 0; i < len(poolDenoms); i++ {
			for j := i + 1; j < len(poolDenoms); j++ {

				hasTakerFee := pairs.Has(poolDenoms[i], poolDenoms[j])
				if hasTakerFee {
					continue
				}

				takerFee, err := s.App.PoolManagerKeeper.GetTradingPairTakerFee(s.Ctx, poolDenoms[i], poolDenoms[j])
				s.Require().NoError(err)

				pairs.SetTakerFee(poolDenoms[i], poolDenoms[j], takerFee)
			}
		}
	}

	return pairs
}

func WithRoutePools(r route.RouteImpl, pools []domain.RoutablePool) route.RouteImpl {
	return routertesting.WithRoutePools(r, pools)
}

func WithCandidateRoutePools(r route.CandidateRoute, pools []route.CandidatePool) route.CandidateRoute {
	return routertesting.WithCandidateRoutePools(r, pools)
}

func (s *RouterTestSuite) setupDefaultMainnetRouter() (*usecase.Router, map[uint64]domain.TickModel, domain.TakerFeeMap) {
	routerConfig := domain.RouterConfig{
		PreferredPoolIDs:          []uint64{},
		MaxRoutes:                 4,
		MaxPoolsPerRoute:          4,
		MaxSplitIterations:        10,
		MinOSMOLiquidity:          10000,
		RouteUpdateHeightInterval: 0,
		RouteCacheEnabled:         false,
	}

	return s.setupMainnetRouter(routerConfig)
}

func (s *RouterTestSuite) setupMainnetRouter(config domain.RouterConfig) (*routerusecase.Router, map[uint64]domain.TickModel, domain.TakerFeeMap) {
	pools, tickMap, err := parsing.ReadPools(relativePathMainnetFiles + poolsFileName)
	s.Require().NoError(err)

	takerFeeMap, err := parsing.ReadTakerFees(relativePathMainnetFiles + takerFeesFileName)
	s.Require().NoError(err)

	logger, err := log.NewLogger(false, "", "info")
	s.Require().NoError(err)
	router := routerusecase.NewRouter(config.PreferredPoolIDs, config.MaxPoolsPerRoute, config.MaxRoutes, config.MaxSplitRoutes, config.MaxSplitIterations, config.MinOSMOLiquidity, logger)
	router = routerusecase.WithSortedPools(router, pools)

	return router, tickMap, takerFeeMap
}
