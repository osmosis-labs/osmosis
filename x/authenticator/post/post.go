package post

import (
	"fmt"
	"strconv"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	authenticatorante "github.com/osmosis-labs/osmosis/v23/x/authenticator/ante"
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/authenticator"
	authenticatorkeeper "github.com/osmosis-labs/osmosis/v23/x/authenticator/keeper"
)

type AuthenticatorDecorator struct {
	authenticatorKeeper *authenticatorkeeper.Keeper
	accountKeeper       *authkeeper.AccountKeeper
	sigModeHandler      authsigning.SignModeHandler
	next                sdk.PostHandler
}

func NewAuthenticatorDecorator(
	authenticatorKeeper *authenticatorkeeper.Keeper,
	accountKeeper *authkeeper.AccountKeeper,
	sigModeHandler authsigning.SignModeHandler,
	next sdk.PostHandler,
) AuthenticatorDecorator {
	return AuthenticatorDecorator{
		authenticatorKeeper: authenticatorKeeper,
		accountKeeper:       accountKeeper,
		sigModeHandler:      sigModeHandler,
		next:                next,
	}
}

// AnteHandle is the authenticator post handler
func (ad AuthenticatorDecorator) PostHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate,
	success bool,
	next sdk.PostHandler,
) (newCtx sdk.Context, err error) {
	// Ensure that the transaction is a authenticator transaction
	active, txOptions := authenticatorante.IsCircuitBreakActive(ctx, tx, ad.authenticatorKeeper)
	if active {
		return ad.next(ctx, tx, simulate, success)
	}

	// Retrieve the selected authenticators from the extension.
	selectedAuthenticatorsFromExtension := txOptions.GetSelectedAuthenticators()

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		// This should never happen
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// The fee payer by default is the first signer of the transaction
	feePayer := feeTx.FeePayer()

	for msgIndex, msg := range tx.GetMsgs() {
		// When using a smart account we enforce one signer per transaction in the AnteHandler, if this is updated changes
		// need to be reflected here
		account := msg.GetSigners()[0]

		selectedAuthenticator, err := ad.authenticatorKeeper.GetInitializedAuthenticatorForAccount(
			ctx,
			account,
			int(selectedAuthenticatorsFromExtension[msgIndex]),
		)
		if err != nil {
			return sdk.Context{}, err
		}

		// We skip replay protection here as it was already checked on authenticate.
		// TODO: We probably want to avoid calling this function again. Can we keep this in cache? maybe in transient store?
		authenticationRequest, err := authenticator.GenerateAuthenticationData(
			ctx,
			ad.accountKeeper,
			ad.sigModeHandler,
			account,
			feePayer,
			msg,
			tx,
			msgIndex,
			simulate,
			authenticator.NoReplayProtection,
		)
		if err != nil {
			return sdk.Context{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized,
				fmt.Sprintf("failed to get authentication data for message %d", msgIndex))
		}

		authenticationRequest.AuthenticatorId = strconv.FormatUint(selectedAuthenticator.Id, 10)

		// Confirm Execution
		err = selectedAuthenticator.Authenticator.ConfirmExecution(ctx, authenticationRequest)
		if err != nil {
			err = errorsmod.Wrapf(err, "execution blocked by authenticator (account = %s, id = %d)", account, selectedAuthenticator.Id)
			err = errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s", err)
			return sdk.Context{}, err
		}

		success = err == nil
	}

	return next(ctx, tx, simulate, success)
}
