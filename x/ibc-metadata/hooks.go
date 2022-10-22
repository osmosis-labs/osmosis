package ibc_metadata

import (
	"encoding/json"
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/osmosis-labs/osmosis/v12/x/ibc-metadata/types"
)

func execute(ctx sdk.Context, contractKeeper *wasmkeeper.PermissionedKeeper, contract string, msg []byte, caller sdk.AccAddress, data types.FungibleTokenPacketData) error {
	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return err
	}

	amount, ok := sdk.NewIntFromString(data.Amount)
	if !ok {
		return sdk.ErrInvalidDecimalStr
	}

	response, err := contractKeeper.Execute(
		ctx, contractAddr, caller,
		[]byte(msg),
		sdk.NewCoins(sdk.NewCoin(data.Denom, amount)),
	)
	if err != nil {
		return err
	}
	fmt.Println(response)

	return nil
}

func WasmHook(im IBCModule, ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) ibcexported.Acknowledgement {
	var data types.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &data); err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf("cannot unmarshal sent packet data: %s", err.Error()))
	}

	metadataBytes := data.GetMetadata()
	if metadataBytes == nil || len(metadataBytes) == 0 {
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	var metadata map[string]interface{}
	err := json.Unmarshal(metadataBytes, &metadata) // ToDo: Be more flexible here? maybe just continue on invalid metadata
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadPacketMetadataMsg, metadata, err.Error()))
	}

	// Check for the wasm key. If it doesn't exist. We continue.
	wasmRaw, ok := metadata["wasm"]
	if !ok {
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	wasm, ok := wasmRaw.(map[string]interface{})
	if !ok {
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	// Get the message
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

	// Set the receiver to the contract. That way it will be able to manage the funds sent in the packet
	data.Receiver = contract
	// Revert the metadata so that the underlying implementation can handle it. This won't be necessary once IBC is upgraded to contain metadata
	data.Metadata = nil
	packet.Data, err = json.Marshal(data)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(types.ErrPacketCreationMsg)
	}

	ack := im.app.OnRecvPacket(ctx, packet, relayer)

	caller, _ := sdk.AccAddressFromBech32(data.Receiver)
	err = execute(ctx, im.ics4Middleware.ContractKeeper, contract, msg, caller, data)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadExecutionMsg, err.Error()))
	}

	// This should actually be done inside the contract
	im.TransferKeeper.SendTransfer(
		ctx,
		packet.GetSourcePort(),
		packet.GetDestPort(),
		sdk.NewCoin("uion", sdk.NewInt(1)),
		sdk.AccAddress(data.Sender),
		data.Receiver,
		clienttypes.NewHeight(0, 100),
		0,
	)

	return ack
}
