package ibc_rate_limit

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
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
	ctx.Logger().Error("DBUG::SEND PACKET!!")
	var params types.Params
	i.ParamSpace.GetIfExists(ctx, []byte("contract"), &params)
	if params.ContractAddress == "" {
		// The contract has not been configured. Continue as usual
		return i.channel.SendPacket(ctx, chanCap, packet)
	}

	amount, denom, err := GetFundsFromPacket(packet)
	ctx.Logger().Error("DBUG::FUNDS!!", amount, denom)
	if err != nil {
		return sdkerrors.Wrap(err, "Rate limited SendPacket")
	}
	channelValue := i.CalculateChannelValue(ctx, denom)
	err = CheckRateLimits(
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

	ctx.Logger().Error("DBUG::Sending packet to the channel!!")

	return i.channel.SendPacket(ctx, chanCap, packet)
}

func (i *ICS4Middleware) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	return i.channel.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

// CalculateChannelValue The value of an IBC channel. This is calculated using the denom supplied by the sender.
// if the denom is not correct, the transfer should fail somewhere else on the call chain
func (i *ICS4Middleware) CalculateChannelValue(ctx sdk.Context, denom string) sdk.Int {
	supply := i.BankKeeper.GetSupplyWithOffset(ctx, denom)
	return supply.Amount
	//locked := i.LockupKeeper.GetModuleLockedCoins(ctx)
	//return supply.Amount.Add(locked.AmountOf(denom))
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

	err = CheckRateLimits(
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
	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface
func (im *IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
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
