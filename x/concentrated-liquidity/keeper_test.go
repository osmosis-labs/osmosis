package concentrated_liquidity_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	concentrated_liquidity "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/clmocks"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"

	cl "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity"

	"github.com/osmosis-labs/osmosis/v16/app/apptesting"
)

var (
	DefaultMinTick, DefaultMaxTick                       = types.MinTick, types.MaxTick
	DefaultLowerPrice                                    = sdk.NewDec(4545)
	DefaultLowerTick                                     = int64(30545000)
	DefaultUpperPrice                                    = sdk.NewDec(5500)
	DefaultUpperTick                                     = int64(31500000)
	DefaultCurrPrice                                     = sdk.NewDec(5000)
	DefaultCurrTick                                int64 = 31000000
	DefaultCurrSqrtPrice, _                              = DefaultCurrPrice.ApproxSqrt() // 70.710678118654752440
	DefaultZeroSpreadFactor                              = sdk.ZeroDec()
	DefaultSpreadRewardAccumCoins                        = sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(50)))
	DefaultPositionId                                    = uint64(1)
	DefaultUnderlyingLockId                              = uint64(0)
	DefaultJoinTime                                      = time.Unix(0, 0).UTC()
	ETH                                                  = "eth"
	DefaultAmt0                                          = sdk.NewInt(1000000)
	DefaultAmt0Expected                                  = sdk.NewInt(998976)
	DefaultCoin0                                         = sdk.NewCoin(ETH, DefaultAmt0)
	USDC                                                 = "usdc"
	DefaultAmt1                                          = sdk.NewInt(5000000000)
	DefaultAmt1Expected                                  = sdk.NewInt(5000000000)
	DefaultCoin1                                         = sdk.NewCoin(USDC, DefaultAmt1)
	DefaultCoins                                         = sdk.NewCoins(DefaultCoin0, DefaultCoin1)
	DefaultLiquidityAmt                                  = sdk.MustNewDecFromStr("1517882343.751510418088349649")
	FullRangeLiquidityAmt                                = sdk.MustNewDecFromStr("70710678.118654752940000000")
	DefaultTickSpacing                                   = uint64(100)
	PoolCreationFee                                      = poolmanagertypes.DefaultParams().PoolCreationFee
	DefaultExponentConsecutivePositionLowerTick, _       = math.PriceToTickRoundDown(sdk.NewDec(5500), DefaultTickSpacing)
	DefaultExponentConsecutivePositionUpperTick, _       = math.PriceToTickRoundDown(sdk.NewDec(6250), DefaultTickSpacing)
	DefaultExponentOverlappingPositionLowerTick, _       = math.PriceToTickRoundDown(sdk.NewDec(4000), DefaultTickSpacing)
	DefaultExponentOverlappingPositionUpperTick, _       = math.PriceToTickRoundDown(sdk.NewDec(4999), DefaultTickSpacing)
	BAR                                                  = "bar"
	FOO                                                  = "foo"
	InsufficientFundsError                               = fmt.Errorf("insufficient funds")
	DefaultAuthorizedUptimes                             = []time.Duration{time.Nanosecond}
	ThreeOrderedConsecutiveAuthorizedUptimes             = []time.Duration{time.Nanosecond, time.Minute, time.Hour, time.Hour * 24}
	ThreeUnorderedNonConsecutiveAuthorizedUptimes        = []time.Duration{time.Nanosecond, time.Hour * 24 * 7, time.Minute}
	AllUptimesAuthorized                                 = types.SupportedUptimes
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
	clk               *concentrated_liquidity.Keeper
	authorizedUptimes []time.Duration
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.clk = s.App.ConcentratedLiquidityKeeper

	if s.authorizedUptimes != nil {
		clParams := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
		clParams.AuthorizedUptimes = s.authorizedUptimes
		s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, clParams)
	}
}

func (s *KeeperTestSuite) SetupDefaultPosition(poolId uint64) {
	s.SetupPosition(poolId, s.TestAccs[0], DefaultCoins, DefaultLowerTick, DefaultUpperTick, false)
}

func (s *KeeperTestSuite) SetupPosition(poolId uint64, owner sdk.AccAddress, providedCoins sdk.Coins, lowerTick, upperTick int64, addRoundingError bool) (sdk.Dec, uint64) {
	roundingErrorCoins := sdk.NewCoins()
	if addRoundingError {
		roundingErrorCoins = sdk.NewCoins(sdk.NewCoin(ETH, roundingError), sdk.NewCoin(USDC, roundingError))
	}

	s.FundAcc(owner, providedCoins.Add(roundingErrorCoins...))
	fmt.Println("owner balances before liq: ", s.App.BankKeeper.GetAllBalances(s.Ctx, owner))
	positionId, actual0, actual1, liquidityDelta, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, poolId, owner, providedCoins, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick, upperTick)
	s.Require().NoError(err)
	liquidity, err := s.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(s.Ctx, positionId)
	s.Require().NoError(err)
	fmt.Println("actual0, actual1, liquidityDelta, liquidity: ", actual0, actual1, liquidityDelta, liquidity)
	return liquidity, positionId
}

// SetupDefaultPositions sets up four different positions to the given pool with different accounts for each position./
// Sets up the following positions:
// 1. Default position
// 2. Full range position
// 3. Position with consecutive price range from the default position
// 4. Position with overlapping price range from the default position
func (s *KeeperTestSuite) SetupDefaultPositions(poolId uint64) {
	// ----------- set up positions ----------
	// 1. Default position
	s.SetupDefaultPosition(poolId)

	// 2. Full range position
	s.SetupFullRangePositionAcc(poolId, s.TestAccs[1])

	// 3. Position with consecutive price range from the default position
	s.SetupOverlappingRangePositionAcc(poolId, s.TestAccs[2])

	// 4. Position with overlapping price range from the default position
	s.SetupOverlappingRangePositionAcc(poolId, s.TestAccs[3])
}

func (s *KeeperTestSuite) SetupDefaultPositionAcc(poolId uint64, owner sdk.AccAddress) uint64 {
	_, positionId := s.SetupPosition(poolId, owner, DefaultCoins, DefaultLowerTick, DefaultUpperTick, false)
	return positionId
}

func (s *KeeperTestSuite) SetupFullRangePositionAcc(poolId uint64, owner sdk.AccAddress) uint64 {
	_, positionId := s.SetupPosition(poolId, owner, DefaultCoins, DefaultMinTick, DefaultMaxTick, false)
	return positionId
}

func (s *KeeperTestSuite) SetupConsecutiveRangePositionAcc(poolId uint64, owner sdk.AccAddress) uint64 {
	_, positionId := s.SetupPosition(poolId, owner, DefaultCoins, DefaultExponentConsecutivePositionLowerTick, DefaultExponentConsecutivePositionUpperTick, false)
	return positionId
}

func (s *KeeperTestSuite) SetupOverlappingRangePositionAcc(poolId uint64, owner sdk.AccAddress) uint64 {
	_, positionId := s.SetupPosition(poolId, owner, DefaultCoins, DefaultExponentOverlappingPositionLowerTick, DefaultExponentOverlappingPositionUpperTick, false)
	return positionId
}

// validatePositionUpdate validates that position with given parameters has expectedRemainingLiquidity left.
func (s *KeeperTestSuite) validatePositionUpdate(ctx sdk.Context, positionId uint64, expectedRemainingLiquidity sdk.Dec) {
	newPositionLiquidity, err := s.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(ctx, positionId)
	s.Require().NoError(err)
	s.Require().Equal(expectedRemainingLiquidity.String(), newPositionLiquidity.String())
	s.Require().True(newPositionLiquidity.GTE(sdk.ZeroDec()))
}

// validateTickUpdates validates that ticks with the given parameters have expectedRemainingLiquidity left.
func (s *KeeperTestSuite) validateTickUpdates(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick int64, upperTick int64, expectedRemainingLiquidity sdk.Dec, expectedLowerSpreadRewardGrowthOppositeDirectionOfLastTraversal, expectedUpperSpreadRewardGrowthOppositeDirectionOfLastTraversal sdk.DecCoins) {
	lowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, poolId, lowerTick)
	s.Require().NoError(err)
	s.Require().Equal(expectedRemainingLiquidity.String(), lowerTickInfo.LiquidityGross.String())
	s.Require().Equal(expectedRemainingLiquidity.String(), lowerTickInfo.LiquidityNet.String())
	s.Require().Equal(expectedLowerSpreadRewardGrowthOppositeDirectionOfLastTraversal.String(), lowerTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal.String())

	upperTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, poolId, upperTick)
	s.Require().NoError(err)
	s.Require().Equal(expectedRemainingLiquidity.String(), upperTickInfo.LiquidityGross.String())
	s.Require().Equal(expectedRemainingLiquidity.Neg().String(), upperTickInfo.LiquidityNet.String())
	s.Require().Equal(expectedUpperSpreadRewardGrowthOppositeDirectionOfLastTraversal.String(), upperTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal.String())
}

func (s *KeeperTestSuite) initializeTick(ctx sdk.Context, currentTick int64, tickIndex int64, initialLiquidity sdk.Dec, spreadRewardGrowthOppositeDirectionOfTraversal sdk.DecCoins, uptimeTrackers []model.UptimeTracker, isLower bool) {
	err := s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(ctx, validPoolId, currentTick, tickIndex, initialLiquidity, isLower)
	s.Require().NoError(err)

	tickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(ctx, validPoolId, tickIndex)
	s.Require().NoError(err)

	tickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal = spreadRewardGrowthOppositeDirectionOfTraversal
	tickInfo.UptimeTrackers = model.UptimeTrackers{
		List: uptimeTrackers,
	}

	s.App.ConcentratedLiquidityKeeper.SetTickInfo(ctx, validPoolId, tickIndex, &tickInfo)
}

// initializeSpreadRewardsAccumulatorPositionWithLiquidity initializes spread factor accumulator position with given parameters and updates it with given liquidity.
func (s *KeeperTestSuite) initializeSpreadRewardAccumulatorPositionWithLiquidity(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64, positionId uint64, liquidity sdk.Dec) {
	err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePositionSpreadRewardAccumulator(ctx, poolId, lowerTick, upperTick, positionId, liquidity)
	s.Require().NoError(err)
}

// addLiquidityToUptimeAccumulators adds shares to all uptime accumulators as defined by the `liquidity` parameter.
// This helper is primarily used to test incentive accrual for specific tick ranges, so we pass in filler values
// for all other components (e.g. join time).
func (s *KeeperTestSuite) addLiquidityToUptimeAccumulators(ctx sdk.Context, poolId uint64, liquidity []sdk.Dec, positionId uint64) {
	s.Require().Equal(len(liquidity), len(types.SupportedUptimes))

	uptimeAccums, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(ctx, poolId)
	s.Require().NoError(err)

	positionName := string(types.KeyPositionId(positionId))

	for uptimeIndex, uptimeAccum := range uptimeAccums {
		err := uptimeAccum.NewPosition(positionName, liquidity[uptimeIndex], &accum.Options{})
		s.Require().NoError(err)
	}
}

// addUptimeGrowthInsideRange adds uptime growth inside the range defined by [lowerTick, upperTick).
//
// By convention, we add additional growth below the range. This translates to the following logic:
//
//   - If currentTick < lowerTick < upperTick, we add to the lower tick's trackers, but not the upper's.
//
//   - If lowerTick <= currentTick < upperTick, we add to just the global accumulators.
//
//   - If lowerTick < upperTick <= currentTick, we add to the upper tick's trackers, but not the lower's.
func (s *KeeperTestSuite) addUptimeGrowthInsideRange(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, currentTick, lowerTick, upperTick int64, uptimeGrowthToAdd []sdk.DecCoins) {
	s.Require().True(lowerTick <= upperTick)

	// Note that we process adds to global accums at the end to ensure that they don't affect the behavior of uninitialized ticks.
	if currentTick < lowerTick {
		// Add to lower tick's uptime trackers
		lowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(ctx, poolId, lowerTick)
		s.Require().NoError(err)
		s.Require().Equal(len(lowerTickInfo.UptimeTrackers.List), len(uptimeGrowthToAdd))

		newLowerUptimeTrackerValues, err := osmoutils.AddDecCoinArrays(cl.GetUptimeTrackerValues(lowerTickInfo.UptimeTrackers.List), uptimeGrowthToAdd)
		s.Require().NoError(err)

		s.initializeTick(ctx, currentTick, lowerTick, lowerTickInfo.LiquidityGross, lowerTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal, wrapUptimeTrackers(newLowerUptimeTrackerValues), true)
	} else if upperTick <= currentTick {
		// Add to upper tick uptime trackers
		upperTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(ctx, poolId, upperTick)
		s.Require().NoError(err)
		s.Require().Equal(len(upperTickInfo.UptimeTrackers.List), len(uptimeGrowthToAdd))

		newUpperUptimeTrackerValues, err := osmoutils.AddDecCoinArrays(cl.GetUptimeTrackerValues(upperTickInfo.UptimeTrackers.List), uptimeGrowthToAdd)
		s.Require().NoError(err)

		s.initializeTick(ctx, currentTick, upperTick, upperTickInfo.LiquidityGross, upperTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal, wrapUptimeTrackers(newUpperUptimeTrackerValues), false)
	}

	// In all cases, global uptime accums need to be updated. If lowerTick <= currentTick < upperTick,
	// nothing more needs to be done.
	err := addToUptimeAccums(ctx, poolId, s.App.ConcentratedLiquidityKeeper, uptimeGrowthToAdd)
	s.Require().NoError(err)
}

// addUptimeGrowthOutsideRange adds uptime growth outside the range defined by [lowerTick, upperTick).
//
// By convention, we add additional growth below the range. This translates to the following logic:
//
//   - If currentTick < lowerTick < upperTick, we add to global accumulators to put the growth
//     below the tick range.
//
//   - If lowerTick <= currentTick < upperTick, we add to lowerTick's uptime trackers to put the
//     growth below the tick range.
//
//   - If lowerTick < upperTick <= currentTick, we add to both lowerTick and upperTick's uptime trackers,
//     the former to put the growth below the tick range and the latter to keep both ticks consistent (since
//     lowerTick's uptime trackers are a subset of upperTick's in this case).
func (s *KeeperTestSuite) addUptimeGrowthOutsideRange(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, currentTick, lowerTick, upperTick int64, uptimeGrowthToAdd []sdk.DecCoins) {
	s.Require().True(lowerTick <= upperTick)

	// Note that we process adds to global accums at the end to ensure that they don't affect the behavior of uninitialized ticks.
	if currentTick < lowerTick || upperTick <= currentTick {
		// Add to lower tick uptime trackers
		lowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(ctx, poolId, lowerTick)
		s.Require().NoError(err)
		s.Require().Equal(len(lowerTickInfo.UptimeTrackers.List), len(uptimeGrowthToAdd))

		newLowerUptimeTrackerValues, err := osmoutils.AddDecCoinArrays(cl.GetUptimeTrackerValues(lowerTickInfo.UptimeTrackers.List), uptimeGrowthToAdd)
		s.Require().NoError(err)

		s.initializeTick(ctx, currentTick, lowerTick, lowerTickInfo.LiquidityGross, lowerTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal, wrapUptimeTrackers(newLowerUptimeTrackerValues), true)

		// Add to upper tick uptime trackers
		upperTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(ctx, poolId, upperTick)
		s.Require().NoError(err)
		s.Require().Equal(len(upperTickInfo.UptimeTrackers.List), len(uptimeGrowthToAdd))

		newUpperUptimeTrackerValues, err := osmoutils.AddDecCoinArrays(cl.GetUptimeTrackerValues(upperTickInfo.UptimeTrackers.List), uptimeGrowthToAdd)
		s.Require().NoError(err)

		s.initializeTick(ctx, currentTick, upperTick, upperTickInfo.LiquidityGross, upperTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal, wrapUptimeTrackers(newUpperUptimeTrackerValues), false)
	} else if currentTick < upperTick {
		// Add to lower tick's uptime trackers
		lowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(ctx, poolId, lowerTick)
		s.Require().NoError(err)
		s.Require().Equal(len(lowerTickInfo.UptimeTrackers.List), len(uptimeGrowthToAdd))

		newLowerUptimeTrackerValues, err := osmoutils.AddDecCoinArrays(cl.GetUptimeTrackerValues(lowerTickInfo.UptimeTrackers.List), uptimeGrowthToAdd)
		s.Require().NoError(err)

		s.initializeTick(ctx, currentTick, lowerTick, lowerTickInfo.LiquidityGross, lowerTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal, wrapUptimeTrackers(newLowerUptimeTrackerValues), true)
	}

	// In all cases, global uptime accums need to be updated. If currentTick < lowerTick,
	// nothing more needs to be done.
	err := addToUptimeAccums(ctx, poolId, s.App.ConcentratedLiquidityKeeper, uptimeGrowthToAdd)
	s.Require().NoError(err)
}

// validatePositionSpreadFactorAccUpdate validates that the position's accumulator with given parameters
// has been updated with liquidity.
func (s *KeeperTestSuite) validatePositionSpreadRewardAccUpdate(ctx sdk.Context, poolId uint64, positionId uint64, liquidity sdk.Dec) {
	accum, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(ctx, poolId)
	s.Require().NoError(err)

	accumulatorPosition, err := accum.GetPositionSize(types.KeySpreadRewardPositionAccumulator(positionId))
	s.Require().NoError(err)

	s.Require().Equal(liquidity.String(), accumulatorPosition.String())
}

// validateListenerCallCount validates that the listeners were invoked the expected number of times.
func (s *KeeperTestSuite) validateListenerCallCount(
	expectedPoolCreatedListenerCallCount,
	expectedInitialPositionCreationListenerCallCount,
	expectedLastPositionWithdrawalListenerCallCount,
	expectedSwapListenerCallCount int,
) {
	// Validate that listeners were called the desired number of times
	listeners := s.App.ConcentratedLiquidityKeeper.GetListenersUnsafe()
	s.Require().Len(listeners, 1)

	mockListener, ok := listeners[0].(*clmocks.ConcentratedLiquidityListenerMock)
	s.Require().True(ok)

	s.Require().Equal(expectedPoolCreatedListenerCallCount, mockListener.AfterConcentratedPoolCreatedCallCount)
	s.Require().Equal(expectedInitialPositionCreationListenerCallCount, mockListener.AfterInitialPoolPositionCreatedCallCount)
	s.Require().Equal(expectedLastPositionWithdrawalListenerCallCount, mockListener.AfterLastPoolPositionRemovedCallCount)
	s.Require().Equal(expectedSwapListenerCallCount, mockListener.AfterConcentratedPoolSwapCallCount)
}

// setListenerMockOnConcentratedLiquidityKeeper injects the mock into the concentrated liquidity keeper
// so that listener invocation can be tested via the mock
func (s *KeeperTestSuite) setListenerMockOnConcentratedLiquidityKeeper() {
	s.App.ConcentratedLiquidityKeeper.SetListenersUnsafe(types.NewConcentratedLiquidityListeners(&clmocks.ConcentratedLiquidityListenerMock{}))
}

// Crosses the tick and charges the fee on the global spread reward accumulator.
// This mimics crossing an initialized tick during a swap and charging the fee on swap completion.
func (s *KeeperTestSuite) crossTickAndChargeSpreadReward(poolId uint64, tickIndexToCross int64) {
	nextTickInfo, err := s.clk.GetTickInfo(s.Ctx, poolId, tickIndexToCross)
	s.Require().NoError(err)

	feeAccum, uptimeAccums, err := s.clk.GetSwapAccumulators(s.Ctx, poolId)
	s.Require().NoError(err)

	// Cross the tick to update it.
	_, err = s.clk.CrossTick(s.Ctx, poolId, tickIndexToCross, &nextTickInfo, DefaultSpreadRewardAccumCoins[0], feeAccum.GetValue(), uptimeAccums)
	s.Require().NoError(err)
	s.AddToSpreadRewardAccumulator(poolId, DefaultSpreadRewardAccumCoins[0])
}

// AddToSpreadRewardAccumulator adds the given fee to pool by updating
// the internal per-pool accumulator that tracks fee growth per one unit of
// liquidity.
func (s *KeeperTestSuite) AddToSpreadRewardAccumulator(poolId uint64, feeUpdate sdk.DecCoin) {
	feeAccumulator, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, poolId)
	s.Require().NoError(err)
	feeAccumulator.AddToAccumulator(sdk.NewDecCoins(feeUpdate))
}

func (s *KeeperTestSuite) validatePositionSpreadRewardGrowth(poolId uint64, positionId uint64, expectedUnclaimedRewards sdk.DecCoins) {
	accum, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, poolId)
	s.Require().NoError(err)
	positionRecord, err := accum.GetPosition(types.KeySpreadRewardPositionAccumulator(positionId))
	s.Require().NoError(err)
	if expectedUnclaimedRewards.IsZero() {
		s.Require().Equal(expectedUnclaimedRewards, positionRecord.UnclaimedRewardsTotal)
	} else {
		s.Require().Equal(expectedUnclaimedRewards[0].Amount.Mul(DefaultLiquidityAmt), positionRecord.UnclaimedRewardsTotal.AmountOf(expectedUnclaimedRewards[0].Denom))
		if expectedUnclaimedRewards.Len() > 1 {
			s.Require().Equal(expectedUnclaimedRewards[1].Amount.Mul(DefaultLiquidityAmt), positionRecord.UnclaimedRewardsTotal.AmountOf(expectedUnclaimedRewards[1].Denom))
		}
	}
}

func (s *KeeperTestSuite) SetBlockTime(timeToSet time.Time) {
	s.Ctx = s.Ctx.WithBlockTime(timeToSet)
}

func (s *KeeperTestSuite) AddBlockTime(timeToAdd time.Duration) {
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(timeToAdd))
}

func (s *KeeperTestSuite) TestValidatePermissionlessPoolCreationEnabled() {
	s.SetupTest()
	// Normally, by default, permissionless pool creation is disabled.
	// SetupTest, however, calls SetupConcentratedLiquidityDenomsAndPoolCreation which enables permissionless pool creation.
	s.Require().NoError(s.App.ConcentratedLiquidityKeeper.ValidatePermissionlessPoolCreationEnabled(s.Ctx))

	// Disable permissionless pool creation.
	defaultParams := types.DefaultParams()
	defaultParams.IsPermissionlessPoolCreationEnabled = false
	s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, defaultParams)

	// Validate that permissionless pool creation is disabled.
	s.Require().Error(s.App.ConcentratedLiquidityKeeper.ValidatePermissionlessPoolCreationEnabled(s.Ctx))
}

// runFungifySetup Sets up a pool with `poolSpreadFactor`, prepares `numPositions` default positions on it (all identical), and sets
// up the passed in incentive records such that they emit on the pool. It also sets the largest authorized uptime to be `fullChargeDuration`.
//
// Returns the pool, expected position ids and the total liquidity created on the pool.
func (s *KeeperTestSuite) runFungifySetup(address sdk.AccAddress, numPositions int, fullChargeDuration time.Duration, poolSpreadFactor sdk.Dec, incentiveRecords []types.IncentiveRecord) (types.ConcentratedPoolExtension, []uint64, sdk.Dec) {
	expectedPositionIds := make([]uint64, numPositions)
	for i := 0; i < numPositions; i++ {
		expectedPositionIds[i] = uint64(i + 1)
	}

	s.TestAccs = apptesting.CreateRandomAccounts(5)
	s.SetBlockTime(defaultBlockTime)
	totalPositionsToCreate := sdk.NewInt(int64(numPositions))
	requiredBalances := sdk.NewCoins(sdk.NewCoin(ETH, DefaultAmt0.Mul(totalPositionsToCreate)), sdk.NewCoin(USDC, DefaultAmt1.Mul(totalPositionsToCreate)))

	// Set test authorized uptime params.
	params := s.clk.GetParams(s.Ctx)
	params.AuthorizedUptimes = []time.Duration{time.Nanosecond, fullChargeDuration}
	s.clk.SetParams(s.Ctx, params)

	// Fund account
	s.FundAcc(address, requiredBalances)

	// Create CL pool
	pool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, DefaultTickSpacing, poolSpreadFactor)

	// Set incentives for pool to ensure accumulators work correctly
	err := s.clk.SetMultipleIncentiveRecords(s.Ctx, incentiveRecords)
	s.Require().NoError(err)

	// Set up fully charged positions
	totalLiquidity := sdk.ZeroDec()
	for i := 0; i < numPositions; i++ {
		_, _, _, liquidityCreated, _, _, err := s.clk.CreatePosition(s.Ctx, defaultPoolId, address, DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
		s.Require().NoError(err)
		totalLiquidity = totalLiquidity.Add(liquidityCreated)
	}

	return pool, expectedPositionIds, totalLiquidity
}

func (s *KeeperTestSuite) runMultipleAuthorizedUptimes(tests func()) {
	authorizedUptimesTested := [][]time.Duration{
		DefaultAuthorizedUptimes,
		ThreeOrderedConsecutiveAuthorizedUptimes,
		ThreeUnorderedNonConsecutiveAuthorizedUptimes,
		AllUptimesAuthorized,
	}

	for _, curAuthorizedUptimes := range authorizedUptimesTested {
		s.authorizedUptimes = curAuthorizedUptimes
		tests()
	}
}

type RangeTestParams struct {
	baseAssets       sdk.Coins
	baseNumPositions int
	numSwapAddresses int
	baseSwapAmount   sdk.Int

	// If false, creates one address per position.
	// Useful for testing fungification.
	singleAddrPerRange                bool
	swapBefore                        bool
	swapAfter                         bool
	swapsBetweenJoins                 bool
	activeIncentives                  bool
	newActiveIncentivesBetweenJoins   bool
	newInactiveIncentivesBetweenJoins bool
	spacedOutJoins                    bool
	fuzzAssets                        bool
	fuzzNumPositions                  bool
	fuzzSwapAmounts                   bool
}

// runMultiplePositionRanges runs various test constructions and invariants on the given position ranges.
func (s *KeeperTestSuite) runMultiplePositionRanges(ranges [][]int64) {
	// Pool setup parameters to vary:
	// 1. Spread factor
	// 2. Incentive records (and all subfields)
	// 3. TestParams struct that we pass into vectors for fuzzing (default should run a single hardcoded set)
	//    * Should include a map of addresses to ranges (ideally could do multiple addr per range but this is harder)
	//    * Default case should be running two hardcoded vectors: one with single addr on all ranges, and one with k addr for k ranges
	//
	// Hard coded vectors (all run for N = 1, N = 2, and N = 37)
	// Case 1: N positions in every range, default liq amounts
	// Case 2: N positions in every range, non-default liq amounts
	// Case 3: N positions in every range, non-default liq amounts, spaced out by time elapsed
	// Case 4: N positions in every range, swaps in between
	// Case 5: N positions in every range, incentive record created & emitted in between
	// Case 6: N positions in every range, incentive record created but not emitted in between (later start time)

	// Invariant checks to run on cached context:
	// 1. Collecting fees on all positions yields expected total fees
	// 2. Collecting incentives on all positions yields expected total incentives
	// 3. The above two invariants hold even during intermediate test steps
	// 4. Join then exit a single position on each range yields same amount (within rounding tolerance)
	// 5. Swap in then out and vice versa yield same amounts minus fees (within rounding tolerance)
	// 6. Removing all positions drains pool address, fee address, and incentive address (even with rounding?)
	// 7. (more complex) At intermediate steps, creating a smaller position then adding to it yields the same end state
	// 8. (more complex) At intermediate steps, creating a larger position then removing from it yields the same end state
	//
	// func (s *KeeperTestSuite) runInvariants(ctx, poolId, positionIds) {}

	// Potential helpers to reference:
	// - ExecuteAndValidateSuccessfulIncentiveClaim
	// - AssertPositionsDoNotExist
	// - GetTotalAccruedRewardsByAccumulator

	// Preset seed to ensure deterministic test runs.
	rand.Seed(2)

	baseAssets := sdk.NewCoins(
		sdk.NewCoin(ETH, sdk.NewInt(5000000000)),
		sdk.NewCoin(USDC, sdk.NewInt(5000000000)),
	)

	defaultParams := RangeTestParams{
		baseNumPositions: 1,
		baseAssets:       baseAssets,
		numSwapAddresses: 1,
		baseSwapAmount:   sdk.NewInt(10000),
		// fuzzNumPositions: true,
		fuzzAssets: true,
		// singleAddrPerRange: true,
	}

	// TODO: make this pool custom with spread factor (fuzzed?)
	pool := s.PrepareConcentratedPool()

	fmt.Println("Current tick: ", pool.GetCurrentTick())

	s.SetupRanges(pool, ranges, defaultParams)
}

// SetupRanges takes in a set of tick ranges
func (s *KeeperTestSuite) SetupRanges(pool types.ConcentratedPoolExtension, ranges [][]int64, testParams RangeTestParams) {

	// binaryFlipOne := rand.Int() % 2
	// binaryFlipTwo := rand.Int() % 2

	// --- Parse test params ---

	// Prepare a slice tracking how many positions to create on each range.
	numPositionSlice, totalPositions := s.prepareNumPositionSlice(ranges, testParams.baseNumPositions, testParams.fuzzNumPositions)
	fmt.Println("numPositionSlice: ", numPositionSlice)

	// Prepare a slice tracking how many assets each position should have (tracked as Coins type).
	assetSlice := s.prepareAssetSlice(ranges, totalPositions, testParams.baseAssets, testParams.fuzzAssets)

	fmt.Println("assetSlice: ", assetSlice)
	// Prepare a slice tracking how much time should elapse after each position creation.
	// TODO: support time elapsing between joins (occasionally no time elapsed)

	// --- Set up addresses ---

	// -- Set up position accounts --
	var positionAddresses []sdk.AccAddress
	if testParams.singleAddrPerRange {
		positionAddresses = apptesting.CreateRandomAccounts(len(ranges))
	} else {
		positionAddresses = apptesting.CreateRandomAccounts(totalPositions)
	}

	// -- Set up swap accounts --
	// TODO: support swaps. Swap amounts should be fuzzed & based on total pool liquidity (e.g. 3-5% of total pool liquidity in the chosen direction)
	// Swap direction & inGivenOut/outGivenIn should be based on separate binary flips

	// Assert that there are a positive number of swap addresses if swaps are enabled
	s.Require().False(testParams.numSwapAddresses <= 0 && (testParams.swapBefore || testParams.swapAfter || testParams.swapsBetweenJoins), "Must have positive number of swap addresses if swaps are enabled")
	// Generate swap accounts
	swapAddresses := apptesting.CreateRandomAccounts(testParams.numSwapAddresses)

	// --- Incentive setup ---
	// TODO: support incentives

	// Set up incentive records.

	// --- Position setup ---

	// Loop over ranges and create positions, setting up behavior as determined by the slices set up above.
	// TODO: add better comments explaining this behavior
	lastVisitedBlockIndex := 0
	totalLiquidity := sdk.ZeroDec()
	allPositionIds := []uint64{}
	for i := range ranges {
		curBlock := 0
		startNumPositions := len(allPositionIds)
		for j := lastVisitedBlockIndex; j < lastVisitedBlockIndex+numPositionSlice[i]; j++ {
			var curAddr sdk.AccAddress
			if testParams.singleAddrPerRange {
				curAddr = positionAddresses[i]
			} else {
				curAddr = positionAddresses[j]
			}

			curAssets := assetSlice[j]
			// Setup position
			fmt.Println("first position assets and range: ", curAssets, ranges[i][0], ranges[i][1])
			curLiquidity, curPositionId := s.SetupPosition(pool.GetId(), curAddr, curAssets, ranges[i][0], ranges[i][1], true)
			fmt.Println("owner balances after liq: ", s.App.BankKeeper.GetAllBalances(s.Ctx, curAddr))
			// curTimeElapsed := curTimeElapsedMap[j]
			fmt.Println("first position addr index and ID: ", j, curPositionId)

			pool, err := s.clk.GetPoolById(s.Ctx, pool.GetId())
			s.Require().NoError(err)
			poolLiquidity := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())
			fmt.Println("Current tick before swap: ", pool.GetCurrentTick())
			fmt.Println("poolLiquidity before swap: ", poolLiquidity)

			s.executeRandomizedSwap(pool, swapAddresses[0], testParams.baseSwapAmount, testParams.fuzzSwapAmounts)

			// amt0Withdrawn, amt1Withdrawn, err := s.clk.WithdrawPosition(s.Ctx, curAddr, curPositionId, curLiquidity)
			// s.Require().NoError(err)
			// fmt.Println("withdrawn asset amounts: ", amt0Withdrawn, amt1Withdrawn)

			pool, err = s.clk.GetPoolById(s.Ctx, pool.GetId())
			s.Require().NoError(err)
			poolLiquidity = s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())
			fmt.Println("Current tick after swap: ", pool.GetCurrentTick())
			_, updatedCurTickSqrtPrice, _ := math.TickToSqrtPrice(pool.GetCurrentTick())
			fmt.Println("current tick/bucket's sqrt price: ", updatedCurTickSqrtPrice)
			fmt.Println("pool's current sqrt price: ", pool.GetCurrentSqrtPrice())
			fmt.Println("poolLiquidity after swap: ", poolLiquidity)
			// Track new position values in global variables
			totalLiquidity = totalLiquidity.Add(curLiquidity)
			allPositionIds = append(allPositionIds, curPositionId)
			curBlock++
		}
		endNumPositions := len(allPositionIds)

		// Ensure the correct number of positions were set up in current range
		s.Require().Equal(numPositionSlice[i], endNumPositions-startNumPositions, "Incorrect number of positions set up in range %d", i)

		lastVisitedBlockIndex += curBlock
	}

	// Ensure that the correct number of positions were set up globally
	s.Require().Equal(totalPositions, len(allPositionIds))

	// Get pool assets
	poolAssets := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())
	fmt.Println("poolAssets: ", poolAssets)
	pool, err := s.clk.GetPoolById(s.Ctx, pool.GetId())
	s.Require().NoError(err)
	fmt.Println("Current tick: ", pool.GetCurrentTick())

	fmt.Println("Last visited index: ", lastVisitedBlockIndex)

	fmt.Println("Values: \n", numPositionSlice, "Length: ", len(numPositionSlice), "\n", positionAddresses, "\n", swapAddresses)
}

// numPositionSlice prepares a slice tracking the number of positions to create on each range, fuzzing the number at each step if applicable.
// Returns a slice representing the number of positions for each range index.
func (s *KeeperTestSuite) prepareNumPositionSlice(ranges [][]int64, baseNumPositions int, fuzzNumPositions bool) ([]int, int) {
	// Create slice representing number of positions for each range index.
	// Default case is `numPositions` on each range unless fuzzing is turned on.
	numPositionsPerRange := make([]int, len(ranges))
	totalPositions := 0

	// Loop through each range and set number of positions, fuzzing if applicable.
	for i := range ranges {
		numPositionsPerRange[i] = baseNumPositions

		// If applicable, fuzz the number of positions on current range
		if fuzzNumPositions {
			// Fuzzed amount should be between 1 and (2 * numPositions) + 1 (up to 100% fuzz both ways from numPositions)
			numPositionsPerRange[i] = (rand.Int() % (2 * baseNumPositions)) + 1
		}

		// Track total positions
		totalPositions += numPositionsPerRange[i]
	}

	return numPositionsPerRange, totalPositions
}

// prepareAssetSlice prepares a slice tracking the assets for each position, fuzzing the amount at each step if applicable.
func (s *KeeperTestSuite) prepareAssetSlice(ranges [][]int64, totalPositions int, baseAssets sdk.Coins, fuzzAssets bool) []sdk.Coins {
	// Create slice representing assets for each position.
	// Default case is `baseAssets` on each range unless fuzzing is turned on.
	assetsByPosition := make([]sdk.Coins, totalPositions)
	totalAssets := sdk.NewCoins()

	// Loop through each range and set number of positions, fuzzing if applicable.
	for i := 0; i < totalPositions; i++ {
		assetsByPosition[i] = baseAssets

		// If applicable, fuzz the number of positions on current range
		if fuzzAssets {
			fuzzedAssets := make([]sdk.Coin, len(baseAssets))
			for coinIndex, coin := range baseAssets {
				// Fuzz +/- 100% of current amount
				newAmount := (rand.Int63() % (2 * coin.Amount.Int64())) + 1
				fuzzedAssets[coinIndex] = sdk.NewCoin(coin.Denom, sdk.NewInt(newAmount))
			}

			assetsByPosition[i] = fuzzedAssets
		}

		// Track total positions
		totalAssets = totalAssets.Add(assetsByPosition[i]...)
	}

	return assetsByPosition
}

func (s *KeeperTestSuite) executeRandomizedSwap(pool types.ConcentratedPoolExtension, swapAddress sdk.AccAddress, baseSwapAmount sdk.Int, fuzzSwap bool) {
	binaryFlip := rand.Int() % 2
	poolLiquidity := s.App.BankKeeper.GetAllBalances(s.Ctx, pool.GetAddress())
	s.Require().True(len(poolLiquidity) == 1 || len(poolLiquidity) == 2, "Pool liquidity should be in one or two tokens")

	// Decide which denom to swap in & out

	var swapInDenom, swapOutDenom string
	if len(poolLiquidity) == 1 {
		// If all pool liquidity is in one token, swap in the other token
		swapOutDenom = poolLiquidity[0].Denom
		if swapOutDenom == pool.GetToken0() {
			swapInDenom = pool.GetToken1()
		} else {
			swapInDenom = pool.GetToken0()
		}
	} else {
		// Otherwise, randomly determine which denom to swap in & out
		if binaryFlip == 0 {
			swapInDenom = pool.GetToken0()
			swapOutDenom = pool.GetToken1()
		} else {
			swapInDenom = pool.GetToken1()
			swapOutDenom = pool.GetToken0()
		}
	}

	// TODO: decide which swap function to use
	// TODO: fuzz swap amounts

	swapInFunded := sdk.NewCoin(swapInDenom, sdk.Int(sdk.MustNewDecFromStr("10000000000000000000000")))
	s.FundAcc(swapAddress, sdk.NewCoins(swapInFunded))

	swapOutCoin := sdk.NewCoin(swapOutDenom, sdk.MinInt(baseSwapAmount, poolLiquidity.AmountOf(swapOutDenom).ToDec().Mul(sdk.MustNewDecFromStr("0.5")).TruncateInt()))
	fmt.Println("remaining swap out amount: ", poolLiquidity.AmountOf(swapOutDenom))
	fmt.Println("asset 0: ", pool.GetToken0())
	fmt.Println("swapInDenom: ", swapInDenom)
	fmt.Println("swapOutCoin: ", swapOutCoin)
	fmt.Println("poolLiquidity: ", poolLiquidity)
	allTicks, _ := s.clk.GetAllInitializedTicksForPool(s.Ctx, pool.GetId())
	fmt.Println("All initialized ticks before next swap: ", allTicks)
	// Note that we set the price limit to zero to ensure that the swap can execute in either direction (gets automatically set to correct limit)
	fmt.Println("------ENTERING SWAP------")
	fmt.Println("swap fields: ", swapOutCoin, swapInDenom, pool.GetSpreadFactor(s.Ctx), sdk.ZeroDec())
	_, _, _, _, _, err := s.clk.SwapInAmtGivenOut(s.Ctx, swapAddress, pool, swapOutCoin, swapInDenom, pool.GetSpreadFactor(s.Ctx), sdk.ZeroDec())
	s.Require().NoError(err)
	fmt.Println("------EXITING SWAP------")
}
