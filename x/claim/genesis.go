package claim

import (
	"github.com/c-osmosis/osmosis/x/claim/keeper"
	"github.com/c-osmosis/osmosis/x/claim/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetModuleAccountBalance(ctx, genState.AirdropAmount)
	k.SetParams(ctx, types.Params{
		AirdropStart:       genState.AirdropStart,
		DurationUntilDecay: genState.DurationUntilDecay,
		DurationOfDecay:    genState.DurationOfDecay,
	})
	k.SetClaimables(ctx, genState.Claimables)
	for _, activities := range genState.Activities {
		user, err := sdk.AccAddressFromBech32(activities.User)
		if err != nil {
			panic(err)
		}
		k.SetUserActions(ctx, user, activities.Actions)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params, _ := k.GetParams(ctx)
	genesis := types.DefaultGenesis()
	genesis.AirdropStart = params.AirdropStart
	genesis.DurationUntilDecay = params.DurationUntilDecay
	genesis.DurationOfDecay = params.DurationOfDecay
	genesis.Claimables = k.GetClaimables(ctx)
	genesis.Activities = k.GetActivities(ctx)

	return genesis
}
