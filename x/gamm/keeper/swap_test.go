package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/tests/mocks"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
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
		name  string
		param param
		// Note: by default swap fee is zero in all tests
		// It is only set to non-zero when this overwrite is non-nil
		swapFeeOverwrite sdk.Dec
		// Note: this is the value by which the original swap fee is divided
		// by if it is non-nil. This is done to test the case where the given
		// parameter swap fee is reduced by more than allowed (max factor of 0.5)
		swapFeeOverwriteQuotient sdk.Dec
		expectPass               bool
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
			name: "boundary valid swap fee given (= 0.5 pool swap fee)",
			param: param{
				tokenIn:           sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutDenom:     "bar",
				tokenOutMinAmount: sdk.NewInt(1),
				expectedTokenOut:  sdk.NewInt(46833),
			},
			swapFeeOverwrite:         sdk.MustNewDecFromStr("0.1"),
			swapFeeOverwriteQuotient: sdk.MustNewDecFromStr("2"),
			expectPass:               true,
		},
		{
			name: "invalid swap fee given (< 0.5 pool swap fee)",
			param: param{
				tokenIn:           sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutDenom:     "bar",
				tokenOutMinAmount: sdk.NewInt(1),
				expectedTokenOut:  sdk.NewInt(49262),
			},
			swapFeeOverwrite:         sdk.MustNewDecFromStr("0.1"),
			swapFeeOverwriteQuotient: sdk.MustNewDecFromStr("3"),
			expectPass:               false,
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
		test := test
		suite.Run(test.name, func() {
			// Init suite for each test.
			suite.SetupTest()
			swapFee := sdk.ZeroDec()
			if !test.swapFeeOverwrite.IsNil() {
				swapFee = test.swapFeeOverwrite
			}
			poolId := suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: swapFee,
				ExitFee: sdk.ZeroDec(),
			})
			if !test.swapFeeOverwriteQuotient.IsNil() {
				swapFee = swapFee.Quo(test.swapFeeOverwriteQuotient)
			}
			keeper := suite.App.GAMMKeeper
			ctx := suite.Ctx
			pool, err := suite.App.GAMMKeeper.GetPool(ctx, poolId)
			suite.NoError(err)

			if test.expectPass {
				spotPriceBefore, err := keeper.CalculateSpotPrice(ctx, poolId, test.param.tokenIn.Denom, test.param.tokenOutDenom)
				suite.NoError(err, "test: %v", test.name)

				prevGasConsumed := suite.Ctx.GasMeter().GasConsumed()
				tokenOutAmount, err := keeper.SwapExactAmountIn(ctx, suite.TestAccs[0], pool, test.param.tokenIn, test.param.tokenOutDenom, test.param.tokenOutMinAmount, swapFee)
				suite.NoError(err, "test: %v", test.name)
				suite.Require().Equal(test.param.expectedTokenOut.String(), tokenOutAmount.String())
				gasConsumedForSwap := suite.Ctx.GasMeter().GasConsumed() - prevGasConsumed
				// We consume `types.GasFeeForSwap` directly, so the extra I/O operation mean we end up consuming more.
				suite.Assert().Greater(gasConsumedForSwap, uint64(types.BalancerGasFeeForSwap))

				suite.AssertEventEmitted(ctx, types.TypeEvtTokenSwapped, 1)

				spotPriceAfter, err := keeper.CalculateSpotPrice(ctx, poolId, test.param.tokenIn.Denom, test.param.tokenOutDenom)
				suite.NoError(err, "test: %v", test.name)

				if !test.swapFeeOverwrite.IsNil() {
					return
				}

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
// TODO: add failing CL pool tests.
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
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			// Init suite for each test.
			suite.SetupTest()
			keeper := suite.App.GAMMKeeper
			ctx := suite.Ctx

			var pool poolmanagertypes.PoolI
			if test.param.poolType == "balancer" {
				poolId := suite.PrepareBalancerPool()
				poolExt, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, poolId)
				suite.NoError(err)
				pool = poolExt.(poolmanagertypes.PoolI)
			} else if test.param.poolType == "stableswap" {
				poolId := suite.PrepareBasicStableswapPool()
				poolExt, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, poolId)
				suite.NoError(err)
				pool = poolExt.(poolmanagertypes.PoolI)
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
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.SetupTest()
			keeper := suite.App.GAMMKeeper
			ctx := suite.Ctx

			var pool poolmanagertypes.PoolI
			var err error

			switch test.param.poolType {
			case "balancer":
				poolId := suite.PrepareBalancerPool()
				poolExt, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, poolId)
				suite.NoError(err)
				pool, _ = poolExt.(poolmanagertypes.PoolI)
			case "stableswap":
				poolId := suite.PrepareBasicStableswapPool()
				poolExt, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, poolId)
				suite.NoError(err)
				pool, _ = poolExt.(poolmanagertypes.PoolI)
			default:
				suite.FailNow("unsupported pool type")
			}

			suite.Require().NotNil(pool)

			swapFee := pool.GetSwapFee(suite.Ctx)

			_, err = keeper.CalcInAmtGivenOut(ctx, pool, test.param.tokenOut, test.param.tokenInDenom, swapFee)

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

// Test two pools -- one is active and should have swaps allowed,
// while the other is inactive and should have swaps frozen.
// As shown in the following test, we can mock a pool by calling
// `mocks.NewMockPool()`, then adding `EXPECT` statements to
// match argument calls, add return values, and more.
// More info at https://github.com/golang/mock
func (suite *KeeperTestSuite) TestInactivePoolFreezeSwaps() {
	// Setup test
	suite.SetupTest()
	testCoin := sdk.NewCoin("foo", sdk.NewInt(10))
	suite.FundAcc(suite.TestAccs[0], defaultAcctFunds)

	// Setup active pool
	activePoolId := suite.PrepareBalancerPool()
	activePool, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, activePoolId)
	suite.Require().NoError(err)

	// Setup mock inactive pool
	gammKeeper := suite.App.GAMMKeeper
	ctrl := gomock.NewController(suite.T())
	defer ctrl.Finish()
	inactivePool := mocks.NewMockCFMMPoolI(ctrl)
	inactivePoolId := activePoolId + 1
	// Add mock return values for pool -- we need to do this because
	// mock objects don't have interface functions implemented by default.
	inactivePool.EXPECT().IsActive(suite.Ctx).Return(false).AnyTimes()
	inactivePool.EXPECT().GetId().Return(inactivePoolId).AnyTimes()
	err = gammKeeper.SetPool(suite.Ctx, inactivePool)
	suite.Require().NoError(err)

	type testCase struct {
		pool       poolmanagertypes.PoolI
		expectPass bool
		name       string
	}
	testCases := []testCase{
		{activePool, true, "swap succeeds on active pool"},
		{inactivePool, false, "swap fails on inactive pool"},
	}

	for _, test := range testCases {
		suite.Run(test.name, func() {
			// Check swaps
			_, swapInErr := suite.App.PoolManagerKeeper.RouteExactAmountIn(
				suite.Ctx,
				suite.TestAccs[0],
				[]poolmanagertypes.SwapAmountInRoute{
					{
						PoolId:        test.pool.GetId(),
						TokenOutDenom: "bar",
					},
				},
				testCoin,
				sdk.ZeroInt(),
			)

			_, swapOutErr := suite.App.PoolManagerKeeper.RouteExactAmountOut(
				suite.Ctx,
				suite.TestAccs[0],
				[]poolmanagertypes.SwapAmountOutRoute{
					{
						PoolId:       test.pool.GetId(),
						TokenInDenom: "bar",
					},
				},
				sdk.NewInt(1000000000000000000),
				testCoin,
			)

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
