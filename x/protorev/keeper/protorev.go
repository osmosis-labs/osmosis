package keeper

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v13/x/protorev/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ---------------------- Trading Stores  ---------------------- //

// GetTokenPairArbRoutes returns the token pair arb routes given two denoms
func (k Keeper) GetTokenPairArbRoutes(ctx sdk.Context, tokenA, tokenB string) (*types.TokenPairArbRoutes, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTokenPairRoutes)
	key := types.GetKeyPrefixRouteForTokenPair(tokenA, tokenB)

	bz := store.Get(key)
	if len(bz) == 0 {
		return nil, fmt.Errorf("no routes found for token pair %s-%s", tokenA, tokenB)
	}

	tokenPairArbRoutes := &types.TokenPairArbRoutes{}
	err := tokenPairArbRoutes.Unmarshal(bz)
	if err != nil {
		return nil, err
	}

	return tokenPairArbRoutes, nil
}

// GetAllTokenPairArbRoutes returns all the token pair arb routes
func (k Keeper) GetAllTokenPairArbRoutes(ctx sdk.Context) ([]*types.TokenPairArbRoutes, error) {
	routes := make([]*types.TokenPairArbRoutes, 0)

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixTokenPairRoutes)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		tokenPairArbRoutes := &types.TokenPairArbRoutes{}
		err := tokenPairArbRoutes.Unmarshal(iterator.Value())
		if err != nil {
			return nil, err
		}

		routes = append(routes, tokenPairArbRoutes)
	}

	return routes, nil
}

// SetTokenPairArbRoutes sets the token pair arb routes given two denoms
func (k Keeper) SetTokenPairArbRoutes(ctx sdk.Context, tokenA, tokenB string, tokenPair *types.TokenPairArbRoutes) (*types.TokenPairArbRoutes, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTokenPairRoutes)
	key := types.GetKeyPrefixRouteForTokenPair(tokenA, tokenB)

	bz, err := tokenPair.Marshal()
	if err != nil {
		return tokenPair, err
	}

	store.Set(key, bz)

	return tokenPair, nil
}

// DeleteAllTokenPairArbRoutes deletes all the token pair arb routes
func (k Keeper) DeleteAllTokenPairArbRoutes(ctx sdk.Context) {
	k.DeleteAllEntriesForKeyPrefix(ctx, types.KeyPrefixTokenPairRoutes)
}

// GetOsmoPool returns the pool id of the Osmo pool for the given denom paired with Osmo
func (k Keeper) GetOsmoPool(ctx sdk.Context, denom string) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixOsmoPools)
	key := types.GetKeyPrefixOsmoPool(denom)

	bz := store.Get(key)
	if len(bz) == 0 {
		return 0, fmt.Errorf("no osmo pool for denom %s", denom)
	}

	poolId := sdk.BigEndianToUint64(bz)
	return poolId, nil
}

// SetOsmoPool sets the pool id of the Osmo pool for the given denom paired with Osmo
func (k Keeper) SetOsmoPool(ctx sdk.Context, denom string, poolId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixOsmoPools)
	key := types.GetKeyPrefixOsmoPool(denom)

	store.Set(key, sdk.Uint64ToBigEndian(poolId))
}

// DeleteAllOsmoPools deletes all the Osmo pools from modules store
func (k Keeper) DeleteAllOsmoPools(ctx sdk.Context) {
	k.DeleteAllEntriesForKeyPrefix(ctx, types.KeyPrefixOsmoPools)
}

// GetAtomPool returns the pool id of the Atom pool for the given denom paired with Atom
func (k Keeper) GetAtomPool(ctx sdk.Context, denom string) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAtomPools)
	key := types.GetKeyPrefixAtomPool(denom)

	bz := store.Get(key)
	if len(bz) == 0 {
		return 0, fmt.Errorf("no atom pool for denom %s", denom)
	}

	poolId := sdk.BigEndianToUint64(bz)
	return poolId, nil
}

// SetAtomPool sets the pool id of the Atom pool for the given denom paired with Atom
func (k Keeper) SetAtomPool(ctx sdk.Context, denom string, poolId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAtomPools)
	key := types.GetKeyPrefixAtomPool(denom)

	store.Set(key, sdk.Uint64ToBigEndian(poolId))
}

// DeleteAllAtomPools deletes all the Atom pools from modules store
func (k Keeper) DeleteAllAtomPools(ctx sdk.Context) {
	k.DeleteAllEntriesForKeyPrefix(ctx, types.KeyPrefixAtomPools)
}

// DeleteAllEntriesForKeyPrefix deletes all the entries from the store for the given key prefix
func (k Keeper) DeleteAllEntriesForKeyPrefix(ctx sdk.Context, keyPrefix []byte) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, keyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}
