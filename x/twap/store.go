package twap

import (
	"encoding/binary"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	"github.com/osmosis-labs/osmosis/v12/x/twap/types"
)

type timeTooOldError struct {
	Time time.Time
}

func (e timeTooOldError) Error() string {
	return fmt.Sprintf("looking for a time thats too old, not in the historical index. "+
		" Try storing the accumulator value. (requested time %s)", e.Time)
}

// just has to not be empty, for store to work / not register as a delete.
var sentinelExistsValue = []byte{1}

// trackChangedPool places an entry into a transient store,
// to track that this pool changed this block.
// This tracking is for use in EndBlock, to create new TWAP records.
func (k Keeper) trackChangedPool(ctx sdk.Context, poolId uint64) {
	store := ctx.TransientStore(k.transientKey)
	poolIdBz := make([]byte, 8)
	binary.LittleEndian.PutUint64(poolIdBz, poolId)

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
	key2 := types.FormatHistoricalPoolIndexTWAPKey(twap.PoolId, twap.Asset0Denom, twap.Asset1Denom, twap.Time)
	osmoutils.MustSet(store, key1, &twap)
	osmoutils.MustSet(store, key2, &twap)
}

// pruneRecordsBeforeTimeButNewest prunes all records for each pool before the given time but the newest
// record. The reason for preserving at least one record earlier than the keep period is
// to ensure that we have a record to interpolate from in case there is only one or no records
// within the keep period.
// For example:
// - Suppose pruning param -48 hour
// - Suppose there are three records at: -51 hour, -50 hour, and -1hour
// If we were to prune everything older than 48 hours,
// we would be left with only one record at -1 hour, and we wouldn't be able to
// get twaps from the [-48 hour, -1 hour] time range.
// So, in order to have correct behavior for the desired guarantee,
// we keep the newest record that is older than the pruning time.
// This is why we would keep the -50 hour and -1hour twaps despite a 48hr pruning period
func (k Keeper) pruneRecordsBeforeTimeButNewest(ctx sdk.Context, lastKeptTime time.Time) error {
	store := ctx.KVStore(k.storeKey)

	// Reverse iterator guarantees that we iterate through the newest per pool first.
	// Due to how it is indexed, we will only iterate times starting from
	// lastKeptTime exclusively down to the oldest record.
	iter := store.ReverseIterator(
		[]byte(types.HistoricalTWAPTimeIndexPrefix),
		types.FormatHistoricalTimeIndexTWAPKey(lastKeptTime, 0, "", ""))
	defer iter.Close()

	// We mark what (pool id, asset 0, asset 1) triplets we've seen.
	// We prune all records for a triplet that we haven't already seen.
	type uniqueTriplet struct {
		poolId uint64
		asset0 string
		asset1 string
	}
	seenPoolAssetTriplets := map[uniqueTriplet]struct{}{}

	for ; iter.Valid(); iter.Next() {
		twapToRemove, err := types.ParseTwapFromBz(iter.Value())
		if err != nil {
			return err
		}

		poolKey := uniqueTriplet{
			poolId: twapToRemove.PoolId,
			asset0: twapToRemove.Asset0Denom,
			asset1: twapToRemove.Asset1Denom,
		}
		_, hasSeenPoolRecord := seenPoolAssetTriplets[poolKey]
		if !hasSeenPoolRecord {
			seenPoolAssetTriplets[poolKey] = struct{}{}
			continue
		}

		k.deleteHistoricalRecord(ctx, twapToRemove)
	}
	return nil
}

func (k Keeper) deleteHistoricalRecord(ctx sdk.Context, twap types.TwapRecord) {
	store := ctx.KVStore(k.storeKey)
	key1 := types.FormatHistoricalTimeIndexTWAPKey(twap.Time, twap.PoolId, twap.Asset0Denom, twap.Asset1Denom)
	key2 := types.FormatHistoricalPoolIndexTWAPKey(twap.PoolId, twap.Asset0Denom, twap.Asset1Denom, twap.Time)
	store.Delete(key1)
	store.Delete(key2)
}

// getMostRecentRecordStoreRepresentation returns the most recent twap record in the store
// for the provided (pool, asset0, asset1) triplet.
// Its called store representation, because most recent record can refer to it being
// interpolated to the current block time, or after events in this block.
// Neither of which apply to the record this returns.
func (k Keeper) getMostRecentRecordStoreRepresentation(ctx sdk.Context, poolId uint64, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	asset0Denom, asset1Denom, err := types.LexicographicalOrderDenoms(asset0Denom, asset1Denom)
	if err != nil {
		return types.TwapRecord{}, err
	}
	store := ctx.KVStore(k.storeKey)
	key := types.FormatMostRecentTWAPKey(poolId, asset0Denom, asset1Denom)
	bz := store.Get(key)
	twap, err := types.ParseTwapFromBz(bz)
	if err != nil {
		err = fmt.Errorf("error in get most recent twap, likely that asset 0 or asset 1 were wrong: %s %s."+
			" Underlying error: %w", asset0Denom, asset1Denom, err)
	}
	return twap, err
}

// getAllMostRecentRecordsForPool returns all most recent twap records
// (in state representation) for the provided pool id.
func (k Keeper) getAllMostRecentRecordsForPool(ctx sdk.Context, poolId uint64) ([]types.TwapRecord, error) {
	store := ctx.KVStore(k.storeKey)
	return types.GetAllMostRecentTwapsForPool(store, poolId)
}

// getAllHistoricalTimeIndexedTWAPs returns all historical TWAPs indexed by time.
func (k Keeper) getAllHistoricalTimeIndexedTWAPs(ctx sdk.Context) ([]types.TwapRecord, error) {
	return osmoutils.GatherValuesFromStorePrefix(ctx.KVStore(k.storeKey), []byte(types.HistoricalTWAPTimeIndexPrefix), types.ParseTwapFromBz)
}

// getAllHistoricalPoolIndexedTWAPs returns all historical TWAPs indexed by pool id.
// nolint: unused
func (k Keeper) getAllHistoricalPoolIndexedTWAPs(ctx sdk.Context) ([]types.TwapRecord, error) {
	return osmoutils.GatherValuesFromStorePrefix(ctx.KVStore(k.storeKey), []byte(types.HistoricalTWAPPoolIndexPrefix), types.ParseTwapFromBz)
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
// * there exists no `t” <= t` in state, where `t' < t”`
//
// This returns an error if:
// * there is no historical record in state at or before t
//   - Occurs if t is older than pruning period, or pool creation date.
//
// * there is no record for the asset pair (asset0, asset1) in particular
//   - e.g. asset not in pool, or provided in wrong order.
func (k Keeper) getRecordAtOrBeforeTime(ctx sdk.Context, poolId uint64, t time.Time, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	asset0Denom, asset1Denom, err := types.LexicographicalOrderDenoms(asset0Denom, asset1Denom)
	if err != nil {
		return types.TwapRecord{}, err
	}
	store := ctx.KVStore(k.storeKey)
	// We make an iteration from time=t, to time=0 for this pool.
	startKey := types.FormatHistoricalPoolIndexTimePrefix(poolId, asset0Denom, asset1Denom)
	endKey := types.FormatHistoricalPoolIndexTimeSuffix(poolId, asset0Denom, asset1Denom, t)
	reverseIterate := true

	twap, err := osmoutils.GetFirstValueInRange(store, startKey, endKey, reverseIterate, types.ParseTwapFromBz)
	if err != nil {
		// diagnose why we have no results by seeing what happens for getMostRecentRecord for this pool id
		_, errDiagnose := k.getMostRecentRecord(ctx, poolId, asset0Denom, asset1Denom)
		if errDiagnose != nil {
			return types.TwapRecord{}, fmt.Errorf(
				"getTwapRecord: querying for assets %s %s that are not in pool id %d",
				asset0Denom, asset1Denom, poolId)
		} else {
			return types.TwapRecord{}, timeTooOldError{Time: t}
		}
	}
	if twap.Asset0Denom != asset0Denom || twap.Asset1Denom != asset1Denom || twap.PoolId != poolId {
		return types.TwapRecord{}, fmt.Errorf("internal error, got twap but its data is wrong")
	}

	return twap, nil
}
