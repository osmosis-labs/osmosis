package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/osmoutils"
)

type LockupHooks interface {
	AfterAddTokensToLock(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins) error
	OnTokenLocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) error
	OnStartUnlock(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) error
	OnTokenUnlocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) error
	OnTokenSlashed(ctx sdk.Context, lockID uint64, amount sdk.Coins) error
	OnLockupExtend(ctx sdk.Context, lockID uint64, prevDuration time.Duration, newDuration time.Duration) error
}

var _ LockupHooks = MultiLockupHooks{}

// combine multiple gamm hooks, all hook functions are run in array sequence.
type MultiLockupHooks []LockupHooks

func NewMultiLockupHooks(hooks ...LockupHooks) MultiLockupHooks {
	return hooks
}

func (h MultiLockupHooks) AfterAddTokensToLock(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins) error {
	for i := range h {
		wrappedHookFn := func(ctx sdk.Context) error {
			return h[i].AfterAddTokensToLock(ctx, address, lockID, amount)
		}
		handleHooksError(ctx, osmoutils.ApplyFuncIfNoError(ctx, wrappedHookFn))
	}
	return nil
}

func (h MultiLockupHooks) OnTokenLocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) error {
	for i := range h {
		wrappedHookFn := func(ctx sdk.Context) error {
			return h[i].OnTokenLocked(ctx, address, lockID, amount, lockDuration, unlockTime)
		}
		handleHooksError(ctx, osmoutils.ApplyFuncIfNoError(ctx, wrappedHookFn))
	}
	return nil
}

func (h MultiLockupHooks) OnStartUnlock(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) error {
	for i := range h {
		wrappedHookFn := func(ctx sdk.Context) error {
			return h[i].OnStartUnlock(ctx, address, lockID, amount, lockDuration, unlockTime)
		}
		handleHooksError(ctx, osmoutils.ApplyFuncIfNoError(ctx, wrappedHookFn))
	}
	return nil
}

func (h MultiLockupHooks) OnTokenUnlocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) error {
	for i := range h {
		wrappedHookFn := func(ctx sdk.Context) error {
			return h[i].OnTokenUnlocked(ctx, address, lockID, amount, lockDuration, unlockTime)
		}
		handleHooksError(ctx, osmoutils.ApplyFuncIfNoError(ctx, wrappedHookFn))
	}
	return nil
}

func (h MultiLockupHooks) OnTokenSlashed(ctx sdk.Context, lockID uint64, amount sdk.Coins) error {
	for i := range h {
		wrappedHookFn := func(ctx sdk.Context) error {
			return h[i].OnTokenSlashed(ctx, lockID, amount)
		}
		handleHooksError(ctx, osmoutils.ApplyFuncIfNoError(ctx, wrappedHookFn))
	}
	return nil
}

func (h MultiLockupHooks) OnLockupExtend(ctx sdk.Context, lockID uint64, prevDuration time.Duration, newDuration time.Duration) error {
	for i := range h {
		wrappedHookFn := func(ctx sdk.Context) error {
			return h[i].OnLockupExtend(ctx, lockID, prevDuration, newDuration)
		}
		handleHooksError(ctx, osmoutils.ApplyFuncIfNoError(ctx, wrappedHookFn))
	}
	return nil
}

// handleHooksError logs the error using the ctx logger
func handleHooksError(ctx sdk.Context, err error) {
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("error in lockup hook %v", err))
	}
}
