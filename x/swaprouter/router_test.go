package swaprouter_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	swaproutertypes "github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
)

const (
	denomA = "denomA"
	denomB = "denomB"
	denomC = "denomC"
)

var (
	defaultInitPoolAmount = sdk.NewInt(1000000000000)
	defaultSwapAmount     = sdk.NewInt(1000000)
)

// TestMultihopSwapExactAmountIn tests that the swaps are routed correctly.
// That is:
// - to the correct module (concentrated-liquidity or gamm)
// - over the right routes (hops)
// This test does not actually validate the math behind the swaps.
func (suite *KeeperTestSuite) TestMultihopSwapExactAmountIn() {
	type param struct {
	}

	tests := []struct {
		name              string
		poolCoins         []sdk.Coins
		routes            []swaproutertypes.SwapAmountInRoute
		tokenIn           sdk.Coin
		tokenOutMinAmount sdk.Int
		swapFee           sdk.Dec
		expectError       bool
		reducedFeeApplied bool
	}{
		{
			name:      "x/gamm - single route",
			poolCoins: []sdk.Coins{sdk.NewCoins(sdk.NewCoin(denomA, defaultInitPoolAmount), sdk.NewCoin(denomB, defaultInitPoolAmount))},
			routes: []swaproutertypes.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: denomB,
				},
			},
			tokenIn:           sdk.NewCoin(denomA, sdk.NewInt(100000)),
			tokenOutMinAmount: sdk.NewInt(1),
		},
		{
			name: "x/gamm - two routes",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(denomA, defaultInitPoolAmount), sdk.NewCoin(denomB, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(denomB, defaultInitPoolAmount), sdk.NewCoin(denomC, defaultInitPoolAmount)), // pool 2.
			},
			routes: []swaproutertypes.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: denomB,
				},
				{
					PoolId:        2,
					TokenOutDenom: denomC,
				},
			},
			tokenIn:           sdk.NewCoin(denomA, sdk.NewInt(100000)),
			tokenOutMinAmount: sdk.NewInt(1),
		},
		// TODO:
		// tests for concentrated liquidity
		// test for multi-hop
		// test for multi hop (osmo routes) -> reduces fee, add an assertion
		// change values in and out to be different with each swap module type
		// create more precise assertions on what the results are
		// edge cases:
		//   * invalid route length
		//   * pool does not exist
		//   * swap errors
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			swaprouterKeeper := suite.App.SwapRouterKeeper

			suite.CreateBalancerPoolsFromCoins(tc.poolCoins)

			tokenOut, err := swaprouterKeeper.RouteExactAmountIn(suite.Ctx, suite.TestAccs[0], tc.routes, tc.tokenIn, tc.tokenOutMinAmount)

			if tc.expectError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().True(tokenOut.GTE(tc.tokenOutMinAmount))
			}
		})
	}
}

// TestMultihopSwapExactAmountOut tests that the swaps are routed correctly.
// That is:
// - to the correct module (concentrated-liquidity or gamm)
// - over the right routes (hops)
// This test does not actually validate the math behind the swaps.
func (suite *KeeperTestSuite) TestMultihopSwapExactAmountOut() {

	tests := []struct {
		name              string
		poolCoins         []sdk.Coins
		routes            []swaproutertypes.SwapAmountOutRoute
		tokenOut          sdk.Coin
		tokenInMaxAmount  sdk.Int
		swapFee           sdk.Dec
		expectError       bool
		reducedFeeApplied bool
	}{
		{
			name:      "x/gamm - single route",
			poolCoins: []sdk.Coins{sdk.NewCoins(sdk.NewCoin(denomA, defaultInitPoolAmount), sdk.NewCoin(denomB, defaultInitPoolAmount))},
			routes: []swaproutertypes.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: denomB,
				},
			},
			tokenOut:         sdk.NewCoin(denomA, defaultSwapAmount),
			tokenInMaxAmount: defaultSwapAmount.Mul(sdk.NewInt(2)), // the amount here is arbitrary
		},
		// TODO:
		// tests for concentrated liquidity
		// test for multi-hop
		// test for multi hop (osmo routes) -> reduces fee, add an assertion
		// change values in and out to be different with each swap module type
		// create more precise assertions on what the results are
		// edge cases:
		//   * invalid route length
		//   * pool does not exist
		//   * swap errors
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			swaprouterKeeper := suite.App.SwapRouterKeeper

			suite.CreateBalancerPoolsFromCoins(tc.poolCoins)

			tokenIn, err := swaprouterKeeper.RouteExactAmountOut(suite.Ctx, suite.TestAccs[0], tc.routes, tc.tokenInMaxAmount, tc.tokenOut)

			if tc.expectError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().True(tokenIn.LTE(tc.tokenInMaxAmount))
			}
		})
	}
}
