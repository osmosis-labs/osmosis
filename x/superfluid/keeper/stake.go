package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

func stakingSuffix(valAddr string) string {
	return fmt.Sprintf("superbonding%s", valAddr)
}

func unstakingSuffix(valAddr string) string {
	return fmt.Sprintf("superunbonding%s", valAddr)
}

func SyntheticDenom(denom, valAddr string) string {
	return denom + stakingSuffix(valAddr)
}

// quick fix for getting the validator addresss from a synthetic denom
func ValidatorAddressFromSuffix(suffix string) (string, error) {
	if strings.Contains(suffix, "superbonding") {
		return strings.TrimLeft(suffix, "superbonding"), nil
	}
	if strings.Contains(suffix, "superunbonding") {
		return strings.TrimLeft(suffix, "superunbonding"), nil
	}
	return "", fmt.Errorf("%s is not a valid synthetic denom suffix", suffix)
}

func (k Keeper) GetTotalSyntheticAssetsLocked(ctx sdk.Context, denom string) sdk.Int {
	return k.lk.GetPeriodLocksAccumulation(ctx, lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         denom,
		Duration:      k.sk.UnbondingTime(ctx),
	})
}

func (k Keeper) GetExpectedDelegationAmount(ctx sdk.Context, acc types.SuperfluidIntermediaryAccount) sdk.Int {
	// Get total number of Osmo this account should have delegated after refresh
	totalSuperfluidDelegation := k.GetTotalSyntheticAssetsLocked(ctx, acc.Denom+stakingSuffix(acc.ValAddr))
	refreshedAmount := k.GetSuperfluidOSMOTokens(ctx, acc.Denom, totalSuperfluidDelegation)
	return refreshedAmount
}

func (k Keeper) RefreshIntermediaryDelegationAmounts(ctx sdk.Context) {
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

		delegation, found := k.sk.GetDelegation(ctx, mAddr, valAddress)
		if !found {
			k.Logger(ctx).Error(fmt.Sprintf("delegation not found for %s with %s", mAddr.String(), acc.ValAddr))
			continue
		}

		currentAmount := validator.TokensFromShares(delegation.Shares).RoundInt()

		refreshedAmount := k.GetExpectedDelegationAmount(ctx, acc)

		if refreshedAmount.GT(currentAmount) {
			//need to mint and delegate
			adjustment := refreshedAmount.Sub(currentAmount)
			err = k.mintOsmoTokensAndDelegate(ctx, adjustment, acc, validator)
			if err != nil {
				panic(err)
			}
		} else if currentAmount.GT(refreshedAmount) {
			// In this case, we want to change the IA's delegated balance to be refreshed Amount
			// which is less than what it already has.
			// This means we need to "InstantUndelegate" some of its delegation (not going through the unbonding queue)
			// and then burn that excessly delegated bits.
			adjustment := currentAmount.Sub(refreshedAmount)

			err := k.forceUndelegateAndBurnOsmoTokens(ctx, adjustment, acc, valAddress)
			if err != nil {
				// TODO: We can't panic here. We can err-wrap though.
				panic(err)
			}
		} else {
			ctx.Logger().Info("Intermediary account already has correct delegation amount? sus. This whp implies the exact same spot price as the last epoch, and no delegation changes.")
		}
	}
}

func (k Keeper) SuperfluidDelegateMore(ctx sdk.Context, lockID uint64, amount sdk.Coins) error {
	intermediaryAccAddr := k.GetLockIdIntermediaryAccountConnection(ctx, lockID)
	if intermediaryAccAddr.Empty() {
		return nil
	}

	acc := k.GetIntermediaryAccount(ctx, intermediaryAccAddr)
	valAddr := acc.ValAddr
	validator, err := k.validateValAddrForSFDelegate(ctx, valAddr)
	if err != nil {
		return err
	}

	suffix := stakingSuffix(valAddr)
	synthLock, err := k.lk.GetSyntheticLockup(ctx, lockID, suffix)
	if err != nil {
		return err
	}
	// TODO: Add safety checks?
	err = k.lk.AddTokensToSyntheticLock(ctx, *synthLock, amount)
	if err != nil {
		return err
	}

	// mint OSMO token based on TWAP of locked denom to denom module account
	amt := k.GetSuperfluidOSMOTokens(ctx, acc.Denom, amount.AmountOf(acc.Denom))
	if amt.IsZero() {
		return nil
	}

	err = k.mintOsmoTokensAndDelegate(ctx, amt, acc, validator)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) validateLockForSFDelegate(ctx sdk.Context, lock *lockuptypes.PeriodLock, sender string) error {
	if lock.Owner != sender {
		return lockuptypes.ErrNotLockOwner
	}

	if lock.Coins.Len() != 1 {
		return types.ErrMultipleCoinsLockupNotSupported
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

	return nil
}

// ensure the valAddr is correctly formatted & corresponds to a real validator on chain.
func (k Keeper) validateValAddrForSFDelegate(ctx sdk.Context, valAddr string) (stakingtypes.Validator, error) {
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

// func (k Keeper) hasBondedSuperfluidDelegation(ctx sdk.Context, lockID int64) bool {
// 	valAddress, err := sdk.ValAddressFromBech32(valAddr)
// 	if err != nil {
// 		return stakingtypes.Validator{}, err
// 	}
// 	validator, found := k.sk.GetValidator(ctx, valAddress)
// 	if !found {
// 		return stakingtypes.Validator{}, stakingtypes.ErrNoValidatorFound
// 	}
// 	return validator, nil
// }

// TODO: Merge a lot of logic with SuperfluidDelegateMore
func (k Keeper) SuperfluidDelegate(ctx sdk.Context, sender string, lockID uint64, valAddr string) error {
	lock, err := k.lk.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	err = k.validateLockForSFDelegate(ctx, lock, sender)
	if err != nil {
		return err
	}
	validator, err := k.validateValAddrForSFDelegate(ctx, valAddr)
	if err != nil {
		return err
	}

	// This guarantees this lockID does not already have a superfluid stake position
	// associated with it.
	// Thus when we stake now, this will be the only superfluid position for this lockID.
	if k.alreadySuperfluidStaking(ctx, lockID) {
		return types.ErrAlreadyUsedSuperfluidLockup
	}

	// Register a synthetic lockup for superfluid staking
	suffix := stakingSuffix(valAddr)
	notUnlocking := false
	unbondingDuration := k.sk.GetParams(ctx).UnbondingTime
	err = k.lk.CreateSyntheticLockup(ctx, lockID, suffix, unbondingDuration, notUnlocking)
	if err != nil {
		return err
	}

	// get the intermediate account for this (denom, validator) pair.
	// This account tracks the amount of osmo being considered as staked.
	// If an intermediary account doesn't exist, then create it + a perpetual gauge.
	acc, err := k.GetOrCreateIntermediaryAccount(ctx, lock.Coins[0].Denom, valAddr)
	if err != nil {
		return err
	}

	// Find how many new osmo tokens this delegation is worth at superfluids current risk adjustment
	// and twap of the denom.
	amount := k.GetSuperfluidOSMOTokens(ctx, acc.Denom, lock.Coins.AmountOf(acc.Denom))
	if amount.IsZero() {
		return types.ErrOsmoEquivalentZeroNotAllowed
	}

	err = k.mintOsmoTokensAndDelegate(ctx, amount, acc, validator)
	if err != nil {
		return err
	}

	// create connection record between lock id and intermediary account
	k.SetLockIdIntermediaryAccountConnection(ctx, lockID, acc)
	return nil
}

func (k Keeper) SuperfluidUndelegate(ctx sdk.Context, sender string, lockID uint64) error {
	lock, err := k.lk.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	if lock.Owner != sender {
		return lockuptypes.ErrNotLockOwner
	}

	// Remove previously created synthetic lockup
	intermediaryAccAddr := k.GetLockIdIntermediaryAccountConnection(ctx, lockID)
	if intermediaryAccAddr.Empty() {
		return types.ErrNotSuperfluidUsedLockup
	}
	intermediaryAcc := k.GetIntermediaryAccount(ctx, intermediaryAccAddr)
	suffix := stakingSuffix(intermediaryAcc.ValAddr)

	synthLock, err := k.lk.GetSyntheticLockup(ctx, lockID, suffix)
	if err != nil {
		return err
	}

	if synthLock.Owner != sender {
		return lockuptypes.ErrNotLockOwner
	}

	err = k.lk.DeleteSyntheticLockup(ctx, lockID, suffix)
	if err != nil {
		return err
	}

	// use synthetic lockup coins for unbonding
	amount := k.GetSuperfluidOSMOTokens(ctx, intermediaryAcc.Denom, synthLock.Coins.AmountOf(intermediaryAcc.Denom+suffix))

	valAddr, err := sdk.ValAddressFromBech32(intermediaryAcc.ValAddr)
	if err != nil {
		return err
	}

	err = k.forceUndelegateAndBurnOsmoTokens(ctx, amount, intermediaryAcc, valAddr)
	if err != nil {
		return err
	}

	unbondingDuration := k.sk.GetParams(ctx).UnbondingTime
	suffix = unstakingSuffix(intermediaryAcc.ValAddr)

	// Note: bonding synthetic lockup amount is always same as native lockup amount in current implementation.
	// If there's the case, it's different, we should create synthetic lockup at deleted bonding
	// synthetic lockup amount
	err = k.lk.CreateSyntheticLockup(ctx, lockID, suffix, unbondingDuration, true)
	if err != nil {
		return err
	}

	k.DeleteLockIdIntermediaryAccountConnection(ctx, lockID)
	return nil
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

// mint osmoAmount of OSMO tokens, and immediately delegate them to validator on behalf of intermediary account
func (k Keeper) mintOsmoTokensAndDelegate(ctx sdk.Context, osmoAmount sdk.Int, intermediaryAccount types.SuperfluidIntermediaryAccount, validator stakingtypes.Validator) error {
	bondDenom := k.sk.BondDenom(ctx)
	coins := sdk.Coins{sdk.NewCoin(bondDenom, osmoAmount)}
	err := k.bk.MintCoins(ctx, types.ModuleName, coins)
	if err != nil {
		return err
	}
	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, intermediaryAccount.GetAccAddress(), coins)
	if err != nil {
		return err
	}

	// make delegation from module account to the validator
	// TODO: What happens here if validator is jailed, tombstoned, or unbonding
	_, err = k.sk.Delegate(ctx,
		intermediaryAccount.GetAccAddress(),
		osmoAmount, stakingtypes.Unbonded, validator, true)
	return err
}

// force undelegate osmoAmount worth of delegation shares from delegations between intermediary account and valAddr
// We take the returned tokens, and then immediately burn them.
func (k Keeper) forceUndelegateAndBurnOsmoTokens(ctx sdk.Context,
	osmoAmount sdk.Int, intermediaryAcc types.SuperfluidIntermediaryAccount, valAddr sdk.ValAddress) error {
	// TODO: Better understand and decide between ValidateUnbondAmount and SharesFromTokens
	// briefly looked into it, did not understand whats correct.
	// TODO: ensure that intermediate account has at least osmoAmount staked.
	shares, err := k.sk.ValidateUnbondAmount(
		ctx, intermediaryAcc.GetAccAddress(), valAddr, osmoAmount,
	)
	if err != nil {
		return err
	}

	undelegatedCoins, err := k.sk.InstantUndelegate(ctx, intermediaryAcc.GetAccAddress(), valAddr, shares)
	if err != nil {
		return err
	}
	err = k.bk.SendCoinsFromAccountToModule(ctx, intermediaryAcc.GetAccAddress(), types.ModuleName, undelegatedCoins)
	if err != nil {
		return err
	}
	err = k.bk.BurnCoins(ctx, types.ModuleName, undelegatedCoins)
	return err
}

// TODO: Need to (eventually) override the existing staking messages and queries, for undelegating, delegating, rewards, and redelegating, to all be going through all superfluid module.
// Want integrators to be able to use the same staking queries and messages
// Eugen’s point: Only rewards message needs to be updated. Rest of messages are fine
// Queries need to be updated
// We can do this at the very end though, since it just relates to queries.
