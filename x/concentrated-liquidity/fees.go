package concentrated_liquidity

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

var emptyCoins = sdk.DecCoins(nil)

// createFeeAccumulator creates an accumulator object in the store using the given poolId.
// The accumulator is initialized with the default(zero) values.
func (k Keeper) createFeeAccumulator(ctx sdk.Context, poolId uint64) error {
	err := accum.MakeAccumulator(ctx.KVStore(k.storeKey), types.KeyFeePoolAccumulator(poolId))
	if err != nil {
		return err
	}
	return nil
}

// GetFeeAccumulator gets the fee accumulator object using the given poolOd
// returns error if accumulator for the given poolId does not exist.
func (k Keeper) GetFeeAccumulator(ctx sdk.Context, poolId uint64) (accum.AccumulatorObject, error) {
	acc, err := accum.GetAccumulator(ctx.KVStore(k.storeKey), types.KeyFeePoolAccumulator(poolId))
	if err != nil {
		return accum.AccumulatorObject{}, err
	}

	return acc, nil
}

// chargeFee charges the given fee on the pool with the given id by updating
// the internal per-pool accumulator that tracks fee growth per one unit of
// liquidity. Returns error if fails to get accumulator.
func (k Keeper) chargeFee(ctx sdk.Context, poolId uint64, feeUpdate sdk.DecCoin) error {
	feeAccumulator, err := k.GetFeeAccumulator(ctx, poolId)
	if err != nil {
		return err
	}

	feeAccumulator.AddToAccumulator(sdk.NewDecCoins(feeUpdate))

	return nil
}

// initOrUpdatePositionFeeAccumulator mutates the fee accumulator position by either creating or updating it
// for the given pool id in the range specified by the given lower and upper ticks, position id and liquidityDelta.
// If liquidityDelta is positive, it adds liquidity. If liquidityDelta is negative, it removes liquidity.
// If this is a new position, liqudityDelta must be positive.
// It checks if the position exists in the fee accumulator. If it does not exist, it creates a new position.
// If it exists, it updates the shares of the position's accumulator in the fee accumulator.
// Upon calling this method, the position's fee accumulator is equal to the fee growth inside the tick range.
// On update, the rewards are moved into the position's unclaimed rewards. See internal method comments for details.
//
// Returns nil on success. Returns error if:
// - fails to get an accumulator for a given pool id
// - fails to determine whether the positive with the given id exists in the accumulator.
// - fails to calculate fee growth outside of the tick range.
// - fails to create a new position.
// - fails to prepare the accumulator for update.
// - fails to update the position's accumulator.
func (k Keeper) initOrUpdatePositionFeeAccumulator(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64, positionId uint64, liquidityDelta sdk.Dec) error {
	// Get the fee accumulator for the position's pool.
	feeAccumulator, err := k.GetFeeAccumulator(ctx, poolId)
	if err != nil {
		return err
	}

	// Get the key for the position's accumulator in the fee accumulator.
	positionKey := types.KeyFeePositionAccumulator(positionId)

	hasPosition, err := feeAccumulator.HasPosition(positionKey)
	if err != nil {
		return err
	}

	feeGrowthOutside, err := k.getFeeGrowthOutside(ctx, poolId, lowerTick, upperTick)
	if err != nil {
		return err
	}

	feeGrowthInside := feeAccumulator.GetValue().Sub(feeGrowthOutside)

	if !hasPosition {
		if !liquidityDelta.IsPositive() {
			return types.NonPositiveLiquidityForNewPositionError{LiquidityDelta: liquidityDelta, PositionId: positionId}
		}

		// Initialize the position with the fee growth inside the tick range
		if err := feeAccumulator.NewPositionIntervalAccumulation(positionKey, liquidityDelta, feeGrowthInside, nil); err != nil {
			return err
		}
	} else {
		// Replace the position's accumulator in the fee accumulator with a new one
		// that has the latest fee growth outside of the tick range.
		// Assume the last time the position was created or modified was at time t.
		// At time t, we track fee growth inside from 0 to t.
		// Then, the update happens at time t + 1. The call below makes the position's
		// accumulator to be "fee growth inside from 0 to t + fee growth outside from 0 to t + 1".
		err = preparePositionAccumulator(feeAccumulator, positionKey, feeGrowthOutside)
		if err != nil {
			return err
		}

		// Update the position's initialFeeAccumulatorValue in the fee accumulator with fee growth inside,
		// taking into account the change in liquidity of the position.
		// Prior to mutating the accumulator, it moves the accumulated rewards into the accumulator position's unclaimed rewards.
		// The move happens by subtracting the "fee growth inside from 0 to t + fee growth outside from 0 to t + 1" from the global
		// fee accumulator growth at time t + 1. This yields the "fee growth inside from t to t + 1". That is, the unclaimed fee growth
		// from the last time the position was either modified or created.
		err = feeAccumulator.UpdatePositionIntervalAccumulation(positionKey, liquidityDelta, feeGrowthInside)
		if err != nil {
			return err
		}
	}

	return nil
}

// getFeeGrowthOutside returns the sum of fee growth above the upper tick and fee growth below the lower tick
// WARNING: this method may mutate the pool, make sure to refetch the pool after calling this method.
// Currently, Tte call to GetTickInfo() may mutate state.
func (k Keeper) getFeeGrowthOutside(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64) (sdk.DecCoins, error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.DecCoins{}, err
	}
	currentTick := pool.GetCurrentTick().Int64()

	// get lower, upper tick info
	lowerTickInfo, err := k.GetTickInfo(ctx, poolId, lowerTick)
	if err != nil {
		return sdk.DecCoins{}, err
	}
	upperTickInfo, err := k.GetTickInfo(ctx, poolId, upperTick)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	poolFeeAccumulator, err := k.GetFeeAccumulator(ctx, poolId)
	if err != nil {
		return sdk.DecCoins{}, err
	}
	poolFeeGrowth := poolFeeAccumulator.GetValue()

	feeGrowthAboveUpperTick := calculateFeeGrowth(upperTick, upperTickInfo.FeeGrowthOppositeDirectionOfLastTraversal, currentTick, poolFeeGrowth, true)
	feeGrowthBelowLowerTick := calculateFeeGrowth(lowerTick, lowerTickInfo.FeeGrowthOppositeDirectionOfLastTraversal, currentTick, poolFeeGrowth, false)

	return feeGrowthAboveUpperTick.Add(feeGrowthBelowLowerTick...), nil
}

// getInitialFeeGrowthOppositeDirectionOfLastTraversalForTick returns what the initial value of the fee growth opposite direction of last traversal field should be for a given tick.
// This value depends on the provided tick's location relative to the current tick. If the provided tick is greater than the current tick,
// then the value is zero. Otherwise, the value is the value of the current global fee growth.
//
// The value is chosen as if all of the fees earned to date had occurred below the tick.
// Returns error if the pool with the given id does exist or if fails to get the fee accumulator.
func (k Keeper) getInitialFeeGrowthOppositeDirectionOfLastTraversalForTick(ctx sdk.Context, poolId uint64, tick int64) (sdk.DecCoins, error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	currentTick := pool.GetCurrentTick().Int64()
	if currentTick >= tick {
		feeAccumulator, err := k.GetFeeAccumulator(ctx, poolId)
		if err != nil {
			return sdk.DecCoins{}, err
		}
		return feeAccumulator.GetValue(), nil
	}

	return emptyCoins, nil
}

// collectFees collects the fees earned by a position and sends them to the owner's account.
// Returns error if the position with the given id does not exist or if fails to get the fee accumulator.
func (k Keeper) collectFees(ctx sdk.Context, sender sdk.AccAddress, positionId uint64) (sdk.Coins, error) {
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Fee collector must be the owner of the position.
	isOwner, err := k.isPositionOwner(ctx, sender, position.PoolId, positionId)
	if err != nil {
		return sdk.Coins{}, err
	}
	if !isOwner {
		return sdk.Coins{}, types.NotPositionOwnerError{Address: sender.String(), PositionId: positionId}
	}

	// Get the amount of fees that the position is eligible to claim.
	// This also mutates the internal state of the fee accumulator.
	feesClaimed, err := k.prepareClaimableFees(ctx, positionId)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Send the claimed fees from the pool's address to the owner's address.
	pool, err := k.getPoolById(ctx, position.PoolId)
	if err != nil {
		return sdk.Coins{}, err
	}
	if err := k.bankKeeper.SendCoins(ctx, pool.GetAddress(), sender, feesClaimed); err != nil {
		return sdk.Coins{}, err
	}

	// Emit an event for the fees collected.
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCollectFees,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyPositionId, strconv.FormatUint(positionId, 10)),
			sdk.NewAttribute(types.AttributeKeyTokensOut, feesClaimed.String()),
		),
	})

	return feesClaimed, nil
}

// GetClaimableFees returns the amount of fees that a position is eligible to claim.
//
// Returns error if:
// - pool with the given id does not exist
// - position given by pool id, owner, lower tick and upper tick does not exist
// - other internal database or math errors.
func (k Keeper) GetClaimableFees(ctx sdk.Context, positionId uint64) (sdk.Coins, error) {
	// Since this is a query, we don't want to modify the state and therefore use a cache context.
	cacheCtx, _ := ctx.CacheContext()
	return k.prepareClaimableFees(cacheCtx, positionId)
}

// prepareClaimableFees returns the amount of fees that a position is eligible to claim.
// Note that it mutates the internal state of the fee accumulator by setting the position's
// unclaimed rewards to zero and update the position's accumulator value to reflect the
// current pool fee accumulator value.
//
// Returns error if:
// - pool with the given id does not exist
// - position given by pool id, owner, lower tick and upper tick does not exist
// - other internal database or math errors.
func (k Keeper) prepareClaimableFees(ctx sdk.Context, positionId uint64) (sdk.Coins, error) {
	// Get the position with the given ID.
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return nil, err
	}

	// Get the fee accumulator for the position's pool.
	feeAccumulator, err := k.GetFeeAccumulator(ctx, position.PoolId)
	if err != nil {
		return nil, err
	}

	// Get the key for the position's accumulator in the fee accumulator.
	positionKey := types.KeyFeePositionAccumulator(positionId)

	// Check if the position exists in the fee accumulator.
	hasPosition, err := feeAccumulator.HasPosition(positionKey)
	if err != nil {
		return nil, err
	}
	if !hasPosition {
		return nil, types.FeePositionNotFoundError{PositionId: positionId}
	}

	// Compute the fee growth outside of the range between the position's lower and upper ticks.
	feeGrowthOutside, err := k.getFeeGrowthOutside(ctx, position.PoolId, position.LowerTick, position.UpperTick)
	if err != nil {
		return nil, err
	}

	// Claim rewards, set the unclaimed rewards to zero, and update the position's accumulator value to reflect the current accumulator value.
	feesClaimed, _, err := prepareAccumAndClaimRewards(feeAccumulator, positionKey, feeGrowthOutside)
	if err != nil {
		return nil, err
	}

	return feesClaimed, nil
}

// calculateFeeGrowth above or below the given tick.
// If calculating fee growth for an upper tick, we consider the following two cases
// 1. currentTick >= upperTick: If current Tick is GTE than the upper Tick, the fee growth would be pool fee growth - uppertick's fee growth outside
// 2. currentTick < upperTick: If current tick is smaller than upper tick, fee growth would be the upper tick's fee growth outside
// this goes vice versa for calculating fee growth for lower tick.
func calculateFeeGrowth(targetTick int64, ticksFeeGrowthOppositeDirectionOfLastTraversal sdk.DecCoins, currentTick int64, feesGrowthGlobal sdk.DecCoins, isUpperTick bool) sdk.DecCoins {
	if (isUpperTick && currentTick >= targetTick) || (!isUpperTick && currentTick < targetTick) {
		return feesGrowthGlobal.Sub(ticksFeeGrowthOppositeDirectionOfLastTraversal)
	}
	return ticksFeeGrowthOppositeDirectionOfLastTraversal
}

// preparePositionAccumulator is called prior to updating unclaimed rewards,
// as we must set the position's accumulator value to the sum of
// - the fee/uptime growth inside at position creation time (position.InitAccumValue)
// - fee/uptime growth outside at the current block time (feeGrowthOutside/uptimeGrowthOutside)
// CONTRACT: position accumulator value prior to this call is equal to the growth inside the position at the time of last update.
func preparePositionAccumulator(accumulator accum.AccumulatorObject, positionKey string, growthOutside sdk.DecCoins) error {
	position, err := accum.GetPosition(accumulator, positionKey)
	if err != nil {
		return err
	}

	// The reason for adding the growth outside to the position's initial accumulator value per share is as follows:
	// - At any time in-between position updates or claiming, a position must have its AccumValuePerShare be equal to growth_inside_at_{last time of update}.
	// - Prior to claiming (the logic below), the position's accumulator is updated to:
	//   growth_inside_at_{last time of update} + growth_outside_at_{current block time of update}
	// - Then, during claiming in osmoutils.ClaimRewards, we perform the following computation:
	// growth_global_at{current block time} - (growth_inside_at_{last time of update} + growth_outside_at_{current block time of update}})
	// which ends up being equal to growth_inside_from_{last_time_of_update}_to_{current block time of update}}.
	intervalAccumulationOutside := position.AccumValuePerShare.Add(growthOutside...)
	err = accumulator.SetPositionIntervalAccumulation(positionKey, intervalAccumulationOutside)
	if err != nil {
		return err
	}
	return nil
}
