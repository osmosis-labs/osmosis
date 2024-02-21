package post

import (
	"fmt"
	"strconv"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/osmosis-labs/osmosis/v23/x/authenticator/authenticator"
	authenticatorkeeper "github.com/osmosis-labs/osmosis/v23/x/authenticator/keeper"
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/utils"
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

	// collect all the contract address that has been updated from transient store
	transientStoreUpdatedContracts := make(map[string]struct{})

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
			authenticationRequest.AuthenticatorId = strconv.FormatUint(accountAuthenticator.Id, 10)

			a := accountAuthenticator.AsAuthenticator(ad.authenticatorKeeper.AuthenticatorManager)

			// If the authenticator is a cosmwasm authenticator, we need to state from the transient store
			// to the contract state
			//
			// TODO: Note that this is a temporary solution. There are issues with the current design:
			// - This will overwrite `runMsgs` changes to the contract state
			// - Any other contract that is used by this contract on `Track` will not be updated
			//   since we only sync the cosmwasm authenticator contract state
			cosmwasmAuthenticator, ok := a.(authenticator.CosmwasmAuthenticator)
			contractAddr := cosmwasmAuthenticator.ContractAddress().String()
			_, isUpdated := transientStoreUpdatedContracts[contractAddr]

			if ok && !isUpdated {
				// sync the transient store state to the committing contract state
				ad.authenticatorKeeper.TransientStore.WriteCosmWasmAuthenticatorStateInto(ctx, &cosmwasmAuthenticator)

				// mark contract as updated
				transientStoreUpdatedContracts[contractAddr] = struct{}{}
			}

			// Confirm Execution
			res := a.ConfirmExecution(ctx, authenticationRequest)

			if res.IsBlock() {
				err = errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "execution blocked by authenticator (account = %s, id = %d)", account, accountAuthenticator.Id)
				err = errorsmod.Wrap(err, fmt.Sprintf("%s", res.Error()))
				return sdk.Context{}, err
			}

			success = res.IsConfirm()
		}
	}

	// All non-ready authenticators should be ready now
	for key := range nonReadyAccountAuthenticatorKeys {
		ad.authenticatorKeeper.MarkAuthenticatorAsReady(ctx, []byte(key))
	}

	return next(ctx, tx, simulate, success)
}
