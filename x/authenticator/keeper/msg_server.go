package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/types"
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

func (m msgServer) AddAuthenticator(
	goCtx context.Context,
	msg *types.MsgAddAuthenticator,
) (*types.MsgAddAuthenticatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: validate based on type of authenticator
	// TODO: do we want to create and account for the data/pubkey
	// here? Or in ante handler

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	err = m.Keeper.AddAuthenticator(ctx, sender, msg.Type, msg.Data)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyAuthenticatorType, msg.Type),
		),
	})

	return &types.MsgAddAuthenticatorResponse{
		Success: true,
	}, nil
}

func (m msgServer) RemoveAuthenticator(goCtx context.Context, msg *types.MsgRemoveAuthenticator) (*types.MsgRemoveAuthenticatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	err = m.Keeper.RemoveAuthenticator(ctx, sender, msg.Id)
	if err != nil {
		return nil, err
	}

	return &types.MsgRemoveAuthenticatorResponse{
		Success: true,
	}, nil
}
