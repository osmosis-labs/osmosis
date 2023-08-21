package types

// TODO: consider moving to a different package

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

type Authenticator[T any] interface {
	GetAuthenticationData(tx sdk.Tx) (T, error)
	Authenticate(ctx sdk.Context, authenticationData T, simulate bool) (err error)
	ConfirmExecution(ctx sdk.Context, msg sdk.Msg, authenticated bool, authenticationData T) bool

	// Optional Hooks. ToDo: Revisit this when adding the authenticator storage and messages
	//OnAuthenticatorAdded(...) bool
	//OnAuthenticatorRemoved(...) bool
}

// Compile time type assertion for the SigVerificationData using the
// SigVerificationAuthenticator struct
var _ Authenticator[SigVerificationData] = &SigVerificationAuthenticator{}

// Secp256k1 signature authenticator
type SigVerificationAuthenticator struct {
	ak      authante.AccountKeeper
	Handler authsigning.SignModeHandler
}

// NewSigVerificationAuthenticator creates a new SigVerificationAuthenticator
func NewSigVerificationAuthenticator(
	ak authante.AccountKeeper,
	Handler authsigning.SignModeHandler,
) SigVerificationAuthenticator {
	return SigVerificationAuthenticator{
		ak:      ak,
		Handler: Handler,
	}
}

// SetAccountKeeper sets the account keeper one the SigVerificationAuthenticator
func (c SigVerificationAuthenticator) SetAccountKeeper(ak authante.AccountKeeper) {
	c.ak = ak
}

// SetAccountKeeper sets the sign mode one the SigVerificationAuthenticator
func (c SigVerificationAuthenticator) SetSignModeHandler(sm *authsigning.SignModeHandler) {
	c.Handler = *sm
}

// SigVerificationData is used to package all the signature data and the tx
// for use in the Authenticate function
type SigVerificationData struct {
	Signers    []sdk.AccAddress
	Signatures []signing.SignatureV2
	Tx         authsigning.Tx
}

// GetAuthenticationData parses the signers and signatures from a transactiom
// then returns a indexed list of both signers and signatures
// NOTE: position in the array is used to associate the signer and signature
func (c SigVerificationAuthenticator) GetAuthenticationData(tx sdk.Tx) (SigVerificationData, error) {
	sigTx, ok := tx.(authsigning.Tx)
	if !ok {
		return SigVerificationData{},
			sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// Get all signers for a transaction
	signers := sigTx.GetSigners()

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	signatures, err := sigTx.GetSignaturesV2()
	if err != nil {
		return SigVerificationData{}, err
	}

	// check that signer length and signature length are the same
	if len(signatures) != len(signers) {
		return SigVerificationData{},
			sdkerrors.Wrapf(
				sdkerrors.ErrUnauthorized,
				"invalid number of signer;  expected: %d, got %d",
				len(signers),
				len(signatures))
	}

	// Get the signature for the message at msgIndex
	return SigVerificationData{
		Signers:    signers,
		Signatures: signatures,
		Tx:         sigTx,
	}, nil
}

// Authenticate takes a SignaturesVerificationData struct and validates
// each signer and signature using Secp256k1 signature verification
func (c SigVerificationAuthenticator) Authenticate(
	ctx sdk.Context,
	verificationData SigVerificationData,
	simulate bool,
) (err error) {
	for i, sig := range verificationData.Signatures {
		acc, err := authante.GetSignerAcc(ctx, c.ak, verificationData.Signers[i])
		if err != nil {
			return err
		}

		// retrieve pubkey
		pubKey := acc.GetPubKey()
		if !simulate && pubKey == nil {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, "pubkey on account is not set")
		}

		// Check account sequence number.
		if sig.Sequence != acc.GetSequence() {
			return sdkerrors.Wrapf(
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
		if !simulate && !ctx.IsReCheckTx() {
			err := authsigning.VerifySignature(pubKey, signerData, sig.Data, c.Handler, verificationData.Tx)
			if err != nil {
				var errMsg string
				if authante.OnlyLegacyAminoSigners(sig.Data) {
					// If all signers are using SIGN_MODE_LEGACY_AMINO, we rely on VerifySignature to check account sequence number,
					// and therefore communicate sequence number as a potential cause of error.
					errMsg = fmt.Sprintf(
						"signature verification failed; please verify account number (%d), sequence (%d) and chain-id (%s)",
						accNum,
						acc.GetSequence(),
						chainID,
					)
				} else {
					errMsg = fmt.Sprintf("signature verification failed; please verify account number (%d) and chain-id (%s)",
						accNum,
						chainID,
					)
				}
				return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, errMsg)

			}
		}
	}

	return
}

func (c SigVerificationAuthenticator) ConfirmExecution(ctx sdk.Context, msg sdk.Msg, authenticated bool, authenticationData SigVerificationData) bool {
	// To be executed in the post handler
	return true
}
