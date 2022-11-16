package keeper

import (
	"encoding/binary"

	"github.com/osmosis-labs/osmosis/v12/x/protorev/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// GetOsmoPool returns the pool id of the Osmo pool for the given denom paired with Osmo
func (k Keeper) GetOsmoPool(ctx sdk.Context, denom string) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.KeyPrefixOsmoPools))
	key := types.GetKeyPrefixOsmoPool(denom)

	bz := store.Get(key)
	if len(bz) == 0 {
		return 0, sdkerrors.Wrapf(ErrNoOsmoPool, "denom: %s", denom)
	}

	poolId := BytesToUInt64(bz)
	return poolId, nil
}

// SetOsmoPool sets the pool id of the Osmo pool for the given denom paired with Osmo
func (k Keeper) SetOsmoPool(ctx sdk.Context, denom string, poolId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.KeyPrefixOsmoPools))
	key := types.GetKeyPrefixOsmoPool(denom)

	store.Set(key, UInt64ToBytes(poolId))
}

// DeleteOsmoPool deletes the pool id of the Osmo pool for the given denom paired with Osmo
func (k Keeper) DeleteOsmoPool(ctx sdk.Context, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.KeyPrefixOsmoPools))
	key := types.GetKeyPrefixOsmoPool(denom)

	store.Delete(key)
}

// GetAtomPool returns the pool id of the Atom pool for the given denom paired with Atom
func (k Keeper) GetAtomPool(ctx sdk.Context, denom string) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.KeyPrefixAtomPools))
	key := types.GetKeyPrefixAtomPool(denom)

	bz := store.Get(key)
	if len(bz) == 0 {
		return 0, sdkerrors.Wrapf(ErrNoAtomPool, "denom: %s", denom)
	}

	poolId := BytesToUInt64(bz)
	return poolId, nil
}

// SetAtomPool sets the pool id of the Atom pool for the given denom paired with Atom
func (k Keeper) SetAtomPool(ctx sdk.Context, denom string, poolId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.KeyPrefixAtomPools))
	key := types.GetKeyPrefixAtomPool(denom)

	store.Set(key, UInt64ToBytes(poolId))
}

// DeleteAtomPool deletes the pool id of the Atom pool for the given denom paired with Atom
func (k Keeper) DeleteAtomPool(ctx sdk.Context, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.KeyPrefixAtomPools))
	key := types.GetKeyPrefixAtomPool(denom)

	store.Delete(key)
}

// GetRoute returns the route given two denoms
func (k Keeper) GetRoute(ctx sdk.Context, denom1 string, denom2 string) (*types.Route, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.KeyPrefixRoutes))
	key := types.GetKeyPrefixRouteForPair(denom1, denom2)

	bz := store.Get(key)
	if len(bz) == 0 {
		return nil, sdkerrors.Wrapf(ErrNoRoute, "denom1: %s, denom2: %s", denom1, denom2)
	}

	route := &types.Route{}
	route.Unmarshal(bz)

	return route, nil
}

// SetRoute sets the route given two denoms
func (k Keeper) SetRoute(ctx sdk.Context, denom1 string, denom2 string, route *types.Route) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.KeyPrefixRoutes))
	key := types.GetKeyPrefixRouteForPair(denom1, denom2)

	bz, _ := route.Marshal()
	store.Set(key, bz)
}

// GetProtoRevStatistics returns the ProtoRevStatistics
func (k Keeper) GetProtoRevStatistics(ctx sdk.Context) (*types.ProtoRevStatistics, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.KeyPrefixProtoRevStatistics))

	bz := store.Get(types.KeyPrefixProtoRevStatistics)
	if len(bz) == 0 {
		// This should never happen because the statistics are initialized on genesis
		return nil, ErrNoProtoRevStatistics
	}

	protoRevStatistics := &types.ProtoRevStatistics{}
	protoRevStatistics.Unmarshal(bz)

	return protoRevStatistics, nil
}

// SetProtoRevStatistics sets the ProtoRevStatistics
func (k Keeper) SetProtoRevStatistics(ctx sdk.Context, protoRevStatistics *types.ProtoRevStatistics) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.KeyPrefixProtoRevStatistics))

	bz, _ := protoRevStatistics.Marshal()
	store.Set(types.KeyPrefixProtoRevStatistics, bz)
}

// Converts a uint64 to a []byte
func UInt64ToBytes(number uint64) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, number)
	return bz
}

// Converts a []byte into a uint64
func BytesToUInt64(bz []byte) uint64 {
	return uint64(binary.LittleEndian.Uint64(bz))
}
