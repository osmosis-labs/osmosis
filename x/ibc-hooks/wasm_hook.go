package ibc_hooks

import (
	"encoding/json"
	"fmt"

	"github.com/osmosis-labs/osmosis/v12/osmoutils"

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

func (h WasmHooks) OnRecvPacketOverride(im IBCMiddleware, ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) ibcexported.Acknowledgement {
	if h.ContractKeeper == nil {
		// Not configured
		return im.App.OnRecvPacket(ctx, packet, relayer)
	}

	var data transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &data); err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf("cannot unmarshal sent packet data: %s", err.Error()))
	}

	// Validate the memo
	shouldBreak, errStr, contractAddr, msgBytes := ValidateMemo(data.GetMemo(), data.Receiver)
	if shouldBreak {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	}
	if errStr != "" {
		return channeltypes.NewErrorAcknowledgement(errStr)
	}
	if msgBytes == nil || contractAddr == nil { // This should never happen
		return channeltypes.NewErrorAcknowledgement("error in wasmhook message validation")
	}

	// The funds sent on this packet need to be transferred to the wasm hooks module address/
	// For this, we override the ICS20 packet's Receiver (essentially hijacking the funds for the module)
	// and execute the underlying OnRecvPacket() call (which should eventually land on the transfer app's
	// relay.go and send the sunds to the module.
	//
	// If that succeeds, we make the contract call
	data.Receiver = WasmHookModuleAccountAddr.String()
	bz, err := json.Marshal(data)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf("cannot marshal the ICS20 packet: %s", err.Error()))
	}
	packet.Data = bz

	// Execute the receive
	ack := im.App.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() {
		return ack
	}

	amount, ok := sdk.NewIntFromString(data.GetAmount())
	if !ok {
		// This should never happen, as it should've been caught in the underlaying call to OnRecvPacket,
		// but returning here for completeness
		return channeltypes.NewErrorAcknowledgement("Invalid packet data: Amount is not an int")
	}

	// The packet's denom is the denom in the sender chain. This needs to be converted to the local denom.
	denom := osmoutils.MustExtractDenomFromPacketOnRecv(packet)
	funds := sdk.NewCoins(sdk.NewCoin(denom, amount))

	result, err := h.ContractKeeper.Execute(ctx, contractAddr, WasmHookModuleAccountAddr, msgBytes, funds)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadExecutionMsg, err.Error()))
	}

	fullAck := ContractAck{ContractResult: result, IbcAck: ack.Acknowledgement()}
	bz, err = json.Marshal(fullAck)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadResponse, err.Error()))
	}

	return channeltypes.NewResultAcknowledgement(bz)
}

func ValidateMemo(memo string, receiver string) (shouldBreak bool, errStr string, contractAddr sdk.AccAddress, msgBytes []byte) {
	// If there is no memo, the packet was either sent with an earlier version of IBC, or the memo was
	// intentionally left blank. Nothing to do here. Ignore the packet and pass it down the stack.
	if len(memo) == 0 {
		return true, "", nil, nil
	}

	// the metadata must be a valid JSON object
	var metadata map[string]interface{}
	err := json.Unmarshal([]byte(memo), &metadata)
	if err != nil {
		return false, fmt.Sprintf(types.ErrBadPacketMetadataMsg, metadata, err.Error()), nil, nil
	}

	// If the key "wasm"  doesn't exist, there's nothing to do on this hook. Continue by passing the packet
	// down the stack
	wasmRaw, ok := metadata["wasm"]
	if !ok {
		return true, "", nil, nil
	}

	// Make sure the wasm key is a map. If it isn't, ignore this packet
	wasm, ok := wasmRaw.(map[string]interface{})
	if !ok {
		return true, "", nil, nil
	}

	// Get the contract
	contract, ok := wasm["contract"].(string)
	if !ok {
		// The tokens will be returned
		return false, fmt.Sprintf(types.ErrBadMetadataFormatMsg, memo, `Could not find key wasm["contract"]`), nil, nil
	}

	contractAddr, err = sdk.AccAddressFromBech32(contract)
	if err != nil {
		return false, fmt.Sprintf(types.ErrBadMetadataFormatMsg, memo, `wasm["contract"] is not a valid bech32 address`), nil, nil
	}

	// The contract and the receiver should be the same for the packet to be valid
	if contract != receiver {
		return false, fmt.Sprintf(types.ErrBadMetadataFormatMsg, memo, `wasm["contract"] should be the same as the receiver of the packet`), nil, nil
	}

	// Ensure the message key is provided
	if wasm["msg"] == nil {
		return false, fmt.Sprintf(types.ErrBadMetadataFormatMsg, memo, `Could not find key wasm["msg"]`), nil, nil
	}

	// Make sure the msg key is a map. If it isn't, return an error
	_, ok = wasm["msg"].(map[string]interface{})
	if !ok {
		return false, fmt.Sprintf(types.ErrBadMetadataFormatMsg, memo, `wasm["msg"] is not an object`), nil, nil
	}

	// Get the message string by serializing the map
	msgBytes, err = json.Marshal(wasm["msg"])
	if err != nil {
		// The tokens will be returned
		return false, fmt.Sprintf(types.ErrBadMetadataFormatMsg, memo, err.Error()), nil, nil
	}

	return false, "", contractAddr, msgBytes
}
