package events

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v10/x/gamm/utils"
	"github.com/osmosis-labs/osmosis/v10/x/lockup/types"
)

// EmitLockToken returns a new event when user lock tokens.
func EmitLockToken(ctx sdk.Context, lock *types.PeriodLock) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		lockTokenEvent(lock),
	})
}

func lockTokenEvent(lock *types.PeriodLock) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtLockTokens,
		sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(lock.ID)),
		sdk.NewAttribute(types.AttributePeriodLockOwner, lock.Owner),
		sdk.NewAttribute(types.AttributePeriodLockAmount, lock.Coins.String()),
		sdk.NewAttribute(types.AttributePeriodLockDuration, lock.Duration.String()),
		sdk.NewAttribute(types.AttributePeriodLockUnlockTime, lock.EndTime.String()),
	)
}

// EmitLockToken returns a new event when user lock tokens.
func EmitExtendLockToken(ctx sdk.Context, lock *types.PeriodLock) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		extendLockTokenEvent(lock),
	})
}

func extendLockTokenEvent(lock *types.PeriodLock) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtLockTokens,
		sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(lock.ID)),
		sdk.NewAttribute(types.AttributePeriodLockOwner, lock.Owner),
		sdk.NewAttribute(types.AttributePeriodLockDuration, lock.Duration.String()),
	)
}

// EmitAddTokenToLock returns a new event when tokens are added to an existing lock.
func EmitAddTokenToLock(ctx sdk.Context, lockId uint64, owner, coins string) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		addTokenToLockEvent(lockId, owner, coins),
	})
}

func addTokenToLockEvent(lockId uint64, owner, coins string) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtAddTokensToLock,
		sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(lockId)),
		sdk.NewAttribute(types.AttributePeriodLockOwner, owner),
		sdk.NewAttribute(types.AttributePeriodLockAmount, coins),
	)
}

// EmitBeginUnlockAll returns a new event when user beings unlocking for all the lock that the account has.
func EmitBeginUnlockAll(ctx sdk.Context, unlockedCoins, owner string) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		beginUnlockAll(unlockedCoins, owner),
	})
}

func beginUnlockAll(unlockedCoins, owner string) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtBeginUnlockAll,
		sdk.NewAttribute(types.AttributePeriodLockOwner, owner),
		sdk.NewAttribute(types.AttributeUnlockedCoins, unlockedCoins),
	)
}

// EmitBeginUnlock returns a new event when user beings unlocking speficic lock.
func EmitBeginUnlock(ctx sdk.Context, lock *types.PeriodLock) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		BeginUnlockEvent(lock),
	})
}

// Question: Any way we can make this private method?
func BeginUnlockEvent(lock *types.PeriodLock) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtBeginUnlock,
		sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(lock.ID)),
		sdk.NewAttribute(types.AttributePeriodLockOwner, lock.Owner),
		sdk.NewAttribute(types.AttributePeriodLockDuration, lock.Duration.String()),
		sdk.NewAttribute(types.AttributePeriodLockUnlockTime, lock.EndTime.String()),
	)
}
