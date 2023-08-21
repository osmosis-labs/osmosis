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

	// Create the signature verification authenticator
	sigVerificationAuthenticator := authenticatortypes.NewSigVerificationAuthenticator(ad.ak, ad.signModeHandler)

	// Get the signer data from the tx
	authData, err := sigVerificationAuthenticator.GetAuthenticationData(tx, simulate)
	if err != nil {
		return ctx, err
	}

	// Validate the signatures for each transaction in the array
	err = sigVerificationAuthenticator.Authenticate(ctx, authData)
	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}
