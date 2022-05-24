package osmolbp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/osmolbp/api"
	"github.com/osmosis-labs/osmosis/x/osmolbp/keeper"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState api.GenesisState) {
	// k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the osmolbp module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *api.GenesisState {
	return &api.GenesisState{}
}
