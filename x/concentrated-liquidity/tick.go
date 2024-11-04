package concentrated_liquidity

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types/genesis"
)

// initOrUpdateTick retrieves the tickInfo from the specified tickIndex and updates both the liquidityNet and LiquidityGross.
// The given currentTick value is used to determine the strategy for updating the spread factor accumulator.
// We update the tick's spread reward growth opposite direction of last traversal accumulator to the spread reward growth global when tick index is <= current tick.
// Otherwise, it is set to zero. If the liquidityDelta causes the tick to be empty, a boolean flags that the tick is empty for the withdrawPosition method to handle later (removes the tick from state).
// Note that liquidityDelta can be either positive or negative depending on whether we are adding or removing liquidity.
// if we are initializing or updating an upper tick, we subtract the liquidityIn from the LiquidityNet
// if we are initializing or updating a lower tick, we add the liquidityIn from the LiquidityNet
// WARNING: this method may mutate the pool, make sure to refetch the pool after calling this method.
func (k Keeper) initOrUpdateTick(ctx sdk.Context, poolId uint64, tickIndex int64, liquidityDelta osmomath.Dec, upper bool) (tickIsEmpty bool, err error) {
	tickInfo, err := k.GetTickInfo(ctx, poolId, tickIndex)
	if err != nil {
		return false, err
	}

	// If both liquidity fields are zero, we consume the base gas spread factor for initializing a tick.
	if tickInfo.LiquidityGross.IsZero() && tickInfo.LiquidityNet.IsZero() {
		ctx.GasMeter().ConsumeGas(uint64(types.BaseGasFeeForInitializingTick), "initialize tick gas spread factor")
	}

	// calculate liquidityGross, which does not care about whether liquidityIn is positive or negative
	liquidityBefore := tickInfo.LiquidityGross

	// note that liquidityIn can be either positive or negative.
	// If negative, this would work as a subtraction from liquidityBefore
	liquidityAfter := liquidityBefore.Add(liquidityDelta)

	tickInfo.LiquidityGross = liquidityAfter

	// calculate liquidityNet, which we take into account and track depending on whether liquidityIn is positive or negative
	if upper {
		tickInfo.LiquidityNet.SubMut(liquidityDelta)
	} else {
		tickInfo.LiquidityNet.AddMut(liquidityDelta)
	}

	// If liquidity is now zero, this tick is flagged to be un-initialized at the end of the withdrawPosition method.
	if tickInfo.LiquidityGross.IsZero() && tickInfo.LiquidityNet.IsZero() {
		tickIsEmpty = true
	}

	k.SetTickInfo(ctx, poolId, tickIndex, &tickInfo)
	return tickIsEmpty, nil
}

// crossTick crosses the given tick. The tick is specified by its index and tick info.
// It updates the given tick's uptime and spread reward accumulators and writes it back to state.
// Prior to updating the tick info and writing it to state, it updates the pool uptime accumulators until the current block time.
// WARNING: this method may mutate the pool, make sure to refetch the pool after calling this method.
// CONTRACT: the caller validates that the pool with the given id exists.
// CONTRACT: caller is responsible for the uptimeAccums to be up-to-date.
// CONTRACT: uptimeAccums are associated with the given pool id.
func (k Keeper) crossTick(ctx sdk.Context, poolId uint64, tickIndex int64, tickInfo *model.TickInfo, swapStateSpreadRewardGrowth sdk.DecCoin, spreadRewardAccumValue sdk.DecCoins, uptimeAccums []*accum.AccumulatorObject) (err error) {
	if tickInfo == nil {
		return types.ErrNextTickInfoNil
	}

	// subtract tick's spread reward growth opposite direction of last traversal from current spread reward growth global, including the spread reward growth of the current swap.
	tickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal = spreadRewardAccumValue.Add(swapStateSpreadRewardGrowth).Sub(tickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal)

	// For each supported uptime, subtract tick's uptime growth outside from the respective uptime accumulator
	// This is functionally equivalent to "flipping" the trackers once the tick is crossed
	updatedUptimeTrackers := tickInfo.UptimeTrackers.List
	for uptimeId := range uptimeAccums {
		updatedUptimeTrackers[uptimeId].UptimeGrowthOutside = uptimeAccums[uptimeId].GetValue().Sub(updatedUptimeTrackers[uptimeId].UptimeGrowthOutside)
	}

	k.SetTickInfo(ctx, poolId, tickIndex, tickInfo)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCrossTick,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
			sdk.NewAttribute(types.AttributeKeyTickIndex, strconv.FormatInt(tickIndex, 10)),
			sdk.NewAttribute(types.AttributeKeySpreadRewardGrowthOppositeDirectionOfLastTraversal, tickInfo.SpreadRewardGrowthOppositeDirectionOfLastTraversal.String()),
			sdk.NewAttribute(types.AttributeKeyUptimeGrowthOppositeDirectionOfLastTraversal, tickInfo.UptimeTrackers.String()),
		),
	})

	return nil
}

// GetTickInfo gets the tickInfo given a poolId and tickIndex. If the tick has not been initialized, it will initialize it.
// If the tick has been initialized, it will return the tickInfo. If the pool does not exist, it will return an error.
// CONTRACT: The caller must check that the pool with given id exists.
// WARNING: this method may mutate the pool, make sure to refetch the pool after calling this method.
func (k Keeper) GetTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64) (tickInfo model.TickInfo, err error) {
	store := ctx.KVStore(k.storeKey)
	tickStruct := model.TickInfo{}
	key := types.KeyTick(poolId, tickIndex)

	found, err := osmoutils.Get(store, key, &tickStruct)
	if !found {
		return k.makeInitialTickInfo(ctx, poolId, tickIndex)
	}
	return tickStruct, err
}

func (k Keeper) makeInitialTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64) (tickStruct model.TickInfo, err error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return tickStruct, err
	}

	// We initialize the spread reward growth opposite direction of last traversal value.
	initialSpreadRewardGrowthOppositeDirectionOfLastTraversal, err := k.getInitialSpreadRewardGrowthOppositeDirectionOfLastTraversalForTick(ctx, pool, tickIndex)
	if err != nil {
		return tickStruct, err
	}

	// Sync global uptime accumulators to ensure the uptime tracker init values are up to date.
	if err := k.updatePoolUptimeAccumulatorsToNowWithPool(ctx, pool); err != nil {
		return tickStruct, err
	}

	// Initialize uptime trackers for the new tick to the appropriate starting values.
	valuesToAdd, err := k.getInitialUptimeGrowthOppositeDirectionOfLastTraversalForTick(ctx, pool, tickIndex)
	if err != nil {
		return tickStruct, err
	}

	initialUptimeTrackers := []model.UptimeTracker{}
	for _, uptimeTrackerValue := range valuesToAdd {
		initialUptimeTrackers = append(initialUptimeTrackers, model.UptimeTracker{UptimeGrowthOutside: uptimeTrackerValue})
	}

	uptimeTrackers := model.UptimeTrackers{List: initialUptimeTrackers}

	// Emit init tick event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtInitTick,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
			sdk.NewAttribute(types.AttributeKeyTickIndex, strconv.FormatInt(tickIndex, 10)),
			sdk.NewAttribute(types.AttributeKeySpreadRewardGrowthOppositeDirectionOfLastTraversal, initialSpreadRewardGrowthOppositeDirectionOfLastTraversal.String()),
			sdk.NewAttribute(types.AttributeKeyUptimeGrowthOppositeDirectionOfLastTraversal, uptimeTrackers.String()),
		),
	})

	return model.TickInfo{LiquidityGross: osmomath.ZeroDec(), LiquidityNet: osmomath.ZeroDec(), SpreadRewardGrowthOppositeDirectionOfLastTraversal: initialSpreadRewardGrowthOppositeDirectionOfLastTraversal, UptimeTrackers: uptimeTrackers}, nil
}

func (k Keeper) SetTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64, tickInfo *model.TickInfo) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyTick(poolId, tickIndex)
	osmoutils.MustSet(store, key, tickInfo)
}

// RemoveTickInfo removes the tickInfo from state.
func (k Keeper) RemoveTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyTick(poolId, tickIndex)
	store.Delete(key)

	// Emit remove tick event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtRemoveTick,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
			sdk.NewAttribute(types.AttributeKeyTickIndex, strconv.FormatInt(tickIndex, 10)),
		),
	})
}

func (k Keeper) GetAllInitializedTicksForPool(ctx sdk.Context, poolId uint64) ([]genesis.FullTick, error) {
	return osmoutils.GatherValuesFromStorePrefixWithKeyParser(ctx.KVStore(k.storeKey), types.KeyTickPrefixByPoolId(poolId), ParseFullTickFromBytes)
}

// validateTickInRangeIsValid validates that given ticks are valid. That is:
// - both lower and upper ticks are divisible by the tick spacing
// - both lower and upper ticks are within MinTick and MaxTick range
// - lower tick must be less than upper tick.
//
// Returns error if validation fails. Otherwise, nil.
func validateTickRangeIsValid(tickSpacing uint64, lowerTick int64, upperTick int64) error {
	// Check if the lower and upper tick values are divisible by the tick spacing.
	if lowerTick%int64(tickSpacing) != 0 || upperTick%int64(tickSpacing) != 0 {
		return types.TickSpacingError{LowerTick: lowerTick, UpperTick: upperTick, TickSpacing: tickSpacing}
	}

	// Check if the lower tick value is within the valid range of MinTick to MaxTick.
	if lowerTick < types.MinInitializedTick || lowerTick >= types.MaxTick {
		return types.InvalidTickError{Tick: lowerTick, IsLower: true, MinTick: types.MinInitializedTick, MaxTick: types.MaxTick}
	}

	// Check if the upper tick value is within the valid range of MinTick to MaxTick.
	if upperTick > types.MaxTick || upperTick <= types.MinInitializedTick {
		return types.InvalidTickError{Tick: upperTick, IsLower: false, MinTick: types.MinInitializedTick, MaxTick: types.MaxTick}
	}

	// Check if the lower tick value is greater than or equal to the upper tick value.
	if lowerTick >= upperTick {
		return types.InvalidLowerUpperTickError{LowerTick: lowerTick, UpperTick: upperTick}
	}
	return nil
}

// roundTickToCanonicalPriceTick takes a tick and determines if multiple ticks can represent the same price as the provided tick. If so, it
// rounds that tick up to the largest tick that can represent the same price that the original tick corresponded to. If one of
// the two ticks happen to be rounded, we re-validate the tick range to ensure that the tick range is still valid.
//
// i.e. the provided tick is -161795100. With our precision, this tick correlates to a sqrtPrice of 0.000000001414213563
// the first tick (given our precision) that is able to represent this price is -161000000, so we use this tick instead.
//
// This really only applies to very small tick values, as the increment of a single tick continues to get smaller as the tick value gets smaller.
func roundTickToCanonicalPriceTick(lowerTick, upperTick int64, sqrtPriceTickLower, sqrtPriceTickUpper osmomath.BigDec, tickSpacing uint64) (int64, int64, error) {
	newLowerTick, err := math.SqrtPriceToTickRoundDownSpacing(sqrtPriceTickLower, tickSpacing)
	if err != nil {
		return 0, 0, err
	}
	newUpperTick, err := math.SqrtPriceToTickRoundDownSpacing(sqrtPriceTickUpper, tickSpacing)
	if err != nil {
		return 0, 0, err
	}

	// If the lower or upper tick has changed, we need to re-validate the tick range.
	if lowerTick != newLowerTick || upperTick != newUpperTick {
		err := validateTickRangeIsValid(tickSpacing, newLowerTick, newUpperTick)
		if err != nil {
			return 0, 0, err
		}
	}
	return newLowerTick, newUpperTick, nil
}
