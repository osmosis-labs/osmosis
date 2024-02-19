package post

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/osmosis-labs/osmosis/v21/x/authenticator/authenticator"

	authenticatorkeeper "github.com/osmosis-labs/osmosis/v21/x/authenticator/keeper"
	"github.com/osmosis-labs/osmosis/v21/x/authenticator/utils"
)

type AuthenticatorDecorator struct {
	authenticatorKeeper *authenticatorkeeper.Keeper
	accountKeeper       *authkeeper.AccountKeeper
	sigModeHandler      authsigning.SignModeHandler
}

func NewAuthenticatorDecorator(
	authenticatorKeeper *authenticatorkeeper.Keeper,
	accountKeeper *authkeeper.AccountKeeper,
	sigModeHandler authsigning.SignModeHandler,
) AuthenticatorDecorator {
	return AuthenticatorDecorator{
		authenticatorKeeper: authenticatorKeeper,
		accountKeeper:       accountKeeper,
		sigModeHandler:      sigModeHandler,
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
	// If this is getting called, all messages succeeded. We can now update the
	// state of the authenticators. If a post handler returns an error, then
	// all state changes are reverted anyway
	ad.authenticatorKeeper.TransientStore.WriteInto(ctx)
	usedAuthenticators := ad.authenticatorKeeper.TransientStore.GetUsedAuthenticators()

	// collect all the keys for authenticators that are not ready
	nonReadyAccountAuthenticatorKeys := make(map[string]struct{})

	for msgIndex, msg := range tx.GetMsgs() {
		account, err := utils.GetAccount(msg)
		if err != nil {
			return sdk.Context{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "unable to get account")
		}

		accountAuthenticators, err := ad.authenticatorKeeper.GetAuthenticatorDataForAccount(ctx, account)
		if err != nil {
			return sdk.Context{}, err
		}

		// We skip replay protection here as it was already checked on authenticate.
		// TODO: We probably want to avoid calling this function again. Can we keep this in cache? maybe in transient store?
		authenticationRequest, err := authenticator.GenerateAuthenticationData(ctx, ad.accountKeeper, ad.sigModeHandler, account, msg, tx, msgIndex, simulate, authenticator.NoReplayProtection)
		if err != nil {
			return sdk.Context{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("failed to get authentication data for message %d", msgIndex))
		}
		for _, accountAuthenticator := range accountAuthenticators { // This should execute on the authenticators used to authenticate the msg
			if usedAuthenticators[msgIndex] != accountAuthenticator.Id {
				continue
			}

			authenticator := accountAuthenticator.AsAuthenticator(ad.authenticatorKeeper.AuthenticatorManager)

			// Confirm Execution
			successfulExecution := authenticator.ConfirmExecution(ctx, authenticationRequest)

			if successfulExecution.IsBlock() {
				return sdk.Context{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "authenticator failed to confirm execution")
			}

			success = successfulExecution.IsConfirm()
		}
	}

	// All non-ready authenticators should be ready now
	for key := range nonReadyAccountAuthenticatorKeys {
		ad.authenticatorKeeper.MarkAuthenticatorAsReady(ctx, []byte(key))
	}

	return next(ctx, tx, simulate, success)
}
