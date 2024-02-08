package post

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	authenticatorkeeper "github.com/osmosis-labs/osmosis/v21/x/authenticator/keeper"
	"github.com/osmosis-labs/osmosis/v21/x/authenticator/types"
	"github.com/osmosis-labs/osmosis/v21/x/authenticator/utils"
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

		for _, accountAuthenticator := range accountAuthenticators { // This should execute on *all* readied authenticators so they can update their state

			// We want to skip `ConfirmExecution` if the authenticator is newly added
			// since Authenticate & Track are called on antehandler but newly added authenticator
			// so Authenticate & Track on newly added authenticator has not been called yet
			// which means the authenticator is not ready to confirm execution
			if !accountAuthenticator.IsReady {
				key := string(types.KeyAccountId(account, accountAuthenticator.Id))
				nonReadyAccountAuthenticatorKeys[key] = struct{}{}
				continue
			}

			authenticator := accountAuthenticator.AsAuthenticator(ad.authenticatorKeeper.AuthenticatorManager)

			// Get the authentication data for the transaction
			authData, err := authenticator.GetAuthenticationData(ctx, tx, msgIndex, simulate)
			if err != nil {
				return ctx, err
			}

			// Confirm Execution
			successfulExecution := authenticator.ConfirmExecution(ctx, account, msg, authData)

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
