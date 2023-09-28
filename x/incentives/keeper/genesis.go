package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v19/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
)

// InitGenesis initializes the incentives module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	k.SetLockableDurations(ctx, genState.LockableDurations)

	for _, gauge := range genState.Gauges {
		gauge := gauge
		if gauge.DistributeTo.LockQueryType == lockuptypes.ByGroup {
			// set gauge directly for byGroup gauges
			err := k.setGauge(ctx, &gauge)
			if err != nil {
				panic(err)
			}
		} else {
			// set gauge refs for all non-byGroup gauges
			err := k.SetGaugeWithRefKey(ctx, &gauge)
			if err != nil {
				panic(err)
			}
		}
	}

	k.SetLastGaugeID(ctx, genState.LastGaugeId)

	for _, group := range genState.Groups {
		k.SetGroup(ctx, group)
	}
}

// ExportGenesis returns the x/incentives module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:            k.GetParams(ctx),
		LockableDurations: k.GetLockableDurations(ctx),
		Gauges:            k.GetNotFinishedGauges(ctx),
		LastGaugeId:       k.GetLastGaugeID(ctx),
		Groups:            k.GetAllGroups(ctx),
	}
}
