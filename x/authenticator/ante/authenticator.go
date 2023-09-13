package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authenticatortypes "github.com/osmosis-labs/osmosis/v19/x/authenticator/authenticator"
	authenticatorkeeper "github.com/osmosis-labs/osmosis/v19/x/authenticator/keeper"
)

type DefaultAccountGetter struct{}

func (DefaultAccountGetter) GetAccount(ctx sdk.Context, msg sdk.Msg, tx sdk.Tx) (sdk.AccAddress, error) {
	if len(msg.GetSigners()) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "no signers")
	}
	return msg.GetSigners()[0], nil
}

var _ authenticatortypes.AccountGetter = DefaultAccountGetter{}

type AuthenticatorDecorator struct {
	authenticatorKeeper *authenticatorkeeper.Keeper
	accountGetter       authenticatortypes.AccountGetter
}

func NewAuthenticatorDecorator(
	authenticatorKeeper *authenticatorkeeper.Keeper,
) AuthenticatorDecorator {
	return AuthenticatorDecorator{
		authenticatorKeeper: authenticatorKeeper,
		accountGetter:       DefaultAccountGetter{},
	}
}

type callData struct {
	authenticator     authenticatortypes.Authenticator
	authenticatorData authenticatortypes.AuthenticatorData
	msg               sdk.Msg
}

// AnteHandle is the authenticator ante handler
func (ad AuthenticatorDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// keep track of called authenticators so they can be notified of failed txs
	calledAuthenticators := make([]callData, 0)

	// Authenticate the accounts of all messages
	for msgIndex, msg := range tx.GetMsgs() {
		// By default, the first signer is the account
		account, err := ad.accountGetter.GetAccount(ctx, msg, tx)
		if err != nil {
			return sdk.Context{}, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("failed to get account for message %d", msgIndex))
		}

		// Get all authenticators for the executing account
		authenticators, err := ad.authenticatorKeeper.GetAuthenticatorsForAccount(ctx, account)
		if err != nil {
			return sdk.Context{}, err
		}

		// If no authenticators are found, use the default authenticator
		// This is done to keep backwards compatibility by defaulting to a signature verifier on accounts without authenticators
		if len(authenticators) == 0 {
			authenticators = append(authenticators, ad.authenticatorKeeper.AuthenticatorManager.GetDefaultAuthenticator())
		}

		msgAuthenticated := false
		// TODO: We should consider adding a way for the user to specify which authenticator to
		// use as part of the tx (likely in the signature)
		// NOTE: we have to make sure that doing that does not make the signature malleable
		for _, authenticator := range authenticators {
			// Get the authentication data for the transaction
			cacheCtx, _ := ctx.CacheContext() // GetAuthenticationData is not allowed to modify the state
			authData, err := authenticator.GetAuthenticationData(cacheCtx, tx, int8(msgIndex), simulate)
			if err != nil {
				return ctx, err
			}

			// Authenticate the message
			calledAuthenticators = append(calledAuthenticators, callData{authenticator: authenticator, authenticatorData: authData, msg: msg})
			authenticated, err := authenticator.Authenticate(ctx, msg, authData)
			if err != nil {
				// TODO: Check this assumption. We want authenticators to return true/false to authenticate or not,
				//       but we also want them to be able to return an error and fully block the tx in that case
				return ctx, err
			}

			if authenticated {
				msgAuthenticated = true
				break
			}
		}

		// if authentation failed, allow reverting of state
		if !msgAuthenticated {
			for _, callData := range calledAuthenticators {
				callData.authenticator.AuthenticationFailed(ctx, callData.authenticatorData, callData.msg)
			}
			return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("authentication failed for message %d", msgIndex))
		}
	}
	return next(ctx, tx, simulate)
}
