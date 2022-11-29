package swapstrategy_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v13/app/apptesting"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/model"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/swapstrategy"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

type StrategyTestSuite struct {
	apptesting.KeeperTestHelper
}

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
			nextSqrtPrice:   sdk.MustNewDecFromStr("70.666662070529219856"), // 4993.777128190373086350
			liquidity:       sdk.MustNewDecFromStr("1517818840.967515822610790519"),
			amountRemaining: sdk.NewDec(13370),
			// sqrt price limit is less than sqrt price target so it does not affect the result
			// TODO: test case where it does affect.
			sqrtPriceLimit:        sdk.MustNewDecFromStr("70.666662070529219856").Sub(sdk.OneDec()),
			zeroForOne:            true,
			expectedSqrtPriceNext: "70.666662070529219856",
			expectedAmountIn:      "13369.999999903622360944",
			expectedAmountOut:     "66808387.149866264039333362",
		},
		"happy path: trade asset1 for asset0": {
			sqrtPCurrent:    sdk.MustNewDecFromStr("70.710678118654752440"), // 5000
			nextSqrtPrice:   sdk.MustNewDecFromStr("70.738349405152439867"), // 5003.91407656543054317
			liquidity:       sdk.MustNewDecFromStr("1517818840.967515822610790519"),
			amountRemaining: sdk.NewDec(42000000),
			// sqrt price limit is less than sqrt price target so it does not affect the result
			// TODO: test case where it does affect.
			sqrtPriceLimit:        sdk.MustNewDecFromStr("70.666662070529219856").Add(sdk.OneDec()),
			zeroForOne:            false,
			expectedSqrtPriceNext: "70.738349405152439867",
			expectedAmountIn:      "42000000.000000000650233591",
			expectedAmountOut:     "8396.714104746015980302",
		},
	}

	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			swapStrategy := swapstrategy.New(tc.zeroForOne, tc.sqrtPriceLimit, suite.App.GetKey(types.ModuleName))
			sqrtPriceNext, amountIn, amountOut := swapStrategy.ComputeSwapStep(tc.sqrtPCurrent, tc.nextSqrtPrice, tc.liquidity, tc.amountRemaining)
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

			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey)

			n, initd := swapStrategy.NextInitializedTick(ctx, 1, 78)
			suite.Require().Equal(int64(84), n)
			suite.Require().True(initd)
		})
		suite.Run("returns tick to right if at initialized tick", func() {
			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey)

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, -55)
			suite.Require().Equal(int64(-4), n)
			suite.Require().True(initd)
		})
		suite.Run("returns the tick directly to the right", func() {
			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey)

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 77)
			suite.Require().Equal(int64(78), n)
			suite.Require().True(initd)
		})
		suite.Run("returns the tick directly to the right", func() {
			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey)

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, -56)
			suite.Require().Equal(int64(-55), n)
			suite.Require().True(initd)
		})
		suite.Run("returns the next words initialized tick if on the right boundary", func() {
			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey)

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, -257)
			suite.Require().Equal(int64(-200), n)
			suite.Require().True(initd)
		})
		suite.Run("returns the next initialized tick from the next word", func() {
			swapStrategy := swapstrategy.New(false, sdk.ZeroDec(), clStoreKey)

			suite.App.ConcentratedLiquidityKeeper.SetTickInfo(suite.Ctx, 1, 340, model.TickInfo{})

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 328)
			suite.Require().Equal(int64(340), n)
			suite.Require().True(initd)
		})
	})

	suite.Run("lte=false", func() {
		suite.Run("returns tick directly to the left of input tick if not initialized", func() {
			swapStrategy := swapstrategy.New(true, sdk.ZeroDec(), clStoreKey)

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 79)
			suite.Require().Equal(int64(78), n)
			suite.Require().True(initd)
		})
		suite.Run("returns same tick if initialized", func() {
			swapStrategy := swapstrategy.New(true, sdk.ZeroDec(), clStoreKey)

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 78)
			suite.Require().Equal(int64(78), n)
			suite.Require().True(initd)
		})
		suite.Run("returns next initialized tick far away", func() {
			swapStrategy := swapstrategy.New(true, sdk.ZeroDec(), clStoreKey)

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 100)
			suite.Require().Equal(int64(84), n)
			suite.Require().True(initd)
		})
	})
}
