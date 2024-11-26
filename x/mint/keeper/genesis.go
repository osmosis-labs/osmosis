package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/mint/types"
)

const developerVestingAmount = 225_000_000_000_000

// InitGenesis new mint genesis.
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	if data == nil {
		panic("nil mint genesis state")
	}

	data.Minter.EpochProvisions = data.Params.GenesisEpochProvisions
	k.SetMinter(ctx, data.Minter)
	k.SetParams(ctx, data.Params)

	// The call to GetModuleAccount creates a module account if it does not exist.
	k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)

	// The account should be exported in the ExportGenesis of the
	// x/auth SDK module. Therefore, we check for existence here
	// to avoid overwriting pre-existing genesis account data.
	if !k.accountKeeper.HasAccount(ctx, k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)) {
		totalDeveloperVestingCoins := sdk.NewCoin(data.Params.MintDenom, osmomath.NewInt(developerVestingAmount))

		if err := k.createDeveloperVestingModuleAccount(ctx, totalDeveloperVestingCoins); err != nil {
			panic(err)
		}

		k.bankKeeper.AddSupplyOffset(ctx, data.Params.MintDenom, osmomath.NewInt(developerVestingAmount).Neg())
	}

	k.setLastReductionEpochNum(ctx, data.ReductionStartedEpoch)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	if params.WeightedDeveloperRewardsReceivers == nil {
		params.WeightedDeveloperRewardsReceivers = make([]types.WeightedAddress, 0)
	}

	lastHalvenEpoch := k.getLastReductionEpochNum(ctx)
	return types.NewGenesisState(minter, params, lastHalvenEpoch)
}
