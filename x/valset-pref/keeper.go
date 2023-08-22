package keeper

import (
	"fmt"
	"math"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/v17/x/valset-pref/types"
)

type Keeper struct {
	storeKey           sdk.StoreKey
	paramSpace         paramtypes.Subspace
	stakingKeeper      types.StakingInterface
	distirbutionKeeper types.DistributionKeeper
	lockupKeeper       types.LockupKeeper
}

func NewKeeper(storeKey sdk.StoreKey,
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

		existingDelegations := k.stakingKeeper.GetDelegatorDelegations(ctx, delAddr, math.MaxUint16)
		if len(existingDelegations) == 0 {
			return types.ValidatorSetPreferences{}, fmt.Errorf("No Existing delegation")
		}

		return types.ValidatorSetPreferences{Preferences: calculateSharesAndFormat(existingDelegations)}, nil
	}

	return valSet, nil
}

// GetValSetPreferencesWithDelegations fetches the delegator's validator set preferences
// considering their existing delegations.
// -If validator set preference does not exist and there are no existing delegations, it returns an error.
// -If validator set preference exists and there are no existing delegations, it returns the existing preference.
// -If there is a validator set preference and existing delegations, or existing delegations
// but no validator set preference, it calculates the delegator's shares in each delegation
// as a ratio of the total shares and returns it as part of ValidatorSetPreferences.
func (k Keeper) GetValSetPreferencesWithDelegations(ctx sdk.Context, delegator string) (types.ValidatorSetPreferences, error) {
	delAddr, err := sdk.AccAddressFromBech32(delegator)
	if err != nil {
		return types.ValidatorSetPreferences{}, err
	}

	valSet, exists := k.GetValidatorSetPreference(ctx, delegator)
	existingDelegations := k.stakingKeeper.GetDelegatorDelegations(ctx, delAddr, math.MaxUint16)

	// No existing delegations for a delegator when valSet does not exist
	if !exists && len(existingDelegations) == 0 {
		return types.ValidatorSetPreferences{}, fmt.Errorf("No Existing delegation to unbond from")
	}

	// Returning existing valSet when there are no existing delegations
	if exists && len(existingDelegations) == 0 {
		return valSet, nil
	}

	// this can either be valSet doesnot exist and existing delegations exist
	// or valset exists and existing delegation exists
	return types.ValidatorSetPreferences{Preferences: calculateSharesAndFormat(existingDelegations)}, nil
}

func calculateSharesAndFormat(existingDelegations []stakingtypes.Delegation) []types.ValidatorPreference {
	existingTotalShares := sdk.NewDec(0)
	for _, existingDelegation := range existingDelegations {
		existingTotalShares = existingTotalShares.Add(existingDelegation.Shares)
	}

	existingDelsValSetFormatted := make([]types.ValidatorPreference, len(existingDelegations))
	for i, existingDelegation := range existingDelegations {
		existingDelsValSetFormatted[i] = types.ValidatorPreference{
			ValOperAddress: existingDelegation.ValidatorAddress,
			Weight:         existingDelegation.Shares.Quo(existingTotalShares),
		}
	}
	return existingDelsValSetFormatted
}
