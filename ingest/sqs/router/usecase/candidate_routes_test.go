package usecase_test

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	routerusecase "github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

type mockPool struct {
	ChainPoolModel       poolmanagertypes.PoolI
	TickModel            *domain.TickModel
	ID                   uint64
	denoms               []string
	totalValueLockedUSDC osmomath.Int
	poolType             poolmanagertypes.PoolType
	tokenOutDenom        string
}

var (
	_ domain.PoolI        = &mockPool{}
	_ domain.RoutablePool = &mockPool{}
)

// GetUnderlyingPool implements routerusecase.RoutablePool.
func (mp *mockPool) GetUnderlyingPool() poolmanagertypes.PoolI {
	return mp.ChainPoolModel
}

// GetSQSPoolModel implements domain.PoolI.
func (mp *mockPool) GetSQSPoolModel() domain.SQSPool {
	return domain.SQSPool{
		TotalValueLockedUSDC: mp.totalValueLockedUSDC,
	}
}

// CalculateTokenOutByTokenIn implements routerusecase.RoutablePool.
func (*mockPool) CalculateTokenOutByTokenIn(tokenIn sdk.Coin) (sdk.Coin, error) {
	panic("unimplemented")
}

// String implements domain.RoutablePool.
func (*mockPool) String() string {
	panic("unimplemented")
}

// GetTickModel implements domain.RoutablePool.
func (mp *mockPool) GetTickModel() (*domain.TickModel, error) {
	return mp.TickModel, nil
}

// Validate implements domain.PoolI.
func (*mockPool) Validate(minUOSMOTVL math.Int) error {
	// Note: always valid for tests.
	return nil
}

// GetTokenOutDenom implements routerusecase.RoutablePool.
func (mp *mockPool) GetTokenOutDenom() string {
	return mp.tokenOutDenom
}

var _ domain.PoolI = &mockPool{}
var _ domain.RoutablePool = &mockPool{}

// GetId implements domain.PoolI.
func (mp *mockPool) GetId() uint64 {
	return mp.ID
}

// GetPoolDenoms implements domain.PoolI.
func (mp *mockPool) GetPoolDenoms() []string {
	return mp.denoms
}

// GetTotalValueLockedUOSMO implements domain.PoolI.
func (mp *mockPool) GetTotalValueLockedUOSMO() math.Int {
	return mp.totalValueLockedUSDC
}

// GetType implements domain.PoolI.
func (mp *mockPool) GetType() poolmanagertypes.PoolType {
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

		// Note these are not deep copied.
		ChainPoolModel: mp.ChainPoolModel,
		tokenOutDenom:  mp.tokenOutDenom,
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

func withChainPoolModel(mockPool *mockPool, chainPool poolmanagertypes.PoolI) *mockPool {
	newPool := deepCopyPool(mockPool)
	newPool.ChainPoolModel = chainPool
	return newPool
}

func denomNum(i int) string {
	return fmt.Sprintf("denom%d", i)
}

func withRoutePools(r domain.Route, pools []domain.RoutablePool) domain.Route {
	newRoute := r.DeepCopy()
	for _, pool := range pools {
		newRoute.AddPool(pool, pool.GetTokenOutDenom())
	}
	return newRoute
}

var _ domain.PoolI = &mockPool{}

type routesTestCase struct {
	pools []domain.PoolI

	maxHops   int
	maxRoutes int

	tokenInDenom           string
	tokenOutDenom          string
	currentRoute           domain.Route
	poolsUsed              []bool
	previousTokenOutDenoms []string

	expectedRoutes []domain.Route
	expectedError  error
}

// Tests that find routes is a greedy algorithm where it does not prioritize the best route
// in terms of the number of hops. It prioritizes the first route that it finds via DFS.
func (s *RouterTestSuite) TestFindRoutes() {
	denomOne := denomOne
	denomTwo := denomTwo

	defaultPool := &mockPool{
		ID:                   1,
		denoms:               []string{denomOne, denomTwo},
		totalValueLockedUSDC: osmomath.NewInt(10),
		poolType:             poolmanagertypes.Balancer,
	}

	tests := map[string]routesTestCase{
		"no pools -> no routes": {
			pools: []domain.PoolI{},

			tokenInDenom:   denomOne,
			tokenOutDenom:  denomTwo,
			currentRoute:   &routerusecase.RouteImpl{},
			poolsUsed:      []bool{},
			expectedRoutes: []domain.Route{},
		},
		"one pool; tokens in & out match -> route created": {
			pools: []domain.PoolI{
				defaultPool,
			},

			maxHops:   1,
			maxRoutes: 1,

			tokenInDenom:  denomOne,
			tokenOutDenom: denomTwo,
			currentRoute:  &routerusecase.RouteImpl{},
			poolsUsed:     []bool{false},
			expectedRoutes: []domain.Route{
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{withTokenOutDenom(defaultPool, denomTwo)}),
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
			currentRoute:   &routerusecase.RouteImpl{},
			poolsUsed:      []bool{false},
			expectedRoutes: []domain.Route{},
		},
		"one pool; tokens in & out match but max router stops route from being found": {
			pools: []domain.PoolI{
				defaultPool,
			},

			maxHops:   3,
			maxRoutes: 0,

			tokenInDenom:   denomOne,
			tokenOutDenom:  denomTwo,
			currentRoute:   &routerusecase.RouteImpl{},
			poolsUsed:      []bool{false},
			expectedRoutes: []domain.Route{},
		},
		"one pool; token out does not match -> no route": {
			pools: []domain.PoolI{
				defaultPool,
			},

			maxHops:   1,
			maxRoutes: 1,

			tokenInDenom:   denomOne,
			tokenOutDenom:  denomThree,
			currentRoute:   &routerusecase.RouteImpl{},
			poolsUsed:      []bool{false},
			expectedRoutes: []domain.Route{},
		},
		"one pool; token in does not match -> no route": {
			pools: []domain.PoolI{
				defaultPool,
			},

			maxHops:   1,
			maxRoutes: 1,

			tokenInDenom:   denomThree,
			tokenOutDenom:  denomTwo,
			currentRoute:   &routerusecase.RouteImpl{},
			poolsUsed:      []bool{false},
			expectedRoutes: []domain.Route{},
		},
		"two pools; valid 2 hop route": {
			pools: []domain.PoolI{
				defaultPool,
				withDenoms(defaultPool, []string{denomTwo, denomThree}),
			},

			maxHops:   2,
			maxRoutes: 1,

			tokenInDenom:  denomOne,
			tokenOutDenom: denomThree,
			currentRoute:  &routerusecase.RouteImpl{},
			poolsUsed:     []bool{false, false},
			expectedRoutes: []domain.Route{
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withTokenOutDenom(defaultPool, denomTwo),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
				}),
			},
		},
		"two pools; max hops of one does not let route to be found": {
			pools: []domain.PoolI{
				defaultPool,
				withDenoms(defaultPool, []string{denomTwo, denomThree}),
			},

			maxHops:   1,
			maxRoutes: 1,

			tokenInDenom:   denomOne,
			tokenOutDenom:  denomThree,
			currentRoute:   &routerusecase.RouteImpl{},
			poolsUsed:      []bool{false, false},
			expectedRoutes: []domain.Route{},
		},
		"4 pools; valid 4 hop route (not in order)": {
			pools: []domain.PoolI{
				defaultPool, // A: denom 1, 2
				withDenoms(defaultPool, []string{denomTwo, denomThree}), // B: denom 2, 3
				withDenoms(defaultPool, []string{denomFour, denomOne}),  // C: denom 4, 1
				withDenoms(defaultPool, []string{denomFour, denomFive}), // D: denom 4, 5
			},

			maxHops:   4,
			maxRoutes: 1,

			// D (denom5 for denom4) -> C (denom4 for denom1) -> A (denom1 for denom2) -> B (denom2 for denom3)
			tokenInDenom:  denomFive,
			tokenOutDenom: denomThree,
			currentRoute:  &routerusecase.RouteImpl{},
			poolsUsed:     []bool{false, false, false, false},
			expectedRoutes: []domain.Route{
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withTokenOutDenom(withDenoms(defaultPool, []string{denomFour, denomFive}), denomFour),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomFour, denomOne}), denomOne),
					withTokenOutDenom(defaultPool, denomTwo),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
				}),
			},
		},
		"2 routes; direct and 2 hop": {
			pools: []domain.PoolI{
				defaultPool, // A: denom 1, 2
				withDenoms(defaultPool, []string{denomTwo, denomThree}), // B: denom 2, 3
				withDenoms(defaultPool, []string{denomOne, denomThree}), // C: denom 1, 3
			},

			maxHops:   2,
			maxRoutes: 2,

			// Route 1: A (denom1 for denom2)
			// Route 2: A (denom1 for denom3) -> B (denom3 for denom2)
			tokenInDenom:  denomOne,
			tokenOutDenom: denomTwo,
			currentRoute:  &routerusecase.RouteImpl{},
			poolsUsed:     []bool{false, false, false},
			expectedRoutes: []domain.Route{
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withTokenOutDenom(defaultPool, denomTwo),
				}),

				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withTokenOutDenom(withDenoms(defaultPool, []string{denomOne, denomThree}), denomThree),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomTwo, denomThree}), denomTwo),
				}),
			},
		},
		"routes: first over 4 hops, second over 1 hop. Second is subroute of first. Token in in intermediary path. Make sure second one is not filtered out": {
			pools: []domain.PoolI{
				withDenoms(defaultPool, []string{denomOne, denomThree}),  // A: denom 1, 3
				withDenoms(defaultPool, []string{denomThree, denomFour}), // B: denom 3, 4
				withDenoms(defaultPool, []string{denomFour, denomOne}),   // C: denom 4, 1
				withDenoms(defaultPool, []string{denomOne, denomTwo}),    // D: denom 1, 2
			},

			maxHops:   4,
			maxRoutes: 2,

			// Route 1: A (denom1 for denom3) -> B (denom3 for denom4) -> C (denom4 for denom1) -> D (denom1 for denom2)
			// Route 2: D(denom1 for denom2)
			//
			// Note that the algorithm detects that in the first route, the A -> B -> C part is obsolete since
			// D can be swapped directly. As a result, it returns duplicate routes.
			tokenInDenom:  denomOne,
			tokenOutDenom: denomTwo,
			currentRoute:  &routerusecase.RouteImpl{},
			poolsUsed:     []bool{false, false, false, false},
			expectedRoutes: []domain.Route{
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{withTokenOutDenom(defaultPool, denomTwo)}),
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{withTokenOutDenom(defaultPool, denomTwo)}),
			},
		},
		"2 possible routes with overlap ": {
			pools: []domain.PoolI{
				withDenoms(defaultPool, []string{denomOne, denomTwo}),    // A: denom 1, 2
				withDenoms(defaultPool, []string{denomTwo, denomThree}),  // B: denom 2, 3
				withDenoms(defaultPool, []string{denomThree, denomFour}), // C: denom 3, 4
				withDenoms(defaultPool, []string{denomFive, denomFour}),  // D: denom 5, 4
				withDenoms(defaultPool, []string{denomThree, denomFive}), // E: denom 3, 5
			},

			maxHops:   4,
			maxRoutes: 2,

			tokenInDenom:  denomOne,
			tokenOutDenom: denomFive,
			currentRoute:  &routerusecase.RouteImpl{},
			poolsUsed:     []bool{false, false, false, false, false},
			// Possible routes:
			// Route 1: A -> B -> C -> D
			// Route 2: A -> B -> E
			//
			// Note that we expect the first one (which is longer) to not be accounted for.
			expectedRoutes: []domain.Route{
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withTokenOutDenom(withDenoms(defaultPool, []string{denomOne, denomTwo}), denomTwo),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomThree, denomFour}), denomFour),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomFive, denomFour}), denomFive),
				}),
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withTokenOutDenom(withDenoms(defaultPool, []string{denomOne, denomTwo}), denomTwo),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomThree, denomFive}), denomFive),
				}),
			},
		},
		"possible routes; overlapping in the beginning but second one is shorter (second not filtered out)": {
			pools: []domain.PoolI{
				withDenoms(defaultPool, []string{denomTwo, denomThree}),            // A: denom 2, 3
				withDenoms(defaultPool, []string{denomThree, denomTwo}),            // B: denom 3, 2
				withDenoms(defaultPool, []string{denomOne, denomTwo}),              // C: denom 1, 2
				withDenoms(defaultPool, []string{denomTwo, denomFour, denomThree}), // D: denom 2, 4, 3
			},

			maxHops:   4,
			maxRoutes: 2,

			tokenInDenom:  denomOne,
			tokenOutDenom: denomFour,
			currentRoute:  &routerusecase.RouteImpl{},
			poolsUsed:     []bool{false, false, false, false},
			// Possible routes:
			// Route 1: C -> A -> B -> D
			// Route 2: C -> A -> D
			expectedRoutes: []domain.Route{
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withTokenOutDenom(withDenoms(defaultPool, []string{denomOne, denomTwo}), denomTwo),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomThree, denomTwo}), denomTwo),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomTwo, denomFour, denomThree}), denomFour),
				}),
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withTokenOutDenom(withDenoms(defaultPool, []string{denomOne, denomTwo}), denomTwo),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomTwo, denomFour, denomThree}), denomFour),
				}),
			},
		},
		// If this test is used with max hops of 10, it will select direct route as the last one.
		"3 routes limit; 4 hop, 4 hop, and 3 hop (better routes not selected)": {
			pools: []domain.PoolI{
				defaultPool, // A: denom 1, 2
				withDenoms(defaultPool, []string{denomTwo, denomThree}),  // B: denom 2, 3
				withDenoms(defaultPool, []string{denomFour, denomSix}),   // C: denom 4, 6
				withDenoms(defaultPool, []string{denomThree, denomFour}), // D: denom 3, 4
				withDenoms(defaultPool, []string{denomOne, denomThree}),  // E: denom 1, 3
				withDenoms(defaultPool, []string{denomThree, denomFive}), // F: denom 3, 5
				withDenoms(defaultPool, []string{denomTwo, denomFour}),   // G: denom 2, 4
				withDenoms(defaultPool, []string{denomOne, denomFive}),   // H: denom 1, 5 // note that direct route is not selected due to max routes
				withDenoms(defaultPool, []string{denomFour, denomFive}),  // I: denom 4, 5
			},

			maxHops:   4,
			maxRoutes: 3,

			// Top 3 routes are selected out:
			// Route 1: A (denom1 for denom2) -> B (denom2 for denom3) -> D (denom3 for denom4) -> I (denom4 for denom5)
			// Route 2: A (denom1 for denom2) -> B (denom2 for denom3) -> E (denom3 for denom1) -> F (denom1 for denom5)
			//    - Note that since F is the direct route, the route is truncated to only have the direct part
			// Route 3: A (denom1 for denom2) -> B (denom2 for denom4) -> I (denom4 for denom5)
			// Route 4: E (denom1 for denom3) -> D (denom3 for denom4) -> I (denom4 for denom5)
			// Route 5: E (denom1 for denom3) -> F (denom3 for denom5) -> G (denom2 for denom4) -> I (denom4 for denom5)
			// Route 6: F (denom1 for denom5)
			tokenInDenom:  denomOne,
			tokenOutDenom: denomFive,
			currentRoute:  &routerusecase.RouteImpl{},
			poolsUsed:     []bool{false, false, false, false, false, false, false, false, false},
			expectedRoutes: []domain.Route{
				// Route 1
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withTokenOutDenom(defaultPool, denomTwo),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomThree, denomFour}), denomFour),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomFour, denomFive}), denomFive),
				}),

				// Route 2
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					// The comments below are left for reference.
					// This is what the route would have been if we did not have detection of obsolete routes.
					// withTokenOutDenom(defaultPool, denomTwo),
					// withTokenOutDenom(withDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
					// withTokenOutDenom(withDenoms(defaultPool, []string{denomOne, denomThree}), denomOne),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomOne, denomFive}), denomFive),
				}),

				// Route 3
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withTokenOutDenom(defaultPool, denomTwo),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomThree, denomFive}), denomFive),
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
			expectedRoutes: []domain.Route{},

			expectedError: routerusecase.ErrNilCurrentRoute,
		},
		"error: sorted pools and pools used mismatch": {
			pools: []domain.PoolI{},

			tokenInDenom:  denomOne,
			tokenOutDenom: denomTwo,
			currentRoute:  &routerusecase.RouteImpl{},
			poolsUsed:     []bool{true, false},

			expectedError: routerusecase.SortedPoolsAndPoolsUsedLengthMismatchError{
				SortedPoolsLen: 0,
				PoolsUsedLen:   2,
			},
		},
		"error: no pools but non empty pools in route": {
			pools: []domain.PoolI{},

			tokenInDenom:  denomOne,
			tokenOutDenom: denomTwo,
			currentRoute:  withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{defaultPool}),
			poolsUsed:     []bool{},

			expectedError: routerusecase.SortedPoolsAndPoolsInRouteLengthMismatchError{
				SortedPoolsLen: 0,
				PoolsInRoute:   1,
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {

			r := routerusecase.NewRouter([]uint64{}, tc.pools, tc.maxHops, tc.maxRoutes, 0, 0, nil)

			routes, err := r.FindRoutes(tc.tokenInDenom, tc.tokenOutDenom, tc.currentRoute, tc.poolsUsed, tc.previousTokenOutDenoms)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError.Error(), err.Error())
				return
			}

			s.Require().NoError(err)

			s.validateRoutes(tc, routes)
		})
	}
}

func (s *RouterTestSuite) TestGetCandidateRoutes() {
	tests := map[string]struct {
		pools []domain.PoolI

		maxHops   int
		maxRoutes int

		tokenInDenom           string
		tokenOutDenom          string
		currentRoute           domain.Route
		poolsUsed              []bool
		previousTokenOutDenoms []string

		expectedRoutes []domain.Route
		expectedError  error
	}{
		"2 possible routes with overlap ": {
			pools: []domain.PoolI{
				withDenoms(defaultPool, []string{denomOne, denomTwo}),                                 // A: denom 1, 2
				withPoolID(withDenoms(defaultPool, []string{denomTwo, denomThree}), defaultPoolID+1),  // B: denom 2, 3
				withPoolID(withDenoms(defaultPool, []string{denomThree, denomFour}), defaultPoolID+2), // C: denom 3, 4
				withPoolID(withDenoms(defaultPool, []string{denomFive, denomFour}), defaultPoolID+3),  // D: denom 5, 4
				withPoolID(withDenoms(defaultPool, []string{denomThree, denomFive}), defaultPoolID+4), // E: denom 3, 5
			},

			maxHops:   4,
			maxRoutes: 2,

			tokenInDenom:  denomOne,
			tokenOutDenom: denomFive,
			currentRoute:  &routerusecase.RouteImpl{},
			poolsUsed:     []bool{false, false, false, false, false},
			// Possible routes:
			// Route 1: A -> B -> C -> D
			// Route 2: A -> B -> E
			//
			// Note that we expect the first one (which is longer) to not be accounted for
			// due to overlapping pool IDs.
			expectedRoutes: []domain.Route{
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withTokenOutDenom(withDenoms(defaultPool, []string{denomOne, denomTwo}), denomTwo),
					withPoolID(withTokenOutDenom(withDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree), defaultPoolID+1),
					withPoolID(withTokenOutDenom(withDenoms(defaultPool, []string{denomThree, denomFive}), denomFive), defaultPoolID+4),
				}),
			},
		},
		// If this test is used with max hops of 10, it will select direct route as the last one.
		"routes limit; 4 hop, 4 hop, and 3 hop (better routes not selected)": {
			pools: []domain.PoolI{
				defaultPool, // A: denom 1, 2
				withPoolID(withDenoms(defaultPool, []string{denomTwo, denomThree}), defaultPoolID+1),  // B: denom 2, 3
				withPoolID(withDenoms(defaultPool, []string{denomFour, denomSix}), defaultPoolID+2),   // C: denom 4, 6
				withPoolID(withDenoms(defaultPool, []string{denomThree, denomFour}), defaultPoolID+3), // D: denom 3, 4
				withPoolID(withDenoms(defaultPool, []string{denomOne, denomThree}), defaultPoolID+4),  // E: denom 1, 3
				withPoolID(withDenoms(defaultPool, []string{denomThree, denomFive}), defaultPoolID+5), // F: denom 3, 5
				withPoolID(withDenoms(defaultPool, []string{denomTwo, denomFour}), defaultPoolID+6),   // G: denom 2, 4
				withPoolID(withDenoms(defaultPool, []string{denomOne, denomFive}), defaultPoolID+7),   // H: denom 1, 5 // note that direct route is not selected due to max routes
				withPoolID(withDenoms(defaultPool, []string{denomFour, denomFive}), defaultPoolID+8),  // I: denom 4, 5
			},

			maxHops:   4,
			maxRoutes: 3,

			// Top 3 routes are selected out:
			// Route 1: A (denom1 for denom2) -> B (denom2 for denom3) -> D (denom3 for denom4) -> I (denom4 for denom5)
			// Route 2: A (denom1 for denom2) -> B (denom2 for denom3) -> E (denom3 for denom1) -> F (denom1 for denom5)
			//    - Note that since F is the direct route, the route is truncated to only have the direct part
			// Route 3: A (denom1 for denom2) -> B (denom2 for denom4) -> I (denom4 for denom5)
			// Route 4: E (denom1 for denom3) -> D (denom3 for denom4) -> I (denom4 for denom5)
			// Route 5: E (denom1 for denom3) -> F (denom3 for denom5) -> G (denom2 for denom4) -> I (denom4 for denom5)
			// Route 6: F (denom1 for denom5)
			tokenInDenom:  denomOne,
			tokenOutDenom: denomFive,
			currentRoute:  &routerusecase.RouteImpl{},
			poolsUsed:     []bool{false, false, false, false, false, false, false, false, false},
			expectedRoutes: []domain.Route{
				// Note that routes get reordered by the number of hops.
				// See similar test in TestFindRoutes for comparison.

				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					// The comments below are left for reference.
					// This is what the route would have been if we did not have detection of obsolete routes.
					// withTokenOutDenom(defaultPool, denomTwo),
					// withTokenOutDenom(withDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
					// withTokenOutDenom(withDenoms(defaultPool, []string{denomOne, denomThree}), denomOne),
					withPoolID(withTokenOutDenom(withDenoms(defaultPool, []string{denomOne, denomFive}), denomFive), defaultPoolID+7),
				}),

				// Route 3
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withTokenOutDenom(defaultPool, denomTwo),
					withPoolID(withTokenOutDenom(withDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree), defaultPoolID+1),
					withPoolID(withTokenOutDenom(withDenoms(defaultPool, []string{denomThree, denomFive}), denomFive), defaultPoolID+5),
				}),

				// Note that the third route is removed. See similar test in TestFindRoutes for comparison.
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {

			r := routerusecase.NewRouter([]uint64{}, tc.pools, tc.maxHops, tc.maxRoutes, 3, 0, nil)

			routes, err := r.GetCandidateRoutes(tc.tokenInDenom, tc.tokenOutDenom)

			if tc.expectedError != nil {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			s.validateRoutes(tc, routes)
		})
	}
}

// validateRoutes validates that the routes are as expected.
func (s *RouterTestSuite) validateRoutes(tc routesTestCase, routes []domain.Route) {
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
}
