package keeper

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/tokenfactory/types"
)

func (k Keeper) GetMintersStore(ctx sdk.Context, denom string) sdk.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, []byte(strings.Join(
		[]string{denom, types.DenomMintersStorePrefix}, "|")))
}

// AddMinters adds minters for a specific denom
func (k Keeper) AddMinters(ctx sdk.Context, denom string, minters []string) {

	store := k.GetMintersStore(ctx, denom)

	for _, minter := range minters {
		// Turn the value into a bool?
		store.Set([]byte(minter), []byte(minter))
	}
}

// AddMinters removes minters for a specific denom
func (k Keeper) RemoveMinters(ctx sdk.Context, denom string, minters []string) {
	store := k.GetMintersStore(ctx, denom)
	for _, minter := range minters {
		store.Delete([]byte(minter))
	}
}

// IsMinter returns if a specific address is a minter for a specific denom
func (k Keeper) IsMinter(ctx sdk.Context, denom string, address string) bool {
	return k.GetMintersStore(ctx, denom).Has([]byte(address))
}
