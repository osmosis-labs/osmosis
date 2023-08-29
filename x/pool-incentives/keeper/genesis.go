package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v19/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
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
			if record.Duration == 0 {
				k.SetPoolGaugeIdNoLock(ctx, record.PoolId, record.GaugeId)
			} else {
				if err := k.SetPoolGaugeIdInternalIncentive(ctx, record.PoolId, record.Duration, record.GaugeId); err != nil {
					panic(err)
				}
			}
		}
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	distrInfo := k.GetDistrInfo(ctx)
	lastPoolId := k.poolmanagerKeeper.GetNextPoolId(ctx)
	lockableDurations := k.GetLockableDurations(ctx)
	var poolToGauges types.PoolToGauges
	for poolId := 1; poolId < int(lastPoolId); poolId++ {
		pool, err := k.poolmanagerKeeper.GetPool(ctx, uint64(poolId))
		if err != nil {
			panic(err)
		}
		isCLPool := pool.GetType() == poolmanagertypes.Concentrated
		if isCLPool {
			incParams := k.incentivesKeeper.GetEpochInfo(ctx)
			gaugeID, err := k.GetPoolGaugeId(ctx, uint64(poolId), incParams.Duration)
			if err != nil {
				panic(err)
			}
			var poolToGauge types.PoolToGauge
			poolToGauge.Duration = incParams.Duration
			poolToGauge.GaugeId = gaugeID
			poolToGauge.PoolId = uint64(poolId)
			poolToGauges.PoolToGauge = append(poolToGauges.PoolToGauge, poolToGauge)
		} else {
			for _, duration := range lockableDurations {
				gaugeID, err := k.GetPoolGaugeId(ctx, uint64(poolId), duration)
				if err != nil {
					// TODO: This error happens on pool export for CosmWasm
					// assocated pools, to fix this we need to assign
					// a gauge to cosmwasm pools on creation

					ctx.Logger().Error(err.Error())
					// panic(err)
				}
				var poolToGauge types.PoolToGauge
				poolToGauge.Duration = duration
				poolToGauge.GaugeId = gaugeID
				poolToGauge.PoolId = uint64(poolId)
				poolToGauges.PoolToGauge = append(poolToGauges.PoolToGauge, poolToGauge)
			}
		}
	}

	return &types.GenesisState{
		Params:            k.GetParams(ctx),
		LockableDurations: k.GetLockableDurations(ctx),
		DistrInfo:         &distrInfo,
		PoolToGauges:      &poolToGauges,
	}
}
