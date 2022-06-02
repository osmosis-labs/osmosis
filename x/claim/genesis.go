package claim

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v3/x/claim/keeper"
	"github.com/osmosis-labs/osmosis/v3/x/claim/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.CreateModuleAccount(ctx, genState.ModuleAccountBalance)

	if genState.Params.AirdropStartTime.Equal(time.Time{}) {
		genState.Params.AirdropStartTime = ctx.BlockTime()
	}

	err := k.SetParams(ctx, genState.Params)
	if err != nil {
		panic(err)
	}
	err = k.SetClaimRecords(ctx, genState.ClaimRecords)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params, _ := k.GetParams(ctx)
	genesis := types.DefaultGenesis()
	genesis.ModuleAccountBalance = k.GetModuleAccountBalance(ctx)
	genesis.Params = params
	genesis.ClaimRecords = k.GetClaimRecords(ctx)
	return genesis
}
