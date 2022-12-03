package concentrated_liquidity

import (
	"errors"

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
// - the provided ticks are out of range / invalid
// - the pool provided does not exist
// - the liquidity delta is zero
// - the amount0 or amount1 returned from the position update is less than the given minimums
// - the pool or user does not have enough tokens to satisfy the requested amount
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

	// calculate liquidity created from this position
	liquidityDelta := math.GetLiquidityFromAmounts(pool.GetCurrentSqrtPrice(), sqrtPriceLowerTick, sqrtPriceUpperTick, amount0Desired, amount1Desired)
	if liquidityDelta.IsZero() {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, errors.New("liquidityDelta calculated equals zero")
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

	// send deposit amount from position owner to pool
	err = k.sendCoinsBetweenPoolAndUser(cacheCtx, pool.GetToken0(), pool.GetToken1(), actualAmount0, actualAmount1, owner, pool.GetAddress())
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
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

	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// send withdraw amount from pool to position owner
	err = k.sendCoinsBetweenPoolAndUser(ctx, pool.GetToken0(), pool.GetToken1(), actualAmount0, actualAmount1, pool.GetAddress(), owner)
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

// sendCoinsBetweenPoolAndUser takes the amounts calculated from a join/exit position and executes the send between pool and user
func (k Keeper) sendCoinsBetweenPoolAndUser(ctx sdk.Context, denom0, denom1 string, amount0, amount1 sdk.Int, sender, receiver sdk.AccAddress) error {
	var finalCoinsToSend sdk.Coins
	if amount0.IsPositive() {
		finalCoinsToSend = append(finalCoinsToSend, sdk.NewCoin(denom0, amount0))
	}
	if amount1.IsPositive() {
		finalCoinsToSend = append(finalCoinsToSend, sdk.NewCoin(denom1, amount1))
	}
	err := k.bankKeeper.SendCoins(ctx, sender, receiver, finalCoinsToSend)
	if err != nil {
		return err
	}
	return nil
}
