package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/farm/types"
)

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) AllocateAssets(goCtx context.Context, msg *types.MsgAllocateAssets) (*types.MsgAllocateAssetsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}

	err = k.AllocateAssetsToFarm(ctx, msg.FarmId, from, msg.Assets)
	if err != nil {
		return nil, err
	}

	return &types.MsgAllocateAssetsResponse{}, nil
}

func (k msgServer) WithdrawRewards(goCtx context.Context, msg *types.MsgWithdrawRewards) (*types.MsgWithdrawRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}

	_, err = k.WithdrawRewardsFromFarm(ctx, msg.FarmId, from)
	if err != nil {
		return nil, err
	}

	return &types.MsgWithdrawRewardsResponse{}, nil
}
