package txfees

import (
	"github.com/osmosis-labs/osmosis/v9/x/txfees/keeper"
	"github.com/osmosis-labs/osmosis/v9/x/txfees/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the txfees module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
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
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Basedenom, _ = k.GetBaseDenom(ctx)
	genesis.Feetokens = k.GetFeeTokens(ctx)
	return genesis
}
