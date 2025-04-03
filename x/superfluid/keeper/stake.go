package keeper

import (
	"context"
	"errors"
	"fmt"
	"strings"

	addresscodec "cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
	valsettypes "github.com/osmosis-labs/osmosis/v27/x/valset-pref/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetTotalSyntheticAssetsLocked returns the total amount of the given denom locked.
func (k Keeper) GetTotalSyntheticAssetsLocked(ctx sdk.Context, denom string) (osmomath.Int, error) {
	unbondingTime, err := k.sk.UnbondingTime(ctx)
	if err != nil {
		return osmomath.Int{}, err
	}
	return k.lk.GetPeriodLocksAccumulation(ctx, lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         denom,
		Duration:      unbondingTime,
	}), nil
}

// GetExpectedDelegationAmount returns the total number of osmo the intermediary account
// has delegated using the most recent osmo equivalent multiplier.
// This is labeled as expected because the way it calculates the amount can
// lead rounding errors from the true delegated amount.
func (k Keeper) GetExpectedDelegationAmount(ctx sdk.Context, acc types.SuperfluidIntermediaryAccount) (osmomath.Int, error) {
	// (1) Find how many tokens total T are locked for (denom, validator) pair
	totalSuperfluidDelegation, err := k.GetTotalSyntheticAssetsLocked(ctx, stakingSyntheticDenom(acc.Denom, acc.ValAddr))
	if err != nil {
		return osmomath.Int{}, err
	}
	// (2) Multiply the T tokens, by the number of superfluid osmo per token, to get the total amount
	// of osmo we expect.
	refreshedAmount, err := k.GetSuperfluidOSMOTokens(ctx, acc.Denom, totalSuperfluidDelegation)
	if err != nil {
		return osmomath.Int{}, err
	}
	return refreshedAmount, nil
}

// RefreshIntermediaryDelegationAmounts refreshes the amount of delegation for all intermediary accounts.
// This method includes minting new osmo if the refreshed delegation amount has increased, and
// instantly undelegating and burning if the refreshed delegation has decreased.
func (k Keeper) RefreshIntermediaryDelegationAmounts(context context.Context, accs []types.SuperfluidIntermediaryAccount) {
	ctx := sdk.UnwrapSDKContext(context)
	// iterate over all intermedairy accounts - every (denom, validator) pair
	for _, acc := range accs {
		mAddr := acc.GetAccAddress()

		valAddress, err := sdk.ValAddressFromBech32(acc.ValAddr)
		if err != nil {
			panic(err)
		}

		validator, err := k.sk.GetValidator(ctx, valAddress)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("validator not found or %s", acc.ValAddr))
			continue
		}

		currentAmount := osmomath.NewInt(0)
		delegation, err := k.sk.GetDelegation(ctx, mAddr, valAddress)
		if err != nil {
			// continue if current delegation return an error, in case its really a dust delegation
			// that becomes worth something after refresh.
			// TODO: We have a correct explanation for this in some github issue, lets amend this correctly.
			k.Logger(ctx).Debug(err.Error())
		} else {
			currentAmount = validator.TokensFromShares(delegation.Shares).RoundInt()
		}

		refreshedAmount, err := k.GetExpectedDelegationAmount(ctx, acc)
		if err != nil {
			ctx.Logger().Error("Error in GetExpectedDelegationAmount (likely that underlying LP share is no longer superfluid capable), state update reverted", err)
		}

		if refreshedAmount.GT(currentAmount) {
			adjustment := refreshedAmount.Sub(currentAmount)
			err = k.mintOsmoTokensAndDelegate(ctx, adjustment, acc)
			if err != nil {
				ctx.Logger().Error("Error in mintOsmoTokensAndDelegate, state update reverted", err)
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
			ctx.Logger().Debug("Intermediary account already has correct delegation amount?" +
				" This with high probability implies the exact same spot price as the last epoch," +
				"and no delegation changes.")
		}
	}
}

// IncreaseSuperfluidDelegation increases the amount of existing superfluid delegation.
// This method would return an error if the lock has not been superfluid delegated before.
func (k Keeper) IncreaseSuperfluidDelegation(ctx sdk.Context, lockID uint64, amount sdk.Coins) error {
	acc, found := k.GetIntermediaryAccountFromLockId(ctx, lockID)
	if !found {
		return nil
	}

	// mint OSMO token based on the most recent osmo equivalent multiplier
	// of locked denom to denom module account
	osmoAmt, err := k.GetSuperfluidOSMOTokens(ctx, acc.Denom, amount.AmountOf(acc.Denom))
	if err != nil {
		return err
	}
	if osmoAmt.IsZero() {
		return nil
	}

	err = k.mintOsmoTokensAndDelegate(ctx, osmoAmt, acc)
	if err != nil {
		return err
	}

	return nil
}

// basic validation for locks to be eligible for superfluid delegation. This includes checking
// - that the sender is the owner of the lock
// - that the lock is consisted of single coin
func (k Keeper) validateLockForSF(lock *lockuptypes.PeriodLock, sender string) error {
	if lock.Owner != sender {
		return lockuptypes.ErrNotLockOwner
	}
	if lock.Coins.Len() != 1 {
		return types.ErrMultipleCoinsLockupNotSupported
	}
	return nil
}

// validateLockForSFDelegate runs the following sanity checks on the lock:
// - the sender is the owner of the lock
// - the lock is consisted of a single coin
// - the asset is registered as a superfluid asset via governance
// - the lock is not unlocking
// - lock duration is greater or equal to the unbonding time
// - lock should not be already superfluid staked
func (k Keeper) validateLockForSFDelegate(ctx sdk.Context, lock *lockuptypes.PeriodLock, sender string) error {
	err := k.validateLockForSF(lock, sender)
	if err != nil {
		return err
	}

	denom := lock.Coins[0].Denom

	// ensure that the locks underlying denom is for an existing superfluid asset
	_, err = k.GetSuperfluidAsset(ctx, denom)
	if err != nil {
		return err
	}

	// prevent unbonding lockups to be not able to be used for superfluid staking
	if lock.IsUnlocking() {
		return errorsmod.Wrapf(types.ErrUnbondingLockupNotSupported, "lock id : %d", lock.ID)
	}

	// ensure that lock duration >= staking.UnbondingTime
	stakingParams, err := k.sk.GetParams(ctx)
	if err != nil {
		return err
	}
	unbondingTime := stakingParams.UnbondingTime
	if lock.Duration < unbondingTime {
		return errorsmod.Wrapf(types.ErrNotEnoughLockupDuration, "lock duration (%d) must be greater than unbonding time (%d)", lock.Duration, unbondingTime)
	}

	// Thus when we stake now, this will be the only superfluid position for this lockID.
	if k.alreadySuperfluidStaking(ctx, lock.ID) {
		return errorsmod.Wrapf(types.ErrAlreadyUsedSuperfluidLockup, "lock id : %d", lock.ID)
	}

	return nil
}

// ensure the valAddr is correctly formatted & corresponds to a real validator on chain.
func (k Keeper) validateValAddrForDelegate(ctx sdk.Context, valAddr string) (stakingtypes.Validator, error) {
	valAddress, err := sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		return stakingtypes.Validator{}, err
	}
	validator, err := k.sk.GetValidator(ctx, valAddress)
	if err != nil {
		return stakingtypes.Validator{}, stakingtypes.ErrNoValidatorFound
	}
	return validator, nil
}

// SuperfluidDelegate superfluid delegates osmo equivalent amount the given lock holds.
// The actual delegation is done by using/creating an intermediary account for the (denom, validator) pair
// and having the intermediary account delegate to the designated validator, not by the sender themselves.
// A state entry of IntermediaryAccountConnection is stored to store the connection between the lock ID
// and the intermediary account, as an intermediary account does not serve for delegations from a single delegator.
// The actual amount of delegation is not equal to the equivalent amount of osmo the lock has. That is,
// the actual amount of delegation is amount * osmo equivalent multiplier * (1 - k.RiskFactor(asset)).
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
	amount, err := k.GetSuperfluidOSMOTokens(ctx, acc.Denom, lockedCoin.Amount)
	if err != nil {
		return err
	}
	if amount.IsZero() {
		return types.ErrOsmoEquivalentZeroNotAllowed
	}

	return k.mintOsmoTokensAndDelegate(ctx, amount, acc)
}

// undelegateCommon is a helper function for SuperfluidUndelegate and superfluidUndelegateToConcentratedPosition.
// It performs the following tasks:
// - checks that the lock is valid for superfluid staking
// - gets the intermediary account associated with the lock id
// - deletes the connection between the lock id and the intermediary account
// - deletes the synthetic lockup associated with the lock id
// - undelegates the superfluid staking position associated with the lock id and burns the underlying osmo tokens
// - returns the intermediary account
func (k Keeper) undelegateCommon(ctx sdk.Context, sender string, lockID uint64) (types.SuperfluidIntermediaryAccount, error) {
	lock, err := k.lk.GetLockByID(ctx, lockID)
	if err != nil {
		return types.SuperfluidIntermediaryAccount{}, err
	}
	err = k.validateLockForSF(lock, sender)
	if err != nil {
		return types.SuperfluidIntermediaryAccount{}, err
	}
	lockedCoin := lock.Coins[0]

	// get the intermediate account associated with lock id, and delete the connection.
	intermediaryAcc, found := k.GetIntermediaryAccountFromLockId(ctx, lockID)
	if !found {
		return types.SuperfluidIntermediaryAccount{}, types.ErrNotSuperfluidUsedLockup
	}
	k.DeleteLockIdIntermediaryAccountConnection(ctx, lockID)

	// Delete the old synthetic lockup
	synthdenom := stakingSyntheticDenom(lockedCoin.Denom, intermediaryAcc.ValAddr)
	err = k.lk.DeleteSyntheticLockup(ctx, lockID, synthdenom)
	if err != nil {
		return types.SuperfluidIntermediaryAccount{}, err
	}

	// undelegate this lock's delegation amount, and burn the minted osmo.
	amount, err := k.GetSuperfluidOSMOTokens(ctx, intermediaryAcc.Denom, lockedCoin.Amount)
	if err != nil {
		return types.SuperfluidIntermediaryAccount{}, err
	}
	err = k.forceUndelegateAndBurnOsmoTokens(ctx, amount, intermediaryAcc)
	if err != nil {
		return types.SuperfluidIntermediaryAccount{}, err
	}
	return intermediaryAcc, nil
}

// SuperfluidUndelegate starts undelegating superfluid delegated position for the given lock.
// Undelegation is done instantly and the equivalent amount is sent to the module account
// where it is burnt. Note that this method does not include unbonding the lock
// itself.
func (k Keeper) SuperfluidUndelegate(ctx sdk.Context, sender string, lockID uint64) error {
	intermediaryAcc, err := k.undelegateCommon(ctx, sender, lockID)
	if err != nil {
		return err
	}
	// Create a new synthetic lockup representing the unstaking side.
	return k.createSyntheticLockup(ctx, lockID, intermediaryAcc, unlockingStatus)
}

// SuperfluidUndelegateToConcentratedPosition starts undelegating superfluid delegated position for the given lock. It behaves similarly to SuperfluidUndelegate,
// however it does not create a new synthetic lockup representing the unstaking side. This is because after the time this function is called, we might
// want to perform more operations prior to creating a lock. Once the actual lock is created, the synthetic lockup representing the unstaking side
// should eventually be created as well. Use this function with caution to avoid accidentally missing synthetic lock creation.
func (k Keeper) SuperfluidUndelegateToConcentratedPosition(ctx sdk.Context, sender string, gammLockID uint64) (types.SuperfluidIntermediaryAccount, error) {
	return k.undelegateCommon(ctx, sender, gammLockID)
}

// SuperfluidUnbondLock unbonds the lock that has been used for superfluid staking.
// This method would return an error if the underlying lock is not superfluid undelegating.
func (k Keeper) SuperfluidUnbondLock(ctx sdk.Context, underlyingLockId uint64, sender string) error {
	_, err := k.unbondLock(ctx, underlyingLockId, sender, sdk.Coins{})
	return err
}

// SuperfluidUndelegateAndUnbondLock unbonds given amount from the
// underlying lock that has been used for superfluid staking.
// This method returns the lock id, same lock id if unlock amount is equal to the
// underlying lock amount. Otherwise it returns the newly created lock id.
// Note that we can either partially or fully undelegate and unbond lock using this method.
func (k Keeper) SuperfluidUndelegateAndUnbondLock(ctx sdk.Context, lockID uint64, sender string, amount osmomath.Int) (uint64, error) {
	lock, err := k.lk.GetLockByID(ctx, lockID)
	if err != nil {
		return 0, err
	}

	coins := sdk.Coins{sdk.NewCoin(lock.Coins[0].Denom, amount)}
	if coins[0].IsZero() {
		return 0, errors.New("amount to unlock must be greater than 0")
	}
	if lock.Coins[0].IsLT(coins[0]) {
		return 0, errors.New("requested amount to unlock exceeds locked tokens")
	}

	// get intermediary account before connection is deleted in SuperfluidUndelegate
	intermediaryAcc, found := k.GetIntermediaryAccountFromLockId(ctx, lockID)
	if !found {
		return 0, types.ErrNotSuperfluidUsedLockup
	}

	// undelegate all
	err = k.SuperfluidUndelegate(ctx, sender, lockID)
	if err != nil {
		return 0, err
	}

	// unbond partial or full locked amount
	newLockID, err := k.unbondLock(ctx, lockID, sender, coins)
	if err != nil {
		return 0, err
	}

	// check new lock id
	// If unbond amount == locked amount, then the underlying lock was not split.
	// So we double check that newLockID == lockID, and return.
	// This has the same effect as calling SuperfluidUndelegate and then SuperfluidUnbondLock.
	// Otherwise unbond amount < locked amount, and the underlying lock was split.
	// lockID contains the amount still locked in the lockup module.
	// newLockID contains the amount unlocked.
	// We double check that newLockID != lockID and then proceed to re-delegate
	// the remainder (locked amount - unbond amount).
	if lock.Coins[0].IsEqual(coins[0]) {
		if newLockID != lockID {
			panic(fmt.Errorf("expected new lock id %v to = lock id %v", newLockID, lockID))
		}
		return lock.ID, nil
	} else {
		if newLockID == lockID {
			panic(fmt.Errorf("expected new lock id %v to != lock id %v", newLockID, lockID))
		}
	}

	// delete synthetic unlocking lock created in the last step of SuperfluidUndelegate
	synthdenom := unstakingSyntheticDenom(lock.Coins[0].Denom, intermediaryAcc.ValAddr)
	err = k.lk.DeleteSyntheticLockup(ctx, lockID, synthdenom)
	if err != nil {
		return 0, err
	}

	// re-delegate remainder
	err = k.SuperfluidDelegate(ctx, sender, lockID, intermediaryAcc.ValAddr)
	if err != nil {
		return 0, err
	}

	// create synthetic unlocking lock for newLockID
	err = k.createSyntheticLockup(ctx, newLockID, intermediaryAcc, unlockingStatus)
	if err != nil {
		return 0, err
	}
	return newLockID, nil
}

// unbondLock unlocks the underlying lock. Same lock id is returned if the amount to unlock
// is equal to the entire locked amount. Otherwise, the amount to unlock is less
// than the amount locked, it will return a new lock id which was created as an unlocking lock.
func (k Keeper) unbondLock(ctx sdk.Context, underlyingLockId uint64, sender string, coins sdk.Coins) (uint64, error) {
	lock, err := k.lk.GetLockByID(ctx, underlyingLockId)
	if err != nil {
		return 0, err
	}
	err = k.validateLockForSF(lock, sender)
	if err != nil {
		return 0, err
	}
	synthLock, _, err := k.lk.GetSyntheticLockupByUnderlyingLockId(ctx, underlyingLockId)
	if err != nil {
		return 0, err
	}
	// TODO: Use !found
	if synthLock == (lockuptypes.SyntheticLock{}) {
		return 0, types.ErrNotSuperfluidUsedLockup
	}
	if !synthLock.IsUnlocking() {
		return 0, types.ErrBondingLockupNotSupported
	}
	return k.lk.BeginForceUnlock(ctx, underlyingLockId, coins)
}

// alreadySuperfluidStaking returns true if underlying lock used in superfluid staking.
// This method would also return true for undelegating position for the lock.
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

	synthLock, _, err := k.lk.GetSyntheticLockupByUnderlyingLockId(ctx, lockID)
	if err != nil {
		return false
	}
	// TODO: return found
	return synthLock != (lockuptypes.SyntheticLock{})
}

// mintOsmoTokensAndDelegate mints osmoAmount of OSMO tokens, and immediately delegate them to validator on behalf of intermediary account.
func (k Keeper) mintOsmoTokensAndDelegate(ctx sdk.Context, osmoAmount osmomath.Int, intermediaryAccount types.SuperfluidIntermediaryAccount) error {
	validator, err := k.validateValAddrForDelegate(ctx, intermediaryAccount.ValAddr)
	if err != nil {
		return err
	}

	err = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
		bondDenom, err := k.sk.BondDenom(cacheCtx)
		if err != nil {
			return err
		}
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

// forceUndelegateAndBurnOsmoTokens force undelegates osmoAmount worth of delegation shares
// from delegations between intermediary account and valAddr.
// We take the returned tokens, and then immediately burn them.
func (k Keeper) forceUndelegateAndBurnOsmoTokens(ctx sdk.Context,
	osmoAmount osmomath.Int, intermediaryAcc types.SuperfluidIntermediaryAccount,
) error {
	valAddr, err := sdk.ValAddressFromBech32(intermediaryAcc.ValAddr)
	if err != nil {
		return err
	}
	// TODO: Better understand and decide between ValidateUnbondAmount and SharesFromTokens
	// briefly looked into it, did not understand what's correct.
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
		bondDenom, err := k.sk.BondDenom(cacheCtx)
		if err != nil {
			return err
		}
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

// IterateBondedValidatorsByPower implements govtypes.StakingKeeper
func (k Keeper) ValidatorAddressCodec() addresscodec.Codec {
	return k.sk.ValidatorAddressCodec()
}

// IterateBondedValidatorsByPower implements govtypes.StakingKeeper
func (k Keeper) IterateBondedValidatorsByPower(ctx context.Context, fn func(int64, stakingtypes.ValidatorI) bool) error {
	return k.sk.IterateBondedValidatorsByPower(ctx, fn)
}

// TotalBondedTokens implements govtypes.StakingKeeper
func (k Keeper) TotalBondedTokens(ctx context.Context) (osmomath.Int, error) {
	return k.sk.TotalBondedTokens(ctx)
}

// IterateDelegations implements govtypes.StakingKeeper
// Iterates through staking keeper's delegations, and then all of the superfluid delegations.
func (k Keeper) IterateDelegations(context context.Context, delegator sdk.AccAddress, fn func(int64, stakingtypes.DelegationI) bool) error {
	ctx := sdk.UnwrapSDKContext(context)
	// call the callback with the non-superfluid delegations
	var index int64
	err := k.sk.IterateDelegations(ctx, delegator, func(i int64, delegation stakingtypes.DelegationI) (stop bool) {
		index = i
		return fn(i, delegation)
	})
	if err != nil {
		return err
	}

	synthlocks := k.lk.GetAllSyntheticLockupsByAddr(ctx, delegator)
	for i, lock := range synthlocks {
		// get locked coin from the lock ID
		interim, ok := k.GetIntermediaryAccountFromLockId(ctx, lock.UnderlyingLockId)
		if !ok {
			return fmt.Errorf("intermediary account not found for lock id %d", lock.UnderlyingLockId)
		}

		lock, err := k.lk.GetLockByID(ctx, lock.UnderlyingLockId)
		if err != nil {
			ctx.Logger().Error("lockup retrieval failed with underlying lock", "Lock", lock, "Error", err)
			return err
		}

		coin, err := lock.SingleCoin()
		if err != nil {
			ctx.Logger().Error("lock fails to meet expected invariant, it contains multiple coins", "Lock", lock, "Error", err)
			return err
		}

		// get osmo-equivalent token amount
		amount, err := k.GetSuperfluidOSMOTokens(ctx, interim.Denom, coin.Amount)
		if err != nil {
			ctx.Logger().Error("failed to get osmo equivalent of token", "Denom", interim.Denom, "Amount", coin.Amount, "Error", err)
			return err
		}

		// get validator shares equivalent to the token amount
		valAddr, err := sdk.ValAddressFromBech32(interim.ValAddr)
		if err != nil {
			ctx.Logger().Error("failed to decode validator address", "Intermediary", interim.ValAddr, "LockID", lock.ID, "Error", err)
			return err
		}

		validator, err := k.sk.GetValidator(ctx, valAddr)
		if err != nil {
			ctx.Logger().Error("validator does not exist for lock", "Validator", valAddr, "LockID", lock.ID)
			return err
		}

		shares, err := validator.SharesFromTokens(amount)
		if err != nil {
			// tokens are not valid. continue.
			return err
		}

		// construct delegation and call callback
		delegation := stakingtypes.Delegation{
			DelegatorAddress: delegator.String(),
			ValidatorAddress: interim.ValAddr,
			Shares:           shares,
		}

		// if valid delegation has been found, increment delegation index
		fn(index+int64(i), delegation)
	}
	return nil
}

// UnbondConvertAndStake converts given lock to osmo and stakes it to given validator.
// Supports conversion of 1)superfluid bonded 2)superfluid undelegating 3)vanilla unlocking.
// Liquid gamm shares will not be supported for conversion.
// Delegation is done in the following logic:
// - If valAddr provided, single delegate.
// - If valAddr not provided and valset exists, valsetpref.Delegate
// - If valAddr not provided and valset delegation is not possible, refer back to original lock's superfluid validator if it was a superfluid lock
// - Else: error
func (k Keeper) UnbondConvertAndStake(ctx sdk.Context, lockID uint64, sender, valAddr string,
	minAmtToStake osmomath.Int, sharesToConvert sdk.Coin) (totalAmtConverted osmomath.Int, err error) {
	senderAddr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return osmomath.Int{}, err
	}

	// use getMigrationType method to check status of lock (either superfluid staked, superfluid unbonding, vanilla locked, unlocked)
	_, migrationType, err := k.getMigrationType(ctx, int64(lockID))
	if err != nil {
		return osmomath.Int{}, err
	}

	// if superfluid bonded, first change it into superfluid undelegate to burn minted osmo and instantly undelegate.
	if migrationType == SuperfluidBonded {
		_, err = k.undelegateCommon(ctx, sender, lockID)
		if err != nil {
			return osmomath.Int{}, err
		}
	}

	if migrationType == SuperfluidBonded || migrationType == SuperfluidUnbonding || migrationType == NonSuperfluid {
		totalAmtConverted, err = k.convertLockToStake(ctx, senderAddr, valAddr, lockID, minAmtToStake)
	} else if migrationType == Unlocked { // liquid gamm shares without locks
		totalAmtConverted, err = k.convertUnlockedToStake(ctx, senderAddr, valAddr, sharesToConvert, minAmtToStake)
	} else { // any other types of migration should fail
		return osmomath.Int{}, errors.New("unsupported staking conversion type")
	}

	if err != nil {
		return osmomath.Int{}, err
	}

	return totalAmtConverted, nil
}

// convertLockToStake handles locks that are superfluid bonded, superfluid unbonding, vanilla locked(unlocking) locks.
// Deletes all associated state, converts the lock itself to staking delegation by going through exit pool and swap.
func (k Keeper) convertLockToStake(ctx sdk.Context, sender sdk.AccAddress, valAddr string, lockId uint64,
	minAmtToStake osmomath.Int) (totalAmtConverted osmomath.Int, err error) {
	lock, err := k.lk.GetLockByID(ctx, lockId)
	if err != nil {
		return osmomath.Int{}, err
	}

	// check lock owner is sender
	if lock.Owner != sender.String() {
		return osmomath.ZeroInt(), types.LockOwnerMismatchError{
			LockId:        lock.ID,
			LockOwner:     lock.Owner,
			ProvidedOwner: sender.String(),
		}
	}

	lockCoin := lock.Coins[0]

	// Ensuring the sharesToMigrate contains gamm pool share prefix.
	if !strings.HasPrefix(lockCoin.Denom, gammtypes.GAMMTokenPrefix) {
		return osmomath.Int{}, types.SharesToMigrateDenomPrefixError{Denom: lockCoin.Denom, ExpectedDenomPrefix: gammtypes.GAMMTokenPrefix}
	}

	poolIdLeaving, err := gammtypes.GetPoolIdFromShareDenom(lockCoin.Denom)
	if err != nil {
		return osmomath.Int{}, err
	}

	var superfluidValAddr string
	interAcc, found := k.GetIntermediaryAccountFromLockId(ctx, lockId)
	if found {
		superfluidValAddr = interAcc.ValAddr
	}

	// Force unlock, validate the provided sharesToStake, and exit the balancer pool.
	// we exit with min token out amount zero since we are checking min amount designated to stake later on anyways.
	exitCoins, err := k.forceUnlockAndExitBalancerPool(ctx, sender, poolIdLeaving, lock, lockCoin, sdk.NewCoins(), false)
	if err != nil {
		return osmomath.Int{}, err
	}

	totalAmtConverted, err = k.convertGammSharesToOsmoAndStake(ctx, sender, valAddr, poolIdLeaving, exitCoins, minAmtToStake, superfluidValAddr)
	if err != nil {
		return osmomath.Int{}, err
	}

	return totalAmtConverted, nil
}

// convertUnlockedToStake converts liquid gamm shares to staking delegation.
// minAmtToStake works as slippage bound for the conversion process.
func (k Keeper) convertUnlockedToStake(ctx sdk.Context, sender sdk.AccAddress, valAddr string, sharesToStake sdk.Coin,
	minAmtToStake osmomath.Int) (totalAmtConverted osmomath.Int, err error) {
	if !strings.HasPrefix(sharesToStake.Denom, gammtypes.GAMMTokenPrefix) {
		return osmomath.Int{}, types.SharesToMigrateDenomPrefixError{Denom: sharesToStake.Denom, ExpectedDenomPrefix: gammtypes.GAMMTokenPrefix}
	}

	// Get the balancer poolId by parsing the gamm share denom.
	poolIdLeaving, err := gammtypes.GetPoolIdFromShareDenom(sharesToStake.Denom)
	if err != nil {
		return osmomath.Int{}, err
	}

	// Exit the balancer pool position.
	// we exit with min token out amount zero since we are checking min amount designated to stake later on anyways.
	exitCoins, err := k.gk.ExitPool(ctx, sender, poolIdLeaving, sharesToStake.Amount, sdk.NewCoins())
	if err != nil {
		return osmomath.Int{}, err
	}

	totalAmtConverted, err = k.convertGammSharesToOsmoAndStake(ctx, sender, valAddr, poolIdLeaving, exitCoins, minAmtToStake, "")
	if err != nil {
		return osmomath.Int{}, err
	}

	return totalAmtConverted, nil
}

// convertGammSharesToOsmoAndStake converts given gamm shares to osmo by swapping in the given pool
// then stakes it to the designated validator.
// minAmtToStake works as slippage bound, and would error if total amount being staked is less than min amount to stake.
// Depending on user inputs, valAddr and originalSuperfluidValAddr could be an empty string,
// each leading to a different delegation scenario.
func (k Keeper) convertGammSharesToOsmoAndStake(
	ctx sdk.Context,
	sender sdk.AccAddress, valAddr string,
	poolIdLeaving uint64, exitCoins sdk.Coins, minAmtToStake osmomath.Int, originalSuperfluidValAddr string,
) (totalAmtCoverted osmomath.Int, err error) {
	var nonOsmoCoins sdk.Coins
	bondDenom, err := k.sk.BondDenom(ctx)
	if err != nil {
		return osmomath.Int{}, err
	}

	// from the exit coins, separate non-bond denom and bond denom.
	for _, exitCoin := range exitCoins {
		// if coin is not uosmo, add it to non-osmo Coins
		if exitCoin.Denom != bondDenom {
			nonOsmoCoins = append(nonOsmoCoins, exitCoin)
		}
	}
	originalBondDenomAmt := exitCoins.AmountOf(bondDenom)

	// track how much non-uosmo tokens we have converted to uosmo
	totalAmtCoverted = osmomath.ZeroInt()

	// iterate over non-bond denom coins and swap them into bond denom
	for _, coinToConvert := range nonOsmoCoins {
		tokenOutAmt, _, err := k.pmk.SwapExactAmountIn(ctx, sender, poolIdLeaving, coinToConvert, bondDenom, osmomath.ZeroInt())
		if err != nil {
			return osmomath.Int{}, err
		}

		totalAmtCoverted = totalAmtCoverted.Add(tokenOutAmt)
	}

	// add the converted amount with the amount of osmo from exit coin to get total amount we would be staking
	totalAmtToStake := originalBondDenomAmt.Add(totalAmtCoverted)

	// check if the total amount to stake after all conversion is greater than provided min amount to stake
	if totalAmtToStake.LT(minAmtToStake) {
		return osmomath.Int{}, types.TokenConvertedLessThenDesiredStakeError{
			ActualTotalAmtToStake:   totalAmtToStake,
			ExpectedTotalAmtToStake: minAmtToStake,
		}
	}

	err = k.delegateBaseOnValsetPref(ctx, sender, valAddr, originalSuperfluidValAddr, totalAmtToStake)
	if err != nil {
		return osmomath.Int{}, err
	}

	return totalAmtToStake, nil
}

// delegateBaseOnValsetPref delegates based on given input parameters.
// valAddr and originalSuperfluidValAddr can be an empty string depending on user input and original lock's status.
// Delegation is done in the following logic:
// - If valAddr provided, single delegate.
// - If valAddr not provided and valset exists, valsetpref.Delegate
// - If valAddr not provided and valset delegation is not possible, refer back to original lock's superfluid validator if it was a superfluid lock
// - Else: error
func (k Keeper) delegateBaseOnValsetPref(ctx sdk.Context, sender sdk.AccAddress, valAddr, originalSuperfluidValAddr string, totalAmtToStake osmomath.Int) error {
	bondDenom, err := k.sk.BondDenom(ctx)
	if err != nil {
		return err
	}

	// if given valAddr is empty, we use delegation preference given from valset-pref module or reference from superfluid staking
	if valAddr == "" {
		err := k.vspk.DelegateToValidatorSet(ctx, sender.String(), sdk.NewCoin(bondDenom, totalAmtToStake))
		// if valset-pref delegation succeeded without error, end method
		if err == nil {
			return nil
		}

		// if valset-pref delegation errored due to no existing delegation existing, fall back and try using superfluid staked validator
		if err == valsettypes.ErrNoDelegation {
			valAddr = originalSuperfluidValAddr
		} else if err != nil { // for other errors, handle error
			return err
		}
	}

	val, err := k.validateValAddrForDelegate(ctx, valAddr)
	if err != nil {
		return err
	}

	// delegate now!
	_, err = k.sk.Delegate(ctx, sender, totalAmtToStake, stakingtypes.Unbonded, val, true)
	if err != nil {
		return err
	}

	return nil
}
