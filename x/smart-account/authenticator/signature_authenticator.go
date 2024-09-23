package authenticator

import (
	"fmt"

	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Compile time type assertion for the SignatureData using the
// SignatureVerification struct
var _ Authenticator = &SignatureVerification{}

const (
	// SignatureVerificationType represents a type of authenticator specifically designed for
	// secp256k1 signature verification.
	SignatureVerificationType = "SignatureVerification"
)

// signature authenticator
type SignatureVerification struct {
	ak     authante.AccountKeeper
	PubKey cryptotypes.PubKey
}

func (sva SignatureVerification) Type() string {
	return SignatureVerificationType
}

func (sva SignatureVerification) StaticGas() uint64 {
	// using 0 gas here. The gas is consumed based on the pubkey type in Authenticate()
	return 0
}

// NewSignatureVerification creates a new SignatureVerification
func NewSignatureVerification(ak authante.AccountKeeper) SignatureVerification {
	return SignatureVerification{ak: ak}
}

// Initialize sets up the public key to the data supplied from the account-authenticator configuration
func (sva SignatureVerification) Initialize(config []byte) (Authenticator, error) {
	if len(config) != secp256k1.PubKeySize {
		sva.PubKey = nil
	}
	sva.PubKey = &secp256k1.PubKey{Key: config}
	return sva, nil
}

// Authenticate takes a SignaturesVerificationData struct and validates
// each signer and signature using signature verification
func (sva SignatureVerification) Authenticate(ctx sdk.Context, request AuthenticationRequest) error {
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

func (sva SignatureVerification) Track(ctx sdk.Context, request AuthenticationRequest) error {
	return nil
}

func (sva SignatureVerification) ConfirmExecution(ctx sdk.Context, request AuthenticationRequest) error {
	return nil
}

func (sva SignatureVerification) OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error {
	// We allow users to pass no data or a valid public key for signature verification.
	if len(config) != secp256k1.PubKeySize {
		return fmt.Errorf("invalid secp256k1 public key size, expected %d, got %d", secp256k1.PubKeySize, len(config))
	}
	return nil
}

func (sva SignatureVerification) OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, config []byte, authenticatorId string) error {
	return nil
}
