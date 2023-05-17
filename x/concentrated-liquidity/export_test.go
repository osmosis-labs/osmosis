package concentrated_liquidity

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

const (
	Uint64Bytes = uint64Bytes
)

var (
	EmptyCoins         = emptyCoins
	HundredFooCoins    = sdk.NewDecCoin("foo", sdk.NewInt(100))
	HundredBarCoins    = sdk.NewDecCoin("bar", sdk.NewInt(100))
	TwoHundredFooCoins = sdk.NewDecCoin("foo", sdk.NewInt(200))
	TwoHundredBarCoins = sdk.NewDecCoin("bar", sdk.NewInt(200))
	// TODO: this is incorrect. Should be grabbing from
	// authorized params instead. Must verify that all tests still make sense.
	// https://github.com/osmosis-labs/osmosis/issues/5039
	FullyChargedDuration = types.SupportedUptimes[len(types.SupportedUptimes)-1]
)

func (k Keeper) SetPool(ctx sdk.Context, pool types.ConcentratedPoolExtension) error {
	return k.setPool(ctx, pool)
}

func (k Keeper) HasPosition(ctx sdk.Context, positionId uint64) bool {
	return k.hasPosition(ctx, positionId)
}

func (k Keeper) DeletePosition(ctx sdk.Context, positionId uint64, owner sdk.AccAddress, poolId uint64) error {
	return k.deletePosition(ctx, positionId, owner, poolId)
}

func (k Keeper) GetPoolById(ctx sdk.Context, poolId uint64) (types.ConcentratedPoolExtension, error) {
	return k.getPoolById(ctx, poolId)
}

func (k Keeper) CrossTick(ctx sdk.Context, poolId uint64, tickIndex int64, swapStateFeeGrowth sdk.DecCoin) (liquidityDelta sdk.Dec, err error) {
	return k.crossTick(ctx, poolId, tickIndex, swapStateFeeGrowth)
}

func (k Keeper) SendCoinsBetweenPoolAndUser(ctx sdk.Context, denom0, denom1 string, amount0, amount1 sdk.Int, sender, receiver sdk.AccAddress) error {
	return k.sendCoinsBetweenPoolAndUser(ctx, denom0, denom1, amount0, amount1, sender, receiver)
}

func (k Keeper) SwapOutAmtGivenIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool types.ConcentratedPoolExtension,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	swapFee sdk.Dec,
	priceLimit sdk.Dec) (calcTokenIn, calcTokenOut sdk.Coin, currentTick int64, liquidity, sqrtPrice sdk.Dec, err error) {
	return k.swapOutAmtGivenIn(ctx, sender, pool, tokenIn, tokenOutDenom, swapFee, priceLimit)
}

func (k Keeper) ComputeOutAmtGivenIn(
	ctx sdk.Context,
	poolId uint64,
	tokenInMin sdk.Coin,
	tokenOutDenom string,
	swapFee sdk.Dec,
	priceLimit sdk.Dec,

) (calcTokenIn, calcTokenOut sdk.Coin, currentTick int64, liquidity, sqrtPrice sdk.Dec, err error) {
	return k.computeOutAmtGivenIn(ctx, poolId, tokenInMin, tokenOutDenom, swapFee, priceLimit)
}

func (k Keeper) SwapInAmtGivenOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool types.ConcentratedPoolExtension,
	desiredTokenOut sdk.Coin,
	tokenInDenom string,
	swapFee sdk.Dec,
	priceLimit sdk.Dec) (calcTokenIn, calcTokenOut sdk.Coin, currentTick int64, liquidity, sqrtPrice sdk.Dec, err error) {
	return k.swapInAmtGivenOut(ctx, sender, pool, desiredTokenOut, tokenInDenom, swapFee, priceLimit)
}

func (k Keeper) ComputeInAmtGivenOut(
	ctx sdk.Context,
	desiredTokenOut sdk.Coin,
	tokenInDenom string,
	swapFee sdk.Dec,
	priceLimit sdk.Dec,
	poolId uint64,

) (calcTokenIn, calcTokenOut sdk.Coin, currentTick int64, liquidity, sqrtPrice sdk.Dec, err error) {
	return k.computeInAmtGivenOut(ctx, desiredTokenOut, tokenInDenom, swapFee, priceLimit, poolId)
}

func (k Keeper) InitOrUpdateTick(ctx sdk.Context, poolId uint64, currentTick int64, tickIndex int64, liquidityIn sdk.Dec, upper bool) (err error) {
	return k.initOrUpdateTick(ctx, poolId, currentTick, tickIndex, liquidityIn, upper)
}

func (k Keeper) InitOrUpdatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityDelta sdk.Dec, joinTime time.Time, positionId uint64) (err error) {
	return k.initOrUpdatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidityDelta, joinTime, positionId)
}

func (k Keeper) GetNextPositionIdAndIncrement(ctx sdk.Context) uint64 {
	return k.getNextPositionIdAndIncrement(ctx)
}

func (k Keeper) PoolExists(ctx sdk.Context, poolId uint64) bool {
	return k.poolExists(ctx, poolId)
}

func (k Keeper) InitializeInitialPositionForPool(ctx sdk.Context, pool types.ConcentratedPoolExtension, amount0Desired, amount1Desired sdk.Int) error {
	return k.initializeInitialPositionForPool(ctx, pool, amount0Desired, amount1Desired)
}

func (k Keeper) CollectFees(ctx sdk.Context, owner sdk.AccAddress, positionId uint64) (sdk.Coins, error) {
	return k.collectFees(ctx, owner, positionId)
}

func (k Keeper) IsPositionOwner(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, positionId uint64) (bool, error) {
	return k.isPositionOwner(ctx, sender, poolId, positionId)
}

func (k Keeper) PrepareClaimableFees(ctx sdk.Context, positionId uint64) (sdk.Coins, error) {
	return k.prepareClaimableFees(ctx, positionId)
}

func ConvertConcentratedToPoolInterface(concentratedPool types.ConcentratedPoolExtension) (poolmanagertypes.PoolI, error) {
	return convertConcentratedToPoolInterface(concentratedPool)
}

func ConvertPoolInterfaceToConcentrated(poolI poolmanagertypes.PoolI) (types.ConcentratedPoolExtension, error) {
	return convertPoolInterfaceToConcentrated(poolI)
}

func (k Keeper) ValidateSwapFee(ctx sdk.Context, params types.Params, swapFee sdk.Dec) bool {
	return k.validateSwapFee(ctx, params, swapFee)
}

func (k Keeper) ValidateTickSpacing(ctx sdk.Context, params types.Params, tickSpacing uint64) bool {
	return k.validateTickSpacing(ctx, params, tickSpacing)
}

func (k Keeper) FungifyChargedPosition(ctx sdk.Context, owner sdk.AccAddress, positionIds []uint64) (uint64, error) {
	return k.fungifyChargedPosition(ctx, owner, positionIds)
}

func (k Keeper) ValidatePositionsAndGetTotalLiquidity(ctx sdk.Context, owner sdk.AccAddress, positionIds []uint64) (uint64, int64, int64, sdk.Dec, error) {
	return k.validatePositionsAndGetTotalLiquidity(ctx, owner, positionIds)
}

func (k Keeper) IsLockMature(ctx sdk.Context, underlyingLockId uint64) (bool, error) {
	return k.isLockMature(ctx, underlyingLockId)
}

func (k Keeper) PositionHasActiveUnderlyingLockAndUpdate(ctx sdk.Context, positionId uint64) (hasActiveUnderlyingLock bool, lockId uint64, err error) {
	return k.positionHasActiveUnderlyingLockAndUpdate(ctx, positionId)
}

func (k Keeper) UpdateFullRangeLiquidityInPool(ctx sdk.Context, poolId uint64, liquidity sdk.Dec) error {
	return k.updateFullRangeLiquidityInPool(ctx, poolId, liquidity)
}

func (k Keeper) MintSharesAndLock(ctx sdk.Context, concentratedPoolId, positionId uint64, owner sdk.AccAddress, remainingLockDuration time.Duration) (concentratedLockID uint64, underlyingLiquidityTokenized sdk.Coins, err error) {
	return k.mintSharesAndLock(ctx, concentratedPoolId, positionId, owner, remainingLockDuration)
}

func (k Keeper) SetPositionIdToLock(ctx sdk.Context, positionId, underlyingLockId uint64) {
	k.setPositionIdToLock(ctx, positionId, underlyingLockId)
}

func RoundTickToCanonicalPriceTick(lowerTick, upperTick int64, priceTickLower, priceTickUpper sdk.Dec, tickSpacing uint64) (int64, int64, error) {
	return roundTickToCanonicalPriceTick(lowerTick, upperTick, priceTickLower, priceTickUpper, tickSpacing)
}

// fees methods
func (k Keeper) CreateFeeAccumulator(ctx sdk.Context, poolId uint64) error {
	return k.createFeeAccumulator(ctx, poolId)
}

func (k Keeper) InitOrUpdatePositionFeeAccumulator(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64, positionId uint64, liquidity sdk.Dec) error {
	return k.initOrUpdatePositionFeeAccumulator(ctx, poolId, lowerTick, upperTick, positionId, liquidity)
}

func (k Keeper) GetFeeGrowthOutside(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64) (sdk.DecCoins, error) {
	return k.getFeeGrowthOutside(ctx, poolId, lowerTick, upperTick)
}

func CalculateFeeGrowth(targetTick int64, feeGrowthOutside sdk.DecCoins, currentTick int64, feesGrowthGlobal sdk.DecCoins, isUpperTick bool) sdk.DecCoins {
	return calculateFeeGrowth(targetTick, feeGrowthOutside, currentTick, feesGrowthGlobal, isUpperTick)
}

func (k Keeper) GetInitialFeeGrowthOppositeDirectionOfLastTraversalForTick(ctx sdk.Context, poolId uint64, tick int64) (sdk.DecCoins, error) {
	return k.getInitialFeeGrowthOppositeDirectionOfLastTraversalForTick(ctx, poolId, tick)
}

func (k Keeper) ChargeFee(ctx sdk.Context, poolId uint64, feeUpdate sdk.DecCoin) error {
	return k.chargeFee(ctx, poolId, feeUpdate)
}

func ValidateTickRangeIsValid(tickSpacing uint64, lowerTick int64, upperTick int64) error {
	return validateTickRangeIsValid(tickSpacing, lowerTick, upperTick)
}

func UpdatePosValueToInitValuePlusGrowthOutside(feeAccumulator accum.AccumulatorObject, positionKey string, feeGrowthOutside sdk.DecCoins) error {
	return updatePositionToInitValuePlusGrowthOutside(feeAccumulator, positionKey, feeGrowthOutside)
}

func UpdatePositionToInitValuePlusGrowthOutside(accumulator accum.AccumulatorObject, positionKey string, growthOutside sdk.DecCoins) error {
	return updatePositionToInitValuePlusGrowthOutside(accumulator, positionKey, growthOutside)
}

func (k Keeper) CreatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, tokensProvided sdk.Coins, amount0Min, amount1Min sdk.Int, lowerTick, upperTick int64) (positionId uint64, actualAmount0 sdk.Int, actualAmount1 sdk.Int, liquidityDelta sdk.Dec, joinTime time.Time, lowerTickResult int64, upperTickResult int64, err error) {
	return k.createPosition(ctx, poolId, owner, tokensProvided, amount0Min, amount1Min, lowerTick, upperTick)
}

func (k Keeper) AddToPosition(ctx sdk.Context, owner sdk.AccAddress, positionId uint64, amount0Added, amount1Added sdk.Int) (uint64, sdk.Int, sdk.Int, error) {
	return k.addToPosition(ctx, owner, positionId, amount0Added, amount1Added)
}

func (ss *SwapState) UpdateFeeGrowthGlobal(feeChargeTotal sdk.Dec) {
	ss.updateFeeGrowthGlobal(feeChargeTotal)
}

// Test helpers.
func (ss *SwapState) SetLiquidity(liquidity sdk.Dec) {
	ss.liquidity = liquidity
}

func (ss *SwapState) SetFeeGrowthGlobal(feeGrowthGlobal sdk.Dec) {
	ss.feeGrowthGlobal = feeGrowthGlobal
}

func (ss *SwapState) GetFeeGrowthGlobal() sdk.Dec {
	return ss.feeGrowthGlobal
}

// incentive methods
func (k Keeper) CreateUptimeAccumulators(ctx sdk.Context, poolId uint64) error {
	return k.createUptimeAccumulators(ctx, poolId)
}

func (k Keeper) GetUptimeAccumulatorValues(ctx sdk.Context, poolId uint64) ([]sdk.DecCoins, error) {
	return k.getUptimeAccumulatorValues(ctx, poolId)
}

func CalcAccruedIncentivesForAccum(ctx sdk.Context, accumUptime time.Duration, qualifyingLiquidity sdk.Dec, timeElapsed sdk.Dec, poolIncentiveRecords []types.IncentiveRecord) (sdk.DecCoins, []types.IncentiveRecord, error) {
	return calcAccruedIncentivesForAccum(ctx, accumUptime, qualifyingLiquidity, timeElapsed, poolIncentiveRecords)
}

func (k Keeper) UpdateUptimeAccumulatorsToNow(ctx sdk.Context, poolId uint64) error {
	return k.updatePoolUptimeAccumulatorsToNow(ctx, poolId)
}

func (k Keeper) SetIncentiveRecord(ctx sdk.Context, incentiveRecord types.IncentiveRecord) error {
	return k.setIncentiveRecord(ctx, incentiveRecord)
}

func (k Keeper) SetMultipleIncentiveRecords(ctx sdk.Context, incentiveRecords []types.IncentiveRecord) error {
	return k.setMultipleIncentiveRecords(ctx, incentiveRecords)
}

func (k Keeper) GetInitialUptimeGrowthOppositeDirectionOfLastTraversalForTick(ctx sdk.Context, poolId uint64, tick int64) ([]sdk.DecCoins, error) {
	return k.getInitialUptimeGrowthOppositeDirectionOfLastTraversalForTick(ctx, poolId, tick)
}

func (k Keeper) InitOrUpdatePositionUptimeAccumulators(ctx sdk.Context, poolId uint64, position sdk.Dec, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityDelta sdk.Dec, positionId uint64) error {
	return k.initOrUpdatePositionUptimeAccumulators(ctx, poolId, position, owner, lowerTick, upperTick, liquidityDelta, positionId)
}

func (k Keeper) GetAllIncentiveRecordsForUptime(ctx sdk.Context, poolId uint64, minUptime time.Duration) ([]types.IncentiveRecord, error) {
	return k.getAllIncentiveRecordsForUptime(ctx, poolId, minUptime)
}

func (k Keeper) CollectIncentives(ctx sdk.Context, owner sdk.AccAddress, positionId uint64) (sdk.Coins, sdk.Coins, error) {
	return k.collectIncentives(ctx, owner, positionId)
}

func GetUptimeTrackerValues(uptimeTrackers []model.UptimeTracker) []sdk.DecCoins {
	return getUptimeTrackerValues(uptimeTrackers)
}

func UpdateAccumAndClaimRewards(accum accum.AccumulatorObject, positionKey string, growthOutside sdk.DecCoins) (sdk.Coins, sdk.DecCoins, error) {
	return updateAccumAndClaimRewards(accum, positionKey, growthOutside)
}

func (k Keeper) ClaimAllIncentivesForPosition(ctx sdk.Context, positionId uint64) (sdk.Coins, sdk.Coins, error) {
	return k.claimAllIncentivesForPosition(ctx, positionId)
}

func FindUptimeIndex(uptime time.Duration) (int, error) {
	return findUptimeIndex(uptime)
}

func (k Keeper) GetAllPositions(ctx sdk.Context) ([]model.Position, error) {
	return k.getAllPositions(ctx)
}

func (k Keeper) UpdatePoolForSwap(ctx sdk.Context, pool types.ConcentratedPoolExtension, sender sdk.AccAddress, tokenIn sdk.Coin, tokenOut sdk.Coin, newCurrentTick int64, newLiquidity sdk.Dec, newSqrtPrice sdk.Dec) error {
	return k.updatePoolForSwap(ctx, pool, sender, tokenIn, tokenOut, newCurrentTick, newLiquidity, newSqrtPrice)
}

func (k Keeper) PrepareBalancerPoolAsFullRange(ctx sdk.Context, clPoolId uint64) (uint64, sdk.Dec, error) {
	return k.prepareBalancerPoolAsFullRange(ctx, clPoolId)
}

func (k Keeper) ClaimAndResetFullRangeBalancerPool(ctx sdk.Context, clPoolId uint64, balPoolId uint64) (sdk.Coins, error) {
	return k.claimAndResetFullRangeBalancerPool(ctx, clPoolId, balPoolId)
}

func (k Keeper) UninitializePool(ctx sdk.Context, poolId uint64) error {
	return k.uninitializePool(ctx, poolId)
}

// SetListenersUnsafe sets the listeners of the module. It is only meant to be used in tests.
// As a result, it is called unsafe.
func (k *Keeper) SetListenersUnsafe(listeners types.ConcentratedLiquidityListeners) {
	k.listeners = listeners
}

// GetListenersUnsafe returns the listeners of the module. It is only meant to be used in tests.
// As a result, it is called unsafe.
func (k Keeper) GetListenersUnsafe() types.ConcentratedLiquidityListeners {
	return k.listeners
}

func ValidateAuthorizedQuoteDenoms(ctx sdk.Context, denom1 string, authorizedQuoteDenoms []string) bool {
	return validateAuthorizedQuoteDenoms(ctx, denom1, authorizedQuoteDenoms)
}

func (k Keeper) ValidatePositionUpdateById(ctx sdk.Context, positionId uint64, updateInitiator sdk.AccAddress, lowerTickGiven int64, upperTickGiven int64, liquidityDeltaGiven sdk.Dec, joinTimeGiven time.Time, poolIdGiven uint64) error {
	return k.validatePositionUpdateById(ctx, positionId, updateInitiator, lowerTickGiven, upperTickGiven, liquidityDeltaGiven, joinTimeGiven, poolIdGiven)
}

func (k Keeper) GetLargestAuthorizedUptimeDuration(ctx sdk.Context) time.Duration {
	return k.getLargestAuthorizedUptimeDuration(ctx)
}

func MoveRewardsToNewPositionAndDeleteOldAcc(ctx sdk.Context, accum accum.AccumulatorObject, oldPositionName, newPositionName string, growthOutside sdk.DecCoins) error {
	return moveRewardsToNewPositionAndDeleteOldAcc(ctx, accum, oldPositionName, newPositionName, growthOutside)
}
