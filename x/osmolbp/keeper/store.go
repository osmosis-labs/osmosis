package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/x/osmolbp"
	"github.com/osmosis-labs/osmosis/x/osmolbp/api"
)

// StoreKey is the store key string for osmolbp
const StoreKey = osmolbp.ModuleName

var (
	lbpSeqStoreKey = []byte{0} // lbp id sequence
	lbpStoreKey    = []byte{1} // lbp objects
	userStoreKey   = byte(2)   // userPosition objects
)

func (k *Keeper) saveLBP(modulestore storetypes.KVStore, id []byte, p *api.LBP) {
	store := k.lbpStore(modulestore)
	store.Set(id, k.cdc.MustMarshal(p))
}

// returns pool, pool bytes id, error
func (k *Keeper) getLBP(modulestore storetypes.KVStore, id uint64) (api.LBP, []byte, error) {
	store := k.lbpStore(modulestore)
	idBz := storeIntIdKey(id)
	bz := store.Get(idBz)
	var p api.LBP
	if bz == nil {
		return p, idBz, errors.Wrap(errors.ErrKeyNotFound, "pool doesn't exist")
	}
	err := k.cdc.Unmarshal(bz, &p)
	return p, idBz, err
}

// gets or creates (when create == true) user position object
// returns pool, error
// return errors.NotFound whene the object is not there and create == false
func (k *Keeper) getUserPosition(modulestore storetypes.KVStore, poolId []byte, user sdk.AccAddress, create bool) (api.UserPosition, error) {
	store := k.userPositionStore(modulestore, poolId)
	bz := store.Get(user)
	var v api.UserPosition
	if bz == nil {
		if create == false {
			return v, errors.ErrNotFound.Wrap("user position for given LBP is not found")
		}
		return newUserPosition(), nil
	}
	err := k.cdc.Unmarshal(bz, &v)
	return v, err
}

// returns pool, found (bool), error
func (k *Keeper) saveUserPosition(modulestore storetypes.KVStore, poolId []byte, addr sdk.AccAddress, v *api.UserPosition) {
	store := k.userPositionStore(modulestore, poolId)
	store.Set(addr, k.cdc.MustMarshal(v))
}

// returns pool, found (bool), error
func (k *Keeper) delUserPosition(modulestore storetypes.KVStore, poolId []byte, addr sdk.AccAddress) {
	store := k.userPositionStore(modulestore, poolId)
	store.Delete(addr)
}

func (k Keeper) lbpStore(moduleStore storetypes.KVStore) prefix.Store {
	return prefix.NewStore(moduleStore, lbpStoreKey)
}

func (k Keeper) userPositionStore(moduleStore storetypes.KVStore, poolId []byte) prefix.Store {
	p := make([]byte, 1+len(poolId))
	p[0] = userStoreKey
	copy(p[1:], poolId)
	return prefix.NewStore(moduleStore, p)
}

// MustLengthPrefix is LengthPrefix with panic on error.
// TODO: use address.MustLengthPrefix when moving to SDK 0.44+
// func mustLengthPrefix(bz []byte) []byte {
// 	res, err := lengthPrefix(bz)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return res
// }

// MaxAddrLen is the maximum allowed length (in bytes) for an address.
const MaxAddrLen = 255

func lengthPrefix(bz []byte) ([]byte, error) {
	bzLen := len(bz)
	if bzLen == 0 {
		return bz, nil
	}

	if bzLen > MaxAddrLen {
		return nil, errors.Wrapf(sdkerrors.ErrUnknownAddress, "address length should be max %d bytes, got %d", MaxAddrLen, bzLen)
	}

	return append([]byte{byte(bzLen)}, bz...), nil
}

// user: bech32 user address
func (k Keeper) getUserAndLBP(modulestore storetypes.KVStore, poolId uint64, user string, create bool) (sdk.AccAddress, *api.LBP, []byte, *api.UserPosition, error) {
	userAddr, err := sdk.AccAddressFromBech32(user)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	p, poolIdBz, err := k.getLBP(modulestore, poolId)
	if err != nil {
		return userAddr, &p, poolIdBz, nil, err
	}
	u, err := k.getUserPosition(modulestore, poolIdBz, userAddr, create)
	return userAddr, &p, poolIdBz, &u, err
}

func storeIntIdKey(id uint64) []byte {
	idBz := make([]byte, 8)
	binary.BigEndian.PutUint64(idBz, id)
	return idBz
}
