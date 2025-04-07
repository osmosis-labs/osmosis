package concentrated_liquidity_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	cl "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

const (
	DefaultFungifyNumPositions       = 3
	DefaultFungifyFullChargeDuration = 24 * time.Hour
)

var (
	DefaultIncentiveRecords = []types.IncentiveRecord{incentiveRecordOne, incentiveRecordTwo, incentiveRecordThree, incentiveRecordFour}
	DefaultBlockTime        = time.Unix(1, 1).UTC()
	DefaultSpreadFactor     = osmomath.NewDecWithPrec(2, 3)
)

// AssertPositionsDoNotExist checks that the positions with the given IDs do not exist on uptime accumulators.
func (s *KeeperTestSuite) AssertPositionsDoNotExist(positionIds []uint64) {
	uptimeAccumulators, err := s.Clk.GetUptimeAccumulators(s.Ctx, defaultPoolId)
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
	uptimeAccumulators, err := s.Clk.GetUptimeAccumulators(s.Ctx, defaultPoolId)
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
func (s *KeeperTestSuite) ExecuteAndValidateSuccessfulIncentiveClaim(positionId uint64, expectedRewards sdk.Coins, expectedForfeited sdk.Coins, poolId uint64) {
	// Initial claim and assertion
	claimedRewards, totalForfeitedRewards, scaledForfeitedRewardsByUptime, err := s.Clk.PrepareClaimAllIncentivesForPosition(s.Ctx, positionId)
	s.Require().NoError(err)

	s.Require().Equal(expectedRewards, claimedRewards)
	s.Require().Equal(expectedForfeited, totalForfeitedRewards)
	s.checkForfeitedCoinsByUptime(totalForfeitedRewards, scaledForfeitedRewardsByUptime)

	// Sanity check that cannot claim again.
	claimedRewards, _, _, err = s.Clk.PrepareClaimAllIncentivesForPosition(s.Ctx, positionId)
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
		liquidityDelta osmomath.Dec
	}

	tests := []struct {
		name                 string
		param                param
		positionExists       bool
		timeElapsedSinceInit time.Duration
		incentiveRecords     []types.IncentiveRecord
		expectedLiquidity    osmomath.Dec
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
			preexistingLiquidity := osmomath.ZeroDec()
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

			timeElapsedSec := osmomath.NewDec(int64(test.timeElapsedSinceInit)).Quo(osmomath.NewDec(10e8))
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
						expectedGrowthCurAccum, _, err := cl.CalcAccruedIncentivesForAccum(s.Ctx, uptime, test.param.liquidityDelta, timeElapsedSec, expectedIncentiveRecords, cl.PerUnitLiqScalingFactor)
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
		expectedPositionLiquidity osmomath.Dec
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
				s.Require().Equal(osmomath.Dec{}, position.Liquidity)
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
				s.Require().Equal(osmomath.Dec{}, positionLiquidity)

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
			_, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, 1, s.TestAccs[0], DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
			s.Require().NoError(err)

			// create a position from the test case
			s.FundAcc(s.TestAccs[1], sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0), sdk.NewCoin(USDC, DefaultAmt1)))
			positionData, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, tc.position.PoolId, s.TestAccs[1], DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), tc.position.LowerTick, tc.position.UpperTick)
			s.Require().NoError(err)
			tc.position.Liquidity = positionData.Liquidity

			if tc.isZeroLiquidity {
				// set the position liquidity to zero
				tc.position.Liquidity = osmomath.ZeroDec()
				positionData.Amount0 = osmomath.ZeroInt()
				positionData.Amount1 = osmomath.ZeroInt()
			}

			// calculate underlying assets from the position
			clPool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, tc.position.PoolId)
			s.Require().NoError(err)
			calculatedCoin0, calculatedCoin1, err := cl.CalculateUnderlyingAssetsFromPosition(s.Ctx, tc.position, clPool)

			s.Require().NoError(err)
			s.Require().Equal(calculatedCoin0.String(), sdk.NewCoin(clPool.GetToken0(), positionData.Amount0).String())
			s.Require().Equal(calculatedCoin1.String(), sdk.NewCoin(clPool.GetToken1(), positionData.Amount1).String())
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
				s.SetupPosition(pos.PoolId, sdk.MustAccAddressFromBech32(pos.Address), DefaultCoins, pos.LowerTick, pos.UpperTick, false)
			}

			// System under test
			actualResult, err := s.App.ConcentratedLiquidityKeeper.HasAnyPositionForPool(s.Ctx, test.poolId)

			s.Require().NoError(err)
			s.Require().Equal(test.expectedResult, actualResult)
		})
	}
}

func (s *KeeperTestSuite) TestCreateFullRangePosition() {
	var (
		positionData       cltypes.CreateFullRangePositionData
		concentratedLockId uint64
		err                error
	)
	invalidCoinsAmount := sdk.NewCoins(DefaultCoin0)
	invalidCoin0Denom := sdk.NewCoins(sdk.NewCoin("invalidDenom", osmomath.NewInt(1000000000000000000)), DefaultCoin1)
	invalidCoin1Denom := sdk.NewCoins(DefaultCoin0, sdk.NewCoin("invalidDenom", osmomath.NewInt(1000000000000000000)))

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
			expectedErr:      types.Amount0IsNegativeError{Amount0: osmomath.ZeroInt()},
		},
		{
			name:             "err: wrong denom 1 provided for a full range",
			coinsForPosition: invalidCoin1Denom,
			expectedErr:      types.Amount1IsNegativeError{Amount1: osmomath.ZeroInt()},
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
				positionData, concentratedLockId, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, clPool.GetId(), defaultAddress, test.coinsForPosition, test.remainingLockDuration)
			} else if test.isUnlocking {
				positionData, concentratedLockId, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionUnlocking(s.Ctx, clPool.GetId(), defaultAddress, test.coinsForPosition, test.remainingLockDuration)
			} else {
				positionData, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool.GetId(), defaultAddress, test.coinsForPosition)
			}

			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, test.expectedErr.Error())
				return
			}

			s.Require().NoError(err)

			// Check position
			_, err = s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionData.ID)
			s.Require().NoError(err)

			// Check lock
			if test.isLocked || test.isUnlocking {
				concentratedLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, concentratedLockId)
				s.Require().NoError(err)
				s.Require().Equal(positionData.Liquidity.TruncateInt().String(), concentratedLock.Coins[0].Amount.String())
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
			liquidity := osmomath.ZeroDec()
			if test.createFullRangePosition {
				var err error
				positionData, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool.GetId(), test.owner, DefaultCoins)
				s.Require().NoError(err)
				positionId = positionData.ID
				liquidity = positionData.Liquidity
			} else {
				var err error
				positionData, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, clPool.GetId(), test.owner, defaultPositionCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), test.lowerTick, test.upperTick)
				s.Require().NoError(err)
				positionId = positionData.ID
				liquidity = positionData.Liquidity
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
			s.Require().Equal(underlyingLiquidityTokenized[0].String(), lockupModuleAccountBalancePost.Sub(lockupModuleAccountBalancePre...).String())

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
				positionData, concentratedLockID, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(
					s.Ctx, clPool.GetId(), owner, defaultPositionCoins, defaultLockDuration)
				s.Require().NoError(err)
				return positionData.ID, concentratedLockID
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
				positionData, concentratedLockID, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionUnlocking(
					s.Ctx, clPool.GetId(), owner, defaultPositionCoins, defaultLockDuration)
				s.Require().NoError(err)
				return positionData.ID, concentratedLockID
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
				positionData, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(
					s.Ctx, clPool.GetId(), owner, defaultPositionCoins)
				s.Require().NoError(err)
				return positionData.ID, 0
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
				positionData, concentratedLockID, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(
					s.Ctx, clPool.GetId(), owner, defaultPositionCoins, defaultLockDuration)
				s.Require().NoError(err)
				return positionData.ID, concentratedLockID
			},
			expectedHasActiveLock:                       true, // lock starts as active
			expectedHasActiveLockAfterTimeUpdate:        true, // since lock is locked, it remains active after time update
			expectedLockError:                           false,
			expectedPositionLockID:                      1,
			expectedPositionLockIDAfterTimeUpdate:       1, // since it stays locked, the mutative method won't change the underlying lock ID
			expectedGetPositionLockIdErr:                false,
			expectedGetPositionLockIdErrAfterTimeUpdate: false,
		},
		{
			name: "position with lock unlocking",
			createPosition: func(s *KeeperTestSuite) (uint64, uint64) {
				s.FundAcc(owner, defaultPositionCoins)
				positionData, concentratedLockID, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionUnlocking(
					s.Ctx, clPool.GetId(), owner, defaultPositionCoins, defaultLockDuration)
				s.Require().NoError(err)
				return positionData.ID, concentratedLockID
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
				positionData, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(
					s.Ctx, clPool.GetId(), owner, defaultPositionCoins)
				s.Require().NoError(err)
				return positionData.ID, 0
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
	positionData, concentratedLockId, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionUnlocking(s.Ctx, clPool.GetId(), owner, defaultPositionCoins, remainingLockDuration)
	s.Require().NoError(err)

	// We should be able to retrieve the lockId from the positionId now
	retrievedLockId, err := s.App.ConcentratedLiquidityKeeper.GetLockIdFromPositionId(s.Ctx, positionData.ID)
	s.Require().NoError(err)
	s.Require().Equal(concentratedLockId, retrievedLockId)

	// Check if lock has position in state
	retrievedPositionId, err := s.App.ConcentratedLiquidityKeeper.GetPositionIdToLockId(s.Ctx, retrievedLockId)
	s.Require().NoError(err)
	s.Require().Equal(positionData.ID, retrievedPositionId)

	// Create a position without a lock
	positionData, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool.GetId(), owner, defaultPositionCoins)
	s.Require().Error(err)

	// Check if position has lock in state, should not
	retrievedLockId, err = s.App.ConcentratedLiquidityKeeper.GetLockIdFromPositionId(s.Ctx, positionData.ID)
	s.Require().Error(err)
	s.Require().Equal(uint64(0), retrievedLockId)

	// Set the position to have a lockId (despite it not actually having a lock)
	s.App.ConcentratedLiquidityKeeper.SetPositionIdToLock(s.Ctx, positionData.ID, concentratedLockId)

	// Check if position has lock in state, it should now
	retrievedLockId, err = s.App.ConcentratedLiquidityKeeper.GetLockIdFromPositionId(s.Ctx, positionData.ID)
	s.Require().NoError(err)
	s.Require().Equal(concentratedLockId, retrievedLockId)

	// Check if lock has position in state
	retrievedPositionId, err = s.App.ConcentratedLiquidityKeeper.GetPositionIdToLockId(s.Ctx, retrievedLockId)
	s.Require().NoError(err)
	s.Require().Equal(positionData.ID, retrievedPositionId)

	// Remove the lockId from the position
	s.App.ConcentratedLiquidityKeeper.RemovePositionIdForLockId(s.Ctx, positionData.ID, retrievedLockId)

	// Check if position has lock in state, should not
	retrievedLockId, err = s.App.ConcentratedLiquidityKeeper.GetLockIdFromPositionId(s.Ctx, positionData.ID)
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
		liquidity        osmomath.Dec
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
		updateLiquidity      osmomath.Dec
	}{
		{
			name:            "full range + position overlapping min tick. update liquidity upwards",
			positionCoins:   sdk.NewCoins(DefaultCoin0, DefaultCoin1),
			lowerTick:       DefaultMinTick,
			upperTick:       DefaultUpperTick, // max tick doesn't overlap, should not count towards full range liquidity
			updateLiquidity: hundredDec,
		},
		{
			name:            "full range + position overlapping max tick. update liquidity downwards",
			positionCoins:   sdk.NewCoins(DefaultCoin0, DefaultCoin1),
			lowerTick:       DefaultLowerTick, // min tick doesn't overlap, should not count towards full range liquidity
			upperTick:       DefaultMaxTick,
			updateLiquidity: osmomath.NewDec(-100),
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
		actualFullRangeLiquidity := osmomath.ZeroDec()

		// Create a full range position.
		positionData, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool.GetId(), owner, tc.positionCoins)
		s.Require().NoError(err)
		actualFullRangeLiquidity = actualFullRangeLiquidity.Add(positionData.Liquidity)

		clPool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, clPoolId)
		s.Require().NoError(err)

		// Get the full range liquidity for the pool.
		expectedFullRangeLiquidity, err := s.App.ConcentratedLiquidityKeeper.GetFullRangeLiquidityInPool(s.Ctx, clPoolId)
		s.Require().NoError(err)
		s.Require().Equal(expectedFullRangeLiquidity, actualFullRangeLiquidity)

		// Create a new position that overlaps with the min tick, but is not full range and therefore should not count towards the full range liquidity.
		s.FundAcc(owner, tc.positionCoins)
		_, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, clPoolId, owner, DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), tc.lowerTick, tc.upperTick)
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
	invalidCoin0Denom := sdk.NewCoins(sdk.NewCoin("invalidDenom", osmomath.NewInt(1000000000000000000)), DefaultCoin1)
	invalidCoin1Denom := sdk.NewCoins(DefaultCoin0, sdk.NewCoin("invalidDenom", osmomath.NewInt(1000000000000000000)))
	zeroCoins := sdk.NewCoins()

	stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	defaultRemainingLockDuration := stakingParams.UnbondingTime

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
			expectedErr:           types.Amount0IsNegativeError{Amount0: osmomath.ZeroInt()},
		},
		{
			name:                  "invalid coin1 denom",
			remainingLockDuration: defaultRemainingLockDuration,
			coinsForPosition:      invalidCoin1Denom,
			expectedErr:           types.Amount1IsNegativeError{Amount1: osmomath.ZeroInt()},
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
			positionData, concentratedLockId, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePositionLocked(s.Ctx, clPool.GetId(), defaultAddress, test.coinsForPosition, test.remainingLockDuration)

			if test.expectedErr != nil {
				s.Require().ErrorContains(err, test.expectedErr.Error())
				return
			}

			s.Require().NoError(err)

			// Check position
			position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionData.ID)
			s.Require().NoError(err)
			s.Require().Equal(s.Ctx.BlockTime(), position.JoinTime)
			s.Require().Equal(types.MaxTick, position.UpperTick)
			s.Require().Equal(types.MinInitializedTick, position.LowerTick)
			s.Require().Equal(positionData.Liquidity, position.Liquidity)

			// Check locked
			concentratedLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, concentratedLockId)
			s.Require().NoError(err)

			s.Require().Equal(concentratedLock.Coins[0].Amount.String(), positionData.Liquidity.TruncateInt().String())
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
	firstPositionAssets := sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(9823358512)), sdk.NewCoin(USDC, osmomath.NewInt(8985893232)))
	firstPosLiq, firstPosId := s.SetupPosition(pool.GetId(), firstPositionAddr, firstPositionAssets, -68720000, -68710000, true)
	secondPositionAssets := sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(9823358512)), sdk.NewCoin(USDC, osmomath.NewInt(8985893232)))
	secondPosLiq, secondPosId := s.SetupPosition(pool.GetId(), secondPositionAddr, secondPositionAssets, -68720000, -68710000, true)

	// Execute a swap that brings the price close enough to the edge of a tick to trigger bankers rounding
	swapAddr := testAccs[2]
	desiredTokenOut := sdk.NewCoin(USDC, osmomath.NewInt(10000))
	s.FundAcc(swapAddr, sdk.NewCoins(sdk.NewCoin(ETH, osmomath.NewInt(1000000000000000000))))
	_, _, _, err := s.Clk.SwapInAmtGivenOut(s.Ctx, swapAddr, pool, desiredTokenOut, ETH, osmomath.ZeroDec(), osmomath.ZeroBigDec())
	s.Require().NoError(err)

	// Both positions should be able to withdraw successfully
	_, _, err = s.Clk.WithdrawPosition(s.Ctx, firstPositionAddr, firstPosId, firstPosLiq)
	s.Require().NoError(err)
	_, _, err = s.Clk.WithdrawPosition(s.Ctx, secondPositionAddr, secondPosId, secondPosLiq)
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
		"two adjacent ranges (flipped order)": {
			// Note: this setup covers both edge cases where initial interval accumulation is negative
			// for spread rewards and incentives
			tickRanges: [][]int64{
				{10000, 20000},
				{-10000, 10000},
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

// This test validates the edge case where the range accumulators become negative.
// It validates both spread and incentive rewards.
// It happens if we initialize a lower tick after the upper AND the current tick is above the position's range.
// To replicate, we create 3 positions full range, A and B.
// Full range position is created to inject some base liquidity.
// Position A is created next at the high range when no swaps are made and no spread rewards are accumulated.
// Some swaps are done to accumulate spread rewards.
// Position B is created where the lower tick is below position A and the upper tick equals to lower tick of position A.
// This results in the lower tick of position A being initialized to a greater value than its upper tick.
// Note, that we ensure that the current tick is above the position's range so that when we compute
// the in-range accumulator, it becomes negative (computed as upper tick acc - lower tick acc when current tick > upper tick of a position).
//
// Note that there is another edge case possible where we initialize an upper tick when current tick > upper tick to a positive value.
// Then, the current tick moves under the lower tick of a future position. As a result, when the position gets initialized,
// the lower tick gets the accumulator value of zero if it is new, resulting in interval accumulation of:
// lower tick accumulator snapshot - upper tick accumulator snapshot = 0 - positive value = negative value.
// This case is covered here implicitly.
//
// Finally, there are 4 sub-tests run to ensure that the total rewards are collected correctly:
// - Current tick is not moved.
// - Current tick is moved under position B's range
// - Current tick is moved in position B's range
// - Current tick is moved under and back above position B's range
func (s *KeeperTestSuite) TestNegativeTickRange_SpreadFactor() {
	s.SetupTest()

	var (
		// Initialize pool with non-zero spread factor.
		spreadFactor     = osmomath.NewDecWithPrec(3, 3)
		pool             = s.PrepareCustomConcentratedPool(s.TestAccs[0], DefaultCoin0.Denom, DefaultCoin1.Denom, 1, spreadFactor)
		poolId           = pool.GetId()
		denom0           = pool.GetToken0()
		denom1           = pool.GetToken1()
		rewardsPerSecond = osmomath.NewDec(1000)
		incentiveCoin    = sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1_000_000))
	)

	s.FundAcc(s.TestAccs[0], sdk.NewCoins(incentiveCoin))
	_, err := s.Clk.CreateIncentive(s.Ctx, poolId, s.TestAccs[0], incentiveCoin, rewardsPerSecond, s.Ctx.BlockTime(), time.Nanosecond)
	s.Require().NoError(err)

	// Estimates how much to swap in to approximately reach the given tick
	// in the zero for one direction (left). Assumes current sqrt price
	// from the refeteched pool as well as its liquidity. Assumes that
	// liquidity is constant between current tick and toTick.
	estimateCoinZeroIn := func(toTick int64) sdk.Coin {
		pool, err := s.Clk.GetPoolById(s.Ctx, poolId)
		s.Require().NoError(err)

		s.Require().True(toTick < pool.GetCurrentTick())

		amountZeroIn := math.CalcAmount0Delta(pool.GetLiquidity(), pool.GetCurrentSqrtPrice(), s.tickToSqrtPrice(toTick), true)
		coinZeroIn := sdk.NewCoin(denom0, amountZeroIn.Dec().TruncateInt())

		return coinZeroIn
	}

	// Estimates how much to swap in to approximately reach the given tick
	// in the one for zero direction (right). Assumes current sqrt price
	// from the refeteched pool as well as its liquidity. Assumes that
	// liquidity is constant between current tick and toTick.
	estimateCoinOneIn := func(toTick int64) sdk.Coin {
		pool, err := s.Clk.GetPoolById(s.Ctx, poolId)
		s.Require().NoError(err)

		s.Require().True(toTick > pool.GetCurrentTick())

		amountOneIn := math.CalcAmount1Delta(pool.GetLiquidity(), pool.GetCurrentSqrtPrice(), s.tickToSqrtPrice(toTick), true)
		coinOneIn := sdk.NewCoin(denom1, amountOneIn.Dec().TruncateInt())

		return coinOneIn
	}

	// Create full range position
	s.FundAcc(s.TestAccs[0], DefaultCoins)
	s.CreateFullRangePosition(pool, DefaultCoins)

	expectedTotalSpreadRewards := sdk.NewCoins()
	expectedTotalIncentiveRewards := osmomath.ZeroDec()

	// Initialize position at a higher range
	s.FundAcc(s.TestAccs[0], DefaultCoins)
	_, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, poolId, s.TestAccs[0], DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), DefaultCurrTick+50, DefaultCurrTick+100)
	s.Require().NoError(err)

	// Estimate how much to swap in to approximately DefaultCurrTick - 50
	coinZeroIn := estimateCoinZeroIn(DefaultCurrTick - 50)

	// Update expected spread rewards
	expectedTotalSpreadRewards = expectedTotalSpreadRewards.Add(sdk.NewCoin(denom0, coinZeroIn.Amount.ToLegacyDec().Mul(spreadFactor).Ceil().TruncateInt()))

	// Increase block time
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Second))
	// Update expected incentive rewards
	expectedTotalIncentiveRewards = expectedTotalIncentiveRewards.Add(rewardsPerSecond)

	s.swapZeroForOneLeftWithSpread(poolId, coinZeroIn, spreadFactor)

	// Refetch pool
	pool, err = s.Clk.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	// Swap to approximately DefaultCurrTick + 150
	coinOneIn := estimateCoinOneIn(DefaultCurrTick + 150)

	// Update expected spread rewards
	expectedTotalSpreadRewards = expectedTotalSpreadRewards.Add(sdk.NewCoin(denom1, coinOneIn.Amount.ToLegacyDec().Mul(spreadFactor).Ceil().TruncateInt()))

	// Increase block time
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Second))
	// Update expected incentive rewards
	expectedTotalIncentiveRewards = expectedTotalIncentiveRewards.Add(rewardsPerSecond)

	s.swapOneForZeroRightWithSpread(poolId, coinOneIn, spreadFactor)

	// Increase block time
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Second))
	// Update expected incentive rewards
	expectedTotalIncentiveRewards = expectedTotalIncentiveRewards.Add(rewardsPerSecond)

	// This previously panicked due to the lack of support for negative range accumulators.
	// See issue: https://github.com/osmosis-labs/osmosis/issues/5854
	// We initialized the lower tick's accumulator (DefaultCurrTick - 25) to be greater than the upper tick's accumulator (DefaultCurrTick + 50)
	// Whenever the current tick is above the position's range, we compute in range accumulator as upper tick accumulator - lower tick accumulator
	// In this case, it ends up being negative, which is now supported.
	negativeIntervalAccumPositionData, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, poolId, s.TestAccs[0], DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), DefaultCurrTick-25, DefaultCurrTick+50)
	s.Require().NoError(err)

	// Increase block time
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Second))
	// Update expected incentive rewards
	expectedTotalIncentiveRewards = expectedTotalIncentiveRewards.Add(rewardsPerSecond)

	s.T().Run("assert rewards when current tick is not moved (stays above position with negative in-range accumulator)", func(t *testing.T) {
		// Assert global invariants
		s.assertGlobalInvariants(ExpectedGlobalRewardValues{
			// Additive tolerance of 1 for each position.
			ExpectedAdditiveSpreadRewardTolerance: osmomath.OneDec().MulInt64(3),
			TotalSpreadRewards:                    expectedTotalSpreadRewards,
			TotalIncentives:                       sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, expectedTotalIncentiveRewards.Ceil().TruncateInt())),
		})
	})

	s.RunTestCaseWithoutStateUpdates("assert rewards when current tick is below the position with negative accumulator", func(t *testing.T) {
		// Make closure-local copy of expectedTotalSpreadRewards
		expectedTotalSpreadRewards := expectedTotalSpreadRewards
		expectedTotalIncentiveRewards := expectedTotalIncentiveRewards

		// Swap third time to cover the newly created position with negative range accumulator
		// Swap to approximately DefaultCurrTick - 50
		coinZeroIn = estimateCoinZeroIn(DefaultCurrTick - 50)

		// Update expected spread rewards
		expectedTotalSpreadRewards = expectedTotalSpreadRewards.Add(sdk.NewCoin(denom0, coinZeroIn.Amount.ToLegacyDec().Mul(spreadFactor).Ceil().TruncateInt()))

		// Increase block time
		s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Second))
		// Update expected incentive rewards
		expectedTotalIncentiveRewards = expectedTotalIncentiveRewards.Add(rewardsPerSecond)

		// Move current tick to be below the expected position
		s.swapZeroForOneLeftWithSpread(poolId, coinZeroIn, spreadFactor)

		// Assert global invariants
		s.assertGlobalInvariants(ExpectedGlobalRewardValues{
			// Additive tolerance of 1 for each position.
			ExpectedAdditiveSpreadRewardTolerance: osmomath.OneDec().MulInt64(3),
			TotalSpreadRewards:                    expectedTotalSpreadRewards,
			TotalIncentives:                       sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, expectedTotalIncentiveRewards.Ceil().TruncateInt())),
		})
	})

	s.RunTestCaseWithoutStateUpdates("assert rewards when current tick is inside the position with negative accumulator", func(t *testing.T) {
		// Make closure-local copy of expectedTotalSpreadRewards
		expectedTotalSpreadRewards := expectedTotalSpreadRewards
		expectedTotalIncentiveRewards := expectedTotalIncentiveRewards

		// Swap third time to cover the newly created position with negative range accumulator
		// Swap to approximately DefaultCurrTick - 10
		coinZeroIn = estimateCoinZeroIn(DefaultCurrTick - 10)

		// Update expected spread rewards
		expectedTotalSpreadRewards = expectedTotalSpreadRewards.Add(sdk.NewCoin(denom0, coinZeroIn.Amount.ToLegacyDec().Mul(spreadFactor).Ceil().TruncateInt()))

		// Increase block time
		s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Second))
		// Update expected incentive rewards
		expectedTotalIncentiveRewards = expectedTotalIncentiveRewards.Add(rewardsPerSecond)

		// Move current tick to be inside of the new position
		s.swapZeroForOneLeftWithSpread(poolId, coinZeroIn, spreadFactor)

		// Assert global invariants
		s.assertGlobalInvariants(ExpectedGlobalRewardValues{
			// Additive tolerance of 1 for each position.
			ExpectedAdditiveSpreadRewardTolerance: osmomath.OneDec().MulInt64(3),
			TotalSpreadRewards:                    expectedTotalSpreadRewards,
			TotalIncentives:                       sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, expectedTotalIncentiveRewards.Ceil().TruncateInt())),
		})
	})

	s.RunTestCaseWithoutStateUpdates("assert rewards when current tick is above the position with negative accumulator", func(t *testing.T) {
		// Make closure-local copy of expectedTotalSpreadRewards
		expectedTotalSpreadRewards := expectedTotalSpreadRewards
		expectedTotalIncentiveRewards := expectedTotalIncentiveRewards

		// Swap third time to cover the newly created position with negative range accumulator
		// Swap to approximately DefaultCurrTick - 50
		coinZeroIn = estimateCoinZeroIn(DefaultCurrTick - 50)

		// Update expected spread rewards
		expectedTotalSpreadRewards = expectedTotalSpreadRewards.Add(sdk.NewCoin(denom0, coinZeroIn.Amount.ToLegacyDec().Mul(spreadFactor).Ceil().TruncateInt()))

		// Increase block time
		s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Second))
		// Update expected incentive rewards
		expectedTotalIncentiveRewards = expectedTotalIncentiveRewards.Add(rewardsPerSecond)

		// Swap inside the new position so that it accumulates rewards
		s.swapZeroForOneLeftWithSpread(poolId, coinZeroIn, spreadFactor)

		// Estimate the next swap to be approximately until DefaultCurrTick + 150
		coinOneIn := estimateCoinOneIn(DefaultCurrTick + 150)

		// Update expected spread rewards
		expectedTotalSpreadRewards = expectedTotalSpreadRewards.Add(sdk.NewCoin(denom1, coinOneIn.Amount.ToLegacyDec().Mul(spreadFactor).Ceil().TruncateInt()))

		// Increase block time
		s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Second))
		// Update expected incentive rewards
		expectedTotalIncentiveRewards = expectedTotalIncentiveRewards.Add(rewardsPerSecond)

		// Swap back to take current tick be above the new position
		s.swapOneForZeroRightWithSpread(poolId, coinOneIn, spreadFactor)

		// Assert global invariants
		s.assertGlobalInvariants(ExpectedGlobalRewardValues{
			// Additive tolerance of 1 for each position.
			ExpectedAdditiveSpreadRewardTolerance: osmomath.OneDec().MulInt64(3),
			TotalSpreadRewards:                    expectedTotalSpreadRewards,
			TotalIncentives:                       sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, expectedTotalIncentiveRewards.Ceil().TruncateInt())),
		})
	})

	// Export and import genesis to make sure that negative accumulation does not lead to unexpected
	// panics in serialization and deserialization.
	spreadRewardAccumulator, err := s.Clk.GetSpreadRewardAccumulator(s.Ctx, poolId)
	s.Require().NoError(err)

	accum, err := spreadRewardAccumulator.GetPosition(types.KeySpreadRewardPositionAccumulator(negativeIntervalAccumPositionData.ID))
	s.Require().NoError(err)

	// Validate that at least one accumulator is negative for the test to be valid.
	s.Require().True(accum.AccumValuePerShare.IsAnyNegative())

	export := s.Clk.ExportGenesis(s.Ctx)

	s.SetupTest()

	s.Clk.InitGenesis(s.Ctx, *export)
}

// TestTransferPositions validates the following:
// - Positions can be transferred from one owner to another.
// - The transfer of positions does not modify the positions that are not transferred.
// - The outstanding incentives and spread rewards go to the old owner after the transfer.
// - The new owner does not receive the outstanding incentives and spread rewards after the transfer.
// - Claiming incentives/spread rewards with the new owner returns nothing after the transfer.
// - Adding incentives/spread rewards and then claiming returns it to the new owner, and the old owner does not get anything.
// - The new owner can withdraw the positions and receive the correct amount of funds.
// The test also checks for expected errors such as:
// - Attempting to transfer a position ID that does not exist.
// - Attempting to transfer a position that the sender does not own.
// - Attempting to transfer the last position in the pool.
func (s *KeeperTestSuite) TestTransferPositions() {
	// expectedUptimes are used for claimable incentives tests
	expectedUptimes := getExpectedUptimes()

	errTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: osmomath.NewDec(1),
		// Actual amount should be less than expected, so we round down
		// This is because when we withdraw the position, we always round in favor of the pool
		RoundingDir: osmomath.RoundDown,
	}

	oldOwner := s.TestAccs[0]
	newOwner := s.TestAccs[1]

	testcases := map[string]struct {
		inRangePositions     []uint64
		outOfRangePositions  []uint64
		positionsToTransfer  []uint64
		setupUnownedPosition bool
		isLastPositionInPool bool
		isGovAddress         bool

		expectedError error
	}{
		"single position ID in range": {
			inRangePositions:    []uint64{DefaultPositionId},
			positionsToTransfer: []uint64{DefaultPositionId},
		},
		"two position IDs, one in range, one out of range": {
			inRangePositions:    []uint64{DefaultPositionId},
			outOfRangePositions: []uint64{DefaultPositionId + 1},
			positionsToTransfer: []uint64{DefaultPositionId, DefaultPositionId + 1},
		},
		"two position IDs in range": {
			inRangePositions:    []uint64{DefaultPositionId, DefaultPositionId + 1},
			positionsToTransfer: []uint64{DefaultPositionId, DefaultPositionId + 1},
		},
		"three position IDs in range": {
			inRangePositions:    []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
			positionsToTransfer: []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
		},
		"three position IDs, two in range, one out of range": {
			inRangePositions:    []uint64{DefaultPositionId, DefaultPositionId + 1},
			outOfRangePositions: []uint64{DefaultPositionId + 2},
			positionsToTransfer: []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
		},
		"three position IDs, two in range, one out of range, only transfer the two in range": {
			inRangePositions:    []uint64{DefaultPositionId, DefaultPositionId + 1},
			outOfRangePositions: []uint64{DefaultPositionId + 2},
			positionsToTransfer: []uint64{DefaultPositionId, DefaultPositionId + 1},
		},
		"three position IDs, one in range, two out of range, transfer one in range and one out of range": {
			inRangePositions:    []uint64{DefaultPositionId},
			outOfRangePositions: []uint64{DefaultPositionId + 1, DefaultPositionId + 2},
			positionsToTransfer: []uint64{DefaultPositionId, DefaultPositionId + 2},
		},
		"three position IDs, not an owner of any of them but caller is gov address": {
			inRangePositions:    []uint64{DefaultPositionId, DefaultPositionId + 1},
			outOfRangePositions: []uint64{DefaultPositionId + 2},
			positionsToTransfer: []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
			isGovAddress:        true,
		},
		"error: two position IDs, second ID does not exist": {
			inRangePositions:    []uint64{DefaultPositionId, DefaultPositionId + 1},
			positionsToTransfer: []uint64{DefaultPositionId, DefaultPositionId + 3},
			expectedError:       types.PositionIdNotFoundError{PositionId: DefaultPositionId + 3},
		},
		"error: three position IDs, not an owner of one of them": {
			inRangePositions:     []uint64{DefaultPositionId, DefaultPositionId + 1},
			positionsToTransfer:  []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
			setupUnownedPosition: true,
			expectedError:        types.PositionOwnerMismatchError{PositionOwner: newOwner.String(), Sender: oldOwner.String()},
		},
		"error: attempt to transfer last position in pool": {
			inRangePositions:     []uint64{DefaultPositionId},
			positionsToTransfer:  []uint64{DefaultPositionId},
			isLastPositionInPool: true,
			expectedError:        types.LastPositionTransferError{PositionId: DefaultPositionId, PoolId: 1},
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()
			pool := s.PrepareConcentratedPool()

			lastPositionId := 0

			// Setup in range positions
			for i := 0; i < len(tc.inRangePositions); i++ {
				s.SetupDefaultPosition(pool.GetId())
				lastPositionId++
			}

			// Setup out of range positions
			for i := 0; i < len(tc.outOfRangePositions); i++ {
				// Position with out of range ticks.
				s.SetupPosition(pool.GetId(), oldOwner, sdk.NewCoins(DefaultCoin1), DefaultMinTick, DefaultMinTick+100, true)
				lastPositionId++
			}

			// Setup unowned position (owned by newOwner)
			if tc.setupUnownedPosition {
				s.SetupDefaultPositionAcc(pool.GetId(), newOwner)
				lastPositionId++
			}

			// Setup a far out of range position that we do not touch, so when we transfer positions we do not transfer the last position in the pool.
			// This is because we special case this logic in the keeper to not allow the last position in the pool to be transferred.
			if !tc.isLastPositionInPool {
				s.SetupPosition(pool.GetId(), s.TestAccs[2], sdk.NewCoins(DefaultCoin0), DefaultMaxTick-100, DefaultMaxTick, true)
			}

			// For each position that is in range, add spread rewards and incentives to their respective addresses
			totalSpreadRewards := s.fundSpreadRewardsAddr(s.Ctx, pool.GetSpreadRewardsAddress(), tc.inRangePositions)
			totalIncentives := s.fundIncentiveAddr(pool.GetIncentivesAddress(), tc.inRangePositions)
			totalExpectedRewards := totalSpreadRewards.Add(totalIncentives...)

			// Add spread rewards and incentives to the pool
			s.addUptimeGrowthInsideRange(s.Ctx, pool.GetId(), apptesting.DefaultLowerTick+1, DefaultLowerTick, DefaultUpperTick, expectedUptimes.hundredTokensMultiDenom)
			s.AddToSpreadRewardAccumulator(pool.GetId(), sdk.NewDecCoin(ETH, osmomath.NewInt(10)))

			// Move block time forward. In the event we have positions in range
			// this allows us to test both collected and forfeited incentives
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour * 24))

			initialUserPositions, err := s.App.ConcentratedLiquidityKeeper.GetUserPositions(s.Ctx, oldOwner, 1)
			s.Require().NoError(err)

			// Account funds of original owner
			preTransferOwnerFunds := s.App.BankKeeper.GetAllBalances(s.Ctx, oldOwner)

			// Account funds of new owner
			preTransferNewOwnerFunds := s.App.BankKeeper.GetAllBalances(s.Ctx, newOwner)

			transferCaller := oldOwner
			if tc.isGovAddress {
				transferCaller = s.App.AccountKeeper.GetModuleAddress(govtypes.ModuleName)
			}

			// System under test
			err = s.App.ConcentratedLiquidityKeeper.TransferPositions(s.Ctx, tc.positionsToTransfer, transferCaller, newOwner)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectedError)
			} else {
				s.Require().NoError(err)

				// Check that the positions we wanted transferred were modified appropriately
				for _, positionId := range tc.positionsToTransfer {
					newPosition, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionId)
					s.Require().NoError(err)

					oldPosition := model.Position{}
					for _, initialPosition := range initialUserPositions {
						if initialPosition.PositionId == newPosition.PositionId {
							oldPosition = initialPosition
							break
						}
					}

					// All position values except the owner should be the same in the new position as it was in the old one.
					s.Require().Equal(oldPosition.UpperTick, newPosition.UpperTick)
					s.Require().Equal(oldPosition.LowerTick, newPosition.LowerTick)
					s.Require().Equal(oldPosition.PoolId, newPosition.PoolId)
					s.Require().Equal(oldPosition.JoinTime, newPosition.JoinTime)
					s.Require().Equal(oldPosition.Liquidity, newPosition.Liquidity)

					// The new position should have the new owner
					s.Require().Equal(newOwner.String(), newPosition.Address)
				}

				allPositions := append(tc.inRangePositions, tc.outOfRangePositions...)
				positionsNotTransfered := osmoutils.DisjointArrays(allPositions, tc.positionsToTransfer)

				// Check that the positions not transferred were not modified
				for _, positionId := range positionsNotTransfered {
					oldPosition, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionId)
					s.Require().NoError(err)

					newPosition := model.Position{}
					for _, initialPosition := range initialUserPositions {
						if initialPosition.PositionId == oldPosition.PositionId {
							newPosition = initialPosition
							break
						}
					}

					// All position values should be the same in the new position as it was in the old one.
					s.Require().Equal(oldPosition, newPosition)
				}

				// Check that the old owner's balance did not change due to the transfer
				postTransferOriginalOwnerFunds := s.App.BankKeeper.GetAllBalances(s.Ctx, oldOwner)
				s.Require().Equal(preTransferOwnerFunds.String(), postTransferOriginalOwnerFunds.String())

				// Check that the new owner's balance did not change due to the transfer
				postTransferNewOwnerFunds := s.App.BankKeeper.GetAllBalances(s.Ctx, newOwner)
				s.Require().Equal(preTransferNewOwnerFunds, postTransferNewOwnerFunds)

				// Claim rewards and ensure that previously accrued incentives and spread rewards go to the new owner
				for _, positionID := range tc.positionsToTransfer {
					_, err = s.App.ConcentratedLiquidityKeeper.CollectSpreadRewards(s.Ctx, newOwner, positionID)
					s.Require().NoError(err)
					_, _, _, err := s.App.ConcentratedLiquidityKeeper.CollectIncentives(s.Ctx, newOwner, positionID)
					s.Require().NoError(err)
				}

				// Ensure all rewards went to the new owner
				postClaimRewardsNewOwnerBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, newOwner)
				s.Require().Equal(totalExpectedRewards, postClaimRewardsNewOwnerBalance)

				// Ensure no rewards went to the old owner
				postClaimRewardsOldOwnerBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, oldOwner)
				s.Require().Equal(preTransferOwnerFunds.String(), postClaimRewardsOldOwnerBalance.String())

				// Test that adding incentives/spread rewards and then claiming returns it to the new owner, and the old owner does not get anything
				totalSpreadRewards := s.fundSpreadRewardsAddr(s.Ctx, pool.GetSpreadRewardsAddress(), tc.inRangePositions)
				totalIncentives := s.fundIncentiveAddr(pool.GetIncentivesAddress(), tc.inRangePositions)
				totalExpectedRewards := totalExpectedRewards.Add(totalSpreadRewards...).Add(totalIncentives...)
				s.addUptimeGrowthInsideRange(s.Ctx, pool.GetId(), apptesting.DefaultLowerTick+1, DefaultLowerTick, DefaultUpperTick, expectedUptimes.hundredTokensMultiDenom)
				s.AddToSpreadRewardAccumulator(pool.GetId(), sdk.NewDecCoin(ETH, osmomath.NewInt(10)))
				for _, positionId := range tc.positionsToTransfer {
					_, _, _, err := s.App.ConcentratedLiquidityKeeper.CollectIncentives(s.Ctx, newOwner, positionId)
					s.Require().NoError(err)
					_, err = s.App.ConcentratedLiquidityKeeper.CollectSpreadRewards(s.Ctx, newOwner, positionId)
					s.Require().NoError(err)
				}
				// New owner balance check
				postSecondTransferNewOwnerFunds := s.App.BankKeeper.GetAllBalances(s.Ctx, newOwner)
				expectedSecondTransferNewOwnerFunds := postTransferNewOwnerFunds.Add(totalExpectedRewards...)
				s.Require().Equal(expectedSecondTransferNewOwnerFunds.String(), postSecondTransferNewOwnerFunds.String())
				// Old owner balance check
				postSecondTransferOriginalOwnerFunds := s.App.BankKeeper.GetAllBalances(s.Ctx, oldOwner)
				s.Require().Equal(postTransferOriginalOwnerFunds, postSecondTransferOriginalOwnerFunds)

				// Test that withdrawing the positions returns the correct amount of funds to the new owner
				for _, positionId := range tc.positionsToTransfer {
					underlyingPositionsValue, err := s.App.ConcentratedLiquidityKeeper.UnderlyingPositionsValue(s.Ctx, []uint64{positionId})
					s.Require().NoError(err)
					position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionId)
					s.Require().NoError(err)
					amt0, amt1, err := s.App.ConcentratedLiquidityKeeper.WithdrawPosition(s.Ctx, newOwner, positionId, position.Liquidity)
					s.Require().NoError(err)
					coinsWithdrawn := sdk.NewCoins(sdk.NewCoin(pool.GetToken0(), amt0), sdk.NewCoin(pool.GetToken1(), amt1))
					// Amount we withdraw is one less than actual value due to rounding in favor of pool
					for i, coin := range coinsWithdrawn {
						osmoassert.Equal(
							s.T(),
							errTolerance,
							underlyingPositionsValue[i].Amount,
							coin.Amount,
						)
					}
				}
			}
		})
	}
}

// fundIncentiveAddr funds the incentive address for each position ID in the provided slice.
// It calculates the expected incentives based on uptime growth and adds these incentives to the total expected rewards.
// It also determines how much position will forfeit and funds this amount to the incentive address.
// The function returns the total expected rewards after funding all the positions.
//
// Parameters:
// - ctx: The context of the operation.
// - incentivesAddress: The address to which the incentives will be funded.
// - positionIds: A slice of position IDs for which the incentives will be funded.
//
// Returns:
// - totalExpectedRewards: The total expected rewards after funding all the positions.
func (s *KeeperTestSuite) fundIncentiveAddr(incentivesAddress sdk.AccAddress, positionIds []uint64) (totalExpectedRewards sdk.Coins) {
	expectedUptimes := getExpectedUptimes()
	for i := 0; i < len(positionIds); i++ {
		coinsToFundForIncentivesToUser := expectedIncentivesFromUptimeGrowth(expectedUptimes.hundredTokensMultiDenom, DefaultLiquidityAmt, time.Hour*24, defaultMultiplier)
		totalExpectedRewards = totalExpectedRewards.Add(coinsToFundForIncentivesToUser...)
		s.FundAcc(incentivesAddress, coinsToFundForIncentivesToUser)
		// Determine how much position will forfeit and fund
		coinsToFundForForefeitToPool := expectedIncentivesFromUptimeGrowth(expectedUptimes.hundredTokensMultiDenom, DefaultLiquidityAmt, time.Hour*24*14, defaultMultiplier).Sub(expectedIncentivesFromUptimeGrowth(expectedUptimes.hundredTokensMultiDenom, DefaultLiquidityAmt, time.Hour*24, defaultMultiplier)...)
		s.FundAcc(incentivesAddress, coinsToFundForForefeitToPool)
	}
	return
}

// fundSpreadRewardsAddr funds the spread rewards address for each position ID in the provided slice.
// It calculates the expected amount to claim based on the position's liquidity and adds these rewards to the total expected rewards.
// The function then funds the spread rewards account with the total expected rewards.
//
// Parameters:
// - ctx: The context of the operation.
// - spreadRewardsAddress: The address to which the spread rewards will be funded.
// - positionIds: A slice of position IDs for which the spread rewards will be funded.
//
// Returns:
// - totalExpectedRewards: The total expected rewards after funding all the positions.
func (s *KeeperTestSuite) fundSpreadRewardsAddr(ctx sdk.Context, spreadRewardsAddress sdk.AccAddress, positionIds []uint64) (totalExpectedRewards sdk.Coins) {
	for _, positionId := range positionIds {
		position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(ctx, positionId)
		s.Require().NoError(err)

		expectedAmountToClaim := position.Liquidity.MulInt(osmomath.NewInt(10)).TruncateInt()
		totalExpectedRewards = totalExpectedRewards.Add(sdk.NewCoin(ETH, expectedAmountToClaim))
		// Fund the spread rewards account with the expected rewards and add to the pool's accum
		s.FundAcc(spreadRewardsAddress, totalExpectedRewards)
	}
	return
}
