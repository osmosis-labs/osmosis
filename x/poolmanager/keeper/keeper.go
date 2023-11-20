package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey

	gammKeeper          types.SwapI
	bankKeeper          types.BankI
	accountKeeper       types.AccountI
	communityPoolKeeper types.CommunityPoolI

	//FIXME: change to list instead of map for determinism
	routes map[types.PoolType]types.SwapI
}

func NewKeeper(storeKey storetypes.StoreKey, gammKeeper types.SwapI, bankKeeper types.BankI, accountKeeper types.AccountI, communityPoolKeeper types.CommunityPoolI) *Keeper {
	routes := map[types.PoolType]types.SwapI{
		types.Balancer: gammKeeper,
	}

	return &Keeper{storeKey: storeKey, gammKeeper: gammKeeper, bankKeeper: bankKeeper, accountKeeper: accountKeeper, communityPoolKeeper: communityPoolKeeper, routes: routes}
}

// InitGenesis initializes the poolmanager module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.SetNextPoolId(ctx, genState.NextPoolId)
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	for _, poolRoute := range genState.PoolRoutes {
		k.SetPoolRoute(ctx, poolRoute.PoolId, poolRoute.PoolType)
	}
}

// ExportGenesis returns the poolmanager module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
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

// SetNextPoolId sets next pool Id.
func (k Keeper) SetNextPoolId(ctx sdk.Context, poolId uint64) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.KeyNextGlobalPoolId, &gogotypes.UInt64Value{Value: poolId})
}
