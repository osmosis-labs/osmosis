package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/x/osmolbp/proto"
)

var (
	poolSeqStoreKey = []byte{0} // pool id sequence
	lbpStoreKey     = []byte{1} // lbp objects
	vaultStoreKey   = byte(2)   // user-pool objects

	poolRes byte = 100 // poolId -> reserves  (token_out balances)
)

func (k *Keeper) savePool(modulestore storetypes.KVStore, id []byte, p *proto.LBP) {
	store := k.lbpStore(modulestore)
	store.Set(id, k.cdc.MustMarshal(p))
}

// returns pool, pool bytes id, error
func (k *Keeper) getPool(modulestore storetypes.KVStore, id uint64) (proto.LBP, []byte, error) {
	store := k.lbpStore(modulestore)
	idBz := make([]byte, 8)
	binary.BigEndian.PutUint64(idBz, id)
	bz := store.Get(idBz)
	var p proto.LBP
	if bz == nil {
		return p, idBz, errors.Wrap(errors.ErrKeyNotFound, "pool doesn't exist")
	}
	err := k.cdc.Unmarshal(bz, &p)
	return p, idBz, err
}

// returns pool, found (bool), error
func (k *Keeper) getUserVault(modulestore storetypes.KVStore, poolId []byte, addr sdk.AccAddress) (proto.UserPosition, bool, error) {
	store := k.userVaultStore(modulestore, poolId)
	bz := store.Get(addr)
	var v proto.UserPosition
	if bz == nil {
		return v, false, nil
	}
	err := k.cdc.Unmarshal(bz, &v)
	return v, true, err
}

// returns pool, found (bool), error
func (k *Keeper) saveUserVault(modulestore storetypes.KVStore, poolId []byte, addr sdk.AccAddress, v *proto.UserPosition) {
	store := k.userVaultStore(modulestore, poolId)
	store.Set(addr, k.cdc.MustMarshal(v))
}

// returns pool, found (bool), error
func (k *Keeper) delUserVault(modulestore storetypes.KVStore, poolId []byte, addr sdk.AccAddress) {
	store := k.userVaultStore(modulestore, poolId)
	store.Delete(addr)
}

func (k Keeper) lbpStore(moduleStore storetypes.KVStore) prefix.Store {
	return prefix.NewStore(moduleStore, lbpStoreKey)
}

func (k Keeper) userVaultStore(moduleStore storetypes.KVStore, poolId []byte) prefix.Store {
	p := make([]byte, 1+len(poolId))
	p[0] = vaultStoreKey
	copy(p[1:], poolId)
	return prefix.NewStore(moduleStore, p)
}

//  TODO: can remove this
// func (k Keeper) depositStore(ctx sdk.Context, addr sdk.AccAddress) prefix.Store {
// 	a := mustLengthPrefix(addr)
// 	p := make([]byte, 1+len(a))
// 	p[0] = deposits
// 	copy(p[1:], addr)
// 	store := ctx.KVStore(k.storeKey)
// 	return prefix.NewStore(store, p)
// }

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
