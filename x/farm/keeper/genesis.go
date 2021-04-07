package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/farm/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	if genState.Farms != nil {
		for _, farm := range genState.Farms {
			err := k.setFarm(ctx, farm)
			if err != nil {
				panic(err)
			}
		}
	}

	if genState.Farmers != nil {
		for _, farmer := range genState.Farmers {
			k.setFarmer(ctx, farmer)
		}
	}

	if genState.HistoricalEntries != nil {
		for _, entry := range genState.HistoricalEntries {
			k.SetHistoricalEntry(ctx, entry.FarmId, entry.CurrentPeriod, entry.Entry)
		}
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	farms := make([]types.Farm, 0)
	k.IterateFarms(ctx, func(farm types.Farm) bool {
		farms = append(farms, farm)
		return false
	})

	farmers := make([]types.Farmer, 0)
	k.IterateFarmers(ctx, func(farmer types.Farmer) bool {
		farmers = append(farmers, farmer)
		return false
	})

	historicalEntries := make([]types.GenesisHistoricalEntry, 0)
	k.IterateHistoricalEntries(ctx, func(entry types.HistoricalEntry, farmId uint64, period uint64) bool {
		historicalEntries = append(historicalEntries, types.GenesisHistoricalEntry{
			Entry:         entry,
			FarmId:        farmId,
			CurrentPeriod: period,
		})
		return false
	})

	return &types.GenesisState{
		Farms:             farms,
		Farmers:           farmers,
		HistoricalEntries: historicalEntries,
	}
}
