package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/pool-incentives/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	k.SetLockableDurations(ctx, genState.LockableDurations)
	if genState.DistrInfo == nil {
		k.SetDistrInfo(ctx, types.DistrInfo{
			TotalWeight: sdk.NewInt(0),
			Records:     nil,
		})
	} else {
		k.SetDistrInfo(ctx, *genState.DistrInfo)
	}
	if genState.PoolToGauges != nil {
		for _, record := range genState.PoolToGauges.PoolToGauge {
			k.SetPoolGaugeId(ctx, record.PoolId, record.Duration, record.GaugeId)
		}
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	distrInfo := k.GetDistrInfo(ctx)
	lastPoolId := k.poolmanagerKeeper.GetNextPoolId(ctx)
	lockableDurations := k.GetLockableDurations(ctx)
	var poolToGauges types.PoolToGauges
	for i := 1; i < int(lastPoolId); i++ {
		for _, duration := range lockableDurations {
			gaugeID, err := k.GetPoolGaugeId(ctx, uint64(i), duration)
			if err != nil {
				panic(err)
			}
			var poolToGauge types.PoolToGauge
			poolToGauge.Duration = duration
			poolToGauge.GaugeId = gaugeID
			poolToGauge.PoolId = uint64(i)
			poolToGauges.PoolToGauge = append(poolToGauges.PoolToGauge, poolToGauge)
		}
	}

	return &types.GenesisState{
		Params:            k.GetParams(ctx),
		LockableDurations: k.GetLockableDurations(ctx),
		DistrInfo:         &distrInfo,
		PoolToGauges:      &poolToGauges,
	}
}
