package ibc_metadata

import (
	"encoding/json"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/osmosis-labs/osmosis/v12/x/ibc-metadata/types"
)

type ICS4Middleware struct {
	channel        porttypes.ICS4Wrapper
	ContractKeeper *wasmkeeper.PermissionedKeeper // ToDo: Turn hooks into an object and move this there
}

func NewICS4Middleware(channel porttypes.ICS4Wrapper, contractKeeper *wasmkeeper.PermissionedKeeper) ICS4Middleware {
	return ICS4Middleware{
		channel:        channel,
		ContractKeeper: contractKeeper,
	}
}

func (i ICS4Middleware) SendPacket(ctx sdk.Context, channelCap *capabilitytypes.Capability, packet ibcexported.PacketI) error {
	//swapAddress := types.GetFundAddress(packet.GetSourcePort(), packet.GetSourceChannel()).String()
	var data types.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &data); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal sent packet data: %s", err.Error())
	}

	return i.channel.SendPacket(ctx, channelCap, packet)
}

func (i ICS4Middleware) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, ack ibcexported.Acknowledgement) error {
	return i.channel.WriteAcknowledgement(ctx, chanCap, packet, ack)
}
