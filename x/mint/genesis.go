package mint

import (
	"github.com/osmosis-labs/osmosis/v9/x/mint/keeper"
	"github.com/osmosis-labs/osmosis/v9/x/mint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis new mint genesis.
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, ak types.AccountKeeper, bk types.BankKeeper, data *types.GenesisState) {
	data.Minter.EpochProvisions = data.Params.GenesisEpochProvisions
	keeper.SetMinter(ctx, data.Minter)
	keeper.SetParams(ctx, data.Params)

	if !ak.HasAccount(ctx, ak.GetModuleAddress(types.ModuleName)) {
		ak.GetModuleAccount(ctx, types.ModuleName)
		totalDeveloperVestingCoins := sdk.NewCoin(data.Params.MintDenom, sdk.NewInt(225_000_000_000_000))
		keeper.CreateDeveloperVestingModuleAccount(ctx, totalDeveloperVestingCoins)
		bk.AddSupplyOffset(ctx, data.Params.MintDenom, sdk.NewInt(225_000_000_000_000).Neg())
	}

	keeper.SetLastHalvenEpochNum(ctx, data.HalvenStartedEpoch)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	minter := keeper.GetMinter(ctx)
	params := keeper.GetParams(ctx)
	lastHalvenEpoch := keeper.GetLastHalvenEpochNum(ctx)
	return types.NewGenesisState(minter, params, lastHalvenEpoch)
}
