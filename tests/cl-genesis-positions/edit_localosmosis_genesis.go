package main

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	osmosisApp "github.com/osmosis-labs/osmosis/v15/app"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	clgenesis "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/genesis"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

func EditLocalOsmosisGenesis(updatedCLGenesis *clgenesis.GenesisState) {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(localOsmosisHomePath)
	config.Moniker = "localosmosis"

	genFile := config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		panic(err)
	}

	encodingConfig := osmosisApp.MakeEncodingConfig()
	cdc := encodingConfig.Marshaler

	var localOsmosisCLGenesis clgenesis.GenesisState
	cdc.MustUnmarshalJSON(appState[cltypes.ModuleName], &localOsmosisCLGenesis)

	var localOsmosisPoolManagerGenesis poolmanagertypes.GenesisState
	cdc.MustUnmarshalJSON(appState[poolmanagertypes.ModuleName], &localOsmosisPoolManagerGenesis)

	nextPoolId := localOsmosisPoolManagerGenesis.NextPoolId
	localOsmosisPoolManagerGenesis.NextPoolId = nextPoolId + 1
	localOsmosisPoolManagerGenesis.PoolRoutes = append(localOsmosisPoolManagerGenesis.PoolRoutes, poolmanagertypes.ModuleRoute{
		PoolType: poolmanagertypes.Concentrated,
		PoolId:   nextPoolId,
	})
	appState[poolmanagertypes.ModuleName] = cdc.MustMarshalJSON(&localOsmosisPoolManagerGenesis)

	// Copy positions
	for _, position := range updatedCLGenesis.Positions {
		position.PoolId = nextPoolId
		localOsmosisCLGenesis.Positions = append(localOsmosisCLGenesis.Positions, position)
	}

	// Copy pool state, including ticks, incentive accums, records, and fee accumulators
	for _, pool := range updatedCLGenesis.PoolData {
		poolAny := pool.Pool

		var clPoolExt cltypes.ConcentratedPoolExtension
		err := cdc.UnpackAny(poolAny, &clPoolExt)
		if err != nil {
			panic(err)
		}

		clPool := clPoolExt.(*model.Pool)
		clPool.Id = nextPoolId

		any, err := codectypes.NewAnyWithValue(clPool)
		if err != nil {
			panic(err)
		}
		anyCopy := *any

		for i := range pool.Ticks {
			pool.Ticks[i].PoolId = nextPoolId
		}

		for i := range pool.IncentiveRecords {
			pool.IncentiveRecords[i].PoolId = nextPoolId
		}

		for i := range pool.IncentivesAccumulators {
			pool.IncentivesAccumulators[i].Name = types.KeyUptimeAccumulator(nextPoolId, uint64(i))
		}

		updatedPoolData := clgenesis.PoolData{
			Pool:                   &anyCopy,
			Ticks:                  pool.Ticks,
			IncentivesAccumulators: pool.IncentivesAccumulators,
			IncentiveRecords:       pool.IncentiveRecords,
			FeeAccumulator: clgenesis.AccumObject{
				Name:         types.KeyFeePoolAccumulator(nextPoolId),
				AccumContent: pool.FeeAccumulator.AccumContent,
			},
		}

		localOsmosisCLGenesis.PoolData = append(localOsmosisCLGenesis.PoolData, updatedPoolData)
	}

	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		panic(err)
	}

	genDoc.AppState = appStateJSON

	if err := genutil.ExportGenesisFile(genDoc, localOsmosisHomePath); err != nil {
		panic(err)
	}
}
