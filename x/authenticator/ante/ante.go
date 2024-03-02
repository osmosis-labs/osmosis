package ante

import (
	"fmt"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/osmosis-labs/osmosis/v23/x/authenticator/authenticator"
	authenticatorkeeper "github.com/osmosis-labs/osmosis/v23/x/authenticator/keeper"
)

// AuthenticatorDecorator is responsible for processing authentication logic
// before transaction execution.
type AuthenticatorDecorator struct {
	authenticatorKeeper *authenticatorkeeper.Keeper
	accountKeeper       authante.AccountKeeper
	sigModeHandler      authsigning.SignModeHandler
}

// NewAuthenticatorDecorator creates a new instance of AuthenticatorDecorator with the provided parameters.
func NewAuthenticatorDecorator(
	authenticatorKeeper *authenticatorkeeper.Keeper,
	accountKeeper authante.AccountKeeper,
	sigModeHandler authsigning.SignModeHandler,
) AuthenticatorDecorator {
	return AuthenticatorDecorator{
		authenticatorKeeper: authenticatorKeeper,
		accountKeeper:       accountKeeper,
		sigModeHandler:      sigModeHandler,
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

	// As long as the gas consumption remains below the fee payer gas limit, exceeding
	// the original limit should be acceptable.
	authenticatorParams := ad.authenticatorKeeper.GetParams(ctx)
	payerGasMeter := sdk.NewGasMeter(authenticatorParams.MaximumUnauthenticatedGas)
	ctx = ctx.WithGasMeter(payerGasMeter)

	// Recover from any OutOfGas panic to return an error with information of the gas limit having been reduced
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case sdk.ErrorOutOfGas:
				log := fmt.Sprintf(
					"FeePayer not authenticated yet. The gas limit has been reduced to %d. Consumed: %d",
					authenticatorParams.MaximumUnauthenticatedGas, payerGasMeter.GasConsumed())
				err = errorsmod.Wrap(sdkerrors.ErrOutOfGas, log)
			default:
				panic(r)
			}
		}
	}()

	cacheCtx, writeCache := ctx.CacheContext()

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// By default, the fee payer is the first signer of the transaction
	feePayer := feeTx.FeePayer()

	msgs := tx.GetMsgs()
	selectedAuthenticators, err := ad.GetSelectedAuthenticators(tx, len(msgs))
	if err != nil {
		return ctx, err
	}

	ak, ok := ad.accountKeeper.(*authkeeper.AccountKeeper)
	if !ok {
		return sdk.Context{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "invalid account keeper type")
	}

	// tracks are used to make sure that we only write to the store after every message is successful
	var tracks []func() error

	// Authenticate the accounts of all messages
	for msgIndex, msg := range msgs {
		signers := msg.GetSigners()

		// Enforce only one signer per message
		if len(signers) != 1 {
			return sdk.Context{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "messages must have exactly one signer")
		}

		// By default, the first signer is the account that is used
		account := signers[0]

		// Ensure the feePayer is the signer of the first message
		if msgIndex == 0 && !feePayer.Equals(account) {
			return sdk.Context{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "feePayer must be the signer of the first message")
		}

		// Get the currently selected authenticator
		selectedAuthenticator, err := ad.authenticatorKeeper.GetInitializedAuthenticatorForAccount(
			cacheCtx,
			account,
			int(selectedAuthenticators[msgIndex]),
		)
		if err != nil {
			return sdk.Context{}, err
		}

		// Generate the authentication request data
		authenticationRequest, err := authenticator.GenerateAuthenticationData(
			ctx,
			ak,
			ad.sigModeHandler,
			account,
			feePayer,
			msg,
			tx,
			msgIndex,
			simulate,
			authenticator.SequenceMatch,
		)
		if err != nil {
			return sdk.Context{},
				errorsmod.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("failed to get authentication data for message %d", msgIndex))
		}

		a11r := selectedAuthenticator.Authenticator
		stringId := strconv.FormatUint(selectedAuthenticator.Id, 10)
		authenticationRequest.AuthenticatorId = stringId

		// Consume the authenticator's static gas
		cacheCtx.GasMeter().ConsumeGas(a11r.StaticGas(), "authenticator static gas")

		// Authenticate should never modify state. That's what track is for
		neverWriteCtx, _ := cacheCtx.CacheContext()
		authErr := a11r.Authenticate(neverWriteCtx, authenticationRequest)

		// If authentication is successful, continue
		if authErr == nil {
			// Once the fee payer is authenticated, we can set the gas limit to its original value
			if account.Equals(feePayer) {
				originalGasMeter.ConsumeGas(payerGasMeter.GasConsumed(), "fee payer gas")

				// Reset this for both contexts
				cacheCtx = cacheCtx.WithGasMeter(originalGasMeter)
				ctx = ctx.WithGasMeter(originalGasMeter)
			}

			// Append the track closure to be called after every message is authenticated
			tracks = append(tracks, func() error {
				err := a11r.Track(cacheCtx, account, feePayer, msg, uint64(msgIndex), stringId)

				if err != nil {
					return err
				}
				return nil
			})
		}

		// If authentication failed, return an error
		if authErr != nil {
			return ctx, errorsmod.Wrap(
				sdkerrors.ErrUnauthorized,
				fmt.Sprintf("authentication failed for message %d: %s", msgIndex, authErr),
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

	writeCache()
	return next(ctx, tx, simulate)
}

// GetSelectedAuthenticators retrieves the selected authenticators for the provided transaction extension
// and matches them with the number of messages in the transaction.
// If no selected authenticators are found in the extension, the function initializes the list with -1 values.
// It returns an array of selected authenticators or an error if the number of selected authenticators does not match
// the number of messages in the transaction.
func (ad AuthenticatorDecorator) GetSelectedAuthenticators(
	tx sdk.Tx,
	msgCount int,
) ([]int64, error) {
	extTx, ok := tx.(authante.HasExtensionOptionsTx)
	if !ok {
		return nil, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a HasExtensionOptionsTx to use Authenticators")
	}

	// Get the selected authenticator options from the transaction.
	txOptions := ad.authenticatorKeeper.GetAuthenticatorExtension(extTx.GetNonCriticalExtensionOptions())
	if txOptions == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest,
			"Cannot get tx ext")
	}
	// Retrieve the selected authenticators from the extension.
	selectedAuthenticators := txOptions.GetSelectedAuthenticators()

	if len(selectedAuthenticators) != msgCount {
		// Return an error if the number of selected authenticators does not match the number of messages.
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest,
			"Mismatch between the number of selected authenticators and messages")
	}

	return selectedAuthenticators, nil
}
