package authenticator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// SignatureVerificationAuthenticator is a Type of authenticator that specific deals with
	// secp256k1 signatures,
	SignatureVerificationAuthenticatorType = "SignatureVerificationAuthenticator"
)

// AuthenticatorData represents the data needed to verify a signer address and message signature
// TODO: make this less like a void pointer
type AuthenticatorData interface{}

// Authenticator is an interface used to represent all the authentication functionality needed to
// verifiy a transaction, pay fees for a transaction and consume gas for verifing a transaction
type Authenticator interface {
	// Type() defines the different types of authenticator, e.g SignatureVerificationAuthenticator
	// or CosmWasmAuthenticator. Each type of authenticator needs to be registered in the AuthenticatorManager
	// and these Types are used to store and link the data structure to the Authenticator logic
	Type() string

	// StaticGas defines what gas the authenticator uses per signatures verification
	StaticGas() uint64

	// Initialize is used when the authenticator associated to an account is retrieved
	// from the store. The data stored for each (account, authenticator) pair will be
	// passed to this method.
	// For example, the SignatureVerificationAuthenticator requires a PublicKey to check
	// the signature, but the auth module is not aware of the public key.
	// By storing the public key along with the authenticator, we can Initialize() the code with that
	// data, which allows the same authenticator code to verify signatures for different public keys
	Initialize(data []byte) (Authenticator, error)

	// GetAuthenticationData gets any authentication data needed from a transaction
	// it returns an interface that is defined as a concrete type by the implementer of the interface.
	// This is used in an ante handler with Authenticate to ensure the user has correct permission to execute
	// a message.
	GetAuthenticationData(
		ctx sdk.Context, // sdk Context is used to get data associated with the authentication data
		tx sdk.Tx, // we pass the transaction into the getter function to parse the signatures and signers
		messageIndex int8, // the message index is used to pull specific signers and signatures from the authentication data
		simulate bool, // simulate is used to simulate transactions
	) (AuthenticatorData, error)

	// Authenticate authenticates a message based on the signer and data parsed from the GetAuthenticationData function
	// the returns true is authenticated or false if not authicated. This is used in an ante handler.
	// NOTE: Consume gas per signature happens in this function.
	Authenticate(
		ctx sdk.Context, // sdk Context is used to get data for use in authentication and to consume gas
		msg sdk.Msg, // a msg is passed into the authenticate function to allow the authentication data to verify the signature
		authenticationData AuthenticatorData, // The authentication data is used to authenticate a message
	) (bool, error)

	// AuthenticationFailed TODO: define
	AuthenticationFailed(
		ctx sdk.Context, // TODO: define
		authenticatorData AuthenticatorData, // TODO: define
		msg sdk.Msg, // TODO: define
	)

	// ConfirmExecution is used in the post handler function to enable transaction rules to be enforces.
	// Rules such as spend and transaction limits. We access the state owned by the account to store and check these values.
	ConfirmExecution(
		ctx sdk.Context, // sdk context is used to set and get account authenticator state
		msg sdk.Msg, // TODO: the message is passed here to check invariants
		authenticated bool, // TODO: define
		authenticationData AuthenticatorData, // TODO: define
	) bool

	// Optional Hooks. TODO: Revisit this when adding the authenticator storage and messages
	// OnAuthenticatorAdded(...) bool
	// OnAuthenticatorRemoved(...) bool
}
