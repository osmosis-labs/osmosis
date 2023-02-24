package ibc_rate_limit

import (
	"encoding/json"
	"strings"

	"github.com/osmosis-labs/osmosis/osmoutils"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v4/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"

	"github.com/osmosis-labs/osmosis/v15/x/ibc-rate-limit/types"
)

type IBCModule struct {
	app            porttypes.IBCModule
	ics4Middleware *ICS4Wrapper
}

func NewIBCModule(app porttypes.IBCModule, ics4 *ICS4Wrapper) IBCModule {
	return IBCModule{
		app:            app,
		ics4Middleware: ics4,
	}
}

// OnChanOpenInit implements the IBCModule interface
func (im *IBCModule) OnChanOpenInit(ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return im.app.OnChanOpenInit(
		ctx,
		order,
		connectionHops,
		portID,
		channelID,
		channelCap,
		counterparty,
		version,
	)
}

// OnChanOpenTry implements the IBCModule interface
func (im *IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, counterpartyVersion)
}

// OnChanOpenAck implements the IBCModule interface
func (im *IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	// Here we can add initial limits when a new channel is open. For now, they can be added manually on the contract
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements the IBCModule interface
func (im *IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// Here we can add initial limits when a new channel is open. For now, they can be added manually on the contract
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCModule interface
func (im *IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// Here we can remove the limits when a new channel is closed. For now, they can remove them  manually on the contract
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm implements the IBCModule interface
func (im *IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// Here we can remove the limits when a new channel is closed. For now, they can remove them  manually on the contract
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

func ValidateReceiverAddress(packet exported.PacketI) error {
	var packetData transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &packetData); err != nil {
		return err
	}
	if len(packetData.Receiver) >= 4096 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "IBC Receiver address too long. Max supported length is %d", 4096)
	}
	return nil
}

// OnRecvPacket implements the IBCModule interface
func (im *IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	if err := ValidateReceiverAddress(packet); err != nil {
		return osmoutils.NewEmitErrorAcknowledgement(ctx, types.ErrBadMessage, err.Error())
	}

	contract := im.ics4Middleware.GetContractAddress(ctx)
	if contract == "" {
		// The contract has not been configured. Continue as usual
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	err := CheckAndUpdateRateLimits(ctx, im.ics4Middleware.ContractKeeper, "recv_packet", contract, packet)
	if err != nil {
		if strings.Contains(err.Error(), "rate limit exceeded") {
			return osmoutils.NewEmitErrorAcknowledgement(ctx, types.ErrRateLimitExceeded)
		}
		fullError := sdkerrors.Wrap(types.ErrContractError, err.Error())
		return osmoutils.NewEmitErrorAcknowledgement(ctx, fullError)
	}

	// if this returns an Acknowledgement that isn't successful, all state changes are discarded
	return im.app.OnRecvPacket(ctx, packet, relayer)
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im *IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	var ack channeltypes.Acknowledgement
	if err := json.Unmarshal(acknowledgement, &ack); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	if osmoutils.IsAckError(acknowledgement) {
		err := im.RevertSentPacket(ctx, packet) // If there is an error here we should still handle the ack
		if err != nil {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventBadRevert,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
					sdk.NewAttribute(types.AttributeKeyFailureType, "acknowledgment"),
					sdk.NewAttribute(types.AttributeKeyPacket, string(packet.GetData())),
					sdk.NewAttribute(types.AttributeKeyAck, string(acknowledgement)),
				),
			)
		}
	}

	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface
func (im *IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	err := im.RevertSentPacket(ctx, packet) // If there is an error here we should still handle the timeout
	if err != nil {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventBadRevert,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(types.AttributeKeyFailureType, "timeout"),
				sdk.NewAttribute(types.AttributeKeyPacket, string(packet.GetData())),
			),
		)
	}
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

// RevertSentPacket Notifies the contract that a sent packet wasn't properly received
func (im *IBCModule) RevertSentPacket(
	ctx sdk.Context,
	packet exported.PacketI,
) error {
	contract := im.ics4Middleware.GetContractAddress(ctx)
	if contract == "" {
		// The contract has not been configured. Continue as usual
		return nil
	}

	if err := UndoSendRateLimit(
		ctx,
		im.ics4Middleware.ContractKeeper,
		contract,
		packet,
	); err != nil {
		return err
	}
	return nil
}

// SendPacket implements the ICS4 Wrapper interface
func (im *IBCModule) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet exported.PacketI,
) error {
	return im.ics4Middleware.SendPacket(ctx, chanCap, packet)
}

// WriteAcknowledgement implements the ICS4 Wrapper interface
func (im *IBCModule) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet exported.PacketI,
	ack exported.Acknowledgement,
) error {
	return im.ics4Middleware.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

func (im *IBCModule) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return im.ics4Middleware.GetAppVersion(ctx, portID, channelID)
}
