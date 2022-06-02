package keeper

import (
	"github.com/osmosis-labs/osmosis/v7/x/txfees/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the txfees module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	err := k.SetBaseDenom(ctx, genState.Basedenom)
	if err != nil {
		panic(err)
	}
	err = k.SetFeeTokens(ctx, genState.Feetokens)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis returns the txfees module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Basedenom, _ = k.GetBaseDenom(ctx)
	genesis.Feetokens = k.GetFeeTokens(ctx)
	return genesis
}
