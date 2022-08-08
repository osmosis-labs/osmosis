package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

var (
	ormPoolID = []byte{0}
)

// Takes a module store, returns an ID (uint64 and it's store bytes representation) for
// the next Pool and updates the sale ID counter in the store.
func (k Keeper) nextSaleID(moduleStore storetypes.KVStore) (uint64, []byte) {
	i, bz := getNextSaleID(moduleStore)
	bzNext := make([]byte, 8)
	binary.BigEndian.PutUint64(bzNext, i+1)
	store := prefix.NewStore(moduleStore, storeSeqStoreKey)
	store.Set(ormPoolID, bzNext)
	return i, bz
}

func getNextSaleID(moduleStore storetypes.KVStore) (uint64, []byte) {
	store := prefix.NewStore(moduleStore, storeSeqStoreKey)
	bz := store.Get(ormPoolID)
	if bz == nil {
		bz = make([]byte, 8)
		store.Set(ormPoolID, bz)
		return 0, bz
	}
	return binary.BigEndian.Uint64(bz), bz
}

func (k Keeper) setNextSaleID(moduleStore storetypes.KVStore, id uint64) {
	bzNext := storeIntIdKey(id)
	store := prefix.NewStore(moduleStore, storeSeqStoreKey)
	store.Set(ormPoolID, bzNext)
}