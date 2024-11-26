package keeper

import (
	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if genState.Params != nil {
		k.SetParams(ctx, *genState.Params)
	} else {
		k.SetParams(ctx, types.DefaultParams())
	}
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
	params := k.GetParams(ctx)
	return &types.GenesisState{
		LastLockId:     k.GetLastLockID(ctx),
		Locks:          locks,
		SyntheticLocks: k.GetAllSyntheticLockups(ctx),
		Params:         &params,
	}
}
