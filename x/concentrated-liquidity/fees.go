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

var (
	emptyCoins = sdk.DecCoins(nil)
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

// chargeFee charges the given fee on the pool with the given id by updating
// the internal per-pool accumulator that tracks fee growth per one unit of
// liquidity. Returns error if fails to get accumulator.
// nolint: unused
func (k Keeper) chargeFee(ctx sdk.Context, poolId uint64, feeUpdate sdk.DecCoin) error {
	feeAccumulator, err := k.getFeeAccumulator(ctx, poolId)
	if err != nil {
		return err
	}

	feeAccumulator.UpdateAccumulator(sdk.NewDecCoins(feeUpdate))

	return nil
}

// initializeFeeAccumulatorPosition initializes the pool fee accumulator with given liquidity delta and zero value for the accumulator.
// nolint: unused
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
// nolint: unused
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

// initOrUpdateFeeAccumulatorPosition either updates or initializes a fee accumulator position.
// if fails upon getting and updating fee accumulator position for the given pool + owner accumulator,
// initializes the fee accumulator position.
// nolint: unused
func (k Keeper) initOrUpdateFeeAccumulatorPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, liquidityDelta sdk.Dec, lowerTick int64, upperTick int64) error {
	// first try updating fee accum position
	err := k.updateFeeAccumulatorPosition(ctx, poolId, owner, liquidityDelta, lowerTick, upperTick)
	if err != nil {
		err = k.initializeFeeAccumulatorPosition(ctx, poolId, owner, liquidityDelta)
		if err != nil {
			return err
		}
	}

	return nil
}

// getFeeGrowthOutside returns fee growth upper tick - fee growth lower tick
// nolint: unused
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

// getInitialFeeGrowthOutsideForTick returns the initial value of fee growth outside for a given tick.
// This value depends on the tick's location relative to the current tick.
//
// feeGrowthOutside = { feeGrowthGlobal current tick >= tick }
//                    { 0               current tick <  tick }
//
// The value is chosen as if all of the fees earned to date had occurrd below the tick.
// Returns error if the pool with the given id does exist or if fails to get the fee accumulator.
func (k Keeper) getInitialFeeGrowthOutsideForTick(ctx sdk.Context, poolId uint64, tick int64) (sdk.DecCoins, error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	currentTick := pool.GetCurrentTick().Int64()
	if currentTick >= tick {
		feeAccumulator, err := k.getFeeAccumulator(ctx, poolId)
		if err != nil {
			return sdk.DecCoins{}, err
		}
		return feeAccumulator.GetValue(), nil
	}

	return emptyCoins, nil
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
// nolint: unused
func calculateFeeGrowth(targetTick int64, feeGrowthOutside sdk.DecCoins, currentTick int64, feesGrowthGlobal sdk.DecCoins, isUpperTick bool) sdk.DecCoins {
	if (isUpperTick && currentTick >= targetTick) || (!isUpperTick && currentTick < targetTick) {
		return feesGrowthGlobal.Sub(feeGrowthOutside)
	}
	return feeGrowthOutside
}
