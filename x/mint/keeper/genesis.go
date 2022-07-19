package keeper

import (
	"github.com/osmosis-labs/osmosis/v10/x/mint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis new mint genesis.
func (k Keeper) InitGenesis(ctx sdk.Context, ak types.AccountKeeper, bk types.BankKeeper, data *types.GenesisState) {
	data.Minter.EpochProvisions = data.Params.GenesisEpochProvisions
	k.SetMinter(ctx, data.Minter)
	k.SetParams(ctx, data.Params)

	if !ak.HasAccount(ctx, ak.GetModuleAddress(types.ModuleName)) {
		ak.GetModuleAccount(ctx, types.ModuleName)
		totalDeveloperVestingCoins := sdk.NewCoin(data.Params.MintDenom, sdk.NewInt(225_000_000_000_000))
		k.CreateDeveloperVestingModuleAccount(ctx, totalDeveloperVestingCoins)
		bk.AddSupplyOffset(ctx, data.Params.MintDenom, sdk.NewInt(225_000_000_000_000).Neg())
	}

	k.SetLastHalvenEpochNum(ctx, data.HalvenStartedEpoch)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)
	lastHalvenEpoch := k.GetLastHalvenEpochNum(ctx)
	return types.NewGenesisState(minter, params, lastHalvenEpoch)
}
