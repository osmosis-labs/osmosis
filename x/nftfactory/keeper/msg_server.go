package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v12/x/nftfactory/types"
)

type msgServer struct {
	keeper *Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

var _ types.MsgServer = msgServer{}

func (server msgServer) CreateDenom(goCtx context.Context, msg *types.MsgCreateDenom) (*types.MsgCreateDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.CreateDenom(ctx, msg.Id, msg.Sender, msg.DenomName, msg.Data)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateDenomResponse{}, nil
}

func (server msgServer) Mint(goCtx context.Context, msg *types.MsgMint) (*types.MsgMintResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.Mint(ctx, msg.Id, msg.Sender, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgMintResponse{}, nil
}
