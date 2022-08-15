package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v11/tests/mocks"
	"github.com/osmosis-labs/osmosis/v11/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v11/x/gamm/types"
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

			if test.expectPass {
				spotPriceBefore, err := keeper.CalculateSpotPrice(ctx, poolId, test.param.tokenIn.Denom, test.param.tokenOutDenom)
				suite.NoError(err, "test: %v", test.name)

				prevGasConsumed := suite.Ctx.GasMeter().GasConsumed()
				tokenOutAmount, err := keeper.SwapExactAmountIn(ctx, suite.TestAccs[0], poolId, test.param.tokenIn, test.param.tokenOutDenom, test.param.tokenOutMinAmount)
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
				_, err := keeper.SwapExactAmountIn(ctx, suite.TestAccs[0], poolId, test.param.tokenIn, test.param.tokenOutDenom, test.param.tokenOutMinAmount)
				suite.Error(err, "test: %v", test.name)
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

			if test.expectPass {
				spotPriceBefore, err := keeper.CalculateSpotPrice(ctx, poolId, test.param.tokenInDenom, test.param.tokenOut.Denom)
				suite.NoError(err, "test: %v", test.name)

				prevGasConsumed := suite.Ctx.GasMeter().GasConsumed()
				tokenInAmount, err := keeper.SwapExactAmountOut(ctx, suite.TestAccs[0], poolId, test.param.tokenInDenom, test.param.tokenInMaxAmount, test.param.tokenOut)
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
				_, err := keeper.SwapExactAmountOut(suite.Ctx, suite.TestAccs[0], poolId, test.param.tokenInDenom, test.param.tokenInMaxAmount, test.param.tokenOut)
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

			foocoin := sdk.NewCoin("foo", sdk.NewInt(10))

			if tc.expectPass {
				_, err := suite.App.GAMMKeeper.SwapExactAmountIn(suite.Ctx, suite.TestAccs[0], poolId, foocoin, "bar", sdk.ZeroInt())
				suite.Require().NoError(err)
				_, err = suite.App.GAMMKeeper.SwapExactAmountOut(suite.Ctx, suite.TestAccs[0], poolId, "bar", sdk.NewInt(1000000000000000000), foocoin)
				suite.Require().NoError(err)
			} else {
				_, err := suite.App.GAMMKeeper.SwapExactAmountIn(suite.Ctx, suite.TestAccs[0], poolId, foocoin, "bar", sdk.ZeroInt())
				suite.Require().Error(err)
				_, err = suite.App.GAMMKeeper.SwapExactAmountOut(suite.Ctx, suite.TestAccs[0], poolId, "bar", sdk.NewInt(1000000000000000000), foocoin)
				suite.Require().Error(err)
			}
		}
	}
}

// Test two pools -- one is active and should have swaps allowed,
// while the other is inactive and should have swaps frozen.
// As shown in the following test, we can mock a pool by calling
// `mocks.NewMockPool()`, then adding `EXPECT` statements to
// match argument calls, add return values, and more.
// More info at https://github.com/golang/mock
func (suite *KeeperTestSuite) TestInactivePoolFreezeSwaps() {
	// Setup test
	suite.SetupTest()
	testCoin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
	suite.FundAcc(suite.TestAccs[0], defaultAcctFunds)

	// Setup active pool
	activePoolId := suite.PrepareBalancerPool()

	// Setup mock inactive pool
	gammKeeper := suite.App.GAMMKeeper
	ctrl := gomock.NewController(suite.T())
	defer ctrl.Finish()
	inactivePool := mocks.NewMockPoolI(ctrl)
	inactivePoolId := gammKeeper.GetNextPoolNumberAndIncrement(suite.Ctx)
	// Add mock return values for pool -- we need to do this because
	// mock objects don't have interface functions implemented by default.
	inactivePool.EXPECT().IsActive(suite.Ctx).Return(false).AnyTimes()
	inactivePool.EXPECT().GetId().Return(inactivePoolId).AnyTimes()
	gammKeeper.SetPool(suite.Ctx, inactivePool)

	type testCase struct {
		poolId     uint64
		expectPass bool
		name       string
	}
	testCases := []testCase{
		{activePoolId, true, "swap succeeds on active pool"},
		{inactivePoolId, false, "swap fails on inactive pool"},
	}

	for _, test := range testCases {
		suite.Run(test.name, func() {
			// Check swaps
			_, swapInErr := gammKeeper.SwapExactAmountIn(suite.Ctx, suite.TestAccs[0], test.poolId, testCoin, "bar", sdk.ZeroInt())
			_, swapOutErr := gammKeeper.SwapExactAmountOut(suite.Ctx, suite.TestAccs[0], test.poolId, "bar", sdk.NewInt(1000000000000000000), testCoin)
			if test.expectPass {
				suite.Require().NoError(swapInErr)
				suite.Require().NoError(swapOutErr)
			} else {
				suite.Require().Error(swapInErr)
				suite.Require().Error(swapOutErr)
			}
		})
	}

}
