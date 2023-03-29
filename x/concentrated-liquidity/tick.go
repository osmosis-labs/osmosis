package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/swapstrategy"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/genesis"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/query"
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

func (k Keeper) GetAllInitializedTicksForPool(ctx sdk.Context, poolId uint64) ([]genesis.FullTick, error) {
	return osmoutils.GatherValuesFromStorePrefixWithKeyParser(ctx.KVStore(k.storeKey), types.KeyTickPrefixByPoolId(poolId), ParseFullTickFromBytes)
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

// GetTickLiquidityForRangeInBatches returns an array of liquidity depth within the given range of lower tick and upper tick.
func (k Keeper) GetTickLiquidityForRange(ctx sdk.Context, poolId uint64) ([]query.LiquidityDepthWithRange, error) {
	// sanity check that pool exists and upper tick is greater than lower tick
	if !k.poolExists(ctx, poolId) {
		return []query.LiquidityDepthWithRange{}, types.PoolNotFoundError{PoolId: poolId}
	}

	// use false for zeroForOne since we're going from lower tick -> upper tick
	zeroForOne := false
	swapStrategy := swapstrategy.New(zeroForOne, sdk.ZeroDec(), k.storeKey, sdk.ZeroDec())

	// get min and max tick for the pool
	p, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return []query.LiquidityDepthWithRange{}, err
	}
	exponentAtPriceOne := p.GetExponentAtPriceOne()
	minTick, maxTick := math.GetMinAndMaxTicksFromExponentAtPriceOneInternal(exponentAtPriceOne)

	// set current tick to min tick, and find the first initialized tick starting from min tick -1.
	// we do -1 to make min tick inclusive.
	currentTick := minTick - 1

	nextTick, ok := swapStrategy.NextInitializedTick(ctx, poolId, currentTick)
	if !ok {
		return []query.LiquidityDepthWithRange{}, types.InvalidTickError{Tick: currentTick, IsLower: false, MinTick: minTick, MaxTick: maxTick}
	}

	tick, err := k.getTickByTickIndex(ctx, poolId, nextTick)
	if err != nil {
		return []query.LiquidityDepthWithRange{}, err
	}

	liquidityDepthsForRange := []query.LiquidityDepthWithRange{}

	// use the smallest tick initialized as the starting point for calculating liquidity.
	currentLiquidity := tick.LiquidityNet
	currentTick = nextTick.Int64()

	totalLiquidityWithinRange := currentLiquidity
	for currentTick <= maxTick {
		nextTick, ok := swapStrategy.NextInitializedTick(ctx, poolId, currentTick)
		// break and return the liquidity as is if
		// - there are no more next tick that is initialized,
		// - we hit upper limit
		if !ok {
			break
		}

		tick, err := k.getTickByTickIndex(ctx, poolId, nextTick)
		if err != nil {
			return []query.LiquidityDepthWithRange{}, err
		}

		liquidityDepthForRange := query.LiquidityDepthWithRange{
			LowerTick:       sdk.NewInt(currentTick),
			UpperTick:       nextTick,
			LiquidityAmount: totalLiquidityWithinRange,
		}
		liquidityDepthsForRange = append(liquidityDepthsForRange, liquidityDepthForRange)

		currentLiquidity = tick.LiquidityNet
		totalLiquidityWithinRange = totalLiquidityWithinRange.Add(currentLiquidity)
		currentTick = nextTick.Int64()
	}

	return liquidityDepthsForRange, nil
}

// GetLiquidityNetInDirection is a method that returns an array of TickLiquidityNet objects representing the net liquidity in a specified direction
// for a given pool. It provides an option to specify the bounds with start tick and bound tick.
// Swap direction is determined by the token in given (zero for one vs one for zero).
// See the swapstrategy package documentation for more details.
// Both start tick and bound tick must be in the appropriate range relative to the current tick and the min/max tick
// depending on the swap strategy and as defined by ValidateSqrtPrice method of the strategy.

// Parameters:

// * ctx (sdk.Context): The context for the SDK.
// * poolId (uint64): The ID of the pool for which the liquidity needs to be checked.
// * tokenIn (string): The token denom that determines the swap direction and strategy.
// * userGivenStartTick (sdk.Int): The starting tick for grabbing liquidities. If not provided, the current tick of the pool is used.
// * boundTick (sdk.Int): An optional bound tick to limit the range of the query. If not provided, the minimum or maximum tick will be used, depending on the strategy.
//
// Returns:

// ([]query.TickLiquidityNet): An array of TickLiquidityNet objects representing the net liquidity in the specified direction.
//
//	Note that the start tick is never included if given. The same goes for the current tick.
//	Returns liquidity net amounts starting from the next tick relative to start/current tick
//
// (error): An error if any issue occurs during the operation.
// Errors:
// * types.PoolNotFoundError: If the given pool does not exist.
// * types.TokenInDenomNotInPoolError: If the given tokenIn is not an asset in the pool.
func (k Keeper) GetTickLiquidityNetInDirection(ctx sdk.Context, poolId uint64, tokenIn string, userGivenStartTick sdk.Int, boundTick sdk.Int) ([]query.TickLiquidityNet, error) {
	// check if pool exists
	if !k.poolExists(ctx, poolId) {
		return []query.TickLiquidityNet{}, types.PoolNotFoundError{PoolId: poolId}
	}

	// get min and max tick for the pool
	p, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return []query.TickLiquidityNet{}, err
	}

	startTick := p.GetCurrentTick()
	// If start tick is set, use it as the current tick for grabbing liquidities from.
	if !userGivenStartTick.IsNil() {
		startTick = userGivenStartTick
	}

	// sanity check that given tokenIn is an asset in pool.
	if tokenIn != p.GetToken0() && tokenIn != p.GetToken1() {
		return []query.TickLiquidityNet{}, types.TokenInDenomNotInPoolError{TokenInDenom: tokenIn}
	}

	// figure out zero for one depending on the token in.
	zeroForOne := p.GetToken0() == tokenIn

	// use max or min tick if provided bound is nil
	exponentAtPriceOne := p.GetExponentAtPriceOne()
	minTick, maxTick := math.GetMinAndMaxTicksFromExponentAtPriceOneInternal(exponentAtPriceOne)
	if boundTick.IsNil() {
		if zeroForOne {
			boundTick = sdk.NewInt(minTick)
		} else {
			boundTick = sdk.NewInt(maxTick)
		}
	}

	liquidityDepths := []query.TickLiquidityNet{}
	swapStrategy := swapstrategy.New(zeroForOne, sdk.ZeroDec(), k.storeKey, sdk.ZeroDec())

	// function to validate that start tick and bound tick are
	// between current tick and the min/max tick depending on the swap direction.
	validateTickIsInValidRange := func(tick sdk.Int) error {
		boundSqrtPrice, err := math.TickToSqrtPrice(boundTick, exponentAtPriceOne)
		if err != nil {
			return err
		}

		if err := swapStrategy.ValidateSqrtPrice(boundSqrtPrice, p.GetCurrentSqrtPrice()); err != nil {
			return err
		}

		return nil
	}

	if err := validateTickIsInValidRange(boundTick); err != nil {
		return []query.TickLiquidityNet{}, err
	}

	if err := validateTickIsInValidRange(startTick); err != nil {
		return []query.TickLiquidityNet{}, err
	}

	checkTickWithInBoundFn := func(zeroForOne bool, currentTick, boundTick sdk.Int) bool {
		return (currentTick.GTE(boundTick) && zeroForOne) || (currentTick.LTE(boundTick) && !zeroForOne)
	}

	nextTick, ok := swapStrategy.NextInitializedTick(ctx, poolId, startTick.Int64())
	if !ok {
		return []query.TickLiquidityNet{}, nil
	}
	for checkTickWithInBoundFn(zeroForOne, nextTick, boundTick) {
		tick, err := k.getTickByTickIndex(ctx, poolId, nextTick)
		if err != nil {
			return []query.TickLiquidityNet{}, err
		}

		liquidityDepth := query.TickLiquidityNet{
			LiquidityNet: tick.LiquidityNet,
			TickIndex:    nextTick,
		}
		liquidityDepths = append(liquidityDepths, liquidityDepth)
		startTick = nextTick

		nextInitTick, ok := swapStrategy.NextInitializedTick(ctx, poolId, startTick.Int64())
		// break and return the liquidity as is if
		// - there are no more next tick that is initialized,
		// - we hit upper limit

		if !ok {
			break
		}
		nextTick = nextInitTick
	}

	return liquidityDepths, nil
}

func (k Keeper) getTickByTickIndex(ctx sdk.Context, poolId uint64, tickIndex sdk.Int) (model.TickInfo, error) {
	store := ctx.KVStore(k.storeKey)
	keyTick := types.KeyTick(poolId, tickIndex.Int64())
	tickStruct := model.TickInfo{}
	found, err := osmoutils.Get(store, keyTick, &tickStruct)
	if err != nil {
		return model.TickInfo{}, err
	}
	if !found {
		return model.TickInfo{}, types.TickNotFoundError{Tick: tickIndex.Int64()}
	}
	return tickStruct, nil
}
