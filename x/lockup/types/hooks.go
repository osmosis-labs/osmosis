package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type LockupHooks interface {
	AfterAddTokensToLock(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins)
	OnTokenLocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time)
	OnStartUnlock(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time)
	OnTokenUnlocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time)
	OnTokenSlashed(ctx sdk.Context, lockID uint64, amount sdk.Coins)
	OnLockupExtend(ctx sdk.Context, lockID uint64, prevDuration time.Duration, newDuration time.Duration)
}

var _ LockupHooks = MultiLockupHooks{}

// combine multiple gamm hooks, all hook functions are run in array sequence.
type MultiLockupHooks []LockupHooks

func NewMultiLockupHooks(hooks ...LockupHooks) MultiLockupHooks {
	return hooks
}

func (h MultiLockupHooks) AfterAddTokensToLock(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins) {
	for i := range h {
		h[i].AfterAddTokensToLock(ctx, address, lockID, amount)
	}
}

func (h MultiLockupHooks) OnTokenLocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
	for i := range h {
		h[i].OnTokenLocked(ctx, address, lockID, amount, lockDuration, unlockTime)
	}
}

func (h MultiLockupHooks) OnStartUnlock(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
	for i := range h {
		h[i].OnStartUnlock(ctx, address, lockID, amount, lockDuration, unlockTime)
	}
}

func (h MultiLockupHooks) OnTokenUnlocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
	for i := range h {
		h[i].OnTokenUnlocked(ctx, address, lockID, amount, lockDuration, unlockTime)
	}
}

func (h MultiLockupHooks) OnTokenSlashed(ctx sdk.Context, lockID uint64, amount sdk.Coins) {
	for i := range h {
		h[i].OnTokenSlashed(ctx, lockID, amount)
	}
}

func (h MultiLockupHooks) OnLockupExtend(ctx sdk.Context, lockID uint64, prevDuration, newDuration time.Duration) {
	for i := range h {
		h[i].OnLockupExtend(ctx, lockID, prevDuration, newDuration)
	}
}
