package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v13/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

var _ = suite.TestingSuite(nil)

func (suite *KeeperTestSuite) TestBalancerPoolSimpleSwapExactAmountIn() {
	type param struct {
		tokenIn           sdk.Coin
		tokenOutDenom     string
		tokenOutMinAmount sdk.Int
		expectedTokenOut  sdk.Int
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "Proper swap",
			param: param{
				tokenIn:           sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutDenom:     "bar",
				tokenOutMinAmount: sdk.NewInt(1),
				expectedTokenOut:  sdk.NewInt(49262),
			},
			expectPass: true,
		},
		{
			name: "Proper swap2",
			param: param{
				tokenIn:           sdk.NewCoin("bar", sdk.NewInt(2451783)),
				tokenOutDenom:     "baz",
				tokenOutMinAmount: sdk.NewInt(1),
				expectedTokenOut:  sdk.NewInt(1167843),
			},
			expectPass: true,
		},
		{
			name: "out is lesser than min amount",
			param: param{
				tokenIn:           sdk.NewCoin("bar", sdk.NewInt(2451783)),
				tokenOutDenom:     "baz",
				tokenOutMinAmount: sdk.NewInt(9000000),
			},
			expectPass: false,
		},
		{
			name: "in and out denom are same",
			param: param{
				tokenIn:           sdk.NewCoin("bar", sdk.NewInt(2451783)),
				tokenOutDenom:     "bar",
				tokenOutMinAmount: sdk.NewInt(1),
			},
			expectPass: false,
		},
		{
			name: "unknown in denom",
			param: param{
				tokenIn:           sdk.NewCoin("bara", sdk.NewInt(2451783)),
				tokenOutDenom:     "bar",
				tokenOutMinAmount: sdk.NewInt(1),
			},
			expectPass: false,
		},
		{
			name: "unknown out denom",
			param: param{
				tokenIn:           sdk.NewCoin("bar", sdk.NewInt(2451783)),
				tokenOutDenom:     "bara",
				tokenOutMinAmount: sdk.NewInt(1),
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			// Init suite for each test.
			suite.SetupTest()
			poolId := suite.PrepareBalancerPool()
			keeper := suite.App.GAMMKeeper
			ctx := suite.Ctx
			pool, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, poolId)
			suite.NoError(err)
			swapFee := pool.GetSwapFee(suite.Ctx)

			if test.expectPass {
				spotPriceBefore, err := keeper.CalculateSpotPrice(ctx, poolId, test.param.tokenIn.Denom, test.param.tokenOutDenom)
				suite.NoError(err, "test: %v", test.name)

				prevGasConsumed := suite.Ctx.GasMeter().GasConsumed()
				tokenOutAmount, err := keeper.SwapExactAmountIn(ctx, suite.TestAccs[0], pool, test.param.tokenIn, test.param.tokenOutDenom, test.param.tokenOutMinAmount, swapFee)
				suite.NoError(err, "test: %v", test.name)
				suite.True(tokenOutAmount.Equal(test.param.expectedTokenOut), "test: %v", test.name)
				gasConsumedForSwap := suite.Ctx.GasMeter().GasConsumed() - prevGasConsumed
				// We consume `types.GasFeeForSwap` directly, so the extra I/O operation mean we end up consuming more.
				suite.Assert().Greater(gasConsumedForSwap, uint64(types.BalancerGasFeeForSwap))

				suite.AssertEventEmitted(ctx, types.TypeEvtTokenSwapped, 1)

				spotPriceAfter, err := keeper.CalculateSpotPrice(ctx, poolId, test.param.tokenIn.Denom, test.param.tokenOutDenom)
				suite.NoError(err, "test: %v", test.name)

				// Ratio of the token out should be between the before spot price and after spot price.
				tradeAvgPrice := test.param.tokenIn.Amount.ToDec().Quo(tokenOutAmount.ToDec())
				suite.True(tradeAvgPrice.GT(spotPriceBefore) && tradeAvgPrice.LT(spotPriceAfter), "test: %v", test.name)
			} else {
				_, err := keeper.SwapExactAmountIn(ctx, suite.TestAccs[0], pool, test.param.tokenIn, test.param.tokenOutDenom, test.param.tokenOutMinAmount, swapFee)
				suite.Error(err, "test: %v", test.name)
			}
		})
	}
}

// TestCalcOutAmtGivenIn only tests that balancer and stableswap pools are type casted correctly while concentratedliquidity pools fail
func (suite *KeeperTestSuite) TestCalcOutAmtGivenIn() {
	type param struct {
		poolType      string
		tokenIn       sdk.Coin
		tokenOutDenom string
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "balancer",
			param: param{
				poolType:      "balancer",
				tokenIn:       sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutDenom: "bar",
			},
			expectPass: true,
		},
		{
			name: "stableswap",
			param: param{
				poolType:      "stableswap",
				tokenIn:       sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutDenom: "bar",
			},
			expectPass: true,
		},
		{
			name: "concentratedliquidity",
			param: param{
				poolType:      "concentratedliquidity",
				tokenIn:       sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutDenom: "bar",
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			// Init suite for each test.
			suite.SetupTest()
			keeper := suite.App.GAMMKeeper
			ctx := suite.Ctx

			var pool swaproutertypes.PoolI
			if test.param.poolType == "balancer" {
				poolId := suite.PrepareBalancerPool()
				poolExt, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, poolId)
				suite.NoError(err)
				pool = poolExt.(swaproutertypes.PoolI)
			} else if test.param.poolType == "stableswap" {
				poolId := suite.PrepareBasicStableswapPool()
				poolExt, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, poolId)
				suite.NoError(err)
				pool = poolExt.(swaproutertypes.PoolI)
			} else if test.param.poolType == "concentratedliquidity" {
				poolExt, err := suite.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(ctx, 1, "bar", "foo", sdk.MustNewDecFromStr("70.710678118654752440"), sdk.NewInt(85176))
				suite.NoError(err)
				pool = poolExt.(swaproutertypes.PoolI)
			}

			swapFee := pool.GetSwapFee(suite.Ctx)

			_, err := keeper.CalcOutAmtGivenIn(ctx, pool, test.param.tokenIn, test.param.tokenOutDenom, swapFee)

			if test.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestCalcInAmtGivenOut only tests that balancer and stableswap pools are type casted correctly while concentratedliquidity pools fail
func (suite *KeeperTestSuite) TestCalcInAmtGivenOut() {
	type param struct {
		poolType     string
		tokenOut     sdk.Coin
		tokenInDenom string
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "balancer",
			param: param{
				poolType:     "balancer",
				tokenOut:     sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenInDenom: "bar",
			},
			expectPass: true,
		},
		{
			name: "stableswap",
			param: param{
				poolType:     "stableswap",
				tokenOut:     sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenInDenom: "bar",
			},
			expectPass: true,
		},
		{
			name: "concentratedliquidity",
			param: param{
				poolType:     "concentratedliquidity",
				tokenOut:     sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenInDenom: "bar",
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			// Init suite for each test.
			suite.SetupTest()
			keeper := suite.App.GAMMKeeper
			ctx := suite.Ctx

			var pool swaproutertypes.PoolI
			if test.param.poolType == "balancer" {
				poolId := suite.PrepareBalancerPool()
				poolExt, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, poolId)
				suite.NoError(err)
				pool = poolExt.(swaproutertypes.PoolI)
			} else if test.param.poolType == "stableswap" {
				poolId := suite.PrepareBasicStableswapPool()
				poolExt, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, poolId)
				suite.NoError(err)
				pool = poolExt.(swaproutertypes.PoolI)
			} else if test.param.poolType == "concentratedliquidity" {
				poolExt, err := suite.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(ctx, 1, "bar", "foo", sdk.MustNewDecFromStr("70.710678118654752440"), sdk.NewInt(85176))
				suite.NoError(err)
				pool = poolExt.(swaproutertypes.PoolI)
			}

			swapFee := pool.GetSwapFee(suite.Ctx)

			_, err := keeper.CalcInAmtGivenOut(ctx, pool, test.param.tokenOut, test.param.tokenInDenom, swapFee)

			if test.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestBalancerPoolSimpleSwapExactAmountOut() {
	type param struct {
		tokenInDenom          string
		tokenInMaxAmount      sdk.Int
		tokenOut              sdk.Coin
		expectedTokenInAmount sdk.Int
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "Proper swap",
			param: param{
				tokenInDenom:          "foo",
				tokenInMaxAmount:      sdk.NewInt(900000000),
				tokenOut:              sdk.NewCoin("bar", sdk.NewInt(100000)),
				expectedTokenInAmount: sdk.NewInt(206165),
			},
			expectPass: true,
		},
		{
			name: "Proper swap2",
			param: param{
				tokenInDenom:          "foo",
				tokenInMaxAmount:      sdk.NewInt(900000000),
				tokenOut:              sdk.NewCoin("baz", sdk.NewInt(316721)),
				expectedTokenInAmount: sdk.NewInt(1084571),
			},
			expectPass: true,
		},
		{
			name: "in is greater than max",
			param: param{
				tokenInDenom:     "foo",
				tokenInMaxAmount: sdk.NewInt(100),
				tokenOut:         sdk.NewCoin("baz", sdk.NewInt(316721)),
			},
			expectPass: false,
		},
		{
			name: "pool doesn't have enough token to out",
			param: param{
				tokenInDenom:     "foo",
				tokenInMaxAmount: sdk.NewInt(900000000),
				tokenOut:         sdk.NewCoin("baz", sdk.NewInt(99316721)),
			},
			expectPass: false,
		},
		{
			name: "unknown in denom",
			param: param{
				tokenInDenom:     "fooz",
				tokenInMaxAmount: sdk.NewInt(900000000),
				tokenOut:         sdk.NewCoin("bar", sdk.NewInt(100000)),
			},
			expectPass: false,
		},
		{
			name: "unknown out denom",
			param: param{
				tokenInDenom:     "foo",
				tokenInMaxAmount: sdk.NewInt(900000000),
				tokenOut:         sdk.NewCoin("barz", sdk.NewInt(100000)),
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			// Init suite for each test.
			suite.SetupTest()
			poolId := suite.PrepareBalancerPool()

			keeper := suite.App.GAMMKeeper
			ctx := suite.Ctx
			pool, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, poolId)
			suite.Require().NoError(err)
			swapFee := pool.GetSwapFee(suite.Ctx)

			if test.expectPass {
				spotPriceBefore, err := keeper.CalculateSpotPrice(ctx, poolId, test.param.tokenInDenom, test.param.tokenOut.Denom)
				suite.NoError(err, "test: %v", test.name)

				prevGasConsumed := suite.Ctx.GasMeter().GasConsumed()

				tokenInAmount, err := keeper.SwapExactAmountOut(ctx, suite.TestAccs[0], pool, test.param.tokenInDenom, test.param.tokenInMaxAmount, test.param.tokenOut, swapFee)
				suite.NoError(err, "test: %v", test.name)
				suite.True(tokenInAmount.Equal(test.param.expectedTokenInAmount),
					"test: %v\n expect_eq actual: %s, expected: %s",
					test.name, tokenInAmount, test.param.expectedTokenInAmount)
				gasConsumedForSwap := suite.Ctx.GasMeter().GasConsumed() - prevGasConsumed
				// We consume `types.GasFeeForSwap` directly, so the extra I/O operation mean we end up consuming more.
				suite.Assert().Greater(gasConsumedForSwap, uint64(types.BalancerGasFeeForSwap))

				suite.AssertEventEmitted(ctx, types.TypeEvtTokenSwapped, 1)

				spotPriceAfter, err := keeper.CalculateSpotPrice(ctx, poolId, test.param.tokenInDenom, test.param.tokenOut.Denom)
				suite.NoError(err, "test: %v", test.name)

				// Ratio of the token out should be between the before spot price and after spot price.
				tradeAvgPrice := tokenInAmount.ToDec().Quo(test.param.tokenOut.Amount.ToDec())
				suite.True(tradeAvgPrice.GT(spotPriceBefore) && tradeAvgPrice.LT(spotPriceAfter), "test: %v", test.name)
			} else {
				_, err := keeper.SwapExactAmountOut(suite.Ctx, suite.TestAccs[0], pool, test.param.tokenInDenom, test.param.tokenInMaxAmount, test.param.tokenOut, swapFee)
				suite.Error(err, "test: %v", test.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestActiveBalancerPoolSwap() {
	type testCase struct {
		blockTime  time.Time
		expectPass bool
	}

	testCases := []testCase{
		{time.Unix(1000, 0), true},
		{time.Unix(2000, 0), true},
	}

	for _, tc := range testCases {
		suite.SetupTest()

		// Mint some assets to the accounts.
		for _, acc := range suite.TestAccs {
			suite.FundAcc(acc, defaultAcctFunds)

			poolId := suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: sdk.NewDec(0),
				ExitFee: sdk.NewDec(0),
			})

			suite.Ctx = suite.Ctx.WithBlockTime(tc.blockTime)
			pool, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, poolId)
			suite.Require().NoError(err)
			swapFee := pool.GetSwapFee(suite.Ctx)

			foocoin := sdk.NewCoin("foo", sdk.NewInt(10))

			if tc.expectPass {
				_, err := suite.App.GAMMKeeper.SwapExactAmountIn(suite.Ctx, suite.TestAccs[0], pool, foocoin, "bar", sdk.ZeroInt(), swapFee)
				suite.Require().NoError(err)
				_, err = suite.App.GAMMKeeper.SwapExactAmountOut(suite.Ctx, suite.TestAccs[0], pool, "bar", sdk.NewInt(1000000000000000000), foocoin, swapFee)
				suite.Require().NoError(err)
			} else {
				_, err := suite.App.GAMMKeeper.SwapExactAmountIn(suite.Ctx, suite.TestAccs[0], pool, foocoin, "bar", sdk.ZeroInt(), swapFee)
				suite.Require().Error(err)
				_, err = suite.App.GAMMKeeper.SwapExactAmountOut(suite.Ctx, suite.TestAccs[0], pool, "bar", sdk.NewInt(1000000000000000000), foocoin, swapFee)
				suite.Require().Error(err)
			}
		}
	}
}
