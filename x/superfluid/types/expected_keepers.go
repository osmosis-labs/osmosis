package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	gammtypes "github.com/osmosis-labs/osmosis/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
)

// LockupKeeper defines the expected interface needed to retrieve locks
type LockupKeeper interface {
	GetLocksPastTimeDenom(ctx sdk.Context, denom string, timestamp time.Time) []lockuptypes.PeriodLock
	GetLocksLongerThanDurationDenom(ctx sdk.Context, denom string, duration time.Duration) []lockuptypes.PeriodLock
	GetPeriodLocksAccumulation(ctx sdk.Context, query lockuptypes.QueryCondition) sdk.Int
	GetAccountPeriodLocks(ctx sdk.Context, addr sdk.AccAddress) []lockuptypes.PeriodLock
	GetLockByID(ctx sdk.Context, lockID uint64) (*lockuptypes.PeriodLock, error)
	CreateSyntheticLockup(ctx sdk.Context, lockID uint64, suffix string, isUnlocking bool) error
}

// GammKeeper defines the expected interface needed for superfluid module
type GammKeeper interface {
	// TODO: should we just calculate Osmo equivalent value of LP token using ExitPool function
	//  or CalculateSpotPrice function?
	CalculateSpotPrice(ctx sdk.Context, poolId uint64, tokenInDenom, tokenOutDenom string) (sdk.Dec, error)
	ExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, tokenOutMins sdk.Coins) (err error)
	GetPool(ctx sdk.Context, poolId uint64) (gammtypes.PoolI, error)
}

type BankKeeper interface {
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// StakingKeeper expected staking keeper (noalias)
type StakingKeeper interface {
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator types.Validator, found bool)
	Delegate(ctx sdk.Context, delAddr sdk.AccAddress, bondAmt sdk.Int, tokenSrc types.BondStatus, validator types.Validator, subtractAccount bool) (newShares sdk.Dec, err error)
}
