package ibc_metadata

import (
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/osmosis-labs/osmosis/v12/x/ibc-metadata/types"
)

type Metadata struct {
	Callback string `json:"callback"`
}

func TestHook(im IBCModule, ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) ibcexported.Acknowledgement {
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
	// Remove the metadata so that the underlying transfer app can process the
	data.Metadata = nil
	packet.Data, err = json.Marshal(data)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(types.ErrPacketCreation)
	}
	return im.app.OnRecvPacket(ctx, packet, relayer)
}
