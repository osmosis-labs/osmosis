package superfluid

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	// initialize superfluid assets
	for _, asset := range genState.SuperfluidAssets {
		k.SetSuperfluidAsset(ctx, asset)
	}

	// initialize epoch twap price
	for _, priceRecord := range genState.OsmoEquivalentMultipliers {
		k.SetOsmoEquivalentMultiplier(ctx, priceRecord.EpochNumber, priceRecord.Denom, priceRecord.Multiplier)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		SuperfluidAssets:          k.GetAllSuperfluidAssets(ctx),
		OsmoEquivalentMultipliers: k.GetAllOsmoEquivalentMultipliers(ctx),
	}
}
