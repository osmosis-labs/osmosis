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

	// get the existing validator set preference from store
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

	// Message 2: Redelegate to valSet Prefereence
	// Get the sum of users delegated amount
	totalTokenAmount := sdk.NewDec(0)
	for _, existingVals := range existingSet.Preferences {
		valAddr, validator, err := server.keeper.GetValAddrAndVal(ctx, existingVals.ValOperAddress)
		if err != nil {
			return nil, err
		}

		newSetFirstValidator, err := sdk.ValAddressFromBech32(msg.Preferences[0].ValOperAddress)
		if err != nil {
			return nil, err
		}

		delegation, found := server.keeper.stakingKeeper.GetDelegation(ctx, delegator, valAddr)
		if !found {
			return nil, fmt.Errorf("No delegation found")
		}

		server.keeper.stakingKeeper.BeginRedelegation(ctx, delegator, valAddr, newSetFirstValidator, delegation.Shares)

		// we want to get the amount not shares so what we can get the sum of total amount
		amountFromShares := validator.TokensFromShares(delegation.Shares).RoundInt()
		totalTokenAmount = totalTokenAmount.Add(amountFromShares.ToDec())
	}

	// Calculate Amount from shares for new set
	for _, newVals := range msg.Preferences {
		amountToStake := newVals.Weight.Mul(totalTokenAmount).RoundInt()

		valAddr, validator, err := server.keeper.GetValAddrAndVal(ctx, newVals.ValOperAddress)
		if err != nil {
			return nil, err
		}

		newSetFirstValidator, err := sdk.ValAddressFromBech32(msg.Preferences[0].ValOperAddress)
		if err != nil {
			return nil, err
		}

		// to make sure that we donot redelegate to the same delegator
		if msg.Preferences[0].ValOperAddress != newVals.ValOperAddress {
			sharesFromAmount, err := validator.SharesFromTokens(amountToStake)
			if err != nil {
				return nil, err
			}

			server.keeper.stakingKeeper.BeginRedelegation(ctx, delegator, newSetFirstValidator, valAddr, sharesFromAmount)
		}
	}

	return &types.MsgRedelegateValidatorSetResponse{}, nil
}

func (server msgServer) WithdrawDelegationRewards(goCtx context.Context, msg *types.MsgWithdrawDelegationRewards) (*types.MsgWithdrawDelegationRewardsResponse, error) {
	return &types.MsgWithdrawDelegationRewardsResponse{}, nil
}
