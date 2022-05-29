package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/x/launchpad"
	"github.com/osmosis-labs/osmosis/x/launchpad/api"
)

// StoreKey is the store key string for launchpad
const StoreKey = launchpad.ModuleName

var (
	storeSeqStoreKey = []byte{0} // sale id sequence
	storeStoreKey    = []byte{1} // sale objects
	userStoreKey     = byte(2)   // userPosition objects
)

func (k *Keeper) saveSale(modulestore storetypes.KVStore, id []byte, p *api.Sale) {
	store := k.saleStore(modulestore)
	store.Set(id, k.cdc.MustMarshal(p))
}

// returns sale, sale bytes id, error
func (k *Keeper) getSale(modulestore storetypes.KVStore, id uint64) (api.Sale, []byte, error) {
	store := k.saleStore(modulestore)
	idBz := storeIntIdKey(id)
	bz := store.Get(idBz)
	var p api.Sale
	if bz == nil {
		return p, idBz, errors.Wrap(errors.ErrKeyNotFound, "sale doesn't exist")
	}
	err := k.cdc.Unmarshal(bz, &p)
	return p, idBz, err
}

// gets or creates (when create == true) user position object
// returns sale, error
// return errors.NotFound whene the object is not there and create == false
func (k *Keeper) getUserPosition(modulestore storetypes.KVStore, saleId []byte, user sdk.AccAddress, create bool) (api.UserPosition, error) {
	store := k.userPositionStore(modulestore, saleId)
	bz := store.Get(user)
	var v api.UserPosition
	if bz == nil {
		if create == false {
			return v, errors.ErrNotFound.Wrap("user position for given Sale is not found")
		}
		return newUserPosition(), nil
	}
	err := k.cdc.Unmarshal(bz, &v)
	return v, err
}

// returns sale, found (bool), error
func (k *Keeper) saveUserPosition(modulestore storetypes.KVStore, saleId []byte, addr sdk.AccAddress, v *api.UserPosition) {
	store := k.userPositionStore(modulestore, saleId)
	store.Set(addr, k.cdc.MustMarshal(v))
}

// returns sale, found (bool), error
func (k *Keeper) delUserPosition(modulestore storetypes.KVStore, saleId []byte, addr sdk.AccAddress) {
	store := k.userPositionStore(modulestore, saleId)
	store.Delete(addr)
}

func (k Keeper) saleStore(moduleStore storetypes.KVStore) prefix.Store {
	return prefix.NewStore(moduleStore, storeStoreKey)
}

func (k Keeper) userPositionStore(moduleStore storetypes.KVStore, saleId []byte) prefix.Store {
	p := make([]byte, 1+len(saleId))
	p[0] = userStoreKey
	copy(p[1:], saleId)
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
func (k Keeper) getUserAndSale(modulestore storetypes.KVStore, saleId uint64, user string, create bool) (sdk.AccAddress, *api.Sale, []byte, *api.UserPosition, error) {
	userAddr, err := sdk.AccAddressFromBech32(user)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	p, saleIdBz, err := k.getSale(modulestore, saleId)
	if err != nil {
		return userAddr, &p, saleIdBz, nil, err
	}
	u, err := k.getUserPosition(modulestore, saleIdBz, userAddr, create)
	return userAddr, &p, saleIdBz, &u, err
}

func storeIntIdKey(id uint64) []byte {
	idBz := make([]byte, 8)
	binary.BigEndian.PutUint64(idBz, id)
	return idBz
}
