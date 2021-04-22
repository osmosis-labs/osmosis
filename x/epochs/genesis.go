package epochs

import (
	"time"

	"github.com/c-osmosis/osmosis/x/epochs/keeper"
	"github.com/c-osmosis/osmosis/x/epochs/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	for _, epoch := range genState.Epochs {
		k.SetEpochInfo(ctx, epoch)
		// TODO: when StartTime and CurrentEpochStartTime are not set use ctx.BlockTime()
		if !epoch.EpochEnded && (epoch.StartTime.Equal(time.Time{}) || epoch.CurrentEpochStartTime.Equal(time.Time{})) {
			epoch.StartTime = ctx.BlockTime()
			epoch.CurrentEpochStartTime = ctx.BlockTime()
		}
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Epochs = k.AllEpochInfos(ctx)
	return genesis
}
