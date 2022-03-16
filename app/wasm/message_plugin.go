package wasm

import (
	"encoding/json"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	wasm "github.com/osmosis-labs/osmosis/v7/app/wasm/bindings"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
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
			return buildSwapMsg(sender, contractMsg.Swap)
		}
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown osmosis query variant"}
	}
}

func buildSwapMsg(sender sdk.AccAddress, swap *wasm.SwapMsg) ([]sdk.Msg, error) {
	if len(swap.Route) != 0 {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "TODO: multi-hop swaps"}
	}
	if swap.Amount.ExactIn != nil {
		routes := []gammtypes.SwapAmountInRoute{{
			PoolId:        swap.First.PoolId,
			TokenOutDenom: swap.First.DenomOut,
		}}
		msg := gammtypes.MsgSwapExactAmountIn{
			Sender: sender.String(),
			Routes: routes,
			TokenIn: sdk.Coin{
				Denom:  swap.First.DenomIn,
				Amount: swap.Amount.ExactIn.Input,
			},
			TokenOutMinAmount: swap.Amount.ExactIn.MinOutput,
		}
		return []sdk.Msg{&msg}, nil
	} else if swap.Amount.ExactOut != nil {
		routes := []gammtypes.SwapAmountOutRoute{{
			PoolId:       swap.First.PoolId,
			TokenInDenom: swap.First.DenomIn,
		}}
		msg := gammtypes.MsgSwapExactAmountOut{
			Sender:           sender.String(),
			Routes:           routes,
			TokenInMaxAmount: swap.Amount.ExactOut.MaxInput,
			TokenOut: sdk.Coin{
				Denom:  swap.First.DenomOut,
				Amount: swap.Amount.ExactOut.Output,
			},
		}
		return []sdk.Msg{&msg}, nil
	} else {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "must support either Swap.ExactIn or Swap.ExactOut"}
	}
}
