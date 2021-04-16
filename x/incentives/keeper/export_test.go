package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) AddPotRefByKey(ctx sdk.Context, key []byte, potID uint64) error {
	return k.addPotRefByKey(ctx, key, potID)
}

func (k Keeper) DeletePotRefByKey(ctx sdk.Context, key []byte, potID uint64) {
	k.deletePotRefByKey(ctx, key, potID)
}

func (k Keeper) GetPotRefs(ctx sdk.Context, key []byte) []uint64 {
	return k.getPotRefs(ctx, key)
}
