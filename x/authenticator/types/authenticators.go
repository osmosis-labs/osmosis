package types

// ToDo: consider moving to a different package

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

type Authenticator[T any] interface {
	GetAuthenticationData(tx sdk.Tx, messageIndex uint8) T
	Authenticate(ctx sdk.Context, msg sdk.Msg, authenticationData T) bool
	ConfirmExecution(ctx sdk.Context, msg sdk.Msg, authenticated bool, authenticationData T) bool

	// Optional Hooks. ToDo: Revisit this when adding the authenticator storage and messages
	//OnAuthenticatorAdded(...) bool
	//OnAuthenticatorRemoved(...) bool
}

type SigVerificationAuthenticator struct {
	//accountKeeper   authsigning.AccountKeeper
	//signModeHandler *txsigning.HandlerMap
}

// These keepers will probably be needed later to validate the signatures
//func (c SigVerificationAuthenticator) SetAccountKeeper(ak authsigning.AccountKeeper) {
//	c.accountKeeper = ak
//}
//
//func (c SigVerificationAuthenticator) SetSignModeHandler(sm *txsigning.HandlerMap) {
//	c.signModeHandler = sm
//}

var _ Authenticator[SigVerificationData] = &SigVerificationAuthenticator{}

type SigVerificationData struct {
	Signer    []byte
	Signature signing.SignatureV2
	// ToDo: we will probably need to provide the whole Tx's data here as signatures are over the whole tx
}

func (c SigVerificationAuthenticator) GetAuthenticationData(tx sdk.Tx, msgIndex uint8) SigVerificationData {
	sigTx, ok := tx.(authsigning.Tx)
	if !ok {
		return SigVerificationData{}
	}

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	signatures, err := sigTx.GetSignaturesV2()
	if err != nil {
		return SigVerificationData{}
	}

	signers := sigTx.GetSigners()

	// check that signer length and signature length are the same
	if len(signatures) != len(signers) {
		return SigVerificationData{}
	}

	if msgIndex >= uint8(len(signatures)) {
		return SigVerificationData{}
	}

	// Get the signature for the message at msgIndex
	return SigVerificationData{
		Signer:    signers[msgIndex],
		Signature: signatures[msgIndex], // ToDo: better marshaling
	}
}

func (c SigVerificationAuthenticator) Authenticate(ctx sdk.Context, msg sdk.Msg, authenticationData SigVerificationData) bool {
	// TODO: Use signature verification abstraction here
	return true
}

func (c SigVerificationAuthenticator) ConfirmExecution(ctx sdk.Context, msg sdk.Msg, authenticated bool, authenticationData SigVerificationData) bool {
	// To be executed in the post handler
	return true
}
