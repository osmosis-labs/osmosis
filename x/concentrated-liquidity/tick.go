package concentrated_liquidity

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	db "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/client/queryproto"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/swapstrategy"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/genesis"
)

// initOrUpdateTick retrieves the tickInfo from the specified tickIndex and updates both the liquidityNet and LiquidityGross.
// The given currentTick value is used to determine the strategy for updating the fee accumulator.
// We update the tick's fee growth opposite direction of last traversal accumulator to the fee growth global when tick index is <= current tick.
// Otherwise, it is set to zero.
// if we are initializing or updating an upper tick, we subtract the liquidityIn from the LiquidityNet
// if we are initializing or updating a lower tick, we add the liquidityIn from the LiquidityNet
// WARNING: this method may mutate the pool, make sure to refetch the pool after calling this method.
func (k Keeper) initOrUpdateTick(ctx sdk.Context, poolId uint64, currentTick int64, tickIndex int64, liquidityIn sdk.Dec, upper bool) (err error) {
	tickInfo, err := k.GetTickInfo(ctx, poolId, tickIndex)
	if err != nil {
		return err
	}

	// If both liquidity fields are zero, we consume the base gas fee for initializing a tick.
	if tickInfo.LiquidityGross.IsZero() && tickInfo.LiquidityNet.IsZero() {
		ctx.GasMeter().ConsumeGas(uint64(types.BaseGasFeeForInitializingTick), "initialize tick gas fee")
	}

	// calculate liquidityGross, which does not care about whether liquidityIn is positive or negative
	liquidityBefore := tickInfo.LiquidityGross

	// if given tickIndex is LTE to the current tick and the liquidityBefore is zero,
	// set the tick's fee growth opposite direction of last traversal to the fee accumulator's value
	if liquidityBefore.IsZero() {
		if tickIndex <= currentTick {
			accum, err := k.GetFeeAccumulator(ctx, poolId)
			if err != nil {
				return err
			}

			tickInfo.FeeGrowthOppositeDirectionOfLastTraversal = accum.GetValue()
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

// WARNING: this method may mutate the pool, make sure to refetch the pool after calling this method.
func (k Keeper) crossTick(ctx sdk.Context, poolId uint64, tickIndex int64, swapStateFeeGrowth sdk.DecCoin) (liquidityDelta sdk.Dec, err error) {
	tickInfo, err := k.GetTickInfo(ctx, poolId, tickIndex)
	if err != nil {
		return sdk.Dec{}, err
	}

	feeAccum, err := k.GetFeeAccumulator(ctx, poolId)
	if err != nil {
		return sdk.Dec{}, err
	}

	// subtract tick's fee growth opposite direction of last traversal from current fee growth global, including the fee growth of the current swap.
	tickInfo.FeeGrowthOppositeDirectionOfLastTraversal = feeAccum.GetValue().Add(swapStateFeeGrowth).Sub(tickInfo.FeeGrowthOppositeDirectionOfLastTraversal)

	// Update global accums to now before uptime outside changes
	if err := k.updatePoolUptimeAccumulatorsToNow(ctx, poolId); err != nil {
		return sdk.Dec{}, err
	}

	uptimeAccums, err := k.GetUptimeAccumulators(ctx, poolId)
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

// GetTickInfo gets the tickInfo given a poolId and tickIndex. If the tick has not been initialized, it will initialize it.
// If the tick has been initialized, it will return the tickInfo. If the pool does not exist, it will return an error.
// WARNING: this method may mutate the pool, make sure to refetch the pool after calling this method.
func (k Keeper) GetTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64) (tickInfo model.TickInfo, err error) {
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
		// the fee growth opposite direction of last traversal value.
		initialFeeGrowthOppositeDirectionOfLastTraversal, err := k.getInitialFeeGrowthOppositeDirectionOfLastTraversalForTick(ctx, poolId, tickIndex)
		if err != nil {
			return tickStruct, err
		}

		// Sync global uptime accumulators to ensure the uptime tracker init values are up to date.
		if err := k.updatePoolUptimeAccumulatorsToNow(ctx, poolId); err != nil {
			return tickStruct, err
		}

		// Initialize uptime trackers for the new tick to the appropriate starting values.
		valuesToAdd, err := k.getInitialUptimeGrowthOppositeDirectionOfLastTraversalForTick(ctx, poolId, tickIndex)
		if err != nil {
			return tickStruct, err
		}

		initialUptimeTrackers := []model.UptimeTracker{}
		for _, uptimeTrackerValue := range valuesToAdd {
			initialUptimeTrackers = append(initialUptimeTrackers, model.UptimeTracker{UptimeGrowthOutside: uptimeTrackerValue})
		}

		return model.TickInfo{LiquidityGross: sdk.ZeroDec(), LiquidityNet: sdk.ZeroDec(), FeeGrowthOppositeDirectionOfLastTraversal: initialFeeGrowthOppositeDirectionOfLastTraversal, UptimeTrackers: initialUptimeTrackers}, nil
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
	if lowerTick < types.MinTick || lowerTick >= types.MaxTick {
		return types.InvalidTickError{Tick: lowerTick, IsLower: true, MinTick: types.MinTick, MaxTick: types.MaxTick}
	}

	// Check if the upper tick value is within the valid range of MinTick to MaxTick.
	if upperTick > types.MaxTick || upperTick <= types.MinTick {
		return types.InvalidTickError{Tick: upperTick, IsLower: false, MinTick: types.MinTick, MaxTick: types.MaxTick}
	}

	// Check if the lower tick value is greater than or equal to the upper tick value.
	if lowerTick >= upperTick {
		return types.InvalidLowerUpperTickError{LowerTick: lowerTick, UpperTick: upperTick}
	}
	return nil
}

// GetTickLiquidityForFullRange returns an array of liquidity depth for all ticks existing from min tick ~ max tick.
func (k Keeper) GetTickLiquidityForFullRange(ctx sdk.Context, poolId uint64) ([]queryproto.LiquidityDepthWithRange, error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return []queryproto.LiquidityDepthWithRange{}, err
	}

	// use false for zeroForOne since we're going from lower tick -> upper tick
	zeroForOne := false
	swapStrategy := swapstrategy.New(zeroForOne, sdk.ZeroDec(), k.storeKey, sdk.ZeroDec(), pool.GetTickSpacing())

	// set current tick to min tick, and find the first initialized tick starting from min tick -1.
	// we do -1 to make min tick inclusive.
	currentTick := types.MinTick - 1

	nextTick, ok := swapStrategy.NextInitializedTick(ctx, poolId, currentTick)
	if !ok {
		return []queryproto.LiquidityDepthWithRange{}, types.InvalidTickError{Tick: currentTick, IsLower: false, MinTick: types.MinTick, MaxTick: types.MaxTick}
	}

	tick, err := k.getTickByTickIndex(ctx, poolId, nextTick)
	if err != nil {
		return []queryproto.LiquidityDepthWithRange{}, err
	}

	liquidityDepthsForRange := []queryproto.LiquidityDepthWithRange{}

	// use the smallest tick initialized as the starting point for calculating liquidity.
	currentLiquidity := tick.LiquidityNet
	currentTick = nextTick.Int64()
	totalLiquidityWithinRange := currentLiquidity

	// iterator assignments
	store := ctx.KVStore(k.storeKey)
	prefixBz := types.KeyTickPrefixByPoolId(poolId)
	prefixStore := prefix.NewStore(store, prefixBz)
	currentTickKey := types.TickIndexToBytes(currentTick)
	maxTickKey := types.TickIndexToBytes(types.MaxTick)
	iterator := prefixStore.Iterator(currentTickKey, storetypes.InclusiveEndBytes(maxTickKey))
	defer iterator.Close()
	previousTickIndex := currentTick

	// start from the next index so that the current tick can become lower tick.
	iterator.Next()
	for ; iterator.Valid(); iterator.Next() {
		tickIndex, err := types.TickIndexFromBytes(iterator.Key())
		if err != nil {
			return []queryproto.LiquidityDepthWithRange{}, err
		}

		tickStruct := model.TickInfo{}
		err = proto.Unmarshal(iterator.Value(), &tickStruct)
		if err != nil {
			return []queryproto.LiquidityDepthWithRange{}, err
		}

		liquidityDepthForRange := queryproto.LiquidityDepthWithRange{
			LowerTick:       sdk.NewInt(previousTickIndex),
			UpperTick:       sdk.NewInt(tickIndex),
			LiquidityAmount: totalLiquidityWithinRange,
		}
		liquidityDepthsForRange = append(liquidityDepthsForRange, liquidityDepthForRange)

		currentLiquidity = tickStruct.LiquidityNet

		previousTickIndex = tickIndex
		totalLiquidityWithinRange = totalLiquidityWithinRange.Add(currentLiquidity)
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
// * boundTick (sdk.Int): An optional bound tick to limit the range of the queryproto. If not provided, the minimum or maximum tick will be used, depending on the strategy.
//
// Returns:

// ([]queryproto.TickLiquidityNet): An array of TickLiquidityNet objects representing the net liquidity in the specified direction.
//
//	Note that the start tick is never included if given. The same goes for the current tick.
//	Returns liquidity net amounts starting from the next tick relative to start/current tick
//
// (error): An error if any issue occurs during the operation.
// Errors:
// * types.PoolNotFoundError: If the given pool does not exist.
// * types.TokenInDenomNotInPoolError: If the given tokenIn is not an asset in the pool.
func (k Keeper) GetTickLiquidityNetInDirection(ctx sdk.Context, poolId uint64, tokenIn string, userGivenStartTick sdk.Int, boundTick sdk.Int) ([]queryproto.TickLiquidityNet, error) {
	// get min and max tick for the pool
	p, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return []queryproto.TickLiquidityNet{}, err
	}

	ctx.Logger().Debug(fmt.Sprintf("userGivenStartTick %s, boundTick %s, currentTick %s\n", userGivenStartTick, boundTick, p.GetCurrentTick()))

	startTick := p.GetCurrentTick()
	// If start tick is set, use it as the current tick for grabbing liquidities from.
	if !userGivenStartTick.IsNil() {
		startTick = userGivenStartTick
		ctx.Logger().Debug(fmt.Sprintf("startTick %s set to user given\n", startTick))
	}

	// sanity check that given tokenIn is an asset in pool.
	if tokenIn != p.GetToken0() && tokenIn != p.GetToken1() {
		return []queryproto.TickLiquidityNet{}, types.TokenInDenomNotInPoolError{TokenInDenom: tokenIn}
	}

	// figure out zero for one depending on the token in.
	zeroForOne := p.GetToken0() == tokenIn

	ctx.Logger().Debug(fmt.Sprintf("is_zero_for_one %t\n", zeroForOne))

	// use max or min tick if provided bound is nil

	ctx.Logger().Debug(fmt.Sprintf("min_tick %d\n", types.MinTick))
	ctx.Logger().Debug(fmt.Sprintf("max_tick %d\n", types.MaxTick))

	if boundTick.IsNil() {
		if zeroForOne {
			boundTick = sdk.NewInt(types.MinTick)
		} else {
			boundTick = sdk.NewInt(types.MaxTick)
		}
	}

	liquidityDepths := []queryproto.TickLiquidityNet{}
	swapStrategy := swapstrategy.New(zeroForOne, sdk.ZeroDec(), k.storeKey, sdk.ZeroDec(), p.GetTickSpacing())

	currentTick := p.GetCurrentTick()
	currentTickSqrtPrice, err := math.TickToSqrtPrice(currentTick)
	if err != nil {
		return []queryproto.TickLiquidityNet{}, err
	}

	ctx.Logger().Debug(fmt.Sprintf("currentTick %s; current tick's sqrt price%s\n", currentTick, currentTickSqrtPrice))

	// function to validate that start tick and bound tick are
	// between current tick and the min/max tick depending on the swap direction.
	validateTickIsInValidRange := func(validateTick sdk.Int) error {
		validateSqrtPrice, err := math.TickToSqrtPrice(validateTick)
		if err != nil {
			return err
		}
		ctx.Logger().Debug(fmt.Sprintf("validateTick %s; validate sqrtPrice %s\n", validateTick.String(), validateSqrtPrice.String()))

		if err := swapStrategy.ValidateSqrtPrice(validateSqrtPrice, currentTickSqrtPrice); err != nil {
			return err
		}

		return nil
	}

	ctx.Logger().Debug("validating bound tick")
	if err := validateTickIsInValidRange(boundTick); err != nil {
		return []queryproto.TickLiquidityNet{}, fmt.Errorf("failed validating bound tick (%s) with current sqrt price of (%s): %w", boundTick, currentTickSqrtPrice, err)
	}

	ctx.Logger().Debug("validating start tick")
	if err := validateTickIsInValidRange(startTick); err != nil {
		return []queryproto.TickLiquidityNet{}, fmt.Errorf("failed validating start tick (%s) with current sqrt price of (%s): %w", startTick, currentTickSqrtPrice, err)
	}

	// iterator assignments
	store := ctx.KVStore(k.storeKey)
	prefixBz := types.KeyTickPrefixByPoolId(poolId)
	prefixStore := prefix.NewStore(store, prefixBz)
	startTickKey := types.TickIndexToBytes(startTick.Int64())
	boundTickKey := types.TickIndexToBytes(boundTick.Int64())

	// define iterator depending on swap strategy
	var iterator db.Iterator
	if zeroForOne {
		iterator = prefixStore.ReverseIterator(boundTickKey, startTickKey)
	} else {
		iterator = prefixStore.Iterator(startTickKey, storetypes.InclusiveEndBytes(boundTickKey))
	}

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		tickIndex, err := types.TickIndexFromBytes(iterator.Key())
		if err != nil {
			return []queryproto.TickLiquidityNet{}, err
		}

		tickStruct := model.TickInfo{}
		err = proto.Unmarshal(iterator.Value(), &tickStruct)
		if err != nil {
			return []queryproto.TickLiquidityNet{}, err
		}

		liquidityDepth := queryproto.TickLiquidityNet{
			LiquidityNet: tickStruct.LiquidityNet,
			TickIndex:    sdk.NewInt(tickIndex),
		}
		liquidityDepths = append(liquidityDepths, liquidityDepth)
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
