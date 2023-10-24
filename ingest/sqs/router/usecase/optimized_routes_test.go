package usecase_test

import (
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

func (s *RouterTestSuite) TestGetBestSplitRoutesQuote() {

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
	}{
		// "valid single route": {
		// 	routes: []domain.Route{
		// 		withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
		// 			withChainPoolModel(withTokenOutDenom(defaultPool, denomNum(1)), defaultBalancerPool),
		// 		})},
		// 	tokenIn: sdk.NewCoin(denomNum(2), sdk.NewInt(100)),

		// 	expectedTokenOutDenom: denomNum(1),
		// },
		// "valid two route single hop": {
		// 	routes: []domain.Route{
		// 		// Route 1
		// 		withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
		// 			withChainPoolModel(withTokenOutDenom(defaultPool, denomNum(1)), defaultBalancerPool),
		// 		}),

		// 		// Route 2
		// 		withRoutePools(&routerusecase.RouteImpl{}, []domain.RoutablePool{
		// 			withPoolID(withChainPoolModel(withTokenOutDenom(defaultPool, denomNum(1)), secondBalancerPoolSameDenoms), 2),
		// 		}),
		// 	},

		// 	maxSplitIterations: 10,

		// 	tokenIn: sdk.NewCoin(denomNum(2), sdk.NewInt(5_000_000)),

		// 	expectedTokenOutDenom: denomNum(1),
		// },
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
		})
	}
}

// This test ensures strict route validation.
// See individual test cases for details.
func (s *RouterTestSuite) TestValidateRoutes() {
	tests := map[string]struct {
		routes       []domain.Route
		tokenInDenom string
		expectError  error
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
		"error: token in is in the route": {
			routes: []domain.Route{withRoutePools(emptyRoute, []domain.RoutablePool{
				withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(1), denomNum(3)}), denomNum(3)),
				withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(2), denomNum(3)}), denomNum(2)),
				withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(2), denomNum(1)}), denomNum(1)),
				withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(1), denomNum(4)}), denomNum(4)),
			}),
			},
			tokenInDenom: denomNum(1),

			expectError: usecase.RoutePoolWithTokenInDenomError{RouteIndex: 0, TokenInDenom: denomNum(1)},
		},
		"error: token out is in the route": {
			routes: []domain.Route{withRoutePools(emptyRoute, []domain.RoutablePool{
				withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(1), denomNum(2)}), denomNum(2)),
				withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(2), denomNum(4)}), denomNum(4)),
				withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(4), denomNum(3)}), denomNum(3)),
				withTokenOutDenom(withDenoms(defaultPool, []string{denomNum(3), denomNum(4)}), denomNum(4)),
			}),
			},
			tokenInDenom: denomNum(1),

			expectError: usecase.RoutePoolWithTokenOutDenomError{RouteIndex: 0, TokenOutDenom: denomNum(4)},
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
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {

			err := usecase.ValidateRoutes(tc.routes, tc.tokenInDenom)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}
			s.Require().NoError(err)
		})
	}
}
