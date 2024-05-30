package poolmanager

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	storetypes "cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

type ShareDenomResponse struct {
	ShareDenom string `json:"share_denom"`
}

type TotalPoolLiquidityResponse struct {
	TotalPoolLiquidity []sdk.Coin `json:"total_pool_liquidity"`
}

//
// Taker Fee Share Agreements
//

// Used for creating the map used for the take fee share agreements cache.
func (k Keeper) GetAllTakerFeeShareAgreementsMap(ctx sdk.Context) (map[string]types.TakerFeeShareAgreement, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyTakerFeeShare)
	defer iterator.Close()

	takerFeeShareAgreementsMap := make(map[string]types.TakerFeeShareAgreement)
	for ; iterator.Valid(); iterator.Next() {
		takerFeeShareAgreement := types.TakerFeeShareAgreement{}
		osmoutils.MustGet(store, iterator.Key(), &takerFeeShareAgreement)
		takerFeeShareAgreementsMap[takerFeeShareAgreement.Denom] = takerFeeShareAgreement
	}

	return takerFeeShareAgreementsMap, nil
}

// Used in the AllTakerFeeShareAgreementsRequest gRPC query.
func (k Keeper) GetAllTakerFeesShareAgreements(ctx sdk.Context) []types.TakerFeeShareAgreement {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyTakerFeeShare)
	defer iterator.Close()

	takerFeeShareAgreements := []types.TakerFeeShareAgreement{}
	for ; iterator.Valid(); iterator.Next() {
		takerFeeShareAgreement := types.TakerFeeShareAgreement{}
		osmoutils.MustGet(store, iterator.Key(), &takerFeeShareAgreement)
		takerFeeShareAgreements = append(takerFeeShareAgreements, takerFeeShareAgreement)
	}

	return takerFeeShareAgreements
}

// Used for initializing the cache for the take fee share agreements.
func (k *Keeper) SetTakerFeeShareAgreementsMapCached(ctx sdk.Context) error {
	takerFeeShareAgreement, err := k.GetAllTakerFeeShareAgreementsMap(ctx)
	if err != nil {
		return err
	}
	k.cachedTakerFeeShareAgreementMap = takerFeeShareAgreement
	return nil
}

// Used in the TakerFeeShareAgreementFromDenomRequest gRPC query.
func (k Keeper) GetTakerFeeShareAgreementFromDenom(ctx sdk.Context, tierDenom string) (types.TakerFeeShareAgreement, bool) {
	takerFeeShareAgreement, found := k.cachedTakerFeeShareAgreementMap[tierDenom]
	if !found {
		return types.TakerFeeShareAgreement{}, false
	}
	return takerFeeShareAgreement, true
}

// Used for setting a specific take fee share agreement in the store.
// Used in the MsgSetTakerFeeShareAgreementForDenom, for governance.
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

// Used in the TakerFeeShareDenomsToAccruedValueRequest gRPC query.
func (k Keeper) GetTakerFeeShareDenomsToAccruedValue(ctx sdk.Context, tierDenom string, takerFeeDenom string) (osmomath.Int, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyTakerFeeShareTier1DenomAccrualForSingleDenom(tierDenom, takerFeeDenom)
	accruedValue := sdk.IntProto{}
	found, err := osmoutils.Get(store, key, &accruedValue)
	if err != nil {
		return osmomath.Int{}, err
	}
	if !found {
		return osmomath.Int{}, fmt.Errorf("no accrued value found for tierDenom %v and takerFeeDenom %s", tierDenom, takerFeeDenom)
	}
	return accruedValue.Int, nil
}

// Used for setting the accrued value for a specific tier denomination and taker fee denomination in the store.
func (k Keeper) SetTakerFeeShareDenomsToAccruedValue(ctx sdk.Context, tierDenom string, takerFeeDenom string, accruedValue osmomath.Int) error {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyTakerFeeShareTier1DenomAccrualForSingleDenom(tierDenom, takerFeeDenom)
	accruedValueProto := sdk.IntProto{Int: accruedValue}
	bz, err := proto.Marshal(&accruedValueProto)
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// Used for increasing the accrued value for a specific tier denomination and taker fee denomination in the store.
func (k Keeper) IncreaseTakerFeeShareDenomsToAccruedValue(ctx sdk.Context, tierDenom string, takerFeeDenom string, additiveValue osmomath.Int) error {
	accruedValueBefore, err := k.GetTakerFeeShareDenomsToAccruedValue(ctx, tierDenom, takerFeeDenom)
	if err != nil {
		if err.Error() == fmt.Errorf("no accrued value found for tierDenom %v and takerFeeDenom %s", tierDenom, takerFeeDenom).Error() {
			accruedValueBefore = osmomath.ZeroInt()
		} else {
			return err
		}
	}

	accruedValueAfter := accruedValueBefore.Add(additiveValue)
	return k.SetTakerFeeShareDenomsToAccruedValue(ctx, tierDenom, takerFeeDenom, accruedValueAfter)
}

// Used in the AllTakerFeeShareAccumulatorsRequest gRPC query.
func (k Keeper) GetAllTakerFeeShareAccumulators(ctx sdk.Context) []types.TakerFeeSkimAccumulator {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.TakerFeeSkimAccrualPrefix)
	defer iterator.Close()

	takerFeeAgreementDenomToCoins := make(map[string]sdk.Coins)
	var denoms []string // Slice to keep track of the keys and ensure deterministic ordering

	for ; iterator.Valid(); iterator.Next() {
		accruedValue := sdk.IntProto{}
		osmoutils.MustGet(store, iterator.Key(), &accruedValue)
		keyParts := strings.Split(string(iterator.Key()), types.KeySeparator)
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

	return takerFeeSkimAccumulators
}

// Used to clear the TakerFeeShareAccumulator records for a specific tier 1 denomination, specifically after the distributions have been completed after epoch.
func (k Keeper) DeleteAllTakerFeeShareAccumulatorsForTierDenom(ctx sdk.Context, tierDenom string) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyTakerFeeShareTier1DenomAccrualForAllDenoms(tierDenom))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}

//
// Registered Alloyed Pool States
//

// Used for setting a specific registered alloyed pool in the store.
// Used in the MsgRegisterAlloyedPool, for governance.
func (k *Keeper) SetRegisteredAlloyedPool(ctx sdk.Context, poolId uint64) error {
	store := ctx.KVStore(k.storeKey)

	cwPool, err := k.GetPool(ctx, poolId)
	if err != nil {
		return err
	}

	// Check if pool is of type CosmWasmPool
	if cwPool.GetType() != types.CosmWasm {
		return fmt.Errorf("pool with id %d is not a CosmWasmPool", poolId)
	}

	contractAddr := cwPool.GetAddress()

	alloyedDenom, err := k.queryAndCheckAlloyedDenom(ctx, contractAddr)
	if err != nil {
		return err
	}

	takerFeeShareAlloyDenoms, err := k.snapshotTakerFeeShareAlloyComposition(ctx, contractAddr)
	if err != nil {
		return err
	}

	registeredAlloyedPool := types.AlloyContractTakerFeeShareState{
		ContractAddress:         contractAddr.String(),
		TakerFeeShareAgreements: takerFeeShareAlloyDenoms,
	}

	bz, err := proto.Marshal(&registeredAlloyedPool)
	if err != nil {
		return err
	}

	key := types.FormatRegisteredAlloyPoolKey(poolId, alloyedDenom)
	store.Set(key, bz)

	// Set cache value
	k.cachedRegisteredAlloyPoolToStateMap[alloyedDenom] = registeredAlloyedPool
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

// Used in the RegisteredAlloyedPoolFromDenomRequest gRPC query.
func (k Keeper) GetRegisteredAlloyedPoolFromDenom(ctx sdk.Context, alloyedDenom string) (types.AlloyContractTakerFeeShareState, bool) {
	registeredAlloyedPool, found := k.cachedRegisteredAlloyPoolToStateMap[alloyedDenom]
	if !found {
		return types.AlloyContractTakerFeeShareState{}, false
	}
	return registeredAlloyedPool, true
}

func (k Keeper) GetRegisteredAlloyedPoolFromPoolId(ctx sdk.Context, poolId uint64) (string, types.AlloyContractTakerFeeShareState, error) {
	store := ctx.KVStore(k.storeKey)
	prefix := types.FormatRegisteredAlloyPoolKeyPoolIdOnly(poolId)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	if !iterator.Valid() {
		return "", types.AlloyContractTakerFeeShareState{}, fmt.Errorf("no registered alloyed pool found for poolId %d", poolId)
	}

	registeredAlloyedPool := types.AlloyContractTakerFeeShareState{}
	osmoutils.MustGet(store, iterator.Key(), &registeredAlloyedPool)

	key := string(iterator.Key())
	parts := strings.Split(key, types.KeySeparator)
	if len(parts) < 3 {
		return "", types.AlloyContractTakerFeeShareState{}, fmt.Errorf("invalid key format")
	}
	alloyedDenom := parts[len(parts)-1]

	return alloyedDenom, registeredAlloyedPool, nil
}

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

// Used for creating the map used for the registered alloyed pools cache.
func (k Keeper) GetAllRegisteredAlloyedPoolsMap(ctx sdk.Context) (map[string]types.AlloyContractTakerFeeShareState, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyRegisteredAlloyPool)
	defer iterator.Close()

	registeredAlloyedPoolsMap := make(map[string]types.AlloyContractTakerFeeShareState)
	for ; iterator.Valid(); iterator.Next() {
		registeredAlloyedPool := types.AlloyContractTakerFeeShareState{}
		osmoutils.MustGet(store, iterator.Key(), &registeredAlloyedPool)

		key := string(iterator.Key())
		parts := strings.Split(key, types.KeySeparator)
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid key format")
		}
		alloyedDenom := parts[len(parts)-1]
		registeredAlloyedPoolsMap[alloyedDenom] = registeredAlloyedPool
	}

	return registeredAlloyedPoolsMap, nil
}

// Used for initializing the cache for the registered alloyed pools.
func (k *Keeper) SetAllRegisteredAlloyedPoolsCached(ctx sdk.Context) error {
	registeredAlloyPools, err := k.GetAllRegisteredAlloyedPoolsMap(ctx)
	if err != nil {
		return err
	}
	k.cachedRegisteredAlloyPoolToStateMap = registeredAlloyPools
	return nil
}

//
// Registered Alloyed Pool Ids
//

// Used for creating the map used for the registered alloyed pools id cache.
func (k Keeper) GetAllRegisteredAlloyedPoolsIdArray(ctx sdk.Context) ([]uint64, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyRegisteredAlloyPool)
	defer iterator.Close()

	registeredAlloyedPoolsIdArray := []uint64{}
	for ; iterator.Valid(); iterator.Next() {
		key := string(iterator.Key())
		parts := strings.Split(key, types.KeySeparator)
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid key format")
		}
		alloyedIdStr := parts[len(parts)-2]
		// Convert the string to uint64
		alloyedId, err := strconv.ParseUint(alloyedIdStr, 10, 64)
		if err != nil {
			return nil, err
		}
		registeredAlloyedPoolsIdArray = append(registeredAlloyedPoolsIdArray, alloyedId)
	}
	sort.Slice(registeredAlloyedPoolsIdArray, func(i, j int) bool { return registeredAlloyedPoolsIdArray[i] < registeredAlloyedPoolsIdArray[j] })

	return registeredAlloyedPoolsIdArray, nil
}

// Used for initializing the cache for the registered alloyed pools id.
func (k *Keeper) SetAllRegisteredAlloyedPoolsIdCached(ctx sdk.Context) error {
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

func (k Keeper) queryAndCheckAlloyedDenom(ctx sdk.Context, contractAddr sdk.AccAddress) (string, error) {
	queryBz := []byte(`{"get_share_denom": {}}`)
	respBz, err := k.wasmKeeper.QuerySmart(ctx, contractAddr, queryBz)
	if err != nil {
		return "", err
	}

	var response ShareDenomResponse
	err = json.Unmarshal(respBz, &response)
	if err != nil {
		return "", err
	}
	alloyedDenom := response.ShareDenom

	parts := strings.Split(alloyedDenom, "/")
	if len(parts) != 4 {
		return "", fmt.Errorf("invalid format for alloyedDenom")
	}

	if parts[0] != "factory" {
		return "", fmt.Errorf("first part of alloyedDenom should be 'factory'")
	}

	if parts[1] != contractAddr.String() {
		return "", fmt.Errorf("second part of alloyedDenom should match contractAddr")
	}

	if parts[2] != "alloyed" {
		return "", fmt.Errorf("third part of alloyedDenom should be 'alloyed'")
	}

	return alloyedDenom, nil
}

func (k Keeper) snapshotTakerFeeShareAlloyComposition(ctx sdk.Context, contractAddr sdk.AccAddress) ([]types.TakerFeeShareAgreement, error) {
	// TODO: Need to add logic for scaling factors
	queryBz := []byte(`{"get_total_pool_liquidity": {}}`)
	respBz, err := k.wasmKeeper.QuerySmart(ctx, contractAddr, queryBz)
	if err != nil {
		return []types.TakerFeeShareAgreement{}, err
	}

	var response TotalPoolLiquidityResponse
	err = json.Unmarshal(respBz, &response)
	if err != nil {
		return []types.TakerFeeShareAgreement{}, err
	}
	totalPoolLiquidity := response.TotalPoolLiquidity

	totalAlloyedLiquidity := osmomath.ZeroDec()
	var assetsWithShareAgreement []sdk.Coin
	var takerFeeShareAgreements []types.TakerFeeShareAgreement
	var skimAddresses []string
	var skimPercents []osmomath.Dec
	for _, coin := range totalPoolLiquidity {
		totalAlloyedLiquidity = totalAlloyedLiquidity.Add(coin.Amount.ToLegacyDec())
		takerFeeShareAgreement, found := k.GetTakerFeeShareAgreementFromDenom(ctx, coin.Denom)
		if !found {
			continue
		}
		assetsWithShareAgreement = append(assetsWithShareAgreement, coin)
		skimAddresses = append(skimAddresses, takerFeeShareAgreement.SkimAddress)
		skimPercents = append(skimPercents, takerFeeShareAgreement.SkimPercent)
	}

	if totalAlloyedLiquidity.IsZero() {
		return []types.TakerFeeShareAgreement{}, fmt.Errorf("totalAlloyedLiquidity is zero")
	}

	for i, coin := range assetsWithShareAgreement {
		scaledSkim := coin.Amount.ToLegacyDec().Quo(totalAlloyedLiquidity).Mul(skimPercents[i])
		takerFeeShareAgreements = append(takerFeeShareAgreements, types.TakerFeeShareAgreement{
			Denom:       coin.Denom,
			SkimPercent: scaledSkim,
			SkimAddress: skimAddresses[i],
		})
	}

	return takerFeeShareAgreements, nil
}

func (k *Keeper) recalculateAndSetTakerFeeShareAlloyComposition(ctx sdk.Context, poolId uint64) error {
	alloyedDenom, registeredAlloyedPoolPrior, err := k.GetRegisteredAlloyedPoolFromPoolId(ctx, poolId)
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

	store := ctx.KVStore(k.storeKey)
	key := types.FormatRegisteredAlloyPoolKey(poolId, alloyedDenom)
	store.Set(key, bz)

	// Set cache value
	k.cachedRegisteredAlloyPoolToStateMap[alloyedDenom] = registeredAlloyedPool

	return nil
}
