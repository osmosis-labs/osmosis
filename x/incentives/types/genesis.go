package types

import "errors"

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: Params{
			DistrEpochIdentifier: "weekly",
		},
		Pots: []Pot{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if gs.Params.DistrEpochIdentifier == "" {
		return errors.New("epoch identifier should NOT be empty")
	}
	return nil
}
