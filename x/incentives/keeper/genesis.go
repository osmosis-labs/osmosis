package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
)

// InitGenesis initializes the incentives module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	k.SetLockableDurations(ctx, genState.LockableDurations)

	for _, gauge := range genState.Gauges {
		gauge := gauge
		// set gauge refs for all non-byGroup gauges
		err := k.SetGaugeWithRefKey(ctx, &gauge)
		if err != nil {
			panic(err)
		}
	}

	for _, groupGauges := range genState.GroupGauges {
		groupGauges := groupGauges
		err := k.setGauge(ctx, &groupGauges)
		if err != nil {
			panic(err)
		}
	}

	k.SetLastGaugeID(ctx, genState.LastGaugeId)

	for _, group := range genState.Groups {
		k.SetGroup(ctx, group)
	}
}

// ExportGenesis returns the x/incentives module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	groupGauges, err := k.GetAllGroupsGauges(ctx)
	if err != nil {
		panic(err)
	}

	groups, err := k.GetAllGroups(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Params:            k.GetParams(ctx),
		LockableDurations: k.GetLockableDurations(ctx),
		Gauges:            k.GetNotFinishedGauges(ctx),
		LastGaugeId:       k.GetLastGaugeID(ctx),
		GroupGauges:       groupGauges,
		Groups:            groups,
	}
}
