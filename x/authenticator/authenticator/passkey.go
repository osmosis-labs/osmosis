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

// Compile time type assertion for the  PassKeyAuthenticator struct
var _ iface.Authenticator = &PassKeyAuthenticator{}

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

// Authenticate takes a SignaturesVerificationData struct and validates
// each signer and signature using  signature verification
func (sva PassKeyAuthenticator) Authenticate(ctx sdk.Context, request iface.AuthenticationRequest) iface.AuthenticationResult {
	// Retrieve pubkey we use either the public key from the authenticator store
	// the public key is added to the sva struct by the Initialize function
	pubKey := sva.PubKey

	// TODO: Why is this a separate function?
	return Authenticate(
		ctx,
		request,
		sva.ak,
		pubKey,
	)
}

func Authenticate(
	ctx sdk.Context,
	request iface.AuthenticationRequest,
	ak authante.AccountKeeper,
	pubKey cryptotypes.PubKey,
) iface.AuthenticationResult {
	// First consume gas for verifing the signature
	params := ak.GetParams(ctx)
	ctx.GasMeter().ConsumeGas(params.SigVerifyCostSecp256r1(), "secp256r1 signature verification")

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
				request.TxData.Sequence,
				request.TxData.ChainID,
			)
			return iface.NotAuthenticated()
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

func (sva PassKeyAuthenticator) ConfirmExecution(ctx sdk.Context, request iface.AuthenticationRequest) iface.ConfirmationResult {
	return iface.Confirm()
}
