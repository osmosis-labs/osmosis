package keeper

import (
	"context"

	"github.com/c-osmosis/osmosis/x/gamm/utils"
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

func (server msgServer) JoinPool(goCtx context.Context, msg *types.MsgJoinPool) (*types.MsgJoinPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	err = server.keeper.JoinPool(ctx, sender, msg.TargetPoolId, msg.PoolAmountOut, msg.MaxAmountsIn)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtPoolCreated,
			sdk.NewAttribute(types.AttributeKeyPoolId, utils.Uint64ToString(msg.TargetPoolId)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
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

	err = server.keeper.ExitPool(ctx, sender, msg.TargetPoolId, msg.PoolAmountIn, msg.MinAmountsOut)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtPoolCreated,
			sdk.NewAttribute(types.AttributeKeyPoolId, utils.Uint64ToString(msg.TargetPoolId)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &types.MsgExitPoolResponse{}, nil
}

func (server msgServer) CreatePool(goCtx context.Context, msg *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	poolId, err := server.keeper.CreatePool(ctx, sender, msg.SwapFee, msg.LpToken, msg.BindTokens)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtPoolCreated,
			sdk.NewAttribute(types.AttributeKeyPoolId, utils.Uint64ToString(poolId)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgCreatePoolResponse{}, nil
}

func (server msgServer) SwapExactAmountIn(goCtx context.Context, msg *types.MsgSwapExactAmountIn) (*types.MsgSwapExactAmountInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	_, _, err = server.keeper.SwapExactAmountIn(ctx, sender, msg.TargetPoolId, msg.TokenIn, msg.TokenOutDenom, msg.MinAmountOut, msg.MaxPrice)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtPoolCreated,
			sdk.NewAttribute(types.AttributeKeyPoolId, utils.Uint64ToString(msg.TargetPoolId)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
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

	_, _, err = server.keeper.SwapExactAmountOut(ctx, sender, msg.TargetPoolId, msg.TokenInDenom, msg.MaxAmountIn, msg.TokenOut, msg.MaxPrice)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtPoolCreated,
			sdk.NewAttribute(types.AttributeKeyPoolId, utils.Uint64ToString(msg.TargetPoolId)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
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

	_, err = server.keeper.JoinPoolWithExternAmountIn(ctx, sender, msg.TargetPool, msg.TokenIn, msg.TokenAmountIn, msg.MinPoolAmountOut)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtPoolCreated,
			sdk.NewAttribute(types.AttributeKeyPoolId, utils.Uint64ToString(msg.TargetPool)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &types.MsgJoinSwapExternAmountInResponse{}, nil
}

func (server msgServer) JoinSwapPoolAmountOut(goCtx context.Context, msg *types.MsgJoinSwapPoolAmountOut) (*types.MsgJoinSwapPoolAmountOut, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	_, err = server.keeper.JoinPoolWithPoolAmountOut(ctx, sender, msg.TargetPool, msg.TokenIn, msg.PoolAmountOut, msg.MaxAmountIn)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtPoolCreated,
			sdk.NewAttribute(types.AttributeKeyPoolId, utils.Uint64ToString(msg.TargetPool)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &types.MsgJoinSwapPoolAmountOut{}, nil
}

func (server msgServer) ExitSwapExternAmountOut(goCtx context.Context, msg *types.MsgExitSwapExternAmountOut) (*types.MsgExitSwapExternAmountOut, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	_, err = server.keeper.ExitPoolWithExternAmountOut(ctx, sender, msg.TargetPool, msg.TokenOut, msg.TokenAmountOut, msg.MaxPoolAmountIn)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtPoolCreated,
			sdk.NewAttribute(types.AttributeKeyPoolId, utils.Uint64ToString(msg.TargetPool)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &types.MsgExitSwapExternAmountOut{}, nil
}

func (server msgServer) ExitSwapPoolAmountIn(goCtx context.Context, msg *types.MsgExitSwapPoolAmountIn) (*types.MsgExitSwapPoolAmountIn, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	_, err = server.keeper.ExitPoolWithPoolAmountIn(ctx, sender, msg.TargetPool, msg.TokenOut, msg.PoolAmountIn, msg.MinAmountOut)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtPoolCreated,
			sdk.NewAttribute(types.AttributeKeyPoolId, utils.Uint64ToString(msg.TargetPool)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &types.MsgExitSwapPoolAmountIn{}, nil
}
