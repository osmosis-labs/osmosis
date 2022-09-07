package keeper

import (
	"github.com/osmosis-labs/osmosis/v12/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetLastLockID(ctx, genState.LastLockId)
	if err := k.InitializeAllLocks(ctx, genState.Locks); err != nil {
		return
	}
	if err := k.InitializeAllSyntheticLocks(ctx, genState.SyntheticLocks); err != nil {
		return
	}
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	locks, err := k.GetPeriodLocks(ctx)
	if err != nil {
		panic(err)
	}
	return &types.GenesisState{
		LastLockId:     k.GetLastLockID(ctx),
		Locks:          locks,
		SyntheticLocks: k.GetAllSyntheticLockups(ctx),
	}
}
