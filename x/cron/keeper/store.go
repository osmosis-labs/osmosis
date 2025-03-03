package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	protobuftypes "github.com/cosmos/gogoproto/types"
	"github.com/osmosis-labs/osmosis/v29/x/cron/types"
)

func (k Keeper) SetCronID(ctx sdk.Context, id uint64) {
	store := k.Store(ctx)
	key := types.LastCronIDKey
	value := k.cdc.MustMarshal(&protobuftypes.UInt64Value{Value: id})
	store.Set(key, value)
}

func (k Keeper) GetCronID(ctx sdk.Context) uint64 {
	store := k.Store(ctx)
	key := types.LastCronIDKey
	value := store.Get(key)
	if value == nil {
		return 0
	}
	var id protobuftypes.UInt64Value
	k.cdc.MustUnmarshal(value, &id)
	return id.GetValue()
}

func (k Keeper) SetCronJob(ctx sdk.Context, msg types.CronJob) {
	store := k.Store(ctx)
	key := types.CronKey(msg.Id)
	value := k.cdc.MustMarshal(&msg)
	store.Set(key, value)
}

func (k Keeper) GetCronJob(ctx sdk.Context, cronID uint64) (cron types.CronJob, found bool) {
	store := k.Store(ctx)
	key := types.CronKey(cronID)
	value := store.Get(key)
	if value == nil {
		return cron, false
	}
	k.cdc.MustUnmarshal(value, &cron)
	return cron, true
}

func (k Keeper) GetCronJobs(ctx sdk.Context) (crons []types.CronJob) {
	store := k.Store(ctx)
	iter := storetypes.KVStorePrefixIterator(store, types.CronJobKeyPrefix)
	defer func(iter storetypes.Iterator) {
		err := iter.Close()
		if err != nil {
			return
		}
	}(iter)
	for ; iter.Valid(); iter.Next() {
		var cron types.CronJob
		k.cdc.MustUnmarshal(iter.Value(), &cron)
		crons = append(crons, cron)
	}
	return crons
}
