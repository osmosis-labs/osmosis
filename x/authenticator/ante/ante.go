package ante

import (
	"fmt"
	"strconv"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/osmosis-labs/osmosis/v23/x/authenticator/authenticator"
	types "github.com/osmosis-labs/osmosis/v23/x/authenticator/iface"
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/utils"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

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

	// Ideally, we would prefer to use min(gasRemaining, maxFeePayerGas) here, but
	// this approach presents challenges due to the implementation of the InfiniteGasMeter.
	// As long as the gas consumption remains below the fee payer gas limit, exceeding
	// the original limit should be acceptable.
	maximumUnauthenticatedGasParam := ad.authenticatorKeeper.GetParams(ctx)
	payerGasMeter := sdk.NewGasMeter(maximumUnauthenticatedGasParam.MaximumUnauthenticatedGas)
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
					maximumUnauthenticatedGasParam.MaximumUnauthenticatedGas, payerGasMeter.GasConsumed())
				err = errorsmod.Wrap(sdkerrors.ErrOutOfGas, log)
			default:
				panic(r)
			}
		}
	}()

	// This is a transient context stored globally throughout the execution of the tx
	// Any changes will to authenticator storage will be written to the store at the end of the tx
	cacheCtx := ad.authenticatorKeeper.TransientStore.ResetTransientContext(ctx)
	// Ensure that no usedAuthenticators are stored in the transient store
	ad.authenticatorKeeper.TransientStore.ResetUsedAuthenticators()

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

	ak, ok := ad.accountKeeper.(*authkeeper.AccountKeeper)
	if !ok {
		return sdk.Context{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "invalid account keeper type")
	}

	var tracks []func() error

	signerToSignature, err := authenticator.GetCommonAuthenticationData(tx, -1)
	if err != nil {
		return sdk.Context{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "failed to get common authentication data")
	}
	authenticationRequest, err := authenticator.GenerateBaseAuthenticationRequest(ctx, tx, signerToSignature, msgs, simulate)
	if err != nil {
		return sdk.Context{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "failed to generate base authentication request")
	}

	// Authenticate the accounts of all messages
	for msgIndex, msg := range msgs {
		// By default, the first signer is the account
		account, err := utils.GetAccount(msg)
		if err != nil {
			return sdk.Context{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("failed to get account for message %d", msgIndex))
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
				return ctx, errorsmod.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("invalid authenticator index for message %d", msgIndex))
			}
			authenticators = []types.InitializedAuthenticator{allAuthenticators[selectedAuthenticators[msgIndex]]}
		}

		// Generate the authentication request data
		err = authenticator.UpdateAuthenticationRequestForAccount(ctx, &authenticationRequest, tx, signerToSignature, ak, ad.sigModeHandler, account, uint64(msgIndex), authenticator.SequenceMatch)
		if err != nil {
			return sdk.Context{}, errorsmod.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("failed to update authentication request with message %d", msgIndex))
		}

		msgAuthenticated := false
		for _, initializedAuthenticator := range authenticators {
			a11r := initializedAuthenticator.Authenticator
			id := initializedAuthenticator.Id
			stringId := strconv.FormatInt(int64(id), 10)

			// Consume the authenticator's static gas
			cacheCtx.GasMeter().ConsumeGas(a11r.StaticGas(), "authenticator static gas")

			authenticationRequest.AuthenticatorId = stringId
			authentication := a11r.Authenticate(cacheCtx, authenticationRequest)
			if authentication.IsRejected() {
				return ctx, authentication.Error()
			}

			if authentication.IsAuthenticated() {
				ad.authenticatorKeeper.TransientStore.AddUsedAuthenticator(id)
				msgAuthenticated = true
				// Once the fee payer is authenticated, we can set the gas limit to its original value
				if !feePayerAuthenticated && account.Equals(feePayer) {
					originalGasMeter.ConsumeGas(payerGasMeter.GasConsumed(), "fee payer gas")
					// Reset this for both contexts
					cacheCtx = ad.authenticatorKeeper.TransientStore.GetTransientContextWithGasMeter(originalGasMeter)
					ctx = ctx.WithGasMeter(originalGasMeter)
					feePayerAuthenticated = true
				}

				// Append the track closure to be called after the fee payer is authenticated
				tracks = append(tracks, func() error {
					err := a11r.Track(ctx, account, msg, stringId)
					if err != nil {
						return err
					}
					return nil
				})

				break
			}
		}

		// if authentication failed, return an error
		if !msgAuthenticated {
			return ctx, errorsmod.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("authentication failed for message %d", msgIndex))
		}
	}

	// Ensure that the fee payer has been authenticated. For this to be true, the fee payer must be
	// the signer of at least one message
	if !feePayerAuthenticated {
		return ctx, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "fee payer not authenticated")
	}

	// If the transaction has been authenticated, we call Track(...) on every message
	// to notify its authenticator so that it can handle any state updates.
	for _, track := range tracks {
		if err := track(); err != nil {
			return sdk.Context{}, err
		}
	}

	return next(ctx, tx, simulate)
}

// GetSelectedAuthenticators retrieves the selected authenticators for the provided transaction extension
// and matches them with the number of messages in the transaction.
// If no selected authenticators are found in the extension, the function initializes the list with -1 values.
// It returns an array of selected authenticators or an error if the number of selected authenticators does not match
// the number of messages in the transaction.
func (ad AuthenticatorDecorator) GetSelectedAuthenticators(extTx authante.HasExtensionOptionsTx, msgCount int) ([]int32, error) {
	// Initialize the list of selected authenticators with -1 values.
	selectedAuthenticators := make([]int32, msgCount)
	for i := range selectedAuthenticators {
		selectedAuthenticators[i] = -1
	}

	// Get the transaction options from the AuthenticatorKeeper extension.
	txOptions := ad.authenticatorKeeper.GetAuthenticatorExtension(extTx.GetNonCriticalExtensionOptions())

	if txOptions != nil {
		// Retrieve the selected authenticators from the extension.
		selectedAuthenticatorsFromExtension := txOptions.GetSelectedAuthenticators()

		if len(selectedAuthenticatorsFromExtension) != msgCount {
			// Return an error if the number of selected authenticators does not match the number of messages.
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Mismatch between the number of selected authenticators and messages")
		}

		// Use the selected authenticators from the extension.
		selectedAuthenticators = selectedAuthenticatorsFromExtension
	}

	return selectedAuthenticators, nil
}
