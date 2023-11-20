package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey

	gammKeeper          types.SwapI
	concentratedKeeper  types.SwapI
	bankKeeper          types.BankI
	accountKeeper       types.AccountI
	communityPoolKeeper types.CommunityPoolI

	poolCreationListeners types.PoolCreationListeners

	routes map[types.PoolType]types.SwapI

	paramSpace paramtypes.Subspace
}

func NewKeeper(storeKey storetypes.StoreKey, paramSpace paramtypes.Subspace, gammKeeper types.SwapI, concentratedKeeper types.SwapI, bankKeeper types.BankI, accountKeeper types.AccountI, communityPoolKeeper types.CommunityPoolI) *Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	routes := map[types.PoolType]types.SwapI{
		types.Balancer: gammKeeper,
	}

	return &Keeper{storeKey: storeKey, paramSpace: paramSpace, gammKeeper: gammKeeper, concentratedKeeper: concentratedKeeper, bankKeeper: bankKeeper, accountKeeper: accountKeeper, communityPoolKeeper: communityPoolKeeper, routes: routes}
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
}

// ExportGenesis returns the poolmanager module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:     k.GetParams(ctx),
		NextPoolId: k.GetNextPoolId(ctx),
		PoolRoutes: k.GetAllPoolRoutes(ctx),
	}
}

// GetNextPoolId returns the next pool id.
func (k Keeper) GetNextPoolId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	nextPoolId := gogotypes.UInt64Value{}
	osmoutils.MustGet(store, types.KeyNextGlobalPoolId, &nextPoolId)
	return nextPoolId.Value
}

// SetPoolCreationListeners sets the pool creation listeners.
func (k *Keeper) SetPoolCreationListeners(listeners types.PoolCreationListeners) *Keeper {
	if k.poolCreationListeners != nil {
		panic("cannot set pool creation listeners twice")
	}

	k.poolCreationListeners = listeners

	return k
}

// SetNextPoolId sets next pool Id.
func (k Keeper) SetNextPoolId(ctx sdk.Context, poolId uint64) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.KeyNextGlobalPoolId, &gogotypes.UInt64Value{Value: poolId})
}
