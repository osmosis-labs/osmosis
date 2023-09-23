package ante

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	authenticatorkeeper "github.com/osmosis-labs/osmosis/v19/x/authenticator/keeper"
)

// GetAccount retrieves the account associated with the first signer of a transaction message.
// It returns the account's address or an error if no signers are present.
func GetAccount(msg sdk.Msg) (sdk.AccAddress, error) {
	if len(msg.GetSigners()) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "no signers")
	}
	return msg.GetSigners()[0], nil
}

// AuthenticatorDecorator is responsible for processing authentication logic
// before transaction execution.
type AuthenticatorDecorator struct {
	authenticatorKeeper *authenticatorkeeper.Keeper
	maxFeePayerGas      uint64
}

// NewAuthenticatorDecorator creates a new instance of AuthenticatorDecorator with the provided parameters.
func NewAuthenticatorDecorator(
	authenticatorKeeper *authenticatorkeeper.Keeper,
	maxFeePayerGas uint64,
) AuthenticatorDecorator {
	return AuthenticatorDecorator{
		authenticatorKeeper: authenticatorKeeper,
		maxFeePayerGas:      maxFeePayerGas,
	}
}

// AnteHandle is the authenticator ante handler responsible for processing authentication
// logic before transaction execution.
func (ad AuthenticatorDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// Performing fee payer authentication with minimal gas allocation
	// serves as a spam-prevention strategy to prevent users from adding multiple
	// authenticators that may excessively consume computational resources.
	// If the fee payer is not authenticated, there would be no entity responsible
	// for covering the transaction's costs. This safeguard ensures that validators
	// are not compelled to expend resources on executing authenticators for transactions
	// that will never be executed.

	// Ideally, we would prefer to use min(gasRemaining, maxFeePayerGas) here, but
	// this approach presents challenges due to the implementation of the InfiniteGasMeter.
	// As long as the gas consumption remains below the fee payer gas limit, exceeding
	// the original limit should be acceptable.
	originalGasMeter := ctx.GasMeter()
	payerGasMeter := sdk.NewGasMeter(ad.maxFeePayerGas)
	ctx = ctx.WithGasMeter(payerGasMeter)

	// Recover from any OutOfGas panic to return an error with information on the reduced gas limit
	defer func() {
		if r := recover(); r != nil {
			// Ensure that the maximum fee payer gas exceeds the gas consumed by the fee payer
			if ad.maxFeePayerGas > payerGasMeter.GasConsumed() {
				panic(r)
			}
			switch r.(type) {
			case sdk.ErrorOutOfGas:
				log := fmt.Sprintf(
					"FeePayer authentication pending. Gas limit reduced to %d. Gas Consumed: %d",
					ad.maxFeePayerGas, payerGasMeter.GasConsumed())
				err = sdkerrors.Wrap(sdkerrors.ErrOutOfGas, log)
			default:
				panic(r)
			}
		}
	}()

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "transaction must be a FeeTx")
	}

	// Determine the fee payer, with the first signer of the transaction as the default.
	feePayer := feeTx.FeePayer()

	// Create a transient context that is globally available during transaction execution.
	// Any changes made to authenticator storage will be written to the store at the end of the transaction.
	cacheCtx := ad.authenticatorKeeper.TransientStore.ResetTransientContext(ctx)

	// Always authenticate the fee payer first to enable gas to be paid.
	// The maximum number of authenticators is limited to 15 in the msg_server.
	feePayerIndex := -1
	for msgIndex, msg := range tx.GetMsgs() {
		account, err := GetAccount(msg)
		if err != nil {
			return sdk.Context{}, sdkerrors.Wrap(sdkerrors.ErrUnauthorized,
				fmt.Sprintf("failed to retrieve the account for message %d", msgIndex))
		}
		if account.Equals(feePayer) {
			feePayerIndex = msgIndex
			cacheCtx, err = ad.AuthTx(cacheCtx, tx, msg, msgIndex, simulate)
			if err != nil {
				return ctx, err
			}

			// Consume gas for fee payer authentication
			originalGasMeter.ConsumeGas(payerGasMeter.GasConsumed(), "fee payer gas consumed, continuing")

			// Reset gas meters for both contexts
			cacheCtx = cacheCtx.WithGasMeter(originalGasMeter)
			ctx = ctx.WithGasMeter(originalGasMeter)
		}
	}

	// This scenario is unlikely but accounted for to optimize computation
	if feePayerIndex == -1 {
		return sdk.Context{}, sdkerrors.Wrap(sdkerrors.ErrUnauthorized,
			fmt.Sprintf("failed to retrieve the account for fee payer %d", feePayer))
	}

	// Authenticate the accounts of all messages.
	for msgIndex, msg := range tx.GetMsgs() {
		// Skip fee payer authentication as it has already been completed.
		if feePayerIndex == msgIndex {
			continue
		}

		// After fee payer authentication, proceed to authenticate the remaining messages.
		_, err := ad.AuthTx(cacheCtx, tx, msg, msgIndex, simulate)
		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

// AuthTx performs the authentication process for a specific message within a transaction.
func (ad AuthenticatorDecorator) AuthTx(
	cacheCtx sdk.Context,
	tx sdk.Tx,
	msg sdk.Msg,
	msgIndex int,
	simulate bool,
) (ctx sdk.Context, err error) {
	account, err := GetAccount(msg)
	if err != nil {
		return sdk.Context{}, sdkerrors.Wrap(sdkerrors.ErrUnauthorized,
			fmt.Sprintf("failed to retrieve the account for message %d", msgIndex))
	}

	// Retrieve all authenticators for the executing account.
	// If no authenticators are found, the default authenticator is used.
	// This ensures backward compatibility by defaulting to a signature verifier
	// for accounts without authenticators.
	authenticators, err := ad.authenticatorKeeper.GetAuthenticatorsForAccountOrDefault(cacheCtx, account)
	if err != nil {
		return sdk.Context{}, err
	}

	msgAuthenticated := false
	// TODO: Consider adding a way for users to specify which authenticator to use as part of the transaction.
	// NOTE: Care must be taken to ensure that doing so does not make the signature malleable.
	for _, authenticator := range authenticators {
		// Consume the authenticator's static gas.
		cacheCtx.GasMeter().ConsumeGas(authenticator.StaticGas(), "authenticator static gas")

		// Retrieve authentication data for the transaction.
		neverWriteCacheCtx, _ := cacheCtx.CacheContext() // GetAuthenticationData should not modify the state.
		authData, err := authenticator.GetAuthenticationData(neverWriteCacheCtx, tx, int8(msgIndex), simulate)
		if err != nil {
			return ctx, err
		}

		authentication := authenticator.Authenticate(cacheCtx, account, msg, authData)
		if authentication.IsRejected() {
			return ctx, authentication.Error()
		}

		if authentication.IsAuthenticated() {
			msgAuthenticated = true
			// We break out of the authenticator loop if one of the authenticators have authenticated the transaction
			break
		}
	}

	// If authentication fails, return an error.
	if !msgAuthenticated {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("authentication failed for message %d", msgIndex))
	}

	return cacheCtx, nil
}
