package simulation

import (
	"fmt"
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	osmosimtypes "github.com/osmosis-labs/osmosis/v15/simulation/simtypes"
	valsetkeeper "github.com/osmosis-labs/osmosis/v15/x/valset-pref"
	"github.com/osmosis-labs/osmosis/v15/x/valset-pref/types"
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
		return nil, fmt.Errorf("%s is not present", sdk.DefaultBondDenom)
	}

	delegationCoin := rand.Intn(int(amount.Int64()))

	return &types.MsgDelegateToValidatorSet{
		Delegator: delegator.Address.String(),
		Coin:      sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(int64(delegationCoin))),
	}, nil
}

func RandomMsgUnDelegateFromValSet(k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*types.MsgUndelegateFromValidatorSet, error) {
	// random delegator account
	delegator := sim.RandomSimAccount()
	delAddr := delegator.Address

	// get delegator valset preferences
	preferences, err := k.GetDelegationPreferences(ctx, delAddr.String())
	if err != nil {
		return nil, fmt.Errorf("no delegations found")
	}

	delegation := preferences.Preferences[rand.Intn(len(preferences.Preferences))]
	val, err := sdk.ValAddressFromBech32(delegation.ValOperAddress)
	if err != nil {
		return nil, fmt.Errorf("validator address not formatted")
	}

	validator, found := sim.StakingKeeper().GetValidator(ctx, val)
	if !found {
		return nil, fmt.Errorf("Validator not found")
	}

	// check if the user has delegated tokens to the valset
	del, found := sim.StakingKeeper().GetDelegation(ctx, delAddr, val)
	if !found {
		return nil, fmt.Errorf("user hasn't delegated tokens to the validator, %s", val.String())
	}

	totalBond := validator.TokensFromShares(del.GetShares()).TruncateInt()
	if !totalBond.IsPositive() {
		return nil, fmt.Errorf("%s is not present", sdk.DefaultBondDenom)
	}

	undelegationCoin := rand.Intn(int(totalBond.Int64()))

	return &types.MsgUndelegateFromValidatorSet{
		Delegator: delAddr.String(),
		Coin:      sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(int64(undelegationCoin))),
	}, nil
}

func RandomMsgReDelegateToValSet(k valsetkeeper.Keeper, sim *osmosimtypes.SimCtx, ctx sdk.Context) (*types.MsgRedelegateValidatorSet, error) {
	// random delegator account
	delegator := sim.RandomSimAccount()
	delAddr := delegator.Address

	// existing delegations
	delegations, err := k.GetDelegationPreferences(ctx, delAddr.String())
	if err != nil {
		return nil, fmt.Errorf("no delegations found")
	}

	for _, dels := range delegations.Preferences {
		val, err := sdk.ValAddressFromBech32(dels.ValOperAddress)
		if err != nil {
			return nil, fmt.Errorf("validator address not formatted")
		}

		if sim.StakingKeeper().HasReceivingRedelegation(ctx, delAddr, val) {
			return nil, fmt.Errorf("receiving redelegation is not allowed for source validators")
		}

		if sim.StakingKeeper().HasMaxUnbondingDelegationEntries(ctx, delAddr, val) {
			return nil, fmt.Errorf("keeper does have a max unbonding delegation entries")
		}

		// check if the user has delegated tokens to the valset
		_, found := sim.StakingKeeper().GetDelegation(ctx, delAddr, val)
		if !found {
			return nil, fmt.Errorf("user hasn't delegated tokens to the validator, %s", val.String())
		}
	}

	// new delegations to redelegate to
	remainingWeight := sdk.NewDec(1)
	preferences, err := GetRandomValAndWeights(ctx, k, sim, remainingWeight)
	if err != nil {
		return nil, err
	}

	// check if redelegation is possible to new validators
	for _, vals := range preferences {
		val, err := sdk.ValAddressFromBech32(vals.ValOperAddress)
		if err != nil {
			return nil, fmt.Errorf("validator address not formatted")
		}

		if sim.StakingKeeper().HasMaxUnbondingDelegationEntries(ctx, delAddr, val) {
			return nil, fmt.Errorf("keeper does have a max unbonding delegation entries")
		}

		if sim.StakingKeeper().HasReceivingRedelegation(ctx, delAddr, val) {
			return nil, fmt.Errorf("receveing redelegation is not allowed for target validators")
		}
	}

	return &types.MsgRedelegateValidatorSet{
		Delegator:   delAddr.String(),
		Preferences: preferences,
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

		randValue := sim.RandomDecAmount(remainingWeight)

		remainingWeight = remainingWeight.Sub(randValue)
		if !randValue.Equal(sdk.ZeroDec()) {
			preferences = append(preferences, types.ValidatorPreference{
				ValOperAddress: randValidator.OperatorAddress,
				Weight:         randValue,
			})
		}
	}

	totalWeight := sdk.ZeroDec()
	// check if all the weights in preferences equal 1
	for _, prefs := range preferences {
		totalWeight = totalWeight.Add(prefs.Weight)
	}

	if !totalWeight.Equal(sdk.OneDec()) {
		return nil, fmt.Errorf("generated weights donot equal 1 got: %d", totalWeight)
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
