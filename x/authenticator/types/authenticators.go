package types

// TODO: consider moving to a different package

import (
	fmt "fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

type AuthenticatorData interface{}

type Authenticator interface {
	Type() string
	Initialize(data []byte) (Authenticator, error)
	GetAuthenticationData(tx sdk.Tx, messageIndex uint8, simulate bool) (AuthenticatorData, error)
	Authenticate(ctx sdk.Context, msg sdk.Msg, authenticationData AuthenticatorData) (bool, error)
	ConfirmExecution(ctx sdk.Context, msg sdk.Msg, authenticated bool, authenticationData AuthenticatorData) bool

	// Optional Hooks. ToDo: Revisit this when adding the authenticator storage and messages
	//OnAuthenticatorAdded(...) bool
	//OnAuthenticatorRemoved(...) bool
}

// Compile time type assertion for the SigVerificationData using the
// SigVerificationAuthenticator struct
var _ Authenticator = &SigVerificationAuthenticator{}
var _ AuthenticatorData = &SigVerificationData{}

// Secp256k1 signature authenticator
type SigVerificationAuthenticator struct {
	ak      *authkeeper.AccountKeeper
	Handler authsigning.SignModeHandler
	PubKey  cryptotypes.PubKey
}

func (c SigVerificationAuthenticator) Type() string {
	return "SigVerification"
}

// NewSigVerificationAuthenticator creates a new SigVerificationAuthenticator
func NewSigVerificationAuthenticator(
	ak *authkeeper.AccountKeeper,
	Handler authsigning.SignModeHandler,
) SigVerificationAuthenticator {
	return SigVerificationAuthenticator{
		ak:      ak,
		Handler: Handler,
	}
}

// SetAccountKeeper sets the account keeper one the SigVerificationAuthenticator
func (c SigVerificationAuthenticator) SetAccountKeeper(ak *authkeeper.AccountKeeper) {
	c.ak = ak
}

// SetAccountKeeper sets the sign mode one the SigVerificationAuthenticator
func (c SigVerificationAuthenticator) SetSignModeHandler(sm *authsigning.SignModeHandler) {
	c.Handler = *sm
}

func (c SigVerificationAuthenticator) Initialize(data []byte) (Authenticator, error) {
	// TODO: I'm concerned about this method. I think we should modify the design here so that these authenticators are
	//       always new objects. Maybe the registered authenticators could be a factory for the actual authenticator.
	if len(data) != secp256k1.PubKeySize {
		c.PubKey = nil
	}
	c.PubKey = &secp256k1.PubKey{Key: data}
	return c, nil
}

// SigVerificationData is used to package all the signature data and the tx
// for use in the Authenticate function
type SigVerificationData struct {
	Signers    []sdk.AccAddress
	Signatures []signing.SignatureV2
	Tx         authsigning.Tx
	Simulate   bool
}

func GetSignersAndSignatures(
	msgs []sdk.Msg,
	allSignatures []signing.SignatureV2,
	feePayer string,
	msgIndex int,
) ([]sdk.AccAddress, []signing.SignatureV2, error) {
	if msgIndex < -1 || msgIndex >= len(msgs) {
		return nil, nil, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid message ID")
	}
	useMsgIndex := msgIndex != -1

	seen := map[string]bool{}
	uniqueSignersCount := 0

	var signersList []sdk.AccAddress
	var signaturesList []signing.SignatureV2

	// Loop through the messages and their signers
	for i, msg := range msgs {
		for _, signer := range msg.GetSigners() {
			signerStr := signer.String()
			if !seen[signerStr] {
				seen[signerStr] = true
				uniqueSignersCount++
				if !useMsgIndex || i == msgIndex {
					signersList = append(signersList, signer)
					signaturesList = append(signaturesList, allSignatures[uniqueSignersCount-1])
					if useMsgIndex {
						break
					}
				}
			}
		}
	}

	// Add FeePayer if not already included
	if feePayer != "" && !seen[feePayer] {
		if uniqueSignersCount != len(allSignatures)-1 {
			//TODO: Better error?
			return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "invalid number of signer;  expected: %d, got %d", uniqueSignersCount, len(allSignatures)-1)
		}
		feePayerAddr, err := sdk.AccAddressFromBech32(feePayer)
		if err != nil {
			return nil, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid fee payer address")
		}
		signersList = append(signersList, feePayerAddr)
		signaturesList = append(signaturesList, allSignatures[len(allSignatures)-1]) // FeePayer's signature is last
	}

	if uniqueSignersCount != len(allSignatures) {
		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "mismatch between signers and signatures;  expected: %d, got %d", len(signersList), len(signaturesList))
	}

	return signersList, signaturesList, nil
}

// GetAuthenticationData parses the signers and signatures from a transactiom
// then returns a indexed list of both signers and signatures
// NOTE: position in the array is used to associate the signer and signature
func (c SigVerificationAuthenticator) GetAuthenticationData(
	tx sdk.Tx,
	messageIndex uint8,
	simulate bool,
) (AuthenticatorData, error) {
	sigTx, ok := tx.(authsigning.Tx)
	if !ok {
		return SigVerificationData{},
			sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	signatures, err := sigTx.GetSignaturesV2()
	if err != nil {
		return SigVerificationData{}, err
	}

	msgs := sigTx.GetMsgs()
	msgSigners, msgSignatures, err := GetSignersAndSignatures(msgs, signatures, "", int(messageIndex)) // TODO: deal with feepayer
	if err != nil {
		return SigVerificationData{}, err
	}

	// Get the signature for the message at msgIndex
	return SigVerificationData{
		Signers:    msgSigners,
		Signatures: msgSignatures,
		Tx:         sigTx,
		Simulate:   simulate,
	}, nil
}

// Authenticate takes a SignaturesVerificationData struct and validates
// each signer and signature using Secp256k1 signature verification
func (c SigVerificationAuthenticator) Authenticate(
	ctx sdk.Context,
	msg sdk.Msg,
	authenticationData AuthenticatorData,
) (success bool, err error) {
	verificationData, ok := authenticationData.(SigVerificationData)
	if !ok {
		return false, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "invalid signature verification data")
	}

	for i, sig := range verificationData.Signatures {
		acc, err := authante.GetSignerAcc(ctx, c.ak, verificationData.Signers[i])
		if err != nil {
			return false, err
		}

		// retrieve pubkey
		pubKey := c.PubKey
		if pubKey == nil {
			// Having a default here keeps this authenticator stateless,
			// that way we don't have to create specific authenticators with the pubkey of each existing account

			pubKey = acc.GetPubKey() // TODO: do we want this default?
		}
		if !verificationData.Simulate && pubKey == nil {
			return false, sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, "pubkey on not set on account or authenticator")
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
				return false, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, errMsg)
			}
		}
	}
	return true, nil
}

func (c SigVerificationAuthenticator) ConfirmExecution(ctx sdk.Context, msg sdk.Msg, authenticated bool, authenticationData AuthenticatorData) bool {
	// To be executed in the post handler
	return true
}
