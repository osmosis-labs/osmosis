package concentrated_liquidity_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/math"
	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

const (
	NoUSDCExpected = ""
	NoETHExpected  = ""
)

// fields used to identify a fee position.
type positionIdentifiers struct {
	poolId         uint64
	owner          sdk.AccAddress
	lowerTick      int64
	upperTick      int64
	freezeDuration time.Duration
}

var (
	oneEth      = sdk.NewDecCoin(ETH, sdk.OneInt())
	oneEthCoins = sdk.NewDecCoins(oneEth)
	onlyUSDC    = [][]string{{USDC}, {USDC}, {USDC}, {USDC}}
	onlyETH     = [][]string{{ETH}, {ETH}, {ETH}, {ETH}}
)

func (s *KeeperTestSuite) TestInitializeFeeAccumulatorPosition() {
	// Setup is done once so that we test
	// the relationship between test cases.
	// For example, that positions with non-zero liquidity
	// cannot be overriden.
	s.SetupTest()
	s.PrepareConcentratedPool()

	defaultAccount := s.TestAccs[0]

	var (
		defaultPoolId     = uint64(1)
		defaultPositionId = positionIdentifiers{
			defaultPoolId,
			defaultAccount,
			DefaultLowerTick,
			DefaultUpperTick,
			DefaultFreezeDuration,
		}
	)

	withOwner := func(posId positionIdentifiers, owner sdk.AccAddress) positionIdentifiers {
		posId.owner = owner
		return posId
	}

	withUpperTick := func(posId positionIdentifiers, upperTick int64) positionIdentifiers {
		posId.upperTick = upperTick
		return posId
	}

	withLowerTick := func(posId positionIdentifiers, lowerTick int64) positionIdentifiers {
		posId.lowerTick = lowerTick
		return posId
	}

	clKeeper := s.App.ConcentratedLiquidityKeeper

	type initFeeAccumTest struct {
		name       string
		positionId positionIdentifiers

		expectedPass bool
	}
	tests := []initFeeAccumTest{
		{
			name:         "first position",
			positionId:   defaultPositionId,
			expectedPass: true,
		},
		{
			name:         "second position",
			positionId:   withLowerTick(defaultPositionId, DefaultLowerTick+1),
			expectedPass: true,
		},
		{
			name:       "overriding first position - error",
			positionId: defaultPositionId,
			// Does not get overwritten by the next test case.
			expectedPass: false,
		},
		{
			name:       "overriding second position - error",
			positionId: withLowerTick(defaultPositionId, DefaultLowerTick+1),
			// Does not get overwritten by the next test case.
			expectedPass: false,
		},
		{
			name: "error: non-existing accumulator (wrong pool)",
			positionId: positionIdentifiers{
				defaultPoolId + 1, // non-existing pool
				defaultAccount,
				DefaultLowerTick,
				DefaultUpperTick,
				DefaultFreezeDuration,
			},
			expectedPass: false,
		},
		{
			name:         "existing accumulator, different owner - different position",
			positionId:   withOwner(defaultPositionId, s.TestAccs[1]),
			expectedPass: true,
		},
		{
			name:         "existing accumulator, different upper tick - different position",
			positionId:   withUpperTick(defaultPositionId, DefaultUpperTick+1),
			expectedPass: true,
		},
		{
			name:         "existing accumulator, different lower tick - different position",
			positionId:   withLowerTick(defaultPositionId, DefaultUpperTick+1),
			expectedPass: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			// system under test
			err := clKeeper.InitializeFeeAccumulatorPosition(s.Ctx, tc.positionId.poolId, tc.positionId.owner, tc.positionId.lowerTick, tc.positionId.upperTick)
			if tc.expectedPass {
				s.Require().NoError(err)

				// get fee accum and see if position size has been properly initialized
				poolFeeAccumulator, err := clKeeper.GetFeeAccumulator(s.Ctx, defaultPoolId)
				s.Require().NoError(err)

				positionKey := cl.FormatPositionAccumulatorKey(tc.positionId.poolId, tc.positionId.owner, tc.positionId.lowerTick, tc.positionId.upperTick)

				positionSize, err := poolFeeAccumulator.GetPositionSize(positionKey)
				s.Require().NoError(err)
				// position should have been properly initialized to zero
				s.Require().Equal(positionSize, sdk.ZeroDec())
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetFeeGrowthOutside() {
	type feeGrowthOutsideTest struct {
		poolSetup bool

		lowerTick                 int64
		upperTick                 int64
		currentTick               int64
		lowerTickFeeGrowthOutside sdk.DecCoins
		upperTickFeeGrowthOutside sdk.DecCoins
		globalFeeGrowth           sdk.DecCoin

		expectedFeeGrowthOutside sdk.DecCoins
		invalidTick              bool
		expectedError            bool
	}

	defaultAccumCoins := sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10)))
	defaultPoolId := uint64(1)
	defaultInitialLiquidity := sdk.OneDec()

	defaultUpperTickIndex := int64(5)
	defaultLowerTickIndex := int64(3)

	emptyUptimeTrackers := wrapUptimeTrackers(getExpectedUptimes().emptyExpectedAccumValues)

	tests := map[string]feeGrowthOutsideTest{
		// imagine single swap over entire position
		// crossing left > right and stopping above upper tick
		// In this case, only the upper tick accumulator must have
		// been updated when crossed.
		// Since we track fees accrued below a tick, upper tick is updated
		// while lower tick is zero.
		"single swap left -> right: 2 ticks, one share - current tick > upper tick": {
			poolSetup:                 true,
			lowerTick:                 0,
			upperTick:                 1,
			currentTick:               2,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"single swap left -> right: 3 ticks, two shares - current tick > upper tick": {
			poolSetup:                 true,
			lowerTick:                 0,
			upperTick:                 2,
			currentTick:               3,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"single swap left -> right: 2 ticks, one share - current tick == lower tick": {
			poolSetup:                 true,
			lowerTick:                 0,
			upperTick:                 1,
			currentTick:               0,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"single swap left -> right: 2 ticks, one share - current tick < lower tick": {
			poolSetup:                 true,
			lowerTick:                 1,
			upperTick:                 2,
			currentTick:               0,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"single swap left -> right: 2 ticks, one share - current tick == upper tick": {
			poolSetup:                 true,
			lowerTick:                 0,
			upperTick:                 1,
			currentTick:               1,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		// imagine single swap over entire position
		// crossing right > left and stopping at lower tick
		// In this case, all fees must have been accrued inside the tick
		// Since we track fees accrued below a tick, both upper and lower position
		// ticks are zero
		"single swap right -> left: 2 ticks, one share - current tick == lower tick": {
			poolSetup:                 true,
			lowerTick:                 0,
			upperTick:                 1,
			currentTick:               0,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"single swap right -> left: 2 ticks, one share - current tick == upper tick": {
			poolSetup:                 true,
			lowerTick:                 0,
			upperTick:                 1,
			currentTick:               1,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"single swap right -> left: 2 ticks, one share - current tick < lower tick": {
			poolSetup:                 true,
			lowerTick:                 0,
			upperTick:                 1,
			currentTick:               -1,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"single swap right -> left: 2 ticks, one share - current tick > lower tick": {
			poolSetup:                 true,
			lowerTick:                 -1,
			upperTick:                 1,
			currentTick:               0,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"single swap right -> left: 2 ticks, one share - current tick > upper tick": {
			poolSetup:                 true,
			lowerTick:                 -1,
			upperTick:                 1,
			currentTick:               2,
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalFeeGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedFeeGrowthOutside:  defaultAccumCoins,
			expectedError:             false,
		},
		"error: pool has not been setup": {
			poolSetup:     false,
			expectedError: true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			// if pool set up true, set up default pool
			var pool types.ConcentratedPoolExtension
			if tc.poolSetup {
				pool = s.PrepareConcentratedPool()
				currentTick := pool.GetCurrentTick().Int64()

				s.initializeTick(s.Ctx, currentTick, tc.lowerTick, defaultInitialLiquidity, tc.lowerTickFeeGrowthOutside, emptyUptimeTrackers, false)
				s.initializeTick(s.Ctx, currentTick, tc.upperTick, defaultInitialLiquidity, tc.upperTickFeeGrowthOutside, emptyUptimeTrackers, true)
				pool.SetCurrentTick(sdk.NewInt(tc.currentTick))
				s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, pool)
				err := s.App.ConcentratedLiquidityKeeper.ChargeFee(s.Ctx, validPoolId, tc.globalFeeGrowth)
				s.Require().NoError(err)
			}

			// system under test
			feeGrowthOutside, err := s.App.ConcentratedLiquidityKeeper.GetFeeGrowthOutside(s.Ctx, defaultPoolId, defaultLowerTickIndex, defaultUpperTickIndex)
			if tc.expectedError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				// check if returned fee growth outside has correct value
				s.Require().Equal(feeGrowthOutside, tc.expectedFeeGrowthOutside)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCalculateFeeGrowth() {
	defaultGeeFrowthGlobal := sdk.NewDecCoins(sdk.NewDecCoin("uosmo", sdk.NewInt(10)))
	defaultGeeFrowthOutside := sdk.NewDecCoins(sdk.NewDecCoin("uosmo", sdk.NewInt(3)))

	defaultSmallerTargetTick := int64(1)
	defaultCurrentTick := int64(2)
	defaultLargerTargetTick := int64(3)

	type calcFeeGrowthTest struct {
		isUpperTick                bool
		isCurrentTickGTETargetTick bool
		expectedFeeGrowth          sdk.DecCoins
	}

	tests := map[string]calcFeeGrowthTest{
		"current Tick is greater than the upper tick": {
			isUpperTick:                true,
			isCurrentTickGTETargetTick: false,
			expectedFeeGrowth:          defaultGeeFrowthOutside,
		},
		"current Tick is less than the upper tick": {
			isUpperTick:                true,
			isCurrentTickGTETargetTick: true,
			expectedFeeGrowth:          defaultGeeFrowthGlobal.Sub(defaultGeeFrowthOutside),
		},
		"current Tick is less than the lower tick": {
			isUpperTick:                false,
			isCurrentTickGTETargetTick: false,
			expectedFeeGrowth:          defaultGeeFrowthGlobal.Sub(defaultGeeFrowthOutside),
		},
		"current Tick is greater than the lower tick": {
			isUpperTick:                false,
			isCurrentTickGTETargetTick: true,
			expectedFeeGrowth:          defaultGeeFrowthOutside,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			var targetTick int64
			if tc.isCurrentTickGTETargetTick {
				targetTick = defaultSmallerTargetTick
			} else {
				targetTick = defaultLargerTargetTick
			}
			feeGrowth := cl.CalculateFeeGrowth(
				targetTick,
				defaultGeeFrowthOutside,
				defaultCurrentTick,
				defaultGeeFrowthGlobal,
				tc.isUpperTick,
			)
			s.Require().Equal(feeGrowth, tc.expectedFeeGrowth)
		})
	}
}

func (suite *KeeperTestSuite) TestGetInitialFeeGrowthOutsideForTick() {
	const (
		validPoolId = 1
	)

	initialPoolTickInt, err := math.PriceToTick(DefaultAmt1.ToDec().Quo(DefaultAmt0.ToDec()), DefaultExponentAtPriceOne)
	initialPoolTick := initialPoolTickInt.Int64()
	suite.Require().NoError(err)

	tests := map[string]struct {
		poolId                   uint64
		tick                     int64
		initialGlobalFeeGrowth   sdk.DecCoin
		shouldAvoidCreatingAccum bool

		expectedInitialFeeGrowthOutside sdk.DecCoins
		expectError                     error
	}{
		"current tick > tick -> fee growth global": {
			poolId:                 validPoolId,
			tick:                   initialPoolTick - 1,
			initialGlobalFeeGrowth: oneEth,

			expectedInitialFeeGrowthOutside: sdk.NewDecCoins(oneEth),
		},
		"current tick == tick -> fee growth global": {
			poolId:                 validPoolId,
			tick:                   initialPoolTick,
			initialGlobalFeeGrowth: oneEth,

			expectedInitialFeeGrowthOutside: sdk.NewDecCoins(oneEth),
		},
		"current tick < tick -> empty coins": {
			poolId:                 validPoolId,
			tick:                   initialPoolTick + 1,
			initialGlobalFeeGrowth: oneEth,

			expectedInitialFeeGrowthOutside: cl.EmptyCoins,
		},
		"pool does not exist": {
			poolId:                 validPoolId + 1,
			tick:                   initialPoolTick - 1,
			initialGlobalFeeGrowth: oneEth,

			expectError: types.PoolNotFoundError{PoolId: validPoolId + 1},
		},
		"accumulator does not exist": {
			poolId:                   validPoolId,
			tick:                     0,
			initialGlobalFeeGrowth:   oneEth,
			shouldAvoidCreatingAccum: true,

			expectError: accum.AccumDoesNotExistError{AccumName: cl.GetFeeAccumulatorName(validPoolId)},
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()
			ctx := suite.Ctx
			clKeeper := suite.App.ConcentratedLiquidityKeeper

			pool, err := clmodel.NewConcentratedLiquidityPool(validPoolId, USDC, ETH, DefaultTickSpacing, DefaultExponentAtPriceOne, DefaultZeroSwapFee)
			suite.Require().NoError(err)

			err = clKeeper.SetPool(ctx, &pool)
			suite.Require().NoError(err)

			if !tc.shouldAvoidCreatingAccum {
				err = clKeeper.CreateFeeAccumulator(ctx, validPoolId)
				suite.Require().NoError(err)

				// Setup test position to make sure that tick is initialized
				// We also set up uptime accums to ensure position creation works as intended
				err = clKeeper.CreateUptimeAccumulators(ctx, validPoolId)
				suite.Require().NoError(err)
				suite.SetupDefaultPosition(validPoolId)

				err = clKeeper.ChargeFee(ctx, validPoolId, tc.initialGlobalFeeGrowth)
				suite.Require().NoError(err)
			}

			// System under test.
			initialFeeGrowthOutside, err := clKeeper.GetInitialFeeGrowthOutsideForTick(ctx, tc.poolId, tc.tick)

			if tc.expectError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expectError)
				return
			}
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedInitialFeeGrowthOutside, initialFeeGrowthOutside)
		})
	}
}

func (suite *KeeperTestSuite) TestChargeFee() {
	// setup once at the beginning.
	suite.SetupTest()

	ctx := suite.Ctx
	clKeeper := suite.App.ConcentratedLiquidityKeeper

	// create fee accumulators with ids 1 and 2 but not 3.
	err := clKeeper.CreateFeeAccumulator(ctx, 1)
	suite.Require().NoError(err)
	err = clKeeper.CreateFeeAccumulator(ctx, 2)
	suite.Require().NoError(err)

	tests := []struct {
		name      string
		poolId    uint64
		feeUpdate sdk.DecCoin

		expectedGlobalGrowth sdk.DecCoins
		expectError          error
	}{
		{
			name:      "pool id 1 - one eth",
			poolId:    1,
			feeUpdate: oneEth,

			expectedGlobalGrowth: sdk.NewDecCoins(oneEth),
		},
		{
			name:      "pool id 1 - 2 usdc",
			poolId:    1,
			feeUpdate: sdk.NewDecCoin(USDC, sdk.NewInt(2)),

			expectedGlobalGrowth: sdk.NewDecCoins(oneEth).Add(sdk.NewDecCoin(USDC, sdk.NewInt(2))),
		},
		{
			name:      "pool id 2 - 1 usdc",
			poolId:    2,
			feeUpdate: oneEth,

			expectedGlobalGrowth: sdk.NewDecCoins(oneEth),
		},
		{
			name:      "accumulator does not exist",
			poolId:    3,
			feeUpdate: oneEth,

			expectError: accum.AccumDoesNotExistError{AccumName: cl.GetFeeAccumulatorName(3)},
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			// System under test.
			err := clKeeper.ChargeFee(ctx, tc.poolId, tc.feeUpdate)

			if tc.expectError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expectError)
				return
			}
			suite.Require().NoError(err)

			feeAcumulator, err := clKeeper.GetFeeAccumulator(ctx, tc.poolId)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedGlobalGrowth, feeAcumulator.GetValue())
		})
	}
}

func (s *KeeperTestSuite) TestQueryAndCollectFees() {
	ownerWithValidPosition := s.TestAccs[0]
	emptyUptimeTrackers := wrapUptimeTrackers(getExpectedUptimes().emptyExpectedAccumValues)

	tests := map[string]struct {
		// setup parameters.
		initialLiquidity          sdk.Dec
		lowerTickFeeGrowthOutside sdk.DecCoins
		upperTickFeeGrowthOutside sdk.DecCoins
		globalFeeGrowth           sdk.DecCoins
		currentTick               int64
		isInvalidPoolIdGiven      bool

		// inputs parameters.
		owner          sdk.AccAddress
		lowerTick      int64
		upperTick      int64
		freezeDuration time.Duration

		// expectations.
		expectedFeesClaimed sdk.Coins
		expectedError       error
	}{
		// imagine single swap over entire position
		// crossing left > right and stopping above upper tick
		// In this case, only the upper tick accumulator must have
		// been updated when crossed.
		// Since we track fees accrued below a tick, upper tick is updated
		// while lower tick is zero.
		"single swap left -> right: 2 ticks, one share, current tick > upper tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:          ownerWithValidPosition,
			lowerTick:      0,
			upperTick:      1,
			freezeDuration: DefaultFreezeDuration,

			currentTick: 2,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},
		"single swap left -> right: 3 ticks, two shares, current tick > upper tick": {
			initialLiquidity: sdk.NewDec(2),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:          ownerWithValidPosition,
			lowerTick:      0,
			upperTick:      2,
			freezeDuration: DefaultFreezeDuration,

			currentTick: 3,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(20))),
		},
		"single swap left -> right: 2 ticks, one share, current tick == upper tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:          ownerWithValidPosition,
			lowerTick:      0,
			upperTick:      1,
			freezeDuration: DefaultFreezeDuration,

			currentTick: 1,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},

		// imagine single swap over entire position
		// crossing right -> left and stopping at lower tick
		// In this case, all fees must have been accrued inside the tick
		// Since we track fees accrued below a tick, both upper and lower position
		// ticks are zero
		"single swap right -> left: 2 ticks, one share, current tick == lower tick": {
			initialLiquidity: sdk.OneDec(),

			// lower tick accumulator is not updated yet.
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:          ownerWithValidPosition,
			lowerTick:      0,
			upperTick:      1,
			freezeDuration: DefaultFreezeDuration,

			currentTick: 0,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, one share, current tick < lower tick": {
			initialLiquidity: sdk.OneDec(),

			// lower tick accumulator updated when crossed.
			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:          ownerWithValidPosition,
			lowerTick:      0,
			upperTick:      1,
			freezeDuration: DefaultFreezeDuration,

			currentTick: -1,

			expectedFeesClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},

		// imagine swap occurring outside of the position
		// As a result, lower and upper ticks are not updated.
		"swap occurs above the position, current tick > upper tick": {
			initialLiquidity: sdk.OneDec(),

			// none are updated.
			lowerTickFeeGrowthOutside: cl.EmptyCoins,
			upperTickFeeGrowthOutside: cl.EmptyCoins,

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:          ownerWithValidPosition,
			lowerTick:      0,
			upperTick:      1,
			freezeDuration: DefaultFreezeDuration,

			currentTick: 5,

			expectedFeesClaimed: sdk.NewCoins(),
		},
		"swap occurs below the position, current tick < lower tick": {
			initialLiquidity: sdk.OneDec(),

			// none are updated.
			lowerTickFeeGrowthOutside: cl.EmptyCoins,
			upperTickFeeGrowthOutside: cl.EmptyCoins,

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:          ownerWithValidPosition,
			lowerTick:      -10,
			upperTick:      -4,
			freezeDuration: DefaultFreezeDuration,

			currentTick: -13,

			expectedFeesClaimed: sdk.NewCoins(),
		},

		// error cases.

		"accumulator does not exist": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:          ownerWithValidPosition,
			lowerTick:      0,
			upperTick:      1,
			freezeDuration: DefaultFreezeDuration,

			currentTick: 2,

			isInvalidPoolIdGiven: true,
			expectedError:        accum.AccumDoesNotExistError{AccumName: "fee/2"},
		},
		"position does not exist": {
			initialLiquidity: sdk.OneDec(),

			lowerTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickFeeGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalFeeGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:          s.TestAccs[1], // different owner from the one who initialized the position.
			lowerTick:      0,
			upperTick:      1,
			freezeDuration: DefaultFreezeDuration,

			currentTick: 2,

			expectedError: cltypes.PositionNotFoundError{PoolId: 1, LowerTick: 0, UpperTick: 1},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			tc := tc
			s.SetupTest()

			validPool := s.PrepareConcentratedPool()
			validPoolId := validPool.GetId()

			s.FundAcc(validPool.GetAddress(), tc.expectedFeesClaimed)

			clKeeper := s.App.ConcentratedLiquidityKeeper
			ctx := s.Ctx

			s.initializeFeeAccumulatorPositionWithLiquidity(ctx, validPoolId, ownerWithValidPosition, tc.lowerTick, tc.upperTick, tc.initialLiquidity)

			s.initializeTick(ctx, tc.currentTick, tc.lowerTick, tc.initialLiquidity, tc.lowerTickFeeGrowthOutside, emptyUptimeTrackers, false)

			s.initializeTick(ctx, tc.currentTick, tc.upperTick, tc.initialLiquidity, tc.upperTickFeeGrowthOutside, emptyUptimeTrackers, true)

			validPool.SetCurrentTick(sdk.NewInt(tc.currentTick))
			clKeeper.SetPool(ctx, validPool)

			err := clKeeper.ChargeFee(ctx, validPoolId, tc.globalFeeGrowth[0])
			s.Require().NoError(err)

			poolBalanceBeforeCollect := s.App.BankKeeper.GetBalance(ctx, validPool.GetAddress(), ETH)
			ownerBalancerBeforeCollect := s.App.BankKeeper.GetBalance(ctx, tc.owner, ETH)

			sutPoolId := validPoolId
			if tc.isInvalidPoolIdGiven {
				sutPoolId = sutPoolId + 1
			}

			var preQueryPosition accum.Record
			positionKey := cl.FormatPositionAccumulatorKey(validPoolId, tc.owner, tc.lowerTick, tc.upperTick)

			// Note the position accumulator before the query to ensure the query in non-mutating.
			accum, err := s.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(ctx, validPoolId)
			s.Require().NoError(err)
			preQueryPosition, _ = accum.GetPosition(positionKey)

			// System under test
			feeQueryAmount, queryErr := clKeeper.QueryClaimableFees(ctx, sutPoolId, tc.owner, tc.lowerTick, tc.upperTick)

			// If the query succeeds, the position should not be updated.
			if queryErr == nil {
				accum, err := s.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(ctx, validPoolId)
				s.Require().NoError(err)
				postQueryPosition, _ := accum.GetPosition(positionKey)
				s.Require().Equal(preQueryPosition, postQueryPosition)
			}

			actualFeesClaimed, err := clKeeper.CollectFees(ctx, sutPoolId, tc.owner, tc.lowerTick, tc.upperTick)

			// Assertions.

			poolBalanceAfterCollect := s.App.BankKeeper.GetBalance(ctx, validPool.GetAddress(), ETH)
			ownerBalancerAfterCollect := s.App.BankKeeper.GetBalance(ctx, tc.owner, ETH)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Error(queryErr)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				s.Require().Equal(sdk.Coins{}, actualFeesClaimed)

				// balances are unchanged
				s.Require().Equal(poolBalanceBeforeCollect, poolBalanceAfterCollect)
				s.Require().Equal(ownerBalancerAfterCollect, ownerBalancerBeforeCollect)
				return
			}

			s.Require().NoError(err)
			s.Require().NoError(queryErr)
			s.Require().Equal(tc.expectedFeesClaimed.String(), actualFeesClaimed.String())
			s.Require().Equal(feeQueryAmount.String(), actualFeesClaimed.String())

			expectedETHAmount := tc.expectedFeesClaimed.AmountOf(ETH)
			s.Require().Equal(expectedETHAmount, poolBalanceBeforeCollect.Sub(poolBalanceAfterCollect).Amount)
			s.Require().Equal(expectedETHAmount, ownerBalancerAfterCollect.Sub(ownerBalancerBeforeCollect).Amount)
		})
	}
}

func (s *KeeperTestSuite) TestUpdateFeeAccumulatorPosition() {
	var (
		ownerOne = s.TestAccs[0]
		ownerTwo = s.TestAccs[1]
	)

	type updateFeeAccumPositionTest struct {
		poolId        uint64
		owner         sdk.AccAddress
		liquidity     sdk.Dec
		lowerTick     int64
		upperTick     int64
		expectedError error
	}

	positions := map[int]updateFeeAccumPositionTest{
		1: {
			lowerTick: DefaultLowerTick,
			upperTick: DefaultUpperTick,
		},
		2: {
			lowerTick: DefaultLowerTick + 1,
			upperTick: DefaultUpperTick,
		},
		3: {
			lowerTick: DefaultLowerTick,
			upperTick: DefaultUpperTick + 1,
		},
	}

	tests := map[string]updateFeeAccumPositionTest{
		"update position with a different pool ID": {
			poolId:    2,
			owner:     ownerOne,
			liquidity: DefaultLiquidityAmt,
			lowerTick: DefaultLowerTick,
			upperTick: DefaultUpperTick,
		},
		"update position with a different owner": {
			poolId:    1,
			owner:     ownerTwo,
			liquidity: DefaultLiquidityAmt,
			lowerTick: DefaultLowerTick,
			upperTick: DefaultUpperTick,
		},
		"update position with a different lower tick": {
			poolId:    1,
			owner:     ownerOne,
			liquidity: DefaultLiquidityAmt,
			lowerTick: DefaultLowerTick + 1,
			upperTick: DefaultUpperTick,
		},
		"update position with a different upper tick": {
			poolId:    1,
			owner:     ownerOne,
			liquidity: DefaultLiquidityAmt,
			lowerTick: DefaultLowerTick,
			upperTick: DefaultUpperTick + 1,
		},
		"err: pool does not exist": {
			poolId:        3,
			expectedError: cltypes.PoolNotFoundError{PoolId: 3},
		},
		"err: position does not exist": {
			poolId:        1,
			owner:         ownerOne,
			liquidity:     DefaultLiquidityAmt,
			lowerTick:     DefaultLowerTick - 1,
			upperTick:     DefaultUpperTick,
			expectedError: accum.NoPositionError{Name: cl.FormatPositionAccumulatorKey(1, ownerOne, DefaultLowerTick-1, DefaultUpperTick)},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			// Setup two cl pools
			poolOne := s.PrepareConcentratedPool()
			poolTwo := s.PrepareConcentratedPool()

			pools := []cltypes.ConcentratedPoolExtension{poolOne, poolTwo}
			owners := []sdk.AccAddress{ownerOne, ownerTwo}

			// Initialize three base positions in each pool for each owner (total of 12 positions)
			for _, pos := range positions {
				for _, pool := range pools {
					for _, owner := range owners {
						s.initializeFeeAccumulatorPositionWithLiquidity(s.Ctx, pool.GetId(), owner, pos.lowerTick, pos.upperTick, DefaultLiquidityAmt)
					}
				}
			}

			// System under test
			// Update one of the positions as per the test case
			err := s.App.ConcentratedLiquidityKeeper.UpdateFeeAccumulatorPosition(s.Ctx, tc.poolId, tc.owner, tc.liquidity, tc.lowerTick, tc.upperTick)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &tc.expectedError)
				return
			}

			s.Require().NoError(err)

			// Validate the test case position was updated and all other positions did not change
			for _, pos := range positions {
				for _, pool := range pools {
					for _, owner := range owners {
						liq := DefaultLiquidityAmt
						if pool.GetId() == tc.poolId && owner.Equals(tc.owner) && pos.lowerTick == tc.lowerTick && pos.upperTick == tc.upperTick {
							liq = DefaultLiquidityAmt.Mul(sdk.NewDec(2))
						}
						s.validatePositionFeeAccUpdate(s.Ctx, pool.GetId(), owner, pos.lowerTick, pos.upperTick, liq)
					}
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestPreparePositionAccumulator() {
	validPositionKey := cl.FormatPositionAccumulatorKey(defaultPoolId, s.TestAccs[0], DefaultLowerTick, DefaultUpperTick)
	invalidPositionKey := cl.FormatPositionAccumulatorKey(defaultPoolId+1, s.TestAccs[0], DefaultLowerTick, DefaultUpperTick+1)
	tests := []struct {
		name               string
		poolId             uint64
		feeGrowthOutside   sdk.DecCoins
		invalidPositionKey bool
		expectError        error
	}{
		{
			name:             "happy path",
			feeGrowthOutside: oneEthCoins,
		},
		{
			name:               "error: non existent accumulator",
			feeGrowthOutside:   oneEthCoins,
			invalidPositionKey: true,
			expectError:        accum.NoPositionError{Name: invalidPositionKey},
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			// Setup test env.
			s.SetupTest()
			clKeeper := s.App.ConcentratedLiquidityKeeper
			s.PrepareConcentratedPool()
			poolFeeAccumulator, err := clKeeper.GetFeeAccumulator(s.Ctx, defaultPoolId)
			s.Require().NoError(err)
			positionKey := validPositionKey

			// Initialize position accumulator.
			err = poolFeeAccumulator.NewPositionCustomAcc(positionKey, sdk.OneDec(), sdk.DecCoins{}, nil)
			s.Require().NoError(err)

			// Record the initial position accumulator value.
			positionPre, err := accum.GetPosition(poolFeeAccumulator, positionKey)
			s.Require().NoError(err)

			// If the test case requires an invalid position key, set it.
			if tc.invalidPositionKey {
				positionKey = invalidPositionKey
			}

			// System under test.
			err = cl.PreparePositionAccumulator(poolFeeAccumulator, positionKey, tc.feeGrowthOutside)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}
			s.Require().NoError(err)

			// Record the final position accumulator value.
			positionPost, err := accum.GetPosition(poolFeeAccumulator, positionKey)
			s.Require().NoError(err)

			// Check that the difference between the new and old position accumulator values is equal to the fee growth outside.
			positionAccumDelta := positionPost.InitAccumValue.Sub(positionPre.InitAccumValue)
			s.Require().Equal(tc.feeGrowthOutside, positionAccumDelta)
		})
	}
}

type Positions struct {
	numSwaps       int
	numAccounts    int
	numFullRange   int
	numNarrowRange int
	numConsecutive int
	numOverlapping int
}

func (s *KeeperTestSuite) TestFunctionalFees() {
	positions := Positions{
		numSwaps:       7,
		numAccounts:    5,
		numFullRange:   4,
		numNarrowRange: 3,
		numConsecutive: 2,
		numOverlapping: 1,
	}
	// Init suite.
	s.Setup()

	// Default setup only creates 3 accounts, but we need 5 for this test.
	s.TestAccs = apptesting.CreateRandomAccounts(positions.numAccounts)

	// Create a default CL pool, but with a 0.3 percent swap fee.
	clPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, DefaultExponentAtPriceOne, sdk.MustNewDecFromStr("0.003"))

	// Setup full range position across all four accounts
	for i := 0; i < positions.numFullRange; i++ {
		s.SetupFullRangePositionAcc(clPool.GetId(), s.TestAccs[i])
	}

	// Setup narrow range position across three of four accounts
	for i := 0; i < positions.numNarrowRange; i++ {
		s.SetupDefaultPositionAcc(clPool.GetId(), s.TestAccs[i])
	}

	// Setup consecutive range position (in relation to narrow range position) across two of four accounts
	for i := 0; i < positions.numConsecutive; i++ {
		s.SetupConsecutiveRangePositionAcc(clPool.GetId(), s.TestAccs[i])
	}

	// Setup overlapping range position (in relation to narrow range position) on one of four accounts
	for i := 0; i < positions.numOverlapping; i++ {
		s.SetupOverlappingRangePositionAcc(clPool.GetId(), s.TestAccs[i])
	}

	// Swap multiple times USDC for ETH, therefore increasing the spot price
	ticksActivatedAfterEachSwap, totalFeesExpected := s.swapAndTrackXTimesInARow(clPool.GetId(), DefaultCoin1, ETH, cltypes.MaxSpotPrice, positions.numSwaps)
	s.CollectAndAssertFees(s.Ctx, clPool.GetId(), totalFeesExpected, [][]sdk.Int{ticksActivatedAfterEachSwap}, onlyUSDC, positions)

	// Swap multiple times ETH for USDC, therefore decreasing the spot price
	ticksActivatedAfterEachSwap, totalFeesExpected = s.swapAndTrackXTimesInARow(clPool.GetId(), DefaultCoin0, USDC, cltypes.MinSpotPrice, positions.numSwaps)
	s.CollectAndAssertFees(s.Ctx, clPool.GetId(), totalFeesExpected, [][]sdk.Int{ticksActivatedAfterEachSwap}, onlyETH, positions)

	// Do the same swaps as before, however this time we collect fees after both swap directions are complete.
	ticksActivatedAfterEachSwapUp, totalFeesExpectedUp := s.swapAndTrackXTimesInARow(clPool.GetId(), DefaultCoin1, ETH, cltypes.MaxSpotPrice, positions.numSwaps)
	ticksActivatedAfterEachSwapDown, totalFeesExpectedDown := s.swapAndTrackXTimesInARow(clPool.GetId(), DefaultCoin0, USDC, cltypes.MinSpotPrice, positions.numSwaps)
	totalFeesExpected = totalFeesExpectedUp.Add(totalFeesExpectedDown...)

	// We expect all positions to have both denoms in their fee accumulators except USDC for the overlapping range position since
	// it was not activated during the USDC -> ETH swap direction but was activated during the ETH -> USDC swap direction.
	ticksActivatedAfterEachSwapTest := [][]sdk.Int{ticksActivatedAfterEachSwapUp, ticksActivatedAfterEachSwapDown}
	denomsExpected := [][]string{{USDC, ETH}, {USDC, ETH}, {USDC, ETH}, {NoUSDCExpected, ETH}}

	s.CollectAndAssertFees(s.Ctx, clPool.GetId(), totalFeesExpected, ticksActivatedAfterEachSwapTest, denomsExpected, positions)
}

// CollectAndAssertFees collects fees from a given pool for all positions and verifies that the total fees collected match the expected total fees.
// The method also checks that if the ticks that were active during the swap lie within the range of a position, then the position's fee accumulators
// are not empty. The total fees collected are compared to the expected total fees within an additive tolerance defined by an error tolerance struct.
func (s *KeeperTestSuite) CollectAndAssertFees(ctx sdk.Context, poolId uint64, totalFees sdk.Coins, activeTicks [][]sdk.Int, expectedFeeDenoms [][]string, positions Positions) {
	var totalFeesCollected sdk.Coins
	// Claim full range position fees across all four accounts
	for i := 0; i < positions.numFullRange; i++ {
		totalFeesCollected = s.collectFeesAndCheckInvariance(ctx, poolId, i, DefaultMinTick, DefaultMaxTick, totalFeesCollected, expectedFeeDenoms[0], activeTicks)
	}

	// Claim narrow range position fees across three of four accounts
	for i := 0; i < positions.numNarrowRange; i++ {
		totalFeesCollected = s.collectFeesAndCheckInvariance(ctx, poolId, i, DefaultLowerTick, DefaultUpperTick, totalFeesCollected, expectedFeeDenoms[1], activeTicks)
	}

	// Claim consecutive range position fees across two of four accounts
	for i := 0; i < positions.numConsecutive; i++ {
		totalFeesCollected = s.collectFeesAndCheckInvariance(ctx, poolId, i, DefaultExponentConsecutivePositionLowerTick.Int64(), DefaultExponentConsecutivePositionUpperTick.Int64(), totalFeesCollected, expectedFeeDenoms[2], activeTicks)
	}

	// Claim overlapping range position fees on one of four accounts
	for i := 0; i < positions.numOverlapping; i++ {
		totalFeesCollected = s.collectFeesAndCheckInvariance(ctx, poolId, i, DefaultExponentOverlappingPositionLowerTick.Int64(), DefaultExponentOverlappingPositionUpperTick.Int64(), totalFeesCollected, expectedFeeDenoms[3], activeTicks)
	}

	// Define error tolerance
	var errTolerance osmomath.ErrTolerance
	errTolerance.AdditiveTolerance = sdk.NewDec(10)
	errTolerance.RoundingDir = osmomath.RoundDown

	// Check that the total fees collected is equal to the total fees (within a tolerance)
	for _, coin := range totalFeesCollected {
		expected := totalFees.AmountOf(coin.Denom)
		actual := coin.Amount
		s.Require().Equal(0, errTolerance.Compare(expected, actual), fmt.Sprintf("expected (%s), actual (%s)", expected, actual))
	}
}

// tickStatusInvariance tests if the swap position was active during the given tick range and
// checks that the fees collected are non-zero if the position was active, or zero otherwise.
func (s *KeeperTestSuite) tickStatusInvariance(ticksActivatedAfterEachSwap [][]sdk.Int, lowerTick, upperTick int64, coins sdk.Coins, expectedFeeDenoms []string) {
	var positionWasActive bool
	// Check if the position was active during the swap
	for i, ticks := range ticksActivatedAfterEachSwap {
		for _, tick := range ticks {
			if tick.GTE(sdk.NewInt(lowerTick)) && tick.LTE(sdk.NewInt(upperTick)) {
				positionWasActive = true
				break
			}
		}
		if positionWasActive {
			// If the position was active, check that the fees collected are non-zero
			if expectedFeeDenoms[i] != NoUSDCExpected && expectedFeeDenoms[i] != NoETHExpected {
				s.Require().True(coins.AmountOf(expectedFeeDenoms[i]).GT(sdk.ZeroInt()))
			}
		} else {
			// If the position was not active, check that the fees collected are zero
			s.Require().Nil(coins)
		}
	}
}

// swapAndTrackXTimesInARow performs `numSwaps` swaps and tracks the tick activated after each swap.
// It also returns the total fees collected.
func (s *KeeperTestSuite) swapAndTrackXTimesInARow(poolId uint64, coinIn sdk.Coin, coinOutDenom string, priceLimit sdk.Dec, numSwaps int) (ticksActivatedAfterEachSwap []sdk.Int, totalFees sdk.Coins) {
	// Retrieve pool
	clPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	// Determine amount needed to fulfill swap numSwaps times and fund account that much
	amountNeededForSwap := coinIn.Amount.Mul(sdk.NewInt(int64(numSwaps)))
	swapCoin := sdk.NewCoin(coinIn.Denom, amountNeededForSwap)
	s.FundAcc(s.TestAccs[4], sdk.NewCoins(swapCoin))

	ticksActivatedAfterEachSwap = append(ticksActivatedAfterEachSwap, clPool.GetCurrentTick())
	totalFees = sdk.NewCoins(sdk.NewCoin(USDC, sdk.ZeroInt()), sdk.NewCoin(ETH, sdk.ZeroInt()))
	// Swap numSwaps times, recording the tick activated after and swap and fees we expect to collect based on the amount in
	for i := 0; i < numSwaps; i++ {
		coinIn, _, _, _, _, err := s.App.ConcentratedLiquidityKeeper.SwapOutAmtGivenIn(s.Ctx, s.TestAccs[4], clPool, coinIn, coinOutDenom, clPool.GetSwapFee(s.Ctx), priceLimit, clPool.GetId())
		s.Require().NoError(err)
		fee := coinIn.Amount.ToDec().Mul(clPool.GetSwapFee(s.Ctx))
		totalFees = totalFees.Add(sdk.NewCoin(coinIn.Denom, fee.TruncateInt()))
		clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPool.GetId())
		s.Require().NoError(err)
		ticksActivatedAfterEachSwap = append(ticksActivatedAfterEachSwap, clPool.GetCurrentTick())
	}
	return ticksActivatedAfterEachSwap, totalFees
}

// collectFeesAndCheckInvariance collects fees from the concentrated liquidity pool and checks the resulting tick status invariance.
func (s *KeeperTestSuite) collectFeesAndCheckInvariance(ctx sdk.Context, poolId uint64, accountIndex int, minTick, maxTick int64, feesCollected sdk.Coins, expectedFeeDenoms []string, activeTicks [][]sdk.Int) (totalFeesCollected sdk.Coins) {
	coins, err := s.App.ConcentratedLiquidityKeeper.CollectFees(ctx, poolId, s.TestAccs[accountIndex], minTick, maxTick)
	s.Require().NoError(err)
	totalFeesCollected = feesCollected.Add(coins...)
	s.tickStatusInvariance(activeTicks, minTick, maxTick, coins, expectedFeeDenoms)
	return totalFeesCollected
}
