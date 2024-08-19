package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v25/x/txfees/types"
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

func (server msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	govAddr := server.keeper.accountKeeper.GetModuleAddress(govtypes.ModuleName)
	if msg.Sender != govAddr.String() {
		return nil, fmt.Errorf("unauthorized: expected sender to be %s, got %s", govAddr, msg.Sender)
	}

	server.keeper.SetParams(ctx, msg.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}
