package keeper

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/tokenfactory/types"
)

func (k Keeper) GetAdminsStore(ctx sdk.Context, denom string) sdk.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, []byte(strings.Join(
		[]string{denom, types.DenomAdminsStorePrefix}, "|")))
}

// AddAdmins adds admins for a specific denom
func (k Keeper) AddAdmins(ctx sdk.Context, denom string, admins []string) {

	store := k.GetAdminsStore(ctx, denom)

	for _, admin := range admins {
		// Turn the value into a bool?
		store.Set([]byte(admin), []byte(admin))
	}
}

// RemoveAdmins removes admins for a specific denom
func (k Keeper) RemoveAdmins(ctx sdk.Context, denom string, admins []string) {
	store := k.GetAdminsStore(ctx, denom)
	for _, admin := range admins {
		store.Delete([]byte(admin))
	}
}

// IsAdmin returns if a specific address is a admin for a specific denom
func (k Keeper) IsAdmin(ctx sdk.Context, denom string, address string) bool {
	return k.GetAdminsStore(ctx, denom).Has([]byte(address))
}
