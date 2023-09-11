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

func (sva SignatureVerificationAuthenticator) Gas() uint64 {
	// The default gas for verifying a secp256k1 signature is 1000
	return 1000
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
	messageIndex uint8,
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

	msgs := sigTx.GetMsgs()

	msgSigners, msgSignatures, err := GetSignersAndSignatures(
		msgs,
		signatures,
		"",
		int(messageIndex),
	)
	if err != nil {
		return SignatureData{}, err
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
func (sva SignatureVerificationAuthenticator) Authenticate(
	ctx sdk.Context,
	msg sdk.Msg,
	authenticationData AuthenticatorData,
) (success bool, err error) {
	verificationData, ok := authenticationData.(SignatureData)
	if !ok {
		return false, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "invalid signature verification data")
	}

	// First consume gas for verifing the signature
	params := sva.ak.GetParams(ctx)
	for _, sig := range verificationData.Signatures {
		err := authante.DefaultSigVerificationGasConsumer(ctx.GasMeter(), sig, params)
		if err != nil {
			return false, err
		}
	}

	// after gas consumption continue to verify signatures
	for i, sig := range verificationData.Signatures {
		acc, err := authante.GetSignerAcc(ctx, sva.ak, verificationData.Signers[i])
		if err != nil {
			return false, err
		}

		// retrieve pubkey
		pubKey := sva.PubKey
		if pubKey == nil {
			// Having a default here keeps this authenticator stateless,
			// that way we don't have to create specific authenticators with the pubkey of each existing account
			pubKey = acc.GetPubKey() // TODO: do we want this default?
		}
		if !verificationData.Simulate && pubKey == nil {
			return false, sdkerrors.Wrap(
				sdkerrors.ErrInvalidPubKey,
				"pubkey on not set on account or authenticator",
			)
		}

		// Check account sequence number.
		if sig.Sequence != acc.GetSequence() {
			return false, sdkerrors.Wrapf(
				sdkerrors.ErrWrongSequence,
				"account sequence mismatch, expected %d, got %d", acc.GetSequence(), sig.Sequence,
			)
		}

		// retrieve signer data
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

		// no need to verify signatures on recheck tx
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
				// return false
				return false, nil
			}
		}
	}
	return true, nil
}

func (sva SignatureVerificationAuthenticator) ConfirmExecution(
	ctx sdk.Context,
	msg sdk.Msg,
	authenticated bool,
	authenticationData AuthenticatorData,
) bool {
	// To be executed in the post handler
	return true
}
