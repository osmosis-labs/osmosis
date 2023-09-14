package ante

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

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
	maxFeePayerGas      uint64
}

func NewAuthenticatorDecorator(
	authenticatorKeeper *authenticatorkeeper.Keeper,
	maxFeePayerGas uint64,
) AuthenticatorDecorator {
	return AuthenticatorDecorator{
		authenticatorKeeper: authenticatorKeeper,
		accountGetter:       DefaultAccountGetter{},
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

	// Authenticate the fee payer

	// Authenticating the fee payer needs to be done with very little gas
	// This is a spam-prevention strategy. If the fee payer is not authenticated, there will be no one to pay
	// for the cost of the tx, which would allow an attacker to force a validator to spend resources by running
	// authenticators on a tx that will never be executed
	originalGasMeter := ctx.GasMeter()
	// TODO: Here we actually want to use min(gasRemaining, maxFeePayerGas), but this may leat to problems because
	//   of the implementation of the InfiniteGasMeter. I think it's ok to allow an overflow here as long as it's
	//   bellow the fee payer gas limit
	payerGasMeter := sdk.NewGasMeter(ad.maxFeePayerGas)
	ctx = ctx.WithGasMeter(payerGasMeter)

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

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

	// Authenticate the accounts of all messages
	for msgIndex, msg := range tx.GetMsgs() {
		// By default, the first signer is the account
		account, err := ad.accountGetter.GetAccount(ctx, msg, tx)
		if err != nil {
			return sdk.Context{}, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("failed to get account for message %d", msgIndex))
		}

		// Get all authenticators for the executing account
		// If no authenticators are found, use the default authenticator
		// This is done to keep backwards compatibility by defaulting to a signature verifier on accounts without authenticators
		authenticators, err := ad.authenticatorKeeper.GetAuthenticatorsForAccountOrDefault(ctx, account)
		if err != nil {
			return sdk.Context{}, err
		}

		msgAuthenticated := false
		// TODO: We should consider adding a way for the user to specify which authenticator to
		// use as part of the tx (likely in the signature)
		// NOTE: we have to make sure that doing that does not make the signature malleable
		for _, authenticator := range authenticators {
			// Consume the authenticator's static gas
			ctx.GasMeter().ConsumeGas(authenticator.StaticGas(), "authenticator static gas")

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
				// Once the fee payer is authenticated, we can set the gas limit to its original value
				if !feePayerAuthenticated && account.Equals(feePayer) {
					originalGasMeter.ConsumeGas(payerGasMeter.GasConsumed(), "fee payer gas")
					ctx = ctx.WithGasMeter(originalGasMeter)
					feePayerAuthenticated = true
				}
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
