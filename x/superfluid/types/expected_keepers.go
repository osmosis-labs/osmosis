package types

import (
	context "context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochstypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v7/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
)

// LockupKeeper defines the expected interface needed to retrieve locks.
type LockupKeeper interface {
	GetLocksPastTimeDenom(ctx sdk.Context, denom string, timestamp time.Time) []lockuptypes.PeriodLock
	GetLocksLongerThanDurationDenom(ctx sdk.Context, denom string, duration time.Duration) []lockuptypes.PeriodLock
	GetAccountLockedLongerDurationDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []lockuptypes.PeriodLock
	GetAccountLockedLongerDurationDenomNotUnlockingOnly(ctx sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []lockuptypes.PeriodLock
	GetPeriodLocksAccumulation(ctx sdk.Context, query lockuptypes.QueryCondition) sdk.Int
	GetAccountPeriodLocks(ctx sdk.Context, addr sdk.AccAddress) []lockuptypes.PeriodLock
	GetPeriodLocks(ctx sdk.Context) ([]lockuptypes.PeriodLock, error)
	GetLockByID(ctx sdk.Context, lockID uint64) (*lockuptypes.PeriodLock, error)
	// Despite the name, BeginForceUnlock is really BeginUnlock
	// TODO: Fix this in future code update
	BeginForceUnlock(ctx sdk.Context, lockID uint64, coins sdk.Coins) error
	ForceUnlock(ctx sdk.Context, lock lockuptypes.PeriodLock) error

	LockTokens(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, duration time.Duration) (lockuptypes.PeriodLock, error)

	SlashTokensFromLockByID(ctx sdk.Context, lockID uint64, coins sdk.Coins) (*lockuptypes.PeriodLock, error)

	GetSyntheticLockup(ctx sdk.Context, lockID uint64, suffix string) (*lockuptypes.SyntheticLock, error)
	GetAllSyntheticLockupsByAddr(ctx sdk.Context, owner sdk.AccAddress) []lockuptypes.SyntheticLock
	CreateSyntheticLockup(ctx sdk.Context, lockID uint64, suffix string, unlockDuration time.Duration, isUnlocking bool) error
	DeleteSyntheticLockup(ctx sdk.Context, lockID uint64, suffix string) error
	GetAllSyntheticLockupsByLockup(ctx sdk.Context, lockID uint64) []lockuptypes.SyntheticLock
}

type LockupMsgServer interface {
	LockTokens(goCtx context.Context, msg *lockuptypes.MsgLockTokens) (*lockuptypes.MsgLockTokensResponse, error)
}

// GammKeeper defines the expected interface needed for superfluid module.
type GammKeeper interface {
	CalculateSpotPrice(ctx sdk.Context, poolId uint64, tokenInDenom, tokenOutDenom string) (sdk.Dec, error)
	GetPoolAndPoke(ctx sdk.Context, poolId uint64) (gammtypes.PoolI, error)
	GetPoolsAndPoke(ctx sdk.Context) (res []gammtypes.PoolI, err error)
	ExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, tokenOutMins sdk.Coins) (exitCoins sdk.Coins, err error)
}

type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error
	GetSupplyOffset(ctx sdk.Context, denom string) sdk.Int
	AddSupplyOffset(ctx sdk.Context, denom string, offsetAmount sdk.Int)
	GetSupplyWithOffset(ctx sdk.Context, denom string) sdk.Coin
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// StakingKeeper expected staking keeper.
type StakingKeeper interface {
	BondDenom(ctx sdk.Context) string
	GetAllValidators(ctx sdk.Context) (validators []stakingtypes.Validator)
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)
	ValidateUnbondAmount(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amt sdk.Int) (shares sdk.Dec, err error)
	Delegate(ctx sdk.Context, delAddr sdk.AccAddress, bondAmt sdk.Int, tokenSrc stakingtypes.BondStatus, validator stakingtypes.Validator, subtractAccount bool) (newShares sdk.Dec, err error)
	Unbond(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, shares sdk.Dec) (amount sdk.Int, err error)
	Undelegate(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec) (time.Time, error)
	InstantUndelegate(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec) (sdk.Coins, error)
	GetDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (delegation stakingtypes.Delegation, found bool)
	GetUnbondingDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (ubd stakingtypes.UnbondingDelegation, found bool)
	UnbondingTime(ctx sdk.Context) time.Duration
	GetParams(ctx sdk.Context) stakingtypes.Params

	IterateBondedValidatorsByPower(ctx sdk.Context, fn func(int64, stakingtypes.ValidatorI) bool)
	TotalBondedTokens(ctx sdk.Context) sdk.Int
	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress, fn func(int64, stakingtypes.DelegationI) bool)
}

// DistrKeeper expected distribution keeper.
type DistrKeeper interface {
	WithdrawDelegationRewards(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error)
}

// IncentivesKeeper expected incentives keeper.
type IncentivesKeeper interface {
	CreateGauge(ctx sdk.Context, isPerpetual bool, owner sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpochsPaidOver uint64) (uint64, error)
	AddToGaugeRewards(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, gaugeID uint64) error

	GetActiveGauges(ctx sdk.Context) []incentivestypes.Gauge
	Distribute(ctx sdk.Context, gauges []incentivestypes.Gauge) (sdk.Coins, error)

	GetParams(ctx sdk.Context) incentivestypes.Params
}

type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochstypes.EpochInfo
	NumBlocksSinceEpochStart(ctx sdk.Context, identifier string) (int64, error)
}
