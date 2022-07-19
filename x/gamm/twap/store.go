package twap

import (
	"encoding/binary"

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

func (k twapkeeper) getChangedPools(ctx sdk.Context) []uint64 {
	store := ctx.TransientStore(&k.transientKey)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	alteredPoolIds := []uint64{}
	for ; iter.Key() != nil; iter.Next() {
		k := iter.Key()
		poolId := binary.LittleEndian.Uint64(k)
		alteredPoolIds = append(alteredPoolIds, poolId)
	}
	return alteredPoolIds
}

func (k twapkeeper) storeHistoricalTWAP(ctx sdk.Context, twap types.TwapRecord) {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatHistoricalTWAPKey(twap.PoolId, twap.Time, twap.Asset0Denom, twap.Asset1Denom)
	osmoutils.MustSet(store, key, &twap)
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
