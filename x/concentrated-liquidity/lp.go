package concentrated_liquidity

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (k Keeper) Mint(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, liquidityIn sdk.Int, lowerTick, upperTick int64) (amtDenom0, amtDenom1 sdk.Int, err error) {
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

	// k.UpdateTickWithNewLiquidity(ctx, poolId, lowerTick, liquidityIn)
	// k.UpdateTickWithNewLiquidity(ctx, poolId, upperTick, liquidityIn)

	// k.updatePositionWithLiquidity(ctx, poolId, owner, lowerTick, upperTick, liquidityIn)

	pool := k.getPoolbyId(ctx, poolId)

	currentSqrtPrice := sdk.NewDecWithPrec(int64(pool.CurrentSqrtPrice.Uint64()), 6)
	fmt.Printf("%v CURRSQRTPRICE \n", currentSqrtPrice)
	sqrtRatioUpperTick, _ := k.tickToPrice(sdk.NewInt(upperTick)).ApproxSqrt()
	fmt.Printf("%v UPPER \n", sqrtRatioUpperTick)
	sqrtRatioLowerTick, _ := k.tickToPrice(sdk.NewInt(lowerTick)).ApproxSqrt()
	fmt.Printf("%v LOWER \n", sqrtRatioLowerTick)

	amtDenom0 = calcAmount0Delta(liquidityIn.ToDec(), currentSqrtPrice, sqrtRatioUpperTick).RoundInt()
	amtDenom1 = calcAmount1Delta(liquidityIn.ToDec(), currentSqrtPrice, sqrtRatioLowerTick).RoundInt()

	return amtDenom0, amtDenom1, nil
}

func (k Keeper) JoinPoolNoSwap(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, err error) {
	return sdk.Int{}, nil
}

func (k Keeper) CalcJoinPoolShares(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, newLiquidity sdk.Coins, err error) {
	return sdk.Int{}, sdk.Coins{}, nil
}

func (k Keeper) CalcJoinPoolNoSwapShares(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, newLiquidity sdk.Coins, err error) {
	return sdk.Int{}, sdk.Coins{}, nil
}

func (k Keeper) ExitPool(ctx sdk.Context, numShares sdk.Int, exitFee sdk.Dec) (exitedCoins sdk.Coins, err error) {
	return sdk.Coins{}, nil
}

func (k Keeper) CalcExitPoolCoinsFromShares(ctx sdk.Context, numShares sdk.Int, exitFee sdk.Dec) (exitedCoins sdk.Coins, err error) {
	return sdk.Coins{}, nil
}
