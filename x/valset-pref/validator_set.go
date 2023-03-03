package keeper

import (
	"fmt"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/protobuf/proto"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	appParams "github.com/osmosis-labs/osmosis/v15/app/params"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v15/x/valset-pref/types"
)

type valSet struct {
	ValAddr string
	Amount  sdk.Dec
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

// SetValidatorSetPreference creates or updates delegators validator set.
// Errors when the given preference is the same as the existing preference in state.
func (k Keeper) SetValidatorSetPreference(ctx sdk.Context, delegator string, preferences []types.ValidatorPreference) (types.ValidatorSetPreferences, error) {
	existingValSet, found := k.GetValidatorSetPreference(ctx, delegator)
	if found {
		// check if the new preferences is the same as the existing preferences
		isEqual := k.IsValidatorSetEqual(existingValSet.Preferences, preferences)
		if isEqual {
			return types.ValidatorSetPreferences{}, fmt.Errorf("The preferences (validator and weights) are the same")
		}
	}

	// checks that all the validators exist on chain
	valSetPref, err := k.IsPreferenceValid(ctx, preferences)
	if err != nil {
		return types.ValidatorSetPreferences{}, fmt.Errorf("The validator preference list is not valid")
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
		return fmt.Errorf("error upon getting delegation preference for addr %s", delegatorAddr)
	}

	delegator, err := sdk.AccAddressFromBech32(delegatorAddr)
	if err != nil {
		return err
	}

	// totalDelAmt is the amount that keeps running track of the amount of tokens delegated
	totalDelAmt := sdk.NewInt(0)
	tokenAmt := sdk.NewInt(0)

	// loop through the validatorSetPreference and delegate the proportion of the tokens based on weights
	for i, val := range existingSet.Preferences {
		_, validator, err := k.getValAddrAndVal(ctx, val.ValOperAddress)
		if err != nil {
			return err
		}

		// in the last valset iteration we dont calculate it from shares using decimals and trucation,
		// we use whats remaining to get more accurate value
		if len(existingSet.Preferences)-1 == i {
			tokenAmt = coin.Amount.Sub(totalDelAmt).ToDec().TruncateInt()
		} else {
			// tokenAmt takes the amount to delegate, calculated by {val_distribution_weight * tokenAmt}
			tokenAmt = val.Weight.Mul(coin.Amount.ToDec()).TruncateInt()
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
// For ex: userA has staked 10tokens with weight {Val->0.5, ValB->0.3, ValC->0.2}
// undelegate 6osmo with validator-set {ValA -> 0.5, ValB -> 0.3, ValC -> 0.2}
// our undelegate logic would attempt to undelegate 3osmo from A, 1.8osmo from B, 1.2osmo from C
// nolint: staticcheck
func (k Keeper) UndelegateFromValidatorSet(ctx sdk.Context, delegatorAddr string, coin sdk.Coin) error {
	// get the existingValSet if it exists, if not check existingStakingPosition and return it
	existingSet, err := k.GetDelegationPreferences(ctx, delegatorAddr)
	if err != nil {
		return fmt.Errorf("user %s doesn't have validator set", delegatorAddr)
	}

	delegator, err := sdk.AccAddressFromBech32(delegatorAddr)
	if err != nil {
		return err
	}

	// the total amount the user wants to undelegate
	tokenAmt := sdk.NewDec(coin.Amount.Int64())

	err = k.CheckUndelegateTotalAmount(tokenAmt, existingSet.Preferences)
	if err != nil {
		return err
	}

	// totalDelAmt is the amount that keeps running track of the amount of tokens undelegated
	totalUnDelAmt := sdk.NewInt(0)
	amountToUnDelegate := sdk.NewInt(0)

	for i, val := range existingSet.Preferences {
		valAddr, validator, err := k.getValAddrAndVal(ctx, val.ValOperAddress)
		if err != nil {
			return err
		}

		// in the last valset iteration we dont calculate it from shares using decimals and trucation,
		// we use whats remaining to get more accurate value
		if len(existingSet.Preferences)-1 == i {
			amountToUnDelegate = coin.Amount.Sub(totalUnDelAmt).ToDec().TruncateInt()
		} else {
			// Calculate the amount to undelegate based on the existing weights
			amountToUnDelegate = val.Weight.Mul(tokenAmt).TruncateInt()
			totalUnDelAmt = totalUnDelAmt.Add(amountToUnDelegate)
		}

		sharesAmt, err := validator.SharesFromTokens(amountToUnDelegate)
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

// CheckUndelegateTotalAmount checks if the tokenAmount equals the total amount calculated from valset weights.
func (k Keeper) CheckUndelegateTotalAmount(tokenAmt sdk.Dec, existingSet []types.ValidatorPreference) error {
	totalAmountFromWeights := sdk.NewDec(0)
	for _, val := range existingSet {
		totalAmountFromWeights = totalAmountFromWeights.Add(val.Weight.Mul(tokenAmt))
	}

	totalAmountFromWeights = totalAmountFromWeights.RoundInt().ToDec()
	tokenAmt = tokenAmt.RoundInt().ToDec()

	if !totalAmountFromWeights.Equal(tokenAmt) {
		return fmt.Errorf("The undelegate total do not add up with the amount calculated from weights expected %s got %s", tokenAmt, totalAmountFromWeights)
	}

	return nil
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
	totalTokenAmount := sdk.NewDec(0)

	// Rearranging the exisingValSet and newValSet to to add extra validator padding
	for _, existingVals := range existingSet {
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
		if diffVal.Amount.TruncateDec().GT(sdk.ZeroDec()) {
			for idx, targetDiffVal := range diffValSets {
				if targetDiffVal.Amount.TruncateDec().LT(sdk.ZeroDec()) && diffVal.ValAddr != targetDiffVal.ValAddr {
					valSource, valTarget, err := k.getValTargetAndSource(ctx, diffVal.ValAddr, targetDiffVal.ValAddr)
					if err != nil {
						return err
					}

					transferAmount := sdk.MinDec(diffVal.Amount, targetDiffVal.Amount.Abs()).TruncateDec()
					if transferAmount.Equal(sdk.ZeroDec()) {
						break
					}

					_, err = k.stakingKeeper.BeginRedelegation(ctx, delegator, valSource, valTarget, transferAmount)
					if err != nil {
						return err
					}

					diffVal.Amount = diffVal.Amount.Sub(transferAmount)
					diffValSets[idx].Amount = targetDiffVal.Amount.Add(transferAmount)

					if diffVal.Amount.Equal(sdk.ZeroDec()) {
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

// WithdrawDelegationRewards withdraws all the delegation rewards from the validator in the val-set.
// If the valset does not exist, it withdraws from existing staking position.
// Delegation reward is collected by the validator and in doing so, they can charge commission to the delegators.
// Rewards are calculated per period, and is updated each time validator delegation changes. For ex: when a delegator
// receives new delgation the rewards can be calculated by taking (total rewards before new delegation - the total current rewards).
func (k Keeper) WithdrawDelegationRewards(ctx sdk.Context, delegatorAddr string) error {
	// get the existingValSet if it exists, if not check existingStakingPosition and return it
	existingSet, err := k.GetDelegationPreferences(ctx, delegatorAddr)
	if err != nil {
		return fmt.Errorf("user %s doesn't have validator set or existing delegations", delegatorAddr)
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
			return fmt.Errorf("validator address not formatted")
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
	synthLocks := k.lockupKeeper.GetAllSyntheticLockupsByLockup(ctx, lockID)
	if len(synthLocks) != 0 {
		return sdk.Coin{}, fmt.Errorf("cannot use DelegateBondedTokens being used for superfluid.")
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
		return nil, stakingtypes.Validator{}, fmt.Errorf("validator address not formatted")
	}

	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return nil, stakingtypes.Validator{}, fmt.Errorf("validator not found %s", validator)
	}

	return valAddr, validator, nil
}

// IsPreferenceValid loops through the validator preferences and checks its existence and validity.
func (k Keeper) IsPreferenceValid(ctx sdk.Context, preferences []types.ValidatorPreference) ([]types.ValidatorPreference, error) {
	var weightsRoundedValPrefList []types.ValidatorPreference
	for _, val := range preferences {
		// round up weights
		valWeightStr := osmomath.SigFigRound(val.Weight, sdk.NewDec(10).Power(2).TruncateInt())

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
func (k Keeper) GetValSetStruct(validator types.ValidatorPreference, amountFromShares sdk.Dec) (valStruct valSet, valStructZeroAmt valSet) {
	val_struct := valSet{
		ValAddr: validator.ValOperAddress,
		Amount:  amountFromShares,
	}

	val_struct_zero_amount := valSet{
		ValAddr: validator.ValOperAddress,
		Amount:  sdk.NewDec(0),
	}

	return val_struct, val_struct_zero_amount
}

// check if lock owner matches the delegator, contains only uosmo and is bonded for <= 2weeks
func (k Keeper) validateLockForForceUnlock(ctx sdk.Context, lockID uint64, delegatorAddr string) (*lockuptypes.PeriodLock, sdk.Int, error) {
	// Checks if sender is lock ID owner
	lock, err := k.lockupKeeper.GetLockByID(ctx, lockID)
	if err != nil {
		return nil, sdk.Int{}, err
	}
	if lock.GetOwner() != delegatorAddr {
		return nil, sdk.Int{}, fmt.Errorf("delegator (%s) and lock owner (%s) does not match", delegatorAddr, lock.Owner)
	}

	lockedOsmoAmount := sdk.NewInt(0)

	// check that lock contains only 1 token
	coin, err := lock.SingleCoin()
	if err != nil {
		return nil, sdk.Int{}, fmt.Errorf("lock fails to meet expected invariant, it contains multiple coins")
	}

	// check that the lock denom is uosmo
	if coin.Denom == appParams.BaseCoinUnit {
		lockedOsmoAmount = lockedOsmoAmount.Add(coin.Amount)
	}

	// check if there is enough uosmo token in the lock
	if lockedOsmoAmount.LTE(sdk.NewInt(0)) {
		return nil, sdk.Int{}, fmt.Errorf("lock does not contain osmo denom, or there isn't enough osmo to unbond")
	}

	// Checks if lock ID is bonded and ensure that the duration is <= 2 weeks
	if lock.IsUnlocking() || lock.Duration > time.Hour*24*7*2 {
		return nil, sdk.Int{}, fmt.Errorf("the tokens have to bonded and the duration has to be <= 2weeks")
	}

	return lock, lockedOsmoAmount, nil
}
