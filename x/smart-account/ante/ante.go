package ante

import (
	"bytes"
	"fmt"
	customante "github.com/osmosis-labs/osmosis/v27/ante"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	storetypes "cosmossdk.io/store/types"
	txsigning "cosmossdk.io/x/tx/signing"

	"github.com/osmosis-labs/osmosis/v27/x/smart-account/authenticator"
	smartaccountkeeper "github.com/osmosis-labs/osmosis/v27/x/smart-account/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/types"
)

// AuthenticatorDecorator is responsible for processing authentication logic
// before transaction execution.
type AuthenticatorDecorator struct {
	smartAccountKeeper *smartaccountkeeper.Keeper
	accountKeeper      authante.AccountKeeper
	sigModeHandler     *txsigning.HandlerMap
	deductFeeDecorator customante.DeductFeeDecorator
	cdc                codec.Codec
}

// NewAuthenticatorDecorator creates a new instance of AuthenticatorDecorator with the provided parameters.
func NewAuthenticatorDecorator(
	cdc codec.Codec,
	smartAccountKeeper *smartaccountkeeper.Keeper,
	accountKeeper authante.AccountKeeper,
	sigModeHandler *txsigning.HandlerMap,
	deductFeeDecorator customante.DeductFeeDecorator,
) AuthenticatorDecorator {
	return AuthenticatorDecorator{
		smartAccountKeeper: smartAccountKeeper,
		accountKeeper:      accountKeeper,
		sigModeHandler:     sigModeHandler,
		deductFeeDecorator: deductFeeDecorator,
		cdc:                cdc,
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
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, types.MeasureKeyAnteHandler)

	// Authenticators don't support manually setting the fee payer
	err = ad.ValidateAuthenticatorFeePayer(tx)
	if err != nil {
		return sdk.Context{}, err
	}

	// Performing fee payer authentication with minimal gas allocation
	// serves as a spam-prevention strategy to prevent users from adding multiple
	// authenticators that may excessively consume computational resources.
	// If the fee payer is not authenticated, there would be no entity responsible
	// for covering the transaction's costs. This safeguard ensures that validators
	// are not compelled to expend resources on executing authenticators for transactions
	// that will never be executed.
	originalGasMeter := ctx.GasMeter()
	prevGasConsumed := originalGasMeter.GasConsumed()

	// As long as the gas consumption remains below the fee payer gas limit, exceeding
	// the original limit should be acceptable.
	authenticatorParams := ad.smartAccountKeeper.GetParams(ctx)
	payerGasMeter := storetypes.NewGasMeter(authenticatorParams.MaximumUnauthenticatedGas)
	ctx = ctx.WithGasMeter(payerGasMeter)

	// Recover from any OutOfGas panic to return an error with information of the gas limit having been reduced
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case storetypes.ErrorOutOfGas:
				log := fmt.Sprintf(
					"FeePayer must be authenticated first because gas consumption has exceeded the free gas limit for authentication process. The gas limit has been reduced to %d. Gas consumed: %d",
					authenticatorParams.MaximumUnauthenticatedGas, payerGasMeter.GasConsumed())
				err = errorsmod.Wrap(sdkerrors.ErrOutOfGas, log)
			default:
				panic(r)
			}
		}
	}()

	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "no messages in transaction")
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// The fee payer is the first signer of the transaction. This should have been enforced by the
	// LimitFeePayerDecorator
	signers, _, err := ad.cdc.GetMsgV1Signers(msgs[0])
	if err != nil {
		return ctx, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "failed to get signers")
	}
	feePayer := sdk.AccAddress(signers[0])
	feeGranter := feeTx.FeeGranter()
	fee := feeTx.GetFee()

	selectedAuthenticators, err := ad.GetSelectedAuthenticators(tx, len(msgs))
	if err != nil {
		return ctx, err
	}

	// tracks are used to make sure that we only write to the store after every message is successful
	var tracks []func() error

	// feesPaid is used to track when fee have been paid, this is to avoid the fee payer being charged twice
	feesPaid := false

	// Authenticate the accounts of all messages
	for msgIndex, msg := range msgs {
		signers, _, err := ad.cdc.GetMsgV1Signers(msg)
		if err != nil {
			return ctx, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "failed to get signers")
		}
		// Enforce only one signer per message
		if len(signers) != 1 {
			return sdk.Context{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "messages must have exactly one signer")
		}

		// By default, the first signer is the account that is used
		account := sdk.AccAddress(signers[0])

		// Get the currently selected authenticator
		selectedAuthenticatorId := int(selectedAuthenticators[msgIndex])
		selectedAuthenticator, err := ad.smartAccountKeeper.GetInitializedAuthenticatorForAccount(
			ctx,
			account,
			selectedAuthenticatorId,
		)
		if err != nil {
			return sdk.Context{},
				errorsmod.Wrapf(err, "failed to get initialized authenticator (account = %s, authenticator id = %d, msg index = %d, msg type url = %s)", account, selectedAuthenticatorId, msgIndex, sdk.MsgTypeURL(msg))
		}

		// Generate the authentication request data
		authenticationRequest, err := authenticator.GenerateAuthenticationRequest(
			ctx,
			ad.cdc,
			ad.accountKeeper,
			ad.sigModeHandler,
			account,
			feePayer,
			feeGranter,
			fee,
			msg,
			tx,
			msgIndex,
			simulate,
			authenticator.SequenceMatch,
		)
		if err != nil {
			return sdk.Context{},
				errorsmod.Wrapf(err, "failed to generate authentication data (account = %s, authenticator id = %d, msg index = %d, msg type url = %s)", account, selectedAuthenticator.Id, msgIndex, sdk.MsgTypeURL(msg))
		}

		a11r := selectedAuthenticator.Authenticator
		stringId := strconv.FormatUint(selectedAuthenticator.Id, 10)
		authenticationRequest.AuthenticatorId = stringId

		// Consume the authenticator's static gas
		ctx.GasMeter().ConsumeGas(a11r.StaticGas(), "authenticator static gas")

		// Authenticate should never modify state. That's what track is for
		neverWriteCtx, _ := ctx.CacheContext()
		authErr := a11r.Authenticate(neverWriteCtx, authenticationRequest)

		// If authentication is successful, continue
		if authErr == nil {
			// Once the fee payer is authenticated, we can set the gas limit to its original value
			if !feesPaid && account.Equals(feePayer) {
				originalGasMeter.ConsumeGas(payerGasMeter.GasConsumed(), "fee payer gas")

				// Once the fee payer is authenticated, we can deduct the fee.
				// This change will persist regardless of whether the rest of the messages pass authentication
				// or not
				_, err := ad.deductFeeDecorator.AnteHandle(ctx, tx, simulate, sdk.ChainAnteDecorators(sdk.Terminator{})) //nolint
				if err != nil {
					return sdk.Context{}, err
				}

				// Write the cache multi store to persist the fee deduction
				cacheMultiStore, ok := ctx.MultiStore().(storetypes.CacheMultiStore)
				if !ok {
					// This should never happen
					return sdk.Context{}, errorsmod.Wrap(sdkerrors.ErrPanic, "expected CacheMultiStore")
				}
				cacheMultiStore.Write()

				// Reset the gas meter
				ctx = ctx.WithGasMeter(originalGasMeter)

				// Set the feesPaid variable to true
				feesPaid = true
			}

			// Append the track closure to be called after every message is authenticated
			// Note: pre-initialize type URL to avoid closure issues from passing a msg
			// loop variable inside the closure.
			currentMsgTypeURL := sdk.MsgTypeURL(msg)
			tracks = append(tracks, func() error {
				err := a11r.Track(ctx, authenticationRequest)
				if err != nil {
					// track should not fail in normal circumstances, since it is intended to update track state before execution.
					// If it does fail, we log the error.
					telemetry.IncrCounter(1, types.CounterKeyTrackFailed)
					ad.smartAccountKeeper.Logger(ctx).Error(
						"track failed", "account", account, "feePayer", feePayer, "msg", currentMsgTypeURL, "authenticatorId", stringId, "error", err)

					return errorsmod.Wrapf(err, "track failed (account = %s, authenticator id = %s, authenticator type, %s, msg index = %d)", account, stringId, a11r.Type(), msgIndex)
				}
				return nil
			})
		}

		// If authentication failed, return an error
		if authErr != nil {
			return ctx, errorsmod.Wrapf(
				authErr,
				"authentication failed for message %d, authenticator id %d, type %s", msgIndex, selectedAuthenticator.Id, selectedAuthenticator.Authenticator.Type(),
			)
		}
	}

	// If the transaction has been authenticated, we call Track(...) on every message
	// to notify its authenticator so that it can handle any state updates.
	for _, track := range tracks {
		if err := track(); err != nil {
			return sdk.Context{}, err
		}
	}

	updatedGasConsumed := ctx.GasMeter().GasConsumed()
	telemetry.SetGauge(float32(updatedGasConsumed-prevGasConsumed), types.GaugeKeyAnteHandlerGasConsumed)

	return next(ctx, tx, simulate)
}

// ValidateAuthenticatorFeePayer enforces that the tx fee payer has not been set manually
// to an account different to the signer of the first message. This is a requirement
// for the authenticator module.
// The only user of a manually set fee payer is with fee grants, which are not
// available on osmosis
func (ad AuthenticatorDecorator) ValidateAuthenticatorFeePayer(tx sdk.Tx) error {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// The fee payer by default is the first signer of the transaction
	feePayer := feeTx.FeePayer()

	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must contain at least one message")
	}
	signers, _, err := ad.cdc.GetMsgV1Signers(msgs[0])
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "failed to get signers")
	}
	if len(signers) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx message must contain at least one signer")
	}

	if !bytes.Equal(feePayer, signers[0]) {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "fee payer must be the first signer")
	}
	return nil
}

// GetSelectedAuthenticators retrieves the selected authenticators for the provided transaction extension
// and matches them with the number of messages in the transaction.
// If no selected authenticators are found in the extension, the function initializes the list with -1 values.
// It returns an array of selected authenticators or an error if the number of selected authenticators does not match
// the number of messages in the transaction.
func (ad AuthenticatorDecorator) GetSelectedAuthenticators(
	tx sdk.Tx,
	msgCount int,
) ([]uint64, error) {
	extTx, ok := tx.(authante.HasExtensionOptionsTx)
	if !ok {
		return nil, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a HasExtensionOptionsTx to use Authenticators")
	}

	// Get the selected authenticator options from the transaction.
	txOptions := ad.smartAccountKeeper.GetAuthenticatorExtension(extTx.GetNonCriticalExtensionOptions())
	if txOptions == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest,
			"Cannot get AuthenticatorTxOptions from tx")
	}
	// Retrieve the selected authenticators from the extension.
	selectedAuthenticators := txOptions.GetSelectedAuthenticators()

	if len(selectedAuthenticators) != msgCount {
		// Return an error if the number of selected authenticators does not match the number of messages.
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"Mismatch between the number of selected authenticators and messages, msg count %d, got %d selected authenticators", msgCount, len(selectedAuthenticators))
	}

	return selectedAuthenticators, nil
}
