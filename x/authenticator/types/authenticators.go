package types

// ToDo: consider moving to a different package

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// ToDo: Use generics for auth data?

type Authenticator[T any] interface {
	GetAuthenticationData(tx sdk.Tx, messageIndex uint8) T
	Authenticate(ctx sdk.Context, msg sdk.Msg, authenticationData T) bool
	ConfirmExecution(ctx sdk.Context, msg sdk.Msg, authenticated bool, authenticationData T) bool

	// Optional Hooks. ToDo: Revisit this when adding the authenticator storage and messages
	//OnAuthenticatorAdded(...) bool
	//OnAuthenticatorRemoved(...) bool
}

// ToDo:  Open Questions
//   * Rename to Authenticate() IsAuthenticated()? I like the name as a verb better, but might not be a best practice
//   * Do we want to enforce that the context has been limited to the authenticators' state before calling the
//    stateful methods? Or should we just leave it up to the caller to ensure that the context is limited?
//   * Sound we add an explicit "account" field to Authenticate() and ConfirmExecution() to represent who we are
//     executing the messages as? The account can be abstracted by the message (GetSigner()) but if we extract it in
//     the caller it may simplify the wasmd case.

type ClassicAuthenticator struct {
	//accountKeeper   authsigning.AccountKeeper
	//signModeHandler *txsigning.HandlerMap
}

// These keepers will probably be needed later to validate the signatures
//func (c ClassicAuthenticator) SetAccountKeeper(ak authsigning.AccountKeeper) {
//	c.accountKeeper = ak
//}
//
//func (c ClassicAuthenticator) SetSignModeHandler(sm *txsigning.HandlerMap) {
//	c.signModeHandler = sm
//}

var _ Authenticator[ClassicAuthData] = &ClassicAuthenticator{}

type ClassicAuthData struct {
	Signer    []byte
	Signature signing.SignatureV2
	// ToDo: we will probably need to provide the whole Tx's data here as signatures are over the whole tx
}

func (c ClassicAuthenticator) GetAuthenticationData(tx sdk.Tx, msgIndex uint8) ClassicAuthData {
	sigTx, ok := tx.(authsigning.Tx)
	if !ok {
		return ClassicAuthData{}
	}

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	signatures, err := sigTx.GetSignaturesV2()
	if err != nil {
		return ClassicAuthData{}
	}

	signers := sigTx.GetSigners()

	// check that signer length and signature length are the same
	if len(signatures) != len(signers) {
		return ClassicAuthData{}
	}

	if msgIndex >= uint8(len(signatures)) {
		return ClassicAuthData{}
	}

	// Get the signature for the message at msgIndex
	return ClassicAuthData{
		Signer:    signers[msgIndex],
		Signature: signatures[msgIndex], // ToDo: better marshaling
	}
}

func (c ClassicAuthenticator) Authenticate(ctx sdk.Context, msg sdk.Msg, authenticationData ClassicAuthData) bool {
	// TODO: Use signature verification abstraction here
	return true
}

func (c ClassicAuthenticator) ConfirmExecution(ctx sdk.Context, msg sdk.Msg, authenticated bool, authenticationData ClassicAuthData) bool {
	// To be executed in the post handler
	return true
}
