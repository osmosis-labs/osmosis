package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/valset-pref/types"
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

	preferences, err := server.keeper.ValidateValidatorSetPreference(ctx, msg.Delegator, msg.Preferences)
	if err != nil {
		return nil, err
	}

	server.keeper.SetValidatorSetPreferences(ctx, msg.Delegator, preferences)
	return &types.MsgSetValidatorSetPreferenceResponse{}, nil
}

// DelegateToValidatorSet delegates to a delegators existing validator-set.
func (server msgServer) DelegateToValidatorSet(goCtx context.Context, msg *types.MsgDelegateToValidatorSet) (*types.MsgDelegateToValidatorSetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.DelegateToValidatorSet(ctx, msg.Delegator, msg.Coin)
	if err != nil {
		return nil, err
	}

	return &types.MsgDelegateToValidatorSetResponse{}, nil
}

// UndelegateFromValidatorSet undelegates {coin} amount from the validator set.
func (server msgServer) UndelegateFromValidatorSet(goCtx context.Context, msg *types.MsgUndelegateFromValidatorSet) (*types.MsgUndelegateFromValidatorSetResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)

	// err := server.keeper.UndelegateFromValidatorSet(ctx, msg.Delegator, msg.Coin)
	// if err != nil {
	// 	return nil, err
	// }

	return &types.MsgUndelegateFromValidatorSetResponse{}, errors.New("not implemented, utilize UndelegateFromRebalancedValidatorSet instead")
}

// UndelegateFromRebalancedValidatorSet undelegates {coin} amount from the validator set, utilizing a user's current delegations
// to their validator set to determine the weights.
func (server msgServer) UndelegateFromRebalancedValidatorSet(goCtx context.Context, msg *types.MsgUndelegateFromRebalancedValidatorSet) (*types.MsgUndelegateFromRebalancedValidatorSetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.UndelegateFromRebalancedValidatorSet(ctx, msg.Delegator, msg.Coin)
	if err != nil {
		return nil, err
	}

	return &types.MsgUndelegateFromRebalancedValidatorSetResponse{}, nil
}

// RedelegateValidatorSet allows delegators to set a new validator set and switch validators.
func (server msgServer) RedelegateValidatorSet(goCtx context.Context, msg *types.MsgRedelegateValidatorSet) (*types.MsgRedelegateValidatorSetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegator, err := sdk.AccAddressFromBech32(msg.Delegator)
	if err != nil {
		return nil, err
	}

	// get existing delegation if there is no valset set, else get valset
	existingSet, err := server.keeper.GetDelegationPreferences(ctx, msg.Delegator)
	if err != nil {
		return nil, errors.New("user has no delegation")
	}

	// Message 1: override the validator set preference set entry
	newPreferences, err := server.keeper.ValidateValidatorSetPreference(ctx, msg.Delegator, msg.Preferences)
	if err != nil {
		return nil, err
	}

	server.keeper.SetValidatorSetPreferences(ctx, msg.Delegator, newPreferences)

	// Message 2: Perform the actual redelegation
	err = server.keeper.PreformRedelegation(ctx, delegator, existingSet.Preferences, newPreferences.Preferences)
	if err != nil {
		return nil, err
	}

	return &types.MsgRedelegateValidatorSetResponse{}, nil
}

// WithdrawDelegationRewards withdraws all the delegation rewards from the validator in the val-set.
func (server msgServer) WithdrawDelegationRewards(goCtx context.Context, msg *types.MsgWithdrawDelegationRewards) (*types.MsgWithdrawDelegationRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.WithdrawDelegationRewards(ctx, msg.Delegator)
	if err != nil {
		return nil, err
	}

	return &types.MsgWithdrawDelegationRewardsResponse{}, nil
}

// DelegateBondedTokens force unlocks bonded uosmo and stakes according to your current validator set preference.
func (server msgServer) DelegateBondedTokens(goCtx context.Context, msg *types.MsgDelegateBondedTokens) (*types.MsgDelegateBondedTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// get the existingValSet if it exists, if not check existingStakingPosition and return it
	_, err := server.keeper.GetDelegationPreferences(ctx, msg.Delegator)
	if err != nil {
		return nil, types.NoValidatorSetOrExistingDelegationsError{DelegatorAddr: msg.Delegator}
	}

	// Message 1: force unlock bonded osmo tokens.
	unlockedOsmoToken, err := server.keeper.ForceUnlockBondedOsmo(ctx, msg.LockID, msg.Delegator)
	if err != nil {
		return nil, err
	}

	delegator, err := sdk.AccAddressFromBech32(msg.Delegator)
	if err != nil {
		return nil, err
	}

	// Message 2: Perform osmo token delegation.
	_, err = server.DelegateToValidatorSet(goCtx, types.NewMsgDelegateToValidatorSet(delegator, unlockedOsmoToken))
	if err != nil {
		return nil, err
	}

	return &types.MsgDelegateBondedTokensResponse{}, nil
}
