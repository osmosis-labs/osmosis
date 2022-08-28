package twap

import (
	"encoding/binary"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/osmoutils"
	"github.com/osmosis-labs/osmosis/v11/x/twap/types"
)

// trackChangedPool places an entry into a transient store,
// to track that this pool changed this block.
// This tracking is for use in EndBlock, to create new TWAP records.
func (k Keeper) trackChangedPool(ctx sdk.Context, poolId uint64) {
	store := ctx.TransientStore(k.transientKey)
	poolIdBz := make([]byte, 8)
	binary.LittleEndian.PutUint64(poolIdBz, poolId)
	// just has to not be empty, for store to work / not register as a delete.
	sentinelExistsValue := []byte{1}
	store.Set(poolIdBz, sentinelExistsValue)
}

// getChangedPools returns all poolIDs that changed this block.
// This is to be guaranteed by trackChangedPool being called on every
// price-affecting pool action.
func (k Keeper) getChangedPools(ctx sdk.Context) []uint64 {
	store := ctx.TransientStore(k.transientKey)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	alteredPoolIds := []uint64{}
	for ; iter.Valid(); iter.Next() {
		k := iter.Key()
		poolId := binary.LittleEndian.Uint64(k)
		alteredPoolIds = append(alteredPoolIds, poolId)
	}
	return alteredPoolIds
}

// storeHistoricalTWAP writes a twap to the store, in all needed indexing.
func (k Keeper) storeHistoricalTWAP(ctx sdk.Context, twap types.TwapRecord) {
	store := ctx.KVStore(k.storeKey)
	key1 := types.FormatHistoricalTimeIndexTWAPKey(twap.Time, twap.PoolId, twap.Asset0Denom, twap.Asset1Denom)
	key2 := types.FormatHistoricalPoolIndexTWAPKey(twap.PoolId, twap.Time, twap.Asset0Denom, twap.Asset1Denom)
	osmoutils.MustSet(store, key1, &twap)
	osmoutils.MustSet(store, key2, &twap)
}

func (k Keeper) pruneRecordsBeforeTime(ctx sdk.Context, lastTime time.Time) error {
	store := ctx.KVStore(k.storeKey)
	iter := store.Iterator([]byte(types.HistoricalTWAPTimeIndexPrefix), types.FormatHistoricalTimeIndexTWAPKey(lastTime, 0, "", ""))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		twapToRemove, err := types.ParseTwapFromBz(iter.Value())
		if err != nil {
			return err
		}
		k.deleteHistoricalRecord(ctx, twapToRemove)
	}
	return nil
}

func (k Keeper) deleteHistoricalRecord(ctx sdk.Context, twap types.TwapRecord) {
	store := ctx.KVStore(k.storeKey)
	key1 := types.FormatHistoricalTimeIndexTWAPKey(twap.Time, twap.PoolId, twap.Asset0Denom, twap.Asset1Denom)
	key2 := types.FormatHistoricalPoolIndexTWAPKey(twap.PoolId, twap.Time, twap.Asset0Denom, twap.Asset1Denom)
	store.Delete(key1)
	store.Delete(key2)
}

// getMostRecentRecordStoreRepresentation returns the most recent twap record in the store
// for the provided (pool, asset0, asset1) triplet.
// Its called store representation, because most recent record can refer to it being
// interpolated to the current block time, or after events in this block.
// Neither of which apply to the record this returns.
func (k Keeper) getMostRecentRecordStoreRepresentation(ctx sdk.Context, poolId uint64, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatMostRecentTWAPKey(poolId, asset0Denom, asset1Denom)
	bz := store.Get(key)
	return types.ParseTwapFromBz(bz)
}

// getAllMostRecentRecordsForPool returns all most recent twap records
// (in state representation) for the provided pool id.
func (k Keeper) getAllMostRecentRecordsForPool(ctx sdk.Context, poolId uint64) ([]types.TwapRecord, error) {
	store := ctx.KVStore(k.storeKey)
	return types.GetAllMostRecentTwapsForPool(store, poolId)
}

// getAllHistoricalTimeIndexedTWAPs returns all historical TWAPs indexed by time.
func (k Keeper) getAllHistoricalTimeIndexedTWAPs(ctx sdk.Context) ([]types.TwapRecord, error) {
	return osmoutils.GatherValuesFromStore(ctx.KVStore(k.storeKey), []byte(types.HistoricalTWAPTimeIndexPrefix), []byte(types.HistoricalTWAPTimeIndexPrefixEnd), types.ParseTwapFromBz)
}

// getAllHistoricalPoolIndexedTWAPs returns all historical TWAPs indexed by pool id.
func (k Keeper) getAllHistoricalPoolIndexedTWAPs(ctx sdk.Context) ([]types.TwapRecord, error) {
	return osmoutils.GatherValuesFromStore(ctx.KVStore(k.storeKey), []byte(types.HistoricalTWAPPoolIndexPrefix), []byte(types.HistoricalTWAPPoolIndexPrefixEnd), types.ParseTwapFromBz)
}

// storeNewRecord stores a record, in both the most recent record store and historical stores.
func (k Keeper) storeNewRecord(ctx sdk.Context, twap types.TwapRecord) {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatMostRecentTWAPKey(twap.PoolId, twap.Asset0Denom, twap.Asset1Denom)
	osmoutils.MustSet(store, key, &twap)
	k.storeHistoricalTWAP(ctx, twap)
}

// getRecordAtOrBeforeTime on a given input (id, t, asset0, asset1)
// returns the TWAP record from state for (id, t', asset0, asset1),
// where t' is such that:
// * t' <= t
// * there exists no `t'' <= t` in state, where `t' < t''`
//
// This returns an error if:
// * there is no historical record in state at or before t
//   - Occurs if t is older than pruning period, or pool creation date.
// * there is no record for the asset pair (asset0, asset1) in particular
//   - e.g. asset not in pool, or provided in wrong order.
func (k Keeper) getRecordAtOrBeforeTime(ctx sdk.Context, poolId uint64, t time.Time, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	store := ctx.KVStore(k.storeKey)
	// We make an iteration from time=t + 1ns, to time=0 for this pool.
	// Note that we cannot get any time entries from t + 1ns, as the key would be `prefix|t+1ns`
	// and the end for a reverse iterator is exclusive. Thus the largest key that can be returned
	// begins with a prefix of `prefix|t`
	startKey := types.FormatHistoricalPoolIndexTimePrefix(poolId, time.Unix(0, 0))
	endKey := types.FormatHistoricalPoolIndexTimePrefix(poolId, t.Add(time.Nanosecond))
	lastParsedTime := time.Time{}
	stopFn := func(key []byte) bool {
		// halt iteration if we can't parse the time, or we've successfully parsed
		// a time, and have iterated beyond records for that time.
		parsedTime, err := types.ParseTimeFromHistoricalPoolIndexKey(key)
		if err != nil {
			return true
		}
		if lastParsedTime.After(parsedTime) {
			return true
		}
		lastParsedTime = parsedTime
		return false
	}

	reverseIterate := true
	twaps, err := osmoutils.GetIterValuesWithStop(store, startKey, endKey, reverseIterate, stopFn, types.ParseTwapFromBz)
	if err != nil {
		return types.TwapRecord{}, err
	}
	if len(twaps) == 0 {
		return types.TwapRecord{}, fmt.Errorf("looking for a time thats too old, not in the historical index. "+
			" Try storing the accumulator value. (requested time %s)", t)
	}

	for _, twap := range twaps {
		if twap.Asset0Denom == asset0Denom && twap.Asset1Denom == asset1Denom {
			return twap, nil
		}
	}
	return types.TwapRecord{}, fmt.Errorf("TWAP not found, but there are other twaps available for this time."+
		" Were provided asset0denom and asset1denom (%s, %s) correct, and in order (asset0 > asset1)?", asset0Denom, asset1Denom)
}
