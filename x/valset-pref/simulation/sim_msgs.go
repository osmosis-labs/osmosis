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

func RandomMsgUnDelegateFromValSet(k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*types.MsgUndelegateFromValidatorSet, error) {
	// get a random validafor
	val, ok := RandSliceElem(sim.StakingKeeper().GetAllValidators(ctx))
	if !ok {
		return nil, fmt.Errorf("validator is not ok")
	}

	// get validator delegations
	valAddr := val.GetOperator()
	delegations := sim.StakingKeeper().GetValidatorDelegations(ctx, val.GetOperator())
	if delegations == nil {
		return nil, fmt.Errorf("keeper does have any delegation entries")
	}

	// get delegation information including the delegator
	delegation := delegations[rand.Intn(len(delegations))]
	delAddr := delegation.GetDelegatorAddr()

	// check if the delegator has valset created
	err := GetRandomExistingValSet(ctx, k, sim, delAddr)
	if err != nil {
		return nil, err
	}

	// check if unbonding delegation has maximum number of entries
	if sim.StakingKeeper().HasMaxUnbondingDelegationEntries(ctx, delAddr, valAddr) {
		return nil, fmt.Errorf("keeper does have a max unbonding delegation entries")
	}

	// Set the max amount to undelegate
	totalBond := val.TokensFromShares(delegation.GetShares()).TruncateInt()
	if !totalBond.IsPositive() {
		return nil, fmt.Errorf("total bond is negative")
	}

	// unbond a random amount form 0 to totalBond(maxAmt)
	unbondAmt := rand.Intn(int(totalBond.Int64()))
	if unbondAmt == 0 {
		return nil, fmt.Errorf("unbond amount is zero")
	}

	return &types.MsgUndelegateFromValidatorSet{
		Delegator: delAddr.String(),
		Coin:      sdk.NewCoin(sim.StakingKeeper().BondDenom(ctx), totalBond),
	}, nil
}

func RandomMsgReDelegateToValSet(k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*types.MsgRedelegateValidatorSet, error) {
	// check if the delegator valset created
	// err := GetRandomExistingValSet(ctx, k, sim, delegator.Address)
	// if err != nil {
	// 	return nil, err
	// }

	valSource, ok := RandSliceElem(sim.StakingKeeper().GetAllValidators(ctx))
	if !ok {
		return nil, fmt.Errorf("validator is not ok")
	}

	srcAddr := valSource.GetOperator()
	delegations := sim.StakingKeeper().GetValidatorDelegations(ctx, srcAddr)
	if delegations == nil {
		return nil, fmt.Errorf("keeper does have any delegation entries")
	}

	// get random delegator from src validator
	delegation := delegations[rand.Intn(len(delegations))]
	delAddr := delegation.GetDelegatorAddr()

	if sim.StakingKeeper().HasReceivingRedelegation(ctx, delAddr, srcAddr) {
		return nil, fmt.Errorf("receveing redelegation is not allowed")
	}

	// Destination validators
	// TODO: might want to check that the destination validators are fine to redelegate to
	remainingWeight := sdk.NewDec(1)
	preferences, err := GetRandomValAndWeights(ctx, k, sim, remainingWeight)
	if err != nil {
		return nil, err
	}

	return &types.MsgRedelegateValidatorSet{
		Delegator:   delAddr.String(),
		Preferences: preferences,
	}, nil
}

func RandomMsgWithdrawRewardsFromValSet(k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*types.MsgWithdrawDelegationRewards, error) {
	delegator := sim.RandomSimAccount()

	err := GetRandomExistingValSet(ctx, k, sim, delegator.Address)
	if err != nil {
		return nil, err
	}

	// check that the delegator has delegated tokens
	delegations, err := GetRandomExistingDelegation(ctx, k, sim, delegator.Address)
	if err != nil {
		return nil, err
	}

	delegation := delegations[rand.Intn(len(delegations))]
	validator := sim.StakingKeeper().Validator(ctx, delegation.GetValidatorAddr())
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

// TODO: Change this to user GetDelegations() once #3857 gets merged, issue created
func GetRandomExistingValSet(ctx sdk.Context, k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, delegatorAddr sdk.AccAddress) error {
	// Get Valset delegations
	_, found := k.GetValidatorSetPreference(ctx, delegatorAddr.String())
	if !found {
		return fmt.Errorf("No val set preference")
	}

	return nil
}

func GetRandomExistingDelegation(ctx sdk.Context, k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, delegatorAddr sdk.AccAddress) ([]stakingtypes.Delegation, error) {
	// gets the existing delegation
	existingDelegations := sim.StakingKeeper().GetDelegatorDelegations(ctx, delegatorAddr, math.MaxUint16)
	if len(existingDelegations) == 0 {
		return nil, fmt.Errorf("No existing delegation")
	}

	return existingDelegations, nil
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
