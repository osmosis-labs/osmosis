package ibc_rate_limit

import (
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/osmosis-labs/osmosis/v11/x/ibc-rate-limit/types"
	lockupkeeper "github.com/osmosis-labs/osmosis/v11/x/lockup/keeper"
)

var (
	_ porttypes.Middleware  = &IBCModule{}
	_ porttypes.ICS4Wrapper = &ICS4Middleware{}
)

type ICS4Middleware struct {
	channel        porttypes.ICS4Wrapper
	accountKeeper  *authkeeper.AccountKeeper
	BankKeeper     *bankkeeper.BaseKeeper
	ContractKeeper *wasmkeeper.PermissionedKeeper
	LockupKeeper   *lockupkeeper.Keeper
	ParamSpace     paramtypes.Subspace
}

func NewICS4Middleware(
	channel porttypes.ICS4Wrapper,
	accountKeeper *authkeeper.AccountKeeper, contractKeeper *wasmkeeper.PermissionedKeeper,
	bankKeeper *bankkeeper.BaseKeeper, lockupKeeper *lockupkeeper.Keeper,
	paramSpace paramtypes.Subspace,
) ICS4Middleware {
	return ICS4Middleware{
		channel:        channel,
		accountKeeper:  accountKeeper,
		ContractKeeper: contractKeeper,
		BankKeeper:     bankKeeper,
		LockupKeeper:   lockupKeeper,
		ParamSpace:     paramSpace,
	}
}

// SendPacket implements the ICS4 interface and is called when sending packets.
// This method retrieves the contract from the middleware's parameters and checks if the limits have been exceeded for
// the current transfer, in which case it returns an error preventing the IBC send from taking place.
// If the contract param is not configured, or the contract doesn't have a configuration for the (channel+denom) being
// used, transfers are not prevented and handled by the wrapped IBC app
func (i *ICS4Middleware) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI) error {
	var params types.Params
	i.ParamSpace.GetIfExists(ctx, []byte("contract"), &params)
	if params.ContractAddress == "" {
		// The contract has not been configured. Continue as usual
		return i.channel.SendPacket(ctx, chanCap, packet)
	}

	amount, denom, err := GetFundsFromPacket(packet)
	if err != nil {
		return sdkerrors.Wrap(err, "Rate limited SendPacket")
	}
	channelValue := i.CalculateChannelValue(ctx, denom)
	err = CheckAndUpdateRateLimits(
		ctx,
		i.ContractKeeper,
		"send_packet",
		params.ContractAddress,
		channelValue,
		packet.GetSourceChannel(),
		denom,
		amount,
	)
	if err != nil {
		return sdkerrors.Wrap(err, "Rate limited SendPacket")
	}

	return i.channel.SendPacket(ctx, chanCap, packet)
}

func (i *ICS4Middleware) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	return i.channel.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

// CalculateChannelValue The value of an IBC channel. This is calculated using the denom supplied by the sender.
// if the denom is not correct, the transfer should fail somewhere else on the call chain
func (i *ICS4Middleware) CalculateChannelValue(ctx sdk.Context, denom string) sdk.Int {
	return i.BankKeeper.GetSupplyWithOffset(ctx, denom).Amount
}

type IBCModule struct {
	app            porttypes.IBCModule
	ics4Middleware *ICS4Middleware
}

func NewIBCModule(app porttypes.IBCModule, ics4 *ICS4Middleware) IBCModule {
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
) error {
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

// OnRecvPacket implements the IBCModule interface
func (im *IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	var params types.Params
	im.ics4Middleware.ParamSpace.GetIfExists(ctx, []byte("contract"), &params)
	if params.ContractAddress == "" {
		// The contract has not been configured. Continue as usual
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}
	amount, denom, err := GetFundsFromPacket(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement("bad packet")
	}
	channelValue := im.ics4Middleware.CalculateChannelValue(ctx, denom)

	err = CheckAndUpdateRateLimits(
		ctx,
		im.ics4Middleware.ContractKeeper,
		"recv_packet",
		params.ContractAddress,
		channelValue,
		packet.GetDestChannel(),
		denom,
		amount,
	)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(types.RateLimitExceededMsg)
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

	if !ack.Success() {
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
	packet channeltypes.Packet,
) error {
	var data transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &data); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}
	var params types.Params
	im.ics4Middleware.ParamSpace.GetIfExists(ctx, []byte("contract"), &params)
	if params.ContractAddress == "" {
		// The contract has not been configured. Continue as usual
		return nil
	}
	channelValue := im.ics4Middleware.CalculateChannelValue(ctx, data.Denom)

	// This could return an error if the "receive" path is full. We should consider adding a message to the
	// contract so that we can force the revert in this case
	_ = CheckAndUpdateRateLimits(
		ctx,
		im.ics4Middleware.ContractKeeper,
		"recv_packet",
		params.ContractAddress,
		channelValue,
		packet.GetDestChannel(),
		data.Denom,
		data.Amount,
	)
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
