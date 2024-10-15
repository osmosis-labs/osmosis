package callback

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v26/x/callback/keeper"
	"github.com/osmosis-labs/osmosis/v26/x/callback/types"
)

// InitGenesis initializes the module genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	params := genState.Params
	err := k.Params.Set(ctx, params)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis exports the module genesis for the current block.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}
	callbacks, err := k.GetAllCallbacks(ctx)
	if err != nil {
		panic(err)
	}
	return types.NewGenesisState(params, callbacks)
}
