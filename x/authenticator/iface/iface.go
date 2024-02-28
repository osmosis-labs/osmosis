package iface

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SignModeData struct {
	Direct  []byte `json:"sign_mode_direct"`
	Textual string `json:"sign_mode_textual"`
}

type LocalAny struct {
	TypeURL string `json:"type_url"`
	Value   []byte `json:"value"`
}

type ExplicitTxData struct {
	ChainID         string     `json:"chain_id"`
	AccountNumber   uint64     `json:"account_number"`
	AccountSequence uint64     `json:"sequence"`
	TimeoutHeight   uint64     `json:"timeout_height"`
	Msgs            []LocalAny `json:"msgs"`
	Memo            string     `json:"memo"`
}

type SimplifiedSignatureData struct {
	Signers    []sdk.AccAddress `json:"signers"`
	Signatures [][]byte         `json:"signatures"`
}

type AuthenticationRequest struct {
	AuthenticatorId string         `json:"authenticator_id"`
	Account         sdk.AccAddress `json:"account"`
	FeePayer        sdk.AccAddress `json:"fee_payer"`
	Msg             LocalAny       `json:"msg"`

	// Since array size is int, and size depends on the system architecture,
	// we use uint64 to cover all available architectures.
	// It is unsigned, so at this point, it can't be negative.
	MsgIndex uint64 `json:"msg_index"`

	// Only allowing messages with a single signer
	Signature           []byte                  `json:"signature"`
	SignModeTxData      SignModeData            `json:"sign_mode_tx_data"`
	TxData              ExplicitTxData          `json:"tx_data"`
	SignatureData       SimplifiedSignatureData `json:"signature_data"`
	Simulate            bool                    `json:"simulate"`
	AuthenticatorParams []byte                  `json:"authenticator_params,omitempty"`
}

// Authenticator is an interface employed to encapsulate all authentication functionalities essential for
// verifying transactions, paying transaction fees, and managing gas consumption during verification.
type Authenticator interface {
	// Type defines the various types of authenticators, such as SignatureVerificationAuthenticator
	// or CosmWasmAuthenticator. Each authenticator type must be registered within the AuthenticatorManager,
	// and these types are used to link the data structure with the authenticator's logic.
	Type() string

	// StaticGas specifies the gas consumption enforced on each call to the authenticator.
	StaticGas() uint64

	// Initialize is used when an authenticator associated with an account is retrieved
	// from storage. The data stored for each (account, authenticator) pair is provided
	// to this method. For instance, the SignatureVerificationAuthenticator requires a PublicKey
	// for signature verification, but the auth module is unaware of the public key. By storing
	// the public key alongside the authenticator, we can Initialize() the code with that data,
	// allowing the same authenticator code to verify signatures for different public keys.
	Initialize(data []byte) (Authenticator, error)

	// Authenticate validates a message based on the signer and data parsed from the GetAuthenticationData function.
	// It returns true if authenticated, or false if not authenticated. This function is used within an ante handler.
	// Note: Gas consumption occurs within this function.
	Authenticate(
		ctx sdk.Context,
		request AuthenticationRequest,
	) error

	// Track is used for authenticators to track any information they may need regardless of how the transactions is
	// authenticated. For instance, if a message is authenticated via authz, ICA, or similar, those entrypoints should
	// call authenticator.Track(...) so that the authenticator can know that the account has executed a specific message
	Track(
		ctx sdk.Context, // The SDK Context is used to access data for authentication and to consume gas.
		account sdk.AccAddress, // The account being authenticated (typically msg.GetSigners()[0]).
		msg sdk.Msg, // A message is passed into the authenticate function, allowing authenticators to utilize its information.
		msgIndex uint64, // The index of the message in the transaction.
		authenticatorId string, // The global authenticator id
	) error

	// ConfirmExecution is employed in the post-handler function to enforce transaction rules,
	// such as spending and transaction limits. It accesses the account's owned state to store
	// and verify these values.
	ConfirmExecution(ctx sdk.Context, request AuthenticationRequest) error

	// OnAuthenticatorAdded is called when an authenticator is added to an account. If the data is not properly formatted
	// or the authenticator is not compatible with the account, an error should be returned.
	OnAuthenticatorAdded(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string) error

	// OnAuthenticatorRemoved is called when an authenticator is removed from an account.
	// This can be used to update any global data that the authenticator is tracking or to prevent removal
	// by returning an error.
	// Removal prevention should be used sparingly and only when absolutely necessary.
	OnAuthenticatorRemoved(ctx sdk.Context, account sdk.AccAddress, data []byte, authenticatorId string) error
}

type InitializedAuthenticator struct {
	Id            uint64
	Authenticator Authenticator
}
