package concentrated_liquidity

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func NewConcentratedPool(poolId uint64, firstDenom string, secondDenom string) (types.Pool, error) {
	denom0 := firstDenom
	denom1 := secondDenom

	// we store token in lexiographical order
	if denom0 < denom1 {
		denom0, denom1 = secondDenom, firstDenom
	}
	pool := types.Pool{
		// TODO: implement pool address method
		// Address: types.NewPoolAddress(poolId).String(),
		Id: poolId,
	}
	return pool, nil
}

func (k Keeper) JoinPool(ctx sdk.Context, tokenIn sdk.Coin, lowerTick sdk.Int, upperTick sdk.Int) (numShares sdk.Int, err error) {
	// first check and validate arguments
	if lowerTick.GTE(types.MaxTick) || lowerTick.LT(types.MinTick) || upperTick.GT(types.MaxTick) {
		// TODO: come back to errors
		return sdk.Int{}, fmt.Errorf("validation fail")
	}

	if tokenIn.Amount.IsZero() {
		return sdk.Int{}, fmt.Errorf("token in amount is zero")
	}

	// update tick with new liquidity
	// k.

	return sdk.Int{}, nil
}

func (k Keeper) SwapOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error) {
	return sdk.Coin{}, nil
}
func (k Keeper) CalcOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error) {
	return sdk.Coin{}, nil
}
func (k Keeper) SwapInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	return sdk.Coin{}, nil
}
func (k Keeper) CalcInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	return sdk.Coin{}, nil
}
func (k Keeper) SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error) {
	return sdk.Dec{}, nil
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
