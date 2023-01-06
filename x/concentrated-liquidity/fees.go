package concentrated_liquidity

import (
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
)

const (
	feeAccumPrefix        = "fee"
	feeAccumNameSeparator = "/"
	uintBase              = 10
)

// createFeeAccumulator creates an accumulator object in the store using the given poolId.
// The accumulator is initialized with the default(zero) values.
func (k Keeper) createFeeAccumulator(ctx sdk.Context, poolId uint64) error {
	err := accum.MakeAccumulator(ctx.KVStore(k.storeKey), getFeeAccumulatorName(poolId))
	if err != nil {
		return err
	}
	return nil
}

// nolint: unused
// getFeeAccumulator gets the fee accumulator object using the given poolOd
// returns error if accumulator for the given poolId does not exist.
func (k Keeper) getFeeAccumulator(ctx sdk.Context, poolId uint64) (accum.AccumulatorObject, error) {
	acc, err := accum.GetAccumulator(ctx.KVStore(k.storeKey), getFeeAccumulatorName(poolId))
	if err != nil {
		return accum.AccumulatorObject{}, err
	}

	return acc, nil
}

// initializeFeeAccumulatorPosition initializes the pool fee accumulator with given liquidity delta and zero value for the accumulator.
func (k Keeper) initializeFeeAccumulatorPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, liquidityDelta sdk.Dec) error {
	// get fee accumulator for the pool
	feeAccumulator, err := k.getFeeAccumulator(ctx, poolId)
	if err != nil {
		return err
	}

	// initialize the owner's position with liquidity Delta and zero accumulator value
	if err := feeAccumulator.NewPositionCustomAcc(owner.String(), liquidityDelta, sdk.NewDecCoins(), nil); err != nil {
		return err
	}

	return nil
}

// updateFeeAccumulatorPosition updates the owner's position
func (k Keeper) updateFeeAccumulatorPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, liquidityDelta sdk.Dec, lowerTick int64, upperTick int64) error {
	feeGrowthOutside, err := k.getFeeGrowthOutside(ctx, poolId, lowerTick, upperTick)
	if err != nil {
		return err
	}

	feeAccumulator, err := k.getFeeAccumulator(ctx, poolId)
	if err != nil {
		return err
	}

	// replace position's accumulator with the updated liquidity and the feeGrowthOutside
	if err := feeAccumulator.UpdatePositionCustomAcc(owner.String(), liquidityDelta, feeGrowthOutside); err != nil {
		return err
	}

	return nil
}

// getFeeGrowthOutside returns fee growth upper tick - fee growth lower tick
func (k Keeper) getFeeGrowthOutside(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64) (sdk.DecCoins, error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.DecCoins{}, err
	}
	currentTick := pool.GetCurrentTick().Int64()

	// get lower, upper tick info
	lowerTickInfo, err := k.getTickInfo(ctx, poolId, lowerTick)
	if err != nil {
		return sdk.DecCoins{}, err
	}
	upperTickInfo, err := k.getTickInfo(ctx, poolId, upperTick)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	poolFeeAccumulator, err := k.getFeeAccumulator(ctx, poolId)
	if err != nil {
		return sdk.DecCoins{}, err
	}
	poolFeeGrowth := poolFeeAccumulator.GetValue()

	// calculate fee growth for upper tick and lower tick
	feeGrowthAboveUpperTick := calculateFeeGrowth(upperTick, upperTickInfo.FeeGrowthOutside, currentTick, poolFeeGrowth, true)
	feeGrowthBelowLowerTick := calculateFeeGrowth(lowerTick, lowerTickInfo.FeeGrowthOutside, currentTick, poolFeeGrowth, false)

	return feeGrowthAboveUpperTick.Add(feeGrowthBelowLowerTick...), nil
}

func getFeeAccumulatorName(poolId uint64) string {
	poolIdStr := strconv.FormatUint(poolId, uintBase)
	return strings.Join([]string{feeAccumPrefix, poolIdStr}, "/")
}

// calculateFeeGrowth for the given targetTicks.
// If calculating fee growth for an upper tick, we consider the following two cases
// 1. currentTick >= upperTick: If current Tick is GTE than the upper Tick, the fee growth would be pool fee growth - uppertick's fee growth outside
// 2. currentTick < upperTick: If current tick is smaller than upper tick, fee growth would be the upper tick's fee growth outside
// this goes vice versa for calculating fee growth for lower tick.
func calculateFeeGrowth(targetTick int64, feeGrowthOutside sdk.DecCoins, currentTick int64, feesGrowthGlobal sdk.DecCoins, isUpperTick bool) sdk.DecCoins {
	if (isUpperTick && currentTick >= targetTick) || (!isUpperTick && currentTick < targetTick) {
		return feesGrowthGlobal.Sub(feeGrowthOutside)
	}
	return feeGrowthOutside
}
