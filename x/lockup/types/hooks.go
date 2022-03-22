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
