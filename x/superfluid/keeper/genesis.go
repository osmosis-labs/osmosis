package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	// initialize superfluid assets
	for _, asset := range genState.SuperfluidAssets {
		k.SetSuperfluidAsset(ctx, asset)
	}

	// initialize osmo equivalent multipliers
	for _, multiplierRecord := range genState.OsmoEquivalentMultipliers {
		k.SetOsmoEquivalentMultiplier(ctx, multiplierRecord.EpochNumber, multiplierRecord.Denom, multiplierRecord.Multiplier)
	}

	for _, intermediaryAcc := range genState.IntermediaryAccounts {
		k.SetIntermediaryAccount(ctx, intermediaryAcc)
	}

	// initialize lock id and intermediary connections
	for _, connection := range genState.IntemediaryAccountConnections {
		acc, err := sdk.AccAddressFromBech32(connection.IntermediaryAccount)
		if err != nil {
			panic(err)
		}
		intermediaryAcc := k.GetIntermediaryAccount(ctx, acc)
		if intermediaryAcc.Denom == "" {
			panic("connection to invalid intermediary account found")
		}
		k.SetLockIdIntermediaryAccountConnection(ctx, connection.LockId, intermediaryAcc)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:                        k.GetParams(ctx),
		SuperfluidAssets:              k.GetAllSuperfluidAssets(ctx),
		OsmoEquivalentMultipliers:     k.GetAllOsmoEquivalentMultipliers(ctx),
		IntermediaryAccounts:          k.GetAllIntermediaryAccounts(ctx),
		IntemediaryAccountConnections: k.GetAllLockIdIntermediaryAccountConnections(ctx),
	}
}
