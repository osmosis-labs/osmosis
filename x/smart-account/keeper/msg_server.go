package keeper

import (
	"context"
	"strconv"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/types"
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

	isSmartAccountActive := m.GetIsSmartAccountActive(ctx)
	if !isSmartAccountActive {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "smartaccount module is not active")
	}

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid sender address")
	}

	// Finally, add the authenticator to the store.
	id, err := m.Keeper.AddAuthenticator(ctx, sender, msg.AuthenticatorType, msg.Data)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyAuthenticatorType, msg.AuthenticatorType),
			sdk.NewAttribute(types.AttributeKeyAuthenticatorId, strconv.FormatUint(id, 10)),
		),
	})

	return &types.MsgAddAuthenticatorResponse{
		Success: true,
	}, nil
}

// RemoveAuthenticator removes an authenticator from the store. The message specifies a sender address and an index.
func (m msgServer) RemoveAuthenticator(goCtx context.Context, msg *types.MsgRemoveAuthenticator) (*types.MsgRemoveAuthenticatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	isSmartAccountActive := m.GetIsSmartAccountActive(ctx)
	if !isSmartAccountActive {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "smartaccount module is not active")
	}

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

	// `MsgSetActiveState` must have only one signer
	signer := msg.GetSigners()[0]

	if msg.Active {
		// Only the circuit breaker governor can set the active state of the authenticator to true
		if !signer.Equals(m.Keeper.CircuitBreakerGovernor) {
			return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "signer is not the circuit breaker governor")
		}
	} else {
		// The circuit breaker can state can be set to false by any of the circuit breaker controllers
		isAuthorized := false
		params := m.Keeper.GetParams(ctx)

		for _, controller := range params.CircuitBreakerControllers {
			if controller == signer.String() {
				isAuthorized = true
				break
			}
		}
		if !isAuthorized {
			return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "signer is not a circuit breaker controller")
		}
	}

	// Set the active state of the authenticator
	m.Keeper.SetActiveState(ctx, msg.Active)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AtrributeKeyIsSmartAccountActive, strconv.FormatBool(msg.Active)),
		),
	})

	return &types.MsgSetActiveStateResponse{}, nil
}
