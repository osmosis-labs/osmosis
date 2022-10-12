package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v12/x/valset-pref/types"
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

// TODO MAYBE: Check if there are any banned assets and ways to handle them
// DelegateToValidatorSet delegates to a delegators existing validator-set.
// For ex: delegate 10osmo with validator-set {ValA -> 0.5, ValB -> 0.3, ValC -> 0.2}
// our delegate logic would attempt to delegate 5osmo to A , 2osmo to B, 3osmo to C
func (server msgServer) DelegateToValidatorSet(goCtx context.Context, msg *types.MsgDelegateToValidatorSet) (*types.MsgDelegateToValidatorSetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.DelegateToValidatorSet(ctx, msg.Delegator, msg.Coin)
	if err != nil {
		return nil, err
	}

	return &types.MsgDelegateToValidatorSetResponse{}, nil
}

func (server msgServer) UndelegateFromValidatorSet(goCtx context.Context, msg *types.MsgUndelegateFromValidatorSet) (*types.MsgUndelegateFromValidatorSetResponse, error) {
	return &types.MsgUndelegateFromValidatorSetResponse{}, nil
}

func (server msgServer) WithdrawDelegationRewards(goCtx context.Context, msg *types.MsgWithdrawDelegationRewards) (*types.MsgWithdrawDelegationRewardsResponse, error) {
	return &types.MsgWithdrawDelegationRewardsResponse{}, nil
}
