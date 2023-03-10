package concentrated_liquidity

import (
	"context"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

type msgServer struct {
	keeper *Keeper
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

func NewMsgCreatorServerImpl(keeper *Keeper) clmodel.MsgCreatorServer {
	return &msgServer{
		keeper: keeper,
	}
}

var (
	_ types.MsgServer          = msgServer{}
	_ clmodel.MsgCreatorServer = msgServer{}
)

// CreateConcentratedPool attempts to create a pool returning a MsgCreateConcentratedPoolResponse or an error upon failure.
// The pool creation fee is used to fund the community pool.
// It will create a dedicated module account for the pool and sends the initial liquidity to the created module account.
func (server msgServer) CreateConcentratedPool(goCtx context.Context, msg *clmodel.MsgCreateConcentratedPool) (*clmodel.MsgCreateConcentratedPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, denomExists := server.keeper.bankKeeper.GetDenomMetaData(ctx, msg.Denom0)
	if !denomExists {
		return nil, fmt.Errorf("received denom0 with invalid metadata: %s", msg.Denom0)
	}

	_, denomExists = server.keeper.bankKeeper.GetDenomMetaData(ctx, msg.Denom1)
	if !denomExists {
		return nil, fmt.Errorf("received denom1 with invalid metadata: %s", msg.Denom1)
	}

	poolId, err := server.keeper.poolmanagerKeeper.CreatePool(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &clmodel.MsgCreateConcentratedPoolResponse{PoolID: poolId}, nil
}

// TODO: tests, including events
func (server msgServer) CreatePosition(goCtx context.Context, msg *types.MsgCreatePosition) (*types.MsgCreatePositionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	actualAmount0, actualAmount1, liquidityCreated, err := server.keeper.createPosition(ctx, msg.PoolId, sender, msg.TokenDesired0.Amount, msg.TokenDesired1.Amount, msg.TokenMinAmount0, msg.TokenMinAmount1, msg.LowerTick, msg.UpperTick, msg.FreezeDuration)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
		sdk.NewEvent(
			types.TypeEvtWithdrawPosition,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(msg.PoolId, 10)),
			sdk.NewAttribute(types.AttributeAmount0, actualAmount0.String()),
			sdk.NewAttribute(types.AttributeAmount1, actualAmount1.String()),
			sdk.NewAttribute(types.AttributeLiquidity, liquidityCreated.String()),
			sdk.NewAttribute(types.AttributeLowerTick, strconv.FormatInt(msg.LowerTick, 10)),
			sdk.NewAttribute(types.AttributeUpperTick, strconv.FormatInt(msg.UpperTick, 10)),
		),
	})

	return &types.MsgCreatePositionResponse{Amount0: actualAmount0, Amount1: actualAmount1, LiquidityCreated: liquidityCreated}, nil
}

// TODO: tests, including events
func (server msgServer) WithdrawPosition(goCtx context.Context, msg *types.MsgWithdrawPosition) (*types.MsgWithdrawPositionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	amount0, amount1, err := server.keeper.withdrawPosition(ctx, msg.PoolId, sender, msg.LowerTick, msg.UpperTick, msg.JoinTime, msg.FreezeDuration, msg.LiquidityAmount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
		sdk.NewEvent(
			types.TypeEvtWithdrawPosition,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(msg.PoolId, 10)),
			sdk.NewAttribute(types.AttributeLiquidity, msg.LiquidityAmount.String()),
			sdk.NewAttribute(types.AttributeAmount0, amount0.String()),
			sdk.NewAttribute(types.AttributeAmount1, amount1.String()),
			sdk.NewAttribute(types.AttributeLowerTick, strconv.FormatInt(msg.LowerTick, 10)),
			sdk.NewAttribute(types.AttributeUpperTick, strconv.FormatInt(msg.UpperTick, 10)),
		),
	})

	return &types.MsgWithdrawPositionResponse{Amount0: amount0, Amount1: amount1}, nil
}

func (server msgServer) CollectFees(goCtx context.Context, msg *types.MsgCollectFees) (*types.MsgCollectFeesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	collectedFees, err := server.keeper.collectFees(ctx, msg.PoolId, sender, msg.LowerTick, msg.UpperTick)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
		sdk.NewEvent(
			types.TypeEvtCollectFees,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(msg.PoolId, 10)),
			sdk.NewAttribute(types.AttributeKeyTokensOut, collectedFees.String()),
			sdk.NewAttribute(types.AttributeLowerTick, strconv.FormatInt(msg.LowerTick, 10)),
			sdk.NewAttribute(types.AttributeUpperTick, strconv.FormatInt(msg.UpperTick, 10)),
		),
	})

	return &types.MsgCollectFeesResponse{CollectedFees: collectedFees}, nil
}
