package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/ibc-hooks/types"
	"strconv"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) EmitIBCAck(goCtx context.Context, msg *types.MsgEmitIBCAck) (*types.MsgEmitIBCAckResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.MsgEmitAckKey,
			sdk.NewAttribute(types.AttributeSender, msg.Sender),
			sdk.NewAttribute(types.AttributeChannel, msg.Channel),
			sdk.NewAttribute(types.AttributePacketSequence, strconv.FormatUint(msg.PacketSequence, 10)),
		),
	)

	ack, err := m.Keeper.EmitIBCAck(ctx, msg.Sender, msg.Channel, msg.PacketSequence)
	if err != nil {
		return nil, err
	}

	return &types.MsgEmitIBCAckResponse{ContractResult: string(ack), IbcAck: string(ack)}, nil
}
