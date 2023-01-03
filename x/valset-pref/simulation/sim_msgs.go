package simulation

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	osmosimtypes "github.com/osmosis-labs/osmosis/v13/simulation/simtypes"
	valsetkeeper "github.com/osmosis-labs/osmosis/v13/x/valset-pref"
	"github.com/osmosis-labs/osmosis/v13/x/valset-pref/types"
)

func RandomMsgSetValSetPreference(k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*types.MsgSetValidatorSetPreference, error) {
	var preferences []types.ValidatorPreference

	// Start with a weight of 1
	remainingWeight := sdk.NewDec(1)

	// Generate a random number of validators
	for remainingWeight.GT(sdk.ZeroDec()) {
		randValidator := RandomValidator(ctx, sim)
		if randValidator == nil {
			return nil, fmt.Errorf("No validator")
		}

		randValue, err := RandomWeight(remainingWeight)
		if err != nil {
			return nil, fmt.Errorf("Error with random weights")
		}

		remainingWeight = remainingWeight.Sub(randValue)

		preferences = append(preferences, types.ValidatorPreference{
			ValOperAddress: randValidator.OperatorAddress,
			Weight:         randValue,
		})
	}

	return &types.MsgSetValidatorSetPreference{
		Delegator:   sim.RandomSimAccount().Address.String(),
		Preferences: preferences,
	}, nil
}

func RandomMsgDelegateToValSet(k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*types.MsgDelegateToValidatorSet, error) {
	delegator := sim.RandomSimAccount()
	// check if the delegator valset created
	err := GetRandomExistingValSet(ctx, k, sim, delegator.Address)
	if err != nil {
		return nil, err
	}

	delegationCoin := sim.RandExponentialCoin(ctx, delegator.Address)

	return &types.MsgDelegateToValidatorSet{
		Delegator: delegator.Address.String(),
		Coin:      delegationCoin,
	}, nil
}

func RandomMsgUnDelegateToValSet(k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*types.MsgUndelegateFromValidatorSet, error) {
	delegator := sim.RandomSimAccount()
	// check if the delegator valset created
	err := GetRandomExistingValSet(ctx, k, sim, delegator.Address)
	if err != nil {
		return nil, err
	}

	// check that the delegator has delegated tokens
	err = GetRandomExistingDelegation(ctx, k, sim, delegator.Address)
	if err != nil {
		return nil, err
	}

	undelegationCoin := sim.RandExponentialCoin(ctx, delegator.Address)
	return &types.MsgUndelegateFromValidatorSet{
		Delegator: delegator.Address.String(),
		Coin:      undelegationCoin,
	}, nil
}

func RandomValidator(ctx sdk.Context, sim *osmosimtypes.SimCtx) *stakingtypes.Validator {
	rand.Seed(time.Now().UnixNano())

	validators := sim.StakingKeeper().GetAllValidators(ctx)
	if len(validators) == 0 {
		return nil
	}

	return &validators[rand.Intn(len(validators))]
}

// TODO: Change this to user GetDelegations() once #3857 gets merged
func GetRandomExistingValSet(ctx sdk.Context, k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, delegatorAddr sdk.AccAddress) error {
	// Get Valset delegations
	_, found := k.GetValidatorSetPreference(ctx, delegatorAddr.String())
	if !found {
		return fmt.Errorf("No val set preference")
	}

	return nil
}

func GetRandomExistingDelegation(ctx sdk.Context, k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, delegatorAddr sdk.AccAddress) error {
	// gets the existing delegation
	existingDelegations := sim.StakingKeeper().GetDelegatorDelegations(ctx, delegatorAddr, math.MaxUint16)
	if len(existingDelegations) == 0 {
		return fmt.Errorf("No existing delegation")
	}

	return nil
}

// Random float point from 0-1
func RandomWeight(maxVal sdk.Dec) (sdk.Dec, error) {
	rand.Seed(time.Now().UnixNano())
	val, err := maxVal.Float64()
	if err != nil {
		return sdk.Dec{}, err
	}

	randVal := rand.Float64() * val
	valWeightStr := fmt.Sprintf("%.2f", randVal)

	return sdk.MustNewDecFromStr(valWeightStr), nil
}
