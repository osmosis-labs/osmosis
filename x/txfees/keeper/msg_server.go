package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/txfees/types"
)

type msgServer struct {
	keeper *Keeper
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

// SetFeeTokens sets the provided fee tokens for the chain. The sender must be whitelisted to set fee tokens.
func (server msgServer) SetFeeTokens(goCtx context.Context, msg *types.MsgSetFeeTokens) (*types.MsgSetFeeTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.SenderValidationSetFeeTokens(ctx, msg.Sender, msg.FeeTokens)
	if err != nil {
		return nil, err
	}

	return &types.MsgSetFeeTokensResponse{}, nil
}
