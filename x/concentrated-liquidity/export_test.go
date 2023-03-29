package concentrated_liquidity

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
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
)

// OrderInitialPoolDenoms sets the pool denoms of a cl pool
func OrderInitialPoolDenoms(denom0, denom1 string) (string, string, error) {
	return cltypes.OrderInitialPoolDenoms(denom0, denom1)
}

func (k Keeper) SetPool(ctx sdk.Context, pool types.ConcentratedPoolExtension) error {
	return k.setPool(ctx, pool)
}

func (k Keeper) HasFullPosition(ctx sdk.Context, positionId uint64) bool {
	return k.hasFullPosition(ctx, positionId)
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

func (k Keeper) GetTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64) (tickInfo model.TickInfo, err error) {
	return k.getTickInfo(ctx, poolId, tickIndex)
}

func (k Keeper) SendCoinsBetweenPoolAndUser(ctx sdk.Context, denom0, denom1 string, amount0, amount1 sdk.Int, sender, receiver sdk.AccAddress) error {
	return k.sendCoinsBetweenPoolAndUser(ctx, denom0, denom1, amount0, amount1, sender, receiver)
}

func (k Keeper) CalcInAmtGivenOutInternal(ctx sdk.Context, desiredTokenOut sdk.Coin, tokenInDenom string, swapFee sdk.Dec, priceLimit sdk.Dec, poolId uint64) (writeCtx func(), tokenIn, tokenOut sdk.Coin, updatedTick sdk.Int, updatedLiquidity, updatedSqrtPrice sdk.Dec, err error) {
	return k.calcInAmtGivenOut(ctx, desiredTokenOut, tokenInDenom, swapFee, priceLimit, poolId)
}

func (k Keeper) CalcOutAmtGivenInInternal(ctx sdk.Context, tokenInMin sdk.Coin, tokenOutDenom string, swapFee sdk.Dec, priceLimit sdk.Dec, poolId uint64) (writeCtx func(), tokenIn, tokenOut sdk.Coin, updatedTick sdk.Int, updatedLiquidity, updatedSqrtPrice sdk.Dec, err error) {
	return k.calcOutAmtGivenIn(ctx, tokenInMin, tokenOutDenom, swapFee, priceLimit, poolId)
}

func (k Keeper) SwapOutAmtGivenIn(ctx sdk.Context, sender sdk.AccAddress, pool types.ConcentratedPoolExtension, tokenIn sdk.Coin, tokenOutDenom string, swapFee sdk.Dec, priceLimit sdk.Dec) (calcTokenIn, calcTokenOut sdk.Coin, currentTick sdk.Int, liquidity, sqrtPrice sdk.Dec, err error) {
	return k.swapOutAmtGivenIn(ctx, sender, pool, tokenIn, tokenOutDenom, swapFee, priceLimit)
}

func (k *Keeper) SwapInAmtGivenOut(ctx sdk.Context, sender sdk.AccAddress, pool types.ConcentratedPoolExtension, desiredTokenOut sdk.Coin, tokenInDenom string, swapFee sdk.Dec, priceLimit sdk.Dec) (calcTokenIn, calcTokenOut sdk.Coin, currentTick sdk.Int, liquidity, sqrtPrice sdk.Dec, err error) {
	return k.swapInAmtGivenOut(ctx, sender, pool, desiredTokenOut, tokenInDenom, swapFee, priceLimit)
}

func (k Keeper) UpdatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityDelta sdk.Dec, joinTime time.Time, positionId uint64) (sdk.Int, sdk.Int, error) {
	return k.updatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidityDelta, joinTime, positionId)
}

func (k Keeper) InitOrUpdateTick(ctx sdk.Context, poolId uint64, currentTick int64, tickIndex int64, liquidityIn sdk.Dec, upper bool) (err error) {
	return k.initOrUpdateTick(ctx, poolId, currentTick, tickIndex, liquidityIn, upper)
}

func (k Keeper) InitOrUpdatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityDelta sdk.Dec, joinTime time.Time, positionId uint64) (err error) {
	return k.initOrUpdatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidityDelta, joinTime, positionId)
}

func (k Keeper) PoolExists(ctx sdk.Context, poolId uint64) bool {
	return k.poolExists(ctx, poolId)
}

func (k Keeper) IsInitialPositionForPool(initialSqrtPrice sdk.Dec, initialTick sdk.Int) bool {
	return k.isInitialPositionForPool(initialSqrtPrice, initialTick)
}

func (k Keeper) InitializeInitialPositionForPool(ctx sdk.Context, pool types.ConcentratedPoolExtension, amount0Desired, amount1Desired sdk.Int) error {
	return k.initializeInitialPositionForPool(ctx, pool, amount0Desired, amount1Desired)
}

func (k Keeper) CollectFees(ctx sdk.Context, owner sdk.AccAddress, positionId uint64) (sdk.Coins, error) {
	return k.collectFees(ctx, owner, positionId)
}

func (k Keeper) QueryClaimableFees(ctx sdk.Context, positionId uint64) (sdk.Coins, error) {
	return k.queryClaimableFees(ctx, positionId)
}

func ConvertConcentratedToPoolInterface(concentratedPool types.ConcentratedPoolExtension) (poolmanagertypes.PoolI, error) {
	return convertConcentratedToPoolInterface(concentratedPool)
}

func ConvertPoolInterfaceToConcentrated(poolI poolmanagertypes.PoolI) (types.ConcentratedPoolExtension, error) {
	return convertPoolInterfaceToConcentrated(poolI)
}

func (k Keeper) SetPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, joinTime time.Time, liquidity sdk.Dec, positionId uint64) {
	k.setPosition(ctx, poolId, owner, lowerTick, upperTick, joinTime, liquidity, positionId)
}

func (k Keeper) ValidateSwapFee(ctx sdk.Context, params types.Params, swapFee sdk.Dec) bool {
	return k.validateSwapFee(ctx, params, swapFee)
}

// fees methods
func (k Keeper) CreateFeeAccumulator(ctx sdk.Context, poolId uint64) error {
	return k.createFeeAccumulator(ctx, poolId)
}

func (k Keeper) GetFeeAccumulator(ctx sdk.Context, poolId uint64) (accum.AccumulatorObject, error) {
	return k.getFeeAccumulator(ctx, poolId)
}

func (k Keeper) InitializeFeeAccumulatorPosition(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64, positionId uint64) error {
	return k.initializeFeeAccumulatorPosition(ctx, poolId, lowerTick, upperTick, positionId)
}

func (k Keeper) UpdateFeeAccumulatorPosition(ctx sdk.Context, liquidityDelta sdk.Dec, positionId uint64) error {
	return k.updateFeeAccumulatorPosition(ctx, liquidityDelta, positionId)
}

func (k Keeper) GetFeeGrowthOutside(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64) (sdk.DecCoins, error) {
	return k.getFeeGrowthOutside(ctx, poolId, lowerTick, upperTick)
}

func CalculateFeeGrowth(targetTick int64, feeGrowthOutside sdk.DecCoins, currentTick int64, feesGrowthGlobal sdk.DecCoins, isUpperTick bool) sdk.DecCoins {
	return calculateFeeGrowth(targetTick, feeGrowthOutside, currentTick, feesGrowthGlobal, isUpperTick)
}

func (k Keeper) GetInitialFeeGrowthOutsideForTick(ctx sdk.Context, poolId uint64, tick int64) (sdk.DecCoins, error) {
	return k.getInitialFeeGrowthOutsideForTick(ctx, poolId, tick)
}

func (k Keeper) ChargeFee(ctx sdk.Context, poolId uint64, feeUpdate sdk.DecCoin) error {
	return k.chargeFee(ctx, poolId, feeUpdate)
}

func ValidateTickInRangeIsValid(tickSpacing uint64, exponentAtPriceOne sdk.Int, lowerTick int64, upperTick int64) error {
	return validateTickRangeIsValid(tickSpacing, exponentAtPriceOne, lowerTick, upperTick)
}

func PreparePositionAccumulator(feeAccumulator accum.AccumulatorObject, positionKey string, feeGrowthOutside sdk.DecCoins) error {
	return preparePositionAccumulator(feeAccumulator, positionKey, feeGrowthOutside)
}

func (k Keeper) CreatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, amount0Desired, amount1Desired, amount0Min, amount1Min sdk.Int, lowerTick, upperTick int64) (uint64, sdk.Int, sdk.Int, sdk.Dec, time.Time, error) {
	return k.createPosition(ctx, poolId, owner, amount0Desired, amount1Desired, amount0Min, amount1Min, lowerTick, upperTick)
}

func (k Keeper) WithdrawPosition(ctx sdk.Context, owner sdk.AccAddress, positionId uint64, requestedLiquidityAmountToWithdraw sdk.Dec) (amtDenom0, amtDenom1 sdk.Int, err error) {
	return k.withdrawPosition(ctx, owner, positionId, requestedLiquidityAmountToWithdraw)
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

func (k Keeper) GetUptimeAccumulators(ctx sdk.Context, poolId uint64) ([]accum.AccumulatorObject, error) {
	return k.getUptimeAccumulators(ctx, poolId)
}

func (k Keeper) GetUptimeAccumulatorValues(ctx sdk.Context, poolId uint64) ([]sdk.DecCoins, error) {
	return k.getUptimeAccumulatorValues(ctx, poolId)
}

func CalcAccruedIncentivesForAccum(ctx sdk.Context, accumUptime time.Duration, qualifyingLiquidity sdk.Dec, timeElapsed sdk.Dec, poolIncentiveRecords []types.IncentiveRecord) (sdk.DecCoins, []types.IncentiveRecord, error) {
	return calcAccruedIncentivesForAccum(ctx, accumUptime, qualifyingLiquidity, timeElapsed, poolIncentiveRecords)
}

func (k Keeper) UpdateUptimeAccumulatorsToNow(ctx sdk.Context, poolId uint64) error {
	return k.updateUptimeAccumulatorsToNow(ctx, poolId)
}

func (k Keeper) SetIncentiveRecord(ctx sdk.Context, incentiveRecord types.IncentiveRecord) {
	k.setIncentiveRecord(ctx, incentiveRecord)
}

func (k Keeper) SetMultipleIncentiveRecords(ctx sdk.Context, incentiveRecords []types.IncentiveRecord) error {
	return k.setMultipleIncentiveRecords(ctx, incentiveRecords)
}

func (k Keeper) GetInitialUptimeGrowthOutsidesForTick(ctx sdk.Context, poolId uint64, tick int64) ([]sdk.DecCoins, error) {
	return k.getInitialUptimeGrowthOutsidesForTick(ctx, poolId, tick)
}

func (k Keeper) InitOrUpdatePositionUptime(ctx sdk.Context, poolId uint64, position sdk.Dec, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityDelta sdk.Dec, joinTime time.Time, positionId uint64) error {
	return k.initOrUpdatePositionUptime(ctx, poolId, position, owner, lowerTick, upperTick, liquidityDelta, joinTime, positionId)
}

func (k Keeper) CollectIncentives(ctx sdk.Context, owner sdk.AccAddress, positionId uint64) (sdk.Coins, error) {
	return k.collectIncentives(ctx, owner, positionId)
}

func GetUptimeTrackerValues(uptimeTrackers []model.UptimeTracker) []sdk.DecCoins {
	return getUptimeTrackerValues(uptimeTrackers)
}

func (k Keeper) CreateIncentive(ctx sdk.Context, poolId uint64, sender sdk.AccAddress, incentiveDenom string, incentiveAmount sdk.Int, emissionRate sdk.Dec, startTime time.Time, minUptime time.Duration) (types.IncentiveRecord, error) {
	return k.createIncentive(ctx, poolId, sender, incentiveDenom, incentiveAmount, emissionRate, startTime, minUptime)
}

func PrepareAccumAndClaimRewards(accum accum.AccumulatorObject, positionKey string, growthOutside sdk.DecCoins) (sdk.Coins, error) {
	return prepareAccumAndClaimRewards(accum, positionKey, growthOutside)
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

func (k Keeper) UpdatePoolForSwap(ctx sdk.Context, pool types.ConcentratedPoolExtension, sender sdk.AccAddress, tokenIn sdk.Coin, tokenOut sdk.Coin, newCurrentTick sdk.Int, newLiquidity sdk.Dec, newSqrtPrice sdk.Dec) error {
	return k.updatePoolForSwap(ctx, pool, sender, tokenIn, tokenOut, newCurrentTick, newLiquidity, newSqrtPrice)
}
