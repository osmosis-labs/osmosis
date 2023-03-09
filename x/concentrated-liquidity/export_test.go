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

func (k Keeper) HasFullPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, joinTime time.Time, freezeDuration time.Duration) bool {
	return k.hasFullPosition(ctx, poolId, owner, lowerTick, upperTick, joinTime, freezeDuration)
}

func (k Keeper) DeletePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, joinTime time.Time, freezeDuration time.Duration) error {
	return k.deletePosition(ctx, poolId, owner, lowerTick, upperTick, joinTime, freezeDuration)
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

func (k Keeper) SwapOutAmtGivenIn(ctx sdk.Context, sender sdk.AccAddress, poolI poolmanagertypes.PoolI, tokenIn sdk.Coin, tokenOutDenom string, swapFee sdk.Dec, priceLimit sdk.Dec, poolId uint64) (calcTokenIn, calcTokenOut sdk.Coin, currentTick sdk.Int, liquidity, sqrtPrice sdk.Dec, err error) {
	return k.swapOutAmtGivenIn(ctx, sender, poolI, tokenIn, tokenOutDenom, swapFee, priceLimit, poolId)
}

func (k Keeper) UpdatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityDelta sdk.Dec, joinTime time.Time, freezeDuration time.Duration) (sdk.Int, sdk.Int, error) {
	return k.updatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidityDelta, joinTime, freezeDuration)
}

func (k Keeper) InitOrUpdateTick(ctx sdk.Context, poolId uint64, currentTick int64, tickIndex int64, liquidityIn sdk.Dec, upper bool) (err error) {
	return k.initOrUpdateTick(ctx, poolId, currentTick, tickIndex, liquidityIn, upper)
}

func (k Keeper) InitOrUpdatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityDelta sdk.Dec, joinTime time.Time, freezeDuration time.Duration) (err error) {
	return k.initOrUpdatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidityDelta, joinTime, freezeDuration)
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

func (k Keeper) CollectFees(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick int64, upperTick int64) (sdk.Coins, error) {
	return k.collectFees(ctx, poolId, owner, lowerTick, upperTick)
}

func (k Keeper) QueryClaimableFees(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick int64, upperTick int64) (sdk.Coins, error) {
	return k.queryClaimableFees(ctx, poolId, owner, lowerTick, upperTick)
}

func ConvertConcentratedToPoolInterface(concentratedPool types.ConcentratedPoolExtension) (poolmanagertypes.PoolI, error) {
	return convertConcentratedToPoolInterface(concentratedPool)
}

func ConvertPoolInterfaceToConcentrated(poolI poolmanagertypes.PoolI) (types.ConcentratedPoolExtension, error) {
	return convertPoolInterfaceToConcentrated(poolI)
}

func (k Keeper) GetAllPositionsWithVaryingFreezeTimes(ctx sdk.Context, poolId uint64, addr sdk.AccAddress, lowerTick, upperTick int64) ([]model.Position, error) {
	return k.getAllPositionsWithVaryingFreezeTimes(ctx, poolId, addr, lowerTick, upperTick)
}

func (k Keeper) SetPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, position *model.Position, joinTime time.Time, freezeDuration time.Duration) {
	k.setPosition(ctx, poolId, owner, lowerTick, upperTick, position, joinTime, freezeDuration)
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

func (k Keeper) InitializeFeeAccumulatorPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64) error {
	return k.initializeFeeAccumulatorPosition(ctx, poolId, owner, lowerTick, upperTick)
}

func (k Keeper) UpdateFeeAccumulatorPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, liquidityDelta sdk.Dec, lowerTick int64, upperTick int64) error {
	return k.updateFeeAccumulatorPosition(ctx, poolId, owner, liquidityDelta, lowerTick, upperTick)
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

func GetFeeAccumulatorName(poolId uint64) string {
	return getFeeAccumulatorName(poolId)
}

func (k Keeper) ChargeFee(ctx sdk.Context, poolId uint64, feeUpdate sdk.DecCoin) error {
	return k.chargeFee(ctx, poolId, feeUpdate)
}

func ValidateTickInRangeIsValid(tickSpacing uint64, exponentAtPriceOne sdk.Int, lowerTick int64, upperTick int64) error {
	return validateTickRangeIsValid(tickSpacing, exponentAtPriceOne, lowerTick, upperTick)
}

func FormatPositionAccumulatorKey(poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64) string {
	return formatFeePositionAccumulatorKey(poolId, owner, lowerTick, upperTick)
}

func PreparePositionAccumulator(feeAccumulator accum.AccumulatorObject, positionKey string, feeGrowthOutside sdk.DecCoins) error {
	return preparePositionAccumulator(feeAccumulator, positionKey, feeGrowthOutside)
}

func (k Keeper) CreatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, amount0Desired, amount1Desired, amount0Min, amount1Min sdk.Int, lowerTick, upperTick int64, freezeDuration time.Duration) (sdk.Int, sdk.Int, sdk.Dec, error) {
	return k.createPosition(ctx, poolId, owner, amount0Desired, amount1Desired, amount0Min, amount1Min, lowerTick, upperTick, freezeDuration)
}

func (k Keeper) WithdrawPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, joinTime time.Time, freezeDuration time.Duration, requestedLiquidityAmountToWithdraw sdk.Dec) (amtDenom0, amtDenom1 sdk.Int, err error) {
	return k.withdrawPosition(ctx, poolId, owner, lowerTick, upperTick, joinTime, freezeDuration, requestedLiquidityAmountToWithdraw)
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

func GetUptimeAccumulatorName(poolId, uptimeIndex uint64) string {
	return getUptimeAccumulatorName(poolId, uptimeIndex)
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

func (k Keeper) SetMultipleIncentiveRecords(ctx sdk.Context, incentiveRecords []types.IncentiveRecord) {
	k.setMultipleIncentiveRecords(ctx, incentiveRecords)
}

func (k Keeper) GetInitialUptimeGrowthOutsidesForTick(ctx sdk.Context, poolId uint64, tick int64) ([]sdk.DecCoins, error) {
	return k.getInitialUptimeGrowthOutsidesForTick(ctx, poolId, tick)
}

func (k Keeper) InitOrUpdatePositionUptime(ctx sdk.Context, poolId uint64, position *model.Position, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityDelta sdk.Dec, joinTime time.Time, freezeDuration time.Duration) error {
	return k.initOrUpdatePositionUptime(ctx, poolId, position, owner, lowerTick, upperTick, liquidityDelta, joinTime, freezeDuration)
}
