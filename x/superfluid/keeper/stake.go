package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

func (k Keeper) GetSuperfluidOSMOTokens(ctx sdk.Context, denom string, amount sdk.Int) sdk.Int {
	twap := k.GetEpochOsmoEquivalentTWAP(ctx, denom)
	if twap.IsZero() {
		return sdk.ZeroInt()
	}

	decAmt := twap.Mul(amount.ToDec())
	asset := k.GetSuperfluidAsset(ctx, denom)
	return k.GetRiskAdjustedOsmoValue(ctx, asset, decAmt.RoundInt())
}

func (k Keeper) GetExpectedDelegationAmount(ctx sdk.Context, acc types.SuperfluidIntermediaryAccount) sdk.Int {
	// Get total number of Osmo this account should have delegated after refresh
	totalSuperfluidDelegation := k.lk.GetPeriodLocksAccumulation(ctx, lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         acc.Denom + stakingSuffix(acc.ValAddr),
		Duration:      time.Hour * 24 * 14, // should be a variable?
	})

	refreshedAmount := k.GetSuperfluidOSMOTokens(ctx, acc.Denom, totalSuperfluidDelegation)
	return refreshedAmount
}

func (k Keeper) RefreshIntermediaryDelegationAmounts(ctx sdk.Context) {
	accs := k.GetAllIntermediaryAccounts(ctx)
	for _, acc := range accs {
		mAddr := acc.GetAccAddress()
		bondDenom := k.sk.BondDenom(ctx)

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
			coins := sdk.NewCoins(sdk.NewCoin(bondDenom, adjustment))
			err = k.bk.MintCoins(ctx, types.ModuleName, coins)
			if err != nil {
				panic(err)
			}
			err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mAddr, coins)
			if err != nil {
				panic(err)
			}
			_, err = k.sk.Delegate(ctx, mAddr, adjustment, stakingtypes.Unbonded, validator, true)
			if err != nil {
				panic(err)
			}
		} else if currentAmount.GT(refreshedAmount) {
			//need to instantUndelegate and burn
			adjustment := currentAmount.Sub(refreshedAmount)
			adjustShares, _ := validator.SharesFromTokens(adjustment)
			if err != nil {
				panic(err)
			}
			res, err := k.sk.InstantUndelegate(ctx, mAddr, valAddress, adjustShares)
			if err != nil {
				panic(err)
			}
			err = k.bk.SendCoinsFromAccountToModule(ctx, mAddr, types.ModuleName, res)
			if err != nil {
				panic(err)
			}
			err = k.bk.BurnCoins(ctx, types.ModuleName, res)
			if err != nil {
				panic(err)
			}

		} else {
			ctx.Logger().Info("Intermediary account already has correct delegation amount? sus.")
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
	bondDenom := k.sk.BondDenom(ctx)
	amt := k.GetSuperfluidOSMOTokens(ctx, acc.Denom, amount.AmountOf(acc.Denom))
	if amt.IsZero() {
		return nil
	}

	coins := sdk.Coins{sdk.NewCoin(bondDenom, amt)}
	err = k.bk.MintCoins(ctx, types.ModuleName, coins)
	if err != nil {
		return err
	}

	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, intermediaryAccAddr, coins)
	if err != nil {
		return err
	}

	// make delegation from module account to the validator
	valAddress, err := sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		return err
	}
	validator, found := k.sk.GetValidator(ctx, valAddress)
	if !found {
		return stakingtypes.ErrNoValidatorFound
	}
	_, err = k.sk.Delegate(ctx, intermediaryAccAddr, amt, stakingtypes.Unbonded, validator, true)
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
		return types.ErrAttemptingToSuperfluidNonSuperfluidAsset
	}

	// prevent unbonding lockups to be not able to be used for superfluid staking
	if lock.IsUnlocking() {
		return types.ErrUnbondingLockupNotSupported
	}

	// length check
	params := k.GetParams(ctx)
	if lock.Duration < params.UnbondingDuration { // if less than bonding, skip
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

	params := k.GetParams(ctx)

	err = k.validateLockForSFDelegate(ctx, lock, sender)
	if err != nil {
		return err
	}
	validator, err := k.validateValAddrForSFDelegate(ctx, valAddr)
	if err != nil {
		return err
	}

	intermediaryAccAddr := k.GetLockIdIntermediaryAccountConnection(ctx, lockID)
	if !intermediaryAccAddr.Empty() {
		return types.ErrAlreadyUsedSuperfluidLockup
	}

	// A lock ID can only have one of three associated superfluid states:
	// 1) Not superfluid'd
	// 2) Superfluid bonded to a single validator
	// 3) Superfluid unbonding from a single validator.
	// If we are in case (2), ensure this is to the same validator.
	//   - TODO: CODE
	// If we are in case (3), disable this delegation.
	// We can wrap (2) and (3) into one check, by checking if we have any synthetic lockups on this ID.
	// and make it more precise later.
	suffix := unstakingSuffix(valAddr)
	_, err = k.lk.GetSyntheticLockup(ctx, lockID, suffix)
	if err == nil {
		return types.ErrUnbondingSyntheticLockupExists
	}

	// Register a synthetic lockup for superfluid staking
	suffix = stakingSuffix(valAddr)
	notUnlocking := false
	err = k.lk.CreateSyntheticLockup(ctx, lockID, suffix, params.UnbondingDuration, notUnlocking)
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
	mAddr := acc.GetAccAddress()

	// mint OSMO token based on TWAP of locked denom to denom module account
	// TODO: Figure out whats going on in next 3 code blocks
	// (1) Get superfluid osmo tokens backing this LP share
	// (2) Mint these as new osmo in minttypes.ModuleName
	// (3) If no account exists, make a new account at this addr
	// (4) send newly minted coins to this account.
	bondDenom := k.sk.BondDenom(ctx)
	amount := k.GetSuperfluidOSMOTokens(ctx, acc.Denom, lock.Coins.AmountOf(acc.Denom))
	if amount.IsZero() {
		return types.ErrOsmoEquivalentZeroNotAllowed
	}

	coins := sdk.Coins{sdk.NewCoin(bondDenom, amount)}
	err = k.bk.MintCoins(ctx, types.ModuleName, coins)
	if err != nil {
		return err
	}
	// TODO: @Dev added this hasAccount gating, think through if theres an edge case that makes it not right
	if !k.ak.HasAccount(ctx, mAddr) {
		// TODO: Why is this a base account, not a module account?
		k.ak.SetAccount(ctx, authtypes.NewBaseAccount(mAddr, nil, 0, 0))
	}
	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, mAddr, coins)
	if err != nil {
		return err
	}

	// make delegation from module account to the validator
	// TODO: What happens here if validator is jailed, tombstoned, or unbonding
	_, err = k.sk.Delegate(ctx, mAddr, amount, stakingtypes.Unbonded, validator, true)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
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

	shares, err := k.sk.ValidateUnbondAmount(
		ctx, intermediaryAcc.GetAccAddress(), valAddr, amount,
	)
	if err != nil {
		return err
	}

	res, err := k.sk.InstantUndelegate(ctx, intermediaryAcc.GetAccAddress(), valAddr, shares)
	if err != nil {
		return err
	}
	err = k.bk.SendCoinsFromAccountToModule(ctx, intermediaryAcc.GetAccAddress(), types.ModuleName, res)
	if err != nil {
		return err
	}
	err = k.bk.BurnCoins(ctx, types.ModuleName, res)
	if err != nil {
		return err
	}

	params := k.GetParams(ctx)
	suffix = unstakingSuffix(intermediaryAcc.ValAddr)

	// Note: bonding synthetic lockup amount is always same as native lockup amount in current implementation.
	// If there's the case, it's different, we should create synthetic lockup at deleted bonding
	// synthetic lockup amount
	err = k.lk.CreateSyntheticLockup(ctx, lockID, suffix, params.UnbondingDuration, true)
	if err != nil {
		return err
	}

	k.DeleteLockIdIntermediaryAccountConnection(ctx, lockID)
	return nil
}

// func (k Keeper) SuperfluidRedelegate(ctx sdk.Context, sender string, lockID uint64, newValAddr string) error {
// 	// Note: we prevent circular redelegations since when unbonding lockup is available from a specific validator,
// 	// not able to redelegate or undelegate again, especially the case for automatic undelegation when native lockup unlock

// 	valAddr, err := k.SuperfluidUndelegate(ctx, sender, lockID)
// 	if err != nil {
// 		return err
// 	}

// 	if valAddr.String() == newValAddr {
// 		return types.ErrSameValidatorRedelegation
// 	}

// 	return k.SuperfluidDelegate(ctx, sender, lockID, newValAddr)
// }

// TODO: Need to (eventually) override the existing staking messages and queries, for undelegating, delegating, rewards, and redelegating, to all be going through all superfluid module.
// Want integrators to be able to use the same staking queries and messages
// Eugen’s point: Only rewards message needs to be updated. Rest of messages are fine
// Queries need to be updated
// We can do this at the very end though, since it just relates to queries.
