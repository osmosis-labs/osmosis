package keeper

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	errorsmod "cosmossdk.io/errors"
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

// LockTokens locks tokens in either two ways.
// 1. Add to an existing lock if a lock with the same owner and same duration exists.
// 2. Create a new lock if not.
// A sanity check to ensure given tokens is a single token is done in ValidateBaic.
// That is, a lock with multiple tokens cannot be created.
func (server msgServer) LockTokens(goCtx context.Context, msg *types.MsgLockTokens) (*types.MsgLockTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	// check if there's an existing lock from the same owner with the same duration.
	// If so, simply add tokens to the existing lock.
	lockExists := server.keeper.HasLock(ctx, owner, msg.Coins[0].Denom, msg.Duration)
	if lockExists {
		lockID, err := server.keeper.AddToExistingLock(ctx, owner, msg.Coins[0], msg.Duration)
		if err != nil {
			return nil, err
		}

		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.TypeEvtAddTokensToLock,
				sdk.NewAttribute(types.AttributePeriodLockID, osmoutils.Uint64ToString(lockID)),
				sdk.NewAttribute(types.AttributePeriodLockOwner, msg.Owner),
				sdk.NewAttribute(types.AttributePeriodLockAmount, msg.Coins.String()),
			),
		})
		return &types.MsgLockTokensResponse{ID: lockID}, nil
	}

	// if the owner + duration combination is new, create a new lock.
	lock, err := server.keeper.CreateLock(ctx, owner, msg.Coins, msg.Duration)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtLockTokens,
			sdk.NewAttribute(types.AttributePeriodLockID, osmoutils.Uint64ToString(lock.ID)),
			sdk.NewAttribute(types.AttributePeriodLockOwner, lock.Owner),
			sdk.NewAttribute(types.AttributePeriodLockAmount, lock.Coins.String()),
			sdk.NewAttribute(types.AttributePeriodLockDuration, lock.Duration.String()),
			sdk.NewAttribute(types.AttributePeriodLockUnlockTime, lock.EndTime.String()),
		),
	})

	return &types.MsgLockTokensResponse{ID: lock.ID}, nil
}

// BeginUnlocking begins unlocking of the specified lock.
// The lock would enter the unlocking queue, with the endtime of the lock set as block time + duration.
func (server msgServer) BeginUnlocking(goCtx context.Context, msg *types.MsgBeginUnlocking) (*types.MsgBeginUnlockingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lock, err := server.keeper.GetLockByID(ctx, msg.ID)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if msg.Owner != lock.Owner {
		return nil, errorsmod.Wrap(types.ErrNotLockOwner, fmt.Sprintf("msg sender (%s) and lock owner (%s) does not match", msg.Owner, lock.Owner))
	}

	unlockingLock, err := server.keeper.BeginUnlock(ctx, lock.ID, msg.Coins)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	// N.B. begin unlock event is emitted downstream in the keeper method.

	return &types.MsgBeginUnlockingResponse{Success: true, UnlockingLockID: unlockingLock}, nil
}

// BeginUnlockingAll begins unlocking for all the locks that the account has by iterating all the not-unlocking locks the account holds.
func (server msgServer) BeginUnlockingAll(goCtx context.Context, msg *types.MsgBeginUnlockingAll) (*types.MsgBeginUnlockingAllResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	unlocks, err := server.keeper.BeginUnlockAllNotUnlockings(ctx, owner)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
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
		lock := lock
		events = events.AppendEvent(createBeginUnlockEvent(&lock))
	}
	ctx.EventManager().EmitEvents(events)

	return &types.MsgBeginUnlockingAllResponse{}, nil
}

func createBeginUnlockEvent(lock *types.PeriodLock) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtBeginUnlock,
		sdk.NewAttribute(types.AttributePeriodLockID, osmoutils.Uint64ToString(lock.ID)),
		sdk.NewAttribute(types.AttributePeriodLockOwner, lock.Owner),
		sdk.NewAttribute(types.AttributePeriodLockDenom, lock.Coins[0].Denom),
		sdk.NewAttribute(types.AttributePeriodLockAmount, lock.Coins[0].Amount.String()),
		sdk.NewAttribute(types.AttributePeriodLockDuration, lock.Duration.String()),
		sdk.NewAttribute(types.AttributePeriodLockUnlockTime, lock.EndTime.String()),
	)
}

// ExtendLockup extends the duration of the existing lock.
// ExtendLockup would fail if the original lock's duration is longer than the new duration,
// OR if the lock is currently unlocking OR if the original lock has a synthetic lock.
func (server msgServer) ExtendLockup(goCtx context.Context, msg *types.MsgExtendLockup) (*types.MsgExtendLockupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	err = server.keeper.ExtendLockup(ctx, msg.ID, owner, msg.Duration)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
	}

	lock, err := server.keeper.GetLockByID(ctx, msg.ID)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtLockTokens,
			sdk.NewAttribute(types.AttributePeriodLockID, osmoutils.Uint64ToString(lock.ID)),
			sdk.NewAttribute(types.AttributePeriodLockOwner, lock.Owner),
			sdk.NewAttribute(types.AttributePeriodLockDuration, lock.Duration.String()),
		),
	})

	return &types.MsgExtendLockupResponse{}, nil
}

// ForceUnlock ignores unlock duration and immediately unlocks the lock.
// This message is only allowed for governance-passed accounts that are kept as parameter in the lockup module.
// Locks that has been superfluid delegated is not supported.
func (server msgServer) ForceUnlock(goCtx context.Context, msg *types.MsgForceUnlock) (*types.MsgForceUnlockResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lock, err := server.keeper.GetLockByID(ctx, msg.ID)
	if err != nil {
		return &types.MsgForceUnlockResponse{Success: false}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	// check if message sender matches lock owner
	if lock.Owner != msg.Owner {
		return &types.MsgForceUnlockResponse{Success: false}, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "Sender (%s) does not match lock owner (%s)", msg.Owner, lock.Owner)
	}

	// check for chain parameter that the address is allowed to force unlock
	forceUnlockAllowedAddresses := server.keeper.GetParams(ctx).ForceUnlockAllowedAddresses
	found := false
	for _, addr := range forceUnlockAllowedAddresses {
		// defense in depth, double checking the message owner and lock owner are both the same and is one of the allowed force unlock addresses
		if addr == lock.Owner && addr == msg.Owner {
			found = true
			break
		}
	}
	if !found {
		return &types.MsgForceUnlockResponse{Success: false}, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "Sender (%s) not allowed to force unlock", lock.Owner)
	}

	// check that given lock is not superfluid staked
	// TODO: Next state break do found, instead !synthlock.IsNil()
	synthLock, _, err := server.keeper.GetSyntheticLockupByUnderlyingLockId(ctx, lock.ID)
	if err != nil {
		return &types.MsgForceUnlockResponse{Success: false}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if !synthLock.IsNil() {
		return &types.MsgForceUnlockResponse{Success: false}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "superfluid delegation exists for lock")
	}

	// force unlock given lock
	// This also supports the case of force unlocking lock as a whole when msg.Coins
	// provided is empty.
	err = server.keeper.PartialForceUnlock(ctx, *lock, msg.Coins)
	if err != nil {
		return &types.MsgForceUnlockResponse{Success: false}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	return &types.MsgForceUnlockResponse{Success: true}, nil
}

func (server msgServer) SetRewardReceiverAddress(goCtx context.Context, msg *types.MsgSetRewardReceiverAddress) (*types.MsgSetRewardReceiverAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	newRewardRecepient, err := sdk.AccAddressFromBech32(msg.RewardReceiver)
	if err != nil {
		return nil, err
	}

	err = server.keeper.SetLockRewardReceiverAddress(ctx, msg.LockID, owner, newRewardRecepient.String())
	if err != nil {
		return &types.MsgSetRewardReceiverAddressResponse{Success: false}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	return &types.MsgSetRewardReceiverAddressResponse{Success: true}, nil
}
