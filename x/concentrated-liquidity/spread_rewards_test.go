package concentrated_liquidity_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	cl "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
	clmath "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/math"
	clmodel "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
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
	liquidity  osmomath.Dec
}

var (
	oneEth      = sdk.NewDecCoin(ETH, osmomath.OneInt())
	oneEthCoins = sdk.NewDecCoins(oneEth)
	onlyUSDC    = [][]string{{USDC}, {USDC}, {USDC}, {USDC}}
	onlyETH     = [][]string{{ETH}, {ETH}, {ETH}, {ETH}}
)

func (s *KeeperTestSuite) TestCreateAndGetSpreadRewardAccumulator() {
	type initSpreadRewardAccumTest struct {
		poolId              uint64
		initializePoolAccum bool

		expectError bool
	}
	tests := map[string]initSpreadRewardAccumTest{
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
				err := clKeeper.CreateSpreadRewardAccumulator(s.Ctx, tc.poolId)
				s.Require().NoError(err)
			}
			poolSpreadRewardAccumulator, err := clKeeper.GetSpreadRewardAccumulator(s.Ctx, tc.poolId)

			if !tc.expectError {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				s.Require().Equal(&accum.AccumulatorObject{}, poolSpreadRewardAccumulator)
			}
		})
	}
}

func (s *KeeperTestSuite) TestInitOrUpdatePositionSpreadRewardAccumulator() {
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

	withLiquidity := func(posFields positionFields, liquidity osmomath.Dec) positionFields {
		posFields.liquidity = liquidity
		return posFields
	}

	clKeeper := s.App.ConcentratedLiquidityKeeper
	secondPosition := withPositionId(withLowerTick(defaultPositionFields, DefaultLowerTick+1), DefaultPositionId+1)

	type initSpreadRewardAccumTest struct {
		name           string
		positionFields positionFields

		expectedLiquidity osmomath.Dec
		expectedError     error
	}
	tests := []initSpreadRewardAccumTest{
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
			positionFields:    secondPosition,
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
			positionFields:    secondPosition,
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
			expectedError: accum.AccumDoesNotExistError{AccumName: types.KeySpreadRewardPoolAccumulator(defaultPoolId + 1)},
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
			err := clKeeper.InitOrUpdatePositionSpreadRewardAccumulator(s.Ctx, tc.positionFields.poolId, tc.positionFields.lowerTick, tc.positionFields.upperTick, tc.positionFields.positionId, tc.positionFields.liquidity)
			if tc.expectedError == nil {
				s.Require().NoError(err)

				// get spread reward accum and see if position size has been properly initialized
				poolSpreadRewardAccumulator, err := clKeeper.GetSpreadRewardAccumulator(s.Ctx, defaultPoolId)
				s.Require().NoError(err)

				positionKey := types.KeySpreadRewardPositionAccumulator(tc.positionFields.positionId)

				positionSize, err := poolSpreadRewardAccumulator.GetPositionSize(positionKey)
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedLiquidity, positionSize)

				positionRecord, err := poolSpreadRewardAccumulator.GetPosition(positionKey)
				s.Require().NoError(err)

				spreadRewardGrowthOutside, err := clKeeper.GetSpreadRewardGrowthOutside(s.Ctx, tc.positionFields.poolId, tc.positionFields.lowerTick, tc.positionFields.upperTick)
				s.Require().NoError(err)

				spreadRewardGrowthInside := poolSpreadRewardAccumulator.GetValue().Sub(spreadRewardGrowthOutside)

				// Position's accumulator must always equal to the spread reward growth inside the position.
				s.Require().Equal(spreadRewardGrowthInside, positionRecord.AccumValuePerShare)

				// Position's spread reward growth must be zero. Note, that on position update,
				// the unclaimed rewards are updated if there was spread reward growth. However,
				// this test case does not set up this condition.
				// It is tested in TestInitOrUpdateSpreadRewardAccumulatorPosition_UpdatingPosition.
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

	defaultAccumCoins := sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))).MulDecTruncate(cl.PerUnitLiqScalingFactor)
	defaultPoolId := uint64(1)
	defaultInitialLiquidity := osmomath.OneDec()

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
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, osmomath.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
		},
		"single swap left -> right: 3 ticks, two shares - current tick > upper tick": {
			poolSetup:                          true,
			lowerTick:                          0,
			upperTick:                          2,
			currentTick:                        3,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, osmomath.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
		},
		"single swap left -> right: 2 ticks, one share - current tick == lower tick": {
			poolSetup:                          true,
			lowerTick:                          0,
			upperTick:                          1,
			currentTick:                        0,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, osmomath.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
		},
		"single swap left -> right: 2 ticks, one share - current tick < lower tick": {
			poolSetup:                          true,
			lowerTick:                          1,
			upperTick:                          2,
			currentTick:                        0,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, osmomath.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
		},
		"single swap left -> right: 2 ticks, one share - current tick == upper tick": {
			poolSetup:                          true,
			lowerTick:                          0,
			upperTick:                          1,
			currentTick:                        1,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, osmomath.NewInt(10)),
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
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, osmomath.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
		},
		"single swap right -> left: 2 ticks, one share - current tick == upper tick": {
			poolSetup:                          true,
			lowerTick:                          0,
			upperTick:                          1,
			currentTick:                        1,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, osmomath.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
		},
		"single swap right -> left: 2 ticks, one share - current tick < lower tick": {
			poolSetup:                          true,
			lowerTick:                          0,
			upperTick:                          1,
			currentTick:                        -1,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, osmomath.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
		},
		"single swap right -> left: 2 ticks, one share - current tick > lower tick": {
			poolSetup:                          true,
			lowerTick:                          -1,
			upperTick:                          1,
			currentTick:                        0,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, osmomath.NewInt(10)),
			expectedSpreadRewardGrowthOutside:  defaultAccumCoins,
			expectedError:                      false,
		},
		"single swap right -> left: 2 ticks, one share - current tick > upper tick": {
			poolSetup:                          true,
			lowerTick:                          -1,
			upperTick:                          1,
			currentTick:                        2,
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
			globalSpreadRewardGrowth:           sdk.NewDecCoin(ETH, osmomath.NewInt(10)),
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

				// Upscale accumulator values
				tc.lowerTickSpreadRewardGrowthOutside = tc.lowerTickSpreadRewardGrowthOutside.MulDecTruncate(cl.PerUnitLiqScalingFactor)
				tc.upperTickSpreadRewardGrowthOutside = tc.upperTickSpreadRewardGrowthOutside.MulDecTruncate(cl.PerUnitLiqScalingFactor)

				s.initializeTick(s.Ctx, tc.lowerTick, defaultInitialLiquidity, tc.lowerTickSpreadRewardGrowthOutside, emptyUptimeTrackers, false)
				s.initializeTick(s.Ctx, tc.upperTick, defaultInitialLiquidity, tc.upperTickSpreadRewardGrowthOutside, emptyUptimeTrackers, true)
				pool.SetCurrentTick(tc.currentTick)
				err := s.App.ConcentratedLiquidityKeeper.SetPool(s.Ctx, pool)
				s.Require().NoError(err)
				s.AddToSpreadRewardAccumulator(validPoolId, tc.globalSpreadRewardGrowth)
			}

			// system under test
			spreadRewardGrowthOutside, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardGrowthOutside(s.Ctx, defaultPoolId, defaultLowerTickIndex, defaultUpperTickIndex)
			if tc.expectedError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				// check if returned spread reward growth outside has correct value
				s.Require().Equal(tc.expectedSpreadRewardGrowthOutside, spreadRewardGrowthOutside)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCalculateSpreadRewardGrowth() {
	defaultGeeFrowthGlobal := sdk.NewDecCoins(sdk.NewDecCoin(appparams.BaseCoinUnit, osmomath.NewInt(10)))
	defaultGeeFrowthOutside := sdk.NewDecCoins(sdk.NewDecCoin(appparams.BaseCoinUnit, osmomath.NewInt(3)))

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

// Test what happens if somehow the accumulator didn't exist.
// TODO: Does this test even matter? We should never be in a situation where the accumulator doesn't exist
func (s *KeeperTestSuite) TestGetInitialSpreadRewardGrowthOppositeDirectionOfLastTraversalForTickAccumDoesntExist() {
	pool, err := clmodel.NewConcentratedLiquidityPool(validPoolId, ETH, USDC, DefaultTickSpacing, DefaultZeroSpreadFactor)
	s.Require().NoError(err)

	// N.B.: we set the listener mock because we would like to avoid
	// utilizing the production listeners, because we are testing a case that should be impossible
	s.setListenerMockOnConcentratedLiquidityKeeper()

	err = s.Clk.SetPool(s.Ctx, &pool)
	s.Require().NoError(err)

	// System under test.
	_, err = s.Clk.GetInitialSpreadRewardGrowthOppositeDirectionOfLastTraversalForTick(s.Ctx, &pool, 0)
	s.Require().Error(err)
	s.Require().ErrorIs(err, accum.AccumDoesNotExistError{AccumName: types.KeySpreadRewardPoolAccumulator(validPoolId)})
}

func (s *KeeperTestSuite) TestGetInitialSpreadRewardGrowthOppositeDirectionOfLastTraversalForTick() {
	sqrtPrice := osmomath.MustMonotonicSqrt(DefaultAmt1.ToLegacyDec().Quo(DefaultAmt0.ToLegacyDec()))
	initialPoolTick, err := clmath.SqrtPriceToTickRoundDownSpacing(osmomath.BigDecFromDec(sqrtPrice), DefaultTickSpacing)
	s.Require().NoError(err)
	initialGlobalSpreadRewardGrowth := oneEth

	tests := map[string]struct {
		tick                            int64
		initialGlobalSpreadRewardGrowth sdk.DecCoin

		expectedInitialSpreadRewardGrowthOppositeDirectionOfLastTraversal sdk.DecCoins
	}{
		"current tick > tick -> spread reward growth global": {
			tick: initialPoolTick - 1,
			expectedInitialSpreadRewardGrowthOppositeDirectionOfLastTraversal: sdk.NewDecCoins(oneEth),
		},
		"current tick == tick -> spread reward growth global": {
			tick: initialPoolTick,
			expectedInitialSpreadRewardGrowthOppositeDirectionOfLastTraversal: sdk.NewDecCoins(oneEth),
		},
		"current tick < tick -> empty coins": {
			tick: initialPoolTick + 1,
			expectedInitialSpreadRewardGrowthOppositeDirectionOfLastTraversal: cl.EmptyCoins,
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			pool := s.preparePoolAndDefaultPosition()
			s.AddToSpreadRewardAccumulator(pool.GetId(), initialGlobalSpreadRewardGrowth)

			// Upscale accumulator values
			tc.expectedInitialSpreadRewardGrowthOppositeDirectionOfLastTraversal = tc.expectedInitialSpreadRewardGrowthOppositeDirectionOfLastTraversal.MulDecTruncate(cl.PerUnitLiqScalingFactor)

			clpool, err := s.Clk.GetPoolById(s.Ctx, pool.GetId())
			s.Require().NoError(err)
			initialSpreadRewardGrowthOppositeDirectionOfLastTraversal, err := s.Clk.GetInitialSpreadRewardGrowthOppositeDirectionOfLastTraversalForTick(s.Ctx, clpool, tc.tick)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedInitialSpreadRewardGrowthOppositeDirectionOfLastTraversal, initialSpreadRewardGrowthOppositeDirectionOfLastTraversal)
		})
	}
}

func (s *KeeperTestSuite) TestQueryAndCollectSpreadRewards() {
	ownerWithValidPosition := s.TestAccs[0]
	emptyUptimeTrackers := wrapUptimeTrackers(getExpectedUptimes().emptyExpectedAccumValues)

	tests := map[string]struct {
		// setup parameters.
		initialLiquidity                   osmomath.Dec
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
		expectedSpreadRewardsClaimed sdk.Coins
		expectedError                error
	}{
		// imagine single swap over entire position
		// crossing left > right and stopping above upper tick
		// In this case, only the upper tick accumulator must have
		// been updated when crossed.
		// Since we track spread rewards accrued below a tick, upper tick is updated
		// while lower tick is zero.
		"single swap left -> right: 2 ticks, one share, current tick > upper tick": {
			initialLiquidity: osmomath.OneDec(),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 2,

			expectedSpreadRewardsClaimed: sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(10))),
		},
		"single swap left -> right: 3 ticks, two shares, current tick > upper tick": {
			initialLiquidity: osmomath.NewDec(2),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   2,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 3,

			expectedSpreadRewardsClaimed: sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(20))),
		},
		"single swap left -> right: 3 ticks, two shares, current tick is in between lower and upper tick": {
			initialLiquidity: osmomath.NewDec(2),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   2,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 1,

			expectedSpreadRewardsClaimed: sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(20))),
		},
		"single swap left -> right: 2 ticks, one share, current tick == upper tick": {
			initialLiquidity: osmomath.OneDec(),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 1,

			expectedSpreadRewardsClaimed: sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(10))),
		},

		// imagine single swap over entire position
		// crossing right -> left and stopping at lower tick
		// In this case, all spread rewards must have been accrued inside the tick
		// Since we track spread rewards accrued below a tick, both upper and lower position
		// ticks are zero
		"single swap right -> left: 2 ticks, one share, current tick == lower tick": {
			initialLiquidity: osmomath.OneDec(),

			// lower tick accumulator is not updated yet.
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 0,

			expectedSpreadRewardsClaimed: sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, one share, current tick < lower tick": {
			initialLiquidity: osmomath.OneDec(),

			// lower tick accumulator updated when crossed.
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: -1,

			expectedSpreadRewardsClaimed: sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, two shares, current tick is in between lower and upper tick": {
			initialLiquidity: osmomath.NewDec(2),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   2,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 1,

			expectedSpreadRewardsClaimed: sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(20))),
		},

		// imagine swap occurring outside of the position
		// As a result, lower and upper ticks are not updated.
		"swap occurs above the position, current tick > upper tick": {
			initialLiquidity: osmomath.OneDec(),

			// none are updated.
			lowerTickSpreadRewardGrowthOutside: cl.EmptyCoins,
			upperTickSpreadRewardGrowthOutside: cl.EmptyCoins,

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: 5,

			expectedSpreadRewardsClaimed: sdk.NewCoins(),
		},
		"swap occurs below the position, current tick < lower tick": {
			initialLiquidity: osmomath.OneDec(),

			// none are updated.
			lowerTickSpreadRewardGrowthOutside: cl.EmptyCoins,
			upperTickSpreadRewardGrowthOutside: cl.EmptyCoins,

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   -10,
			upperTick:                   -4,
			positionIdToCollectAndQuery: DefaultPositionId,

			currentTick: -13,

			expectedSpreadRewardsClaimed: sdk.NewCoins(),
		},

		// error cases.

		"position does not exist": {
			initialLiquidity: osmomath.OneDec(),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			owner:                       ownerWithValidPosition,
			lowerTick:                   0,
			upperTick:                   1,
			positionIdToCollectAndQuery: DefaultPositionId + 1, // position id does not exist.

			currentTick: 2,

			expectedError: types.PositionIdNotFoundError{PositionId: 2},
		},
		"non owner attempts to collect": {
			initialLiquidity: osmomath.OneDec(),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

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

			// Upscale accumulator values
			tc.lowerTickSpreadRewardGrowthOutside = tc.lowerTickSpreadRewardGrowthOutside.MulDecTruncate(cl.PerUnitLiqScalingFactor)
			tc.upperTickSpreadRewardGrowthOutside = tc.upperTickSpreadRewardGrowthOutside.MulDecTruncate(cl.PerUnitLiqScalingFactor)

			s.FundAcc(validPool.GetSpreadRewardsAddress(), tc.expectedSpreadRewardsClaimed)

			clKeeper := s.App.ConcentratedLiquidityKeeper
			ctx := s.Ctx

			// Set the position in store, otherwise querying via position id will fail.
			err := clKeeper.SetPosition(ctx, validPoolId, ownerWithValidPosition, tc.lowerTick, tc.upperTick, time.Now().UTC(), tc.initialLiquidity, DefaultPositionId, DefaultUnderlyingLockId)
			s.Require().NoError(err)

			s.initializeSpreadRewardAccumulatorPositionWithLiquidity(ctx, validPoolId, tc.lowerTick, tc.upperTick, DefaultPositionId, tc.initialLiquidity)

			s.initializeTick(ctx, tc.lowerTick, tc.initialLiquidity, tc.lowerTickSpreadRewardGrowthOutside, emptyUptimeTrackers, false)

			s.initializeTick(ctx, tc.upperTick, tc.initialLiquidity, tc.upperTickSpreadRewardGrowthOutside, emptyUptimeTrackers, true)

			validPool.SetCurrentTick(tc.currentTick)
			err = clKeeper.SetPool(ctx, validPool)
			s.Require().NoError(err)

			s.AddToSpreadRewardAccumulator(validPoolId, tc.globalSpreadRewardGrowth[0])

			poolSpreadRewardBalanceBeforeCollect := s.App.BankKeeper.GetBalance(ctx, validPool.GetSpreadRewardsAddress(), ETH)
			ownerBalancerBeforeCollect := s.App.BankKeeper.GetBalance(ctx, tc.owner, ETH)

			var preQueryPosition accum.Record
			positionKey := types.KeySpreadRewardPositionAccumulator(DefaultPositionId)

			// Note the position accumulator before the query to ensure the query in non-mutating.
			accum, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(ctx, validPoolId)
			s.Require().NoError(err)
			preQueryPosition, _ = accum.GetPosition(positionKey)

			// System under test
			spreadRewardQueryAmount, queryErr := clKeeper.GetClaimableSpreadRewards(ctx, tc.positionIdToCollectAndQuery)

			// If the query succeeds, the position should not be updated.
			if queryErr == nil {
				accum, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(ctx, validPoolId)
				s.Require().NoError(err)
				postQueryPosition, _ := accum.GetPosition(positionKey)
				s.Require().Equal(preQueryPosition, postQueryPosition)
			}

			actualSpreadRewardsClaimed, err := clKeeper.CollectSpreadRewards(ctx, tc.owner, tc.positionIdToCollectAndQuery)

			// Assertions.

			poolSpreadRewardBalanceAfterCollect := s.App.BankKeeper.GetBalance(ctx, validPool.GetSpreadRewardsAddress(), ETH)
			ownerBalancerAfterCollect := s.App.BankKeeper.GetBalance(ctx, tc.owner, ETH)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				s.Require().Equal(sdk.Coins{}, actualSpreadRewardsClaimed)

				// balances are unchanged
				s.Require().Equal(poolSpreadRewardBalanceAfterCollect, poolSpreadRewardBalanceBeforeCollect)
				s.Require().Equal(ownerBalancerAfterCollect, ownerBalancerBeforeCollect)
				return
			}

			s.Require().NoError(err)
			s.Require().NoError(queryErr)
			s.Require().Equal(tc.expectedSpreadRewardsClaimed.String(), actualSpreadRewardsClaimed.String())
			s.Require().Equal(spreadRewardQueryAmount.String(), actualSpreadRewardsClaimed.String())

			expectedETHAmount := tc.expectedSpreadRewardsClaimed.AmountOf(ETH)
			s.Require().Equal(expectedETHAmount.String(), poolSpreadRewardBalanceBeforeCollect.Sub(poolSpreadRewardBalanceAfterCollect).Amount.String())
			s.Require().Equal(expectedETHAmount.String(), ownerBalancerAfterCollect.Sub(ownerBalancerBeforeCollect).Amount.String())
		})
	}
}

func (s *KeeperTestSuite) TestPrepareClaimableSpreadRewards() {
	emptyUptimeTrackers := wrapUptimeTrackers(getExpectedUptimes().emptyExpectedAccumValues)

	tests := map[string]struct {
		// setup parameters.
		initialLiquidity                   osmomath.Dec
		lowerTickSpreadRewardGrowthOutside sdk.DecCoins
		upperTickSpreadRewardGrowthOutside sdk.DecCoins
		globalSpreadRewardGrowth           sdk.DecCoins
		expectedReinvestedDustAmount       osmomath.Dec
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
			initialLiquidity: osmomath.OneDec(),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			lowerTick:           0,
			upperTick:           1,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 2,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
		},
		"single swap left -> right: 3 ticks, two shares, current tick > upper tick": {
			initialLiquidity: osmomath.NewDec(2),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			lowerTick:           0,
			upperTick:           2,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 3,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
		},
		"single swap left -> right: 3 ticks, two shares, current tick in between lower and upper tick": {
			initialLiquidity: osmomath.NewDec(2),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			lowerTick:           0,
			upperTick:           2,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 1,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
		},
		"single swap left -> right: 2 ticks, one share, current tick == upper tick": {
			initialLiquidity: osmomath.OneDec(),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			lowerTick:           0,
			upperTick:           1,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 1,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, one share, current tick == lower tick": {
			initialLiquidity: osmomath.OneDec(),

			// lower tick accumulator is not updated yet.
			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			lowerTick:           0,
			upperTick:           1,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 0,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, one share, current tick < lower tick": {
			initialLiquidity: osmomath.OneDec(),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			lowerTick:           0,
			upperTick:           1,
			positionIdToPrepare: DefaultPositionId,

			currentTick: -1,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
		},
		"single swap right -> left: 2 ticks, two shares, current tick in between lower and upper tick": {
			initialLiquidity: osmomath.NewDec(2),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			lowerTick:           0,
			upperTick:           2,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 1,

			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),
		},
		"dust: single swap right -> left: 2 ticks, two shares, current tick in between lower and upper tick": {
			initialLiquidity: osmomath.NewDec(2),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoinFromDec(ETH, osmomath.MustNewDecFromStr("3.3"))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			lowerTick:           0,
			upperTick:           2,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 1,

			// expected = global - below lower - above upper = 10 - 3.3 = 6.7
			expectedInitAccumValue: sdk.NewDecCoins(sdk.NewDecCoinFromDec(ETH, osmomath.MustNewDecFromStr("6.7"))),
			// we no longer reinvest dust, so we expect this to be thrown out
			expectedReinvestedDustAmount: osmomath.ZeroDec(),
		},
		"swap occurs above the position, current tick > upper tick": {
			initialLiquidity: osmomath.OneDec(),

			lowerTickSpreadRewardGrowthOutside: cl.EmptyCoins,
			upperTickSpreadRewardGrowthOutside: cl.EmptyCoins,

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			lowerTick:           0,
			upperTick:           1,
			positionIdToPrepare: DefaultPositionId,

			currentTick: 5,

			expectedInitAccumValue: sdk.DecCoins(nil),
		},
		"swap occurs below the position, current tick < lower tick": {
			initialLiquidity: osmomath.OneDec(),

			lowerTickSpreadRewardGrowthOutside: cl.EmptyCoins,
			upperTickSpreadRewardGrowthOutside: cl.EmptyCoins,

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			lowerTick:           -10,
			upperTick:           -4,
			positionIdToPrepare: DefaultPositionId,

			currentTick: -13,

			expectedInitAccumValue: sdk.DecCoins(nil),
		},

		// error cases.
		"position does not exist": {
			initialLiquidity: osmomath.OneDec(),

			lowerTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(0))),
			upperTickSpreadRewardGrowthOutside: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

			globalSpreadRewardGrowth: sdk.NewDecCoins(sdk.NewDecCoin(ETH, osmomath.NewInt(10))),

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

			// Upscale accumulator value
			tc.lowerTickSpreadRewardGrowthOutside = tc.lowerTickSpreadRewardGrowthOutside.MulDecTruncate(cl.PerUnitLiqScalingFactor)
			tc.upperTickSpreadRewardGrowthOutside = tc.upperTickSpreadRewardGrowthOutside.MulDecTruncate(cl.PerUnitLiqScalingFactor)
			tc.expectedInitAccumValue = tc.expectedInitAccumValue.MulDecTruncate(cl.PerUnitLiqScalingFactor)

			// Set the position in store.
			err := clKeeper.SetPosition(ctx, validPoolId, s.TestAccs[0], tc.lowerTick, tc.upperTick, time.Now().UTC(), tc.initialLiquidity, DefaultPositionId, DefaultUnderlyingLockId)
			s.Require().NoError(err)

			s.initializeSpreadRewardAccumulatorPositionWithLiquidity(ctx, validPoolId, tc.lowerTick, tc.upperTick, DefaultPositionId, tc.initialLiquidity)
			s.initializeTick(ctx, tc.lowerTick, tc.initialLiquidity, tc.lowerTickSpreadRewardGrowthOutside, emptyUptimeTrackers, false)
			s.initializeTick(ctx, tc.upperTick, tc.initialLiquidity, tc.upperTickSpreadRewardGrowthOutside, emptyUptimeTrackers, true)
			validPool.SetCurrentTick(tc.currentTick)

			_ = clKeeper.SetPool(ctx, validPool)

			s.AddToSpreadRewardAccumulator(validPoolId, tc.globalSpreadRewardGrowth[0])

			positionKey := types.KeySpreadRewardPositionAccumulator(DefaultPositionId)

			// Note the position accumulator before calling prepare
			originalAccum, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(ctx, validPoolId)
			s.Require().NoError(err)

			originalAccumValue := originalAccum.GetValue()

			// System under test
			actualSpreadRewardsClaimed, err := clKeeper.PrepareClaimableSpreadRewards(ctx, tc.positionIdToPrepare)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				s.Require().Equal(sdk.Coins(nil), actualSpreadRewardsClaimed)
				return
			}
			s.Require().NoError(err)

			accum, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(ctx, validPoolId)
			s.Require().NoError(err)

			postPreparePosition, err := accum.GetPosition(positionKey)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedInitAccumValue, postPreparePosition.AccumValuePerShare)
			s.Require().Equal(tc.initialLiquidity, postPreparePosition.NumShares)

			expectedClaimedAmountDec := tc.expectedInitAccumValue.AmountOf(ETH).QuoTruncate(cl.PerUnitLiqScalingFactor).Mul(tc.initialLiquidity)
			expectedSpreadRewardClaimAmount := expectedClaimedAmountDec.TruncateInt()
			s.Require().Equal(expectedSpreadRewardClaimAmount, actualSpreadRewardsClaimed.AmountOf(ETH))

			// validate that truncated dust amount is reinvested back into the global accumulator
			if expectedClaimedAmountDec.GT(expectedSpreadRewardClaimAmount.ToLegacyDec()) {
				accumDelta, _ := accum.GetValue().SafeSub(originalAccumValue)
				s.Require().Equal(tc.expectedReinvestedDustAmount, accumDelta.AmountOf(ETH))
			}
		})
	}
}

// This test ensures that the position's spread reward accumulator is updated correctly when the spread reward grows.
// It validates that another position within the same tick does not affect the current position.
// It also validates that the position's changes are applied at the right time relative to position's
// spread reward accumulator creation or update.
func (s *KeeperTestSuite) TestInitOrUpdateSpreadRewardAccumulatorPosition_UpdatingPosition() {
	type updateSpreadRewardAccumPositionTest struct {
		doesSpreadRewardGrowBeforeFirstCall           bool
		doesSpreadRewardGrowBetweenFirstAndSecondCall bool
		doesSpreadRewardGrowBetweenSecondAndThirdCall bool
		doesSpreadRewardGrowAfterThirdCall            bool

		expectedUnclaimedRewardsPositionOne sdk.DecCoins
		expectedUnclaimedRewardsPositionTwo sdk.DecCoins
	}

	tests := map[string]updateSpreadRewardAccumPositionTest{
		"1: spread reward charged prior to first call to InitOrUpdateSpreadRewardAccumulatorPosition with position one": {
			doesSpreadRewardGrowBeforeFirstCall: true,

			// Growing spread reward before first position has no effect on the unclaimed rewards
			// of either position because they are not initialized at that point.
			expectedUnclaimedRewardsPositionOne: cl.EmptyCoins,
			// For position two, growing spread reward has no effect on the unclaimed rewards
			// because we never update it, only create it.
			expectedUnclaimedRewardsPositionTwo: cl.EmptyCoins,
		},
		"2: spread reward charged between first and second call to InitOrUpdateSpreadRewardAccumulatorPosition, after position one is created and before position two is created": {
			doesSpreadRewardGrowBetweenFirstAndSecondCall: true,

			// Position one's unclaimed rewards increase.
			expectedUnclaimedRewardsPositionOne: DefaultSpreadRewardAccumCoins,
			// For position two, growing spread reward has no effect on the unclaimed rewards
			// because we never update it, only create it.
			expectedUnclaimedRewardsPositionTwo: cl.EmptyCoins,
		},
		"3: spread reward charged between second and third call to InitOrUpdateSpreadRewardAccumulatorPosition, after position two is created and before position 1 is updated": {
			doesSpreadRewardGrowBetweenSecondAndThirdCall: true,

			// spread reward charged because it grows between the second and third position being created.
			// when third position is created, the rewards are moved to unclaimed.
			expectedUnclaimedRewardsPositionOne: DefaultSpreadRewardAccumCoins,
			// For position two, growing spread reward has no effect on the unclaimed rewards
			// because we never update it, only create it.
			expectedUnclaimedRewardsPositionTwo: cl.EmptyCoins,
		},
		"4: spread reward charged after third call to InitOrUpdateSpreadRewardAccumulatorPosition, after position 1 is updated": {
			doesSpreadRewardGrowAfterThirdCall: true,

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
			if tc.doesSpreadRewardGrowBeforeFirstCall {
				s.crossTickAndChargeSpreadReward(poolId, DefaultLowerTick)
			}

			_, err := s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, poolId, DefaultLowerTick, DefaultLiquidityAmt, false)
			s.Require().NoError(err)

			_, err = s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(s.Ctx, poolId, DefaultUpperTick, DefaultLiquidityAmt, true)
			s.Require().NoError(err)

			// InitOrUpdateSpreadRewardAccumulatorPosition #1 lower tick to upper tick
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdatePositionSpreadRewardAccumulator(s.Ctx, poolId, DefaultLowerTick, DefaultUpperTick, DefaultPositionId, DefaultLiquidityAmt)
			s.Require().NoError(err)

			// Imaginary spread reward charge #2.
			if tc.doesSpreadRewardGrowBetweenFirstAndSecondCall {
				s.crossTickAndChargeSpreadReward(poolId, DefaultLowerTick)
			}

			// InitOrUpdateSpreadRewardAccumulatorPosition # 2 lower tick to upper tick with a different position id.
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdatePositionSpreadRewardAccumulator(s.Ctx, poolId, DefaultLowerTick, DefaultUpperTick, DefaultPositionId+1, DefaultLiquidityAmt)
			s.Require().NoError(err)

			// Imaginary spread reward charge #3.
			if tc.doesSpreadRewardGrowBetweenSecondAndThirdCall {
				s.crossTickAndChargeSpreadReward(poolId, DefaultLowerTick)
			}

			// InitOrUpdateSpreadRewardAccumulatorPosition # 3 lower tick to upper tick with the original position id.
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdatePositionSpreadRewardAccumulator(s.Ctx, poolId, DefaultLowerTick, DefaultUpperTick, DefaultPositionId, DefaultLiquidityAmt)
			s.Require().NoError(err)

			// Imaginary spread reward charge #4.
			if tc.doesSpreadRewardGrowAfterThirdCall {
				s.crossTickAndChargeSpreadReward(poolId, DefaultLowerTick)
			}

			// Validate original position's spread reward growth.
			s.validatePositionSpreadRewardGrowth(poolId, DefaultPositionId, tc.expectedUnclaimedRewardsPositionOne)

			// Validate second position's spread reward growth.
			s.validatePositionSpreadRewardGrowth(poolId, DefaultPositionId+1, tc.expectedUnclaimedRewardsPositionTwo)

			// Validate position one was updated with default liquidity twice.
			s.validatePositionSpreadRewardAccUpdate(s.Ctx, poolId, DefaultPositionId, DefaultLiquidityAmt.MulInt64(2))

			// Validate position two was updated with default liquidity once.
			s.validatePositionSpreadRewardAccUpdate(s.Ctx, poolId, DefaultPositionId+1, DefaultLiquidityAmt)
		})
	}
}

func (s *KeeperTestSuite) TestUpdatePosValueToInitValuePlusGrowthOutside() {
	validPositionKey := types.KeySpreadRewardPositionAccumulator(1)
	invalidPositionKey := types.KeySpreadRewardPositionAccumulator(2)
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
			poolSpreadRewardAccumulator, err := clKeeper.GetSpreadRewardAccumulator(s.Ctx, defaultPoolId)
			s.Require().NoError(err)
			positionKey := validPositionKey

			// Initialize position accumulator.
			err = poolSpreadRewardAccumulator.NewPosition(positionKey, osmomath.OneDec(), nil)
			s.Require().NoError(err)

			// Record the initial position accumulator value.
			positionPre, err := accum.GetPosition(poolSpreadRewardAccumulator, positionKey)
			s.Require().NoError(err)

			// If the test case requires an invalid position key, set it.
			if tc.invalidPositionKey {
				positionKey = invalidPositionKey
			}

			// System under test.
			err = cl.UpdatePosValueToInitValuePlusGrowthOutside(poolSpreadRewardAccumulator, positionKey, tc.spreadRewardGrowthOutside)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}
			s.Require().NoError(err)

			// Record the final position accumulator value.
			positionPost, err := accum.GetPosition(poolSpreadRewardAccumulator, positionKey)
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

func (s *KeeperTestSuite) TestFunctional_SpreadRewards_Swaps() {
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
	clPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, osmomath.MustNewDecFromStr("0.002"))

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
	ticksActivatedAfterEachSwap, totalSpreadRewardsExpected, _, _ := s.swapAndTrackXTimesInARow(clPool.GetId(), DefaultCoin1, ETH, types.MaxSpotPriceBigDec, positions.numSwaps)
	s.CollectAndAssertSpreadRewards(s.Ctx, clPool.GetId(), totalSpreadRewardsExpected, positionIds, [][]int64{ticksActivatedAfterEachSwap}, onlyUSDC, positions)

	// Swap multiple times ETH for USDC, therefore decreasing the spot price
	ticksActivatedAfterEachSwap, totalSpreadRewardsExpected, _, _ = s.swapAndTrackXTimesInARow(clPool.GetId(), DefaultCoin0, USDC, types.MinSpotPriceBigDec, positions.numSwaps)
	s.CollectAndAssertSpreadRewards(s.Ctx, clPool.GetId(), totalSpreadRewardsExpected, positionIds, [][]int64{ticksActivatedAfterEachSwap}, onlyETH, positions)

	// Do the same swaps as before, however this time we collect spread rewards after both swap directions are complete.
	ticksActivatedAfterEachSwapUp, totalSpreadRewardsExpectedUp, _, _ := s.swapAndTrackXTimesInARow(clPool.GetId(), DefaultCoin1, ETH, types.MaxSpotPriceBigDec, positions.numSwaps)
	ticksActivatedAfterEachSwapDown, totalSpreadRewardsExpectedDown, _, _ := s.swapAndTrackXTimesInARow(clPool.GetId(), DefaultCoin0, USDC, types.MinSpotPriceBigDec, positions.numSwaps)
	totalSpreadRewardsExpected = totalSpreadRewardsExpectedUp.Add(totalSpreadRewardsExpectedDown...)

	// We expect all positions to have both denoms in their spread reward accumulators except USDC for the overlapping range position since
	// it was not activated during the USDC -> ETH swap direction but was activated during the ETH -> USDC swap direction.
	ticksActivatedAfterEachSwapTest := [][]int64{ticksActivatedAfterEachSwapUp, ticksActivatedAfterEachSwapDown}
	denomsExpected := [][]string{{USDC, ETH}, {USDC, ETH}, {USDC, ETH}, {NoUSDCExpected, ETH}}

	s.CollectAndAssertSpreadRewards(s.Ctx, clPool.GetId(), totalSpreadRewardsExpected, positionIds, ticksActivatedAfterEachSwapTest, denomsExpected, positions)
}

// This test focuses on various functional testing around spread rewards and LP logic.
// It tests invariants such as the following:
// - can create positions in the same range, swap between them and yet collect the correct spread rewards.
// - correct proportions of spread rewards for overlapping positions are withdrawn.
// - withdrawing full liquidity claims correctly under the hood.
// - withdrawing partial liquidity does not withdraw but still lets spread reward claim as desired.
func (s *KeeperTestSuite) TestFunctional_SpreadRewards_LP() {
	// Setup.
	s.SetupTest()
	s.TestAccs = apptesting.CreateRandomAccounts(5)

	var (
		ctx                         = s.Ctx
		concentratedLiquidityKeeper = s.App.ConcentratedLiquidityKeeper
		owner                       = s.TestAccs[0]
	)

	// Create pool with 0.2% spread factor.
	pool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, osmomath.MustNewDecFromStr("0.002"))
	fundCoins := sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0.MulRaw(2)), sdk.NewCoin(USDC, DefaultAmt1.MulRaw(2)))
	s.FundAcc(owner, fundCoins)

	// Errors since no position.
	_, _, _, err := s.App.ConcentratedLiquidityKeeper.SwapOutAmtGivenIn(s.Ctx, owner, pool, sdk.NewCoin(ETH, osmomath.OneInt()), USDC, pool.GetSpreadFactor(s.Ctx), types.MaxSpotPriceBigDec)
	s.Require().Error(err)

	// Create position in the default range 1.
	positionDataOne, err := concentratedLiquidityKeeper.CreatePosition(ctx, pool.GetId(), owner, DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
	s.Require().NoError(err)

	// Swap once.
	ticksActivatedAfterEachSwap, totalSpreadRewardsExpected, _, _ := s.swapAndTrackXTimesInARow(pool.GetId(), DefaultCoin1, ETH, types.MaxSpotPriceBigDec, 1)

	// Withdraw half.
	halfLiquidity := positionDataOne.Liquidity.Mul(osmomath.NewDecWithPrec(5, 1))
	_, _, err = concentratedLiquidityKeeper.WithdrawPosition(ctx, owner, positionDataOne.ID, halfLiquidity)
	s.Require().NoError(err)

	// Collect spread rewards.
	spreadRewardsCollected := s.collectSpreadRewardsAndCheckInvariance(ctx, 0, DefaultMinTick, DefaultMaxTick, positionDataOne.ID, sdk.NewCoins(), []string{USDC}, [][]int64{ticksActivatedAfterEachSwap})
	expectedSpreadRewardsTruncated := totalSpreadRewardsExpected
	for i, spreadRewardToken := range totalSpreadRewardsExpected {
		// We run expected spread rewards through a cycle of division and multiplication by liquidity to capture appropriate rounding behavior
		expectedSpreadRewardsTruncated[i] = sdk.NewCoin(spreadRewardToken.Denom, spreadRewardToken.Amount.ToLegacyDec().QuoTruncate(positionDataOne.Liquidity).MulTruncate(positionDataOne.Liquidity).TruncateInt())
	}
	s.Require().Equal(expectedSpreadRewardsTruncated, spreadRewardsCollected)

	// Unclaimed rewards should be emptied since spread rewards were collected.
	s.validatePositionSpreadRewardGrowth(pool.GetId(), positionDataOne.ID, cl.EmptyCoins)

	// Create position in the default range 2.
	positionDataTwo, err := concentratedLiquidityKeeper.CreatePosition(ctx, pool.GetId(), owner, DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
	s.Require().NoError(err)
	fullLiquidity := positionDataTwo.Liquidity

	// Swap once in the other direction.
	ticksActivatedAfterEachSwap, totalSpreadRewardsExpected, _, _ = s.swapAndTrackXTimesInARow(pool.GetId(), DefaultCoin0, USDC, types.MinSpotPriceBigDec, 1)

	// This should claim under the hood for position 2 since full liquidity is removed.
	balanceBeforeWithdraw := s.App.BankKeeper.GetBalance(ctx, owner, ETH)
	amtDenom0, _, err := concentratedLiquidityKeeper.WithdrawPosition(ctx, owner, positionDataTwo.ID, positionDataTwo.Liquidity)
	s.Require().NoError(err)
	balanceAfterWithdraw := s.App.BankKeeper.GetBalance(ctx, owner, ETH)

	// Validate that the correct amount of ETH was collected in withdraw for position two.
	// total spread rewards * full liquidity / (full liquidity + half liquidity)
	expectedPositionToWithdraw := totalSpreadRewardsExpected.AmountOf(ETH).ToLegacyDec().Mul(fullLiquidity.Quo(fullLiquidity.Add(halfLiquidity))).TruncateInt()
	s.Require().Equal(expectedPositionToWithdraw.String(), balanceAfterWithdraw.Sub(balanceBeforeWithdraw).Amount.Sub(amtDenom0).String())

	// Validate cannot claim for withdrawn position.
	_, err = s.App.ConcentratedLiquidityKeeper.CollectSpreadRewards(ctx, owner, positionDataTwo.ID)
	s.Require().Error(err)

	spreadRewardsCollected = s.collectSpreadRewardsAndCheckInvariance(ctx, 0, DefaultMinTick, DefaultMaxTick, positionDataOne.ID, sdk.NewCoins(), []string{ETH}, [][]int64{ticksActivatedAfterEachSwap})

	// total spread rewards * half liquidity / (full liquidity + half liquidity)
	expectesSpreadRewardsCollected := totalSpreadRewardsExpected.AmountOf(ETH).ToLegacyDec().Mul(halfLiquidity.Quo(fullLiquidity.Add(halfLiquidity))).TruncateInt()
	s.Require().Equal(expectesSpreadRewardsCollected.String(), spreadRewardsCollected.AmountOf(ETH).String())

	// Create position in the default range 3.
	positionDataThree, err := concentratedLiquidityKeeper.CreatePosition(ctx, pool.GetId(), owner, DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
	s.Require().NoError(err)

	collectedThree, err := s.App.ConcentratedLiquidityKeeper.CollectSpreadRewards(ctx, owner, positionDataThree.ID)
	s.Require().NoError(err)
	s.Require().Equal(sdk.Coins{}, collectedThree)
}

// CollectAndAssertSpreadRewards collects spread rewards from a given pool for all positions and verifies that the total spread rewards collected match the expected total spread rewards.
// The method also checks that if the ticks that were active during the swap lie within the range of a position, then the position's spread reward accumulators
// are not empty. The total spread rewards collected are compared to the expected total spread rewards within an additive tolerance defined by an error tolerance struct.
func (s *KeeperTestSuite) CollectAndAssertSpreadRewards(ctx sdk.Context, poolId uint64, totalSpreadRewards sdk.Coins, positionIds [][]uint64, activeTicks [][]int64, expectedSpreadRewardDenoms [][]string, positions Positions) {
	var totalSpreadRewardsCollected sdk.Coins
	// Claim full range position spread rewards across all four accounts
	for i := 0; i < positions.numFullRange; i++ {
		totalSpreadRewardsCollected = s.collectSpreadRewardsAndCheckInvariance(ctx, i, DefaultMinTick, DefaultMaxTick, positionIds[0][i], totalSpreadRewardsCollected, expectedSpreadRewardDenoms[0], activeTicks)
	}

	// Claim narrow range position spread rewards across three of four accounts
	for i := 0; i < positions.numNarrowRange; i++ {
		totalSpreadRewardsCollected = s.collectSpreadRewardsAndCheckInvariance(ctx, i, DefaultLowerTick, DefaultUpperTick, positionIds[1][i], totalSpreadRewardsCollected, expectedSpreadRewardDenoms[1], activeTicks)
	}

	// Claim consecutive range position spread rewards across two of four accounts
	for i := 0; i < positions.numConsecutive; i++ {
		totalSpreadRewardsCollected = s.collectSpreadRewardsAndCheckInvariance(ctx, i, DefaultExponentConsecutivePositionLowerTick, DefaultExponentConsecutivePositionUpperTick, positionIds[2][i], totalSpreadRewardsCollected, expectedSpreadRewardDenoms[2], activeTicks)
	}

	// Claim overlapping range position spread rewards on one of four accounts
	for i := 0; i < positions.numOverlapping; i++ {
		totalSpreadRewardsCollected = s.collectSpreadRewardsAndCheckInvariance(ctx, i, DefaultExponentOverlappingPositionLowerTick, DefaultExponentOverlappingPositionUpperTick, positionIds[3][i], totalSpreadRewardsCollected, expectedSpreadRewardDenoms[3], activeTicks)
	}

	// Define error tolerance
	var errTolerance osmomath.ErrTolerance
	errTolerance.AdditiveTolerance = osmomath.NewDec(10)
	errTolerance.RoundingDir = osmomath.RoundDown

	// Check that the total spread rewards collected is equal to the total spread rewards (within a tolerance)
	for _, coin := range totalSpreadRewardsCollected {
		expected := totalSpreadRewards.AmountOf(coin.Denom)
		actual := coin.Amount
		osmoassert.Equal(s.T(), errTolerance, expected, actual)
	}
}

// tickStatusInvariance tests if the swap position was active during the given tick range and
// checks that the spread rewards collected are non-zero if the position was active, or zero otherwise.
func (s *KeeperTestSuite) tickStatusInvariance(ticksActivatedAfterEachSwap [][]int64, lowerTick, upperTick int64, coins sdk.Coins, expectedSpreadRewardDenoms []string) {
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
			if expectedSpreadRewardDenoms[i] != NoUSDCExpected && expectedSpreadRewardDenoms[i] != NoETHExpected {
				s.Require().True(coins.AmountOf(expectedSpreadRewardDenoms[i]).GT(osmomath.ZeroInt()), "denom: %s", expectedSpreadRewardDenoms[i])
			}
		} else {
			// If the position was not active, check that the spread rewards collected are zero
			s.Require().Equal(sdk.Coins{}, coins)
		}
	}
}

// swapAndTrackXTimesInARow performs `numSwaps` swaps and tracks the tick activated after each swap.
// It also returns the total spread rewards collected, the total token in, and the total token out.
func (s *KeeperTestSuite) swapAndTrackXTimesInARow(poolId uint64, coinIn sdk.Coin, coinOutDenom string, priceLimit osmomath.BigDec, numSwaps int) (ticksActivatedAfterEachSwap []int64, totalSpreadRewards sdk.Coins, totalTokenIn sdk.Coin, totalTokenOut sdk.Coin) {
	// Retrieve pool
	clPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	// Determine amount needed to fulfill swap numSwaps times and fund account that much
	amountNeededForSwap := coinIn.Amount.Mul(osmomath.NewInt(int64(numSwaps)))
	swapCoin := sdk.NewCoin(coinIn.Denom, amountNeededForSwap)
	s.FundAcc(s.TestAccs[4], sdk.NewCoins(swapCoin))

	ticksActivatedAfterEachSwap = append(ticksActivatedAfterEachSwap, clPool.GetCurrentTick())
	totalSpreadRewards = sdk.NewCoins(sdk.NewCoin(USDC, osmomath.ZeroInt()), sdk.NewCoin(ETH, osmomath.ZeroInt()))
	// Swap numSwaps times, recording the tick activated after and swap and spread rewards we expect to collect based on the amount in
	totalTokenIn = sdk.NewCoin(coinIn.Denom, osmomath.ZeroInt())
	totalTokenOut = sdk.NewCoin(coinOutDenom, osmomath.ZeroInt())
	for i := 0; i < numSwaps; i++ {
		coinIn, coinOut, _, err := s.App.ConcentratedLiquidityKeeper.SwapOutAmtGivenIn(s.Ctx, s.TestAccs[4], clPool, coinIn, coinOutDenom, clPool.GetSpreadFactor(s.Ctx), priceLimit)
		s.Require().NoError(err)
		spreadReward := coinIn.Amount.ToLegacyDec().Mul(clPool.GetSpreadFactor(s.Ctx))
		totalSpreadRewards = totalSpreadRewards.Add(sdk.NewCoin(coinIn.Denom, spreadReward.TruncateInt()))
		clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPool.GetId())
		s.Require().NoError(err)
		ticksActivatedAfterEachSwap = append(ticksActivatedAfterEachSwap, clPool.GetCurrentTick())
		totalTokenIn = totalTokenIn.Add(coinIn)
		totalTokenOut = totalTokenOut.Add(coinOut)
	}
	return ticksActivatedAfterEachSwap, totalSpreadRewards, totalTokenIn, totalTokenOut
}

// collectSpreadRewardsAndCheckInvariance collects spread rewards from the concentrated liquidity pool and checks the resulting tick status invariance.
func (s *KeeperTestSuite) collectSpreadRewardsAndCheckInvariance(ctx sdk.Context, accountIndex int, minTick, maxTick int64, positionId uint64, spreadRewardsCollected sdk.Coins, expectedSpreadRewardDenoms []string, activeTicks [][]int64) (totalSpreadRewardsCollected sdk.Coins) {
	coins, err := s.App.ConcentratedLiquidityKeeper.CollectSpreadRewards(ctx, s.TestAccs[accountIndex], positionId)
	s.Require().NoError(err)
	totalSpreadRewardsCollected = spreadRewardsCollected.Add(coins...)
	s.tickStatusInvariance(activeTicks, minTick, maxTick, coins, expectedSpreadRewardDenoms)
	return totalSpreadRewardsCollected
}

func (s *KeeperTestSuite) TestScaleDownSpreadRewardAmount() {
	tests := []struct {
		name            string
		incentiveAmount osmomath.Int
		scalingFactor   osmomath.Dec
		expectedAmount  osmomath.Int
	}{
		{
			name:            "PerUnitLiqScalingFactor with no remainder",
			incentiveAmount: osmomath.MustNewDecFromStr("123456789").Mul(apptesting.PerUnitLiqScalingFactor).TruncateInt(),
			scalingFactor:   apptesting.PerUnitLiqScalingFactor,
			expectedAmount:  osmomath.NewInt(123456789),
		},
		{
			name:            "PerUnitLiqScalingFactor with remainder",
			incentiveAmount: osmomath.MustNewDecFromStr("123456789.123456789123456789").Mul(apptesting.PerUnitLiqScalingFactor).TruncateInt(),
			scalingFactor:   apptesting.PerUnitLiqScalingFactor,
			expectedAmount:  osmomath.NewInt(123456789),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			scaledAmount := cl.ScaleDownSpreadRewardAmount(test.incentiveAmount, test.scalingFactor)
			s.Require().Equal(test.expectedAmount, scaledAmount, "scaledAmount does not match")
		})
	}
}
