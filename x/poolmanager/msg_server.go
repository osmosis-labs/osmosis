package poolmanager

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/osmosis-labs/osmosis/osmoutils/observability"

	"github.com/osmosis-labs/osmosis/v23/x/poolmanager/types"
)

var tracer = otel.Tracer(types.ModuleName)

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
	ctx, span := observability.InitSDKCtxWithSpan(goCtx, tracer, "msg_swap_exact_amount_in")
	defer span.End()

	// Set input span attributes.
	span.SetAttributes(
		attribute.Stringer("token_in", msg.TokenIn),
		attribute.String("token_out_denom", msg.TokenOutDenom()),
		attribute.String("sender", msg.Sender),
		attribute.Stringer("token_out_min_amount", msg.TokenOutMinAmount),
	)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	tokenOutAmount, err := server.keeper.RouteExactAmountIn(ctx, sender, msg.Routes, msg.TokenIn, msg.TokenOutMinAmount)
	if err != nil {
		return nil, err
	}

	// Swap event is handled elsewhere

	// Set output span attributes.
	span.SetAttributes(
		attribute.Stringer("token_out_amount", tokenOutAmount),
	)

	return &types.MsgSwapExactAmountInResponse{TokenOutAmount: tokenOutAmount}, nil
}

// TODO: spec and tests, including events
func (server msgServer) SwapExactAmountOut(goCtx context.Context, msg *types.MsgSwapExactAmountOut) (*types.MsgSwapExactAmountOutResponse, error) {
	ctx, span := observability.InitSDKCtxWithSpan(goCtx, tracer, "msg_swap_exact_amount_out")
	defer span.End()

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	// Set input span attributes.
	span.SetAttributes(
		attribute.Stringer("token_out", msg.TokenOut),
		attribute.String("token_out_denom", msg.TokenInDenom()),
		attribute.String("sender", msg.Sender),
		attribute.Stringer("token_in_max_amount", msg.TokenInMaxAmount),
	)

	tokenInAmount, err := server.keeper.RouteExactAmountOut(ctx, sender, msg.Routes, msg.TokenInMaxAmount, msg.TokenOut)
	if err != nil {
		return nil, err
	}

	// Set output span attributes.
	span.SetAttributes(
		attribute.Stringer("token_in_amount", tokenInAmount),
	)

	// Swap event is handled elsewhere

	return &types.MsgSwapExactAmountOutResponse{TokenInAmount: tokenInAmount}, nil
}

func (server msgServer) SplitRouteSwapExactAmountIn(goCtx context.Context, msg *types.MsgSplitRouteSwapExactAmountIn) (*types.MsgSplitRouteSwapExactAmountInResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	ctx, span := observability.InitSDKCtxWithSpan(goCtx, tracer, "msg_split_route_swap_exact_amount_in")
	defer span.End()

	// Set input span attributes.
	span.SetAttributes(
		attribute.String("token_in_denom", msg.TokenInDenom),
		attribute.String("sender", msg.Sender),
		attribute.Stringer("token_out_min_amount", msg.TokenOutMinAmount),
	)

	tokenOutAmount, err := server.keeper.SplitRouteExactAmountIn(ctx, sender, msg.Routes, msg.TokenInDenom, msg.TokenOutMinAmount)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(
		attribute.Stringer("token_out_amount", tokenOutAmount),
	)

	// Swap event is handled in each pool module's SwapExactAmountIn

	return &types.MsgSplitRouteSwapExactAmountInResponse{TokenOutAmount: tokenOutAmount}, nil
}

func (server msgServer) SplitRouteSwapExactAmountOut(goCtx context.Context, msg *types.MsgSplitRouteSwapExactAmountOut) (*types.MsgSplitRouteSwapExactAmountOutResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	ctx, span := observability.InitSDKCtxWithSpan(goCtx, tracer, "msg_split_route_swap_exact_amount_out")
	defer span.End()

	// Set input span attributes.
	span.SetAttributes(
		attribute.String("token_out_denom", msg.TokenOutDenom),
		attribute.String("sender", msg.Sender),
		attribute.Stringer("token_in_max_amount", msg.TokenInMaxAmount),
	)

	tokenInAmount, err := server.keeper.SplitRouteExactAmountOut(ctx, sender, msg.Routes, msg.TokenOutDenom, msg.TokenInMaxAmount)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(
		attribute.Stringer("token_in_amount", tokenInAmount),
	)

	// Swap event is handled in each pool module's SwapExactAmountOut

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

	return &types.MsgSetDenomPairTakerFeeResponse{Success: true}, nil
}
