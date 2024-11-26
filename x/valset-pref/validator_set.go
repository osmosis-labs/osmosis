package keeper

import (
	"errors"
	"fmt"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	appParams "github.com/osmosis-labs/osmosis/v27/app/params"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v27/x/valset-pref/types"
)

type valSet struct {
	ValAddr string
	Amount  osmomath.Dec
}

type ValRatio struct {
	ValAddr       sdk.ValAddress
	Weight        osmomath.Dec
	DelegatedAmt  osmomath.Int
	UndelegateAmt osmomath.Int
	VRatio        osmomath.Dec
}

// SetValidatorSetPreferences sets a new valset position for a delegator in modules state.
func (k Keeper) SetValidatorSetPreferences(ctx sdk.Context, delegator string, validators types.ValidatorSetPreferences) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, []byte(delegator), &validators)
}

// GetValidatorSetPreference returns the existing valset position for a delegator.
func (k Keeper) GetValidatorSetPreference(ctx sdk.Context, delegator string) (types.ValidatorSetPreferences, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(delegator))
	if bz == nil {
		return types.ValidatorSetPreferences{}, false
	}

	// valset delegation exists, so return it
	var valsetPref types.ValidatorSetPreferences
	if err := proto.Unmarshal(bz, &valsetPref); err != nil {
		return types.ValidatorSetPreferences{}, false
	}

	return valsetPref, true
}

// ValidateValidatorSetPreference derives given validator set.
// It validates the list and formats the inputs such as rounding.
// Errors when the given preference is the same as the existing preference in state.
// NOTE: this function does not add valset to the state
func (k Keeper) ValidateValidatorSetPreference(ctx sdk.Context, delegator string, preferences []types.ValidatorPreference) (types.ValidatorSetPreferences, error) {
	existingValSet, found := k.GetValidatorSetPreference(ctx, delegator)
	if found {
		// check if the new preferences is the same as the existing preferences
		isEqual := k.IsValidatorSetEqual(existingValSet.Preferences, preferences)
		if isEqual {
			return types.ValidatorSetPreferences{}, errors.New("The preferences (validator and weights) are the same")
		}
	}

	// checks that all the validators exist on chain
	valSetPref, err := k.IsPreferenceValid(ctx, preferences)
	if err != nil {
		return types.ValidatorSetPreferences{}, errors.New("The validator preference list is not valid")
	}

	return types.ValidatorSetPreferences{Preferences: valSetPref}, nil
}

// DelegateToValidatorSet delegates to a delegators existing validator-set.
// If the valset does not exist, it delegates to existing staking position.
// For ex: delegate 10osmo with validator-set {ValA -> 0.5, ValB -> 0.3, ValC -> 0.2}
// our delegate logic would attempt to delegate 5osmo to A , 2osmo to B, 3osmo to C
// nolint: staticcheck
func (k Keeper) DelegateToValidatorSet(ctx sdk.Context, delegatorAddr string, coin sdk.Coin) error {
	// get valset formatted delegation either from existing val set preference or existing delegations
	existingSet, err := k.GetDelegationPreferences(ctx, delegatorAddr)
	if err != nil {
		return err
	}

	delegator, err := sdk.AccAddressFromBech32(delegatorAddr)
	if err != nil {
		return err
	}

	// totalDelAmt is the amount that keeps running track of the amount of tokens delegated
	totalDelAmt := osmomath.NewInt(0)
	tokenAmt := osmomath.NewInt(0)

	// loop through the validatorSetPreference and delegate the proportion of the tokens based on weights
	for i, val := range existingSet.Preferences {
		_, validator, err := k.getValAddrAndVal(ctx, val.ValOperAddress)
		if err != nil {
			return err
		}

		// in the last valset iteration we dont calculate it from shares using decimals and truncation,
		// we use what's remaining to get more accurate value
		if len(existingSet.Preferences)-1 == i {
			tokenAmt = coin.Amount.Sub(totalDelAmt).ToLegacyDec().TruncateInt()
		} else {
			// tokenAmt takes the amount to delegate, calculated by {val_distribution_weight * tokenAmt}
			tokenAmt = val.Weight.MulInt(coin.Amount).TruncateInt()
			totalDelAmt = totalDelAmt.Add(tokenAmt)
		}
		// TODO: What happens here if validator unbonding
		// Delegate the unbonded tokens
		_, err = k.stakingKeeper.Delegate(ctx, delegator, tokenAmt, stakingtypes.Unbonded, validator, true)
		if err != nil {
			return err
		}
	}

	return nil
}

// UndelegateFromValidatorSet undelegates {coin} amount from the validator set.
// If the valset does not exist, it undelegates from existing staking position.
// Ex: user A has staked 10tokens with weight {Val->0.5, ValB->0.3, ValC->0.2}
// undelegate 6osmo with validator-set {ValA -> 0.5, ValB -> 0.3, ValC -> 0.2}
// our undelegate logic would attempt to undelegate 3osmo from A, 1.8osmo from B, 1.2osmo from C
// Truncation ensures we do not undelegate more than the user has staked with the validator set.
// NOTE: check README.md for more verbose description of the algorithm.
// TODO: This is currently disabled.
// Properly implement for vratio > 1 to hit steps 5-7, then re-enable
// https://github.com/osmosis-labs/osmosis/issues/6686
func (k Keeper) UndelegateFromValidatorSet(ctx sdk.Context, delegatorAddr string, undelegation sdk.Coin) error {
	// TODO: Change to GetDelegationPreferences
	existingSet, err := k.GetValSetPreferencesWithDelegations(ctx, delegatorAddr)
	if err != nil {
		return types.NoValidatorSetOrExistingDelegationsError{DelegatorAddr: delegatorAddr}
	}

	delegator := sdk.MustAccAddressFromBech32(delegatorAddr)

	// Step 1,2: compute the total amount delegated and the amount to undelegate for each validator
	// under valset-ratios.
	valSetRatio, validators, totalDelegatedAmt, err := k.getValsetRatios(ctx, delegator, existingSet.Preferences, undelegation.Amount)
	if err != nil {
		return err
	}

	if undelegation.Amount.ToLegacyDec().GT(totalDelegatedAmt) {
		return types.UndelegateMoreThanDelegatedError{TotalDelegatedAmt: totalDelegatedAmt, UndelegationAmt: undelegation.Amount}
	}

	// Step 3: Sort validators in descending order of VRatio.
	sort.Slice(valSetRatio, func(i, j int) bool {
		return valSetRatio[i].VRatio.GT(valSetRatio[j].VRatio)
	})

	totalUnDelAmt := osmomath.NewInt(0)
	var amountToUnDelegate osmomath.Int
	// Step 4: if largest V Ratio is under 1, happy path, simply
	// undelegate target amount from each validator
	if valSetRatio[0].VRatio.LTE(osmomath.OneDec()) {
		for index, val := range valSetRatio {
			validator := validators[val.ValAddr.String()]

			// in the last valset iteration we don't calculate it from shares using decimals and truncation,
			// we use what's remaining to get more accurate value
			if len(existingSet.Preferences)-1 == index {
				amountToUnDelegate = undelegation.Amount.Sub(totalUnDelAmt).ToLegacyDec().TruncateInt()
			} else {
				// Calculate the amount to undelegate based on the existing weightxs
				amountToUnDelegate = val.UndelegateAmt
				totalUnDelAmt = totalUnDelAmt.Add(amountToUnDelegate)
			}
			sharesAmt, err := validator.SharesFromTokens(amountToUnDelegate)
			if err != nil {
				return err
			}

			_, _, err = k.stakingKeeper.Undelegate(ctx, delegator, val.ValAddr, sharesAmt) // this has to be shares amount
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Step 5
	// `targetRatio`: This is a threshold value that is used to decide how to unbond tokens from validators.
	// It starts as 1 and is recalculated each time a validator is fully unbonded and removed from the unbonding process.
	// By reducing the target ratio using the ratio of the removed validator, we adjust the proportions we are aiming for with the remaining validators.
	targetRatio := osmomath.OneDec()
	amountRemaining := undelegation.Amount

	// Step 6
	for len(valSetRatio) > 0 && valSetRatio[0].VRatio.GT(targetRatio) {
		_, _, err = k.stakingKeeper.Undelegate(ctx, delegator, valSetRatio[0].ValAddr, valSetRatio[0].DelegatedAmt.ToLegacyDec()) // this has to be shares amount
		if err != nil {
			return err
		}
		amountRemaining = amountRemaining.Sub(valSetRatio[0].DelegatedAmt)
		targetRatio = targetRatio.Mul(osmomath.OneDec().Sub(valSetRatio[0].Weight))
		valSetRatio = valSetRatio[1:]
	}

	// Step 7
	for _, val := range valSetRatio {
		_, validator, err := k.getValAddrAndVal(ctx, val.ValAddr.String())
		if err != nil {
			return err
		}

		sharesAmt, err := validator.SharesFromTokens(val.UndelegateAmt)
		if err != nil {
			return err
		}

		_, _, err = k.stakingKeeper.Undelegate(ctx, delegator, val.ValAddr, sharesAmt) // this has to be shares amount
		if err != nil {
			return err
		}
	}

	return nil
}

// UndelegateFromRebalancedValidatorSet undelegates a specified amount of tokens from a delegator's existing validator set,
// but takes into consideration the user's existing delegations to the validators in the set.
// The method first fetches the delegator's validator set preferences, checks their existing delegations, and
// returns a set with modified weights that consider their existing delegations.
// If there is no existing delegation, it returns an error.
// The method then computes the total amount delegated and the amount to undelegate for each validator under this
// newly calculated valset-ratio set.
//
// If the undelegation amount is greater than the total delegated amount, it returns an error.
// The validators are then sorted in descending order of VRatio.
// The method ensures that the largest VRatio is under 1. If it is greater than 1, it returns an error.
// Finally, the method undelegates the target amount from each validator.
// If an error occurs during the undelegation process, it is returned.
func (k Keeper) UndelegateFromRebalancedValidatorSet(ctx sdk.Context, delegatorAddr string, undelegation sdk.Coin) error {
	// GetValSetPreferencesWithDelegations fetches the delegator's validator set preferences, but returns a set with
	// modified weights that consider their existing delegations. If there is no existing delegation, it returns an error.
	// The new weights based on the existing delegations is returned, but the original valset preferences
	// are not modified.
	// For example, if someone's valset is 50/50 between two validators, but they have 10 OSMO delegated to validator A,
	// and 90 OSMO delegated to validator B, the returned valset preference weight will be 10/90.

	existingSet, err := k.GetValSetPreferencesWithDelegations(ctx, delegatorAddr)
	if err != nil {
		return types.NoValidatorSetOrExistingDelegationsError{DelegatorAddr: delegatorAddr}
	}

	delegator := sdk.MustAccAddressFromBech32(delegatorAddr)

	// Step 1,2: compute the total amount delegated and the amount to undelegate for each validator
	// under valset-ratios.
	valSetRatio, validators, totalDelegatedAmt, err := k.getValsetRatios(ctx, delegator, existingSet.Preferences, undelegation.Amount)
	if err != nil {
		return err
	}

	if undelegation.Amount.ToLegacyDec().GT(totalDelegatedAmt) {
		return types.UndelegateMoreThanDelegatedError{TotalDelegatedAmt: totalDelegatedAmt, UndelegationAmt: undelegation.Amount}
	}

	// Step 3: Sort validators in descending order of VRatio.
	sort.Slice(valSetRatio, func(i, j int) bool {
		return valSetRatio[i].VRatio.GT(valSetRatio[j].VRatio)
	})

	totalUnDelAmt := osmomath.NewInt(0)
	var amountToUnDelegate osmomath.Int

	// Ensure largest VRatio is under 1.
	// Since we called GetValSetPreferencesWithDelegations, there should be no VRatio > 1
	if valSetRatio[0].VRatio.GT(osmomath.OneDec()) {
		return types.ValsetRatioGreaterThanOneError{ValsetRatio: valSetRatio[0].VRatio}
	}

	// Step 4: Undelegate target amount from each validator
	for index, val := range valSetRatio {
		validator := validators[val.ValAddr.String()]

		// in the last valset iteration we don't calculate it from shares using decimals and truncation,
		// we use what's remaining to get more accurate value
		if len(existingSet.Preferences)-1 == index {
			// Directly retrieve the delegation to the last validator
			// Use the min between our undelegation amount calculated via iterations of undelegating
			// and the amount actually delegated to the validator. This is done to prevent an error
			// in the event some rounding issue increases our calculated undelegation amount.
			delegation, err := k.stakingKeeper.GetDelegation(ctx, delegator, val.ValAddr)
			if err != nil {
				return err
			}
			delegationToVal := delegation.Shares.TruncateInt()
			calculatedUndelegationAmt := undelegation.Amount.Sub(totalUnDelAmt).ToLegacyDec().TruncateInt()
			amountToUnDelegate = osmomath.MinInt(delegationToVal, calculatedUndelegationAmt)
		} else {
			// Calculate the amount to undelegate based on the existing weightxs
			amountToUnDelegate = val.UndelegateAmt
			totalUnDelAmt = totalUnDelAmt.Add(amountToUnDelegate)
		}
		sharesAmt, err := validator.SharesFromTokens(amountToUnDelegate)
		if err != nil {
			return err
		}

		_, _, err = k.stakingKeeper.Undelegate(ctx, delegator, val.ValAddr, sharesAmt) // this has to be shares amount
		if err != nil {
			return err
		}
	}
	return nil
}

// getValsetRatios returns the valRatio array calculated based on the given delegator, valset prefs, and undelegating amount.
// Errors when given delegator does not have delegation towards all of the validators given in the valsetPrefs
func (k Keeper) getValsetRatios(ctx sdk.Context, delegator sdk.AccAddress,
	prefs []types.ValidatorPreference, undelegateAmt osmomath.Int) ([]ValRatio, map[string]stakingtypes.Validator, osmomath.Dec, error) {
	// total amount user has delegated
	totalDelegatedAmt := osmomath.ZeroDec()
	var valSetRatios []ValRatio
	validators := map[string]stakingtypes.Validator{}

	for _, val := range prefs {
		amountToUnDelegate := val.Weight.MulInt(undelegateAmt).TruncateInt()
		valAddr, validator, err := k.getValAddrAndVal(ctx, val.ValOperAddress)
		if err != nil {
			return nil, map[string]stakingtypes.Validator{}, osmomath.ZeroDec(), err
		}
		validators[valAddr.String()] = validator

		delegation, err := k.stakingKeeper.GetDelegation(ctx, delegator, valAddr)
		if err != nil {
			return nil, map[string]stakingtypes.Validator{}, osmomath.ZeroDec(), err
		}

		undelegateSharesAmt, err := validator.SharesFromTokens(amountToUnDelegate)
		if err != nil {
			return nil, map[string]stakingtypes.Validator{}, osmomath.ZeroDec(), err
		}

		// vRatio = undelegating amount / total delegated shares
		// vRatio equals to 1 when undelegating full amount
		vRatio := undelegateSharesAmt.Quo(delegation.Shares)
		totalDelegatedAmt = totalDelegatedAmt.AddMut(validator.TokensFromShares(delegation.Shares))
		valSetRatios = append(valSetRatios, ValRatio{
			ValAddr:       valAddr,
			UndelegateAmt: amountToUnDelegate,
			VRatio:        vRatio,
			DelegatedAmt:  delegation.Shares.TruncateInt(),
			Weight:        val.Weight,
		})
	}
	return valSetRatios, validators, totalDelegatedAmt, nil
}

// The redelegation command allows delegators to instantly switch validators.
// Once the unbonding period has passed, the redelegation is automatically completed in the EndBlocker.
// A redelegation object is created every time a redelegation occurs. To prevent "redelegation hopping" where delegatorA can redelegate
// between many validators over small period of time, redelegations may not occur under the following situation:
// 1. delegatorA attempts to redelegate to the same validator
//   - valA --redelegate--> valB
//   - valB --redelegate--> valB (ERROR: Self redelegation is not allowed)
//
// 2. delegatorA attempts to redelegate to an immature redelegation validator
//   - valA --redelegate--> valB
//   - valB --redelegate--> valA	(ERROR: Redelegation to ValB is already in progress)
//
// 3. delegatorA attempts to redelegate while unbonding is in progress
//   - unbond (10osmo) from valA
//   - valA --redelegate--> valB (ERROR: new redelegation while unbonding is in progress)
func (k Keeper) PreformRedelegation(ctx sdk.Context, delegator sdk.AccAddress, existingSet []types.ValidatorPreference, newSet []types.ValidatorPreference) error {
	var existingValSet []valSet
	var newValSet []valSet
	totalTokenAmount := osmomath.NewDec(0)

	// Rearranging the exisingValSet and newValSet to to add extra validator padding
	for _, existingVals := range existingSet {
		valAddr, validator, err := k.GetValidatorInfo(ctx, existingVals.ValOperAddress)
		if err != nil {
			return err
		}

		// check if the user has delegated tokens to the valset
		delegation, err := k.stakingKeeper.GetDelegation(ctx, delegator, valAddr)
		if err != nil {
			return err
		}

		tokenFromShares := validator.TokensFromShares(delegation.Shares)
		existing_val, existing_val_zero_amount := k.GetValSetStruct(existingVals, tokenFromShares)
		existingValSet = append(existingValSet, existing_val)
		newValSet = append(newValSet, existing_val_zero_amount)
		totalTokenAmount = totalTokenAmount.Add(tokenFromShares)
	}

	for _, newVals := range newSet {
		amountToDelegate := newVals.Weight.Mul(totalTokenAmount)

		new_val, new_val_zero_amount := k.GetValSetStruct(newVals, amountToDelegate)
		newValSet = append(newValSet, new_val)
		existingValSet = append(existingValSet, new_val_zero_amount)
	}

	// calculate the difference between two sets
	var diffValSets []*valSet
	for i, newVals := range existingValSet {
		diffAmount := newVals.Amount.Sub(newValSet[i].Amount)

		diff_val := valSet{
			ValAddr: newVals.ValAddr,
			Amount:  diffAmount,
		}
		diffValSets = append(diffValSets, &diff_val)
	}

	// Algorithm starts here, verbose explanation in README.md
	for _, diffVal := range diffValSets {
		if diffVal.Amount.TruncateDec().IsPositive() {
			for idx, targetDiffVal := range diffValSets {
				if targetDiffVal.Amount.TruncateDec().IsNegative() && diffVal.ValAddr != targetDiffVal.ValAddr {
					valSource, valTarget, err := k.getValTargetAndSource(ctx, diffVal.ValAddr, targetDiffVal.ValAddr)
					if err != nil {
						return err
					}

					transferAmount := osmomath.MinDec(diffVal.Amount, targetDiffVal.Amount.Abs()).TruncateDec()
					if transferAmount.IsZero() {
						break
					}

					_, err = k.stakingKeeper.BeginRedelegation(ctx, delegator, valSource, valTarget, transferAmount)
					if err != nil {
						return err
					}

					diffVal.Amount = diffVal.Amount.Sub(transferAmount)
					diffValSets[idx].Amount = targetDiffVal.Amount.Add(transferAmount)

					if diffVal.Amount.IsZero() {
						break
					}
				}
			}
		}
	}

	return nil
}

// getValTargetAndSource formats the validator address and returns sdk.ValAddress formatted value.
func (k Keeper) getValTargetAndSource(ctx sdk.Context, valSource, valTarget string) (sdk.ValAddress, sdk.ValAddress, error) {
	validatorSource, _, err := k.GetValidatorInfo(ctx, valSource)
	if err != nil {
		return nil, nil, err
	}

	validatorTarget, _, err := k.GetValidatorInfo(ctx, valTarget)
	if err != nil {
		return nil, nil, err
	}

	return validatorSource, validatorTarget, nil
}

// WithdrawDelegationRewards withdraws all the delegation rewards from all validators the user is delegated to, disregarding the val-set.
// If the valset does not exist, it withdraws from existing staking position.
// Delegation reward is collected by the validator and in doing so, they can charge commission to the delegators.
// Rewards are calculated per period, and is updated each time validator delegation changes. For ex: when a delegator
// receives new delegation the rewards can be calculated by taking (total rewards before new delegation - the total current rewards).
func (k Keeper) WithdrawDelegationRewards(ctx sdk.Context, delegatorAddr string) error {
	// Get all validators the user is delegated to, and create a set from it.
	existingSet, err := k.GetValSetPreferencesWithDelegations(ctx, delegatorAddr)
	if err != nil {
		return types.NoValidatorSetOrExistingDelegationsError{DelegatorAddr: delegatorAddr}
	}

	delegator, err := sdk.AccAddressFromBech32(delegatorAddr)
	if err != nil {
		return err
	}

	err = k.withdrawExistingValSetStakingPosition(ctx, delegator, existingSet.Preferences)
	if err != nil {
		return err
	}

	return nil
}

// withdrawExistingValSetStakingPosition takes the existing valset delegator delegations and withdraws the rewards.
func (k Keeper) withdrawExistingValSetStakingPosition(ctx sdk.Context, delegator sdk.AccAddress, delegations []types.ValidatorPreference) error {
	for _, dels := range delegations {
		valAddr, err := sdk.ValAddressFromBech32(dels.ValOperAddress)
		if err != nil {
			return errors.New("validator address not formatted")
		}

		_, err = k.distirbutionKeeper.WithdrawDelegationRewards(ctx, delegator, valAddr)
		if err != nil {
			return err
		}
	}
	return nil
}

// ForceUnlockBondedOsmo allows breaking of a bonded lockup (by ID) of osmo, of length <= 2 weeks.
// We want to later have osmo incentives get auto-staked, we want people w/ no staking positions to
// get their osmo auto-locked. This function takes all that osmo and stakes according to your
// current validator set preference.
// (Note: Noting that there is an implicit valset preference if you've already staked)
// CONTRACT: This method should **never** be used alone.
func (k Keeper) ForceUnlockBondedOsmo(ctx sdk.Context, lockID uint64, delegatorAddr string) (sdk.Coin, error) {
	lock, lockedOsmoAmount, err := k.validateLockForForceUnlock(ctx, lockID, delegatorAddr)
	if err != nil {
		return sdk.Coin{}, err
	}

	// Ensured the lock has no superfluid relation by checking that there are no synthetic locks
	synthLocks, _, err := k.lockupKeeper.GetSyntheticLockupByUnderlyingLockId(ctx, lockID)
	if err != nil {
		return sdk.Coin{}, err
	}
	// TODO: use found
	if synthLocks != (lockuptypes.SyntheticLock{}) {
		return sdk.Coin{}, errors.New("cannot use DelegateBondedTokens being used for superfluid.")
	}

	// ForceUnlock ignores lockup duration and unlock tokens immediately.
	err = k.lockupKeeper.ForceUnlock(ctx, *lock)
	if err != nil {
		return sdk.Coin{}, err
	}

	// Takes unlocked osmo, and delegate according to valset pref
	unlockedOsmoCoin := sdk.Coin{Denom: appParams.BaseCoinUnit, Amount: lockedOsmoAmount}

	return unlockedOsmoCoin, nil
}

// GetValAddrAndVal checks if the validator address is valid and the validator provided exists on chain.
func (k Keeper) getValAddrAndVal(ctx sdk.Context, valOperAddress string) (sdk.ValAddress, stakingtypes.Validator, error) {
	valAddr, err := sdk.ValAddressFromBech32(valOperAddress)
	if err != nil {
		return nil, stakingtypes.Validator{}, errors.New("validator address not formatted")
	}

	validator, err := k.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, stakingtypes.Validator{}, fmt.Errorf("validator not found %s", validator.String())
	}

	return valAddr, validator, nil
}

// IsPreferenceValid loops through the validator preferences and checks its existence and validity.
func (k Keeper) IsPreferenceValid(ctx sdk.Context, preferences []types.ValidatorPreference) ([]types.ValidatorPreference, error) {
	var weightsRoundedValPrefList []types.ValidatorPreference
	for _, val := range preferences {
		// round up weights
		valWeightStr := osmomath.SigFigRound(val.Weight, osmomath.NewDec(10).Power(2).TruncateInt())

		_, _, err := k.GetValidatorInfo(ctx, val.ValOperAddress)
		if err != nil {
			return nil, err
		}

		weightsRoundedValPrefList = append(weightsRoundedValPrefList, types.ValidatorPreference{
			ValOperAddress: val.ValOperAddress,
			Weight:         valWeightStr,
		})
	}

	return weightsRoundedValPrefList, nil
}

// IsValidatorSetEqual returns true if the two preferences are equal.
func (k Keeper) IsValidatorSetEqual(newPreferences, existingPreferences []types.ValidatorPreference) bool {
	var isEqual bool
	// check if the two validator-set length are equal
	if len(newPreferences) != len(existingPreferences) {
		return false
	}

	// sort the new validator-set
	sort.Slice(newPreferences, func(i, j int) bool {
		return newPreferences[i].ValOperAddress < newPreferences[j].ValOperAddress
	})

	// sort the existing validator-set
	sort.Slice(existingPreferences, func(i, j int) bool {
		return existingPreferences[i].ValOperAddress < existingPreferences[j].ValOperAddress
	})

	// make sure that both valAddress and weights cannot be the same in the new val-set
	// if we just find one difference between two sets we can guarantee that they are different
	for i := range newPreferences {
		if newPreferences[i].ValOperAddress != existingPreferences[i].ValOperAddress ||
			!newPreferences[i].Weight.Equal(existingPreferences[i].Weight) {
			isEqual = false
			break
		} else {
			isEqual = true
		}
	}

	return isEqual
}

func (k Keeper) GetValidatorInfo(ctx sdk.Context, existingValAddr string) (sdk.ValAddress, stakingtypes.Validator, error) {
	valAddr, validator, err := k.getValAddrAndVal(ctx, existingValAddr)
	if err != nil {
		return nil, stakingtypes.Validator{}, err
	}
	return valAddr, validator, nil
}

// GetValSetStruct initializes valSet struct with valAddr, weight and amount.
// It also creates an extra struct with zero amount, that can be appended to newValSet that will be created.
// We do this to make sure the struct array length is the same to calculate their difference.
func (k Keeper) GetValSetStruct(validator types.ValidatorPreference, amountFromShares osmomath.Dec) (valStruct valSet, valStructZeroAmt valSet) {
	val_struct := valSet{
		ValAddr: validator.ValOperAddress,
		Amount:  amountFromShares,
	}

	val_struct_zero_amount := valSet{
		ValAddr: validator.ValOperAddress,
		Amount:  osmomath.NewDec(0),
	}

	return val_struct, val_struct_zero_amount
}

// check if lock owner matches the delegator, contains only uosmo and is bonded for <= 2weeks
func (k Keeper) validateLockForForceUnlock(ctx sdk.Context, lockID uint64, delegatorAddr string) (*lockuptypes.PeriodLock, osmomath.Int, error) {
	// Checks if sender is lock ID owner
	lock, err := k.lockupKeeper.GetLockByID(ctx, lockID)
	if err != nil {
		return nil, osmomath.Int{}, err
	}
	if lock.GetOwner() != delegatorAddr {
		return nil, osmomath.Int{}, fmt.Errorf("delegator (%s) and lock owner (%s) does not match", delegatorAddr, lock.Owner)
	}

	lockedOsmoAmount := osmomath.NewInt(0)

	// check that lock contains only 1 token
	coin, err := lock.SingleCoin()
	if err != nil {
		return nil, osmomath.Int{}, errors.New("lock fails to meet expected invariant, it contains multiple coins")
	}

	// check that the lock denom is uosmo
	if coin.Denom == appParams.BaseCoinUnit {
		lockedOsmoAmount = lockedOsmoAmount.Add(coin.Amount)
	}

	// check if there is enough uosmo token in the lock
	if lockedOsmoAmount.LTE(osmomath.NewInt(0)) {
		return nil, osmomath.Int{}, errors.New("lock does not contain osmo denom, or there isn't enough osmo to unbond")
	}

	// Checks if lock ID is bonded and ensure that the duration is <= 2 weeks
	if lock.IsUnlocking() || lock.Duration > time.Hour*24*7*2 {
		return nil, osmomath.Int{}, errors.New("the tokens have to bonded and the duration has to be <= 2weeks")
	}

	return lock, lockedOsmoAmount, nil
}
