package ante

import (
	errorsmod "cosmossdk.io/errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	authenticatortypes "github.com/osmosis-labs/osmosis/v19/x/authenticator/authenticator"
	authenticatorkeeper "github.com/osmosis-labs/osmosis/v19/x/authenticator/keeper"
)

func GetAccount(msg sdk.Msg) (sdk.AccAddress, error) {
	if len(msg.GetSigners()) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "no signers")
	}
	return msg.GetSigners()[0], nil
}

type AuthenticatorDecorator struct {
	authenticatorKeeper *authenticatorkeeper.Keeper
	maxFeePayerGas      uint64
}

func NewAuthenticatorDecorator(
	authenticatorKeeper *authenticatorkeeper.Keeper,
	maxFeePayerGas uint64,
) AuthenticatorDecorator {
	return AuthenticatorDecorator{
		authenticatorKeeper: authenticatorKeeper,
		maxFeePayerGas:      maxFeePayerGas,
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
	// keep track of called authenticators so that they can be notified of failed txs
	calledAuthenticators := make([]callData, 0)

	// Authenticating the fee payer needs to be done with very little gas
	// This is a spam-prevention strategy. This protects from a user adding multiple
	// authenticators that overuse compute.
	//
	// If the fee payer is not authenticated, there will be no one to pay
	// for the cost of the tx, which would allow an attacker to force a
	// validator to spend resources by running authenticators on a tx
	// that will never be executed
	originalGasMeter := ctx.GasMeter()
	// Ideally we would want to use min(gasRemaining, maxFeePayerGas) here, but this leads to problems because
	// of the implementation of the InfiniteGasMeter. I think it's ok to allow potentially going over
	// the original limit as long as it's bellow the fee payer gas limit
	payerGasMeter := sdk.NewGasMeter(ad.maxFeePayerGas)
	ctx = ctx.WithGasMeter(payerGasMeter)

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// The fee payer by default is the first signer of the transaction
	feePayer := feeTx.FeePayer()
	feePayerAuthenticated := false

	// Recover from any OutOfGas panic to return an error with information of the gas limit having been reduced
	defer func() {
		if r := recover(); r != nil {
			if feePayerAuthenticated {
				panic(r)
			}
			switch r.(type) {
			case sdk.ErrorOutOfGas:
				log := fmt.Sprintf(
					"FeePayer not authenticated yet. The gas limit has been reduced to %d. Consumed: %d",
					ad.maxFeePayerGas, payerGasMeter.GasConsumed())
				err = sdkerrors.Wrap(sdkerrors.ErrOutOfGas, log)
			default:
				panic(r)
			}
		}
	}()

	// This is a transient context stored globally throughout the execution of the tx
	// Any changes will to authenticator storage will be written to the store at the end of the tx
	cacheCtx := ad.authenticatorKeeper.TransientStore.GetTransientContext(ctx)

	// Authenticate the accounts of all messages
	for msgIndex, msg := range tx.GetMsgs() {
		// By default, the first signer is the account
		account, err := GetAccount(msg)
		if err != nil {
			return sdk.Context{}, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("failed to get account for message %d", msgIndex))
		}

		// Get all authenticators for the executing account
		// If no authenticators are found, use the default authenticator
		// This is done to keep backwards compatibility by defaulting to a signature verifier on accounts without authenticators
		authenticators, err := ad.authenticatorKeeper.GetAuthenticatorsForAccountOrDefault(cacheCtx, account)
		if err != nil {
			return sdk.Context{}, err
		}

		msgAuthenticated := false
		// TODO: We should consider adding a way for the user to specify which authenticator to
		// use as part of the tx (likely in the signature)
		// NOTE: we have to make sure that doing that does not make the signature malleable
		for _, authenticator := range authenticators {
			// Consume the authenticator's static gas
			cacheCtx.GasMeter().ConsumeGas(authenticator.StaticGas(), "authenticator static gas")

			// Get the authentication data for the transaction
			neverWriteCacheCtx, _ := cacheCtx.CacheContext() // GetAuthenticationData is not allowed to modify the state
			authData, err := authenticator.GetAuthenticationData(neverWriteCacheCtx, tx, int8(msgIndex), simulate)
			if err != nil {
				return ctx, err
			}

			// Keep track of all called authenticators
			calledAuthenticators = append(calledAuthenticators, callData{authenticator: authenticator, authenticatorData: authData, msg: msg})

			authentication := authenticator.Authenticate(cacheCtx, account, msg, authData)
			if authentication.IsRejected() {
				return ctx, authentication.Error()

			}

			if authentication.IsAuthenticated() {
				msgAuthenticated = true
				// Once the fee payer is authenticated, we can set the gas limit to its original value
				if !feePayerAuthenticated && account.Equals(feePayer) {
					originalGasMeter.ConsumeGas(payerGasMeter.GasConsumed(), "fee payer gas")
					// Reset this for both contexts
					cacheCtx = cacheCtx.WithGasMeter(originalGasMeter)
					ctx = ctx.WithGasMeter(originalGasMeter)
					feePayerAuthenticated = true
				}
				break
			}
		}

		// if authentication failed, return an error
		if !msgAuthenticated {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("authentication failed for message %d", msgIndex))
		}
	}

	return next(ctx, tx, simulate)
}
