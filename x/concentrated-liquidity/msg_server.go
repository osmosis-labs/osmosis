package concentrated_liquidity

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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

// CreateConcentratedPool attempts to create a concentrated liquidity pool via the poolmanager module, returning a MsgCreateConcentratedPoolResponse or an error upon failure.
// The pool creation fee is used to fund the community pool. It will also create a dedicated module account for the pool.
func (server msgServer) CreateConcentratedPool(goCtx context.Context, msg *clmodel.MsgCreateConcentratedPool) (*clmodel.MsgCreateConcentratedPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

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

	positionId, actualAmount0, actualAmount1, liquidityCreated, joinTime, err := server.keeper.createPosition(ctx, msg.PoolId, sender, msg.TokenDesired0.Amount, msg.TokenDesired1.Amount, msg.TokenMinAmount0, msg.TokenMinAmount1, msg.LowerTick, msg.UpperTick)
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

func (server msgServer) AddToPosition(goCtx context.Context, msg *types.MsgAddToPosition) (*types.MsgAddToPositionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	positionId, actualAmount0, actualAmount1, err := server.keeper.addToPosition(ctx, sender, msg.PositionId, msg.TokenDesired0.Amount, msg.TokenDesired1.Amount)
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

	return &types.MsgAddToPositionResponse{PositionId: positionId, Amount0: actualAmount0, Amount1: actualAmount1}, nil
}

// TODO: tests, including events
func (server msgServer) WithdrawPosition(goCtx context.Context, msg *types.MsgWithdrawPosition) (*types.MsgWithdrawPositionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	amount0, amount1, err := server.keeper.WithdrawPosition(ctx, sender, msg.PositionId, msg.LiquidityAmount)
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

// CollectFees collects the fees earned by each position ID provided and sends them to the owner's account.
// Returns error if one of the provided position IDs do not exist or if the function fails to get the fee accumulator.
func (server msgServer) CollectFees(goCtx context.Context, msg *types.MsgCollectFees) (*types.MsgCollectFeesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	totalCollectedFees := sdk.NewCoins()
	for _, positionId := range msg.PositionIds {
		collectedFees, err := server.keeper.collectFees(ctx, sender, positionId)
		if err != nil {
			return nil, err
		}
		totalCollectedFees = totalCollectedFees.Add(collectedFees...)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
		sdk.NewEvent(
			types.TypeEvtTotalCollectFees,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyTokensOut, totalCollectedFees.String()),
		),
	})

	return &types.MsgCollectFeesResponse{CollectedFees: totalCollectedFees}, nil
}

// CollectIncentives collects incentives for all given PositionIds in a given range that belong to sender.
func (server msgServer) CollectIncentives(goCtx context.Context, msg *types.MsgCollectIncentives) (*types.MsgCollectIncentivesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	totalCollectedIncentives := sdk.NewCoins()
	totalForefeitedIncentives := sdk.NewCoins()
	for _, positionId := range msg.PositionIds {
		collectedIncentives, forfeitedIncentives, err := server.keeper.collectIncentives(ctx, sender, positionId)
		if err != nil {
			return nil, err
		}
		totalCollectedIncentives = totalCollectedIncentives.Add(collectedIncentives...)
		totalForefeitedIncentives = totalForefeitedIncentives.Add(forfeitedIncentives...)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
		sdk.NewEvent(
			types.TypeEvtTotalCollectIncentives,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyTokensOut, totalCollectedIncentives.String()),
		),
	})

	return &types.MsgCollectIncentivesResponse{CollectedIncentives: totalCollectedIncentives, ForfeitedIncentives: totalForefeitedIncentives}, nil
}

func (server msgServer) CreateIncentive(goCtx context.Context, msg *types.MsgCreateIncentive) (*types.MsgCreateIncentiveResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	incentiveRecord, err := server.keeper.CreateIncentive(ctx, msg.PoolId, sender, msg.IncentiveDenom, msg.IncentiveAmount, msg.EmissionRate, msg.StartTime, msg.MinUptime)
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
		IncentiveAmount: incentiveRecord.IncentiveRecordBody.RemainingAmount,
		EmissionRate:    incentiveRecord.IncentiveRecordBody.EmissionRate,
		StartTime:       incentiveRecord.IncentiveRecordBody.StartTime,
		MinUptime:       incentiveRecord.MinUptime,
	}, nil
}

func (server msgServer) FungifyChargedPositions(goCtx context.Context, msg *types.MsgFungifyChargedPositions) (*types.MsgFungifyChargedPositionsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	newPositionId, err := server.keeper.fungifyChargedPosition(ctx, sender, msg.PositionIds)
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
			types.TypeEvtFungifyChargedPosition,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeInputPositionIds, strings.Trim(strings.Join(strings.Fields(fmt.Sprint(msg.PositionIds)), ","), "[]")),
			sdk.NewAttribute(types.AttributeOutputPositionId, strconv.FormatUint(newPositionId, 10)),
		),
	})

	return &types.MsgFungifyChargedPositionsResponse{
		NewPositionId: newPositionId,
	}, nil
}
