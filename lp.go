package concentrated_liquidity

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/math"
	types "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
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
	if err := validateTickRangeIsValid(lowerTick, upperTick); err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}

	sqrtPriceLowerTick, sqrtPriceUpperTick, err := math.TicksToSqrtPrice(lowerTick, upperTick)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}

	// now calculate amount for token0 and token1
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}

	liquidityDelta := math.GetLiquidityFromAmounts(pool.GetCurrentSqrtPrice(), sqrtPriceLowerTick, sqrtPriceUpperTick, amount0Desired, amount1Desired)
	if liquidityDelta.IsZero() {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, fmt.Errorf("liquidity delta zero")
	}

	// N.B. we only write cache context if actual amounts
	// returned are greater than the given minimums.
	cacheCtx, writeCacheCtx := ctx.CacheContext()

	actualAmount0, actualAmount1, err := k.updatePosition(cacheCtx, poolId, owner, lowerTick, upperTick, liquidityDelta)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}

	if actualAmount0.LT(amount0Min) {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, types.InsufficientLiquidityCreatedError{Actual: actualAmount0, Minimum: amount0Min, IsTokenZero: true}
	}

	if actualAmount1.LT(amount1Min) {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, types.InsufficientLiquidityCreatedError{Actual: actualAmount1, Minimum: amount1Min}
	}

	// only persist updates if amount validation passed.
	writeCacheCtx()

	return actualAmount0, actualAmount1, liquidityDelta, nil
}

// withdrawPosition attempts to withdraw liquidityAmount from a position with the given pool id in the given tick range.
// On success, returns a positive amount of each token withdrawn.
// Returns error if
// - there is no position in the given tick ranges
// - if tick ranges are invalid
// - if attempts to withdraw an amount higher than originally provided in createPosition for a given range.
func (k Keeper) withdrawPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, requestedLiqudityAmountToWithdraw sdk.Dec) (amtDenom0, amtDenom1 sdk.Int, err error) {
	if err := validateTickRangeIsValid(lowerTick, upperTick); err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	position, err := k.getPosition(ctx, poolId, owner, lowerTick, upperTick)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	availableLiquidity := position.Liquidity

	if requestedLiqudityAmountToWithdraw.GT(availableLiquidity) {
		return sdk.Int{}, sdk.Int{}, types.InsufficientLiquidityError{Actual: requestedLiqudityAmountToWithdraw, Available: availableLiquidity}
	}

	liquidityDelta := requestedLiqudityAmountToWithdraw.Neg()

	actualAmount0, actualAmount1, err := k.updatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidityDelta)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	return actualAmount0.Neg(), actualAmount1.Neg(), nil
}

// updatePosition updates the position in the given pool id and in the given tick range and liquidityAmount.
// Negative liquidityDelta implies withdrawing liquidity.
// Positive liquidityDelta implies adding liquidity.
// Updates ticks and pool liquidity. Returns how much of each token is either added or removed.
// Negative returned amounts imply that tokens are removed from the pool.
// Positive returned amounts imply that tokens are added to the pool.
// TODO: tests.
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
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	sqrtPriceLowerTick, sqrtPriceUpperTick, err := math.TicksToSqrtPrice(lowerTick, upperTick)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	actualAmount0, actualAmount1 := pool.CalcActualAmounts(ctx, lowerTick, upperTick, sqrtPriceLowerTick, sqrtPriceUpperTick, liquidityDelta)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	pool.UpdateLiquidityIfActivePosition(ctx, lowerTick, upperTick, liquidityDelta)

	if err := k.setPool(ctx, pool); err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// The returned amounts are rounded down to avoid returning more to clients than they actually deposited.
	return actualAmount0.TruncateInt(), actualAmount1.TruncateInt(), nil
}
