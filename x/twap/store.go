package twap

import (
	"encoding/binary"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	storetypes "cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/twap/types"
)

// NumRecordsToPrunePerBlock is the number of twap records indexed by pool ID to prune per block.
// One record indexed by pool ID is deleted per incentive record.
// Therefore, setting this to 200 means 200 complete incentive records are deleted per block.
// The choice is somewhat arbitrary
// However, th intuition is that the number should be low enough to not make blocks take longer but
// not too small where it would take all the way to the next epoch.
var NumRecordsToPrunePerBlock uint16 = 200

// NumDeprecatedRecordsToPrunePerBlock is the number of twap records indexed by time to prune per block.
// This is the same as NumRecordsToPrunePerBlock, but is used for the deprecated historical twap records.
// This is to be used in the upgrade handler, to clear out the now-obsolete historical twap records
// that were indexed by time. It is expected that these records will be pruned shortly after the upgrade.
// After all these records are pruned, this logic can be removed for a future upgrade.
var NumDeprecatedRecordsToPrunePerBlock uint16 = 200

type timeTooOldError struct {
	Time time.Time
}

func (e timeTooOldError) Error() string {
	return fmt.Sprintf("looking for a time that's too old, not in the historical index. "+
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

// storeHistoricalTWAP writes a twap to the store, indexed by pool id.
func (k Keeper) StoreHistoricalTWAP(ctx sdk.Context, twap types.TwapRecord) {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatHistoricalPoolIndexTWAPKey(twap.PoolId, twap.Asset0Denom, twap.Asset1Denom, twap.Time)
	osmoutils.MustSet(store, key, &twap)
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
//
// If we reach the per block pruning limit, we store the last key seen in the pruning state.
// This is so that we can continue pruning from where we left off in the next block.
// If we have pruned all records, we set the pruning state to not pruning.
func (k Keeper) pruneRecordsBeforeTimeButNewest(ctx sdk.Context, state types.PruningState) error {
	store := ctx.KVStore(k.storeKey)

	var numPruned uint16
	var lastPoolIdCompleted uint64

	for poolId := state.LastSeenPoolId; poolId > 0; poolId-- {
		denoms, err := k.poolmanagerKeeper.RouteGetPoolDenoms(ctx, poolId)
		if err != nil {
			return err
		}

		// Notice, if we hit the prune limit in the middle of a pool, we will re-iterate over the completed pruned pool records.
		// This is acceptable overhead for the simplification this provides.
		denomPairs := types.GetAllUniqueDenomPairs(denoms)
		for _, denomPair := range denomPairs {
			// Reverse iterator guarantees that we iterate through the newest per pool first.
			// Due to how it is indexed, we will only iterate times starting from
			// lastKeptTime exclusively down to the oldest record.
			iter := store.ReverseIterator(
				types.FormatHistoricalPoolIndexDenomPairTWAPKey(poolId, denomPair.Denom0, denomPair.Denom1),
				types.FormatHistoricalPoolIndexTWAPKey(poolId, denomPair.Denom0, denomPair.Denom1, state.LastKeptTime))
			defer iter.Close()

			firstIteration := true
			for ; iter.Valid(); iter.Next() {
				if !firstIteration {
					// We have stored the newest record, so we can prune the rest.
					timeIndexKey := iter.Key()
					store.Delete(timeIndexKey)
					numPruned += 1

					if numPruned >= NumRecordsToPrunePerBlock {
						// We have hit the limit in the middle of a pool.
						// We store this pool as the last seen pool in the pruning state.
						// We accept re-iterating over denomPairs as acceptable overhead.
						state.LastSeenPoolId = poolId
						k.SetPruningState(ctx, state)
						return nil
					}
				} else {
					// If this is the first iteration after we have gotten through the records after lastKeptTime, we
					// still keep the record in order to allow interpolation (see function description for more details).
					firstIteration = false
				}
			}
		}
		lastPoolIdCompleted = poolId
	}

	if lastPoolIdCompleted == 1 {
		// We have pruned all records.
		state.IsPruning = false
		k.SetPruningState(ctx, state)
	}
	return nil
}

func (k Keeper) DeleteHistoricalRecord(ctx sdk.Context, twap types.TwapRecord) {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatHistoricalPoolIndexTWAPKey(twap.PoolId, twap.Asset0Denom, twap.Asset1Denom, twap.Time)
	store.Delete(key)
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

// GetAllMostRecentRecordsForPool returns all most recent twap records
// (in state representation) for the provided pool id.
func (k Keeper) GetAllMostRecentRecordsForPool(ctx sdk.Context, poolId uint64) ([]types.TwapRecord, error) {
	store := ctx.KVStore(k.storeKey)
	return types.GetAllMostRecentTwapsForPool(store, poolId)
}

// GetAllMostRecentRecordsForPool returns all most recent twap records
// (in state representation) for the provided pool id.
func (k Keeper) GetAllMostRecentRecordsForPoolWithDenoms(ctx sdk.Context, poolId uint64, denoms []string) ([]types.TwapRecord, error) {
	store := ctx.KVStore(k.storeKey)
	// if length != 2, use iterator
	if len(denoms) != 2 {
		return types.GetAllMostRecentTwapsForPool(store, poolId)
	}
	// else, directly fetch the key.
	asset0Denom, asset1Denom, err := types.LexicographicalOrderDenoms(denoms[0], denoms[1])
	if err != nil {
		return []types.TwapRecord{}, err
	}
	record, err := types.GetMostRecentTwapForPool(store, poolId, asset0Denom, asset1Denom)
	return []types.TwapRecord{record}, err
}

// getAllHistoricalPoolIndexedTWAPs returns all historical TWAPs indexed by pool id.
func (k Keeper) getAllHistoricalPoolIndexedTWAPs(ctx sdk.Context) ([]types.TwapRecord, error) {
	return osmoutils.GatherValuesFromStorePrefix(ctx.KVStore(k.storeKey), []byte(types.HistoricalTWAPPoolIndexPrefix), types.ParseTwapFromBz)
}

// GetAllHistoricalPoolIndexedTWAPsForPoolId returns HistoricalTwapRecord for a pool give poolId.
func (k Keeper) GetAllHistoricalPoolIndexedTWAPsForPoolId(ctx sdk.Context, poolId uint64) ([]types.TwapRecord, error) {
	return osmoutils.GatherValuesFromStorePrefix(ctx.KVStore(k.storeKey), types.FormatKeyPoolTwapRecords(poolId), types.ParseTwapFromBz)
}

// StoreNewRecord stores a record, in both the most recent record store and historical stores.
func (k Keeper) StoreNewRecord(ctx sdk.Context, twap types.TwapRecord) {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatMostRecentTWAPKey(twap.PoolId, twap.Asset0Denom, twap.Asset1Denom)
	osmoutils.MustSet(store, key, &twap)
	k.StoreHistoricalTWAP(ctx, twap)
}

// DeleteMostRecentRecord deletes a given record in most recent record store.
// Note that if there are entries in historical indexes for this record, they are not deleted by this method.
func (k Keeper) DeleteMostRecentRecord(ctx sdk.Context, twap types.TwapRecord) {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatMostRecentTWAPKey(twap.PoolId, twap.Asset0Denom, twap.Asset1Denom)
	store.Delete(key)
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

// DeleteHistoricalTimeIndexedTWAPs deletes every historical twap record indexed by time (now deprecated) up till the limit.
// This is to be used in the upgrade handler, to clear out the now-obsolete historical twap records
// that were indexed by time.
func (k Keeper) DeleteHistoricalTimeIndexedTWAPs(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, []byte("historical_time_index"))
	defer iter.Close()

	iterationCounter := uint16(0)
	for iter.Valid() {
		store.Delete(iter.Key())
		iterationCounter++
		if iterationCounter >= NumDeprecatedRecordsToPrunePerBlock {
			ctx.Logger().Info("Deleted deprecated historical time indexed twaps", "count", iterationCounter)
			return
		}
		iter.Next()
	}

	ctx.Logger().Info("Deleted deprecated historical time indexed twaps", "count", iterationCounter)

	if iterationCounter == 0 {
		// We have pruned all records, so we can delete the pruning key.
		ctx.Logger().Info("All deprecated historical time indexed twaps have been deleted")
		store.Delete(types.DeprecatedHistoricalTWAPsIsPruningKey)
	}
}

// DeleteDeprecatedHistoricalTWAPsIsPruning the state entry that determines if we are still
// executing pruning logic in the end blocker.
// TODO: Remove this in v26
func (k Keeper) DeleteDeprecatedHistoricalTWAPsIsPruning(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.DeprecatedHistoricalTWAPsIsPruningKey)
}
