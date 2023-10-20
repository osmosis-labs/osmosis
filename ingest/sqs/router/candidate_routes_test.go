package router_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/app/apptesting"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
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
		preferredPoolIDs = []uint64{allPool.BalancerPoolID, allPool.StableSwapPoolID}
		maxHops          = 3
		maxRoutes        = 5
		defaultAllPools  = []domain.PoolI{
			&domain.Pool{
				UnderlyingPool:       balancerPool,
				TotalValueLockedUSDC: osmomath.NewInt(5), // 5
			},
			&domain.Pool{
				UnderlyingPool:       stableswapPool,
				TotalValueLockedUSDC: osmomath.OneInt(),
			},
			&domain.Pool{
				UnderlyingPool:       concentratedPool,
				TotalValueLockedUSDC: osmomath.NewInt(4), // 4
			},
			&domain.Pool{
				UnderlyingPool:       cosmWasmPool,
				TotalValueLockedUSDC: osmomath.NewInt(3), // 3
			},
		}

		// Expected
		// First, preferred pool IDs, then sorted by TVL.
		expectedSortedPoolIDs = []uint64{allPool.BalancerPoolID, allPool.StableSwapPoolID, allPool.ConcentratedPoolID, allPool.CosmWasmPoolID}
	)

	// System under test
	router := router.NewRouter(preferredPoolIDs, defaultAllPools, maxHops, maxRoutes)

	// Assert
	s.Require().Equal(maxHops, router.GetMaxHops())
	s.Require().Equal(maxRoutes, router.GetMaxRoutes())
	s.Require().Equal(expectedSortedPoolIDs, router.GetSortedPoolIDs())
}

type mockPool struct {
	ID                   uint64
	denoms               []string
	totalValueLockedUSDC osmomath.Int
	poolType             types.PoolType
	tokenOutDenom        string
}

// GetTokenOutDenom implements router.RoutablePool.
func (mp *mockPool) GetTokenOutDenom() string {
	return mp.tokenOutDenom
}

var _ domain.PoolI = &mockPool{}
var _ router.RoutablePool = &mockPool{}

// GetId implements domain.PoolI.
func (mp *mockPool) GetId() uint64 {
	return mp.ID
}

// GetPoolDenoms implements domain.PoolI.
func (mp *mockPool) GetPoolDenoms() []string {
	return mp.denoms
}

// GetTotalValueLockedUSDC implements domain.PoolI.
func (mp *mockPool) GetTotalValueLockedUSDC() math.Int {
	return mp.totalValueLockedUSDC
}

// GetType implements domain.PoolI.
func (mp *mockPool) GetType() types.PoolType {
	return mp.poolType
}

func deepCopyPool(mp *mockPool) *mockPool {

	newDenoms := make([]string, len(mp.denoms))
	copy(newDenoms, mp.denoms)

	newTotalValueLocker := osmomath.NewIntFromBigInt(mp.totalValueLockedUSDC.BigInt())

	return &mockPool{
		ID:                   mp.ID,
		denoms:               newDenoms,
		totalValueLockedUSDC: newTotalValueLocker,
		poolType:             mp.poolType,
	}
}

func withPoolID(mockPool *mockPool, id uint64) *mockPool {
	newPool := deepCopyPool(mockPool)
	newPool.ID = id
	return newPool
}

func withDenoms(mockPool *mockPool, denoms []string) *mockPool {
	newPool := deepCopyPool(mockPool)
	newPool.denoms = denoms
	return newPool
}

func withTokenOutDenom(mockPool *mockPool, tokenOutDenom string) *mockPool {
	newPool := deepCopyPool(mockPool)
	newPool.tokenOutDenom = tokenOutDenom
	return newPool
}

func denomNum(i int) string {
	return fmt.Sprintf("denom%d", i)
}

func withRoutePools(r router.Route, pools []router.RoutablePool) router.Route {
	newRoute := r.DeepCopy()
	for _, pool := range pools {
		newRoute.AddPool(pool, pool.GetTokenOutDenom())
	}
	return newRoute
}

var _ domain.PoolI = &mockPool{}

// Tests that find routes is a greedy algorithm where it does not prioritize the best route
// in terms of the number of hops. It prioritizes the first route that it finds via DFS.
func (s *RouterTestSuite) TestFindRoutes() {
	denomOne := denomNum(1)
	denomTwo := denomNum(2)

	defaultPool := &mockPool{
		ID:                   1,
		denoms:               []string{denomNum(1), denomNum(2)},
		totalValueLockedUSDC: osmomath.NewInt(10),
		poolType:             types.Balancer,
	}

	tests := map[string]struct {
		pools []domain.PoolI

		maxHops   int
		maxRoutes int

		tokenInDenom           string
		tokenOutDenom          string
		currentRoute           router.Route
		poolsUsed              []bool
		previousTokenOutDenoms []string

		expectedRoutes []router.Route
		expectedError  error
	}{
		"no pools -> no routes": {
			pools: []domain.PoolI{},

			tokenInDenom:   denomOne,
			tokenOutDenom:  denomTwo,
			currentRoute:   &router.RouteImpl{},
			poolsUsed:      []bool{},
			expectedRoutes: []router.Route{},
		},
		"one pool; tokens in & out match -> route created": {
			pools: []domain.PoolI{
				defaultPool,
			},

			maxHops:   1,
			maxRoutes: 1,

			tokenInDenom:  denomOne,
			tokenOutDenom: denomTwo,
			currentRoute:  &router.RouteImpl{},
			poolsUsed:     []bool{false},
			expectedRoutes: []router.Route{
				withRoutePools(&router.RouteImpl{}, []router.RoutablePool{withTokenOutDenom(defaultPool, denomTwo)}),
			},
		},
		"one pool; tokens in & out match but max hops stops route from being found": {
			pools: []domain.PoolI{
				defaultPool,
			},

			maxHops:   0,
			maxRoutes: 3,

			tokenInDenom:   denomOne,
			tokenOutDenom:  denomTwo,
			currentRoute:   &router.RouteImpl{},
			poolsUsed:      []bool{false},
			expectedRoutes: []router.Route{},
		},
		"one pool; tokens in & out match but max router stops route from being found": {
			pools: []domain.PoolI{
				defaultPool,
			},

			maxHops:   3,
			maxRoutes: 0,

			tokenInDenom:   denomOne,
			tokenOutDenom:  denomTwo,
			currentRoute:   &router.RouteImpl{},
			poolsUsed:      []bool{false},
			expectedRoutes: []router.Route{},
		},
		"one pool; token out does not match -> no route": {
			pools: []domain.PoolI{
				defaultPool,
			},

			maxHops:   1,
			maxRoutes: 1,

			tokenInDenom:   denomOne,
			tokenOutDenom:  denomNum(3),
			currentRoute:   &router.RouteImpl{},
			poolsUsed:      []bool{false},
			expectedRoutes: []router.Route{},
		},
		"one pool; token in does not match -> no route": {
			pools: []domain.PoolI{
				defaultPool,
			},

			maxHops:   1,
			maxRoutes: 1,

			tokenInDenom:   denomNum(3),
			tokenOutDenom:  denomTwo,
			currentRoute:   &router.RouteImpl{},
			poolsUsed:      []bool{false},
			expectedRoutes: []router.Route{},
		},
		"two pools; valid 2 hop route": {
			pools: []domain.PoolI{
				defaultPool,
				withDenoms(defaultPool, []string{denomNum(2), denomNum(3)}),
			},

			maxHops:   2,
			maxRoutes: 1,

			tokenInDenom:  denomOne,
			tokenOutDenom: denomNum(3),
			currentRoute:  &router.RouteImpl{},
			poolsUsed:     []bool{false, false},
			expectedRoutes: []router.Route{
				withRoutePools(&router.RouteImpl{}, []router.RoutablePool{
					withTokenOutDenom(defaultPool, denomTwo),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(2), denomNum(3)}), denomNum(3)),
				}),
			},
		},
		"two pools; max hops of one does not let route to be found": {
			pools: []domain.PoolI{
				defaultPool,
				withDenoms(defaultPool, []string{denomNum(2), denomNum(3)}),
			},

			maxHops:   1,
			maxRoutes: 1,

			tokenInDenom:   denomOne,
			tokenOutDenom:  denomNum(3),
			currentRoute:   &router.RouteImpl{},
			poolsUsed:      []bool{false, false},
			expectedRoutes: []router.Route{},
		},
		"4 pools; valid 4 hop route (not in order)": {
			pools: []domain.PoolI{
				defaultPool, // A: denom 1, 2
				withDenoms(defaultPool, []string{denomNum(2), denomNum(3)}), // B: denom 2, 3
				withDenoms(defaultPool, []string{denomNum(4), denomNum(1)}), // C: denom 4, 1
				withDenoms(defaultPool, []string{denomNum(4), denomNum(5)}), // D: denom 4, 5
			},

			maxHops:   4,
			maxRoutes: 1,

			// D (denom5 for denom4) -> C (denom4 for denom1) -> A (denom1 for denom2) -> B (denom2 for denom3)
			tokenInDenom:  denomNum(5),
			tokenOutDenom: denomNum(3),
			currentRoute:  &router.RouteImpl{},
			poolsUsed:     []bool{false, false, false, false},
			expectedRoutes: []router.Route{
				withRoutePools(&router.RouteImpl{}, []router.RoutablePool{
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(4), denomNum(5)}), denomNum(4)),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(4), denomNum(1)}), denomNum(1)),
					withTokenOutDenom(defaultPool, denomTwo),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(2), denomNum(3)}), denomNum(3)),
				}),
			},
		},
		"2 routes; direct and 2 hop": {
			pools: []domain.PoolI{
				defaultPool, // A: denom 1, 2
				withDenoms(defaultPool, []string{denomNum(2), denomNum(3)}), // B: denom 2, 3
				withDenoms(defaultPool, []string{denomNum(1), denomNum(3)}), // C: denom 1, 3
			},

			maxHops:   2,
			maxRoutes: 2,

			// Route 1: A (denom1 for denom2)
			// Route 2: A (denom1 for denom3) -> B (denom3 for denom2)
			tokenInDenom:  denomNum(1),
			tokenOutDenom: denomNum(2),
			currentRoute:  &router.RouteImpl{},
			poolsUsed:     []bool{false, false, false},
			expectedRoutes: []router.Route{
				withRoutePools(&router.RouteImpl{}, []router.RoutablePool{
					withTokenOutDenom(defaultPool, denomTwo),
				}),

				withRoutePools(&router.RouteImpl{}, []router.RoutablePool{
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(1), denomNum(3)}), denomNum(3)),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(2), denomNum(3)}), denomNum(2)),
				}),
			},
		},
		// If this test is used with max hops of 10, it will select direct route as the last one.
		"3 routes limit; 4 hop, 4 hop, and 3 hop (better routes not selected)": {
			pools: []domain.PoolI{
				defaultPool, // A: denom 1, 2
				withDenoms(defaultPool, []string{denomNum(2), denomNum(3)}), // B: denom 2, 3
				withDenoms(defaultPool, []string{denomNum(4), denomNum(6)}), // C: denom 4, 6
				withDenoms(defaultPool, []string{denomNum(3), denomNum(4)}), // D: denom 3, 4
				withDenoms(defaultPool, []string{denomNum(1), denomNum(3)}), // E: denom 1, 3
				withDenoms(defaultPool, []string{denomNum(3), denomNum(5)}), // F: denom 3, 5
				withDenoms(defaultPool, []string{denomNum(2), denomNum(4)}), // G: denom 2, 4
				withDenoms(defaultPool, []string{denomNum(1), denomNum(5)}), // H: denom 1, 5 // note that direct route is not selected due to max routes
				withDenoms(defaultPool, []string{denomNum(4), denomNum(5)}), // I: denom 4, 5
			},

			maxHops:   4,
			maxRoutes: 3,

			// Top 3 routes are selected out:
			// Route 1: A (denom1 for denom2) -> B (denom2 for denom3) -> D (denom3 for denom4) -> I (denom4 for denom5)
			// Route 2: A (denom1 for denom2) -> B (denom2 for denom3) -> E (denom3 for denom1) -> F (denom1 for denom5)
			// Route 3: A (denom1 for denom2) -> B (denom2 for denom4) -> I (denom4 for denom5)
			// Route 4: E (denom1 for denom3) -> D (denom3 for denom4) -> I (denom4 for denom5)
			// Route 5: E (denom1 for denom3) -> F (denom3 for denom5) -> G (denom2 for denom4) -> I (denom4 for denom5)
			// Route 6: F (denom1 for denom5)
			tokenInDenom:  denomNum(1),
			tokenOutDenom: denomNum(5),
			currentRoute:  &router.RouteImpl{},
			poolsUsed:     []bool{false, false, false, false, false, false, false, false, false},
			expectedRoutes: []router.Route{
				// Route 1
				withRoutePools(&router.RouteImpl{}, []router.RoutablePool{
					withTokenOutDenom(defaultPool, denomTwo),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(2), denomNum(3)}), denomNum(3)),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(3), denomNum(4)}), denomNum(4)),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(4), denomNum(5)}), denomNum(5)),
				}),

				// Route 2
				withRoutePools(&router.RouteImpl{}, []router.RoutablePool{
					withTokenOutDenom(defaultPool, denomTwo),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(2), denomNum(3)}), denomNum(3)),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(1), denomNum(3)}), denomNum(1)),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(1), denomNum(5)}), denomNum(5)),
				}),

				// Route 3
				withRoutePools(&router.RouteImpl{}, []router.RoutablePool{
					withTokenOutDenom(defaultPool, denomTwo),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(2), denomNum(3)}), denomNum(3)),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(3), denomNum(5)}), denomNum(5)),
				}),
			},
		},
		// errors
		"error: nil route": {
			pools: []domain.PoolI{},

			tokenInDenom:   denomOne,
			tokenOutDenom:  denomTwo,
			currentRoute:   nil,
			poolsUsed:      []bool{},
			expectedRoutes: []router.Route{},

			expectedError: router.ErrNilCurrentRoute,
		},
		"error: sorted pools and pools used mismatch": {
			pools: []domain.PoolI{},

			tokenInDenom:  denomOne,
			tokenOutDenom: denomTwo,
			currentRoute:  &router.RouteImpl{},
			poolsUsed:     []bool{true, false},

			expectedError: router.SortedPoolsAndPoolsUsedLengthMismatchError{
				SortedPoolsLen: 0,
				PoolsUsedLen:   2,
			},
		},
		"error: no pools but non empty pools in route": {
			pools: []domain.PoolI{},

			tokenInDenom:  denomOne,
			tokenOutDenom: denomTwo,
			currentRoute:  withRoutePools(&router.RouteImpl{}, []router.RoutablePool{defaultPool}),
			poolsUsed:     []bool{},

			expectedError: router.SortedPoolsAndPoolsInRouteLengthMismatchError{
				SortedPoolsLen: 0,
				PoolsInRoute:   1,
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {

			r := router.NewRouter([]uint64{}, tc.pools, tc.maxHops, tc.maxRoutes)

			routes, err := r.FindRoutes(tc.tokenInDenom, tc.tokenOutDenom, tc.currentRoute, tc.poolsUsed, tc.previousTokenOutDenoms)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError.Error(), err.Error())
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(len(tc.expectedRoutes), len(routes))

			for i, expectedRoute := range tc.expectedRoutes {
				actualRoute := routes[i]

				expectedPools := expectedRoute.GetPools()
				actualPools := actualRoute.GetPools()

				s.Require().Equal(len(expectedPools), len(actualPools))

				for j, expectedPool := range expectedPools {
					s.Require().Equal(expectedPool.GetId(), actualPools[j].GetId())
					s.Require().Equal(expectedPool.GetTokenOutDenom(), actualPools[j].GetTokenOutDenom())
					s.Require().Equal(expectedPool.GetPoolDenoms(), actualPools[j].GetPoolDenoms())
				}
			}
		})
	}
}
