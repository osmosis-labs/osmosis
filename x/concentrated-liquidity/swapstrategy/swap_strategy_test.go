package swapstrategy_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/swapstrategy"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

type StrategyTestSuite struct {
	apptesting.KeeperTestHelper
}

var (
	two                   = sdk.NewDec(2)
	three                 = sdk.NewDec(2)
	four                  = sdk.NewDec(4)
	five                  = sdk.NewDec(5)
	defaultSqrtPriceLower = sdk.MustNewDecFromStr("70.688664163408836321") // approx 4996.89
	defaultSqrtPriceUpper = sdk.MustNewDecFromStr("70.710678118654752440") // 5000
	defaultAmountOne      = sdk.MustNewDecFromStr("66829187.967824033199646915")
	defaultAmountZero     = sdk.MustNewDecFromStr("13369.999999999998920002")
	defaultLiquidity      = sdk.MustNewDecFromStr("3035764687.503020836176699298")
	defaultFee            = sdk.MustNewDecFromStr("0.03")
	defaultTickSpacing    = uint64(100)
)

func TestStrategyTestSuite(t *testing.T) {
	suite.Run(t, new(StrategyTestSuite))
}

func (suite *StrategyTestSuite) SetupTest() {
	suite.Setup()
}

// TODO: split up this test case to be separate for each strategy.
func (suite *StrategyTestSuite) TestNextInitializedTick() {
	suite.SetupTest()
	ctx := suite.Ctx

	liquidityTicks := []int64{-200, -55, -4, 70, 78, 84, 139, 240, 535}
	for _, t := range liquidityTicks {
		suite.App.ConcentratedLiquidityKeeper.SetTickInfo(ctx, 1, t, model.TickInfo{})
	}

	_, err := suite.App.ConcentratedLiquidityKeeper.GetAllInitializedTicksForPool(ctx, 1)
	suite.Require().NoError(err)

	clStoreKey := suite.App.GetKey(types.ModuleName)

	suite.Run("lte=true", func() {
		suite.Run("returns tick to right if at initialized tick", func() {
			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec(), defaultTickSpacing)

			n, initd := swapStrategy.NextInitializedTick(ctx, 1, 78)
			suite.Require().Equal(sdk.NewInt(84), n)
			suite.Require().True(initd)
		})
		suite.Run("returns tick to right if at initialized tick", func() {
			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec(), defaultTickSpacing)

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, -55)
			suite.Require().Equal(sdk.NewInt(-4), n)
			suite.Require().True(initd)
		})
		suite.Run("returns the tick directly to the right", func() {
			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec(), defaultTickSpacing)

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 77)
			suite.Require().Equal(sdk.NewInt(78), n)
			suite.Require().True(initd)
		})
		suite.Run("returns the tick directly to the right", func() {
			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec(), defaultTickSpacing)

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, -56)
			suite.Require().Equal(sdk.NewInt(-55), n)
			suite.Require().True(initd)
		})
		suite.Run("returns the next words initialized tick if on the right boundary", func() {
			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec(), defaultTickSpacing)

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, -257)
			suite.Require().Equal(sdk.NewInt(-200), n)
			suite.Require().True(initd)
		})
		suite.Run("returns the next initialized tick from the next word", func() {
			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec(), defaultTickSpacing)

			suite.App.ConcentratedLiquidityKeeper.SetTickInfo(suite.Ctx, 1, 340, model.TickInfo{})

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 328)
			suite.Require().Equal(sdk.NewInt(340), n)
			suite.Require().True(initd)
		})
	})

	suite.Run("lte=false", func() {
		suite.Run("returns tick directly to the left of input tick if not initialized", func() {
			swapStrategy := swapstrategy.New(true, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec(), defaultTickSpacing)

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 79)
			suite.Require().Equal(sdk.NewInt(78), n)
			suite.Require().True(initd)
		})
		suite.Run("returns previous tick even though given is initialized", func() {
			swapStrategy := swapstrategy.New(true, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec(), defaultTickSpacing)

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 78)
			suite.Require().Equal(sdk.NewInt(70), n)
			suite.Require().True(initd)
		})
		suite.Run("returns next initialized tick far away", func() {
			swapStrategy := swapstrategy.New(true, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec(), defaultTickSpacing)

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 100)
			suite.Require().Equal(sdk.NewInt(84), n)
			suite.Require().True(initd)
		})
	})
}

// TestComputeSwapState_Inverse validates that out given in and in given out compute swap steps
// for one strategy produce the same results as the other given inverse inputs.
// That is, if we swap in A of token0 and expect to get out B of token 1,
// we should be able to get A of token 0 in when swapping out for B of token 1.
func (suite *StrategyTestSuite) TestComputeSwapState_Inverse() {
	var (
		errToleranceOne = osmomath.ErrTolerance{
			AdditiveTolerance: sdk.OneDec(),
			RoundingDir:       osmomath.RoundUp,
		}

		errToleranceSmall = osmomath.ErrTolerance{
			AdditiveTolerance: sdk.NewDecFromIntWithPrec(sdk.OneInt(), 5),
		}
	)

	testCases := map[string]struct {
		sqrtPriceCurrent sdk.Dec
		sqrtPriceTarget  sdk.Dec
		liquidity        sdk.Dec
		amountIn         sdk.Dec
		amountOut        sdk.Dec
		zeroForOne       bool
		swapFee          sdk.Dec

		expectedSqrtPriceNextOutGivenIn sdk.Dec
		expectedSqrtPriceNextInGivenOut sdk.Dec
		expectedAmountIn                sdk.Dec
		expectedAmountOut               sdk.Dec
	}{
		"1: one_for_zero__not_equal_target__no_fee": {
			sqrtPriceCurrent: sdk.MustNewDecFromStr("70.710678118654752440"), // 5000
			sqrtPriceTarget:  sdk.MustNewDecFromStr("70.724818840347693039"), // 5002
			liquidity:        sdk.MustNewDecFromStr("3035764687.503020836176699298"),
			amountIn:         sdk.NewDec(42000000),
			amountOut:        sdk.NewDec(8398),
			zeroForOne:       false,
			swapFee:          sdk.ZeroDec(),

			// from token_in:   sqrt_next = sqrt_cur + token_in / liq2 = 70.72451318306962507883763621
			expectedSqrtPriceNextOutGivenIn: sdk.MustNewDecFromStr("70.724513183069625078"), // approx 5001.96
			// from token_out:  sqrt_next = liq2 * sqrt_cur / (liq2 - token_out * sqrt_cur)
			expectedSqrtPriceNextInGivenOut: sdk.MustNewDecFromStr("70.724513183069625078"), // approx 5001.96

			expectedAmountIn:  sdk.NewDec(42000000),
			expectedAmountOut: sdk.NewDec(8398),
		},
		"2: zero_for_one__not_equal_target__no_fee": {
			sqrtPriceCurrent: sdk.MustNewDecFromStr("70.710678118654752440"), // 5000
			sqrtPriceTarget:  sdk.MustNewDecFromStr("70.682388188289167342"), // 4996
			liquidity:        sdk.MustNewDecFromStr("3035764687.503020836176699298"),
			amountIn:         sdk.NewDec(13370),
			amountOut:        sdk.NewDec(66829187),
			zeroForOne:       true,
			swapFee:          sdk.ZeroDec(),

			// from amount in: sqrt_next = liq2 * sqrt_cur / (liq2 + token_in * sqrt_cur) quo round up = 70.68866416340883631930670240
			expectedSqrtPriceNextOutGivenIn: sdk.MustNewDecFromStr("70.688664163408836320"), // approx 4996.89

			// from amount out: sqrt_next = sqrt_cur - token_out / liq2 quo round down
			expectedSqrtPriceNextInGivenOut: sdk.MustNewDecFromStr("70.688664163408836320"), // approx 4996.89

			expectedAmountIn:  sdk.NewDec(13370),
			expectedAmountOut: sdk.NewDec(66829187),
		},
		"3: one_for_zero__equal_target__no_fee": {
			sqrtPriceCurrent: sdk.MustNewDecFromStr("70.710678118654752440"), // 5000
			sqrtPriceTarget:  sdk.MustNewDecFromStr("70.724513183069625078"), // approx 5001.96
			liquidity:        sdk.MustNewDecFromStr("3035764687.503020836176699298"),
			amountIn:         sdk.NewDec(42000000),
			amountOut:        sdk.NewDec(8398),
			swapFee:          sdk.ZeroDec(),

			zeroForOne: false,
			// from token_in:   sqrt_next = sqrt_cur + token_in / liq2 = 70.72451318306962507883763621
			expectedSqrtPriceNextOutGivenIn: sdk.MustNewDecFromStr("70.724513183069625078"), // approx 5001.96
			// from token_out:  sqrt_next = liq2 * sqrt_cur / (liq2 - token_out * sqrt_cur)
			expectedSqrtPriceNextInGivenOut: sdk.MustNewDecFromStr("70.724513183069625078"), // approx 5001.96

			expectedAmountIn:  sdk.NewDec(42000000),
			expectedAmountOut: sdk.NewDec(8398),
		},
		"4: zero_for_one__equal_target__no_fee": {
			sqrtPriceCurrent: sdk.MustNewDecFromStr("70.710678118654752440"), // 5000
			sqrtPriceTarget:  sdk.MustNewDecFromStr("70.688664163408836320"), // approx 4996.89
			liquidity:        sdk.MustNewDecFromStr("3035764687.503020836176699298"),
			amountIn:         sdk.NewDec(13370),
			amountOut:        sdk.NewDec(66829187),
			zeroForOne:       true,
			swapFee:          sdk.ZeroDec(),

			// from amount in: sqrt_next = liq2 * sqrt_cur / (liq2 + token_in * sqrt_cur) = 70.68866416340883631930670240
			expectedSqrtPriceNextOutGivenIn: sdk.MustNewDecFromStr("70.688664163408836320"), // approx 4996.89

			// from amount out: sqrt_next = sqrt_cur - token_out / liq2
			expectedSqrtPriceNextInGivenOut: sdk.MustNewDecFromStr("70.688664163408836320"), // approx 4996.89

			expectedAmountIn:  sdk.NewDec(13370),
			expectedAmountOut: sdk.NewDec(66829187),
		},
	}

	for name, tc := range testCases {
		tc := tc
		suite.Run(name, func() {
			sut := swapstrategy.New(tc.zeroForOne, sdk.ZeroDec(), suite.App.GetKey(types.ModuleName), sdk.ZeroDec(), defaultTickSpacing)
			sqrtPriceNextOutGivenIn, amountInOutGivenIn, amountOutOutGivenIn, _ := sut.ComputeSwapStepOutGivenIn(tc.sqrtPriceCurrent, tc.sqrtPriceTarget, tc.liquidity, tc.amountIn)
			suite.Require().Equal(tc.expectedSqrtPriceNextOutGivenIn.String(), sqrtPriceNextOutGivenIn.String())

			sqrtPriceNextInGivenOut, amountOutInGivenOut, amountInInGivenOut, _ := sut.ComputeSwapStepInGivenOut(tc.sqrtPriceCurrent, tc.sqrtPriceTarget, tc.liquidity, amountOutOutGivenIn)

			suite.Require().Equal(tc.expectedSqrtPriceNextInGivenOut.String(), sqrtPriceNextInGivenOut.String())

			// Tolerance of 1 with rounding up because we round up for in given out.
			// This is to ensure that inflow into the pool is rounded in favor of the pool.
			suite.Require().Equal(0, errToleranceOne.CompareBigDec(
				osmomath.BigDecFromSDKDec(amountInOutGivenIn),
				osmomath.BigDecFromSDKDec(amountInInGivenOut)),
				fmt.Sprintf("amount in out given in: %s, amount in in given out: %s", amountInOutGivenIn, amountInInGivenOut))

			// These should be approximately equal. The difference stems from minor roundings and truncations in the intermediary calculations.
			suite.Require().Equal(0, errToleranceSmall.CompareBigDec(
				osmomath.BigDecFromSDKDec(amountOutOutGivenIn),
				osmomath.BigDecFromSDKDec(amountOutInGivenOut)),
				fmt.Sprintf("amount out out given in: %s, amount out in given out: %s", amountOutOutGivenIn, amountOutInGivenOut))
		})
	}
}

func (suite *StrategyTestSuite) TestGetPriceLimit() {
	tests := map[string]struct {
		zeroForOne bool
		expected   sdk.Dec
	}{
		"zero for one -> min": {
			zeroForOne: true,
			expected:   types.MinSpotPrice,
		},
		"one for zero -> max": {
			zeroForOne: false,
			expected:   types.MaxSpotPrice,
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			priceLimit := swapstrategy.GetPriceLimit(tc.zeroForOne)
			suite.Require().Equal(tc.expected, priceLimit)
		})
	}
}
