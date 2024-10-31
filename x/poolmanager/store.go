package poolmanager

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	storetypes "cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	v3 "github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/cosmwasm/msg/v3"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

//
// Taker Fee Share Agreements
//

// getAllTakerFeeShareAgreementsMap creates the map used for the taker fee share agreements cache.
func (k Keeper) getAllTakerFeeShareAgreementsMap(ctx sdk.Context) (map[string]types.TakerFeeShareAgreement, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyTakerFeeShare)
	defer iterator.Close()

	takerFeeShareAgreementsMap := make(map[string]types.TakerFeeShareAgreement)
	for ; iterator.Valid(); iterator.Next() {
		takerFeeShareAgreement := types.TakerFeeShareAgreement{}
		if err := proto.Unmarshal(iterator.Value(), &takerFeeShareAgreement); err != nil {
			return nil, err
		}
		takerFeeShareAgreementsMap[takerFeeShareAgreement.Denom] = takerFeeShareAgreement
	}

	return takerFeeShareAgreementsMap, nil
}

// GetAllTakerFeesShareAgreements creates a slice of all taker fee share agreements.
// Used in the AllTakerFeeShareAgreementsRequest gRPC query.
func (k Keeper) GetAllTakerFeesShareAgreements(ctx sdk.Context) ([]types.TakerFeeShareAgreement, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyTakerFeeShare)
	defer iterator.Close()

	takerFeeShareAgreements := []types.TakerFeeShareAgreement{}
	for ; iterator.Valid(); iterator.Next() {
		takerFeeShareAgreement := types.TakerFeeShareAgreement{}
		if err := proto.Unmarshal(iterator.Value(), &takerFeeShareAgreement); err != nil {
			return nil, err
		}
		takerFeeShareAgreements = append(takerFeeShareAgreements, takerFeeShareAgreement)
	}

	return takerFeeShareAgreements, nil
}

// setTakerFeeShareAgreementsMapCached is used for initializing the cache for the taker fee share agreements.
func (k *Keeper) setTakerFeeShareAgreementsMapCached(ctx sdk.Context) error {
	takerFeeShareAgreement, err := k.getAllTakerFeeShareAgreementsMap(ctx)
	if err != nil {
		return err
	}
	k.cachedTakerFeeShareAgreementMap = takerFeeShareAgreement
	return nil
}

// getTakerFeeShareAgreementFromDenom retrieves a specific taker fee share agreement from the store.
func (k Keeper) getTakerFeeShareAgreementFromDenom(takerFeeShareDenom string) (types.TakerFeeShareAgreement, bool) {
	takerFeeShareAgreement, found := k.cachedTakerFeeShareAgreementMap[takerFeeShareDenom]
	return takerFeeShareAgreement, found
}

// GetTakerFeeShareAgreementFromDenomUNSAFE is used to expose an internal method to gRPC query. This method should not be used in other modules, since the cache is not populated in those keepers.
// Used in the TakerFeeShareAgreementFromDenomRequest gRPC query.
func (k Keeper) GetTakerFeeShareAgreementFromDenomUNSAFE(takerFeeShareDenom string) (types.TakerFeeShareAgreement, bool) {
	return k.getTakerFeeShareAgreementFromDenom(takerFeeShareDenom)
}

// GetTakerFeeShareAgreementFromDenom retrieves a specific taker fee share agreement from the store, bypassing cache.
func (k Keeper) GetTakerFeeShareAgreementFromDenomNoCache(ctx sdk.Context, takerFeeShareDenom string) (types.TakerFeeShareAgreement, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatTakerFeeShareAgreementKey(takerFeeShareDenom)
	bz := store.Get(key)
	if bz == nil {
		return types.TakerFeeShareAgreement{}, false
	}

	var takerFeeShareAgreement types.TakerFeeShareAgreement
	if err := proto.Unmarshal(bz, &takerFeeShareAgreement); err != nil {
		return types.TakerFeeShareAgreement{}, false
	}

	return takerFeeShareAgreement, true
}

// SetTakerFeeShareAgreementForDenom is used for setting a specific taker fee share agreement in the store.
// Used in the MsgSetTakerFeeShareAgreementForDenom, by the gov address only.
func (k *Keeper) SetTakerFeeShareAgreementForDenom(ctx sdk.Context, takerFeeShare types.TakerFeeShareAgreement) error {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatTakerFeeShareAgreementKey(takerFeeShare.Denom)
	bz, err := proto.Marshal(&takerFeeShare)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	// Set cache value
	k.cachedTakerFeeShareAgreementMap[takerFeeShare.Denom] = takerFeeShare

	// Check if this denom is in the registered alloyed pools, if so we need to recalculate the taker fee share composition
	poolIds, err := k.getAllRegisteredAlloyedPoolsIdArray(ctx)
	if err != nil {
		return err
	}
	for _, poolId := range poolIds {
		pool, err := k.cosmwasmpoolKeeper.GetPool(ctx, poolId)
		if err != nil {
			return err
		}
		poolDenoms := pool.GetPoolDenoms(ctx)
		for _, denom := range poolDenoms {
			if denom == takerFeeShare.Denom {
				// takerFeeShare.Denom is one of the poolDenoms
				err := k.recalculateAndSetTakerFeeShareAlloyComposition(ctx, poolId)
				if err != nil {
					return err
				}
				break
			}
		}
	}

	return nil
}

//
// Taker Fee Share Accumulators
//

// GetTakerFeeShareDenomsToAccruedValue retrieves the accrued value for a specific taker fee share denomination and taker fee charged denomination from the store.
// Used in the TakerFeeShareDenomsToAccruedValueRequest gRPC query.
func (k Keeper) GetTakerFeeShareDenomsToAccruedValue(ctx sdk.Context, takerFeeShareDenom string, takerFeeChargedDenom string) (osmomath.Int, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyTakerFeeShareDenomAccrualForTakerFeeChargedDenom(takerFeeShareDenom, takerFeeChargedDenom)
	accruedValue := sdk.IntProto{}
	found, err := osmoutils.Get(store, key, &accruedValue)
	if err != nil {
		return osmomath.Int{}, err
	}
	if !found {
		return osmomath.Int{}, types.NoAccruedValueError{TakerFeeShareDenom: takerFeeShareDenom, TakerFeeChargedDenom: takerFeeChargedDenom}
	}
	return accruedValue.Int, nil
}

// SetTakerFeeShareDenomsToAccruedValue sets the accrued value for a specific taker fee share denomination and taker fee charged denomination in the store.
func (k Keeper) SetTakerFeeShareDenomsToAccruedValue(ctx sdk.Context, takerFeeShareDenom string, takerFeeChargedDenom string, accruedValue osmomath.Int) error {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyTakerFeeShareDenomAccrualForTakerFeeChargedDenom(takerFeeShareDenom, takerFeeChargedDenom)
	accruedValueProto := sdk.IntProto{Int: accruedValue}
	bz, err := proto.Marshal(&accruedValueProto)
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// increaseTakerFeeShareDenomsToAccruedValue increases (adds to, not replace) the accrued value for a specific taker fee share denomination and taker fee charged denomination in the store.
func (k Keeper) increaseTakerFeeShareDenomsToAccruedValue(ctx sdk.Context, takerFeeShareDenom string, takerFeeChargedDenom string, additiveValue osmomath.Int) error {
	accruedValueBefore, err := k.GetTakerFeeShareDenomsToAccruedValue(ctx, takerFeeShareDenom, takerFeeChargedDenom)
	if err != nil {
		if _, ok := err.(types.NoAccruedValueError); ok {
			accruedValueBefore = osmomath.ZeroInt()
		} else {
			return err
		}
	}

	accruedValueAfter := accruedValueBefore.Add(additiveValue)
	return k.SetTakerFeeShareDenomsToAccruedValue(ctx, takerFeeShareDenom, takerFeeChargedDenom, accruedValueAfter)
}

// GetAllTakerFeeShareAccumulators creates a slice of all taker fee share accumulators.
// Used in the AllTakerFeeShareAccumulatorsRequest gRPC query.
func (k Keeper) GetAllTakerFeeShareAccumulators(ctx sdk.Context) ([]types.TakerFeeSkimAccumulator, error) {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.TakerFeeSkimAccrualPrefix)
	defer iter.Close()

	takerFeeAgreementDenomToCoins := make(map[string]sdk.Coins)
	var denoms []string // Slice to keep track of the keys and ensure deterministic ordering

	for ; iter.Valid(); iter.Next() {
		accruedValue := sdk.IntProto{}
		if err := proto.Unmarshal(iter.Value(), &accruedValue); err != nil {
			return nil, err
		}
		keyParts := strings.Split(string(iter.Key()), types.KeySeparator)
		tierDenom := keyParts[1]
		takerFeeDenom := keyParts[2]
		accruedValueInt := accruedValue.Int
		currentCoins := takerFeeAgreementDenomToCoins[tierDenom]

		// Add the denom to the slice if it's not already there
		if _, exists := takerFeeAgreementDenomToCoins[tierDenom]; !exists {
			denoms = append(denoms, tierDenom)
		}

		takerFeeAgreementDenomToCoins[tierDenom] = currentCoins.Add(sdk.NewCoin(takerFeeDenom, accruedValueInt))
	}

	takerFeeSkimAccumulators := []types.TakerFeeSkimAccumulator{}
	for _, denom := range denoms {
		takerFeeSkimAccumulators = append(takerFeeSkimAccumulators, types.TakerFeeSkimAccumulator{
			Denom:            denom,
			SkimmedTakerFees: takerFeeAgreementDenomToCoins[denom],
		})
	}

	return takerFeeSkimAccumulators, nil
}

// DeleteAllTakerFeeShareAccumulatorsForTakerFeeShareDenom clears the TakerFeeShareAccumulator records for a specific taker fee share denom.
// Is specifically used after the distributions have been completed after epoch for each denom.
func (k Keeper) DeleteAllTakerFeeShareAccumulatorsForTakerFeeShareDenom(ctx sdk.Context, takerFeeShareDenom string) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyTakerFeeShareDenomAccrualForAllDenoms(takerFeeShareDenom))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}

//
// Registered Alloyed Pool States
//

// setRegisteredAlloyedPool sets a specific registered alloyed pool in the store.
// Used in the MsgRegisterAlloyedPool, by the gov address only.
func (k *Keeper) setRegisteredAlloyedPool(ctx sdk.Context, poolId uint64) error {
	store := ctx.KVStore(k.storeKey)

	cwPool, err := k.GetPool(ctx, poolId)
	if err != nil {
		return err
	}

	if cwPool.GetType() != types.CosmWasm {
		return types.NotCosmWasmPoolError{PoolId: poolId}
	}

	contractAddr := cwPool.GetAddress()

	alloyedDenom, err := k.queryAndCheckAlloyedDenom(ctx, contractAddr)
	if err != nil {
		return err
	}

	takerFeeShareAgreements, err := k.snapshotTakerFeeShareAlloyComposition(ctx, contractAddr)
	if err != nil {
		return err
	}

	registeredAlloyedPool := types.AlloyContractTakerFeeShareState{
		ContractAddress:         contractAddr.String(),
		TakerFeeShareAgreements: takerFeeShareAgreements,
	}

	bz, err := proto.Marshal(&registeredAlloyedPool)
	if err != nil {
		return err
	}

	// Just to be safe, if a pool is registered with the same ID already but different alloyed denom, remove it.
	iterator := storetypes.KVStorePrefixIterator(store, types.FormatRegisteredAlloyPoolKeyPoolIdOnly(poolId))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}

	key := types.FormatRegisteredAlloyPoolKey(poolId, alloyedDenom)
	store.Set(key, bz)

	// Set cache value
	k.cachedRegisteredAlloyPoolByAlloyDenomMap[alloyedDenom] = registeredAlloyedPool

	return nil
}

// getRegisteredAlloyedPoolFromDenom retrieves a specific registered alloyed pool from the store via the alloyed denom.
func (k Keeper) getRegisteredAlloyedPoolFromDenom(alloyedDenom string) (types.AlloyContractTakerFeeShareState, bool) {
	registeredAlloyedPool, found := k.cachedRegisteredAlloyPoolByAlloyDenomMap[alloyedDenom]
	if !found {
		return types.AlloyContractTakerFeeShareState{}, false
	}
	return registeredAlloyedPool, true
}

// GetRegisteredAlloyedPoolFromDenomUNSAFE is used to expose an internal method to gRPC query. This method should not be used in other modules, since the cache is not populated in those keepers.
// Used in the RegisteredAlloyedPoolFromDenomRequest gRPC query.
func (k Keeper) GetRegisteredAlloyedPoolFromDenomUNSAFE(alloyedDenom string) (types.AlloyContractTakerFeeShareState, bool) {
	return k.getRegisteredAlloyedPoolFromDenom(alloyedDenom)
}

// getRegisteredAlloyedPoolFromPoolId retrieves a specific registered alloyed pool from the store via the pool id.
func (k Keeper) getRegisteredAlloyedPoolFromPoolId(ctx sdk.Context, poolId uint64) (types.AlloyContractTakerFeeShareState, error) {
	alloyedDenom, err := k.getAlloyedDenomFromPoolId(ctx, poolId)
	if err != nil {
		return types.AlloyContractTakerFeeShareState{}, err
	}
	registeredAlloyedPool, found := k.getRegisteredAlloyedPoolFromDenom(alloyedDenom)
	if !found {
		return types.AlloyContractTakerFeeShareState{}, types.NoRegisteredAlloyedPoolError{PoolId: poolId}
	}
	return registeredAlloyedPool, nil
}

// GetRegisteredAlloyedPoolFromPoolIdUNSAFE is used to expose an internal method to gRPC query. This method should not be used in other modules, since the cache is not populated in those keepers.
// Used in the RegisteredAlloyedPoolFromPoolIdRequest gRPC query.
func (k Keeper) GetRegisteredAlloyedPoolFromPoolIdUNSAFE(ctx sdk.Context, poolId uint64) (types.AlloyContractTakerFeeShareState, error) {
	return k.getRegisteredAlloyedPoolFromPoolId(ctx, poolId)
}

// GetAllRegisteredAlloyedPools creates a slice of all registered alloyed pools.
// Used in the AllRegisteredAlloyedPoolsRequest gRPC query.
func (k Keeper) GetAllRegisteredAlloyedPools(ctx sdk.Context) ([]types.AlloyContractTakerFeeShareState, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyRegisteredAlloyPool)
	defer iterator.Close()

	var registeredAlloyedPools []types.AlloyContractTakerFeeShareState
	for ; iterator.Valid(); iterator.Next() {
		registeredAlloyedPool := types.AlloyContractTakerFeeShareState{}
		err := proto.Unmarshal(iterator.Value(), &registeredAlloyedPool)
		if err != nil {
			return nil, err
		}

		registeredAlloyedPools = append(registeredAlloyedPools, registeredAlloyedPool)
	}

	return registeredAlloyedPools, nil
}

// GetAllRegisteredAlloyedPoolsByDenomMap creates the map used for the registered alloyed pools cache.
func (k Keeper) getAllRegisteredAlloyedPoolsByDenomMap(ctx sdk.Context) (map[string]types.AlloyContractTakerFeeShareState, error) {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.KeyRegisteredAlloyPool)
	defer iter.Close()

	registeredAlloyedPoolsMap := make(map[string]types.AlloyContractTakerFeeShareState)
	for ; iter.Valid(); iter.Next() {
		registeredAlloyedPool := types.AlloyContractTakerFeeShareState{}
		if err := proto.Unmarshal(iter.Value(), &registeredAlloyedPool); err != nil {
			return nil, err
		}

		key := string(iter.Key())
		parts := strings.Split(key, types.KeySeparator)
		if len(parts) < 3 {
			return nil, types.ErrInvalidKeyFormat
		}
		alloyedDenom := parts[len(parts)-1]
		registeredAlloyedPoolsMap[alloyedDenom] = registeredAlloyedPool
	}

	return registeredAlloyedPoolsMap, nil
}

// setAllRegisteredAlloyedPoolsByDenomCached initializes the cache for the registered alloyed pools.
func (k *Keeper) setAllRegisteredAlloyedPoolsByDenomCached(ctx sdk.Context) error {
	registeredAlloyPools, err := k.getAllRegisteredAlloyedPoolsByDenomMap(ctx)
	if err != nil {
		return err
	}
	k.cachedRegisteredAlloyPoolByAlloyDenomMap = registeredAlloyPools
	return nil
}

//
// Registered Alloyed Pool Ids
//

// getAllRegisteredAlloyedPoolsIdArray creates an array of all registered alloyed pools IDs.
func (k Keeper) getAllRegisteredAlloyedPoolsIdArray(ctx sdk.Context) ([]uint64, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyRegisteredAlloyPool)
	defer iterator.Close()

	registeredAlloyedPoolsIdArray := []uint64{}
	for ; iterator.Valid(); iterator.Next() {
		key := string(iterator.Key())
		parts := strings.Split(key, types.KeySeparator)
		if len(parts) < 3 {
			return nil, types.ErrInvalidKeyFormat
		}
		alloyedIdStr := parts[len(parts)-2]
		// Convert the string to uint64
		alloyedId, err := strconv.ParseUint(alloyedIdStr, 10, 64)
		if err != nil {
			return nil, types.InvalidAlloyedPoolIDError{AlloyedIDStr: alloyedIdStr, Err: err}
		}
		registeredAlloyedPoolsIdArray = append(registeredAlloyedPoolsIdArray, alloyedId)
	}
	sort.Slice(registeredAlloyedPoolsIdArray, func(i, j int) bool { return registeredAlloyedPoolsIdArray[i] < registeredAlloyedPoolsIdArray[j] })

	return registeredAlloyedPoolsIdArray, nil
}

//
// Helpers
//

// queryAndCheckAlloyedDenom queries the smart contract for the alloyed denomination and validates its format.
// It sends a query to the contract address to get the share denomination, then checks if the denomination
// follows the expected format "factory/{contractAddr}/alloyed/{denom}".
// Returns the alloyed denomination if valid, otherwise returns an error.
func (k Keeper) queryAndCheckAlloyedDenom(ctx sdk.Context, contractAddr sdk.AccAddress) (string, error) {
	queryBz := []byte(`{"get_share_denom": {}}`)
	respBz, err := k.wasmKeeper.QuerySmart(ctx, contractAddr, queryBz)
	if err != nil {
		return "", err
	}

	var response v3.ShareDenomResponse
	err = json.Unmarshal(respBz, &response)
	if err != nil {
		return "", err
	}
	alloyedDenom := response.ShareDenom

	parts := strings.Split(alloyedDenom, "/")
	if len(parts) != 4 {
		return "", types.InvalidAlloyedDenomFormatError{PartsLength: len(parts)}
	}

	if parts[0] != "factory" {
		return "", types.InvalidAlloyedDenomPartError{PartIndex: 0, Expected: "factory", Actual: parts[0]}
	}

	if parts[1] != contractAddr.String() {
		return "", types.InvalidAlloyedDenomPartError{PartIndex: 1, Expected: contractAddr.String(), Actual: parts[1]}
	}

	if parts[2] != "alloyed" {
		return "", types.InvalidAlloyedDenomPartError{PartIndex: 2, Expected: "alloyed", Actual: parts[2]}
	}

	return alloyedDenom, nil
}

// snapshotTakerFeeShareAlloyComposition queries the smart contract for the total pool liquidity and calculates
// the taker fee share agreements based on the liquidity of each asset in the pool. It returns a slice of
// TakerFeeShareAgreement objects, each containing the denomination, skim percent, and skim address.
// If the total alloyed liquidity is zero, it returns an error.
func (k Keeper) snapshotTakerFeeShareAlloyComposition(ctx sdk.Context, contractAddr sdk.AccAddress) ([]types.TakerFeeShareAgreement, error) {
	totalPoolLiquidity, err := k.queryTotalPoolLiquidity(ctx, contractAddr)
	if err != nil {
		return nil, err
	}

	assetConfigs, err := k.queryAssetConfigs(ctx, contractAddr)
	if err != nil {
		return nil, err
	}

	normalizationFactors, err := k.createNormalizationFactorsMap(assetConfigs)
	if err != nil {
		return nil, err
	}

	return k.calculateTakerFeeShareAgreements(totalPoolLiquidity, normalizationFactors)
}

// queryTotalPoolLiquidity queries the smart contract for the total pool liquidity.
// It sends a query to the contract address and unmarshals the response into a slice of sdk.Coin.
// Returns the total pool liquidity if successful, otherwise returns an error.
func (k Keeper) queryTotalPoolLiquidity(ctx sdk.Context, contractAddr sdk.AccAddress) ([]sdk.Coin, error) {
	queryBz := []byte(`{"get_total_pool_liquidity": {}}`)
	respBz, err := k.wasmKeeper.QuerySmart(ctx, contractAddr, queryBz)
	if err != nil {
		return nil, err
	}

	var liquidityResponse v3.TotalPoolLiquidityResponse
	if err := json.Unmarshal(respBz, &liquidityResponse); err != nil {
		return nil, err
	}

	return liquidityResponse.TotalPoolLiquidity, nil
}

// queryAssetConfigs queries the smart contract for the asset configurations.
// It sends a query to the contract address and unmarshals the response into a slice of v3.AssetConfig.
// Returns the asset configurations if successful, otherwise returns an error.
func (k Keeper) queryAssetConfigs(ctx sdk.Context, contractAddr sdk.AccAddress) ([]v3.AssetConfig, error) {
	queryBz := []byte(`{"list_asset_configs": {}}`)
	respBz, err := k.wasmKeeper.QuerySmart(ctx, contractAddr, queryBz)
	if err != nil {
		return nil, err
	}

	var assetConfigsResponse v3.ListAssetConfigsResponse
	if err := json.Unmarshal(respBz, &assetConfigsResponse); err != nil {
		return nil, err
	}

	return assetConfigsResponse.AssetConfigs, nil
}

// createNormalizationFactorsMap creates a map of normalization factors from the given asset configurations.
// It iterates through the asset configurations and converts the normalization factor string to osmomath.Dec.
// Returns the normalization factors map if successful, otherwise returns an error.
func (k Keeper) createNormalizationFactorsMap(assetConfigs []v3.AssetConfig) (map[string]osmomath.Dec, error) {
	normalizationFactors := make(map[string]osmomath.Dec)
	for _, config := range assetConfigs {
		factor, err := osmomath.NewDecFromStr(config.NormalizationFactor)
		if err != nil {
			return nil, err
		}
		normalizationFactors[config.Denom] = factor
	}
	return normalizationFactors, nil
}

// calculateTakerFeeShareAgreements calculates the taker fee share agreements based on the total pool liquidity
// and normalization factors. It iterates through the pool liquidity, normalizes the amounts, and calculates
// the scaled skim percentages for each asset with a share agreement. Returns a slice of TakerFeeShareAgreement
// objects if successful, otherwise returns an error.
func (k Keeper) calculateTakerFeeShareAgreements(totalPoolLiquidity []sdk.Coin, normalizationFactors map[string]osmomath.Dec) ([]types.TakerFeeShareAgreement, error) {
	totalAlloyedLiquidity := types.ZeroDec
	var assetsWithShareAgreement []sdk.Coin
	var takerFeeShareAgreements []types.TakerFeeShareAgreement
	var skimAddresses []string
	var skimPercents []osmomath.Dec

	for _, coin := range totalPoolLiquidity {
		normalizationFactor := normalizationFactors[coin.Denom]
		normalizedAmount := coin.Amount.ToLegacyDec().Quo(normalizationFactor)
		totalAlloyedLiquidity = totalAlloyedLiquidity.Add(normalizedAmount)

		takerFeeShareAgreement, found := k.getTakerFeeShareAgreementFromDenom(coin.Denom)
		if !found {
			continue
		}
		assetsWithShareAgreement = append(assetsWithShareAgreement, coin)
		skimAddresses = append(skimAddresses, takerFeeShareAgreement.SkimAddress)
		skimPercents = append(skimPercents, takerFeeShareAgreement.SkimPercent)
	}

	if totalAlloyedLiquidity.IsZero() {
		return nil, types.ErrTotalAlloyedLiquidityIsZero
	}

	for i, coin := range assetsWithShareAgreement {
		normalizationFactor := normalizationFactors[coin.Denom]
		normalizedAmount := coin.Amount.ToLegacyDec().Quo(normalizationFactor)
		scaledSkim := normalizedAmount.Quo(totalAlloyedLiquidity).Mul(skimPercents[i])
		takerFeeShareAgreements = append(takerFeeShareAgreements, types.TakerFeeShareAgreement{
			Denom:       coin.Denom,
			SkimPercent: scaledSkim,
			SkimAddress: skimAddresses[i],
		})
	}

	return takerFeeShareAgreements, nil
}

// recalculateAndSetTakerFeeShareAlloyComposition recalculates the taker fee share composition for a given pool.
// It retrieves the registered alloyed pool, calculates the new taker fee share agreements, and updates the store and cache with the new state.
func (k *Keeper) recalculateAndSetTakerFeeShareAlloyComposition(ctx sdk.Context, poolId uint64) error {
	registeredAlloyedPoolPrior, err := k.getRegisteredAlloyedPoolFromPoolId(ctx, poolId)
	if err != nil {
		return err
	}

	takerFeeShareAlloyDenoms, err := k.snapshotTakerFeeShareAlloyComposition(ctx, sdk.MustAccAddressFromBech32(registeredAlloyedPoolPrior.ContractAddress))
	if err != nil {
		return err
	}

	registeredAlloyedPool := types.AlloyContractTakerFeeShareState{
		ContractAddress:         registeredAlloyedPoolPrior.ContractAddress,
		TakerFeeShareAgreements: takerFeeShareAlloyDenoms,
	}

	bz, err := proto.Marshal(&registeredAlloyedPool)
	if err != nil {
		return err
	}

	alloyedDenom, err := k.getAlloyedDenomFromPoolId(ctx, poolId)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	key := types.FormatRegisteredAlloyPoolKey(poolId, alloyedDenom)
	store.Set(key, bz)

	// Set cache value
	k.cachedRegisteredAlloyPoolByAlloyDenomMap[alloyedDenom] = registeredAlloyedPool

	return nil
}

// getAlloyedDenomFromPoolId retrieves the alloyed denomination associated with a given pool ID from the store.
// It iterates through the registered alloyed pools and matches the pool ID to find the corresponding alloyed denomination.
// Returns the alloyed denomination if found, otherwise returns an error.
func (k Keeper) getAlloyedDenomFromPoolId(ctx sdk.Context, poolId uint64) (string, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyRegisteredAlloyPool)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := string(iterator.Key())
		parts := strings.Split(key, types.KeySeparator)
		if len(parts) < 3 {
			return "", types.ErrInvalidKeyFormat
		}
		alloyedIdStr := parts[len(parts)-2]
		// Convert the string to uint64
		alloyedId, err := strconv.ParseUint(alloyedIdStr, 10, 64)
		if err != nil {
			return "", types.InvalidAlloyedPoolIDError{AlloyedIDStr: alloyedIdStr, Err: err}
		}
		if alloyedId == poolId {
			alloyedDenom := parts[len(parts)-1]
			return alloyedDenom, nil
		}
	}
	return "", types.NoRegisteredAlloyedPoolError{PoolId: poolId}
}
