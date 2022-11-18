package concentrated_liquidity

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

// createPosition creates a concentrated liquidity position in range between lowerTick and upperTick
// in a given `PoolId with the desired amount of each token. Since LPs are only allowed to provide
// liquidity proportional to the existing reserves, the actual amount of tokens used might differ from requested.
// As a result, LPs may also provide the minimum amount of each token to be used so that the system fails
// to create position if the desired amounts cannot be satisfied.
// On success, returns an actual amount of each token used and liquidity created.
// Returns error if:
// TODO: list error cases
// TODO: table-driven tests
func (k Keeper) createPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, amount0Desired, amount1Desired, amount0Min, amount1Min sdk.Int, lowerTick, upperTick int64) (sdk.Int, sdk.Int, sdk.Dec, error) {
	// ensure types.MinTick <= lowerTick < types.MaxTick
	// TODO (bez): Add unit tests.
	if lowerTick < types.MinTick || lowerTick >= types.MaxTick {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, fmt.Errorf("invalid lower tick: %d", lowerTick)
	}

	// ensure types.MaxTick < upperTick <= types.MinTick
	// TODO (bez): Add unit tests.
	if upperTick > types.MaxTick || upperTick <= types.MinTick {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, fmt.Errorf("invalid upper tick: %d", upperTick)
	}

	sqrtPriceLowerTick, sqrtPriceUpperTick, err := ticksToSqrtPrice(lowerTick, upperTick)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}

	// now calculate amount for token0 and token1
	pool := k.getPoolbyId(ctx, poolId)

	liquidityDelta := getLiquidityFromAmounts(pool.CurrentSqrtPrice, sqrtPriceLowerTick, sqrtPriceUpperTick, amount0Desired, amount1Desired)
	if liquidityDelta.IsZero() {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, fmt.Errorf("liquidity delta zero")
	}

	actualAmount0, actualAmount1, err := k.updatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidityDelta)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}

	// TODO: handle amount0Min, amount1Min

	return actualAmount0, actualAmount1, liquidityDelta, nil
}

// withdrawPosition withdraws a concentrated liquidity position from the given pool id in the given tick range and liquidityAmount.
// On success, returns an amount of each token withdrawn.
// Returns error if
// - there is no position in the given tick ranges
// - if tick ranges are invalid
// - if attempts to withdraw an amount higher than originally provided in createPosition for a given range
// TODO: implement and table-driven tests
func (k Keeper) withdrawPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, requestedLiqudityAmountToWithdraw sdk.Int) (amtDenom0, amtDenom1 sdk.Int, err error) {

	// check if requested liquidiyt matches available

	position, err := k.GetPosition(ctx, poolId, owner, lowerTick, upperTick)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	availableLiquidity := position.Liquidity

	if requestedLiqudityAmountToWithdraw.ToDec().GT(availableLiquidity) {
		// TODO: format error correctly.
		return sdk.Int{}, sdk.Int{}, fmt.Errorf("cannot withdraw more than available")
	}

	liquidityDelta := availableLiquidity.Sub(requestedLiqudityAmountToWithdraw.ToDec())

	actualAmount0, actualAmount1, err := k.updatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidityDelta)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	return actualAmount0, actualAmount1, nil
}

// TODO: spec and tests.
func (k Keeper) updatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityDelta sdk.Dec) (sdk.Int, sdk.Int, error) {

	// update tickInfo state
	// TODO: come back to sdk.Int vs sdk.Dec state & truncation
	err := k.initOrUpdateTick(ctx, poolId, lowerTick, liquidityDelta, false)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// TODO: come back to sdk.Int vs sdk.Dec state & truncation
	err = k.initOrUpdateTick(ctx, poolId, upperTick, liquidityDelta, true)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// update position state
	// TODO: come back to sdk.Int vs sdk.Dec state & truncation
	err = k.initOrUpdatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidityDelta)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// now calculate amount for token0 and token1
	pool := k.getPoolbyId(ctx, poolId)

	sqrtPriceLowerTick, sqrtPriceUpperTick, err := ticksToSqrtPrice(lowerTick, upperTick)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	actualAmount0, actualAmount1 := pool.calcActualAmounts(ctx, lowerTick, upperTick, sqrtPriceLowerTick, sqrtPriceUpperTick, liquidityDelta)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	pool.updateLiquidityIfActivePosition(ctx, lowerTick, upperTick, liquidityDelta)

	k.setPoolById(ctx, pool.Id, pool)

	return actualAmount0, actualAmount1, nil
}

// calcActualAmounts calculates and returns actual amounts based on where the current tick is located relative to position's
// lower and upper ticks.
// There are 3 possible cases:
// -The position is active ( lowerTick <= p.CurrentTick < upperTick).
//    * The provided liqudity is distributed in both tokens.
//    * Actual amounts might differ from desired because we recalculate them from liquidity delta and sqrt price.
//      the calculations lead to amounts being off. // TODO: confirm logic is correct
// - Current tick is below the position ( p.CurrentTick < lowerTick).
//    * The provided liquidity is distributed in token0 only.
// - Current tick is above the position ( p.CurrentTick >= p.upperTick ).
//    * The provided liquidity is distributed in token1 only.
// TODO: add tests.
func (p Pool) calcActualAmounts(ctx sdk.Context, lowerTick, upperTick int64, sqrtRatioLowerTick, sqrtRatioUpperTick sdk.Dec, liquidityDelta sdk.Dec) (actualAmountDenom0 sdk.Int, actualAmountDenom1 sdk.Int) {

	if p.isPositionActive(lowerTick, upperTick) {
		// outcome one: the current price falls within the position
		// if this is the case, we attempt to provide liquidity evenly between asset0 and asset1
		// we also update the pool liquidity since the virtual liquidity is modified by this position's creation
		currentSqrtPrice := p.CurrentSqrtPrice
		actualAmountDenom0 = calcAmount0Delta(liquidityDelta, currentSqrtPrice, sqrtRatioUpperTick, false).RoundInt()
		actualAmountDenom1 = calcAmount1Delta(liquidityDelta, currentSqrtPrice, sqrtRatioLowerTick, false).RoundInt()
	} else if p.CurrentTick.LT(sdk.NewInt(lowerTick)) {
		// outcome two: position is below current price
		// this means position is solely made up of asset0
		actualAmountDenom1 = sdk.ZeroInt()
		actualAmountDenom0 = calcAmount0Delta(liquidityDelta, sqrtRatioLowerTick, sqrtRatioUpperTick, false).RoundInt()
	} else {
		// outcome three: position is above current price
		// this means position is solely made up of asset1
		actualAmountDenom0 = sdk.ZeroInt()
		actualAmountDenom1 = calcAmount1Delta(liquidityDelta, sqrtRatioLowerTick, sqrtRatioUpperTick, false).RoundInt()
	}

	return actualAmountDenom0, actualAmountDenom1
}

// TODO: add tests.
func (p Pool) isPositionActive(lowerTick, upperTick int64) bool {
	return p.CurrentTick.GTE(sdk.NewInt(lowerTick)) && p.CurrentTick.LT(sdk.NewInt(upperTick))
}

// TODO: add tests.
func (p *Pool) updateLiquidityIfActivePosition(ctx sdk.Context, lowerTick, upperTick int64, liquidityDelta sdk.Dec) bool {
	if p.isPositionActive(lowerTick, upperTick) {
		p.Liquidity = p.Liquidity.Add(liquidityDelta)
		return true
	}
	return false
}

// TODO: spec and tests
func ticksToSqrtPrice(lowerTick, upperTick int64) (sdk.Dec, sdk.Dec, error) {
	sqrtPriceUpperTick, err := tickToSqrtPrice(sdk.NewInt(upperTick))
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	sqrtPriceLowerTick, err := tickToSqrtPrice(sdk.NewInt(lowerTick))
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	return sqrtPriceLowerTick, sqrtPriceUpperTick, nil
}
