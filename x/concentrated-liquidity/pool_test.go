package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity"
	cltypes "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (s *KeeperTestSuite) TestCalcOutAmtGivenIn() {
	currPrice := sdk.NewDec(5000)
	currSqrtPrice, err := currPrice.ApproxSqrt() // 70.710678118654752440
	s.Require().NoError(err)
	currTick := cl.PriceToTick(currPrice) // 85176
	lowerPrice := sdk.NewDec(4545)
	s.Require().NoError(err)
	lowerTick := cl.PriceToTick(lowerPrice) // 84222
	upperPrice := sdk.NewDec(5500)
	s.Require().NoError(err)
	upperTick := cl.PriceToTick(upperPrice) // 86129

	defaultAmt0 := sdk.NewInt(1000000)
	defaultAmt1 := sdk.NewInt(5000000000)

	swapFee := sdk.ZeroDec()

	tests := map[string]struct {
		positionAmount0  sdk.Int
		positionAmount1  sdk.Int
		addPositions     func(ctx sdk.Context, poolId uint64)
		tokenIn          sdk.Coin
		tokenOutDenom    string
		priceLimit       sdk.Dec
		expectedTokenIn  sdk.Coin
		expectedTokenOut sdk.Coin
		expectedTick     sdk.Int
		newLowerPrice    sdk.Dec
		newUpperPrice    sdk.Dec
		poolLiqAmount0   sdk.Int
		poolLiqAmount1   sdk.Int
	}{
		//  One price range
		//
		//          5000
		//  4545 -----|----- 5500
		"single position within one tick: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			tokenOutDenom: "eth",
			priceLimit:    sdk.NewDec(5004),
			// params
			// liquidity: 		 1517818840.967515822610790519
			// sqrtPriceNext:    70.738349405152439867 which is 5003.914076565430543175 https://www.wolframalpha.com/input?i=70.710678118654752440+%2B+42000000+%2F+1517818840.967515822610790519
			// sqrtPriceCurrent: 70.710678118654752440 which is 5000
			// expectedTokenIn:  42000000.0000 rounded up https://www.wolframalpha.com/input?i=1517818840.967515822610790519+*+%2870.738349405152439867+-+70.710678118654752440%29
			// expectedTokenOut: 8396.714105 rounded down https://www.wolframalpha.com/input?i=%281517818840.967515822610790519+*+%2870.738349405152439867+-+70.710678118654752440+%29%29+%2F+%2870.710678118654752440+*+70.738349405152439867%29
			// expectedTick: 	 85184.0 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C5003.914076565430543175%5D
			expectedTokenIn:  sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			expectedTokenOut: sdk.NewCoin("eth", sdk.NewInt(8396)),
			expectedTick:     sdk.NewInt(85184),
		},
		"single position within one tick: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("eth", sdk.NewInt(13370)),
			tokenOutDenom: "usdc",
			priceLimit:    sdk.NewDec(4993),
			// params
			// liquidity: 		 1517818840.967515822610790519
			// sqrtPriceNext:    70.666662070529219856 which is 4993.777128190373086350 https://www.wolframalpha.com/input?i=%28%281517818840.967515822610790519%29%29+%2F+%28%28%281517818840.967515822610790519%29+%2F+%2870.710678118654752440%29%29+%2B+%2813370%29%29
			// expectedTokenIn:  13369.9999 rounded up https://www.wolframalpha.com/input?i=%281517818840.967515822610790519+*+%2870.710678118654752440+-+70.666662070529219856+%29%29+%2F+%2870.666662070529219856+*+70.710678118654752440%29
			// expectedTokenOut: 66808387.149 rounded down https://www.wolframalpha.com/input?i=1517818840.967515822610790519+*+%2870.710678118654752440+-+70.666662070529219856%29
			// expectedTick: 	 85163.7 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C4993.777128190373086350%5D
			expectedTokenIn:  sdk.NewCoin("eth", sdk.NewInt(13370)),
			expectedTokenOut: sdk.NewCoin("usdc", sdk.NewInt(66808387)),
			expectedTick:     sdk.NewInt(85163),
		},
		//  Two equal price ranges
		//
		//          5000
		//  4545 -----|----- 5500
		//  4545 -----|----- 5500
		"two positions within one tick: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)

				// add second position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			tokenOutDenom: "eth",
			priceLimit:    sdk.NewDec(5002),
			// params
			// liquidity: 		 3035637681.935031645221581038
			// sqrtPriceNext:    70.724513761903596153 which is 5001.956846857691162236 https://www.wolframalpha.com/input?i=70.710678118654752440%2B%2842000000+%2F+3035637681.935031645221581038%29
			// sqrtPriceCurrent: 70.710678118654752440 which is 5000
			// expectedTokenIn:  41999999.999 rounded up https://www.wolframalpha.com/input?i=3035637681.935031645221581038+*+%2870.724513761903596153+-+70.710678118654752440%29
			// expectedTokenOut: 8398.3567 rounded down https://www.wolframalpha.com/input?i=%283035637681.935031645221581038+*+%2870.724513761903596153+-+70.710678118654752440+%29%29+%2F+%2870.710678118654752440+*+70.724513761903596153%29
			// expectedTick:     85180.1 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C5003.914076565430543175%5D
			expectedTokenIn:  sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			expectedTokenOut: sdk.NewCoin("eth", sdk.NewInt(8398)),
			expectedTick:     sdk.NewInt(85180),
			// two positions with same liquidity entered
			poolLiqAmount0: sdk.NewInt(1000000).MulRaw(2),
			poolLiqAmount1: sdk.NewInt(5000000000).MulRaw(2),
		},
		"two positions within one tick: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)

				// add second position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("eth", sdk.NewInt(13370)),
			tokenOutDenom: "usdc",
			priceLimit:    sdk.NewDec(4996),
			// params
			// liquidity: 		 3035637681.935031645221581038
			// sqrtPriceNext:    70.688663242671855280 which is 4996.887111035867053835 https://www.wolframalpha.com/input?i=%28%283035637681.935031645221581038%29%29+%2F+%28%28%283035637681.935031645221581038%29+%2F+%2870.710678118654752440%29%29+%2B+%2813370%29%29
			// sqrtPriceCurrent: 70.710678118654752440 which is 5000
			// expectedTokenIn:  13369.9999 rounded up https://www.wolframalpha.com/input?i=%283035637681.935031645221581038+*+%2870.710678118654752440+-+70.688663242671855280+%29%29+%2F+%2870.688663242671855280+*+70.710678118654752440%29
			// expectedTokenOut: 66829187.096 rounded down https://www.wolframalpha.com/input?i=3035637681.935031645221581038+*+%2870.710678118654752440+-+70.688663242671855280%29
			// expectedTick: 	 85170.00 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C4996.887111035867053835%5D
			expectedTokenIn:  sdk.NewCoin("eth", sdk.NewInt(13370)),
			expectedTokenOut: sdk.NewCoin("usdc", sdk.NewInt(66829187)),
			expectedTick:     sdk.NewInt(85169), // TODO: should be 85170, is 85169 due to log precision
			// two positions with same liquidity entered
			poolLiqAmount0: sdk.NewInt(1000000).MulRaw(2),
			poolLiqAmount1: sdk.NewInt(5000000000).MulRaw(2),
		},
		//  Consecutive price ranges
		//
		//          5000
		//  4545 -----|----- 5500
		//             5500 ----------- 6250
		//
		"two positions with consecutive price ranges": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
				// params
				// liquidity (1st):  1517818840.967515822610790519
				// sqrtPriceNext:    74.160724590951092256 which is 5499.813071854898049815 (this is calculated by finding the closest tick LTE the upper range of the first range) https://www.wolframalpha.com/input?i2d=true&i=Power%5B1.0001%2CDivide%5B86129%2C2%5D%5D
				// sqrtPriceCurrent: 70.710678118654752440 which is 5000
				// expectedTokenIn:  5236545537.865 rounded up https://www.wolframalpha.com/input?i=1517818840.967515822610790519+*+%2874.160724590951092256+-+70.710678118654752440%29
				// expectedTokenOut: 998587.023 rounded down https://www.wolframalpha.com/input?i=%281517818840.967515822610790519+*+%2874.161984870956629487+-+70.710678118654752440+%29%29+%2F+%2870.710678118654752440+*+74.161984870956629487%29
				// expectedTick:     86129.0 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C5499.813071854898049815%5D

				// create second position parameters
				newLowerPrice := sdk.NewDec(5500)
				s.Require().NoError(err)
				newLowerTick := cl.PriceToTick(newLowerPrice) // 86129
				newUpperPrice := sdk.NewDec(6250)
				s.Require().NoError(err)
				newUpperTick := cl.PriceToTick(newUpperPrice) // 87407

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[2], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64())
				s.Require().NoError(err)
				// params
				// liquidity (2nd):  1198107969.043944887658592210
				// sqrtPriceNext:    78.136538612066568296 which is 6105.473934424522538231 https://www.wolframalpha.com/input?i=74.160724590951092256+%2B+4763454462.135+%2F+1198107969.043944887658592210
				// sqrtPriceCurrent: 74.160724590951092256 which is 5499.813071854898049815
				// expectedTokenIn:  4763454462.135 rounded up https://www.wolframalpha.com/input?i=1198107969.043944887658592210+*+%2878.136538612066568296+-+74.160724590951092256%29
				// expectedTokenOut: 822041.769 rounded down https://www.wolframalpha.com/input?i=%281198107969.043944887658592210+*+%2878.136538612066568296+-+74.160724590951092256+%29%29+%2F+%2874.160724590951092256+*+78.136538612066568296%29
				// expectedTick:     87173.8 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C6105.473934424522538231%5D
			},
			tokenIn:       sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			tokenOutDenom: "eth",
			priceLimit:    sdk.NewDec(6106),
			// expectedTokenIn:  5236545537.865 + 4763454462.135 = 1000000000 usdc
			// expectedTokenOut: 998587.023 + 822041.769 = 1820628.792 round down = 1.820628 eth
			expectedTokenIn:  sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			expectedTokenOut: sdk.NewCoin("eth", sdk.NewInt(1820628)),
			expectedTick:     sdk.NewInt(87173),
			newLowerPrice:    sdk.NewDec(5500),
			newUpperPrice:    sdk.NewDec(6250),
		},
		//  Partially overlapping price ranges
		//
		//          5000
		//  4545 -----|----- 5500
		//        5001 ----------- 6250
		//
		"two positions with partially overlapping price ranges": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
				// params
				// liquidity (1st):  1517818840.967515822610790519
				// sqrtPriceNext:    74.160724590951092256 which is 5499.813071854898049815 (this is calculated by finding the closest tick LTE the upper range of the first range) https://www.wolframalpha.com/input?i2d=true&i=Power%5B1.0001%2CDivide%5B86129%2C2%5D%5D
				// sqrtPriceCurrent: 70.710678118654752440 which is 5000
				// expectedTokenIn:  5236545537.864897 rounded up https://www.wolframalpha.com/input?i=1517818840.967515822610790519+*+%2874.160724590951092256+-+70.710678118654752440%29
				// expectedTokenOut: 998934.824728 rounded down https://www.wolframalpha.com/input?i=%281517818840.967515822610790519+*+%2874.161984870956629487+-+70.710678118654752440+%29%29+%2F+%2870.710678118654752440+*+74.161984870956629487%29
				// expectedTick:     86129.0 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C5499.813071854898049815%5D

				// create second position parameters
				newLowerPrice := sdk.NewDec(5501)
				s.Require().NoError(err)
				newLowerTick := cl.PriceToTick(newLowerPrice) // 86131
				newUpperPrice := sdk.NewDec(6250)
				s.Require().NoError(err)
				newUpperTick := cl.PriceToTick(newUpperPrice) // 87407

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[2], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64())
				s.Require().NoError(err)
				// params
				// liquidity (2nd):  1200046517.432645168443803695
				// sqrtPriceNext:    78.137532176937376749 which is 6105.473934701923906716 https://www.wolframalpha.com/input?i=74.168140663410187419++%2B++4763454462.135+%2F+1200046517.432645168443803695
				// sqrtPriceCurrent: 74.168140663410187419 which is 5500.913089467399755950
				// expectedTokenIn:  4763454462.135 rounded up https://www.wolframalpha.com/input?i=1200046517.432645168443803695+*+%2878.137532176937376749+-+74.168140663410187419%29
				// expectedTokenOut: 821949.120898 rounded down https://www.wolframalpha.com/input?i=%281200046517.432645168443803695+*+%2878.137532176937376749+-+74.168140663410187419+%29%29+%2F+%2874.168140663410187419+*+78.137532176937376749%29
				// expectedTick:     87173.8 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C6105.473934424522538231%5D
			},
			tokenIn:       sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			tokenOutDenom: "eth",
			priceLimit:    sdk.NewDec(6106),
			// expectedTokenIn:  5236545537.865 + 4763454462.135 = 1000000000 usdc
			// expectedTokenOut: 998587.023 + 821949.120898 = 1820536.143 round down = 1.820536 eth
			expectedTokenIn:  sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			expectedTokenOut: sdk.NewCoin("eth", sdk.NewInt(1820536)),
			expectedTick:     sdk.NewInt(87173),
			newLowerPrice:    sdk.NewDec(5501),
			newUpperPrice:    sdk.NewDec(6250),
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.Setup()
			// create pool
			pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, "eth", "usdc", currSqrtPrice, currTick)
			s.Require().NoError(err)

			// add positions
			test.addPositions(s.Ctx, pool.Id)

			tokenIn, tokenOut, updatedTick, updatedLiquidity, _, err := s.App.ConcentratedLiquidityKeeper.CalcOutAmtGivenIn(
				s.Ctx,
				test.tokenIn, test.tokenOutDenom,
				swapFee, test.priceLimit, pool.Id)
			s.Require().NoError(err)

			s.Require().Equal(test.expectedTokenIn.String(), tokenIn.String())
			s.Require().Equal(test.expectedTokenOut.String(), tokenOut.String())
			s.Require().Equal(test.expectedTick.String(), updatedTick.String())

			if test.newLowerPrice.IsNil() && test.newUpperPrice.IsNil() {
				test.newLowerPrice = lowerPrice
				test.newUpperPrice = upperPrice
			}

			newLowerTick := cl.PriceToTick(test.newLowerPrice)
			newUpperTick := cl.PriceToTick(test.newUpperPrice)

			lowerSqrtPrice, err := cl.TickToSqrtPrice(newLowerTick)
			s.Require().NoError(err)
			upperSqrtPrice, err := cl.TickToSqrtPrice(newUpperTick)
			s.Require().NoError(err)

			if test.poolLiqAmount0.IsNil() && test.poolLiqAmount1.IsNil() {
				test.poolLiqAmount0 = defaultAmt0
				test.poolLiqAmount1 = defaultAmt1
			}

			expectedLiquidity := cl.GetLiquidityFromAmounts(currSqrtPrice, lowerSqrtPrice, upperSqrtPrice, test.poolLiqAmount0, test.poolLiqAmount1)
			s.Require().Equal(expectedLiquidity.String(), updatedLiquidity.String())
		})

	}
}

func (s *KeeperTestSuite) TestSwapOutAmtGivenIn() {
	currPrice := sdk.NewDec(5000)
	currSqrtPrice, err := currPrice.ApproxSqrt() // 70.710678118654752440
	s.Require().NoError(err)
	currTick := cl.PriceToTick(currPrice) // 85176
	lowerPrice := sdk.NewDec(4545)
	s.Require().NoError(err)
	lowerTick := cl.PriceToTick(lowerPrice) // 84222
	upperPrice := sdk.NewDec(5500)
	s.Require().NoError(err)
	upperTick := cl.PriceToTick(upperPrice) // 86129

	defaultAmt0 := sdk.NewInt(1000000)
	defaultAmt1 := sdk.NewInt(5000000000)

	swapFee := sdk.ZeroDec()

	tests := map[string]struct {
		positionAmount0  sdk.Int
		positionAmount1  sdk.Int
		addPositions     func(ctx sdk.Context, poolId uint64)
		tokenIn          sdk.Coin
		tokenOutDenom    string
		priceLimit       sdk.Dec
		expectedTokenOut sdk.Coin
		expectedTick     sdk.Int
		newLowerPrice    sdk.Dec
		newUpperPrice    sdk.Dec
		poolLiqAmount0   sdk.Int
		poolLiqAmount1   sdk.Int
	}{
		//  One price range
		//
		//          5000
		//  4545 -----|----- 5500
		"single position within one tick: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:          sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			tokenOutDenom:    "eth",
			priceLimit:       sdk.NewDec(5004),
			expectedTokenOut: sdk.NewCoin("eth", sdk.NewInt(8396)),
			expectedTick:     sdk.NewInt(85184),
		},
		"single position within one tick: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:          sdk.NewCoin("eth", sdk.NewInt(13370)),
			tokenOutDenom:    "usdc",
			priceLimit:       sdk.NewDec(4993),
			expectedTokenOut: sdk.NewCoin("usdc", sdk.NewInt(66808387)),
			expectedTick:     sdk.NewInt(85163),
		},
		//  Two equal price ranges
		//
		//          5000
		//  4545 -----|----- 5500
		//  4545 -----|----- 5500
		"two positions within one tick: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)

				// add second position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:          sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			tokenOutDenom:    "eth",
			priceLimit:       sdk.NewDec(5002),
			expectedTokenOut: sdk.NewCoin("eth", sdk.NewInt(8398)),
			expectedTick:     sdk.NewInt(85180),
			// two positions with same liquidity entered
			poolLiqAmount0: sdk.NewInt(1000000).MulRaw(2),
			poolLiqAmount1: sdk.NewInt(5000000000).MulRaw(2),
		},
		"two positions within one tick: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)

				// add second position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:          sdk.NewCoin("eth", sdk.NewInt(13370)),
			tokenOutDenom:    "usdc",
			priceLimit:       sdk.NewDec(4996),
			expectedTokenOut: sdk.NewCoin("usdc", sdk.NewInt(66829187)),
			expectedTick:     sdk.NewInt(85169),
			// two positions with same liquidity entered
			poolLiqAmount0: sdk.NewInt(1000000).MulRaw(2),
			poolLiqAmount1: sdk.NewInt(5000000000).MulRaw(2),
		},
		//  Consecutive price ranges
		//
		//          5000
		//  4545 -----|----- 5500
		//             5500 ----------- 6250
		//
		"two positions with consecutive price ranges": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)

				// create second position parameters
				newLowerPrice := sdk.NewDec(5500)
				s.Require().NoError(err)
				newLowerTick := cl.PriceToTick(newLowerPrice) // 84222
				newUpperPrice := sdk.NewDec(6250)
				s.Require().NoError(err)
				newUpperTick := cl.PriceToTick(newUpperPrice) // 87407

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[2], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:          sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			tokenOutDenom:    "eth",
			priceLimit:       sdk.NewDec(6106),
			expectedTokenOut: sdk.NewCoin("eth", sdk.NewInt(1820628)),
			expectedTick:     sdk.NewInt(87173),
			newLowerPrice:    sdk.NewDec(5500),
			newUpperPrice:    sdk.NewDec(6250),
		},
		//  Partially overlapping price ranges
		//
		//          5000
		//  4545 -----|----- 5500
		//        5001 ----------- 6250
		//
		"two positions with partially overlapping price ranges": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)

				// create second position parameters
				newLowerPrice := sdk.NewDec(5501)
				s.Require().NoError(err)
				newLowerTick := cl.PriceToTick(newLowerPrice) // 86131
				newUpperPrice := sdk.NewDec(6250)
				s.Require().NoError(err)
				newUpperTick := cl.PriceToTick(newUpperPrice) // 87407

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[2], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:          sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			tokenOutDenom:    "eth",
			priceLimit:       sdk.NewDec(6106),
			expectedTokenOut: sdk.NewCoin("eth", sdk.NewInt(1820536)),
			expectedTick:     sdk.NewInt(87173),
			newLowerPrice:    sdk.NewDec(5501),
			newUpperPrice:    sdk.NewDec(6250),
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.Setup()
			// create pool
			pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, "eth", "usdc", currSqrtPrice, currTick)
			s.Require().NoError(err)

			// add positions
			test.addPositions(s.Ctx, pool.Id)

			// execute internal swap function
			tokenOut, err := s.App.ConcentratedLiquidityKeeper.SwapOutAmtGivenIn(
				s.Ctx,
				test.tokenIn, test.tokenOutDenom,
				swapFee, test.priceLimit, pool.Id)
			s.Require().NoError(err)

			pool = s.App.ConcentratedLiquidityKeeper.GetPoolbyId(s.Ctx, pool.Id)
			s.Require().NoError(err)

			// check that we produced the same token out from the swap function that we expected
			s.Require().Equal(test.expectedTokenOut.String(), tokenOut.String())

			// check that the pool's current tick was updated correctly
			s.Require().Equal(test.expectedTick.String(), pool.CurrentTick.String())

			// the following is needed to get the expected liquidity to later compare to what the pool was updated to
			if test.newLowerPrice.IsNil() && test.newUpperPrice.IsNil() {
				test.newLowerPrice = lowerPrice
				test.newUpperPrice = upperPrice
			}

			newLowerTick := cl.PriceToTick(test.newLowerPrice)
			newUpperTick := cl.PriceToTick(test.newUpperPrice)

			lowerSqrtPrice, err := cl.TickToSqrtPrice(newLowerTick)
			s.Require().NoError(err)
			upperSqrtPrice, err := cl.TickToSqrtPrice(newUpperTick)
			s.Require().NoError(err)

			if test.poolLiqAmount0.IsNil() && test.poolLiqAmount1.IsNil() {
				test.poolLiqAmount0 = defaultAmt0
				test.poolLiqAmount1 = defaultAmt1
			}

			expectedLiquidity := cl.GetLiquidityFromAmounts(currSqrtPrice, lowerSqrtPrice, upperSqrtPrice, test.poolLiqAmount0, test.poolLiqAmount1)
			// check that the pools liquidity was updated correctly
			s.Require().Equal(expectedLiquidity.String(), pool.Liquidity.String())

			// TODO: need to figure out a good way to test that the currentSqrtPrice that the pool is set to makes sense
			// right now we calculate this value through iterations, so unsure how to do this here / if its needed
		})

	}
}

func (s *KeeperTestSuite) TestOrderInitialPoolDenoms() {
	denom0, denom1, err := cltypes.OrderInitialPoolDenoms("axel", "osmo")
	s.Require().NoError(err)
	s.Require().Equal(denom0, "axel")
	s.Require().Equal(denom1, "osmo")

	denom0, denom1, err = cltypes.OrderInitialPoolDenoms("usdc", "eth")
	s.Require().NoError(err)
	s.Require().Equal(denom0, "eth")
	s.Require().Equal(denom1, "usdc")

	denom0, denom1, err = cltypes.OrderInitialPoolDenoms("usdc", "usdc")
	s.Require().Error(err)

}

func (suite *KeeperTestSuite) TestPriceToTick() {
	testCases := []struct {
		name         string
		price        sdk.Dec
		tickExpected string
	}{
		{
			"happy path",
			sdk.NewDec(5000),
			"85176",
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			tick := cl.PriceToTick(tc.price)
			suite.Require().Equal(tc.tickExpected, tick.String())
		})
	}
}

// func (s *KeeperTestSuite) TestCalcInAmtGivenOut() {
// 	ctx := s.Ctx
// 	pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, "eth", "usdc", sdk.MustNewDecFromStr("70.710678"), sdk.NewInt(85176))
// 	s.Require().NoError(err)
// 	s.SetupPosition(pool.Id)

// 	// test asset a to b logic
// 	tokenOut := sdk.NewCoin("usdc", sdk.NewInt(4199999999))
// 	tokenInDenom := "eth"
// 	swapFee := sdk.NewDec(0)
// 	minPrice := sdk.NewDec(4500)
// 	maxPrice := sdk.NewDec(5500)

// 	amountIn, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// 	s.Require().NoError(err)
// 	s.Require().Equal(sdk.NewDec(805287), amountIn.Amount.ToDec())

// 	// test asset b to a logic
// 	tokenOut = sdk.NewCoin("eth", sdk.NewInt(133700))
// 	tokenInDenom = "usdc"
// 	swapFee = sdk.NewDec(0)

// 	amountIn, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// 	s.Require().NoError(err)
// 	s.Require().Equal(sdk.NewDec(666975610), amountIn.Amount.ToDec())

// 	// test asset a to b logic
// 	tokenOut = sdk.NewCoin("usdc", sdk.NewInt(4199999999))
// 	tokenInDenom = "eth"
// 	swapFee = sdk.NewDecWithPrec(2, 2)

// 	amountIn, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// 	s.Require().NoError(err)
// 	s.Require().Equal(sdk.NewDec(821722), amountIn.Amount.ToDec())
// }

// func (s *KeeperTestSuite) TestSwapInAmtGivenOut() {
// 	ctx := s.Ctx
// 	pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(ctx, 1, "eth", "usdc", sdk.MustNewDecFromStr("70.710678"), sdk.NewInt(85176))
// 	s.Require().NoError(err)
// 	fmt.Printf("%v pool liq pre \n", pool.Liquidity)
// 	lowerTick := int64(84222)
// 	upperTick := int64(86129)
// 	amount0Desired := sdk.NewInt(1)
// 	amount1Desired := sdk.NewInt(5000)

// 	s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, pool.Id, s.TestAccs[0], amount0Desired, amount1Desired, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick, upperTick)

// 	// test asset a to b logic
// 	tokenOut := sdk.NewCoin("usdc", sdk.NewInt(4199999999))
// 	tokenInDenom := "eth"
// 	swapFee := sdk.NewDec(0)
// 	minPrice := sdk.NewDec(4500)
// 	maxPrice := sdk.NewDec(5500)

// 	amountIn, err := s.App.ConcentratedLiquidityKeeper.SwapInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// 	s.Require().NoError(err)
// 	fmt.Printf("%v amountIn \n", amountIn)
// 	pool = s.App.ConcentratedLiquidityKeeper.GetPoolbyId(ctx, pool.Id)

// // test asset a to b logic
// tokenOut := sdk.NewCoin("usdc", sdk.NewInt(4199999999))
// tokenInDenom := "eth"
// swapFee := sdk.NewDec(0)
// minPrice := sdk.NewDec(4500)
// maxPrice := sdk.NewDec(5500)

// amountIn, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// s.Require().NoError(err)
// s.Require().Equal(sdk.NewDec(805287), amountIn.Amount.ToDec())

// // test asset b to a logic
// tokenOut = sdk.NewCoin("eth", sdk.NewInt(133700))
// tokenInDenom = "usdc"
// swapFee = sdk.NewDec(0)

// amountIn, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// s.Require().NoError(err)
// s.Require().Equal(sdk.NewDec(666975610), amountIn.Amount.ToDec())

// // test asset a to b logic
// tokenOut = sdk.NewCoin("usdc", sdk.NewInt(4199999999))
// tokenInDenom = "eth"
// swapFee = sdk.NewDecWithPrec(2, 2)

// amountIn, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// s.Require().NoError(err)
// s.Require().Equal(sdk.NewDec(821722), amountIn.Amount.ToDec())
// }
