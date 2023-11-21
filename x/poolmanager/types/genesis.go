package types

import "errors"

// DefaultGenesis returns the default poolmanager genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		NextPoolId: 1,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if gs.NextPoolId == 0 {
		return errors.New("next pool id cannot be 0")
	}
	return nil
}
