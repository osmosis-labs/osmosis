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
		err := k.SetGaugeWithRefKey(ctx, &gauge)
		if err != nil {
			panic(err)
		}
	}
	for _, genesisReward := range genState.GenesisReward {
		currentReward := genesisReward.CurrentReward
		denom := currentReward.Denom
		duration := currentReward.Duration
		err := k.SetCurrentReward(ctx, currentReward, denom, duration)
		if err != nil {
			panic(err)
		}
		for _, historicalReward := range genesisReward.HistoricalReward {
			err := k.AddHistoricalReward(ctx, historicalReward, denom, duration, historicalReward.Period, int64(historicalReward.LastEligibleEpoch))
			if err != nil {
				panic(err)
			}
		}
	}
	for _, periodLockReward := range genState.PeriodLockReward {
		err := k.SetPeriodLockReward(ctx, periodLockReward)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:            k.GetParams(ctx),
		LockableDurations: k.GetLockableDurations(ctx),
		Gauges:            k.GetNotFinishedGauges(ctx),
		GenesisReward:     k.GetGenesisRewards(ctx),
		PeriodLockReward:  k.GetAllPeriodLockReward(ctx),
	}
}
