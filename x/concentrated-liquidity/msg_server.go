package concentrated_liquidity

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	clmodel "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

type msgServer struct {
	keeper *Keeper
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

func NewMsgCreatorServerImpl(keeper *Keeper) clmodel.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

var (
	_ types.MsgServer   = msgServer{}
	_ clmodel.MsgServer = msgServer{}
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

func (server msgServer) CreatePosition(goCtx context.Context, msg *types.MsgCreatePosition) (*types.MsgCreatePositionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	positionData, err := server.keeper.CreatePosition(ctx, msg.PoolId, sender, msg.TokensProvided, msg.TokenMinAmount0, msg.TokenMinAmount1, msg.LowerTick, msg.UpperTick)
	if err != nil {
		return nil, err
	}

	// Note: create position event is emitted in keeper.createPosition(...)

	return &types.MsgCreatePositionResponse{PositionId: positionData.ID, Amount0: positionData.Amount0, Amount1: positionData.Amount1, LiquidityCreated: positionData.Liquidity, LowerTick: positionData.LowerTick, UpperTick: positionData.UpperTick}, nil
}

func (server msgServer) AddToPosition(goCtx context.Context, msg *types.MsgAddToPosition) (*types.MsgAddToPositionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	if msg.TokenMinAmount0.IsNil() {
		msg.TokenMinAmount0 = osmomath.ZeroInt()
	}
	if msg.TokenMinAmount1.IsNil() {
		msg.TokenMinAmount1 = osmomath.ZeroInt()
	}

	positionId, actualAmount0, actualAmount1, err := server.keeper.addToPosition(ctx, sender, msg.PositionId, msg.Amount0, msg.Amount1, msg.TokenMinAmount0, msg.TokenMinAmount1)
	if err != nil {
		return nil, err
	}

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

	// Note: withdraw position event is emitted in keeper.withdrawPosition(...)

	return &types.MsgWithdrawPositionResponse{Amount0: amount0, Amount1: amount1}, nil
}

// CollectSpreadRewards collects the fees earned by each position ID provided and sends them to the owner's account.
// Returns error if one of the provided position IDs do not exist or if the function fails to get the fee accumulator.
func (server msgServer) CollectSpreadRewards(goCtx context.Context, msg *types.MsgCollectSpreadRewards) (*types.MsgCollectSpreadRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	totalCollectedSpreadRewards := sdk.NewCoins()
	for _, positionId := range msg.PositionIds {
		collectedFees, err := server.keeper.collectSpreadRewards(ctx, sender, positionId)
		if err != nil {
			return nil, err
		}
		totalCollectedSpreadRewards = totalCollectedSpreadRewards.Add(collectedFees...)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtTotalCollectSpreadRewards,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyTokensOut, totalCollectedSpreadRewards.String()),
		),
	})

	return &types.MsgCollectSpreadRewardsResponse{CollectedSpreadRewards: totalCollectedSpreadRewards}, nil
}

// CollectIncentives collects incentives for all positions in given range that belong to sender
func (server msgServer) CollectIncentives(goCtx context.Context, msg *types.MsgCollectIncentives) (*types.MsgCollectIncentivesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	totalCollectedIncentives := sdk.NewCoins()
	totalForefeitedIncentives := sdk.NewCoins()
	for _, positionId := range msg.PositionIds {
		collectedIncentives, forfeitedIncentives, _, err := server.keeper.collectIncentives(ctx, sender, positionId)
		if err != nil {
			return nil, err
		}
		totalCollectedIncentives = totalCollectedIncentives.Add(collectedIncentives...)
		totalForefeitedIncentives = totalForefeitedIncentives.Add(forfeitedIncentives...)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtTotalCollectIncentives,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyTokensOut, totalCollectedIncentives.String()),
		),
	})

	return &types.MsgCollectIncentivesResponse{CollectedIncentives: totalCollectedIncentives, ForfeitedIncentives: totalForefeitedIncentives}, nil
}

func (server msgServer) TransferPositions(goCtx context.Context, msg *types.MsgTransferPositions) (*types.MsgTransferPositionsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	newOwner, err := sdk.AccAddressFromBech32(msg.NewOwner)
	if err != nil {
		return nil, err
	}

	err = server.keeper.transferPositions(ctx, msg.PositionIds, sender, newOwner)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtTransferPositions,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeNewOwner, msg.NewOwner),
			sdk.NewAttribute(types.AttributeInputPositionIds, osmocli.ParseUint64SliceToString(msg.PositionIds)),
		),
	})

	return &types.MsgTransferPositionsResponse{}, nil
}
