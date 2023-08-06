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

		existingDelsValSetFormatted, err := k.GetExistingStakingDelegations(ctx, delAddr)
		if err != nil {
			return types.ValidatorSetPreferences{}, err
		}

		return types.ValidatorSetPreferences{Preferences: existingDelsValSetFormatted}, nil
	}

	return valSet, nil
}

// getValsetDelegationsAndPreferences retrieves the validator preferences and
// delegations for a given delegator. It returns the ValidatorSetPreferences,
// a slice of Delegations, and an error if any issues occur during the process.
//
// The function first retrieves the validator set preferences associated with the delegator.
// If preferences exist, it iterates over them and fetches each associated delegation,
// adding it to a slice of delegations. If no preferences exist, it gets all delegator
// delegations.
//
// Params:
//
//	ctx        - The sdk.Context representing the current context.
//	delegator  - The address (as a string) of the delegator whose preferences and
//	             delegations are to be fetched.
//
// Returns:
//
//	The ValidatorSetPreferences object associated with the delegator.
//	A slice of Delegation objects for the delegator.
//	An error if any issues occur during the process.
func (k Keeper) getValsetDelegationsAndPreferences(ctx sdk.Context, delegator string) (types.ValidatorSetPreferences, []stakingtypes.Delegation, error) {
	delAddr, err := sdk.AccAddressFromBech32(delegator)
	if err != nil {
		return types.ValidatorSetPreferences{}, nil, err
	}

	valSet, exists := k.GetValidatorSetPreference(ctx, delegator)
	var delegations []stakingtypes.Delegation
	if exists {
		for _, val := range valSet.Preferences {
			del, found := k.stakingKeeper.GetDelegation(ctx, delAddr, sdk.ValAddress(val.ValOperAddress))
			if !found {
				del = stakingtypes.Delegation{DelegatorAddress: delegator, ValidatorAddress: (val.ValOperAddress), Shares: sdk.ZeroDec()}
			}
			delegations = append(delegations, del)
		}
	}
	if !exists {
		delegations = k.stakingKeeper.GetDelegatorDelegations(ctx, delAddr, math.MaxUint16)
		if len(delegations) == 0 {
			return types.ValidatorSetPreferences{}, nil, fmt.Errorf("No Existing delegation to unbond from")
		}
	}

	return valSet, delegations, nil
}

// GetExistingStakingDelegations returns the existing delegation that's not valset.
// This function also formats the output into ValidatorSetPreference struct {valAddr, weight}.
// The weight is calculated based on (valDelegation / totalDelegations) for each validator.
// This method erros when given address does not have any existing delegations.
func (k Keeper) GetExistingStakingDelegations(ctx sdk.Context, delAddr sdk.AccAddress) ([]types.ValidatorPreference, error) {
	var existingDelsValSetFormatted []types.ValidatorPreference

	existingDelegations := k.stakingKeeper.GetDelegatorDelegations(ctx, delAddr, math.MaxUint16)
	if len(existingDelegations) == 0 {
		return nil, fmt.Errorf("No Existing delegation")
	}

	existingTotalShares := sdk.NewDec(0)
	// calculate total shares that currently exists
	for _, existingDelegation := range existingDelegations {
		existingTotalShares = existingTotalShares.Add(existingDelegation.Shares)
	}

	// for each delegation format it in types.ValidatorSetPreferences format
	for _, existingDelegation := range existingDelegations {
		existingDelsValSetFormatted = append(existingDelsValSetFormatted, types.ValidatorPreference{
			ValOperAddress: existingDelegation.ValidatorAddress,
			Weight:         existingDelegation.Shares.Quo(existingTotalShares),
		})
	}

	return existingDelsValSetFormatted, nil
}
