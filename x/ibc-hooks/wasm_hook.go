package ibc_hooks

import (
	"encoding/json"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/osmosis-labs/osmosis/v12/x/ibc-hooks/types"
)

type ContractAck struct {
	ContractResult []byte `json:"contract_result"`
	IbcAck         []byte `json:"ibc_ack"`
}

type WasmHooks struct {
	ContractKeeper *wasmkeeper.PermissionedKeeper
}

func NewWasmHooks(contractKeeper *wasmkeeper.PermissionedKeeper) WasmHooks {
	return WasmHooks{ContractKeeper: contractKeeper}
}

func (h WasmHooks) ExecuteContract(ctx sdk.Context, contract string, msg []byte, caller sdk.AccAddress, data transfertypes.FungibleTokenPacketData) ([]byte, error) {
	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return nil, err
	}

	result, err := h.ContractKeeper.Execute(ctx, contractAddr, caller, msg, sdk.NewCoins())
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (h WasmHooks) OnRecvPacketOverride(im IBCMiddleware, ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) ibcexported.Acknowledgement {
	if h.ContractKeeper == nil {
		// Not configured
		return im.App.OnRecvPacket(ctx, packet, relayer)
	}

	var data transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &data); err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf("cannot unmarshal sent packet data: %s", err.Error()))
	}

	// If there is no memo, the packet was either sent with an earlier version of IBC, or the memo was
	// intentionally left blank. Nothing to do here. Ignore the packet and pass it down the stack.
	memo := data.GetMemo()
	if len(memo) == 0 {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	}

	// Start of validation of the memo for this hook

	// the metadata must be a valid JSON object
	var metadata map[string]interface{}
	err := json.Unmarshal([]byte(memo), &metadata)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadPacketMetadataMsg, metadata, err.Error()))
	}

	// If the key "wasm"  doesn't exist, there's nothing to do on this hook. Continue by passing the packet
	// down the stack
	wasmRaw, ok := metadata["wasm"]
	if !ok {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	}

	// Make sure the wasm key is a map. If it isn't, ignore this packet
	wasm, ok := wasmRaw.(map[string]interface{})
	if !ok {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	}

	// Get the contract
	contract, ok := wasm["contract"].(string)
	if !ok {
		// The tokens will be returned
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadMetadataFormatMsg, memo, `Could not find key "contract"`))
	}

	// Ensure the message key is provided
	if wasm["msg"] == nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadMetadataFormatMsg, memo, `Could not find key "msg"`))
	}

	// Make sure the msg key is a map. If it isn't, return an error
	_, ok = wasm["msg"].(map[string]interface{})
	if !ok {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadMetadataFormatMsg, memo, `wasm["msg"] is not an object`))
	}

	// Get the message string by serializing the map
	msg, err := json.Marshal(wasm["msg"])
	if err != nil {
		// The tokens will be returned
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadMetadataFormatMsg, memo, err.Error()))
	}
	// End of validation of the memo for this hook

	// Execute the receive
	ack := im.App.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() { // ToDO: Fix this with the proper ack handling
		return ack
	}
	caller, _ := sdk.AccAddressFromBech32(data.Sender)
	result, err := h.ExecuteContract(ctx, contract, msg, caller, data)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadExecutionMsg, err.Error()))
	}

	fullAck := ContractAck{ContractResult: result, IbcAck: ack.Acknowledgement()}
	bz, err := json.Marshal(fullAck)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadResponse, err.Error()))
	}

	return channeltypes.NewResultAcknowledgement(bz)
}
