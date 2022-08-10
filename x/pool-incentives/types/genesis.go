package types

import (
	"encoding/json"
	"errors"
	"fmt"
	time "time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewGenesisState(params Params, lockableDurations []time.Duration, distrInfo *DistrInfo) *GenesisState {
	return &GenesisState{
		Params:            params,
		LockableDurations: lockableDurations,
		DistrInfo:         distrInfo,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		LockableDurations: []time.Duration{
			time.Hour,
			time.Hour * 3,
			time.Hour * 7,
		},
		DistrInfo: &DistrInfo{
			TotalWeight: sdk.ZeroInt(),
			Records:     nil,
		},
	}
}

// GetGenesisStateFromAppState returns x/pool-yield GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

// ValidateGenesis validates the provided pool-yield genesis state to ensure the
// expected invariants holds. (i.e. params in correct bounds).
func ValidateGenesis(data *GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	if data.DistrInfo.TotalWeight.LT(sdk.NewInt(0)) {
		return errors.New("distrinfo weight should not be negative")
	}

	return validateLockableDurations(data.LockableDurations)
}

// validateLockableDurations ensures the provided lockable duration is a valid parameter.
func validateLockableDurations(i interface{}) error {
	_, ok := i.([]time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
