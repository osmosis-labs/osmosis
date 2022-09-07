package keeper

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/v12/x/validator-preference/types"
)

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

// ValidatePreferences loops through the validator preferences and checks its existence and validity.
func (k Keeper) ValidatePreferences(ctx sdk.Context, preferences []types.ValidatorPreference) error {
	for _, val := range preferences {
		_, _, err := k.getValAddrAndVal(ctx, val.ValOperAddress)
		if err != nil {
			return err
		}
	}
	return nil
}

// ChargeForCreateValSet gets the creationFee (default 10osmo) and funds it to the community pool.
func (k Keeper) ChargeForCreateValSet(ctx sdk.Context, delegatorAddr string) error {
	// Send creation fee to community pool
	creationFee := k.GetParams(ctx).ValsetCreationFee
	if creationFee == nil {
		return fmt.Errorf("creation fee cannot be nil or 0 ")
	}

	accAddr, err := sdk.AccAddressFromBech32(delegatorAddr)
	if err != nil {
		return err
	}

	if err := k.distrKeeper.FundCommunityPool(ctx, creationFee, accAddr); err != nil {
		return err
	}

	return nil
}

// SortAndCompareValidatorSet returns true if the two preferences are equal
func (k Keeper) SortAndCompareValidatorSet(newPreferences, existingPreferences []types.ValidatorPreference) bool {
	if len(newPreferences) != len(existingPreferences) {
		return false
	}

	sort.Slice(newPreferences, func(i, j int) bool {
		return newPreferences[i].ValOperAddress < newPreferences[j].ValOperAddress
	})

	sort.Slice(existingPreferences, func(i, j int) bool {
		return existingPreferences[i].ValOperAddress < existingPreferences[j].ValOperAddress
	})

	// make sure that both valAddress and weights cannot be the same in the new val-set
	for i := range newPreferences {
		if newPreferences[i].ValOperAddress != existingPreferences[i].ValOperAddress ||
			!newPreferences[i].Weight.Equal(existingPreferences[i].Weight) {
			return false
		}
	}

	return true
}
