package twap

import (
	"encoding/binary"
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/osmoutils"
	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

func (k twapkeeper) trackChangedPool(ctx sdk.Context, poolId uint64) {
	store := ctx.TransientStore(&k.transientKey)
	poolIdBz := make([]byte, 8)
	binary.LittleEndian.PutUint64(poolIdBz, poolId)
	// just has to not be empty, for store to work / not register as a delete.
	sentinelExistsValue := []byte{1}
	store.Set(poolIdBz, sentinelExistsValue)
}

func (k twapkeeper) hasPoolChangedThisBlock(ctx sdk.Context, poolId uint64) bool {
	store := ctx.TransientStore(&k.transientKey)
	poolIdBz := make([]byte, 8)
	binary.LittleEndian.PutUint64(poolIdBz, poolId)
	return store.Has(poolIdBz)
}

func (k twapkeeper) storeHistoricalTWAP(ctx sdk.Context, twap types.TwapRecord) {
	store := ctx.KVStore(k.storeKey)
	key1 := types.FormatHistoricalTimeIndexTWAPKey(twap.Time, twap.PoolId, twap.Asset0Denom, twap.Asset1Denom)
	key2 := types.FormatHistoricalPoolIndexTWAPKey(twap.PoolId, twap.Time, twap.Asset0Denom, twap.Asset1Denom)
	osmoutils.MustSet(store, key1, &twap)
	osmoutils.MustSet(store, key2, &twap)
}

func (k twapkeeper) deleteHistoricalTWAP(ctx sdk.Context, twap types.TwapRecord) {
	store := ctx.KVStore(k.storeKey)
	key1 := types.FormatHistoricalTimeIndexTWAPKey(twap.Time, twap.PoolId, twap.Asset0Denom, twap.Asset1Denom)
	key2 := types.FormatHistoricalPoolIndexTWAPKey(twap.PoolId, twap.Time, twap.Asset0Denom, twap.Asset1Denom)
	store.Delete(key1)
	store.Delete(key2)
}

func (k twapkeeper) getMostRecentTWAP(ctx sdk.Context, poolId uint64, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatMostRecentTWAPKey(poolId, asset0Denom, asset1Denom)
	bz := store.Get(key)
	return types.ParseTwapFromBz(bz)
}

func (k twapkeeper) getAllMostRecentTWAPsForPool(ctx sdk.Context, poolId uint64) ([]types.TwapRecord, error) {
	store := ctx.KVStore(k.storeKey)
	return types.GetAllMostRecentTwapsForPool(store, poolId)
}

func (k twapkeeper) storeMostRecentTWAP(ctx sdk.Context, twap types.TwapRecord) {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatMostRecentTWAPKey(twap.PoolId, twap.Asset0Denom, twap.Asset1Denom)
	osmoutils.MustSet(store, key, &twap)
}

// returns an error if theres no historical record at or before time.
// (Asking for a time too far back)
func (k twapkeeper) getTwapBeforeTime(ctx sdk.Context, poolId uint64, time time.Time, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	store := ctx.KVStore(k.storeKey)
	startKey := types.FormatHistoricalPoolIndexTimePrefix(poolId, time)
	// TODO: Optimize to cut down search on asset0Denom, asset1denom.
	// Not really important, since primarily envisioning 2 asset pools
	stopFn := func(key []byte) bool {
		return types.ParseTimeFromHistoricalPoolIndexKey(key).After(time)
	}

	twaps, err := osmoutils.GetValuesUntilDerivedStop(store, startKey, stopFn, types.ParseTwapFromBz)
	if err != nil {
		return types.TwapRecord{}, err
	}
	if len(twaps) == 0 {
		return types.TwapRecord{}, errors.New("looking for a time thats too old, not in the historical index. " +
			" Try storing the accumulator value.")
	}

	for _, twap := range twaps {
		if twap.Asset0Denom == asset0Denom && twap.Asset1Denom == asset1Denom {
			return twap, nil
		}
	}
	return types.TwapRecord{}, errors.New("Something went wrong - TWAP not found, but there are twaps available for this time." +
		" Were provided asset0denom and asset1denom correct?")
}
