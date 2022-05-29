package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/launchpad/api"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k Keeper, genState api.GenesisState) {
	// TODO setSales, setNextSaleNumber
	//TODO  k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the launchpad module's exported genesis.
func ExportGenesis(ctx sdk.Context, k Keeper) *api.GenesisState {
	// TODO export genesis -- GetSales, GetUserPositions, GetNextSaleNumber, GetParams
	return &api.GenesisState{}
}
