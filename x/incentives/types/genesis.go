package types

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: Params{
			DistrEpochIdentifier: "week",
			MinAutostakingRate:   sdk.NewDecWithPrec(5, 1),
		},
		Gauges: []Gauge{},
		LockableDurations: []time.Duration{
			time.Second,
			time.Hour,
			time.Hour * 3,
			time.Hour * 7,
		},
	}
}

// GetGenesisStateFromAppState returns x/incentives GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONMarshaler, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if gs.Params.DistrEpochIdentifier == "" {
		return errors.New("epoch identifier should NOT be empty")
	}
	if gs.Params.MinAutostakingRate.IsNil() {
		return errors.New("MinAutostakingRate is nil")
	}
	if gs.Params.MinAutostakingRate.LT(sdk.ZeroDec()) {
		return errors.New("MinAutostakingRate should NOT be negative")
	}
	if gs.Params.MinAutostakingRate.GT(sdk.OneDec()) {
		return errors.New("MinAutostakingRate should NOT be bigger than 1")
	}
	return nil
}
