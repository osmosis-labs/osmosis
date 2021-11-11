package keeper

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/tokenfactory/types"
)

func (k Keeper) GetBurnersStore(ctx sdk.Context, denom string) sdk.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, []byte(strings.Join(
		[]string{denom, types.DenomBurnersStorePrefix}, "|")))
}

// AddBurners adds burners for a specific denom
func (k Keeper) AddBurners(ctx sdk.Context, denom string, burners []string) {
	store := k.GetBurnersStore(ctx, denom)
	for _, burner := range burners {
		// Turn the value into a bool?
		store.Set([]byte(burner), []byte(burner))
	}
}

// RemoveBurners removes burners for a specific denom
func (k Keeper) RemoveBurners(ctx sdk.Context, denom string, burners []string) {
	store := k.GetBurnersStore(ctx, denom)
	for _, burner := range burners {
		store.Delete([]byte(burner))
	}
}

// IsBurner returns if a specific address is a burner for a specific denom
func (k Keeper) IsBurner(ctx sdk.Context, denom string, address string) bool {
	return k.GetBurnersStore(ctx, denom).Has([]byte(address))
}
