package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	cltypes "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

var _ = suite.TestingSuite(nil)

func (s *KeeperTestSuite) TestSwapExactAmountIn() {
	type param struct {
		tokenIn           sdk.Coin
		tokenOutDenom     string
		tokenOutMinAmount sdk.Int
		expectedTokenOut  sdk.Int
	}

	tests := []struct {
		name        string
		param       param
		expectedErr error
	}{
		{
			name: "Proper swap usdc > eth",
			// liquidity: 		 1517818840.967515822610790519
			// sqrtPriceNext:    70.738349405152439867 which is 5003.914076565430543175 https://www.wolframalpha.com/input?i=70.710678118654752440+%2B+42000000+%2F+1517818840.967515822610790519
			// sqrtPriceCurrent: 70.710678118654752440 which is 5000
			// expectedTokenIn:  42000000.0000 rounded up https://www.wolframalpha.com/input?i=1517818840.967515822610790519+*+%2870.738349405152439867+-+70.710678118654752440%29
			// expectedTokenOut: 8396.714105 rounded down https://www.wolframalpha.com/input?i=%281517818840.967515822610790519+*+%2870.738349405152439867+-+70.710678118654752440+%29%29+%2F+%2870.710678118654752440+*+70.738349405152439867%29
			param: param{
				tokenIn:           sdk.NewCoin(USDC, sdk.NewInt(42000000)),
				tokenOutDenom:     ETH,
				tokenOutMinAmount: types.LowerPriceLimit.RoundInt(),
				expectedTokenOut:  sdk.NewInt(8396),
			},
		},
		{
			name: "Proper swap eth > usdc",
			// params
			// liquidity: 		 1517818840.967515822610790519
			// sqrtPriceNext:    70.666662070529219856 which is 4993.777128190373086350 https://www.wolframalpha.com/input?i=%28%281517818840.967515822610790519%29%29+%2F+%28%28%281517818840.967515822610790519%29+%2F+%2870.710678118654752440%29%29+%2B+%2813370%29%29
			// expectedTokenIn:  13369.9999 rounded up https://www.wolframalpha.com/input?i=%281517818840.967515822610790519+*+%2870.710678118654752440+-+70.666662070529219856+%29%29+%2F+%2870.666662070529219856+*+70.710678118654752440%29
			// expectedTokenOut: 66808387.149 rounded down https://www.wolframalpha.com/input?i=1517818840.967515822610790519+*+%2870.710678118654752440+-+70.666662070529219856%29
			// expectedTick: 	 85163.7 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C4993.777128190373086350%5D
			param: param{
				tokenIn:           sdk.NewCoin(ETH, sdk.NewInt(13370)),
				tokenOutDenom:     USDC,
				tokenOutMinAmount: types.LowerPriceLimit.RoundInt(),
				expectedTokenOut:  sdk.NewInt(66808387),
			},
		},
		{
			name: "out is lesser than min amount",
			param: param{
				tokenIn:           sdk.NewCoin(USDC, sdk.NewInt(42000000)),
				tokenOutDenom:     ETH,
				tokenOutMinAmount: sdk.NewInt(8397),
			},
			expectedErr: types.AmountLessThanMinError{TokenAmount: sdk.NewInt(8396), TokenMin: sdk.NewInt(8397)},
		},
		{
			name: "in and out denom are same",
			param: param{
				tokenIn:           sdk.NewCoin(ETH, sdk.NewInt(13370)),
				tokenOutDenom:     ETH,
				tokenOutMinAmount: types.LowerPriceLimit.RoundInt(),
			},
			expectedErr: types.DenomDuplicatedError{TokenInDenom: ETH, TokenOutDenom: ETH},
		},
		{
			name: "unknown in denom",
			param: param{
				tokenIn:           sdk.NewCoin("etha", sdk.NewInt(13370)),
				tokenOutDenom:     ETH,
				tokenOutMinAmount: types.LowerPriceLimit.RoundInt(),
			},
			expectedErr: types.TokenInDenomNotInPoolError{TokenInDenom: "etha"},
		},
		{
			name: "unknown out denom",
			param: param{
				tokenIn:           sdk.NewCoin(ETH, sdk.NewInt(13370)),
				tokenOutDenom:     "etha",
				tokenOutMinAmount: types.LowerPriceLimit.RoundInt(),
			},
			expectedErr: types.TokenOutDenomNotInPoolError{TokenOutDenom: "etha"},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()

			// Create a usdc - eth concentrated liquidity pool
			pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, "eth", "usdc", DefaultCurrSqrtPrice, DefaultCurrTick)
			s.Require().NoError(err)

			// Check the test case to see if we are swapping asset0 for asset1 or vice versa
			asset0 := pool.GetToken0()
			zeroForOne := test.param.tokenIn.Denom == asset0

			// Fund the test account with usdc and eth, then create a default position to the pool created earlier
			s.SetupPosition(1)

			// Note spot price and gas used prior to swap
			spotPriceBefore := pool.GetCurrentSqrtPrice().Power(2)
			prevGasConsumed := s.Ctx.GasMeter().GasConsumed()

			// Execute the swap directed in the test case
			tokenOutAmount, err := s.App.ConcentratedLiquidityKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], pool.(swaproutertypes.PoolI), test.param.tokenIn, test.param.tokenOutDenom, test.param.tokenOutMinAmount, DefaultZeroSwapFee)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, test.expectedErr.Error())
			} else {
				s.Require().NoError(err)
				s.Require().Equal(test.param.expectedTokenOut.String(), tokenOutAmount.String())

				gasConsumedForSwap := s.Ctx.GasMeter().GasConsumed() - prevGasConsumed

				// Check that we consume enough gas that a CL pool swap warrants
				// We consume `types.GasFeeForSwap` directly, so the extra I/O operation mean we end up consuming more.
				s.Require().Greater(gasConsumedForSwap, uint64(cltypes.ConcentratedGasFeeForSwap))

				// Assert events
				s.AssertEventEmitted(s.Ctx, swaproutertypes.TypeEvtTokenSwapped, 1)

				// Retrieve pool again post swap
				pool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
				s.Require().NoError(err)

				spotPriceAfter := pool.GetCurrentSqrtPrice().Power(2)

				// Ratio of the token out should be between the before spot price and after spot price.
				tradeAvgPrice := tokenOutAmount.ToDec().Quo(test.param.tokenIn.Amount.ToDec())

				if zeroForOne {
					s.Require().True(tradeAvgPrice.LT(spotPriceBefore))
					s.Require().True(tradeAvgPrice.GT(spotPriceAfter))
				} else {
					tradeAvgPrice = sdk.OneDec().Quo(tradeAvgPrice)
					s.Require().True(tradeAvgPrice.GT(spotPriceBefore))
					s.Require().True(tradeAvgPrice.LT(spotPriceAfter))
				}

			}
		})
	}
}

func (s *KeeperTestSuite) TestSwapExactAmountOut() {
	type param struct {
		tokenOut         sdk.Coin
		tokenInDenom     string
		tokenInMaxAmount sdk.Int
		expectedTokenIn  sdk.Int
	}

	tests := []struct {
		name        string
		param       param
		expectedErr error
	}{
		{
			name: "Proper swap eth > usdc",
			// liquidity: 		 1517818840.967515822610790519
			// sqrtPriceNext:    70.738349405152439867 which is 5003.914076565430543175 https://www.wolframalpha.com/input?i=70.710678118654752440+%2B+42000000+%2F+1517818840.967515822610790519
			// sqrtPriceCurrent: 70.710678118654752440 which is 5000
			// expectedTokenIn:  42000000.0000 rounded up https://www.wolframalpha.com/input?i=1517818840.967515822610790519+*+%2870.738349405152439867+-+70.710678118654752440%29
			// expectedTokenOut: 8396.714105 rounded down https://www.wolframalpha.com/input?i=%281517818840.967515822610790519+*+%2870.738349405152439867+-+70.710678118654752440+%29%29+%2F+%2870.710678118654752440+*+70.738349405152439867%29
			param: param{
				tokenOut:         sdk.NewCoin(USDC, sdk.NewInt(42000000)),
				tokenInDenom:     ETH,
				tokenInMaxAmount: types.UpperPriceLimit.RoundInt(),
				expectedTokenIn:  sdk.NewInt(8396),
			},
		},
		{
			name: "Proper swap usdc > eth",
			// params
			// liquidity: 		 1517818840.967515822610790519
			// sqrtPriceNext:    70.666662070529219856 which is 4993.777128190373086350 https://www.wolframalpha.com/input?i=%28%281517818840.967515822610790519%29%29+%2F+%28%28%281517818840.967515822610790519%29+%2F+%2870.710678118654752440%29%29+%2B+%2813370%29%29
			// expectedTokenIn:  13369.9999 rounded up https://www.wolframalpha.com/input?i=%281517818840.967515822610790519+*+%2870.710678118654752440+-+70.666662070529219856+%29%29+%2F+%2870.666662070529219856+*+70.710678118654752440%29
			// expectedTokenOut: 66808387.149 rounded down https://www.wolframalpha.com/input?i=1517818840.967515822610790519+*+%2870.710678118654752440+-+70.666662070529219856%29
			// expectedTick: 	 85163.7 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C4993.777128190373086350%5D
			param: param{
				tokenOut:         sdk.NewCoin(ETH, sdk.NewInt(13370)),
				tokenInDenom:     USDC,
				tokenInMaxAmount: types.UpperPriceLimit.RoundInt(),
				expectedTokenIn:  sdk.NewInt(66808387),
			},
		},
		{
			name: "out is more than max amount",
			param: param{
				tokenOut:         sdk.NewCoin(USDC, sdk.NewInt(42000000)),
				tokenInDenom:     ETH,
				tokenInMaxAmount: types.LowerPriceLimit.RoundInt(),
			},
			expectedErr: types.AmountGreaterThanMaxError{TokenAmount: sdk.NewInt(8396), TokenMax: types.LowerPriceLimit.RoundInt()},
		},
		{
			name: "in and out denom are same",
			param: param{
				tokenOut:         sdk.NewCoin(ETH, sdk.NewInt(13370)),
				tokenInDenom:     ETH,
				tokenInMaxAmount: types.UpperPriceLimit.RoundInt(),
			},
			expectedErr: types.DenomDuplicatedError{TokenInDenom: ETH, TokenOutDenom: ETH},
		},
		{
			name: "unknown out denom",
			param: param{
				tokenOut:         sdk.NewCoin("etha", sdk.NewInt(13370)),
				tokenInDenom:     ETH,
				tokenInMaxAmount: types.UpperPriceLimit.RoundInt(),
			},
			expectedErr: types.TokenOutDenomNotInPoolError{TokenOutDenom: "etha"},
		},
		{
			name: "unknown in denom",
			param: param{
				tokenOut:         sdk.NewCoin(ETH, sdk.NewInt(13370)),
				tokenInDenom:     "etha",
				tokenInMaxAmount: types.UpperPriceLimit.RoundInt(),
			},
			expectedErr: types.TokenInDenomNotInPoolError{TokenInDenom: "etha"},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()

			// Create a usdc - eth concentrated liquidity pool
			pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, "eth", "usdc", DefaultCurrSqrtPrice, DefaultCurrTick)
			s.Require().NoError(err)

			// Check the test case to see if we are swapping asset0 for asset1 or vice versa
			asset0 := pool.GetToken0()
			zeroForOne := test.param.tokenOut.Denom == asset0

			// Fund the test account with usdc and eth, then create a default position to the pool created earlier
			s.SetupPosition(1)

			// Note spot price and gas used prior to swap
			spotPriceBefore := pool.GetCurrentSqrtPrice().Power(2)
			prevGasConsumed := s.Ctx.GasMeter().GasConsumed()

			// Execute the swap directed in the test case
			tokenIn, err := s.App.ConcentratedLiquidityKeeper.SwapExactAmountOut(s.Ctx, s.TestAccs[0], pool.(swaproutertypes.PoolI), test.param.tokenInDenom, test.param.tokenInMaxAmount, test.param.tokenOut, DefaultZeroSwapFee)

			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, test.expectedErr.Error())
			} else {
				s.Require().NoError(err)
				s.Require().Equal(test.param.expectedTokenIn.String(), tokenIn.String())

				gasConsumedForSwap := s.Ctx.GasMeter().GasConsumed() - prevGasConsumed
				// Check that we consume enough gas that a CL pool swap warrants
				// We consume `types.GasFeeForSwap` directly, so the extra I/O operation mean we end up consuming more.
				s.Require().Greater(gasConsumedForSwap, uint64(cltypes.ConcentratedGasFeeForSwap))

				// Assert events
				s.AssertEventEmitted(s.Ctx, swaproutertypes.TypeEvtTokenSwapped, 1)

				// Retrieve pool again post swap
				pool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
				s.Require().NoError(err)

				spotPriceAfter := pool.GetCurrentSqrtPrice().Power(2)

				// Ratio of the token out should be between the before spot price and after spot price.
				tradeAvgPrice := tokenIn.ToDec().Quo(test.param.tokenOut.Amount.ToDec())

				if zeroForOne {
					s.Require().True(tradeAvgPrice.LT(spotPriceBefore))
					s.Require().True(tradeAvgPrice.GT(spotPriceAfter))
				} else {
					tradeAvgPrice = sdk.OneDec().Quo(tradeAvgPrice)
					s.Require().True(tradeAvgPrice.GT(spotPriceBefore))
					s.Require().True(tradeAvgPrice.LT(spotPriceAfter))
				}

			}
		})
	}
}
