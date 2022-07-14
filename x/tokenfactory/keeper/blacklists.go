package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetAuthorityMetadata returns the authority metadata for a specific denom
func (k Keeper) GetBlacklistStatus(ctx sdk.Context, denom string, address string) bool {
	return k.GetDenomBlacklistPrefixStore(ctx, denom).Has([]byte(address))
}

// setBlacklistStatus stores authority metadata for a specific denom
func (k Keeper) setBlacklistStatus(ctx sdk.Context, denom string, address string, status bool) {
	store := k.GetDenomBlacklistPrefixStore(ctx, denom)

	if status == true {
		store.Set([]byte(address), []byte{1})
	} else {
		store.Delete([]byte(address))
	}
}
