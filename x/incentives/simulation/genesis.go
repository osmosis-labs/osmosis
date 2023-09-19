package simulation

// DONTCOVER

import (
	"time"

	"github.com/cosmos/cosmos-sdk/types/module"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v19/x/incentives/types"
)

// Simulation parameter constants.
const (
	ParamsDistrEpochIdentifier = "distr_epoch_identifier"
)

// RandomizedGenState generates a random GenesisState for the incentives module.
func RandomizedGenState(simState *module.SimulationState) {
	// TODO: Make this read off of what mint set for its genesis value.
	distrEpochIdentifier := "day"
	createGaugeFee := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(50000000)))
	addToGaugeFee := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(25000000)))

	incentivesGenesis := types.GenesisState{
		Params: types.Params{
			DistrEpochIdentifier: distrEpochIdentifier,
			CreateGaugeFee:       createGaugeFee,
			AddToGaugeFee:        addToGaugeFee,
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
