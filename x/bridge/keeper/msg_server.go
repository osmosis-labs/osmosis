package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
	goCtx context.Context,
	msg *types.MsgInboundTransfer,
) (*types.MsgInboundTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	return new(types.MsgInboundTransferResponse), m.Keeper.InboundTransfer(ctx, *msg)
}

func (m msgServer) OutboundTransfer(
	goCtx context.Context,
	msg *types.MsgOutboundTransfer,
) (*types.MsgOutboundTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	return new(types.MsgOutboundTransferResponse), m.Keeper.OutboundTransfer(ctx, *msg)
}

func (m msgServer) UpdateParams(
	goCtx context.Context,
	msg *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	result := m.Keeper.UpdateParams(ctx, msg.NewParams)

	err := ctx.EventManager().EmitTypedEvent(&types.EventUpdateParams{
		NewSigners:     msg.NewParams.Signers,
		CreatedSigners: result.signersToCreate,
		DeletedSigners: result.signersToDelete,
		NewAssets:      msg.NewParams.Assets,
		CreatedAssets:  result.assetsToCreate,
		DeletedAssets:  result.assetsToDelete,
	})
	if err != nil {
		return nil, err
	}

	return new(types.MsgUpdateParamsResponse), nil
}

func (m msgServer) ChangeAssetStatus(
	goCtx context.Context,
	msg *types.MsgChangeAssetStatus,
) (*types.MsgChangeAssetStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	return new(types.MsgChangeAssetStatusResponse), m.Keeper.ChangeAssetStatus(ctx, *msg)
}
