package types

import (
	"encoding/json"
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: Params{
			RefreshEpochIdentifier: "day",
		},
		SuperfluidAssets:     []SuperfluidAsset{},
		SuperfluidAssetInfos: []SuperfluidAssetInfo{},
	}
}

// GetGenesisStateFromAppState returns x/incentives GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONMarshaler, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if gs.Params.RefreshEpochIdentifier == "" {
		return errors.New("refresh identifier should NOT be empty")
	}
	return nil
}
