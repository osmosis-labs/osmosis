package app

import (
	"encoding/json"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// The genesis state of the blockchain is represented here as a map of raw json
// messages key'd by a identifier string.
// The identifier is used to determine which module genesis information belongs
// to so it may be appropriately routed during init chain.
// Within this application default genesis information is retrieved from
// the ModuleBasicManager which populates json from each BasicModule
// object provided to it during init.
type GenesisState map[string]json.RawMessage

const (
	// DefaultMaxWasmCodeSize limit max bytes read to prevent gzip bombs
	// 600 KB is copied from x/wasm, but you can customize here as desired
	DefaultMaxWasmCodeSize = 600 * 1024 * 2
)

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState() GenesisState {
	encCfg := MakeEncodingConfig()
	gen := ModuleBasics.DefaultGenesis(encCfg.Marshaler)

	// here we override wasm config to make it permissioned by default
	wasmGen := wasm.GenesisState{
		Params: wasmtypes.Params{
			CodeUploadAccess:             wasmtypes.AllowNobody,
			InstantiateDefaultPermission: wasmtypes.AccessTypeEverybody,
			MaxWasmCodeSize:              DefaultMaxWasmCodeSize,
		},
	}
	gen[wasm.ModuleName] = encCfg.Marshaler.MustMarshalJSON(&wasmGen)
	return gen
}
