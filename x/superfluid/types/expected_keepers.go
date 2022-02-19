package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	epochstypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
)

// LockupKeeper defines the expected interface needed to retrieve locks
type LockupKeeper interface {
	GetLocksPastTimeDenom(ctx sdk.Context, denom string, timestamp time.Time) []lockuptypes.PeriodLock
	GetLocksLongerThanDurationDenom(ctx sdk.Context, denom string, duration time.Duration) []lockuptypes.PeriodLock
	GetPeriodLocksAccumulation(ctx sdk.Context, query lockuptypes.QueryCondition) sdk.Int
	GetAccountPeriodLocks(ctx sdk.Context, addr sdk.AccAddress) []lockuptypes.PeriodLock
	GetPeriodLocks(ctx sdk.Context) ([]lockuptypes.PeriodLock, error)
	GetLockByID(ctx sdk.Context, lockID uint64) (*lockuptypes.PeriodLock, error)
	GetSyntheticLockup(ctx sdk.Context, lockID uint64, suffix string) (*lockuptypes.SyntheticLock, error)
	CreateSyntheticLockup(ctx sdk.Context, lockID uint64, suffix string, unlockDuration time.Duration, isUnlocking bool) error
	DeleteSyntheticLockup(ctx sdk.Context, lockID uint64, suffix string) error
	SlashTokensFromLockByID(ctx sdk.Context, lockID uint64, coins sdk.Coins) (*lockuptypes.PeriodLock, error)

	AddTokensToSyntheticLock(ctx sdk.Context, lock lockuptypes.SyntheticLock, amount sdk.Coins) error
}

// GammKeeper defines the expected interface needed for superfluid module
type GammKeeper interface {
	CalculateSpotPrice(ctx sdk.Context, poolId uint64, tokenInDenom, tokenOutDenom string) (sdk.Dec, error)
	ExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, tokenOutMins sdk.Coins) (err error)
	GetPool(ctx sdk.Context, poolId uint64) (gammtypes.PoolI, error)
	GetPools(ctx sdk.Context) (res []gammtypes.PoolI, err error)
}

type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// StakingKeeper expected staking keeper
type StakingKeeper interface {
	BondDenom(ctx sdk.Context) string
	GetAllValidators(ctx sdk.Context) (validators []stakingtypes.Validator)
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)
	ValidateUnbondAmount(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amt sdk.Int) (shares sdk.Dec, err error)
	Delegate(ctx sdk.Context, delAddr sdk.AccAddress, bondAmt sdk.Int, tokenSrc stakingtypes.BondStatus, validator stakingtypes.Validator, subtractAccount bool) (newShares sdk.Dec, err error)
	Unbond(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, shares sdk.Dec) (amount sdk.Int, err error)
	Undelegate(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec) (time.Time, error)
	GetDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (delegation stakingtypes.Delegation, found bool)
	GetUnbondingDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (ubd types.UnbondingDelegation, found bool)
}

// DistrKeeper expected distribution keeper
type DistrKeeper interface {
	WithdrawDelegationRewards(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error)
}

// IncentivesKeeper expected incentives keeper
type IncentivesKeeper interface {
	CreateGauge(ctx sdk.Context, isPerpetual bool, owner sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpochsPaidOver uint64) (uint64, error)
	AddToGaugeRewards(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, gaugeID uint64) error
}

type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochstypes.EpochInfo
}
