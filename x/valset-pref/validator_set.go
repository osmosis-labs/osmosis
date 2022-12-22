package keeper

import (
	"fmt"
	"math"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/v13/x/valset-pref/types"
)

type valSet struct {
	valAddr string
	weight  sdk.Dec
	amount  sdk.Dec
}

func (k Keeper) SetValidatorSetPreference(ctx sdk.Context, delegator string, preferences []types.ValidatorPreference) error {
	// check if a user already has a validator-set created
	existingValidators, found := k.GetValidatorSetPreference(ctx, delegator)
	if found {
		// check if the new preferences is the same as the existing preferences
		isEqual := k.IsValidatorSetEqual(preferences, existingValidators.Preferences)
		if isEqual {
			return fmt.Errorf("The preferences (validator and weights) are the same")
		}
	}

	// checks that all the validators exist on chain
	isValid := k.IsPreferenceValid(ctx, preferences)
	if !isValid {
		return fmt.Errorf("The validator preference list is not valid")
	}

	return nil
}

// DelegateToValidatorSet delegates to a delegators existing validator-set.
// For ex: delegate 10osmo with validator-set {ValA -> 0.5, ValB -> 0.3, ValC -> 0.2}
// our delegate logic would attempt to delegate 5osmo to A , 2osmo to B, 3osmo to C
func (k Keeper) DelegateToValidatorSet(ctx sdk.Context, delegatorAddr string, coin sdk.Coin) error {
	// get the existing validator set preference from store
	existingSet, found := k.GetValidatorSetPreference(ctx, delegatorAddr)
	if !found {
		return fmt.Errorf("user %s doesn't have validator set", delegatorAddr)
	}

	delegator, err := sdk.AccAddressFromBech32(delegatorAddr)
	if err != nil {
		return err
	}

	// loop through the validatorSetPreference and delegate the proportion of the tokens based on weights
	for _, val := range existingSet.Preferences {
		_, validator, err := k.getValAddrAndVal(ctx, val.ValOperAddress)
		if err != nil {
			return err
		}

		// tokenAmt takes the amount to delegate, calculated by {val_distribution_weight * tokenAmt}
		tokenAmt := val.Weight.Mul(coin.Amount.ToDec()).TruncateInt()

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
// For ex: userA has staked 10tokens with weight {Val->0.5, ValB->0.3, ValC->0.2}
// undelegate 6osmo with validator-set {ValA -> 0.5, ValB -> 0.3, ValC -> 0.2}
// our undelegate logic would attempt to undelegate 3osmo from A , 1.8osmo from B, 1.2osmo from C
func (k Keeper) UndelegateFromValidatorSet(ctx sdk.Context, delegatorAddr string, coin sdk.Coin) error {
	// get the existing validator set preference
	existingSet, found := k.GetValidatorSetPreference(ctx, delegatorAddr)
	if !found {
		return fmt.Errorf("user %s doesn't have validator set", delegatorAddr)
	}

	delegator, err := sdk.AccAddressFromBech32(delegatorAddr)
	if err != nil {
		return err
	}

	// the total amount the user wants to undelegate
	tokenAmt := sdk.NewDec(coin.Amount.Int64())

	totalAmountFromWeights := sdk.NewDec(0)
	for _, val := range existingSet.Preferences {
		totalAmountFromWeights = totalAmountFromWeights.Add(val.Weight.Mul(tokenAmt))
	}

	if !totalAmountFromWeights.Equal(tokenAmt) {
		return fmt.Errorf("The undelegate total do not add up with the amount calculated from weights expected %s got %s", tokenAmt, totalAmountFromWeights)
	}

	for _, val := range existingSet.Preferences {
		// Calculate the amount to undelegate based on the existing weights
		amountToUnDelegate := val.Weight.Mul(tokenAmt)

		valAddr, validator, err := k.getValAddrAndVal(ctx, val.ValOperAddress)
		if err != nil {
			return err
		}

		sharesAmt, err := validator.SharesFromTokens(amountToUnDelegate.TruncateInt())
		if err != nil {
			return err
		}

		_, err = k.stakingKeeper.Undelegate(ctx, delegator, valAddr, sharesAmt) // this has to be shares amount
		if err != nil {
			return err
		}
	}
	return nil
}

// The redelegation command allows delegators to instantly switch validators.
// Once the unbonding period has passed, the redelegation is automatically completed in the EndBlocker.
// A redelegation object is created every time a redelegation occurs. To prevent "redelegation hopping" redelegations may not occur under the situation that:
// 1. the (re)delegator already has another immature redelegation in progress with a destination to a validator (let's call it Validator X)
// 2. the (re)delegator is attempting to create a new redelegation where the source validator for this new redelegation is Validator X
// 3. the (re)delegator cannot create a new redelegation until the unbonding period i.e. 21 days.
func (k Keeper) PreformRedelegation(ctx sdk.Context, delegator sdk.AccAddress, existingSet types.ValidatorSetPreferences, newSet []types.ValidatorPreference) error {
	var existingValSet []valSet
	var newValSet []valSet
	totalTokenAmount := sdk.NewDec(0)

	// Rearranging the exisingValSet and newValSet to to add extra validator padding
	for _, existingVals := range existingSet.Preferences {
		valAddr, validator, err := k.GetValidatorInfo(ctx, existingVals.ValOperAddress)
		if err != nil {
			return err
		}

		// check if the user has delegated tokens to the valset
		delegation, found := k.stakingKeeper.GetDelegation(ctx, delegator, valAddr)
		if !found {
			return fmt.Errorf("No delegation found")
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
	var diffValSet []*valSet
	for i, newVals := range existingValSet {
		diffAmount := newVals.amount.Sub(newValSet[i].amount)

		diff_val := valSet{
			valAddr: newVals.valAddr,
			weight:  newVals.amount,
			amount:  diffAmount,
		}
		diffValSet = append(diffValSet, &diff_val)
	}

	// Algorithm starts here
	for _, diff_val := range diffValSet {
		for diff_val.amount.GT(sdk.NewDec(0)) {
			source_val := diff_val.valAddr
			// FindMin returns the index and MinAmt of the minimum amount in diffValSet
			target_val, idx := k.FindMin(diffValSet)

			validator_source, err := sdk.ValAddressFromBech32(source_val)
			if err != nil {
				return fmt.Errorf("validator address not formatted")
			}

			validator_target, err := sdk.ValAddressFromBech32(target_val.valAddr)
			if err != nil {
				return fmt.Errorf("validator address not formatted")
			}

			// reDelegationAmt to is the amount to redelegate, which is the min of diffAmount and target_validator
			reDelegationAmt := sdk.MinDec(target_val.amount.Abs(), diff_val.amount)
			_, err = k.stakingKeeper.BeginRedelegation(ctx, delegator, validator_source, validator_target, reDelegationAmt)
			if err != nil {
				return err
			}

			// Update the current diffAmount by subtracting it with the reDelegationAmount
			diff_val.amount = diff_val.amount.Sub(reDelegationAmt)
			// Find target_validator through idx in diffValSet and set that to (target_validatorAmount - reDelegationAmount)
			diffValSet[idx].amount = target_val.amount.Add(reDelegationAmt)
		}
	}

	return nil
}

// WithdrawDelegationRewards withdraws all the delegation rewards from the validator in the val-set.
// Delegation reward is collected by the validator and in doing so, they can charge commission to the delegators.
// Rewards are calculated per period, and is updated each time validator delegation changes. For ex: when a delegator
// receives new delgation the rewards can be calculated by taking (total rewards before new delegation - the total current rewards).
func (k Keeper) WithdrawDelegationRewards(ctx sdk.Context, delegatorAddr string) error {
	delegator, err := sdk.AccAddressFromBech32(delegatorAddr)
	if err != nil {
		return err
	}

	// check if there is existing staking position that's not val-set
	delegations := k.stakingKeeper.GetDelegatorDelegations(ctx, delegator, math.MaxUint16)

	// get the existing validator set preference
	existingSet, found := k.GetValidatorSetPreference(ctx, delegatorAddr)
	if !found && len(delegations) == 0 {
		return fmt.Errorf("user %s doesn't have validator set or existing delegations", delegatorAddr)
	}

	// there is existing staking position, but it's not valset
	if !found && len(delegations) != 0 {
		err := k.withdrawExistingStakingPosition(ctx, delegator, delegations)
		if err != nil {
			return err
		}
		return nil
	}

	// there is no existing staking position, but there is val-set delegation
	if found && len(delegations) == 0 {
		err := k.withdrawExistingValSetStakingPosition(ctx, delegator, existingSet.Preferences)
		if err != nil {
			return err
		}
		return nil
	}

	// there is staking position delegation, as well as val-set delegation
	err = k.withdrawExistingStakingPosition(ctx, delegator, delegations)
	if err != nil {
		return err
	}

	err = k.withdrawExistingValSetStakingPosition(ctx, delegator, existingSet.Preferences)
	if err != nil {
		return err
	}

	return nil
}

// withdrawExistingStakingPosition takes the existing staking delegator delegations and withdraws the rewards.
func (k Keeper) withdrawExistingStakingPosition(ctx sdk.Context, delegator sdk.AccAddress, delegations []stakingtypes.Delegation) error {
	for _, dels := range delegations {
		_, err := k.distirbutionKeeper.WithdrawDelegationRewards(ctx, delegator, dels.GetValidatorAddr())
		if err != nil {
			return err
		}
	}
	return nil
}

// withdrawExistingValSetStakingPosition takes the existing valset delegator delegations and withdraws the rewards.
func (k Keeper) withdrawExistingValSetStakingPosition(ctx sdk.Context, delegator sdk.AccAddress, delegations []types.ValidatorPreference) error {
	for _, dels := range delegations {
		valAddr, err := sdk.ValAddressFromBech32(dels.ValOperAddress)
		if err != nil {
			return fmt.Errorf("validator address not formatted")
		}

		_, err = k.distirbutionKeeper.WithdrawDelegationRewards(ctx, delegator, valAddr)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetValAddrAndVal checks if the validator address is valid and the validator provided exists on chain.
func (k Keeper) getValAddrAndVal(ctx sdk.Context, valOperAddress string) (sdk.ValAddress, stakingtypes.Validator, error) {
	valAddr, err := sdk.ValAddressFromBech32(valOperAddress)
	if err != nil {
		return nil, stakingtypes.Validator{}, fmt.Errorf("validator address not formatted")
	}

	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return nil, stakingtypes.Validator{}, fmt.Errorf("validator not found %s", validator)
	}

	return valAddr, validator, nil
}

// IsPreferenceValid loops through the validator preferences and checks its existence and validity.
func (k Keeper) IsPreferenceValid(ctx sdk.Context, preferences []types.ValidatorPreference) bool {
	for _, val := range preferences {
		_, _, err := k.GetValidatorInfo(ctx, val.ValOperAddress)
		if err != nil {
			return false
		}
	}
	return true
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
func (k Keeper) GetValSetStruct(validator types.ValidatorPreference, amountFromShares sdk.Dec) (valSet, valSet) {
	val_struct := valSet{
		valAddr: validator.ValOperAddress,
		weight:  validator.Weight,
		amount:  amountFromShares,
	}

	val_struct_zero_amount := valSet{
		valAddr: validator.ValOperAddress,
		weight:  validator.Weight,
		amount:  sdk.NewDec(0),
	}

	return val_struct, val_struct_zero_amount
}

// FindMin takes in a valSet struct array and computes the minimum val set based on the amount delegated to a validator.
func (k Keeper) FindMin(valPrefs []*valSet) (min valSet, idx int) {
	min = *valPrefs[0]
	idx = 0
	for i, val := range valPrefs {
		if val.amount.LT(min.amount) {
			min = *val
			idx = i
		}
	}
	return min, idx
}

// FindMax takes in a valSet struct array and computes the maximum val set based the amount delegated to a validator.
func (k Keeper) FindMax(valPrefs []*valSet) (max valSet, idx int) {
	max = *valPrefs[0]
	idx = 0
	for i, val := range valPrefs {
		if val.amount.GT(max.amount) {
			max = *val
			idx = i
		}
	}
	return max, idx
}
