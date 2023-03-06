package swapstrategy_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/swapstrategy"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

type StrategyTestSuite struct {
	apptesting.KeeperTestHelper
}

var (
	two   = sdk.NewDec(2)
	three = sdk.NewDec(2)
	four  = sdk.NewDec(4)
	five  = sdk.NewDec(5)
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
			swapStrategy := swapstrategy.New(tc.zeroForOne, true, tc.sqrtPriceLimit, suite.App.GetKey(types.ModuleName), sdk.ZeroDec())
			sqrtPriceNext, amountIn, amountOut, _ := swapStrategy.ComputeSwapStep(tc.sqrtPCurrent, tc.nextSqrtPrice, tc.liquidity, tc.amountRemaining)
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

			swapStrategy := swapstrategy.New(false, false, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

			n, initd := swapStrategy.NextInitializedTick(ctx, 1, 78)
			suite.Require().Equal(sdk.NewInt(84), n)
			suite.Require().True(initd)
		})
		suite.Run("returns tick to right if at initialized tick", func() {
			swapStrategy := swapstrategy.New(false, true, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, -55)
			suite.Require().Equal(sdk.NewInt(-4), n)
			suite.Require().True(initd)
		})
		suite.Run("returns the tick directly to the right", func() {
			swapStrategy := swapstrategy.New(false, true, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 77)
			suite.Require().Equal(sdk.NewInt(78), n)
			suite.Require().True(initd)
		})
		suite.Run("returns the tick directly to the right", func() {
			swapStrategy := swapstrategy.New(false, true, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, -56)
			suite.Require().Equal(sdk.NewInt(-55), n)
			suite.Require().True(initd)
		})
		suite.Run("returns the next words initialized tick if on the right boundary", func() {
			swapStrategy := swapstrategy.New(false, true, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, -257)
			suite.Require().Equal(sdk.NewInt(-200), n)
			suite.Require().True(initd)
		})
		suite.Run("returns the next initialized tick from the next word", func() {
			swapStrategy := swapstrategy.New(false, true, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

			suite.App.ConcentratedLiquidityKeeper.SetTickInfo(suite.Ctx, 1, 340, model.TickInfo{})

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 328)
			suite.Require().Equal(sdk.NewInt(340), n)
			suite.Require().True(initd)
		})
	})

	suite.Run("lte=false", func() {
		suite.Run("returns tick directly to the left of input tick if not initialized", func() {
			swapStrategy := swapstrategy.New(true, true, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 79)
			suite.Require().Equal(sdk.NewInt(78), n)
			suite.Require().True(initd)
		})
		suite.Run("returns previous tick even though given is initialized", func() {
			swapStrategy := swapstrategy.New(true, true, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 78)
			suite.Require().Equal(sdk.NewInt(70), n)
			suite.Require().True(initd)
		})
		suite.Run("returns next initialized tick far away", func() {
			swapStrategy := swapstrategy.New(true, true, sdk.ZeroDec(), clStoreKey, sdk.ZeroDec())

			n, initd := swapStrategy.NextInitializedTick(suite.Ctx, 1, 100)
			suite.Require().Equal(sdk.NewInt(84), n)
			suite.Require().True(initd)
		})
	})
}
