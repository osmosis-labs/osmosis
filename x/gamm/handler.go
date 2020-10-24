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
	// TODO: 밑의 메소드에서 실제 자산을 옮기지않음. 자산 처리를 키퍼의 다른 함수나 이 핸들러 안에서 해야됨 일단은 이렇게 놔둠.
	_, _, err := k.SwapExactAmountIn(ctx, msg.Sender, msg.TargetPoolId, msg.TokenIn, msg.TokenAmountIn, msg.TokenOut, msg.MinAmountOut, msg.MaxPrice)
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
	// TODO: 밑의 메소드에서 실제 자산을 옮기지않음. 자산 처리를 키퍼의 다른 함수나 이 핸들러 안에서 해야됨 일단은 이렇게 놔둠.
	_, _, err := k.SwapExactAmountOut(ctx, msg.Sender, msg.TargetPoolId, msg.TokenIn, msg.MaxAmountIn, msg.TokenOut, msg.TokenAmountOut, msg.MaxPrice)
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
