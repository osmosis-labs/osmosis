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

	// initialize osmo equivalent multipliers
	for _, multiplierRecord := range genState.OsmoEquivalentMultipliers {
		k.SetOsmoEquivalentMultiplier(ctx, multiplierRecord.EpochNumber, multiplierRecord.Denom, multiplierRecord.Multiplier)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:                        k.GetParams(ctx),
		SuperfluidAssets:              k.GetAllSuperfluidAssets(ctx),
		OsmoEquivalentMultipliers:     k.GetAllOsmoEquivalentMultipliers(ctx),
		IntermediaryAccounts:          k.GetAllIntermediaryAccounts(ctx),
		IntemediaryAccountConnections: k.GetAllLockIdIntermediaryAccountConnections(ctx),
	}
}
