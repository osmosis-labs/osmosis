package concentrated_liquidity

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

var emptyCoins = sdk.DecCoins(nil)

// createSpreadRewardsAccumulator creates an accumulator object in the store using the given poolId.
// The accumulator is initialized with the default(zero) values.
func (k Keeper) createSpreadRewardsAccumulator(ctx sdk.Context, poolId uint64) error {
	err := accum.MakeAccumulator(ctx.KVStore(k.storeKey), types.KeySpreadFactorPoolAccumulator(poolId))
	if err != nil {
		return err
	}
	return nil
}

// GetSpreadRewardsAccumulator gets the spread factors accumulator object using the given poolOd
// returns error if accumulator for the given poolId does not exist.
func (k Keeper) GetSpreadRewardsAccumulator(ctx sdk.Context, poolId uint64) (accum.AccumulatorObject, error) {
	acc, err := accum.GetAccumulator(ctx.KVStore(k.storeKey), types.KeySpreadFactorPoolAccumulator(poolId))
	if err != nil {
		return accum.AccumulatorObject{}, err
	}

	return acc, nil
}

// chargeSpreadRewards charges the given spread factors on the pool with the given id by updating
// the internal per-pool accumulator that tracks spread reward growth per one unit of
// liquidity. Returns error if fails to get accumulator.
func (k Keeper) chargeSpreadRewards(ctx sdk.Context, poolId uint64, spreadRewardUpdate sdk.DecCoin) error {
	spreadRewardAccumulator, err := k.GetSpreadRewardsAccumulator(ctx, poolId)
	if err != nil {
		return err
	}

	spreadRewardAccumulator.AddToAccumulator(sdk.NewDecCoins(spreadRewardUpdate))

	return nil
}

// initOrUpdatePositionSpreadRewardsAccumulator mutates the spread factors accumulator position by either creating or updating it
// for the given pool id in the range specified by the given lower and upper ticks, position id and liquidityDelta.
// If liquidityDelta is positive, it adds liquidity. If liquidityDelta is negative, it removes liquidity.
// If this is a new position, liqudityDelta must be positive.
// It checks if the position exists in the spread factors accumulator. If it does not exist, it creates a new position.
// If it exists, it updates the shares of the position's accumulator in the spread factors accumulator.
// Upon calling this method, the position's spread factors accumulator is equal to the spread reward growth inside the tick range.
// On update, the rewards are moved into the position's unclaimed rewards. See internal method comments for details.
//
// Returns nil on success. Returns error if:
// - fails to get an accumulator for a given pool id
// - fails to determine whether the positive with the given id exists in the accumulator.
// - fails to calculate spread reward growth outside of the tick range.
// - fails to create a new position.
// - fails to prepare the accumulator for update.
// - fails to update the position's accumulator.
func (k Keeper) initOrUpdatePositionSpreadRewardsAccumulator(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64, positionId uint64, liquidityDelta sdk.Dec) error {
	// Get the spread factors accumulator for the position's pool.
	spreadRewardAccumulator, err := k.GetSpreadRewardsAccumulator(ctx, poolId)
	if err != nil {
		return err
	}

	// Get the key for the position's accumulator in the spread factors accumulator.
	positionKey := types.KeySpreadFactorPositionAccumulator(positionId)

	hasPosition, err := spreadRewardAccumulator.HasPosition(positionKey)
	if err != nil {
		return err
	}

	spreadRewardGrowthOutside, err := k.getSpreadRewardGrowthOutside(ctx, poolId, lowerTick, upperTick)
	if err != nil {
		return err
	}

	spreadRewardGrowthInside := spreadRewardAccumulator.GetValue().Sub(spreadRewardGrowthOutside)

	if !hasPosition {
		if !liquidityDelta.IsPositive() {
			return types.NonPositiveLiquidityForNewPositionError{LiquidityDelta: liquidityDelta, PositionId: positionId}
		}

		// Initialize the position with the spread reward growth inside the tick range
		if err := spreadRewardAccumulator.NewPositionIntervalAccumulation(positionKey, liquidityDelta, spreadRewardGrowthInside, nil); err != nil {
			return err
		}
	} else {
		// Replace the position's accumulator in the spread factors accumulator with a new one
		// that has the latest spread reward growth outside of the tick range.
		// Assume the last time the position was created or modified was at time t.
		// At time t, we track spread reward growth inside from 0 to t.
		// Then, the update happens at time t + 1. The call below makes the position's
		// accumulator to be "spread reward growth inside from 0 to t + spread reward growth outside from 0 to t + 1".
		err = updatePositionToInitValuePlusGrowthOutside(spreadRewardAccumulator, positionKey, spreadRewardGrowthOutside)
		if err != nil {
			return err
		}

		// Update the position's initialSpreadRewardsAccumulatorValue in the spread factors accumulator with spread reward growth inside,
		// taking into account the change in liquidity of the position.
		// Prior to mutating the accumulator, it moves the accumulated rewards into the accumulator position's unclaimed rewards.
		// The move happens by subtracting the "spread reward growth inside from 0 to t + spread reward growth outside from 0 to t + 1" from the global
		// spread factors accumulator growth at time t + 1. This yields the "spread reward growth inside from t to t + 1". That is, the unclaimed spread reward growth
		// from the last time the position was either modified or created.
		err = spreadRewardAccumulator.UpdatePositionIntervalAccumulation(positionKey, liquidityDelta, spreadRewardGrowthInside)
		if err != nil {
			return err
		}
	}

	return nil
}

// getSpreadRewardGrowthOutside returns the sum of spread reward growth above the upper tick and spread reward growth below the lower tick
// WARNING: this method may mutate the pool, make sure to refetch the pool after calling this method.
// Currently, Tte call to GetTickInfo() may mutate state.
func (k Keeper) getSpreadRewardGrowthOutside(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64) (sdk.DecCoins, error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.DecCoins{}, err
	}
	currentTick := pool.GetCurrentTick()

	// get lower, upper tick info
	lowerTickInfo, err := k.GetTickInfo(ctx, poolId, lowerTick)
	if err != nil {
		return sdk.DecCoins{}, err
	}
	upperTickInfo, err := k.GetTickInfo(ctx, poolId, upperTick)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	poolSpreadRewardsAccumulator, err := k.GetSpreadRewardsAccumulator(ctx, poolId)
	if err != nil {
		return sdk.DecCoins{}, err
	}
	poolSpreadRewardGrowth := poolSpreadRewardsAccumulator.GetValue()

	spreadRewardGrowthAboveUpperTick := calculateSpreadRewardGrowth(upperTick, upperTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal, currentTick, poolSpreadRewardGrowth, true)
	spreadRewardGrowthBelowLowerTick := calculateSpreadRewardGrowth(lowerTick, lowerTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal, currentTick, poolSpreadRewardGrowth, false)

	return spreadRewardGrowthAboveUpperTick.Add(spreadRewardGrowthBelowLowerTick...), nil
}

// getInitialSpreadRewardGrowthOppositeDirectionOfLastTraversalForTick returns what the initial value of the spread reward growth opposite direction of last traversal field should be for a given tick.
// This value depends on the provided tick's location relative to the current tick. If the provided tick is greater than the current tick,
// then the value is zero. Otherwise, the value is the value of the current global spread reward growth.
//
// The value is chosen as if all of the spread factors earned to date had occurred below the tick.
// Returns error if the pool with the given id does exist or if fails to get the spread factors accumulator.
func (k Keeper) getInitialSpreadRewardGrowthOppositeDirectionOfLastTraversalForTick(ctx sdk.Context, poolId uint64, tick int64) (sdk.DecCoins, error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	currentTick := pool.GetCurrentTick()
	if currentTick >= tick {
		spreadRewardAccumulator, err := k.GetSpreadRewardsAccumulator(ctx, poolId)
		if err != nil {
			return sdk.DecCoins{}, err
		}
		return spreadRewardAccumulator.GetValue(), nil
	}

	return emptyCoins, nil
}

// collectSpreadRewards collects the spread factors earned by a position and sends them to the owner's account.
// Returns error if the position with the given id does not exist or if fails to get the spread factors accumulator.
func (k Keeper) collectSpreadRewards(ctx sdk.Context, sender sdk.AccAddress, positionId uint64) (sdk.Coins, error) {
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Spread factor collector must be the owner of the position.
	isOwner, err := k.isPositionOwner(ctx, sender, position.PoolId, positionId)
	if err != nil {
		return sdk.Coins{}, err
	}
	if !isOwner {
		return sdk.Coins{}, types.NotPositionOwnerError{Address: sender.String(), PositionId: positionId}
	}

	// Get the amount of spread factors that the position is eligible to claim.
	// This also mutates the internal state of the spread factors accumulator.
	spreadRewardsClaimed, err := k.prepareClaimableSpreadRewards(ctx, positionId)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Send the claimed spread factors from the pool's address to the owner's address.
	pool, err := k.getPoolById(ctx, position.PoolId)
	if err != nil {
		return sdk.Coins{}, err
	}
	if err := k.bankKeeper.SendCoins(ctx, pool.GetSpreadRewardsAddress(), sender, spreadRewardsClaimed); err != nil {
		return sdk.Coins{}, err
	}

	// Emit an event for the spread factors collected.
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCollectSpreadRewards,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyPositionId, strconv.FormatUint(positionId, 10)),
			sdk.NewAttribute(types.AttributeKeyTokensOut, spreadRewardsClaimed.String()),
		),
	})

	return spreadRewardsClaimed, nil
}

// GetClaimableSpreadRewards returns the amount of spread factors that a position is eligible to claim.
//
// Returns error if:
// - pool with the given id does not exist
// - position given by pool id, owner, lower tick and upper tick does not exist
// - other internal database or math errors.
func (k Keeper) GetClaimableSpreadRewards(ctx sdk.Context, positionId uint64) (sdk.Coins, error) {
	// Since this is a query, we don't want to modify the state and therefore use a cache context.
	cacheCtx, _ := ctx.CacheContext()
	return k.prepareClaimableSpreadRewards(cacheCtx, positionId)
}

// prepareClaimableSpreadRewards returns the amount of spread factors that a position is eligible to claim.
// Note that it mutates the internal state of the spread factors accumulator by setting the position's
// unclaimed rewards to zero and update the position's accumulator value to reflect the
// current pool spread factors accumulator value.
//
// Returns error if:
// - pool with the given id does not exist
// - position given by pool id, owner, lower tick and upper tick does not exist
// - other internal database or math errors.
func (k Keeper) prepareClaimableSpreadRewards(ctx sdk.Context, positionId uint64) (sdk.Coins, error) {
	// Get the position with the given ID.
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return nil, err
	}

	// Get the spread factors accumulator for the position's pool.
	spreadRewardAccumulator, err := k.GetSpreadRewardsAccumulator(ctx, position.PoolId)
	if err != nil {
		return nil, err
	}

	// Get the key for the position's accumulator in the spread factors accumulator.
	positionKey := types.KeySpreadFactorPositionAccumulator(positionId)

	// Check if the position exists in the spread factors accumulator.
	hasPosition, err := spreadRewardAccumulator.HasPosition(positionKey)
	if err != nil {
		return nil, err
	}
	if !hasPosition {
		return nil, types.SpreadFactorPositionNotFoundError{PositionId: positionId}
	}

	// Compute the spread reward growth outside of the range between the position's lower and upper ticks.
	spreadRewardGrowthOutside, err := k.getSpreadRewardGrowthOutside(ctx, position.PoolId, position.LowerTick, position.UpperTick)
	if err != nil {
		return nil, err
	}

	// Claim rewards, set the unclaimed rewards to zero, and update the position's accumulator value to reflect the current accumulator value.
	spreadRewardsClaimed, _, err := updateAccumAndClaimRewards(spreadRewardAccumulator, positionKey, spreadRewardGrowthOutside)
	if err != nil {
		return nil, err
	}

	return spreadRewardsClaimed, nil
}

// calculateSpreadRewardGrowth above or below the given tick.
// If calculating spread reward growth for an upper tick, we consider the following two cases
// 1. currentTick >= upperTick: If current Tick is GTE than the upper Tick, the spread reward growth would be pool spread reward growth - uppertick's spread reward growth outside
// 2. currentTick < upperTick: If current tick is smaller than upper tick, spread reward growth would be the upper tick's spread reward growth outside
// this goes vice versa for calculating spread reward growth for lower tick.
func calculateSpreadRewardGrowth(targetTick int64, ticksSpreadRewardGrowthOppositeDirectionOfLastTraversal sdk.DecCoins, currentTick int64, spreadRewardsGrowthGlobal sdk.DecCoins, isUpperTick bool) sdk.DecCoins {
	if (isUpperTick && currentTick >= targetTick) || (!isUpperTick && currentTick < targetTick) {
		return spreadRewardsGrowthGlobal.Sub(ticksSpreadRewardGrowthOppositeDirectionOfLastTraversal)
	}
	return ticksSpreadRewardGrowthOppositeDirectionOfLastTraversal
}

// updatePositionToInitValuePlusGrowthOutside is called prior to updating unclaimed rewards,
// as we must set the position's accumulator value to the sum of
// - the spread factor/uptime growth inside at position creation time (position.InitAccumValue)
// - spread factor/uptime growth outside at the current block time (spreadRewardGrowthOutside/uptimeGrowthOutside)
func updatePositionToInitValuePlusGrowthOutside(accumulator accum.AccumulatorObject, positionKey string, growthOutside sdk.DecCoins) error {
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
