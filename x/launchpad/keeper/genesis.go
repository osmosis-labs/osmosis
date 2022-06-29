package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/launchpad/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k Keeper, genState types.GenesisState) {
	// TODO setSales, setNextSaleNumber
	//TODO  k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the launchpad module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	// TODO export genesis -- GetSales, GetUserPositions, GetNextSaleNumber, GetParams
	return &types.GenesisState{}
}
