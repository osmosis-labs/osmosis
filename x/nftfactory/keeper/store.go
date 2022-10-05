package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// HasDenomID returns whether the specified denom ID exists
func (k Keeper) HasDenomID(ctx sdk.Context, id string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.KeyDenomID(id))
}

// SetDenom is responsible for saving the definition of denom
func (k Keeper) SetDenom(ctx sdk.Context, denom types.Denom) error {
	if k.HasDenomID(ctx, denom.Id) {
		return sdkerrors.Wrapf(types.ErrInvalidDenom, "denomID %s has already exists", denom.Id)
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&denom)

	store.Set(types.KeyDenomID(denom.Id), bz)
	store.Set(types.KeyDenomName(denom.denomName), []byte(denom.Id))
	return nil
}
