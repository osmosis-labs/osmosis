package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func (k Keeper) GetTotalLiquidity(ctx sdk.Context) sdk.Coins {
	coins := sdk.Coins{}
	k.IterateDenomLiquidity(ctx, func(coin sdk.Coin) bool {
		coins = coins.Add(coin)
		return false
	})
	return coins
}

func (k Keeper) SetTotalLiquidity(ctx sdk.Context, coins sdk.Coins) {
	for _, coin := range coins {
		k.SetDenomLiquidity(ctx, coin.Denom, coin.Amount)
	}
}

func (k Keeper) SetDenomLiquidity(ctx sdk.Context, denom string, amount sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	bz, err := amount.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set(types.GetDenomPrefix(denom), bz)
}

func (k Keeper) GetDenomLiquidity(ctx sdk.Context, denom string) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetDenomPrefix(denom))
	if bz == nil {
		return sdk.NewInt(0)
	}

	var amount sdk.Int
	if err := amount.Unmarshal(bz); err != nil {
		panic(err)
	}
	return amount
}

func (k Keeper) IterateDenomLiquidity(ctx sdk.Context, cb func(sdk.Coin) bool) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyTotalLiquidity)

	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var amount sdk.Int
		if err := amount.Unmarshal(iterator.Value()); err != nil {
			panic(err)
		}

		if cb(sdk.NewCoin(string(iterator.Key()), amount)) {
			break
		}
	}
}

func (k Keeper) RecordTotalLiquidityIncrease(ctx sdk.Context, coins sdk.Coins) {
	for _, coin := range coins {
		amount := k.GetDenomLiquidity(ctx, coin.Denom)
		amount = amount.Add(coin.Amount)
		k.SetDenomLiquidity(ctx, coin.Denom, amount)
	}
}

func (k Keeper) RecordTotalLiquidityDecrease(ctx sdk.Context, coins sdk.Coins) {
	for _, coin := range coins {
		amount := k.GetDenomLiquidity(ctx, coin.Denom)
		amount = amount.Sub(coin.Amount)
		k.SetDenomLiquidity(ctx, coin.Denom, amount)
	}
}
