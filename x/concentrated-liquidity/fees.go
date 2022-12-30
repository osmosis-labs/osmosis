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
	zero = sdk.ZeroDec()
)

func (k Keeper) createFeeAccumulator(ctx sdk.Context, poolId uint64) error {
	err := accum.MakeAccumulator(ctx.KVStore(k.storeKey), getFeeAccumulatorName(poolId))
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) getFeeAccumulator(ctx sdk.Context, poolId uint64) (accum.AccumulatorObject, error) {
	acc, err := accum.GetAccumulator(ctx.KVStore(k.storeKey), getFeeAccumulatorName(poolId))
	if err != nil {
		return accum.AccumulatorObject{}, err
	}

	return acc, nil
}

func (k Keeper) chargeFee(ctx sdk.Context, poolId uint64, feeUpdate sdk.DecCoin) error {
	feeAccumulator, err := k.getFeeAccumulator(ctx, poolId)
	if err != nil {
		return err
	}

	feeAccumulator.UpdateAccumulator(sdk.NewDecCoins(feeUpdate))

	return nil
}

func (k Keeper) initializeFeeAccumulatorPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, liquidityDelta sdk.Dec) error {
	feeAccumulator, err := k.getFeeAccumulator(ctx, poolId)
	if err != nil {
		return err
	}

	if err := feeAccumulator.NewPositionCustom(owner, zero, sdk.NewDecCoins(), nil); err != nil {
		return err
	}

	return nil
}

func (k Keeper) updateFeeAccumulatorPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, liquidityDelta sdk.Dec, lowerTick int64, upperTick int64) error {
	feeGrowthOutside, err := k.getFeeGrowthOutside(ctx, poolId, owner, lowerTick, upperTick)
	if err != nil {
		return err
	}

	feeAccumulator, err := k.getFeeAccumulator(ctx, poolId)
	if err != nil {
		return err
	}

	if err := feeAccumulator.UpdatePositionCustom(owner, liquidityDelta, feeGrowthOutside); err != nil {
		return err
	}

	return nil
}

func (k Keeper) getFeeGrowthOutside(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64) (sdk.DecCoins, error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	lowerTickInfo, err := k.getTickInfo(ctx, poolId, lowerTick)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	upperTickInfo, err := k.getTickInfo(ctx, poolId, upperTick)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	currentTick := pool.GetCurrentTick().Int64()

	feeGlobalAccumulator, err := k.getFeeAccumulator(ctx, poolId)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	feeGrowthGlobal := feeGlobalAccumulator.GetValue()

	feeGrowthAboveUpperTick := calculateFeeGrowthAbove(upperTick, upperTickInfo.FeeGrowthOutside, currentTick, feeGrowthGlobal)
	feeGrowthBelowLowerTick := calculateFeeGrowthBelow(lowerTick, lowerTickInfo.FeeGrowthOutside, currentTick, feeGrowthGlobal)

	return feeGrowthAboveUpperTick.Add(feeGrowthBelowLowerTick...), nil
}

// getInitialFeeGrowthOtsideForTick returns the initial value of fee growth outside for a given tick.
// This value depends on the tick's location relative to the current tick.
//
// feeGrowthOutside = { feeGrowthGlobal current tick >= tick }
//                    { 0               current tick <  tick }
//
// The value is chosen as if all of the fees earned to date had occurrd below the tick.
func (k Keeper) getInitialFeeGrowthOtsideForTick(ctx sdk.Context, poolId uint64, tick int64) (sdk.DecCoins, error) {
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

	return sdk.NewDecCoins(), nil
}

func (k Keeper) collectFees(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick int64, upperTick int64) error {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return err
	}

	feeAccumulator, err := k.getFeeAccumulator(ctx, poolId)
	if err != nil {
		return err
	}

	// We need to update the position's accumulator before we claim rewards.
	// Note that liquidity delta is zero in this case.
	if err := k.updateFeeAccumulatorPosition(ctx, poolId, owner, zero, lowerTick, upperTick); err != nil {
		return err
	}

	rewardsClaimed, err := feeAccumulator.ClaimRewards(owner)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoins(ctx, pool.GetAddress(), owner, rewardsClaimed); err != nil {
		return err
	}

	return nil
}

func getFeeAccumulatorName(poolId uint64) string {
	poolIdStr := strconv.FormatUint(poolId, uintBase)
	return strings.Join([]string{feeAccumPrefix, poolIdStr}, "/")
}

func calculateFeeGrowthAbove(upperTick int64, feeGrowthOutsideUpperTick sdk.DecCoins, currentTick int64, feesGrowthGlobal sdk.DecCoins) sdk.DecCoins {
	if currentTick >= upperTick {
		return feesGrowthGlobal.Sub(feeGrowthOutsideUpperTick)
	}
	return feeGrowthOutsideUpperTick
}

func calculateFeeGrowthBelow(lowerTick int64, feeGrowthOutsideLowerTick sdk.DecCoins, currentTick int64, feesGrowthGlobal sdk.DecCoins) sdk.DecCoins {
	if currentTick >= lowerTick {
		return feeGrowthOutsideLowerTick
	}
	return feesGrowthGlobal.Sub(feeGrowthOutsideLowerTick)
}
