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

// ToDo: Split this into its own package

type Metadata struct {
	Callback string `json:"callback"`
}

func ExecuteSwap(ctx sdk.Context, contractKeeper *wasmkeeper.PermissionedKeeper, contract string, caller sdk.AccAddress, data types.FungibleTokenPacketData) error {
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
		[]byte(fmt.Sprintf(`
	{"swap": 
	  {"input_coin": {"amount": "%s", "denom": "uosmo"}, 
	   "output_denom": "uion", 
	   "slipage": {"max_price_impact_percentage": "10"}}
    }`, amount)),
		sdk.NewCoins(sdk.NewCoin(data.Denom, amount)),
	)
	if err != nil {
		return err
	}
	fmt.Println(response)

	return nil
}

func SwapHook(im IBCModule, ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) ibcexported.Acknowledgement {
	var data types.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &data); err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf("cannot unmarshal sent packet data: %s", err.Error()))
	}

	metadataBytes := data.GetMetadata()
	if metadataBytes == nil || len(metadataBytes) == 0 {
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	var metadata Metadata
	err := json.Unmarshal(metadataBytes, &metadata)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadPacketMetadataMsg, metadata, err.Error()))
	}

	data.Metadata = nil
	packet.Data, err = json.Marshal(data)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(types.ErrPacketCreationMsg)
	}
	ack := im.app.OnRecvPacket(ctx, packet, relayer)

	caller, _ := sdk.AccAddressFromBech32(data.Receiver)
	err = ExecuteSwap(ctx, im.ics4Middleware.ContractKeeper, metadata.Callback, caller, data)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadExecutionMsg, err.Error()))
	}

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
