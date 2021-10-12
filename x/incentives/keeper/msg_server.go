package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/osmosis-labs/osmosis/x/gamm/utils"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an instance of MsgServer
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

var _ types.MsgServer = msgServer{}

func (server msgServer) CreateGauge(goCtx context.Context, msg *types.MsgCreateGauge) (*types.MsgCreateGaugeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	gaugeID, err := server.keeper.CreateGauge(ctx, msg.IsPerpetual, owner, msg.Coins, msg.DistributeTo, msg.StartTime, msg.NumEpochsPaidOver)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCreateGauge,
			sdk.NewAttribute(types.AttributeGaugeID, utils.Uint64ToString(gaugeID)),
		),
	})

	return &types.MsgCreateGaugeResponse{}, nil
}

func (server msgServer) AddToGauge(goCtx context.Context, msg *types.MsgAddToGauge) (*types.MsgAddToGaugeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}
	err = server.keeper.AddToGaugeRewards(ctx, owner, msg.Rewards, msg.GaugeId)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtAddToGauge,
			sdk.NewAttribute(types.AttributeGaugeID, utils.Uint64ToString(msg.GaugeId)),
		),
	})

	return &types.MsgAddToGaugeResponse{}, nil
}

func (server msgServer) ClaimLockReward(goCtx context.Context, msg *types.MsgClaimLockReward) (*types.MsgClaimLockRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lockID := msg.ID
	lock, err := server.keeper.lk.GetLockByID(ctx, lockID)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if msg.Owner != lock.Owner {
		ctx.Logger().Debug(fmt.Sprintf("msg sender(%s) and lock owner(%s) does not match", msg.Owner, lock.Owner))
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("msg sender(%s) and lock owner(%s) does not match", msg.Owner, lock.Owner))
	}

	lockReward, err := server.keeper.GetPeriodLockReward(ctx, lockID)
	if err != nil {
		ctx.Logger().Debug(fmt.Sprintf("F1::: ClaimLockReward::: error:%s", err.Error()))
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	epochInfo := server.keeper.GetEpochInfo(ctx)
	lockableDurations := server.keeper.GetLockableDurations(ctx)
	err = server.keeper.UpdateHistoricalRewardFromCurrentReward(ctx, lock.Coins, lock.Duration, epochInfo, lockableDurations)
	if err != nil {
		ctx.Logger().Debug(fmt.Sprintf("F1::: ClaimLockReward::: error:%s", err.Error()))
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	lockReward, err = server.keeper.GetRewardForLock(ctx, *lock, lockReward, epochInfo, lockableDurations)
	if err != nil {
		ctx.Logger().Debug(fmt.Sprintf("F1::: ClaimLockReward::: error:%s", err.Error()))
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	sentRewards, err := server.keeper.ClaimRewardForLock(ctx, *lock, lockReward, lockableDurations)
	if err != nil {
		ctx.Logger().Debug(fmt.Sprintf("F1::: ClaimLockReward::: error:%s", err.Error()))
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.Logger().Debug(fmt.Sprintf("F1::: ClaimLockReward::: reward:%s", sentRewards.String()))

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtClaimLockReward,
			sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(lock.ID)),
			sdk.NewAttribute(types.AttributePeriodLockOwner, lock.Owner),
			sdk.NewAttribute(types.AttributeRewardCoins, sentRewards.String()),
		),
	})

	return &types.MsgClaimLockRewardResponse{}, nil
}

func (server msgServer) ClaimLockRewardAll(goCtx context.Context, msg *types.MsgClaimLockRewardAll) (*types.MsgClaimLockRewardResponseAll, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	lockRewards := []sdk.Coins{}
	sentRewards := sdk.Coins{}
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	locks := server.keeper.lk.GetAccountPeriodLocks(ctx, sdk.AccAddress(owner))
	ctx.Logger().Debug(fmt.Sprintf("F1::: ClaimLockRewardAll::: locks:%d", len(locks)))

	for _, lock := range locks {
		ctx.Logger().Debug(fmt.Sprintf("F1::: ClaimLockRewardAll::: lock:%d", lock.ID))
		if msg.Owner != lock.Owner {
			ctx.Logger().Debug(fmt.Sprintf("msg sender(%s) and lock owner(%s) does not match", msg.Owner, lock.Owner))
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("msg sender(%s) and lock owner(%s) does not match", msg.Owner, lock.Owner))
		}
		lockID := lock.ID

		lockReward, err := server.keeper.GetPeriodLockReward(ctx, lockID)
		if err != nil {
			ctx.Logger().Debug(fmt.Sprintf("F1::: ClaimLockRewardAll::: error:%s", err.Error()))
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
		}
		epochInfo := server.keeper.GetEpochInfo(ctx)
		lockableDurations := server.keeper.GetLockableDurations(ctx)
		err = server.keeper.UpdateHistoricalRewardFromCurrentReward(ctx, lock.Coins, lock.Duration, epochInfo, lockableDurations)
		if err != nil {
			ctx.Logger().Debug(fmt.Sprintf("F1::: ClaimLockRewardAll::: error:%s", err.Error()))
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
		}
		lockReward, err = server.keeper.GetRewardForLock(ctx, lock, lockReward, epochInfo, lockableDurations)
		if err != nil {
			ctx.Logger().Debug(fmt.Sprintf("F1::: ClaimLockRewardAll::: error:%s", err.Error()))
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
		}
		rewards, err := server.keeper.ClaimRewardForLock(ctx, lock, lockReward, lockableDurations)
		if err != nil {
			ctx.Logger().Debug(fmt.Sprintf("F1::: ClaimLockRewardAll::: error:%s", err.Error()))
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
		}

		// TODO: refactor
		lockRewards = append(lockRewards, rewards)
		sentRewards = sentRewards.Add(rewards...)

		ctx.Logger().Debug(fmt.Sprintf("F1::: ClaimLockRewardAll::: reward:%s", rewards.String()))
	}
	ctx.Logger().Debug(fmt.Sprintf("F1::: ClaimLockRewardAll::: total reward:%s", sentRewards.String()))

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtClaimLockRewardAll,
			sdk.NewAttribute(types.AttributePeriodLockOwner, msg.Owner),
			sdk.NewAttribute(types.AttributeRewardCoins, sentRewards.String()),
		),
	})
	for i, lock := range locks {
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.TypeEvtClaimLockReward,
				sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(lock.ID)),
				sdk.NewAttribute(types.AttributePeriodLockOwner, lock.Owner),
				sdk.NewAttribute(types.AttributeRewardCoins, lockRewards[i].String()),
			),
		})
	}

	return &types.MsgClaimLockRewardResponseAll{}, nil
}
