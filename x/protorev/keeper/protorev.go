package keeper

import (
	"errors"
	"fmt"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"

	"cosmossdk.io/store/prefix"

	"github.com/osmosis-labs/osmosis/osmoutils"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ---------------------- Trading Stores  ---------------------- //

// GetTokenPairArbRoutes returns the token pair arb routes given two denoms
func (k Keeper) GetTokenPairArbRoutes(ctx sdk.Context, tokenA, tokenB string) (types.TokenPairArbRoutes, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTokenPairRoutes)
	key := types.GetKeyPrefixRouteForTokenPair(tokenA, tokenB)

	bz := store.Get(key)
	if len(bz) == 0 {
		return types.TokenPairArbRoutes{}, fmt.Errorf("no routes found for token pair %s-%s", tokenA, tokenB)
	}

	tokenPairArbRoutes := types.TokenPairArbRoutes{}
	err := tokenPairArbRoutes.Unmarshal(bz)
	if err != nil {
		return types.TokenPairArbRoutes{}, err
	}

	return tokenPairArbRoutes, nil
}

// GetAllTokenPairArbRoutes returns all the token pair arb routes
func (k Keeper) GetAllTokenPairArbRoutes(ctx sdk.Context) ([]types.TokenPairArbRoutes, error) {
	routes := make([]types.TokenPairArbRoutes, 0)

	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefixTokenPairRoutes)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		tokenPairArbRoutes := types.TokenPairArbRoutes{}
		err := tokenPairArbRoutes.Unmarshal(iterator.Value())
		if err != nil {
			return nil, err
		}

		routes = append(routes, tokenPairArbRoutes)
	}

	return routes, nil
}

// SetTokenPairArbRoutes sets the token pair arb routes given two denoms
func (k Keeper) SetTokenPairArbRoutes(ctx sdk.Context, tokenA, tokenB string, tokenPair types.TokenPairArbRoutes) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTokenPairRoutes)
	key := types.GetKeyPrefixRouteForTokenPair(tokenA, tokenB)

	bz, err := tokenPair.Marshal()
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

// DeleteAllTokenPairArbRoutes deletes all the token pair arb routes
func (k Keeper) DeleteAllTokenPairArbRoutes(ctx sdk.Context) {
	k.DeleteAllEntriesForKeyPrefix(ctx, types.KeyPrefixTokenPairRoutes)
}

// DeprecatedGetAllBaseDenoms returns all of the base denoms (sorted by priority in descending order) used to build cyclic arbitrage routes
// After v24 upgrade, this method should be deleted. We now use the param store.
func (k Keeper) DeprecatedGetAllBaseDenoms(ctx sdk.Context) ([]types.BaseDenom, error) {
	baseDenoms := make([]types.BaseDenom, 0)

	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefixDeprecatedBaseDenoms)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		baseDenom := types.BaseDenom{}
		err := baseDenom.Unmarshal(iterator.Value())
		if err != nil {
			return []types.BaseDenom{}, err
		}

		baseDenoms = append(baseDenoms, baseDenom)
	}

	return baseDenoms, nil
}

// DeprecatedSetBaseDenoms sets all of the base denoms used to build cyclic arbitrage routes. The base denoms priority
// order is going to match the order of the base denoms in the slice.
// After v24 upgrade, this method should be deleted. We now use the param store.
func (k Keeper) DeprecatedSetBaseDenoms(ctx sdk.Context, baseDenoms []types.BaseDenom) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDeprecatedBaseDenoms)

	for i, baseDenom := range baseDenoms {
		key := types.DeprecatedGetKeyPrefixBaseDenom(uint64(i))

		bz, err := baseDenom.Marshal()
		if err != nil {
			return err
		}
		store.Set(key, bz)
	}

	return nil
}

// DeprecatedDeleteBaseDenoms deletes all of the base denoms.
// After v24 upgrade, this method should be deleted. We now use the param store.
func (k Keeper) DeprecatedDeleteBaseDenoms(ctx sdk.Context) {
	k.DeleteAllEntriesForKeyPrefix(ctx, types.KeyPrefixDeprecatedBaseDenoms)
}

// GetAllBaseDenoms returns all of the base denoms (sorted by priority in descending order) used to build cyclic arbitrage routes
func (k Keeper) GetAllBaseDenoms(ctx sdk.Context) ([]types.BaseDenom, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixBaseDenoms)
	baseDenoms := types.BaseDenoms{}
	err := baseDenoms.Unmarshal(bz)
	if err != nil {
		return []types.BaseDenom{}, err
	}
	return baseDenoms.BaseDenoms, nil
}

// SetBaseDenoms sets all of the base denoms used to build cyclic arbitrage routes. The base denoms priority
// order is going to match the order of the base denoms in the slice.
func (k Keeper) SetBaseDenoms(ctx sdk.Context, baseDenoms []types.BaseDenom) error {
	newBaseDenoms := types.BaseDenoms{BaseDenoms: baseDenoms}
	store := ctx.KVStore(k.storeKey)
	test, err := newBaseDenoms.Marshal()
	if err != nil {
		return err
	}

	store.Set(types.KeyPrefixBaseDenoms, test)
	return nil
}

// GetPoolForDenomPair returns the id of the highest liquidity pool between the base denom and the denom to match
func (k Keeper) GetPoolForDenomPair(ctx sdk.Context, baseDenom, denomToMatch string) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDenomPairToPool)
	key := types.GetKeyPrefixDenomPairToPool(baseDenom, denomToMatch)

	bz := store.Get(key)
	if len(bz) == 0 {
		return 0, types.NoPoolForDenomPairError{BaseDenom: baseDenom, MatchDenom: denomToMatch}
	}

	poolId := sdk.BigEndianToUint64(bz)
	return poolId, nil
}

// GetPoolForDenomPairNoOrder returns the id of the pool between the two denoms.
// It is order-independent. That is, tokenA can either be a base or a quote. Both cases are handled.
// If no pool exists, an error is returned.
// TODO: unit test
func (k Keeper) GetPoolForDenomPairNoOrder(ctx sdk.Context, tokenA, tokenB string) (uint64, error) {
	poolId, err := k.GetPoolForDenomPair(ctx, tokenA, tokenB)
	if err != nil {
		if errors.Is(err, types.NoPoolForDenomPairError{BaseDenom: tokenA, MatchDenom: tokenB}) {
			// Attempt changing base and match denoms.
			poolId, err = k.GetPoolForDenomPair(ctx, tokenB, tokenA)
			if err != nil {
				return 0, err
			}
		} else {
			return 0, err
		}
	}
	return poolId, nil
}

// SetPoolForDenomPair sets the id of the highest liquidty pool between the base denom and the denom to match
func (k Keeper) SetPoolForDenomPair(ctx sdk.Context, baseDenom, denomToMatch string, poolId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDenomPairToPool)
	key := types.GetKeyPrefixDenomPairToPool(baseDenom, denomToMatch)

	store.Set(key, sdk.Uint64ToBigEndian(poolId))
}

// DeleteAllPoolsForBaseDenom deletes all the pools for the given base denom
func (k Keeper) DeleteAllPoolsForBaseDenom(ctx sdk.Context, baseDenom string) {
	key := append(types.KeyPrefixDenomPairToPool, types.GetKeyPrefixDenomPairToPool(baseDenom, "")...)
	k.DeleteAllEntriesForKeyPrefix(ctx, key)
}

// SetSwapsToBackrun sets the swaps to backrun, updated via hooks
func (k Keeper) SetSwapsToBackrun(ctx sdk.Context, swapsToBackrun types.Route) error {
	store := prefix.NewStore(ctx.TransientStore(k.transientKey), types.KeyPrefixSwapsToBackrun)
	bz, err := swapsToBackrun.Marshal()
	if err != nil {
		return err
	}

	store.Set(types.KeyPrefixSwapsToBackrun, bz)

	return nil
}

// GetSwapsToBackrun returns the swaps to backrun, updated via hooks
func (k Keeper) GetSwapsToBackrun(ctx sdk.Context) (types.Route, error) {
	store := prefix.NewStore(ctx.TransientStore(k.transientKey), types.KeyPrefixSwapsToBackrun)
	bz := store.Get(types.KeyPrefixSwapsToBackrun)

	swapsToBackrun := types.Route{}
	err := swapsToBackrun.Unmarshal(bz)
	if err != nil {
		return types.Route{}, err
	}

	return swapsToBackrun, nil
}

// DeleteSwapsToBackrun deletes the swaps to backrun
func (k Keeper) DeleteSwapsToBackrun(ctx sdk.Context) {
	store := prefix.NewStore(ctx.TransientStore(k.transientKey), types.KeyPrefixSwapsToBackrun)
	store.Delete(types.KeyPrefixSwapsToBackrun)
}

// AddSwapToSwapsToBackrun appends a swap to the swaps to backrun
func (k Keeper) AddSwapsToSwapsToBackrun(ctx sdk.Context, swaps []types.Trade) error {
	swapsToBackrun, err := k.GetSwapsToBackrun(ctx)
	if err != nil {
		return err
	}

	swapsToBackrun.Trades = append(swapsToBackrun.Trades, swaps...)

	err = k.SetSwapsToBackrun(ctx, swapsToBackrun)
	if err != nil {
		return err
	}

	return nil
}

// DeleteAllEntriesForKeyPrefix deletes all the entries from the store for the given key prefix
func (k Keeper) DeleteAllEntriesForKeyPrefix(ctx sdk.Context, keyPrefix []byte) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, keyPrefix)

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
		return 0, errors.New("days since module genesis not found")
	}

	daysSinceGenesis := sdk.BigEndianToUint64(bz)

	return daysSinceGenesis, nil
}

// SetDaysSinceModuleGenesis updates the number of days since genesis
func (k Keeper) SetDaysSinceModuleGenesis(ctx sdk.Context, daysSinceGenesis uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDaysSinceGenesis)
	store.Set(types.KeyPrefixDaysSinceGenesis, sdk.Uint64ToBigEndian(daysSinceGenesis))
}

// Deprecated: Can be removed in v16
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

// Deprecated: Used in v16 upgrade, can be removed in v17
// GetAllDeveloperFees returns all the developer fees the developer account can withdraw
func (k Keeper) GetAllDeveloperFees(ctx sdk.Context) ([]sdk.Coin, error) {
	fees := make([]sdk.Coin, 0)

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDeveloperFees)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefixDeveloperFees)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		developerFees := sdk.Coin{}
		if err := developerFees.Unmarshal(iterator.Value()); err != nil {
			return nil, fmt.Errorf("error unmarshalling developer fees: %w", err)
		}

		fees = append(fees, developerFees)
	}

	return fees, nil
}

// Deprecated: Can be removed in v16
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

// Deprecated: Used in v16 upgrade, can be removed in v17
// DeleteDeveloperFees deletes the developer fees given a denom
func (k Keeper) DeleteDeveloperFees(ctx sdk.Context, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDeveloperFees)
	key := types.GetKeyPrefixDeveloperFees(denom)
	store.Delete(key)
}

// GetProtoRevEnabled returns whether protorev is enabled
func (k Keeper) GetProtoRevEnabled(ctx sdk.Context) bool {
	params := k.GetParams(ctx)
	return params.Enabled
}

// SetProtoRevEnabled sets whether the protorev post handler is enabled
func (k Keeper) SetProtoRevEnabled(ctx sdk.Context, enabled bool) {
	params := k.GetParams(ctx)
	params.Enabled = enabled
	k.SetParams(ctx, params)
}

// GetPointCountForBlock returns the number of pool points that have been consumed in the current block
func (k Keeper) GetPointCountForBlock(ctx sdk.Context) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixPointCountForBlock)
	bz := store.Get(types.KeyPrefixPointCountForBlock)
	if bz == nil {
		// This should never happen as this is set to 0 on genesis
		return 0, errors.New("current pool point count has not been set in state")
	}

	res := sdk.BigEndianToUint64(bz)

	return res, nil
}

// SetPointCountForBlock sets the number of pool points that have been consumed in the current block
func (k Keeper) SetPointCountForBlock(ctx sdk.Context, pointCount uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixPointCountForBlock)
	store.Set(types.KeyPrefixPointCountForBlock, sdk.Uint64ToBigEndian(pointCount))
}

// IncrementPointCountForBlock increments the number of pool points that have been consumed in the current block
func (k Keeper) IncrementPointCountForBlock(ctx sdk.Context, amount uint64) error {
	pointCount, err := k.GetPointCountForBlock(ctx)
	if err != nil {
		return err
	}

	k.SetPointCountForBlock(ctx, pointCount+amount)

	return nil
}

// GetLatestBlockHeight returns the latest block height that protorev was run on
func (k Keeper) GetLatestBlockHeight(ctx sdk.Context) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixLatestBlockHeight)
	bz := store.Get(types.KeyPrefixLatestBlockHeight)
	if bz == nil {
		// This should never happen as the module is initialized on genesis and reset in the post handler
		return 0, errors.New("block height has not been set in state")
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
func (k Keeper) GetAdminAccount(ctx sdk.Context) sdk.AccAddress {
	params := k.GetParams(ctx)
	return sdk.MustAccAddressFromBech32(params.Admin)
}

// SetAdminAccount sets the admin account for protorev
func (k Keeper) SetAdminAccount(ctx sdk.Context, adminAccount sdk.AccAddress) {
	params := k.GetParams(ctx)
	params.Admin = adminAccount.String()
	k.SetParams(ctx, params)
}

// GetDeveloperAccount returns the developer account for protorev
func (k Keeper) GetDeveloperAccount(ctx sdk.Context) (sdk.AccAddress, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDeveloperAccount)
	bz := store.Get(types.KeyPrefixDeveloperAccount)
	if bz == nil {
		return nil, errors.New("developer account not found, it has not been initialized by the admin account")
	}

	return sdk.AccAddress(bz), nil
}

// SetDeveloperAccount sets the developer account for protorev that will receive a portion of arbitrage profits
func (k Keeper) SetDeveloperAccount(ctx sdk.Context, developerAccount sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDeveloperAccount)
	store.Set(types.KeyPrefixDeveloperAccount, developerAccount.Bytes())
}

// GetMaxPointsPerTx returns the max number of pool points that can be consumed per transaction. A pool point is roughly
// equivalent to 1 ms of simulation & execution time.
func (k Keeper) GetMaxPointsPerTx(ctx sdk.Context) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixMaxPointsPerTx)
	bz := store.Get(types.KeyPrefixMaxPointsPerTx)
	if bz == nil {
		// This should never happen as it is set to the default value on genesis
		return 0, errors.New("max pool points per tx has not been set in state")
	}

	res := sdk.BigEndianToUint64(bz)
	return res, nil
}

// SetMaxPointsPerTx sets the max number of pool points that can be consumed per transaction. A pool point is roughly
// equivalent to 1 ms of simulation & execution time.
func (k Keeper) SetMaxPointsPerTx(ctx sdk.Context, maxPoints uint64) error {
	if maxPoints == 0 || maxPoints > types.MaxPoolPointsPerTx {
		return fmt.Errorf("max pool points must be between 1 and %d", types.MaxPoolPointsPerTx)
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixMaxPointsPerTx)
	bz := sdk.Uint64ToBigEndian(maxPoints)
	store.Set(types.KeyPrefixMaxPointsPerTx, bz)

	return nil
}

// GetMaxPointsPerBlock returns the max number of pool points that can be consumed per block. A pool point is roughly
// equivalent to 1 ms of simulation & execution time.
func (k Keeper) GetMaxPointsPerBlock(ctx sdk.Context) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixMaxPointsPerBlock)
	bz := store.Get(types.KeyPrefixMaxPointsPerBlock)
	if bz == nil {
		// This should never happen as it is set to the default value on genesis
		return 0, errors.New("max pool points per block has not been set in state")
	}

	res := sdk.BigEndianToUint64(bz)
	return res, nil
}

// SetMaxPointsPerBlock sets the max number of pool points that can be consumed per block. A pool point is roughly
// equivalent to 1 ms of simulation & execution time.
func (k Keeper) SetMaxPointsPerBlock(ctx sdk.Context, maxPoints uint64) error {
	if maxPoints == 0 || maxPoints > types.MaxPoolPointsPerBlock {
		return fmt.Errorf("max pool points per block must be between 1 and %d", types.MaxPoolPointsPerBlock)
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixMaxPointsPerBlock)
	bz := sdk.Uint64ToBigEndian(maxPoints)
	store.Set(types.KeyPrefixMaxPointsPerBlock, bz)

	return nil
}

// GetInfoByPoolType retrieves the metadata about the different pool types. This is used to determine the execution costs of
// different pool types when calculating the optimal route (in terms of time and gas consumption).
func (k Keeper) GetInfoByPoolType(ctx sdk.Context) types.InfoByPoolType {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixInfoByPoolType)
	poolWeights := &types.InfoByPoolType{}
	osmoutils.MustGet(store, types.KeyPrefixInfoByPoolType, poolWeights)
	return *poolWeights
}

// SetInfoByPoolType sets the pool type information.
func (k Keeper) SetInfoByPoolType(ctx sdk.Context, poolWeights types.InfoByPoolType) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixInfoByPoolType)
	osmoutils.MustSet(store, types.KeyPrefixInfoByPoolType, &poolWeights)
}

// GetAllProtocolRevenue returns all types of protocol revenue (txfees, taker fees, and cyclic arb profits), as well as the block height from which we started accounting
// for each of these revenue sources.
func (k Keeper) GetAllProtocolRevenue(ctx sdk.Context) types.AllProtocolRevenue {
	currentCyclicArb := k.GetAllProfits(ctx)
	currentCyclicArbCoins := osmoutils.ConvertCoinArrayToCoins(currentCyclicArb)

	cyclicArbTracker := types.CyclicArbTracker{
		CyclicArb:                  currentCyclicArbCoins.Sub(k.GetCyclicArbProfitTrackerValue(ctx)...),
		HeightAccountingStartsFrom: k.GetCyclicArbProfitTrackerStartHeight(ctx),
	}

	takerFeesTracker := poolmanagertypes.TakerFeesTracker{
		TakerFeesToStakers:         k.poolmanagerKeeper.GetTakerFeeTrackerForStakers(ctx),
		TakerFeesToCommunityPool:   k.poolmanagerKeeper.GetTakerFeeTrackerForCommunityPool(ctx),
		HeightAccountingStartsFrom: k.poolmanagerKeeper.GetTakerFeeTrackerStartHeight(ctx),
	}

	return types.AllProtocolRevenue{
		TakerFeesTracker: takerFeesTracker,
		CyclicArbTracker: cyclicArbTracker,
	}
}
