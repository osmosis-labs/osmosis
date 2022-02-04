package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func (k Keeper) SetLegacyTotalLiquidity(ctx sdk.Context, coins sdk.Coins) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyTotalLiquidity, []byte(coins.String()))
}

func (k Keeper) GetLegacyTotalLiquidity(ctx sdk.Context) sdk.Coins {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(types.KeyTotalLiquidity) {
		return sdk.Coins{}
	}

	bz := store.Get(types.KeyTotalLiquidity)
	coins, err := sdk.ParseCoinsNormalized(string(bz))
	if err != nil {
		panic("invalid total liquidity value set")
	}

	return coins
}

func (k Keeper) DeleteLegacyTotalLiquidity(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyTotalLiquidity)
}
