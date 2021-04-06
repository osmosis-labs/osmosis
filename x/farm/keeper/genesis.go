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
			k.setFarmer(ctx, &farmer)
		}
	}

	if genState.HistoricalRecords != nil {
		for _, record := range genState.HistoricalRecords {
			k.SetHistoricalRecord(ctx, record.FarmId, record.CurrentPeriod, record.Record)
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

	historicalRecords := make([]types.GenesisHistoricalRecord, 0)
	k.IterateHistoricalRecords(ctx, func(record types.HistoricalRecord, farmId uint64, period uint64) bool {
		historicalRecords = append(historicalRecords, types.GenesisHistoricalRecord{
			Record:        record,
			FarmId:        farmId,
			CurrentPeriod: period,
		})
		return false
	})

	return &types.GenesisState{
		Farms:             farms,
		Farmers:           farmers,
		HistoricalRecords: historicalRecords,
	}
}
