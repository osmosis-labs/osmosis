package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

var _ = suite.TestingSuite(nil)

func (s *KeeperTestSuite) TestCLPoolSimpleSwapExactAmountIn() {
	currPrice := sdk.NewDec(5000)
	currSqrtPrice, err := currPrice.ApproxSqrt() // 70.710678118654752440
	s.Require().NoError(err)
	currTick := math.PriceToTick(currPrice) // 85176
	lowerPrice := sdk.NewDec(4545)
	lowerTick := math.PriceToTick(lowerPrice) // 84222
	upperPrice := sdk.NewDec(5500)
	upperTick := math.PriceToTick(upperPrice) // 86129

	defaultAmt0 := sdk.NewInt(1000000)
	defaultAmt1 := sdk.NewInt(5000000000)

	swapFee := sdk.ZeroDec()

	type param struct {
		tokenIn           sdk.Coin
		tokenOutDenom     string
		tokenOutMinAmount sdk.Int
		expectedTokenOut  sdk.Int
	}

	tests := []struct {
		name      string
		param     param
		expectErr bool
	}{
		{
			name: "Proper swap foo > bar",
			param: param{
				tokenIn:           sdk.NewCoin("foo", sdk.NewInt(42000000)),
				tokenOutDenom:     "bar",
				tokenOutMinAmount: sdk.NewInt(1),
				expectedTokenOut:  sdk.NewInt(8396),
			},
		},
		{
			name: "Proper swap bar > foo",
			param: param{
				tokenIn:           sdk.NewCoin("bar", sdk.NewInt(13370)),
				tokenOutDenom:     "foo",
				tokenOutMinAmount: sdk.NewInt(1),
				expectedTokenOut:  sdk.NewInt(66808387),
			},
		},
		{
			name: "out is lesser than min amount",
			param: param{
				tokenIn:           sdk.NewCoin("foo", sdk.NewInt(42000000)),
				tokenOutDenom:     "bar",
				tokenOutMinAmount: sdk.NewInt(8397),
			},
			expectErr: true,
		},
		{
			name: "in and out denom are same",
			param: param{
				tokenIn:           sdk.NewCoin("bar", sdk.NewInt(13370)),
				tokenOutDenom:     "bar",
				tokenOutMinAmount: sdk.NewInt(1),
			},
			expectErr: true,
		},
		{
			name: "unknown in denom",
			param: param{
				tokenIn:           sdk.NewCoin("bara", sdk.NewInt(13370)),
				tokenOutDenom:     "bar",
				tokenOutMinAmount: sdk.NewInt(1),
			},
			expectErr: true,
		},
		{
			name: "unknown out denom",
			param: param{
				tokenIn:           sdk.NewCoin("bar", sdk.NewInt(13370)),
				tokenOutDenom:     "bara",
				tokenOutMinAmount: sdk.NewInt(1),
			},
			expectErr: true,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()

			// Create a foo - bar concentrated liquidity pool
			pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, "bar", "foo", currSqrtPrice, currTick)
			s.Require().NoError(err)

			// Check the test case to see if we are swapping asset0 for asset1 or vice versa
			asset0 := pool.GetToken0()
			zeroForOne := test.param.tokenIn.Denom == asset0

			// Fund the test account with foo and bar, then create a default position to the pool created earlier
			s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(10000000000000)), sdk.NewCoin("foo", sdk.NewInt(1000000000000))))
			_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, 1, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
			s.Require().NoError(err)

			if test.expectErr {
				pool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
				_, err = s.App.ConcentratedLiquidityKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], pool.(swaproutertypes.PoolI), test.param.tokenIn, test.param.tokenOutDenom, test.param.tokenOutMinAmount, swapFee)
				s.Require().Error(err)
			} else {
				spotPriceBefore := pool.GetCurrentSqrtPrice().Power(2)
				prevGasConsumed := s.Ctx.GasMeter().GasConsumed()

				// Execute the swap directed in the test case
				tokenOutAmount, err := s.App.ConcentratedLiquidityKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], pool.(swaproutertypes.PoolI), test.param.tokenIn, test.param.tokenOutDenom, test.param.tokenOutMinAmount, swapFee)
				s.Require().NoError(err)
				s.Require().Equal(test.param.expectedTokenOut.String(), tokenOutAmount.String())

				gasConsumedForSwap := s.Ctx.GasMeter().GasConsumed() - prevGasConsumed
				// TODO: make a CLGasFeeForSwap
				// Check that we consume enough gas that a CL pool swap warrants
				// We consume `types.GasFeeForSwap` directly, so the extra I/O operation mean we end up consuming more.
				s.Require().Greater(gasConsumedForSwap, uint64(types.BalancerGasFeeForSwap))

				// Assert events
				s.AssertEventEmitted(s.Ctx, types.TypeEvtTokenSwapped, 1)

				// Retrieve pool again post swap
				pool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
				s.Require().NoError(err)

				spotPriceAfter := pool.GetCurrentSqrtPrice().Power(2)

				// Ratio of the token out should be between the before spot price and after spot price.
				tradeAvgPrice := tradeAvgPrice = tokenOutAmount.ToDec().Quo(test.param.tokenIn.Amount.ToDec())
				if !zeroForOne {
					tradeAvgPrice = sdk.OneDec().Quo(tradeAvgPrice)
				}

				if zeroForOne {
					s.Require().True(tradeAvgPrice.LT(spotPriceBefore))
					s.Require().True(tradeAvgPrice.GT(spotPriceAfter))
				} else {
					s.Require().True(tradeAvgPrice.GT(spotPriceBefore))
					s.Require().True(tradeAvgPrice.LT(spotPriceAfter))
				}

			}
		})
	}
}
