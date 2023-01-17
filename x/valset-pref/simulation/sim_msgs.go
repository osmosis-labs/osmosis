package simulation

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	osmosimtypes "github.com/osmosis-labs/osmosis/v14/simulation/simtypes"
	valsetkeeper "github.com/osmosis-labs/osmosis/v14/x/valset-pref"
	"github.com/osmosis-labs/osmosis/v14/x/valset-pref/types"
)

func RandomMsgSetValSetPreference(k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*types.MsgSetValidatorSetPreference, error) {
	// Start with a weight of 1
	remainingWeight := sdk.NewDec(1)

	preferences, err := GetRandomValAndWeights(ctx, k, sim, remainingWeight)
	if err != nil {
		return nil, err
	}

	return &types.MsgSetValidatorSetPreference{
		Delegator:   sim.RandomSimAccount().Address.String(),
		Preferences: preferences,
	}, nil
}

func RandomMsgDelegateToValSet(k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*types.MsgDelegateToValidatorSet, error) {
	delegator := sim.RandomSimAccount()
	// check if the delegator valset created
	_, err := GetRandomDelegations(ctx, k, sim, delegator.Address)
	if err != nil {
		return nil, err
	}

	amount := sim.BankKeeper().GetBalance(ctx, delegator.Address, sdk.DefaultBondDenom).Amount
	if !amount.IsPositive() {
		return nil, fmt.Errorf(" balance is negative")
	}

	delegationCoin := rand.Intn(int(amount.Int64()))

	return &types.MsgDelegateToValidatorSet{
		Delegator: delegator.Address.String(),
		Coin:      sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(int64(delegationCoin))),
	}, nil
}

func RandomMsgUnDelegateFromValSet(k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*types.MsgUndelegateFromValidatorSet, error) {
	// random delegator
	delegator := sim.RandomSimAccount()

	// get a random validafor
	validatorDelegations, err := GetRandomDelegations(ctx, k, sim, delegator.Address)
	if err != nil {
		return nil, err
	}

	delegatorDels := sim.StakingKeeper().GetAllDelegatorDelegations(ctx, delegator.Address)
	if len(delegatorDels) == 0 {
		return nil, fmt.Errorf("number of delegators equal 0")
	}

	for _, vals := range validatorDelegations {
		dels := sim.StakingKeeper().GetValidatorDelegations(ctx, sdk.ValAddress(vals.ValOperAddress))
		if dels == nil {
			return nil, fmt.Errorf("keeper does have any delegation entries")
		}

	}

	undelegationCoin := sim.RandExponentialCoin(ctx, delegator.Address)
	return &types.MsgUndelegateFromValidatorSet{
		Delegator: delegator.Address.String(),
		Coin:      undelegationCoin,
	}, nil
}

func RandomMsgReDelegateToValSet(k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*types.MsgRedelegateValidatorSet, error) {

	delegator := sim.RandomSimAccount()
	delAddr := delegator.Address

	// source validator
	_, err := GetRandomDelegations(ctx, k, sim, delAddr)
	if err != nil {
		return nil, err
	}

	// gets the existing delegation
	existingDelegations := sim.StakingKeeper().GetDelegatorDelegations(ctx, delAddr, math.MaxUint16)
	if len(existingDelegations) == 0 {
		return nil, fmt.Errorf("No existing delegation")
	}

	for _, exVals := range existingDelegations {
		if len(sim.StakingKeeper().GetRedelegationsFromSrcValidator(ctx, exVals.GetValidatorAddr())) != 0 || sim.StakingKeeper().HasReceivingRedelegation(ctx, delAddr, exVals.GetValidatorAddr()) {
			return nil, fmt.Errorf("receveing redelegation is not allowed for source validators")
		}
	}

	// Destination validators
	remainingWeight := sdk.NewDec(1)
	preferences, err := GetRandomValAndWeights(ctx, k, sim, remainingWeight)
	if err != nil {
		return nil, err
	}

	for _, vals := range preferences {
		if len(sim.StakingKeeper().GetRedelegationsFromSrcValidator(ctx, sdk.ValAddress(vals.ValOperAddress))) != 0 || sim.StakingKeeper().HasReceivingRedelegation(ctx, delAddr, sdk.ValAddress(vals.ValOperAddress)) {
			return nil, fmt.Errorf("receveing redelegation is not allowed for target validators")
		}
	}

	return &types.MsgRedelegateValidatorSet{
		Delegator:   delAddr.String(),
		Preferences: preferences,
	}, nil
}

func RandomMsgWithdrawRewardsFromValSet(k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*types.MsgWithdrawDelegationRewards, error) {
	delegator := sim.RandomSimAccount()

	delegations, err := GetRandomDelegations(ctx, k, sim, delegator.Address)
	if err != nil {
		return nil, err
	}

	delegation := delegations[rand.Intn(len(delegations))]
	validator := sim.StakingKeeper().Validator(ctx, sdk.ValAddress(delegation.ValOperAddress))
	if validator == nil {
		return nil, fmt.Errorf("validator not found")
	}

	return &types.MsgWithdrawDelegationRewards{
		Delegator: delegator.Address.String(),
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

func GetRandomValAndWeights(ctx sdk.Context, k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, remainingWeight sdk.Dec) ([]types.ValidatorPreference, error) {
	var preferences []types.ValidatorPreference

	// Generate random validators with random weights that sums to 1
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
		if !randValue.Equal(sdk.ZeroDec()) {
			preferences = append(preferences, types.ValidatorPreference{
				ValOperAddress: randValidator.OperatorAddress,
				Weight:         randValue,
			})
		}
	}

	return preferences, nil
}

func GetRandomDelegations(ctx sdk.Context, k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, delegatorAddr sdk.AccAddress) ([]types.ValidatorPreference, error) {
	// Get Valset delegations
	delegations, err := k.GetDelegationPreferences(ctx, delegatorAddr.String())
	if err != nil {
		return nil, fmt.Errorf("No delegations found")
	}

	return delegations.Preferences, err
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

func RandSliceElem[E any](elems []E) (E, bool) {
	if len(elems) == 0 {
		var e E
		return e, false
	}

	return elems[rand.Intn(len(elems))], true
}
