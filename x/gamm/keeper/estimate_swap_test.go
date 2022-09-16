package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

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
	}{
		{
			name: "Proper swap - foo -> bar(pool 1) - bar(pool 2) -> baz",
			param: param{
				routes: []types.SwapAmountInRoute{
					{
						PoolId:        1,
						TokenOutDenom: "bar",
					},
					{
						PoolId:        2,
						TokenOutDenom: "baz",
					},
				},
				estimateRoutes: []types.SwapAmountInRoute{
					{
						PoolId:        3,
						TokenOutDenom: "bar",
					},
					{
						PoolId:        4,
						TokenOutDenom: "baz",
					},
				},
				tokenIn:           sdk.NewCoin("foo", sdk.NewInt(100000)),
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
						TokenOutDenom: "uosmo",
					},
					{
						PoolId:        2,
						TokenOutDenom: "baz",
					},
				},
				estimateRoutes: []types.SwapAmountInRoute{
					{
						PoolId:        3,
						TokenOutDenom: "uosmo",
					},
					{
						PoolId:        4,
						TokenOutDenom: "baz",
					},
				},
				tokenIn:           sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutMinAmount: sdk.NewInt(1),
			},
			reducedFeeApplied: true,
			expectPass:        true,
		},
	}

	for _, test := range tests {
		// Init suite for each test.
		suite.SetupTest()

		suite.Run(test.name, func() {
			keeper := suite.App.GAMMKeeper
			poolDefaultSwapFee := sdk.NewDecWithPrec(1, 2) // 1%

			// Prepare 4 pools,
			// Two pools for calculating `MultihopSwapExactAmountIn`
			// and two pools for calculating `EstimateMultihopSwapExactAmountIn`
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee,
				ExitFee: sdk.NewDec(0),
			})
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee,
				ExitFee: sdk.NewDec(0),
			})

			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee,
				ExitFee: sdk.NewDec(0),
			})
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee,
				ExitFee: sdk.NewDec(0),
			})

			// calculate token out amount using `MultihopSwapExactAmountIn`
			multihopTokenOutAmount, err := keeper.MultihopSwapExactAmountIn(
				suite.Ctx,
				suite.TestAccs[0],
				test.param.routes,
				test.param.tokenIn,
				test.param.tokenOutMinAmount)
			suite.Require().NoError(err)

			// calculate token out amount using `EstimateMultihopSwapExactAmountIn`
			estimateMultihopTokenOutAmount, err := keeper.EstimateMultihopSwapExactAmountIn(
				suite.Ctx,
				test.param.estimateRoutes,
				test.param.tokenIn,
				test.param.tokenOutMinAmount)
			suite.Require().NoError(err)

			// assure that the token out amount is same
			suite.Require().Equal(multihopTokenOutAmount, estimateMultihopTokenOutAmount)
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
	}{
		{
			name: "Proper swap: foo -> bar (pool 1), bar -> baz (pool 2)",
			param: param{
				routes: []types.SwapAmountOutRoute{
					{
						PoolId:       1,
						TokenInDenom: "foo",
					},
					{
						PoolId:       2,
						TokenInDenom: "bar",
					},
				},
				estimateRoutes: []types.SwapAmountOutRoute{
					{
						PoolId:       3,
						TokenInDenom: "foo",
					},
					{
						PoolId:       4,
						TokenInDenom: "bar",
					},
				},
				tokenInMaxAmount: sdk.NewInt(90000000),
				tokenOut:         sdk.NewCoin("baz", sdk.NewInt(100000)),
			},
			expectPass: true,
		},
		{
			name: "Swap - foo -> uosmo(pool 1) - uosmo(pool 2) -> baz with a half fee applied",
			param: param{
				routes: []types.SwapAmountOutRoute{
					{
						PoolId:       1,
						TokenInDenom: "foo",
					},
					{
						PoolId:       2,
						TokenInDenom: "uosmo",
					},
				},
				estimateRoutes: []types.SwapAmountOutRoute{
					{
						PoolId:       3,
						TokenInDenom: "foo",
					},
					{
						PoolId:       4,
						TokenInDenom: "uosmo",
					},
				},
				tokenInMaxAmount: sdk.NewInt(90000000),
				tokenOut:         sdk.NewCoin("baz", sdk.NewInt(100000)),
			},
			expectPass:        true,
			reducedFeeApplied: true,
		},
	}

	for _, test := range tests {
		// Init suite for each test.
		suite.SetupTest()

		suite.Run(test.name, func() {
			keeper := suite.App.GAMMKeeper
			poolDefaultSwapFee := sdk.NewDecWithPrec(1, 2) // 1%

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
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee, // 1%
				ExitFee: sdk.NewDec(0),
			})
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee,
				ExitFee: sdk.NewDec(0),
			})

			multihopTokenInAmount, err := keeper.MultihopSwapExactAmountOut(
				suite.Ctx,
				suite.TestAccs[0],
				test.param.routes,
				test.param.tokenInMaxAmount,
				test.param.tokenOut)
			suite.Require().NoError(err, "test: %v", test.name)

			estimateMultihopTokenInAmount, err := keeper.EstimateMultihopSwapExactAmountOut(
				suite.Ctx,
				test.param.estimateRoutes,
				test.param.tokenInMaxAmount,
				test.param.tokenOut)
			suite.Require().NoError(err, "test: %v", test.name)

			suite.Require().Equal(multihopTokenInAmount, estimateMultihopTokenInAmount)
		})
	}
}
