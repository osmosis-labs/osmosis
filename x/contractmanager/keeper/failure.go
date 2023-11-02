package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/x/contractmanager/types"
)

// AddContractFailure adds a specific failure to the store using address as the key
func (k Keeper) AddContractFailure(ctx sdk.Context, channelID, address string, ackID uint64, ackType string) {
	failure := types.Failure{
		ChannelId: channelID,
		Address:   address,
		AckId:     ackID,
		AckType:   ackType,
	}
	nextFailureID := k.GetNextFailureIDKey(ctx, failure.GetAddress())

	store := ctx.KVStore(k.storeKey)

	failure.Id = nextFailureID
	b := k.cdc.MustMarshal(&failure)
	store.Set(types.GetFailureKey(failure.GetAddress(), nextFailureID), b)
}

func (k Keeper) GetNextFailureIDKey(ctx sdk.Context, address string) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetFailureKeyPrefix(address))
	iterator := sdk.KVStoreReversePrefixIterator(store, []byte{})
	defer iterator.Close()

	if iterator.Valid() {
		var val types.Failure
		k.cdc.MustUnmarshal(iterator.Value(), &val)

		return val.Id + 1
	}

	return 0
}

// GetAllFailure returns all failures
func (k Keeper) GetAllFailures(ctx sdk.Context) (list []types.Failure) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ContractFailuresKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Failure
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
