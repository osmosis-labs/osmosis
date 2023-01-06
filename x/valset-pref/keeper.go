package keeper

import (
	"fmt"
	"math"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/gogo/protobuf/proto"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v14/x/valset-pref/types"
)

type Keeper struct {
	storeKey           sdk.StoreKey
	paramSpace         paramtypes.Subspace
	stakingKeeper      types.StakingInterface
	distirbutionKeeper types.DistributionKeeper
}

func NewKeeper(storeKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	stakingKeeper types.StakingInterface,
	distirbutionKeeper types.DistributionKeeper,
) Keeper {
	return Keeper{
		storeKey:           storeKey,
		paramSpace:         paramSpace,
		stakingKeeper:      stakingKeeper,
		distirbutionKeeper: distirbutionKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
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

// TODO: Write unit test
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

// GetExistingStakingDelegations returns the existing delegation that's not valset.
// This function also formats the output into ValidatorSetPreference struct where with {valAddr, weight}.
// The weight is calculated based on (valDelegation / totalDelegations) for each validator.
func (k Keeper) GetExistingStakingDelegations(ctx sdk.Context, delAddr sdk.AccAddress) ([]types.ValidatorPreference, error) {
	var existingDelsValSetFormatted []types.ValidatorPreference

	// valset delegation does not exist, so get all the existing delegations
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
