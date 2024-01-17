package usecase_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/pools"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/route"
	"github.com/osmosis-labs/osmosis/v21/x/gamm/pool-models/balancer"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
)

// TestPrepareResult prepares the result of the quote for output to the client.
// First, it strips away unnecessary fields from each pool in the route.
// Additionally, it computes the effective spread factor from all routes.
//
// The test structure is as follows:
// - Set up a 50-50 split route
// Route 1: 2 hop
// Route 2: 1 hop
//
// Validate that the effective swap fee is computed correctly.
// TODO: validate that taker fees are accounted for.
func (s *RouterTestSuite) TestPrepareResult() {
	s.SetupTest()

	var (
		takerFeeOne   = osmomath.NewDecWithPrec(2, 2)
		takerFeeTwo   = osmomath.NewDecWithPrec(4, 4)
		takerFeeThree = osmomath.NewDecWithPrec(3, 3)

		defaultAmount = sdk.NewInt(100_000_00)

		totalInAmount  = defaultAmount
		totalOutAmount = defaultAmount.MulRaw(4)

		poolOneBalances = sdk.NewCoins(
			sdk.NewCoin(USDT, defaultAmount.MulRaw(5)),
			sdk.NewCoin(ETH, defaultAmount),
		)

		poolTwoBalances = sdk.NewCoins(
			sdk.NewCoin(USDC, defaultAmount),
			sdk.NewCoin(USDT, defaultAmount),
		)

		poolThreeBalances = sdk.NewCoins(
			sdk.NewCoin(ETH, defaultAmount),
			sdk.NewCoin(USDC, defaultAmount.MulRaw(4)),
		)
	)

	// Prepare 2 routes
	// Route 1: 2 hops
	// Route 2: 1 hop

	// Pool USDT / ETH -> 0.01 spread factor & 5 USDTfor 1 ETH
	poolIDOne := s.PrepareCustomBalancerPool([]balancer.PoolAsset{
		{
			Token:  sdk.NewCoin(USDT, defaultAmount.MulRaw(5)),
			Weight: sdk.NewInt(100),
		},
		{
			Token:  sdk.NewCoin(ETH, defaultAmount),
			Weight: sdk.NewInt(100),
		},
	}, balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: osmomath.ZeroDec(),
	})

	poolOne, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolIDOne)
	s.Require().NoError(err)

	// Pool USDC / USDT -> 0.01 spread factor & 1 USDC for 1 USDT
	poolIDTwo := s.PrepareCustomBalancerPool([]balancer.PoolAsset{
		{
			Token:  sdk.NewCoin(USDC, defaultAmount),
			Weight: sdk.NewInt(100),
		},
		{
			Token:  sdk.NewCoin(USDT, defaultAmount),
			Weight: sdk.NewInt(100),
		},
	}, balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(3, 2),
		ExitFee: osmomath.ZeroDec(),
	})

	poolTwo, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolIDTwo)
	s.Require().NoError(err)

	// Pool ETH / USDC -> 0.005 spread factor & 4 USDC for 1 ETH
	poolIDThree := s.PrepareCustomBalancerPool([]balancer.PoolAsset{
		{
			Token:  sdk.NewCoin(ETH, defaultAmount),
			Weight: sdk.NewInt(100),
		},
		{
			Token:  sdk.NewCoin(USDC, defaultAmount.MulRaw(4)),
			Weight: sdk.NewInt(100),
		},
	}, balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(5, 3),
		ExitFee: osmomath.ZeroDec(),
	})

	poolThree, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolIDThree)
	s.Require().NoError(err)

	testQuote := &usecase.QuoteImpl{
		AmountIn:  sdk.NewCoin(ETH, totalInAmount),
		AmountOut: totalOutAmount,

		// 2 routes with 50-50 split, each single hop
		Route: []domain.SplitRoute{

			// Route 1
			&usecase.RouteWithOutAmount{
				RouteImpl: route.RouteImpl{
					Pools: []domain.RoutablePool{
						pools.NewRoutablePool(
							domain.NewPool(poolOne, poolOne.GetSpreadFactor(sdk.Context{}), poolOneBalances),
							USDT,
							takerFeeOne,
						),
						pools.NewRoutablePool(
							domain.NewPool(poolTwo, poolTwo.GetSpreadFactor(sdk.Context{}), poolTwoBalances),
							USDC,
							takerFeeTwo,
						),
					},
				},

				InAmount:  totalInAmount.QuoRaw(2),
				OutAmount: totalOutAmount.QuoRaw(2),
			},

			// Route 2
			&usecase.RouteWithOutAmount{
				RouteImpl: route.RouteImpl{
					Pools: []domain.RoutablePool{
						pools.NewRoutablePool(
							domain.NewPool(poolThree, poolThree.GetSpreadFactor(sdk.Context{}), poolThreeBalances),
							USDC,
							takerFeeThree,
						),
					},
				},

				InAmount:  totalInAmount.QuoRaw(2),
				OutAmount: totalOutAmount.QuoRaw(2),
			},
		},
		EffectiveFee: osmomath.ZeroDec(),
	}

	expectedRoutes := []domain.SplitRoute{

		// Route 1
		&usecase.RouteWithOutAmount{
			RouteImpl: route.RouteImpl{
				Pools: []domain.RoutablePool{
					pools.NewRoutableResultPool(
						poolIDOne,
						poolmanagertypes.Balancer,
						poolOne.GetSpreadFactor(sdk.Context{}),
						USDT,
						takerFeeOne,
					),
					pools.NewRoutableResultPool(
						poolIDTwo,
						poolmanagertypes.Balancer,
						poolTwo.GetSpreadFactor(sdk.Context{}),
						USDC,
						takerFeeTwo,
					),
				},
			},

			InAmount:  totalInAmount.QuoRaw(2),
			OutAmount: totalOutAmount.QuoRaw(2),
		},

		// Route 2
		&usecase.RouteWithOutAmount{
			RouteImpl: route.RouteImpl{
				Pools: []domain.RoutablePool{
					pools.NewRoutableResultPool(
						poolIDThree,
						poolmanagertypes.Balancer,
						poolThree.GetSpreadFactor(sdk.Context{}),
						USDC,
						takerFeeThree,
					),
				},
			},

			InAmount:  totalInAmount.QuoRaw(2),
			OutAmount: totalOutAmount.QuoRaw(2),
		},
	}

	// Compute expected total fee and validate against actual
	expectedPoolOneTotalFee := poolOne.GetSpreadFactor(sdk.Context{}).Add(takerFeeOne)
	expectedPoolTwoTotalFee := poolTwo.GetSpreadFactor(sdk.Context{}).Add(takerFeeTwo)
	expectedPoolThreeTotalFee := poolThree.GetSpreadFactor(sdk.Context{}).Add(takerFeeThree)

	expectedRouteOneFee := expectedPoolOneTotalFee.Add(osmomath.OneDec().Sub(expectedPoolOneTotalFee).MulMut(expectedPoolTwoTotalFee)).MulMut(osmomath.NewDecWithPrec(5, 1))
	expectedRouteTwoFee := expectedPoolThreeTotalFee.MulMut(osmomath.NewDecWithPrec(5, 1))

	// ((0.01 + 0.02) + (1 - (0.01 + 0.02)) * (0.03 + 0.0004)) * 0.5 + (0.005 + 0.003) * 0.5
	expectedEffectiveSpreadFactor := expectedRouteOneFee.Add(expectedRouteTwoFee)

	// System under test
	routes, effectiveSpreadFactor := testQuote.PrepareResult()

	// Validate routes.
	s.validateRoutes(expectedRoutes, routes)
	s.validateRoutes(expectedRoutes, testQuote.GetRoute())

	// Validate effective spread factor.
	s.Require().Equal(expectedEffectiveSpreadFactor.String(), effectiveSpreadFactor.String())
	s.Require().Equal(expectedEffectiveSpreadFactor.String(), testQuote.GetEffectiveSpreadFactor().String())
}

// validateRoutes validates that the given routes are equal.
// Specifically, validates:
// - Pools
// - In amount
// - Out amount
func (s *RouterTestSuite) validateRoutes(expectedRoutes []domain.SplitRoute, actualRoutes []domain.SplitRoute) {
	s.Require().Equal(len(expectedRoutes), len(actualRoutes))
	for i, expectedRoute := range expectedRoutes {
		actualRoute := actualRoutes[i]

		// Validate pools
		s.ValidateRoutePools(expectedRoute.GetPools(), actualRoute.GetPools())

		// Validate in amount
		s.Require().Equal(expectedRoute.GetAmountIn().String(), actualRoute.GetAmountIn().String())

		// Validate out amount
		s.Require().Equal(expectedRoute.GetAmountOut().String(), actualRoute.GetAmountOut().String())
	}
}
