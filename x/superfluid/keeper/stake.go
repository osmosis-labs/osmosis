package keeper

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) GetTotalSyntheticAssetsLocked(ctx sdk.Context, denom string) sdk.Int {
	return k.lk.GetPeriodLocksAccumulation(ctx, lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         denom,
		Duration:      k.sk.UnbondingTime(ctx),
	})
}

func (k Keeper) GetExpectedDelegationAmount(ctx sdk.Context, acc types.SuperfluidIntermediaryAccount) sdk.Int {
	// Get total number of Osmo this account should have delegated after refresh
	// (1) Find how many tokens total T are locked for (denom, validator) pair
	totalSuperfluidDelegation := k.GetTotalSyntheticAssetsLocked(ctx, stakingSyntheticDenom(acc.Denom, acc.ValAddr))
	// (2) Multiply the T tokens, by the number of superfluid osmo per token, to get the total amount
	// of osmo we expect.
	refreshedAmount := k.GetSuperfluidOSMOTokens(ctx, acc.Denom, totalSuperfluidDelegation)
	return refreshedAmount
}

func (k Keeper) RefreshIntermediaryDelegationAmounts(ctx sdk.Context) {
	// iterate over every (denom, validator) pair
	accs := k.GetAllIntermediaryAccounts(ctx)
	for _, acc := range accs {
		mAddr := acc.GetAccAddress()

		valAddress, err := sdk.ValAddressFromBech32(acc.ValAddr)
		if err != nil {
			panic(err)
		}

		validator, found := k.sk.GetValidator(ctx, valAddress)
		if !found {
			k.Logger(ctx).Error(fmt.Sprintf("validator not found or %s", acc.ValAddr))
			continue
		}

		currentAmount := sdk.NewInt(0)
		delegation, found := k.sk.GetDelegation(ctx, mAddr, valAddress)
		if !found {
			// continue if current delegation is 0, in case its really a dust delegation
			// that becomes worth something after refresh.
			k.Logger(ctx).Info(fmt.Sprintf("Existing delegation not found for %s with %s during superfluid refresh."+
				" It may have been previously bonded, but now unbonded.", mAddr.String(), acc.ValAddr))
		} else {
			// TODO: Be consistent withn TokensFromShares vs ValidateFromUnbondAmount
			currentAmount = validator.TokensFromShares(delegation.Shares).RoundInt()
		}

		refreshedAmount := k.GetExpectedDelegationAmount(ctx, acc)

		if refreshedAmount.GT(currentAmount) {
			adjustment := refreshedAmount.Sub(currentAmount)
			err = k.mintOsmoTokensAndDelegate(ctx, adjustment, acc)
			if err != nil {
				ctx.Logger().Error("Error in forceUndelegateAndBurnOsmoTokens, state update reverted", err)
			}
		} else if currentAmount.GT(refreshedAmount) {
			// In this case, we want to change the IA's delegated balance to be refreshed Amount
			// which is less than what it already has.
			// This means we need to "InstantUndelegate" some of its delegation (not going through the unbonding queue)
			// and then burn that excessly delegated bits.
			adjustment := currentAmount.Sub(refreshedAmount)

			err := k.forceUndelegateAndBurnOsmoTokens(ctx, adjustment, acc)
			if err != nil {
				ctx.Logger().Error("Error in forceUndelegateAndBurnOsmoTokens, state update reverted", err)
			}
		} else {
			ctx.Logger().Info("Intermediary account already has correct delegation amount?" +
				" This with high probability implies the exact same spot price as the last epoch," +
				"and no delegation changes.")
		}
	}
}

func (k Keeper) IncreaseSuperfluidDelegation(ctx sdk.Context, lockID uint64, amount sdk.Coins) error {
	acc, found := k.GetIntermediaryAccountFromLockId(ctx, lockID)
	if !found {
		return nil
	}

	// mint OSMO token based on TWAP of locked denom to denom module account
	osmoAmt := k.GetSuperfluidOSMOTokens(ctx, acc.Denom, amount.AmountOf(acc.Denom))
	if osmoAmt.IsZero() {
		return nil
	}

	err := k.mintOsmoTokensAndDelegate(ctx, osmoAmt, acc)
	if err != nil {
		return err
	}

	return nil
}

// basic validation for locks, that sender is correct, and that the lock length is correct.
func (k Keeper) validateLockForSF(ctx sdk.Context, lock *lockuptypes.PeriodLock, sender string) error {
	if lock.Owner != sender {
		return lockuptypes.ErrNotLockOwner
	}
	if lock.Coins.Len() != 1 {
		return types.ErrMultipleCoinsLockupNotSupported
	}
	return nil
}

func (k Keeper) validateLockForSFDelegate(ctx sdk.Context, lock *lockuptypes.PeriodLock, sender string) error {
	err := k.validateLockForSF(ctx, lock, sender)
	if err != nil {
		return err
	}
	defaultSuperfluidAsset := types.SuperfluidAsset{}
	if k.GetSuperfluidAsset(ctx, lock.Coins[0].Denom) == defaultSuperfluidAsset {
		return types.ErrNonSuperfluidAsset
	}

	// prevent unbonding lockups to be not able to be used for superfluid staking
	if lock.IsUnlocking() {
		return types.ErrUnbondingLockupNotSupported
	}

	// ensure that lock duration >= staking.UnbondingTime
	if lock.Duration < k.sk.GetParams(ctx).UnbondingTime {
		return types.ErrNotEnoughLockupDuration
	}

	// Thus when we stake now, this will be the only superfluid position for this lockID.
	if k.alreadySuperfluidStaking(ctx, lock.ID) {
		return types.ErrAlreadyUsedSuperfluidLockup
	}

	return nil
}

// ensure the valAddr is correctly formatted & corresponds to a real validator on chain.
func (k Keeper) validateValAddrForDelegate(ctx sdk.Context, valAddr string) (stakingtypes.Validator, error) {
	valAddress, err := sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		return stakingtypes.Validator{}, err
	}
	validator, found := k.sk.GetValidator(ctx, valAddress)
	if !found {
		return stakingtypes.Validator{}, stakingtypes.ErrNoValidatorFound
	}
	return validator, nil
}

// TODO: Merge a lot of logic with IncreaseSuperfluidDelegation.
func (k Keeper) SuperfluidDelegate(ctx sdk.Context, sender string, lockID uint64, valAddr string) error {
	lock, err := k.lk.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	// This guarantees the lockID does not already have a superfluid stake position
	// associated with it, the lock is sufficiently long, the lock only locks one asset, etc.
	// Thus when we stake this lock, it will be the only superfluid position for this lockID.
	err = k.validateLockForSFDelegate(ctx, lock, sender)
	if err != nil {
		return err
	}
	lockedCoin := lock.Coins[0]

	// get the intermediate account for this (denom, validator) pair.
	// This account tracks the amount of osmo being considered as staked.
	// If an intermediary account doesn't exist, then create it + a perpetual gauge.
	acc, err := k.GetOrCreateIntermediaryAccount(ctx, lockedCoin.Denom, valAddr)
	if err != nil {
		return err
	}
	// create connection record between lock id and intermediary account
	k.SetLockIdIntermediaryAccountConnection(ctx, lockID, acc)

	// Register a synthetic lockup for superfluid staking
	err = k.createSyntheticLockup(ctx, lockID, acc, bondedStatus)
	if err != nil {
		return err
	}

	// Find how many new osmo tokens this delegation is worth at superfluids current risk adjustment
	// and twap of the denom.
	amount := k.GetSuperfluidOSMOTokens(ctx, acc.Denom, lockedCoin.Amount)
	if amount.IsZero() {
		return types.ErrOsmoEquivalentZeroNotAllowed
	}

	return k.mintOsmoTokensAndDelegate(ctx, amount, acc)
}

func (k Keeper) SuperfluidUndelegate(ctx sdk.Context, sender string, lockID uint64) error {
	lock, err := k.lk.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}
	err = k.validateLockForSF(ctx, lock, sender)
	if err != nil {
		return err
	}
	lockedCoin := lock.Coins[0]

	// get the intermediate acct asscd. with lock id, and delete the connection.
	intermediaryAcc, found := k.GetIntermediaryAccountFromLockId(ctx, lockID)
	if !found {
		return types.ErrNotSuperfluidUsedLockup
	}
	k.DeleteLockIdIntermediaryAccountConnection(ctx, lockID)

	// Delete the old synthetic lockup, and create a new synthetic lockup representing the unstaking
	synthdenom := stakingSyntheticDenom(lockedCoin.Denom, intermediaryAcc.ValAddr)
	err = k.lk.DeleteSyntheticLockup(ctx, lockID, synthdenom)
	if err != nil {
		return err
	}

	// undelegate this lock's delegation amount, and burn the minted osmo.
	amount := k.GetSuperfluidOSMOTokens(ctx, intermediaryAcc.Denom, lockedCoin.Amount)
	err = k.forceUndelegateAndBurnOsmoTokens(ctx, amount, intermediaryAcc)
	if err != nil {
		return err
	}

	// Create a new synthetic lockup representing the unstaking side.
	return k.createSyntheticLockup(ctx, lockID, intermediaryAcc, unlockingStatus)
}

func (k Keeper) SuperfluidUnbondLock(ctx sdk.Context, underlyingLockId uint64, sender string) error {
	lock, err := k.lk.GetLockByID(ctx, underlyingLockId)
	if err != nil {
		return err
	}
	err = k.validateLockForSF(ctx, lock, sender)
	if err != nil {
		return err
	}
	synthLocks := k.lk.GetAllSyntheticLockupsByLockup(ctx, underlyingLockId)
	if len(synthLocks) != 1 {
		return types.ErrNotSuperfluidUsedLockup
	}
	if !synthLocks[0].IsUnlocking() {
		return types.ErrBondingLockupNotSupported
	}
	return k.lk.BeginForceUnlock(ctx, underlyingLockId, sdk.Coins{})
}

func (k Keeper) alreadySuperfluidStaking(ctx sdk.Context, lockID uint64) bool {
	// We need to catch two cases:
	// (1) lockID has another superfluid bond
	// (2) lockID has a superfluid unbonding
	// we check (1) by looking for presence of an intermediary account lock ID connection
	// we check (2) (and re-check 1 for suredness) by looking for the existence of
	// synthetic locks for this.
	intermediaryAccAddr := k.GetLockIdIntermediaryAccountConnection(ctx, lockID)
	if !intermediaryAccAddr.Empty() {
		return true
	}

	synthLocks := k.lk.GetAllSyntheticLockupsByLockup(ctx, lockID)
	return len(synthLocks) > 0
}

// mint osmoAmount of OSMO tokens, and immediately delegate them to validator on behalf of intermediary account.
func (k Keeper) mintOsmoTokensAndDelegate(ctx sdk.Context, osmoAmount sdk.Int, intermediaryAccount types.SuperfluidIntermediaryAccount) error {
	validator, err := k.validateValAddrForDelegate(ctx, intermediaryAccount.ValAddr)
	if err != nil {
		return err
	}

	err = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
		bondDenom := k.sk.BondDenom(cacheCtx)
		coins := sdk.Coins{sdk.NewCoin(bondDenom, osmoAmount)}
		err = k.bk.MintCoins(cacheCtx, types.ModuleName, coins)
		if err != nil {
			return err
		}
		k.bk.AddSupplyOffset(cacheCtx, bondDenom, osmoAmount.Neg())
		err = k.bk.SendCoinsFromModuleToAccount(cacheCtx, types.ModuleName, intermediaryAccount.GetAccAddress(), coins)
		if err != nil {
			return err
		}

		// make delegation from module account to the validator
		// TODO: What happens here if validator is jailed, tombstoned, or unbonding
		// For now, we don't worry since worst case it errors, in which case we revert mint.
		_, err = k.sk.Delegate(cacheCtx,
			intermediaryAccount.GetAccAddress(),
			osmoAmount, stakingtypes.Unbonded, validator, true)
		return err
	})
	return err
}

// force undelegate osmoAmount worth of delegation shares from delegations between intermediary account and valAddr
// We take the returned tokens, and then immediately burn them.
func (k Keeper) forceUndelegateAndBurnOsmoTokens(ctx sdk.Context,
	osmoAmount sdk.Int, intermediaryAcc types.SuperfluidIntermediaryAccount,
) error {
	valAddr, err := sdk.ValAddressFromBech32(intermediaryAcc.ValAddr)
	if err != nil {
		return err
	}
	// TODO: Better understand and decide between ValidateUnbondAmount and SharesFromTokens
	// briefly looked into it, did not understand whats correct.
	// TODO: ensure that intermediate account has at least osmoAmount staked.
	shares, err := k.sk.ValidateUnbondAmount(
		ctx, intermediaryAcc.GetAccAddress(), valAddr, osmoAmount,
	)
	if err == stakingtypes.ErrNoDelegation {
		return nil
	} else if err != nil {
		return err
	}
	err = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
		undelegatedCoins, err := k.sk.InstantUndelegate(cacheCtx, intermediaryAcc.GetAccAddress(), valAddr, shares)
		if err != nil {
			return err
		}

		// TODO: Should we compare undelegatedCoins vs osmoAmount?
		err = k.bk.SendCoinsFromAccountToModule(cacheCtx, intermediaryAcc.GetAccAddress(), types.ModuleName, undelegatedCoins)
		if err != nil {
			return err
		}
		err = k.bk.BurnCoins(cacheCtx, types.ModuleName, undelegatedCoins)
		if err != nil {
			return err
		}
		bondDenom := k.sk.BondDenom(cacheCtx)
		k.bk.AddSupplyOffset(cacheCtx, bondDenom, undelegatedCoins.AmountOf(bondDenom))

		return err
	})

	return err
}

// TODO: Need to (eventually) override the existing staking messages and queries, for undelegating, delegating, rewards, and redelegating, to all be going through all superfluid module.
// Want integrators to be able to use the same staking queries and messages
// Eugen’s point: Only rewards message needs to be updated. Rest of messages are fine
// Queries need to be updated
// We can do this at the very end though, since it just relates to queries.
