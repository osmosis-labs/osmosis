package swaprouter

import (
	"context"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	balancerv2 "github.com/osmosis-labs/osmosis/v13/x/gamm/pool-models/balancer/v2"
	stableswapv2 "github.com/osmosis-labs/osmosis/v13/x/gamm/pool-models/stableswap/v2"
	gammtypes "github.com/osmosis-labs/osmosis/v13/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

type msgServer struct {
	keeper *Keeper
}

var (
	_ balancerv2.MsgCreatorServer   = (*msgServer)(nil)
	_ stableswapv2.MsgCreatorServer = (*msgServer)(nil)
)

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

func NewBalancerMsgServerImpl(keeper *Keeper) balancerv2.MsgCreatorServer {
	return &msgServer{
		keeper: keeper,
	}
}

func NewStableswapMsgServerImpl(keeper *Keeper) stableswapv2.MsgCreatorServer {
	return &msgServer{
		keeper: keeper,
	}
}

// CreateBalancerPool is a create balancer pool message.
func (server msgServer) CreateBalancerPool(goCtx context.Context, msg *balancerv2.MsgCreateBalancerPool) (*balancerv2.MsgCreateBalancerPoolResponse, error) {
	poolId, err := server.CreatePool(goCtx, msg)
	if err != nil {
		return nil, err
	}
	return &balancerv2.MsgCreateBalancerPoolResponse{PoolID: poolId}, nil
}

func (server msgServer) CreateStableswapPool(goCtx context.Context, msg *stableswapv2.MsgCreateStableswapPool) (*stableswapv2.MsgCreateStableswapPoolResponse, error) {
	poolId, err := server.CreatePool(goCtx, msg)
	if err != nil {
		return nil, err
	}
	return &stableswapv2.MsgCreateStableswapPoolResponse{PoolID: poolId}, nil
}

// CreatePool attempts to create a pool returning the newly created pool ID or an error upon failure.
// The pool creation fee is used to fund the community pool.
// It will create a dedicated module account for the pool and sends the initial liquidity to the created module account.
func (server msgServer) CreatePool(goCtx context.Context, msg types.CreatePoolMsg) (poolId uint64, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	poolId, err = server.keeper.CreatePool(ctx, msg)
	if err != nil {
		return 0, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			gammtypes.TypeEvtPoolCreated,
			sdk.NewAttribute(gammtypes.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.PoolCreator().String()),
		),
	})

	return poolId, nil
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
