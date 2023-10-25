package usecase_test

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/coinutil"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/log"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase"
	routerusecase "github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

// TODO: copy exists in candidate_routes_test.go - share & reuse
var (
	defaultPool = &mockPool{
		ID:                   1,
		denoms:               []string{denomNum(1), denomNum(2)},
		totalValueLockedUSDC: osmomath.NewInt(10),
		poolType:             poolmanagertypes.Balancer,
	}
	emptyRoute = &routerusecase.RouteImpl{}
)

// This test validates that we are able to split over multiple routes.
// We know that higher liquidity pools should be more optimal than the lower liquidity pools
// all else equal.
// Given that, we initialize several pools with different amounts of liquidity.
// We define an expected order of the amountsIn and Out for the split routes based on liquidity.
// Lastly, we assert that the actual order of the split routes is the same as the expected order.
func (s *RouterTestSuite) TestGetBestSplitRoutesQuote() {
	type routeWithOrder struct {
		route domain.SplitRoute
		order int
	}

	s.Setup()

	xLiquidity := sdk.NewCoins(
		sdk.NewCoin(denomNum(1), sdk.NewInt(1_000_000_000_000)),
		sdk.NewCoin(denomNum(2), sdk.NewInt(2_000_000_000_000)),
	)

	// X Liquidity
	defaultBalancerPoolID := s.PrepareBalancerPoolWithCoins(xLiquidity...)

	// 2X liquidity
	// Note that the second pool has more liquidity than the first so it should be preferred
	secondBalancerPoolIDSameDenoms := s.PrepareBalancerPoolWithCoins(coinutil.MulRaw(xLiquidity, 2)...)

	// 4X liquidity
	// Note that the third pool has more liquidity than first and second so it should be preferred
	thirdBalancerPoolIDSameDenoms := s.PrepareBalancerPoolWithCoins(coinutil.MulRaw(xLiquidity, 4)...)

	// Get the defaultBalancerPool from the store
	defaultBalancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, defaultBalancerPoolID)
	s.Require().NoError(err)

	// Get the secondBalancerPool from the store
	secondBalancerPoolSameDenoms, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, secondBalancerPoolIDSameDenoms)
	s.Require().NoError(err)

	// Get the thirdBalancerPool from the store
	thirdBalancerPoolSameDenoms, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, thirdBalancerPoolIDSameDenoms)
	s.Require().NoError(err)

	tests := map[string]struct {
		maxSplitIterations int

		routes      []domain.Route
		tokenIn     sdk.Coin
		expectError error

		expectedTokenOutDenom string

		// Ascending order in terms of which route is preferred
		// and uses the largest amount of the token in
		expectedProportionInOrder []int
	}{
		"valid single route": {
			routes: []domain.Route{
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withChainPoolModel(withTokenOutDenom(defaultPool, denomNum(1)), defaultBalancerPool),
				})},
			tokenIn: sdk.NewCoin(denomNum(2), sdk.NewInt(100)),

			expectedTokenOutDenom: denomNum(1),

			expectedProportionInOrder: []int{0},
		},
		"valid two route single hop": {
			routes: []domain.Route{
				// Route 1
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withChainPoolModel(withTokenOutDenom(defaultPool, denomNum(1)), defaultBalancerPool),
				}),

				// Route 2
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withPoolID(withChainPoolModel(withTokenOutDenom(defaultPool, denomNum(1)), secondBalancerPoolSameDenoms), 2),
				}),
			},

			maxSplitIterations: 10,

			tokenIn: sdk.NewCoin(denomNum(2), sdk.NewInt(5_000_000)),

			expectedTokenOutDenom: denomNum(1),

			// Route 2 is preferred because it has 2x the liquidity of Route 1
			expectedProportionInOrder: []int{0, 1},
		},
		"valid three route single hop": {
			routes: []domain.Route{
				// Route 1
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withChainPoolModel(withTokenOutDenom(defaultPool, denomNum(1)), defaultBalancerPool),
				}),

				// Route 2
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withPoolID(withChainPoolModel(withTokenOutDenom(defaultPool, denomNum(1)), thirdBalancerPoolSameDenoms), 3),
				}),

				// Route 3
				withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
					withPoolID(withChainPoolModel(withTokenOutDenom(defaultPool, denomNum(1)), secondBalancerPoolSameDenoms), 2),
				}),
			},

			maxSplitIterations: 17,

			tokenIn: sdk.NewCoin(denomNum(2), sdk.NewInt(56_789_321)),

			expectedTokenOutDenom: denomNum(1),

			// Route 2 is preferred because it has 4x the liquidity of Route 1
			// and 2X the liquidity of Route 3
			expectedProportionInOrder: []int{2, 0, 1},
		},

		// TODO: cover error cases
		// TODO: multi route multi hop
		// TODO: assert that split ratios are correct
	}

	for name, tc := range tests {
		s.Run(name, func() {

			logger, err := log.NewLogger(false)
			s.Require().NoError(err)

			r := routerusecase.NewRouter([]uint64{}, []domain.PoolI{}, 0, 0, tc.maxSplitIterations, logger)

			quote, err := r.GetBestSplitRoutesQuote(tc.routes, tc.tokenIn)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(tc.expectError, err)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.tokenIn, quote.GetAmountIn())

			quoteCoinOut := quote.GetAmountOut()
			// We only validate that some amount is returned. The correctness of the amount is to be calculated at a different level
			// of abstraction.
			s.Require().NotNil(quoteCoinOut)

			// Validate that amounts in in the quote split routes add up to the original amount in
			routes := quote.GetRoute()
			actualTotalFromSplits := sdk.ZeroInt()
			for _, splitRoute := range routes {
				actualTotalFromSplits = actualTotalFromSplits.Add(splitRoute.GetAmountIn())
			}

			s.Require().Equal(tc.tokenIn.Amount, actualTotalFromSplits)

			// Route must not be nil
			actualRoutes := quote.GetRoute()
			s.Require().NotNil(actualRoutes)

			s.Require().Equal(len(tc.expectedProportionInOrder), len(actualRoutes))

			routesWithOrder := make([]routeWithOrder, len(actualRoutes))
			for i, route := range actualRoutes {
				routesWithOrder[i] = routeWithOrder{
					route: route,
					order: tc.expectedProportionInOrder[i],
				}
			}

			// Sort actual routes in the expected order
			// and assert that the token in is strictly decreasing
			sort.Slice(routesWithOrder, func(i, j int) bool {
				return routesWithOrder[i].order < routesWithOrder[j].order
			})
			s.Require().NotEmpty(routesWithOrder)

			// Iterate over sorted routes with order and validate that the
			// amounts in and out are strictly decreasing as per the expected order.
			previousRouteAmountIn := routesWithOrder[0].route.GetAmountIn()
			previousRouteAmountOut := routesWithOrder[0].route.GetAmountOut()
			for i := 1; i < len(routesWithOrder)-1; i++ {
				currentRouteAmountIn := routesWithOrder[i].route.GetAmountIn()
				currentRouteAmountOut := routesWithOrder[i].route.GetAmountOut()

				// Both in and out amounts must be strictly decreasing
				s.Require().True(previousRouteAmountIn.GT(currentRouteAmountIn))
				s.Require().True(previousRouteAmountOut.GT(currentRouteAmountOut))
			}
		})
	}
}

// This test ensures strict route validation.
// See individual test cases for details.
func (s *RouterTestSuite) TestValidateAndFilterRoutes() {
	tests := map[string]struct {
		routes         []domain.Route
		tokenInDenom   string
		expectError    error
		expectFiltered bool
	}{
		"valid single route single hop": {
			routes: []domain.Route{
				withRoutePools(emptyRoute, []domain.RoutablePool{
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(1), denomNum(3)}), denomNum(3)),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(2), denomNum(3)}), denomNum(2)),
				}),
			},

			tokenInDenom: denomNum(1),
		},
		"valid single route multi-hop": {
			routes: []domain.Route{
				withRoutePools(emptyRoute, []domain.RoutablePool{
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(1), denomNum(3)}), denomNum(3)),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(2), denomNum(3)}), denomNum(2)),
				}),
			},

			tokenInDenom: denomNum(1),
		},
		"valid multi route": {
			routes: []domain.Route{
				withRoutePools(emptyRoute, []domain.RoutablePool{
					withTokenOutDenom(defaultPool, denomNum(2)),
				}),
				withRoutePools(emptyRoute, []domain.RoutablePool{
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(1), denomNum(3)}), denomNum(3)),
					withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(2), denomNum(3)}), denomNum(2)),
				}),
			},

			tokenInDenom: denomNum(1),
		},

		// errors

		"error: no pools in route": {
			routes: []domain.Route{
				withRoutePools(emptyRoute, []domain.RoutablePool{}),
			},

			tokenInDenom: denomNum(2),

			expectError: usecase.NoPoolsInRoute{RouteIndex: 0},
		},
		"error: token out mismatch between multiple routes": {
			routes: []domain.Route{withRoutePools(emptyRoute, []domain.RoutablePool{
				withTokenOutDenom(defaultPool, denomNum(2)),
			}),
				withRoutePools(emptyRoute, []domain.RoutablePool{
					withTokenOutDenom(defaultPool, denomNum(1)),
				}),
			},

			tokenInDenom: denomNum(2),

			expectError: usecase.TokenOutMismatchBetweenRoutesError{TokenOutDenomRouteA: denomNum(2), TokenOutDenomRouteB: denomNum(1)},
		},
		"error: token in matches token out": {
			routes: []domain.Route{withRoutePools(emptyRoute, []domain.RoutablePool{
				withTokenOutDenom(defaultPool, denomNum(1)),
			}),
			},
			tokenInDenom: denomNum(1),

			expectError: usecase.TokenOutDenomMatchesTokenInDenomError{Denom: denomNum(1)},
		},
		"error: token in does not match pool denoms": {
			routes: []domain.Route{withRoutePools(emptyRoute, []domain.RoutablePool{
				withTokenOutDenom(defaultPool, denomNum(1)),
			}),
			},
			tokenInDenom: denomNum(3),

			expectError: usecase.PreviousTokenOutDenomNotInPoolError{RouteIndex: 0, PoolId: defaultPool.GetId(), PreviousTokenOutDenom: denomNum(3)},
		},
		"error: token out does not match pool denoms": {
			routes: []domain.Route{withRoutePools(emptyRoute, []domain.RoutablePool{
				withTokenOutDenom(defaultPool, denomNum(3)),
			}),
			},
			tokenInDenom: denomNum(1),

			expectError: usecase.CurrentTokenOutDenomNotInPoolError{RouteIndex: 0, PoolId: defaultPool.GetId(), CurrentTokenOutDenom: denomNum(3)},
		},

		// Routes filtered
		"filtered: token in is in the route": {
			routes: []domain.Route{withRoutePools(emptyRoute, []domain.RoutablePool{
				withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(1), denomNum(3)}), denomNum(3)),
				withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(2), denomNum(3)}), denomNum(2)),
				withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(2), denomNum(1)}), denomNum(1)),
				withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(1), denomNum(4)}), denomNum(4)),
			}),
			},
			tokenInDenom: denomNum(1),

			expectFiltered: true,
		},
		"filtered: token out is in the route": {
			routes: []domain.Route{withRoutePools(emptyRoute, []domain.RoutablePool{
				withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(1), denomNum(2)}), denomNum(2)),
				withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(2), denomNum(4)}), denomNum(4)),
				withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(4), denomNum(3)}), denomNum(3)),
				withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(3), denomNum(4)}), denomNum(4)),
			}),
			},
			tokenInDenom: denomNum(1),

			expectFiltered: true,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {

			router := routerusecase.NewRouter([]uint64{}, []domain.PoolI{}, 0, 0, 0, &log.NoOpLogger{})

			// TODO: validate filtered routes
			filteredRoutes, err := router.ValidateAndFilterRoutes(tc.routes, tc.tokenInDenom)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}
			s.Require().NoError(err)

			if tc.expectFiltered {
				s.Require().NotEqual(len(tc.routes), len(filteredRoutes))
				s.Require().Len(filteredRoutes, 0)
				return
			}

			s.Require().Equal(len(tc.routes), len(filteredRoutes))

		})
	}
}
