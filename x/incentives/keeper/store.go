package keeper

import (
	"encoding/json"
	"fmt"

	"github.com/osmosis-labs/osmosis/v11/x/incentives/types"

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
