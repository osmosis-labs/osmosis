package keeper

import (
	"context"

	"github.com/osmosis-labs/osmosis/v12/x/nftfactory/types"
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

func (server msgServer) CreateDenom(goCtx context.Context, msg *types.MsgCreateDenom) (*types.MsgCreateDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	denom, err := server.Keeper.CreateDenom(ctx, msg.Id, msg.Sender, msg.DenomName, msg.Data)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateDenomResponse{}, nil
}
