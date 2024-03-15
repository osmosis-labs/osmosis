package keeper

import (
	"context"
	"fmt"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/authenticator/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// AddAuthenticator allows the addition of various types of authenticators to an account.
// This method serves as a versatile function for adding diverse authenticator types
// to an account, making it highly adaptable for different use cases.
func (m msgServer) AddAuthenticator(
	goCtx context.Context,
	msg *types.MsgAddAuthenticator,
) (*types.MsgAddAuthenticatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid sender address")
	}

	authenticators, err := m.Keeper.GetAuthenticatorDataForAccount(ctx, sender)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "failed to get authenticators for account %s", sender)
	}

	// Limit the number of authenticators to prevent excessive iteration in the ante handler.
	if len(authenticators) >= 15 {
		return nil, fmt.Errorf("maximum authenticators reached (%d), attempting to add more than the maximum allowed", 15)
	}

	// Finally, add the authenticator to the store.
	id, err := m.Keeper.AddAuthenticator(ctx, sender, msg.Type, msg.Data)
	if err != nil {
		return nil, err
	}

	stringId := strconv.FormatUint(id, 10)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyAuthenticatorType, msg.Type),
			sdk.NewAttribute(types.AttributeKeyAuthenticatorId, stringId),
		),
	})

	return &types.MsgAddAuthenticatorResponse{
		Success: true,
	}, nil
}

// RemoveAuthenticator removes an authenticator from the store. The message specifies a sender address and an index.
func (m msgServer) RemoveAuthenticator(goCtx context.Context, msg *types.MsgRemoveAuthenticator) (*types.MsgRemoveAuthenticatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid sender address")
	}

	// At this point, we assume that verification has occurred on the account, and we
	// proceed to remove the authenticator from the store.
	err = m.Keeper.RemoveAuthenticator(ctx, sender, msg.Id)
	if err != nil {
		return nil, err
	}

	return &types.MsgRemoveAuthenticatorResponse{
		Success: true,
	}, nil
}

func (m msgServer) SetActiveState(goCtx context.Context, msg *types.MsgSetActiveState) (*types.MsgSetActiveStateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the account address from the message
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid sender address")
	}

	// TODO: check if sender is one of circuit breaker controller accounts
	_ = sender

	// Set the active state of the authenticator
	err = m.Keeper.SetActiveState(ctx, msg.Active)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AtrributeKeyAuthenticatorActiveState, strconv.FormatBool(msg.Active)),
		),
	})

	return &types.MsgSetActiveStateResponse{}, nil
}
