package simulation

// DONTCOVER

import (
	"time"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
)

// Simulation parameter constants.
const (
	ParamsDistrEpochIdentifier = "distr_epoch_identifier"
)

// RandomizedGenState generates a random GenesisState for the incentives module.
func RandomizedGenState(simState *module.SimulationState) {
	// TODO: Make this read off of what mint set for its genesis value.
	distrEpochIdentifier := "day"

	incentivesGenesis := types.GenesisState{
		Params: types.Params{
			DistrEpochIdentifier: distrEpochIdentifier,
		},
		LockableDurations: []time.Duration{
			time.Second,
			time.Hour,
			time.Hour * 3,
			time.Hour * 7,
		},
	}

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&incentivesGenesis)
}
