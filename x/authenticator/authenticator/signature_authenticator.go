package authenticator

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v23/x/authenticator/iface"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// Compile time type assertion for the SignatureData using the
// SignatureVerificationAuthenticator struct
var _ iface.Authenticator = &SignatureVerificationAuthenticator{}

const (
	// SignatureVerificationAuthenticatorType represents a type of authenticator specifically designed for
	// secp256k1 signature verification.
	SignatureVerificationAuthenticatorType = "SignatureVerificationAuthenticator"
)

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
// when the app starts a SignatureVerificationAuthenticator is passed to the authentication manager,
// it will only be instantiated once then used for each signature authentication.
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
// which should have a public key in the data field.
func (sva SignatureVerificationAuthenticator) Initialize(
	data []byte,
) (iface.Authenticator, error) {
	if len(data) != secp256k1.PubKeySize {
		sva.PubKey = nil
	}
	sva.PubKey = &secp256k1.PubKey{Key: data}
	return sva, nil
}

// Authenticate takes a SignaturesVerificationData struct and validates
// each signer and signature using  signature verification
func (sva SignatureVerificationAuthenticator) Authenticate(ctx sdk.Context, request iface.AuthenticationRequest) iface.AuthenticationResult {
	// First consume gas for verifying the signature
	params := sva.ak.GetParams(ctx)
	ctx.GasMeter().ConsumeGas(params.SigVerifyCostSecp256k1, "secp256k1 signature verification")

	// Retrieve pubkey we use either the public key from the authenticator store
	// if that's not available query the original auth store for the public key
	// the public key is added to the sva struct by the Initialize function
	pubKey := sva.PubKey
	if pubKey == nil {
		// Having a default here keeps this authenticator stateless,
		// that way we don't have to create specific authenticators with the pubkey of each existing account
		acc, err := authante.GetSignerAcc(ctx, sva.ak, request.Account)
		if err != nil {
			return iface.Rejected("couldn't get signer account", err)
		}
		pubKey = acc.GetPubKey()
	}

	// after gas consumption continue to verify signatures
	if !request.Simulate && pubKey == nil {
		return iface.Rejected("pubkey on not set on account or authenticator", sdkerrors.ErrInvalidPubKey)
	}

	// No need to verify signatures on recheck tx
	if !request.Simulate && !ctx.IsReCheckTx() {
		if !pubKey.VerifySignature(request.SignModeTxData.Direct, request.Signature) {
			ctx.Logger().Debug(
				"signature verification failed; please verify account number (%d), sequence (%d) and chain-id (%s)",
				request.TxData.AccountNumber,
				request.TxData.AccountSequence,
				request.TxData.ChainID,
			)
			return iface.NotAuthenticated()
		}
	}
	return iface.Authenticated()
}

func (sva SignatureVerificationAuthenticator) Track(ctx sdk.Context, account sdk.AccAddress, msg sdk.Msg, authenticatorId string) error {
	return nil
}

func (sva SignatureVerificationAuthenticator) ConfirmExecution(ctx sdk.Context, request iface.AuthenticationRequest) iface.ConfirmationResult {
	return iface.Confirm()
}

func (sva SignatureVerificationAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	// We allow users to pass no data or a valid public key for signature verification.
	// Users can pass no data if the public key is already contained in the auth store.
	if len(data) == 0 || len(data) != secp256k1.PubKeySize {
		return fmt.Errorf("invalid secp256k1 public key size, expected %d, got %d", secp256k1.PubKeySize, len(data))
	}
	return nil
}

func (sva SignatureVerificationAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte) error {
	return nil
}
