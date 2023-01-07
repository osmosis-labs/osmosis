package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	cltypes "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

var _ = suite.TestingSuite(nil)

func (s *KeeperTestSuite) TestCalcAndSwapOutAmtGivenIn() {
	tests := map[string]struct {
		positionAmount0   sdk.Int
		positionAmount1   sdk.Int
		addPositions      func(ctx sdk.Context, poolId uint64)
		tokenIn           sdk.Coin
		tokenOutDenom     string
		priceLimit        sdk.Dec
		expectedTokenIn   sdk.Coin
		expectedTokenOut  sdk.Coin
		expectedTick      sdk.Int
		expectedSqrtPrice sdk.Dec
		newLowerPrice     sdk.Dec
		newUpperPrice     sdk.Dec
		poolLiqAmount0    sdk.Int
		poolLiqAmount1    sdk.Int
		expectErr         bool
	}{
		//  One price range
		//
		//          5000
		//  4545 -----|----- 5500
		"single position within one tick: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			tokenOutDenom: "eth",
			priceLimit:    sdk.NewDec(5004),
			// params
			// liquidity: 		 1517818840.967415409394235163
			// sqrtPriceNext:    70.738349405152441697 which is 5003.914076565430802105 https://www.wolframalpha.com/input?i=70.710678118654752440+%2B+42000000+%2F+1517818840.967415409394235163
			// sqrtPriceCurrent: 70.710678118654752440 which is 5000
			// expectedTokenIn:  42000000.0000 rounded up https://www.wolframalpha.com/input?i=1517818840.967415409394235163+*+%2870.738349405152441697+-+70.710678118654752440%29
			// expectedTokenOut: 8396.71410474607902463597 rounded down https://www.wolframalpha.com/input?i=%281517818840.967415409394235163+*+%2870.738349405152441697+-+70.710678118654752440+%29%29+%2F+%2870.710678118654752440+*+70.738349405152441697%29
			// expectedTick: 	 85184.0 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C5003.914076565430802105%5D
			expectedTokenIn:   sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			expectedTokenOut:  sdk.NewCoin("eth", sdk.NewInt(8396)),
			expectedTick:      sdk.NewInt(85184),
			expectedSqrtPrice: sdk.MustNewDecFromStr("70.738349405152441697"),
		},
		"single position within one tick: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("eth", sdk.NewInt(13370)),
			tokenOutDenom: "usdc",
			priceLimit:    sdk.NewDec(4993),
			// params
			// liquidity: 		 1517818840.967415409394235163
			// sqrtPriceNext:    70.666662070528898353 which is 4993.7771281903276472619 https://www.wolframalpha.com/input?i=%28%281517818840.967415409394235163%29%29+%2F+%28%28%281517818840.967415409394235163%29+%2F+%2870.710678118654752440%29%29+%2B+%2813370%29%29
			// sqrtPriceCurrent: 70.710678118654752440 which is 5000
			// expectedTokenIn:  13370.0000 rounded up https://www.wolframalpha.com/input?i=%281517818840.967415409394235163+*+%2870.710678118654752440+-+70.666662070528898353+%29%29+%2F+%2870.666662070528898353+*+70.710678118654752440%29
			// expectedTokenOut: 66808387.150 rounded down https://www.wolframalpha.com/input?i=1517818840.967415409394235163+*+%2870.710678118654752440+-+70.666662070528898353%29
			// expectedTick: 	 85163.7 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C4993.7771281903276472619%5D
			expectedTokenIn:   sdk.NewCoin("eth", sdk.NewInt(13370)),
			expectedTokenOut:  sdk.NewCoin("usdc", sdk.NewInt(66808387)),
			expectedTick:      sdk.NewInt(85163),
			expectedSqrtPrice: sdk.MustNewDecFromStr("70.666662070528898354"), // ends in 4 instead of 3 because we round up on token0 > token1 swaps
		},
		//  Two equal price ranges
		//
		//          5000
		//  4545 -----|----- 5500
		//  4545 -----|----- 5500
		"two positions within one tick: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)

				// add second position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			tokenOutDenom: "eth",
			priceLimit:    sdk.NewDec(5002),
			// params
			// liquidity: 		 3035637681.934830818788470326
			// sqrtPriceNext:    70.724513761903597068 which is 5001.956846857691162236 https://www.wolframalpha.com/input?i=70.710678118654752440%2B%2842000000+%2F+3035637681.934830818788470326+%29
			// sqrtPriceCurrent: 70.710678118654752440 which is 5000
			// expectedTokenIn:  41999999.999 rounded up https://www.wolframalpha.com/input?i=3035637681.934830818788470326+*+%2870.724513761903596153+-+70.710678118654752440%29
			// expectedTokenOut: 8398.3567 rounded down https://www.wolframalpha.com/input?i=%283035637681.934830818788470326+*+%2870.724513761903596153+-+70.710678118654752440+%29%29+%2F+%2870.710678118654752440+*+70.724513761903596153%29
			// expectedTick:     85180.1 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C5003.914076565430543175%5D
			expectedTokenIn:   sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			expectedTokenOut:  sdk.NewCoin("eth", sdk.NewInt(8398)),
			expectedTick:      sdk.NewInt(85180),
			expectedSqrtPrice: sdk.MustNewDecFromStr("70.724513761903597069"),
			// two positions with same liquidity entered
			poolLiqAmount0: sdk.NewInt(1000000).MulRaw(2),
			poolLiqAmount1: sdk.NewInt(5000000000).MulRaw(2),
		},
		"two positions within one tick: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)

				// add second position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("eth", sdk.NewInt(13370)),
			tokenOutDenom: "usdc",
			priceLimit:    sdk.NewDec(4996),
			// params
			// liquidity: 		 3035637681.934830818788470326
			// sqrtPriceNext:    70.688663242671673182 which is 4996.8871110358413093253114 https://www.wolframalpha.com/input?i=%28%283035637681.934830818788470326%29%29+%2F+%28%28%283035637681.934830818788470326%29+%2F+%2870.710678118654752440%29%29+%2B+%2813370%29%29
			// sqrtPriceCurrent: 70.710678118654752440 which is 5000
			// expectedTokenIn:  13370.000 rounded up https://www.wolframalpha.com/input?i=%283035637681.934830818788470326+*+%2870.710678118654752440+-+70.688663242671673182+%29%29+%2F+%2870.688663242671673182+*+70.710678118654752440%29
			// expectedTokenOut: 66829187.0973574985351 rounded down https://www.wolframalpha.com/input?i=3035637681.934830818788470326+*+%2870.710678118654752440+-+70.688663242671673182%29
			// expectedTick: 	 85169.96 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C4996.8871110358413093253114%5D
			expectedTokenIn:   sdk.NewCoin("eth", sdk.NewInt(13370)),
			expectedTokenOut:  sdk.NewCoin("usdc", sdk.NewInt(66829187)),
			expectedTick:      sdk.NewInt(85169),
			expectedSqrtPrice: sdk.MustNewDecFromStr("70.688663242671673183"), // ends in 3 instead of 2 because we round up on token0 > token1 swaps
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
		"two positions with consecutive price ranges: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
				// params
				// liquidity (1st):  1517818840.967415409394235163
				// sqrtPriceNext:    74.160724590950847045 which is 5499.813071854861679877 https://www.wolframalpha.com/input?i2d=true&i=Power%5B1.0001%2CDivide%5B86129%2C2%5D%5D
				// sqrtPriceCurrent: 70.710678118654752440 which is 5000
				// expectedTokenIn:  5236545537.864178570316821 rounded up https://www.wolframalpha.com/input?i=1517818840.967415409394235163+*+%2874.160724590950847045+-+70.710678118654752440%29
				// expectedTokenOut: 998587.023047 rounded down https://www.wolframalpha.com/input?i=%281517818840.967415409394235163+*+%2874.160724590950847045+-+70.710678118654752440+%29%29+%2F+%2870.710678118654752440+*+74.160724590950847045%29
				// expectedTick:     86129.0 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C5499.813071854861679877%5D

				// create second position parameters
				newLowerPrice := sdk.NewDec(5500)
				s.Require().NoError(err)
				newLowerTick := math.PriceToTick(newLowerPrice) // 86129
				newUpperPrice := sdk.NewDec(6250)
				s.Require().NoError(err)
				newUpperTick := math.PriceToTick(newUpperPrice) // 87407

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64(), DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
				// params
				// liquidity (2nd):  1198107969.043941799742686113
				// sqrtPriceNext:    78.136538612066933125 which is 6105.318666275026731905 https://www.wolframalpha.com/input?i=74.160724590950847045+%2B+4763454462.135821429683179+%2F+1198107969.043941799742686113
				// sqrtPriceCurrent: 74.160724590950847045 which is 5499.813071854861679877 https://www.wolframalpha.com/input?i2d=true&i=Power%5B1.0001%2CDivide%5B86129%2C2%5D%5D
				// expectedTokenIn:  4763454462.13582143006801819 rounded up https://www.wolframalpha.com/input?i=1198107969.043941799742686113+*+%2878.136538612066933125+-+74.160724590950847045%29
				// expectedTokenOut: 822041.7685 rounded down https://www.wolframalpha.com/input?i=%281198107969.043941799742686113+*+%2878.136538612066933125+-+74.160724590950847045+%29%29+%2F+%2874.160724590950847045+*+78.136538612066933125%29
				// expectedTick:     87173.5 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C6105.318666275026731905%5D
			},
			tokenIn:       sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			tokenOutDenom: "eth",
			priceLimit:    sdk.NewDec(6106),
			// expectedTokenIn:  5236545537.864178570316821 + 4763454462.13582143006801819 = 10000000000 usdc
			// expectedTokenOut: 998587.023047 + 822041.7685 = 1820628.791 round down = 1.820628 eth
			expectedTokenIn:   sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			expectedTokenOut:  sdk.NewCoin("eth", sdk.NewInt(1820628)),
			expectedTick:      sdk.NewInt(87173),
			expectedSqrtPrice: sdk.MustNewDecFromStr("78.136538612066933125"),
			newLowerPrice:     sdk.NewDec(5500),
			newUpperPrice:     sdk.NewDec(6250),
		},
		//  Consecutive price ranges
		//
		//                     5000
		//             4545 -----|----- 5500
		//  4000 ----------- 4545
		//
		"two positions with consecutive price ranges: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
				// params
				// liquidity (1st):  1517818840.967415409394235163
				// sqrtPriceNext:    67.416477345120317059 which is 4544.98141762512095360 (this is calculated by finding the closest tick LTE the upper range of the first range) https://www.wolframalpha.com/input?key=&i2d=true&i=Power%5B1.0001%2CDivide%5B84222%2C2%5D%5D
				// sqrtPriceCurrent: 70.710678118654752440 which is 5000
				// expectedTokenIn:  1048863.4367036217364698923810 rounded up https://www.wolframalpha.com/input?i=%281517818840.967415409394235163+*+%2870.710678118654752440+-+67.416477345120317059%29%29+%2F+%2867.416477345120317059+*+70.710678118654752440%29
				// expectedTokenOut: 5000000000.000 rounded down https://www.wolframalpha.com/input?key=&i=1517818840.967415409394235163+*+%2870.710678118654752440-+67.416477345120317059%29
				// expectedTick:     84222.0 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C4544.98141762512095360%5D

				// create second position parameters
				newLowerPrice := sdk.NewDec(4000)
				s.Require().NoError(err)
				newLowerTick := math.PriceToTick(newLowerPrice) // 82944
				newUpperPrice := sdk.NewDec(4545)
				s.Require().NoError(err)
				newUpperTick := math.PriceToTick(newUpperPrice) // 84222

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64(), DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
				// params
				// liquidity (2nd):  1198190689.904481374284815563
				// sqrtPriceNext:    63.991892380367981543 which is 4094.96229042061943099 https://www.wolframalpha.com/input?i=%281198190689.904481374284815563+*+67.416477345120317059%29+%2F+%28%281198190689.904481374284815563%29+%2B+%28951136.5632963782635301076189857337+*+67.416477345120317059%29%29
				// sqrtPriceCurrent: 67.416477345120317059 which is 4544.98141762512095360 https://www.wolframalpha.com/input?key=&i2d=true&i=Power%5B1.0001%2CDivide%5B84222%2C2%5D%5D
				// expectedTokenIn:  951136.56329637826376673854082567826779 rounded up https://www.wolframalpha.com/input?i=%281198190689.904481374284815563+*+%2867.416477345120317059+-+63.991892380367981543%29%29+%2F+%2863.991892380367981543+*+67.416477345120317059%29
				// expectedTokenOut: 4103305821.5531149215495196159979505795 rounded down https://www.wolframalpha.com/input?i=1198190689.904481374284815563+*+%2867.416477345120317059-+63.991892380367981543%29
				// expectedTick:     83179.3 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C4094.96229042061943099%5D
			},
			tokenIn:       sdk.NewCoin("eth", sdk.NewInt(2000000)),
			tokenOutDenom: "usdc",
			priceLimit:    sdk.NewDec(4094),
			// expectedTokenIn:  1048863.43670362173646989238101426629093898924360016563101536 + 951136.5632963782635301076189857337 = 2000000 eth
			// expectedTokenOut: 5000000000.000 + 4103305821.5679708 = 9103305821.5679708 round down = 9103.305821 usdc
			expectedTokenIn:   sdk.NewCoin("eth", sdk.NewInt(2000000)),
			expectedTokenOut:  sdk.NewCoin("usdc", sdk.NewInt(9103305821)),
			expectedTick:      sdk.NewInt(83179),
			expectedSqrtPrice: sdk.MustNewDecFromStr("63.991892380367981544"), // ends in 4 instead of 3 because we round up on token0 > token1 swaps
			newLowerPrice:     sdk.NewDec(4000),
			newUpperPrice:     sdk.NewDec(4545),
		},
		//  Partially overlapping price ranges
		//
		//          5000
		//  4545 -----|----- 5500
		//        5001 ----------- 6250
		//
		"two positions with partially overlapping price ranges: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
				// params
				// liquidity (1st):  1517818840.967415409394235163
				// sqrtPriceNext:    74.160724590950847045 which is 5499.813071854861679877199 (this is calculated by finding the closest tick LTE the upper range of the first range) https://www.wolframalpha.com/input?i2d=true&i=Power%5B1.0001%2CDivide%5B86129%2C2%5D%5D
				// sqrtPriceCurrent: 70.710678118654752440 which is 5000
				// expectedTokenIn:  5236545537.86417857031682136476 rounded up https://www.wolframalpha.com/input?i=1517818840.967415409394235163+*+%2874.160724590950847045+-+70.710678118654752440%29
				// expectedTokenOut: 998587.02304743550633886321769 rounded down https://www.wolframalpha.com/input?i=%281517818840.967415409394235163+*+%2874.160724590950847045+-+70.710678118654752440+%29%29+%2F+%2870.710678118654752440+*+74.160724590950847045%29
				// expectedTick:     86129.0 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C5499.813071854861679877199%5D

				// create second position parameters
				newLowerPrice := sdk.NewDec(5001)
				s.Require().NoError(err)
				newLowerTick := math.PriceToTick(newLowerPrice) // 85178
				newUpperPrice := sdk.NewDec(6250)
				s.Require().NoError(err)
				newUpperTick := math.PriceToTick(newUpperPrice) // 87407

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64(), DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
				// params
				// liquidity (2nd):  670565280.937709473686409722
				// sqrtPriceNext:    77.820305833374877582 which is 6056.0000000000000000 we hit the price limit here, so we just use the user defined max (6056)
				// sqrtPriceCurrent: 70.717075849691041259 which is 5000.9048167309559050
				// expectedTokenIn:  4763179409.57411318488036031546107 rounded up https://www.wolframalpha.com/input?i=670565280.937709473686409722+*+%2877.820305833374877582+-+70.717075849691041259%29
				// expectedTokenOut: 865525.190787795280 rounded down https://www.wolframalpha.com/input?i=%28670565280.937709473686409722+*+%2877.820305833374877582+-+70.717075849691041259+%29%29+%2F+%2870.717075849691041259+*+77.820305833374877582%29
				// expectedTick:     87092.4 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C6056.0000000000000000%5D
			},
			tokenIn:       sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			tokenOutDenom: "eth",
			priceLimit:    sdk.NewDec(6056),
			// expectedTokenIn:  5236545537.86417857031682136476 + 4763179409.57411318488036031546107 = 9999724947.43829175519718168022107 = 999972.49 usdc
			// expectedTokenOut: 998587.023 + 865525.190 = 1864112.213 round down = 1.864112 eth
			expectedTokenIn:   sdk.NewCoin("usdc", sdk.NewInt(9999724947)),
			expectedTokenOut:  sdk.NewCoin("eth", sdk.NewInt(1864112)),
			expectedTick:      sdk.NewInt(87092),
			expectedSqrtPrice: sdk.MustNewDecFromStr("77.820305833374877582"),
			newLowerPrice:     sdk.NewDec(5001),
			newUpperPrice:     sdk.NewDec(6250),
		},
		"two positions with partially overlapping price ranges, not utilizing full liquidity of second position: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
				// params
				// liquidity (1st):  1517818840.967415409394235163
				// sqrtPriceNext:    74.160724590950847045 which is 5499.8130718548616798771996903351 (this is calculated by finding the closest tick LTE the upper range of the first range) https://www.wolframalpha.com/input?i2d=true&i=Power%5B1.0001%2CDivide%5B86129%2C2%5D%5D
				// sqrtPriceCurrent: 70.710678118654752440 which is 5000
				// expectedTokenIn:  5236545537.864178570316821364761994285865595615 rounded up https://www.wolframalpha.com/input?i=1517818840.967415409394235163+*+%2874.160724590950847045+-+70.710678118654752440%29
				// expectedTokenOut: 998587.023047435506338863217691350 rounded down https://www.wolframalpha.com/input?i=%281517818840.967415409394235163+*+%2874.160724590950847045+-+70.710678118654752440+%29%29+%2F+%2870.710678118654752440+*+74.160724590950847045%29
				// expectedTick:     86129.0 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C5499.8130718548616798771996903351%5D

				// create second position parameters
				newLowerPrice := sdk.NewDec(5001)
				s.Require().NoError(err)
				newLowerTick := math.PriceToTick(newLowerPrice) // 85178
				newUpperPrice := sdk.NewDec(6250)
				s.Require().NoError(err)
				newUpperTick := math.PriceToTick(newUpperPrice) // 87407

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64(), DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
				// params
				// liquidity (2nd):  670565280.937709473686409722
				// sqrtPriceNext:    75.583797338122787359 which is 5712.9104200502884761 https://www.wolframalpha.com/input?i=70.717075849691041259+%2B+%283263454462.135821429683178635238005714134404385+%2F+670565280.937709473686409722%29
				// sqrtPriceCurrent: 70.717075849691041259 which is 5000.9048167309559050
				// expectedTokenIn:  3263454462.1358214299310811 rounded up https://www.wolframalpha.com/input?i=670565280.937709473686409722+*+%2875.583797338122787359+-+70.717075849691041259%29
				// expectedTokenOut: 610554.667370699018656 rounded down https://www.wolframalpha.com/input?i=%28670565280.937709473686409722+*+%2875.583797338122787359+-+70.717075849691041259+%29%29+%2F+%2870.717075849691041259+*+75.583797338122787359%29
				// expectedTick:     86509.2 rounded down https://www.wolframalpha.com/input?key=&i2d=true&i=Log%5B1.0001%2C5712.9104200502884761%5D
			},
			tokenIn:       sdk.NewCoin("usdc", sdk.NewInt(8500000000)),
			tokenOutDenom: "eth",
			priceLimit:    sdk.NewDec(6056),
			// expectedTokenIn:  5236545537.864178570316821364761994285865595615 + 3263454462.135821429683178635238005714134404385 = 8500000000.000 = 8500.00 usdc
			// expectedTokenOut: 998587.023 + 610554.667 = 1609141.69 round down = 1.609141 eth
			expectedTokenIn:   sdk.NewCoin("usdc", sdk.NewInt(8500000000)),
			expectedTokenOut:  sdk.NewCoin("eth", sdk.NewInt(1609141)),
			expectedTick:      sdk.NewInt(86509),
			expectedSqrtPrice: sdk.MustNewDecFromStr("75.583797338122787359"),
			newLowerPrice:     sdk.NewDec(5001),
			newUpperPrice:     sdk.NewDec(6250),
		},
		//  Partially overlapping price ranges
		//
		//                5000
		//        4545 -----|----- 5500
		//  4000 ----------- 4999
		//
		"two positions with partially overlapping price ranges: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
				// params
				// liquidity (1st):  1517818840.967415409394235163
				// sqrtPriceNext:    67.416477345120317059 which is 4544.98141762512095360 https://www.wolframalpha.com/input?i2d=true&i=Sqrt%5BPower%5B1.0001%2C84222%5D%5D
				// sqrtPriceCurrent: 70.710678118654752440 which is 5000
				// expectedTokenIn:  1048863.43670362173646989238101426629093898924360016563101536 rounded up https://www.wolframalpha.com/input?i=%281517818840.967415409394235163+*+%2870.710678118654752440+-+67.416477345120317059%29%29+%2F+%2867.416477345120317059+*+70.710678118654752440%29
				// expectedTokenOut: 5000000000.000 rounded down https://www.wolframalpha.com/input?i=1517818840.967415409394235163+*+%2870.710678118654752440-+67.416477345120317059%29
				// expectedTick:     84222.0 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C4544.98141762512095360%5D

				// create second position parameters
				newLowerPrice := sdk.NewDec(4000)
				s.Require().NoError(err)
				newLowerTick := math.PriceToTick(newLowerPrice) // 82944
				newUpperPrice := sdk.NewDec(4999)
				s.Require().NoError(err)
				newUpperTick := math.PriceToTick(newUpperPrice) // 85174

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64(), DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
				// params
				// liquidity (2nd):  670293788.068824290816957143
				// sqrtPriceNext:    64.256329884061506898 which is 4128.875930169335868589 https://www.wolframalpha.com/input?i=%28%28670293788.068824290816957143%29+%2F+%28%28%28670293788.068824290816957143%29+%2F+%2870.702934555750545592%29%29+%2B+%28951136.56329637826353010761898573370906101075639983436898464%29%29
				// sqrtPriceCurrent: 70.702934555750545592 which is 4998.904954794744599825 https://www.wolframalpha.com/input?i2d=true&i=Sqrt%5BPower%5B1.0001%2C85174%5D%5D
				// expectedTokenIn:  951136.563296378263628439905390538 rounded up https://www.wolframalpha.com/input?i=%28670293788.068824290816957143+*+%2870.702934555750545592+-+64.256329884061506898%29%29+%2F+%2864.256329884061506898+*+70.702934555750545592%29
				// expectedTokenOut: 4321119065.56862509898611380197721 rounded down https://www.wolframalpha.com/input?i=670293788.068824290816957143+*+%2870.702934555750545592-+64.256329884061506898%29
				// expectedTick:     83261.8 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C4128.875930169335868589%5D
			},
			tokenIn:       sdk.NewCoin("eth", sdk.NewInt(2000000)),
			tokenOutDenom: "usdc",
			priceLimit:    sdk.NewDec(4128),
			// expectedTokenIn:  1048863.43670362173646989238101426629093898924360016563101536 + 951136.56329637826353010761898573370906101075639983436898464 = 2000000 eth
			// expectedTokenOut: 5000000000.000 + 4321119065.568625098986113 = 9321119065.568625098986113 round down = 9321.119065 usdc
			expectedTokenIn:   sdk.NewCoin("eth", sdk.NewInt(2000000)),
			expectedTokenOut:  sdk.NewCoin("usdc", sdk.NewInt(9321119065)),
			expectedTick:      sdk.NewInt(83261),
			expectedSqrtPrice: sdk.MustNewDecFromStr("64.256329884061506899"), // ends in 9 instead of 8 because we round up on token0 > token1 swaps
			newLowerPrice:     sdk.NewDec(4000),
			newUpperPrice:     sdk.NewDec(4999),
		},
		"two positions with partially overlapping price ranges, not utilizing full liquidity of second position: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
				// params
				// liquidity (1st):  1517818840.967415409394235163
				// sqrtPriceNext:    67.416477345120317059 which is 4544.98141762512095360 https://www.wolframalpha.com/input?i2d=true&i=Sqrt%5BPower%5B1.0001%2C84222%5D%5D
				// sqrtPriceCurrent: 70.710678118654752440 which is 5000
				// expectedTokenIn:  1048863.43670362173646989238101426629093898924360016563101536 rounded up https://www.wolframalpha.com/input?i=%281517818840.967415409394235163+*+%2870.710678118654752440+-+67.416477345120317059%29%29+%2F+%2867.416477345120317059+*+70.710678118654752440%29
				// expectedTokenOut: 5000000000.000 rounded down https://www.wolframalpha.com/input?i=1517818840.967415409394235163+*+%2870.710678118654752440-+67.416477345120317059%29
				// expectedTick:     84222.0 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C4544.98141762512095360%5D

				// create second position parameters
				newLowerPrice := sdk.NewDec(4000)
				s.Require().NoError(err)
				newLowerTick := math.PriceToTick(newLowerPrice) // 82944
				newUpperPrice := sdk.NewDec(4999)
				s.Require().NoError(err)
				newUpperTick := math.PriceToTick(newUpperPrice) // 85174

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64(), DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
				// params
				// liquidity (2nd):  670293788.068824290816957143
				// sqrtPriceNext:    65.512371527681108800 which is 4291.870823180922417 https://www.wolframalpha.com/input?i=%28%28670293788.068824290816957143%29+%2F+%28%28%28670293788.068824290816957143%29+%2F+%2870.702934555750545592%29%29+%2B+%28751136.56329637826353010761898573370906101075639983436898464%29%29
				// sqrtPriceCurrent: 70.702934555750545592 which is 4998.9044954 https://www.wolframalpha.com/input?i2d=true&i=Sqrt%5BPower%5B1.0001%2C85174%5D%5D
				// expectedTokenIn:  751136.5632963782636688068845751300415841768 rounded up https://www.wolframalpha.com/input?i=%28670293788.068824290816957143+*+%2870.702934555750545592+-+65.512371527681108800%29%29+%2F+%2865.512371527681108800+*+70.702934555750545592%29
				// expectedTokenOut: 3479202154.294649933683844701072876595280842 rounded down https://www.wolframalpha.com/input?i=670293788.068824290816957143+*+%2870.702934555750545592-+65.512371527681108800%29
				// expectedTick:     83649.0 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C4291.870823180922417%5D
			},
			tokenIn:       sdk.NewCoin("eth", sdk.NewInt(1800000)),
			tokenOutDenom: "usdc",
			priceLimit:    sdk.NewDec(4128),
			// expectedTokenIn:  1048863.43670362173646989238101426629093898924360016563101536 + 751136.56329637826353010761898573370906101075639983436898464 = 1.800000 eth
			// expectedTokenOut: 5000000000.000 + 3479202154.310192937 = 8479202154.310192937 round down = 8479.202154 usdc
			expectedTokenIn:   sdk.NewCoin("eth", sdk.NewInt(1800000)),
			expectedTokenOut:  sdk.NewCoin("usdc", sdk.NewInt(8479202154)),
			expectedTick:      sdk.NewInt(83648),
			expectedSqrtPrice: sdk.MustNewDecFromStr("65.512371527681108801"), // ends in 1 instead of 0 because we round up on token0 > token1 swaps
			newLowerPrice:     sdk.NewDec(4000),
			newUpperPrice:     sdk.NewDec(4999),
		},
		//  Sequential price ranges with a gap
		//
		//          5000
		//  4545 -----|----- 5500
		//              5501 ----------- 6250
		//
		"two sequential positions with a gap": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
				// params
				// liquidity (1st):  1517818840.967415409394235163
				// sqrtPriceNext:    74.160724590950847045 which is 5499.813071854861679877 (this is calculated by finding the closest tick LTE the upper range of the first range) https://www.wolframalpha.com/input?i2d=true&i=Power%5B1.0001%2CDivide%5B86129%2C2%5D%5D
				// sqrtPriceCurrent: 70.710678118654752440 which is 5000
				// expectedTokenIn:  5236545537.8641785703168213647619942 rounded up https://www.wolframalpha.com/input?i=1517818840.967415409394235163+*+%2874.160724590950847045+-+70.710678118654752440%29
				// expectedTokenOut: 998587.023047435506338863217691350 rounded down https://www.wolframalpha.com/input?i=%281517818840.967415409394235163+*+%2874.160724590950847045+-+70.710678118654752440+%29%29+%2F+%2870.710678118654752440+*+74.160724590950847045%29
				// expectedTick:     86129.0 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C5499.813071854861679877%5D

				// create second position parameters
				newLowerPrice := sdk.NewDec(5501)
				s.Require().NoError(err)
				newLowerTick := math.PriceToTick(newLowerPrice) // 86131
				newUpperPrice := sdk.NewDec(6250)
				s.Require().NoError(err)
				newUpperTick := math.PriceToTick(newUpperPrice) // 87407

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64(), DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
				// params
				// liquidity (2nd):  1200046517.432642062946883869
				// sqrtPriceNext:    78.137532176937826230 which is 6105.473934701923906716 https://www.wolframalpha.com/input?i=74.168140663409942130++%2B++4763454462.1358214296831786352380058+%2F1200046517.432642062946883869
				// sqrtPriceCurrent: 74.168140663409942130 which is 5500.9130894673633706922277169 https://www.wolframalpha.com/input?i2d=true&i=Power%5B1.0001%2CDivide%5B86131%2C2%5D%5D
				// expectedTokenIn:  4763454462.13582142964 rounded up https://www.wolframalpha.com/input?i=1200046517.432642062946883869+*+%2878.137532176937826230+-+74.168140663409942130%29
				// expectedTokenOut: 821949.120898595865033 rounded down https://www.wolframalpha.com/input?i=%281200046517.432642062946883869+*+%2878.137532176937826230+-+74.168140663409942130+%29%29+%2F+%2874.168140663409942130+*+78.137532176937826230%29
				// expectedTick:     87173.8 rounded down https://www.wolframalpha.com/input?i2d=true&i=Log%5B1.0001%2C6105.473934701923906716%5D
			},
			tokenIn:       sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			tokenOutDenom: "eth",
			priceLimit:    sdk.NewDec(6106),
			// expectedTokenIn:  5236545537.8641785703168213647619942 + 4763454462.13582142964 = 10000000000 usdc
			// expectedTokenOut: 998587.023047435506338863217691350 + 821949.120898595865033 = 1820536.143 round down = 1.820536 eth
			expectedTokenIn:   sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			expectedTokenOut:  sdk.NewCoin("eth", sdk.NewInt(1820536)),
			expectedTick:      sdk.NewInt(87173),
			expectedSqrtPrice: sdk.MustNewDecFromStr("78.137532176937826230"),
			newLowerPrice:     sdk.NewDec(5501),
			newUpperPrice:     sdk.NewDec(6250),
		},
		// Slippage protection doesn't cause a failure but interrupts early.
		"single position within one tick, trade completes but slippage protection interrupts trade early: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("eth", sdk.NewInt(13370)),
			tokenOutDenom: "usdc",
			priceLimit:    sdk.NewDec(4994),
			// params
			// liquidity: 		 1517818840.967415409394235163
			// sqrtPriceNext:    70.668238976219012614 which is 4994 https://www.wolframalpha.com/input?i=70.710678118654752440+%2B+42000000+%2F+1517818840.967415409394235163
			// sqrtPriceCurrent: 70.710678118654752440 which is 5000
			// expectedTokenIn:  12890.72275 rounded up https://www.wolframalpha.com/input?key=&i=%281517818840.967415409394235163+*+%2870.710678118654752440+-+70.668238976219012614+%29%29+%2F+%2870.710678118654752440+*+70.668238976219012614%29
			// expectedTokenOut: 64414929.9834 rounded down https://www.wolframalpha.com/input?key=&i=1517818840.967415409394235163+*+%2870.710678118654752440+-+70.668238976219012614%29
			// expectedTick: 	 85164.2 rounded down https://www.wolframalpha.com/input?key=&i2d=true&i=Log%5B1.0001%2C4994%5D
			expectedTokenIn:   sdk.NewCoin("eth", sdk.NewInt(12891)),
			expectedTokenOut:  sdk.NewCoin("usdc", sdk.NewInt(64414929)),
			expectedTick:      sdk.NewInt(85164),
			expectedSqrtPrice: sdk.MustNewDecFromStr("70.668238976219012614"),
		},
		"single position within one tick, trade does not complete due to lack of liquidity: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("usdc", sdk.NewInt(5300000000)),
			tokenOutDenom: "eth",
			priceLimit:    sdk.NewDec(6000),
			expectErr:     true,
		},
		"single position within one tick, trade does not complete due to lack of liquidity: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("eth", sdk.NewInt(1100000)),
			tokenOutDenom: "usdc",
			priceLimit:    sdk.NewDec(4000),
			expectErr:     true,
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.Setup()
			s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(10000000000000)), sdk.NewCoin("usdc", sdk.NewInt(1000000000000))))
			s.FundAcc(s.TestAccs[1], sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(10000000000000)), sdk.NewCoin("usdc", sdk.NewInt(1000000000000))))

			// Create default CL pool
			pool := s.PrepareConcentratedPool()

			// add positions
			test.addPositions(s.Ctx, pool.GetId())

			// perform calc
			tokenIn, tokenOut, updatedTick, updatedLiquidity, updatedSqrtPrice, err := s.App.ConcentratedLiquidityKeeper.CalcOutAmtGivenInInternal(
				s.Ctx,
				test.tokenIn, test.tokenOutDenom,
				DefaultZeroSwapFee, test.priceLimit, pool.GetId())
			if test.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				// check that tokenIn, tokenOut, tick, and sqrtPrice from CalcOut are all what we expected
				s.Require().Equal(test.expectedTick.String(), updatedTick.String())
				s.Require().Equal(test.expectedSqrtPrice.String(), updatedSqrtPrice.String())
				s.Require().Equal(test.expectedTokenIn.String(), tokenIn.String())
				s.Require().Equal(test.expectedTokenOut.String(), tokenOut.String())

				if test.newLowerPrice.IsNil() && test.newUpperPrice.IsNil() {
					test.newLowerPrice = DefaultLowerPrice
					test.newUpperPrice = DefaultUpperPrice
				}

				newLowerTick := math.PriceToTick(test.newLowerPrice)
				newUpperTick := math.PriceToTick(test.newUpperPrice)

				lowerSqrtPrice, err := math.TickToSqrtPrice(newLowerTick)
				s.Require().NoError(err)
				upperSqrtPrice, err := math.TickToSqrtPrice(newUpperTick)
				s.Require().NoError(err)

				if test.poolLiqAmount0.IsNil() && test.poolLiqAmount1.IsNil() {
					test.poolLiqAmount0 = DefaultAmt0
					test.poolLiqAmount1 = DefaultAmt1
				}

				// check that liquidity is what we expected
				expectedLiquidity := math.GetLiquidityFromAmounts(DefaultCurrSqrtPrice, lowerSqrtPrice, upperSqrtPrice, test.poolLiqAmount0, test.poolLiqAmount1)
				s.Require().Equal(expectedLiquidity.String(), updatedLiquidity.String())
			}

			// perform swap
			tokenIn, tokenOut, updatedTick, updatedLiquidity, updatedSqrtPrice, err = s.App.ConcentratedLiquidityKeeper.SwapOutAmtGivenIn(
				s.Ctx,
				test.tokenIn, test.tokenOutDenom,
				DefaultZeroSwapFee, test.priceLimit, pool.GetId())
			if test.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				s.Require().Equal(test.expectedTokenIn.String(), tokenIn.String())

				pool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
				s.Require().NoError(err)

				// check that tokenIn, tokenOut, tick, and sqrtPrice from SwapOut are all what we expected
				s.Require().Equal(test.expectedTick.String(), updatedTick.String())
				s.Require().Equal(test.expectedSqrtPrice.String(), updatedSqrtPrice.String())
				s.Require().Equal(test.expectedTokenIn.String(), tokenIn.String())
				s.Require().Equal(test.expectedTokenOut.String(), tokenOut.String())
				// also ensure the pool's currentTick and currentSqrtPrice was updated due to calling a mutative method
				s.Require().Equal(test.expectedTick.String(), pool.GetCurrentTick().String())
				s.Require().Equal(test.expectedSqrtPrice.String(), pool.GetCurrentSqrtPrice().String())

				if test.newLowerPrice.IsNil() && test.newUpperPrice.IsNil() {
					test.newLowerPrice = DefaultLowerPrice
					test.newUpperPrice = DefaultUpperPrice
				}

				newLowerTick := math.PriceToTick(test.newLowerPrice)
				newUpperTick := math.PriceToTick(test.newUpperPrice)

				lowerSqrtPrice, err := math.TickToSqrtPrice(newLowerTick)
				s.Require().NoError(err)
				upperSqrtPrice, err := math.TickToSqrtPrice(newUpperTick)
				s.Require().NoError(err)

				if test.poolLiqAmount0.IsNil() && test.poolLiqAmount1.IsNil() {
					test.poolLiqAmount0 = DefaultAmt0
					test.poolLiqAmount1 = DefaultAmt1
				}

				expectedLiquidity := math.GetLiquidityFromAmounts(DefaultCurrSqrtPrice, lowerSqrtPrice, upperSqrtPrice, test.poolLiqAmount0, test.poolLiqAmount1)
				// check that liquidity is what we expected
				s.Require().Equal(expectedLiquidity.String(), pool.GetLiquidity().String())
				// also ensure the pool's currentLiquidity was updated due to calling a mutative method
				s.Require().Equal(expectedLiquidity.String(), updatedLiquidity.String())
			}
		})

	}
}

func (s *KeeperTestSuite) TestCalcAndSwapInAmtGivenOut() {
	tests := map[string]struct {
		positionAmount0   sdk.Int
		positionAmount1   sdk.Int
		addPositions      func(ctx sdk.Context, poolId uint64)
		tokenOut          sdk.Coin
		tokenInDenom      string
		priceLimit        sdk.Dec
		expectedTokenIn   sdk.Coin
		expectedTokenOut  sdk.Coin
		expectedTick      sdk.Int
		expectedSqrtPrice sdk.Dec
		newLowerPrice     sdk.Dec
		newUpperPrice     sdk.Dec
		poolLiqAmount0    sdk.Int
		poolLiqAmount1    sdk.Int
		expectErr         bool
	}{
		//  One price range
		//
		//          5000
		//  4545 -----|----- 5500
		"single position within one tick: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenOut:          sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			tokenInDenom:      "eth",
			priceLimit:        sdk.NewDec(5004),
			expectedTokenOut:  sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			expectedTokenIn:   sdk.NewCoin("eth", sdk.NewInt(8396)),
			expectedTick:      sdk.NewInt(85184),
			expectedSqrtPrice: sdk.MustNewDecFromStr("70.738349405152441697"),
		},
		"single position within one tick: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenOut:          sdk.NewCoin("eth", sdk.NewInt(13370)),
			tokenInDenom:      "usdc",
			priceLimit:        sdk.NewDec(4993),
			expectedTokenOut:  sdk.NewCoin("eth", sdk.NewInt(13370)),
			expectedTokenIn:   sdk.NewCoin("usdc", sdk.NewInt(66808387)),
			expectedTick:      sdk.NewInt(85163),
			expectedSqrtPrice: sdk.MustNewDecFromStr("70.666662070528898354"),
		},
		//  Two equal price ranges
		//
		//          5000
		//  4545 -----|----- 5500
		//  4545 -----|----- 5500
		"two positions within one tick: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)

				// add second position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenOut:          sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			tokenInDenom:      "eth",
			priceLimit:        sdk.NewDec(5002),
			expectedTokenOut:  sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			expectedTokenIn:   sdk.NewCoin("eth", sdk.NewInt(8398)),
			expectedTick:      sdk.NewInt(85180),
			expectedSqrtPrice: sdk.MustNewDecFromStr("70.724513761903597069"),
			// two positions with same liquidity entered
			poolLiqAmount0: sdk.NewInt(1000000).MulRaw(2),
			poolLiqAmount1: sdk.NewInt(5000000000).MulRaw(2),
		},
		"two positions within one tick: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)

				// add second position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenOut:          sdk.NewCoin("eth", sdk.NewInt(13370)),
			tokenInDenom:      "usdc",
			priceLimit:        sdk.NewDec(4996),
			expectedTokenOut:  sdk.NewCoin("eth", sdk.NewInt(13370)),
			expectedTokenIn:   sdk.NewCoin("usdc", sdk.NewInt(66829187)),
			expectedTick:      sdk.NewInt(85169),
			expectedSqrtPrice: sdk.MustNewDecFromStr("70.688663242671673183"),
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
		"two positions with consecutive price ranges: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)

				// create second position parameters
				newLowerPrice := sdk.NewDec(5500)
				s.Require().NoError(err)
				newLowerTick := math.PriceToTick(newLowerPrice) // 86129
				newUpperPrice := sdk.NewDec(6250)
				s.Require().NoError(err)
				newUpperTick := math.PriceToTick(newUpperPrice) // 87407

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64(), DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenOut:          sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			tokenInDenom:      "eth",
			priceLimit:        sdk.NewDec(6106),
			expectedTokenOut:  sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			expectedTokenIn:   sdk.NewCoin("eth", sdk.NewInt(1820628)),
			expectedTick:      sdk.NewInt(87173),
			expectedSqrtPrice: sdk.MustNewDecFromStr("78.136538612066933125"),
			newLowerPrice:     sdk.NewDec(5500),
			newUpperPrice:     sdk.NewDec(6250),
		},
		//  Consecutive price ranges
		//
		//                     5000
		//             4545 -----|----- 5500
		//  4000 ----------- 4545
		//
		"two positions with consecutive price ranges: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)

				// create second position parameters
				newLowerPrice := sdk.NewDec(4000)
				s.Require().NoError(err)
				newLowerTick := math.PriceToTick(newLowerPrice) // 82944
				newUpperPrice := sdk.NewDec(4545)
				s.Require().NoError(err)
				newUpperTick := math.PriceToTick(newUpperPrice) // 84222

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64(), DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenOut:          sdk.NewCoin("eth", sdk.NewInt(2000000)),
			tokenInDenom:      "usdc",
			priceLimit:        sdk.NewDec(4094),
			expectedTokenOut:  sdk.NewCoin("eth", sdk.NewInt(2000000)),
			expectedTokenIn:   sdk.NewCoin("usdc", sdk.NewInt(9103305821)),
			expectedTick:      sdk.NewInt(83179),
			expectedSqrtPrice: sdk.MustNewDecFromStr("63.991892380367981544"),
			newLowerPrice:     sdk.NewDec(4000),
			newUpperPrice:     sdk.NewDec(4545),
		},
		//  Partially overlapping price ranges
		//
		//          5000
		//  4545 -----|----- 5500
		//        5001 ----------- 6250
		//
		"two positions with partially overlapping price ranges: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)

				// create second position parameters
				newLowerPrice := sdk.NewDec(5001)
				s.Require().NoError(err)
				newLowerTick := math.PriceToTick(newLowerPrice) // 85178
				newUpperPrice := sdk.NewDec(6250)
				s.Require().NoError(err)
				newUpperTick := math.PriceToTick(newUpperPrice) // 87407

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64(), DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenOut:          sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			tokenInDenom:      "eth",
			priceLimit:        sdk.NewDec(6056),
			expectedTokenOut:  sdk.NewCoin("usdc", sdk.NewInt(9999724947)),
			expectedTokenIn:   sdk.NewCoin("eth", sdk.NewInt(1864112)),
			expectedTick:      sdk.NewInt(87092),
			expectedSqrtPrice: sdk.MustNewDecFromStr("77.820305833374877582"),
			newLowerPrice:     sdk.NewDec(5001),
			newUpperPrice:     sdk.NewDec(6250),
		},
		"two positions with partially overlapping price ranges, not utilizing full liquidity of second position: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)

				// create second position parameters
				newLowerPrice := sdk.NewDec(5001)
				s.Require().NoError(err)
				newLowerTick := math.PriceToTick(newLowerPrice) // 85178
				newUpperPrice := sdk.NewDec(6250)
				s.Require().NoError(err)
				newUpperTick := math.PriceToTick(newUpperPrice) // 87407

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64(), DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenOut:          sdk.NewCoin("usdc", sdk.NewInt(8500000000)),
			tokenInDenom:      "eth",
			priceLimit:        sdk.NewDec(6056),
			expectedTokenOut:  sdk.NewCoin("usdc", sdk.NewInt(8500000000)),
			expectedTokenIn:   sdk.NewCoin("eth", sdk.NewInt(1609141)),
			expectedTick:      sdk.NewInt(86509),
			expectedSqrtPrice: sdk.MustNewDecFromStr("75.583797338122787359"),
			newLowerPrice:     sdk.NewDec(5001),
			newUpperPrice:     sdk.NewDec(6250),
		},
		//  Partially overlapping price ranges
		//
		//                5000
		//        4545 -----|----- 5500
		//  4000 ----------- 4999
		//
		"two positions with partially overlapping price ranges: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)

				// create second position parameters
				newLowerPrice := sdk.NewDec(4000)
				s.Require().NoError(err)
				newLowerTick := math.PriceToTick(newLowerPrice) // 82944
				newUpperPrice := sdk.NewDec(4999)
				s.Require().NoError(err)
				newUpperTick := math.PriceToTick(newUpperPrice) // 85174

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64(), DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenOut:          sdk.NewCoin("eth", sdk.NewInt(2000000)),
			tokenInDenom:      "usdc",
			priceLimit:        sdk.NewDec(4128),
			expectedTokenOut:  sdk.NewCoin("eth", sdk.NewInt(2000000)),
			expectedTokenIn:   sdk.NewCoin("usdc", sdk.NewInt(9321119065)),
			expectedTick:      sdk.NewInt(83261),
			expectedSqrtPrice: sdk.MustNewDecFromStr("64.256329884061506899"),
			newLowerPrice:     sdk.NewDec(4000),
			newUpperPrice:     sdk.NewDec(4999),
		},
		"two positions with partially overlapping price ranges, not utilizing full liquidity of second position: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)

				// create second position parameters
				newLowerPrice := sdk.NewDec(4000)
				s.Require().NoError(err)
				newLowerTick := math.PriceToTick(newLowerPrice) // 82944
				newUpperPrice := sdk.NewDec(4999)
				s.Require().NoError(err)
				newUpperTick := math.PriceToTick(newUpperPrice) // 85174

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64(), DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenOut:          sdk.NewCoin("eth", sdk.NewInt(1800000)),
			tokenInDenom:      "usdc",
			priceLimit:        sdk.NewDec(4128),
			expectedTokenOut:  sdk.NewCoin("eth", sdk.NewInt(1800000)),
			expectedTokenIn:   sdk.NewCoin("usdc", sdk.NewInt(8479202154)),
			expectedTick:      sdk.NewInt(83648),
			expectedSqrtPrice: sdk.MustNewDecFromStr("65.512371527681108801"),
			newLowerPrice:     sdk.NewDec(4000),
			newUpperPrice:     sdk.NewDec(4999),
		},
		//  Sequential price ranges with a gap
		//
		//          5000
		//  4545 -----|----- 5500
		//              5501 ----------- 6250
		//
		"two sequential positions with a gap": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)

				// create second position parameters
				newLowerPrice := sdk.NewDec(5501)
				s.Require().NoError(err)
				newLowerTick := math.PriceToTick(newLowerPrice) // 86131
				newUpperPrice := sdk.NewDec(6250)
				s.Require().NoError(err)
				newUpperTick := math.PriceToTick(newUpperPrice) // 87407

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64(), DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenOut:          sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			tokenInDenom:      "eth",
			priceLimit:        sdk.NewDec(6106),
			expectedTokenOut:  sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			expectedTokenIn:   sdk.NewCoin("eth", sdk.NewInt(1820536)),
			expectedTick:      sdk.NewInt(87173),
			expectedSqrtPrice: sdk.MustNewDecFromStr("78.137532176937826230"),
			newLowerPrice:     sdk.NewDec(5501),
			newUpperPrice:     sdk.NewDec(6250),
		},
		// Slippage protection doesn't cause a failure but interrupts early.
		"single position within one tick, trade completes but slippage protection interrupts trade early: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenOut:          sdk.NewCoin("eth", sdk.NewInt(13370)),
			tokenInDenom:      "usdc",
			priceLimit:        sdk.NewDec(4994),
			expectedTokenOut:  sdk.NewCoin("eth", sdk.NewInt(12891)),
			expectedTokenIn:   sdk.NewCoin("usdc", sdk.NewInt(64414929)),
			expectedTick:      sdk.NewInt(85164),
			expectedSqrtPrice: sdk.MustNewDecFromStr("70.668238976219012614"),
		},
		"single position within one tick, trade does not complete due to lack of liquidity: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenOut:     sdk.NewCoin("usdc", sdk.NewInt(5300000000)),
			tokenInDenom: "eth",
			priceLimit:   sdk.NewDec(6000),
			expectErr:    true,
		},
		"single position within one tick, trade does not complete due to lack of liquidity: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick, DefaultIncentiveIDsCommittedTo)
				s.Require().NoError(err)
			},
			tokenOut:     sdk.NewCoin("eth", sdk.NewInt(1100000)),
			tokenInDenom: "usdc",
			priceLimit:   sdk.NewDec(4000),
			expectErr:    true,
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.Setup()
			s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(10000000000000)), sdk.NewCoin("usdc", sdk.NewInt(1000000000000))))
			s.FundAcc(s.TestAccs[1], sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(10000000000000)), sdk.NewCoin("usdc", sdk.NewInt(1000000000000))))

			// Create default CL pool
			pool := s.PrepareConcentratedPool()

			// add positions
			test.addPositions(s.Ctx, pool.GetId())

			// perform calc
			tokenIn, tokenOut, updatedTick, updatedLiquidity, updatedSqrtPrice, err := s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOutInternal(
				s.Ctx,
				test.tokenOut, test.tokenInDenom,
				DefaultZeroSwapFee, test.priceLimit, pool.GetId())
			if test.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				// check that tokenIn, tokenOut, tick, and sqrtPrice from CalcOut are all what we expected
				s.Require().Equal(test.expectedTokenOut.String(), tokenOut.String())
				s.Require().Equal(test.expectedTokenIn.String(), tokenIn.String())
				s.Require().Equal(test.expectedTick.String(), updatedTick.String())
				s.Require().Equal(test.expectedSqrtPrice.String(), updatedSqrtPrice.String())

				if test.newLowerPrice.IsNil() && test.newUpperPrice.IsNil() {
					test.newLowerPrice = DefaultLowerPrice
					test.newUpperPrice = DefaultUpperPrice
				}

				newLowerTick := math.PriceToTick(test.newLowerPrice)
				newUpperTick := math.PriceToTick(test.newUpperPrice)

				lowerSqrtPrice, err := math.TickToSqrtPrice(newLowerTick)
				s.Require().NoError(err)
				upperSqrtPrice, err := math.TickToSqrtPrice(newUpperTick)
				s.Require().NoError(err)

				if test.poolLiqAmount0.IsNil() && test.poolLiqAmount1.IsNil() {
					test.poolLiqAmount0 = DefaultAmt0
					test.poolLiqAmount1 = DefaultAmt1
				}

				// check that liquidity is what we expected
				expectedLiquidity := math.GetLiquidityFromAmounts(DefaultCurrSqrtPrice, lowerSqrtPrice, upperSqrtPrice, test.poolLiqAmount0, test.poolLiqAmount1)
				s.Require().Equal(expectedLiquidity.String(), updatedLiquidity.String())
			}

			// perform swap
			tokenIn, tokenOut, updatedTick, updatedLiquidity, updatedSqrtPrice, err = s.App.ConcentratedLiquidityKeeper.SwapInAmtGivenOut(
				s.Ctx,
				test.tokenOut, test.tokenInDenom,
				DefaultZeroSwapFee, test.priceLimit, pool.GetId())
			if test.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				pool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
				s.Require().NoError(err)

				// check that tokenIn, tokenOut, tick, and sqrtPrice from SwapOut are all what we expected
				s.Require().Equal(test.expectedTick.String(), updatedTick.String())
				s.Require().Equal(test.expectedSqrtPrice.String(), updatedSqrtPrice.String())
				s.Require().Equal(test.expectedTokenIn.String(), tokenIn.String())
				s.Require().Equal(test.expectedTokenOut.String(), tokenOut.String())
				// also ensure the pool's currentTick and currentSqrtPrice was updated due to calling a mutative method
				s.Require().Equal(test.expectedTick.String(), pool.GetCurrentTick().String())
				s.Require().Equal(test.expectedSqrtPrice.String(), pool.GetCurrentSqrtPrice().String())

				if test.newLowerPrice.IsNil() && test.newUpperPrice.IsNil() {
					test.newLowerPrice = DefaultLowerPrice
					test.newUpperPrice = DefaultUpperPrice
				}

				newLowerTick := math.PriceToTick(test.newLowerPrice)
				newUpperTick := math.PriceToTick(test.newUpperPrice)

				lowerSqrtPrice, err := math.TickToSqrtPrice(newLowerTick)
				s.Require().NoError(err)
				upperSqrtPrice, err := math.TickToSqrtPrice(newUpperTick)
				s.Require().NoError(err)

				if test.poolLiqAmount0.IsNil() && test.poolLiqAmount1.IsNil() {
					test.poolLiqAmount0 = DefaultAmt0
					test.poolLiqAmount1 = DefaultAmt1
				}

				expectedLiquidity := math.GetLiquidityFromAmounts(DefaultCurrSqrtPrice, lowerSqrtPrice, upperSqrtPrice, test.poolLiqAmount0, test.poolLiqAmount1)
				// check that liquidity is what we expected
				s.Require().Equal(expectedLiquidity.String(), pool.GetLiquidity().String())
				// also ensure the pool's currentLiquidity was updated due to calling a mutative method
				s.Require().Equal(expectedLiquidity.String(), updatedLiquidity.String())
			}
		})

	}
}

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
			// liquidity: 		 1517818840.967415409394235163
			// sqrtPriceNext:    70.738349405152439867 which is 5003.914076565430543175 https://www.wolframalpha.com/input?i=70.710678118654752440+%2B+42000000+%2F+1517818840.967415409394235163
			// sqrtPriceCurrent: 70.710678118654752440 which is 5000
			// expectedTokenIn:  42000000.0000 rounded up https://www.wolframalpha.com/input?i=1517818840.967415409394235163+*+%2870.738349405152439867+-+70.710678118654752440%29
			// expectedTokenOut: 8396.714105 rounded down https://www.wolframalpha.com/input?i=%281517818840.967415409394235163+*+%2870.738349405152439867+-+70.710678118654752440+%29%29+%2F+%2870.710678118654752440+*+70.738349405152439867%29
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
			// liquidity: 		 1517818840.967415409394235163
			// sqrtPriceNext:    70.666662070529219856 which is 4993.777128190373086350 https://www.wolframalpha.com/input?i=%28%281517818840.967415409394235163%29%29+%2F+%28%28%281517818840.967415409394235163%29+%2F+%2870.710678118654752440%29%29+%2B+%2813370%29%29
			// expectedTokenIn:  13369.9999 rounded up https://www.wolframalpha.com/input?i=%281517818840.967415409394235163+*+%2870.710678118654752440+-+70.666662070529219856+%29%29+%2F+%2870.666662070529219856+*+70.710678118654752440%29
			// expectedTokenOut: 66808387.149 rounded down https://www.wolframalpha.com/input?i=1517818840.967415409394235163+*+%2870.710678118654752440+-+70.666662070529219856%29
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

			// Create a default CL pool
			pool := s.PrepareConcentratedPool()

			// Check the test case to see if we are swapping asset0 for asset1 or vice versa
			asset0 := pool.GetToken0()
			zeroForOne := test.param.tokenIn.Denom == asset0

			// Fund the test account with usdc and eth, then create a default position to the pool created earlier
			s.SetupIncentivizedPosition(1)

			// Retrieve pool post position set up
			pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
			s.Require().NoError(err)

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
				s.AssertEventEmitted(s.Ctx, cltypes.TypeEvtTokenSwapped, 1)

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
			// liquidity: 		 1517818840.967415409394235163
			// sqrtPriceNext:    70.738349405152439867 which is 5003.914076565430543175 https://www.wolframalpha.com/input?i=70.710678118654752440+%2B+42000000+%2F+1517818840.967415409394235163
			// sqrtPriceCurrent: 70.710678118654752440 which is 5000
			// expectedTokenIn:  42000000.0000 rounded up https://www.wolframalpha.com/input?i=1517818840.967415409394235163+*+%2870.738349405152439867+-+70.710678118654752440%29
			// expectedTokenOut: 8396.714105 rounded down https://www.wolframalpha.com/input?i=%281517818840.967415409394235163+*+%2870.738349405152439867+-+70.710678118654752440+%29%29+%2F+%2870.710678118654752440+*+70.738349405152439867%29
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
			// liquidity: 		 1517818840.967415409394235163
			// sqrtPriceNext:    70.666662070529219856 which is 4993.777128190373086350 https://www.wolframalpha.com/input?i=%28%281517818840.967415409394235163%29%29+%2F+%28%28%281517818840.967415409394235163%29+%2F+%2870.710678118654752440%29%29+%2B+%2813370%29%29
			// expectedTokenIn:  13369.9999 rounded up https://www.wolframalpha.com/input?i=%281517818840.967415409394235163+*+%2870.710678118654752440+-+70.666662070529219856+%29%29+%2F+%2870.666662070529219856+*+70.710678118654752440%29
			// expectedTokenOut: 66808387.149 rounded down https://www.wolframalpha.com/input?i=1517818840.967415409394235163+*+%2870.710678118654752440+-+70.666662070529219856%29
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

			// Create a default CL pool
			pool := s.PrepareConcentratedPool()

			// Check the test case to see if we are swapping asset0 for asset1 or vice versa
			asset0 := pool.GetToken0()
			zeroForOne := test.param.tokenOut.Denom == asset0

			// Fund the test account with usdc and eth, then create a default position to the pool created earlier
			s.SetupIncentivizedPosition(1)

			// Retrieve pool post position set up
			pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
			s.Require().NoError(err)

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
				s.AssertEventEmitted(s.Ctx, cltypes.TypeEvtTokenSwapped, 1)

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
