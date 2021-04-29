package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) getTotalLockedDenom(ctx sdk.Context, denom string) (res sdk.Int) {
	key := getTotalLockedDenomKey(denom)
	store := ctx.KVStore(k.storeKey)
	if !store.Has(key) {
		return sdk.ZeroInt()
	}
	bz := store.Get(getTotalLockedDenomKey(denom))
	err := (&res).UnmarshalJSON(bz)
	if err != nil {
		panic(err)
	}
	return
}

func (k Keeper) setTotalLockedDenom(ctx sdk.Context, denom string, amount sdk.Int) {
	bz, err := amount.MarshalJSON()
	if err != nil {
		panic(err)
	}
	ctx.KVStore(k.storeKey).Set(getTotalLockedDenomKey(denom), bz)
}

func (k Keeper) IncreaseTotalLocked(ctx sdk.Context, amount sdk.Coins) {
	for _, coin := range amount {
		x := k.getTotalLockedDenom(ctx, coin.Denom)
		k.setTotalLockedDenom(ctx, coin.Denom, x.Add(coin.Amount))
	}
}

func (k Keeper) DecreaseTotalLocked(ctx sdk.Context, amount sdk.Coins) {
	for _, coin := range amount {
		x := k.getTotalLockedDenom(ctx, coin.Denom)
		k.setTotalLockedDenom(ctx, coin.Denom, x.Sub(coin.Amount))
		// XXX: invariant no negative
	}
}
