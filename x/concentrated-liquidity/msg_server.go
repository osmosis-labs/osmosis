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

	positionId, actualAmount0, actualAmount1, liquidityCreated, joinTime, err := server.keeper.createPosition(ctx, msg.PoolId, sender, msg.TokenDesired0.Amount, msg.TokenDesired1.Amount, msg.TokenMinAmount0, msg.TokenMinAmount1, msg.LowerTick, msg.UpperTick, msg.FreezeDuration)
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

	// Note: create position event is emitted in keeper.createPosition(...)

	return &types.MsgCreatePositionResponse{PositionId: positionId, Amount0: actualAmount0, Amount1: actualAmount1, JoinTime: joinTime, LiquidityCreated: liquidityCreated}, nil
}

// TODO: tests, including events
func (server msgServer) WithdrawPosition(goCtx context.Context, msg *types.MsgWithdrawPosition) (*types.MsgWithdrawPositionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	amount0, amount1, err := server.keeper.withdrawPosition(ctx, sender, msg.PositionId, msg.LiquidityAmount)
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

	// Note: withdraw position event is emitted in keeper.withdrawPosition(...)

	return &types.MsgWithdrawPositionResponse{Amount0: amount0, Amount1: amount1}, nil
}

func (server msgServer) CollectFees(goCtx context.Context, msg *types.MsgCollectFees) (*types.MsgCollectFeesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	collectedFees, err := server.keeper.collectFees(ctx, sender, msg.PositionId)
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
			sdk.NewAttribute(types.AttributeKeyTokensOut, collectedFees.String()),
		),
	})

	return &types.MsgCollectFeesResponse{CollectedFees: collectedFees}, nil
}

// CollectIncentives collects incentives for all positions in given range that belong to sender
func (server msgServer) CollectIncentives(goCtx context.Context, msg *types.MsgCollectIncentives) (*types.MsgCollectIncentivesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	collectedIncentives, err := server.keeper.collectIncentives(ctx, sender, msg.PositionId)
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
			types.TypeEvtCollectIncentives,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyTokensOut, collectedIncentives.String()),
		),
	})

	return &types.MsgCollectIncentivesResponse{CollectedIncentives: collectedIncentives}, nil
}

func (server msgServer) CreateIncentive(goCtx context.Context, msg *types.MsgCreateIncentive) (*types.MsgCreateIncentiveResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	incentiveRecord, err := server.keeper.createIncentive(ctx, msg.PoolId, sender, msg.IncentiveDenom, msg.IncentiveAmount, msg.EmissionRate, msg.StartTime, msg.MinUptime)
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
			types.TypeEvtCreateIncentive,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(msg.PoolId, 10)),
			sdk.NewAttribute(types.AttributeIncentiveDenom, msg.IncentiveDenom),
			sdk.NewAttribute(types.AttributeIncentiveAmount, msg.IncentiveAmount.String()),
			sdk.NewAttribute(types.AttributeIncentiveEmissionRate, msg.EmissionRate.String()),
			sdk.NewAttribute(types.AttributeIncentiveStartTime, msg.StartTime.String()),
			sdk.NewAttribute(types.AttributeIncentiveMinUptime, msg.MinUptime.String()),
		),
	})

	return &types.MsgCreateIncentiveResponse{
		IncentiveDenom:  incentiveRecord.IncentiveDenom,
		IncentiveAmount: incentiveRecord.RemainingAmount,
		EmissionRate:    incentiveRecord.EmissionRate,
		StartTime:       incentiveRecord.StartTime,
		MinUptime:       incentiveRecord.MinUptime,
	}, nil
}
