package genesis

import (
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

// DefaultGenesis returns the default GenesisState for the concentrated-liquidity module.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		PoolData: []PoolData{},
		Params:   types.DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return nil
}
