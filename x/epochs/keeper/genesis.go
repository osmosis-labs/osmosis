package keeper

import (
	"time"

	"github.com/osmosis-labs/osmosis/v10/x/epochs/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	// set epoch info from genesis
	for _, epoch := range genState.Epochs {
		// Initialize empty epoch values via Cosmos SDK
		if epoch.StartTime.Equal(time.Time{}) {
			epoch.StartTime = ctx.BlockTime()
		}

		epoch.CurrentEpochStartHeight = ctx.BlockHeight()

		k.SetEpochInfo(ctx, epoch)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Epochs = k.AllEpochInfos(ctx)
	return genesis
}
