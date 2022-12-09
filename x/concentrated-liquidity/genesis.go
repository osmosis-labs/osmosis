package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

// InitGenesis initializes the concentrated-liquidity module with default parameters
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, types.DefaultParams())
}

// ExportGenesis returns the concentrated-liquidity module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}
