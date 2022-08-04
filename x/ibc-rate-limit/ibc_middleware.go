package ibc_rate_limit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
)

var _ porttypes.Middleware = &RateLimitMiddleware{}

type RateLimitMiddleware struct {
	porttypes.IBCModule

	app    porttypes.IBCModule
	keeper *ibckeeper.Keeper
}

func NewRateLimitMiddleware(app porttypes.IBCModule, k *ibckeeper.Keeper) RateLimitMiddleware {
	return RateLimitMiddleware{
		app:    app,
		keeper: k,
	}
}

//// OnChanOpenInit implements the IBCModule interface
//func (im RateLimitMiddleware) OnChanOpenInit(ctx sdk.Context,
//	order channeltypes.Order,
//	connectionHops []string,
//	portID string,
//	channelID string,
//	channelCap *capabilitytypes.Capability,
//	counterparty channeltypes.Counterparty,
//	version string,
//) error {
//	return im.app.OnChanOpenInit(
//		ctx,
//		order,
//		connectionHops,
//		portID,
//		channelID,
//		channelCap,
//		counterparty,
//		version, // note we only pass app version here
//	)
//}
//
//// OnChanOpenTry implements the IBCModule interface
//func (im RateLimitMiddleware) OnChanOpenTry(
//	ctx sdk.Context,
//	order channeltypes.Order,
//	connectionHops []string,
//	portID,
//	channelID string,
//	channelCap *capabilitytypes.Capability,
//	counterparty channeltypes.Counterparty,
//	counterpartyVersion string,
//) (string, error) {
//	// call underlying app's OnChanOpenTry callback with the app versions
//	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, counterpartyVersion)
//}
//
//// OnChanOpenAck implements the IBCModule interface
//func (im RateLimitMiddleware) OnChanOpenAck(
//	ctx sdk.Context,
//	portID,
//	channelID string,
//	counterpartyChannelID string,
//	counterpartyVersion string,
//) error {
//	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
//}
//
//// OnChanOpenConfirm implements the IBCModule interface
//func (im RateLimitMiddleware) OnChanOpenConfirm(
//	ctx sdk.Context,
//	portID,
//	channelID string,
//) error {
//	//doCustomLogic()
//	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
//}
//
//// OnChanCloseInit implements the IBCModule interface
//func (im RateLimitMiddleware) OnChanCloseInit(
//	ctx sdk.Context,
//	portID,
//	channelID string,
//) error {
//	return im.app.OnChanCloseInit(ctx, portID, channelID)
//}
//
//// OnChanCloseConfirm implements the IBCModule interface
//func (im RateLimitMiddleware) OnChanCloseConfirm(
//	ctx sdk.Context,
//	portID,
//	channelID string,
//) error {
//	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
//}
//
//// OnRecvPacket implements the IBCModule interface
//func (im RateLimitMiddleware) OnRecvPacket(
//	ctx sdk.Context,
//	packet channeltypes.Packet,
//	relayer sdk.AccAddress,
//) exported.Acknowledgement {
//	return im.app.OnRecvPacket(ctx, packet, relayer)
//}
//
//// OnAcknowledgementPacket implements the IBCModule interface
//func (im RateLimitMiddleware) OnAcknowledgementPacket(
//	ctx sdk.Context,
//	packet channeltypes.Packet,
//	acknowledgement []byte,
//	relayer sdk.AccAddress,
//) error {
//	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
//}
//
//// OnTimeoutPacket implements the IBCModule interface
//func (im RateLimitMiddleware) OnTimeoutPacket(
//	ctx sdk.Context,
//	packet channeltypes.Packet,
//	relayer sdk.AccAddress,
//) error {
//	return im.app.OnTimeoutPacket(ctx, packet, relayer)
//}

// SendPacket implements the ICS4 Wrapper interface
func (im RateLimitMiddleware) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet exported.PacketI,
) error {
	return im.keeper.ChannelKeeper.SendPacket(ctx, chanCap, packet)
}

// WriteAcknowledgement implements the ICS4 Wrapper interface
func (im RateLimitMiddleware) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet exported.PacketI,
	ack exported.Acknowledgement,
) error {
	return im.keeper.ChannelKeeper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}
