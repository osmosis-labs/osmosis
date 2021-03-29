package incentives

import (
	"github.com/c-osmosis/osmosis/x/incentives/keeper"
	"github.com/c-osmosis/osmosis/x/incentives/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetCurrentEpochInfo(ctx, genState.CurrentEpoch, genState.EpochBeginBlock)
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	currentEpoch, epochBeginBlock := k.GetCurrentEpochInfo(ctx)
	return &types.GenesisState{
		Params:          k.GetParams(ctx),
		CurrentEpoch:    currentEpoch,
		EpochBeginBlock: epochBeginBlock,
	}
}
