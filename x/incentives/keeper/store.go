package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v3/x/incentives/types"
)

// getLastGaugeID returns ID used last time
func (k Keeper) getLastGaugeID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyLastGaugeID)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// setLastGaugeID save ID used by last gauge
func (k Keeper) setLastGaugeID(ctx sdk.Context, ID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyLastGaugeID, sdk.Uint64ToBigEndian(ID))
}

// gaugeStoreKey returns action store key from ID
func gaugeStoreKey(ID uint64) []byte {
	return combineKeys(types.KeyPrefixPeriodGauge, sdk.Uint64ToBigEndian(ID))
}

func gaugeDenomStoreKey(denom string) []byte {
	return combineKeys(types.KeyPrefixGaugesByDenom, []byte(denom))
}

// getGaugeRefs get gauge IDs specified on the provided key
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

// addGaugeRefByKey append gauge ID into an array associated to provided key
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

// deleteGaugeRefByKey removes gauge ID from an array associated to provided key
//nolint:ineffassign
func (k Keeper) deleteGaugeRefByKey(ctx sdk.Context, key []byte, gaugeID uint64) error {
	index := -1
	store := ctx.KVStore(k.storeKey)
	gaugeIDs := k.getGaugeRefs(ctx, key)
	gaugeIDs, index = removeValue(gaugeIDs, gaugeID)
	if index < 0 {
		return fmt.Errorf("specific gauge with ID %d not found", gaugeID)
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

// getAllGaugeIDsByDenom returns all active gauge-IDs associated with lockups of denomination `denom`
func (k Keeper) getAllGaugeIDsByDenom(ctx sdk.Context, denom string) []uint64 {
	return k.getGaugeRefs(ctx, gaugeDenomStoreKey(denom))
}

// deleteGaugeIDForDenom deletes ID from the list of gauge ID's associated with denomination `denom`
func (k Keeper) deleteGaugeIDForDenom(ctx sdk.Context, ID uint64, denom string) error {
	return k.deleteGaugeRefByKey(ctx, gaugeDenomStoreKey(denom), ID)
}

// addGaugeIDForDenom adds ID to the list of gauge ID's associated with denomination `denom`
func (k Keeper) addGaugeIDForDenom(ctx sdk.Context, ID uint64, denom string) error {
	return k.addGaugeRefByKey(ctx, gaugeDenomStoreKey(denom), ID)
}
