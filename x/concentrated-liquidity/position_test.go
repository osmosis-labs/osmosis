package concentrated_liquidity_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v16/app/apptesting"
	cl "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
)

const (
	DefaultFungifyNumPositions       = 3
	DefaultFungifyFullChargeDuration = 24 * time.Hour
)

var (
	DefaultIncentiveRecords = []types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo, incentiveRecordThree, incentiveRecordFour}
	DefaultBlockTime        = time.Unix(1, 1).UTC()
	DefaultSpreadFactor     = sdk.NewDecWithPrec(2, 3)
)

// AssertPositionsDoNotExist checks that the positions with the given IDs do not exist on uptime accumulators.
func (s *KeeperTestSuite) AssertPositionsDoNotExist(positionIds []uint64) {
	uptimeAccumulators, err := s.clk.GetUptimeAccumulators(s.Ctx, defaultPoolId)
	s.Require().NoError(err)

	for _, positionId := range positionIds {
		oldPositionName := string(types.KeyPositionId(positionId))
		for _, uptimeAccum := range uptimeAccumulators {
			// Check if the accumulator contains the position.
			hasPosition := uptimeAccum.HasPosition(oldPositionName)
			s.Require().False(hasPosition)
		}

		// Check that the old position has been deleted.
		_, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionId)
		s.Require().Error(err)
		s.Require().ErrorAs(err, &types.PositionIdNotFoundError{})
	}
}

// GetTotalAccruedRewardsByAccumulator returns the total accrued rewards for the given position on each uptime accumulator.
func (s *KeeperTestSuite) GetTotalAccruedRewardsByAccumulator(positionId uint64, requireHasPosition bool) []sdk.DecCoins {
	uptimeAccumulators, err := s.clk.GetUptimeAccumulators(s.Ctx, defaultPoolId)
	s.Require().NoError(err)

	unclaimedRewardsForEachUptimeNewPosition := make([]sdk.DecCoins, len(uptimeAccumulators))

	for i, uptimeAccum := range uptimeAccumulators {
		newPositionName := string(types.KeyPositionId(positionId))
		// Check if the accumulator contains the position.
		hasPosition := uptimeAccum.HasPosition(newPositionName)
		if requireHasPosition {
			s.Require().True(hasPosition)
		}

		if hasPosition {
			// Move the unclaimed rewards to the new position.
			// Get the unclaimed rewards for the old position.
			position, err := accum.GetPosition(uptimeAccum, newPositionName)
			s.Require().NoError(err)

			unclaimedRewardsForPosition := accum.GetTotalRewards(uptimeAccum, position)

			unclaimedRewardsForEachUptimeNewPosition[i] = unclaimedRewardsForEachUptimeNewPosition[i].Add(unclaimedRewardsForPosition...)
		}
	}

	return unclaimedRewardsForEachUptimeNewPosition
}

// ExecuteAndValidateSuccessfulIncentiveClaim claims incentives for position Id and asserts its output is as expected.
// It also asserts that no more incentives can be claimed for the position.
func (s *KeeperTestSuite) ExecuteAndValidateSuccessfulIncentiveClaim(positionId uint64, expectedRewards sdk.Coins, expectedForfeited sdk.Coins) {
	// Initial claim and assertion
	claimedRewards, forfeitedRewards, err := s.clk.PrepareClaimAllIncentivesForPosition(s.Ctx, positionId)
	s.Require().NoError(err)

	s.Require().Equal(expectedRewards, claimedRewards)
	s.Require().Equal(expectedForfeited, forfeitedRewards)

	// Sanity check that cannot claim again.
	claimedRewards, _, err = s.clk.PrepareClaimAllIncentivesForPosition(s.Ctx, positionId)
	s.Require().NoError(err)

	s.Require().Equal(sdk.Coins(nil), claimedRewards)
}

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

				recordExists := newUptimeAccums[uptimeIndex].HasPosition(positionName)
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
					if timeElapsedSec.IsPositive() {
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

type positionOwnershipTest struct {
	queryPositionOwner sdk.AccAddress
	queryPositionId    uint64
	expPass            bool

	setupPositions []sdk.AccAddress
	poolId         uint64
}

func (s *KeeperTestSuite) TestGetUserPositions() {
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
				{1, 1, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick},
			},
		},
		{
			name:   "Get current users multiple position same pool",
			sender: defaultAddress,
			setupPositions: []position{
				{1, 1, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick},
				{2, 1, defaultAddress, DefaultCoins, DefaultLowerTick + 100, DefaultUpperTick + 100},
				{3, 1, defaultAddress, DefaultCoins, DefaultLowerTick + 200, DefaultUpperTick + 200},
			},
		},
		{
			name:   "Get current users multiple position multiple pools",
			sender: secondAddress,
			setupPositions: []position{
				{1, 1, secondAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick},
				{2, 2, secondAddress, DefaultCoins, DefaultLowerTick + 100, DefaultUpperTick + 100},
				{3, 3, secondAddress, DefaultCoins, DefaultLowerTick + 200, DefaultUpperTick + 200},
			},
		},
		{
			name:   "User has positions over multiple pools, but filter by one pool",
			sender: secondAddress,
			poolId: 2,
			setupPositions: []position{
				{1, 1, secondAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick},
				{2, 2, secondAddress, DefaultCoins, DefaultLowerTick + 100, DefaultUpperTick + 100},
				{3, 3, secondAddress, DefaultCoins, DefaultLowerTick + 200, DefaultUpperTick + 200},
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
				liquidity, _ := s.SetupPosition(pos.poolId, pos.acc, pos.coins, pos.lowerTick, pos.upperTick, false)
				if pos.acc.Equals(pos.acc) {
					if test.poolId == 0 || test.poolId == pos.poolId {
						expectedUserPositions = append(expectedUserPositions, model.Position{
							PositionId: pos.positionId,
							PoolId:     pos.poolId,
							Address:    pos.acc.String(),
							LowerTick:  pos.lowerTick,
							UpperTick:  pos.upperTick,
							JoinTime:   s.Ctx.BlockTime(),
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

func (s *KeeperTestSuite) TestGetUserPositionsSerialized() {
	type position struct {
		positionId uint64
		poolId     uint64
		acc        sdk.AccAddress
		coins      sdk.Coins
		lowerTick  int64
		upperTick  int64
	}

	defaultAddress := s.TestAccs[0]
	alternateAddress := s.TestAccs[1]

	tests := []struct {
		name                    string
		addressToQuery          sdk.AccAddress
		poolIdToQuery           uint64
		paginationLimit         uint64
		expectedNumberOfRecords int
	}{
		{
			name:                    "Get default users positions in all pools",
			addressToQuery:          defaultAddress,
			poolIdToQuery:           0,
			expectedNumberOfRecords: 5,
			paginationLimit:         10,
		},
		{
			name:                    "Get default users positions in pool 1",
			addressToQuery:          defaultAddress,
			poolIdToQuery:           1,
			expectedNumberOfRecords: 4,
			paginationLimit:         10,
		},
		{
			name:                    "Get default users positions in pool 1, cut off last record with pagination",
			addressToQuery:          defaultAddress,
			poolIdToQuery:           1,
			expectedNumberOfRecords: 3,
			paginationLimit:         3,
		},
		{
			name:                    "Get default users positions in pool 1, cut off last record with pagination",
			addressToQuery:          defaultAddress,
			poolIdToQuery:           1,
			expectedNumberOfRecords: 3,
			paginationLimit:         3,
		},
		{
			name:                    "Get alternate users positions in pool 1",
			addressToQuery:          alternateAddress,
			poolIdToQuery:           1,
			expectedNumberOfRecords: 1,
			paginationLimit:         10,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {

			s.SetupTest()
			k := s.App.ConcentratedLiquidityKeeper

			//s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0), sdk.NewCoin(USDC, DefaultAmt1)))

			s.PrepareConcentratedPool()
			s.PrepareConcentratedPool()

			positions := []position{
				{1, 1, defaultAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick},
				{2, 1, defaultAddress, DefaultCoins, DefaultLowerTick + 100, DefaultUpperTick + 100},
				{3, 1, defaultAddress, DefaultCoins, DefaultLowerTick + 200, DefaultUpperTick + 200},
				{4, 1, defaultAddress, DefaultCoins, DefaultLowerTick + 300, DefaultUpperTick + 300},
				{5, 1, alternateAddress, DefaultCoins, DefaultLowerTick, DefaultUpperTick},
				{6, 2, defaultAddress, DefaultCoins, DefaultLowerTick + 100, DefaultUpperTick + 100},
				{7, 2, alternateAddress, DefaultCoins, DefaultLowerTick + 200, DefaultUpperTick + 200},
			}

			expectedUserPositions := []model.FullPositionBreakdown{}
			count := test.paginationLimit
			for _, pos := range positions {
				// if position does not exist this errors
				s.SetupPosition(pos.poolId, pos.acc, pos.coins, pos.lowerTick, pos.upperTick, false)
				if pos.acc.Equals(test.addressToQuery) && (test.poolIdToQuery == 0 || test.poolIdToQuery == pos.poolId) && count > 0 {
					position, err := k.GetPosition(s.Ctx, pos.positionId)
					s.Require().NoError(err)

					positionPool, err := k.GetConcentratedPoolById(s.Ctx, position.PoolId)
					s.Require().NoError(err)

					asset0, asset1, err := cl.CalculateUnderlyingAssetsFromPosition(s.Ctx, position, positionPool)
					s.Require().NoError(err)

					claimableSpreadRewards, err := k.GetClaimableSpreadRewards(s.Ctx, pos.positionId)
					s.Require().NoError(err)

					claimableIncentives, forfeitedIncentives, err := k.GetClaimableIncentives(s.Ctx, pos.positionId)
					s.Require().NoError(err)

					expectedUserPositions = append(expectedUserPositions, model.FullPositionBreakdown{
						Position:               position,
						Asset0:                 asset0,
						Asset1:                 asset1,
						ClaimableSpreadRewards: claimableSpreadRewards,
						ClaimableIncentives:    claimableIncentives,
						ForfeitedIncentives:    forfeitedIncentives,
					})
					count--
				}
			}

			paginationReq := &query.PageRequest{
				Limit:      test.paginationLimit,
				CountTotal: true,
			}

			userPositions, _, err := k.GetUserPositionsSerialized(s.Ctx, test.addressToQuery, test.poolIdToQuery, paginationReq)
			s.Require().NoError(err)
			s.Require().Equal(test.expectedNumberOfRecords, len(userPositions))
			s.Require().Equal(expectedUserPositions, userPositions)
		})
	}
}

func (s *KeeperTestSuite) TestDeletePosition() {
	defaultPoolId := uint64(1)
	DefaultJoinTime := s.Ctx.BlockTime()
	defaultCreator := s.TestAccs[0]

	tests := []struct {
		name             string
		positionId       uint64
		underlyingLockId uint64
		creator          sdk.AccAddress
		poolId           uint64
		expectedErr      error
	}{
		{
			name:             "Valid case: Delete position info on existing pool and existing position (no underlying lock)",
			underlyingLockId: 0,
			poolId:           defaultPoolId,
			creator:          defaultCreator,
			positionId:       DefaultPositionId,
		},
		{
			name:             "Valid case: Delete position info on existing pool and existing position (has underlying lock)",
			underlyingLockId: 1,
			poolId:           defaultPoolId,
			creator:          defaultCreator,
			positionId:       DefaultPositionId,
		},
		{
			name:             "Invalid case: Delete a non existing position",
			positionId:       DefaultPositionId + 1,
			underlyingLockId: 0,
			poolId:           defaultPoolId,
			creator:          defaultCreator,
			expectedErr:      types.PositionIdNotFoundError{PositionId: DefaultPositionId + 1},
		},
		{
			name:             "Invalid case: non-existent the address-pool-position ID to position mapping",
			positionId:       DefaultPositionId,
			underlyingLockId: 0,
			poolId:           defaultPoolId,
			creator:          s.TestAccs[1],
			expectedErr:      types.AddressPoolPositionIdNotFoundError{PoolId: DefaultPositionId, PositionId: DefaultPositionId, Owner: s.TestAccs[1].String()},
		},
		{
			name:             "Invalid case: non-existent pool-position ID mapping",
			poolId:           3,
			positionId:       DefaultPositionId,
			creator:          defaultCreator,
			underlyingLockId: 0,
			expectedErr:      types.PoolPositionIdNotFoundError{PoolId: 3, PositionId: DefaultPositionId},
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
			err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePosition(s.Ctx, validPoolId, defaultCreator, DefaultLowerTick, DefaultUpperTick, DefaultLiquidityAmt, DefaultJoinTime, DefaultPositionId)
			s.Require().NoError(err)

			if test.underlyingLockId != 0 {
				err = s.App.ConcentratedLiquidityKeeper.SetPosition(s.Ctx, validPoolId, defaultCreator, DefaultLowerTick, DefaultUpperTick, DefaultJoinTime, DefaultLiquidityAmt, 1, test.underlyingLockId)
				s.Require().NoError(err)
			}

			// Check stores exist
			// Retrieve the position from the store via position ID and compare to expected values.
			position := model.Position{}
			positionIdToPositionKey := types.KeyPositionId(DefaultPositionId)
			osmoutils.MustGet(store, positionIdToPositionKey, &position)
			s.Require().Equal(DefaultPositionId, position.PositionId)
			s.Require().Equal(defaultPoolId, position.PoolId)
			s.Require().Equal(defaultCreator.String(), position.Address)
			s.Require().Equal(DefaultLowerTick, position.LowerTick)
			s.Require().Equal(DefaultUpperTick, position.UpperTick)
			s.Require().Equal(DefaultJoinTime, position.JoinTime)
			s.Require().Equal(DefaultLiquidityAmt, position.Liquidity)

			// Retrieve the position ID from the store via owner/poolId key and compare to expected value (true).
			ownerPoolIdToPositionIdKey := types.KeyAddressPoolIdPositionId(defaultCreator, defaultPoolId, DefaultPositionId)
			valueBytes := store.Get(ownerPoolIdToPositionIdKey)
			s.Require().Equal([]byte{1}, valueBytes)

			// Retrieve the position ID from the store via poolId key and compare to expected value (true).
			poolIdtoPositionIdKey := types.KeyPoolPositionPositionId(defaultPoolId, DefaultPositionId)
			valueBytes = store.Get(poolIdtoPositionIdKey)
			s.Require().Equal([]byte{1}, valueBytes)

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
			positionIdBytes := store.Get(lockIdToPositionIdKey)
			if test.underlyingLockId != 0 {
				s.Require().Equal(DefaultPositionId, sdk.BigEndianToUint64(positionIdBytes))
			} else {
				s.Require().Nil(positionIdBytes)
			}

			err = s.App.ConcentratedLiquidityKeeper.DeletePosition(s.Ctx, test.positionId, test.creator, test.poolId)
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
				ownerPoolIdToPositionIdKey = types.KeyAddressPoolIdPositionId(defaultCreator, defaultPoolId, DefaultPositionId)
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
			_, _, _, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, 1, s.TestAccs[0], DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
			s.Require().NoError(err)

			// create a position from the test case
			s.FundAcc(s.TestAccs[1], sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0), sdk.NewCoin(USDC, DefaultAmt1)))
			_, actualAmount0, actualAmount1, liquidity, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, tc.position.PoolId, s.TestAccs[1], DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), tc.position.LowerTick, tc.position.UpperTick)
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
		},
		{
			name:                       "Error: No position to fungify",
			setupFullyChargedPositions: []position{},
			setupUnchargedPositions:    []position{},
			positionIdsToMigrate:       []uint64{1},
			accountCallingMigration:    defaultAddress,
			expectedNewPositionId:      0,
			expectedErr:                types.PositionQuantityTooLowError{MinNumPositions: cl.MinNumPositions, NumPositions: 1},
		},
		{
			name: "Error: one of the full range positions is locked",
			setupFullyChargedPositions: []position{
				{1, defaultPoolId, defaultAddress, DefaultCoins, types.MinInitializedTick, types.MaxTick, unlocked},
				{2, defaultPoolId, defaultAddress, DefaultCoins, types.MinInitializedTick, types.MaxTick, locked},
				{3, defaultPoolId, defaultAddress, DefaultCoins, types.MinInitializedTick, types.MaxTick, unlocked},
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
				{1, defaultPoolId, defaultAddress, DefaultCoins, types.MinInitializedTick, types.MaxTick, unlocked},
				{2, defaultPoolId, defaultAddress, DefaultCoins, types.MinInitializedTick, types.MaxTick, locked},
				{3, defaultPoolId, defaultAddress, DefaultCoins, types.MinInitializedTick, types.MaxTick, unlocked},
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
			err := s.clk.SetMultipleIncentiveRecords(s.Ctx, DefaultIncentiveRecords)
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
					_, _, _, liquidityCreated, _, err = s.clk.CreateFullRangePositionUnlocking(s.Ctx, pos.poolId, pos.acc, pos.coins, lockDuration)
					s.Require().NoError(err)
				} else {
					_, _, _, liquidityCreated, _, _, err = s.clk.CreatePosition(s.Ctx, pos.poolId, pos.acc, pos.coins, sdk.ZeroInt(), sdk.ZeroInt(), pos.lowerTick, pos.upperTick)
					s.Require().NoError(err)
				}

				totalLiquidity = totalLiquidity.Add(liquidityCreated)
			}

			// Increase block time by the fully charged duration to make sure previously added positions are charged.
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(testFullChargeDuration))

			// Set up uncharged positions
			for _, pos := range test.setupUnchargedPositions {
				_, _, _, _, _, _, err := s.clk.CreatePosition(s.Ctx, pos.poolId, pos.acc, pos.coins, sdk.ZeroInt(), sdk.ZeroInt(), pos.lowerTick, pos.upperTick)
				s.Require().NoError(err)
			}

			// Increase block time by one more day - 1 ns to ensure that the previously added positions are not fully charged.
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(testFullChargeDuration - time.Nanosecond))

			// Get the longest authorized uptime, which is the fully charged duration.
			fullyChargedDuration := s.clk.GetLargestAuthorizedUptimeDuration(s.Ctx)

			// First run non mutative validation and check results
			poolId, lowerTick, upperTick, liquidity, err := s.clk.ValidatePositionsAndGetTotalLiquidity(s.Ctx, test.accountCallingMigration, test.positionIdsToMigrate, fullyChargedDuration)
			if test.expectedErr != nil {
				s.Require().ErrorIs(err, test.expectedErr)
				s.Require().Equal(uint64(0), poolId)
				s.Require().Equal(int64(0), lowerTick)
				s.Require().Equal(int64(0), upperTick)
				s.Require().Equal(sdk.Dec{}, liquidity)
			} else {
				s.Require().NoError(err)

				// Check that the poolId, lowerTick, upperTick, and liquidity are correct
				for _, posId := range test.positionIdsToMigrate {
					position, err := s.clk.GetPosition(s.Ctx, posId)
					s.Require().NoError(err)
					s.Require().Equal(poolId, position.PoolId)
					s.Require().Equal(lowerTick, position.LowerTick)
					s.Require().Equal(upperTick, position.UpperTick)
				}
				s.Require().Equal(totalLiquidity, liquidity)
			}

			// Update the accumulators for defaultPoolId to the current time
			err = s.clk.UpdatePoolUptimeAccumulatorsToNow(s.Ctx, defaultPoolId)
			s.Require().NoError(err)

			// Get the unclaimed rewards for all the positions that are being migrated
			unclaimedRewardsForEachUptimeAcrossAllOldPositions := make([]sdk.DecCoins, len(types.SupportedUptimes))
			for _, positionId := range test.positionIdsToMigrate {
				unclaimedRewardsForPosition := s.GetTotalAccruedRewardsByAccumulator(positionId, false)
				unclaimedRewardsForEachUptimeAcrossAllOldPositions, err = osmoutils.AddDecCoinArrays(unclaimedRewardsForEachUptimeAcrossAllOldPositions, unclaimedRewardsForPosition)
				s.Require().NoError(err)
			}

			// Next, run the mutative function and check results
			newPositionId, err := s.clk.FungifyChargedPosition(s.Ctx, test.accountCallingMigration, test.positionIdsToMigrate)
			if test.expectedErr != nil {
				s.Require().ErrorIs(err, test.expectedErr)
				s.Require().Equal(uint64(0), newPositionId)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(test.expectedNewPositionId, newPositionId)

				// Since the positionLiquidity of the old position should have been deleted, retrieving it should return an error.
				for _, posId := range test.positionIdsToMigrate {
					positionLiquidity, err := s.clk.GetPositionLiquidity(s.Ctx, posId)
					s.Require().ErrorIs(err, types.PositionIdNotFoundError{PositionId: posId})
					s.Require().Equal(sdk.Dec{}, positionLiquidity)
				}

				// --- New position assertions ---

				newPosition, err := s.clk.GetPosition(s.Ctx, newPositionId)

				// Check that the liquidity is equal to the sum of the old positions.
				s.Require().NoError(err)
				s.Require().Equal(totalLiquidity, newPosition.Liquidity)

				// The new position's join time should be the current block time minus the fully charged duration.
				fullCharge := s.clk.GetLargestAuthorizedUptimeDuration(s.Ctx)
				s.Require().Equal(s.Ctx.BlockTime().Add(-fullCharge), newPosition.JoinTime)

				// Get the unclaimed rewards for the new position
				unclaimedRewardsForEachUptimeNewPosition := s.GetTotalAccruedRewardsByAccumulator(newPositionId, true)

				// Check that the old positions have been deleted.
				s.AssertPositionsDoNotExist(test.positionIdsToMigrate)

				// The new position's unclaimed rewards should be the sum of the old positions' unclaimed rewards.
				s.Require().Equal(unclaimedRewardsForEachUptimeAcrossAllOldPositions, unclaimedRewardsForEachUptimeNewPosition)

				// Get the final amount expected to be claimed by merging the unclaimed rewards across all uptimes.
				// Note that the second value is dust, not an error.
				expectedRewardsToClaim, _ := osmoutils.CollapseDecCoinsArray(unclaimedRewardsForEachUptimeAcrossAllOldPositions).TruncateDecimal()

				// Claim all the rewards for the new position and check that the rewards match the unclaimed rewards.
				s.ExecuteAndValidateSuccessfulIncentiveClaim(newPositionId, expectedRewardsToClaim, sdk.Coins(nil))

				// Check that cannot claim rewards for the old positions.
				for _, positionId := range test.positionIdsToMigrate {
					_, _, err := s.clk.PrepareClaimAllIncentivesForPosition(s.Ctx, positionId)
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
				s.SetupPosition(pos.PoolId, sdk.AccAddress(pos.Address), DefaultCoins, pos.LowerTick, pos.UpperTick, false)
			}

			// System under test
			actualResult, err := s.App.ConcentratedLiquidityKeeper.HasAnyPositionForPool(s.Ctx, test.poolId)

			s.Require().NoError(err)
			s.Require().Equal(test.expectedResult, actualResult)
		})
	}
}

// This test specifically tests that spread reward collection works as expected
// after fungifying positions.
func (s *KeeperTestSuite) TestFungifyChargedPositions_SwapAndClaimSpreadRewards() {
	// Init suite for the test.
	s.SetupTest()

	const swapAmount = 1_000_000
	var defaultAddress = s.TestAccs[0]

	// Set up pool, positions, and incentive records
	_, expectedPositionIds, totalLiquidity := s.runFungifySetup(defaultAddress, DefaultFungifyNumPositions, DefaultFungifyFullChargeDuration, DefaultSpreadFactor, DefaultIncentiveRecords)

	// Perform a swap to earn spread rewards
	swapAmountIn := sdk.NewCoin(ETH, sdk.NewInt(swapAmount))
	expectedSpreadReward := swapAmountIn.Amount.ToDec().Mul(DefaultSpreadFactor)
	// We run expected spread rewards through a cycle of divison and multiplication by liquidity to capture appropriate rounding behavior.
	// Note that we truncate the int at the end since it is not possible to have a decimal spread reward amount collected (the QuoTruncate
	// and MulTruncates are much smaller operations that round down for values past the 18th decimal place).
	expectedSpreadRewardTruncated := expectedSpreadReward.QuoTruncate(totalLiquidity).MulTruncate(totalLiquidity).TruncateInt()
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(swapAmountIn))
	s.swapAndTrackXTimesInARow(defaultPoolId, swapAmountIn, USDC, types.MinSpotPrice, 1)

	// Increase block time by the fully charged duration
	s.AddBlockTime(DefaultFungifyFullChargeDuration)

	// First run non mutative validation and check results
	newPositionId, err := s.clk.FungifyChargedPosition(s.Ctx, defaultAddress, expectedPositionIds)
	s.Require().NoError(err)

	// Claim spread rewards
	collected, err := s.clk.CollectSpreadRewards(s.Ctx, defaultAddress, newPositionId)
	s.Require().NoError(err)

	// Validate that the correct spread reward amount was collected.
	s.Require().Equal(expectedSpreadRewardTruncated, collected.AmountOf(swapAmountIn.Denom))

	// Check that cannot claim again.
	collected, err = s.clk.CollectSpreadRewards(s.Ctx, defaultAddress, newPositionId)
	s.Require().NoError(err)
	s.Require().Equal(sdk.Coins(nil), collected)

	spreadRewardAccum, err := s.clk.GetSpreadRewardAccumulator(s.Ctx, defaultPoolId)
	s.Require().NoError(err)

	// Check that cannot claim old positions
	for _, oldPositionId := range expectedPositionIds {
		collected, err = s.clk.CollectSpreadRewards(s.Ctx, defaultAddress, oldPositionId)
		s.Require().Error(err)
		s.Require().Equal(sdk.Coins{}, collected)

		hasPosition := s.clk.HasPosition(s.Ctx, oldPositionId)
		s.Require().False(hasPosition)

		hasSpreadRewardPositionTracker := spreadRewardAccum.HasPosition(types.KeySpreadRewardPositionAccumulator(oldPositionId))
		s.Require().False(hasSpreadRewardPositionTracker)
	}
}

func (s *KeeperTestSuite) TestFungifyChargedPositions_ClaimIncentives() {
	// Init suite for the test.
	s.SetupTest()
	var defaultAddress = s.TestAccs[0]

	// Set incentives for pool to ensure accumulators work correctly
	testIncentiveRecord := types.IncentiveRecord{
		PoolId: 1,
		IncentiveRecordBody: types.IncentiveRecordBody{
			RemainingCoin: sdk.NewDecCoinFromDec(USDC, sdk.NewDec(1000000000000000000)),
			EmissionRate:  sdk.NewDec(1), // 1 per second
			StartTime:     defaultBlockTime,
		},
		MinUptime: time.Nanosecond,
	}

	// Set up pool, positions, and incentive records
	pool, expectedPositionIds, _ := s.runFungifySetup(defaultAddress, DefaultFungifyNumPositions, DefaultFungifyFullChargeDuration, DefaultSpreadFactor, []types.IncentiveRecord{testIncentiveRecord})

	// an error of 1 for each position
	roundingError := int64(DefaultFungifyNumPositions)
	roundingTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: sdk.NewDec(roundingError),
		RoundingDir:       osmomath.RoundDown,
	}
	expectedAmount := sdk.NewInt(60 * 60 * 24) // 1 day in seconds * 1 per second
	s.FundAcc(pool.GetIncentivesAddress(), sdk.NewCoins(sdk.NewCoin(USDC, expectedAmount)))

	// Increase block time by the fully charged duration
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(DefaultFungifyFullChargeDuration))

	// sync accumulators
	// We use cache context to update uptime accumulators for estimating claimable incentives
	// prior to running fungify. However, we do not want the mutations made in test setup to have
	// impact on the system under test because it (fungify) must update the uptime accumulators itself.
	cacheCtx, _ := s.Ctx.CacheContext()
	err := s.clk.UpdatePoolUptimeAccumulatorsToNow(cacheCtx, pool.GetId())
	s.Require().NoError(err)

	claimableIncentives := sdk.NewCoins()
	for i := 0; i < DefaultFungifyNumPositions; i++ {
		positionIncentives, forfeitedIncentives, err := s.clk.GetClaimableIncentives(cacheCtx, uint64(i+1))
		s.Require().NoError(err)
		s.Require().Equal(sdk.Coins(nil), forfeitedIncentives)
		claimableIncentives = claimableIncentives.Add(positionIncentives...)
	}

	actualClaimedAmount := claimableIncentives.AmountOf(USDC)
	s.Require().Equal(0, roundingTolerance.Compare(expectedAmount, actualClaimedAmount), "expected: %s, got: %s", expectedAmount, actualClaimedAmount)

	// System under test
	newPositionId, err := s.clk.FungifyChargedPosition(s.Ctx, defaultAddress, expectedPositionIds)
	s.Require().NoError(err)

	// Claim incentives.
	collected, _, err := s.clk.CollectIncentives(s.Ctx, defaultAddress, newPositionId)
	s.Require().NoError(err)

	// Validate that the correct incentives amount was collected.
	actualClaimedAmount = collected.AmountOf(USDC)
	s.Require().Equal(1, len(collected))
	s.Require().Equal(0, roundingTolerance.Compare(expectedAmount, actualClaimedAmount), "expected: %s, got: %s", expectedAmount, actualClaimedAmount)

	// Check that cannot claim again.
	collected, _, err = s.clk.CollectIncentives(s.Ctx, defaultAddress, newPositionId)
	s.Require().NoError(err)
	s.Require().Equal(sdk.Coins(nil), collected)

	// Check that cannot claim old positions
	for i := 0; i < DefaultFungifyNumPositions; i++ {
		collected, _, err = s.clk.CollectIncentives(s.Ctx, defaultAddress, uint64(i+1))
		s.Require().Error(err)
		s.Require().Equal(sdk.Coins{}, collected)
	}
}

// TestFunctionalFungifyChargedPositions is a functional test that covers more complex scenarios related to fee/incentive claiming
// in the context of fungified positions.
//
// Testing strategy:
// 1. Create a pool with 6 positions in groups of 2 such that each group of 2 is adjacent to each other
// 2. Swap out all the USDC in the pool to generate spread rewards
// 3. Emit incentives (to left two positions)
// 4. Collect all spread rewards and incentives on cached ctx and make assertions
// 5. Fungify each set of positions
// 6. Collect all spread rewards and incentives on and make assertions
func (s *KeeperTestSuite) TestFunctionalFungifyChargedPositions() {
	s.SetupTest()
	s.Ctx = s.Ctx.WithBlockTime(defaultBlockTime)

	// Set incentives for pool to ensure accumulators work correctly
	testIncentiveRecord := types.IncentiveRecord{
		PoolId: 1,
		IncentiveRecordBody: types.IncentiveRecordBody{
			RemainingCoin: sdk.NewDecCoinFromDec(USDC, sdk.NewDec(1000000000000000000)),
			EmissionRate:  sdk.NewDec(1), // 1 per second
			StartTime:     defaultBlockTime,
		},
		MinUptime: time.Nanosecond,
	}

	// --- Set up positions ---

	// Create the relevant positions with these acccounts such that the left, middle, and right positions
	// are exactly adjacent to each other in terms of tick ranges.
	defaultPositionWidth := DefaultUpperTick - DefaultLowerTick

	// middleAddress refers to the owner of the position in the middle of the three we create in this test
	middleAddress := s.TestAccs[0]

	// Set up pool, default incentive records, and a single default position
	pool, middlePositionIds, _ := s.runFungifySetup(middleAddress, 2, DefaultFungifyFullChargeDuration, DefaultSpreadFactor, []types.IncentiveRecord{testIncentiveRecord})

	// Create two new addresses to hold a position to the left and right of the one we created above
	testAccs := apptesting.CreateRandomAccounts(2)
	leftAddress := testAccs[0]
	rightAddress := testAccs[1]

	// Set up left positions
	leftPositionLowerTick := DefaultLowerTick - defaultPositionWidth
	leftPositionUpperTick := DefaultLowerTick
	_, leftOne := s.SetupPosition(pool.GetId(), leftAddress, DefaultCoins, leftPositionLowerTick, leftPositionUpperTick, false)
	_, leftTwo := s.SetupPosition(pool.GetId(), leftAddress, DefaultCoins, leftPositionLowerTick, leftPositionUpperTick, false)

	// Set up right positions
	rightPositionLowerTick := DefaultUpperTick
	rightPositionUpperTick := DefaultUpperTick + defaultPositionWidth
	_, rightOne := s.SetupPosition(pool.GetId(), rightAddress, DefaultCoins, rightPositionLowerTick, rightPositionUpperTick, true)
	_, rightTwo := s.SetupPosition(pool.GetId(), rightAddress, DefaultCoins, rightPositionLowerTick, rightPositionUpperTick, true)

	// --- Set up large swap ---

	// Calculate input and output amounts for swap based on pool liquidity
	pool, err := s.clk.GetPoolById(s.Ctx, pool.GetId())
	s.Require().NoError(err)
	poolLiquidity := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())
	usdcSupply := poolLiquidity.FilterDenoms([]string{USDC})[0]
	usdcSupply = sdk.NewCoin(USDC, usdcSupply.Amount.Sub(sdk.NewInt(1)))
	ethFunded := sdk.NewCoins(sdk.NewCoin(ETH, poolLiquidity.AmountOf(ETH).MulRaw(2)))

	// --- Execute large swap ---

	s.TestAccs = apptesting.CreateRandomAccounts(5)
	s.FundAcc(s.TestAccs[4], ethFunded)
	coinIn, _, _, err := s.clk.SwapInAmtGivenOut(s.Ctx, s.TestAccs[4], pool, usdcSupply, ETH, DefaultSpreadFactor, types.MinSpotPrice)
	s.Require().NoError(err)

	// --- Set up expected spread rewards and incentives ---

	// Set up expected spread rewards
	expectedTotalSpreadReward := coinIn.Amount.ToDec().Mul(DefaultSpreadFactor).Ceil().TruncateInt()
	expectedTotalSpreadRewardCoins := sdk.NewCoins(sdk.NewCoin(coinIn.Denom, expectedTotalSpreadReward))

	// Set up expected incentives
	expectedIncentivesAmount := sdk.NewInt(int64(DefaultFungifyFullChargeDuration.Seconds()))
	expectedIncentivesCoins := sdk.NewCoins(sdk.NewCoin(USDC, expectedIncentivesAmount))
	s.FundAcc(pool.GetIncentivesAddress(), expectedIncentivesCoins)

	// --- Emit incentives ---

	// Increase block time by the fully charged duration
	// Note: claiming incentives should already trigger update incentives accumulators
	s.AddBlockTime(DefaultFungifyFullChargeDuration)

	// --- Assertions on non-fungified positions ---

	// We operate and claim on cached context so we can compare against behavior with fungified positions
	cacheCtx, _ := s.Ctx.CacheContext()
	allPositionIds := []uint64{leftOne, leftTwo, middlePositionIds[0], middlePositionIds[1], rightOne, rightTwo}
	positionOwners := []sdk.AccAddress{leftAddress, leftAddress, middleAddress, middleAddress, rightAddress, rightAddress}

	// Set up trackers for individual and total collected rewards
	collectedSpreadRewardsMap := make(map[uint64]sdk.Coins, len(allPositionIds))
	collectedIncentivesMap := make(map[uint64]sdk.Coins, len(allPositionIds))
	totalCollectedSpread := sdk.NewCoins()
	totalCollectedIncentives := sdk.NewCoins()

	for i, id := range allPositionIds {
		// Collect spread rewards and incentives on cached context
		collectedSpread, err := s.clk.CollectSpreadRewards(cacheCtx, positionOwners[i], id)
		s.Require().NoError(err)
		collectedIncentives, forfeited, err := s.clk.CollectIncentives(cacheCtx, positionOwners[i], id)
		s.Require().NoError(err)
		s.Require().True(forfeited.Empty())

		// Ensure positions that aren't touched don't collect any spread rewards or incentives
		if id == rightOne || id == rightTwo {
			s.Require().True(collectedSpread.Empty())
			s.Require().True(collectedIncentives.Empty())
		}

		// Middle positions collect no incentives either since we emit after the swap
		if id == middlePositionIds[0] || id == middlePositionIds[1] {
			s.Require().True(collectedIncentives.Empty())
		}

		// Track total amounts
		collectedSpreadRewardsMap[id] = collectedSpread
		collectedIncentivesMap[id] = collectedIncentives
		totalCollectedSpread = totalCollectedSpread.Add(collectedSpread...)
		totalCollectedIncentives = totalCollectedIncentives.Add(collectedIncentives...)
	}

	// Ensure that identical positions collected the same amounts
	s.Require().Equal(collectedSpreadRewardsMap[leftOne], collectedSpreadRewardsMap[leftTwo])
	s.Require().Equal(collectedSpreadRewardsMap[middlePositionIds[0]], collectedSpreadRewardsMap[middlePositionIds[1]])
	s.Require().Equal(collectedSpreadRewardsMap[rightOne], collectedSpreadRewardsMap[rightTwo])

	s.Require().Equal(collectedIncentivesMap[leftOne], collectedIncentivesMap[leftTwo])
	s.Require().Equal(collectedIncentivesMap[middlePositionIds[0]], collectedIncentivesMap[middlePositionIds[1]])
	s.Require().Equal(collectedIncentivesMap[rightOne], collectedIncentivesMap[rightTwo])

	// Sanity check that majority of spread rewards went to the positions that provided majority of liquidity
	s.Require().True(collectedSpreadRewardsMap[leftOne].IsAllGT(collectedSpreadRewardsMap[middlePositionIds[0]]))

	// Ensure that the total spread rewards collected is correct
	roundingTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: sdk.NewDec(int64(len(allPositionIds))),
		RoundingDir:       osmomath.RoundDown,
	}
	for _, spreadRewardCoin := range expectedTotalSpreadRewardCoins {
		denom := spreadRewardCoin.Denom
		s.Require().Equal(0, roundingTolerance.Compare(expectedTotalSpreadRewardCoins.AmountOf(denom), totalCollectedSpread.AmountOf(denom)))
	}

	// Ensure that the total incentives collected is correct
	for _, incentiveCoin := range expectedIncentivesCoins {
		denom := incentiveCoin.Denom
		s.Require().Equal(0, roundingTolerance.Compare(expectedIncentivesCoins.AmountOf(denom), totalCollectedIncentives.AmountOf(denom)), "expected: %s, got: %s", expectedIncentivesCoins.AmountOf(denom), totalCollectedIncentives.AmountOf(denom))
	}

	// --- System under test: Fungify positions ---

	fungifiedLeft, err := s.clk.FungifyChargedPosition(s.Ctx, leftAddress, []uint64{leftOne, leftTwo})
	s.Require().NoError(err)
	fungifiedMiddle, err := s.clk.FungifyChargedPosition(s.Ctx, middleAddress, middlePositionIds)
	s.Require().NoError(err)
	fungifiedRight, err := s.clk.FungifyChargedPosition(s.Ctx, rightAddress, []uint64{rightOne, rightTwo})

	// --- Spread reward assertions on fungified positions ---

	// Set up variables to represent loss due to truncation since expected values
	// are derived from individual position claims (each of which truncate)
	truncatedETHCoins := sdk.NewCoins(sdk.NewCoin(ETH, roundingError))
	truncatedUSDCCoins := sdk.NewCoins(sdk.NewCoin(USDC, roundingError))

	// Left position spread reward assertion
	fungifiedLeftSpread, err := s.clk.CollectSpreadRewards(s.Ctx, leftAddress, fungifiedLeft)
	s.Require().NoError(err)
	s.Require().Equal(collectedSpreadRewardsMap[leftOne].Add(collectedSpreadRewardsMap[leftTwo]...).Add(truncatedETHCoins...), fungifiedLeftSpread)

	// Middle position spread reward assertion
	fungifiedMiddleSpread, err := s.clk.CollectSpreadRewards(s.Ctx, middleAddress, fungifiedMiddle)
	s.Require().NoError(err)
	s.Require().Equal(collectedSpreadRewardsMap[middlePositionIds[0]].Add(collectedSpreadRewardsMap[middlePositionIds[1]]...).Add(truncatedETHCoins...), fungifiedMiddleSpread)

	// Right position spread reward assertion
	fungifiedRightSpread, err := s.clk.CollectSpreadRewards(s.Ctx, rightAddress, fungifiedRight)
	s.Require().NoError(err)
	s.Require().True(fungifiedRightSpread.Empty())

	// --- Incentive assertions on fungified positions ---

	// Left position incentives assertion
	fungifiedLeftIncentives, forfeited, err := s.clk.CollectIncentives(s.Ctx, leftAddress, fungifiedLeft)
	s.Require().NoError(err)
	s.Require().True(forfeited.Empty())
	s.Require().Equal(collectedIncentivesMap[leftOne].Add(collectedIncentivesMap[leftTwo]...).Add(truncatedUSDCCoins...), fungifiedLeftIncentives)

	// Middle position incentives assertion
	fungifiedMiddleIncentives, forfeited, err := s.clk.CollectIncentives(s.Ctx, middleAddress, fungifiedMiddle)
	s.Require().NoError(err)
	s.Require().True(forfeited.Empty())
	s.Require().True(fungifiedMiddleIncentives.Empty())

	// Right position incentives assertion
	fungifiedRightIncentives, forfeited, err := s.clk.CollectIncentives(s.Ctx, rightAddress, fungifiedRight)
	s.Require().NoError(err)
	s.Require().True(forfeited.Empty())
	s.Require().True(fungifiedRightIncentives.Empty())
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
				positionId, _, _, liquidity, concentratedLockId, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, clPool.GetId(), defaultAddress, test.coinsForPosition, test.remainingLockDuration)
			} else if test.isUnlocking {
				positionId, _, _, liquidity, concentratedLockId, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionUnlocking(s.Ctx, clPool.GetId(), defaultAddress, test.coinsForPosition, test.remainingLockDuration)
			} else {
				positionId, _, _, liquidity, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool.GetId(), defaultAddress, test.coinsForPosition)
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

func (s *KeeperTestSuite) TestMintSharesAndLock() {
	var (
		defaultPositionCoins = sdk.NewCoins(DefaultCoin0, DefaultCoin1)
		defaultAddress       = s.TestAccs[0]
	)

	tests := []struct {
		name                    string
		owner                   sdk.AccAddress
		remainingLockDuration   time.Duration
		createFullRangePosition bool
		lowerTick               int64
		upperTick               int64
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
			name:                    "err: lower tick is not min tick",
			owner:                   defaultAddress,
			createFullRangePosition: false,
			lowerTick:               DefaultLowerTick,
			upperTick:               types.MaxTick,
			remainingLockDuration:   24 * time.Hour,
			expectedErr:             types.PositionNotFullRangeError{PositionId: 1, LowerTick: DefaultLowerTick, UpperTick: types.MaxTick},
		},
		{
			name:                    "err: upper tick is not max tick",
			owner:                   defaultAddress,
			createFullRangePosition: false,
			lowerTick:               types.MinInitializedTick,
			upperTick:               DefaultUpperTick,
			remainingLockDuration:   24 * time.Hour,
			expectedErr:             types.PositionNotFullRangeError{PositionId: 1, LowerTick: types.MinInitializedTick, UpperTick: DefaultUpperTick},
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
			s.FundAcc(test.owner, DefaultCoins)

			// Create a position
			positionId := uint64(0)
			liquidity := sdk.ZeroDec()
			if test.createFullRangePosition {
				var err error
				positionId, _, _, liquidity, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool.GetId(), test.owner, DefaultCoins)
				s.Require().NoError(err)
			} else {
				var err error
				positionId, _, _, liquidity, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, clPool.GetId(), test.owner, defaultPositionCoins, sdk.ZeroInt(), sdk.ZeroInt(), test.lowerTick, test.upperTick)
				s.Require().NoError(err)
			}

			lockupModuleAccountBalancePre := s.App.LockupKeeper.GetModuleBalance(s.Ctx)

			// System under test
			concentratedLockId, underlyingLiquidityTokenized, err := s.App.ConcentratedLiquidityKeeper.MintSharesAndLock(s.Ctx, clPool.GetId(), positionId, test.owner, test.remainingLockDuration)
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
		name                                       string
		createPosition                             func(s *KeeperTestSuite) (uint64, uint64)
		expectedHasActiveLock                      bool
		expectedHasActiveLockAfterTimeUpdate       bool
		expectedLockError                          bool
		expectedPositionHasActiveUnderlyingLockErr bool
		expectedPositionLockID                     uint64
		expectedGetPositionLockIdErr               bool
	}

	tests := []testParams{
		{
			name: "position with lock locked",
			createPosition: func(s *KeeperTestSuite) (uint64, uint64) {
				s.FundAcc(owner, defaultPositionCoins)
				positionID, _, _, _, concentratedLockID, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(
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
				positionID, _, _, _, concentratedLockID, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionUnlocking(
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
				positionID, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(
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
		{
			name: "invalid position, invalid lock: should return false",
			createPosition: func(s *KeeperTestSuite) (uint64, uint64) {
				return 100, 0
			},
			expectedHasActiveLock:                      false,
			expectedHasActiveLockAfterTimeUpdate:       false,
			expectedLockError:                          true,
			expectedPositionHasActiveUnderlyingLockErr: true,
			expectedPositionLockID:                     0,
			expectedGetPositionLockIdErr:               true,
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
		name                                             string
		createPosition                                   func(s *KeeperTestSuite) (uint64, uint64)
		expectedHasActiveLock                            bool
		expectedHasActiveLockAfterTimeUpdate             bool
		expectedLockError                                bool
		expectedPositionLockID                           uint64
		expectedPositionLockIDAfterTimeUpdate            uint64
		expectedGetPositionLockIdErr                     bool
		expectedGetPositionLockIdErrAfterTimeUpdate      bool
		expectedPositionHasActiveUnderlyingLockAndUpdate bool
	}

	tests := []testParams{
		{
			name: "position with lock locked",
			createPosition: func(s *KeeperTestSuite) (uint64, uint64) {
				s.FundAcc(owner, defaultPositionCoins)
				positionID, _, _, _, concentratedLockID, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(
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
				positionID, _, _, _, concentratedLockID, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionUnlocking(
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
				positionID, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(
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
		{
			name: "invalid position without lock ",
			// we return invalid lock id with no-op to trigger error
			createPosition: func(s *KeeperTestSuite) (uint64, uint64) {
				return 10, 0
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
			if tc.expectedPositionHasActiveUnderlyingLockAndUpdate {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
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
	positionId, _, _, _, concentratedLockId, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionUnlocking(s.Ctx, clPool.GetId(), owner, defaultPositionCoins, remainingLockDuration)
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
	positionId, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool.GetId(), owner, defaultPositionCoins)
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
	s.App.ConcentratedLiquidityKeeper.RemovePositionIdForLockId(s.Ctx, positionId, retrievedLockId)

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

		// Retrieve the position from the store via owner/poolId/positionId and compare to expected value (true).
		key = types.KeyAddressPoolIdPositionId(tc.owner, tc.poolId, tc.positionId)
		valueBytes := store.Get(key)
		s.Require().Equal([]byte{1}, valueBytes)

		// Retrieve the position from the store via poolId/positionId and compare to expected value (true).
		key = types.KeyPoolPositionPositionId(tc.poolId, tc.positionId)
		valueBytes = store.Get(key)
		s.Require().Equal([]byte{1}, valueBytes)

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
		_, _, _, liquidity, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool.GetId(), owner, tc.positionCoins)
		s.Require().NoError(err)
		actualFullRangeLiquidity = actualFullRangeLiquidity.Add(liquidity)

		clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPoolId)
		s.Require().NoError(err)

		// Get the full range liquidity for the pool.
		expectedFullRangeLiquidity, err := s.App.ConcentratedLiquidityKeeper.GetFullRangeLiquidityInPool(s.Ctx, clPoolId)
		s.Require().NoError(err)
		s.Require().Equal(expectedFullRangeLiquidity, actualFullRangeLiquidity)

		// Create a new position that overlaps with the min tick, but is not full range and therefore should not count towards the full range liquidity.
		s.FundAcc(owner, tc.positionCoins)
		_, _, _, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, clPoolId, owner, DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), tc.lowerTick, tc.upperTick)
		s.Require().NoError(err)

		clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPoolId)
		s.Require().NoError(err)

		// Test updating the full range liquidity.
		err = s.App.ConcentratedLiquidityKeeper.UpdateFullRangeLiquidityInPool(s.Ctx, clPoolId, tc.updateLiquidity)
		s.Require().NoError(err)
		actualFullRangeLiquidity, err = s.App.ConcentratedLiquidityKeeper.GetFullRangeLiquidityInPool(s.Ctx, clPoolId)
		s.Require().NoError(err)
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

	positionOne, err := clKeeper.GetAllPositionIdsForPoolId(s.Ctx, types.PositionPrefix, clPoolOne.GetId())
	s.Require().NoError(err)

	positionTwo, err := clKeeper.GetAllPositionIdsForPoolId(s.Ctx, types.PositionPrefix, clPooltwo.GetId())
	s.Require().NoError(err)

	s.Require().Equal(expectedPositionOneIds, positionOne)
	s.Require().Equal(expectedPositionTwoIds, positionTwo)
}

func (s *KeeperTestSuite) TestCreateFullRangePositionLocked() {
	invalidCoinsAmount := sdk.NewCoins(DefaultCoin0)
	invalidCoin0Denom := sdk.NewCoins(sdk.NewCoin("invalidDenom", sdk.NewInt(1000000000000000000)), DefaultCoin1)
	invalidCoin1Denom := sdk.NewCoins(DefaultCoin0, sdk.NewCoin("invalidDenom", sdk.NewInt(1000000000000000000)))
	zeroCoins := sdk.NewCoins()

	defaultRemainingLockDuration := s.App.StakingKeeper.GetParams(s.Ctx).UnbondingTime

	tests := []struct {
		name                  string
		remainingLockDuration time.Duration
		coinsForPosition      sdk.Coins
		expectedErr           error
	}{
		{
			name:                  "valid test",
			remainingLockDuration: defaultRemainingLockDuration,
			coinsForPosition:      DefaultCoins,
		},
		{
			name:                  "invalid coin0 denom",
			remainingLockDuration: defaultRemainingLockDuration,
			coinsForPosition:      invalidCoin0Denom,
			expectedErr:           types.Amount0IsNegativeError{Amount0: sdk.ZeroInt()},
		},
		{
			name:                  "invalid coin1 denom",
			remainingLockDuration: defaultRemainingLockDuration,
			coinsForPosition:      invalidCoin1Denom,
			expectedErr:           types.Amount1IsNegativeError{Amount1: sdk.ZeroInt()},
		},
		{
			name:                  "invalid coins amount",
			remainingLockDuration: defaultRemainingLockDuration,
			coinsForPosition:      invalidCoinsAmount,
			expectedErr:           types.NumCoinsError{NumCoins: len(invalidCoinsAmount)},
		},
		{
			name:                  "edge: both coins amounts' are zero",
			remainingLockDuration: defaultRemainingLockDuration,
			coinsForPosition:      zeroCoins,
			expectedErr:           types.NumCoinsError{NumCoins: 0},
		},
	}

	for _, test := range tests {
		test := test
		s.Run(test.name, func() {
			// Init suite for each test.
			s.SetupTest()

			// Create a default CL pool
			clPool := s.PrepareConcentratedPoolWithCoins(ETH, USDC)

			// Fund the owner account
			defaultAddress := s.TestAccs[0]
			s.FundAcc(defaultAddress, test.coinsForPosition)

			// System under test
			positionId, _, _, liquidity, concentratedLockId, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, clPool.GetId(), defaultAddress, test.coinsForPosition, test.remainingLockDuration)

			if test.expectedErr != nil {
				s.Require().ErrorContains(err, test.expectedErr.Error())
				return
			}

			s.Require().NoError(err)

			// Check position
			position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionId)
			s.Require().NoError(err)
			s.Require().Equal(s.Ctx.BlockTime(), position.JoinTime)
			s.Require().Equal(types.MaxTick, position.UpperTick)
			s.Require().Equal(types.MinInitializedTick, position.LowerTick)
			s.Require().Equal(liquidity, position.Liquidity)

			// Check locked
			concentratedLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, concentratedLockId)
			s.Require().NoError(err)

			s.Require().Equal(concentratedLock.Coins[0].Amount.String(), liquidity.TruncateInt().String())
			s.Require().False(concentratedLock.IsUnlocking())
		})
	}
}

// TestTickRoundingEdgeCase tests an edge case where incorrect tick rounding would cause LP funds to be drained.
func (s *KeeperTestSuite) TestTickRoundingEdgeCase() {
	s.SetupTest()
	pool := s.PrepareConcentratedPool()

	testAccs := apptesting.CreateRandomAccounts(3)
	firstPositionAddr := testAccs[0]
	secondPositionAddr := testAccs[1]

	// Create two identical positions with the initial assets set such that both positions are fully in one asset
	firstPositionAssets := sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(9823358512)), sdk.NewCoin(USDC, sdk.NewInt(8985893232)))
	firstPosLiq, firstPosId := s.SetupPosition(pool.GetId(), firstPositionAddr, firstPositionAssets, -68720000, -68710000, true)
	secondPositionAssets := sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(9823358512)), sdk.NewCoin(USDC, sdk.NewInt(8985893232)))
	secondPosLiq, secondPosId := s.SetupPosition(pool.GetId(), secondPositionAddr, secondPositionAssets, -68720000, -68710000, true)

	// Execute a swap that brings the price close enough to the edge of a tick to trigger bankers rounding
	swapAddr := testAccs[2]
	desiredTokenOut := sdk.NewCoin(USDC, sdk.NewInt(10000))
	s.FundAcc(swapAddr, sdk.NewCoins(sdk.NewCoin(ETH, sdk.NewInt(1000000000000000000))))
	_, _, _, err := s.clk.SwapInAmtGivenOut(s.Ctx, swapAddr, pool, desiredTokenOut, ETH, sdk.ZeroDec(), sdk.ZeroDec())
	s.Require().NoError(err)

	// Both positions should be able to withdraw successfully
	_, _, err = s.clk.WithdrawPosition(s.Ctx, firstPositionAddr, firstPosId, firstPosLiq)
	s.Require().NoError(err)
	_, _, err = s.clk.WithdrawPosition(s.Ctx, secondPositionAddr, secondPosId, secondPosLiq)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestMultipleRanges() {
	tests := map[string]struct {
		tickRanges      [][]int64
		rangeTestParams RangeTestParams
	}{
		"one range, default params": {
			tickRanges: [][]int64{
				{0, 10000},
			},
			rangeTestParams: DefaultRangeTestParams,
		},
		"one min width range": {
			tickRanges: [][]int64{
				{0, 100},
			},
			rangeTestParams: withTickSpacing(DefaultRangeTestParams, DefaultTickSpacing),
		},
		"two adjacent ranges": {
			tickRanges: [][]int64{
				{-10000, 10000},
				{10000, 20000},
			},
			rangeTestParams: DefaultRangeTestParams,
		},
		"two adjacent ranges with current tick smaller than both": {
			tickRanges: [][]int64{
				{-10000, 10000},
				{10000, 20000},
			},
			rangeTestParams: withCurrentTick(DefaultRangeTestParams, -20000),
		},
		"two adjacent ranges with current tick larger than both": {
			tickRanges: [][]int64{
				{-10000, 10000},
				{10000, 20000},
			},
			rangeTestParams: withCurrentTick(DefaultRangeTestParams, 30000),
		},
		"two adjacent ranges with current tick exactly on lower bound": {
			tickRanges: [][]int64{
				{-10000, 10000},
				{10000, 20000},
			},
			rangeTestParams: withCurrentTick(DefaultRangeTestParams, -10000),
		},
		"two adjacent ranges with current tick exactly between both": {
			tickRanges: [][]int64{
				{-10000, 10000},
				{10000, 20000},
			},
			rangeTestParams: withCurrentTick(DefaultRangeTestParams, 10000),
		},
		"two adjacent ranges with current tick exactly on upper bound": {
			tickRanges: [][]int64{
				{-10000, 10000},
				{10000, 20000},
			},
			rangeTestParams: withCurrentTick(DefaultRangeTestParams, 20000),
		},
		"two non-adjacent ranges": {
			tickRanges: [][]int64{
				{-10000, 10000},
				{20000, 30000},
			},
			rangeTestParams: DefaultRangeTestParams,
		},
		"two ranges with one tick gap in between, which is equal to current tick": {
			tickRanges: [][]int64{
				{799221, 799997},
				{799997 + 2, 812343},
			},
			rangeTestParams: withCurrentTick(DefaultRangeTestParams, 799997+1),
		},
		"one range on large tick": {
			tickRanges: [][]int64{
				{207000000, 207000000 + 100},
			},
			rangeTestParams: withTickSpacing(DefaultRangeTestParams, DefaultTickSpacing),
		},
		"one position adjacent to left of current tick (no swaps)": {
			tickRanges: [][]int64{
				{-1, 0},
			},
			rangeTestParams: RangeTestParamsNoFuzzNoSwap,
		},
		"one position on left of current tick with gap (no swaps)": {
			tickRanges: [][]int64{
				{-2, -1},
			},
			rangeTestParams: RangeTestParamsNoFuzzNoSwap,
		},
		"one position adjacent to right of current tick (no swaps)": {
			tickRanges: [][]int64{
				{0, 1},
			},
			rangeTestParams: RangeTestParamsNoFuzzNoSwap,
		},
		"one position on right of current tick with gap (no swaps)": {
			tickRanges: [][]int64{
				{1, 2},
			},
			rangeTestParams: RangeTestParamsNoFuzzNoSwap,
		},
		"one range on small tick": {
			tickRanges: [][]int64{
				{-107000000, -107000000 + 100},
			},
			rangeTestParams: withDoubleFundedLP(DefaultRangeTestParams),
		},
		"one range on min tick": {
			tickRanges: [][]int64{
				{types.MinInitializedTick, types.MinInitializedTick + 100},
			},
			rangeTestParams: withDoubleFundedLP(DefaultRangeTestParams),
		},
		"initial current tick equal to min initialized tick": {
			tickRanges: [][]int64{
				{0, 1},
			},
			rangeTestParams: withCurrentTick(DefaultRangeTestParams, types.MinInitializedTick),
		},
		"three overlapping ranges with no swaps, current tick in one": {
			tickRanges: [][]int64{
				{-10000, 10000},
				{0, 20000},
				{-7300, 12345},
			},
			rangeTestParams: withNoSwap(withCurrentTick(DefaultRangeTestParams, -9000)),
		},
		"three overlapping ranges with no swaps, current tick in two of three": {
			tickRanges: [][]int64{
				{-10000, 10000},
				{0, 20000},
				{-7300, 12345},
			},
			rangeTestParams: withNoSwap(withCurrentTick(DefaultRangeTestParams, -7231)),
		},
		"three overlapping ranges with no swaps, current tick in all three": {
			tickRanges: [][]int64{
				{-10000, 10000},
				{0, 20000},
				{-7300, 12345},
			},
			rangeTestParams: withNoSwap(withCurrentTick(DefaultRangeTestParams, 109)),
		},
		/* TODO: uncomment when infinite loop bug is fixed
		"one range on max tick": {
			tickRanges: [][]int64{
				{types.MaxTick - 100, types.MaxTick},
			},
			rangeTestParams: withTickSpacing(DefaultRangeTestParams, DefaultTickSpacing),
		},
		"initial current tick equal to max tick": {
			tickRanges: [][]int64{
				{0, 1},
			},
			rangeTestParams: withCurrentTick(withTickSpacing(DefaultRangeTestParams, uint64(1)), types.MaxTick),
		},
		*/
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.runMultiplePositionRanges(tc.tickRanges, tc.rangeTestParams)
		})
	}
}

// This test reproduces the panic stemming from the negative range accumulator whenever
// lower tick accumulator is greater than upper tick accumulator and current tick is above the position's range.
func (s *KeeperTestSuite) TestNegativeTickRange_SpreadFactor() {
	s.SetupTest()
	// Initialize pool with non-zero spread factor.
	spreadFactor := sdk.NewDecWithPrec(3, 3)
	pool := s.PrepareCustomConcentratedPool(s.TestAccs[0], DefaultCoin0.Denom, DefaultCoin1.Denom, 1, spreadFactor)
	poolId := pool.GetId()

	// Create full range position
	s.FundAcc(s.TestAccs[0], DefaultCoins)
	s.CreateFullRangePosition(pool, DefaultCoins)

	// Initialize position at a higher range
	s.FundAcc(s.TestAccs[0], DefaultCoins)
	_, _, _, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, poolId, s.TestAccs[0], DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), DefaultCurrTick+50, DefaultCurrTick+100)
	s.Require().NoError(err)

	// Refetch pool
	pool, err = s.clk.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	// Swap to approximately tick 50 below current
	amountZeroIn := math.CalcAmount0Delta(osmomath.BigDecFromSDKDec(pool.GetLiquidity()), pool.GetCurrentSqrtPrice(), osmomath.BigDecFromSDKDec(s.tickToSqrtPrice(DefaultCurrTick-50)), true)
	coinZeroIn := sdk.NewCoin(pool.GetToken0(), amountZeroIn.SDKDec().TruncateInt())

	s.FundAcc(s.TestAccs[0], sdk.NewCoins(coinZeroIn))
	_, err = s.clk.SwapExactAmountIn(s.Ctx, s.TestAccs[0], pool, coinZeroIn, pool.GetToken1(), sdk.ZeroInt(), spreadFactor)
	s.Require().NoError(err)

	// Refetch pool
	pool, err = s.clk.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	// Swap to approximately DefaultCurrTick + 150
	amountOneIn := math.CalcAmount1Delta(osmomath.BigDecFromSDKDec(pool.GetLiquidity()), pool.GetCurrentSqrtPrice(), osmomath.BigDecFromSDKDec(s.tickToSqrtPrice(DefaultCurrTick+150)), true)
	coinOneIn := sdk.NewCoin(pool.GetToken1(), amountOneIn.SDKDec().TruncateInt())

	s.FundAcc(s.TestAccs[0], sdk.NewCoins(coinOneIn))
	_, err = s.clk.SwapExactAmountIn(s.Ctx, s.TestAccs[0], pool, coinOneIn, pool.GetToken0(), sdk.ZeroInt(), spreadFactor)
	s.Require().NoError(err)

	// This currently panics due to the lack of support for negative range accumulators.
	// We initialized the lower tick's accumulator (DefaultCurrTick - 25) to be greater than the upper tick's accumulator (DefaultCurrTick + 50)
	// Whenever the current tick is above the position's range, we compute in range accumulator as upper tick accumulator - lower tick accumulator
	// In this case, it ends up being negative, which is not supported.
	// The fix is to be implmeneted in: https://github.com/osmosis-labs/osmosis/issues/5854
	osmoassert.ConditionalPanic(s.T(), true, func() {
		s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, poolId, s.TestAccs[0], DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), DefaultCurrTick-25, DefaultCurrTick+50)
	})
}
