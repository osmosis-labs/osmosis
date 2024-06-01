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
	alloyedpooltypes "github.com/osmosis-labs/osmosis/v25/x/cosmwasmpool/cosmwasm/msg/v3"
	"github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

//
// Taker Fee Share Agreements
//

// GetAllTakerFeeShareAgreementsMap creates the map used for the taker fee share agreements cache.
func (k Keeper) GetAllTakerFeeShareAgreementsMap(ctx sdk.Context) (map[string]types.TakerFeeShareAgreement, error) {
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

// SetTakerFeeShareAgreementsMapCached is used for initializing the cache for the taker fee share agreements.
func (k *Keeper) SetTakerFeeShareAgreementsMapCached(ctx sdk.Context) error {
	takerFeeShareAgreement, err := k.GetAllTakerFeeShareAgreementsMap(ctx)
	if err != nil {
		return err
	}
	k.cachedTakerFeeShareAgreementMap = takerFeeShareAgreement
	return nil
}

// GetTakerFeeShareAgreementFromDenom retrieves a specific taker fee share agreement from the store.
// Used in the TakerFeeShareAgreementFromDenomRequest gRPC query.
func (k Keeper) GetTakerFeeShareAgreementFromDenom(ctx sdk.Context, takerFeeShareDenom string) (types.TakerFeeShareAgreement, bool) {
	takerFeeShareAgreement, found := k.cachedTakerFeeShareAgreementMap[takerFeeShareDenom]
	if !found {
		return types.TakerFeeShareAgreement{}, false
	}
	return takerFeeShareAgreement, true
}

// SetTakerFeeShareAgreementForDenom is used for setting a specific take fee share agreement in the store.
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
	for _, poolId := range k.cachedRegisteredAlloyedPoolIdArray {
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

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgSetTakerFeeShareAgreementForDenomPair,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyTakerFeeShareDenom, takerFeeShare.Denom),
			sdk.NewAttribute(types.AttributeKeyTakerFeeShareSkimPercent, takerFeeShare.SkimPercent.String()),
			sdk.NewAttribute(types.AttributeKeyTakerFeeShareSkimAddress, takerFeeShare.SkimAddress),
		),
	})

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

// IncreaseTakerFeeShareDenomsToAccruedValue increases (adds to, not replace) the accrued value for a specific taker fee share denomination and taker fee charged denomination in the store.
func (k Keeper) IncreaseTakerFeeShareDenomsToAccruedValue(ctx sdk.Context, takerFeeShareDenom string, takerFeeChargedDenom string, additiveValue osmomath.Int) error {
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

// SetRegisteredAlloyedPool sets a specific registered alloyed pool in the store.
// Used in the MsgRegisterAlloyedPool, by the gov address only.
func (k *Keeper) SetRegisteredAlloyedPool(ctx sdk.Context, poolId uint64) error {
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

	// Set cache values
	k.cachedRegisteredAlloyPoolByAlloyDenomMap[alloyedDenom] = registeredAlloyedPool
	k.cachedRegisteredAlloyedPoolIdArray = append(k.cachedRegisteredAlloyedPoolIdArray, poolId)
	sort.Slice(k.cachedRegisteredAlloyedPoolIdArray, func(i, j int) bool {
		return k.cachedRegisteredAlloyedPoolIdArray[i] < k.cachedRegisteredAlloyedPoolIdArray[j]
	})

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgSetRegisteredAlloyedPool,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
		),
	})

	return nil
}

// GetRegisteredAlloyedPoolFromDenom retrieves a specific registered alloyed pool from the store via the alloyed denom.
// Used in the RegisteredAlloyedPoolFromDenomRequest gRPC query.
func (k Keeper) GetRegisteredAlloyedPoolFromDenom(ctx sdk.Context, alloyedDenom string) (types.AlloyContractTakerFeeShareState, bool) {
	registeredAlloyedPool, found := k.cachedRegisteredAlloyPoolByAlloyDenomMap[alloyedDenom]
	if !found {
		return types.AlloyContractTakerFeeShareState{}, false
	}
	return registeredAlloyedPool, true
}

// GetRegisteredAlloyedPoolFromPoolId retrieves a specific registered alloyed pool from the store via the pool id.
// Used in the RegisteredAlloyedPoolFromPoolIdRequest gRPC query.
func (k Keeper) GetRegisteredAlloyedPoolFromPoolId(ctx sdk.Context, poolId uint64) (types.AlloyContractTakerFeeShareState, error) {
	alloyedDenom, err := k.getAlloyedDenomFromPoolId(ctx, poolId)
	if err != nil {
		return types.AlloyContractTakerFeeShareState{}, err
	}
	registeredAlloyedPool, found := k.GetRegisteredAlloyedPoolFromDenom(ctx, alloyedDenom)
	if !found {
		return types.AlloyContractTakerFeeShareState{}, types.NoRegisteredAlloyedPoolError{PoolId: poolId}
	}
	return registeredAlloyedPool, nil
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
func (k Keeper) GetAllRegisteredAlloyedPoolsByDenomMap(ctx sdk.Context) (map[string]types.AlloyContractTakerFeeShareState, error) {
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

// SetAllRegisteredAlloyedPoolsByDenomCached initializes the cache for the registered alloyed pools.
func (k *Keeper) SetAllRegisteredAlloyedPoolsByDenomCached(ctx sdk.Context) error {
	registeredAlloyPools, err := k.GetAllRegisteredAlloyedPoolsByDenomMap(ctx)
	if err != nil {
		return err
	}
	k.cachedRegisteredAlloyPoolByAlloyDenomMap = registeredAlloyPools
	return nil
}

//
// Registered Alloyed Pool Ids
//

// GetAllRegisteredAlloyedPoolsIdArray creates the array used for the registered alloyed pools id array cache.
func (k Keeper) GetAllRegisteredAlloyedPoolsIdArray(ctx sdk.Context) ([]uint64, error) {
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

// SetAllRegisteredAlloyedPoolIdArrayCached initializes the cache for the registered alloyed pools id array.
func (k *Keeper) SetAllRegisteredAlloyedPoolIdArrayCached(ctx sdk.Context) error {
	registeredAlloyPoolIds, err := k.GetAllRegisteredAlloyedPoolsIdArray(ctx)
	if err != nil {
		return err
	}
	k.cachedRegisteredAlloyedPoolIdArray = registeredAlloyPoolIds
	return nil
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

	var response alloyedpooltypes.ShareDenomResponse
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
	// Query for total pool liquidity
	queryBz := []byte(`{"get_total_pool_liquidity": {}}`)
	respBz, err := k.wasmKeeper.QuerySmart(ctx, contractAddr, queryBz)
	if err != nil {
		return []types.TakerFeeShareAgreement{}, err
	}

	var liquidityResponse alloyedpooltypes.TotalPoolLiquidityResponse
	err = json.Unmarshal(respBz, &liquidityResponse)
	if err != nil {
		return []types.TakerFeeShareAgreement{}, err
	}
	totalPoolLiquidity := liquidityResponse.TotalPoolLiquidity

	// Query for asset configs
	queryBz = []byte(`{"list_asset_configs": {}}`)
	respBz, err = k.wasmKeeper.QuerySmart(ctx, contractAddr, queryBz)
	if err != nil {
		return []types.TakerFeeShareAgreement{}, err
	}

	var assetConfigsResponse alloyedpooltypes.ListAssetConfigsResponse
	err = json.Unmarshal(respBz, &assetConfigsResponse)
	if err != nil {
		return []types.TakerFeeShareAgreement{}, err
	}
	assetConfigs := assetConfigsResponse.AssetConfigs

	// Create a map for quick lookup of normalization factors
	normalizationFactors := make(map[string]osmomath.Dec)
	for _, config := range assetConfigs {
		factor, err := osmomath.NewDecFromStr(config.NormalizationFactor)
		if err != nil {
			return []types.TakerFeeShareAgreement{}, err
		}
		normalizationFactors[config.Denom] = factor
	}

	totalAlloyedLiquidity := osmomath.ZeroDec()
	var assetsWithShareAgreement []sdk.Coin
	var takerFeeShareAgreements []types.TakerFeeShareAgreement
	var skimAddresses []string
	var skimPercents []osmomath.Dec

	for _, coin := range totalPoolLiquidity {
		normalizationFactor := normalizationFactors[coin.Denom]
		normalizedAmount := coin.Amount.ToLegacyDec().Mul(normalizationFactor)
		totalAlloyedLiquidity = totalAlloyedLiquidity.Add(normalizedAmount)

		takerFeeShareAgreement, found := k.GetTakerFeeShareAgreementFromDenom(ctx, coin.Denom)
		if !found {
			continue
		}
		assetsWithShareAgreement = append(assetsWithShareAgreement, coin)
		skimAddresses = append(skimAddresses, takerFeeShareAgreement.SkimAddress)
		skimPercents = append(skimPercents, takerFeeShareAgreement.SkimPercent)
	}

	if totalAlloyedLiquidity.IsZero() {
		return []types.TakerFeeShareAgreement{}, types.ErrTotalAlloyedLiquidityIsZero
	}

	for i, coin := range assetsWithShareAgreement {
		normalizationFactor := normalizationFactors[coin.Denom]
		normalizedAmount := coin.Amount.ToLegacyDec().Mul(normalizationFactor)
		scaledSkim := normalizedAmount.Quo(totalAlloyedLiquidity).Mul(skimPercents[i])
		takerFeeShareAgreements = append(takerFeeShareAgreements, types.TakerFeeShareAgreement{
			Denom:       coin.Denom,
			SkimPercent: scaledSkim,
			SkimAddress: skimAddresses[i],
		})
	}

	return takerFeeShareAgreements, nil
}

// recalculateAndSetTakerFeeShareAlloyComposition recalculates the taker fee share composition for a given pool
// and updates the store and cache with the new values. It retrieves the registered alloyed pool, calculates
// the new taker fee share agreements, and updates the store and cache with the new state.
func (k *Keeper) recalculateAndSetTakerFeeShareAlloyComposition(ctx sdk.Context, poolId uint64) error {
	registeredAlloyedPoolPrior, err := k.GetRegisteredAlloyedPoolFromPoolId(ctx, poolId)
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