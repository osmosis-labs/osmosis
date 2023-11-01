package poolmanager

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type Keeper struct {
	storeKey sdk.StoreKey

	gammKeeper           types.PoolModuleI
	concentratedKeeper   types.PoolModuleI
	cosmwasmpoolKeeper   types.PoolModuleI
	poolIncentivesKeeper types.PoolIncentivesKeeperI
	bankKeeper           types.BankI
	accountKeeper        types.AccountI
	communityPoolKeeper  types.CommunityPoolI
	stakingKeeper        types.StakingKeeper
	protorevKeeper       types.ProtorevKeeper

	// routes is a map to get the pool module by id.
	routes map[types.PoolType]types.PoolModuleI

	// poolModules is a list of all pool modules.
	// It is used when an operation has to be applied to all pool
	// modules. Since map iterations are non-deterministic, we
	// use this list to ensure deterministic iteration.
	poolModules []types.PoolModuleI

	paramSpace paramtypes.Subspace
}

func NewKeeper(storeKey sdk.StoreKey, paramSpace paramtypes.Subspace, gammKeeper types.PoolModuleI, concentratedKeeper types.PoolModuleI, cosmwasmpoolKeeper types.PoolModuleI, bankKeeper types.BankI, accountKeeper types.AccountI, communityPoolKeeper types.CommunityPoolI, stakingKeeper types.StakingKeeper, protorevKeeper types.ProtorevKeeper) *Keeper {
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

	return &Keeper{
		storeKey:            storeKey,
		paramSpace:          paramSpace,
		gammKeeper:          gammKeeper,
		concentratedKeeper:  concentratedKeeper,
		cosmwasmpoolKeeper:  cosmwasmpoolKeeper,
		bankKeeper:          bankKeeper,
		accountKeeper:       accountKeeper,
		communityPoolKeeper: communityPoolKeeper,
		routes:              routesMap,
		poolModules:         routesList,
		stakingKeeper:       stakingKeeper,
		protorevKeeper:      protorevKeeper,
	}
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
	// If the values were exported, we set them here.
	// If the values were not exported, we initialize the tracker to zero and set the accounting height to the current height.
	if genState.TakerFeesToStakersTracker != nil {
		k.SetTakerFeeTrackerForStakers(ctx, genState.TakerFeesToStakersTracker.TakerFeesToStakers)
		k.SetTakerFeeTrackerStartHeight(ctx, genState.TakerFeesToStakersTracker.HeightAccountingStartsFrom)
	} else {
		k.SetTakerFeeTrackerForStakers(ctx, sdk.NewCoins())
		k.SetTakerFeeTrackerStartHeight(ctx, uint64(ctx.BlockHeight()))
	}
	if genState.TakerFeesToCommunityPoolTracker != nil {
		k.SetTakerFeeTrackerForCommunityPool(ctx, genState.TakerFeesToCommunityPoolTracker.TakerFeesToCommunityPool)
		k.SetTakerFeeTrackerStartHeight(ctx, genState.TakerFeesToCommunityPoolTracker.HeightAccountingStartsFrom)
	} else {
		k.SetTakerFeeTrackerForCommunityPool(ctx, sdk.NewCoins())
		k.SetTakerFeeTrackerStartHeight(ctx, uint64(ctx.BlockHeight()))
	}
}

// ExportGenesis returns the poolmanager module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	// Export KVStore values to the genesis state so they can be imported in init genesis.
	takerFeesToStakersTracker := types.TakerFeesToStakersTracker{
		TakerFeesToStakers:         k.GetTakerFeeTrackerForStakers(ctx),
		HeightAccountingStartsFrom: k.GetTakerFeeTrackerStartHeight(ctx),
	}
	takerFeesToCommunityPoolTracker := types.TakerFeesToCommunityPoolTracker{
		TakerFeesToCommunityPool:   k.GetTakerFeeTrackerForCommunityPool(ctx),
		HeightAccountingStartsFrom: k.GetTakerFeeTrackerStartHeight(ctx),
	}

	return &types.GenesisState{
		Params:                          k.GetParams(ctx),
		NextPoolId:                      k.GetNextPoolId(ctx),
		PoolRoutes:                      k.getAllPoolRoutes(ctx),
		TakerFeesToStakersTracker:       &takerFeesToStakersTracker,
		TakerFeesToCommunityPoolTracker: &takerFeesToCommunityPoolTracker,
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
