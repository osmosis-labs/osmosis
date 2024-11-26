package poolmanager

import (
	"fmt"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"

	storetypes "cosmossdk.io/store/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey

	gammKeeper           types.PoolModuleI
	concentratedKeeper   types.ConcentratedI
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

	cachedTakerFeeShareAgreementMap          map[string]types.TakerFeeShareAgreement
	cachedRegisteredAlloyPoolByAlloyDenomMap map[string]types.AlloyContractTakerFeeShareState
}

func NewKeeper(storeKey storetypes.StoreKey, paramSpace paramtypes.Subspace, gammKeeper types.PoolModuleI, concentratedKeeper types.ConcentratedI, cosmwasmpoolKeeper types.PoolModuleI, bankKeeper types.BankI, accountKeeper types.AccountI, communityPoolKeeper types.CommunityPoolI, stakingKeeper types.StakingKeeper, protorevKeeper types.ProtorevKeeper, wasmKeeper types.WasmKeeper) *Keeper {
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

	return &Keeper{
		storeKey:                                 storeKey,
		paramSpace:                               paramSpace,
		gammKeeper:                               gammKeeper,
		concentratedKeeper:                       concentratedKeeper,
		cosmwasmpoolKeeper:                       cosmwasmpoolKeeper,
		bankKeeper:                               bankKeeper,
		accountKeeper:                            accountKeeper,
		communityPoolKeeper:                      communityPoolKeeper,
		routes:                                   routesMap,
		poolModules:                              routesList,
		stakingKeeper:                            stakingKeeper,
		protorevKeeper:                           protorevKeeper,
		wasmKeeper:                               wasmKeeper,
		cachedPoolModules:                        cachedPoolModules,
		cachedTakerFeeShareAgreementMap:          cachedTakerFeeShareAgreementMap,
		cachedRegisteredAlloyPoolByAlloyDenomMap: cachedRegisteredAlloyPoolMap,
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
		k.SetDenomPairTakerFee(ctx, denomPairTakerFee.TokenInDenom, denomPairTakerFee.TokenOutDenom, denomPairTakerFee.TakerFee)
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
	// Here, the only time in which these caches are empty is during the start up of the node.
	// Once the node has started up and runs the first BeginBlock of the poolmanager module,
	// it will populate the caches. Every single subsequent BeginBlock, this logic will be a no-op.
	if len(k.cachedTakerFeeShareAgreementMap) == 0 || len(k.cachedRegisteredAlloyPoolByAlloyDenomMap) == 0 {
		err := k.setTakerFeeShareAgreementsMapCached(ctx)
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("%w", types.ErrSetTakerFeeShareAgreementsMapCached).Error())
		}
		err = k.setAllRegisteredAlloyedPoolsByDenomCached(ctx)
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("%w", types.ErrSetAllRegisteredAlloyedPoolsByDenomCached).Error())
		}
	}
}

// AlloyedAssetCompositionUpdateRate is the rate in blocks at which the taker fee share alloy composition is updated in the end block.
var AlloyedAssetCompositionUpdateRate = int64(700)

// EndBlock updates the taker fee share alloy composition for all registered alloyed pools
// if the current block height is a multiple of the alloyedAssetCompositionUpdateRate.
func (k *Keeper) EndBlock(ctx sdk.Context) {
	if ctx.BlockHeight()%AlloyedAssetCompositionUpdateRate == 0 {
		registeredAlloyPoolIds, err := k.getAllRegisteredAlloyedPoolsIdArray(ctx)
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("unable to get all registered alloyed pools: %w", err).Error())
			return
		}
		for _, id := range registeredAlloyPoolIds {
			err := k.recalculateAndSetTakerFeeShareAlloyComposition(ctx, id)
			if err != nil {
				ctx.Logger().Error(fmt.Errorf(
					"%s for pool id %d: %v", types.ErrSetRegisteredAlloyedPool, id, err,
				).Error())
			}
		}
	}
}
