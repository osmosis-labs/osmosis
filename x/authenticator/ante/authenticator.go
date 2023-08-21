package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	authenticatortypes "github.com/osmosis-labs/osmosis/v17/x/authenticator/types"
)

// Verify all signatures for a tx and return an error if any are invalid. Note,
// the AuthenticatorDecorator will not check signatures on ReCheck.
//
// CONTRACT: Pubkeys are set in context for all signers before this decorator runs
// CONTRACT: Tx must implement SigVerifiableTx interface
type AuthenticatorDecorator struct {
	ak              authante.AccountKeeper
	signModeHandler authsigning.SignModeHandler
}

func NewAuthenticatorDecorator(
	ak authante.AccountKeeper,
	signModeHandler authsigning.SignModeHandler,
) AuthenticatorDecorator {
	return AuthenticatorDecorator{
		ak:              ak,
		signModeHandler: signModeHandler,
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

	for i, msg := range tx.GetMsgs() {

		// Todo: Replace getting the authenticator for something like this:
		//ad.authenticatorKeeper.GetAuthenticatorsForAccount(msg.GetSigners()[0])  // ToDo: How do we deal with multiple signers?
		authenticator := authenticatortypes.NewSigVerificationAuthenticator(ad.ak, ad.signModeHandler)

		// Get the authentication data for the transaction
		authData, err := authenticator.GetAuthenticationData(tx, uint8(i), simulate)
		if err != nil {
			return ctx, err
		}

		// Authenticate the message
		_, err = authenticator.Authenticate(ctx, msg, authData)
		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}
