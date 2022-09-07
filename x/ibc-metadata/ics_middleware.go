package ibc_metadata

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"
)

type ICS4Middleware struct {
	channel porttypes.ICS4Wrapper
}

func NewICS4Middleware(channel porttypes.ICS4Wrapper) ICS4Middleware {
	return ICS4Middleware{
		channel: channel,
	}
}

func (i ICS4Middleware) SendPacket(ctx sdk.Context, channelCap *capabilitytypes.Capability, packet ibcexported.PacketI) error {
	//swapAddress := types.GetFundAddress(packet.GetSourcePort(), packet.GetSourceChannel()).String()
	//var data types.FungibleTokenPacketData
	//if err := json.Unmarshal(packet.GetData(), &data); err != nil {
	//	return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal sent packet data: %s", err.Error())
	//}
	//
	//// if Sent from the swap address, skip packet
	//if swapAddress == data.Sender {
	//	return nil
	//}
	//
	//return w.channelKeeper.SendPacket(ctx, channelCap, packet)
	return i.channel.SendPacket(ctx, channelCap, packet)
}

func (i ICS4Middleware) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, ack ibcexported.Acknowledgement) error {
	return i.channel.WriteAcknowledgement(ctx, chanCap, packet, ack)
}
