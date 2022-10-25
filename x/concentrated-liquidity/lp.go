package concentrated_liquidity

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (k Keeper) createPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, amount0Desired, amount1Desired, amount0Min, amount1Min sdk.Int, lowerTick, upperTick int64) (amtDenom0, amtDenom1 sdk.Int, err error) {
	// TODO: calculate from amounts given
	liquidityIn := sdk.MustNewDecFromStr("1517.882323")

	// ensure types.MinTick <= lowerTick < types.MaxTick
	// TODO (bez): Add unit tests.
	if lowerTick < types.MinTick || lowerTick >= types.MaxTick {
		return sdk.Int{}, sdk.Int{}, fmt.Errorf("invalid lower tick: %d", lowerTick)
	}

	// ensure types.MaxTick < upperTick <= types.MinTick
	// TODO (bez): Add unit tests.
	if upperTick > types.MaxTick || upperTick <= types.MinTick {
		return sdk.Int{}, sdk.Int{}, fmt.Errorf("invalid upper tick: %d", upperTick)
	}

	if liquidityIn.IsZero() {
		return sdk.Int{}, sdk.Int{}, fmt.Errorf("token in amount is zero")
	}
	// TODO: we need to check if TickInfo is initialized on a specific tick before updating, otherwise get wont work
	// k.setTickInfo(ctx, poolId, lowerTick, TickInfo{})
	// k.UpdateTickWithNewLiquidity(ctx, poolId, lowerTick, liquidityIn)
	// k.setTickInfo(ctx, poolId, upperTick, TickInfo{})
	// k.UpdateTickWithNewLiquidity(ctx, poolId, upperTick, liquidityIn)
	// k.setPosition(ctx, poolId, owner, lowerTick, upperTick, Position{})
	// k.updatePositionWithLiquidity(ctx, poolId, owner, lowerTick, upperTick, liquidityIn)

	pool := k.getPoolbyId(ctx, poolId)

	currentSqrtPrice := pool.CurrentSqrtPrice
	sqrtRatioUpperTick, _ := k.tickToSqrtPrice(sdk.NewInt(upperTick))
	sqrtRatioLowerTick, _ := k.tickToSqrtPrice(sdk.NewInt(lowerTick))

	amtDenom0 = calcAmount0Delta(liquidityIn, currentSqrtPrice, sqrtRatioUpperTick).RoundInt()
	amtDenom1 = calcAmount1Delta(liquidityIn, currentSqrtPrice, sqrtRatioLowerTick).RoundInt()

	return amtDenom0, amtDenom1, nil
}

func (k Keeper) withdrawPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityAmount sdk.Int) (amtDenom0, amtDenom1 sdk.Int, err error) {
	panic("not implemented")
}
