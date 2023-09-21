package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v19/x/authenticator/authenticator"
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

// AddAuthenticator adds any types of authenticator to an account
func (m msgServer) AddAuthenticator(
	goCtx context.Context,
	msg *types.MsgAddAuthenticator,
) (*types.MsgAddAuthenticatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	auths, err := m.Keeper.GetAuthenticatorsForAccount(ctx, sender)
	if err != nil {
		return nil, err
	}

	// Validate that the data is correct for the type of authenticator
	switch msg.Type {
	case authenticator.SignatureVerificationAuthenticatorType:
		if len(msg.Data) != secp256k1.PubKeySize {
			return nil, fmt.Errorf("invalid secp256k1 pub key size expected %d, got %d", secp256k1.PubKeySize, len(msg.Data))
		}
		// If there are no other authenticators ensure that the first authenticator is associated
		// with the original account
		if len(auths) == 0 {
			pubKey := secp256k1.PubKey{Key: msg.Data}
			senderDataAccount := sdk.AccAddress(pubKey.Address())
			if !senderDataAccount.Equals(sender) {
				return nil, fmt.Errorf("first authenticator must be associated with account expected %s, got %s", sender, senderDataAccount)
			}
		}
	}

	// Limit the number of authenticators to stop over iteration in the ante handler
	if len(auths) >= 15 {
		return nil, fmt.Errorf("max authenticators: %d, tried to add more than the max amount of authenticator", 15)
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
