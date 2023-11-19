package usecase_test

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/coinutil"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain/mocks"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/log"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase"
	routerusecase "github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase/route"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/router/usecase/routertesting"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

const defaultPoolID = uint64(1)

// TODO: copy exists in candidate_routes_test.go - share & reuse
var (
	DefaultTakerFee     = osmomath.MustNewDecFromStr("0.002")
	DefaultPoolBalances = sdk.NewCoins(
		sdk.NewCoin(DenomOne, DefaultAmt0),
		sdk.NewCoin(DenomTwo, DefaultAmt1),
	)
	DefaultSpreadFactor = osmomath.MustNewDecFromStr("0.005")

	DefaultPool = &mocks.MockRoutablePool{
		ID:                   defaultPoolID,
		Denoms:               []string{DenomOne, DenomTwo},
		TotalValueLockedUSDC: osmomath.NewInt(10),
		PoolType:             poolmanagertypes.Balancer,
		Balances:             DefaultPoolBalances,
		TakerFee:             DefaultTakerFee,
		SpreadFactor:         DefaultSpreadFactor,
	}
	EmptyRoute = &route.RouteImpl{}

	// Test denoms
	DenomOne   = routertesting.DenomOne
	DenomTwo   = routertesting.DenomTwo
	DenomThree = routertesting.DenomThree
	DenomFour  = routertesting.DenomFour
	DenomFive  = routertesting.DenomFive
	DenomSix   = routertesting.DenomSix
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
		sdk.NewCoin(DenomOne, sdk.NewInt(1_000_000_000_000)),
		sdk.NewCoin(DenomTwo, sdk.NewInt(2_000_000_000_000)),
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
				WithRoutePools(&route.RouteImpl{}, []domain.RoutablePool{
					mocks.WithChainPoolModel(mocks.WithTokenOutDenom(DefaultPool, DenomOne), defaultBalancerPool),
				})},
			tokenIn: sdk.NewCoin(DenomTwo, sdk.NewInt(100)),

			expectedTokenOutDenom: DenomOne,

			expectedProportionInOrder: []int{0},
		},
		"valid two route single hop": {
			routes: []domain.Route{
				// Route 1
				WithRoutePools(&route.RouteImpl{}, []domain.RoutablePool{
					mocks.WithChainPoolModel(mocks.WithTokenOutDenom(DefaultPool, DenomOne), defaultBalancerPool),
				}),

				// Route 2
				WithRoutePools(&route.RouteImpl{}, []domain.RoutablePool{
					mocks.WithPoolID(mocks.WithChainPoolModel(mocks.WithTokenOutDenom(DefaultPool, DenomOne), secondBalancerPoolSameDenoms), 2),
				}),
			},

			maxSplitIterations: 10,

			tokenIn: sdk.NewCoin(DenomTwo, sdk.NewInt(5_000_000)),

			expectedTokenOutDenom: DenomOne,

			// Route 2 is preferred because it has 2x the liquidity of Route 1
			expectedProportionInOrder: []int{0, 1},
		},
		"valid three route single hop": {
			routes: []domain.Route{
				// Route 1
				WithRoutePools(&route.RouteImpl{}, []domain.RoutablePool{
					mocks.WithChainPoolModel(mocks.WithTokenOutDenom(DefaultPool, DenomOne), defaultBalancerPool),
				}),

				// Route 2
				WithRoutePools(&route.RouteImpl{}, []domain.RoutablePool{
					mocks.WithPoolID(mocks.WithChainPoolModel(mocks.WithTokenOutDenom(DefaultPool, DenomOne), thirdBalancerPoolSameDenoms), 3),
				}),

				// Route 3
				WithRoutePools(&route.RouteImpl{}, []domain.RoutablePool{
					mocks.WithPoolID(mocks.WithChainPoolModel(mocks.WithTokenOutDenom(DefaultPool, DenomOne), secondBalancerPoolSameDenoms), 2),
				}),
			},

			maxSplitIterations: 17,

			tokenIn: sdk.NewCoin(DenomTwo, sdk.NewInt(56_789_321)),

			expectedTokenOutDenom: DenomOne,

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

			logger, err := log.NewLogger(false, "", "")
			s.Require().NoError(err)

			r := routerusecase.NewRouter([]uint64{}, domain.TakerFeeMap{}, 0, 0, tc.maxSplitIterations, 0, logger)

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
		routes                    []domain.Route
		tokenInDenom              string
		expectError               error
		expectFiltered            bool
		expectFilteredRouteLength int
	}{
		"valid single route single hop": {
			routes: []domain.Route{
				WithRoutePools(EmptyRoute, []domain.RoutablePool{
					mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomOne, DenomThree}), DenomThree),
					mocks.WithPoolID(mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomTwo, DenomThree}), DenomTwo), defaultPoolID+1),
				}),
			},

			tokenInDenom: DenomOne,
		},
		"valid single route multi-hop": {
			routes: []domain.Route{
				WithRoutePools(EmptyRoute, []domain.RoutablePool{
					mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomOne, DenomThree}), DenomThree),
					mocks.WithPoolID(mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomTwo, DenomThree}), DenomTwo), defaultPoolID+1),
				}),
			},

			tokenInDenom: DenomOne,
		},
		"valid multi route": {
			routes: []domain.Route{
				WithRoutePools(EmptyRoute, []domain.RoutablePool{
					mocks.WithTokenOutDenom(DefaultPool, DenomTwo),
				}),
				WithRoutePools(EmptyRoute, []domain.RoutablePool{
					mocks.WithPoolID(mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomOne, DenomThree}), DenomThree), defaultPoolID+1),
					mocks.WithPoolID(mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomTwo, DenomThree}), DenomTwo), defaultPoolID+2),
				}),
			},

			tokenInDenom: DenomOne,
		},

		// errors

		"error: no pools in route": {
			routes: []domain.Route{
				WithRoutePools(EmptyRoute, []domain.RoutablePool{}),
			},

			tokenInDenom: DenomTwo,

			expectError: usecase.NoPoolsInRouteError{RouteIndex: 0},
		},
		"error: token out mismatch between multiple routes": {
			routes: []domain.Route{WithRoutePools(EmptyRoute, []domain.RoutablePool{
				mocks.WithTokenOutDenom(DefaultPool, DenomTwo),
			}),
				WithRoutePools(EmptyRoute, []domain.RoutablePool{
					mocks.WithPoolID(mocks.WithTokenOutDenom(DefaultPool, DenomOne), defaultPoolID+1),
				}),
			},

			tokenInDenom: DenomTwo,

			expectError: usecase.TokenOutMismatchBetweenRoutesError{TokenOutDenomRouteA: DenomTwo, TokenOutDenomRouteB: DenomOne},
		},
		"error: token in matches token out": {
			routes: []domain.Route{WithRoutePools(EmptyRoute, []domain.RoutablePool{
				mocks.WithTokenOutDenom(DefaultPool, DenomOne),
			}),
			},
			tokenInDenom: DenomOne,

			expectError: usecase.TokenOutDenomMatchesTokenInDenomError{Denom: DenomOne},
		},
		"error: token in does not match pool denoms": {
			routes: []domain.Route{WithRoutePools(EmptyRoute, []domain.RoutablePool{
				mocks.WithTokenOutDenom(DefaultPool, DenomOne),
			}),
			},
			tokenInDenom: DenomThree,

			expectError: usecase.PreviousTokenOutDenomNotInPoolError{RouteIndex: 0, PoolId: DefaultPool.GetId(), PreviousTokenOutDenom: DenomThree},
		},
		"error: token out does not match pool denoms": {
			routes: []domain.Route{WithRoutePools(EmptyRoute, []domain.RoutablePool{
				mocks.WithTokenOutDenom(DefaultPool, DenomThree),
			}),
			},
			tokenInDenom: DenomOne,

			expectError: usecase.CurrentTokenOutDenomNotInPoolError{RouteIndex: 0, PoolId: DefaultPool.GetId(), CurrentTokenOutDenom: DenomThree},
		},

		// Routes filtered
		"filtered: token in is in the route": {
			routes: []domain.Route{WithRoutePools(EmptyRoute, []domain.RoutablePool{
				mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomOne, DenomThree}), DenomThree),
				mocks.WithPoolID(mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomTwo, DenomThree}), DenomTwo), defaultPoolID+1),
				mocks.WithPoolID(mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomTwo, DenomOne}), DenomOne), defaultPoolID+2),
				mocks.WithPoolID(mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomOne, DenomFour}), DenomFour), defaultPoolID+3),
			}),
			},
			tokenInDenom: DenomOne,

			expectFiltered: true,
		},
		"filtered: token out is in the route": {
			routes: []domain.Route{WithRoutePools(EmptyRoute, []domain.RoutablePool{
				mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomOne, DenomTwo}), DenomTwo),
				mocks.WithPoolID(mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomTwo, DenomFour}), DenomFour), defaultPoolID+1),
				mocks.WithPoolID(mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomFour, DenomThree}), DenomThree), defaultPoolID+2),
				mocks.WithPoolID(mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomThree, DenomFour}), DenomFour), defaultPoolID+3),
			}),
			},
			tokenInDenom: DenomOne,

			expectFiltered: true,
		},
		"filtered: same pool id within only route": {
			routes: []domain.Route{WithRoutePools(EmptyRoute, []domain.RoutablePool{
				mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomOne, DenomTwo}), DenomTwo),
				mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomTwo, DenomFour}), DenomFour),
			}),
			},
			tokenInDenom: DenomOne,

			expectFiltered: true,
		},
		"filtered: same pool id between routes - second removed": {
			routes: []domain.Route{
				WithRoutePools(EmptyRoute, []domain.RoutablePool{
					mocks.WithTokenOutDenom(DefaultPool, DenomTwo), // ID 1
				}),
				WithRoutePools(EmptyRoute, []domain.RoutablePool{
					mocks.WithPoolID(mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomOne, DenomThree}), DenomThree), defaultPoolID+1),
					mocks.WithTokenOutDenom(mocks.WithDenoms(DefaultPool, []string{DenomTwo, DenomThree}), DenomTwo), // ID 1
				}),
			},
			tokenInDenom: DenomOne,

			expectFiltered:            true,
			expectFilteredRouteLength: 1,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {

			router := routerusecase.NewRouter([]uint64{}, domain.TakerFeeMap{}, 0, 0, 0, 0, &log.NoOpLogger{})

			filteredRoutes, err := router.ValidateAndFilterRoutes(tc.routes, tc.tokenInDenom)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}
			s.Require().NoError(err)

			if tc.expectFiltered {
				s.Require().NotEqual(len(tc.routes), len(filteredRoutes))
				s.Require().Len(filteredRoutes, tc.expectFilteredRouteLength)
				return
			}

			s.Require().Equal(len(tc.routes), len(filteredRoutes))
		})
	}
}

// Basic test validating that a non-zero quote is returned from mainnet state
func (s *RouterTestSuite) TestGetOptimalQuote_Mainnet() {
	router := s.setupMainnetRouter()

	const (
		UOSMO = "uosmo"
		ATOM  = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	)

	candidateRoutes, err := router.GetCandidateRoutes(UOSMO, ATOM)
	s.Require().NoError(err)

	// ATOM / OSMO
	quote, err := router.GetOptimalQuote(sdk.NewCoin(UOSMO, sdk.NewInt(100_000_000)), ATOM, candidateRoutes)
	s.Require().NoError(err)

	// at least 6 ATOM
	s.Require().True(quote.GetAmountOut().GT(osmomath.NewInt(6_000_000)))
}
