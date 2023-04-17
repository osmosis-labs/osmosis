package concentrated_liquidity_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

var (
	DefaultIncentiveRecords = []types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo, incentiveRecordThree, incentiveRecordFour}
)

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
		liquidityIn    sdk.Dec
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
			s.Setup()

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
			s.App.ConcentratedLiquidityKeeper.SetMultipleIncentiveRecords(s.Ctx, test.incentiveRecords)

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
			actualUptimeAccumDelta, expectedUptimeAccumValueGrowth, expectedIncentiveRecords, expectedGrowthCurAccum := emptyAccumValues, emptyAccumValues, test.incentiveRecords, sdk.DecCoins{}

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
				s.Require().Equal(expectedInitAccumValues[uptimeIndex], positionRecord.InitAccumValue)
				s.Require().Equal(test.expectedLiquidity, positionRecord.NumShares)

				// Accumulator value related checks

				if test.positionExists {
					// Track how much the current uptime accum has grown by
					actualUptimeAccumDelta[uptimeIndex] = newUptimeAccumValues[uptimeIndex].Sub(initUptimeAccumValues[uptimeIndex])
					if timeElapsedSec.GT(sdk.ZeroDec()) {
						expectedGrowthCurAccum, expectedIncentiveRecords, err = cl.CalcAccruedIncentivesForAccum(s.Ctx, uptime, test.param.liquidityDelta, timeElapsedSec, expectedIncentiveRecords)
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
		name             string
		positionId       uint64
		expectedPosition sdk.Dec
		expectedErr      error
	}{
		{
			name:             "Get position info on existing pool and existing position",
			positionId:       DefaultPositionId,
			expectedPosition: DefaultLiquidityAmt,
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
			s.Setup()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)
			// Create a default CL pool
			s.PrepareConcentratedPool()

			// Set up a default initialized position
			err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePosition(s.Ctx, validPoolId, s.TestAccs[0], DefaultLowerTick, DefaultUpperTick, DefaultLiquidityAmt, DefaultJoinTime, DefaultPositionId)

			// System under test
			positionLiquidity, err := s.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(s.Ctx, test.positionId)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, test.expectedErr)
				s.Require().Equal(sdk.Dec{}, positionLiquidity)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(test.expectedPosition, positionLiquidity)
			}
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
		coin0      sdk.Coin
		coin1      sdk.Coin
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
				{1, 1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, DefaultJoinTime},
			},
		},
		{
			name:   "Get current users multiple position same pool",
			sender: defaultAddress,
			setupPositions: []position{
				{1, 1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, DefaultJoinTime},
				{2, 1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick + 1, DefaultUpperTick + 1, DefaultJoinTime},
				{3, 1, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick + 2, DefaultUpperTick + 2, DefaultJoinTime},
			},
		},
		{
			name:   "Get current users multiple position multiple pools",
			sender: secondAddress,
			setupPositions: []position{
				{1, 1, secondAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, DefaultJoinTime},
				{2, 2, secondAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick + 1, DefaultUpperTick + 1, DefaultJoinTime},
				{3, 3, secondAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick + 2, DefaultUpperTick + 2, DefaultJoinTime},
			},
		},
		{
			name:   "User has positions over multiple pools, but filter by one pool",
			sender: secondAddress,
			poolId: 2,
			setupPositions: []position{
				{1, 1, secondAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick, DefaultJoinTime},
				{2, 2, secondAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick + 1, DefaultUpperTick + 1, DefaultJoinTime},
				{3, 3, secondAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick + 2, DefaultUpperTick + 2, DefaultJoinTime},
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			// Create a default CL pools
			s.PrepareMultipleConcentratedPools(3)

			expectedUserPositions := []model.Position{}
			for _, pos := range test.setupPositions {
				// if position does not exist this errors
				liquidity, _ := s.SetupPosition(pos.poolId, pos.acc, pos.coin0, pos.coin1, pos.lowerTick, pos.upperTick, pos.joinTime)
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
			s.Setup()
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
			key := types.KeyPositionId(DefaultPositionId)
			osmoutils.MustGet(store, key, &position)
			s.Require().Equal(DefaultPositionId, position.PositionId)
			s.Require().Equal(defaultPoolId, position.PoolId)
			s.Require().Equal(s.TestAccs[0].String(), position.Address)
			s.Require().Equal(DefaultLowerTick, position.LowerTick)
			s.Require().Equal(DefaultUpperTick, position.UpperTick)
			s.Require().Equal(DefaultJoinTime, position.JoinTime)
			s.Require().Equal(DefaultLiquidityAmt, position.Liquidity)

			// Retrieve the position from the store via owner/poolId/positionId and compare to expected values.
			key = types.KeyAddressPoolIdPositionId(s.TestAccs[0], defaultPoolId, DefaultPositionId)
			positionIdBytes := store.Get(key)
			s.Require().Equal(DefaultPositionId, sdk.BigEndianToUint64(positionIdBytes))

			// Retrieve the position from the store via poolId/positionId and compare to expected values.
			key = types.KeyPoolPositionPositionId(defaultPoolId, DefaultPositionId)
			positionIdBytes = store.Get(key)
			s.Require().Equal(DefaultPositionId, sdk.BigEndianToUint64(positionIdBytes))

			// Retrieve the position ID to underlying lock ID mapping from the store and compare to expected values.
			key = types.KeyPositionIdForLock(DefaultPositionId)
			underlyingLockIdBytes := store.Get(key)
			if test.underlyingLockId != 0 {
				s.Require().Equal(test.underlyingLockId, sdk.BigEndianToUint64(underlyingLockIdBytes))
			} else {
				s.Require().Nil(underlyingLockIdBytes)
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
				key := types.KeyPositionId(DefaultPositionId)
				osmoutils.Get(store, key, &position)
				s.Require().Equal(model.Position{}, position)

				// Retrieve the position from the store via owner/poolId/positionId and compare to expected values.
				key = types.KeyAddressPoolIdPositionId(s.TestAccs[0], defaultPoolId, DefaultPositionId)
				positionIdBytes := store.Get(key)
				s.Require().Nil(positionIdBytes)

				// Retrieve the position from the store via poolId/positionId and compare to expected values.
				key = types.KeyPoolPositionPositionId(defaultPoolId, DefaultPositionId)
				positionIdBytes = store.Get(key)
				s.Require().Nil(positionIdBytes)

				// Retrieve the position ID to underlying lock ID mapping from the store and compare to expected values.
				key = types.KeyPositionIdForLock(DefaultPositionId)
				underlyingLockIdBytes := store.Get(key)
				s.Require().Nil(underlyingLockIdBytes)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCalculateUnderlyingAssetsFromPosition() {
	tests := []struct {
		name           string
		position       model.Position
		expectedAsset0 sdk.Dec
		expectedAsset1 sdk.Dec
	}{
		{
			name:     "Default range position",
			position: model.Position{PoolId: 1, LowerTick: DefaultLowerTick, UpperTick: DefaultUpperTick},
		},
		{
			name:     "Full range position",
			position: model.Position{PoolId: 1, LowerTick: DefaultMinTick, UpperTick: DefaultMaxTick},
		},
		{
			name:     "Below current tick position",
			position: model.Position{PoolId: 1, LowerTick: DefaultLowerTick, UpperTick: DefaultLowerTick + 1},
		},
		{
			name:     "Above current tick position",
			position: model.Position{PoolId: 1, LowerTick: DefaultUpperTick, UpperTick: DefaultUpperTick + 1},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// prepare concentrated pool with a default position
			clPool := s.PrepareConcentratedPool()
			s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0), sdk.NewCoin(USDC, DefaultAmt1)))
			s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, 1, s.TestAccs[0], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)

			// create a position from the test case
			s.FundAcc(s.TestAccs[1], sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0), sdk.NewCoin(USDC, DefaultAmt1)))
			_, actualAmount0, actualAmount1, liquidity, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, tc.position.PoolId, s.TestAccs[1], DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), tc.position.LowerTick, tc.position.UpperTick)
			s.Require().NoError(err)
			tc.position.Liquidity = liquidity

			// calculate underlying assets from the position
			clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, tc.position.PoolId)
			s.Require().NoError(err)
			calculatedCoin0, calculatedCoin1, err := cl.CalculateUnderlyingAssetsFromPosition(s.Ctx, tc.position, clPool)

			s.Require().NoError(err)
			s.Require().Equal(calculatedCoin0.String(), sdk.NewCoin(clPool.GetToken0(), actualAmount0).String())
			s.Require().Equal(calculatedCoin1.String(), sdk.NewCoin(clPool.GetToken1(), actualAmount1).String())
		})
	}
}

func (s *KeeperTestSuite) TestValidateAndFungifyChargedPositions() {
	defaultAddress := s.TestAccs[0]
	secondAddress := s.TestAccs[1]
	defaultBlockTime = time.Unix(1, 1).UTC()

	type position struct {
		positionId uint64
		poolId     uint64
		acc        sdk.AccAddress
		coin0      sdk.Coin
		coin1      sdk.Coin
		lowerTick  int64
		upperTick  int64
	}

	tests := []struct {
		name                       string
		setupFullyChargedPositions []position
		setupUnchargedPositions    []position
		positionIdsToMigrate       []uint64
		accountCallingMigration    sdk.AccAddress
		expectedNewPositionId      uint64
		expectedErr                error
	}{
		{
			name: "Happy path: Fungify three fully charged positions",
			setupFullyChargedPositions: []position{
				{1, defaultPoolId, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick},
				{2, defaultPoolId, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick},
				{3, defaultPoolId, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick},
			},
			positionIdsToMigrate:    []uint64{1, 2, 3},
			accountCallingMigration: defaultAddress,
			expectedNewPositionId:   4,
		},
		{
			name: "Error: Fungify three positions, but one of them is not fully charged",
			setupFullyChargedPositions: []position{
				{1, defaultPoolId, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick},
				{2, defaultPoolId, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick},
			},
			setupUnchargedPositions: []position{
				{3, defaultPoolId, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick},
			},
			positionIdsToMigrate:    []uint64{1, 2, 3},
			accountCallingMigration: defaultAddress,
			expectedNewPositionId:   0,
			expectedErr:             types.PositionNotFullyChargedError{PositionId: 3, PositionJoinTime: defaultBlockTime.Add(cl.FullyChargedDuration), FullyChargedMinTimestamp: defaultBlockTime.Add(cl.FullyChargedDuration + time.Hour*24*7)},
		},
		{
			name: "Error: Fungify three positions, but one of them is not in the same pool",
			setupFullyChargedPositions: []position{
				{1, defaultPoolId, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick},
				{2, 2, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick},
				{3, defaultPoolId, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick},
			},
			positionIdsToMigrate:    []uint64{1, 2, 3},
			accountCallingMigration: defaultAddress,
			expectedNewPositionId:   0,
			expectedErr:             types.PositionsNotInSamePoolError{Position1PoolId: 2, Position2PoolId: 1},
		},
		{
			name: "Error: Fungify three positions, but one of them is not owned by the same owner",
			setupFullyChargedPositions: []position{
				{1, defaultPoolId, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick},
				{2, defaultPoolId, secondAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick},
				{3, defaultPoolId, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick},
			},
			positionIdsToMigrate:    []uint64{1, 2, 3},
			accountCallingMigration: defaultAddress,
			expectedNewPositionId:   0,
			expectedErr:             types.PositionOwnerMismatchError{PositionOwner: secondAddress.String(), Sender: defaultAddress.String()},
		},
		{
			name: "Error: Fungify three positions, but one of them is not in the same range",
			setupFullyChargedPositions: []position{
				{1, defaultPoolId, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick},
				{2, defaultPoolId, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick - 1, DefaultUpperTick},
				{3, defaultPoolId, defaultAddress, DefaultCoin0, DefaultCoin1, DefaultLowerTick, DefaultUpperTick},
			},
			positionIdsToMigrate:    []uint64{1, 2, 3},
			accountCallingMigration: defaultAddress,
			expectedNewPositionId:   0,
			expectedErr:             types.PositionsNotInSameTickRangeError{Position1TickLower: DefaultLowerTick - 1, Position1TickUpper: DefaultUpperTick, Position2TickLower: DefaultLowerTick, Position2TickUpper: DefaultUpperTick},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()
			s.Ctx = s.Ctx.WithBlockTime(defaultBlockTime)
			totalPositionsToCreate := sdk.NewInt(int64(len(test.setupFullyChargedPositions) + len(test.setupUnchargedPositions)))
			requiredBalances := sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0.Mul(totalPositionsToCreate)), sdk.NewCoin(USDC, DefaultAmt1.Mul(totalPositionsToCreate)))

			// Fund accounts
			s.FundAcc(defaultAddress, requiredBalances)
			s.FundAcc(secondAddress, requiredBalances)

			// Create two default CL pools
			s.PrepareConcentratedPool()
			s.PrepareConcentratedPool()

			// Set incentives for pool to ensure accumulators work correctly
			s.App.ConcentratedLiquidityKeeper.SetMultipleIncentiveRecords(s.Ctx, DefaultIncentiveRecords)

			// Set up fully charged positions
			totalLiquidity := sdk.ZeroDec()
			for _, pos := range test.setupFullyChargedPositions {
				_, _, _, liquidityCreated, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, pos.poolId, pos.acc, pos.coin0.Amount, pos.coin1.Amount, sdk.ZeroInt(), sdk.ZeroInt(), pos.lowerTick, pos.upperTick)
				s.Require().NoError(err)
				totalLiquidity = totalLiquidity.Add(liquidityCreated)
			}

			// Increase block time by the fully charged duration
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(cl.FullyChargedDuration))

			// Set up uncharged positions
			for _, pos := range test.setupUnchargedPositions {
				_, _, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, pos.poolId, pos.acc, pos.coin0.Amount, pos.coin1.Amount, sdk.ZeroInt(), sdk.ZeroInt(), pos.lowerTick, pos.upperTick)
				s.Require().NoError(err)
			}

			// Increase block time by one more day
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour * 24))

			// First run non mutative validation and check results
			poolId, lowerTick, upperTick, liquidity, err := s.App.ConcentratedLiquidityKeeper.ValidatePositionsAndGetTotalLiquidity(s.Ctx, test.accountCallingMigration, test.positionIdsToMigrate)
			if test.expectedErr != nil {
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

			unclaimedRewardsForAllOldPositions := make([]sdk.DecCoins, len(uptimeAccumulators))

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
						unclaimedRewardsForAllOldPositions[i] = unclaimedRewardsForAllOldPositions[i].Add(unclaimedRewardsForPosition...)

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
				s.Require().Equal(s.Ctx.BlockTime().Add(-cl.FullyChargedDuration), newPosition.JoinTime)

				// Get the uptime accumulators for the poolId of the new position.
				uptimeAccumulators, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, poolId)
				s.Require().NoError(err)

				unclaimedRewardsForNewPosition := make([]sdk.DecCoins, len(uptimeAccumulators))

				// Get the unclaimed rewards for the new position
				for i, uptimeAccum := range uptimeAccumulators {
					newPositionName := string(types.KeyPositionId(newPositionId))
					// Check if the accumulator contains the position.
					hasPosition, err := uptimeAccum.HasPosition(newPositionName)
					s.Require().NoError(err)
					// If the accumulator contains the position, move the unclaimed rewards to the new position.
					if hasPosition {
						// Get the unclaimed rewards for the old position.
						position, err := accum.GetPosition(uptimeAccum, newPositionName)
						s.Require().NoError(err)

						unclaimedRewardsForPosition := accum.GetTotalRewards(uptimeAccum, position)

						unclaimedRewardsForNewPosition[i] = unclaimedRewardsForNewPosition[i].Add(unclaimedRewardsForPosition...)

					}
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
				}

				// The new position's unclaimed rewards should be the sum of the old positions' unclaimed rewards.
				s.Require().Equal(unclaimedRewardsForAllOldPositions, unclaimedRewardsForNewPosition)

				// Claim all the rewards for the new position and check that the rewards match the unclaimed rewards.
				claimedRewards, forfeitedRewards, err := s.App.ConcentratedLiquidityKeeper.ClaimAllIncentivesForPosition(s.Ctx, newPositionId)
				s.Require().NoError(err)

				for _, coin := range unclaimedRewardsForNewPosition {
					for _, c := range coin {
						s.Require().Equal(c.Amount.TruncateInt().String(), claimedRewards.AmountOf(c.Denom).String())
					}
				}

				s.Require().Equal(sdk.Coins{}, forfeitedRewards)

			}
		})
	}
}

func (s *KeeperTestSuite) TestHasAnyPosition() {
	s.Setup()
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
			s.Setup()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			// Create a default CL pools
			s.PrepareConcentratedPool()

			for _, pos := range test.setupPositions {
				s.SetupPosition(pos.PoolId, sdk.AccAddress(pos.Address), DefaultCoin0, DefaultCoin1, pos.LowerTick, pos.UpperTick, DefaultJoinTime)
			}

			// System under test
			actualResult, err := s.App.ConcentratedLiquidityKeeper.HasAnyPositionForPool(s.Ctx, test.poolId)

			s.Require().NoError(err)
			s.Require().Equal(test.expectedResult, actualResult)
		})
	}
}

func (s *KeeperTestSuite) TestCreateFullRangePosition() {
	s.Setup()
	defaultAddress := s.TestAccs[0]
	DefaultJoinTime := s.Ctx.BlockTime()
	defaultPositionCoins := sdk.NewCoins(DefaultCoin0, DefaultCoin1)

	tests := []struct {
		name                  string
		remainingLockDuration time.Duration
		isLocked              bool
		isUnlocking           bool
	}{
		{
			name: "full range position",
		},
		{
			name:                  "full range position: locked",
			remainingLockDuration: 24 * time.Hour * 14,
			isLocked:              true,
		},
		{
			name:                  "full range position: unlocking",
			remainingLockDuration: 24 * time.Hour,
			isUnlocking:           true,
		},
	}

	for _, test := range tests {
		test := test
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()
			s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)

			// Create a default CL pools
			clPool := s.PrepareConcentratedPool()

			var (
				positionId         uint64
				liquidity          sdk.Dec
				concentratedLockId uint64
				err                error
			)

			// Fund the owner account
			s.FundAcc(defaultAddress, defaultPositionCoins)

			// System under test
			if test.isLocked {
				positionId, _, _, liquidity, _, concentratedLockId, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, clPool, defaultAddress, defaultPositionCoins, test.remainingLockDuration)
			} else if test.isUnlocking {
				positionId, _, _, liquidity, _, concentratedLockId, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionUnlocking(s.Ctx, clPool, defaultAddress, defaultPositionCoins, test.remainingLockDuration)
			} else {
				positionId, _, _, liquidity, _, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool, defaultAddress, defaultPositionCoins)
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
	}

	for _, test := range tests {
		test := test
		s.Run(test.name, func() {
			// Init suite for each test.
			s.Setup()
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
				positionId, _, _, liquidity, _, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool, test.owner, defaultPositionCoins)
				s.Require().NoError(err)
			}

			lockupModuleAccountBalancePre := s.App.LockupKeeper.GetModuleBalance(s.Ctx)

			// System under test
			concentratedLockId, underlyingLiquidityTokenized, err := s.App.ConcentratedLiquidityKeeper.MintSharesLockAndUpdate(s.Ctx, clPool.GetId(), positionId, test.owner, test.remainingLockDuration, liquidity)
			s.Require().NoError(err)

			lockupModuleAccountBalancePost := s.App.LockupKeeper.GetModuleBalance(s.Ctx)

			// Check that the lockup module account balance increased by the amount expected to be locked
			s.Require().Equal(underlyingLiquidityTokenized[0].String(), lockupModuleAccountBalancePost.Sub(lockupModuleAccountBalancePre).String())

			// Check that the positionId is mapped to the lockId
			positionLockId, err := s.App.ConcentratedLiquidityKeeper.GetPositionIdToLock(s.Ctx, positionId)
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

func (s *KeeperTestSuite) TestPositionToLockCRUD() {
	// Init suite for each test.
	s.Setup()
	s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)
	owner := s.TestAccs[0]
	remainingLockDuration := 24 * time.Hour
	defaultPositionCoins := sdk.NewCoins(DefaultCoin0, DefaultCoin1)

	// Create a default CL pools
	clPool := s.PrepareConcentratedPool()

	// Fund the owner account
	s.FundAcc(owner, defaultPositionCoins)

	// Create a position with a lock
	positionId, _, _, _, _, concentratedLockId, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionUnlocking(s.Ctx, clPool, owner, defaultPositionCoins, remainingLockDuration)
	s.Require().NoError(err)

	// We should be able to retrieve the lockId from the positionId now
	retrievedLockId, err := s.App.ConcentratedLiquidityKeeper.GetPositionIdToLock(s.Ctx, positionId)
	s.Require().NoError(err)
	s.Require().Equal(concentratedLockId, retrievedLockId)

	// Check if position has lock in state, should be true
	hasLockInState, err := s.App.ConcentratedLiquidityKeeper.PositionHasUnderlyingLockInState(s.Ctx, positionId)
	s.Require().NoError(err)
	s.Require().True(hasLockInState)

	// Create a position without a lock
	positionId, _, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool, owner, defaultPositionCoins)

	// Check if position has lock in state, should not
	retrievedLockId, err = s.App.ConcentratedLiquidityKeeper.GetPositionIdToLock(s.Ctx, positionId)
	s.Require().Error(err)
	s.Require().Equal(uint64(0), retrievedLockId)

	// Check if position has lock in state, should be false
	hasLockInState, err = s.App.ConcentratedLiquidityKeeper.PositionHasUnderlyingLockInState(s.Ctx, positionId)
	s.Require().NoError(err)
	s.Require().False(hasLockInState)

	// Set the position to have a lockId (despite it not actually having a lock)
	s.App.ConcentratedLiquidityKeeper.SetPositionIdToLock(s.Ctx, positionId, concentratedLockId)

	// Check if position has lock in state, it should now
	retrievedLockId, err = s.App.ConcentratedLiquidityKeeper.GetPositionIdToLock(s.Ctx, positionId)
	s.Require().NoError(err)
	s.Require().Equal(concentratedLockId, retrievedLockId)

	// Check if position has lock in state, should now be true
	hasLockInState, err = s.App.ConcentratedLiquidityKeeper.PositionHasUnderlyingLockInState(s.Ctx, positionId)
	s.Require().NoError(err)
	s.Require().True(hasLockInState)

	// Remove the lockId from the position
	s.App.ConcentratedLiquidityKeeper.RemovePositionIdToLock(s.Ctx, positionId)

	// Check if position has lock in state, should not
	retrievedLockId, err = s.App.ConcentratedLiquidityKeeper.GetPositionIdToLock(s.Ctx, positionId)
	s.Require().Error(err)
	s.Require().Equal(uint64(0), retrievedLockId)

	// Check if position has lock in state, should be false
	hasLockInState, err = s.App.ConcentratedLiquidityKeeper.PositionHasUnderlyingLockInState(s.Ctx, positionId)
	s.Require().NoError(err)
	s.Require().False(hasLockInState)

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
		s.Setup()
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
		s.Setup()
		s.Ctx = s.Ctx.WithBlockTime(DefaultJoinTime)
		owner := s.TestAccs[0]
		s.FundAcc(owner, tc.positionCoins)

		// Create a new pool.
		clPool := s.PrepareConcentratedPool()
		clPoolId := clPool.GetId()
		actualFullRangeLiquidity := sdk.ZeroDec()

		// Create a full range position.
		_, _, _, liquidity, _, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool, owner, tc.positionCoins)
		s.Require().NoError(err)
		actualFullRangeLiquidity = actualFullRangeLiquidity.Add(liquidity)

		clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPoolId)
		s.Require().NoError(err)

		// Get the full range liquidity for the pool.
		expectedFullRangeLiquidity := s.App.ConcentratedLiquidityKeeper.MustGetFullRangeLiquidityInPool(s.Ctx, clPoolId)
		s.Require().Equal(expectedFullRangeLiquidity, actualFullRangeLiquidity)

		// Create a new position that overlaps with the min tick, but is not full range and therefore should not count towards the full range liquidity.
		s.FundAcc(owner, tc.positionCoins)
		_, _, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, clPoolId, owner, DefaultAmt0, DefaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), tc.lowerTick, tc.upperTick)
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
