package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/v12/x/nftfactory/types"
)

// HasDenomID returns whether the specified denom ID exists
func (k Keeper) HasDenomID(ctx sdk.Context, id string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.KeyDenomID(id))
}

// SetDenom is responsible for saving the definition of denom
func (k Keeper) SetDenom(ctx sdk.Context, denom types.Denom) error {
	if k.HasDenomID(ctx, denom.Id) {
		return fmt.Errorf("denomID %s has already exists", denom.Id)
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(&denom)
	if err != nil {
		return err
	}

	store.Set(types.KeyDenomID(denom.Id), bz)
	store.Set(types.KeyDenomName(denom.DenomName), []byte(denom.Id))
	return nil
}
