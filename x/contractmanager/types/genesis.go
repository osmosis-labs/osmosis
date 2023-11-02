package types

import (
	"fmt"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		FailuresList: []Failure{},
		Params:       DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in failure
	failureIndexMap := make(map[string]struct{})

	for _, elem := range gs.FailuresList {
		index := string(GetFailureKey(elem.Address, elem.Id))
		if _, ok := failureIndexMap[index]; ok {
			return fmt.Errorf("duplicated address for failure")
		}
		failureIndexMap[index] = struct{}{}
	}

	return gs.Params.Validate()
}
