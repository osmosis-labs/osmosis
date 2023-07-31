package types

import (
	context "context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/model"
	cltypes "github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v17/x/gamm/types"
	gammmigration "github.com/osmosis-labs/osmosis/v17/x/gamm/types/migration"
	incentivestypes "github.com/osmosis-labs/osmosis/v17/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v17/x/lockup/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

// LockupKeeper defines the expected interface needed to retrieve locks.
type LockupKeeper interface {
	GetLocksLongerThanDurationDenom(ctx sdk.Context, denom string, duration time.Duration) []lockuptypes.PeriodLock
	GetAccountLockedLongerDurationDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []lockuptypes.PeriodLock
	GetAccountLockedLongerDurationDenomNotUnlockingOnly(ctx sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []lockuptypes.PeriodLock
	GetPeriodLocksAccumulation(ctx sdk.Context, query lockuptypes.QueryCondition) sdk.Int
	GetAccountPeriodLocks(ctx sdk.Context, addr sdk.AccAddress) []lockuptypes.PeriodLock
	GetPeriodLocks(ctx sdk.Context) ([]lockuptypes.PeriodLock, error)
	GetLockByID(ctx sdk.Context, lockID uint64) (*lockuptypes.PeriodLock, error)
	// Despite the name, BeginForceUnlock is really BeginUnlock
	// TODO: Fix this in future code update
	BeginForceUnlock(ctx sdk.Context, lockID uint64, coins sdk.Coins) (uint64, error)
	ForceUnlock(ctx sdk.Context, lock lockuptypes.PeriodLock) error
	PartialForceUnlock(ctx sdk.Context, lock lockuptypes.PeriodLock, coins sdk.Coins) error
	SplitLock(ctx sdk.Context, lock lockuptypes.PeriodLock, coins sdk.Coins, forceUnlock bool) (lockuptypes.PeriodLock, error)

	CreateLock(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, duration time.Duration) (lockuptypes.PeriodLock, error)

	SlashTokensFromLockByID(ctx sdk.Context, lockID uint64, coins sdk.Coins) (*lockuptypes.PeriodLock, error)
	SlashTokensFromLockByIDSendUnderlyingAndBurn(ctx sdk.Context, lockID uint64, liquiditySharesInLock, underlyingPositionAssets sdk.Coins, poolAddress sdk.AccAddress) (*lockuptypes.PeriodLock, error)

	GetSyntheticLockup(ctx sdk.Context, lockID uint64, suffix string) (*lockuptypes.SyntheticLock, error)
	GetAllSyntheticLockupsByAddr(ctx sdk.Context, owner sdk.AccAddress) []lockuptypes.SyntheticLock
	GetAllSyntheticLockups(ctx sdk.Context) []lockuptypes.SyntheticLock
	CreateSyntheticLockup(ctx sdk.Context, lockID uint64, suffix string, unlockDuration time.Duration, isUnlocking bool) error
	DeleteSyntheticLockup(ctx sdk.Context, lockID uint64, suffix string) error
	GetSyntheticLockupByUnderlyingLockId(ctx sdk.Context, lockID uint64) (lockuptypes.SyntheticLock, bool, error)
}

type LockupMsgServer interface {
	LockTokens(goCtx context.Context, msg *lockuptypes.MsgLockTokens) (*lockuptypes.MsgLockTokensResponse, error)
}

// GammKeeper defines the expected interface needed for superfluid module.
type GammKeeper interface {
	GetPoolAndPoke(ctx sdk.Context, poolId uint64) (gammtypes.CFMMPoolI, error)
	GetPoolsAndPoke(ctx sdk.Context) (res []gammtypes.CFMMPoolI, err error)
	ExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, tokenOutMins sdk.Coins) (exitCoins sdk.Coins, err error)
	GetAllMigrationInfo(ctx sdk.Context) (gammmigration.MigrationRecords, error)
	GetLinkedConcentratedPoolID(ctx sdk.Context, poolIdLeaving uint64) (poolIdEntering uint64, err error)
	MigrateUnlockedPositionFromBalancerToConcentrated(ctx sdk.Context, sender sdk.AccAddress, sharesToMigrate sdk.Coin, tokenOutMins sdk.Coins) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, poolIdLeaving, poolIdEntering uint64, err error)
}

type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error
	AddSupplyOffset(ctx sdk.Context, denom string, offsetAmount sdk.Int)
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
	InstantUndelegate(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec) (sdk.Coins, error)
	GetDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (delegation stakingtypes.Delegation, found bool)
	UnbondingTime(ctx sdk.Context) time.Duration
	GetParams(ctx sdk.Context) stakingtypes.Params

	IterateBondedValidatorsByPower(ctx sdk.Context, fn func(int64, stakingtypes.ValidatorI) bool)
	TotalBondedTokens(ctx sdk.Context) sdk.Int
	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress, fn func(int64, stakingtypes.DelegationI) bool)
}

// CommunityPoolKeeper expected distribution keeper.
type CommunityPoolKeeper interface {
	WithdrawDelegationRewards(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error)
}

// IncentivesKeeper expected incentives keeper.
type IncentivesKeeper interface {
	CreateGauge(ctx sdk.Context, isPerpetual bool, owner sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpochsPaidOver uint64, poolId uint64) (uint64, error)
	AddToGaugeRewards(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, gaugeID uint64) error

	GetActiveGauges(ctx sdk.Context) []incentivestypes.Gauge
	Distribute(ctx sdk.Context, gauges []incentivestypes.Gauge) (sdk.Coins, error)

	GetParams(ctx sdk.Context) incentivestypes.Params
}

type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochstypes.EpochInfo
	NumBlocksSinceEpochStart(ctx sdk.Context, identifier string) (int64, error)
}

type ConcentratedKeeper interface {
	GetPosition(ctx sdk.Context, positionId uint64) (model.Position, error)
	SetPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, joinTime time.Time, liquidity sdk.Dec, positionId uint64, underlyingLockId uint64) error
	UpdatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityDelta sdk.Dec, joinTime time.Time, positionId uint64) (sdk.Int, sdk.Int, bool, bool, error)
	GetConcentratedPoolById(ctx sdk.Context, poolId uint64) (cltypes.ConcentratedPoolExtension, error)
	CreateFullRangePositionLocked(ctx sdk.Context, clPoolId uint64, owner sdk.AccAddress, coins sdk.Coins, remainingLockDuration time.Duration) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, concentratedLockID uint64, err error)
	CreateFullRangePositionUnlocking(ctx sdk.Context, clPoolId uint64, owner sdk.AccAddress, coins sdk.Coins, remainingLockDuration time.Duration) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, concentratedLockID uint64, err error)
	GetPositionIdToLockId(ctx sdk.Context, underlyingLockId uint64) (uint64, error)
	GetFullRangeLiquidityInPool(ctx sdk.Context, poolId uint64) (sdk.Dec, error)
	PositionHasActiveUnderlyingLock(ctx sdk.Context, positionId uint64) (bool, uint64, error)
	HasAnyPositionForPool(ctx sdk.Context, poolId uint64) (bool, error)
	WithdrawPosition(ctx sdk.Context, owner sdk.AccAddress, positionId uint64, requestedLiquidityAmountToWithdraw sdk.Dec) (amtDenom0, amtDenom1 sdk.Int, err error)
	GetUserPositions(ctx sdk.Context, addr sdk.AccAddress, poolId uint64) ([]model.Position, error)
	GetLockIdFromPositionId(ctx sdk.Context, positionId uint64) (uint64, error)
}
