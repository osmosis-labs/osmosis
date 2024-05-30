package poolmanager

import (
	"fmt"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"

	storetypes "cosmossdk.io/store/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey

	gammKeeper           types.PoolModuleI
	concentratedKeeper   types.PoolModuleI
	cosmwasmpoolKeeper   types.PoolModuleI
	poolIncentivesKeeper types.PoolIncentivesKeeperI
	bankKeeper           types.BankI
	accountKeeper        types.AccountI
	communityPoolKeeper  types.CommunityPoolI
	stakingKeeper        types.StakingKeeper
	protorevKeeper       types.ProtorevKeeper
	wasmKeeper           types.WasmKeeper

	// routes is a map to get the pool module by id.
	routes map[types.PoolType]types.PoolModuleI

	// map from poolId to the swap module + Gas consumed amount
	// note that after getPoolModule doesn't return an error
	// it will always return the same result. Meaning its perfect for a sync.map cache.
	cachedPoolModules *sync.Map

	// poolModules is a list of all pool modules.
	// It is used when an operation has to be applied to all pool
	// modules. Since map iterations are non-deterministic, we
	// use this list to ensure deterministic iteration.
	poolModules []types.PoolModuleI

	paramSpace paramtypes.Subspace

	defaultTakerFeeBz  []byte
	defaultTakerFeeVal osmomath.Dec

	cachedTakerFeeShareAgreementMap     map[string]types.TakerFeeShareAgreement
	cachedRegisteredAlloyPoolToStateMap map[string]types.AlloyContractTakerFeeShareState
	cachedRegisteredAlloyedPoolIdArray  []uint64
}

func NewKeeper(storeKey storetypes.StoreKey, paramSpace paramtypes.Subspace, gammKeeper types.PoolModuleI, concentratedKeeper types.PoolModuleI, cosmwasmpoolKeeper types.PoolModuleI, bankKeeper types.BankI, accountKeeper types.AccountI, communityPoolKeeper types.CommunityPoolI, stakingKeeper types.StakingKeeper, protorevKeeper types.ProtorevKeeper, wasmKeeper types.WasmKeeper) *Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	routesMap := map[types.PoolType]types.PoolModuleI{
		types.Balancer:     gammKeeper,
		types.Stableswap:   gammKeeper,
		types.Concentrated: concentratedKeeper,
		types.CosmWasm:     cosmwasmpoolKeeper,
	}

	routesList := []types.PoolModuleI{
		gammKeeper, concentratedKeeper, cosmwasmpoolKeeper,
	}

	cachedPoolModules := &sync.Map{}
	cachedTakerFeeShareAgreementMap := make(map[string]types.TakerFeeShareAgreement)
	cachedRegisteredAlloyPoolMap := make(map[string]types.AlloyContractTakerFeeShareState)
	cachedRegistedAlloyedPoolIdArray := []uint64{}

	return &Keeper{
		storeKey:                            storeKey,
		paramSpace:                          paramSpace,
		gammKeeper:                          gammKeeper,
		concentratedKeeper:                  concentratedKeeper,
		cosmwasmpoolKeeper:                  cosmwasmpoolKeeper,
		bankKeeper:                          bankKeeper,
		accountKeeper:                       accountKeeper,
		communityPoolKeeper:                 communityPoolKeeper,
		routes:                              routesMap,
		poolModules:                         routesList,
		stakingKeeper:                       stakingKeeper,
		protorevKeeper:                      protorevKeeper,
		wasmKeeper:                          wasmKeeper,
		cachedPoolModules:                   cachedPoolModules,
		cachedTakerFeeShareAgreementMap:     cachedTakerFeeShareAgreementMap,
		cachedRegisteredAlloyPoolToStateMap: cachedRegisteredAlloyPoolMap,
		cachedRegisteredAlloyedPoolIdArray:  cachedRegistedAlloyedPoolIdArray,
	}
}

func (k *Keeper) ResetCaches() {
	k.cachedPoolModules = &sync.Map{}
}

// GetParams returns the total set of poolmanager parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of poolmanager parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// SetParam sets a specific poolmanger module's parameter with the provided parameter.
func (k Keeper) SetParam(ctx sdk.Context, key []byte, value interface{}) {
	k.paramSpace.Set(ctx, key, value)
}

// InitGenesis initializes the poolmanager module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.SetNextPoolId(ctx, genState.NextPoolId)
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	k.SetParams(ctx, genState.Params)

	for _, poolRoute := range genState.PoolRoutes {
		k.SetPoolRoute(ctx, poolRoute.PoolId, poolRoute.PoolType)
	}

	// We track taker fees generated in the module's KVStore.
	for _, coin := range genState.TakerFeesTracker.TakerFeesToStakers {
		if err := k.UpdateTakerFeeTrackerForStakersByDenom(ctx, coin.Denom, coin.Amount); err != nil {
			panic(err)
		}
	}
	for _, coin := range genState.TakerFeesTracker.TakerFeesToCommunityPool {
		if err := k.UpdateTakerFeeTrackerForCommunityPoolByDenom(ctx, coin.Denom, coin.Amount); err != nil {
			panic(err)
		}
	}
	k.SetTakerFeeTrackerStartHeight(ctx, genState.TakerFeesTracker.HeightAccountingStartsFrom)

	// Set the pool volumes KVStore.
	for _, poolVolume := range genState.PoolVolumes {
		k.SetVolume(ctx, poolVolume.PoolId, poolVolume.PoolVolume)
	}

	// Set the denom pair taker fees KVStore.
	for _, denomPairTakerFee := range genState.DenomPairTakerFeeStore {
		k.SetDenomPairTakerFee(ctx, denomPairTakerFee.Denom0, denomPairTakerFee.Denom1, denomPairTakerFee.TakerFee)
	}
}

// ExportGenesis returns the poolmanager module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	pools, err := k.AllPools(ctx)
	if err != nil {
		panic(err)
	}

	// Utilize poolVolumes struct to export pool volumes from KVStore.
	poolVolumes := make([]*types.PoolVolume, len(pools))
	for i, pool := range pools {
		poolVolume := k.GetTotalVolumeForPool(ctx, pool.GetId())
		poolVolumes[i] = &types.PoolVolume{
			PoolId:     pool.GetId(),
			PoolVolume: poolVolume,
		}
	}

	// Utilize denomPairTakerFee struct to export taker fees from KVStore.
	denomPairTakerFees, err := k.GetAllTradingPairTakerFees(ctx)
	if err != nil {
		panic(err)
	}

	// Export KVStore values to the genesis state so they can be imported in init genesis.
	takerFeesTracker := types.TakerFeesTracker{
		TakerFeesToStakers:         k.GetTakerFeeTrackerForStakers(ctx),
		TakerFeesToCommunityPool:   k.GetTakerFeeTrackerForCommunityPool(ctx),
		HeightAccountingStartsFrom: k.GetTakerFeeTrackerStartHeight(ctx),
	}
	return &types.GenesisState{
		Params:                 k.GetParams(ctx),
		NextPoolId:             k.GetNextPoolId(ctx),
		PoolRoutes:             k.getAllPoolRoutes(ctx),
		TakerFeesTracker:       &takerFeesTracker,
		PoolVolumes:            poolVolumes,
		DenomPairTakerFeeStore: denomPairTakerFees,
	}
}

// GetNextPoolId returns the next pool id.
func (k Keeper) GetNextPoolId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	nextPoolId := gogotypes.UInt64Value{}
	osmoutils.MustGet(store, types.KeyNextGlobalPoolId, &nextPoolId)
	return nextPoolId.Value
}

// SetNextPoolId sets next pool Id.
func (k Keeper) SetNextPoolId(ctx sdk.Context, poolId uint64) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.KeyNextGlobalPoolId, &gogotypes.UInt64Value{Value: poolId})
}

// SetPoolIncentivesKeeper sets pool incentives keeper
func (k *Keeper) SetPoolIncentivesKeeper(poolIncentivesKeeper types.PoolIncentivesKeeperI) {
	k.poolIncentivesKeeper = poolIncentivesKeeper
}

// SetStakingKeeper sets staking keeper
func (k *Keeper) SetStakingKeeper(stakingKeeper types.StakingKeeper) {
	k.stakingKeeper = stakingKeeper
}

// SetProtorevKeeper sets protorev keeper
func (k *Keeper) SetProtorevKeeper(protorevKeeper types.ProtorevKeeper) {
	k.protorevKeeper = protorevKeeper
}

// SetWasmKeeper sets wasm keeper
func (k *Keeper) SetWasmKeeper(wasmKeeper types.WasmKeeper) {
	k.wasmKeeper = wasmKeeper
}

// BeginBlock sets the poolmanager caches if they are empty
func (k *Keeper) BeginBlock(ctx sdk.Context) {
	// Set the caches if they are empty
	if len(k.cachedTakerFeeShareAgreementMap) == 0 || len(k.cachedRegisteredAlloyPoolToStateMap) == 0 || len(k.cachedRegisteredAlloyedPoolIdArray) == 0 {
		err := k.SetTakerFeeShareAgreementsMapCached(ctx)
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("error in setting taker fee share agreements map cached: %w", err).Error())
		}
		err = k.SetAllRegisteredAlloyedPoolsCached(ctx)
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("error in setting all registered alloyed pools cached: %w", err).Error())
		}
		err = k.SetAllRegisteredAlloyedPoolsIdCached(ctx)
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("error in setting all registered alloyed pools id cached: %w", err).Error())
		}
	}
}

var alloyedAssetCompositionUpdateRate = int64(1)

// EndBlock recalculates the taker fee share alloy composition for all altered registered alloyed pools
func (k *Keeper) EndBlock(ctx sdk.Context) {
	if ctx.BlockHeight()%alloyedAssetCompositionUpdateRate == 0 {
		for _, id := range k.cachedRegisteredAlloyedPoolIdArray {
			err := k.recalculateAndSetTakerFeeShareAlloyComposition(ctx, id)
			if err != nil {
				ctx.Logger().Error(fmt.Errorf(
					"error in setting registered alloyed pool for pool id %d: %w", id, err,
				).Error())
			}
		}
	}
}

// getCacheTrackers returns the cached trackers, used for testing.
// nolint: unused
func (k Keeper) getCacheTrackers() (map[string]types.TakerFeeShareAgreement, map[string]types.AlloyContractTakerFeeShareState, []uint64) {
	return k.cachedTakerFeeShareAgreementMap, k.cachedRegisteredAlloyPoolToStateMap, k.cachedRegisteredAlloyedPoolIdArray
}

// setCacheTrackers sets the cached trackers, used for testing.
// nolint: unused
func (k *Keeper) setCacheTrackers(takerFeeShareAgreement map[string]types.TakerFeeShareAgreement, registeredAlloyPoolToState map[string]types.AlloyContractTakerFeeShareState, registeredAlloyedPoolId []uint64) {
	if takerFeeShareAgreement != nil {
		k.cachedTakerFeeShareAgreementMap = takerFeeShareAgreement
	}
	if registeredAlloyPoolToState != nil {
		k.cachedRegisteredAlloyPoolToStateMap = registeredAlloyPoolToState
	}
	if registeredAlloyedPoolId != nil {
		k.cachedRegisteredAlloyedPoolIdArray = registeredAlloyedPoolId
	}
}
