package keeper_test

import (
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

var _ = suite.TestingSuite(nil)

func (s *KeeperTestSuite) TestBalancerPoolSimpleSwapExactAmountIn() {
	type param struct {
		tokenIn           sdk.Coin
		tokenOutDenom     string
		tokenOutMinAmount osmomath.Int
		expectedTokenOut  osmomath.Int
	}

	tests := []struct {
		name  string
		param param
		// Note: by default spread factor is zero in all tests
		// It is only set to non-zero when this overwrite is non-nil
		spreadFactorOverwrite osmomath.Dec
		// Note: this is the value by which the original spread factor is divided
		// by if it is non-nil. This is done to test the case where the given
		// parameter spread factor is reduced by more than allowed (max factor of 0.5)
		spreadFactorOverwriteQuotient osmomath.Dec
		expectPass                    bool
	}{
		{
			name: "Proper swap",
			param: param{
				tokenIn:           sdk.NewCoin("foo", osmomath.NewInt(100000)),
				tokenOutDenom:     "bar",
				tokenOutMinAmount: osmomath.NewInt(1),
				expectedTokenOut:  osmomath.NewInt(49262),
			},
			expectPass: true,
		},
		{
			name: "Proper swap2",
			param: param{
				tokenIn:           sdk.NewCoin("bar", osmomath.NewInt(2451783)),
				tokenOutDenom:     "baz",
				tokenOutMinAmount: osmomath.NewInt(1),
				expectedTokenOut:  osmomath.NewInt(1167843),
			},
			expectPass: true,
		},
		{
			name: "boundary valid spread factor given (= 0.5 pool spread factor)",
			param: param{
				tokenIn:           sdk.NewCoin("foo", osmomath.NewInt(100000)),
				tokenOutDenom:     "bar",
				tokenOutMinAmount: osmomath.NewInt(1),
				expectedTokenOut:  osmomath.NewInt(46833),
			},
			spreadFactorOverwrite:         osmomath.MustNewDecFromStr("0.1"),
			spreadFactorOverwriteQuotient: osmomath.MustNewDecFromStr("2"),
			expectPass:                    true,
		},
		{
			name: "invalid spread factor given (< 0.5 pool spread factor)",
			param: param{
				tokenIn:           sdk.NewCoin("foo", osmomath.NewInt(100000)),
				tokenOutDenom:     "bar",
				tokenOutMinAmount: osmomath.NewInt(1),
				expectedTokenOut:  osmomath.NewInt(49262),
			},
			spreadFactorOverwrite:         osmomath.MustNewDecFromStr("0.1"),
			spreadFactorOverwriteQuotient: osmomath.MustNewDecFromStr("3"),
			expectPass:                    false,
		},
		{
			name: "out is lesser than min amount",
			param: param{
				tokenIn:           sdk.NewCoin("bar", osmomath.NewInt(2451783)),
				tokenOutDenom:     "baz",
				tokenOutMinAmount: osmomath.NewInt(9000000),
			},
			expectPass: false,
		},
		{
			name: "in and out denom are same",
			param: param{
				tokenIn:           sdk.NewCoin("bar", osmomath.NewInt(2451783)),
				tokenOutDenom:     "bar",
				tokenOutMinAmount: osmomath.NewInt(1),
			},
			expectPass: false,
		},
		{
			name: "unknown in denom",
			param: param{
				tokenIn:           sdk.NewCoin("bara", osmomath.NewInt(2451783)),
				tokenOutDenom:     "bar",
				tokenOutMinAmount: osmomath.NewInt(1),
			},
			expectPass: false,
		},
		{
			name: "unknown out denom",
			param: param{
				tokenIn:           sdk.NewCoin("bar", osmomath.NewInt(2451783)),
				tokenOutDenom:     "bara",
				tokenOutMinAmount: osmomath.NewInt(1),
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		test := test
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()
			spreadFactor := osmomath.ZeroDec()
			if !test.spreadFactorOverwrite.IsNil() {
				spreadFactor = test.spreadFactorOverwrite
			}
			poolId := s.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: spreadFactor,
				ExitFee: osmomath.ZeroDec(),
			})
			if !test.spreadFactorOverwriteQuotient.IsNil() {
				spreadFactor = spreadFactor.Quo(test.spreadFactorOverwriteQuotient)
			}
			keeper := s.App.GAMMKeeper
			ctx := s.Ctx
			pool, err := s.App.GAMMKeeper.GetPool(ctx, poolId)
			s.NoError(err)

			if test.expectPass {
				spotPriceBefore, err := keeper.CalculateSpotPrice(ctx, poolId, test.param.tokenIn.Denom, test.param.tokenOutDenom)
				s.NoError(err, "test: %v", test.name)

				prevGasConsumed := s.Ctx.GasMeter().GasConsumed()
				tokenOutAmount, err := keeper.SwapExactAmountIn(ctx, s.TestAccs[0], pool, test.param.tokenIn, test.param.tokenOutDenom, test.param.tokenOutMinAmount, spreadFactor)
				s.NoError(err, "test: %v", test.name)
				s.Require().Equal(test.param.expectedTokenOut.String(), tokenOutAmount.String())
				gasConsumedForSwap := s.Ctx.GasMeter().GasConsumed() - prevGasConsumed
				// We consume `types.GasFeeForSwap` directly, so the extra I/O operation mean we end up consuming more.
				s.Assert().Greater(gasConsumedForSwap, uint64(types.BalancerGasFeeForSwap))

				s.AssertEventEmitted(ctx, types.TypeEvtTokenSwapped, 1)

				spotPriceAfter, err := keeper.CalculateSpotPrice(ctx, poolId, test.param.tokenIn.Denom, test.param.tokenOutDenom)
				s.NoError(err, "test: %v", test.name)

				if !test.spreadFactorOverwrite.IsNil() {
					return
				}

				// Ratio of the token out should be between the before spot price and after spot price.
				tradeAvgPrice := osmomath.BigDecFromDec(test.param.tokenIn.Amount.ToLegacyDec().Quo(tokenOutAmount.ToLegacyDec()))
				s.True(tradeAvgPrice.GT(spotPriceBefore) && tradeAvgPrice.LT(spotPriceAfter), "test: %v", test.name)
			} else {
				_, err := keeper.SwapExactAmountIn(ctx, s.TestAccs[0], pool, test.param.tokenIn, test.param.tokenOutDenom, test.param.tokenOutMinAmount, spreadFactor)
				s.Error(err, "test: %v", test.name)
			}
		})
	}
}

// TestCalcOutAmtGivenIn only tests that balancer and stableswap pools are type casted correctly while concentratedliquidity pools fail
// TODO: add failing CL pool tests.
func (s *KeeperTestSuite) TestCalcOutAmtGivenIn() {
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
				tokenIn:       sdk.NewCoin("foo", osmomath.NewInt(100000)),
				tokenOutDenom: "bar",
			},
			expectPass: true,
		},
		{
			name: "stableswap",
			param: param{
				poolType:      "stableswap",
				tokenIn:       sdk.NewCoin("foo", osmomath.NewInt(100000)),
				tokenOutDenom: "bar",
			},
			expectPass: true,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()
			keeper := s.App.GAMMKeeper
			ctx := s.Ctx

			var pool poolmanagertypes.PoolI
			if test.param.poolType == "balancer" {
				poolId := s.PrepareBalancerPool()
				poolExt, err := s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
				s.NoError(err)
				pool = poolExt.(poolmanagertypes.PoolI)
			} else if test.param.poolType == "stableswap" {
				poolId := s.PrepareBasicStableswapPool()
				poolExt, err := s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
				s.NoError(err)
				pool = poolExt.(poolmanagertypes.PoolI)
			}

			spreadFactor := pool.GetSpreadFactor(s.Ctx)

			_, err := keeper.CalcOutAmtGivenIn(ctx, pool, test.param.tokenIn, test.param.tokenOutDenom, spreadFactor)

			if test.expectPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCalcInAmtGivenOut() {
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
				tokenOut:     sdk.NewCoin("foo", osmomath.NewInt(100000)),
				tokenInDenom: "bar",
			},
			expectPass: true,
		},
		{
			name: "stableswap",
			param: param{
				poolType:     "stableswap",
				tokenOut:     sdk.NewCoin("foo", osmomath.NewInt(100000)),
				tokenInDenom: "bar",
			},
			expectPass: true,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			keeper := s.App.GAMMKeeper
			ctx := s.Ctx

			var pool poolmanagertypes.PoolI
			var err error

			switch test.param.poolType {
			case "balancer":
				poolId := s.PrepareBalancerPool()
				poolExt, err := s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
				s.NoError(err)
				pool, _ = poolExt.(poolmanagertypes.PoolI)
			case "stableswap":
				poolId := s.PrepareBasicStableswapPool()
				poolExt, err := s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
				s.NoError(err)
				pool, _ = poolExt.(poolmanagertypes.PoolI)
			default:
				s.FailNow("unsupported pool type")
			}

			s.Require().NotNil(pool)

			spreadFactor := pool.GetSpreadFactor(s.Ctx)

			_, err = keeper.CalcInAmtGivenOut(ctx, pool, test.param.tokenOut, test.param.tokenInDenom, spreadFactor)

			if test.expectPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestBalancerPoolSimpleSwapExactAmountOut() {
	type param struct {
		tokenInDenom          string
		tokenInMaxAmount      osmomath.Int
		tokenOut              sdk.Coin
		expectedTokenInAmount osmomath.Int
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
				tokenInMaxAmount:      osmomath.NewInt(900000000),
				tokenOut:              sdk.NewCoin("bar", osmomath.NewInt(100000)),
				expectedTokenInAmount: osmomath.NewInt(206165),
			},
			expectPass: true,
		},
		{
			name: "Proper swap2",
			param: param{
				tokenInDenom:          "foo",
				tokenInMaxAmount:      osmomath.NewInt(900000000),
				tokenOut:              sdk.NewCoin("baz", osmomath.NewInt(316721)),
				expectedTokenInAmount: osmomath.NewInt(1084571),
			},
			expectPass: true,
		},
		{
			name: "in is greater than max",
			param: param{
				tokenInDenom:     "foo",
				tokenInMaxAmount: osmomath.NewInt(100),
				tokenOut:         sdk.NewCoin("baz", osmomath.NewInt(316721)),
			},
			expectPass: false,
		},
		{
			name: "pool doesn't have enough token to out",
			param: param{
				tokenInDenom:     "foo",
				tokenInMaxAmount: osmomath.NewInt(900000000),
				tokenOut:         sdk.NewCoin("baz", osmomath.NewInt(99316721)),
			},
			expectPass: false,
		},
		{
			name: "unknown in denom",
			param: param{
				tokenInDenom:     "fooz",
				tokenInMaxAmount: osmomath.NewInt(900000000),
				tokenOut:         sdk.NewCoin("bar", osmomath.NewInt(100000)),
			},
			expectPass: false,
		},
		{
			name: "unknown out denom",
			param: param{
				tokenInDenom:     "foo",
				tokenInMaxAmount: osmomath.NewInt(900000000),
				tokenOut:         sdk.NewCoin("barz", osmomath.NewInt(100000)),
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()
			poolId := s.PrepareBalancerPool()

			keeper := s.App.GAMMKeeper
			ctx := s.Ctx
			pool, err := s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
			s.Require().NoError(err)
			spreadFactor := pool.GetSpreadFactor(s.Ctx)

			if test.expectPass {
				spotPriceBefore, err := keeper.CalculateSpotPrice(ctx, poolId, test.param.tokenInDenom, test.param.tokenOut.Denom)
				s.NoError(err, "test: %v", test.name)

				prevGasConsumed := s.Ctx.GasMeter().GasConsumed()

				tokenInAmount, err := keeper.SwapExactAmountOut(ctx, s.TestAccs[0], pool, test.param.tokenInDenom, test.param.tokenInMaxAmount, test.param.tokenOut, spreadFactor)
				s.NoError(err, "test: %v", test.name)
				s.True(tokenInAmount.Equal(test.param.expectedTokenInAmount),
					"test: %v\n expect_eq actual: %s, expected: %s",
					test.name, tokenInAmount, test.param.expectedTokenInAmount)
				gasConsumedForSwap := s.Ctx.GasMeter().GasConsumed() - prevGasConsumed
				// We consume `types.GasFeeForSwap` directly, so the extra I/O operation mean we end up consuming more.
				s.Assert().Greater(gasConsumedForSwap, uint64(types.BalancerGasFeeForSwap))

				s.AssertEventEmitted(ctx, types.TypeEvtTokenSwapped, 1)

				spotPriceAfter, err := keeper.CalculateSpotPrice(ctx, poolId, test.param.tokenInDenom, test.param.tokenOut.Denom)
				s.NoError(err, "test: %v", test.name)

				// Ratio of the token out should be between the before spot price and after spot price.
				tradeAvgPrice := osmomath.BigDecFromDec(tokenInAmount.ToLegacyDec().Quo(test.param.tokenOut.Amount.ToLegacyDec()))
				s.True(tradeAvgPrice.GT(spotPriceBefore) && tradeAvgPrice.LT(spotPriceAfter), "test: %v", test.name)
			} else {
				_, err := keeper.SwapExactAmountOut(s.Ctx, s.TestAccs[0], pool, test.param.tokenInDenom, test.param.tokenInMaxAmount, test.param.tokenOut, spreadFactor)
				s.Error(err, "test: %v", test.name)
			}
		})
	}
}

func (s *KeeperTestSuite) TestActiveBalancerPoolSwap() {
	type testCase struct {
		blockTime  time.Time
		expectPass bool
	}

	testCases := []testCase{
		{time.Unix(1000, 0), true},
		{time.Unix(2000, 0), true},
	}

	for _, tc := range testCases {
		s.SetupTest()

		// Mint some assets to the accounts.
		for _, acc := range s.TestAccs {
			s.FundAcc(acc, defaultAcctFunds)

			poolId := s.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: osmomath.NewDec(0),
				ExitFee: osmomath.NewDec(0),
			})

			s.Ctx = s.Ctx.WithBlockTime(tc.blockTime)
			pool, err := s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
			s.Require().NoError(err)
			spreadFactor := pool.GetSpreadFactor(s.Ctx)

			foocoin := sdk.NewCoin("foo", osmomath.NewInt(10))

			if tc.expectPass {
				_, err := s.App.GAMMKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], pool, foocoin, "bar", osmomath.ZeroInt(), spreadFactor)
				s.Require().NoError(err)
				_, err = s.App.GAMMKeeper.SwapExactAmountOut(s.Ctx, s.TestAccs[0], pool, "bar", osmomath.NewInt(1000000000000000000), foocoin, spreadFactor)
				s.Require().NoError(err)
			} else {
				_, err := s.App.GAMMKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], pool, foocoin, "bar", osmomath.ZeroInt(), spreadFactor)
				s.Require().Error(err)
				_, err = s.App.GAMMKeeper.SwapExactAmountOut(s.Ctx, s.TestAccs[0], pool, "bar", osmomath.NewInt(1000000000000000000), foocoin, spreadFactor)
				s.Require().Error(err)
			}
		}
	}
}

func (s *KeeperTestSuite) TestOutOfGasError() {
	s.SetupTest()
	poolId := s.PrepareBalancerPool()

	pool, err := s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
	s.Require().NoError(err)
	foocoin := sdk.NewCoin("foo", osmomath.NewInt(10))
	spreadFactor := pool.GetSpreadFactor(s.Ctx)
	ctx := s.Ctx.WithGasMeter(storetypes.NewGasMeter(10))
	_, err = s.App.GAMMKeeper.SwapExactAmountIn(ctx, s.TestAccs[0], pool, foocoin, "bar", osmomath.ZeroInt(), spreadFactor)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "lack of gas")
}

// UNFORKINGNOTE: This test really wasn't testing anything important
// With the unfork, we can no longer utilize mocks when calling SetPools, since
// the interface needs to be registered with codec, and the mocks aren't wired to do that.
//
// // Test two pools -- one is active and should have swaps allowed,
// // while the other is inactive and should have swaps frozen.
// // As shown in the following test, we can mock a pool by calling
// // `mocks.NewMockPool()`, then adding `EXPECT` statements to
// // match argument calls, add return values, and more.
// // More info at https://github.com/golang/mock
// func (s *KeeperTestSuite) TestInactivePoolFreezeSwaps() {
// 	// Setup test
// 	s.SetupTest()
// 	testCoin := sdk.NewCoin("foo", osmomath.NewInt(10))
// 	s.FundAcc(s.TestAccs[0], defaultAcctFunds)

// 	// Setup active pool
// 	activePoolId := s.PrepareBalancerPool()
// 	activePool, err := s.App.GAMMKeeper.GetPool(s.Ctx, activePoolId)
// 	s.Require().NoError(err)

// 	// Setup mock inactive pool
// 	gammKeeper := s.App.GAMMKeeper
// 	ctrl := gomock.NewController(s.T())
// 	defer ctrl.Finish()
// 	inactivePool := mocks.NewMockCFMMPoolI(ctrl)
// 	inactivePoolId := activePoolId + 1
// 	// Add mock return values for pool -- we need to do this because
// 	// mock objects don't have interface functions implemented by default.
// 	inactivePool.EXPECT().IsActive(s.Ctx).Return(false).AnyTimes()
// 	inactivePool.EXPECT().GetId().Return(inactivePoolId).AnyTimes()
// 	err = gammKeeper.SetPool(s.Ctx, inactivePool)
// 	s.Require().NoError(err)

// 	type testCase struct {
// 		pool       poolmanagertypes.PoolI
// 		expectPass bool
// 		name       string
// 	}
// 	testCases := []testCase{
// 		{activePool, true, "swap succeeds on active pool"},
// 		{inactivePool, false, "swap fails on inactive pool"},
// 	}

// 	for _, test := range testCases {
// 		s.Run(test.name, func() {
// 			// Check swaps
// 			_, swapInErr := s.App.PoolManagerKeeper.RouteExactAmountIn(
// 				s.Ctx,
// 				s.TestAccs[0],
// 				[]poolmanagertypes.SwapAmountInRoute{
// 					{
// 						PoolId:        test.pool.GetId(),
// 						TokenOutDenom: "bar",
// 					},
// 				},
// 				testCoin,
// 				osmomath.ZeroInt(),
// 			)

// 			_, swapOutErr := s.App.PoolManagerKeeper.RouteExactAmountOut(
// 				s.Ctx,
// 				s.TestAccs[0],
// 				[]poolmanagertypes.SwapAmountOutRoute{
// 					{
// 						PoolId:       test.pool.GetId(),
// 						TokenInDenom: "bar",
// 					},
// 				},
// 				osmomath.NewInt(1000000000000000000),
// 				testCoin,
// 			)

// 			if test.expectPass {
// 				s.Require().NoError(swapInErr)
// 				s.Require().NoError(swapOutErr)
// 			} else {
// 				s.Require().Error(swapInErr)
// 				s.Require().Error(swapOutErr)
// 			}
// 		})
// 	}
// }
