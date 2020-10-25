package gamm

import (
	"github.com/c-osmosis/osmosis/x/gamm/keeper"
	"github.com/c-osmosis/osmosis/x/gamm/types"
	"github.com/c-osmosis/osmosis/x/gamm/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// TODO: proto 정의가 완료되는 대로 적용하기
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgSwapExactAmountIn:
			return handleSwapExactAmountIn(ctx, k, msg)
		case *types.MsgSwapExactAmountOut:
			return handleSwapExactAmountOut(ctx, k, msg)
		case *types.MsgJoinPool:
			return handleJoinPool(ctx, k, msg)
		case *types.MsgExitPool:
			return handleExitPool(ctx, k, msg)
		case *types.MsgCreatePool:
			return handleCreatePool(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized gamm message type: %T", msg)
		}
	}
}

func handleCreatePool(ctx sdk.Context, k keeper.Keeper, msg *types.MsgCreatePool) (*sdk.Result, error) {
	poolId, err := k.CreatePool(ctx, msg.Sender, msg.SwapFee, msg.LpToken, msg.BindTokens)
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
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	})

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleSwapExactAmountIn(ctx sdk.Context, k keeper.Keeper, msg *types.MsgSwapExactAmountIn) (*sdk.Result, error) {
	_, _, err := k.SwapExactAmountIn(ctx, msg.Sender, msg.TargetPoolId, msg.TokenIn, msg.TokenOutDenom, msg.MinAmountOut, msg.MaxPrice)
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

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleSwapExactAmountOut(ctx sdk.Context, k keeper.Keeper, msg *types.MsgSwapExactAmountOut) (*sdk.Result, error) {
	_, _, err := k.SwapExactAmountOut(ctx, msg.Sender, msg.TargetPoolId, msg.TokenInDenom, msg.MaxAmountIn, msg.TokenOut, msg.MaxPrice)
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

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleJoinPool(ctx sdk.Context, k keeper.Keeper, msg *types.MsgJoinPool) (*sdk.Result, error) {
	err := k.JoinPool(ctx, msg.Sender, msg.TargetPoolId, msg.PoolAmountOut, msg.MaxAmountsIn)
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

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleExitPool(ctx sdk.Context, k keeper.Keeper, msg *types.MsgExitPool) (*sdk.Result, error) {
	err := k.ExitPool(ctx, msg.Sender, msg.TargetPoolId, msg.PoolAmountIn, msg.MinAmountsOut)
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

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}
