package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
)

func (suite *KeeperTestSuite) TestBalancerPoolSimpleSwapExactAmountIn() {
	type param struct {
		tokenIn           sdk.Coin
		tokenOutDenom     string
		tokenOutMinAmount sdk.Int
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
			},
			expectPass: true,
		},
		{
			name: "Proper swap2",
			param: param{
				tokenIn:           sdk.NewCoin("bar", sdk.NewInt(2451783)),
				tokenOutDenom:     "baz",
				tokenOutMinAmount: sdk.NewInt(1),
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
		// Init suite for each test.
		suite.SetupTest()
		poolId := suite.prepareBalancerPool()

		keeper := suite.app.GAMMKeeper

		if test.expectPass {
			spotPriceBefore, err := keeper.CalculateSpotPriceWithSwapFee(suite.ctx, poolId, test.param.tokenIn.Denom, test.param.tokenOutDenom)
			suite.NoError(err, "test: %v", test.name)

			tokenOutAmount, _, err := keeper.SwapExactAmountIn(suite.ctx, acc1, poolId, test.param.tokenIn, test.param.tokenOutDenom, test.param.tokenOutMinAmount)
			suite.NoError(err, "test: %v", test.name)

			spotPriceAfter, err := keeper.CalculateSpotPriceWithSwapFee(suite.ctx, poolId, test.param.tokenIn.Denom, test.param.tokenOutDenom)
			suite.NoError(err, "test: %v", test.name)

			// Ratio of the token out should be between the before spot price and after spot price.
			tradeAvgPrice := test.param.tokenIn.Amount.ToDec().Quo(tokenOutAmount.ToDec())
			suite.True(tradeAvgPrice.GT(spotPriceBefore) && tradeAvgPrice.LT(spotPriceAfter), "test: %v", test.name)
		} else {
			_, _, err := keeper.SwapExactAmountIn(suite.ctx, acc1, poolId, test.param.tokenIn, test.param.tokenOutDenom, test.param.tokenOutMinAmount)
			suite.Error(err, "test: %v", test.name)
		}
	}
}

func (suite *KeeperTestSuite) TestBalancerPoolSimpleSwapExactAmountOut() {
	type param struct {
		tokenInDenom     string
		tokenInMaxAmount sdk.Int
		tokenOut         sdk.Coin
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "Proper swap",
			param: param{
				tokenInDenom:     "foo",
				tokenInMaxAmount: sdk.NewInt(900000000),
				tokenOut:         sdk.NewCoin("bar", sdk.NewInt(100000)),
			},
			expectPass: true,
		},
		{
			name: "Proper swap2",
			param: param{
				tokenInDenom:     "foo",
				tokenInMaxAmount: sdk.NewInt(900000000),
				tokenOut:         sdk.NewCoin("baz", sdk.NewInt(316721)),
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
		// Init suite for each test.
		suite.SetupTest()
		poolId := suite.prepareBalancerPool()

		keeper := suite.app.GAMMKeeper

		if test.expectPass {
			spotPriceBefore, err := keeper.CalculateSpotPriceWithSwapFee(suite.ctx, poolId, test.param.tokenInDenom, test.param.tokenOut.Denom)
			suite.NoError(err, "test: %v", test.name)

			tokenInAmount, _, err := keeper.SwapExactAmountOut(suite.ctx, acc1, poolId, test.param.tokenInDenom, test.param.tokenInMaxAmount, test.param.tokenOut)
			suite.NoError(err, "test: %v", test.name)

			spotPriceAfter, err := keeper.CalculateSpotPriceWithSwapFee(suite.ctx, poolId, test.param.tokenInDenom, test.param.tokenOut.Denom)
			suite.NoError(err, "test: %v", test.name)

			// Ratio of the token out should be between the before spot price and after spot price.
			tradeAvgPrice := tokenInAmount.ToDec().Quo(test.param.tokenOut.Amount.ToDec())
			suite.True(tradeAvgPrice.GT(spotPriceBefore) && tradeAvgPrice.LT(spotPriceAfter), "test: %v", test.name)
		} else {
			_, _, err := keeper.SwapExactAmountOut(suite.ctx, acc1, poolId, test.param.tokenInDenom, test.param.tokenInMaxAmount, test.param.tokenOut)
			suite.Error(err, "test: %v", test.name)
		}
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
		for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
			err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, acc, sdk.NewCoins(
				sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
				sdk.NewCoin("foo", sdk.NewInt(10000000)),
				sdk.NewCoin("bar", sdk.NewInt(10000000)),
				sdk.NewCoin("baz", sdk.NewInt(10000000)),
			))
			if err != nil {
				panic(err)
			}

			poolId := suite.prepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: sdk.NewDec(0),
				ExitFee: sdk.NewDec(0),
			})

			suite.ctx = suite.ctx.WithBlockTime(tc.blockTime)

			foocoin := sdk.NewCoin("foo", sdk.NewInt(10))

			if tc.expectPass {
				_, _, err = suite.app.GAMMKeeper.SwapExactAmountIn(suite.ctx, acc1, poolId, foocoin, "bar", sdk.ZeroInt())
				suite.Require().NoError(err)
				_, _, err = suite.app.GAMMKeeper.SwapExactAmountOut(suite.ctx, acc1, poolId, "bar", sdk.NewInt(1000000000000000000), foocoin)
				suite.Require().NoError(err)
			} else {
				_, _, err = suite.app.GAMMKeeper.SwapExactAmountIn(suite.ctx, acc1, poolId, foocoin, "bar", sdk.ZeroInt())
				suite.Require().Error(err)
				_, _, err = suite.app.GAMMKeeper.SwapExactAmountOut(suite.ctx, acc1, poolId, "bar", sdk.NewInt(1000000000000000000), foocoin)
				suite.Require().Error(err)
			}
		}
	}
}
