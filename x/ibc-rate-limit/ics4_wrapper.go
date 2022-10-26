package ibc_rate_limit

import (
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"strings"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
)

var (
	_ porttypes.Middleware  = &IBCModule{}
	_ porttypes.ICS4Wrapper = &ICS4Wrapper{}
)

type ICS4Wrapper struct {
	channel        porttypes.ICS4Wrapper
	accountKeeper  *authkeeper.AccountKeeper
	bankKeeper     *bankkeeper.BaseKeeper
	ContractKeeper *wasmkeeper.PermissionedKeeper
	paramSpace     paramtypes.Subspace
}

func NewICS4Middleware(
	channel porttypes.ICS4Wrapper,
	accountKeeper *authkeeper.AccountKeeper, contractKeeper *wasmkeeper.PermissionedKeeper,
	bankKeeper *bankkeeper.BaseKeeper, paramSpace paramtypes.Subspace,
) ICS4Wrapper {
	return ICS4Wrapper{
		channel:        channel,
		accountKeeper:  accountKeeper,
		ContractKeeper: contractKeeper,
		bankKeeper:     bankKeeper,
		paramSpace:     paramSpace,
	}
}

// SendPacket implements the ICS4 interface and is called when sending packets.
// This method retrieves the contract from the middleware's parameters and checks if the limits have been exceeded for
// the current transfer, in which case it returns an error preventing the IBC send from taking place.
// If the contract param is not configured, or the contract doesn't have a configuration for the (channel+denom) being
// used, transfers are not prevented and handled by the wrapped IBC app
func (i *ICS4Wrapper) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI) error {
	contract := i.GetParams(ctx)
	if contract == "" {
		// The contract has not been configured. Continue as usual
		return i.channel.SendPacket(ctx, chanCap, packet)
	}

	//amount, denom, _, _, err := GetFundsFromPacket(packet)
	//if err != nil {
	//	return sdkerrors.Wrap(err, "Rate limit SendPacket")
	//}
	//
	//// Calculate the channel value using the denom from the packet.
	//// This will be the denom for native tokens, and "ibc/..." for foreign ones being sent back
	//channelValue := i.CalculateChannelValue(ctx, denom, packet)

	fullPacket, ok := packet.(channeltypes.Packet)
	if !ok {
		return sdkerrors.ErrInvalidRequest
	}

	err := CheckAndUpdateRateLimits(
		ctx,
		i.ContractKeeper,
		"send_packet",
		contract,
		fullPacket,
	)

	if err != nil {
		return sdkerrors.Wrap(err, "bad packet in rate limit's SendPacket")
	}

	return i.channel.SendPacket(ctx, chanCap, packet)
}

func (i *ICS4Wrapper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	return i.channel.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

func (i *ICS4Wrapper) GetParams(ctx sdk.Context) (contract string) {
	i.paramSpace.GetIfExists(ctx, []byte("contract"), &contract)
	return contract
}

// CalculateChannelValue The value of an IBC channel. This is calculated using the denom supplied by the sender.
// if the denom is not correct, the transfer should fail somewhere else on the call chain
func (i *ICS4Wrapper) CalculateChannelValue(ctx sdk.Context, denom string, packet exported.PacketI) sdk.Int {
	// If the source is the counterparty chain, this should be
	if strings.HasPrefix(denom, "ibc/") {
		return i.bankKeeper.GetSupplyWithOffset(ctx, denom).Amount
	}

	escrowAddress := transfertypes.GetEscrowAddress(packet.GetSourcePort(), packet.GetSourceChannel())
	return i.bankKeeper.GetBalance(ctx, escrowAddress, denom).Amount
}
