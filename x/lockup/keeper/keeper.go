package keeper

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/utils"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/lockup/types"
)

// Keeper provides a way to manage module storage
type Keeper struct {
	cdc      codec.Codec
	storeKey sdk.StoreKey

	hooks types.LockupHooks

	ak types.AccountKeeper
	bk types.BankKeeper
	dk types.DistrKeeper
}

// NewKeeper returns an instance of Keeper
func NewKeeper(cdc codec.Codec, storeKey sdk.StoreKey, ak types.AccountKeeper, bk types.BankKeeper, dk types.DistrKeeper) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		ak:       ak,
		bk:       bk,
		dk:       dk,
	}
}

// Logger returns a logger instance
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Set the lockup hooks
func (k *Keeper) SetHooks(lh types.LockupHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set lockup hooks twice")
	}

	k.hooks = lh

	return k
}

// AdminKeeper defines a god priviledge keeper functions to remove tokens from locks and create new locks
// For the governance system of token pools, we want a "ragequit" feature
// So governance changes will take 1 week to go into effect
// During that time, people can choose to "ragequit" which means they would leave the original pool
// and form a new pool with the old parameters but if they still had 2 months of lockup left,
// their liquidity still needs to be 2 month lockup-ed, just in the new pool
// And we need to replace their pool1 LP tokens with pool2 LP tokens with the same lock duration and end time
type AdminKeeper struct {
	Keeper
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

type msgServer struct {
	keeper *Keeper
}

// LockKeeperImpl returns an instance of LockKeeper
func LockKeeperImpl(keeper *Keeper) types.LockKeeper {
	return &msgServer{
		keeper: keeper,
	}
}

var _ types.LockKeeper = msgServer{}

func (server msgServer) LockTokens(goCtx context.Context, msg *types.MsgLockTokens) (*types.MsgLockTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// we only allow locks with one denom for now
	if msg.Coins.Len() != 1 {
		return nil, errors.Wrap(errors.ErrInvalidRequest,
			fmt.Sprintf("Lockups can only have one denom per lockID, got %v", msg.Coins))
	}

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	if len(msg.Coins) == 1 {
		locks := server.keeper.GetAccountLockedDurationNotUnlockingOnly(ctx, owner, msg.Coins[0].Denom, msg.Duration)
		// if existing lock with same duration and denom exists, just add there
		if len(locks) > 0 {
			lock := locks[0]
			if lock.Coins.Len() != 1 {
				return nil, errors.Wrap(errors.ErrInvalidRequest, err.Error())
			}

			if lock.Owner != owner.String() {
				return nil, types.ErrNotLockOwner
			}

			_, err = server.keeper.AddTokensToLockByID(ctx, lock.ID, msg.Coins)
			if err != nil {
				return nil, errors.Wrap(errors.ErrInvalidRequest, err.Error())
			}

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

	lock, err := server.keeper.LockTokens(ctx, owner, msg.Coins, msg.Duration)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInvalidRequest, err.Error())
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
		return nil, errors.Wrap(errors.ErrInvalidRequest, err.Error())
	}

	err = server.keeper.BeginUnlock(ctx, lock.ID, msg.Coins)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInvalidRequest, err.Error())
	}

	if msg.Owner != lock.Owner {
		return nil, errors.Wrap(types.ErrNotLockOwner, fmt.Sprintf("msg sender(%s) and lock owner(%s) does not match", msg.Owner, lock.Owner))
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
		return nil, errors.Wrap(errors.ErrInvalidRequest, err.Error())
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
