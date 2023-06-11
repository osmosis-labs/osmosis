package cosmwasmpool

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/x/cosmwasmpool/types"
)

// InitGenesis initializes the store state from a genesis state.
func (k *Keeper) InitGenesis(ctx sdk.Context, gen *types.GenesisState) {
	k.SetParams(ctx, gen.Params)
}

// ExportGenesis returns the cosmwasm pool's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)
	return &types.GenesisState{
		Params: params,
	}
}
