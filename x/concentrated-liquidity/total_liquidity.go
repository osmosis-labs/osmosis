package concentrated_liquidity

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

func (k Keeper) GetTotalLiquidity(ctx sdk.Context) (sdk.Coins, error) {
	coins := sdk.Coins{}
	k.IterateDenomLiquidity(ctx, func(coin sdk.Coin) bool {
		coins = coins.Add(coin)
		return false
	})
	return coins, nil
}

func (k Keeper) setTotalLiquidity(ctx sdk.Context, coins sdk.Coins) {
	for _, coin := range coins {
		k.setDenomLiquidity(ctx, coin.Denom, coin.Amount)
	}
}

func (k Keeper) setDenomLiquidity(ctx sdk.Context, denom string, amount osmomath.Int) {
	store := ctx.KVStore(k.storeKey)
	bz, err := amount.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set(types.GetDenomPrefix(denom), bz)
}

func (k Keeper) GetDenomLiquidity(ctx sdk.Context, denom string) osmomath.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetDenomPrefix(denom))
	if bz == nil {
		return osmomath.NewInt(0)
	}

	var amount osmomath.Int
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
		var amount osmomath.Int
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
		k.setDenomLiquidity(ctx, coin.Denom, amount)
	}
}

func (k Keeper) RecordTotalLiquidityDecrease(ctx sdk.Context, coins sdk.Coins) {
	for _, coin := range coins {
		amount := k.GetDenomLiquidity(ctx, coin.Denom)
		amount = amount.Sub(coin.Amount)
		k.setDenomLiquidity(ctx, coin.Denom, amount)
	}
}
