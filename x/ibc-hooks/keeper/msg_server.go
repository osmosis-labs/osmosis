package keeper

import (
	"context"
	"fmt"
	"strconv"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/x/ibc-hooks/types"
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

// Gov messages

func (server msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	govAddr := server.Keeper.accountKeeper.GetModuleAddress(govtypes.ModuleName)
	if msg.Sender != govAddr.String() {
		return nil, fmt.Errorf("unauthorized: expected sender to be %s, got %s", govAddr, msg.Sender)
	}

	server.Keeper.SetParams(ctx, msg.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}
