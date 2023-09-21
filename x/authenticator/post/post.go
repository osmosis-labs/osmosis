package post

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authenticatorkeeper "github.com/osmosis-labs/osmosis/v19/x/authenticator/keeper"
)

type AuthenticatorDecorator struct {
	authenticatorKeeper *authenticatorkeeper.Keeper
}

func NewAuthenticatorDecorator(
	authenticatorKeeper *authenticatorkeeper.Keeper,
) AuthenticatorDecorator {
	return AuthenticatorDecorator{
		authenticatorKeeper: authenticatorKeeper,
	}
}

// AnteHandle is the authenticator ante handler
func (ad AuthenticatorDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	for msgIndex, msg := range tx.GetMsgs() {
		account := msg.GetSigners()[0]
		authenticators, err := ad.authenticatorKeeper.GetAuthenticatorsForAccount(ctx, account)
		if err != nil {
			return sdk.Context{}, err
		}

		if len(authenticators) == 0 {
			authenticators = append(authenticators, ad.authenticatorKeeper.AuthenticatorManager.GetDefaultAuthenticator())
		}
		for _, authenticator := range authenticators { // This should execute on *all* authenticators so they can update their state
			// Get the authentication data for the transaction
			authData, err := authenticator.GetAuthenticationData(ctx, tx, int8(msgIndex), simulate)
			if err != nil {
				return ctx, err
			}

			// Authenticate the message
			success := authenticator.ConfirmExecution(ctx, account, msg, authData)

			if success.IsBlock() {
				return sdk.Context{}, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "authenticator failed to confirm execution")
			}
		}
	}

	ad.authenticatorKeeper.TransientStore.Write(ctx)
	return next(ctx, tx, simulate)
}
