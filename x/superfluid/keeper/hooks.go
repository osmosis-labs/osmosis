package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	epochstypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"
)

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks
// Don't do anything pre epoch start
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}

// lockup hooks
func (h Hooks) AfterAddTokensToLock(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins) {
	intermediaryAccAddr := h.k.GetLockIdIntermediaryAccountConnection(ctx, lockID)
	if !intermediaryAccAddr.Empty() {
		// superfluid delegate for additional amount
		err := h.k.SuperfluidDelegateMore(ctx, lockID, amount)
		if err != nil {
			h.k.Logger(ctx).Error(err.Error())
		}
	}
}

func (h Hooks) OnTokenLocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
}

func (h Hooks) OnStartUnlock(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
	// undelegate automatically when start unlocking if superfluid staking is available
	intermediaryAccAddr := h.k.GetLockIdIntermediaryAccountConnection(ctx, lockID)
	if !intermediaryAccAddr.Empty() {
		_, err := h.k.SuperfluidUndelegate(ctx, address.String(), lockID)
		if err != nil {
			h.k.Logger(ctx).Error(err.Error())
			// TODO: If not panic, there could be the case user get infinite amount of rewards without actual lockup
			panic(err)
		}
	}
}

func (h Hooks) OnTokenUnlocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {

}

func (h Hooks) OnTokenSlashed(ctx sdk.Context, lockID uint64, amount sdk.Coins) {

}

// staking hooks
func (h Hooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress)   {}
func (h Hooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {}
func (h Hooks) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, slashFactor sdk.Dec) {
	if slashFactor == sdk.ZeroDec() {
		return
	}
	h.k.SlashLockupsForValidatorSlash(ctx, valAddr, slashFactor)
}
func (h Hooks) BeforeSlashingUnbondingDelegation(ctx sdk.Context, unbondingDelegation stakingtypes.UnbondingDelegation,
	infractionHeight int64, slashFactor sdk.Dec) {
	if slashFactor == sdk.ZeroDec() {
		return
	}
	h.k.SlashLockupsForUnbondingDelegationSlash(ctx, unbondingDelegation.DelegatorAddress, unbondingDelegation.ValidatorAddress, slashFactor)
}

// Not used right now, as we don't allow superfluid redelegations
func (h Hooks) BeforeSlashingRedelegation(ctx sdk.Context, srcValidator stakingtypes.Validator, redelegation stakingtypes.Redelegation,
	infractionHeight int64, slashFactor sdk.Dec) {
}
