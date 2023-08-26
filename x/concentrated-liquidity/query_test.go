package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/client/queryproto"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types/genesis"
)

func (s *KeeperTestSuite) TestGetTickLiquidityForFullRange() {
	defaultTick := withPoolId(defaultTick, defaultPoolId)

	tests := []struct {
		name        string
		presetTicks []genesis.FullTick

		expectedLiquidityDepthForRange []queryproto.LiquidityDepthWithRange
	}{
		{
			name: "one full range position, testing range in between",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
			},
			expectedLiquidityDepthForRange: []queryproto.LiquidityDepthWithRange{
				{
					LiquidityAmount: sdk.NewDec(10),
					LowerTick:       DefaultMinTick,
					UpperTick:       DefaultMaxTick,
				},
			},
		},
		{
			name: "one ranged position, testing range with greater range than initialized ticks",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, 5, sdk.NewDec(-10)),
			},
			expectedLiquidityDepthForRange: []queryproto.LiquidityDepthWithRange{
				{
					LiquidityAmount: sdk.NewDec(10),
					LowerTick:       DefaultMinTick,
					UpperTick:       5,
				},
			},
		},
		//  	   	10 ----------------- 30
		//  -20 ------------- 20
		{
			name: "two ranged positions, testing overlapping positions",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -20, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, 20, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(50)),
				withLiquidityNetandTickIndex(defaultTick, 30, sdk.NewDec(-50)),
			},
			expectedLiquidityDepthForRange: []queryproto.LiquidityDepthWithRange{
				{
					LiquidityAmount: sdk.NewDec(10),
					LowerTick:       -20,
					UpperTick:       10,
				},
				{
					LiquidityAmount: sdk.NewDec(60),
					LowerTick:       10,
					UpperTick:       20,
				},
				{
					LiquidityAmount: sdk.NewDec(50),
					LowerTick:       20,
					UpperTick:       30,
				},
			},
		},
		//  	   	       10 ----------------- 30
		//  min tick --------------------------------------max tick
		{
			name: "one full ranged position, one narrow position",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(50)),
				withLiquidityNetandTickIndex(defaultTick, 30, sdk.NewDec(-50)),
			},
			expectedLiquidityDepthForRange: []queryproto.LiquidityDepthWithRange{
				{
					LiquidityAmount: sdk.NewDec(10),
					LowerTick:       DefaultMinTick,
					UpperTick:       10,
				},
				{
					LiquidityAmount: sdk.NewDec(60),
					LowerTick:       10,
					UpperTick:       30,
				},
				{
					LiquidityAmount: sdk.NewDec(10),
					LowerTick:       30,
					UpperTick:       DefaultMaxTick,
				},
			},
		},
		//              11--13
		//         10 ----------------- 30
		//  -20 ------------- 20
		{
			name: "three ranged positions, testing overlapping positions",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -20, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, 20, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(50)),
				withLiquidityNetandTickIndex(defaultTick, 30, sdk.NewDec(-50)),
				withLiquidityNetandTickIndex(defaultTick, 11, sdk.NewDec(100)),
				withLiquidityNetandTickIndex(defaultTick, 13, sdk.NewDec(-100)),
			},
			expectedLiquidityDepthForRange: []queryproto.LiquidityDepthWithRange{
				{
					LiquidityAmount: sdk.NewDec(10),
					LowerTick:       -20,
					UpperTick:       10,
				},
				{
					LiquidityAmount: sdk.NewDec(60),
					LowerTick:       10,
					UpperTick:       11,
				},
				{
					LiquidityAmount: sdk.NewDec(160),
					LowerTick:       11,
					UpperTick:       13,
				},
				{
					LiquidityAmount: sdk.NewDec(60),
					LowerTick:       13,
					UpperTick:       20,
				},
				{
					LiquidityAmount: sdk.NewDec(50),
					LowerTick:       20,
					UpperTick:       30,
				},
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()

			// Create a default CL pool
			s.PrepareConcentratedPool()
			for _, tick := range test.presetTicks {
				s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, tick.PoolId, tick.TickIndex, &tick.Info)
			}

			liquidityForRange, err := s.App.ConcentratedLiquidityKeeper.GetTickLiquidityForFullRange(s.Ctx, defaultPoolId)
			s.Require().NoError(err)
			s.Require().Equal(liquidityForRange, test.expectedLiquidityDepthForRange)
		})
	}
}

func (s *KeeperTestSuite) TestGetTickLiquidityNetInDirection() {
	defaultTick := withPoolId(defaultTick, defaultPoolId)

	tests := []struct {
		name        string
		presetTicks []genesis.FullTick

		// testing params
		poolId          uint64
		tokenIn         string
		currentPoolTick int64
		startTick       sdk.Int
		boundTick       sdk.Int

		// expected values
		expectedLiquidityDepths []queryproto.TickLiquidityNet
		expectedError           bool
	}{
		{
			name: "one full range position, zero for one true",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
			},

			poolId:    defaultPoolId,
			tokenIn:   ETH,
			boundTick: sdk.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(10),
					TickIndex:    DefaultMinTick,
				},
			},
		},
		{
			name: "one full range position, zero for one false",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
			},

			poolId:    defaultPoolId,
			tokenIn:   USDC,
			boundTick: sdk.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(-10),
					TickIndex:    DefaultMaxTick,
				},
			},
		},
		{
			name: "one full range position, one range position above current tick, zero for one true",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, 5, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:    defaultPoolId,
			tokenIn:   ETH,
			boundTick: sdk.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(10),
					TickIndex:    DefaultMinTick,
				},
			},
		},
		{
			name: "one full range position, one range position above current tick, zero for one false",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, 5, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:    defaultPoolId,
			tokenIn:   USDC,
			boundTick: sdk.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(20),
					TickIndex:    5,
				},
				{
					LiquidityNet: sdk.NewDec(-20),
					TickIndex:    10,
				},
				{
					LiquidityNet: sdk.NewDec(-10),
					TickIndex:    DefaultMaxTick,
				},
			},
		},
		{
			name: "one full range position, one range position above current tick, zero for one false, bound tick below with non-empty ticks",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:    defaultPoolId,
			tokenIn:   ETH,
			boundTick: sdk.NewInt(-15),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(20),
					TickIndex:    -10,
				},
			},
		},
		{
			name: "one ranged position, returned empty array",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:                  defaultPoolId,
			tokenIn:                 ETH,
			boundTick:               sdk.NewInt(-5),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{},
		},
		{
			name: "one full range position, one range position above current tick, zero for one false, bound tick below with non-empty ticks",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:    defaultPoolId,
			tokenIn:   USDC,
			boundTick: sdk.NewInt(10),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(-20),
					TickIndex:    10,
				},
			},
		},
		{
			name: "one full range position, two ranged positions, zero for one true",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, -5, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 5, sdk.NewDec(-20)),
				withLiquidityNetandTickIndex(defaultTick, 2, sdk.NewDec(40)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-40)),
			},

			poolId:    defaultPoolId,
			tokenIn:   ETH,
			boundTick: sdk.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(20),
					TickIndex:    -5,
				},
				{
					LiquidityNet: sdk.NewDec(10),
					TickIndex:    DefaultMinTick,
				},
			},
		},
		{
			name: "one full range position, two ranged positions, zero for one false",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, -5, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 5, sdk.NewDec(-20)),
				withLiquidityNetandTickIndex(defaultTick, 2, sdk.NewDec(40)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-40)),
			},

			poolId:    defaultPoolId,
			tokenIn:   USDC,
			boundTick: sdk.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(40),
					TickIndex:    2,
				},
				{
					LiquidityNet: sdk.NewDec(-20),
					TickIndex:    5,
				},
				{
					LiquidityNet: sdk.NewDec(-40),
					TickIndex:    10,
				},
				{
					LiquidityNet: sdk.NewDec(-10),
					TickIndex:    DefaultMaxTick,
				},
			},
		},
		{
			name: "current pool tick == start tick, zero for one",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			tokenIn:         ETH,
			currentPoolTick: 10,
			startTick:       sdk.NewInt(10),
			boundTick:       sdk.NewInt(-15),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(-20),
					TickIndex:    10,
				},
				{
					LiquidityNet: sdk.NewDec(20),
					TickIndex:    -10,
				},
			},
		},
		{
			name: "current pool tick != start tick, zero for one",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			tokenIn:         ETH,
			currentPoolTick: 21,
			startTick:       sdk.NewInt(10),
			boundTick:       sdk.NewInt(-15),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{

					LiquidityNet: sdk.NewDec(-20),
					TickIndex:    10,
				},
				{

					LiquidityNet: sdk.NewDec(20),
					TickIndex:    -10,
				},
			},
		},
		{
			name: "current pool tick == start tick, one for zero",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			tokenIn:         USDC,
			currentPoolTick: 5,
			startTick:       sdk.NewInt(5),
			boundTick:       sdk.NewInt(15),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(-20),
					TickIndex:    10,
				},
			},
		},
		{
			name: "current pool tick != start tick, one for zero",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			tokenIn:         USDC,
			currentPoolTick: -50,
			startTick:       sdk.NewInt(5),
			boundTick:       sdk.NewInt(15),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: sdk.NewDec(-20),
					TickIndex:    10,
				},
			},
		},

		// error cases
		{
			name: "error: invalid pool id",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
			},

			poolId:        5,
			tokenIn:       "invalid_token",
			boundTick:     sdk.NewInt(-5),
			expectedError: true,
		},
		{
			name: "error: invalid token in",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
			},

			poolId:        defaultPoolId,
			tokenIn:       "invalid_token",
			boundTick:     sdk.NewInt(-5),
			expectedError: true,
		},
		{
			name: "error: wrong direction of bound ticks",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			poolId:        defaultPoolId,
			tokenIn:       USDC,
			boundTick:     sdk.NewInt(-5),
			expectedError: true,
		},
		{
			name: "error: bound tick is greater than max tick",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
			},

			poolId:        defaultPoolId,
			tokenIn:       USDC,
			boundTick:     sdk.NewInt(DefaultMaxTick + 1),
			expectedError: true,
		},
		{
			name: "error: bound tick is greater than min tick",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, sdk.NewDec(10)),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, sdk.NewDec(-10)),
			},

			poolId:        defaultPoolId,
			tokenIn:       ETH,
			boundTick:     sdk.NewInt(DefaultMinCurrentTick - 1),
			expectedError: true,
		},
		{
			name: "start tick is in invalid range relative to current pool tick, zero for one",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			tokenIn:         ETH,
			currentPoolTick: 10,
			startTick:       sdk.NewInt(21),
			boundTick:       sdk.NewInt(-15),
			expectedError:   true,
		},
		{
			name: "start tick is in invalid range relative to current pool tick, one for zero",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, sdk.NewDec(20)),
				withLiquidityNetandTickIndex(defaultTick, 10, sdk.NewDec(-20)),
			},

			tokenIn:         USDC,
			currentPoolTick: 5,
			startTick:       sdk.NewInt(-50),
			boundTick:       sdk.NewInt(15),
			expectedError:   true,
		},
	}

	for _, test := range tests {
		test := test
		if test.poolId == 0 {
			test.poolId = defaultPoolId
		}
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()

			// Create a default CL pool
			pool := s.PrepareConcentratedPool()
			for _, tick := range test.presetTicks {
				s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, tick.PoolId, tick.TickIndex, &tick.Info)
			}

			// Force initialize current sqrt price to 1.
			// Normally, initialized during position creation.
			// We only initialize ticks in this test for simplicity.
			curPrice := sdk.OneDec()
			// TODO: consider adding tests for GetTickLiquidityNetInDirection
			// with tick spacing > 1, requiring price to tick conversion with rounding.
			curTick, err := math.CalculateSqrtPriceToTick(osmomath.BigDecFromSDKDec(osmomath.MustMonotonicSqrt(curPrice)))
			s.Require().NoError(err)
			var curSqrtPrice osmomath.BigDec = osmomath.OneDec()
			if test.currentPoolTick > 0 {
				_, sqrtPrice, err := math.TickToSqrtPrice(test.currentPoolTick)
				s.Require().NoError(err)

				curTick = test.currentPoolTick
				curSqrtPrice = osmomath.BigDecFromSDKDec(sqrtPrice)
			}
			pool.SetCurrentSqrtPrice(curSqrtPrice)
			pool.SetCurrentTick(curTick)

			err = s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, pool)
			s.Require().NoError(err)

			// system under test
			liquidityForRange, err := s.App.ConcentratedLiquidityKeeper.GetTickLiquidityNetInDirection(s.Ctx, test.poolId, test.tokenIn, test.startTick, test.boundTick)
			if test.expectedError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(test.expectedLiquidityDepths, liquidityForRange)
		})
	}
}
