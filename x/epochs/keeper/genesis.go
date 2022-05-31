package keeper

import (
	"time"

<<<<<<< HEAD:x/epochs/genesis.go
	"github.com/osmosis-labs/osmosis/v10/x/epochs/keeper"
	"github.com/osmosis-labs/osmosis/v10/x/epochs/types"

=======
>>>>>>> 61a207f8 (chore: move init export genesis to keepers (#1631)):x/epochs/keeper/genesis.go
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/epochs/types"
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
