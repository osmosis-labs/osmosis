package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// DefaultGenesis creates a default GenesisState object.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Pools:            []*codectypes.Any{},
		NextPoolNumber:   1,
		Params:           DefaultParams(),
		MigrationRecords: &MigrationRecords{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return nil
}
