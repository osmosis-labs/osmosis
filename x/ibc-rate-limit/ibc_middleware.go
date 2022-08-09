package ibc_rate_limit

import (
	"errors"
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/osmosis-labs/osmosis/v10/x/ibc-rate-limit/types"
)

var _ porttypes.Middleware = &IBCModule{}
var _ porttypes.ICS4Wrapper = &ICS4Middleware{}

type ICS4Middleware struct {
	channel       porttypes.ICS4Wrapper
	accountKeeper *authkeeper.AccountKeeper
	ParamSpace    paramtypes.Subspace
	WasmKeeper    *wasmkeeper.Keeper
}

func NewICS4Middleware(channel porttypes.ICS4Wrapper, accountKeeper *authkeeper.AccountKeeper, wasmKeeper *wasmkeeper.Keeper, paramSpace paramtypes.Subspace) ICS4Middleware {
	return ICS4Middleware{
		channel:       channel,
		accountKeeper: accountKeeper,
		WasmKeeper:    wasmKeeper,
		ParamSpace:    paramSpace,
	}
}

func (i ICS4Middleware) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI) error {
	fmt.Println("Sending package through middleware")
	contract := i.ParamSpace.GetRaw(ctx, []byte("contract"))
	if contract == nil {
		// The contract has not been configured. Continue as usual
		return i.channel.SendPacket(ctx, chanCap, packet)
	}

	sendPacketMsg := `{"send_packet": {"channel_id": "test", "channel_value": "100", "funds": "1"}}`
	sender := i.accountKeeper.GetModuleAccount(ctx, transfertypes.ModuleName)

	// ToDo: This shoiuld probably be done through the message dispatcher
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(i.WasmKeeper)
	response, err := contractKeeper.Execute(ctx, contract, sender.GetAddress(), []byte(sendPacketMsg), sdk.Coins{})

	if err != nil {
		// Handle potential errors
		if !errors.Is(err, wasmtypes.ErrNotFound) { // Contract not found. This means the rate limiter is not configured
			// ToDo: Improve error handling here
			return sdkerrors.Wrap(types.ErrRateLimitExceeded, "SendPacket")
		}
	}
	fmt.Println(string(response))
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
