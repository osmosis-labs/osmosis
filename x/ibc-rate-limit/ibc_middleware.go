package ibc_rate_limit

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
)

var _ porttypes.Middleware = &IBCModule{}
var _ porttypes.ICS4Wrapper = &ICS4Middleware{}

type ICS4Middleware struct {
	channel porttypes.ICS4Wrapper
}

func NewICS4Middleware(channel porttypes.ICS4Wrapper) ICS4Middleware {
	fmt.Println("Initializing ics4")
	return ICS4Middleware{
		channel: channel,
	}
}

func (i ICS4Middleware) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI) error {
	fmt.Println("Sending package through middleware")
	//return sdkerrors.Wrap(types.ErrRateLimitExceeded, "test")
	return i.channel.SendPacket(ctx, chanCap, packet)
}

func (i ICS4Middleware) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	fmt.Println("WriteAcknowledgement middleware")
	return i.channel.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

type IBCModule struct {
	app            porttypes.IBCModule
	ics4Middleware ICS4Middleware
}

func NewIBCModule(app porttypes.IBCModule, ics4 ICS4Middleware) IBCModule {
	fmt.Println("Initializing middleware")
	return IBCModule{
		app:            app,
		ics4Middleware: ics4,
	}
}

// OnChanOpenInit implements the IBCModule interface
func (im IBCModule) OnChanOpenInit(ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) error {
	fmt.Println("OnChanOpenInit Middleware")
	return im.app.OnChanOpenInit(
		ctx,
		order,
		connectionHops,
		portID,
		channelID,
		channelCap,
		counterparty,
		version, // note we only pass app version here
	)
}

// OnChanOpenTry implements the IBCModule interface
func (im IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	fmt.Println("OnChanOpenTry Middleware")
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, counterpartyVersion)
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	fmt.Println("OnChanOpenAck Middleware")
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	fmt.Println("OnChanOpenConfirm Middleware")
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	fmt.Println("OnChanCloseInit Middleware")
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	fmt.Println("OnChanCloseConfirm Middleware")
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnRecvPacket implements the IBCModule interface
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	fmt.Println("OnRecvPacket Middleware")
	//return channeltypes.NewErrorAcknowledgement(types.RateLimitExceededMsg)

	return im.app.OnRecvPacket(ctx, packet, relayer)
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	fmt.Println("OnAcknowledgementPacket Middleware")
	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	fmt.Println("OnTimeoutPacket Middleware")
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

// SendPacket implements the ICS4 Wrapper interface
func (im IBCModule) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet exported.PacketI,
) error {
	fmt.Println("Sending package through middleware")
	return im.ics4Middleware.SendPacket(ctx, chanCap, packet)
}

// WriteAcknowledgement implements the ICS4 Wrapper interface
func (im IBCModule) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet exported.PacketI,
	ack exported.Acknowledgement,
) error {
	fmt.Println("WriteAcknowledgement middleware")
	return im.ics4Middleware.WriteAcknowledgement(ctx, chanCap, packet, ack)
}
