package protorev

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/protorev/keeper"
	"github.com/osmosis-labs/osmosis/v12/x/protorev/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Perform stateless validation on the genesis state
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	// Init module parameters
	k.SetParams(ctx, genState.Params)

	// Init the statistics on genesis
	defaultStatistics := &types.ProtoRevStatistics{}
	k.SetProtoRevStatistics(ctx, defaultStatistics)

	// Init all of the searcher route
	for _, searcherRoutes := range genState.Routes {
		k.SetSearcherRoutes(ctx, searcherRoutes.PoolId, &searcherRoutes)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	return genesis
}
