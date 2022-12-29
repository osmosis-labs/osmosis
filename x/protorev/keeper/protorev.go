package keeper

import (
	"fmt"
	"strconv"

	"github.com/osmosis-labs/osmosis/v14/x/protorev/types"

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

// ---------------------- Config Stores  ---------------------- //

// GetDaysSinceModuleGenesis returns the number of days since the module was initialized
func (k Keeper) GetDaysSinceModuleGenesis(ctx sdk.Context) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDaysSinceGenesis)
	bz := store.Get(types.KeyPrefixDaysSinceGenesis)
	if bz == nil {
		// This should never happen as the module is initialized with 0 days on genesis
		return 0, fmt.Errorf("days since module genesis not found")
	}

	daysSinceGenesis := sdk.BigEndianToUint64(bz)

	return daysSinceGenesis, nil
}

// SetDaysSinceModuleGenesis updates the number of days since genesis
func (k Keeper) SetDaysSinceModuleGenesis(ctx sdk.Context, daysSinceGenesis uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDaysSinceGenesis)
	store.Set(types.KeyPrefixDaysSinceGenesis, sdk.Uint64ToBigEndian(daysSinceGenesis))
}

// GetDeveloperFees returns the fees the developers can withdraw from the module account
func (k Keeper) GetDeveloperFees(ctx sdk.Context, denom string) (sdk.Coin, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDeveloperFees)
	key := types.GetKeyPrefixDeveloperFees(denom)

	bz := store.Get(key)
	if bz == nil {
		return sdk.Coin{}, fmt.Errorf("developer fees for %s not found", denom)
	}

	developerFees := sdk.Coin{}
	err := developerFees.Unmarshal(bz)
	if err != nil {
		return sdk.Coin{}, err
	}

	return developerFees, nil
}

// GetAllDeveloperFees returns all the developer fees (Osmo and Atom since these are the only two tradable assets)
func (k Keeper) GetAllDeveloperFees(ctx sdk.Context) []sdk.Coin {
	fees := make([]sdk.Coin, 0)

	// Get Osmo fees
	if fee, err := k.GetDeveloperFees(ctx, types.OsmosisDenomination); err == nil {
		fees = append(fees, fee)
	}

	// Get Atom fees
	if fee, err := k.GetDeveloperFees(ctx, types.AtomDenomination); err == nil {
		fees = append(fees, fee)
	}

	return fees
}

// SetDeveloperFees sets the fees the developers can withdraw from the module account
func (k Keeper) SetDeveloperFees(ctx sdk.Context, developerFees sdk.Coin) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDeveloperFees)
	key := types.GetKeyPrefixDeveloperFees(developerFees.Denom)

	bz, err := developerFees.Marshal()
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

// DeleteDeveloperFees deletes the developer fees given a denom
func (k Keeper) DeleteDeveloperFees(ctx sdk.Context, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDeveloperFees)
	key := types.GetKeyPrefixDeveloperFees(denom)
	store.Delete(key)
}

// GetProtoRevEnabled returns whether protorev is enabled
func (k Keeper) GetProtoRevEnabled(ctx sdk.Context) (bool, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixProtoRevEnabled)
	bz := store.Get(types.KeyPrefixProtoRevEnabled)
	if bz == nil {
		// This should never happen as the module is initialized on genesis
		return false, fmt.Errorf("protorev enabled/disabled configuration has not been set in state")
	}

	res, err := strconv.ParseBool(string(bz))
	if err != nil {
		return false, err
	}

	return res, nil
}

// SetProtoRevEnabled sets whether the protorev post handler is enabled
func (k Keeper) SetProtoRevEnabled(ctx sdk.Context, enabled bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixProtoRevEnabled)
	bz := []byte(strconv.FormatBool(enabled))
	store.Set(types.KeyPrefixProtoRevEnabled, bz)
}

// GetRouteCountForBlock returns the number of routes that have been traversed in the current block
func (k Keeper) GetRouteCountForBlock(ctx sdk.Context) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixRouteCountForBlock)
	bz := store.Get(types.KeyPrefixRouteCountForBlock)
	if bz == nil {
		// This should never happen as this is set to 0 on genesis
		return 0, fmt.Errorf("current route count has not been set in state")
	}

	res := sdk.BigEndianToUint64(bz)

	return res, nil
}

// SetRouteCountForBlock sets the number of routes that have been traversed in the current block
func (k Keeper) SetRouteCountForBlock(ctx sdk.Context, txCount uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixRouteCountForBlock)
	store.Set(types.KeyPrefixRouteCountForBlock, sdk.Uint64ToBigEndian(txCount))
}

// IncrementRouteCountForBlock increments the number of routes that have been traversed in the current block
func (k Keeper) IncrementRouteCountForBlock(ctx sdk.Context, amount uint64) error {
	routeCount, err := k.GetRouteCountForBlock(ctx)
	if err != nil {
		return err
	}

	k.SetRouteCountForBlock(ctx, routeCount+amount)

	return nil
}

// GetLatestBlockHeight returns the latest block height that protorev was run on
func (k Keeper) GetLatestBlockHeight(ctx sdk.Context) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixLatestBlockHeight)
	bz := store.Get(types.KeyPrefixLatestBlockHeight)
	if bz == nil {
		// This should never happen as the module is initialized on genesis and reset in the post handler
		return 0, fmt.Errorf("block height has not been set in state")
	}

	res := sdk.BigEndianToUint64(bz)

	return res, nil
}

// SetLatestBlockHeight sets the latest block height that protorev was run on
func (k Keeper) SetLatestBlockHeight(ctx sdk.Context, blockHeight uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixLatestBlockHeight)
	store.Set(types.KeyPrefixLatestBlockHeight, sdk.Uint64ToBigEndian(blockHeight))
}

// ---------------------- Admin Stores  ---------------------- //

// GetAdminAccount returns the admin account for protorev
func (k Keeper) GetAdminAccount(ctx sdk.Context) (sdk.AccAddress, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAdminAccount)
	bz := store.Get(types.KeyPrefixAdminAccount)
	if bz == nil {
		return nil, fmt.Errorf("admin account not found, it has not been initialized through governance")
	}

	return sdk.AccAddress(bz), nil
}

// SetAdminAccount sets the admin account for protorev
func (k Keeper) SetAdminAccount(ctx sdk.Context, adminAccount sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAdminAccount)
	store.Set(types.KeyPrefixAdminAccount, adminAccount.Bytes())
}

// GetDeveloperAccount returns the developer account for protorev
func (k Keeper) GetDeveloperAccount(ctx sdk.Context) (sdk.AccAddress, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDeveloperAccount)
	bz := store.Get(types.KeyPrefixDeveloperAccount)
	if bz == nil {
		return nil, fmt.Errorf("developer account not found, it has not been initialized by the admin account")
	}

	return sdk.AccAddress(bz), nil
}

// SetDeveloperAccount sets the developer account for protorev that will receive a portion of arbitrage profits
func (k Keeper) SetDeveloperAccount(ctx sdk.Context, developerAccount sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDeveloperAccount)
	store.Set(types.KeyPrefixDeveloperAccount, developerAccount.Bytes())
}

// GetMaxRoutesPerTx returns the max number of routes that can be iterated per transaction
func (k Keeper) GetMaxRoutesPerTx(ctx sdk.Context) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixMaxRoutesPerTx)
	bz := store.Get(types.KeyPrefixMaxRoutesPerTx)
	if bz == nil {
		// This should never happen as it is set to the default value on genesis
		return 0, fmt.Errorf("max routes configuration has not been set in state")
	}

	res := sdk.BigEndianToUint64(bz)
	return res, nil
}

// SetMaxRoutesPerTx sets the max number of routes that can be iterated per transaction
func (k Keeper) SetMaxRoutesPerTx(ctx sdk.Context, maxRoutes uint64) error {
	if maxRoutes == 0 || maxRoutes > types.MaxIterableRoutesPerTx {
		return fmt.Errorf("max routes must be between 1 and %d", types.MaxIterableRoutesPerTx)
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixMaxRoutesPerTx)
	bz := sdk.Uint64ToBigEndian(maxRoutes)
	store.Set(types.KeyPrefixMaxRoutesPerTx, bz)

	return nil
}

// GetMaxRoutesPerBlock returns the max number of routes that can be iterated per block
func (k Keeper) GetMaxRoutesPerBlock(ctx sdk.Context) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixMaxRoutesPerBlock)
	bz := store.Get(types.KeyPrefixMaxRoutesPerBlock)
	if bz == nil {
		// This should never happen as it is set to the default value on genesis
		return 0, fmt.Errorf("max routes configuration has not been set in state")
	}

	res := sdk.BigEndianToUint64(bz)
	return res, nil
}

// SetMaxRoutesPerBlock sets the max number of routes that can be iterated per block
func (k Keeper) SetMaxRoutesPerBlock(ctx sdk.Context, maxRoutes uint64) error {
	if maxRoutes == 0 || maxRoutes > types.MaxIterableRoutesPerBlock {
		return fmt.Errorf("max routes per block must be between 1 and %d", types.MaxIterableRoutesPerBlock)
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixMaxRoutesPerBlock)
	bz := sdk.Uint64ToBigEndian(maxRoutes)
	store.Set(types.KeyPrefixMaxRoutesPerBlock, bz)

	return nil
}
