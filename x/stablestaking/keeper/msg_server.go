package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v27/x/stablestaking/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the stablestaking MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) StakeTokens(goCtx context.Context, msg *types.MsgStakeTokens) (*types.MsgStakeTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		return nil, err
	}

	return m.handleStakeTokensRequest(ctx, addr, msg.Amount)
}

func (m msgServer) UnstakeTokens(goCtx context.Context, msg *types.MsgUnstakeTokens) (*types.MsgUnstakeTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		return nil, err
	}

	return m.handleUnStakeTokensRequest(ctx, addr, msg.Amount)
}

func (m msgServer) handleStakeTokensRequest(ctx sdk.Context, staker sdk.AccAddress, amount sdk.Coin) (*types.MsgStakeTokensResponse, error) {
	resp, err := m.Keeper.StakeTokens(ctx, staker, amount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventStake,
			sdk.NewAttribute(types.AttributeKeyStaker, staker.String()),
			sdk.NewAttribute(types.AttributeKeyStakeCoin, amount.Denom),
			sdk.NewAttribute(types.AttributeKeyStakeAmount, amount.Amount.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})
	return resp, nil
}

func (m msgServer) handleUnStakeTokensRequest(ctx sdk.Context, staker sdk.AccAddress, amount sdk.Coin) (*types.MsgUnstakeTokensResponse, error) {
	resp, err := m.Keeper.UnStakeTokens(ctx, staker, amount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventUnstake,
			sdk.NewAttribute(types.AttributeKeyStaker, staker.String()),
			sdk.NewAttribute(types.AttributeKeyUnStakeCoin, amount.Denom),
			sdk.NewAttribute(types.AttributeKeyUnStakeAmount, amount.Amount.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})
	return resp, nil
}
