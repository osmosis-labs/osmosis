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
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.WithdrawDelegationRewards(ctx, msg.Delegator)
	if err != nil {
		return nil, err
	}

	return &types.MsgWithdrawDelegationRewardsResponse{}, nil
}
