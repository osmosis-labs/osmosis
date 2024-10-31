package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types/genesis"
)

var tenDec = osmomath.NewDec(10)
var negTenDec = osmomath.NewDec(-10)
var twentyDec = osmomath.NewDec(20)
var negTwentyDec = osmomath.NewDec(-20)
var fortyDec = osmomath.NewDec(40)
var negFortyDec = osmomath.NewDec(-40)
var fiftyDec = osmomath.NewDec(50)
var negFiftyDec = osmomath.NewDec(-50)
var sixtyDec = osmomath.NewDec(60)
var hundredDec = osmomath.NewDec(100)

// This test validates GetTickLiquidityForFullRange query by force-setting the tick and their net liquidity
// values as well as the current pool tick.
// It then checks if the returned range is as expected.
func (s *KeeperTestSuite) TestGetTickLiquidityForFullRange() {
	defaultTick := withPoolId(defaultTick, defaultPoolId)
	type testcase struct {
		name             string
		presetTicks      []genesis.FullTick
		currentTickIndex int64

		expectedLiquidityDepthForRange []queryproto.LiquidityDepthWithRange

		// Current tick is always 0 so must be pointing to the appropriate bucket
		// within which tick 0 is contained.
		expectedCurrentBucketIndex int64
	}

	defaultUpperTick := int64(5)

	defaultCase := testcase{
		name: "one ranged position, testing range with greater range than initialized ticks",
		presetTicks: []genesis.FullTick{
			withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
			withLiquidityNetandTickIndex(defaultTick, defaultUpperTick, negTenDec),
		},
		expectedLiquidityDepthForRange: []queryproto.LiquidityDepthWithRange{
			{
				LiquidityAmount: tenDec,
				LowerTick:       DefaultMinTick,
				UpperTick:       defaultUpperTick,
			},
		},
	}

	withCurrentTickAndBucketIndex := func(desiredCurrentTick, expectedCurrentBucketIndex int64, appendNameSuffix string) testcase {
		// deep copy default case
		test := testcase{
			name:                           defaultCase.name,
			presetTicks:                    make([]genesis.FullTick, len(defaultCase.presetTicks)),
			expectedLiquidityDepthForRange: make([]queryproto.LiquidityDepthWithRange, len(defaultCase.expectedLiquidityDepthForRange)),
		}
		copy(test.presetTicks, defaultCase.presetTicks)
		copy(test.expectedLiquidityDepthForRange, defaultCase.expectedLiquidityDepthForRange)

		test.name = test.name + " " + appendNameSuffix
		test.currentTickIndex = desiredCurrentTick
		test.expectedCurrentBucketIndex = expectedCurrentBucketIndex
		return test
	}

	tests := []testcase{
		{
			name: "one full range position, testing range in between",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
			},
			currentTickIndex: 100,

			expectedLiquidityDepthForRange: []queryproto.LiquidityDepthWithRange{
				{
					LiquidityAmount: tenDec,
					LowerTick:       DefaultMinTick,
					UpperTick:       DefaultMaxTick,
				},
			},
			expectedCurrentBucketIndex: 0,
		},
		withCurrentTickAndBucketIndex(DefaultMinTick-1, -1, "current tick below min tick"),
		withCurrentTickAndBucketIndex(DefaultMinTick, 0, "current tick at min tick"),
		withCurrentTickAndBucketIndex(defaultUpperTick-1, 0, "current tick one below max"),
		// Corresponds to length since the current tick is at the max tick
		withCurrentTickAndBucketIndex(defaultUpperTick, int64(len(defaultCase.expectedLiquidityDepthForRange)), "current tick at max"),
		// Corresponds to length since the current tick is above the max tick
		withCurrentTickAndBucketIndex(defaultUpperTick+1, int64(len(defaultCase.expectedLiquidityDepthForRange)), "current tick above max"),
		//  	   	10 ----------------- 30
		//  -20 ------------- 20
		{
			name: "two ranged positions, testing overlapping positions",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -20, tenDec),
				withLiquidityNetandTickIndex(defaultTick, 20, negTenDec),
				withLiquidityNetandTickIndex(defaultTick, 10, fiftyDec),
				withLiquidityNetandTickIndex(defaultTick, 30, negFiftyDec),
			},
			currentTickIndex: 15,

			expectedLiquidityDepthForRange: []queryproto.LiquidityDepthWithRange{
				{
					LiquidityAmount: tenDec,
					LowerTick:       -20,
					UpperTick:       10,
				},
				{
					LiquidityAmount: sixtyDec,
					LowerTick:       10,
					UpperTick:       20,
				},
				{
					LiquidityAmount: fiftyDec,
					LowerTick:       20,
					UpperTick:       30,
				},
			},

			expectedCurrentBucketIndex: 1,
		},
		//  	   	       10 ----------------- 30
		//  min tick --------------------------------------max tick
		{
			name: "one full ranged position, one narrow position",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
				withLiquidityNetandTickIndex(defaultTick, 10, fiftyDec),
				withLiquidityNetandTickIndex(defaultTick, 30, negFiftyDec),
			},
			currentTickIndex: 30,

			expectedLiquidityDepthForRange: []queryproto.LiquidityDepthWithRange{
				{
					LiquidityAmount: tenDec,
					LowerTick:       DefaultMinTick,
					UpperTick:       10,
				},
				{
					LiquidityAmount: sixtyDec,
					LowerTick:       10,
					UpperTick:       30,
				},
				{
					LiquidityAmount: tenDec,
					LowerTick:       30,
					UpperTick:       DefaultMaxTick,
				},
			},

			expectedCurrentBucketIndex: 2,
		},
		//              11--13
		//         10 ----------------- 30
		//  -20 ------------- 20
		{
			name: "three ranged positions, testing overlapping positions",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -20, tenDec),
				withLiquidityNetandTickIndex(defaultTick, 20, negTenDec),
				withLiquidityNetandTickIndex(defaultTick, 10, fiftyDec),
				withLiquidityNetandTickIndex(defaultTick, 30, negFiftyDec),
				withLiquidityNetandTickIndex(defaultTick, 11, hundredDec),
				withLiquidityNetandTickIndex(defaultTick, 13, osmomath.NewDec(-100)),
			},
			currentTickIndex: 30,

			expectedLiquidityDepthForRange: []queryproto.LiquidityDepthWithRange{
				{
					LiquidityAmount: tenDec,
					LowerTick:       -20,
					UpperTick:       10,
				},
				{
					LiquidityAmount: sixtyDec,
					LowerTick:       10,
					UpperTick:       11,
				},
				{
					LiquidityAmount: osmomath.NewDec(160),
					LowerTick:       11,
					UpperTick:       13,
				},
				{
					LiquidityAmount: sixtyDec,
					LowerTick:       13,
					UpperTick:       20,
				},
				{
					LiquidityAmount: fiftyDec,
					LowerTick:       20,
					UpperTick:       30,
				},
			},

			// Equals to length since current tick is above max tick
			expectedCurrentBucketIndex: 5,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()

			// Create a default CL pool
			concentratedPool := s.PrepareConcentratedPool()
			// Set current tick to the configured value
			concentratedPool.SetCurrentTick(test.currentTickIndex)

			currentTickLiquidity := osmomath.ZeroDec()
			for i, tick := range test.presetTicks {
				if i > 0 {
					lowerTick := test.presetTicks[i-1].TickIndex
					upperTick := tick.TickIndex

					// Set current liquidity corresponding to the appropriate bucket
					if concentratedPool.IsCurrentTickInRange(lowerTick, upperTick) {
						concentratedPool.UpdateLiquidity(currentTickLiquidity)
					}
				}

				s.App.ConcentratedLiquidityKeeper.SetTickInfo(s.Ctx, tick.PoolId, tick.TickIndex, &tick.Info)

				currentTickLiquidity = currentTickLiquidity.Add(tick.Info.LiquidityNet)
			}

			// Write updates pool to state
			err := s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, concentratedPool)
			s.Require().NoError(err)

			liquidityForRange, currentBucketIndex, err := s.App.ConcentratedLiquidityKeeper.GetTickLiquidityForFullRange(s.Ctx, defaultPoolId)
			s.Require().NoError(err)
			s.Require().Equal(liquidityForRange, test.expectedLiquidityDepthForRange)

			s.Require().Equal(test.expectedCurrentBucketIndex, currentBucketIndex)
		})
	}
}

// Tests GetTickLiquidityForFullRange by creating a position as opposed to directly
// setting tick net liquidity values
func (s *KeeperTestSuite) TestGetTickLiquidityForFullRange_CreatePosition() {
	// Init suite for each test.
	s.SetupTest()

	var (
		positionOneLowerTick   = int64(-500000)
		posititionOneUpperTick = int64(500000)

		positionTwoLowerTick = int64(-100000)
		positionTwoUpperTick = int64(1250000)

		defaultTokenAmount = osmomath.NewInt(1000000000000000000)
		defaultToken0      = sdk.NewCoin(ETH, defaultTokenAmount)
		defaultToken1      = sdk.NewCoin(USDC, defaultTokenAmount.MulRaw(5))
		defaultCoins       = sdk.NewCoins(defaultToken0, defaultToken1)

		expectedLiquidityDepthForRange = []queryproto.LiquidityDepthWithRange{
			{
				// This gets initializes after position creation
				LiquidityAmount: osmomath.ZeroDec(),
				LowerTick:       positionOneLowerTick,
				UpperTick:       positionTwoLowerTick,
			},
			{
				// This gets initializes after position creation
				LiquidityAmount: osmomath.ZeroDec(),
				LowerTick:       positionTwoLowerTick,
				UpperTick:       posititionOneUpperTick,
			},
			{
				// This gets initializes after position creation
				LiquidityAmount: osmomath.ZeroDec(),
				LowerTick:       posititionOneUpperTick,
				UpperTick:       positionTwoUpperTick,
			},
		}

		// points to the bucket between positionTwo lower tick and positionOne upper tick
		expectedCurrentBucketIndex = int64(3)
	)

	// Create a default CL pool
	concentratedPool := s.PrepareConcentratedPool()

	// Fund account with enough tokens for both positions
	s.FundAcc(s.TestAccs[0], defaultCoins.Add(defaultCoins...))

	// Create first position
	positionOneData, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, concentratedPool.GetId(), s.TestAccs[0], defaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), positionOneLowerTick, posititionOneUpperTick)
	s.Require().NoError(err)

	// Create second position
	positionTwoData, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, concentratedPool.GetId(), s.TestAccs[0], defaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), positionTwoLowerTick, positionTwoUpperTick)
	s.Require().NoError(err)

	s.Require().Len(expectedLiquidityDepthForRange, 3)
	// We take CreatePosition as correct since it is tested for correctness at a lower level
	// of abstraction
	expectedLiquidityDepthForRange[0].LiquidityAmount = positionOneData.Liquidity
	expectedLiquidityDepthForRange[1].LiquidityAmount = positionOneData.Liquidity.Add(positionTwoData.Liquidity)
	expectedLiquidityDepthForRange[2].LiquidityAmount = positionTwoData.Liquidity

	liquidityForRange, currentBucketIndex, err := s.App.ConcentratedLiquidityKeeper.GetTickLiquidityForFullRange(s.Ctx, defaultPoolId)
	s.Require().NoError(err)
	s.Require().Equal(liquidityForRange, expectedLiquidityDepthForRange)

	s.Require().Equal(expectedCurrentBucketIndex, currentBucketIndex)
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
		startTick       osmomath.Int
		boundTick       osmomath.Int

		// expected values
		expectedLiquidityDepths []queryproto.TickLiquidityNet
		expectedError           bool
	}{
		{
			name: "one full range position, zero for one true",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
			},

			poolId:    defaultPoolId,
			tokenIn:   ETH,
			boundTick: osmomath.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: tenDec,
					TickIndex:    DefaultMinTick,
				},
			},
		},
		{
			name: "one full range position, zero for one false",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
			},

			poolId:    defaultPoolId,
			tokenIn:   USDC,
			boundTick: osmomath.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: negTenDec,
					TickIndex:    DefaultMaxTick,
				},
			},
		},
		{
			name: "one full range position, one range position above current tick, zero for one true",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
				withLiquidityNetandTickIndex(defaultTick, 5, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negTwentyDec),
			},

			poolId:    defaultPoolId,
			tokenIn:   ETH,
			boundTick: osmomath.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: tenDec,
					TickIndex:    DefaultMinTick,
				},
			},
		},
		{
			name: "one full range position, one range position above current tick, zero for one false",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
				withLiquidityNetandTickIndex(defaultTick, 5, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negTwentyDec),
			},

			poolId:    defaultPoolId,
			tokenIn:   USDC,
			boundTick: osmomath.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: twentyDec,
					TickIndex:    5,
				},
				{
					LiquidityNet: negTwentyDec,
					TickIndex:    10,
				},
				{
					LiquidityNet: negTenDec,
					TickIndex:    DefaultMaxTick,
				},
			},
		},
		{
			name: "one full range position, one range position above current tick, zero for one false, bound tick below with non-empty ticks",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
				withLiquidityNetandTickIndex(defaultTick, -10, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negTwentyDec),
			},

			poolId:    defaultPoolId,
			tokenIn:   ETH,
			boundTick: osmomath.NewInt(-15),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: twentyDec,
					TickIndex:    -10,
				},
			},
		},
		{
			name: "one ranged position, returned empty array",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negTwentyDec),
			},

			poolId:                  defaultPoolId,
			tokenIn:                 ETH,
			boundTick:               osmomath.NewInt(-5),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{},
		},
		{
			name: "one full range position, one range position above current tick, zero for one false, bound tick below with non-empty ticks",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
				withLiquidityNetandTickIndex(defaultTick, -10, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negTwentyDec),
			},

			poolId:    defaultPoolId,
			tokenIn:   USDC,
			boundTick: osmomath.NewInt(10),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: negTwentyDec,
					TickIndex:    10,
				},
			},
		},
		{
			name: "one full range position, two ranged positions, zero for one true",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
				withLiquidityNetandTickIndex(defaultTick, -5, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 5, negTwentyDec),
				withLiquidityNetandTickIndex(defaultTick, 2, fortyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negFortyDec),
			},

			poolId:    defaultPoolId,
			tokenIn:   ETH,
			boundTick: osmomath.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: twentyDec,
					TickIndex:    -5,
				},
				{
					LiquidityNet: tenDec,
					TickIndex:    DefaultMinTick,
				},
			},
		},
		{
			name: "one full range position, two ranged positions, zero for one false",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
				withLiquidityNetandTickIndex(defaultTick, -5, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 5, negTwentyDec),
				withLiquidityNetandTickIndex(defaultTick, 2, fortyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negFortyDec),
			},

			poolId:    defaultPoolId,
			tokenIn:   USDC,
			boundTick: osmomath.Int{},
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: fortyDec,
					TickIndex:    2,
				},
				{
					LiquidityNet: negTwentyDec,
					TickIndex:    5,
				},
				{
					LiquidityNet: negFortyDec,
					TickIndex:    10,
				},
				{
					LiquidityNet: negTenDec,
					TickIndex:    DefaultMaxTick,
				},
			},
		},
		{
			name: "current pool tick == start tick, zero for one",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negTwentyDec),
			},

			tokenIn:         ETH,
			currentPoolTick: 10,
			startTick:       osmomath.NewInt(10),
			boundTick:       osmomath.NewInt(-15),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: negTwentyDec,
					TickIndex:    10,
				},
				{
					LiquidityNet: twentyDec,
					TickIndex:    -10,
				},
			},
		},
		{
			name: "current pool tick != start tick, zero for one",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negTwentyDec),
			},

			tokenIn:         ETH,
			currentPoolTick: 21,
			startTick:       osmomath.NewInt(10),
			boundTick:       osmomath.NewInt(-15),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{

					LiquidityNet: negTwentyDec,
					TickIndex:    10,
				},
				{

					LiquidityNet: twentyDec,
					TickIndex:    -10,
				},
			},
		},
		{
			name: "current pool tick == start tick, one for zero",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negTwentyDec),
			},

			tokenIn:         USDC,
			currentPoolTick: 5,
			startTick:       osmomath.NewInt(5),
			boundTick:       osmomath.NewInt(15),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: negTwentyDec,
					TickIndex:    10,
				},
			},
		},
		{
			name: "current pool tick != start tick, one for zero",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negTwentyDec),
			},

			tokenIn:         USDC,
			currentPoolTick: -50,
			startTick:       osmomath.NewInt(5),
			boundTick:       osmomath.NewInt(15),
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: negTwentyDec,
					TickIndex:    10,
				},
			},
		},

		// error cases
		{
			name: "error: invalid pool id",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
			},

			poolId:        5,
			tokenIn:       "invalid_token",
			boundTick:     osmomath.NewInt(-5),
			expectedError: true,
		},
		{
			name: "error: invalid token in",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
			},

			poolId:        defaultPoolId,
			tokenIn:       "invalid_token",
			boundTick:     osmomath.NewInt(-5),
			expectedError: true,
		},
		{
			name: "error: wrong direction of bound ticks",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
				withLiquidityNetandTickIndex(defaultTick, -10, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negTwentyDec),
			},

			poolId:        defaultPoolId,
			tokenIn:       USDC,
			boundTick:     osmomath.NewInt(-5),
			expectedError: true,
		},
		{
			name: "error: bound tick is greater than max tick",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
			},

			poolId:        defaultPoolId,
			tokenIn:       USDC,
			boundTick:     osmomath.NewInt(DefaultMaxTick + 1),
			expectedError: true,
		},
		{
			name: "error: bound tick is greater than min tick",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
			},

			poolId:        defaultPoolId,
			tokenIn:       ETH,
			boundTick:     osmomath.NewInt(DefaultMinCurrentTick - 1),
			expectedError: true,
		},
		{
			name: "start tick is in invalid range relative to current pool tick, zero for one",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negTwentyDec),
			},

			tokenIn:         ETH,
			currentPoolTick: 10,
			startTick:       osmomath.NewInt(21),
			boundTick:       osmomath.NewInt(-15),
			expectedError:   true,
		},
		{
			name: "start tick is in invalid range relative to current pool tick, one for zero",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negTwentyDec),
			},

			tokenIn:         USDC,
			currentPoolTick: 5,
			startTick:       osmomath.NewInt(-50),
			boundTick:       osmomath.NewInt(15),
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
			curPrice := osmomath.OneDec()
			// TODO: consider adding tests for GetTickLiquidityNetInDirection
			// with tick spacing > 1, requiring price to tick conversion with rounding.
			curTick, err := math.CalculateSqrtPriceToTick(osmomath.BigDecFromDec(osmomath.MustMonotonicSqrt(curPrice)))
			s.Require().NoError(err)
			var curSqrtPrice osmomath.BigDec = osmomath.OneBigDec()
			if test.currentPoolTick > 0 {
				sqrtPrice, err := math.TickToSqrtPrice(test.currentPoolTick)
				s.Require().NoError(err)

				curTick = test.currentPoolTick
				curSqrtPrice = sqrtPrice
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

func (s *KeeperTestSuite) TestGetNumNextInitializedTicks() {
	defaultTick := withPoolId(defaultTick, defaultPoolId)

	tests := []struct {
		name        string
		presetTicks []genesis.FullTick

		// testing params
		poolId                       uint64
		tokenInDenom                 string
		currentPoolTick              int64
		numberOfNextInitializedTicks uint64

		// expected values
		expectedLiquidityDepths []queryproto.TickLiquidityNet
		expectedError           bool
	}{
		{
			name: "one full range position, zero for one true",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
			},

			poolId:                       defaultPoolId,
			tokenInDenom:                 ETH,
			numberOfNextInitializedTicks: 1,
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: tenDec,
					TickIndex:    DefaultMinTick,
				},
			},
		},
		{
			name: "one full range position, zero for one false",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
			},

			poolId:                       defaultPoolId,
			tokenInDenom:                 USDC,
			numberOfNextInitializedTicks: 1,
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: negTenDec,
					TickIndex:    DefaultMaxTick,
				},
			},
		},
		{
			name: "one full range position, one range position above current tick, zero for one true",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
				withLiquidityNetandTickIndex(defaultTick, 5, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negTwentyDec),
			},

			poolId:                       defaultPoolId,
			tokenInDenom:                 ETH,
			numberOfNextInitializedTicks: 1,
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: tenDec,
					TickIndex:    DefaultMinTick,
				},
			},
		},
		{
			name: "one full range position, one range position above current tick, zero for one false",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
				withLiquidityNetandTickIndex(defaultTick, 5, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negTwentyDec),
			},

			poolId:                       defaultPoolId,
			tokenInDenom:                 USDC,
			numberOfNextInitializedTicks: 3,
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: twentyDec,
					TickIndex:    5,
				},
				{
					LiquidityNet: negTwentyDec,
					TickIndex:    10,
				},
				{
					LiquidityNet: negTenDec,
					TickIndex:    DefaultMaxTick,
				},
			},
		},
		{
			name: "one full range position, one range position above current tick, zero for one false, bound num ticks below with non-empty ticks",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
				withLiquidityNetandTickIndex(defaultTick, -10, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negTwentyDec),
			},

			poolId:                       defaultPoolId,
			tokenInDenom:                 ETH,
			numberOfNextInitializedTicks: 1,
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: twentyDec,
					TickIndex:    -10,
				},
			},
		},
		{
			name: "one ranged position, returned empty array",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, -10, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negTwentyDec),
			},

			poolId:                       defaultPoolId,
			tokenInDenom:                 ETH,
			numberOfNextInitializedTicks: 0,
			expectedLiquidityDepths:      []queryproto.TickLiquidityNet{},
		},
		{
			name: "one full range position, one range position above current tick, zero for one false, bound num ticks below with non-empty ticks",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
				withLiquidityNetandTickIndex(defaultTick, -10, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negTwentyDec),
			},

			poolId:                       defaultPoolId,
			tokenInDenom:                 USDC,
			numberOfNextInitializedTicks: 1,
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: negTwentyDec,
					TickIndex:    10,
				},
			},
		},
		{
			name: "one full range position, two ranged positions, zero for one true",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
				withLiquidityNetandTickIndex(defaultTick, -5, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 5, negTwentyDec),
				withLiquidityNetandTickIndex(defaultTick, 2, fortyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negFortyDec),
			},

			poolId:                       defaultPoolId,
			tokenInDenom:                 ETH,
			numberOfNextInitializedTicks: 2,
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: twentyDec,
					TickIndex:    -5,
				},
				{
					LiquidityNet: tenDec,
					TickIndex:    DefaultMinTick,
				},
			},
		},
		{
			name: "one full range position, two ranged positions, zero for one false",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
				withLiquidityNetandTickIndex(defaultTick, -5, twentyDec),
				withLiquidityNetandTickIndex(defaultTick, 5, negTwentyDec),
				withLiquidityNetandTickIndex(defaultTick, 2, fortyDec),
				withLiquidityNetandTickIndex(defaultTick, 10, negFortyDec),
			},

			poolId:                       defaultPoolId,
			tokenInDenom:                 USDC,
			numberOfNextInitializedTicks: 4,
			expectedLiquidityDepths: []queryproto.TickLiquidityNet{
				{
					LiquidityNet: fortyDec,
					TickIndex:    2,
				},
				{
					LiquidityNet: negTwentyDec,
					TickIndex:    5,
				},
				{
					LiquidityNet: negFortyDec,
					TickIndex:    10,
				},
				{
					LiquidityNet: negTenDec,
					TickIndex:    DefaultMaxTick,
				},
			},
		},

		// error cases
		{
			name: "error: invalid pool id",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
			},

			poolId:                       5,
			tokenInDenom:                 "invalid_token",
			numberOfNextInitializedTicks: 1,
			expectedError:                true,
		},
		{
			name: "error: invalid token in",
			presetTicks: []genesis.FullTick{
				withLiquidityNetandTickIndex(defaultTick, DefaultMinTick, tenDec),
				withLiquidityNetandTickIndex(defaultTick, DefaultMaxTick, negTenDec),
			},

			poolId:                       defaultPoolId,
			tokenInDenom:                 "invalid_token",
			numberOfNextInitializedTicks: 1,
			expectedError:                true,
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
			curPrice := osmomath.OneDec()
			curTick, err := math.CalculateSqrtPriceToTick(osmomath.BigDecFromDec(osmomath.MustMonotonicSqrt(curPrice)))
			s.Require().NoError(err)
			var curSqrtPrice osmomath.BigDec = osmomath.OneBigDec()
			if test.currentPoolTick > 0 {
				sqrtPrice, err := math.TickToSqrtPrice(test.currentPoolTick)
				s.Require().NoError(err)

				curTick = test.currentPoolTick
				curSqrtPrice = sqrtPrice
			}
			pool.SetCurrentSqrtPrice(curSqrtPrice)
			pool.SetCurrentTick(curTick)

			err = s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, pool)
			s.Require().NoError(err)

			// system under test
			liquidityForRange, err := s.App.ConcentratedLiquidityKeeper.GetNumNextInitializedTicks(s.Ctx, test.poolId, test.numberOfNextInitializedTicks, test.tokenInDenom)
			if test.expectedError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(test.expectedLiquidityDepths, liquidityForRange)
		})
	}
}
