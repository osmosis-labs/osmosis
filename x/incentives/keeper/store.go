package keeper

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetLastGaugeID returns the last used gauge ID.
func (k Keeper) GetLastGaugeID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyLastGaugeID)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetLastGaugeID sets the last used gauge ID to the provided ID.
func (k Keeper) SetLastGaugeID(ctx sdk.Context, ID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyLastGaugeID, sdk.Uint64ToBigEndian(ID))
}

// gaugeStoreKey returns the combined byte array (store key) of the provided gauge ID's key prefix and the ID itself.
func gaugeStoreKey(ID uint64) []byte {
	return combineKeys(types.KeyPrefixPeriodGauge, sdk.Uint64ToBigEndian(ID))
}

// gaugeDenomStoreKey returns the combined byte array (store key) of the provided gauge denom key prefix and the denom itself.
func gaugeDenomStoreKey(denom string) []byte {
	return combineKeys(types.KeyPrefixGaugesByDenom, []byte(denom))
}

// getGaugeRefs returns the gauge IDs specified by the provided key.
func (k Keeper) getGaugeRefs(ctx sdk.Context, key []byte) []uint64 {
	store := ctx.KVStore(k.storeKey)
	gaugeIDs := []uint64{}
	if store.Has(key) {
		bz := store.Get(key)
		err := json.Unmarshal(bz, &gaugeIDs)
		if err != nil {
			panic(err)
		}
	}
	return gaugeIDs
}

// addGaugeRefByKey appends the provided gauge ID into an array associated with the provided key.
func (k Keeper) addGaugeRefByKey(ctx sdk.Context, key []byte, gaugeID uint64) error {
	store := ctx.KVStore(k.storeKey)
	gaugeIDs := k.getGaugeRefs(ctx, key)
	if findIndex(gaugeIDs, gaugeID) > -1 {
		return fmt.Errorf("gauge with same ID exist: %d", gaugeID)
	}
	gaugeIDs = append(gaugeIDs, gaugeID)
	bz, err := json.Marshal(gaugeIDs)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// deleteGaugeRefByKey removes the provided gauge ID from an array associated with the provided key.
func (k Keeper) deleteGaugeRefByKey(ctx sdk.Context, key []byte, gaugeID uint64) error {
	store := ctx.KVStore(k.storeKey)
	gaugeIDs := k.getGaugeRefs(ctx, key)
	gaugeIDs, index := removeValue(gaugeIDs, gaugeID)
	if index < 0 {
		return fmt.Errorf("specific gauge with ID %d not found by reference %s", gaugeID, key)
	}
	if len(gaugeIDs) == 0 {
		store.Delete(key)
	} else {
		bz, err := json.Marshal(gaugeIDs)
		if err != nil {
			return err
		}
		store.Set(key, bz)
	}
	return nil
}

// getAllGaugeIDsByDenom returns all active gauge-IDs associated with lockups of the provided denom.
func (k Keeper) getAllGaugeIDsByDenom(ctx sdk.Context, denom string) []uint64 {
	return k.getGaugeRefs(ctx, gaugeDenomStoreKey(denom))
}

// deleteGaugeIDForDenom deletes the provided ID from the list of gauge ID's associated with the provided denom.
func (k Keeper) deleteGaugeIDForDenom(ctx sdk.Context, ID uint64, denom string) error {
	return k.deleteGaugeRefByKey(ctx, gaugeDenomStoreKey(denom), ID)
}

// addGaugeIDForDenom adds the provided ID to the list of gauge ID's associated with the provided denom.
func (k Keeper) addGaugeIDForDenom(ctx sdk.Context, ID uint64, denom string) error {
	return k.addGaugeRefByKey(ctx, gaugeDenomStoreKey(denom), ID)
}

// SetGroup sets groupGroup for a specific key.
// TODO: explore if we can store this better, this has GroupGaugeId in key and value
func (k Keeper) SetGroup(ctx sdk.Context, group types.Group) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.KeyGroupByGaugeID(group.GroupGaugeId), &group)
}

// GetAllGroupGauges gets all the groupGauges that is in state.
func (k Keeper) GetAllGroups(ctx sdk.Context) ([]types.Group, error) {
	return osmoutils.GatherValuesFromStorePrefix(ctx.KVStore(k.storeKey), types.KeyPrefixGroup, k.ParseGroupFromBz)
}

// GetAllGroupsGauges iterates through all groups, sequentially pulls the gauges from each, and returns just the gauges.
func (k Keeper) GetAllGroupsGauges(ctx sdk.Context) ([]types.Gauge, error) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixGroup)
	iter := prefixStore.Iterator(nil, nil)

	var gauges []types.Gauge
	for ; iter.Valid(); iter.Next() {
		group, err := k.ParseGroupFromBz(iter.Value())
		if err != nil {
			iter.Close()
			panic(fmt.Errorf("invalid group key (%s): %v", string(iter.Key()), err))
		}

		gauge, err := k.GetGaugeByID(ctx, group.GroupGaugeId)
		if err != nil {
			iter.Close()
			return nil, err
		}
		gauges = append(gauges, *gauge)
	}
	return gauges, nil
}

// GetAllGroupsWithGauge iterates through all groups, sequentially pulls the gauges from each, and returns the groups with their associated gauge.
func (k Keeper) GetAllGroupsWithGauge(ctx sdk.Context) ([]types.GroupsWithGauge, error) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixGroup)
	iter := prefixStore.Iterator(nil, nil)

	var groupsWithGauge []types.GroupsWithGauge
	for ; iter.Valid(); iter.Next() {
		group, err := k.ParseGroupFromBz(iter.Value())
		if err != nil {
			iter.Close()
			panic(fmt.Errorf("invalid group key (%s): %v", string(iter.Key()), err))
		}

		gauge, err := k.GetGaugeByID(ctx, group.GroupGaugeId)
		if err != nil {
			iter.Close()
			return nil, err
		}
		groupsWithGauge = append(groupsWithGauge, types.GroupsWithGauge{
			Group: group,
			Gauge: *gauge,
		})
	}
	return groupsWithGauge, nil
}

func (k Keeper) ParseGroupFromBz(bz []byte) (group types.Group, err error) {
	if len(bz) == 0 {
		return types.Group{}, errors.New("group gauge not found")
	}
	err = proto.Unmarshal(bz, &group)

	return group, err
}

// GetGroupByGaugeID gets group struct for a given gauge ID. Note that Group and group's associated gauge
// are 1:1 mapped. As a result, they share the same ID.
func (k Keeper) GetGroupByGaugeID(ctx sdk.Context, gaugeID uint64) (types.Group, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyGroupByGaugeID(gaugeID)
	bz := store.Get(key)
	if bz == nil {
		return types.Group{}, types.GroupNotFoundError{GroupGaugeId: gaugeID}
	}

	var group types.Group
	if err := proto.Unmarshal(bz, &group); err != nil {
		return types.Group{}, err
	}

	return group, nil
}
