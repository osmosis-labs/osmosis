package concentrated_liquidity

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

// initOrUpdateTick retrieves the tickInfo from the specified tickIndex and updates both the liquidityNet and LiquidityGross.
// The given currentTick value is used to determine the strategy for updating the fee accumulator.
// We update the tick's fee growth outside accumulator to the fee growth global when tick index is <= current tick.
// Otherwise, it is set to zero.
// if we are initializing or updating an upper tick, we subtract the liquidityIn from the LiquidityNet
// if we are initializing or updating an lower tick, we add the liquidityIn from the LiquidityNet
func (k Keeper) initOrUpdateTick(ctx sdk.Context, poolId uint64, currentTick int64, tickIndex int64, liquidityIn sdk.Dec, upper bool) (err error) {
	tickInfo, err := k.getTickInfo(ctx, poolId, tickIndex)
	if err != nil {
		return err
	}

	// calculate liquidityGross, which does not care about whether liquidityIn is positive or negative
	liquidityBefore := tickInfo.LiquidityGross

	// if given tickIndex is LTE to the current tick and the liquidityBefore is zero,
	// set the tick's fee growth outside to the fee accumulator's value
	if liquidityBefore.IsZero() {
		if tickIndex <= currentTick {
			accum, err := k.getFeeAccumulator(ctx, poolId)
			if err != nil {
				return err
			}

			tickInfo.FeeGrowthOutside = accum.GetValue()
		}
	}

	// note that liquidityIn can be either positive or negative.
	// If negative, this would work as a subtraction from liquidityBefore
	liquidityAfter := math.AddLiquidity(liquidityBefore, liquidityIn)

	tickInfo.LiquidityGross = liquidityAfter

	// calculate liquidityNet, which we take into account and track depending on whether liquidityIn is positive or negative
	if upper {
		tickInfo.LiquidityNet = tickInfo.LiquidityNet.Sub(liquidityIn)
	} else {
		tickInfo.LiquidityNet = tickInfo.LiquidityNet.Add(liquidityIn)
	}

	k.SetTickInfo(ctx, poolId, tickIndex, tickInfo)

	return nil
}

func (k Keeper) crossTick(ctx sdk.Context, poolId uint64, tickIndex int64, swapStateFeeGrowth sdk.DecCoin) (liquidityDelta sdk.Dec, err error) {
	tickInfo, err := k.getTickInfo(ctx, poolId, tickIndex)
	if err != nil {
		return sdk.Dec{}, err
	}

	feeAccum, err := k.getFeeAccumulator(ctx, poolId)
	if err != nil {
		return sdk.Dec{}, err
	}

	// subtract tick's fee growth outside from current fee growth global, including the fee growth of the current swap.
	tickInfo.FeeGrowthOutside = feeAccum.GetValue().Add(swapStateFeeGrowth).Sub(tickInfo.FeeGrowthOutside)

	// Update global accums to now before uptime outside changes
	if err := k.updateUptimeAccumulatorsToNow(ctx, poolId); err != nil {
		return sdk.Dec{}, err
	}

	uptimeAccums, err := k.getUptimeAccumulators(ctx, poolId)
	if err != nil {
		return sdk.Dec{}, err
	}

	// For each supported uptime, subtract tick's uptime growth outside from the respective uptime accumulator
	// This is functionally equivalent to "flipping" the trackers once the tick is crossed
	updatedUptimeTrackers := tickInfo.UptimeTrackers
	for uptimeId, uptimeAccum := range uptimeAccums {
		updatedUptimeTrackers[uptimeId].UptimeGrowthOutside = uptimeAccum.GetValue().Sub(updatedUptimeTrackers[uptimeId].UptimeGrowthOutside)
	}

	k.SetTickInfo(ctx, poolId, tickIndex, tickInfo)

	return tickInfo.LiquidityNet, nil
}

// getTickInfo gets tickInfo given poolId and tickIndex. Returns a boolean field that returns true if value is found for given key.
func (k Keeper) getTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64) (tickInfo model.TickInfo, err error) {
	store := ctx.KVStore(k.storeKey)
	tickStruct := model.TickInfo{}
	key := types.KeyTick(poolId, tickIndex)
	if !k.poolExists(ctx, poolId) {
		return model.TickInfo{}, types.PoolNotFoundError{PoolId: poolId}
	}

	found, err := osmoutils.Get(store, key, &tickStruct)
	// return 0 values if key has not been initialized
	if !found {
		// If tick has not yet been initialized, we create a new one and initialize
		// the fee growth outside.
		initialFeeGrowthOutside, err := k.getInitialFeeGrowthOutsideForTick(ctx, poolId, tickIndex)
		if err != nil {
			return tickStruct, err
		}

		// Sync global uptime accumulators to ensure the uptime tracker init values are up to date.
		if err := k.updateUptimeAccumulatorsToNow(ctx, poolId); err != nil {
			return tickStruct, err
		}

		// Initialize uptime trackers for the new tick to the appropriate starting values.
		valuesToAdd, err := k.getInitialUptimeGrowthOutsidesForTick(ctx, poolId, tickIndex)
		if err != nil {
			return tickStruct, err
		}

		initialUptimeTrackers := []model.UptimeTracker{}
		for _, uptimeTrackerValue := range valuesToAdd {
			initialUptimeTrackers = append(initialUptimeTrackers, model.UptimeTracker{UptimeGrowthOutside: uptimeTrackerValue})
		}

		return model.TickInfo{LiquidityGross: sdk.ZeroDec(), LiquidityNet: sdk.ZeroDec(), FeeGrowthOutside: initialFeeGrowthOutside, UptimeTrackers: initialUptimeTrackers}, nil
	}
	if err != nil {
		return tickStruct, err
	}

	return tickStruct, nil
}

func (k Keeper) SetTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64, tickInfo model.TickInfo) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyTick(poolId, tickIndex)
	osmoutils.MustSet(store, key, &tickInfo)
}

// validateTickInRangeIsValid validates that given ticks are valid.
// That is, both lower and upper ticks are within MinTick and MaxTick range for the given exponentAtPriceOne.
// Also, lower tick must be less than upper tick.
// Returns error if validation fails. Otherwise, nil.
func validateTickRangeIsValid(tickSpacing uint64, exponentAtPriceOne sdk.Int, lowerTick int64, upperTick int64) error {
	minTick, maxTick := GetMinAndMaxTicksFromExponentAtPriceOne(exponentAtPriceOne)
	// Check if the lower and upper tick values are divisible by the tick spacing.
	if lowerTick%int64(tickSpacing) != 0 || upperTick%int64(tickSpacing) != 0 {
		return types.TickSpacingError{LowerTick: lowerTick, UpperTick: upperTick, TickSpacing: tickSpacing}
	}

	// Check if the lower tick value is within the valid range of MinTick to MaxTick.
	if lowerTick < minTick || lowerTick >= maxTick {
		return types.InvalidTickError{Tick: lowerTick, IsLower: true, MinTick: minTick, MaxTick: maxTick}
	}

	// Check if the upper tick value is within the valid range of MinTick to MaxTick.
	if upperTick > maxTick || upperTick <= minTick {
		return types.InvalidTickError{Tick: upperTick, IsLower: false, MinTick: minTick, MaxTick: maxTick}
	}

	// Check if the lower tick value is greater than or equal to the upper tick value.
	if lowerTick >= upperTick {
		return types.InvalidLowerUpperTickError{LowerTick: lowerTick, UpperTick: upperTick}
	}
	return nil
}

// GetMinAndMaxTicksFromExponentAtPriceOne determines min and max ticks allowed for a given exponentAtPriceOne value
// This allows for a min spot price of 0.000000000000000001 and a max spot price of 100000000000000000000000000000000000000 for every exponentAtPriceOne value
func GetMinAndMaxTicksFromExponentAtPriceOne(exponentAtPriceOne sdk.Int) (minTick, maxTick int64) {
	return math.GetMinAndMaxTicksFromExponentAtPriceOneInternal(exponentAtPriceOne)
}

// GetPerTickLiquidityDepthFromRange uses the given lower tick and upper tick, iterates over ticks, creates and returns LiquidityDepth array.
// LiquidityNet from the tick is used to indicate liquidity depths.
func (k Keeper) GetPerTickLiquidityDepthFromRange(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64) ([]types.LiquidityDepth, error) {
	if !k.poolExists(ctx, poolId) {
		return []types.LiquidityDepth{}, types.PoolNotFoundError{PoolId: poolId}
	}
	store := ctx.KVStore(k.storeKey)
	prefixBz := types.KeyTickPrefix(poolId)
	prefixStore := prefix.NewStore(store, prefixBz)

	lowerKey := types.TickIndexToBytes(lowerTick)
	upperKey := types.TickIndexToBytes(upperTick)
	iterator := prefixStore.Iterator(lowerKey, storetypes.InclusiveEndBytes(upperKey))

	liquidityDepths := []types.LiquidityDepth{}

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		tickIndex, err := types.TickIndexFromBytes(iterator.Key())
		if err != nil {
			return []types.LiquidityDepth{}, err
		}

		keyTick := types.KeyTick(poolId, tickIndex)
		tickStruct := model.TickInfo{}
		found, err := osmoutils.Get(store, keyTick, &tickStruct)
		if err != nil {
			return []types.LiquidityDepth{}, err
		}

		if !found {
			continue
		}

		liquidityDepth := types.LiquidityDepth{
			TickIndex:    sdk.NewInt(tickIndex),
			LiquidityNet: tickStruct.LiquidityNet,
		}
		liquidityDepths = append(liquidityDepths, liquidityDepth)
	}

	return liquidityDepths, nil
}
