package poolmanager

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
)

type msgServer struct {
	keeper *Keeper
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

// TODO: spec and tests, including events
func (server msgServer) SwapExactAmountIn(goCtx context.Context, msg *types.MsgSwapExactAmountIn) (*types.MsgSwapExactAmountInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	tokenOutAmount, err := server.keeper.RouteExactAmountIn(ctx, sender, msg.Routes, msg.TokenIn, msg.TokenOutMinAmount)
	if err != nil {
		return nil, err
	}

	// Swap event is handled elsewhere
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgSwapExactAmountInResponse{TokenOutAmount: tokenOutAmount}, nil
}

// TODO: spec and tests, including events
func (server msgServer) SwapExactAmountOut(goCtx context.Context, msg *types.MsgSwapExactAmountOut) (*types.MsgSwapExactAmountOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	tokenInAmount, err := server.keeper.RouteExactAmountOut(ctx, sender, msg.Routes, msg.TokenInMaxAmount, msg.TokenOut)
	if err != nil {
		return nil, err
	}

	// Swap event is handled elsewhere
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgSwapExactAmountOutResponse{TokenInAmount: tokenInAmount}, nil
}

func (server msgServer) SplitRouteSwapExactAmountIn(goCtx context.Context, msg *types.MsgSplitRouteSwapExactAmountIn) (*types.MsgSplitRouteSwapExactAmountInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	tokenOutAmount, err := server.keeper.SplitRouteExactAmountIn(ctx, sender, msg.Routes, msg.TokenInDenom, msg.TokenOutMinAmount)
	if err != nil {
		return nil, err
	}

	// Swap event is handled in each pool module's SwapExactAmountIn
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgSplitRouteSwapExactAmountInResponse{TokenOutAmount: tokenOutAmount}, nil
}

func (server msgServer) SplitRouteSwapExactAmountOut(goCtx context.Context, msg *types.MsgSplitRouteSwapExactAmountOut) (*types.MsgSplitRouteSwapExactAmountOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	tokenInAmount, err := server.keeper.SplitRouteExactAmountOut(ctx, sender, msg.Routes, msg.TokenOutDenom, msg.TokenInMaxAmount)
	if err != nil {
		return nil, err
	}

	// Swap event is handled in each pool module's SwapExactAmountOut
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgSplitRouteSwapExactAmountOutResponse{TokenInAmount: tokenInAmount}, nil
}

func (server msgServer) SetDenomPairTakerFee(goCtx context.Context, msg *types.MsgSetDenomPairTakerFee) (*types.MsgSetDenomPairTakerFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	for _, denomPair := range msg.DenomPairTakerFee {
		err := server.keeper.SenderValidationSetDenomPairTakerFee(ctx, msg.Sender, denomPair.Denom0, denomPair.Denom1, denomPair.TakerFee)
		if err != nil {
			return nil, err
		}
	}

	// Set denom pair taker fee event is handled in each iteration of the loop above
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgSetDenomPairTakerFeeResponse{Success: true}, nil
}
