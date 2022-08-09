package ibc_rate_limit

import (
	"encoding/json"
	"fmt"
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
	"github.com/osmosis-labs/osmosis/v10/x/ibc-rate-limit/types"
	lockupkeeper "github.com/osmosis-labs/osmosis/v10/x/lockup/keeper"
	"strings"
)

var _ porttypes.Middleware = &IBCModule{}
var _ porttypes.ICS4Wrapper = &ICS4Middleware{}

type ICS4Middleware struct {
	channel       porttypes.ICS4Wrapper
	accountKeeper *authkeeper.AccountKeeper
	BankKeeper    *bankkeeper.BaseKeeper
	WasmKeeper    *wasmkeeper.Keeper
	LockupKeeper  *lockupkeeper.Keeper
	ParamSpace    paramtypes.Subspace
}

func NewICS4Middleware(
	channel porttypes.ICS4Wrapper,
	accountKeeper *authkeeper.AccountKeeper, wasmKeeper *wasmkeeper.Keeper,
	bankKeeper *bankkeeper.BaseKeeper, lockupKeeper *lockupkeeper.Keeper,
	paramSpace paramtypes.Subspace,
) ICS4Middleware {
	return ICS4Middleware{
		channel:       channel,
		accountKeeper: accountKeeper,
		WasmKeeper:    wasmKeeper,
		BankKeeper:    bankKeeper,
		LockupKeeper:  lockupKeeper,
		ParamSpace:    paramSpace,
	}
}

func (i ICS4Middleware) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI) error {
	contractRaw := i.ParamSpace.GetRaw(ctx, []byte("contract"))
	if contractRaw == nil {
		// The contract has not been configured. Continue as usual
		return i.channel.SendPacket(ctx, chanCap, packet)
	}

	contract := strings.Trim(string(contractRaw), `"`) // ToDo: Why is this stored with ""
	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return err
	}

	var packetData map[string]interface{} // ToDo: Do this with a struct
	err = json.Unmarshal(packet.GetData(), &packetData)
	if err != nil {
		return err
	}

	sendPacketMsg := i.BuildWasmExecMsg(
		ctx,
		packet.GetSourceChannel(),
		packetData["denom"].(string),
		packetData["amount"].(string),
	)
	sender := i.accountKeeper.GetModuleAccount(ctx, transfertypes.ModuleName)

	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(i.WasmKeeper)
	// ToDo: Why doesn't this return a response
	_, err = contractKeeper.Execute(ctx, contractAddr, sender.GetAddress(), []byte(sendPacketMsg), sdk.Coins{})

	if err != nil {
		// ToDo: catch the wasm error and return err if it's something unexpected
		fmt.Println(err)
		return sdkerrors.Wrap(types.ErrRateLimitExceeded, "SendPacket")
	}

	return i.channel.SendPacket(ctx, chanCap, packet)
}

func (i ICS4Middleware) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	fmt.Println("WriteAcknowledgement middleware")
	return i.channel.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

func (i *ICS4Middleware) BuildWasmExecMsg(ctx sdk.Context, sourceChannel, denom, amount string) string {
	// ToDo: Do this with a struct
	return fmt.Sprintf(
		`{"send_packet": {"channel_id": "%s", "channel_value": "%s", "funds": "%s"}}`,
		sourceChannel,
		i.CalculateChannelValue(ctx, denom),
		amount,
	)
}

// CalculateChannelValue The value of an IBC channel. This is calculated using the denom supplied by the sender.
//  if the denom is not correct, the transfer should fail somewhere else on the call chain
func (i *ICS4Middleware) CalculateChannelValue(ctx sdk.Context, denom string) sdk.Int {
	supply := i.BankKeeper.GetSupply(ctx, denom)
	locked := i.LockupKeeper.GetModuleLockedCoins(ctx)
	return supply.Amount.Add(locked.AmountOf(denom))
}

type IBCModule struct {
	app            porttypes.IBCModule
	ics4Middleware ICS4Middleware
}

func NewIBCModule(app porttypes.IBCModule, ics4 ICS4Middleware) IBCModule {
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
