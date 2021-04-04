package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

func NewGenesisState(mintDenom string, lockableDurations []time.Duration) *GenesisState {
	return &GenesisState{
		MintDenom:         mintDenom,
		LockableDurations: lockableDurations,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		MintDenom: "osmo",
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
	err := validateMintDenom(data.MintDenom)
	if err != nil {
		return err
	}
	return validateLockableDurations(data.LockableDurations)
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateLockableDurations(i interface{}) error {
	_, ok := i.([]time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
