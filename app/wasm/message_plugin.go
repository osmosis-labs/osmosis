package wasm

import (
	"encoding/json"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	wasm "github.com/osmosis-labs/osmosis/v7/app/wasm/bindings"
)

func CustomEncoder(osmoKeeper *QueryPlugin) func(sender sdk.AccAddress, msg json.RawMessage) ([]sdk.Msg, error) {
	return func(sender sdk.AccAddress, msg json.RawMessage) ([]sdk.Msg, error) {
		var contractMsg wasm.OsmosisMsg
		if err := json.Unmarshal(msg, &contractMsg); err != nil {
			return nil, sdkerrors.Wrap(err, "osmosis msg")
		}

		if contractMsg.MintTokens != nil {
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "not implemented: mint tokens"}
		}
		if contractMsg.Swap != nil {
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "not implemented: swap"}
		}
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown osmosis query variant"}
	}
}
