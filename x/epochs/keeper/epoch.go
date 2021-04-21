package keeper

import (
	"encoding/json"

	"github.com/c-osmosis/osmosis/x/epochs/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetEpochInfo returns epoch info by identifier
func (k Keeper) GetEpochInfo(ctx sdk.Context, identifier string) types.Epoch {
	epoch := types.Epoch{}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(append(types.KeyPrefixEpoch, []byte(identifier)...))
	if b == nil {
		return epoch
	}
	err := json.Unmarshal(b, &epoch)
	if err != nil {
		panic(err)
	}
	return epoch
}

// SetEpochInfo set epoch info
func (k Keeper) SetEpochInfo(ctx sdk.Context, epoch types.Epoch) {
	store := ctx.KVStore(k.storeKey)
	value, err := json.Marshal(epoch)
	if err != nil {
		panic(err)
	}
	store.Set(append(types.KeyPrefixEpoch, []byte(epoch.Identifier)...), value)
}

// IterateEpochInfo iterate through epochs
func (k Keeper) IterateEpochInfo(ctx sdk.Context, fn func(index int64, epochInfo types.Epoch) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixEpoch)
	defer iterator.Close()

	i := int64(0)

	for ; iterator.Valid(); iterator.Next() {
		epoch := types.Epoch{}
		err := json.Unmarshal(iterator.Value(), &epoch)
		if err != nil {
			panic(err)
		}
		stop := fn(i, epoch)

		if stop {
			break
		}
		i++
	}
}

func (k Keeper) AllEpochInfos(ctx sdk.Context) []types.Epoch {
	epochs := []types.Epoch{}
	k.IterateEpochInfo(ctx, func(index int64, epochInfo types.Epoch) (stop bool) {
		epochs = append(epochs, epochInfo)
		return false
	})
	return epochs
}
