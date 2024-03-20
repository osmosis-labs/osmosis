package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

// InitGenesis initializes the bridge module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	createdAssets, err := k.createAssets(ctx, genState.Params.Assets)
	if err != nil {
		panic(fmt.Errorf("can't create assets on x/bridge genesis: %w", err))
	}
	genState.Params.Assets = createdAssets

	// don't need to specifically create the signers, just save them

	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the bridge module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}
