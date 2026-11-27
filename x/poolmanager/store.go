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
	cosmwasmutils "github.com/osmosis-labs/osmosis/osmoutils/cosmwasm"
	v3 "github.com/osmosis-labs/osmosis/v32/x/cosmwasmpool/cosmwasm/msg/v3"
	"github.com/osmosis-labs/osmosis/v32/x/poolmanager/types"
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

// setTakerFeeShareAgreementsMapCached replaces the taker fee share agreements cache with the contents of the store.
// If the store cannot be read, the cache is emptied rather than left as it was, so that it never depends on
// what a node happened to hold in memory.
func (k *Keeper) setTakerFeeShareAgreementsMapCached(ctx sdk.Context) error {
	takerFeeShareAgreement, err := k.getAllTakerFeeShareAgreementsMap(ctx)
	if err != nil {
		k.cachedTakerFeeShareAgreementMap = make(map[string]types.TakerFeeShareAgreement)
		return err
	}
	k.cachedTakerFeeShareAgreementMap = takerFeeShareAgreement
	return nil
}

// getTakerFeeShareAgreementFromDenom retrieves a specific taker fee share agreement from the cache.
// The cache holds the store contents as of the start of the current block, see BeginBlock.
func (k Keeper) getTakerFeeShareAgreementFromDenom(takerFeeShareDenom string) (types.TakerFeeShareAgreement, bool) {
	takerFeeShareAgreement, found := k.cachedTakerFeeShareAgreementMap[takerFeeShareDenom]
	return takerFeeShareAgreement, found
}

// GetTakerFeeShareAgreementFromDenomUNSAFE is used to expose an internal method to gRPC query. This method should not be used in other modules, since it reads the cache, which only reflects the store as of the start of the current block.
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
// It only writes the store. The cache is not touched here, because a write made on a context that is later
// discarded would stay in the cache. The cache picks the change up in the next BeginBlock.
func (k *Keeper) SetTakerFeeShareAgreementForDenom(ctx sdk.Context, takerFeeShare types.TakerFeeShareAgreement) error {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatTakerFeeShareAgreementKey(takerFeeShare.Denom)
	bz, err := proto.Marshal(&takerFeeShare)
	if err != nil {
		return err
	}

	store.Set(key, bz)

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
		// Query the liquidity through queryTotalPoolLiquidity rather than pool.GetPoolDenoms, which panics on failure
		// and caps gas at the node-local query gas limit.
		poolLiquidity, err := k.queryTotalPoolLiquidity(ctx, pool.GetAddress())
		if err != nil {
			return err
		}
		for _, coin := range poolLiquidity {
			if coin.Denom == takerFeeShare.Denom {
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
// It only writes the store. The cache picks the change up in the next BeginBlock.
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

	return nil
}

// getRegisteredAlloyedPoolFromDenom retrieves a specific registered alloyed pool from the cache via the alloyed denom.
// The cache holds the store contents as of the start of the current block, see BeginBlock.
func (k Keeper) getRegisteredAlloyedPoolFromDenom(alloyedDenom string) (types.AlloyContractTakerFeeShareState, bool) {
	registeredAlloyedPool, found := k.cachedRegisteredAlloyPoolByAlloyDenomMap[alloyedDenom]
	if !found {
		return types.AlloyContractTakerFeeShareState{}, false
	}
	return registeredAlloyedPool, true
}

// GetRegisteredAlloyedPoolFromDenomUNSAFE is used to expose an internal method to gRPC query. This method should not be used in other modules, since it reads the cache, which only reflects the store as of the start of the current block.
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
	bz := ctx.KVStore(k.storeKey).Get(types.FormatRegisteredAlloyPoolKey(poolId, alloyedDenom))
	if bz == nil {
		return types.AlloyContractTakerFeeShareState{}, types.NoRegisteredAlloyedPoolError{PoolId: poolId}
	}
	registeredAlloyedPool := types.AlloyContractTakerFeeShareState{}
	if err := proto.Unmarshal(bz, &registeredAlloyedPool); err != nil {
		return types.AlloyContractTakerFeeShareState{}, err
	}
	return registeredAlloyedPool, nil
}

// GetRegisteredAlloyedPoolFromPoolIdUNSAFE is used to expose an internal method to gRPC query. This method should not be used in other modules.
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

// setAllRegisteredAlloyedPoolsByDenomCached replaces the registered alloyed pools cache with the contents of the store.
// If the store cannot be read, the cache is emptied rather than left as it was.
func (k *Keeper) setAllRegisteredAlloyedPoolsByDenomCached(ctx sdk.Context) error {
	registeredAlloyPools, err := k.getAllRegisteredAlloyedPoolsByDenomMap(ctx)
	if err != nil {
		k.cachedRegisteredAlloyPoolByAlloyDenomMap = make(map[string]types.AlloyContractTakerFeeShareState)
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
	sort.SliceStable(registeredAlloyedPoolsIdArray, func(i, j int) bool { return registeredAlloyedPoolsIdArray[i] < registeredAlloyedPoolsIdArray[j] })

	return registeredAlloyedPoolsIdArray, nil
}

//
// Helpers
//

// fixedQueryGasLimitWasmKeeper caps smart queries at a constant gas limit. The wasm keeper's own
// QueryGasLimit is the node's query_gas_limit setting from app.toml, which can differ between validators.
// Where the gas meter is infinite (EndBlock, gov proposal execution) that setting is the whole cap, so using it
// would let nodes reach different results for the same query.
type fixedQueryGasLimitWasmKeeper struct {
	types.WasmKeeper
}

func (fixedQueryGasLimitWasmKeeper) QueryGasLimit() storetypes.Gas {
	return cosmwasmutils.DefaultContractCallGasLimit
}

// queryAlloyedPoolContract sends a smart query to an alloyed pool contract and unmarshals the response.
// The query runs on its own gas meter, capped at cosmwasmutils.DefaultContractCallGasLimit, and a panic in the
// contract call is returned as an error. These queries also run in EndBlock, where the gas meter is infinite and
// nothing recovers a panic.
func queryAlloyedPoolContract[K any](ctx sdk.Context, wasmKeeper types.WasmKeeper, contractAddr sdk.AccAddress, query string) (K, error) {
	return cosmwasmutils.Query[json.RawMessage, K](ctx, fixedQueryGasLimitWasmKeeper{wasmKeeper}, contractAddr.String(), json.RawMessage(query))
}

// queryAndCheckAlloyedDenom queries the smart contract for the alloyed denomination and validates its format.
// It sends a query to the contract address to get the share denomination, then checks if the denomination
// follows the expected format "factory/{contractAddr}/alloyed/{denom}".
// Returns the alloyed denomination if valid, otherwise returns an error.
func (k Keeper) queryAndCheckAlloyedDenom(ctx sdk.Context, contractAddr sdk.AccAddress) (string, error) {
	response, err := queryAlloyedPoolContract[v3.ShareDenomResponse](ctx, k.wasmKeeper, contractAddr, `{"get_share_denom": {}}`)
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

	return k.calculateTakerFeeShareAgreements(ctx, totalPoolLiquidity, normalizationFactors)
}

// queryTotalPoolLiquidity queries the smart contract for the total pool liquidity.
// It sends a query to the contract address and unmarshals the response into a slice of sdk.Coin.
// Returns the total pool liquidity if successful, otherwise returns an error.
func (k Keeper) queryTotalPoolLiquidity(ctx sdk.Context, contractAddr sdk.AccAddress) ([]sdk.Coin, error) {
	liquidityResponse, err := queryAlloyedPoolContract[v3.TotalPoolLiquidityResponse](ctx, k.wasmKeeper, contractAddr, `{"get_total_pool_liquidity": {}}`)
	if err != nil {
		return nil, err
	}

	return liquidityResponse.TotalPoolLiquidity, nil
}

// queryAssetConfigs queries the smart contract for the asset configurations.
// It sends a query to the contract address and unmarshals the response into a slice of v3.AssetConfig.
// Returns the asset configurations if successful, otherwise returns an error.
func (k Keeper) queryAssetConfigs(ctx sdk.Context, contractAddr sdk.AccAddress) ([]v3.AssetConfig, error) {
	assetConfigsResponse, err := queryAlloyedPoolContract[v3.ListAssetConfigsResponse](ctx, k.wasmKeeper, contractAddr, `{"list_asset_configs": {}}`)
	if err != nil {
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
// The share agreements are read from the store, not from the cache, so that an agreement set earlier in the
// same block is taken into account.
func (k Keeper) calculateTakerFeeShareAgreements(ctx sdk.Context, totalPoolLiquidity []sdk.Coin, normalizationFactors map[string]osmomath.Dec) ([]types.TakerFeeShareAgreement, error) {
	totalAlloyedLiquidity := types.ZeroDec
	var assetsWithShareAgreement []sdk.Coin
	var takerFeeShareAgreements []types.TakerFeeShareAgreement
	var skimAddresses []string
	var skimPercents []osmomath.Dec

	for _, coin := range totalPoolLiquidity {
		normalizationFactor := normalizationFactors[coin.Denom]
		normalizedAmount := coin.Amount.ToLegacyDec().Quo(normalizationFactor)
		totalAlloyedLiquidity = totalAlloyedLiquidity.Add(normalizedAmount)

		takerFeeShareAgreement, found := k.GetTakerFeeShareAgreementFromDenomNoCache(ctx, coin.Denom)
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
// It retrieves the registered alloyed pool, calculates the new taker fee share agreements, and updates the store with the new state.
// The cache picks the change up in the next BeginBlock.
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
