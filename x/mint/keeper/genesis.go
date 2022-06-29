package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/mint/types"
)

const developerVestingAmount = 225_000_000_000_000

// InitGenesis new mint genesis.
func (k Keeper) InitGenesis(ctx sdk.Context, ak types.AccountKeeper, bk types.BankKeeper, data *types.GenesisState) {
	if data == nil {
		panic("nil mint genesis state")
	}

	data.Minter.EpochProvisions = data.Params.GenesisEpochProvisions
	k.SetMinter(ctx, data.Minter)
	k.SetParams(ctx, data.Params)

	// The account should be exported in the ExportGenesis of the
	// x/auth SDK module. Therefore, we check for existence here
	// to avoid overwriting pre-existing genesis account data.
	if !ak.HasAccount(ctx, ak.GetModuleAddress(types.DeveloperVestingModuleAcctName)) {
		totalDeveloperVestingCoins := sdk.NewCoin(data.Params.MintDenom, sdk.NewInt(developerVestingAmount))

		if err := k.CreateDeveloperVestingModuleAccount(ctx, totalDeveloperVestingCoins); err != nil {
			panic(err)
		}

		bk.AddSupplyOffset(ctx, data.Params.MintDenom, sdk.NewInt(developerVestingAmount).Neg())
	}

	k.SetLastHalvenEpochNum(ctx, data.HalvenStartedEpoch)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	if params.WeightedDeveloperRewardsReceivers == nil {
		params.WeightedDeveloperRewardsReceivers = make([]types.WeightedAddress, 0)
	}

	lastHalvenEpoch := k.GetLastHalvenEpochNum(ctx)
	return types.NewGenesisState(minter, params, lastHalvenEpoch)
}
