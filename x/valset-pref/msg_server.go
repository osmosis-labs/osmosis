package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v12/x/valset-pref/types"
)

type msgServer struct {
	keeper *Keeper
}

type valSet struct {
	valAddr string
	weight  sdk.Dec
	amount  sdk.Dec
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

	err := server.keeper.SetupValidatorSetPreference(ctx, msg.Delegator, msg.Preferences)
	if err != nil {
		return nil, err
	}

	// create/update the validator-set based on what user provides
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

// UnStakeFromValidatorSet unstakes all the tokens from the validator set.
// For ex: undelegate 10osmo with validator-set {ValA -> 0.5, ValB -> 0.3, ValC -> 0.2}
// our undelegate logic would attempt to undelegate 5osmo from A , 2osmo from B, 3osmo from C
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

	// Message 1: override the validator set preference set entry
	delegator, err := sdk.AccAddressFromBech32(msg.Delegator)
	if err != nil {
		return nil, err
	}

	_, err = server.SetValidatorSetPreference(goCtx, &types.MsgSetValidatorSetPreference{
		Delegator:   msg.Delegator,
		Preferences: msg.Preferences,
	})
	if err != nil {
		return nil, err
	}

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
		fmt.Println("SHARE AMOUNT, ", amountFromShares)
		existing_val := valSet{
			valAddr: existingVals.ValOperAddress,
			weight:  existingVals.Weight,
			amount:  amountFromShares,
		}

		existingvalSet = append(existingvalSet, existing_val)
		totalTokenAmount = totalTokenAmount.Add(amountFromShares)
	}

	fmt.Println(existingvalSet)

	// The total delegated sum by the user (totalTokenAmount)
	for _, newVals := range msg.Preferences {
		amountToStake := newVals.Weight.Mul(totalTokenAmount)
		fmt.Println("STAKE AMOUNT, ", amountToStake)
		new_val := valSet{
			valAddr: newVals.ValOperAddress,
			weight:  newVals.Weight,
			amount:  amountToStake,
		}
		newValSet = append(newValSet, new_val)
	}

	fmt.Println(newValSet)

	// calculate the difference
	var diffValSet []valSet
	for i, newVals := range msg.Preferences {
		diffAmount := existingvalSet[i].amount.Sub(newValSet[i].amount)
		fmt.Println("DIFF AMOUNT, ", diffAmount)
		diff_val := valSet{
			valAddr: newVals.ValOperAddress,
			weight:  newVals.Weight,
			amount:  diffAmount,
		}
		diffValSet = append(newValSet, diff_val)
	}

	fmt.Println(diffValSet)

	// Algorithm starts here
	for i, diff_val := range diffValSet {
		if diff_val.amount.GT(sdk.NewDec(0)) {
			source_large := diff_val.valAddr
			target_large := server.keeper.FindMin(diffValSet)

			valAddrSrc_large, err := sdk.ValAddressFromBech32(source_large)
			if err != nil {
				return nil, fmt.Errorf("validator address not formatted")
			}

			valAddrTarget_large, err := sdk.ValAddressFromBech32(target_large.valAddr)
			if err != nil {
				return nil, fmt.Errorf("validator address not formatted")
			}

			server.keeper.stakingKeeper.BeginRedelegation(ctx, delegator, valAddrSrc_large, valAddrTarget_large, diff_val.amount)
			diffValSet[i].amount = sdk.NewDec(0)
		}

		if diff_val.amount.LT(sdk.NewDec(0)) {
			source_small := diff_val.valAddr
			target_small := server.keeper.FindMax(diffValSet)

			valAddrSrc_small, err := sdk.ValAddressFromBech32(source_small)
			if err != nil {
				return nil, fmt.Errorf("validator address not formatted")
			}

			valAddrTarget_small, err := sdk.ValAddressFromBech32(target_small.valAddr)
			if err != nil {
				return nil, fmt.Errorf("validator address not formatted")
			}

			server.keeper.stakingKeeper.BeginRedelegation(ctx, delegator, valAddrSrc_small, valAddrTarget_small, diff_val.amount.Abs())
			diffValSet[i].amount = diffValSet[i].amount.Add(diff_val.amount.Abs())
		}
	}

	return &types.MsgRedelegateValidatorSetResponse{}, nil
}

func (server msgServer) WithdrawDelegationRewards(goCtx context.Context, msg *types.MsgWithdrawDelegationRewards) (*types.MsgWithdrawDelegationRewardsResponse, error) {
	return &types.MsgWithdrawDelegationRewardsResponse{}, nil
}

// ctx := sdk.UnwrapSDKContext(goCtx)

// // get the existing validator set preference from store
// existingSet, found := server.keeper.GetValidatorSetPreference(ctx, msg.Delegator)
// if !found {
// 	return nil, fmt.Errorf("user %s doesn't have validator set", msg.Delegator)
// }

// // Message 1: override the validator set preference set entry
// delegator, err := sdk.AccAddressFromBech32(msg.Delegator)
// if err != nil {
// 	return nil, err
// }

// _, err = server.SetValidatorSetPreference(goCtx, &types.MsgSetValidatorSetPreference{
// 	Delegator:   msg.Delegator,
// 	Preferences: msg.Preferences,
// })
// if err != nil {
// 	return nil, err
// }

// // Message 2: Redelegate to valSet Prefereence
// // Get the sum of users delegated amount
// totalTokenAmount := sdk.NewDec(0)
// for _, existingVals := range existingSet.Preferences {
// 	valAddr, validator, newSetFirstValidator, err := server.keeper.GetValidatorInfo(ctx, existingVals.ValOperAddress, msg.Preferences[0].ValOperAddress)
// 	if err != nil {
// 		return nil, err
// 	}

// 	delegation, found := server.keeper.stakingKeeper.GetDelegation(ctx, delegator, valAddr)
// 	if !found {
// 		return nil, fmt.Errorf("No delegation found")
// 	}

// 	server.keeper.stakingKeeper.BeginRedelegation(ctx, delegator, valAddr, newSetFirstValidator, delegation.Shares)

// 	// we want to get the amount not shares so what we can get the sum of total amount
// 	amountFromShares := validator.TokensFromShares(delegation.Shares).RoundInt()
// 	totalTokenAmount = totalTokenAmount.Add(amountFromShares.ToDec())
// }

// // Calculate Amount from shares for new set
// for _, newVals := range msg.Preferences {
// 	amountToStake := newVals.Weight.Mul(totalTokenAmount).RoundInt()

// 	valAddr, validator, newSetFirstValidator, err := server.keeper.GetValidatorInfo(ctx, newVals.ValOperAddress, msg.Preferences[0].ValOperAddress)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// to make sure that we donot redelegate to the same delegator
// 	if msg.Preferences[0].ValOperAddress != newVals.ValOperAddress {
// 		sharesFromAmount, err := validator.SharesFromTokens(amountToStake)
// 		if err != nil {
// 			return nil, err
// 		}

// 		server.keeper.stakingKeeper.BeginRedelegation(ctx, delegator, newSetFirstValidator, valAddr, sharesFromAmount)
// 	}
// }

// return &types.MsgRedelegateValidatorSetResponse{}, nil
