package authenticator

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v21/x/authenticator/iface"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// Compile time type assertion for the SignatureData using the
// PassKeyAuthenticator struct
var _ iface.Authenticator = &PassKeyAuthenticator{}
var _ iface.AuthenticatorData = &SignatureData{}

const (
	// PassKeyAuthenticatorType represents a type of authenticator specifically designed for
	// secp256r1 signature verification.
	PassKeyAuthenticatorType = "PassKeyAuthenticator"
)

type PassKeyAuthenticator struct {
	ak      *authkeeper.AccountKeeper
	Handler authsigning.SignModeHandler
	PubKey  cryptotypes.PubKey
}

func (sva PassKeyAuthenticator) Type() string {
	return PassKeyAuthenticatorType
}

func (sva PassKeyAuthenticator) StaticGas() uint64 {
	// using 0 gas here. The gas is consumed based on the pubkey type in Authenticate()
	return 0
}

// NewPassKeyAuthenticator creates a new PassKeyAuthenticator
// when the app starts a PassKeyAuthenticator is passed to the authentation manager,
// it will only be instantiated once then used for each signature authentation.
func NewPassKeyAuthenticator(
	ak *authkeeper.AccountKeeper,
	Handler authsigning.SignModeHandler,
) PassKeyAuthenticator {
	// NOTE: We generate a private key here that is not used
	// we do this to allow the Initialize function access to the PubKey struct
	priv, err := secp256r1.GenPrivKey()
	if err != nil {
		panic(err)
	}
	pk := priv.PubKey()

	return PassKeyAuthenticator{
		ak:      ak,
		Handler: Handler,
		PubKey:  pk,
	}
}

// Initialize is used when a secondary account is used as an authenticator,
// this is used to verify a signature from an account that does not have a public key
// in the store. In this case we Initialize the authenticator from the authenticators store
// which should have a public key in the data field.
func (sva PassKeyAuthenticator) Initialize(
	data []byte,
) (iface.Authenticator, error) {
	pk, ok := sva.PubKey.(*secp256r1.PubKey)
	if !ok {
		return sva, fmt.Errorf("pubkic key cannot be cast as secp256r1.PubKey type")
	}
	err := pk.Key.Unmarshal(data)
	if err != nil {
		return sva, err
	}
	return sva, nil
}

// PassKeySignatureData is used to package all the signature data and the tx
// for use in the Authenticate function
type PassKeySignatureData = SignatureData

// GetAuthenticationData parses the signers and signatures from a transactiom
// then returns a indexed list of both signers and signatures
// NOTE: position in the array is used to associate the signer and signature
func (sva PassKeyAuthenticator) GetAuthenticationData(
	ctx sdk.Context,
	tx sdk.Tx,
	messageIndex int,
	simulate bool,
) (iface.AuthenticatorData, error) {
	signers, signatures, signingTx, err := GetCommonAuthenticationData(ctx, tx, messageIndex, simulate)
	if err != nil {
		return SignatureData{}, err
	}

	// Get the signature for the message at msgIndex
	return SignatureData{
		Signers:    signers,
		Signatures: signatures,
		Tx:         signingTx,
		Simulate:   simulate,
	}, nil
}

// Authenticate takes a SignaturesVerificationData struct and validates
// each signer and signature using  signature verification
func (sva PassKeyAuthenticator) Authenticate(
	ctx sdk.Context,
	account sdk.AccAddress,
	msg sdk.Msg,
	authenticationData iface.AuthenticatorData,
) iface.AuthenticationResult {
	// Retrieve pubkey we use either the public key from the authenticator store
	// the public key is added to the sva struct by the Initialize function
	pubKey := sva.PubKey

	return Authenticate(
		ctx,
		msg,
		authenticationData,
		sva.ak,
		pubKey,
		sva.Handler,
	)
}

func Authenticate(
	ctx sdk.Context,
	msg sdk.Msg,
	authenticationData iface.AuthenticatorData,
	ak authante.AccountKeeper,
	pubKey cryptotypes.PubKey,
	handler authsigning.SignModeHandler,
) iface.AuthenticationResult {
	verificationData, ok := authenticationData.(SignatureData)
	if !ok {
		return iface.Rejected("invalid signature verification data", sdkerrors.ErrInvalidType)
	}

	// First consume gas for verifing the signature
	params := ak.GetParams(ctx)
	for _, sig := range verificationData.Signatures {
		err := authante.DefaultSigVerificationGasConsumer(ctx.GasMeter(), sig, params)
		if err != nil {
			return iface.Rejected("couldn't get gas consumer", err)
		}
	}
	// after gas consumption continue to verify signatures
	for i, sig := range verificationData.Signatures {
		acc, err := authante.GetSignerAcc(ctx, ak, verificationData.Signers[i])
		if err != nil {
			return iface.Rejected("couldn't get signer account", err)
		}

		if !verificationData.Simulate && pubKey == nil {
			return iface.Rejected("pubkey on not set on account or authenticator", sdkerrors.ErrInvalidPubKey)
		}

		// Check account sequence number.
		if sig.Sequence != acc.GetSequence() {
			return iface.Rejected(
				fmt.Sprintf("account sequence mismatch, expected %d, got %d", acc.GetSequence(), sig.Sequence),
				sdkerrors.ErrInvalidPubKey,
			)
		}

		// Retrieve and build the signer data struct
		genesis := ctx.BlockHeight() == 0
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
				handler,
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
				// return NotAuthenticated()
				return iface.NotAuthenticated()
			}
		}
	}
	return iface.Authenticated()
}

func (sva PassKeyAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg) error {
	return nil
}

func (sva PassKeyAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	// We allow users to pass a valid public key for signature verification.
	// we assume secp256r1.PubKeySize == secp256r1.PublicKey
	if len(data) != secp256k1.PubKeySize {
		return fmt.Errorf("invalid secp256r1 public key size, expected %d, got %d", secp256k1.PubKeySize, len(data))
	}
	return nil
}

func (sva PassKeyAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	return nil
}

func (sva PassKeyAuthenticator) ConfirmExecution(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticationData iface.AuthenticatorData) iface.ConfirmationResult {
	return iface.Confirm()
}
