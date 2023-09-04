package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
)

// StakingInterface expected staking keeper.
type StakingInterface interface {
	GetAllValidators(ctx sdk.Context) (validators []stakingtypes.Validator)
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)
	Delegate(ctx sdk.Context, delAddr sdk.AccAddress, bondAmt osmomath.Int, tokenSrc stakingtypes.BondStatus, validator stakingtypes.Validator, subtractAccount bool) (newShares osmomath.Dec, err error)
	GetDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (delegation stakingtypes.Delegation, found bool)
	Undelegate(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount osmomath.Dec) (time.Time, error)
	BeginRedelegation(ctx sdk.Context, delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress, sharesAmount osmomath.Dec) (completionTime time.Time, err error)
	GetDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress, maxRetrieve uint16) (delegations []stakingtypes.Delegation)
	GetValidators(ctx sdk.Context, maxRetrieve uint32) (validators []stakingtypes.Validator)
}

type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

type DistributionKeeper interface {
	WithdrawDelegationRewards(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error)
	IncrementValidatorPeriod(ctx sdk.Context, val stakingtypes.ValidatorI) uint64
	CalculateDelegationRewards(ctx sdk.Context, val stakingtypes.ValidatorI, del stakingtypes.DelegationI, endingPeriod uint64) (rewards sdk.DecCoins)
	AllocateTokensToValidator(ctx sdk.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins)
}
type LockupKeeper interface {
	GetLockByID(ctx sdk.Context, lockID uint64) (*lockuptypes.PeriodLock, error)
	GetSyntheticLockupByUnderlyingLockId(ctx sdk.Context, lockID uint64) (lockuptypes.SyntheticLock, bool, error)
	ForceUnlock(ctx sdk.Context, lock lockuptypes.PeriodLock) error
	BeginUnlock(ctx sdk.Context, lockID uint64, coins sdk.Coins) (uint64, error)
	GetPeriodLocks(ctx sdk.Context) ([]lockuptypes.PeriodLock, error)
}
