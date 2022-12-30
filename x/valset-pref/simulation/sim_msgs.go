package simulation

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	osmosimtypes "github.com/osmosis-labs/osmosis/v13/simulation/simtypes"

	valsetkeeper "github.com/osmosis-labs/osmosis/v13/x/valset-pref"
	"github.com/osmosis-labs/osmosis/v13/x/valset-pref/types"
)

func RandomMsgSetValSetPreference(k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*types.MsgSetValidatorSetPreference, error) {
	// Generate random digit from 0-1 with
	//randWeight := RandomWeight()

	// Generate random validators
	randValidator := RandomValidator(ctx, sim)
	if randValidator == "" {
		return nil, fmt.Errorf("No validator")
	}

	return &types.MsgSetValidatorSetPreference{
		Delegator: sim.RandomSimAccount().Address.String(),
		Preferences: []types.ValidatorPreference{
			{
				Weight:         sdk.NewDec(1),
				ValOperAddress: randValidator,
			},
		},
	}, nil
}

func RandomMsgDelegateToValSet(k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*types.MsgDelegateToValidatorSet, error) {
	delegator := sim.RandomSimAccount()
	// check if the delegator has either a valset created
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
	// check if the delegator either a valset created
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

func RandomValidator(ctx sdk.Context, sim *osmosimtypes.SimCtx) string {
	validators := sim.StakingKeeper().GetValidators(ctx, 100)
	if len(validators) == 0 {
		return ""
	}

	valAddr := validators[rand.Intn(len(validators))]
	return valAddr.OperatorAddress
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
func RandomWeight() sdk.Dec {
	rand.Seed(time.Now().UnixNano())
	valWeightStr := fmt.Sprintf("%.2f", rand.Float64())

	return sdk.MustNewDecFromStr(valWeightStr)
}
