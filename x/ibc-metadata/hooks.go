package ibc_metadata

import (
	"encoding/json"
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	fmt.Println(data)
	_, err = contractKeeper.Execute(
		ctx, contractAddr, caller,
		[]byte(fmt.Sprintf(`{"swap": {"input_coin": {"amount": "1", "denom": "%s"}, "output_denom": "uatom", "minimum_output_amount": "1"}}`, data.Denom)),
		sdk.NewCoins(sdk.NewCoin(data.Denom, amount)),
	)
	if err != nil {
		return err
	}
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

	err = ExecuteSwap(ctx, im.ics4Middleware.ContractKeeper, metadata.Callback, relayer, data)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf(types.ErrBadExecutionMsg, err.Error()))
	}

	return ack
}
