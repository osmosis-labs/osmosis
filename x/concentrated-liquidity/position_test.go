package concentrated_liquidity_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

var DefaultIncentiveRecords = []types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo, incentiveRecordThree, incentiveRecordFour}

func (s *KeeperTestSuite) TestInitOrUpdatePosition() {
	const (
		validPoolId   = 1
		invalidPoolId = 2
	)
	defaultJoinTime := s.Ctx.BlockTime()
	supportedUptimes := types.SupportedUptimes
	emptyAccumValues := getExpectedUptimes().emptyExpectedAccumValues
	type param struct {
		poolId         uint64
		lowerTick      int64
		upperTick      int64
		joinTime       time.Time
		positionId     uint64
		liquidityDelta sdk.Dec
	}

	tests := []struct {
		name                 string
		param                param
		positionExists       bool
		timeElapsedSinceInit time.Duration
		incentiveRecords     []types.IncentiveRecord
		expectedLiquidity    sdk.Dec
		expectedErr          error
	}{
		{
			name: "Init position from -50 to 50 with DefaultLiquidityAmt liquidity",
			param: param{
				poolId:         validPoolId,
				lowerTick:      -50,
				upperTick:      50,
				liquidityDelta: DefaultLiquidityAmt,
				positionId:     1,
				joinTime:       defaultJoinTime,
			},
			timeElapsedSinceInit: time.Hour,
			incentiveRecords:     DefaultIncentiveRecords,
			positionExists:       false,
			expectedLiquidity:    DefaultLiquidityAmt,
		},
		{
			name: "Update position from -50 to 50 that already contains DefaultLiquidityAmt liquidity with DefaultLiquidityAmt more liquidity",
			param: param{
				poolId:         validPoolId,
				lowerTick:      -50,
				upperTick:      50,
				liquidityDelta: DefaultLiquidityAmt,
				positionId:     1,
				joinTime:       defaultJoinTime,
			},
			positionExists:    true,
			expectedLiquidity: DefaultLiquidityAmt.Add(DefaultLiquidityAmt),
		},
		{
			name: "Init position for non-existing pool",
			param: param{
				poolId:         invalidPoolId,
				lowerTick:      -50,
				upperTick:      50,
				liquidityDelta: DefaultLiquidityAmt,
				positionId:     1,
				joinTime:       defaultJoinTime,
			},
			positionExists: false,
			expectedErr:    types.PoolNotFoundError{PoolId: 2},
		},
		{
			name: "Init position from -50 to 50 with negative DefaultLiquidityAmt liquidity",
			param: param{
				poolId:         validPoolId,
				lowerTick:      -50,
				upperTick:      50,
				liquidityDelta: DefaultLiquidityAmt.Neg(),
				positionId:     1,
				joinTime:       defaultJoinTime,
			},
			positionExists: false,
			expectedErr:    types.NegativeLiquidityError{Liquidity: DefaultLiquidityAmt.Neg()},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()

			// Set blocktime to fixed UTC value for consistency
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)

			// Create a default CL pool
			clPool := s.PrepareConcentratedPool()

			// We get initial uptime accum values for comparison later
			initUptimeAccumValues, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulatorValues(s.Ctx, test.param.poolId)
			if test.param.poolId == invalidPoolId {
				s.Require().Error(err)
				// Ensure that no accumulators are retrieved upon error
				s.Require().Equal([]sdk.DecCoins{}, initUptimeAccumValues)
			} else {
				s.Require().NoError(err)
				// Ensure initial uptime accums are empty
				s.Require().Equal(getExpectedUptimes().emptyExpectedAccumValues, initUptimeAccumValues)
			}

			// Set incentives for pool to ensure accumulators work correctly
			err = s.App.ConcentratedLiquidityKeeper.SetMultipleIncentiveRecords(s.Ctx, test.incentiveRecords)
			s.Require().NoError(err)

			// If positionExists set, initialize the specified position with defaultLiquidityAmt
			preexistingLiquidity := sdk.ZeroDec()
			if test.positionExists {
				// We let some fixed amount of time to elapse so we can ensure LastLiquidityUpdate time is
				// tracked properly even with no liquidity.
				s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime.Add(time.Minute * 5))

				err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePosition(s.Ctx, test.param.poolId, s.TestAccs[0], test.param.lowerTick, test.param.upperTick, test.param.liquidityDelta, test.param.joinTime, test.param.positionId)
				s.Require().NoError(err)
				preexistingLiquidity = test.param.liquidityDelta

				// Since this is the pool's initial liquidity, uptime accums should not have increased in value
				newUptimeAccumValues, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulatorValues(s.Ctx, test.param.poolId)
				s.Require().NoError(err)
				s.Require().Equal(initUptimeAccumValues, newUptimeAccumValues)

				// LastLiquidityUpdate time should be moved up nonetheless
				clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPool.GetId())
				s.Require().NoError(err)
				s.Require().Equal(s.Ctx.BlockTime(), clPool.GetLastLiquidityUpdate())
			}

			// Move up blocktime by time we want to elapse
			// We keep track of init blocktime to test error cases
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(test.timeElapsedSinceInit))

			// Get the position liquidity for poolId 1
			liquidity, err := s.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(s.Ctx, test.param.positionId)
			if test.positionExists {
				// If we had a position before, ensure the position info displays proper liquidity
				s.Require().NoError(err)
				s.Require().Equal(preexistingLiquidity, liquidity)
			} else {
				// If we did not have a position before, ensure getting the non-existent position returns an error
				s.Require().Error(err)
				s.Require().ErrorContains(err, types.PositionIdNotFoundError{PositionId: test.param.positionId}.Error())
			}

			// System under test. Initialize or update the position according to the test case
			err = s.App.ConcentratedLiquidityKeeper.InitOrUpdatePosition(s.Ctx, test.param.poolId, s.TestAccs[0], test.param.lowerTick, test.param.upperTick, test.param.liquidityDelta, test.param.joinTime, test.param.positionId)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, test.expectedErr.Error())

				// If the error is due to a nonexistent pool, we exit before pool-level checks
				if test.param.poolId == invalidPoolId {
					return
				}

				// Uptime accumulators should not be updated upon error
				newUptimeAccumValues, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulatorValues(s.Ctx, test.param.poolId)
				s.Require().NoError(err)
				s.Require().Equal(initUptimeAccumValues, newUptimeAccumValues)

				// LastLiquidityUpdate should not have moved up since init upon error
				clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPool.GetId())
				s.Require().NoError(err)
				s.Require().Equal(defaultJoinTime, clPool.GetLastLiquidityUpdate())
				return
			}
			s.Require().NoError(err)

			// Get the position liquidity for poolId 1
			liquidity, err = s.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(s.Ctx, test.param.positionId)
			s.Require().NoError(err)

			// Check that the initialized or updated position matches our expectation
			s.Require().Equal(test.expectedLiquidity, liquidity)

			// ---Tests for ensuring uptime accumulators behaved as expected---

			// Get updated accumulators and accum values
			newUptimeAccums, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, test.param.poolId)
			s.Require().NoError(err)
			newUptimeAccumValues, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulatorValues(s.Ctx, test.param.poolId)
			s.Require().NoError(err)
			expectedInitAccumValues, err := s.App.ConcentratedLiquidityKeeper.GetUptimeGrowthInsideRange(s.Ctx, clPool.GetId(), test.param.lowerTick, test.param.upperTick)
			s.Require().NoError(err)

			// Setup for checks
			actualUptimeAccumDelta, expectedUptimeAccumValueGrowth, expectedIncentiveRecords, _ := emptyAccumValues, emptyAccumValues, test.incentiveRecords, sdk.DecCoins{}

			timeElapsedSec := sdk.NewDec(int64(test.timeElapsedSinceInit)).Quo(sdk.NewDec(10e8))
			positionName := string(types.KeyPositionId(test.param.positionId))

			// Loop through each supported uptime for pool and ensure that:
			// 1. Position is properly updated on it
			// 2. Accum value has changed by the correct amount
			for uptimeIndex, uptime := range supportedUptimes {
				// Position-related checks

				recordExists, err := newUptimeAccums[uptimeIndex].HasPosition(positionName)
				s.Require().NoError(err)
				s.Require().True(recordExists)

				// Ensure position's record has correct values
				positionRecord, err := accum.GetPosition(newUptimeAccums[uptimeIndex], positionName)
				s.Require().NoError(err)

				// We expect the position's accum record to be initialized to the uptime growth *inside* its range
				s.Require().Equal(expectedInitAccumValues[uptimeIndex], positionRecord.AccumValuePerShare)
				s.Require().Equal(test.expectedLiquidity, positionRecord.NumShares)

				// Accumulator value related checks

				if test.positionExists {
					// Track how much the current uptime accum has grown by
					actualUptimeAccumDelta[uptimeIndex] = newUptimeAccumValues[uptimeIndex].Sub(initUptimeAccumValues[uptimeIndex])
					if timeElapsedSec.GT(sdk.ZeroDec()) {
						expectedGrowthCurAccum, _, err := cl.CalcAccruedIncentivesForAccum(s.Ctx, uptime, test.param.liquidityDelta, timeElapsedSec, expectedIncentiveRecords)
						s.Require().NoError(err)
						expectedUptimeAccumValueGrowth[uptimeIndex] = expectedGrowthCurAccum
					}
				} else {
					// if no position init, should remain empty
					s.Require().Equal(initUptimeAccumValues[uptimeIndex], newUptimeAccumValues[uptimeIndex])
				}
			}

			// Ensure uptime accumulators have grown by the expected amount
			s.Require().Equal(expectedUptimeAccumValueGrowth, actualUptimeAccumDelta)

			// Ensure incentive records have been properly updated in state. Note that we do a two-way contains check since records
			// get reordered lexicographically by denom in state.
			actualIncentiveRecords, err := s.App.ConcentratedLiquidityKeeper.GetAllIncentiveRecordsForPool(s.Ctx, test.param.poolId)
			s.Require().NoError(err)
			s.Require().ElementsMatch(expectedIncentiveRecords, actualIncentiveRecords)
		})
	}
}

func (s *KeeperTestSuite) TestGetPosition() {
	tests := []struct {
		name                      string
		positionId                uint64
		expectedPositionLiquidity sdk.Dec
		expectedErr               error
	}{
		{
			name:                      "Get position info on existing pool and existing position",
			positionId:                DefaultPositionId,
			expectedPositionLiquidity: DefaultLiquidityAmt,
		},
		{
			name:        "Get position info on a non-existent positionId",
			positionId:  DefaultPositionId + 1,
			expectedErr: types.PositionIdNotFoundError{PositionId: DefaultPositionId + 1},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			// Create a default CL pool
			s.PrepareConcentratedPool()

			// Set up a default initialized position
			err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePosition(s.Ctx, validPoolId, s.TestAccs[0], DefaultLowerTick, DefaultUpperTick, DefaultLiquidityAmt, DefaultJoinTime, DefaultPositionId)
			s.Require().NoError(err)

			// System under test
			position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, test.positionId)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, test.expectedErr)
				s.Require().Equal(sdk.Dec{}, position.Liquidity)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(test.expectedPositionLiquidity, position.Liquidity)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetNextPositionAndIncrement() {
	// Init suite for each test.
	s.SetupTest()
	s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)
	// Create a default CL pool
	pool := s.PrepareConcentratedPool()

	// Set up a default initialized position
	s.SetupDefaultPosition(pool.GetId())

	// System under test
	positionId := s.App.ConcentratedLiquidityKeeper.GetNextPositionIdAndIncrement(s.Ctx)
	s.Require().Equal(positionId, uint64(2))

	// try incrementing one more time
	positionId = s.App.ConcentratedLiquidityKeeper.GetNextPositionIdAndIncrement(s.Ctx)
	s.Require().Equal(positionId, uint64(3))
}

func (s *KeeperTestSuite) TestIsPositionOwner() {
	actualOwner := s.TestAccs[0]
	nonOwner := s.TestAccs[1]

	tests := []struct {
		name         string
		ownerToQuery sdk.AccAddress
		poolId       uint64
		positionId   uint64
		isOwner      bool
	}{
		{
			name:         "Happy path",
			ownerToQuery: actualOwner,
			poolId:       1,
			positionId:   DefaultPositionId,
			isOwner:      true,
		},
		{
			name:         "query non owner",
			ownerToQuery: nonOwner,
			poolId:       1,
			positionId:   DefaultPositionId,
			isOwner:      false,
		},
		{
			name:         "different pool ID, not the owner",
			ownerToQuery: actualOwner,
			poolId:       2,
			positionId:   DefaultPositionId,
			isOwner:      false,
		},
		{
			name:         "different position ID, not the owner",
			ownerToQuery: actualOwner,
			poolId:       1,
			positionId:   DefaultPositionId + 1,
			isOwner:      false,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			// Create a default CL pool.
			s.PrepareConcentratedPool()

			// Set up a default initialized position.
			err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePosition(s.Ctx, validPoolId, actualOwner, DefaultLowerTick, DefaultUpperTick, DefaultLiquidityAmt, DefaultJoinTime, DefaultPositionId)
			s.Require().NoError(err)

			// System under test.
			isOwner, err := s.App.ConcentratedLiquidityKeeper.IsPositionOwner(s.Ctx, test.ownerToQuery, test.poolId, test.positionId)
			s.Require().Equal(test.isOwner, isOwner)
			s.Require().NoError(err)
		})
	}
}

func (s *KeeperTestSuite) TestGetAllUserPositions() {
	s.Setup()
	defaultAddress := s.TestAccs[0]
	secondAddress := s.TestAccs[1]
	DefaultJoinTime := s.Ctx.BlockTime()
	type position struct {
		positionId uint64
		poolId     uint64
		acc        sdk.AccAddress
		coins      sdk.Coins
		lowerTick  int64
		upperTick  int64
		joinTime   time.Time
	}

	tests := []struct {
		name           string
		sender         sdk.AccAddress
		poolId         uint64
		setupPositions []position
		expectedErr    error
	}{
		{
			name:   "Get current user one position",
			sender: defaultAddress,
			setupPositions: []position{
				{1, 1, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, DefaultJoinTime},
			},
		},
		{
			name:   "Get current users multiple position same pool",
			sender: defaultAddress,
			setupPositions: []position{
				{1, 1, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, DefaultJoinTime},
				{2, 1, defaultAddress, DefaultCoins, DefaultLowerTick + 100, DefaultUpperTick + 100, DefaultJoinTime},
				{3, 1, defaultAddress, DefaultCoins, DefaultLowerTick + 200, DefaultUpperTick + 200, DefaultJoinTime},
			},
		},
		{
			name:   "Get current users multiple position multiple pools",
			sender: secondAddress,
			setupPositions: []position{
				{1, 1, secondAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, DefaultJoinTime},
				{2, 2, secondAddress, DefaultCoins, DefaultLowerTick + 100, DefaultUpperTick + 100, DefaultJoinTime},
				{3, 3, secondAddress, DefaultCoins, DefaultLowerTick + 200, DefaultUpperTick + 200, DefaultJoinTime},
			},
		},
		{
			name:   "User has positions over multiple pools, but filter by one pool",
			sender: secondAddress,
			poolId: 2,
			setupPositions: []position{
				{1, 1, secondAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, DefaultJoinTime},
				{2, 2, secondAddress, DefaultCoins, DefaultLowerTick + 100, DefaultUpperTick + 100, DefaultJoinTime},
				{3, 3, secondAddress, DefaultCoins, DefaultLowerTick + 200, DefaultUpperTick + 200, DefaultJoinTime},
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			// Create a default CL pools
			s.PrepareMultipleConcentratedPools(3)

			expectedUserPositions := []model.Position{}
			for _, pos := range test.setupPositions {
				// if position does not exist this errors
				liquidity, _ := s.SetupPosition(pos.poolId, pos.acc, pos.coins, pos.lowerTick, pos.upperTick, pos.joinTime)
				if pos.acc.Equals(pos.acc) {
					if test.poolId == 0 || test.poolId == pos.poolId {
						expectedUserPositions = append(expectedUserPositions, model.Position{
							PositionId: pos.positionId,
							PoolId:     pos.poolId,
							Address:    pos.acc.String(),
							LowerTick:  pos.lowerTick,
							UpperTick:  pos.upperTick,
							JoinTime:   pos.joinTime,
							Liquidity:  liquidity,
						})
					}
				}
			}

			// System under test
			position, err := s.App.ConcentratedLiquidityKeeper.GetUserPositions(s.Ctx, test.sender, test.poolId)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, test.expectedErr)
				s.Require().Nil(position)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(expectedUserPositions, position)
			}
		})
	}
}

func (s *KeeperTestSuite) TestDeletePosition() {
	defaultPoolId := uint64(1)
	DefaultJoinTime := s.Ctx.BlockTime()

	tests := []struct {
		name             string
		positionId       uint64
		underlyingLockId uint64
		expectedErr      error
	}{
		{
			name:             "Delete position info on existing pool and existing position (no underlying lock)",
			underlyingLockId: 0,
			positionId:       DefaultPositionId,
		},
		{
			name:             "Delete position info on existing pool and existing position (has underlying lock)",
			underlyingLockId: 1,
			positionId:       DefaultPositionId,
		},
		{
			name:             "Delete a non existing position",
			positionId:       DefaultPositionId + 1,
			underlyingLockId: 0,
			expectedErr:      types.PositionIdNotFoundError{PositionId: DefaultPositionId + 1},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)
			store := s.Ctx.KVStore(s.App.GetKey(types.StoreKey))

			// Create a default CL pool
			s.PrepareConcentratedPool()

			// Set up a default initialized position
			err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePosition(s.Ctx, validPoolId, s.TestAccs[0], DefaultLowerTick, DefaultUpperTick, DefaultLiquidityAmt, DefaultJoinTime, DefaultPositionId)
			s.Require().NoError(err)

			if test.underlyingLockId != 0 {
				err = s.App.ConcentratedLiquidityKeeper.SetPosition(s.Ctx, validPoolId, s.TestAccs[0], DefaultLowerTick, DefaultUpperTick, DefaultJoinTime, DefaultLiquidityAmt, 1, test.underlyingLockId)
				s.Require().NoError(err)
			}

			// Check stores exist
			// Retrieve the position from the store via position ID and compare to expected values.
			position := model.Position{}
			positionIdToPositionKey := types.KeyPositionId(DefaultPositionId)
			osmoutils.MustGet(store, positionIdToPositionKey, &position)
			s.Require().Equal(DefaultPositionId, position.PositionId)
			s.Require().Equal(defaultPoolId, position.PoolId)
			s.Require().Equal(s.TestAccs[0].String(), position.Address)
			s.Require().Equal(DefaultLowerTick, position.LowerTick)
			s.Require().Equal(DefaultUpperTick, position.UpperTick)
			s.Require().Equal(DefaultJoinTime, position.JoinTime)
			s.Require().Equal(DefaultLiquidityAmt, position.Liquidity)

			// Retrieve the position ID from the store via owner/poolId key and compare to expected values.
			ownerPoolIdToPositionIdKey := types.KeyAddressPoolIdPositionId(s.TestAccs[0], defaultPoolId, DefaultPositionId)
			positionIdBytes := store.Get(ownerPoolIdToPositionIdKey)
			s.Require().Equal(DefaultPositionId, sdk.BigEndianToUint64(positionIdBytes))

			// Retrieve the position ID from the store via poolId key and compare to expected values.
			poolIdtoPositionIdKey := types.KeyPoolPositionPositionId(defaultPoolId, DefaultPositionId)
			positionIdBytes = store.Get(poolIdtoPositionIdKey)
			s.Require().Equal(DefaultPositionId, sdk.BigEndianToUint64(positionIdBytes))

			// Retrieve the position ID to underlying lock ID mapping from the store and compare to expected values.
			positionIdToLockIdKey := types.KeyPositionIdForLock(DefaultPositionId)
			underlyingLockIdBytes := store.Get(positionIdToLockIdKey)
			if test.underlyingLockId != 0 {
				s.Require().Equal(test.underlyingLockId, sdk.BigEndianToUint64(underlyingLockIdBytes))
			} else {
				s.Require().Nil(underlyingLockIdBytes)
			}

			// Retrieve the lock ID to position ID mapping from the store and compare to expected values.
			lockIdToPositionIdKey := types.KeyLockIdForPositionId(test.underlyingLockId)
			positionIdBytes = store.Get(lockIdToPositionIdKey)
			if test.underlyingLockId != 0 {
				s.Require().Equal(DefaultPositionId, sdk.BigEndianToUint64(positionIdBytes))
			} else {
				s.Require().Nil(positionIdBytes)
			}

			err = s.App.ConcentratedLiquidityKeeper.DeletePosition(s.Ctx, test.positionId, s.TestAccs[0], defaultPoolId)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, test.expectedErr)
			} else {
				s.Require().NoError(err)

				// Since the positionLiquidity is deleted, retrieving it should return an error.
				positionLiquidity, err := s.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(s.Ctx, test.positionId)
				s.Require().Error(err)
				s.Require().ErrorIs(err, types.PositionIdNotFoundError{PositionId: test.positionId})
				s.Require().Equal(sdk.Dec{}, positionLiquidity)

				// Check that stores were deleted
				// Retrieve the position from the store via position ID and compare to expected values.
				position := model.Position{}
				positionIdToPositionKey := types.KeyPositionId(DefaultPositionId)
				_, err = osmoutils.Get(store, positionIdToPositionKey, &position)
				s.Require().NoError(err)
				s.Require().Equal(model.Position{}, position)

				// Retrieve the position ID from the store via owner/poolId key and compare to expected values.
				ownerPoolIdToPositionIdKey = types.KeyAddressPoolIdPositionId(s.TestAccs[0], defaultPoolId, DefaultPositionId)
				positionIdBytes := store.Get(ownerPoolIdToPositionIdKey)
				s.Require().Nil(positionIdBytes)

				// Retrieve the position ID from the store via poolId key and compare to expected values.
				poolIdtoPositionIdKey = types.KeyPoolPositionPositionId(defaultPoolId, DefaultPositionId)
				positionIdBytes = store.Get(poolIdtoPositionIdKey)
				s.Require().Nil(positionIdBytes)

				// Retrieve the position ID to underlying lock ID mapping from the store and compare to expected values.
				positionIdToLockIdKey = types.KeyPositionIdForLock(DefaultPositionId)
				underlyingLockIdBytes := store.Get(positionIdToLockIdKey)
				s.Require().Nil(underlyingLockIdBytes)

				// Retrieve the lock ID to position ID mapping from the store and compare to expected values.
				lockIdToPositionIdKey := types.KeyLockIdForPositionId(test.underlyingLockId)
				positionIdBytes = store.Get(lockIdToPositionIdKey)
				s.Require().Nil(positionIdBytes)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCalculateUnderlyingAssetsFromPosition() {
	tests := []struct {
		name            string
		position        model.Position
		isZeroLiquidity bool
	}{
		{
			name:     "Default range position",
			position: model.Position{PoolId: 1, LowerTick: DefaultLowerTick, UpperTick: DefaultUpperTick},
		},
		{
			name:            "Zero liquidity",
			isZeroLiquidity: true,
			position:        model.Position{PoolId: 1, LowerTick: DefaultLowerTick, UpperTick: DefaultUpperTick},
		},
		{
			name:     "Full range position",
			position: model.Position{PoolId: 1, LowerTick: DefaultMinTick, UpperTick: DefaultMaxTick},
		},
		{
			name:     "Below current tick position",
			position: model.Position{PoolId: 1, LowerTick: DefaultLowerTick, UpperTick: DefaultLowerTick + 100},
		},
		{
			name:     "Above current tick position",
			position: model.Position{PoolId: 1, LowerTick: DefaultUpperTick, UpperTick: DefaultUpperTick + 100},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// prepare concentrated pool with a default position
			s.PrepareConcentratedPool()
			s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0), sdk.NewCoin(USDC, DefaultAmt1)))
			_, _, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, 1, s.TestAccs[0], DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
			s.Require().NoError(err)

			// create a position from the test case
			s.FundAcc(s.TestAccs[1], sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0), sdk.NewCoin(USDC, DefaultAmt1)))
			_, actualAmount0, actualAmount1, liquidity, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, tc.position.PoolId, s.TestAccs[1], DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), tc.position.LowerTick, tc.position.UpperTick)
			s.Require().NoError(err)
			tc.position.Liquidity = liquidity

			if tc.isZeroLiquidity {
				// set the position liquidity to zero
				tc.position.Liquidity = sdk.ZeroDec()
				actualAmount0 = sdk.ZeroInt()
				actualAmount1 = sdk.ZeroInt()
			}

			// calculate underlying assets from the position
			clPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, tc.position.PoolId)
			s.Require().NoError(err)
			calculatedCoin0, calculatedCoin1, err := cl.CalculateUnderlyingAssetsFromPosition(s.Ctx, tc.position, clPool)

			s.Require().NoError(err)
			s.Require().Equal(calculatedCoin0.String(), sdk.NewCoin(clPool.GetToken0(), actualAmount0).String())
			s.Require().Equal(calculatedCoin1.String(), sdk.NewCoin(clPool.GetToken1(), actualAmount1).String())
		})
	}
}

func (s *KeeperTestSuite) TestValidateAndFungifyChargedPositions() {
	const (
		locked   = true
		unlocked = !locked
	)

	var (
		defaultAddress         = s.TestAccs[0]
		secondAddress          = s.TestAccs[1]
		defaultBlockTime       = time.Unix(1, 1).UTC()
		testFullChargeDuration = time.Hour * 24
	)

	type position struct {
		positionId uint64
		poolId     uint64
		acc        sdk.AccAddress
		coins      sdk.Coins
		lowerTick  int64
		upperTick  int64
		isLocked   bool
	}

	tests := []struct {
		name                       string
		setupFullyChargedPositions []position
		setupUnchargedPositions    []position
		lockPositionIds            []uint64
		positionIdsToMigrate       []uint64
		accountCallingMigration    sdk.AccAddress
		unlockBeforeBlockTimeMs    time.Duration
		expectedNewPositionId      uint64
		expectedErr                error
		doesValidatePass           bool
	}{
		{
			name: "Happy path: Fungify three fully charged positions",
			setupFullyChargedPositions: []position{
				{1, defaultPoolId, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, unlocked},
				{2, defaultPoolId, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, unlocked},
				{3, defaultPoolId, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, unlocked},
			},
			positionIdsToMigrate:    []uint64{1, 2, 3},
			accountCallingMigration: defaultAddress,
			expectedNewPositionId:   4,
		},
		{
			name: "Error: Fungify three positions, but one of them is not fully charged",
			setupFullyChargedPositions: []position{
				{1, defaultPoolId, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, unlocked},
				{2, defaultPoolId, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, unlocked},
			},
			setupUnchargedPositions: []position{
				{3, defaultPoolId, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, unlocked},
			},
			positionIdsToMigrate:    []uint64{1, 2, 3},
			accountCallingMigration: defaultAddress,
			expectedNewPositionId:   0,
			expectedErr:             types.PositionNotFullyChargedError{PositionId: 3, PositionJoinTime: defaultBlockTime.Add(testFullChargeDuration), FullyChargedMinTimestamp: defaultBlockTime.Add(testFullChargeDuration).Add(testFullChargeDuration)},
		},
		{
			name: "Error: Fungify three positions, but one of them is not in the same pool",
			setupFullyChargedPositions: []position{
				{1, defaultPoolId, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, unlocked},
				{2, 2, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, unlocked},
				{3, defaultPoolId, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, unlocked},
			},
			positionIdsToMigrate:    []uint64{1, 2, 3},
			accountCallingMigration: defaultAddress,
			expectedNewPositionId:   0,
			expectedErr:             types.PositionsNotInSamePoolError{Position1PoolId: 2, Position2PoolId: 1},
		},
		{
			name: "Error: Fungify three positions, but one of them is not owned by the same owner",
			setupFullyChargedPositions: []position{
				{1, defaultPoolId, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, unlocked},
				{2, defaultPoolId, secondAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, unlocked},
				{3, defaultPoolId, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, unlocked},
			},
			positionIdsToMigrate:    []uint64{1, 2, 3},
			accountCallingMigration: defaultAddress,
			expectedNewPositionId:   0,
			expectedErr:             types.PositionOwnerMismatchError{PositionOwner: secondAddress.String(), Sender: defaultAddress.String()},
		},
		{
			name: "Error: Fungify three positions, but one of them is not in the same range",
			setupFullyChargedPositions: []position{
				{1, defaultPoolId, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, unlocked},
				{2, defaultPoolId, defaultAddress, DefaultCoins, DefaultLowerTick - 100, DefaultUpperTick, unlocked},
				{3, defaultPoolId, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, unlocked},
			},
			positionIdsToMigrate:    []uint64{1, 2, 3},
			accountCallingMigration: defaultAddress,
			expectedNewPositionId:   0,
			expectedErr:             types.PositionsNotInSameTickRangeError{Position1TickLower: DefaultLowerTick - 100, Position1TickUpper: DefaultUpperTick, Position2TickLower: DefaultLowerTick, Position2TickUpper: DefaultUpperTick},
		},
		{
			name: "Error: Fungify one position, must have at least two",
			setupFullyChargedPositions: []position{
				{1, defaultPoolId, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick, unlocked},
			},
			setupUnchargedPositions: []position{},
			positionIdsToMigrate:    []uint64{1},
			accountCallingMigration: defaultAddress,
			expectedNewPositionId:   0,
			expectedErr:             types.PositionQuantityTooLowError{MinNumPositions: cl.MinNumPositions, NumPositions: 1},
			doesValidatePass:        true,
		},
		{
			name: "Error: one of the full range positions is locked",
			setupFullyChargedPositions: []position{
				{1, defaultPoolId, defaultAddress, DefaultCoins, types.MinTick, types.MaxTick, unlocked},
				{2, defaultPoolId, defaultAddress, DefaultCoins, types.MinTick, types.MaxTick, locked},
				{3, defaultPoolId, defaultAddress, DefaultCoins, types.MinTick, types.MaxTick, unlocked},
			},
			lockPositionIds:         []uint64{2},
			positionIdsToMigrate:    []uint64{1, 2, 3},
			accountCallingMigration: defaultAddress,
			expectedNewPositionId:   0,
			expectedErr:             types.LockNotMatureError{PositionId: 2, LockId: 1},
		},
		{
			name: "Pass: one of the full range positions was locked but got unlocked 1ms before fungification",
			setupFullyChargedPositions: []position{
				{1, defaultPoolId, defaultAddress, DefaultCoins, types.MinTick, types.MaxTick, unlocked},
				{2, defaultPoolId, defaultAddress, DefaultCoins, types.MinTick, types.MaxTick, locked},
				{3, defaultPoolId, defaultAddress, DefaultCoins, types.MinTick, types.MaxTick, unlocked},
			},

			lockPositionIds:         []uint64{2},
			positionIdsToMigrate:    []uint64{1, 2, 3},
			accountCallingMigration: defaultAddress,
			// Subtracting one millisecond from the block time (when it's supposed to be unlocked
			// by default, makes the lock mature)
			unlockBeforeBlockTimeMs: time.Millisecond * -1,
			expectedNewPositionId:   4,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultBlockTime)
			totalPositionsToCreate := sdk.NewInt(int64(len(test.setupFullyChargedPositions) + len(test.setupUnchargedPositions)))
			requiredBalances := sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0.Mul(totalPositionsToCreate)), sdk.NewCoin(USDC, DefaultAmt1.Mul(totalPositionsToCreate)))

			params := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
			params.AuthorizedUptimes = []time.Duration{time.Nanosecond, testFullChargeDuration}
			s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, params)

			// Fund accounts
			s.FundAcc(defaultAddress, requiredBalances)
			s.FundAcc(secondAddress, requiredBalances)

			// Create two default CL pools
			s.PrepareConcentratedPool()
			s.PrepareConcentratedPool()

			// Set incentives for pool to ensure accumulators work correctly
			err := s.App.ConcentratedLiquidityKeeper.SetMultipleIncentiveRecords(s.Ctx, DefaultIncentiveRecords)
			s.Require().NoError(err)

			// Set up fully charged positions
			totalLiquidity := sdk.ZeroDec()

			// See increases in the test below.
			// The reason we double testFullChargeDurationis is because that is by how much we increase block time in total
			// to set up the fully charged positions.
			lockDuration := testFullChargeDuration + testFullChargeDuration + test.unlockBeforeBlockTimeMs
			for _, pos := range test.setupFullyChargedPositions {
				var (
					liquidityCreated sdk.Dec
					err              error
				)
				if pos.isLocked {
					_, _, _, liquidityCreated, _, _, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionUnlocking(s.Ctx, pos.poolId, pos.acc, pos.coins, lockDuration)
					s.Require().NoError(err)
				} else {
					_, _, _, liquidityCreated, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, pos.poolId, pos.acc, pos.coins, sdk.ZeroInt(), sdk.ZeroInt(), pos.lowerTick, pos.upperTick)
					s.Require().NoError(err)
				}

				totalLiquidity = totalLiquidity.Add(liquidityCreated)
			}

			// Increase block time by the fully charged duration to make sure previously added positions are charged.
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(testFullChargeDuration))

			// Set up uncharged positions
			for _, pos := range test.setupUnchargedPositions {
				_, _, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, pos.poolId, pos.acc, pos.coins, sdk.ZeroInt(), sdk.ZeroInt(), pos.lowerTick, pos.upperTick)
				s.Require().NoError(err)
			}

			// Increase block time by one more day - 1 ns to ensure that the previously added positions are not fully charged.
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(testFullChargeDuration - time.Nanosecond))

			// First run non mutative validation and check results
			poolId, lowerTick, upperTick, liquidity, err := s.App.ConcentratedLiquidityKeeper.ValidatePositionsAndGetTotalLiquidity(s.Ctx, test.accountCallingMigration, test.positionIdsToMigrate)
			if test.expectedErr != nil && !test.doesValidatePass {
				s.Require().Error(err)
				s.Require().ErrorIs(err, test.expectedErr)
				s.Require().Equal(uint64(0), poolId)
				s.Require().Equal(int64(0), lowerTick)
				s.Require().Equal(int64(0), upperTick)
				s.Require().Equal(sdk.Dec{}, liquidity)
			} else {
				s.Require().NoError(err)

				// Check that the poolId, lowerTick, upperTick, and liquidity are correct
				for _, posId := range test.positionIdsToMigrate {
					position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, posId)
					s.Require().NoError(err)
					s.Require().Equal(poolId, position.PoolId)
					s.Require().Equal(lowerTick, position.LowerTick)
					s.Require().Equal(upperTick, position.UpperTick)
				}
				s.Require().Equal(totalLiquidity, liquidity)
			}

			// Update the accumulators for defaultPoolId to the current time
			err = s.App.ConcentratedLiquidityKeeper.UpdateUptimeAccumulatorsToNow(s.Ctx, defaultPoolId)
			s.Require().NoError(err)

			// Get the uptime accumulators for defaultPoolId
			uptimeAccumulators, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, defaultPoolId)
			s.Require().NoError(err)

			unclaimedRewardsForEachUptimeAcrossAllOldPositions := make([]sdk.DecCoins, len(uptimeAccumulators))

			// Get the unclaimed rewards for all the positions that are being migrated
			for _, positionId := range test.positionIdsToMigrate {
				oldPositionName := string(types.KeyPositionId(positionId))
				for i, uptimeAccum := range uptimeAccumulators {
					// Check if the accumulator contains the position.
					hasPosition, err := uptimeAccum.HasPosition(oldPositionName)
					s.Require().NoError(err)

					// If the accumulator contains the position, note the unclaimed rewards.
					if hasPosition {
						// Get the unclaimed rewards for the old position.
						position, err := accum.GetPosition(uptimeAccum, oldPositionName)
						s.Require().NoError(err)

						unclaimedRewardsForPosition := accum.GetTotalRewards(uptimeAccum, position)

						// Add the unclaimed rewards to the total unclaimed rewards for all the old positions.
						unclaimedRewardsForEachUptimeAcrossAllOldPositions[i] = unclaimedRewardsForEachUptimeAcrossAllOldPositions[i].Add(unclaimedRewardsForPosition...)
					}
				}
			}

			// Next, run the mutative function and check results
			newPositionId, err := s.App.ConcentratedLiquidityKeeper.FungifyChargedPosition(s.Ctx, test.accountCallingMigration, test.positionIdsToMigrate)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, test.expectedErr)
				s.Require().Equal(uint64(0), newPositionId)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(test.expectedNewPositionId, newPositionId)

				// Since the positionLiquidity of the old position should have been deleted, retrieving it should return an error.
				for _, posId := range test.positionIdsToMigrate {
					positionLiquidity, err := s.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(s.Ctx, posId)
					s.Require().Error(err)
					s.Require().ErrorIs(err, types.PositionIdNotFoundError{PositionId: posId})
					s.Require().Equal(sdk.Dec{}, positionLiquidity)
				}

				// Retrieve the new position and check that the liquidity is equal to the sum of the old positions.
				newPosition, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, newPositionId)
				s.Require().NoError(err)
				s.Require().Equal(totalLiquidity, newPosition.Liquidity)

				// The new position's join time should be the current block time minus the fully charged duration.
				fullCharge := s.App.ConcentratedLiquidityKeeper.GetLargestAuthorizedUptimeDuration(s.Ctx)
				s.Require().Equal(s.Ctx.BlockTime().Add(-fullCharge), newPosition.JoinTime)

				// Get the uptime accumulators for the poolId of the new position.
				uptimeAccumulators, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, poolId)
				s.Require().NoError(err)

				unclaimedRewardsForEachUptimeNewPosition := make([]sdk.DecCoins, len(uptimeAccumulators))

				// Get the unclaimed rewards for the new position
				for i, uptimeAccum := range uptimeAccumulators {
					newPositionName := string(types.KeyPositionId(newPositionId))
					// Check if the accumulator contains the position.
					hasPosition, err := uptimeAccum.HasPosition(newPositionName)
					s.Require().NoError(err)
					s.Require().True(hasPosition)

					// Move the unclaimed rewards to the new position.
					// Get the unclaimed rewards for the old position.
					position, err := accum.GetPosition(uptimeAccum, newPositionName)
					s.Require().NoError(err)

					unclaimedRewardsForPosition := accum.GetTotalRewards(uptimeAccum, position)

					unclaimedRewardsForEachUptimeNewPosition[i] = unclaimedRewardsForEachUptimeNewPosition[i].Add(unclaimedRewardsForPosition...)
				}

				// Check that the old uptime accumulators and positions have been deleted.
				for _, positionId := range test.positionIdsToMigrate {
					oldPositionName := string(types.KeyPositionId(positionId))
					for _, uptimeAccum := range uptimeAccumulators {
						// Check if the accumulator contains the position.
						hasPosition, err := uptimeAccum.HasPosition(oldPositionName)
						s.Require().NoError(err)
						s.Require().False(hasPosition)
					}

					// Check that the old position has been deleted.
					_, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionId)
					s.Require().Error(err)
					s.Require().ErrorAs(err, &types.PositionIdNotFoundError{})
				}

				// The new position's unclaimed rewards should be the sum of the old positions' unclaimed rewards.
				s.Require().Equal(unclaimedRewardsForEachUptimeAcrossAllOldPositions, unclaimedRewardsForEachUptimeNewPosition)

				// Get the final amount expected to be claimed by merging the unclaimed rewards across
				// all uptimes.
				expectedRewardToClaimDecCoins := sdk.NewDecCoins()
				for _, uptimeCoins := range unclaimedRewardsForEachUptimeAcrossAllOldPositions {
					expectedRewardToClaimDecCoins = expectedRewardToClaimDecCoins.Add(uptimeCoins...)
				}

				expectedRewardsToClaim, _ := expectedRewardToClaimDecCoins.TruncateDecimal()

				// Claim all the rewards for the new position and check that the rewards match the unclaimed rewards.
				claimedRewards, forfeitedRewards, err := s.App.ConcentratedLiquidityKeeper.ClaimAllIncentivesForPosition(s.Ctx, newPositionId)
				s.Require().NoError(err)

				s.Require().Equal(expectedRewardsToClaim, claimedRewards)
				s.Require().Equal(sdk.Coins(nil), forfeitedRewards)

				// Sanity check that cannot claim again.
				claimedRewards, _, err = s.App.ConcentratedLiquidityKeeper.ClaimAllIncentivesForPosition(s.Ctx, newPositionId)
				s.Require().NoError(err)

				s.Require().Equal(sdk.Coins(nil), claimedRewards)
				s.Require().Equal(sdk.Coins(nil), claimedRewards)

				// Check that cannot claim rewards for the old positions.
				for _, positionId := range test.positionIdsToMigrate {
					_, _, err := s.App.ConcentratedLiquidityKeeper.ClaimAllIncentivesForPosition(s.Ctx, positionId)
					s.Require().Error(err)
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestHasAnyPositionForPool() {
	s.SetupTest()
	defaultAddress := s.TestAccs[0]
	DefaultJoinTime := s.Ctx.BlockTime()

	tests := []struct {
		name           string
		poolId         uint64
		setupPositions []model.Position
		expectedResult bool
	}{
		{
			name:           "no positions exist",
			poolId:         defaultPoolId,
			expectedResult: false,

			setupPositions: []model.Position{},
		},
		{
			name:           "one position",
			poolId:         defaultPoolId,
			expectedResult: true,

			setupPositions: []model.Position{
				{
					PoolId:    defaultPoolId,
					Address:   defaultAddress.String(),
					LowerTick: DefaultLowerTick,
					UpperTick: DefaultUpperTick,
				},
			},
		},
		{
			name:           "two positions per pool",
			poolId:         defaultPoolId,
			expectedResult: true,

			setupPositions: []model.Position{
				{
					PoolId:    defaultPoolId,
					Address:   defaultAddress.String(),
					LowerTick: DefaultLowerTick,
					UpperTick: DefaultUpperTick,
				},
				{
					PoolId:    defaultPoolId,
					Address:   defaultAddress.String(),
					LowerTick: DefaultLowerTick,
					UpperTick: DefaultUpperTick,
				},
			},
		},
		{
			name:           "two positions for a different pool; returns false",
			poolId:         defaultPoolId + 1,
			expectedResult: false,

			setupPositions: []model.Position{
				{
					PoolId:    defaultPoolId,
					Address:   defaultAddress.String(),
					LowerTick: DefaultLowerTick,
					UpperTick: DefaultUpperTick,
				},
				{
					PoolId:    defaultPoolId,
					Address:   defaultAddress.String(),
					LowerTick: DefaultLowerTick,
					UpperTick: DefaultUpperTick,
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			// Create a default CL pools
			s.PrepareConcentratedPool()

			for _, pos := range test.setupPositions {
				s.SetupPosition(pos.PoolId, sdk.AccAddress(pos.Address), DefaultCoins, pos.LowerTick, pos.UpperTick, DefaultJoinTime)
			}

			// System under test
			actualResult, err := s.App.ConcentratedLiquidityKeeper.HasAnyPositionForPool(s.Ctx, test.poolId)

			s.Require().NoError(err)
			s.Require().Equal(test.expectedResult, actualResult)
		})
	}
}

// This test specifically tests that fee collection works as expected
// after fungifying positions.
func (s *KeeperTestSuite) TestFungifyChargedPositions_SwapAndClaimFees() {
	// Init suite for the test.
	s.SetupTest()

	const (
		numPositions           = 3
		testFullChargeDuration = time.Hour * 24
		swapAmount             = 1_000_000
	)

	var (
		defaultAddress   = s.TestAccs[0]
		defaultBlockTime = time.Unix(1, 1).UTC()
		swapFee          = sdk.NewDecWithPrec(2, 3)
	)

	expectedPositionIds := make([]uint64, numPositions)
	for i := 0; i < numPositions; i++ {
		expectedPositionIds[i] = uint64(i + 1)
	}

	s.TestAccs = apptesting.CreateRandomAccounts(5)
	s.Ctx = s.Ctx.WithBlockTime(defaultBlockTime)
	totalPositionsToCreate := sdk.NewInt(int64(numPositions))
	requiredBalances := sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0.Mul(totalPositionsToCreate)), sdk.NewCoin(USDC, DefaultAmt1.Mul(totalPositionsToCreate)))

	// Set test authorized uptime params.
	params := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
	params.AuthorizedUptimes = []time.Duration{time.Nanosecond, testFullChargeDuration}
	s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, params)

	// Fund account
	s.FundAcc(defaultAddress, requiredBalances)

	// Create CL pool
	s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, swapFee)

	// Set incentives for pool to ensure accumulators work correctly
	err := s.App.ConcentratedLiquidityKeeper.SetMultipleIncentiveRecords(s.Ctx, DefaultIncentiveRecords)
	s.Require().NoError(err)

	// Set up fully charged positions
	totalLiquidity := sdk.ZeroDec()
	for i := 0; i < numPositions; i++ {
		_, _, _, liquidityCreated, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, defaultPoolId, defaultAddress, DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
		s.Require().NoError(err)
		totalLiquidity = totalLiquidity.Add(liquidityCreated)
	}

	// Perform a swap to earn fees
	swapAmountIn := sdk.NewCoin(ETH, sdk.NewInt(swapAmount))
	expectedFee := swapAmountIn.Amount.ToDec().Mul(swapFee)
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(swapAmountIn))
	s.swapAndTrackXTimesInARow(defaultPoolId, swapAmountIn, USDC, types.MinSpotPrice, 1)

	// Increase block time by the fully charged duration
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(testFullChargeDuration))

	// First run non mutative validation and check results
	newPositionId, err := s.App.ConcentratedLiquidityKeeper.FungifyChargedPosition(s.Ctx, defaultAddress, expectedPositionIds)
	s.Require().NoError(err)

	// Claim fees
	collected, err := s.App.ConcentratedLiquidityKeeper.CollectFees(s.Ctx, defaultAddress, newPositionId)
	s.Require().NoError(err)

	// Validate that the correct fee amount was collected.
	s.Require().Equal(expectedFee, collected.AmountOf(swapAmountIn.Denom).ToDec())

	// Check that cannot claim again.
	collected, err = s.App.ConcentratedLiquidityKeeper.CollectFees(s.Ctx, defaultAddress, newPositionId)
	s.Require().NoError(err)
	s.Require().Equal(sdk.Coins(nil), collected)

	feeAccum, err := s.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(s.Ctx, defaultPoolId)
	s.Require().NoError(err)

	// Check that cannot claim old positions
	for _, oldPositionId := range expectedPositionIds {
		collected, err = s.App.ConcentratedLiquidityKeeper.CollectFees(s.Ctx, defaultAddress, oldPositionId)
		s.Require().Error(err)
		s.Require().Equal(sdk.Coins{}, collected)

		hasPosition := s.App.ConcentratedLiquidityKeeper.HasPosition(s.Ctx, oldPositionId)
		s.Require().False(hasPosition)

		hasFeePositionTracker, err := feeAccum.HasPosition(types.KeyFeePositionAccumulator(oldPositionId))
		s.Require().NoError(err)
		s.Require().False(hasFeePositionTracker)
	}
}

func (s *KeeperTestSuite) TestFungifyChargedPositions_ClaimIncentives() {
	// Init suite for the test.
	s.SetupTest()

	const (
		numPositions           = 3
		testFullChargeDuration = 24 * time.Hour
	)

	var (
		defaultAddress   = s.TestAccs[0]
		defaultBlockTime = time.Unix(1, 1).UTC()
		swapFee          = sdk.NewDecWithPrec(2, 3)
	)

	expectedPositionIds := make([]uint64, numPositions)
	for i := 0; i < numPositions; i++ {
		expectedPositionIds[i] = uint64(i + 1)
	}

	s.TestAccs = apptesting.CreateRandomAccounts(5)
	s.Ctx = s.Ctx.WithBlockTime(defaultBlockTime)
	totalPositionsToCreate := sdk.NewInt(int64(numPositions))
	requiredBalances := sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0.Mul(totalPositionsToCreate)), sdk.NewCoin(USDC, DefaultAmt1.Mul(totalPositionsToCreate)))

	// Set test authorized uptime params.
	params := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
	params.AuthorizedUptimes = []time.Duration{time.Nanosecond, testFullChargeDuration}
	s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, params)

	// Fund accounts
	s.FundAcc(defaultAddress, requiredBalances)

	// Create CL pool
	pool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, swapFee)

	// an error of 1 for each position
	roundingError := int64(numPositions)
	roundingTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: sdk.NewDec(roundingError),
		RoundingDir:       osmomath.RoundDown,
	}
	expectedAmount := sdk.NewInt(60 * 60 * 24) // 1 day in seconds * 1 per second

	s.FundAcc(pool.GetIncentivesAddress(), sdk.NewCoins(sdk.NewCoin(USDC, expectedAmount)))
	// Set incentives for pool to ensure accumulators work correctly
	testIncentiveRecord := types.IncentiveRecord{
		PoolId:               1,
		IncentiveDenom:       USDC,
		IncentiveCreatorAddr: s.TestAccs[0].String(),
		IncentiveRecordBody: types.IncentiveRecordBody{
			RemainingAmount: sdk.NewDec(1000000000000000000),
			EmissionRate:    sdk.NewDec(1), // 1 per second
			StartTime:       defaultBlockTime,
		},
		MinUptime: time.Nanosecond,
	}
	err := s.App.ConcentratedLiquidityKeeper.SetMultipleIncentiveRecords(s.Ctx, []types.IncentiveRecord{testIncentiveRecord})
	s.Require().NoError(err)

	// Set up fully charged positions
	totalLiquidity := sdk.ZeroDec()
	for i := 0; i < numPositions; i++ {
		_, _, _, liquidityCreated, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, defaultPoolId, defaultAddress, DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
		s.Require().NoError(err)
		totalLiquidity = totalLiquidity.Add(liquidityCreated)
	}

	// Increase block time by the fully charged duration
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(testFullChargeDuration))

	// sync accumulators
	// We use cache context to update uptime accumulators for estimating claimable incentives
	// prior to running fungify. However, we do not want the mutations made in test setup to have
	// impact on the system under test because it (fungify) must update the uptime accumulators itself.
	cacheCtx, _ := s.Ctx.CacheContext()
	err = s.App.ConcentratedLiquidityKeeper.UpdateUptimeAccumulatorsToNow(cacheCtx, pool.GetId())
	s.Require().NoError(err)

	claimableIncentives := sdk.NewCoins()
	for i := 0; i < numPositions; i++ {
		positionIncentices, forfeitedIncentives, err := s.App.ConcentratedLiquidityKeeper.GetClaimableIncentives(cacheCtx, uint64(i+1))
		s.Require().NoError(err)
		s.Require().Equal(sdk.Coins(nil), forfeitedIncentives)
		claimableIncentives = claimableIncentives.Add(positionIncentices...)
	}

	actualClaimedAmount := claimableIncentives.AmountOf(USDC)
	s.Require().Equal(0, roundingTolerance.Compare(expectedAmount, actualClaimedAmount), "expected: %s, got: %s", expectedAmount, actualClaimedAmount)

	// System under test
	newPositionId, err := s.App.ConcentratedLiquidityKeeper.FungifyChargedPosition(s.Ctx, defaultAddress, expectedPositionIds)
	s.Require().NoError(err)

	// Claim incentives.
	collected, _, err := s.App.ConcentratedLiquidityKeeper.CollectIncentives(s.Ctx, defaultAddress, newPositionId)
	s.Require().NoError(err)

	// Validate that the correct incentives amount was collected.
	actualClaimedAmount = collected.AmountOf(USDC)
	s.Require().Equal(1, len(collected))
	s.Require().Equal(0, roundingTolerance.Compare(expectedAmount, actualClaimedAmount), "expected: %s, got: %s", expectedAmount, actualClaimedAmount)

	// Check that cannot claim again.
	collected, _, err = s.App.ConcentratedLiquidityKeeper.CollectIncentives(s.Ctx, defaultAddress, newPositionId)
	s.Require().NoError(err)
	s.Require().Equal(sdk.Coins(nil), collected)

	// Check that cannot claim old positions
	for i := 0; i < numPositions; i++ {
		collected, _, err = s.App.ConcentratedLiquidityKeeper.CollectIncentives(s.Ctx, defaultAddress, uint64(i+1))
		s.Require().Error(err)
		s.Require().Equal(sdk.Coins{}, collected)
	}
}

func (s *KeeperTestSuite) TestCreateFullRangePosition() {
	var (
		positionId         uint64
		liquidity          sdk.Dec
		concentratedLockId uint64
		err                error
	)
	invalidCoinsAmount := sdk.NewCoins(DefaultCoin0)
	invalidCoin0Denom := sdk.NewCoins(sdk.NewCoin("invalidDenom", sdk.NewInt(1000000000000000000)), DefaultCoin1)
	invalidCoin1Denom := sdk.NewCoins(DefaultCoin0, sdk.NewCoin("invalidDenom", sdk.NewInt(1000000000000000000)))

	tests := []struct {
		name                  string
		remainingLockDuration time.Duration
		coinsForPosition      sdk.Coins
		isLocked              bool
		isUnlocking           bool
		expectedErr           error
	}{
		{
			name:             "full range position",
			coinsForPosition: DefaultCoins,
		},
		{
			name:                  "full range position: locked",
			remainingLockDuration: 24 * time.Hour * 14,
			coinsForPosition:      DefaultCoins,
			isLocked:              true,
		},
		{
			name:                  "full range position: unlocking",
			remainingLockDuration: 24 * time.Hour,
			coinsForPosition:      DefaultCoins,
			isUnlocking:           true,
		},
		{
			name:             "err: only one asset provided for a full range",
			coinsForPosition: invalidCoinsAmount,
			expectedErr:      types.NumCoinsError{NumCoins: 1},
		},
		{
			name:                  "err: only one asset provided for a full range locked",
			remainingLockDuration: 24 * time.Hour * 14,
			coinsForPosition:      invalidCoinsAmount,
			isLocked:              true,
			expectedErr:           types.NumCoinsError{NumCoins: 1},
		},
		{
			name:                  "err: only one asset provided for a full range unlocking",
			remainingLockDuration: 24 * time.Hour,
			coinsForPosition:      invalidCoinsAmount,
			isUnlocking:           true,
			expectedErr:           types.NumCoinsError{NumCoins: 1},
		},
		{
			name:             "err: wrong denom 0 provided for a full range",
			coinsForPosition: invalidCoin0Denom,
			expectedErr:      types.Amount0IsNegativeError{Amount0: sdk.ZeroInt()},
		},
		{
			name:             "err: wrong denom 1 provided for a full range",
			coinsForPosition: invalidCoin1Denom,
			expectedErr:      types.Amount1IsNegativeError{Amount1: sdk.ZeroInt()},
		},
	}

	for _, test := range tests {
		test := test
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()
			DefaultJoinTime := s.Ctx.BlockTime()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			// Create a default CL pool
			// We prepare it with a position already registered to prevent any oddities that we are not testing for here.
			clPool := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(ETH, USDC)

			// Fund the owner account
			defaultAddress := s.TestAccs[0]
			s.FundAcc(defaultAddress, test.coinsForPosition)

			// System under test
			if test.isLocked {
				positionId, _, _, liquidity, _, concentratedLockId, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, clPool.GetId(), defaultAddress, test.coinsForPosition, test.remainingLockDuration)
			} else if test.isUnlocking {
				positionId, _, _, liquidity, _, concentratedLockId, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionUnlocking(s.Ctx, clPool.GetId(), defaultAddress, test.coinsForPosition, test.remainingLockDuration)
			} else {
				positionId, _, _, liquidity, _, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool.GetId(), defaultAddress, test.coinsForPosition)
			}

			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, test.expectedErr.Error())
				return
			}

			s.Require().NoError(err)

			// Check position
			_, err = s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionId)
			s.Require().NoError(err)

			// Check lock
			if test.isLocked || test.isUnlocking {
				concentratedLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, concentratedLockId)
				s.Require().NoError(err)
				s.Require().Equal(liquidity.TruncateInt().String(), concentratedLock.Coins[0].Amount.String())
				isUnlocking := concentratedLock.IsUnlocking()
				s.Require().Equal(!test.isLocked, isUnlocking)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMintSharesLockAndUpdate() {
	defaultAddress := s.TestAccs[0]
	defaultPositionCoins := sdk.NewCoins(DefaultCoin0, DefaultCoin1)

	tests := []struct {
		name                    string
		owner                   sdk.AccAddress
		remainingLockDuration   time.Duration
		createFullRangePosition bool
		expectedErr             error
	}{
		{
			name:                    "2 week lock",
			owner:                   defaultAddress,
			createFullRangePosition: true,
			remainingLockDuration:   24 * time.Hour * 14,
		},
		{
			name:                    "1 day lock",
			owner:                   defaultAddress,
			createFullRangePosition: true,
			remainingLockDuration:   24 * time.Hour,
		},
		{
			name:                    "err: not a full range position",
			owner:                   defaultAddress,
			createFullRangePosition: false,
			remainingLockDuration:   24 * time.Hour,
			expectedErr:             types.PositionNotFullRangeError{PositionId: 1, LowerTick: DefaultLowerTick, UpperTick: DefaultUpperTick},
		},
	}

	for _, test := range tests {
		test := test
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			// Create a default CL pools
			clPool := s.PrepareConcentratedPool()

			// Fund the owner account
			s.FundAcc(test.owner, defaultPositionCoins)

			// Create a position
			positionId := uint64(0)
			liquidity := sdk.ZeroDec()
			if test.createFullRangePosition {
				var err error
				positionId, _, _, liquidity, _, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool.GetId(), test.owner, defaultPositionCoins)
				s.Require().NoError(err)
			} else {
				var err error
				positionId, _, _, liquidity, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, clPool.GetId(), test.owner, defaultPositionCoins, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
				s.Require().NoError(err)
			}

			lockupModuleAccountBalancePre := s.App.LockupKeeper.GetModuleBalance(s.Ctx)

			// System under test
			concentratedLockId, underlyingLiquidityTokenized, err := s.App.ConcentratedLiquidityKeeper.MintSharesLockAndUpdate(s.Ctx, clPool.GetId(), positionId, test.owner, test.remainingLockDuration)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, test.expectedErr)
				return
			}
			s.Require().NoError(err)

			// Check that the underlying liquidity tokenized is equal to the liquidity of the position
			s.Require().Equal(liquidity.TruncateInt().String(), underlyingLiquidityTokenized[0].Amount.String())

			lockupModuleAccountBalancePost := s.App.LockupKeeper.GetModuleBalance(s.Ctx)

			// Check that the lockup module account balance increased by the amount expected to be locked
			s.Require().Equal(underlyingLiquidityTokenized[0].String(), lockupModuleAccountBalancePost.Sub(lockupModuleAccountBalancePre).String())

			// Check that the positionId is mapped to the lockId
			positionLockId, err := s.App.ConcentratedLiquidityKeeper.GetLockIdFromPositionId(s.Ctx, positionId)
			s.Require().NoError(err)
			s.Require().Equal(positionLockId, concentratedLockId)

			// Check total supply of cl liquidity token increased by the amount expected to be minted
			clPositionSharesInSupply := s.App.BankKeeper.GetSupply(s.Ctx, underlyingLiquidityTokenized[0].Denom)
			s.Require().Equal(underlyingLiquidityTokenized[0].String(), clPositionSharesInSupply.String())

			// Check specific lock params
			concentratedLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, concentratedLockId)
			s.Require().NoError(err)
			s.Require().Equal(underlyingLiquidityTokenized[0].Amount.String(), concentratedLock.Coins[0].Amount.String())
			s.Require().Equal(test.remainingLockDuration, concentratedLock.Duration)
		},
		)
	}
}

func (s *KeeperTestSuite) TestPositionHasActiveUnderlyingLock() {
	defaultLockDuration := 24 * time.Hour
	clPool := s.PrepareConcentratedPool()
	owner := s.TestAccs[0]
	defaultPositionCoins := sdk.NewCoins(DefaultCoin0, DefaultCoin1)

	type testParams struct {
		name                                 string
		createPosition                       func(s *KeeperTestSuite) (uint64, uint64)
		expectedHasActiveLock                bool
		expectedHasActiveLockAfterTimeUpdate bool
		expectedLockError                    bool
		expectedPositionLockID               uint64
		expectedGetPositionLockIdErr         bool
	}

	tests := []testParams{
		{
			name: "position with lock locked",
			createPosition: func(s *KeeperTestSuite) (uint64, uint64) {
				s.FundAcc(owner, defaultPositionCoins)
				positionID, _, _, _, _, concentratedLockID, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(
					s.Ctx, clPool.GetId(), owner, defaultPositionCoins, defaultLockDuration)
				s.Require().NoError(err)
				return positionID, concentratedLockID
			},
			expectedHasActiveLock:                true, // lock starts as active
			expectedHasActiveLockAfterTimeUpdate: true, // since lock is locked, it remains active after time update
			expectedLockError:                    false,
			expectedPositionLockID:               1,
			expectedGetPositionLockIdErr:         false,
		},
		{
			name: "position with lock unlocking",
			createPosition: func(s *KeeperTestSuite) (uint64, uint64) {
				s.FundAcc(owner, defaultPositionCoins)
				positionID, _, _, _, _, concentratedLockID, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionUnlocking(
					s.Ctx, clPool.GetId(), owner, defaultPositionCoins, defaultLockDuration)
				s.Require().NoError(err)
				return positionID, concentratedLockID
			},
			expectedHasActiveLock:                true,  // lock starts as active
			expectedHasActiveLockAfterTimeUpdate: false, // since lock is unlocking, it should no longer be active after time update
			expectedLockError:                    false,
			expectedPositionLockID:               2,
			expectedGetPositionLockIdErr:         false,
		},
		{
			name: "position without lock",
			createPosition: func(s *KeeperTestSuite) (uint64, uint64) {
				s.FundAcc(owner, defaultPositionCoins)
				positionID, _, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(
					s.Ctx, clPool.GetId(), owner, defaultPositionCoins)
				s.Require().NoError(err)
				return positionID, 0
			},
			expectedHasActiveLock:                false,
			expectedHasActiveLockAfterTimeUpdate: false,
			expectedLockError:                    true,
			expectedPositionLockID:               0,
			expectedGetPositionLockIdErr:         true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			positionID, lockID := tc.createPosition(s)

			retrievedLockID, err := s.App.ConcentratedLiquidityKeeper.GetLockIdFromPositionId(s.Ctx, positionID)
			if tc.expectedLockError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(lockID, retrievedLockID)
			}

			// System under test (non mutative)
			hasActiveLockInState, retrievedLockID, err := s.App.ConcentratedLiquidityKeeper.PositionHasActiveUnderlyingLock(s.Ctx, positionID)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedHasActiveLock, hasActiveLockInState)
			s.Require().Equal(tc.expectedPositionLockID, retrievedLockID)

			// Position ID to lock ID mapping should not change
			retrievedPositionID, err := s.App.ConcentratedLiquidityKeeper.GetPositionIdToLockId(s.Ctx, lockID)
			if tc.expectedGetPositionLockIdErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(positionID, retrievedPositionID)
			}

			// Move time forward by the lock duration
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(defaultLockDuration + 1))

			// System under test (non mutative)
			hasActiveLockInState, retrievedLockID, err = s.App.ConcentratedLiquidityKeeper.PositionHasActiveUnderlyingLock(s.Ctx, positionID)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedHasActiveLockAfterTimeUpdate, hasActiveLockInState)
			s.Require().Equal(tc.expectedPositionLockID, retrievedLockID)

			// Position ID to lock ID mapping should not change, even though underlying lock might be mature
			retrievedPositionID, err = s.App.ConcentratedLiquidityKeeper.GetPositionIdToLockId(s.Ctx, lockID)
			if tc.expectedGetPositionLockIdErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(positionID, retrievedPositionID)
			}
		})
	}
}

func (s *KeeperTestSuite) TestPositionHasActiveUnderlyingLockAndUpdate() {
	defaultLockDuration := 24 * time.Hour
	clPool := s.PrepareConcentratedPool()
	owner := s.TestAccs[0]
	defaultPositionCoins := sdk.NewCoins(DefaultCoin0, DefaultCoin1)

	type testParams struct {
		name                                        string
		createPosition                              func(s *KeeperTestSuite) (uint64, uint64)
		expectedHasActiveLock                       bool
		expectedHasActiveLockAfterTimeUpdate        bool
		expectedLockError                           bool
		expectedPositionLockID                      uint64
		expectedPositionLockIDAfterTimeUpdate       uint64
		expectedGetPositionLockIdErr                bool
		expectedGetPositionLockIdErrAfterTimeUpdate bool
	}

	tests := []testParams{
		{
			name: "position with lock locked",
			createPosition: func(s *KeeperTestSuite) (uint64, uint64) {
				s.FundAcc(owner, defaultPositionCoins)
				positionID, _, _, _, _, concentratedLockID, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(
					s.Ctx, clPool.GetId(), owner, defaultPositionCoins, defaultLockDuration)
				s.Require().NoError(err)
				return positionID, concentratedLockID
			},
			expectedHasActiveLock:                       true, // lock starts as active
			expectedHasActiveLockAfterTimeUpdate:        true, // since lock is locked, it remains active after time update
			expectedLockError:                           false,
			expectedPositionLockID:                      1,
			expectedPositionLockIDAfterTimeUpdate:       1, // since it stays locked, the mutative method wont change the underlying lock ID
			expectedGetPositionLockIdErr:                false,
			expectedGetPositionLockIdErrAfterTimeUpdate: false,
		},
		{
			name: "position with lock unlocking",
			createPosition: func(s *KeeperTestSuite) (uint64, uint64) {
				s.FundAcc(owner, defaultPositionCoins)
				positionID, _, _, _, _, concentratedLockID, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionUnlocking(
					s.Ctx, clPool.GetId(), owner, defaultPositionCoins, defaultLockDuration)
				s.Require().NoError(err)
				return positionID, concentratedLockID
			},
			expectedHasActiveLock:                       true,  // lock starts as active
			expectedHasActiveLockAfterTimeUpdate:        false, // since lock is unlocking, it should no longer be active after time update
			expectedLockError:                           false,
			expectedPositionLockID:                      2,
			expectedPositionLockIDAfterTimeUpdate:       0, // since it becomes unlocked, the mutative method will change the underlying lock ID to 0
			expectedGetPositionLockIdErr:                false,
			expectedGetPositionLockIdErrAfterTimeUpdate: true, // since it becomes unlocked, the mutative method will change the underlying lock ID to 0 and this now errors
		},
		{
			name: "position without lock",
			createPosition: func(s *KeeperTestSuite) (uint64, uint64) {
				s.FundAcc(owner, defaultPositionCoins)
				positionID, _, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(
					s.Ctx, clPool.GetId(), owner, defaultPositionCoins)
				s.Require().NoError(err)
				return positionID, 0
			},
			expectedHasActiveLock:                       false,
			expectedHasActiveLockAfterTimeUpdate:        false,
			expectedLockError:                           true,
			expectedPositionLockID:                      0,
			expectedPositionLockIDAfterTimeUpdate:       0,
			expectedGetPositionLockIdErr:                true,
			expectedGetPositionLockIdErrAfterTimeUpdate: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			positionID, lockID := tc.createPosition(s)

			retrievedLockID, err := s.App.ConcentratedLiquidityKeeper.GetLockIdFromPositionId(s.Ctx, positionID)
			if tc.expectedLockError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(lockID, retrievedLockID)
			}

			// System under test (mutative)
			hasActiveLockInState, retrievedLockID, err := s.App.ConcentratedLiquidityKeeper.PositionHasActiveUnderlyingLockAndUpdate(s.Ctx, positionID)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedHasActiveLock, hasActiveLockInState)
			s.Require().Equal(tc.expectedPositionLockID, retrievedLockID)

			// Position ID to lock ID mapping should not change
			retrievedPositionID, err := s.App.ConcentratedLiquidityKeeper.GetPositionIdToLockId(s.Ctx, lockID)
			if tc.expectedGetPositionLockIdErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(positionID, retrievedPositionID)
			}

			// Move time forward by the lock duration
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(defaultLockDuration + 1))

			// System under test (mutative)
			hasActiveLockInState, retrievedLockID, err = s.App.ConcentratedLiquidityKeeper.PositionHasActiveUnderlyingLockAndUpdate(s.Ctx, positionID)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedHasActiveLockAfterTimeUpdate, hasActiveLockInState)
			s.Require().Equal(tc.expectedPositionLockIDAfterTimeUpdate, retrievedLockID)

			// Position ID to lock ID mapping should not change, even though underlying lock might be mature
			retrievedPositionID, err = s.App.ConcentratedLiquidityKeeper.GetPositionIdToLockId(s.Ctx, lockID)
			s.Require().Equal(tc.expectedPositionLockIDAfterTimeUpdate, retrievedPositionID)
			if tc.expectedGetPositionLockIdErrAfterTimeUpdate {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// Tests Set, Get, and Remove methods for positionId -> lockId mappings
func (s *KeeperTestSuite) TestPositionToLockCRUD() {
	// Init suite for each test.
	s.SetupTest()
	s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)
	owner := s.TestAccs[0]
	remainingLockDuration := 24 * time.Hour
	defaultPositionCoins := sdk.NewCoins(DefaultCoin0, DefaultCoin1)

	// Create a default CL pools
	clPool := s.PrepareConcentratedPool()

	// Fund the owner account
	s.FundAcc(owner, defaultPositionCoins)

	// Create a position with a lock
	positionId, _, _, _, _, concentratedLockId, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionUnlocking(s.Ctx, clPool.GetId(), owner, defaultPositionCoins, remainingLockDuration)
	s.Require().NoError(err)

	// We should be able to retrieve the lockId from the positionId now
	retrievedLockId, err := s.App.ConcentratedLiquidityKeeper.GetLockIdFromPositionId(s.Ctx, positionId)
	s.Require().NoError(err)
	s.Require().Equal(concentratedLockId, retrievedLockId)

	// Check if lock has position in state
	retrievedPositionId, err := s.App.ConcentratedLiquidityKeeper.GetPositionIdToLockId(s.Ctx, retrievedLockId)
	s.Require().NoError(err)
	s.Require().Equal(positionId, retrievedPositionId)

	// Create a position without a lock
	positionId, _, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool.GetId(), owner, defaultPositionCoins)
	s.Require().Error(err)

	// Check if position has lock in state, should not
	retrievedLockId, err = s.App.ConcentratedLiquidityKeeper.GetLockIdFromPositionId(s.Ctx, positionId)
	s.Require().Error(err)
	s.Require().Equal(uint64(0), retrievedLockId)

	// Set the position to have a lockId (despite it not actually having a lock)
	s.App.ConcentratedLiquidityKeeper.SetPositionIdToLock(s.Ctx, positionId, concentratedLockId)

	// Check if position has lock in state, it should now
	retrievedLockId, err = s.App.ConcentratedLiquidityKeeper.GetLockIdFromPositionId(s.Ctx, positionId)
	s.Require().NoError(err)
	s.Require().Equal(concentratedLockId, retrievedLockId)

	// Check if lock has position in state
	retrievedPositionId, err = s.App.ConcentratedLiquidityKeeper.GetPositionIdToLockId(s.Ctx, retrievedLockId)
	s.Require().NoError(err)
	s.Require().Equal(positionId, retrievedPositionId)

	// Remove the lockId from the position
	s.App.ConcentratedLiquidityKeeper.RemovePositionIdToLock(s.Ctx, positionId, retrievedLockId)

	// Check if position has lock in state, should not
	retrievedLockId, err = s.App.ConcentratedLiquidityKeeper.GetLockIdFromPositionId(s.Ctx, positionId)
	s.Require().Error(err)
	s.Require().Equal(uint64(0), retrievedLockId)
}

func (s *KeeperTestSuite) TestSetPosition() {
	defaultAddress := s.TestAccs[0]
	DefaultJoinTime := s.Ctx.BlockTime()

	testCases := []struct {
		name             string
		poolId           uint64
		owner            sdk.AccAddress
		lowerTick        int64
		upperTick        int64
		joinTime         time.Time
		liquidity        sdk.Dec
		positionId       uint64
		underlyingLockId uint64
	}{
		{
			name:             "basic set position",
			poolId:           1,
			owner:            defaultAddress,
			lowerTick:        DefaultLowerTick,
			upperTick:        DefaultUpperTick,
			joinTime:         DefaultJoinTime,
			liquidity:        DefaultLiquidityAmt,
			positionId:       1,
			underlyingLockId: 0,
		},
		{
			name:             "set position with underlying lock",
			poolId:           1,
			owner:            defaultAddress,
			lowerTick:        DefaultLowerTick,
			upperTick:        DefaultUpperTick,
			joinTime:         DefaultJoinTime,
			liquidity:        DefaultLiquidityAmt,
			positionId:       2,
			underlyingLockId: 3,
		},
	}

	// Loop through test cases.
	for _, tc := range testCases {
		s.SetupTest()
		s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)
		store := s.Ctx.KVStore(s.App.GetKey(types.StoreKey))
		s.PrepareConcentratedPool()

		// Call the SetPosition function with test case parameters.
		err := s.App.ConcentratedLiquidityKeeper.SetPosition(
			s.Ctx,
			tc.poolId,
			tc.owner,
			tc.lowerTick,
			tc.upperTick,
			tc.joinTime,
			tc.liquidity,
			tc.positionId,
			tc.underlyingLockId,
		)
		s.Require().NoError(err)

		// Retrieve the position from the store via position ID and compare to expected values.
		position := model.Position{}
		key := types.KeyPositionId(tc.positionId)
		osmoutils.MustGet(store, key, &position)
		s.Require().Equal(tc.positionId, position.PositionId)
		s.Require().Equal(tc.poolId, position.PoolId)
		s.Require().Equal(tc.owner.String(), position.Address)
		s.Require().Equal(tc.lowerTick, position.LowerTick)
		s.Require().Equal(tc.upperTick, position.UpperTick)
		s.Require().Equal(tc.joinTime, position.JoinTime)
		s.Require().Equal(tc.liquidity, position.Liquidity)

		// Retrieve the position from the store via owner/poolId/positionId and compare to expected values.
		key = types.KeyAddressPoolIdPositionId(tc.owner, tc.poolId, tc.positionId)
		positionIdBytes := store.Get(key)
		s.Require().Equal(tc.positionId, sdk.BigEndianToUint64(positionIdBytes))

		// Retrieve the position from the store via poolId/positionId and compare to expected values.
		key = types.KeyPoolPositionPositionId(tc.poolId, tc.positionId)
		positionIdBytes = store.Get(key)
		s.Require().Equal(tc.positionId, sdk.BigEndianToUint64(positionIdBytes))

		// Retrieve the position ID to underlying lock ID mapping from the store and compare to expected values.
		key = types.KeyPositionIdForLock(tc.positionId)
		underlyingLockIdBytes := store.Get(key)
		if tc.underlyingLockId != 0 {
			s.Require().Equal(tc.underlyingLockId, sdk.BigEndianToUint64(underlyingLockIdBytes))
		} else {
			s.Require().Nil(underlyingLockIdBytes)
		}
	}
}

func (s *KeeperTestSuite) TestGetAndUpdateFullRangeLiquidity() {
	testCases := []struct {
		name                 string
		positionCoins        sdk.Coins
		lowerTick, upperTick int64
		updateLiquidity      sdk.Dec
	}{
		{
			name:            "full range + position overlapping min tick. update liquidity upwards",
			positionCoins:   sdk.NewCoins(DefaultCoin0, DefaultCoin1),
			lowerTick:       DefaultMinTick,
			upperTick:       DefaultUpperTick, // max tick doesn't overlap, should not count towards full range liquidity
			updateLiquidity: sdk.NewDec(100),
		},
		{
			name:            "full range + position overlapping max tick. update liquidity downwards",
			positionCoins:   sdk.NewCoins(DefaultCoin0, DefaultCoin1),
			lowerTick:       DefaultLowerTick, // min tick doesn't overlap, should not count towards full range liquidity
			upperTick:       DefaultMaxTick,
			updateLiquidity: sdk.NewDec(-100),
		},
	}

	for _, tc := range testCases {
		s.SetupTest()
		s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)
		owner := s.TestAccs[0]
		s.FundAcc(owner, tc.positionCoins)

		// Create a new pool.
		clPool := s.PrepareConcentratedPool()
		clPoolId := clPool.GetId()
		actualFullRangeLiquidity := sdk.ZeroDec()

		// Create a full range position.
		_, _, _, liquidity, _, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool.GetId(), owner, tc.positionCoins)
		s.Require().NoError(err)
		actualFullRangeLiquidity = actualFullRangeLiquidity.Add(liquidity)

		clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPoolId)
		s.Require().NoError(err)

		// Get the full range liquidity for the pool.
		expectedFullRangeLiquidity := s.App.ConcentratedLiquidityKeeper.MustGetFullRangeLiquidityInPool(s.Ctx, clPoolId)
		s.Require().Equal(expectedFullRangeLiquidity, actualFullRangeLiquidity)

		// Create a new position that overlaps with the min tick, but is not full range and therefore should not count towards the full range liquidity.
		s.FundAcc(owner, tc.positionCoins)
		_, _, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, clPoolId, owner, DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), tc.lowerTick, tc.upperTick)
		s.Require().NoError(err)

		clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPoolId)
		s.Require().NoError(err)

		// Test updating the full range liquidity.
		err = s.App.ConcentratedLiquidityKeeper.UpdateFullRangeLiquidityInPool(s.Ctx, clPoolId, tc.updateLiquidity)
		s.Require().NoError(err)
		actualFullRangeLiquidity = s.App.ConcentratedLiquidityKeeper.MustGetFullRangeLiquidityInPool(s.Ctx, clPoolId)
		s.Require().Equal(expectedFullRangeLiquidity.Add(tc.updateLiquidity), actualFullRangeLiquidity)
	}
}

func (s *KeeperTestSuite) TestGetAllPositionIdsForPoolId() {
	s.SetupTest()
	clKeeper := s.App.ConcentratedLiquidityKeeper
	s.Ctx = s.Ctx.WithBlockTime(defaultStartTime)

	// Set up test pool
	clPoolOne := s.PrepareConcentratedPool()

	s.SetupDefaultPositionAcc(clPoolOne.GetId(), s.TestAccs[0])
	s.SetupDefaultPositionAcc(clPoolOne.GetId(), s.TestAccs[1])
	s.SetupDefaultPositionAcc(clPoolOne.GetId(), s.TestAccs[2])

	clPooltwo := s.PrepareConcentratedPool()

	s.SetupDefaultPositionAcc(clPooltwo.GetId(), s.TestAccs[0])
	s.SetupDefaultPositionAcc(clPooltwo.GetId(), s.TestAccs[1])
	s.SetupDefaultPositionAcc(clPooltwo.GetId(), s.TestAccs[2])

	expectedPositionOneIds := []uint64{1, 2, 3}
	expectedPositionTwoIds := []uint64{4, 5, 6}

	positionOne, err := clKeeper.GetAllPositionIdsForPoolId(s.Ctx, clPoolOne.GetId())
	s.Require().NoError(err)

	positionTwo, err := clKeeper.GetAllPositionIdsForPoolId(s.Ctx, clPooltwo.GetId())
	s.Require().NoError(err)

	s.Require().Equal(expectedPositionOneIds, positionOne)
	s.Require().Equal(expectedPositionTwoIds, positionTwo)
}
