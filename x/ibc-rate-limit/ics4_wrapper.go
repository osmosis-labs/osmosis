package ibc_rate_limit

import (
	errorsmod "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/osmosis-labs/osmosis/v21/x/ibc-rate-limit/types"
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
func (i *ICS4Wrapper) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, sourcePort, sourceChannel string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64, data []byte) (uint64, error) {
	contract := i.GetContractAddress(ctx)
	if contract == "" {
		// The contract has not been configured. Continue as usual
		return i.channel.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}

	// We need the full packet so the contract can process it. If it can't be cast to a channeltypes.Packet, this
	// should fail. The only reason that would happen is if another middleware is modifying the packet, though. In
	// that case we can modify the middleware order or change this cast so we have all the data we need.
	// UNFORKINGTODO OQ: The full packet data is not available here. Specifically, the sequence, destPort and destChannel are not available.
	// This is silly as it means we cannot filter packets based on destination (the sequence could be obtained by calling channel.SendPacket() before checking the rate limits)
	// I think this works with the current contracts as destination is not checked for sends, but would need to double check to be 100% sure.
	// Should we modify what the contracts expect so that there's no risk of them trying to rely on the missing data? Alt. just document this
	// UNFORKINGTODO N: I am setting the sequence to 0 so it can compile, but note that this needs to be addressed.
	fullPacket := channeltypes.Packet{
		Sequence:           0,
		SourcePort:         sourcePort,
		SourceChannel:      sourceChannel,
		DestinationPort:    "omitted",
		DestinationChannel: "omitted",
		Data:               data,
		TimeoutTimestamp:   timeoutTimestamp,
		TimeoutHeight:      timeoutHeight,
	}

	err := CheckAndUpdateRateLimits(ctx, i.ContractKeeper, "send_packet", contract, fullPacket)
	if err != nil {
		return 0, errorsmod.Wrap(err, "rate limit SendPacket failed to authorize transfer")
	}

	return i.channel.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

func (i *ICS4Wrapper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	return i.channel.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

func (i *ICS4Wrapper) GetContractAddress(ctx sdk.Context) (contract string) {
	return i.GetParams(ctx).ContractAddress
}

func (i *ICS4Wrapper) GetParams(ctx sdk.Context) (params types.Params) {
	// This was previously done via i.paramSpace.GetParamSet(ctx, &params). That will
	// panic if the params don't exist. This is a workaround to avoid that panic.
	// Params should be refactored to just use a raw kvstore.
	empty := types.Params{}
	for _, pair := range params.ParamSetPairs() {
		i.paramSpace.GetIfExists(ctx, pair.Key, pair.Value)
	}
	if params == empty {
		return types.DefaultParams()
	}
	return params
}

func (i *ICS4Wrapper) SetParams(ctx sdk.Context, params types.Params) {
	i.paramSpace.SetParamSet(ctx, &params)
}
