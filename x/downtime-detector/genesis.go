package downtimedetector

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/downtime-detector/types"
)

func (k *Keeper) InitGenesis(ctx sdk.Context, gen *types.GenesisState) {
	k.StoreLastBlockTime(ctx, gen.LastBlockTime)
	// set all default genesis down times, in case the provided list in genesis misses some.
	k.setGenDowntimes(ctx, types.DefaultGenesis().GetDowntimes())
	// override with genesis list
	k.setGenDowntimes(ctx, gen.Downtimes)
}

func (k *Keeper) setGenDowntimes(ctx sdk.Context, genDowntimes []types.GenesisDowntimeEntry) {
	for _, downtime := range genDowntimes {
		k.StoreLastDowntimeOfLength(ctx, downtime.Duration, downtime.LastDowntime)
	}
}

// ExportGenesis returns the downtime detector module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	t, err := k.GetLastBlockTime(ctx)
	if err != nil {
		panic(err)
	}
	return &types.GenesisState{
		Downtimes:     k.getGenDowntimes(ctx),
		LastBlockTime: t,
	}
}

func (k *Keeper) getGenDowntimes(ctx sdk.Context) []types.GenesisDowntimeEntry {
	downtimes := []types.GenesisDowntimeEntry{}
	for _, downtime := range types.DowntimeToDuration.Keys() {
		t, err := k.GetLastDowntimeOfLength(ctx, downtime)
		if err != nil {
			panic(err)
		}
		downtimes = append(downtimes, types.GenesisDowntimeEntry{
			Duration:     downtime,
			LastDowntime: t,
		})
	}
	return downtimes
}
