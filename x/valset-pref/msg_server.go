package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v13/x/valset-pref/types"
)

type msgServer struct {
	keeper *Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

var _ types.MsgServer = msgServer{}

func (server msgServer) SetValidatorSetPreference(goCtx context.Context, msg *types.MsgSetValidatorSetPreference) (*types.MsgSetValidatorSetPreferenceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.SetValidatorSetPreference(ctx, msg.Delegator, msg.Preferences)
	if err != nil {
		return nil, err
	}

	setMsg := types.ValidatorSetPreferences{
		Preferences: msg.Preferences,
	}

	server.keeper.SetValidatorSetPreferences(ctx, msg.Delegator, setMsg)
	return &types.MsgSetValidatorSetPreferenceResponse{}, nil
}

func (server msgServer) DelegateToValidatorSet(goCtx context.Context, msg *types.MsgDelegateToValidatorSet) (*types.MsgDelegateToValidatorSetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.DelegateToValidatorSet(ctx, msg.Delegator, msg.Coin)
	if err != nil {
		return nil, err
	}

	return &types.MsgDelegateToValidatorSetResponse{}, nil
}

func (server msgServer) UndelegateFromValidatorSet(goCtx context.Context, msg *types.MsgUndelegateFromValidatorSet) (*types.MsgUndelegateFromValidatorSetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.UndelegateFromValidatorSet(ctx, msg.Delegator, msg.Coin)
	if err != nil {
		return nil, err
	}

	return &types.MsgUndelegateFromValidatorSetResponse{}, nil
}

func (server msgServer) RedelegateValidatorSet(goCtx context.Context, msg *types.MsgRedelegateValidatorSet) (*types.MsgRedelegateValidatorSetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	existingSet, found := server.keeper.GetValidatorSetPreference(ctx, msg.Delegator)
	if !found {
		return nil, fmt.Errorf("user %s doesn't have validator set", msg.Delegator)
	}

	delegator, err := sdk.AccAddressFromBech32(msg.Delegator)
	if err != nil {
		return nil, err
	}

	// Message 1: override the validator set preference set entry
	_, err = server.SetValidatorSetPreference(goCtx, &types.MsgSetValidatorSetPreference{
		Delegator:   msg.Delegator,
		Preferences: msg.Preferences,
	})
	if err != nil {
		return nil, err
	}

	// Message 2: Perform the actual redelegation
	err = server.keeper.PreformRedelegation(ctx, delegator, existingSet, msg.Preferences)
	if err != nil {
		return nil, err
	}

	return &types.MsgRedelegateValidatorSetResponse{}, nil
}

func (server msgServer) WithdrawDelegationRewards(goCtx context.Context, msg *types.MsgWithdrawDelegationRewards) (*types.MsgWithdrawDelegationRewardsResponse, error) {
	return &types.MsgWithdrawDelegationRewardsResponse{}, nil
}

/**
// Message 2: Work on redelegation
	var existingvalSet []valSet
	var newValSet []valSet
	totalTokenAmount := sdk.NewDec(0)
	for _, existingVals := range existingSet.Preferences {
		valAddr, validator, err := server.keeper.GetValidatorInfo(ctx, existingVals.ValOperAddress)
		if err != nil {
			return nil, err
		}

		delegation, found := server.keeper.stakingKeeper.GetDelegation(ctx, delegator, valAddr)
		if !found {
			return nil, fmt.Errorf("No delegation found")
		}

		amountFromShares := validator.TokensFromShares(delegation.Shares)

		existing_val, existing_val_test := server.keeper.SetupStructs(existingVals, amountFromShares)

		existingvalSet = append(existingvalSet, existing_val)
		newValSet = append(newValSet, existing_val_test)

		totalTokenAmount = totalTokenAmount.Add(amountFromShares)
		fmt.Println("EXISTING VAL: ", existingVals.ValOperAddress, amountFromShares)
	}

	// The total delegated sum by the user (totalTokenAmount)
	for _, newVals := range msg.Preferences {
		amountToStake := newVals.Weight.Mul(totalTokenAmount)

		new_val, new_val_test := server.keeper.SetupStructs(newVals, amountToStake)

		newValSet = append(newValSet, new_val)
		existingvalSet = append(existingvalSet, new_val_test)

		fmt.Println("NEW VAL: ", newVals.ValOperAddress, amountToStake)
	}

	// calculate the difference
	var diffValSet []*valSet
	for i, newVals := range existingvalSet {
		diffAmount := newVals.amount.Sub(newValSet[i].amount)

		fmt.Println("Internal DIFF AMOUNT", newVals.valAddr, diffAmount)
		diff_val := valSet{
			valAddr: newVals.valAddr,
			weight:  newVals.amount,
			amount:  diffAmount,
		}

		diffValSet = append(diffValSet, &diff_val)
	}

	// Algorithm starts here
	for _, diff_val := range diffValSet {
		for diff_val.amount.GT(sdk.NewDec(0)) {
			source_large := diff_val.valAddr
			target_large, idx := server.keeper.FindMin(diffValSet)

			validator_source, err := sdk.ValAddressFromBech32(source_large)
			if err != nil {
				return nil, fmt.Errorf("validator address not formatted")
			}

			validator_target, err := sdk.ValAddressFromBech32(target_large.valAddr)
			if err != nil {
				return nil, fmt.Errorf("validator address not formatted")
			}

			amount := sdk.MinDec(target_large.amount.Abs(), diff_val.amount)
			server.keeper.stakingKeeper.BeginRedelegation(ctx, delegator, validator_source, validator_target, amount)

			// Find target value in diffValSet and set that to (sourceAmt - targetAmt)
			diff_val.amount = diff_val.amount.Sub(amount)            // set the source to 0
			diffValSet[idx].amount = target_large.amount.Add(amount) // set the target to (sourceAmt - targetAmt)

			fmt.Println("FIRST", idx, source_large, target_large.valAddr, amount)
		}
	}

	for _, val := range diffValSet {
		valAddrSrc_small, err := sdk.ValAddressFromBech32(val.valAddr)
		if err != nil {
			return nil, fmt.Errorf("validator address not formatted")
		}

		validator, found := server.keeper.stakingKeeper.GetValidator(ctx, valAddrSrc_small)
		if !found {
			return nil, fmt.Errorf("validator not found %s", validator)
		}

		fmt.Println("Validators: ", validator.OperatorAddress, validator.DelegatorShares)

	}

**/
