package concentrated_liquidity_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/math"
	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

const (
	NoUSDCExpected = ""
	NoETHExpected  = ""
)

// fields used to identify a spread reward position.
type positionFields struct {
	poolId     uint64
	owner      sdk.AccAddress
	lowerTick  int64
	upperTick  int64
	positionId uint64
	liquidity  sdk.Dec
}

var (
	oneEth      = sdk.NewDecCoin(ETH, sdk.OneInt())
	oneEthCoins = sdk.NewDecCoins(oneEth)
	onlyUSDC    = [][]string{{USDC}, {USDC}, {USDC}, {USDC}}
	onlyETH     = [][]string{{ETH}, {ETH}, {ETH}, {ETH}}
)

func (s *KeeperTestSuite) TestCreateAndGetSpreadRewardsAccumulator() {
	type initSpreadFactorAccumTest struct {
		poolId              uint64
		initializePoolAccum bool

		expectError bool
	}
	tests := map[string]initSpreadFactorAccumTest{
		"default pool setup": {
			poolId:              defaultPoolId,
			initializePoolAccum: true,
		},
		"setup with different poolId": {
			poolId:              defaultPoolId + 1,
			initializePoolAccum: true,
		},
		"pool not initialized": {
			initializePoolAccum: false,
			poolId:              defaultPoolId,
			expectError:         true,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			clKeeper := s.App.ConcentratedLiquidityKeeper

			// system under test
			if tc.initializePoolAccum {
				err := clKeeper.CreateSpreadRewardsAccumulator(s.Ctx, tc.poolId)
				s.Require().NoError(err)
			}
			poolSpreadRewardsAccumulator, err := clKeeper.GetSpreadRewardsAccumulator(s.Ctx, tc.poolId)

			if !tc.expectError {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				s.Require().Equal(accum.AccumulatorObject{}, poolSpreadRewardsAccumulator)
			}
		})
	}
}

func (s *KeeperTestSuite) TestInitOrUpdatePositionSpreadRewardsAccumulator() {
	// Setup is done once so that we test
	// the relationship between test cases.
	// For example, that positions with non-zero liquidity
	// cannot be overridden.
	s.SetupTest()
	s.PrepareConcentratedPool()

	defaultAccount := s.TestAccs[0]

	var (
		defaultPoolId         = uint64(1)
		defaultPositionFields = positionFields{
			defaultPoolId,
			defaultAccount,
			DefaultLowerTick,
			DefaultUpperTick,
			DefaultPositionId,
			DefaultLiquidityAmt,
		}
	)

	withOwner := func(posFields positionFields, owner sdk.AccAddress) positionFields {
		posFields.owner = owner
		return posFields
	}

	withUpperTick := func(posFields positionFields, upperTick int64) positionFields {
		posFields.upperTick = upperTick
		return posFields
	}

	withLowerTick := func(posFields positionFields, lowerTick int64) positionFields {
		posFields.lowerTick = lowerTick
		return posFields
	}

	withPositionId := func(posFields positionFields, positionId uint64) positionFields {
		posFields.positionId = positionId
		return posFields
	}

	withLiquidity := func(posFields positionFields, liquidity sdk.Dec) positionFields {
		posFields.liquidity = liquidity
		return posFields
	}

	clKeeper := s.App.ConcentratedLiquidityKeeper

	type initSpreadFactorAccumTest struct {
		name           string
		positionFields positionFields

		expectedLiquidity sdk.Dec
		expectedError     error
	}
	tests := []initSpreadFactorAccumTest{
		{
			name:           "error: negative liquidity for the first position",
			positionFields: withLiquidity(defaultPositionFields, DefaultLiquidityAmt.Neg()),
			expectedError:  types.NonPositiveLiquidityForNewPositionError{LiquidityDelta: DefaultLiquidityAmt.Neg(), PositionId: DefaultPositionId},
		},
		{
			name:              "first position",
			positionFields:    defaultPositionFields,
			expectedLiquidity: defaultPositionFields.liquidity,
		},
		{
			name:              "second position",
			positionFields:    withPositionId(withLowerTick(defaultPositionFields, DefaultLowerTick+1), DefaultPositionId+1),
			expectedLiquidity: defaultPositionFields.liquidity,
		},
		{
			name:              "adding to first position",
			positionFields:    defaultPositionFields,
			expectedLiquidity: defaultPositionFields.liquidity.MulInt64(2),
		},
		{
			name:              "removing from first position",
			positionFields:    withLiquidity(defaultPositionFields, defaultPositionFields.liquidity.Neg()),
			expectedLiquidity: defaultPositionFields.liquidity,
		},
		{
			name:              "adding to second position",
			positionFields:    withPositionId(withLowerTick(defaultPositionFields, DefaultLowerTick+1), DefaultPositionId+1),
			expectedLiquidity: defaultPositionFields.liquidity.MulInt64(2),
		},
		{
			name: "error: non-existing accumulator (wrong pool)",
			positionFields: positionFields{
				defaultPoolId + 1, // non-existing pool
				defaultAccount,
				DefaultLowerTick,
				DefaultUpperTick,
				DefaultPositionId,
				DefaultLiquidityAmt,
			},
			expectedError: accum.AccumDoesNotExistError{AccumName: types.KeySpreadFactorPoolAccumulator(defaultPoolId + 1)},
		},
		{
			name:              "existing accumulator, different owner - different position",
			positionFields:    withPositionId(withOwner(defaultPositionFields, s.TestAccs[1]), DefaultPositionId+2),
			expectedLiquidity: defaultPositionFields.liquidity,
		},
		{
			name:              "existing accumulator, different upper tick - different position",
			positionFields:    withPositionId(withUpperTick(defaultPositionFields, DefaultUpperTick+1), DefaultPositionId+3),
			expectedLiquidity: defaultPositionFields.liquidity,
		},
		{
			name:              "existing accumulator, different lower tick - different position",
			positionFields:    withPositionId(withLowerTick(defaultPositionFields, DefaultUpperTick+1), DefaultPositionId+4),
			expectedLiquidity: defaultPositionFields.liquidity,
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			// System under test
			err := clKeeper.InitOrUpdatePositionSpreadRewardsAccumulator(s.Ctx, tc.positionFields.poolId, tc.positionFields.lowerTick, tc.positionFields.upperTick, tc.positionFields.positionId, tc.positionFields.liquidity)
			if tc.expectedError == nil {
				s.Require().NoError(err)

				// get spread reward accum and see if position size has been properly initialized
				poolSpreadRewardsAccumulator, err := clKeeper.GetSpreadRewardsAccumulator(s.Ctx, defaultPoolId)
				s.Require().NoError(err)

				positionKey := types.KeySpreadFactorPositionAccumulator(tc.positionFields.positionId)

				positionSize, err := poolSpreadRewardsAccumulator.GetPositionSize(positionKey)
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedLiquidity, positionSize)

				positionRecord, err := poolSpreadRewardsAccumulator.GetPosition(positionKey)
				s.Require().NoError(err)

				spreadRewardGrowthOutside, err := clKeeper.GetSpreadRewardGrowthOutside(s.Ctx, tc.positionFields.poolId, tc.positionFields.lowerTick, tc.positionFields.upperTick)
				s.Require().NoError(err)

				spreadRewardGrowthInside := poolSpreadRewardsAccumulator.GetValue().Sub(spreadRewardGrowthOutside)

				// Position's accumulator must always equal to the spread reward growth inside the position.
				s.Require().Equal(spreadRewardGrowthInside, positionRecord.AccumValuePerShare)

				// Position's spread reward growth must be zero. Note, that on position update,
				// the unclaimed rewards are updated if there was spread reward growth. However,
				// this test case does not set up this condition.
				// It is tested in TestInitOrUpdateSpreadRewardsAccumulatorPosition_UpdatingPosition.
				s.Require().Equal(cl.EmptyCoins, positionRecord.UnclaimedRewardsTotal)
			} else {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetSpreadRewardGrowthOutside() {
	type spreadRewardGrowthOutsideTest struct {
		poolSetup bool

		lowerTick                          int64
		upperTick                          int64
		currentTick                        int64
		lowerTickSpreadRewardGrowthOutside sdk.DecCoins
		upperTickSpreadRewardGrowthOutside sdk.DecCoins
		globalSpreadRewardGrowth           sdk.DecCoin

		expectedSpreadRewardGrowthOutside sdk.DecCoins
		expectedError                     bool
	}

	defaultAccumCoins := sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10)))
	defaultPoolId := uint64(1)
	defaultInitialLiquidity := sdk.OneDec()

	defaultUpperTickIndex := int64(5)
	defaultLowerTickIndex := int64(3)

	emptyUptimeTrackers := wrapUptimeTrackers(getExpectedUptimes().emptyExpectedAccumValues)

	tests := map[string]spreadRewardGrowthOutsideTest{
		// imagine single swap over entire position
		// crossing left > right and stopping above upper tick
		// In this case, only the upper tick accumulator must have
		// been updated when crossed.
		// Since we track spread rewards accrued below a tick, upper tick is updated
		// while lower tick is zero.
		"single swap left -> right: 2 ticks, one share - current tick > upper tick": {
			poolSetup:                          true,
			lowerTick:                          0,
			upperTick:                          1,
			currentTick:                        2,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
		},
		"single swap left -> right: 3 ticks, two shares - current tick > upper tick": {
			poolSetup:                          true,
			lowerTick:                          0,
			upperTick:                          2,
			currentTick:                        3,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
		},
		"single swap left -> right: 2 ticks, one share - current tick == lower tick": {
			poolSetup:                          true,
			lowerTick:                          0,
			upperTick:                          1,
			currentTick:                        0,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
		},
		"single swap left -> right: 2 ticks, one share - current tick < lower tick": {
			poolSetup:                          true,
			lowerTick:                          1,
			upperTick:                          2,
			currentTick:                        0,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
		},
		"single swap left -> right: 2 ticks, one share - current tick == upper tick": {
			poolSetup:                          true,
			lowerTick:                          0,
			upperTick:                          1,
			currentTick:                        1,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
		},
		// imagine single swap over entire position
		// crossing right > left and stopping at lower tick
		// In this case, all spread rewards must have been accrued inside the tick
		// Since we track spread rewards accrued below a tick, both upper and lower position
		// ticks are zero
		"single swap right -> left: 2 ticks, one share - current tick == lower tick": {
			poolSetup:                          true,
			lowerTick:                          0,
			upperTick:                          1,
			currentTick:                        0,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
		},
		"single swap right -> left: 2 ticks, one share - current tick == upper tick": {
			poolSetup:                          true,
			lowerTick:                          0,
			upperTick:                          1,
			currentTick:                        1,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
		},
		"single swap right -> left: 2 ticks, one share - current tick < lower tick": {
			poolSetup:                          true,
			lowerTick:                          0,
			upperTick:                          1,
			currentTick:                        -1,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
		},
		"single swap right -> left: 2 ticks, one share - current tick > lower tick": {
			poolSetup:                          true,
			lowerTick:                          -1,
			upperTick:                          1,
			currentTick:                        0,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
		},
		"single swap right -> left: 2 ticks, one share - current tick > upper tick": {
			poolSetup:                          true,
			lowerTick:                          -1,
			upperTick:                          1,
			currentTick:                        2,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, sdk.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
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
				currentTick := pool.GetCurrentTick()

				s.initializeTick(s.Ctx, currentTick, tc.lowerTick, defaultInitialLiquidity, tc.lowerTickSpreadRewardGrowthOutside, emptyUptimeTrackers, false)
				s.initializeTick(s.Ctx, currentTick, tc.upperTick, defaultInitialLiquidity, tc.upperTickSpreadRewardGrowthOutside, emptyUptimeTrackers, true)
				pool.SetCurrentTick(tc.currentTick)
				err := s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, pool)
				s.Require().NoError(err)
				err = s.App.ConcentratedLiquidityKeeper.ChargeSpreadRewards(s.Ctx, validPoolId, tc.globalSpreadRewardGrowth)
				s.Require().NoError(err)
			}

			// system under test
			spreadRewardGrowthOutside, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardGrowthOutside(s.Ctx, defaultPoolId, defaultLowerTickIndex, defaultUpperTickIndex)
			if tc.expectedError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				// check if returned spread reward growth outside has correct value
				s.Require().Equal(spreadRewardGrowthOutside, tc.expectedSpreadRewardGrowthOutside)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCalculateSpreadRewardGrowth() {
	defaultGeeFrowthGlobal := sdk.NewDecCoins(sdk.NewDecCoin("uosmo", sdk.NewInt(10)))
	defaultGeeFrowthOutside := sdk.NewDecCoins(sdk.NewDecCoin("uosmo", sdk.NewInt(3)))

	defaultSmallerTargetTick := int64(1)
	defaultCurrentTick := int64(2)
	defaultLargerTargetTick := int64(3)

	type calcSpreadRewardGrowthTest struct {
		isUpperTick                bool
		isCurrentTickGTETargetTick bool
		expectedSpreadRewardGrowth sdk.DecCoins
	}

	tests := map[string]calcSpreadRewardGrowthTest{
		"current Tick is greater than the upper tick": {
			isUpperTick:                true,
			isCurrentTickGTETargetTick: false,
			expectedSpreadRewardGrowth: defaultGeeFrowthOutside,
		},
		"current Tick is less than the upper tick": {
			isUpperTick:                true,
			isCurrentTickGTETargetTick: true,
			expectedSpreadRewardGrowth: defaultGeeFrowthGlobal.Sub(defaultGeeFrowthOutside),
		},
		"current Tick is less than the lower tick": {
			isUpperTick:                false,
			isCurrentTickGTETargetTick: false,
			expectedSpreadRewardGrowth: defaultGeeFrowthGlobal.Sub(defaultGeeFrowthOutside),
		},
		"current Tick is greater than the lower tick": {
			isUpperTick:                false,
			isCurrentTickGTETargetTick: true,
			expectedSpreadRewardGrowth: defaultGeeFrowthOutside,
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
			spreadRewardGrowth := cl.CalculateSpreadRewardGrowth(
				targetTick,
				defaultGeeFrowthOutside,
				defaultCurrentTick,
				defaultGeeFrowthGlobal,
				tc.isUpperTick,
			)
			s.Require().Equal(spreadRewardGrowth, tc.expectedSpreadRewardGrowth)
		})
	}
}

func (s *KeeperTestSuite) TestGetInitialSpreadRewardGrowthOppositeDirectionOfLastTraversalForTick() {
	const (
		validPoolId = 1
	)

	initialPoolTick, err := math.PriceToTickRoundDown(DefaultAmt1.ToDec().Quo(DefaultAmt0.ToDec()), DefaultTickSpacing)
	s.Require().NoError(err)

	tests := map[string]struct {
		poolId                          uint64
		tick                            int64
		initialGlobalSpreadRewardGrowth sdk.DecCoin
		shouldAvoidCreatingAccum        bool

		expectedInitialSpreadRewardGrowthOppositeDirectionOfLastTraversal sdk.DecCoins
		expectError                                                       error
	}{
		"current tick > tick -> spread reward growth global": {
			poolId:                          validPoolId,
			tick:                            initialPoolTick - 1,
			initialGlobalSpreadRewardGrowth: oneEth,

			expectedInitialSpreadRewardGrowthOppositeDirectionOfLastTraversal: sdk.NewDecCoins(oneEth),
		},
		"current tick == tick -> spread reward growth global": {
			poolId:                          validPoolId,
			tick:                            initialPoolTick,
			initialGlobalSpreadRewardGrowth: oneEth,

			expectedInitialSpreadRewardGrowthOppositeDirectionOfLastTraversal: sdk.NewDecCoins(oneEth),
		},
		"current tick < tick -> empty coins": {
			poolId:                          validPoolId,
			tick:                            initialPoolTick + 1,
			initialGlobalSpreadRewardGrowth: oneEth,

			expectedInitialSpreadRewardGrowthOppositeDirectionOfLastTraversal: cl.EmptyCoins,
		},
		"pool does not exist": {
			poolId:                          validPoolId + 1,
			tick:                            initialPoolTick - 1,
			initialGlobalSpreadRewardGrowth: oneEth,

			expectError: types.PoolNotFoundError{PoolId: validPoolId + 1},
		},
		"accumulator does not exist": {
			poolId:                          validPoolId,
			tick:                            0,
			initialGlobalSpreadRewardGrowth: oneEth,
			shouldAvoidCreatingAccum:        true,

			expectError: accum.AccumDoesNotExistError{AccumName: types.KeySpreadFactorPoolAccumulator(validPoolId)},
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			ctx := s.Ctx
			clKeeper := s.App.ConcentratedLiquidityKeeper

			pool, err := clmodel.NewConcentratedLiquidityPool(validPoolId, ETH, USDC, DefaultTickSpacing, DefaultZeroSpreadFactor)
			s.Require().NoError(err)

			// N.B.: we set the listener mock because we would like to avoid
			// utilizing the production listeners. The production listeners
			// are irrelevant in the context of the system under test. However,
			// setting them up would require compromising being able to set up other
			// edge case tests. For example, the test case where spread reward accumulator
			// is not initialized.
			s.setListenerMockOnConcentratedLiquidityKeeper()

			err = clKeeper.SetPool(ctx, &pool)
			s.Require().NoError(err)

			if !tc.shouldAvoidCreatingAccum {
				err = clKeeper.CreateSpreadRewardsAccumulator(ctx, validPoolId)
				s.Require().NoError(err)

				// Setup test position to make sure that tick is initialized
				// We also set up uptime accums to ensure position creation works as intended
				err = clKeeper.CreateUptimeAccumulators(ctx, validPoolId)
				s.Require().NoError(err)
				s.SetupDefaultPosition(validPoolId)

				err = clKeeper.ChargeSpreadRewards(ctx, validPoolId, tc.initialGlobalSpreadRewardGrowth)
				s.Require().NoError(err)
			}

			// System under test.
			initialSpreadRewardGrowthOppositeDirectionOfLastTraversal, err := clKeeper.GetInitialSpreadRewardGrowthOppositeDirectionOfLastTraversalForTick(ctx, tc.poolId, tc.tick)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedInitialSpreadRewardGrowthOppositeDirectionOfLastTraversal, initialSpreadRewardGrowthOppositeDirectionOfLastTraversal)
		})
	}
}

func (s *KeeperTestSuite) TestChargeSpreadRewards() {
	// setup once at the beginning.
	s.SetupTest()

	ctx := s.Ctx
	clKeeper := s.App.ConcentratedLiquidityKeeper

	// create spread reward accumulators with ids 1 and 2 but not 3.
	err := clKeeper.CreateSpreadRewardsAccumulator(ctx, 1)
	s.Require().NoError(err)
	err = clKeeper.CreateSpreadRewardsAccumulator(ctx, 2)
	s.Require().NoError(err)

	tests := []struct {
		name               string
		poolId             uint64
		spreadRewardUpdate sdk.DecCoin

		expectedGlobalGrowth sdk.DecCoins
		expectError          error
	}{
		{
			name:               "pool id 1 - one eth",
			poolId:             1,
			spreadRewardUpdate: oneEth,

			expectedGlobalGrowth: sdk.NewDecCoins(oneEth),
		},
		{
			name:               "pool id 1 - 2 usdc",
			poolId:             1,
			spreadRewardUpdate: sdk.NewDecCoin(USDC, sdk.NewInt(2)),

			expectedGlobalGrowth: sdk.NewDecCoins(oneEth).Add(sdk.NewDecCoin(USDC, sdk.NewInt(2))),
		},
		{
			name:               "pool id 2 - 1 usdc",
			poolId:             2,
			spreadRewardUpdate: oneEth,

			expectedGlobalGrowth: sdk.NewDecCoins(oneEth),
		},
		{
			name:               "accumulator does not exist",
			poolId:             3,
			spreadRewardUpdate: oneEth,

			expectError: accum.AccumDoesNotExistError{AccumName: types.KeySpreadFactorPoolAccumulator(3)},
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			// System under test.
			err := clKeeper.ChargeSpreadRewards(ctx, tc.poolId, tc.spreadRewardUpdate)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}
			s.Require().NoError(err)

			spreadFactorAcumulator, err := clKeeper.GetSpreadRewardsAccumulator(ctx, tc.poolId)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedGlobalGrowth, spreadFactorAcumulator.GetValue())
		})
	}
}

func (s *KeeperTestSuite) TestQueryAndCollectSpreadRewards() {
	ownerWithValidPosition := s.TestAccs[0]
	emptyUptimeTrackers := wrapUptimeTrackers(getExpectedUptimes().emptyExpectedAccumValues)

	tests := map[string]struct {
		// setup parameters.
		initialLiquidity                   sdk.Dec
		lowerTickSpreadRewardGrowthOutside sdk.DecCoins
		upperTickSpreadRewardGrowthOutside sdk.DecCoins
		globalSpreadRewardGrowth           sdk.DecCoins
		currentTick                        int64
		isInvalidPoolIdGiven               bool

		// inputs parameters.
		owner                       sdk.AccAddress
		lowerTick                   int64
		upperTick                   int64
		positionIdToCollectAndQuery uint64

		// expectations.
		expectedSpreadFactorsClaimed sdk.Coins
		expectedError                error
	}{
		// imagine single swap over entire position
		// crossing left > right and stopping above upper tick
		// In this case, only the upper tick accumulator must have
		// been updated when crossed.
		// Since we track spread rewards accrued below a tick, upper tick is updated
		// while lower tick is zero.
		"single swap left -> right: 2 ticks, one share, current tick > upper tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 2,

			expectedSpreadFactorsClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},
		"single swap left -> right: 3 ticks, two shares, current tick > upper tick": {
			initialLiquidity: sdk.NewDec(2),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   2,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 3,

			expectedSpreadFactorsClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(20))),
		},
		"single swap left -> right: 3 ticks, two shares, current tick is in between lower and upper tick": {
			initialLiquidity: sdk.NewDec(2),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   2,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 1,

			expectedSpreadFactorsClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(20))),
		},
		"single swap left -> right: 2 ticks, one share, current tick == upper tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 1,

			expectedSpreadFactorsClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},

		// imagine single swap over entire position
		// crossing right -> left and stopping at lower tick
		// In this case, all spread rewards must have been accrued inside the tick
		// Since we track spread rewards accrued below a tick, both upper and lower position
		// ticks are zero
		"single swap right -> left: 2 ticks, one share, current tick == lower tick": {
			initialLiquidity: sdk.OneDec(),

			// lower tick accumulator is not updated yet.
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 0,

			expectedSpreadFactorsClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, one share, current tick < lower tick": {
			initialLiquidity: sdk.OneDec(),

			// lower tick accumulator updated when crossed.
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: -1,

			expectedSpreadFactorsClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, two shares, current tick is in between lower and upper tick": {
			initialLiquidity: sdk.NewDec(2),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   2,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 1,

			expectedSpreadFactorsClaimed: sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(20))),
		},

		// imagine swap occurring outside of the position
		// As a result, lower and upper ticks are not updated.
		"swap occurs above the position, current tick > upper tick": {
			initialLiquidity: sdk.OneDec(),

			// none are updated.
			lowerTickSpreadRewardGrowthOutside: cl.EmptyCoins,
			upperTickSpreadRewardGrowthOutside: cl.EmptyCoins,

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 5,

			expectedSpreadFactorsClaimed: sdk.NewCoins(),
		},
		"swap occurs below the position, current tick < lower tick": {
			initialLiquidity: sdk.OneDec(),

			// none are updated.
			lowerTickSpreadRewardGrowthOutside: cl.EmptyCoins,
			upperTickSpreadRewardGrowthOutside: cl.EmptyCoins,

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   -10,
			upperTick:                   -4,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: -13,

			expectedSpreadFactorsClaimed: sdk.NewCoins(),
		},

		// error cases.

		"position does not exist": {
			initialLiquidity: sdk.OneDec(),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId + 1, // position id does not exist.

			currentTick: 2,

			expectedError: types.PositionIdNotFoundError{PositionId: 2},
		},
		"non owner attempts to collect": {
			initialLiquidity: sdk.OneDec(),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			owner:                       s.TestAccs[1], // different owner from the one who initialized the position.
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 2,

			expectedError: types.NotPositionOwnerError{Address: s.TestAccs[1].String(), PositionId: DefaultPositionId},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			tc := tc
			s.SetupTest()

			validPool := s.PrepareConcentratedPool()
			validPoolId := validPool.GetId()

			s.FundAcc(validPool.GetSpreadRewardsAddress(), tc.expectedSpreadFactorsClaimed)

			clKeeper := s.App.ConcentratedLiquidityKeeper
			ctx := s.Ctx

			// Set the position in store, otherwise querying via position id will fail.
			err := clKeeper.SetPosition(ctx, validPoolId, ownerWithValidPosition, tc.lowerTick, tc.upperTick, time.Now().UTC(), tc.initialLiquidity, DefaultPositionId, DefaultUnderlyingLockId)
			s.Require().NoError(err)

			s.initializeSpreadRewardsAccumulatorPositionWithLiquidity(ctx, validPoolId, tc.lowerTick, tc.upperTick, DefaultPositionId, tc.initialLiquidity)

			s.initializeTick(ctx, tc.currentTick, tc.lowerTick, tc.initialLiquidity, tc.lowerTickSpreadRewardGrowthOutside, emptyUptimeTrackers, false)

			s.initializeTick(ctx, tc.currentTick, tc.upperTick, tc.initialLiquidity, tc.upperTickSpreadRewardGrowthOutside, emptyUptimeTrackers, true)

			validPool.SetCurrentTick(tc.currentTick)
			err = clKeeper.SetPool(ctx, validPool)
			s.Require().NoError(err)

			err = clKeeper.ChargeSpreadRewards(ctx, validPoolId, tc.globalSpreadRewardGrowth[0])
			s.Require().NoError(err)

			poolSpreadFactorBalanceBeforeCollect := s.App.BankKeeper.GetBalance(ctx, validPool.GetSpreadRewardsAddress(), ETH)
			ownerBalancerBeforeCollect := s.App.BankKeeper.GetBalance(ctx, tc.owner, ETH)

			var preQueryPosition accum.Record
			positionKey := types.KeySpreadFactorPositionAccumulator(DefaultPositionId)

			// Note the position accumulator before the query to ensure the query in non-mutating.
			accum, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardsAccumulator(ctx, validPoolId)
			s.Require().NoError(err)
			preQueryPosition, _ = accum.GetPosition(positionKey)

			// System under test
			spreadFactorQueryAmount, queryErr := clKeeper.GetClaimableSpreadRewards(ctx, tc.positionIdToCollectAndQuery)

			// If the query succeeds, the position should not be updated.
			if queryErr == nil {
				accum, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardsAccumulator(ctx, validPoolId)
				s.Require().NoError(err)
				postQueryPosition, _ := accum.GetPosition(positionKey)
				s.Require().Equal(preQueryPosition, postQueryPosition)
			}

			actualSpreadFactorsClaimed, err := clKeeper.CollectSpreadRewards(ctx, tc.owner, tc.positionIdToCollectAndQuery)

			// Assertions.

			poolSpreadFactorBalanceAfterCollect := s.App.BankKeeper.GetBalance(ctx, validPool.GetSpreadRewardsAddress(), ETH)
			ownerBalancerAfterCollect := s.App.BankKeeper.GetBalance(ctx, tc.owner, ETH)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				s.Require().Equal(sdk.Coins{}, actualSpreadFactorsClaimed)

				// balances are unchanged
				s.Require().Equal(poolSpreadFactorBalanceAfterCollect, poolSpreadFactorBalanceBeforeCollect)
				s.Require().Equal(ownerBalancerAfterCollect, ownerBalancerBeforeCollect)
				return
			}

			s.Require().NoError(err)
			s.Require().NoError(queryErr)
			s.Require().Equal(tc.expectedSpreadFactorsClaimed.String(), actualSpreadFactorsClaimed.String())
			s.Require().Equal(spreadFactorQueryAmount.String(), actualSpreadFactorsClaimed.String())

			expectedETHAmount := tc.expectedSpreadFactorsClaimed.AmountOf(ETH)
			s.Require().Equal(expectedETHAmount.String(), poolSpreadFactorBalanceBeforeCollect.Sub(poolSpreadFactorBalanceAfterCollect).Amount.String())
			s.Require().Equal(expectedETHAmount.String(), ownerBalancerAfterCollect.Sub(ownerBalancerBeforeCollect).Amount.String())
		})
	}
}

func (s *KeeperTestSuite) TestPrepareClaimableSpreadRewards() {
	emptyUptimeTrackers := wrapUptimeTrackers(getExpectedUptimes().emptyExpectedAccumValues)

	tests := map[string]struct {
		// setup parameters.
		initialLiquidity                   sdk.Dec
		lowerTickSpreadRewardGrowthOutside sdk.DecCoins
		upperTickSpreadRewardGrowthOutside sdk.DecCoins
		globalSpreadRewardGrowth           sdk.DecCoins
		currentTick                        int64
		isInvalidPoolIdGiven               bool

		// inputs parameters.
		lowerTick           int64
		upperTick           int64
		positionIdToPrepare uint64

		// expectations.
		expectedInitAccumValue sdk.DecCoins
		expectedError          error
	}{
		"single swap left -> right: 2 ticks, one share, current tick > upper tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           1,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 2,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
		},
		"single swap left -> right: 3 ticks, two shares, current tick > upper tick": {
			initialLiquidity: sdk.NewDec(2),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           2,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 3,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
		},
		"single swap left -> right: 3 ticks, two shares, current tick in between lower and upper tick": {
			initialLiquidity: sdk.NewDec(2),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           2,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 1,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
		},
		"single swap left -> right: 2 ticks, one share, current tick == upper tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           1,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 1,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, one share, current tick == lower tick": {
			initialLiquidity: sdk.OneDec(),

			// lower tick accumulator is not updated yet.
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           1,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 0,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, one share, current tick < lower tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           1,
			positionIdToPrepare: DefaultPositionId,

			currentTick: -1,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, two shares, current tick in between lower and upper tick": {
			initialLiquidity: sdk.NewDec(2),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           2,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 1,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),
		},
		"swap occurs above the position, current tick > upper tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickSpreadRewardGrowthOutside: cl.EmptyCoins,
			upperTickSpreadRewardGrowthOutside: cl.EmptyCoins,

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           1,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 5,

			expectedInitAccumValue: sdk.DecCoins(nil),
		},
		"swap occurs below the position, current tick < lower tick": {
			initialLiquidity: sdk.OneDec(),

			lowerTickSpreadRewardGrowthOutside: cl.EmptyCoins,
			upperTickSpreadRewardGrowthOutside: cl.EmptyCoins,

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           -10,
			upperTick:           -4,
			positionIdToPrepare: DefaultPositionId,

			currentTick: -13,

			expectedInitAccumValue: sdk.DecCoins(nil),
		},

		// error cases.
		"position does not exist": {
			initialLiquidity: sdk.OneDec(),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, sdk.NewInt(10))),

			lowerTick:           0,
			upperTick:           1,
			positionIdToPrepare: DefaultPositionId + 1, // position id does not exist.

			currentTick: 2,

			expectedError: types.PositionIdNotFoundError{PositionId: 2},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			tc := tc
			s.SetupTest()

			validPool := s.PrepareConcentratedPool()
			validPoolId := validPool.GetId()

			clKeeper := s.App.ConcentratedLiquidityKeeper
			ctx := s.Ctx

			// Set the position in store.
			err := clKeeper.SetPosition(ctx, validPoolId, s.TestAccs[0], tc.lowerTick, tc.upperTick, time.Now().UTC(), tc.initialLiquidity, DefaultPositionId, DefaultUnderlyingLockId)
			s.Require().NoError(err)

			s.initializeSpreadRewardsAccumulatorPositionWithLiquidity(ctx, validPoolId, tc.lowerTick, tc.upperTick, DefaultPositionId, tc.initialLiquidity)
			s.initializeTick(ctx, tc.currentTick, tc.lowerTick, tc.initialLiquidity, tc.lowerTickSpreadRewardGrowthOutside, emptyUptimeTrackers, false)
			s.initializeTick(ctx, tc.currentTick, tc.upperTick, tc.initialLiquidity, tc.upperTickSpreadRewardGrowthOutside, emptyUptimeTrackers, true)
			validPool.SetCurrentTick(tc.currentTick)

			_ = clKeeper.SetPool(ctx, validPool)

			err = clKeeper.ChargeSpreadRewards(ctx, validPoolId, tc.globalSpreadRewardGrowth[0])
			s.Require().NoError(err)

			positionKey := types.KeySpreadFactorPositionAccumulator(DefaultPositionId)

			// Note the position accumulator before calling prepare
			_, err = s.App.ConcentratedLiquidityKeeper.GetSpreadRewardsAccumulator(ctx, validPoolId)
			s.Require().NoError(err)

			// System under test
			actualSpreadFactorsClaimed, err := clKeeper.PrepareClaimableSpreadRewards(ctx, tc.positionIdToPrepare)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				s.Require().Equal(sdk.Coins(nil), actualSpreadFactorsClaimed)
				return
			}
			s.Require().NoError(err)

			accum, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardsAccumulator(ctx, validPoolId)
			s.Require().NoError(err)

			postPreparePosition, err := accum.GetPosition(positionKey)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedInitAccumValue, postPreparePosition.AccumValuePerShare)
			s.Require().Equal(tc.initialLiquidity, postPreparePosition.NumShares)

			expectedSpreadFactorClaimAmount := tc.expectedInitAccumValue.AmountOf(ETH).Mul(tc.initialLiquidity).TruncateInt()
			s.Require().Equal(expectedSpreadFactorClaimAmount, actualSpreadFactorsClaimed.AmountOf(ETH))
		})
	}
}

// This test ensures that the position's spread reward accumulator is updated correctly when the spread reward grows.
// It validates that another position within the same tick does not affect the current position.
// It also validates that the position's changes are applied at the right time relative to position's
// spread reward accumulator creation or update.
func (s *KeeperTestSuite) TestInitOrUpdateSpreadRewardsAccumulatorPosition_UpdatingPosition() {
	type updateSpreadFactorAccumPositionTest struct {
		doesSpreadFactorGrowBeforeFirstCall           bool
		doesSpreadFactorGrowBetweenFirstAndSecondCall bool
		doesSpreadFactorGrowBetweenSecondAndThirdCall bool
		doesSpreadFactorGrowAfterThirdCall            bool

		expectedUnclaimedRewardsPositionOne sdk.DecCoins
		expectedUnclaimedRewardsPositionTwo sdk.DecCoins
	}

	tests := map[string]updateSpreadFactorAccumPositionTest{
		"1: spread reward charged prior to first call to InitOrUpdateSpreadRewardsAccumulatorPosition with position one": {
			doesSpreadFactorGrowBeforeFirstCall: true,

			// Growing spread reward before first position has no effect on the unclaimed rewards
			// of either position because they are not initialized at that point.
			expectedUnclaimedRewardsPositionOne: cl.EmptyCoins,
			// For position two, growing spread reward has no effect on the unclaimed rewards
			// because we never update it, only create it.
			expectedUnclaimedRewardsPositionTwo: cl.EmptyCoins,
		},
		"2: spread reward charged between first and second call to InitOrUpdateSpreadRewardsAccumulatorPosition, after position one is created and before position two is created": {
			doesSpreadFactorGrowBetweenFirstAndSecondCall: true,

			// Position one's unclaimed rewards increase.
			expectedUnclaimedRewardsPositionOne: DefaultSpreadRewardAccumCoins,
			// For position two, growing spread reward has no effect on the unclaimed rewards
			// because we never update it, only create it.
			expectedUnclaimedRewardsPositionTwo: cl.EmptyCoins,
		},
		"3: spread reward charged between second and third call to InitOrUpdateSpreadRewardsAccumulatorPosition, after position two is created and before position 1 is updated": {
			doesSpreadFactorGrowBetweenSecondAndThirdCall: true,

			// spread reward charged because it grows between the second and third position being created.
			// when third position is created, the rewards are moved to unclaimed.
			expectedUnclaimedRewardsPositionOne: DefaultSpreadRewardAccumCoins,
			// For position two, growing spread reward has no effect on the unclaimed rewards
			// because we never update it, only create it.
			expectedUnclaimedRewardsPositionTwo: cl.EmptyCoins,
		},
		"4: spread reward charged after third call to InitOrUpdateSpreadRewardsAccumulatorPosition, after position 1 is updated": {
			doesSpreadFactorGrowAfterThirdCall: true,

			// no spread reward charged because it grows after the position is updated and the rewards are moved to unclaimed.
			expectedUnclaimedRewardsPositionOne: cl.EmptyCoins,
			// For position two, growing spread reward has no effect on the unclaimed rewards
			// because we never update it, only create it.
			expectedUnclaimedRewardsPositionTwo: cl.EmptyCoins,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			pool := s.PrepareConcentratedPool()
			poolId := pool.GetId()

			pool.SetCurrentTick(DefaultCurrTick)

			// Imaginary spread reward charge #1.
			if tc.doesSpreadFactorGrowBeforeFirstCall {
				s.crossTickAndChargeSpreadRewards(poolId, DefaultLowerTick)
			}

			err := s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, poolId, pool.GetCurrentTick(), DefaultLowerTick, DefaultLiquidityAmt, false)
			s.Require().NoError(err)

			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, poolId, pool.GetCurrentTick(), DefaultUpperTick, DefaultLiquidityAmt, true)
			s.Require().NoError(err)

			// InitOrUpdateSpreadRewardsAccumulatorPosition #1 lower tick to upper tick
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdatePositionSpreadRewardsAccumulator(s.Ctx, poolId, DefaultLowerTick, DefaultUpperTick, DefaultPositionId, DefaultLiquidityAmt)
			s.Require().NoError(err)

			// Imaginary spread reward charge #2.
			if tc.doesSpreadFactorGrowBetweenFirstAndSecondCall {
				s.crossTickAndChargeSpreadRewards(poolId, DefaultLowerTick)
			}

			// InitOrUpdateSpreadRewardsAccumulatorPosition # 2 lower tick to upper tick with a different position id.
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdatePositionSpreadRewardsAccumulator(s.Ctx, poolId, DefaultLowerTick, DefaultUpperTick, DefaultPositionId+1, DefaultLiquidityAmt)
			s.Require().NoError(err)

			// Imaginary spread reward charge #3.
			if tc.doesSpreadFactorGrowBetweenSecondAndThirdCall {
				s.crossTickAndChargeSpreadRewards(poolId, DefaultLowerTick)
			}

			// InitOrUpdateSpreadRewardsAccumulatorPosition # 3 lower tick to upper tick with the original position id.
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdatePositionSpreadRewardsAccumulator(s.Ctx, poolId, DefaultLowerTick, DefaultUpperTick, DefaultPositionId, DefaultLiquidityAmt)
			s.Require().NoError(err)

			// Imaginary spread reward charge #4.
			if tc.doesSpreadFactorGrowAfterThirdCall {
				s.crossTickAndChargeSpreadRewards(poolId, DefaultLowerTick)
			}

			// Validate original position's spread reward growth.
			s.validatePositionSpreadRewardGrowth(poolId, DefaultPositionId, tc.expectedUnclaimedRewardsPositionOne)

			// Validate second position's spread reward growth.
			s.validatePositionSpreadRewardGrowth(poolId, DefaultPositionId+1, tc.expectedUnclaimedRewardsPositionTwo)

			// Validate position one was updated with default liquidity twice.
			s.validatePositionSpreadFactorAccUpdate(s.Ctx, poolId, DefaultPositionId, DefaultLiquidityAmt.MulInt64(2))

			// Validate position two was updated with default liquidity once.
			s.validatePositionSpreadFactorAccUpdate(s.Ctx, poolId, DefaultPositionId+1, DefaultLiquidityAmt)
		})
	}
}

func (s *KeeperTestSuite) TestUpdatePosValueToInitValuePlusGrowthOutside() {
	validPositionKey := types.KeySpreadFactorPositionAccumulator(1)
	invalidPositionKey := types.KeySpreadFactorPositionAccumulator(2)
	tests := []struct {
		name                      string
		poolId                    uint64
		spreadRewardGrowthOutside sdk.DecCoins
		invalidPositionKey        bool
		expectError               error
	}{
		{
			name:                      "happy path",
			spreadRewardGrowthOutside: oneEthCoins,
		},
		{
			name:                      "error: non existent accumulator",
			spreadRewardGrowthOutside: oneEthCoins,
			invalidPositionKey:        true,
			expectError:               accum.NoPositionError{Name: invalidPositionKey},
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			// Setup test env.
			s.SetupTest()
			clKeeper := s.App.ConcentratedLiquidityKeeper
			s.PrepareConcentratedPool()
			poolSpreadRewardsAccumulator, err := clKeeper.GetSpreadRewardsAccumulator(s.Ctx, defaultPoolId)
			s.Require().NoError(err)
			positionKey := validPositionKey

			// Initialize position accumulator.
			err = poolSpreadRewardsAccumulator.NewPosition(positionKey, sdk.OneDec(), nil)
			s.Require().NoError(err)

			// Record the initial position accumulator value.
			positionPre, err := accum.GetPosition(poolSpreadRewardsAccumulator, positionKey)
			s.Require().NoError(err)

			// If the test case requires an invalid position key, set it.
			if tc.invalidPositionKey {
				positionKey = invalidPositionKey
			}

			// System under test.
			err = cl.UpdatePosValueToInitValuePlusGrowthOutside(poolSpreadRewardsAccumulator, positionKey, tc.spreadRewardGrowthOutside)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}
			s.Require().NoError(err)

			// Record the final position accumulator value.
			positionPost, err := accum.GetPosition(poolSpreadRewardsAccumulator, positionKey)
			s.Require().NoError(err)

			// Check that the difference between the new and old position accumulator values is equal to the spread reward growth outside.
			positionAccumDelta := positionPost.AccumValuePerShare.Sub(positionPre.AccumValuePerShare)
			s.Require().Equal(tc.spreadRewardGrowthOutside, positionAccumDelta)
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

func (s *KeeperTestSuite) TestFunctional_SpreadFactors_Swaps() {
	positions := Positions{
		numSwaps:       7,
		numAccounts:    5,
		numFullRange:   4,
		numNarrowRange: 3,
		numConsecutive: 2,
		numOverlapping: 1,
	}
	// Init suite.
	s.SetupTest()

	// Default setup only creates 3 accounts, but we need 5 for this test.
	s.TestAccs = apptesting.CreateRandomAccounts(positions.numAccounts)

	// Create a default CL pool, but with a 0.3 percent spread factor.
	clPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, sdk.MustNewDecFromStr("0.002"))

	positionIds := make([][]uint64, 4)
	// Setup full range position across all four accounts
	for i := 0; i < positions.numFullRange; i++ {
		positionId := s.SetupFullRangePositionAcc(clPool.GetId(), s.TestAccs[i])
		positionIds[0] = append(positionIds[0], positionId)
	}

	// Setup narrow range position across three of four accounts
	for i := 0; i < positions.numNarrowRange; i++ {
		positionId := s.SetupDefaultPositionAcc(clPool.GetId(), s.TestAccs[i])
		positionIds[1] = append(positionIds[1], positionId)
	}

	// Setup consecutive range position (in relation to narrow range position) across two of four accounts
	for i := 0; i < positions.numConsecutive; i++ {
		positionId := s.SetupConsecutiveRangePositionAcc(clPool.GetId(), s.TestAccs[i])
		positionIds[2] = append(positionIds[2], positionId)
	}

	// Setup overlapping range position (in relation to narrow range position) on one of four accounts
	for i := 0; i < positions.numOverlapping; i++ {
		positionId := s.SetupOverlappingRangePositionAcc(clPool.GetId(), s.TestAccs[i])
		positionIds[3] = append(positionIds[3], positionId)
	}

	// Swap multiple times USDC for ETH, therefore increasing the spot price
	ticksActivatedAfterEachSwap, totalSpreadFactorsExpected, _, _ := s.swapAndTrackXTimesInARow(clPool.GetId(), DefaultCoin1, ETH, types.MaxSpotPrice, positions.numSwaps)
	s.CollectAndAssertSpreadFactors(s.Ctx, clPool.GetId(), totalSpreadFactorsExpected, positionIds, [][]int64{ticksActivatedAfterEachSwap}, onlyUSDC, positions)

	// Swap multiple times ETH for USDC, therefore decreasing the spot price
	ticksActivatedAfterEachSwap, totalSpreadFactorsExpected, _, _ = s.swapAndTrackXTimesInARow(clPool.GetId(), DefaultCoin0, USDC, types.MinSpotPrice, positions.numSwaps)
	s.CollectAndAssertSpreadFactors(s.Ctx, clPool.GetId(), totalSpreadFactorsExpected, positionIds, [][]int64{ticksActivatedAfterEachSwap}, onlyETH, positions)

	// Do the same swaps as before, however this time we collect spread rewards after both swap directions are complete.
	ticksActivatedAfterEachSwapUp, totalSpreadFactorsExpectedUp, _, _ := s.swapAndTrackXTimesInARow(clPool.GetId(), DefaultCoin1, ETH, types.MaxSpotPrice, positions.numSwaps)
	ticksActivatedAfterEachSwapDown, totalSpreadFactorsExpectedDown, _, _ := s.swapAndTrackXTimesInARow(clPool.GetId(), DefaultCoin0, USDC, types.MinSpotPrice, positions.numSwaps)
	totalSpreadFactorsExpected = totalSpreadFactorsExpectedUp.Add(totalSpreadFactorsExpectedDown...)

	// We expect all positions to have both denoms in their spread reward accumulators except USDC for the overlapping range position since
	// it was not activated during the USDC -> ETH swap direction but was activated during the ETH -> USDC swap direction.
	ticksActivatedAfterEachSwapTest := [][]int64{ticksActivatedAfterEachSwapUp, ticksActivatedAfterEachSwapDown}
	denomsExpected := [][]string{{USDC, ETH}, {USDC, ETH}, {USDC, ETH}, {NoUSDCExpected, ETH}}

	s.CollectAndAssertSpreadFactors(s.Ctx, clPool.GetId(), totalSpreadFactorsExpected, positionIds, ticksActivatedAfterEachSwapTest, denomsExpected, positions)
}

// This test focuses on various functional testing around spread rewards and LP logic.
// It tests invariants such as the following:
// - can create positions in the same range, swap between them and yet collect the correct spread rewards.
// - correct proportions of spread rewards for overlapping positions are withdrawn.
// - withdrawing full liquidity claims correctly under the hood.
// - withdrawing partial liquidity does not withdraw but still lets spread reward claim as desired.
func (s *KeeperTestSuite) TestFunctional_SpreadFactors_LP() {
	// Setup.
	s.SetupTest()
	s.TestAccs = apptesting.CreateRandomAccounts(5)

	var (
		ctx                         = s.Ctx
		concentratedLiquidityKeeper = s.App.ConcentratedLiquidityKeeper
		owner                       = s.TestAccs[0]
	)

	// Create pool with 0.2% spread factor.
	pool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, sdk.MustNewDecFromStr("0.002"))
	fundCoins := sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0.MulRaw(2)), sdk.NewCoin(USDC, DefaultAmt1.MulRaw(2)))
	s.FundAcc(owner, fundCoins)

	// Errors since no position.
	_, _, _, _, _, err := s.App.ConcentratedLiquidityKeeper.SwapOutAmtGivenIn(s.Ctx, owner, pool, sdk.NewCoin(ETH, sdk.OneInt()), USDC, pool.GetSpreadFactor(s.Ctx), types.MaxSpotPrice)
	s.Require().Error(err)

	// Create position in the default range 1.
	positionIdOne, _, _, liquidity, _, _, _, err := concentratedLiquidityKeeper.CreatePosition(ctx, pool.GetId(), owner, DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
	s.Require().NoError(err)

	// Swap once.
	ticksActivatedAfterEachSwap, totalSpreadFactorsExpected, _, _ := s.swapAndTrackXTimesInARow(pool.GetId(), DefaultCoin1, ETH, types.MaxSpotPrice, 1)

	// Withdraw half.
	halfLiquidity := liquidity.Mul(sdk.NewDecWithPrec(5, 1))
	_, _, err = concentratedLiquidityKeeper.WithdrawPosition(ctx, owner, positionIdOne, halfLiquidity)
	s.Require().NoError(err)

	// Collect spread rewards.
	spreadRewardsCollected := s.collectSpreadRewardsAndCheckInvariance(ctx, 0, DefaultMinTick, DefaultMaxTick, positionIdOne, sdk.NewCoins(), []string{USDC}, [][]int64{ticksActivatedAfterEachSwap})
	expectedSpreadFactorsTruncated := totalSpreadFactorsExpected
	for i, spreadFactorToken := range totalSpreadFactorsExpected {
		// We run expected spread rewards through a cycle of divison and multiplication by liquidity to capture appropriate rounding behavior
		expectedSpreadFactorsTruncated[i] = sdk.NewCoin(spreadFactorToken.Denom, spreadFactorToken.Amount.ToDec().QuoTruncate(liquidity).MulTruncate(liquidity).TruncateInt())
	}
	s.Require().Equal(expectedSpreadFactorsTruncated, spreadRewardsCollected)

	// Unclaimed rewards should be emptied since spread rewards were collected.
	s.validatePositionSpreadRewardGrowth(pool.GetId(), positionIdOne, cl.EmptyCoins)

	// Create position in the default range 2.
	positionIdTwo, _, _, fullLiquidity, _, _, _, err := concentratedLiquidityKeeper.CreatePosition(ctx, pool.GetId(), owner, DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
	s.Require().NoError(err)

	// Swap once in the other direction.
	ticksActivatedAfterEachSwap, totalSpreadFactorsExpected, _, _ = s.swapAndTrackXTimesInARow(pool.GetId(), DefaultCoin0, USDC, types.MinSpotPrice, 1)

	// This should claim under the hood for position 2 since full liquidity is removed.
	balanceBeforeWithdraw := s.App.BankKeeper.GetBalance(ctx, owner, ETH)
	amtDenom0, _, err := concentratedLiquidityKeeper.WithdrawPosition(ctx, owner, positionIdTwo, fullLiquidity)
	s.Require().NoError(err)
	balanceAfterWithdraw := s.App.BankKeeper.GetBalance(ctx, owner, ETH)

	// Validate that the correct amount of ETH was collected in withdraw for position two.
	// total spread rewards * full liquidity / (full liquidity + half liquidity)
	expectedPositionToWithdraw := totalSpreadFactorsExpected.AmountOf(ETH).ToDec().Mul(fullLiquidity.Quo(fullLiquidity.Add(halfLiquidity))).TruncateInt()
	s.Require().Equal(expectedPositionToWithdraw.String(), balanceAfterWithdraw.Sub(balanceBeforeWithdraw).Amount.Sub(amtDenom0).String())

	// Validate cannot claim for withdrawn position.
	_, err = s.App.ConcentratedLiquidityKeeper.CollectSpreadRewards(ctx, owner, positionIdTwo)
	s.Require().Error(err)

	spreadRewardsCollected = s.collectSpreadRewardsAndCheckInvariance(ctx, 0, DefaultMinTick, DefaultMaxTick, positionIdOne, sdk.NewCoins(), []string{ETH}, [][]int64{ticksActivatedAfterEachSwap})

	// total spread rewards * half liquidity / (full liquidity + half liquidity)
	expectedSpreadRewardsCollected := totalSpreadFactorsExpected.AmountOf(ETH).ToDec().Mul(halfLiquidity.Quo(fullLiquidity.Add(halfLiquidity))).TruncateInt()
	s.Require().Equal(expectedSpreadRewardsCollected.String(), spreadRewardsCollected.AmountOf(ETH).String())

	// Create position in the default range 3.
	positionIdThree, _, _, fullLiquidity, _, _, _, err := concentratedLiquidityKeeper.CreatePosition(ctx, pool.GetId(), owner, DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
	s.Require().NoError(err)

	collectedThree, err := s.App.ConcentratedLiquidityKeeper.CollectSpreadRewards(ctx, owner, positionIdThree)
	s.Require().NoError(err)
	s.Require().Equal(sdk.Coins(nil), collectedThree)
}

// CollectAndAssertSpreadFactors collects spread rewards from a given pool for all positions and verifies that the total spread rewards collected match the expected total spread rewards.
// The method also checks that if the ticks that were active during the swap lie within the range of a position, then the position's spread reward accumulators
// are not empty. The total spread rewards collected are compared to the expected total spread rewards within an additive tolerance defined by an error tolerance struct.
func (s *KeeperTestSuite) CollectAndAssertSpreadFactors(ctx sdk.Context, poolId uint64, totalSpreadFactors sdk.Coins, positionIds [][]uint64, activeTicks [][]int64, expectedSpreadFactorDenoms [][]string, positions Positions) {
	var totalSpreadRewardsCollected sdk.Coins
	// Claim full range position spread rewards across all four accounts
	for i := 0; i < positions.numFullRange; i++ {
		totalSpreadRewardsCollected = s.collectSpreadRewardsAndCheckInvariance(ctx, i, DefaultMinTick, DefaultMaxTick, positionIds[0][i], totalSpreadRewardsCollected, expectedSpreadFactorDenoms[0], activeTicks)
	}

	// Claim narrow range position spread rewards across three of four accounts
	for i := 0; i < positions.numNarrowRange; i++ {
		totalSpreadRewardsCollected = s.collectSpreadRewardsAndCheckInvariance(ctx, i, DefaultLowerTick, DefaultUpperTick, positionIds[1][i], totalSpreadRewardsCollected, expectedSpreadFactorDenoms[1], activeTicks)
	}

	// Claim consecutive range position spread rewards across two of four accounts
	for i := 0; i < positions.numConsecutive; i++ {
		totalSpreadRewardsCollected = s.collectSpreadRewardsAndCheckInvariance(ctx, i, DefaultExponentConsecutivePositionLowerTick, DefaultExponentConsecutivePositionUpperTick, positionIds[2][i], totalSpreadRewardsCollected, expectedSpreadFactorDenoms[2], activeTicks)
	}

	// Claim overlapping range position spread rewards on one of four accounts
	for i := 0; i < positions.numOverlapping; i++ {
		totalSpreadRewardsCollected = s.collectSpreadRewardsAndCheckInvariance(ctx, i, DefaultExponentOverlappingPositionLowerTick, DefaultExponentOverlappingPositionUpperTick, positionIds[3][i], totalSpreadRewardsCollected, expectedSpreadFactorDenoms[3], activeTicks)
	}

	// Define error tolerance
	var errTolerance osmomath.ErrTolerance
	errTolerance.AdditiveTolerance = sdk.NewDec(10)
	errTolerance.RoundingDir = osmomath.RoundDown

	// Check that the total spread rewards collected is equal to the total spread rewards (within a tolerance)
	for _, coin := range totalSpreadRewardsCollected {
		expected := totalSpreadFactors.AmountOf(coin.Denom)
		actual := coin.Amount
		s.Require().Equal(0, errTolerance.Compare(expected, actual), fmt.Sprintf("expected (%s), actual (%s)", expected, actual))
	}
}

// tickStatusInvariance tests if the swap position was active during the given tick range and
// checks that the spread rewards collected are non-zero if the position was active, or zero otherwise.
func (s *KeeperTestSuite) tickStatusInvariance(ticksActivatedAfterEachSwap [][]int64, lowerTick, upperTick int64, coins sdk.Coins, expectedSpreadFactorDenoms []string) {
	var positionWasActive bool
	// Check if the position was active during the swap
	for i, ticks := range ticksActivatedAfterEachSwap {
		for _, tick := range ticks {
			if tick >= lowerTick && tick <= upperTick {
				positionWasActive = true
				break
			}
		}
		if positionWasActive {
			// If the position was active, check that the spread rewards collected are non-zero
			if expectedSpreadFactorDenoms[i] != NoUSDCExpected && expectedSpreadFactorDenoms[i] != NoETHExpected {
				s.Require().True(coins.AmountOf(expectedSpreadFactorDenoms[i]).GT(sdk.ZeroInt()), "denom: %s", expectedSpreadFactorDenoms[i])
			}
		} else {
			// If the position was not active, check that the spread rewards collected are zero
			s.Require().Nil(coins)
		}
	}
}

// swapAndTrackXTimesInARow performs `numSwaps` swaps and tracks the tick activated after each swap.
// It also returns the total spread rewards collected, the total token in, and the total token out.
func (s *KeeperTestSuite) swapAndTrackXTimesInARow(poolId uint64, coinIn sdk.Coin, coinOutDenom string, priceLimit sdk.Dec, numSwaps int) (ticksActivatedAfterEachSwap []int64, totalSpreadFactors sdk.Coins, totalTokenIn sdk.Coin, totalTokenOut sdk.Coin) {
	// Retrieve pool
	clPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	// Determine amount needed to fulfill swap numSwaps times and fund account that much
	amountNeededForSwap := coinIn.Amount.Mul(sdk.NewInt(int64(numSwaps)))
	swapCoin := sdk.NewCoin(coinIn.Denom, amountNeededForSwap)
	s.FundAcc(s.TestAccs[4], sdk.NewCoins(swapCoin))

	ticksActivatedAfterEachSwap = append(ticksActivatedAfterEachSwap, clPool.GetCurrentTick())
	totalSpreadFactors = sdk.NewCoins(sdk.NewCoin(USDC, sdk.ZeroInt()), sdk.NewCoin(ETH, sdk.ZeroInt()))
	// Swap numSwaps times, recording the tick activated after and swap and spread rewards we expect to collect based on the amount in
	totalTokenIn = sdk.NewCoin(coinIn.Denom, sdk.ZeroInt())
	totalTokenOut = sdk.NewCoin(coinOutDenom, sdk.ZeroInt())
	for i := 0; i < numSwaps; i++ {
		coinIn, coinOut, _, _, _, err := s.App.ConcentratedLiquidityKeeper.SwapOutAmtGivenIn(s.Ctx, s.TestAccs[4], clPool, coinIn, coinOutDenom, clPool.GetSpreadFactor(s.Ctx), priceLimit)
		s.Require().NoError(err)
		spreadFactor := coinIn.Amount.ToDec().Mul(clPool.GetSpreadFactor(s.Ctx))
		totalSpreadFactors = totalSpreadFactors.Add(sdk.NewCoin(coinIn.Denom, spreadFactor.TruncateInt()))
		clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPool.GetId())
		s.Require().NoError(err)
		ticksActivatedAfterEachSwap = append(ticksActivatedAfterEachSwap, clPool.GetCurrentTick())
		totalTokenIn = totalTokenIn.Add(coinIn)
		totalTokenOut = totalTokenOut.Add(coinOut)
	}
	return ticksActivatedAfterEachSwap, totalSpreadFactors, totalTokenIn, totalTokenOut
}

// collectSpreadRewardsAndCheckInvariance collects spread rewards from the concentrated liquidity pool and checks the resulting tick status invariance.
func (s *KeeperTestSuite) collectSpreadRewardsAndCheckInvariance(ctx sdk.Context, accountIndex int, minTick, maxTick int64, positionId uint64, spreadRewardsCollected sdk.Coins, expectedSpreadFactorDenoms []string, activeTicks [][]int64) (totalSpreadRewardsCollected sdk.Coins) {
	coins, err := s.App.ConcentratedLiquidityKeeper.CollectSpreadRewards(ctx, s.TestAccs[accountIndex], positionId)
	s.Require().NoError(err)
	totalSpreadRewardsCollected = spreadRewardsCollected.Add(coins...)
	s.tickStatusInvariance(activeTicks, minTick, maxTick, coins, expectedSpreadFactorDenoms)
	return totalSpreadRewardsCollected
}
