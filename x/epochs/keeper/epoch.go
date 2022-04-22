package keeper

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/v7/x/epochs/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetEpochInfo returns epoch info by identifier.
func (k Keeper) GetEpochInfo(ctx sdk.Context, identifier string) types.EpochInfo {
	epoch := types.EpochInfo{}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(append(types.KeyPrefixEpoch, []byte(identifier)...))
	if b == nil {
		return epoch
	}
	err := proto.Unmarshal(b, &epoch)
	if err != nil {
		panic(err)
	}
	return epoch
}

// SetEpochInfo set epoch info.
func (k Keeper) SetEpochInfo(ctx sdk.Context, epoch types.EpochInfo) {
	store := ctx.KVStore(k.storeKey)
	value, err := proto.Marshal(&epoch)
	if err != nil {
		panic(err)
	}
	store.Set(append(types.KeyPrefixEpoch, []byte(epoch.Identifier)...), value)
}

// DeleteEpochInfo delete epoch info.
func (k Keeper) DeleteEpochInfo(ctx sdk.Context, identifier string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(append(types.KeyPrefixEpoch, []byte(identifier)...))
}

// IterateEpochInfo iterate through epochs.
func (k Keeper) IterateEpochInfo(ctx sdk.Context, fn func(index int64, epochInfo types.EpochInfo) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixEpoch)
	defer iterator.Close()

	i := int64(0)

	for ; iterator.Valid(); iterator.Next() {
		epoch := types.EpochInfo{}
		err := proto.Unmarshal(iterator.Value(), &epoch)
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

func (k Keeper) AllEpochInfos(ctx sdk.Context) []types.EpochInfo {
	epochs := []types.EpochInfo{}
	k.IterateEpochInfo(ctx, func(index int64, epochInfo types.EpochInfo) (stop bool) {
		epochs = append(epochs, epochInfo)
		return false
	})
	return epochs
}

// NumBlocksSinceEpochStart returns the number of blocks since the epoch started.
// if the epoch started on block N, then calling this during block N (after BeforeEpochStart)
// would return 0.
// Calling it any point in block N+1 (assuming the epoch doesn't increment) would return 1.
func (k Keeper) NumBlocksSinceEpochStart(ctx sdk.Context, identifier string) (int64, error) {
	epoch := k.GetEpochInfo(ctx, identifier)
	if (epoch == types.EpochInfo{}) {
		return 0, fmt.Errorf("epoch with identifier %s not found", identifier)
	}
	return ctx.BlockHeight() - epoch.CurrentEpochStartHeight, nil
}
