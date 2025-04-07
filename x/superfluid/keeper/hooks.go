package keeper

import (
	"context"
	"time"

	"github.com/osmosis-labs/osmosis/osmomath"
	epochstypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/keeper/internal/events"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Hooks wrapper struct for incentives keeper.
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Return the wrapper struct.
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// GetModuleName implements types.EpochHooks.
func (Hooks) GetModuleName() string {
	return types.ModuleName
}

// epochs hooks
// Don't do anything pre epoch start.
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}

// lockup hooks
// if you add tokens to a lock that is superfluid unbonding, nothing happens superfluid side.
// This lock does as an edge case take on the slashing risk as well for historical slashes.
// This is deemed as fine, governance can re-pay if it occurs on mainnet.
func (h Hooks) AfterAddTokensToLock(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins) {
	intermediaryAccAddr := h.k.GetLockIdIntermediaryAccountConnection(ctx, lockID)
	if !intermediaryAccAddr.Empty() {
		// superfluid delegate for additional amount
		err := h.k.IncreaseSuperfluidDelegation(ctx, lockID, amount)
		if err != nil {
			h.k.Logger(ctx).Error(err.Error())
		} else {
			events.EmitSuperfluidIncreaseDelegationEvent(ctx, lockID, amount)
		}
	}
}

func (h Hooks) OnTokenLocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
}

func (h Hooks) OnStartUnlock(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
}

func (h Hooks) OnTokenUnlocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
}

func (h Hooks) OnTokenSlashed(ctx sdk.Context, lockID uint64, amount sdk.Coins) {
}

func (h Hooks) OnLockupExtend(ctx sdk.Context, lockID uint64, oldDuration, newDuration time.Duration) {
}

// staking hooks.
func (h Hooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBonded(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterUnbondingInitiated(_ context.Context, _ uint64) error {
	return nil
}

func (h Hooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, slashFactor osmomath.Dec) error {
	if slashFactor.IsZero() {
		return nil
	}
	h.k.SlashLockupsForValidatorSlash(ctx, valAddr, slashFactor)
	return nil
}

func (h Hooks) AfterValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, infractionHeight int64, slashFactor osmomath.Dec, effectiveSlashFactor osmomath.Dec) {
	if slashFactor.IsZero() {
		return
	}
	accs := h.k.GetAllIntermediaryAccounts(ctx)
	h.k.RefreshIntermediaryDelegationAmounts(ctx, accs)
}
