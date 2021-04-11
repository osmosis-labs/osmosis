package mint

import (
	"github.com/c-osmosis/osmosis/x/mint/keeper"
	"github.com/c-osmosis/osmosis/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis new mint genesis
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, ak types.AccountKeeper, data *types.GenesisState) {
	keeper.SetMinter(ctx, data.Minter)
	keeper.SetParams(ctx, data.Params)
	ak.GetModuleAccount(ctx, types.ModuleName)
	keeper.SetLastEpochTime(ctx, ctx.BlockTime())
	keeper.SetEpochNum(ctx, data.CurrentEpoch)
	keeper.SetLastHalvenEpochNum(ctx, data.HalvenStartedEpoch)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	minter := keeper.GetMinter(ctx)
	params := keeper.GetParams(ctx)
	curEpoch := keeper.GetEpochNum(ctx)
	lastEpoch := keeper.GetLastHalvenEpochNum(ctx)
	return types.NewGenesisState(minter, params, curEpoch, lastEpoch)
}
