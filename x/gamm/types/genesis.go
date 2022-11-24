package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// DefaultGenesis creates a default GenesisState object.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Pools:          []*codectypes.Any{},
		NextPoolNumber: 1,
	}
}
