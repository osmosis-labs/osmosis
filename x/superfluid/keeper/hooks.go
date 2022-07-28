package keeper

import (
	"time"

	epochstypes "github.com/osmosis-labs/osmosis/v10/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v10/x/superfluid/keeper/internal/events"

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

// epochs hooks
// Don't do anything pre epoch start.
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
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

func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, infractionHeight int64, slashFactor sdk.Dec, effectiveSlashFactor sdk.Dec) {
	if slashFactor.IsZero() {
		return
	}
	h.k.SlashLockupsForValidatorSlash(ctx, valAddr, infractionHeight, slashFactor)
}

func (h Hooks) AfterValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, infractionHeight int64, slashFactor sdk.Dec, effectiveSlashFactor sdk.Dec) {
	if slashFactor.IsZero() {
		return
	}
	h.k.RefreshIntermediaryDelegationAmounts(ctx)
}
