package authenticator

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Compile time type assertion for the SignatureData using the
// SignatureVerificationAuthenticator struct
var _ Authenticator = &SignatureVerificationAuthenticator{}

const (
	// SignatureVerificationAuthenticatorType represents a type of authenticator specifically designed for
	// secp256k1 signature verification.
	SignatureVerificationAuthenticatorType = "SignatureVerificationAuthenticator"
)

// signature authenticator
type SignatureVerificationAuthenticator struct {
	ak     *authkeeper.AccountKeeper
	PubKey cryptotypes.PubKey
}

func (sva SignatureVerificationAuthenticator) Type() string {
	return SignatureVerificationAuthenticatorType
}

func (sva SignatureVerificationAuthenticator) StaticGas() uint64 {
	// using 0 gas here. The gas is consumed based on the pubkey type in Authenticate()
	return 0
}

// NewSignatureVerificationAuthenticator creates a new SignatureVerificationAuthenticator
func NewSignatureVerificationAuthenticator(ak *authkeeper.AccountKeeper) SignatureVerificationAuthenticator {
	return SignatureVerificationAuthenticator{ak: ak}
}

// Initialize sets up the public key to the data supplied from the account-authenticator configuration
func (sva SignatureVerificationAuthenticator) Initialize(data []byte) (Authenticator, error) {
	if len(data) != secp256k1.PubKeySize {
		sva.PubKey = nil
	}
	sva.PubKey = &secp256k1.PubKey{Key: data}
	return sva, nil
}

// Authenticate takes a SignaturesVerificationData struct and validates
// each signer and signature using  signature verification
func (sva SignatureVerificationAuthenticator) Authenticate(ctx sdk.Context, request AuthenticationRequest) error {
	// First consume gas for verifying the signature
	params := sva.ak.GetParams(ctx)
	ctx.GasMeter().ConsumeGas(params.SigVerifyCostSecp256k1, "secp256k1 signature verification")
	// after gas consumption continue to verify signatures

	if request.Simulate || ctx.IsReCheckTx() {
		return nil
	}
	if sva.PubKey == nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidPubKey, "pubkey on not set on account or authenticator")
	}

	if !sva.PubKey.VerifySignature(request.SignModeTxData.Direct, request.Signature) {
		return errorsmod.Wrapf(
			sdkerrors.ErrUnauthorized,
			"signature verification failed; please verify account number (%d), sequence (%d) and chain-id (%s)",
			request.TxData.AccountNumber,
			request.TxData.AccountSequence,
			request.TxData.ChainID,
		)
	}
	return nil
}

func (sva SignatureVerificationAuthenticator) Track(ctx sdk.Context, request AuthenticationRequest) error {
	return nil
}

func (sva SignatureVerificationAuthenticator) ConfirmExecution(ctx sdk.Context, request AuthenticationRequest) error {
	return nil
}

func (sva SignatureVerificationAuthenticator) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string) error {
	// We allow users to pass no data or a valid public key for signature verification.
	// Users can pass no data if the public key is already contained in the auth store.
	if len(data) != secp256k1.PubKeySize {
		return fmt.Errorf("invalid secp256k1 public key size, expected %d, got %d", secp256k1.PubKeySize, len(data))
	}
	return nil
}

func (sva SignatureVerificationAuthenticator) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string) error {
	return nil
}
