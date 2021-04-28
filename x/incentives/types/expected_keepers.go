package types

import (
	time "time"

	lockuptypes "github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

// LockupKeeper defines the expected interface needed to retrieve locks
type LockupKeeper interface {
	GetLocksPastTimeDenom(ctx sdk.Context, denom string, timestamp time.Time) []lockuptypes.PeriodLock
	GetLocksLongerThanDurationDenom(ctx sdk.Context, denom string, duration time.Duration) []lockuptypes.PeriodLock
	GetAccountLockedPastTimeDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, timestamp time.Time) []lockuptypes.PeriodLock
	GetAccountLockedLongerDurationDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []lockuptypes.PeriodLock
}

type IncentivesKeeper interface {
	IncreaseTotalLocked(ctx sdk.Context, amount sdk.Coins)
	DecreaseTotalLocked(ctx sdk.Context, amount sdk.Coins)
}
