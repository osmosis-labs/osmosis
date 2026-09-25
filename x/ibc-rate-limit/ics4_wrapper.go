package ibc_rate_limit

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/osmosis-labs/osmosis/v31/x/ibc-rate-limit/types"
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
	var packetdata transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(data, &packetdata); err != nil {
		return i.channel.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}
	if packetdata.Denom == "" || packetdata.Amount == "" {
		return i.channel.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}
	contract := i.GetContractAddress(ctx)
	if contract == "" {
		// The contract has not been configured. Continue as usual
		return i.channel.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}

	// Authorise and charge the quota before anything is committed. The
	// contract does not know the sequence yet (it is only assigned by the
	// inner send), so it is passed as 0 and associated afterwards.
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

	sequence, err := i.channel.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	if err != nil {
		return 0, err
	}

	// Now that the sequence is known, let the contract associate it with the
	// charge so a later failure of this packet can be refunded while its quota
	// window is still active. If this fails the send stays charged and is
	// simply not refundable, which is the conservative outcome, so it must not
	// fail the transfer.
	fullPacket.Sequence = sequence
	if err := RecordSendRateLimit(ctx, i.ContractKeeper, contract, fullPacket); err != nil {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventRecordSendFailed,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(types.AttributeKeyPacket, string(data)),
				sdk.NewAttribute(types.AttributeKeyError, err.Error()),
			),
		)
	}

	return sequence, nil
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
