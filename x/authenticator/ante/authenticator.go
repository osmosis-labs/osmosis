package ante

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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

// AnteHandle is the authenticator ante handler responsible for processing authentication
// logic before transaction execution.
func (ad AuthenticatorDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "transaction must be a FeeTx")
	}

	// Determine the fee payer, which defaults to the first signer of the transaction.
	feePayer := feeTx.FeePayer()

	// Create a transient context that is globally available during transaction execution.
	// Any changes made to authenticator storage will be written to the store at the end of the transaction.
	cacheCtx := ad.authenticatorKeeper.TransientStore.ResetTransientContext(ctx)

	// Always authenticate the fee payer first to enable gas to be paid.
	// The maximum number of authenticators is limited to 15 in the msg_server.
	feePayerIndex := 0
	for msgIndex, msg := range tx.GetMsgs() {
		account, err := GetAccount(msg)
		if err != nil {
			return sdk.Context{}, sdkerrors.Wrap(sdkerrors.ErrUnauthorized,
				fmt.Sprintf("failed to retrieve the account for message %d", msgIndex))
		}
		if account.Equals(feePayer) {
			feePayerIndex = msgIndex
			ctx, err := ad.AuthTx(cacheCtx, tx, msg, msgIndex, simulate)
			if err != nil {
				return ctx, err
			}
		}
	}

	// Authenticate the accounts of all messages.
	for msgIndex, msg := range tx.GetMsgs() {
		// Skip fee payer authentication as it has already been completed.
		if feePayerIndex == msgIndex {
			continue
		}

		// After the fee payer has been authenticated, proceed to authenticate the remaining messages.
		ctx, err := ad.AuthTx(cacheCtx, tx, msg, msgIndex, simulate)
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
		}
	}

	// If authentication fails, return an error.
	if !msgAuthenticated {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("authentication failed for message %d", msgIndex))
	}

	return cacheCtx, nil
}
