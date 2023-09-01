package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authenticatorkeeper "github.com/osmosis-labs/osmosis/v19/x/authenticator/keeper"
	authenticatortypes "github.com/osmosis-labs/osmosis/v19/x/authenticator/types"
)

// Verify all signatures for a tx and return an error if any are invalid. Note,
// the AuthenticatorDecorator will not check signatures on ReCheck.
//
// CONTRACT: Pubkeys are set in context for all signers before this decorator runs
// CONTRACT: Tx must implement SigVerifiableTx interface
type AuthenticatorDecorator struct {
	accountKeeper       authante.AccountKeeper
	authenticatorKeeper *authenticatorkeeper.Keeper
	signModeHandler     authsigning.SignModeHandler
}

func NewAuthenticatorDecorator(
	ak authante.AccountKeeper,
	authenticatorKeeper *authenticatorkeeper.Keeper,
	signModeHandler authsigning.SignModeHandler,
) AuthenticatorDecorator {
	return AuthenticatorDecorator{
		accountKeeper:       ak,
		authenticatorKeeper: authenticatorKeeper,
		signModeHandler:     signModeHandler,
	}
}

// AnteHandle is the authenticator decorator ante handler
// this is used to validate multiple signatures
func (ad AuthenticatorDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {

	for msgIndex, msg := range tx.GetMsgs() {

		// Todo: Replace getting the authenticator for something like this:
		authenticators, err := ad.authenticatorKeeper.GetAuthenticatorsForAccount(ctx, msg.GetSigners()[0])
		if err != nil {
			return sdk.Context{}, err
		}

		if len(authenticators) == 0 {
			authenticators = append(authenticators, authenticatortypes.GetDefaultAuthenticator())
		}

		// ToDo: Add a way for the user to specify which authenticator to use as part of the tx (likely in the signature)
		// Note: we have to make sure that doing that does not make the signature malleable

		for _, authenticator := range authenticators {
			// Get the authentication data for the transaction
			authData, err := authenticator.GetAuthenticationData(tx, uint8(msgIndex), simulate)
			if err != nil {
				return ctx, err
			}

			// Authenticate the message
			authenticated, err := authenticator.Authenticate(ctx, msg, authData)
			if err != nil {
				return ctx, err
			}

			if authenticated {
				return next(ctx, tx, simulate)
			}
		}
	}
	return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "authentication failed")

}
