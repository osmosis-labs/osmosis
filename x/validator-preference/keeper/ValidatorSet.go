package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/v12/x/validator-preference/types"
)

// ValidateValidator checks if the validator address is valid and the validator provided exists onchain.
func (k Keeper) ValidateValidator(ctx sdk.Context, valOperAddress string) (sdk.ValAddress, stakingtypes.Validator, error) {
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

// ValidatePreferences checks if the sum of the validator set equals 1.
func (k Keeper) ValidatePreferences(ctx sdk.Context, preferences []types.ValidatorPreference) error {
	total_weight := sdk.NewDec(0)
	for _, val := range preferences {
		// validation to check that the validator given is valid
		_, _, err := k.ValidateValidator(ctx, val.ValOperAddress)
		if err != nil {
			return err
		}

		total_weight = total_weight.Add(val.Weight)
	}

	// check if the total validator distribution weights equal 1
	if !total_weight.Equal(sdk.NewDec(1)) {
		return fmt.Errorf("The weights allocated to the validators do not add up to 1, %d", total_weight)
	}

	return nil
}
