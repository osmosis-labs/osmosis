package swaprouter

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k Keeper) GetNextPoolIdAndIncrement(ctx sdk.Context) uint64 {
	return k.getNextPoolIdAndIncrement(ctx)
}
