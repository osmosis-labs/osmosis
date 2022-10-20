package concentrated_liquidity

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (k Keeper) Mint(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, liquidityIn sdk.Int, lowerTick sdk.Int, upperTick sdk.Int) (amtDenom0, amtDenom1 sdk.Int, err error) {
	// ensure that lower tick is always smaller than upper tick
	if lowerTick.GTE(types.MaxTick) || lowerTick.LT(types.MinTick) || upperTick.GT(types.MaxTick) {
		return sdk.Int{}, sdk.Int{}, fmt.Errorf("validation fail")
	}

	if liquidityIn.IsZero() {
		return sdk.Int{}, sdk.Int{}, fmt.Errorf("token in amount is zero")
	}

	// k.UpdateTickWithNewLiquidity(ctx, poolId, lowerTick, liquidityIn)
	// k.UpdateTickWithNewLiquidity(ctx, poolId, upperTick, liquidityIn)

	// k.updatePositionWithLiquidity(ctx, poolId, owner.String(), lowerTick, upperTick, liquidityIn)

	pool := k.getPoolbyId(ctx, poolId)

	currentSqrtPrice := pool.CurrentSqrtPrice
	sqrtRatioUpperTick, err := k.getSqrtRatioAtTick(upperTick)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}
	sqrtRatioLowerTick, err := k.getSqrtRatioAtTick(lowerTick)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	amtDenom0 = calcAmount0Delta(currentSqrtPrice.ToDec(), sqrtRatioUpperTick, liquidityIn.ToDec()).RoundInt()
	amtDenom1 = calcAmount1Delta(currentSqrtPrice.ToDec(), sqrtRatioLowerTick, liquidityIn.ToDec()).RoundInt()

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
