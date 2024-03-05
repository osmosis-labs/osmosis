package keeper

import (
	"context"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (m msgServer) InboundTransfer(
	ctx context.Context,
	msg *types.MsgInboundTransfer,
) (*types.MsgInboundTransferResponse, error) {
	return new(types.MsgInboundTransferResponse), m.Keeper.InboundTransfer(ctx, *msg)
}

func (m msgServer) OutboundTransfer(
	ctx context.Context,
	msg *types.MsgOutboundTransfer,
) (*types.MsgOutboundTransferResponse, error) {
	return new(types.MsgOutboundTransferResponse), m.Keeper.OutboundTransfer(ctx, *msg)
}

func (m msgServer) UpdateParams(
	ctx context.Context,
	msg *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	return new(types.MsgUpdateParamsResponse), m.Keeper.UpdateParams(ctx, *msg)
}

func (m msgServer) ChangeAssetStatus(
	ctx context.Context,
	msg *types.MsgChangeAssetStatus,
) (*types.MsgChangeAssetStatusResponse, error) {
	return new(types.MsgChangeAssetStatusResponse), m.Keeper.ChangeAssetStatus(ctx, *msg)
}
