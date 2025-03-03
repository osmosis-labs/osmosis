package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v29/x/cron/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}
	var (
		cronID uint64
	)
	for _, item := range genState.CronJobs {
		k.SetCronJob(ctx, item)
		// Set the last cron ID
		cronID = item.Id
	}
	k.SetCronID(ctx, cronID)
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:   k.GetParams(ctx),
		CronJobs: k.GetCronJobs(ctx),
	}
}
