package concentrated_liquidity_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/clmocks"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"

	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
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
	DefaultZeroSwapFee                                   = sdk.ZeroDec()
	DefaultFeeAccumCoins                                 = sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(50)))
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
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
}

func (s *KeeperTestSuite) SetupDefaultPosition(poolId uint64) {
	s.SetupPosition(poolId, s.TestAccs[0], DefaultCoins, DefaultLowerTick, DefaultUpperTick, s.Ctx.BlockTime())
}

func (s *KeeperTestSuite) SetupPosition(poolId uint64, owner sdk.AccAddress, providedCoins sdk.Coins, lowerTick, upperTick int64, joinTime time.Time) (sdk.Dec, uint64) {
	s.FundAcc(owner, providedCoins)
	positionId, _, _, _, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, poolId, owner, providedCoins, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick, upperTick)
	s.Require().NoError(err)
	liquidity, err := s.App.ConcentratedLiquidityKeeper.GetPositionLiquidity(s.Ctx, positionId)
	s.Require().NoError(err)
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
	_, positionId := s.SetupPosition(poolId, owner, DefaultCoins, DefaultLowerTick, DefaultUpperTick, s.Ctx.BlockTime())
	return positionId
}

func (s *KeeperTestSuite) SetupFullRangePositionAcc(poolId uint64, owner sdk.AccAddress) uint64 {
	_, positionId := s.SetupPosition(poolId, owner, DefaultCoins, DefaultMinTick, DefaultMaxTick, s.Ctx.BlockTime())
	return positionId
}

func (s *KeeperTestSuite) SetupConsecutiveRangePositionAcc(poolId uint64, owner sdk.AccAddress) uint64 {
	_, positionId := s.SetupPosition(poolId, owner, DefaultCoins, DefaultExponentConsecutivePositionLowerTick, DefaultExponentConsecutivePositionUpperTick, s.Ctx.BlockTime())
	return positionId
}

func (s *KeeperTestSuite) SetupOverlappingRangePositionAcc(poolId uint64, owner sdk.AccAddress) uint64 {
	_, positionId := s.SetupPosition(poolId, owner, DefaultCoins, DefaultExponentOverlappingPositionLowerTick, DefaultExponentOverlappingPositionUpperTick, s.Ctx.BlockTime())
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
func (s *KeeperTestSuite) validateTickUpdates(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick int64, upperTick int64, expectedRemainingLiquidity sdk.Dec, expectedLowerFeeGrowthOppositeDirectionOfLastTraversal, expectedUpperFeeGrowthOppositeDirectionOfLastTraversal sdk.DecCoins) {
	lowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, poolId, lowerTick)
	s.Require().NoError(err)
	s.Require().Equal(expectedRemainingLiquidity.String(), lowerTickInfo.LiquidityGross.String())
	s.Require().Equal(expectedRemainingLiquidity.String(), lowerTickInfo.LiquidityNet.String())
	s.Require().Equal(expectedLowerFeeGrowthOppositeDirectionOfLastTraversal.String(), lowerTickInfo.FeeGrowthOppositeDirectionOfLastTraversal.String())

	upperTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, poolId, upperTick)
	s.Require().NoError(err)
	s.Require().Equal(expectedRemainingLiquidity.String(), upperTickInfo.LiquidityGross.String())
	s.Require().Equal(expectedRemainingLiquidity.Neg().String(), upperTickInfo.LiquidityNet.String())
	s.Require().Equal(expectedUpperFeeGrowthOppositeDirectionOfLastTraversal.String(), upperTickInfo.FeeGrowthOppositeDirectionOfLastTraversal.String())
}

func (s *KeeperTestSuite) initializeTick(ctx sdk.Context, currentTick int64, tickIndex int64, initialLiquidity sdk.Dec, feeGrowthOppositeDirectionOfTraversal sdk.DecCoins, uptimeTrackers []model.UptimeTracker, isLower bool) {
	err := s.App.ConcentratedLiquidityKeeper.InitOrUpdateTick(ctx, validPoolId, currentTick, tickIndex, initialLiquidity, isLower)
	s.Require().NoError(err)

	tickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(ctx, validPoolId, tickIndex)
	s.Require().NoError(err)

	tickInfo.FeeGrowthOppositeDirectionOfLastTraversal = feeGrowthOppositeDirectionOfTraversal
	tickInfo.UptimeTrackers = uptimeTrackers

	s.App.ConcentratedLiquidityKeeper.SetTickInfo(ctx, validPoolId, tickIndex, tickInfo)
}

// initializeFeeAccumulatorPositionWithLiquidity initializes fee accumulator position with given parameters and updates it with given liquidity.
func (s *KeeperTestSuite) initializeFeeAccumulatorPositionWithLiquidity(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64, positionId uint64, liquidity sdk.Dec) {
	err := s.App.ConcentratedLiquidityKeeper.InitOrUpdatePositionFeeAccumulator(ctx, poolId, lowerTick, upperTick, positionId, liquidity)
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
		s.Require().Equal(len(lowerTickInfo.UptimeTrackers), len(uptimeGrowthToAdd))

		newLowerUptimeTrackerValues, err := addDecCoinsArray(cl.GetUptimeTrackerValues(lowerTickInfo.UptimeTrackers), uptimeGrowthToAdd)
		s.Require().NoError(err)

		s.initializeTick(ctx, currentTick, lowerTick, lowerTickInfo.LiquidityGross, lowerTickInfo.FeeGrowthOppositeDirectionOfLastTraversal, wrapUptimeTrackers(newLowerUptimeTrackerValues), true)
	} else if upperTick <= currentTick {
		// Add to upper tick uptime trackers
		upperTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(ctx, poolId, upperTick)
		s.Require().NoError(err)
		s.Require().Equal(len(upperTickInfo.UptimeTrackers), len(uptimeGrowthToAdd))

		newUpperUptimeTrackerValues, err := addDecCoinsArray(cl.GetUptimeTrackerValues(upperTickInfo.UptimeTrackers), uptimeGrowthToAdd)
		s.Require().NoError(err)

		s.initializeTick(ctx, currentTick, upperTick, upperTickInfo.LiquidityGross, upperTickInfo.FeeGrowthOppositeDirectionOfLastTraversal, wrapUptimeTrackers(newUpperUptimeTrackerValues), false)
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
		s.Require().Equal(len(lowerTickInfo.UptimeTrackers), len(uptimeGrowthToAdd))

		newLowerUptimeTrackerValues, err := addDecCoinsArray(cl.GetUptimeTrackerValues(lowerTickInfo.UptimeTrackers), uptimeGrowthToAdd)
		s.Require().NoError(err)

		s.initializeTick(ctx, currentTick, lowerTick, lowerTickInfo.LiquidityGross, lowerTickInfo.FeeGrowthOppositeDirectionOfLastTraversal, wrapUptimeTrackers(newLowerUptimeTrackerValues), true)

		// Add to upper tick uptime trackers
		upperTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(ctx, poolId, upperTick)
		s.Require().NoError(err)
		s.Require().Equal(len(upperTickInfo.UptimeTrackers), len(uptimeGrowthToAdd))

		newUpperUptimeTrackerValues, err := addDecCoinsArray(cl.GetUptimeTrackerValues(upperTickInfo.UptimeTrackers), uptimeGrowthToAdd)
		s.Require().NoError(err)

		s.initializeTick(ctx, currentTick, upperTick, upperTickInfo.LiquidityGross, upperTickInfo.FeeGrowthOppositeDirectionOfLastTraversal, wrapUptimeTrackers(newUpperUptimeTrackerValues), false)
	} else if currentTick < upperTick {
		// Add to lower tick's uptime trackers
		lowerTickInfo, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(ctx, poolId, lowerTick)
		s.Require().NoError(err)
		s.Require().Equal(len(lowerTickInfo.UptimeTrackers), len(uptimeGrowthToAdd))

		newLowerUptimeTrackerValues, err := addDecCoinsArray(cl.GetUptimeTrackerValues(lowerTickInfo.UptimeTrackers), uptimeGrowthToAdd)
		s.Require().NoError(err)

		s.initializeTick(ctx, currentTick, lowerTick, lowerTickInfo.LiquidityGross, lowerTickInfo.FeeGrowthOppositeDirectionOfLastTraversal, wrapUptimeTrackers(newLowerUptimeTrackerValues), true)
	}

	// In all cases, global uptime accums need to be updated. If currentTick < lowerTick,
	// nothing more needs to be done.
	err := addToUptimeAccums(ctx, poolId, s.App.ConcentratedLiquidityKeeper, uptimeGrowthToAdd)
	s.Require().NoError(err)
}

// validatePositionFeeAccUpdate validates that the position's accumulator with given parameters
// has been updated with liquidity.
func (s *KeeperTestSuite) validatePositionFeeAccUpdate(ctx sdk.Context, poolId uint64, positionId uint64, liquidity sdk.Dec) {
	accum, err := s.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(ctx, poolId)
	s.Require().NoError(err)

	accumulatorPosition, err := accum.GetPositionSize(types.KeyFeePositionAccumulator(positionId))
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

// Crosses the tick and charges the fee on the global fee accumulator.
// This mimics crossing an initialized tick during a swap and charging the fee on swap completion.
func (s *KeeperTestSuite) crossTickAndChargeFee(poolId uint64, tickIndexToCross int64) {
	// Cross the tick to update it.
	_, err := s.App.ConcentratedLiquidityKeeper.CrossTick(s.Ctx, poolId, tickIndexToCross, DefaultFeeAccumCoins[0])
	s.Require().NoError(err)
	err = s.App.ConcentratedLiquidityKeeper.ChargeFee(s.Ctx, poolId, DefaultFeeAccumCoins[0])
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) validatePositionFeeGrowth(poolId uint64, positionId uint64, expectedUnclaimedRewards sdk.DecCoins) {
	accum, err := s.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(s.Ctx, poolId)
	s.Require().NoError(err)
	positionRecord, err := accum.GetPosition(types.KeyFeePositionAccumulator(positionId))
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
