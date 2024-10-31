package concentrated_liquidity

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

var emptyCoins = sdk.DecCoins(nil)

// createSpreadRewardAccumulator creates an accumulator object in the store using the given poolId.
// The accumulator is initialized with the default(zero) values.
func (k Keeper) createSpreadRewardAccumulator(ctx sdk.Context, poolId uint64) error {
	err := accum.MakeAccumulator(ctx.KVStore(k.storeKey), types.KeySpreadRewardPoolAccumulator(poolId))
	if err != nil {
		return err
	}
	return nil
}

// GetSpreadRewardAccumulator gets the spread reward accumulator object using the given poolOd
// returns error if accumulator for the given poolId does not exist.
func (k Keeper) GetSpreadRewardAccumulator(ctx sdk.Context, poolId uint64) (*accum.AccumulatorObject, error) {
	return accum.GetAccumulator(ctx.KVStore(k.storeKey), types.KeySpreadRewardPoolAccumulator(poolId))
}

// initOrUpdatePositionSpreadRewardAccumulator mutates the spread reward accumulator position by either creating or updating it
// for the given pool id in the range specified by the given lower and upper ticks, position id and liquidityDelta.
// If liquidityDelta is positive, it adds liquidity. If liquidityDelta is negative, it removes liquidity.
// If this is a new position, liqudityDelta must be positive.
// It checks if the position exists in the spread reward accumulator. If it does not exist, it creates a new position.
// If it exists, it updates the shares of the position's accumulator in the spread reward accumulator.
// Upon calling this method, the position's spread reward accumulator is equal to the spread reward growth inside the tick range.
// On update, the rewards are moved into the position's unclaimed rewards. See internal method comments for details.
//
// Returns nil on success. Returns error if:
// - fails to get an accumulator for a given pool id
// - fails to determine whether the positive with the given id exists in the accumulator.
// - fails to calculate spread reward growth outside of the tick range.
// - fails to create a new position.
// - fails to prepare the accumulator for update.
// - fails to update the position's accumulator.
func (k Keeper) initOrUpdatePositionSpreadRewardAccumulator(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64, positionId uint64, liquidityDelta osmomath.Dec) error {
	// Get the spread reward accumulator for the position's pool.
	spreadRewardAccumulator, err := k.GetSpreadRewardAccumulator(ctx, poolId)
	if err != nil {
		return err
	}

	// Get the key for the position's accumulator in the spread reward accumulator.
	positionKey := types.KeySpreadRewardPositionAccumulator(positionId)

	hasPosition := spreadRewardAccumulator.HasPosition(positionKey)

	spreadRewardGrowthOutside, err := k.getSpreadRewardGrowthOutside(ctx, poolId, lowerTick, upperTick)
	if err != nil {
		return err
	}

	// Note: this is SafeSub because interval accumulation is allowed to be negative.
	spreadRewardGrowthInside, _ := spreadRewardAccumulator.GetValue().SafeSub(spreadRewardGrowthOutside)

	if !hasPosition {
		if !liquidityDelta.IsPositive() {
			return types.NonPositiveLiquidityForNewPositionError{LiquidityDelta: liquidityDelta, PositionId: positionId}
		}

		// Initialize the position with the spread reward growth inside the tick range
		if err := spreadRewardAccumulator.NewPositionIntervalAccumulation(positionKey, liquidityDelta, spreadRewardGrowthInside, nil); err != nil {
			return err
		}
	} else {
		// Replace the position's accumulator in the spread reward accumulator with a new one
		// that has the latest spread reward growth outside of the tick range.
		// Assume the last time the position was created or modified was at time t.
		// At time t, we track spread reward growth inside from 0 to t.
		// Then, the update happens at time t + 1. The call below makes the position's
		// accumulator to be "spread reward growth inside from 0 to t + spread reward growth outside from 0 to t + 1".
		err = updatePositionToInitValuePlusGrowthOutside(spreadRewardAccumulator, positionKey, spreadRewardGrowthOutside)
		if err != nil {
			return err
		}

		// Update the position's initialSpreadRewardAccumulatorValue in the spread reward accumulator with spread reward growth inside,
		// taking into account the change in liquidity of the position.
		// Prior to mutating the accumulator, it moves the accumulated rewards into the accumulator position's unclaimed rewards.
		// The move happens by subtracting the "spread reward growth inside from 0 to t + spread reward growth outside from 0 to t + 1" from the global
		// spread reward accumulator growth at time t + 1. This yields the "spread reward growth inside from t to t + 1". That is, the unclaimed spread reward growth
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
// Currently, the call to GetTickInfo() may mutate state.
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

	poolSpreadRewardAccumulator, err := k.GetSpreadRewardAccumulator(ctx, poolId)
	if err != nil {
		return sdk.DecCoins{}, err
	}
	poolSpreadRewardGrowth := poolSpreadRewardAccumulator.GetValue()

	spreadRewardGrowthAboveUpperTick := calculateSpreadRewardGrowth(upperTick, upperTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal, currentTick, poolSpreadRewardGrowth, true)
	spreadRewardGrowthBelowLowerTick := calculateSpreadRewardGrowth(lowerTick, lowerTickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal, currentTick, poolSpreadRewardGrowth, false)

	return spreadRewardGrowthAboveUpperTick.Add(spreadRewardGrowthBelowLowerTick...), nil
}

// getInitialSpreadRewardGrowthOppositeDirectionOfLastTraversalForTick returns what the initial value of the spread reward growth opposite direction of last traversal field should be for a given tick.
// This value depends on the provided tick's location relative to the current tick. If the provided tick is greater than the current tick,
// then the value is zero. Otherwise, the value is the value of the current global spread reward growth.
//
// The value is chosen as if all of the spread rewards earned to date had occurred below the tick.
// Returns error if the pool with the given id does exist or if fails to get the spread reward accumulator.
func (k Keeper) getInitialSpreadRewardGrowthOppositeDirectionOfLastTraversalForTick(ctx sdk.Context, pool types.ConcentratedPoolExtension, tick int64) (sdk.DecCoins, error) {
	currentTick := pool.GetCurrentTick()
	if currentTick >= tick {
		spreadRewardAccumulator, err := k.GetSpreadRewardAccumulator(ctx, pool.GetId())
		if err != nil {
			return sdk.DecCoins{}, err
		}
		return spreadRewardAccumulator.GetValue(), nil
	}

	return emptyCoins, nil
}

// collectSpreadRewards collects the spread reward earned by a position and sends them to the owner's account.
// Returns error if the position with the given id does not exist or if fails to get the spread reward accumulator.
func (k Keeper) collectSpreadRewards(ctx sdk.Context, sender sdk.AccAddress, positionId uint64) (sdk.Coins, error) {
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Spread reward collector must be the owner of the position.
	if sender.String() != position.Address {
		return sdk.Coins{}, types.NotPositionOwnerError{
			PositionId: positionId,
			Address:    sender.String(),
		}
	}

	// Get the amount of spread rewards that the position is eligible to claim.
	// This also mutates the internal state of the spread reward accumulator.
	spreadRewardsClaimed, err := k.prepareClaimableSpreadRewards(ctx, positionId)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Early return, emit no events if there is no spread rewards to claim.
	if spreadRewardsClaimed.IsZero() {
		return sdk.Coins{}, nil
	}

	// Send the claimed spread rewards from the pool's address to the owner's address.
	pool, err := k.getPoolById(ctx, position.PoolId)
	if err != nil {
		return sdk.Coins{}, err
	}
	if err := k.bankKeeper.SendCoins(ctx, pool.GetSpreadRewardsAddress(), sender, spreadRewardsClaimed); err != nil {
		return sdk.Coins{}, err
	}

	// Emit an event for the spread rewards collected.
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCollectSpreadRewards,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(pool.GetId(), 10)),
			sdk.NewAttribute(types.AttributeKeyPositionId, strconv.FormatUint(positionId, 10)),
			sdk.NewAttribute(types.AttributeKeyTokensOut, spreadRewardsClaimed.String()),
		),
	})

	return spreadRewardsClaimed, nil
}

// GetClaimableSpreadRewards returns the amount of spread rewards that a position is eligible to claim.
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

// prepareClaimableSpreadRewards returns the amount of spread rewards that a position is eligible to claim.
// Note that it mutates the internal state of the spread reward accumulator by setting the position's
// unclaimed rewards to zero and update the position's accumulator value to reflect the
// current pool spread reward accumulator value. If there is any dust left over, it is added back to the
// global accumulator as long as there are shares remaining in the accumulator. If not, the dust
// is ignored.
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

	// Get the spread reward accumulator for the position's pool.
	spreadRewardAccumulator, err := k.GetSpreadRewardAccumulator(ctx, position.PoolId)
	if err != nil {
		return nil, err
	}

	// Get the key for the position's accumulator in the spread reward accumulator.
	positionKey := types.KeySpreadRewardPositionAccumulator(positionId)

	// Check if the position exists in the spread reward accumulator.
	hasPosition := spreadRewardAccumulator.HasPosition(positionKey)
	if !hasPosition {
		return nil, types.SpreadRewardPositionNotFoundError{PositionId: positionId}
	}

	// Compute the spread reward growth outside of the range between the position's lower and upper ticks.
	spreadRewardGrowthOutside, err := k.getSpreadRewardGrowthOutside(ctx, position.PoolId, position.LowerTick, position.UpperTick)
	if err != nil {
		return nil, err
	}

	// Claim rewards, set the unclaimed rewards to zero, and update the position's accumulator value to reflect the current accumulator value.
	spreadRewardsClaimedScaled, forfeitedDustScaled, err := updateAccumAndClaimRewards(spreadRewardAccumulator, positionKey, spreadRewardGrowthOutside)
	if err != nil {
		return nil, err
	}

	spreadFactorScalingFactor, err := k.getSpreadFactorScalingFactorForPool(ctx, position.PoolId)
	if err != nil {
		return nil, err
	}

	// We scale the spread factor per-unit of liquidity accumulator up to avoid truncation to zero.
	// However, once we compute the total for the liquidity entitlement, we must scale it back down.
	// We always truncate down in the pool's favor.
	spreadRewardsClaimed := sdk.NewCoins()
	forfeitedDust := sdk.DecCoins{}
	if spreadFactorScalingFactor.Equal(oneDec) {
		// If the scaling factor is 1, we don't need to scale down the spread rewards.
		// We also use the forfeited dust calculated updateAccumAndClaimRewards since it is already scaled down.
		spreadRewardsClaimed = spreadRewardsClaimedScaled
		forfeitedDust = forfeitedDustScaled
	} else {
		// If the scaling factor is not 1, we scale down the spread rewards and throw away the dust.
		for _, coin := range spreadRewardsClaimedScaled {
			scaledCoinAmt := scaleDownSpreadRewardAmount(coin.Amount, spreadFactorScalingFactor)
			if !scaledCoinAmt.IsZero() {
				spreadRewardsClaimed = append(spreadRewardsClaimed, sdk.NewCoin(coin.Denom, scaledCoinAmt))
			}
		}
	}

	// add forfeited dust back to the global accumulator
	if !forfeitedDust.IsZero() {
		// Refetch the spread reward accumulator as the number of shares has changed after claiming.
		spreadRewardAccumulator, err := k.GetSpreadRewardAccumulator(ctx, position.PoolId)
		if err != nil {
			return nil, err
		}

		totalSharesRemaining := spreadRewardAccumulator.GetTotalShares()

		// if there are no shares remaining, the dust is ignored. Otherwise, it is added back to the global accumulator.
		// Total shares remaining can be zero if we claim in withdrawPosition for the last position in the pool.
		// The shares are decremented in osmoutils/accum.ClaimRewards.
		if !totalSharesRemaining.IsZero() {
			forfeitedDustPerShareScaled := forfeitedDust.QuoDecTruncate(totalSharesRemaining)
			spreadRewardAccumulator.AddToAccumulator(forfeitedDustPerShareScaled)
		}
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
// - the spread reward/uptime growth inside at position creation time (position.InitAccumValue)
// - spread reward/uptime growth outside at the current block time (spreadRewardGrowthOutside/uptimeGrowthOutside)
func updatePositionToInitValuePlusGrowthOutside(accumulator *accum.AccumulatorObject, positionKey string, growthOutside sdk.DecCoins) error {
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

// scaleDownSpreadRewardAmount scales down the spread reward amount by the scaling factor.
func scaleDownSpreadRewardAmount(incentiveAmount osmomath.Int, scalingFactor osmomath.Dec) (scaledTotalEmittedAmount osmomath.Int) {
	return incentiveAmount.ToLegacyDec().QuoTruncateMut(scalingFactor).TruncateInt()
}

// getSpreadFactorScalingFactorForPool returns the spread factor scaling factor for the given pool.
// It returns perUnitLiqScalingFactor if the pool is migrated or if the pool ID is greater than the migration threshold.
// It returns oneDecScalingFactor otherwise.
func (k Keeper) getSpreadFactorScalingFactorForPool(ctx sdk.Context, poolID uint64) (osmomath.Dec, error) {
	migrationThreshold, err := k.GetSpreadFactorPoolIDMigrationThreshold(ctx)
	if err != nil {
		return osmomath.Dec{}, err
	}

	// If the given pool ID is greater than the migration threshold, we return the perUnitLiqScalingFactor.
	if poolID > migrationThreshold {
		return perUnitLiqScalingFactor, nil
	}

	// If the given pool ID is one of the migrated spread factor accumulator pool IDs, we return the perUnitLiqScalingFactor.
	_, isMigrated := types.MigratedSpreadFactorAccumulatorPoolIDsV25[poolID]
	if isMigrated {
		return perUnitLiqScalingFactor, nil
	}

	// Otherwise, we return the oneDecScalingFactor.
	return oneDecScalingFactor, nil
}

// SetSpreadFactorPoolIDMigrationThreshold sets the pool ID migration threshold to the last pool ID for spread factor accumulators.
func (k Keeper) SetSpreadFactorPoolIDMigrationThreshold(ctx sdk.Context, poolIDThreshold uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeySpreadRewardAccumulatorMigrationThreshold, sdk.Uint64ToBigEndian(poolIDThreshold))
}

// GetSpreadFactorPoolIDMigrationThreshold returns the pool ID migration threshold for spread factor accumulators.
func (k Keeper) GetSpreadFactorPoolIDMigrationThreshold(ctx sdk.Context) (uint64, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeySpreadRewardAccumulatorMigrationThreshold)
	if bz == nil {
		return 0, fmt.Errorf("spread reward accumulator migration threshold not found")
	}

	threshold := sdk.BigEndianToUint64(bz)

	return threshold, nil
}
