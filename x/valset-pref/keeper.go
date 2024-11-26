package keeper

import (
	"errors"
	"fmt"
	"math"

	"cosmossdk.io/log"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/valset-pref/types"

	storetypes "cosmossdk.io/store/types"
)

type Keeper struct {
	storeKey           storetypes.StoreKey
	paramSpace         paramtypes.Subspace
	stakingKeeper      types.StakingInterface
	distirbutionKeeper types.DistributionKeeper
	lockupKeeper       types.LockupKeeper
}

func NewKeeper(storeKey storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	stakingKeeper types.StakingInterface,
	distirbutionKeeper types.DistributionKeeper,
	lockupKeeper types.LockupKeeper,
) Keeper {
	return Keeper{
		storeKey:           storeKey,
		paramSpace:         paramSpace,
		stakingKeeper:      stakingKeeper,
		distirbutionKeeper: distirbutionKeeper,
		lockupKeeper:       lockupKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetDelegationPreferences checks if valset position exists, if it does return that
// else return existing delegation that's not valset.
func (k Keeper) GetDelegationPreferences(ctx sdk.Context, delegator string) (types.ValidatorSetPreferences, error) {
	valSet, exists := k.GetValidatorSetPreference(ctx, delegator)
	if !exists {
		delAddr, err := sdk.AccAddressFromBech32(delegator)
		if err != nil {
			return types.ValidatorSetPreferences{}, err
		}
		existingDelegations, err := k.stakingKeeper.GetDelegatorDelegations(ctx, delAddr, math.MaxUint16)
		if err != nil {
			return types.ValidatorSetPreferences{}, err
		}
		if len(existingDelegations) == 0 {
			return types.ValidatorSetPreferences{}, types.ErrNoDelegation
		}

		preferences, err := k.formatToValPrefArr(ctx, existingDelegations)
		if err != nil {
			return types.ValidatorSetPreferences{}, err
		}
		return types.ValidatorSetPreferences{Preferences: preferences}, nil
	}

	return valSet, nil
}

// GetValSetPreferencesWithDelegations fetches the delegator's validator set preferences
// considering their existing delegations.
// -If validator set preference does not exist and there are no existing delegations, it returns an error.
// -If validator set preference exists and there are no existing delegations, it returns the existing preference.
// -If there is any existing delegation:
// calculates the delegator's shares in each delegation
// as a ratio of the total shares and returns it as part of ValidatorSetPreferences.
func (k Keeper) GetValSetPreferencesWithDelegations(ctx sdk.Context, delegator string) (types.ValidatorSetPreferences, error) {
	delAddr, err := sdk.AccAddressFromBech32(delegator)
	if err != nil {
		return types.ValidatorSetPreferences{}, err
	}

	valSet, exists := k.GetValidatorSetPreference(ctx, delegator)
	existingDelegations, err := k.stakingKeeper.GetDelegatorDelegations(ctx, delAddr, math.MaxUint16)
	if err != nil {
		return types.ValidatorSetPreferences{}, err
	}

	// No existing delegations for a delegator when valSet does not exist
	if !exists && len(existingDelegations) == 0 {
		return types.ValidatorSetPreferences{}, errors.New("No Existing delegation to unbond from")
	}

	// Returning existing valSet when there are no existing delegations
	if exists && len(existingDelegations) == 0 {
		return valSet, nil
	}

	// when existing delegation exists, have it based upon the existing delegation
	// regardless of the delegator having valset pref or not
	preferences, err := k.formatToValPrefArr(ctx, existingDelegations)
	if err != nil {
		return types.ValidatorSetPreferences{}, err
	}
	return types.ValidatorSetPreferences{Preferences: preferences}, nil
}

// formatToValPrefArr iterates over given delegations array, formats it into ValidatorPreference array.
// Used to calculate weights for the each delegation towards validator.
// CONTRACT: This method assumes no duplicated ValOperAddress exists in the given delegation.
func (k Keeper) formatToValPrefArr(ctx sdk.Context, delegations []stakingtypes.Delegation) ([]types.ValidatorPreference, error) {
	totalTokens := osmomath.NewDec(0)

	// We cache token amounts for each delegation to avoid a second set of reads
	tokenDelegations := make(map[stakingtypes.Delegation]osmomath.Dec)
	for _, existingDelegation := range delegations {
		// Fetch validator corresponding to current delegation
		valAddr, err := sdk.ValAddressFromBech32(existingDelegation.ValidatorAddress)
		if err != nil {
			return []types.ValidatorPreference{}, err
		}
		validator, err := k.stakingKeeper.GetValidator(ctx, valAddr)
		if err != nil {
			return []types.ValidatorPreference{}, types.ValidatorNotFoundError{ValidatorAddr: existingDelegation.ValidatorAddress}
		}

		// Convert shares to underlying token amounts
		currentDelegationTokens := validator.TokensFromShares(existingDelegation.Shares)

		// Cache token amounts for each delegation and track total tokens
		tokenDelegations[existingDelegation] = currentDelegationTokens
		totalTokens = totalTokens.Add(currentDelegationTokens)
	}

	// Build ValidatorPreference array from delegations
	valPrefs := make([]types.ValidatorPreference, len(delegations))
	for i, delegation := range delegations {
		valPrefs[i] = types.ValidatorPreference{
			ValOperAddress: delegation.ValidatorAddress,
			// We accept bankers rounding here as rounding direction is not critical
			// and we want to minimize rounding error.
			Weight: tokenDelegations[delegation].Quo(totalTokens),
		}
	}
	return valPrefs, nil
}
