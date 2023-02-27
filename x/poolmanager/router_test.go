package poolmanager_test

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	gamm "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v15/x/pool-incentives/types"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

const (
	foo   = "foo"
	bar   = "bar"
	baz   = "baz"
	uosmo = "uosmo"
)

var (
	defaultInitPoolAmount     = sdk.NewInt(1000000000000)
	DefaultExponentAtPriceOne = sdk.NewInt(-4)
	defaultPoolSwapFee        = sdk.NewDecWithPrec(1, 2) // 1% pool swap fee default
	defaultSwapAmount         = sdk.NewInt(1000000)
	gammKeeperType            = reflect.TypeOf(&gamm.Keeper{})
	concentratedKeeperType    = reflect.TypeOf(&cl.Keeper{})
)

// TestGetPoolModule tests that the correct pool module is returned for a given pool id.
// Additionally, validates that the expected errors are produced when expected.
func (suite *KeeperTestSuite) TestGetPoolModule() {
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
		"valid stableswap pool": {
			preCreatePoolType: types.Stableswap,
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
				types.Stableswap: &gamm.Keeper{}, // undefined for balancer.
			},

			expectError: types.UndefinedRouteError{PoolId: 1, PoolType: types.Balancer},
		},
		// TODO: valid concentrated liquidity test case.
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()
			poolmanagerKeeper := suite.App.PoolManagerKeeper

			suite.createPoolFromType(tc.preCreatePoolType)

			if len(tc.routesOverwrite) > 0 {
				poolmanagerKeeper.SetPoolRoutesUnsafe(tc.routesOverwrite)
			}

			swapModule, err := poolmanagerKeeper.GetPoolModule(suite.Ctx, tc.poolId)

			if tc.expectError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expectError)
				suite.Require().Nil(swapModule)
				return
			}

			suite.Require().NoError(err)
			suite.Require().NotNil(swapModule)

			suite.Require().Equal(tc.expectedModule, reflect.TypeOf(swapModule))
		})
	}
}

// TestMultihopSwapExactAmountIn tests that the swaps are routed correctly.
// That is:
// - to the correct module (concentrated-liquidity or gamm)
// - over the right routes (hops)
// - fee reduction is applied correctly
func (suite *KeeperTestSuite) TestMultihopSwapExactAmountIn() {
	tests := []struct {
		name                    string
		poolCoins               []sdk.Coins
		poolFee                 []sdk.Dec
		routes                  []poolmanagertypes.SwapAmountInRoute
		incentivizedGauges      []uint64
		tokenIn                 sdk.Coin
		tokenOutMinAmount       sdk.Int
		swapFee                 sdk.Dec
		expectError             bool
		expectReducedFeeApplied bool
	}{
		{
			name:      "One route: Swap - [foo -> bar], 1 percent fee",
			poolCoins: []sdk.Coins{sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount))},
			poolFee:   []sdk.Dec{defaultPoolSwapFee},
			routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: bar,
				},
			},
			tokenIn:           sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount: sdk.NewInt(1),
		},
		{
			name: "Two routes: Swap - [foo -> bar](pool 1) - [bar -> baz](pool 2), both pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount)), // pool 2.
			},
			poolFee: []sdk.Dec{defaultPoolSwapFee, defaultPoolSwapFee},
			routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: bar,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
			},
			incentivizedGauges: []uint64{},
			tokenIn:            sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount:  sdk.NewInt(1),
		},
		{
			name: "Two routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) with a half fee applied, both pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(baz, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.
			},
			poolFee: []sdk.Dec{defaultPoolSwapFee, defaultPoolSwapFee},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenIn:                 sdk.NewCoin("foo", sdk.NewInt(100000)),
			tokenOutMinAmount:       sdk.NewInt(1),
			expectReducedFeeApplied: true,
		},
		{
			name: "Two routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) with a half fee applied, (pool 1) 1 percent fee, (pool 2) 10 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(baz, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.
			},
			poolFee: []sdk.Dec{defaultPoolSwapFee, sdk.NewDecWithPrec(1, 1)},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenIn:                 sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount:       sdk.NewInt(1),
			expectReducedFeeApplied: true,
		},
		{
			name: "Three routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(baz, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount)),   // pool 3.
			},
			poolFee: []sdk.Dec{defaultPoolSwapFee, defaultPoolSwapFee, defaultPoolSwapFee},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
				{
					PoolId:        3,
					TokenOutDenom: bar,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenIn:                 sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount:       sdk.NewInt(1),
			expectReducedFeeApplied: false,
		},
		{
			name: "Two routes: Swap between four asset pools - [foo -> bar](pool 1) - [bar -> baz](pool 2), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.                                                                                     // pool 3.
			},
			poolFee: []sdk.Dec{defaultPoolSwapFee, defaultPoolSwapFee},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: bar,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenIn:                 sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount:       sdk.NewInt(1),
			expectReducedFeeApplied: false,
		},
		{
			name: "Two routes: Swap between four asset pools - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2), with a half fee applied, both pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.                                                                                     // pool 3.
			},
			poolFee: []sdk.Dec{defaultPoolSwapFee, defaultPoolSwapFee},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenIn:                 sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount:       sdk.NewInt(1),
			expectReducedFeeApplied: true,
		},
		{
			name: "Three routes: Swap between four asset pools - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 3.                                                                                      // pool 3.
			},
			poolFee: []sdk.Dec{defaultPoolSwapFee, defaultPoolSwapFee, defaultPoolSwapFee},
			routes: []types.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: uosmo,
				},
				{
					PoolId:        2,
					TokenOutDenom: baz,
				},
				{
					PoolId:        3,
					TokenOutDenom: bar,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9},
			tokenIn:                 sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount:       sdk.NewInt(1),
			expectReducedFeeApplied: false,
		},
		// TODO:
		// tests for concentrated liquidity
		// change values in and out to be different with each swap module type
		// tests for stable-swap pools
		// edge cases:
		//   * invalid route length
		//   * pool does not exist
		//   * swap errors
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			poolmanagerKeeper := suite.App.PoolManagerKeeper

			suite.createBalancerPoolsFromCoinsWithSwapFee(tc.poolCoins, tc.poolFee)

			// if test specifies incentivized gauges, set them here
			if len(tc.incentivizedGauges) > 0 {
				suite.makeGaugesIncentivized(tc.incentivizedGauges)
			}

			if tc.expectError {
				// execute the swap
				_, err := poolmanagerKeeper.RouteExactAmountIn(suite.Ctx, suite.TestAccs[0], tc.routes, tc.tokenIn, tc.tokenOutMinAmount)
				suite.Require().Error(err)
			} else {
				// calculate the swap as separate swaps with either the reduced swap fee or normal fee
				expectedMultihopTokenOutAmount := suite.calcInAmountAsSeparateSwaps(tc.expectReducedFeeApplied, tc.routes, tc.tokenIn)
				// execute the swap
				multihopTokenOutAmount, err := poolmanagerKeeper.RouteExactAmountIn(suite.Ctx, suite.TestAccs[0], tc.routes, tc.tokenIn, tc.tokenOutMinAmount)
				// compare the expected tokenOut to the actual tokenOut
				suite.Require().NoError(err)
				suite.Require().Equal(expectedMultihopTokenOutAmount.Amount.String(), multihopTokenOutAmount.String())
			}
		})
	}
}

// TestMultihopSwapExactAmountOut tests that the swaps are routed correctly.
// That is:
// - to the correct module (concentrated-liquidity or gamm)
// - over the right routes (hops)
// - fee reduction is applied correctly
func (suite *KeeperTestSuite) TestMultihopSwapExactAmountOut() {
	tests := []struct {
		name                    string
		poolCoins               []sdk.Coins
		poolFee                 []sdk.Dec
		routes                  []poolmanagertypes.SwapAmountOutRoute
		incentivizedGauges      []uint64
		tokenOut                sdk.Coin
		tokenInMaxAmount        sdk.Int
		swapFee                 sdk.Dec
		expectError             bool
		expectReducedFeeApplied bool
	}{
		{
			name:      "One route: Swap - [foo -> bar], 1 percent fee",
			poolCoins: []sdk.Coins{sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount))},
			poolFee:   []sdk.Dec{defaultPoolSwapFee},
			routes: []poolmanagertypes.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: bar,
				},
			},
			tokenInMaxAmount: sdk.NewInt(90000000),
			tokenOut:         sdk.NewCoin(foo, defaultSwapAmount),
		},
		{
			name: "Two routes: Swap - [foo -> bar](pool 1) - [bar -> baz](pool 2), both pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount)), // pool 2.
			},
			poolFee: []sdk.Dec{defaultPoolSwapFee, defaultPoolSwapFee},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
				{
					PoolId:       2,
					TokenInDenom: bar,
				},
			},
			incentivizedGauges: []uint64{},

			tokenInMaxAmount: sdk.NewInt(90000000),
			tokenOut:         sdk.NewCoin(baz, sdk.NewInt(100000)),
		},
		{
			name: "Two routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) with a half fee applied, both pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(baz, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.
			},
			poolFee: []sdk.Dec{defaultPoolSwapFee, defaultPoolSwapFee},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
				{
					PoolId:       2,
					TokenInDenom: uosmo,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenInMaxAmount:        sdk.NewInt(90000000),
			tokenOut:                sdk.NewCoin(baz, sdk.NewInt(100000)),
			expectReducedFeeApplied: true,
		},
		{
			name: "Two routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) with a half fee applied, (pool 1) 1 percent fee, (pool 2) 10 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(baz, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.
			},
			poolFee: []sdk.Dec{defaultPoolSwapFee, sdk.NewDecWithPrec(1, 1)},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
				{
					PoolId:       2,
					TokenInDenom: uosmo,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenInMaxAmount:        sdk.NewInt(90000000),
			tokenOut:                sdk.NewCoin(baz, sdk.NewInt(100000)),
			expectReducedFeeApplied: true,
		},
		{
			name: "Three routes: Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(baz, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount)),   // pool 3.
			},
			poolFee: []sdk.Dec{defaultPoolSwapFee, defaultPoolSwapFee, defaultPoolSwapFee},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
				{
					PoolId:       2,
					TokenInDenom: uosmo,
				},
				{
					PoolId:       3,
					TokenInDenom: baz,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenInMaxAmount:        sdk.NewInt(90000000),
			tokenOut:                sdk.NewCoin(bar, sdk.NewInt(100000)),
			expectReducedFeeApplied: false,
		},
		{
			name: "Two routes: Swap between four asset pools - [foo -> bar](pool 1) - [bar -> baz](pool 2), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.                                                                                     // pool 3.
			},
			poolFee: []sdk.Dec{defaultPoolSwapFee, defaultPoolSwapFee},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
				{
					PoolId:       2,
					TokenInDenom: bar,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenOut:                sdk.NewCoin(baz, sdk.NewInt(100000)),
			tokenInMaxAmount:        sdk.NewInt(90000000),
			expectReducedFeeApplied: false,
		},
		{
			name: "Two routes: Swap between four asset pools - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2), with a half fee applied, both pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.                                                                                     // pool 3.
			},
			poolFee: []sdk.Dec{defaultPoolSwapFee, defaultPoolSwapFee},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
				{
					PoolId:       2,
					TokenInDenom: uosmo,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6},
			tokenOut:                sdk.NewCoin(baz, sdk.NewInt(100000)),
			tokenInMaxAmount:        sdk.NewInt(90000000),
			expectReducedFeeApplied: true,
		},
		{
			name: "Three routes: Swap between four asset pools - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
			poolCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 1.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 2.
				sdk.NewCoins(sdk.NewCoin(bar, defaultInitPoolAmount), sdk.NewCoin(baz, defaultInitPoolAmount),
					sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(uosmo, defaultInitPoolAmount)), // pool 3.                                                                                    // pool 3.
			},
			poolFee: []sdk.Dec{defaultPoolSwapFee, defaultPoolSwapFee, defaultPoolSwapFee},
			routes: []types.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: foo,
				},
				{
					PoolId:       2,
					TokenInDenom: uosmo,
				},
				{
					PoolId:       3,
					TokenInDenom: baz,
				},
			},
			incentivizedGauges:      []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9},
			tokenOut:                sdk.NewCoin(bar, sdk.NewInt(100000)),
			tokenInMaxAmount:        sdk.NewInt(90000000),
			expectReducedFeeApplied: false,
		},
		// TODO:
		// tests for concentrated liquidity
		// tests for stable-swap pools
		// change values in and out to be different with each swap module type
		// edge cases:
		//   * invalid route length
		//   * pool does not exist
		//   * swap errors
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			poolmanagerKeeper := suite.App.PoolManagerKeeper

			suite.createBalancerPoolsFromCoinsWithSwapFee(tc.poolCoins, tc.poolFee)

			// if test specifies incentivized gauges, set them here
			if len(tc.incentivizedGauges) > 0 {
				suite.makeGaugesIncentivized(tc.incentivizedGauges)
			}

			if tc.expectError {
				// execute the swap
				_, err := poolmanagerKeeper.RouteExactAmountOut(suite.Ctx, suite.TestAccs[0], tc.routes, tc.tokenInMaxAmount, tc.tokenOut)
				suite.Require().Error(err)
			} else {
				// calculate the swap as separate swaps with either the reduced swap fee or normal fee
				expectedMultihopTokenOutAmount := suite.calcOutAmountAsSeparateSwaps(tc.expectReducedFeeApplied, tc.routes, tc.tokenOut)
				// execute the swap
				multihopTokenOutAmount, err := poolmanagerKeeper.RouteExactAmountOut(suite.Ctx, suite.TestAccs[0], tc.routes, tc.tokenInMaxAmount, tc.tokenOut)
				// compare the expected tokenOut to the actual tokenOut
				suite.Require().NoError(err)
				suite.Require().Equal(expectedMultihopTokenOutAmount.Amount.String(), multihopTokenOutAmount.String())
			}
		})
	}
}

// TestEstimateMultihopSwapExactAmountIn tests that the estimation done via `EstimateSwapExactAmountIn`
// results in the same amount of token out as the actual swap.
func (suite *KeeperTestSuite) TestEstimateMultihopSwapExactAmountIn() {
	type param struct {
		routes            []types.SwapAmountInRoute
		estimateRoutes    []types.SwapAmountInRoute
		tokenIn           sdk.Coin
		tokenOutMinAmount sdk.Int
	}

	tests := []struct {
		name              string
		param             param
		expectPass        bool
		reducedFeeApplied bool
		poolType          types.PoolType
	}{
		{
			name: "Proper swap - foo -> bar(pool 1) - bar(pool 2) -> baz",
			param: param{
				routes: []types.SwapAmountInRoute{
					{
						PoolId:        1,
						TokenOutDenom: bar,
					},
					{
						PoolId:        2,
						TokenOutDenom: baz,
					},
				},
				estimateRoutes: []types.SwapAmountInRoute{
					{
						PoolId:        3,
						TokenOutDenom: bar,
					},
					{
						PoolId:        4,
						TokenOutDenom: baz,
					},
				},
				tokenIn:           sdk.NewCoin(foo, sdk.NewInt(100000)),
				tokenOutMinAmount: sdk.NewInt(1),
			},
			expectPass: true,
		},
		{
			name: "Swap - foo -> uosmo(pool 1) - uosmo(pool 2) -> baz with a half fee applied",
			param: param{
				routes: []types.SwapAmountInRoute{
					{
						PoolId:        1,
						TokenOutDenom: uosmo,
					},
					{
						PoolId:        2,
						TokenOutDenom: baz,
					},
				},
				estimateRoutes: []types.SwapAmountInRoute{
					{
						PoolId:        3,
						TokenOutDenom: uosmo,
					},
					{
						PoolId:        4,
						TokenOutDenom: baz,
					},
				},
				tokenIn:           sdk.NewCoin(foo, sdk.NewInt(100000)),
				tokenOutMinAmount: sdk.NewInt(1),
			},
			reducedFeeApplied: true,
			expectPass:        true,
		},
		{
			name: "Proper swap (stableswap pool) - foo -> bar(pool 1) - bar(pool 2) -> baz",
			param: param{
				routes: []types.SwapAmountInRoute{
					{
						PoolId:        1,
						TokenOutDenom: bar,
					},
					{
						PoolId:        2,
						TokenOutDenom: baz,
					},
				},
				estimateRoutes: []types.SwapAmountInRoute{
					{
						PoolId:        3,
						TokenOutDenom: bar,
					},
					{
						PoolId:        4,
						TokenOutDenom: baz,
					},
				},
				tokenIn:           sdk.NewCoin(foo, sdk.NewInt(100000)),
				tokenOutMinAmount: sdk.NewInt(1),
			},
			expectPass: true,
			poolType:   types.Stableswap,
		},
		{
			name: "Asserts panic catching in MultihopEstimateOutGivenExactAmountIn works: tokenOut more than pool reserves",
			param: param{
				routes: []types.SwapAmountInRoute{
					{
						PoolId:        1,
						TokenOutDenom: bar,
					},
					{
						PoolId:        2,
						TokenOutDenom: baz,
					},
				},
				estimateRoutes: []types.SwapAmountInRoute{
					{
						PoolId:        3,
						TokenOutDenom: bar,
					},
					{
						PoolId:        4,
						TokenOutDenom: baz,
					},
				},
				tokenIn:           sdk.NewCoin(foo, sdk.NewInt(9000000000000000000)),
				tokenOutMinAmount: sdk.NewInt(1),
			},
			expectPass: false,
			poolType:   types.Stableswap,
		},
	}

	for _, test := range tests {
		// Init suite for each test.
		suite.SetupTest()

		suite.Run(test.name, func() {
			poolmanagerKeeper := suite.App.PoolManagerKeeper

			firstEstimatePoolId, secondEstimatePoolId := suite.setupPools(test.poolType, defaultPoolSwapFee)

			firstEstimatePool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, firstEstimatePoolId)
			suite.Require().NoError(err)
			secondEstimatePool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, secondEstimatePoolId)
			suite.Require().NoError(err)

			// calculate token out amount using `MultihopSwapExactAmountIn`
			multihopTokenOutAmount, errMultihop := poolmanagerKeeper.RouteExactAmountIn(
				suite.Ctx,
				suite.TestAccs[0],
				test.param.routes,
				test.param.tokenIn,
				test.param.tokenOutMinAmount)

			// calculate token out amount using `EstimateMultihopSwapExactAmountIn`
			estimateMultihopTokenOutAmount, errEstimate := poolmanagerKeeper.MultihopEstimateOutGivenExactAmountIn(
				suite.Ctx,
				test.param.estimateRoutes,
				test.param.tokenIn)

			if test.expectPass {
				suite.Require().NoError(errMultihop, "test: %v", test.name)
				suite.Require().NoError(errEstimate, "test: %v", test.name)
				suite.Require().Equal(multihopTokenOutAmount, estimateMultihopTokenOutAmount)
			} else {
				suite.Require().Error(errMultihop, "test: %v", test.name)
				suite.Require().Error(errEstimate, "test: %v", test.name)
			}
			// ensure that pool state has not been altered after estimation
			firstEstimatePoolAfterSwap, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, firstEstimatePoolId)
			suite.Require().NoError(err)
			secondEstimatePoolAfterSwap, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, secondEstimatePoolId)
			suite.Require().NoError(err)

			suite.Require().Equal(firstEstimatePool, firstEstimatePoolAfterSwap)
			suite.Require().Equal(secondEstimatePool, secondEstimatePoolAfterSwap)
		})
	}
}

// TestEstimateMultihopSwapExactAmountOut tests that the estimation done via `EstimateSwapExactAmountOut`
// results in the same amount of token in as the actual swap.
func (suite *KeeperTestSuite) TestEstimateMultihopSwapExactAmountOut() {
	type param struct {
		routes           []types.SwapAmountOutRoute
		estimateRoutes   []types.SwapAmountOutRoute
		tokenInMaxAmount sdk.Int
		tokenOut         sdk.Coin
	}

	tests := []struct {
		name              string
		param             param
		expectPass        bool
		reducedFeeApplied bool
		poolType          types.PoolType
	}{
		{
			name: "Proper swap: foo -> bar (pool 1), bar -> baz (pool 2)",
			param: param{
				routes: []types.SwapAmountOutRoute{
					{
						PoolId:       1,
						TokenInDenom: foo,
					},
					{
						PoolId:       2,
						TokenInDenom: bar,
					},
				},
				estimateRoutes: []types.SwapAmountOutRoute{
					{
						PoolId:       3,
						TokenInDenom: foo,
					},
					{
						PoolId:       4,
						TokenInDenom: bar,
					},
				},
				tokenInMaxAmount: sdk.NewInt(90000000),
				tokenOut:         sdk.NewCoin(baz, sdk.NewInt(100000)),
			},
			expectPass: true,
		},
		{
			name: "Swap - foo -> uosmo(pool 1) - uosmo(pool 2) -> baz with a half fee applied",
			param: param{
				routes: []types.SwapAmountOutRoute{
					{
						PoolId:       1,
						TokenInDenom: foo,
					},
					{
						PoolId:       2,
						TokenInDenom: uosmo,
					},
				},
				estimateRoutes: []types.SwapAmountOutRoute{
					{
						PoolId:       3,
						TokenInDenom: foo,
					},
					{
						PoolId:       4,
						TokenInDenom: uosmo,
					},
				},
				tokenInMaxAmount: sdk.NewInt(90000000),
				tokenOut:         sdk.NewCoin(baz, sdk.NewInt(100000)),
			},
			expectPass:        true,
			reducedFeeApplied: true,
		},
		{
			name: "Proper swap, stableswap pool: foo -> bar (pool 1), bar -> baz (pool 2)",
			param: param{
				routes: []types.SwapAmountOutRoute{
					{
						PoolId:       1,
						TokenInDenom: foo,
					},
					{
						PoolId:       2,
						TokenInDenom: bar,
					},
				},
				estimateRoutes: []types.SwapAmountOutRoute{
					{
						PoolId:       3,
						TokenInDenom: foo,
					},
					{
						PoolId:       4,
						TokenInDenom: bar,
					},
				},
				tokenInMaxAmount: sdk.NewInt(90000000),
				tokenOut:         sdk.NewCoin(baz, sdk.NewInt(100000)),
			},
			expectPass: true,
			poolType:   types.Stableswap,
		},
		{
			name: "Asserts panic catching in MultihopEstimateInGivenExactAmountOut works: tokenOut more than pool reserves",
			param: param{
				routes: []types.SwapAmountOutRoute{
					{
						PoolId:       1,
						TokenInDenom: foo,
					},
					{
						PoolId:       2,
						TokenInDenom: bar,
					},
				},
				estimateRoutes: []types.SwapAmountOutRoute{
					{
						PoolId:       3,
						TokenInDenom: foo,
					},
					{
						PoolId:       4,
						TokenInDenom: bar,
					},
				},
				tokenInMaxAmount: sdk.NewInt(90000000),
				tokenOut:         sdk.NewCoin(baz, sdk.NewInt(9000000000000000000)),
			},
			expectPass: false,
			poolType:   types.Stableswap,
		},
	}

	for _, test := range tests {
		// Init suite for each test.
		suite.SetupTest()

		suite.Run(test.name, func() {
			poolmanagerKeeper := suite.App.PoolManagerKeeper

			firstEstimatePoolId, secondEstimatePoolId := suite.setupPools(test.poolType, defaultPoolSwapFee)

			firstEstimatePool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, firstEstimatePoolId)
			suite.Require().NoError(err)
			secondEstimatePool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, secondEstimatePoolId)
			suite.Require().NoError(err)

			multihopTokenInAmount, errMultihop := poolmanagerKeeper.RouteExactAmountOut(
				suite.Ctx,
				suite.TestAccs[0],
				test.param.routes,
				test.param.tokenInMaxAmount,
				test.param.tokenOut)

			estimateMultihopTokenInAmount, errEstimate := poolmanagerKeeper.MultihopEstimateInGivenExactAmountOut(
				suite.Ctx,
				test.param.estimateRoutes,
				test.param.tokenOut)

			if test.expectPass {
				suite.Require().NoError(errMultihop, "test: %v", test.name)
				suite.Require().NoError(errEstimate, "test: %v", test.name)
				suite.Require().Equal(multihopTokenInAmount, estimateMultihopTokenInAmount)
			} else {
				suite.Require().Error(errMultihop, "test: %v", test.name)
				suite.Require().Error(errEstimate, "test: %v", test.name)
			}

			// ensure that pool state has not been altered after estimation
			firstEstimatePoolAfterSwap, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, firstEstimatePoolId)
			suite.Require().NoError(err)
			secondEstimatePoolAfterSwap, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, secondEstimatePoolId)
			suite.Require().NoError(err)

			suite.Require().Equal(firstEstimatePool, firstEstimatePoolAfterSwap)
			suite.Require().Equal(secondEstimatePool, secondEstimatePoolAfterSwap)
		})
	}
}

func (suite *KeeperTestSuite) makeGaugesIncentivized(incentivizedGauges []uint64) {
	var records []poolincentivestypes.DistrRecord
	totalWeight := sdk.NewInt(int64(len(incentivizedGauges)))
	for _, gauge := range incentivizedGauges {
		records = append(records, poolincentivestypes.DistrRecord{GaugeId: gauge, Weight: sdk.OneInt()})
	}
	distInfo := poolincentivestypes.DistrInfo{
		TotalWeight: totalWeight,
		Records:     records,
	}
	suite.App.PoolIncentivesKeeper.SetDistrInfo(suite.Ctx, distInfo)
}

func (suite *KeeperTestSuite) calcOutAmountAsSeparateSwaps(osmoFeeReduced bool, routes []poolmanagertypes.SwapAmountOutRoute, tokenOut sdk.Coin) sdk.Coin {
	cacheCtx, _ := suite.Ctx.CacheContext()
	if osmoFeeReduced {
		// extract route from swap
		route := types.SwapAmountOutRoutes(routes)
		// utilizing the extracted route, determine the routeSwapFee and sumOfSwapFees
		// these two variables are used to calculate the overall swap fee utilizing the following formula
		// swapFee = routeSwapFee * ((pool_fee) / (sumOfSwapFees))
		routeSwapFee, sumOfSwapFees, err := suite.App.PoolManagerKeeper.GetOsmoRoutedMultihopTotalSwapFee(suite.Ctx, route)
		suite.Require().NoError(err)
		nextTokenOut := tokenOut
		for i := len(routes) - 1; i >= 0; i-- {
			hop := routes[i]
			// extract the current pool's swap fee
			hopPool, err := suite.App.GAMMKeeper.GetPoolAndPoke(cacheCtx, hop.PoolId)
			suite.Require().NoError(err)
			currentPoolSwapFee := hopPool.GetSwapFee(cacheCtx)
			// utilize the routeSwapFee, sumOfSwapFees, and current pool swap fee to calculate the new reduced swap fee
			swapFee := routeSwapFee.Mul((currentPoolSwapFee.Quo(sumOfSwapFees)))
			// we then do individual swaps until we reach the end of the swap route
			tokenOut, err := suite.App.GAMMKeeper.SwapExactAmountOut(cacheCtx, suite.TestAccs[0], hopPool, hop.TokenInDenom, sdk.NewInt(100000000), nextTokenOut, swapFee)
			suite.Require().NoError(err)
			nextTokenOut = sdk.NewCoin(hop.TokenInDenom, tokenOut)
		}
		return nextTokenOut
	} else {
		nextTokenOut := tokenOut
		for i := len(routes) - 1; i >= 0; i-- {
			hop := routes[i]
			hopPool, err := suite.App.GAMMKeeper.GetPoolAndPoke(cacheCtx, hop.PoolId)
			suite.Require().NoError(err)
			updatedPoolSwapFee := hopPool.GetSwapFee(cacheCtx)
			tokenOut, err := suite.App.GAMMKeeper.SwapExactAmountOut(cacheCtx, suite.TestAccs[0], hopPool, hop.TokenInDenom, sdk.NewInt(100000000), nextTokenOut, updatedPoolSwapFee)
			suite.Require().NoError(err)
			nextTokenOut = sdk.NewCoin(hop.TokenInDenom, tokenOut)
		}
		return nextTokenOut
	}
}

func (suite *KeeperTestSuite) calcInAmountAsSeparateSwaps(osmoFeeReduced bool, routes []poolmanagertypes.SwapAmountInRoute, tokenIn sdk.Coin) sdk.Coin {
	cacheCtx, _ := suite.Ctx.CacheContext()
	if osmoFeeReduced {
		// extract route from swap
		route := types.SwapAmountInRoutes(routes)
		// utilizing the extracted route, determine the routeSwapFee and sumOfSwapFees
		// these two variables are used to calculate the overall swap fee utilizing the following formula
		// swapFee = routeSwapFee * ((pool_fee) / (sumOfSwapFees))
		routeSwapFee, sumOfSwapFees, err := suite.App.PoolManagerKeeper.GetOsmoRoutedMultihopTotalSwapFee(suite.Ctx, route)
		suite.Require().NoError(err)
		nextTokenIn := tokenIn
		for _, hop := range routes {
			// extract the current pool's swap fee
			hopPool, err := suite.App.GAMMKeeper.GetPoolAndPoke(cacheCtx, hop.PoolId)
			suite.Require().NoError(err)
			currentPoolSwapFee := hopPool.GetSwapFee(cacheCtx)
			// utilize the routeSwapFee, sumOfSwapFees, and current pool swap fee to calculate the new reduced swap fee
			swapFee := routeSwapFee.Mul((currentPoolSwapFee.Quo(sumOfSwapFees)))
			// we then do individual swaps until we reach the end of the swap route
			tokenOut, err := suite.App.GAMMKeeper.SwapExactAmountIn(cacheCtx, suite.TestAccs[0], hopPool, nextTokenIn, hop.TokenOutDenom, sdk.OneInt(), swapFee)
			suite.Require().NoError(err)
			nextTokenIn = sdk.NewCoin(hop.TokenOutDenom, tokenOut)
		}
		return nextTokenIn
	} else {
		nextTokenIn := tokenIn
		for _, hop := range routes {
			hopPool, err := suite.App.GAMMKeeper.GetPoolAndPoke(cacheCtx, hop.PoolId)
			suite.Require().NoError(err)
			updatedPoolSwapFee := hopPool.GetSwapFee(cacheCtx)
			tokenOut, err := suite.App.GAMMKeeper.SwapExactAmountIn(cacheCtx, suite.TestAccs[0], hopPool, nextTokenIn, hop.TokenOutDenom, sdk.OneInt(), updatedPoolSwapFee)
			suite.Require().NoError(err)
			nextTokenIn = sdk.NewCoin(hop.TokenOutDenom, tokenOut)
		}
		return nextTokenIn
	}
}

func (suite *KeeperTestSuite) TestSingleSwapExactAmountIn() {
	tests := []struct {
		name                   string
		poolId                 uint64
		poolCoins              sdk.Coins
		poolFee                sdk.Dec
		tokenIn                sdk.Coin
		tokenOutDenom          string
		tokenOutMinAmount      sdk.Int
		expectedTokenOutAmount sdk.Int
		expectError            bool
	}{
		// We have:
		//  - foo: 1000000000000
		//  - bar: 1000000000000
		//  - swapFee: 1%
		//  - foo in: 100000
		//  - bar amount out will be calculated according to the formula
		// 		https://www.wolframalpha.com/input?i=solve+%2810%5E12+%2B+10%5E5+x+0.99%29%2810%5E12+-+x%29+%3D+10%5E24
		// We round down the token amount out, get the result is 98999
		{
			name:                   "Swap - [foo -> bar], 1 percent fee",
			poolId:                 1,
			poolCoins:              sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)),
			poolFee:                defaultPoolSwapFee,
			tokenIn:                sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount:      sdk.NewInt(1),
			tokenOutDenom:          bar,
			expectedTokenOutAmount: sdk.NewInt(98999),
		},
		{
			name:              "Wrong pool id",
			poolId:            2,
			poolCoins:         sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)),
			poolFee:           defaultPoolSwapFee,
			tokenIn:           sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount: sdk.NewInt(1),
			tokenOutDenom:     bar,
			expectError:       true,
		},
		{
			name:              "In denom not exist",
			poolId:            1,
			poolCoins:         sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)),
			poolFee:           defaultPoolSwapFee,
			tokenIn:           sdk.NewCoin(baz, sdk.NewInt(100000)),
			tokenOutMinAmount: sdk.NewInt(1),
			tokenOutDenom:     bar,
			expectError:       true,
		},
		{
			name:              "Out denom not exist",
			poolId:            1,
			poolCoins:         sdk.NewCoins(sdk.NewCoin(foo, defaultInitPoolAmount), sdk.NewCoin(bar, defaultInitPoolAmount)),
			poolFee:           defaultPoolSwapFee,
			tokenIn:           sdk.NewCoin(foo, sdk.NewInt(100000)),
			tokenOutMinAmount: sdk.NewInt(1),
			tokenOutDenom:     baz,
			expectError:       true,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			poolmanagerKeeper := suite.App.PoolManagerKeeper

			suite.FundAcc(suite.TestAccs[0], tc.poolCoins)
			suite.PrepareCustomBalancerPoolFromCoins(tc.poolCoins, balancer.PoolParams{
				SwapFee: tc.poolFee,
				ExitFee: sdk.ZeroDec(),
			})

			// execute the swap
			multihopTokenOutAmount, err := poolmanagerKeeper.SwapExactAmountIn(suite.Ctx, suite.TestAccs[0], tc.poolId, tc.tokenIn, tc.tokenOutDenom, tc.tokenOutMinAmount)
			if tc.expectError {
				suite.Require().Error(err)
			} else {
				// compare the expected tokenOut to the actual tokenOut
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expectedTokenOutAmount.String(), multihopTokenOutAmount.String())
			}
		})
	}
}

// setupPools creates pools of desired type and returns their IDs
func (suite *KeeperTestSuite) setupPools(poolType types.PoolType, poolDefaultSwapFee sdk.Dec) (firstEstimatePoolId, secondEstimatePoolId uint64) {
	switch poolType {
	case types.Stableswap:
		// Prepare 4 pools,
		// Two pools for calculating `MultihopSwapExactAmountOut`
		// and two pools for calculating `EstimateMultihopSwapExactAmountOut`
		suite.PrepareBasicStableswapPool()
		suite.PrepareBasicStableswapPool()

		firstEstimatePoolId = suite.PrepareBasicStableswapPool()

		secondEstimatePoolId = suite.PrepareBasicStableswapPool()
		return
	default:
		// Prepare 4 pools,
		// Two pools for calculating `MultihopSwapExactAmountOut`
		// and two pools for calculating `EstimateMultihopSwapExactAmountOut`
		suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
			SwapFee: poolDefaultSwapFee, // 1%
			ExitFee: sdk.NewDec(0),
		})
		suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
			SwapFee: poolDefaultSwapFee,
			ExitFee: sdk.NewDec(0),
		})

		firstEstimatePoolId = suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
			SwapFee: poolDefaultSwapFee, // 1%
			ExitFee: sdk.NewDec(0),
		})

		secondEstimatePoolId = suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
			SwapFee: poolDefaultSwapFee,
			ExitFee: sdk.NewDec(0),
		})
		return
	}
}
