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

func (k Keeper) SetValidatorSetPreferences(ctx sdk.Context, delegator string, validators types.ValidatorSetPreferences) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, []byte(delegator), &validators)
}

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

func (k Keeper) GetDelegations(ctx sdk.Context, delegator string) (types.ValidatorSetPreferences, error) {
	valSet, exists := k.GetValidatorSetPreference(ctx, delegator)

	if !exists {
		existingDelsValSetFormatted, err := k.GetExistingStakingDelegations(ctx, delegator)
		if err != nil {
			return types.ValidatorSetPreferences{}, err
		}

		return types.ValidatorSetPreferences{Preferences: existingDelsValSetFormatted}, nil
	}

	return valSet, nil
}

func (k Keeper) GetExistingStakingDelegations(ctx sdk.Context, delegator string) ([]types.ValidatorPreference, error) {
	var existingDelsValSetFormatted []types.ValidatorPreference

	delAddr, err := sdk.AccAddressFromBech32(delegator)
	if err != nil {
		return nil, err
	}

	// valset delegation does not exist, so get all the existing delegations
	existingDelegations := k.stakingKeeper.GetDelegatorDelegations(ctx, delAddr, math.MaxUint16)
	existingTotalShares := sdk.NewDec(0)

	// calculate total shares that currently exists
	for _, existingdels := range existingDelegations {
		existingTotalShares = existingTotalShares.Add(existingdels.Shares)
	}

	// for each delegation format it in types.ValidatorSetPreferences format
	for _, existingdels := range existingDelegations {
		existingDelsValSetFormatted = append(existingDelsValSetFormatted, types.ValidatorPreference{
			ValOperAddress: existingdels.ValidatorAddress,
			Weight:         existingdels.Shares.Quo(existingTotalShares), // TODO: only 3 places decimal
		})
	}

	return existingDelsValSetFormatted, nil
}
