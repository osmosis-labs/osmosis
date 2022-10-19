package concentrated_liquidity

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (k Keeper) Mint(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, tokenIn sdk.Int, lowerTick sdk.Int, upperTick sdk.Int) (numShares sdk.Int, err error) {
	// first check and validate arguments
	if lowerTick.GTE(types.MaxTick) || lowerTick.LT(types.MinTick) || upperTick.GT(types.MaxTick) {
		// TODO: come back to errors
		return sdk.Int{}, fmt.Errorf("validation fail")
	}

	if tokenIn.IsZero() {
		return sdk.Int{}, fmt.Errorf("token in amount is zero")
	}

	k.UpdateTickWithNewLiquidity(ctx, poolId, lowerTick, tokenIn)
	k.UpdateTickWithNewLiquidity(ctx, poolId, upperTick, tokenIn)

	// update tick with new liquidity
	k.updatePositionWithLiquidity(ctx, poolId, owner.String(), lowerTick, upperTick, tokenIn)

	return sdk.Int{}, nil
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
