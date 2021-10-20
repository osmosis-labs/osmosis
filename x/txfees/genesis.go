package txfees

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/txfees/keeper"
	"github.com/osmosis-labs/osmosis/x/txfees/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetBaseDenom(ctx, genState.Basetoken)
	k.SetFeeTokens(ctx, genState.Feetokens)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Basetoken, _ = k.GetBaseDenom(ctx)
	genesis.Feetokens = k.GetFeeTokens(ctx)
	return genesis
}
