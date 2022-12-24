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

/**
	**SetValidatorSetPreference**
	Questions;
	1. Should we store the existing staking position in modules state?
	2. If so, how are we going to make sure the to sync with val-set state and vice versa?

	Existing Staking Position:
	ValA -> 99osmo -> 0.6
	ValB -> 66osmo -> 0.4

	New ValSet Position:
	ValC -> 0.2 ->
	ValD -> 0.4 ->
	ValE -> 0.4 ->

	Note: User cannot SetValSet if existingSet(Staking position) already exist without modifying the existing weights

	**Delegate Logic flow**
	Existing Staking Position: (coming from modules state or not?)
	ValA -> 99osmo -> 0.6
	ValB -> 66osmo -> 0.4

	Delegate (100osmo)
	Existing:
	ValA -> 99 + 60 -> 159osmo
	ValB -> 66 + 40 -> 106osmo

	New ValSet Position:
	ValC -> 0.2 -> 20osmo
	ValD -> 0.4 -> 40osmo
	ValE -> 0.4 -> 40osmo

	**UnDelegate Logic flow**
	Existing Staking Position: (coming from modules state or not?)
	ValA -> 159osmo -> 0.6
	ValB -> 106osmo -> 0.4

	Undelegate (80osmo)
	Existing:
	ValA -> 159 - 48 -> 111osmo -> 0.6
	ValB -> 106 - 32 -> 74osmo -> 0.4


	New ValSet Position:
	ValC -> 0.2 -> 16osmo
	ValD -> 0.4 -> 32osmo
	ValE -> 0.4 -> 32osmo


	**SetValidatorSetPreference**
	Case1:
	- userA has existingStakingPosition(nonValSet) {ValA -> 99osmo, ValB-> 66osmo}
	- userA wants to convert his existingStakingPosition to valSetPosition
	- userA valsetPosition = {ValA -> 0.6(99osmo), ValB -> 0.4(66osmo)}

	Case2: What we do right now!
	- userA doesnot have existingStakingPosition
	- userA wants to create valSetPosition
	- userA valSetPosition = {ValA -> 0.3, ValB -> 0.2, ValC -> 0.5}

	Case3:
	- userA has existingStakingPosition(nonValSet) = {ValA -> 99osmo(0.6), ValB-> 66osmo(0.4)}
	- userA also wants to create a valSetPosition while maintaining existingStakingPosition
	- userA valSetPosition =  {ValC -> 0.3, ValD -> 0.2, ValE-> 0.4}
	- ISSUE: userA already has delegated tokens and staking position weight == 1
	- ISSUE: how do we make sure (existingStakingPosition + valSetPosition) weights == 1
	- QA: Do we store existingStakingPosition in state? How do we ensure state sync between valset and stakingposition?


	**Delegate Logic flow**
	- userA delegates to val-set with {ValAddr, weights}
	- userA delegates to existingStakingPosition for ex: {ValA -> 0.6(99osmo), ValB -> 0.4(66osmo)}
		- calculate the weight based on shares ratio

	- (maybe) userA has both val-set and existingStakingPosition, delegate to all validators?

	**UnDelegate Logic flow**
	- userA undelegates from val-set with {ValAddr, weights}
	- userA undelegates from existingStakingPosition for ex: {ValA -> 0.6(99osmo), ValB -> 0.4(66osmo)}
		- calculate the weight based on shares ratio

	- (maybe) userA has both val-set and existingStakingPosition, undelegate from all validators?

	**WithdrawDelegationReward**
	- if valset exists withdraw from there
	- else withdraw from current staking position
	- if none then error

	- (maybe): If both exists withdraw from both

	**Redelegation**
	- get existing staking position, return [] if it doesnot exist
	- get existing val-set position, return [] if it doesnot exist
	- if none exist error

	- (maybe) if both exist merge them both and redelegate as a whole
	- redelegate to a new set that the user provides

**/
