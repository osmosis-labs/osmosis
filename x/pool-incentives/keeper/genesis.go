package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	k.SetLockableDurations(ctx, genState.LockableDurations)
	if genState.DistrInfo == nil {
		k.SetDistrInfo(ctx, types.DistrInfo{
			TotalWeight: osmomath.NewInt(0),
			Records:     nil,
		})
	} else {
		k.SetDistrInfo(ctx, *genState.DistrInfo)
	}
	if genState.AnyPoolToInternalGauges != nil {
		for _, record := range genState.AnyPoolToInternalGauges.PoolToGauge {
			if err := k.SetPoolGaugeIdInternalIncentive(ctx, record.PoolId, record.Duration, record.GaugeId); err != nil {
				panic(err)
			}
		}
	}
	if genState.ConcentratedPoolToNoLockGauges != nil {
		for _, record := range genState.ConcentratedPoolToNoLockGauges.PoolToGauge {
			k.SetPoolGaugeIdNoLock(ctx, record.PoolId, record.GaugeId, record.Duration)
		}
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	distrInfo := k.GetDistrInfo(ctx)
	lastPoolId := k.poolmanagerKeeper.GetNextPoolId(ctx)
	lockableDurations := k.GetLockableDurations(ctx)
	var (
		anyPoolToInternalGauges        types.AnyPoolToInternalGauges
		concentratedPoolToNoLockGauges types.ConcentratedPoolToNoLockGauges
	)
	for poolId := 1; poolId < int(lastPoolId); poolId++ {
		pool, err := k.poolmanagerKeeper.GetPool(ctx, uint64(poolId))
		if err != nil {
			panic(err)
		}

		if pool.GetType() == poolmanagertypes.CosmWasm {
			// TODO: remove this post-v19. In v19 we did not create a hook for cw pool gauges.
			// Fix tracked in:
			// https://github.com/osmosis-labs/osmosis/issues/6122
			continue
		}

		isCLPool := pool.GetType() == poolmanagertypes.Concentrated
		if isCLPool {
			// This creates a link for the internal pool gauge.
			// Every CL pool has one such gauge.
			incParams := k.incentivesKeeper.GetEpochInfo(ctx)
			gaugeID, err := k.GetPoolGaugeId(ctx, uint64(poolId), incParams.Duration)
			if err != nil {
				panic(err)
			}
			var poolToGauge types.PoolToGauge
			poolToGauge.Duration = incParams.Duration
			poolToGauge.GaugeId = gaugeID
			poolToGauge.PoolId = uint64(poolId)
			anyPoolToInternalGauges.PoolToGauge = append(anyPoolToInternalGauges.PoolToGauge, poolToGauge)

			// All concentrated pools need an additional link for the no-lock gauge.
			gaugeIDs, err := k.GetNoLockGaugeIdsFromPool(ctx, uint64(poolId))
			if err != nil {
				panic(err)
			}
			for _, gaugeID := range gaugeIDs {
				poolToGauge := types.PoolToGauge{
					GaugeId: gaugeID,
					PoolId:  uint64(poolId),
				}
				concentratedPoolToNoLockGauges.PoolToGauge = append(concentratedPoolToNoLockGauges.PoolToGauge, poolToGauge)
			}
		} else {
			for _, duration := range lockableDurations {
				gaugeID, err := k.GetPoolGaugeId(ctx, uint64(poolId), duration)
				if err != nil {
					panic(err)
				}
				var poolToGauge types.PoolToGauge
				poolToGauge.Duration = duration
				poolToGauge.GaugeId = gaugeID
				poolToGauge.PoolId = uint64(poolId)
				anyPoolToInternalGauges.PoolToGauge = append(anyPoolToInternalGauges.PoolToGauge, poolToGauge)
			}
		}
	}

	return &types.GenesisState{
		Params:                         k.GetParams(ctx),
		LockableDurations:              k.GetLockableDurations(ctx),
		DistrInfo:                      &distrInfo,
		AnyPoolToInternalGauges:        &anyPoolToInternalGauges,
		ConcentratedPoolToNoLockGauges: &concentratedPoolToNoLockGauges,
	}
}
