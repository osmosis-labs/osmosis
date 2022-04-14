package keeper

import (
	"github.com/osmosis-labs/osmosis/v7/x/claim/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams get params.
func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.ParamsKey))
	params := types.Params{}
	err := k.cdc.UnmarshalJSON(bz, &params)
	return params, err
}

// SetParams set params.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.MarshalJSON(&params)
	if err != nil {
		return err
	}
	store.Set([]byte(types.ParamsKey), bz)
	return nil
}
