package incentives

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/incentives/keeper"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	k.SetLockableDurations(ctx, genState.LockableDurations)
	for _, gauge := range genState.Gauges {
		k.SetGaugeWithRefKey(ctx, &gauge)
	}
	for _, autostaking := range genState.Autostakings {
		k.SetAutostaking(ctx, &autostaking)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:            k.GetParams(ctx),
		LockableDurations: k.GetLockableDurations(ctx),
		Gauges:            k.GetNotFinishedGauges(ctx),
		Autostakings:      k.AllAutoStakings(ctx),
	}
}
