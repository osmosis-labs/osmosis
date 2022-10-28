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
func (k Keeper) createPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, amount0Desired, amount1Desired, amount0Min, amount1Min sdk.Int, lowerTick, upperTick int64) (amtDenom0, amtDenom1 sdk.Int, liquidityCreated sdk.Dec, err error) {
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

	// now calculate amount for token0 and token1
	pool := k.getPoolbyId(ctx, poolId)

	currentSqrtPrice := pool.CurrentSqrtPrice
	sqrtRatioUpperTick, _ := k.tickToSqrtPrice(sdk.NewInt(upperTick))
	sqrtRatioLowerTick, _ := k.tickToSqrtPrice(sdk.NewInt(lowerTick))

	liquidity := getLiquidityFromAmounts(currentSqrtPrice, sqrtRatioLowerTick, sqrtRatioUpperTick, amount0Desired, amount1Desired)
	if liquidity.IsZero() {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, fmt.Errorf("token in amount is zero")
	}

	// update tickInfo state
	// TODO: come back to sdk.Int vs sdk.Dec state & truncation
	err = k.initOrUpdateTick(ctx, poolId, lowerTick, liquidity.TruncateInt(), false)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}

	// TODO: come back to sdk.Int vs sdk.Dec state & truncation
	err = k.initOrUpdateTick(ctx, poolId, upperTick, liquidity.TruncateInt(), true)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}

	// update position state
	// TODO: come back to sdk.Int vs sdk.Dec state & truncation
	err = k.initOrUpdatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidity.TruncateInt())
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}

	// update pool state as well only if we're injecting liquidity at the current tick of the pool
	if pool.CurrentTick.LT(sdk.NewInt(lowerTick)) {
		amtDenom0 = calcAmount0Delta(liquidity, sqrtRatioLowerTick, sqrtRatioUpperTick).RoundInt()
		amtDenom1 = sdk.ZeroInt()
	} else if pool.CurrentTick.LT(sdk.NewInt(upperTick)) {
		amtDenom0 = calcAmount0Delta(liquidity, currentSqrtPrice, sqrtRatioUpperTick).RoundInt()
		amtDenom1 = calcAmount1Delta(liquidity, currentSqrtPrice, sqrtRatioLowerTick).RoundInt()
		pool.Liquidity = pool.Liquidity.Add(liquidity)
	} else {
		amtDenom0 = sdk.ZeroInt()
		amtDenom1 = calcAmount1Delta(liquidity, sqrtRatioLowerTick, sqrtRatioUpperTick).RoundInt()
	}

	k.setPoolById(ctx, pool.Id, pool)

	return amtDenom0, amtDenom1, liquidityCreated, nil
}

// withdrawPosition withdraws a concentrated liquidity position from the given pool id in the given tick range and liquidityAmount.
// On success, returns an amount of each token withdrawn.
// Returns error if
// - there is no position in the given tick ranges
// - if tick ranges are invalid
// - if attempts to withdraw an amount higher than originally provided in createPosition for a given range
// TODO: implement and table-driven tests
func (k Keeper) withdrawPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityAmount sdk.Int) (amtDenom0, amtDenom1 sdk.Int, err error) {
	panic("not implemented")
}
