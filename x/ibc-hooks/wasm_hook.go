package ibc_hooks

import (
	"encoding/json"
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/osmosis-labs/osmosis/v12/x/ibc-hooks/types"
)

type ContractAck struct {
	contractResult []byte
	ibcAck         []byte
}

type WasmHooks struct {
	contractKeeper *wasmkeeper.PermissionedKeeper
}

func (h WasmHooks) ExecuteContract(ctx sdk.Context, contract string, msg []byte, caller sdk.AccAddress, data transfertypes.FungibleTokenPacketData) ([]byte, error) {
	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return nil, err
	}

	amount, ok := sdk.NewIntFromString(data.Amount)
	if !ok {
		return nil, sdkerrors.New("WasmHooks", 10, fmt.Sprintf("Invalid amount provided in packet: %v", data.Amount))
	}
	coins := sdk.NewCoins(sdk.NewCoin(data.Denom, amount))
	result, err := h.contractKeeper.Execute(ctx, contractAddr, caller, msg, coins)

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (h WasmHooks) OnRecvPacketOverride(im IBCMiddleware, ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) ibcexported.Acknowledgement {

	var data transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &data); err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf("cannot unmarshal sent packet data: %s", err.Error()))
	}

	memo := data.GetMemo()
	if len(memo) == 0 {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	}

	var metadata map[string]interface{}
	err := json.Unmarshal([]byte(memo), &metadata) // ToDo: Be more flexible here? maybe just continue on invalid metadata
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadPacketMetadataMsg, metadata, err.Error()))
	}

	// Check for the wasm key. If it doesn't exist. We continue.
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
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadPacketMetadataMsg, metadata, err.Error()))
	}

	// Get the message
	msg, err := json.Marshal(wasm["execute"])
	if err != nil {
		// The tokens will be returned
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadPacketMetadataMsg, metadata, err.Error()))
	}

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

	fullAck := ContractAck{contractResult: result, ibcAck: ack.Acknowledgement()}
	bz, err := json.Marshal(fullAck)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadResponse, err.Error()))
	}

	return channeltypes.NewResultAcknowledgement(bz)
}
