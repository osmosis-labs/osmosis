package types

import (
	"encoding/json"
	"fmt"
	time "time"

	"github.com/cosmos/cosmos-sdk/codec"
)

func NewGenesisState(lockableDurations []time.Duration) *GenesisState {
	return &GenesisState{
		LockableDurations: lockableDurations,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		LockableDurations: []time.Duration{
			time.Hour,
			time.Hour * 3,
			time.Hour * 7,
		},
	}
}

// GetGenesisStateFromAppState returns x/pool-yield GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONMarshaler, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

// ValidateGenesis validates the provided pool-yield genesis state to ensure the
// expected invariants holds. (i.e. params in correct bounds)
func ValidateGenesis(data *GenesisState) error {
	return validateLockableDurations(data.LockableDurations)
}

func validateLockableDurations(i interface{}) error {
	_, ok := i.([]time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
