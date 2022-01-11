package incentives

import (
	"fmt"

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
		duration := currentReward.LockDuration
		err := k.SetCurrentReward(ctx, currentReward, denom, duration)
		if err != nil {
			panic(err)
		}
		for _, historicalReward := range genesisReward.HistoricalReward {
			err := k.SetHistoricalReward(ctx, historicalReward.CumulativeRewardRatio, denom, duration, historicalReward.Period, int64(historicalReward.LastEligibleEpoch))
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
		GenesisReward:     GetGenesisRewards(ctx, k),
		PeriodLockReward:  k.GetAllPeriodLockReward(ctx),
	}
}

func GetGenesisRewards(ctx sdk.Context, k keeper.Keeper) []types.GenesisReward {
	var genesisRewards []types.GenesisReward
	for _, currentReward := range k.GetAllCurrentReward(ctx) {
		denom := currentReward.Denom
		duration := currentReward.LockDuration
		genesisReward := types.GenesisReward{}
		genesisReward.CurrentReward = currentReward
		var historicalRewards []types.HistoricalReward
		for i := uint64(1); i < currentReward.Period; i++ {
			historicalReward, err := k.GetHistoricalReward(ctx, denom, duration, i)
			if err != nil {
				panic(fmt.Sprintf("unable to retrieve historical reward for denom(%v) d(%v) period(%v)", denom, duration, i))
			}
			historicalRewards = append(historicalRewards, historicalReward)
		}
		genesisReward.HistoricalReward = historicalRewards
		genesisRewards = append(genesisRewards, genesisReward)
	}
	return genesisRewards
}
