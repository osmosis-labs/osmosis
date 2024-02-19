package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k Keeper) SwapNonNativeFeeToDenom(ctx sdk.Context, denomToSwapTo string, feeCollectorAddress sdk.AccAddress, coinsToSwap sdk.Coins) {
	k.swapNonNativeFeeToDenom(ctx, denomToSwapTo, feeCollectorAddress, coinsToSwap)
}
