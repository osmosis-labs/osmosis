package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/gamm/types"
)

type msgServer struct {
	keeper Keeper
}

func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

var _ types.MsgServer = msgServer{}

func (server msgServer) CreatePool(goCtx context.Context, msg *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	poolId, err := server.keeper.CreatePool(ctx, sender, msg.PoolParams, msg.PoolAssets, msg.FuturePoolGovernor)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtPoolCreated,
			sdk.NewAttribute(types.AttributeKeyPoolId, fmt.Sprintf("%d", poolId)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgCreatePoolResponse{}, nil
}

func (server msgServer) JoinPool(goCtx context.Context, msg *types.MsgJoinPool) (*types.MsgJoinPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	err = server.keeper.JoinPool(ctx, sender, msg.PoolId, msg.ShareOutAmount, msg.TokenInMaxs)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgJoinPoolResponse{}, nil
}

func (server msgServer) ExitPool(goCtx context.Context, msg *types.MsgExitPool) (*types.MsgExitPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	err = server.keeper.ExitPool(ctx, sender, msg.PoolId, msg.ShareInAmount, msg.TokenOutMins)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgExitPoolResponse{}, nil
}

func (server msgServer) SwapExactAmountIn(goCtx context.Context, msg *types.MsgSwapExactAmountIn) (*types.MsgSwapExactAmountInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	_, err = server.keeper.MultihopSwapExactAmountIn(ctx, sender, msg.Routes, msg.TokenIn, msg.TokenOutMinAmount)

	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgSwapExactAmountInResponse{}, nil
}

func (server msgServer) SwapExactAmountOut(goCtx context.Context, msg *types.MsgSwapExactAmountOut) (*types.MsgSwapExactAmountOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	_, err = server.keeper.MultihopSwapExactAmountOut(ctx, sender, msg.Routes, msg.TokenInMaxAmount, msg.TokenOut)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgSwapExactAmountOutResponse{}, nil
}

func (server msgServer) JoinSwapExternAmountIn(goCtx context.Context, msg *types.MsgJoinSwapExternAmountIn) (*types.MsgJoinSwapExternAmountInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	_, err = server.keeper.JoinSwapExternAmountIn(ctx, sender, msg.PoolId, msg.TokenIn, msg.ShareOutMinAmount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgJoinSwapExternAmountInResponse{}, nil
}

func (server msgServer) JoinSwapShareAmountOut(goCtx context.Context, msg *types.MsgJoinSwapShareAmountOut) (*types.MsgJoinSwapShareAmountOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	_, err = server.keeper.JoinSwapShareAmountOut(ctx, sender, msg.PoolId, msg.TokenInDenom, msg.ShareOutAmount, msg.TokenInMaxAmount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgJoinSwapShareAmountOutResponse{}, nil
}

func (server msgServer) ExitSwapExternAmountOut(goCtx context.Context, msg *types.MsgExitSwapExternAmountOut) (*types.MsgExitSwapExternAmountOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	_, err = server.keeper.ExitSwapExternAmountOut(ctx, sender, msg.PoolId, msg.TokenOut, msg.ShareInMaxAmount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgExitSwapExternAmountOutResponse{}, nil
}

func (server msgServer) ExitSwapShareAmountIn(goCtx context.Context, msg *types.MsgExitSwapShareAmountIn) (*types.MsgExitSwapShareAmountInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	_, err = server.keeper.ExitSwapShareAmountIn(ctx, sender, msg.PoolId, msg.TokenOutDenom, msg.ShareInAmount, msg.TokenOutMinAmount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgExitSwapShareAmountInResponse{}, nil
}
