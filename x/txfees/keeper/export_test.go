package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k Keeper) SwapNonNativeFeeToDenom(ctx sdk.Context, denomToSwapTo string, feeCollectorAddress sdk.AccAddress) {
	k.swapNonNativeFeeToDenom(ctx, denomToSwapTo, feeCollectorAddress)
}

func (k Keeper) ClearTakerFeeShareAccumulators(ctx sdk.Context) {
	k.clearTakerFeeShareAccumulators(ctx)
}
