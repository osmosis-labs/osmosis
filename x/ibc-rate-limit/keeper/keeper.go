package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channelkeeper "github.com/cosmos/ibc-go/v3/modules/core/04-channel/keeper"
	ibcchanneltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	portkeeper "github.com/cosmos/ibc-go/v3/modules/core/05-port/keeper"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/osmosis-labs/osmosis/v10/x/ibc-rate-limit/types"
)

// Middleware must implement types.ChannelKeeper and types.PortKeeper expected interfaces
// so that it can wrap IBC channel and port logic for underlying application.
var (
	_ transfertypes.ICS4Wrapper = Keeper{}
	_ types.ChannelKeeper       = Keeper{}
	_ types.PortKeeper          = Keeper{}
)

// Keeper defines the IBC fungible transfer keeper
type Keeper struct {
	Ics4Wrapper   transfertypes.ICS4Wrapper
	ChannelKeeper channelkeeper.Keeper
	PortKeeper    portkeeper.Keeper
}

func NewKeeper(
	ics4Wrapper transfertypes.ICS4Wrapper, channelKeeper channelkeeper.Keeper, portKeeper portkeeper.Keeper,
) Keeper {
	return Keeper{
		Ics4Wrapper:   ics4Wrapper,
		ChannelKeeper: channelKeeper,
		PortKeeper:    portKeeper,
	}
}

func (k Keeper) SendPacket(ctx sdk.Context, channelCap *capabilitytypes.Capability, packet exported.PacketI) error {
	fmt.Println("Sending package from keeper")
	return k.ChannelKeeper.SendPacket(ctx, channelCap, packet)
}

func (k Keeper) GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel ibcchanneltypes.Channel, found bool) {
	return k.ChannelKeeper.GetChannel(ctx, srcPort, srcChan)
}

func (k Keeper) GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool) {
	return k.GetNextSequenceSend(ctx, portID, channelID)
}
func (k Keeper) GetPacketCommitment(ctx sdk.Context, portID, channelID string, sequence uint64) []byte {
	return k.ChannelKeeper.GetPacketCommitment(ctx, portID, channelID, sequence)
}

func (k Keeper) BindPort(ctx sdk.Context, portID string) *capabilitytypes.Capability {
	return k.PortKeeper.BindPort(ctx, portID)
}
