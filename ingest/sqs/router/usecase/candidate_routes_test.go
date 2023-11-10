package usecase_test

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain/mocks"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	routerusecase "github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

func denomNum(i int) string {
	return fmt.Sprintf("denom%d", i)
}

func withRoutePools(r domain.Route, pools []domain.RoutablePool) domain.Route {
	newRoute := r.DeepCopy()
	for _, pool := range pools {
		newRoute.AddPool(pool, pool.GetTokenOutDenom(), pool.GetTakerFee())
	}
	return newRoute
}

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

	defaultPool := &mocks.MockRoutablePool{
		ID:                   1,
		Denoms:               []string{denomOne, denomTwo},
		TotalValueLockedUSDC: osmomath.NewInt(10),
		PoolType:             poolmanagertypes.Balancer,
		TakerFee:             osmomath.ZeroDec(),
		SpreadFactor:         osmomath.ZeroDec(),
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
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{mocks.WithTokenOutDenom(defaultPool, denomTwo)}),
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
				mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}),
			},

			maxHops:   2,
			maxRoutes: 1,

			tokenInDenom:  denomOne,
			tokenOutDenom: denomThree,
			currentRoute:  &routerusecase.RouteImpl{},
			poolsUsed:     []bool{false, false},
			expectedRoutes: []domain.Route{
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					mocks.WithTokenOutDenom(defaultPool, denomTwo),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
				}),
			},
		},
		"two pools; max hops of one does not let route to be found": {
			pools: []domain.PoolI{
				defaultPool,
				mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}),
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
				mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}), // B: denom 2, 3
				mocks.WithDenoms(defaultPool, []string{denomFour, denomOne}),  // C: denom 4, 1
				mocks.WithDenoms(defaultPool, []string{denomFour, denomFive}), // D: denom 4, 5
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
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomFour, denomFive}), denomFour),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomFour, denomOne}), denomOne),
					mocks.WithTokenOutDenom(defaultPool, denomTwo),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
				}),
			},
		},
		"2 routes; direct and 2 hop": {
			pools: []domain.PoolI{
				defaultPool, // A: denom 1, 2
				mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}), // B: denom 2, 3
				mocks.WithDenoms(defaultPool, []string{denomOne, denomThree}), // C: denom 1, 3
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
					mocks.WithTokenOutDenom(defaultPool, denomTwo),
				}),

				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomOne, denomThree}), denomThree),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}), denomTwo),
				}),
			},
		},
		"routes: first over 4 hops, second over 1 hop. Second is subroute of first. Token in in intermediary path. Make sure second one is not filtered out": {
			pools: []domain.PoolI{
				mocks.WithDenoms(defaultPool, []string{denomOne, denomThree}),  // A: denom 1, 3
				mocks.WithDenoms(defaultPool, []string{denomThree, denomFour}), // B: denom 3, 4
				mocks.WithDenoms(defaultPool, []string{denomFour, denomOne}),   // C: denom 4, 1
				mocks.WithDenoms(defaultPool, []string{denomOne, denomTwo}),    // D: denom 1, 2
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
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{mocks.WithTokenOutDenom(defaultPool, denomTwo)}),
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{mocks.WithTokenOutDenom(defaultPool, denomTwo)}),
			},
		},
		"2 possible routes with overlap ": {
			pools: []domain.PoolI{
				mocks.WithDenoms(defaultPool, []string{denomOne, denomTwo}),    // A: denom 1, 2
				mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}),  // B: denom 2, 3
				mocks.WithDenoms(defaultPool, []string{denomThree, denomFour}), // C: denom 3, 4
				mocks.WithDenoms(defaultPool, []string{denomFive, denomFour}),  // D: denom 5, 4
				mocks.WithDenoms(defaultPool, []string{denomThree, denomFive}), // E: denom 3, 5
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
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomOne, denomTwo}), denomTwo),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomThree, denomFour}), denomFour),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomFive, denomFour}), denomFive),
				}),
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomOne, denomTwo}), denomTwo),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomThree, denomFive}), denomFive),
				}),
			},
		},
		"possible routes; overlapping in the beginning but second one is shorter (second not filtered out)": {
			pools: []domain.PoolI{
				mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}),            // A: denom 2, 3
				mocks.WithDenoms(defaultPool, []string{denomThree, denomTwo}),            // B: denom 3, 2
				mocks.WithDenoms(defaultPool, []string{denomOne, denomTwo}),              // C: denom 1, 2
				mocks.WithDenoms(defaultPool, []string{denomTwo, denomFour, denomThree}), // D: denom 2, 4, 3
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
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomOne, denomTwo}), denomTwo),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomThree, denomTwo}), denomTwo),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomTwo, denomFour, denomThree}), denomFour),
				}),
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomOne, denomTwo}), denomTwo),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomTwo, denomFour, denomThree}), denomFour),
				}),
			},
		},
		// If this test is used with max hops of 10, it will select direct route as the last one.
		"3 routes limit; 4 hop, 4 hop, and 3 hop (better routes not selected)": {
			pools: []domain.PoolI{
				defaultPool, // A: denom 1, 2
				mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}),  // B: denom 2, 3
				mocks.WithDenoms(defaultPool, []string{denomFour, denomSix}),   // C: denom 4, 6
				mocks.WithDenoms(defaultPool, []string{denomThree, denomFour}), // D: denom 3, 4
				mocks.WithDenoms(defaultPool, []string{denomOne, denomThree}),  // E: denom 1, 3
				mocks.WithDenoms(defaultPool, []string{denomThree, denomFive}), // F: denom 3, 5
				mocks.WithDenoms(defaultPool, []string{denomTwo, denomFour}),   // G: denom 2, 4
				mocks.WithDenoms(defaultPool, []string{denomOne, denomFive}),   // H: denom 1, 5 // note that direct route is not selected due to max routes
				mocks.WithDenoms(defaultPool, []string{denomFour, denomFive}),  // I: denom 4, 5
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
					mocks.WithTokenOutDenom(defaultPool, denomTwo),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomThree, denomFour}), denomFour),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomFour, denomFive}), denomFive),
				}),

				// Route 2
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					// The comments below are left for reference.
					// This is what the route would have been if we did not have detection of obsolete routes.
					// mocks.WithTokenOutDenom(defaultPool, denomTwo),
					// mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
					// mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomOne, denomThree}), denomOne),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomOne, denomFive}), denomFive),
				}),

				// Route 3
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					mocks.WithTokenOutDenom(defaultPool, denomTwo),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomThree, denomFive}), denomFive),
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

			// Get taker fees for all pools.
			takerFees := s.getTakerFeeMapForAllPoolTokenPairs(tc.pools)

			r := routerusecase.NewRouter([]uint64{}, tc.pools, takerFees, tc.maxHops, tc.maxRoutes, 0, 0, nil)

			routes, err := r.FindRoutes(tc.tokenInDenom, tc.tokenOutDenom, tc.currentRoute, tc.poolsUsed, tc.previousTokenOutDenoms)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError.Error(), err.Error())
				return
			}

			s.Require().NoError(err)

			s.validateFoundRoutes(tc, routes)
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
				mocks.WithDenoms(defaultPool, []string{denomOne, denomTwo}),                                 // A: denom 1, 2
				mocks.WithPoolID(mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}), defaultPoolID+1),  // B: denom 2, 3
				mocks.WithPoolID(mocks.WithDenoms(defaultPool, []string{denomThree, denomFour}), defaultPoolID+2), // C: denom 3, 4
				mocks.WithPoolID(mocks.WithDenoms(defaultPool, []string{denomFive, denomFour}), defaultPoolID+3),  // D: denom 5, 4
				mocks.WithPoolID(mocks.WithDenoms(defaultPool, []string{denomThree, denomFive}), defaultPoolID+4), // E: denom 3, 5
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
					mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomOne, denomTwo}), denomTwo),
					mocks.WithPoolID(mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree), defaultPoolID+1),
					mocks.WithPoolID(mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomThree, denomFive}), denomFive), defaultPoolID+4),
				}),
			},
		},
		// If this test is used with max hops of 10, it will select direct route as the last one.
		"routes limit; 4 hop, 4 hop, and 3 hop (better routes not selected)": {
			pools: []domain.PoolI{
				defaultPool, // A: denom 1, 2
				mocks.WithPoolID(mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}), defaultPoolID+1),  // B: denom 2, 3
				mocks.WithPoolID(mocks.WithDenoms(defaultPool, []string{denomFour, denomSix}), defaultPoolID+2),   // C: denom 4, 6
				mocks.WithPoolID(mocks.WithDenoms(defaultPool, []string{denomThree, denomFour}), defaultPoolID+3), // D: denom 3, 4
				mocks.WithPoolID(mocks.WithDenoms(defaultPool, []string{denomOne, denomThree}), defaultPoolID+4),  // E: denom 1, 3
				mocks.WithPoolID(mocks.WithDenoms(defaultPool, []string{denomThree, denomFive}), defaultPoolID+5), // F: denom 3, 5
				mocks.WithPoolID(mocks.WithDenoms(defaultPool, []string{denomTwo, denomFour}), defaultPoolID+6),   // G: denom 2, 4
				mocks.WithPoolID(mocks.WithDenoms(defaultPool, []string{denomOne, denomFive}), defaultPoolID+7),   // H: denom 1, 5 // note that direct route is not selected due to max routes
				mocks.WithPoolID(mocks.WithDenoms(defaultPool, []string{denomFour, denomFive}), defaultPoolID+8),  // I: denom 4, 5
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
					// mocks.WithTokenOutDenom(defaultPool, denomTwo),
					// mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree),
					// mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomOne, denomThree}), denomOne),
					mocks.WithPoolID(mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomOne, denomFive}), denomFive), defaultPoolID+7),
				}),

				// Route 3
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					mocks.WithTokenOutDenom(defaultPool, denomTwo),
					mocks.WithPoolID(mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomTwo, denomThree}), denomThree), defaultPoolID+1),
					mocks.WithPoolID(mocks.WithTokenOutDenom(mocks.WithDenoms(defaultPool, []string{denomThree, denomFive}), denomFive), defaultPoolID+5),
				}),

				// Note that the third route is removed. See similar test in TestFindRoutes for comparison.
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {

			takerFees := s.getTakerFeeMapForAllPoolTokenPairs(tc.pools)

			r := routerusecase.NewRouter([]uint64{}, tc.pools, takerFees, tc.maxHops, tc.maxRoutes, 3, 0, nil)

			routes, err := r.GetCandidateRoutes(tc.tokenInDenom, tc.tokenOutDenom)

			if tc.expectedError != nil {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			s.validateFoundRoutes(tc, routes)
		})
	}
}

// validateFoundRoutes validates that the routes are as expected.
func (s *RouterTestSuite) validateFoundRoutes(tc routesTestCase, routes []domain.Route) {
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
