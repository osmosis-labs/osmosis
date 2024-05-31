package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(sPoolDelta sdk.Dec, params Params) *GenesisState {
	return &GenesisState{
		OsmosisPoolDelta: sPoolDelta,
		Params:           params,
	}
}

// DefaultGenesisState returns raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		OsmosisPoolDelta: sdk.ZeroDec(),
		Params:           DefaultParams(),
	}
}

// ValidateGenesis validates the provided market genesis state
func ValidateGenesis(data *GenesisState) error {
	return data.Params.Validate()
}

// GetGenesisStateFromAppState returns x/market GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}
