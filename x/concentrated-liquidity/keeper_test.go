package concentrated_liquidity_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	concentrated_liquidity "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/clmocks"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"

	cl "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
)

var (
	DefaultMinTick, DefaultMaxTick       = types.MinInitializedTick, types.MaxTick
	DefaultMinCurrentTick                = types.MinCurrentTick
	DefaultLowerPrice                    = apptesting.DefaultLowerPrice
	DefaultLowerTick                     = apptesting.DefaultLowerTick
	DefaultUpperPrice                    = apptesting.DefaultUpperPrice
	DefaultUpperTick                     = apptesting.DefaultUpperTick
	DefaultCurrPrice                     = apptesting.DefaultCurrPrice
	DefaultCurrTick                int64 = apptesting.DefaultCurrTick
	DefaultCurrSqrtPrice                 = apptesting.DefaultCurrSqrtPrice

	DefaultZeroSpreadFactor       = osmomath.ZeroDec()
	DefaultSpreadRewardAccumCoins = apptesting.DefaultSpreadRewardAccumCoins
	DefaultPositionId             = uint64(1)
	DefaultUnderlyingLockId       = uint64(0)
	DefaultJoinTime               = time.Unix(0, 0).UTC()
	ETH                           = apptesting.ETH
	DefaultAmt0                   = apptesting.DefaultAmt0
	DefaultAmt0Expected           = apptesting.DefaultAmt0Expected
	DefaultCoin0                  = apptesting.DefaultCoin0
	USDC                          = apptesting.USDC
	DefaultAmt1                   = apptesting.DefaultAmt1
	DefaultAmt1Expected           = apptesting.DefaultAmt1Expected
	DefaultCoin1                  = apptesting.DefaultCoin1
	DefaultCoins                  = apptesting.DefaultCoins

	// Both of the following liquidity values are calculated in x/concentrated-liquidity/python/swap_test.py
	DefaultLiquidityAmt   = apptesting.DefaultLiquidityAmt
	FullRangeLiquidityAmt = apptesting.FullRangeLiquidityAmt

	DefaultTickSpacing                             = uint64(100)
	PoolCreationFee                                = poolmanagertypes.DefaultParams().PoolCreationFee
	sqrt4000                                       = osmomath.MustNewDecFromStr("63.245553203367586640")
	sqrt4994                                       = apptesting.Sqrt4994
	sqrt4999                                       = osmomath.MustNewDecFromStr("70.703606697254136613")
	sqrt5500                                       = osmomath.MustNewDecFromStr("74.161984870956629488")
	sqrt6250                                       = osmomath.MustNewDecFromStr("79.056941504209483300")
	DefaultExponentConsecutivePositionLowerTick, _ = math.SqrtPriceToTickRoundDownSpacing(osmomath.BigDecFromDec(sqrt5500), DefaultTickSpacing)
	DefaultExponentConsecutivePositionUpperTick, _ = math.SqrtPriceToTickRoundDownSpacing(osmomath.BigDecFromDec(sqrt6250), DefaultTickSpacing)
	DefaultExponentOverlappingPositionLowerTick, _ = math.SqrtPriceToTickRoundDownSpacing(osmomath.BigDecFromDec(sqrt4000), DefaultTickSpacing)
	DefaultExponentOverlappingPositionUpperTick, _ = math.SqrtPriceToTickRoundDownSpacing(osmomath.BigDecFromDec(sqrt4999), DefaultTickSpacing)
	BAR                                            = "bar"
	FOO                                            = "foo"
	InsufficientFundsError                         = fmt.Errorf("insufficient funds")
	DefaultAuthorizedUptimes                       = []time.Duration{time.Nanosecond}
	ThreeOrderedConsecutiveAuthorizedUptimes       = []time.Duration{time.Nanosecond, time.Minute, time.Hour, time.Hour * 24}
	ThreeUnorderedNonConsecutiveAuthorizedUptimes  = []time.Duration{time.Nanosecond, time.Hour * 24 * 7, time.Minute}
	AllUptimesAuthorized                           = types.SupportedUptimes
)

func TestConstants(t *testing.T) {
	lowerSqrtPrice, _ := osmomath.MonotonicSqrt(DefaultLowerPrice)
	upperSqrtPrice, _ := osmomath.MonotonicSqrt(DefaultUpperPrice)
	liq := math.GetLiquidityFromAmounts(DefaultCurrSqrtPrice,
		osmomath.BigDecFromDec(lowerSqrtPrice), osmomath.BigDecFromDec(upperSqrtPrice), DefaultAmt0, DefaultAmt1)
	require.Equal(t, DefaultLiquidityAmt, liq)
}

type FuzzTestSuite struct {
	positionData    []positionAndLiquidity
	iteration       int
	seed            int64
	collectedErrors []error
}

type positionAndLiquidity struct {
	positionId   uint64
	liquidity    osmomath.Dec
	accountIndex int
}

type KeeperTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
	clk *concentrated_liquidity.Keeper
	FuzzTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupDefaultPosition(poolId uint64) {
	s.SetupPosition(poolId, s.TestAccs[0], DefaultCoins, DefaultLowerTick, DefaultUpperTick, false)
}

func (s *KeeperTestSuite) SetupPosition(poolId uint64, owner sdk.AccAddress, providedCoins sdk.Coins, lowerTick, upperTick int64, addRoundingError bool) (osmomath.Dec, uint64) {
	roundingErrorCoins := sdk.NewCoins()
	if addRoundingError {
		roundingErrorCoins = sdk.NewCoins(sdk.NewCoin(ETH, roundingError), sdk.NewCoin(USDC, roundingError))
	}

	s.FundAcc(owner, providedCoins.Add(roundingErrorCoins...))
	positionData, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, poolId, owner, providedCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), lowerTick, upperTick)
	s.Require().NoError(err)
	liquidity, err := s.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(s.Ctx, positionData.ID)
	s.Require().NoError(err)
	return liquidity, positionData.ID
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
func (s *KeeperTestSuite) validatePositionUpdate(ctx sdk.Context, positionId uint64, expectedRemainingLiquidity osmomath.Dec) {
	newPositionLiquidity, err := s.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(ctx, positionId)
	s.Require().NoError(err)
	s.Require().Equal(expectedRemainingLiquidity.String(), newPositionLiquidity.String())
	s.Require().True(newPositionLiquidity.GTE(osmomath.ZeroDec()))
}

// validateTickUpdates validates that ticks with the given parameters have expectedRemainingLiquidity left.
func (s *KeeperTestSuite) validateTickUpdates(poolId uint64, lowerTick int64, upperTick int64, expectedRemainingLiquidity osmomath.Dec, expectedLowerSpreadRewardGrowthOppositeDirectionOfLastTraversal, expectedUpperSpreadRewardGrowthOppositeDirectionOfLastTraversal sdk.DecCoins) {
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

func (s *KeeperTestSuite) initializeTick(ctx sdk.Context, tickIndex int64, initialLiquidity osmomath.Dec, spreadRewardGrowthOppositeDirectionOfTraversal sdk.DecCoins, uptimeTrackers []model.UptimeTracker, isLower bool) {
	_, err := s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(ctx, validPoolId, tickIndex, initialLiquidity, isLower)
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
func (s *KeeperTestSuite) initializeSpreadRewardAccumulatorPositionWithLiquidity(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64, positionId uint64, liquidity osmomath.Dec) {
	err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePositionSpreadRewardAccumulator(ctx, poolId, lowerTick, upperTick, positionId, liquidity)
	s.Require().NoError(err)
}

// addLiquidityToUptimeAccumulators adds shares to all uptime accumulators as defined by the `liquidity` parameter.
// This helper is primarily used to test incentive accrual for specific tick ranges, so we pass in filler values
// for all other components (e.g. join time).
func (s *KeeperTestSuite) addLiquidityToUptimeAccumulators(ctx sdk.Context, poolId uint64, liquidity []osmomath.Dec, positionId uint64) {
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
func (s *KeeperTestSuite) addUptimeGrowthInsideRange(ctx sdk.Context, poolId uint64, currentTick, lowerTick, upperTick int64, uptimeGrowthToAdd []sdk.DecCoins) {
	s.Require().True(lowerTick <= upperTick)

	// Add spread rewards and incentives to the pool
	scaledGrowth := s.scaleUptimeAccumulators(uptimeGrowthToAdd)

	// Note that we process adds to global accums at the end to ensure that they don't affect the behavior of uninitialized ticks.
	if currentTick < lowerTick {
		// Add to lower tick's uptime trackers
		lowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(ctx, poolId, lowerTick)
		s.Require().NoError(err)
		s.Require().Equal(len(lowerTickInfo.UptimeTrackers.List), len(scaledGrowth))

		newLowerUptimeTrackerValues, err := osmoutils.AddDecCoinArrays(cl.GetUptimeTrackerValues(lowerTickInfo.UptimeTrackers.List), scaledGrowth)
		s.Require().NoError(err)

		s.initializeTick(ctx, lowerTick, lowerTickInfo.LiquidityGross, lowerTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal, wrapUptimeTrackers(newLowerUptimeTrackerValues), true)
	} else if upperTick <= currentTick {
		// Add to upper tick uptime trackers
		upperTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(ctx, poolId, upperTick)
		s.Require().NoError(err)
		s.Require().Equal(len(upperTickInfo.UptimeTrackers.List), len(scaledGrowth))

		newUpperUptimeTrackerValues, err := osmoutils.AddDecCoinArrays(cl.GetUptimeTrackerValues(upperTickInfo.UptimeTrackers.List), scaledGrowth)
		s.Require().NoError(err)

		s.initializeTick(ctx, upperTick, upperTickInfo.LiquidityGross, upperTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal, wrapUptimeTrackers(newUpperUptimeTrackerValues), false)
	}

	// In all cases, global uptime accums need to be updated. If lowerTick <= currentTick < upperTick,
	// nothing more needs to be done.
	// We do not provide scaled growth here as addToUptimeAccums will scale the growth by the scaling factor.
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
func (s *KeeperTestSuite) addUptimeGrowthOutsideRange(ctx sdk.Context, poolId uint64, currentTick, lowerTick, upperTick int64, uptimeGrowthToAdd []sdk.DecCoins) {
	s.Require().True(lowerTick <= upperTick)

	// Add spread rewards and incentives to the pool
	growthCopy := s.scaleUptimeAccumulators(uptimeGrowthToAdd)

	// Note that we process adds to global accums at the end to ensure that they don't affect the behavior of uninitialized ticks.
	if currentTick < lowerTick || upperTick <= currentTick {
		// Add to lower tick uptime trackers
		lowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(ctx, poolId, lowerTick)
		s.Require().NoError(err)
		s.Require().Equal(len(lowerTickInfo.UptimeTrackers.List), len(growthCopy))

		newLowerUptimeTrackerValues, err := osmoutils.AddDecCoinArrays(cl.GetUptimeTrackerValues(lowerTickInfo.UptimeTrackers.List), growthCopy)
		s.Require().NoError(err)

		s.initializeTick(ctx, lowerTick, lowerTickInfo.LiquidityGross, lowerTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal, wrapUptimeTrackers(newLowerUptimeTrackerValues), true)

		// Add to upper tick uptime trackers
		upperTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(ctx, poolId, upperTick)
		s.Require().NoError(err)
		s.Require().Equal(len(upperTickInfo.UptimeTrackers.List), len(growthCopy))

		newUpperUptimeTrackerValues, err := osmoutils.AddDecCoinArrays(cl.GetUptimeTrackerValues(upperTickInfo.UptimeTrackers.List), growthCopy)
		s.Require().NoError(err)

		s.initializeTick(ctx, upperTick, upperTickInfo.LiquidityGross, upperTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal, wrapUptimeTrackers(newUpperUptimeTrackerValues), false)
	} else if currentTick < upperTick {
		// Add to lower tick's uptime trackers
		lowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(ctx, poolId, lowerTick)
		s.Require().NoError(err)
		s.Require().Equal(len(lowerTickInfo.UptimeTrackers.List), len(growthCopy))

		newLowerUptimeTrackerValues, err := osmoutils.AddDecCoinArrays(cl.GetUptimeTrackerValues(lowerTickInfo.UptimeTrackers.List), growthCopy)
		s.Require().NoError(err)

		s.initializeTick(ctx, lowerTick, lowerTickInfo.LiquidityGross, lowerTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal, wrapUptimeTrackers(newLowerUptimeTrackerValues), true)
	}

	// In all cases, global uptime accums need to be updated. If currentTick < lowerTick,
	// nothing more needs to be done.
	// Note that addToUptimeAccums(...) applies the scaling factor to the input before writing to the accumulator.
	err := addToUptimeAccums(ctx, poolId, s.App.ConcentratedLiquidityKeeper, uptimeGrowthToAdd)
	s.Require().NoError(err)
}

// validatePositionSpreadFactorAccUpdate validates that the position's accumulator with given parameters
// has been updated with liquidity.
func (s *KeeperTestSuite) validatePositionSpreadRewardAccUpdate(ctx sdk.Context, poolId uint64, positionId uint64, liquidity osmomath.Dec) {
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
	nextTickInfo, err := s.Clk.GetTickInfo(s.Ctx, poolId, tickIndexToCross)
	s.Require().NoError(err)

	feeAccum, err := s.Clk.GetSpreadRewardAccumulator(s.Ctx, poolId)
	s.Require().NoError(err)
	uptimeAccums, err := s.Clk.GetUptimeAccumulators(s.Ctx, poolId)
	s.Require().NoError(err)

	// Cross the tick to update it.
	err = s.Clk.CrossTick(s.Ctx, poolId, tickIndexToCross, &nextTickInfo, DefaultSpreadRewardAccumCoins[0], feeAccum.GetValue(), uptimeAccums)
	s.Require().NoError(err)
	s.AddToSpreadRewardAccumulator(poolId, DefaultSpreadRewardAccumCoins[0])
}

// AddToSpreadRewardAccumulator adds the given fee to pool by updating
// the internal per-pool accumulator that tracks fee growth per one unit of
// liquidity.
//
// NOTE: This method takes the unscaled fee amount and scales it by the liquidity scaling factor.
func (s *KeeperTestSuite) AddToSpreadRewardAccumulator(poolId uint64, feeUpdate sdk.DecCoin) {
	feeAccumulator, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, poolId)
	s.Require().NoError(err)
	feeUpdate.Amount = feeUpdate.Amount.MulTruncate((cl.PerUnitLiqScalingFactor))
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

func (s *KeeperTestSuite) TestIsPermissionlessPoolCreationEnabled() {
	// Permissionless pool creation is enabled in tests by default
	s.Require().True(s.App.ConcentratedLiquidityKeeper.IsPermissionlessPoolCreationEnabled(s.Ctx))

	// Disable permissionless pool creation.
	defaultParams := types.DefaultParams()
	defaultParams.IsPermissionlessPoolCreationEnabled = false
	s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, defaultParams)

	// Validate that permissionless pool creation is disabled.
	s.Require().False(s.App.ConcentratedLiquidityKeeper.IsPermissionlessPoolCreationEnabled(s.Ctx))
}

func (s *KeeperTestSuite) runMultipleAuthorizedUptimes(tests func()) {
	authorizedUptimesTested := [][]time.Duration{
		DefaultAuthorizedUptimes,
		ThreeOrderedConsecutiveAuthorizedUptimes,
		ThreeUnorderedNonConsecutiveAuthorizedUptimes,
		AllUptimesAuthorized,
	}

	for _, curAuthorizedUptimes := range authorizedUptimesTested {
		s.AuthorizedUptimes = curAuthorizedUptimes
		tests()
	}
}

// runFungifySetup Sets up a pool with `poolSpreadFactor`, prepares `numPositions` default positions on it (all identical), and sets
// up the passed in incentive records such that they emit on the pool. It also sets the largest authorized uptime to be `fullChargeDuration`.
//
// Returns the pool, expected position ids and the total liquidity created on the pool.
func (s *KeeperTestSuite) runFungifySetup(address sdk.AccAddress, numPositions int, fullChargeDuration time.Duration, poolSpreadFactor osmomath.Dec, incentiveRecords []types.IncentiveRecord) (types.ConcentratedPoolExtension, []uint64, osmomath.Dec) {
	expectedPositionIds := make([]uint64, numPositions)
	for i := 0; i < numPositions; i++ {
		expectedPositionIds[i] = uint64(i + 1)
	}

	s.TestAccs = apptesting.CreateRandomAccounts(5)
	s.SetBlockTime(defaultBlockTime)
	totalPositionsToCreate := osmomath.NewInt(int64(numPositions))
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
	totalLiquidity := osmomath.ZeroDec()
	for i := 0; i < numPositions; i++ {
		positionData, err := s.clk.CreatePosition(s.Ctx, defaultPoolId, address, DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), DefaultLowerTick, DefaultUpperTick)
		s.Require().NoError(err)
		totalLiquidity = totalLiquidity.Add(positionData.Liquidity)
	}

	return pool, expectedPositionIds, totalLiquidity
}

// runMultiplePositionRanges runs various test constructions and invariants on the given position ranges.
func (s *KeeperTestSuite) runMultiplePositionRanges(ranges [][]int64, rangeTestParams RangeTestParams) {
	// Preset seed to ensure deterministic test runs.
	rand.Seed(2)

	// TODO: add pool-related fuzz params (spread factor & number of pools)
	pool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, rangeTestParams.tickSpacing, rangeTestParams.spreadFactor)

	// Run full state determined by params while asserting invariants at each intermediate step
	s.setupRangesAndAssertInvariants(pool, ranges, rangeTestParams)

	// Assert global invariants on final state
	s.assertGlobalInvariants(ExpectedGlobalRewardValues{})
}

// validates that the given position ID is eligible for claiming given expected incentives.
func (s *KeeperTestSuite) validateClaimableIncentives(positionID uint64, expectedClaimableIncentives sdk.Coins) {
	claimableIncentives, _, err := s.App.ConcentratedLiquidityKeeper.GetClaimableIncentives(s.Ctx, positionID)
	s.Require().NoError(err)
	s.Require().Equal(expectedClaimableIncentives.String(), claimableIncentives.String())
}

// validates that the given position ID has the given expected accumulator growth in the pool accumulator.
func (s *KeeperTestSuite) validateUptimePositionAccumulator(uptimeAccumulator *accum.AccumulatorObject, positionID uint64, expectedAccumulatorGrowth sdk.DecCoins) {
	positionAcc, err := uptimeAccumulator.GetPosition(string(types.KeyPositionId(positionID)))
	s.Require().NoError(err)
	s.Require().Equal(expectedAccumulatorGrowth.String(), positionAcc.GetAccumValuePerShare().String())
}
