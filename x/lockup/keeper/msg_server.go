package keeper

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/v9/x/gamm/utils"
	"github.com/osmosis-labs/osmosis/v9/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type msgServer struct {
	keeper *Keeper
}

// NewMsgServerImpl returns an instance of MsgServer.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

var _ types.MsgServer = msgServer{}

func (server msgServer) LockTokens(goCtx context.Context, msg *types.MsgLockTokens) (*types.MsgLockTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	// if there is an existing lock
	if len(msg.Coins) == 1 {
		locks, err := server.keeper.AddToExistingLock(ctx, owner, msg.Coins[0], msg.Duration)
		if err != nil {
			return nil, err
		}
		if len(locks) > 0 {
			ctx.EventManager().EmitEvents(sdk.Events{
				sdk.NewEvent(
					types.TypeEvtAddTokensToLock,
					sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(locks[0].ID)),
					sdk.NewAttribute(types.AttributePeriodLockOwner, msg.Owner),
					sdk.NewAttribute(types.AttributePeriodLockAmount, msg.Coins.String()),
				),
			})
			return &types.MsgLockTokensResponse{ID: locks[0].ID}, nil
		}
	}

	lock, err := server.keeper.CreateLock(ctx, owner, msg.Coins, msg.Duration)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtLockTokens,
			sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(lock.ID)),
			sdk.NewAttribute(types.AttributePeriodLockOwner, lock.Owner),
			sdk.NewAttribute(types.AttributePeriodLockAmount, lock.Coins.String()),
			sdk.NewAttribute(types.AttributePeriodLockDuration, lock.Duration.String()),
			sdk.NewAttribute(types.AttributePeriodLockUnlockTime, lock.EndTime.String()),
		),
	})

	return &types.MsgLockTokensResponse{ID: lock.ID}, nil
}

func (server msgServer) BeginUnlocking(goCtx context.Context, msg *types.MsgBeginUnlocking) (*types.MsgBeginUnlockingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lock, err := server.keeper.GetLockByID(ctx, msg.ID)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if msg.Owner != lock.Owner {
		return nil, sdkerrors.Wrap(types.ErrNotLockOwner, fmt.Sprintf("msg sender (%s) and lock owner (%s) does not match", msg.Owner, lock.Owner))
	}

	err = server.keeper.BeginUnlock(ctx, lock.ID, msg.Coins)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	lock, err = server.keeper.GetLockByID(ctx, msg.ID)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		createBeginUnlockEvent(lock),
	})

	return &types.MsgBeginUnlockingResponse{}, nil
}

func (server msgServer) BeginUnlockingAll(goCtx context.Context, msg *types.MsgBeginUnlockingAll) (*types.MsgBeginUnlockingAllResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	unlocks, err := server.keeper.BeginUnlockAllNotUnlockings(ctx, owner)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	// Create the events for this message
	unlockedCoins := server.keeper.getCoinsFromLocks(unlocks)
	events := sdk.Events{
		sdk.NewEvent(
			types.TypeEvtBeginUnlockAll,
			sdk.NewAttribute(types.AttributePeriodLockOwner, msg.Owner),
			sdk.NewAttribute(types.AttributeUnlockedCoins, unlockedCoins.String()),
		),
	}
	for _, lock := range unlocks {
		events = events.AppendEvent(createBeginUnlockEvent(&lock))
	}
	ctx.EventManager().EmitEvents(events)

	return &types.MsgBeginUnlockingAllResponse{}, nil
}

func createBeginUnlockEvent(lock *types.PeriodLock) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtBeginUnlock,
		sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(lock.ID)),
		sdk.NewAttribute(types.AttributePeriodLockOwner, lock.Owner),
		sdk.NewAttribute(types.AttributePeriodLockDuration, lock.Duration.String()),
		sdk.NewAttribute(types.AttributePeriodLockUnlockTime, lock.EndTime.String()),
	)
}

func (server msgServer) ExtendLockup(goCtx context.Context, msg *types.MsgExtendLockup) (*types.MsgExtendLockupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lock, err := server.keeper.GetLockByID(ctx, msg.ID)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if msg.Owner != lock.Owner {
		return nil, sdkerrors.Wrapf(types.ErrNotLockOwner, fmt.Sprintf("msg sender (%s) and lock owner (%s) does not match", msg.Owner, lock.Owner))
	}

	err = server.keeper.ExtendLockup(ctx, *lock, msg.Duration)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtLockTokens,
			sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(lock.ID)),
			sdk.NewAttribute(types.AttributePeriodLockOwner, lock.Owner),
			sdk.NewAttribute(types.AttributePeriodLockDuration, lock.Duration.String()),
		),
	})

	return &types.MsgExtendLockupResponse{}, nil
}
