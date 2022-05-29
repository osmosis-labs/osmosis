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
// the next Pool and updates the pool ID counter.
func (k Keeper) nextPoolID(store storetypes.KVStore) (uint64, []byte) {
	store = prefix.NewStore(store, storeSeqStoreKey)
	bz := store.Get(ormPoolID)
	if bz == nil {
		bz = make([]byte, 8)
		store.Set(ormPoolID, bz)
		return 0, bz
	}
	i := binary.BigEndian.Uint64(bz)
	binary.BigEndian.PutUint64(bz, i+1)
	store.Set(ormPoolID, bz)
	return i, bz
}
