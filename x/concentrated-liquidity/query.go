package concentrated_liquidity

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/client/queryproto"
	"github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/swapstrategy"
	types "github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/types"
)

// This file contains query-related helper functions for the Concentrated Liquidity module

// GetTickLiquidityForFullRange returns an array of liquidity depth for all ticks existing from min tick ~ max tick.
func (k Keeper) GetTickLiquidityForFullRange(ctx sdk.Context, poolId uint64) ([]queryproto.LiquidityDepthWithRange, error) {
	// use false for zeroForOne since we're going from lower tick -> upper tick
	zeroForOne := false
	swapStrategy := swapstrategy.New(zeroForOne, sdk.ZeroDec(), k.storeKey, sdk.ZeroDec())

	// set current tick to min tick, and find the first initialized tick starting from min tick -1.
	// we do -1 to make min tick inclusive.
	currentTick := types.MinCurrentTick

	nextTickIter := swapStrategy.InitializeNextTickIterator(ctx, poolId, currentTick)
	defer nextTickIter.Close()
	if !nextTickIter.Valid() {
		return []queryproto.LiquidityDepthWithRange{}, types.RanOutOfTicksForPoolError{PoolId: poolId}
	}

	nextTick, err := types.TickIndexFromBytes(nextTickIter.Key())
	if err != nil {
		return []queryproto.LiquidityDepthWithRange{}, err
	}

	tick, err := k.getTickByTickIndex(ctx, poolId, nextTick)
	if err != nil {
		return []queryproto.LiquidityDepthWithRange{}, err
	}

	liquidityDepthsForRange := []queryproto.LiquidityDepthWithRange{}

	// use the smallest tick initialized as the starting point for calculating liquidity.
	currentLiquidity := tick.LiquidityNet
	currentTick = nextTick
	totalLiquidityWithinRange := currentLiquidity

	previousTickIndex := currentTick

	// start from the next index so that the current tick can become lower tick.
	nextTickIter.Next()
	for ; nextTickIter.Valid(); nextTickIter.Next() {
		tickIndex, err := types.TickIndexFromBytes(nextTickIter.Key())
		if err != nil {
			return []queryproto.LiquidityDepthWithRange{}, err
		}

		tickStruct, err := ParseTickFromBz(nextTickIter.Value())
		if err != nil {
			return []queryproto.LiquidityDepthWithRange{}, err
		}

		liquidityDepthForRange := queryproto.LiquidityDepthWithRange{
			LowerTick:       previousTickIndex,
			UpperTick:       tickIndex,
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

	ctx.Logger().Debug(fmt.Sprintf("userGivenStartTick %s, boundTick %s, currentTick %d\n", userGivenStartTick, boundTick, p.GetCurrentTick()))

	startTick := p.GetCurrentTick()
	// If start tick is set, use it as the current tick for grabbing liquidities from.
	if !userGivenStartTick.IsNil() {
		startTick = userGivenStartTick.Int64()
		ctx.Logger().Debug(fmt.Sprintf("startTick %d set to user given\n", startTick))
	}

	// sanity check that given tokenIn is an asset in pool.
	if tokenIn != p.GetToken0() && tokenIn != p.GetToken1() {
		return []queryproto.TickLiquidityNet{}, types.TokenInDenomNotInPoolError{TokenInDenom: tokenIn}
	}

	// figure out zero for one depending on the token in.
	zeroForOne := p.GetToken0() == tokenIn

	ctx.Logger().Debug(fmt.Sprintf("is_zero_for_one %t\n", zeroForOne))

	// use max or min tick if provided bound is nil

	ctx.Logger().Debug(fmt.Sprintf("min_tick %d\n", types.MinInitializedTick))
	ctx.Logger().Debug(fmt.Sprintf("max_tick %d\n", types.MaxTick))

	if boundTick.IsNil() {
		if zeroForOne {
			boundTick = sdk.NewInt(types.MinInitializedTick)
		} else {
			boundTick = sdk.NewInt(types.MaxTick)
		}
	}

	liquidityDepths := []queryproto.TickLiquidityNet{}
	swapStrategy := swapstrategy.New(zeroForOne, sdk.ZeroDec(), k.storeKey, sdk.ZeroDec())

	currentTick := p.GetCurrentTick()
	_, currentTickSqrtPrice, err := math.TickToSqrtPrice(currentTick)
	if err != nil {
		return []queryproto.TickLiquidityNet{}, err
	}

	ctx.Logger().Debug(fmt.Sprintf("currentTick %d; current tick's sqrt price%s\n", currentTick, currentTickSqrtPrice))

	// function to validate that start tick and bound tick are
	// between current tick and the min/max tick depending on the swap direction.
	validateTickIsInValidRange := func(validateTick sdk.Int) error {
		_, validateSqrtPrice, err := math.TickToSqrtPrice(validateTick.Int64())
		if err != nil {
			return err
		}
		ctx.Logger().Debug(fmt.Sprintf("validateTick %s; validate sqrtPrice %s\n", validateTick.String(), validateSqrtPrice.String()))

		if err := swapStrategy.ValidateSqrtPrice(validateSqrtPrice, osmomath.BigDecFromSDKDec(currentTickSqrtPrice)); err != nil {
			return err
		}

		return nil
	}

	ctx.Logger().Debug("validating bound tick")
	if err := validateTickIsInValidRange(boundTick); err != nil {
		return []queryproto.TickLiquidityNet{}, fmt.Errorf("failed validating bound tick (%s) with current sqrt price of (%s): %w", boundTick, currentTickSqrtPrice, err)
	}

	ctx.Logger().Debug("validating start tick")
	if err := validateTickIsInValidRange(sdk.NewInt(startTick)); err != nil {
		return []queryproto.TickLiquidityNet{}, fmt.Errorf("failed validating start tick (%d) with current sqrt price of (%s): %w", startTick, currentTickSqrtPrice, err)
	}

	// iterator assignments
	store := ctx.KVStore(k.storeKey)
	prefixBz := types.KeyTickPrefixByPoolId(poolId)
	prefixStore := prefix.NewStore(store, prefixBz)

	// If zero for one, we use reverse iterator. As a result, we need to increment the start tick by 1
	// so that we include the start tick in the search.
	//
	// If one for zero, we use forward iterator. However, our definition of the active range is inclusive
	// of the lower bound. As a result, current liquidity must already include the lower bound tick
	// so we skip it.
	startTick = startTick + 1

	startTickKey := types.TickIndexToBytes(startTick)
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

		tickStruct, err := ParseTickFromBz(iterator.Value())
		if err != nil {
			return []queryproto.TickLiquidityNet{}, err
		}

		liquidityDepth := queryproto.TickLiquidityNet{
			LiquidityNet: tickStruct.LiquidityNet,
			TickIndex:    tickIndex,
		}
		liquidityDepths = append(liquidityDepths, liquidityDepth)
	}

	return liquidityDepths, nil
}

func (k Keeper) getTickByTickIndex(ctx sdk.Context, poolId uint64, tickIndex int64) (model.TickInfo, error) {
	store := ctx.KVStore(k.storeKey)
	keyTick := types.KeyTick(poolId, tickIndex)
	tickStruct := model.TickInfo{}
	found, err := osmoutils.Get(store, keyTick, &tickStruct)
	if err != nil {
		return model.TickInfo{}, err
	}
	if !found {
		return model.TickInfo{}, types.TickNotFoundError{Tick: tickIndex}
	}
	return tickStruct, nil
}
