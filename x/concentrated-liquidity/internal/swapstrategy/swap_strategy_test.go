package swapstrategy_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/swapstrategy"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

type StrategyTestSuite struct {
	apptesting.KeeperTestHelper
}

var (
	two              = sdk.NewDec(2)
	three            = sdk.NewDec(2)
	four             = sdk.NewDec(4)
	five             = sdk.NewDec(5)
	defaultLiquidity = sdk.MustNewDecFromStr("3035764687.503020836176699298")

	defaultFee = sdk.MustNewDecFromStr("0.03")
)

func TestStrategyTestSuite(t *testing.T) {
	suite.Run(t, new(StrategyTestSuite))
}

func (suite *StrategyTestSuite) SetupTest() {
	suite.Setup()
}

// TODO: split up this test case to be separate for each strategy.
func (suite *StrategyTestSuite) TestComputeSwapState() {
	testCases := map[string]struct {
		sqrtPCurrent          sdk.Dec
		nextSqrtPrice         sdk.Dec
		liquidity             sdk.Dec
		amountRemaining       sdk.Dec
		sqrtPriceLimit        sdk.Dec
		zeroForOne            bool
		expectedSqrtPriceNext string
		expectedAmountIn      string
		expectedAmountOut     string
	}{
		"happy path: trade asset0 for asset1": {
			sqrtPCurrent:    sdk.MustNewDecFromStr("70.710678118654752440"), // 5000
			nextSqrtPrice:   sdk.MustNewDecFromStr("70.666662070529219856"), // 4993.7771281903730
			liquidity:       sdk.MustNewDecFromStr("1517882343.751510418088349649"),
			amountRemaining: sdk.NewDec(13370),
			// sqrt price limit is less than sqrt price target so it does not affect the result
			// TODO: test case where it does affect.
			sqrtPriceLimit:        sdk.MustNewDecFromStr("70.661163307718052314").Sub(sdk.OneDec()), // 4993
			zeroForOne:            true,
			expectedSqrtPriceNext: "70.666663910857144332",
			expectedAmountIn:      "13370.000000000000000000",
			expectedAmountOut:     "66808388.890199400470645012",
		},
		"happy path: trade asset1 for asset0": {
			sqrtPCurrent:    sdk.MustNewDecFromStr("70.710678118654752440"), // 5000
			nextSqrtPrice:   sdk.MustNewDecFromStr("70.738349405152439867"), // 5003.91407656543054317
			liquidity:       sdk.MustNewDecFromStr("1517882343.751510418088349649"),
			amountRemaining: sdk.NewDec(42000000),
			// sqrt price limit is less than sqrt price target so it does not affect the result
			// TODO: test case where it does affect.
			sqrtPriceLimit:        sdk.MustNewDecFromStr("70.738956735309575810").Add(sdk.OneDec()), // 5003
			zeroForOne:            false,
			expectedSqrtPriceNext: "70.738348247484497717",
			expectedAmountIn:      "42000000.000000000000000000",
			expectedAmountOut:     "8396.714242162444943332",
		},
	}

	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			swapStrategy := swapstrategy.New(tc.zeroForOne, tc.sqrtPriceLimit, suite.App.GetKey(types.ModuleName), sdk.ZeroDec())
			sqrtPriceNext, amountIn, amountOut, _ := swapStrategy.ComputeSwapStepOutGivenIn(tc.sqrtPCurrent, tc.nextSqrtPrice, tc.liquidity, tc.amountRemaining)
			suite.Require().Equal(tc.expectedSqrtPriceNext, sqrtPriceNext.String())
			suite.Require().Equal(tc.expectedAmountIn, amountIn.String())
			suite.Require().Equal(tc.expectedAmountOut, amountOut.String())
		})
	}
}

// TODO: split up this test case to be separate for each strategy.
func (suite *StrategyTestSuite) TestNextInitializedTick() {
	suite.SetupTest()
	ctx := suite.Ctx

	liquidityTicks := []int64{-200, -55, -4, 70, 78, 84, 139, 240, 535}
	for _, t := range liquidityTicks {
		suite.App.ConcentratedLiquidityKeeper.SetTickInfo(ctx, 1, t, model.TickInfo{})
	}

	clStoreKey := suite.App.GetKey(types.ModuleName)

	suite.Run("lte=true", func() {
		suite.Run("returns tick to right if at initialized tick", func() {

			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

			n, initd := swapStrategy.NextInitializedTick(ctx, 1, 78)
			suite.Require().Equal(sdk.NewInt(84), n)
			suite.Require().True(initd)
		})
		suite.Run("returns tick to right if at initialized tick", func() {
			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, -55)
			suite.Require().Equal(sdk.NewInt(-4), n)
			suite.Require().True(initd)
		})
		suite.Run("returns the tick directly to the right", func() {
			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 77)
			suite.Require().Equal(sdk.NewInt(78), n)
			suite.Require().True(initd)
		})
		suite.Run("returns the tick directly to the right", func() {
			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, -56)
			suite.Require().Equal(sdk.NewInt(-55), n)
			suite.Require().True(initd)
		})
		suite.Run("returns the next words initialized tick if on the right boundary", func() {
			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, -257)
			suite.Require().Equal(sdk.NewInt(-200), n)
			suite.Require().True(initd)
		})
		suite.Run("returns the next initialized tick from the next word", func() {
			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

			suite.App.ConcentratedLiquidityKeeper.SetTickInfo(suite.Ctx, 1, 340, model.TickInfo{})

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 328)
			suite.Require().Equal(sdk.NewInt(340), n)
			suite.Require().True(initd)
		})
	})

	suite.Run("lte=false", func() {
		suite.Run("returns tick directly to the left of input tick if not initialized", func() {
			swapStrategy := swapstrategy.New(true, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 79)
			suite.Require().Equal(sdk.NewInt(78), n)
			suite.Require().True(initd)
		})
		suite.Run("returns previous tick even though given is initialized", func() {
			swapStrategy := swapstrategy.New(true, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 78)
			suite.Require().Equal(sdk.NewInt(70), n)
			suite.Require().True(initd)
		})
		suite.Run("returns next initialized tick far away", func() {
			swapStrategy := swapstrategy.New(true, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

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
			expectedSqrtPriceNextInGivenOut: sdk.MustNewDecFromStr("70.688664163408836319"), // approx 4996.89

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
			sut := swapstrategy.New(tc.zeroForOne, sdk.ZeroDec(), suite.App.GetKey(types.ModuleName), sdk.ZeroDec())
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

			// These should be approximately equal. The difference stems from minor roundings and truncatios in the intermediary calculations.
			suite.Require().Equal(0, errToleranceSmall.CompareBigDec(
				osmomath.BigDecFromSDKDec(amountOutOutGivenIn),
				osmomath.BigDecFromSDKDec(amountOutInGivenOut)),
				fmt.Sprintf("amount out out given in: %s, amount out in given out: %s", amountOutOutGivenIn, amountOutInGivenOut))
		})
	}
}
