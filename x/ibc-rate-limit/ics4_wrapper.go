package ibc_rate_limit

import (
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v4/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"

	"github.com/osmosis-labs/osmosis/v14/x/ibc-rate-limit/types"
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

func (i *ICS4Wrapper) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return i.channel.GetAppVersion(ctx, portID, channelID)
}

func NewICS4Middleware(
	channel porttypes.ICS4Wrapper,
	accountKeeper *authkeeper.AccountKeeper, contractKeeper *wasmkeeper.PermissionedKeeper,
	bankKeeper *bankkeeper.BaseKeeper, paramSpace paramtypes.Subspace,
) ICS4Wrapper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}
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
	contract := i.GetContractAddress(ctx)
	if contract == "" {
		// The contract has not been configured. Continue as usual
		return i.channel.SendPacket(ctx, chanCap, packet)
	}

	// We need the full packet so the contract can process it. If it can't be cast to a channeltypes.Packet, this
	// should fail. The only reason that would happen is if another middleware is modifying the packet, though. In
	// that case we can modify the middleware order or change this cast so we have all the data we need.
	fullPacket, ok := packet.(channeltypes.Packet)
	if !ok {
		return sdkerrors.ErrInvalidRequest
	}

	err := CheckAndUpdateRateLimits(ctx, i.ContractKeeper, "send_packet", contract, fullPacket)
	if err != nil {
		return sdkerrors.Wrap(err, "rate limit SendPacket failed to authorize transfer")
	}

	return i.channel.SendPacket(ctx, chanCap, packet)
}

func (i *ICS4Wrapper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	return i.channel.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

func (i *ICS4Wrapper) GetContractAddress(ctx sdk.Context) (contract string) {
	return i.GetParams(ctx).ContractAddress
}

func (i *ICS4Wrapper) GetParams(ctx sdk.Context) (params types.Params) {
	if !i.paramSpace.Has(ctx, []byte("contract")) {
		fmt.Println("contract does not exist")
	}
	i.paramSpace.GetParamSet(ctx, &params)
	fmt.Println("params", params)
	return params
}

func (i *ICS4Wrapper) SetParams(ctx sdk.Context, params types.Params) {
	i.paramSpace.SetParamSet(ctx, &params)
}
