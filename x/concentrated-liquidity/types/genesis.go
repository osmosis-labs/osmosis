package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// DefaultGenesis returns the default GenesisState for the concentrated-liquidity module.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Pools:  []*codectypes.Any{},
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return nil
}
