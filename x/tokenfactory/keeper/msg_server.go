package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/x/tokenfactory/types"
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

	err := server.Keeper.CreateDenom(ctx, msg.Sender, msg.Nonce)

	if err != nil {
		return nil, err
	}

	// TODO: events
	// ctx.EventManager().EmitEvents(sdk.Events{})

	return &types.MsgCreateDenomResponse{}, nil
}

var _ types.MsgServer = msgServer{}
