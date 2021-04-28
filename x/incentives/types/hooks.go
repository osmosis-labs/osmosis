package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IncentiveHooks interface {
	AfterCreatePot(ctx sdk.Context, potId uint64)
	AfterAddToPot(ctx sdk.Context, potId uint64)
	AfterStartDistribution(ctx sdk.Context, potId uint64)
	AfterFinishDistribution(ctx sdk.Context, potId uint64)
	AfterDistribute(ctx sdk.Context, potId uint64)
}

var _ IncentiveHooks = MultiIncentiveHooks{}

// combine multiple incentive hooks, all hook functions are run in array sequence
type MultiIncentiveHooks []IncentiveHooks

func NewMultiIncentiveHooks(hooks ...IncentiveHooks) MultiIncentiveHooks {
	return hooks
}

func (h MultiIncentiveHooks) AfterCreatePot(ctx sdk.Context, potId uint64) {
	for i := range h {
		h[i].AfterCreatePot(ctx, potId)
	}
}

func (h MultiIncentiveHooks) AfterAddToPot(ctx sdk.Context, potId uint64) {
	for i := range h {
		h[i].AfterAddToPot(ctx, potId)
	}
}

func (h MultiIncentiveHooks) AfterStartDistribution(ctx sdk.Context, potId uint64) {
	for i := range h {
		h[i].AfterStartDistribution(ctx, potId)
	}
}

func (h MultiIncentiveHooks) AfterFinishDistribution(ctx sdk.Context, potId uint64) {
	for i := range h {
		h[i].AfterFinishDistribution(ctx, potId)
	}
}

func (h MultiIncentiveHooks) AfterDistribute(ctx sdk.Context, potId uint64) {
	for i := range h {
		h[i].AfterDistribute(ctx, potId)
	}
}

type LockupHook struct {
	k IncentivesKeeper
}

func NewLockupHook(k IncentivesKeeper) LockupHook {
	return LockupHook{k}
}

func (h LockupHook) OnTokenLocked(ctx sdk.Context, _ sdk.AccAddress, _ uint64, amount sdk.Coins, _ time.Duration, _ time.Time) {
	h.k.IncreaseTotalLocked(ctx, amount)
}

func (h LockupHook) OnTokenUnlocked(ctx sdk.Context, _ sdk.AccAddress, _ uint64, amount sdk.Coins, _ time.Duration, _ time.Time) {
	h.k.DecreaseTotalLocked(ctx, amount)
}
