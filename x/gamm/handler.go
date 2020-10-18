package gamm

import (
	"github.com/c-osmosis/osmosis/x/gamm/keeper"
	"github.com/c-osmosis/osmosis/x/gamm/types"
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
	return nil, sdkerrors.Wrapf(sdkerrors.ErrPanic, "unimplemented")
}

func handleSwapExactAmountIn(ctx sdk.Context, k keeper.Keeper, msg *types.MsgSwapExactAmountIn) (*sdk.Result, error) {
	return nil, sdkerrors.Wrapf(sdkerrors.ErrPanic, "unimplemented")
}

func handleSwapExactAmountOut(ctx sdk.Context, k keeper.Keeper, msg *types.MsgSwapExactAmountOut) (*sdk.Result, error) {
	return nil, sdkerrors.Wrapf(sdkerrors.ErrPanic, "unimplemented")
}

func handleJoinPool(ctx sdk.Context, k keeper.Keeper, msg *types.MsgJoinPool) (*sdk.Result, error) {
	return nil, sdkerrors.Wrapf(sdkerrors.ErrPanic, "unimplemented")
}

func handleExitPool(ctx sdk.Context, k keeper.Keeper, msg *types.MsgExitPool) (*sdk.Result, error) {
	return nil, sdkerrors.Wrapf(sdkerrors.ErrPanic, "unimplemented")
}
