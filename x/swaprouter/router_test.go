package swaprouter_test

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gamm "github.com/osmosis-labs/osmosis/v12/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
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
	gammKeeperType        = reflect.TypeOf(&gamm.Keeper{})
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

			suite.createBalancerPoolsFromCoins(tc.poolCoins)

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

			suite.createBalancerPoolsFromCoins(tc.poolCoins)

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

// TestGetSwapModule tests that the correct swap module is returned for a given pool id.
// Additionally, validates that the expected errors are produced when expected.
func (suite *KeeperTestSuite) TestGetSwapModule() {

	tests := map[string]struct {
		poolId            uint64
		preCreatePoolType types.PoolType
		routesOverwrite   map[types.PoolType]types.SwapI

		expectedModule reflect.Type
		expectError    error
	}{
		"valid balancer pool": {
			preCreatePoolType: types.Balancer,
			poolId:            1,
			expectedModule:    gammKeeperType,
		},
		"non-existent pool": {
			preCreatePoolType: types.Balancer,
			poolId:            2,
			expectedModule:    gammKeeperType,

			expectError: types.FailedToFindRouteError{PoolId: 2},
		},
		"undefined route": {
			preCreatePoolType: types.Balancer,
			poolId:            1,
			routesOverwrite: map[types.PoolType]types.SwapI{
				types.StableSwap: &gamm.Keeper{}, // undefined for balancer.
			},

			expectError: types.UndefinedRouteError{PoolId: 1, PoolType: types.Balancer},
		},
		// TODO: valid stableswap test case.
		// TODO: valid concentrated liquidity test case.
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()
			swaprouterKeeper := suite.App.SwapRouterKeeper

			suite.createPoolFromType(tc.preCreatePoolType)

			if len(tc.routesOverwrite) > 0 {
				swaprouterKeeper.SetPoolRoutesUnsafe(tc.routesOverwrite)
			}

			swapModule, err := swaprouterKeeper.GetSwapModule(suite.Ctx, tc.poolId)

			if tc.expectError != nil {
				suite.Require().Error(err)
				suite.Require().Nil(swapModule)
				return
			}

			suite.Require().NoError(err)
			suite.Require().NotNil(swapModule)

			suite.Require().Equal(tc.expectedModule, reflect.TypeOf(swapModule))
		})
	}
}
