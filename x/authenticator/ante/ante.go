package ante

import (
	"fmt"

	types "github.com/osmosis-labs/osmosis/v19/x/authenticator/iface"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/utils"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authenticatorkeeper "github.com/osmosis-labs/osmosis/v19/x/authenticator/keeper"
)

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
	originalGasMeter := ctx.GasMeter()

	// Ideally, we would prefer to use min(gasRemaining, maxFeePayerGas) here, but
	// this approach presents challenges due to the implementation of the InfiniteGasMeter.
	// As long as the gas consumption remains below the fee payer gas limit, exceeding
	// the original limit should be acceptable.
	payerGasMeter := sdk.NewGasMeter(ad.maxFeePayerGas)
	ctx = ctx.WithGasMeter(payerGasMeter)

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		// This should never happen
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
	cacheCtx := ad.authenticatorKeeper.TransientStore.ResetTransientContext(ctx)

	extTx, ok := tx.(authante.HasExtensionOptionsTx)
	if !ok {
		// This should never happen
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a HasExtensionOptionsTx")
	}

	msgs := tx.GetMsgs()
	selectedAuthenticators, err := ad.GetSelectedAuthenticators(extTx, len(msgs))
	if err != nil {
		return ctx, err
	}

	// Authenticate the accounts of all messages
	for msgIndex, msg := range msgs {
		// By default, the first signer is the account
		account, err := utils.GetAccount(msg)
		if err != nil {
			return sdk.Context{}, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("failed to get account for message %d", msgIndex))
		}

		// Get all authenticators for the executing account
		// If no authenticators are found, use the default authenticator
		// This is done to keep backwards compatibility by defaulting to a signature verifier on accounts without authenticators
		allAuthenticators, err := ad.authenticatorKeeper.GetAuthenticatorsForAccountOrDefault(cacheCtx, account)
		if err != nil {
			return sdk.Context{}, err
		}

		// Check if there has been a selected authenticator in the transaction
		authenticators := allAuthenticators
		if selectedAuthenticators[msgIndex] >= 0 {
			if int(selectedAuthenticators[msgIndex]) >= len(allAuthenticators) {
				return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("invalid authenticator index for message %d", msgIndex))
			}
			authenticators = []types.Authenticator{allAuthenticators[selectedAuthenticators[msgIndex]]}
		}

		msgAuthenticated := false
		for _, authenticator := range authenticators {
			// Consume the authenticator's static gas
			cacheCtx.GasMeter().ConsumeGas(authenticator.StaticGas(), "authenticator static gas")

			// Get the authentication data for the transaction
			neverWriteCacheCtx, _ := cacheCtx.CacheContext() // GetAuthenticationData is not allowed to modify the state
			authData, err := authenticator.GetAuthenticationData(neverWriteCacheCtx, tx, msgIndex, simulate)
			if err != nil {
				return ctx, err
			}

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
					cacheCtx = ad.authenticatorKeeper.TransientStore.GetTransientContextWithGasMeter(originalGasMeter)
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

	// Ensure that the fee payer has been authenticated. For this to be true, the fee payer must be
	// the signer of at least one message
	if !feePayerAuthenticated {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "fee payer not authenticated")
	}

	// If the transaction has been authenticated, we call TrackMessages(...) to
	// notify every authenticator so that they can handle any storage updates
	// that need to happen regardless of how the message was authorized.
	err = utils.TrackMessages(cacheCtx, ad.authenticatorKeeper, msgs)
	if err != nil {
		return sdk.Context{}, err
	}

	return next(ctx, tx, simulate)
}

func (ad AuthenticatorDecorator) GetSelectedAuthenticators(extTx authante.HasExtensionOptionsTx, msgCount int) ([]int32, error) {
	selectedAuthenticators := make([]int32, msgCount)
	for i := range selectedAuthenticators {
		selectedAuthenticators[i] = -1
	}

	txOptions := ad.authenticatorKeeper.GetAuthenticatorExtension(extTx.GetNonCriticalExtensionOptions())
	if txOptions != nil {
		selectedAuthenticatorsFromExtension := txOptions.GetSelectedAuthenticators()
		if len(selectedAuthenticatorsFromExtension) != msgCount {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Mismatch between the number of selected authenticators and messages")
		}
		selectedAuthenticators = selectedAuthenticatorsFromExtension
	}
	return selectedAuthenticators, nil
}
