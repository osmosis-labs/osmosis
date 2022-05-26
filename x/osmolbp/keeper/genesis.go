package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/osmolbp/api"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k Keeper, genState api.GenesisState) {
	// TODO setLBPs, setNextLBPNumber
	//TODO  k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the osmolbp module's exported genesis.
func ExportGenesis(ctx sdk.Context, k Keeper) *api.GenesisState {
	// TODO export genesis -- GetLBPs, GetUserPositions, GetNextLBPNumber, GetParams
	return &api.GenesisState{}
}
