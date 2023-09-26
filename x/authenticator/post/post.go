package post

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authenticatorkeeper "github.com/osmosis-labs/osmosis/v19/x/authenticator/keeper"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/utils"
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

// AnteHandle is the authenticator post handler
func (ad AuthenticatorDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// If this is getting called, all messages succeeded. We can now update the
	// state of the authenticators. If a post handler returns an error, then
	// all state changes are reverted anyway
	ad.authenticatorKeeper.TransientStore.Write(ctx)

	for msgIndex, msg := range tx.GetMsgs() {
		account, err := utils.GetAccount(msg)
		if err != nil {
			return sdk.Context{}, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "unable to get account")
		}
		authenticators, err := ad.authenticatorKeeper.GetAuthenticatorsForAccountOrDefault(ctx, account)
		if err != nil {
			return sdk.Context{}, err
		}

		for _, authenticator := range authenticators { // This should execute on *all* authenticators so they can update their state
			// Get the authentication data for the transaction
			authData, err := authenticator.GetAuthenticationData(ctx, tx, msgIndex, simulate)
			if err != nil {
				return ctx, err
			}

			// Confirm Execution
			success := authenticator.ConfirmExecution(ctx, account, msg, authData)

			if success.IsBlock() {
				return sdk.Context{}, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "authenticator failed to confirm execution")
			}
		}
	}

	return next(ctx, tx, simulate)
}
