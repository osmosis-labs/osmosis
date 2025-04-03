package keeper

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/v27/x/oracle/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the oracle MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (ms *msgServer) AggregateExchangeRatePrevote(goCtx context.Context, msg *types.MsgAggregateExchangeRatePrevote) (*types.MsgAggregateExchangeRatePrevoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valAddr, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}

	feederAddr, err := sdk.AccAddressFromBech32(msg.Feeder)
	if err != nil {
		return nil, err
	}

	if err := ms.ValidateFeeder(ctx, feederAddr, valAddr); err != nil {
		return nil, err
	}

	// Convert hex string to votehash
	voteHash, err := types.AggregateVoteHashFromHexString(msg.Hash)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidHash, err.Error())
	}

	// lazy init currentVotePeriodEpochCounter when it is necessary
	currentEpochCounter := ms.epochKeeper.GetEpochInfo(ctx, ms.GetParams(ctx).VotePeriodEpochIdentifier).CurrentEpoch
	aggregatePrevote := types.NewAggregateExchangeRatePrevote(voteHash, valAddr, uint64(currentEpochCounter))
	ms.SetAggregateExchangeRatePrevote(ctx, valAddr, aggregatePrevote)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeAggregatePrevote,
			sdk.NewAttribute(types.AttributeKeyVoter, msg.Validator),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Feeder),
		),
	})

	return &types.MsgAggregateExchangeRatePrevoteResponse{}, nil
}

func (ms msgServer) AggregateExchangeRateVote(goCtx context.Context, msg *types.MsgAggregateExchangeRateVote) (*types.MsgAggregateExchangeRateVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valAddr, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}

	feederAddr, err := sdk.AccAddressFromBech32(msg.Feeder)
	if err != nil {
		return nil, err
	}

	if err := ms.ValidateFeeder(ctx, feederAddr, valAddr); err != nil {
		return nil, err
	}

	aggregatePrevote, err := ms.GetAggregateExchangeRatePrevote(ctx, valAddr)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrNoAggregatePrevote, msg.Validator)
	}

	// at this point we have a guarantee that currentVotePeriodEpochCounter is initialized since there has to be
	// a prevote to be able to vote

	// Check a msg is submitted proper period
	currentEpochCounter := ms.epochKeeper.GetEpochInfo(ctx, ms.GetParams(ctx).VotePeriodEpochIdentifier).CurrentEpoch
	if (uint64(currentEpochCounter) - aggregatePrevote.SubmitEpochCounter) != 1 {
		return nil, errorsmod.Wrapf(types.ErrRevealPeriodMissMatch, "expect %d - %d = 1",
			currentEpochCounter, aggregatePrevote.SubmitEpochCounter)
	}

	exchangeRateTuples, err := types.ParseExchangeRateTuples(msg.ExchangeRates)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidCoins, err.Error())
	}

	// check all denoms are in the vote target
	for _, tuple := range exchangeRateTuples {
		if !ms.IsVoteTarget(ctx, tuple.Denom) {
			return nil, errorsmod.Wrap(types.ErrUnknownDenom, tuple.Denom)
		}
	}

	// Verify a exchange rate with aggregate prevote hash
	hash := types.GetAggregateVoteHash(msg.Salt, msg.ExchangeRates, valAddr)
	if aggregatePrevote.Hash != hash.String() {
		return nil, errorsmod.Wrapf(types.ErrVerificationFailed, "must be given %s not %s", aggregatePrevote.Hash, hash)
	}

	// Move aggregate prevote to aggregate vote with given exchange rates
	ms.SetAggregateExchangeRateVote(ctx, valAddr, types.NewAggregateExchangeRateVote(exchangeRateTuples, valAddr))
	ms.DeleteAggregateExchangeRatePrevote(ctx, valAddr)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeAggregateVote,
			sdk.NewAttribute(types.AttributeKeyVoter, msg.Validator),
			sdk.NewAttribute(types.AttributeKeyExchangeRates, msg.ExchangeRates),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Feeder),
		),
	})

	return &types.MsgAggregateExchangeRateVoteResponse{}, nil
}

func (ms msgServer) DelegateFeedConsent(goCtx context.Context, msg *types.MsgDelegateFeedConsent) (*types.MsgDelegateFeedConsentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAddr, err := sdk.ValAddressFromBech32(msg.Operator)
	if err != nil {
		return nil, err
	}

	delegateAddr, err := sdk.AccAddressFromBech32(msg.Delegate)
	if err != nil {
		return nil, err
	}

	// Check the delegator is a validator
	_, err = ms.StakingKeeper.GetValidator(ctx, operatorAddr)
	if err != nil {
		return nil, errorsmod.Wrap(stakingtypes.ErrNoValidatorFound, msg.Operator)
	}

	// Set the delegation
	ms.SetFeederDelegation(ctx, operatorAddr, delegateAddr)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeFeedDelegate,
			sdk.NewAttribute(types.AttributeKeyFeeder, msg.Delegate),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Operator),
		),
	})

	return &types.MsgDelegateFeedConsentResponse{}, nil
}
