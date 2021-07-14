package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
)

// autostakingStoreKey returns storekey by address
func autostakingStoreKey(address string) []byte {
	return combineKeys(types.KeyPrefixAutostaking, []byte(address))
}

// IterateAutoStaking iterate through autostaking configurations
func (k Keeper) IterateAutoStaking(ctx sdk.Context, fn func(index int64, autostaking types.AutoStaking) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixAutostaking)
	defer iterator.Close()

	i := int64(0)

	for ; iterator.Valid(); iterator.Next() {
		autostaking := types.AutoStaking{}
		err := proto.Unmarshal(iterator.Value(), &autostaking)
		if err != nil {
			panic(err)
		}
		stop := fn(i, autostaking)

		if stop {
			break
		}
		i++
	}
}

func (k Keeper) AllAutoStakings(ctx sdk.Context) []types.AutoStaking {
	autostakings := []types.AutoStaking{}
	k.IterateAutoStaking(ctx, func(index int64, autostaking types.AutoStaking) (stop bool) {
		autostakings = append(autostakings, autostaking)
		return false
	})
	return autostakings
}

// SetAutostaking set the autostaking configuration into the store
func (k Keeper) SetAutostaking(ctx sdk.Context, autostaking *types.AutoStaking) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(autostaking)
	if err != nil {
		return err
	}
	store.Set(autostakingStoreKey(autostaking.Address), bz)
	return nil
}

// GetAutostakingByAddress regurns autostaking config by address
func (k Keeper) GetAutostakingByAddress(ctx sdk.Context, address string) *types.AutoStaking {
	autostaking := types.AutoStaking{}
	store := ctx.KVStore(k.storeKey)
	autostakingKey := autostakingStoreKey(address)
	if !store.Has(autostakingKey) {
		return nil
	}
	bz := store.Get(autostakingKey)
	proto.Unmarshal(bz, &autostaking)
	return &autostaking
}
