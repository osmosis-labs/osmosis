package authenticator

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// Compile time type assertion for the SignatureData using the
// SignatureVerificationAuthenticator struct
var _ Authenticator = &SignatureVerificationAuthenticator{}
var _ AuthenticatorData = &SignatureData{}

// signature authenticator
type SignatureVerificationAuthenticator struct {
	ak      *authkeeper.AccountKeeper
	Handler authsigning.SignModeHandler
	PubKey  cryptotypes.PubKey
}

func (sva SignatureVerificationAuthenticator) Type() string {
	return SignatureVerificationAuthenticatorType
}

func (sva SignatureVerificationAuthenticator) StaticGas() uint64 {
	// using 0 gas here. The gas is consumed based on the pubkey type in Authenticate()
	return 0
}

// NewSignatureVerificationAuthenticator creates a new SignatureVerificationAuthenticator
func NewSignatureVerificationAuthenticator(
	ak *authkeeper.AccountKeeper,
	Handler authsigning.SignModeHandler,
) SignatureVerificationAuthenticator {
	return SignatureVerificationAuthenticator{
		ak:      ak,
		Handler: Handler,
	}
}

// Initialize is used when a secondary account is used as an authenticator,
// this is used to verify a signature from an account that does not have a public key
// in the store. In this case we Initialize the authenticator from the authenticators store
func (sva SignatureVerificationAuthenticator) Initialize(
	data []byte,
) (Authenticator, error) {
	if len(data) != secp256k1.PubKeySize {
		sva.PubKey = nil
	}
	sva.PubKey = &secp256k1.PubKey{Key: data}
	return sva, nil
}

// SignatureData is used to package all the signature data and the tx
// for use in the Authenticate function
type SignatureData struct {
	Signers    []sdk.AccAddress
	Signatures []signing.SignatureV2
	Tx         authsigning.Tx
	Simulate   bool
}

// GetAuthenticationData parses the signers and signatures from a transactiom
// then returns a indexed list of both signers and signatures
// NOTE: position in the array is used to associate the signer and signature
func (sva SignatureVerificationAuthenticator) GetAuthenticationData(
	ctx sdk.Context,
	tx sdk.Tx,
	// TODO: revisit msg index functionality
	messageIndex int8,
	simulate bool,
) (AuthenticatorData, error) {
	sigTx, ok := tx.(authsigning.Tx)
	if !ok {
		return SignatureData{},
			sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	signatures, err := sigTx.GetSignaturesV2()
	if err != nil {
		return SignatureData{}, err
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// We get the signers here for an invariant check
	signers := sigTx.GetSigners()
	msgs := sigTx.GetMsgs()

	msgSigners, msgSignatures, err := GetSignersAndSignatures(
		msgs,
		signatures,
		feeTx.FeePayer().String(),
		// TODO: We need to clearly define why the message index is needed here.
		int(messageIndex),
	)
	if err != nil {
		return SignatureData{}, err
	}

	// NOTE: added signer invariant check to ensure our code is working as before
	if len(signers) != len(msgSigners) {
		return SignatureData{},
			sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invariant check failed, old signers don't match new signers")
	}

	// NOTE: added signature invariant check to ensure our code is working as before
	if len(signatures) != len(msgSignatures) {
		return SignatureData{},
			sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invariant check failed, old signatures don't match new signatures")
	}

	// Get the signature for the message at msgIndex
	return SignatureData{
		Signers:    msgSigners,
		Signatures: msgSignatures,
		Tx:         sigTx,
		Simulate:   simulate,
	}, nil
}

// Authenticate takes a SignaturesVerificationData struct and validates
// each signer and signature using  signature verification
func (sva SignatureVerificationAuthenticator) Authenticate(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData AuthenticatorData) AuthenticationResult {
	verificationData, ok := authenticationData.(SignatureData)
	if !ok {
		return Rejected("invalid signature verification data", sdkerrors.ErrInvalidType)
	}

	// First consume gas for verifing the signature
	params := sva.ak.GetParams(ctx)
	for _, sig := range verificationData.Signatures {
		err := authante.DefaultSigVerificationGasConsumer(ctx.GasMeter(), sig, params)
		if err != nil {
			return Rejected("couldn't get gas consumer", err)
		}
	}

	// after gas consumption continue to verify signatures
	for i, sig := range verificationData.Signatures {
		acc, err := authante.GetSignerAcc(ctx, sva.ak, verificationData.Signers[i])
		if err != nil {
			return Rejected("couldn't get signer account", err)
		}

		// Retrieve pubkey we use either the public key from the authenticator store
		// if that's not available query the original auth store for the public key
		// the public key is added to the sva struct by the Initialize function
		pubKey := sva.PubKey
		if pubKey == nil {
			// Having a default here keeps this authenticator stateless,
			// that way we don't have to create specific authenticators with the pubkey of each existing account
			pubKey = acc.GetPubKey() // TODO: do we want this default?
		}
		if !verificationData.Simulate && pubKey == nil {
			return Rejected("pubkey on not set on account or authenticator", sdkerrors.ErrInvalidPubKey)
		}

		// Check account sequence number.
		fmt.Println(account, acc.GetSequence(), sig.Sequence, acc.GetAddress())
		if sig.Sequence != acc.GetSequence() {
			return Rejected(
				fmt.Sprintf("account sequence mismatch, expected %d, got %d", acc.GetSequence(), sig.Sequence),
				sdkerrors.ErrInvalidPubKey,
			)
		}

		// Retrieve and build the signer data struct
		genesis := ctx.IsGenesis() || ctx.BlockHeight() == 0
		chainID := ctx.ChainID()
		var accNum uint64
		if !genesis {
			accNum = acc.GetAccountNumber()
		}
		signerData := authsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNum,
			Sequence:      acc.GetSequence(),
		}

		// No need to verify signatures on recheck tx
		if !verificationData.Simulate && !ctx.IsReCheckTx() {
			err := authsigning.VerifySignature(
				pubKey,
				signerData,
				sig.Data,
				sva.Handler,
				verificationData.Tx,
			)
			if err != nil {
				if authante.OnlyLegacyAminoSigners(sig.Data) {
					// If all signers are using SIGN_MODE_LEGACY_AMINO, we rely on VerifySignature to check account sequence number,
					// and therefore communicate sequence number as a potential cause of error.
					ctx.Logger().Debug(
						"signature verification failed; please verify account number (%d), sequence (%d) and chain-id (%s)",
						accNum,
						acc.GetSequence(),
						chainID,
					)
				} else {
					ctx.Logger().Debug(fmt.Sprintf("signature verification failed; please verify account number (%d) and chain-id (%s)",
						accNum,
						chainID,
					))
				}
				// Errors are reserved for when something unexpected happened. Here authentication just failed, so we
				// return skip
				return NotAuthenticated()
			}
		}
	}
	return Authenticated()
}

func (sva SignatureVerificationAuthenticator) ConfirmExecution(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData AuthenticatorData) ConfirmationResult {
	return Confirm()
}
