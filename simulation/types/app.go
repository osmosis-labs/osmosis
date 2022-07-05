package simulation

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
)

type App interface {
	*baseapp.BaseApp
	// The genesis state of the blockchain is represented here as a map of raw json
	// messages key'd by a identifier string.
	// The identifier is used to determine which module genesis information belongs
	// to so it may be appropriately routed during init chain.
	// Within this application default genesis information is retrieved from
	// the ModuleBasicManager which populates json from each BasicModule
	// object provided to it during init.
	NewDefaultGenesisState() map[string]json.RawMessage
	Cdc() codec.JSONCodec
}
