package keeper

import (
	"fmt"
	"math"

	"github.com/tendermint/tendermint/libs/log"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/v19/x/valset-pref/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
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
			return types.ValidatorSetPreferences{}, types.ErrNoDelegation
		}

		return types.ValidatorSetPreferences{Preferences: formatToValPrefArr(existingDelegations)}, nil
	}

	return valSet, nil
}

// formatToValPrefArr iterates over given delegations array, formats it into ValidatorPreference array.
// Used to calculate weights for the each delegation towards validator.
// CONTRACT: This method assumes no duplicated ValOperAddress exists in the given delegation.
func formatToValPrefArr(delegations []stakingtypes.Delegation) []types.ValidatorPreference {
	totalShares := sdk.NewDec(0)
	for _, existingDelegation := range delegations {
		totalShares = totalShares.Add(existingDelegation.Shares)
	}

	valPrefs := make([]types.ValidatorPreference, len(delegations))
	for i, delegation := range delegations {
		valPrefs[i] = types.ValidatorPreference{
			ValOperAddress: delegation.ValidatorAddress,
			Weight:         delegation.Shares.Quo(totalShares),
		}
	}
	return valPrefs
}
