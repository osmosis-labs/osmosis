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
	types "github.com/osmosis-labs/osmosis/v23/x/authenticator/iface"
	authenticatorkeeper "github.com/osmosis-labs/osmosis/v23/x/authenticator/keeper"
	authenticatortypes "github.com/osmosis-labs/osmosis/v23/x/authenticator/types"
)

// AuthenticatorDecorator is responsible for processing authentication logic
// before transaction execution.
type AuthenticatorDecorator struct {
	authenticatorKeeper *authenticatorkeeper.Keeper
	accountKeeper       authante.AccountKeeper
	sigModeHandler      authsigning.SignModeHandler
	next                sdk.AnteHandler
}

// NewAuthenticatorDecorator creates a new instance of AuthenticatorDecorator with the provided parameters.
func NewAuthenticatorDecorator(
	authenticatorKeeper *authenticatorkeeper.Keeper,
	accountKeeper authante.AccountKeeper,
	sigModeHandler authsigning.SignModeHandler,
	next sdk.AnteHandler,

) AuthenticatorDecorator {
	return AuthenticatorDecorator{
		authenticatorKeeper: authenticatorKeeper,
		accountKeeper:       accountKeeper,
		sigModeHandler:      sigModeHandler,
		next:                next,
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
	// Check that the authenticator flow is active by querying the params
	authenticatorParams := ad.authenticatorKeeper.GetParams(ctx)
	if !authenticatorParams.AreSmartAccountsActive {
		return ad.next(ctx, tx, simulate)
	}

	extTx, ok := tx.(authante.HasExtensionOptionsTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a HasExtensionOptionsTx to use Authenticators")
	}

	// Get the selected authenticator options from the transaction.
	txOptions := ad.authenticatorKeeper.GetAuthenticatorExtension(extTx.GetNonCriticalExtensionOptions())
	if txOptions == nil {
		return ad.next(ctx, tx, simulate)
	}
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

	// Ensure that no usedAuthenticators are stored in the transient store
	ad.authenticatorKeeper.UsedAuthenticators.ResetUsedAuthenticators()

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// The fee payer by default is the first signer of the transaction
	feePayer := feeTx.FeePayer()

	msgs := tx.GetMsgs()
	selectedAuthenticators, err := ad.GetSelectedAuthenticators(txOptions, len(msgs))
	if err != nil {
		return ctx, err
	}

	ak, ok := ad.accountKeeper.(*authkeeper.AccountKeeper)
	if !ok {
		return sdk.Context{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "invalid account keeper type")
	}

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

		// Get all authenticators for the executing account
		// If no authenticators are found, use the default authenticator
		// This is done to keep backwards compatibility by defaulting to a signature verifier on accounts without authenticators
		// TODO: only return the selected account authenticator (no defaults)
		allAuthenticators, err := ad.authenticatorKeeper.GetAuthenticatorsForAccountOrDefault(cacheCtx, account)
		if err != nil {
			return sdk.Context{}, err
		}

		// Check if there has been a selected authenticator in the transaction
		authenticators := allAuthenticators
		if selectedAuthenticators[msgIndex] >= 0 {
			if int(selectedAuthenticators[msgIndex]) >= len(allAuthenticators) {
				return ctx, errorsmod.Wrap(sdkerrors.ErrUnauthorized,
					fmt.Sprintf("invalid authenticator index for message %d", msgIndex))
			}
			authenticators = []types.InitializedAuthenticator{allAuthenticators[selectedAuthenticators[msgIndex]]}
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

		var authErr error
		for _, initializedAuthenticator := range authenticators {
			a11r := initializedAuthenticator.Authenticator
			id := initializedAuthenticator.Id
			stringId := strconv.FormatInt(int64(id), 10)

			// Consume the authenticator's static gas
			cacheCtx.GasMeter().ConsumeGas(a11r.StaticGas(), "authenticator static gas")

			authenticationRequest.AuthenticatorId = stringId
			// Authenticate should never modify state. That's what track is for
			neverWriteCtx, _ := cacheCtx.CacheContext()
			authErr = a11r.Authenticate(neverWriteCtx, authenticationRequest)

			// if authentication is successful, continue
			if authErr == nil {
				// authentication succeeded, add the authenticator to the used authenticators
				ad.authenticatorKeeper.UsedAuthenticators.AddUsedAuthenticator(id)
				// Once the fee payer is authenticated, we can set the gas limit to its original value
				if account.Equals(feePayer) {
					originalGasMeter.ConsumeGas(payerGasMeter.GasConsumed(), "fee payer gas")
					// Reset this for both contexts
					cacheCtx = cacheCtx.WithGasMeter(originalGasMeter)
					ctx = ctx.WithGasMeter(originalGasMeter)
				}

				// Append the track closure to be called after the fee payer is authenticated
				tracks = append(tracks, func() error {
					err := a11r.Track(cacheCtx, account, feePayer, msg, uint64(msgIndex), stringId)

					if err != nil {
						return err
					}
					return nil
				})

				// skip the rest if found a successful authenticator
				break
			}
		}

		// if authentication failed, return an error
		if authErr != nil {
			return ctx, errorsmod.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("authentication failed for message %d: %s", msgIndex, authErr))
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
func (ad AuthenticatorDecorator) GetSelectedAuthenticators(txOptions authenticatortypes.AuthenticatorTxOptions, msgCount int) ([]int32, error) {
	// Initialize the list of selected authenticators with -1 values.
	selectedAuthenticators := make([]int32, msgCount)
	for i := range selectedAuthenticators {
		selectedAuthenticators[i] = -1
	}

	if txOptions != nil {
		// Retrieve the selected authenticators from the extension.
		selectedAuthenticatorsFromExtension := txOptions.GetSelectedAuthenticators()

		if len(selectedAuthenticatorsFromExtension) != msgCount {
			// Return an error if the number of selected authenticators does not match the number of messages.
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest,
				"Mismatch between the number of selected authenticators and messages")
		}

		// Use the selected authenticators from the extension.
		selectedAuthenticators = selectedAuthenticatorsFromExtension
	}

	return selectedAuthenticators, nil
}
